import { metrics, pushtx } from './stores';

const metrics_ws_url = 'ws://localhost:8080/ws/metrics';
const tx_ws_url = 'ws://localhost:8080/ws/tx';

export function connectmetrics(): () => void {
	let ws: WebSocket | null = null;
	let shouldreconnect = true;
	let retrydelay = 500;

	const connect = () => {
		ws = new WebSocket(metrics_ws_url);

		ws.onopen = () => {
			retrydelay = 500;
			metrics.update((m) => ({ ...m, isconnected: true }));
		};

		ws.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				metrics.update((m) => {
					const baseline = m.session_start_value === null ? data.processed : m.session_start_value;

					return {
						...m,
						...data,
						session_start_value: baseline,
						isconnected: true
					};
				});
			} catch (err) {
				console.error('failed to parse metrics:', err);
			}
		};

		ws.onclose = () => {
			metrics.update((m) => ({ ...m, isconnected: false, session_start_value: null }));

			if (shouldreconnect) {
				setTimeout(connect, retrydelay);
				retrydelay = Math.min(retrydelay * 1.5, 5000);
			}
		};

		ws.onerror = () => ws?.close();
	};

	connect();
	return () => {
		shouldreconnect = false;
		ws?.close();
	};
}

export function connecttxstream(): () => void {
	let ws: WebSocket | null = null;
	let shouldreconnect = true;

	const connect = () => {
		ws = new WebSocket(tx_ws_url);

		ws.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);

				if (Array.isArray(data)) {
					for (let i = 0; i < data.length; i++) {
						const tx = data[i];
						if (tx && tx.id !== undefined) {
							pushtx(tx);
						}
					}
				} else if (data && data.id !== undefined) {
					pushtx(data);
				}
			} catch (err) {
				console.error('failed to parse tx:', err);
			}
		};

		ws.onclose = () => {
			if (shouldreconnect) setTimeout(connect, 1000);
		};

		ws.onerror = () => ws?.close();
	};

	connect();
	return () => {
		shouldreconnect = false;
		ws?.close();
	};
}
