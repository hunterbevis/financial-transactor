const BASE_URL = 'http://localhost:8080';

export function submitTransactions(count: number) {
	return fetch(`${BASE_URL}/submit`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ count })
	});
}

export function resizeWorkers(workers: number) {
	return fetch(`${BASE_URL}/resize`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ workers })
	});
}

export function resetEngine() {
	return fetch(`${BASE_URL}/reset`, { method: 'POST' });
}

// new: update the synthetic work delay
export function setLatency(ms: number) {
	return fetch(`${BASE_URL}/latency`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ latency: ms })
	});
}
