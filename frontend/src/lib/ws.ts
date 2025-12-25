import { metrics } from './stores';

const WS_URL = 'ws://localhost:8080/ws/metrics';

export function connectMetrics(): () => void {
	let ws: WebSocket | null = null;
	let shouldReconnect = true;
	let retryDelay = 500;

	const connect = () => {
		ws = new WebSocket(WS_URL);

		ws.onopen = () => {
			retryDelay = 500;
			metrics.update((m) => ({ ...m, isConnected: true }));
		};

		ws.onmessage = (event) => {
			// keep processing even if hidden so the charts don't "jump" when returning
			try {
				const data = JSON.parse(event.data);
				metrics.update((m) => ({ ...m, ...data, isConnected: true }));
			} catch (err) {
				console.error('failed to parse ws message', err);
			}
		};

		ws.onclose = () => {
			metrics.update((m) => ({ ...m, isConnected: false }));
			if (shouldReconnect) {
				setTimeout(connect, retryDelay);
				retryDelay = Math.min(retryDelay * 1.5, 5000);
			}
		};

		ws.onerror = () => ws?.close();
	};

	connect();
	return () => {
		shouldReconnect = false;
		ws?.close();
	};
}
