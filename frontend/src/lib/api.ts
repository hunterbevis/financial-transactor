const base_url = 'http://localhost:8080';

async function handleresponse(res: Response) {
	if (res.status === 429) {
		const error = await res.json().catch(() => ({ message: 'rate_limited' }));
		console.error('system congested:', error.message);
		throw new Error('rate_limited');
	}
	if (!res.ok) throw new Error(`http_${res.status}`);
	return res;
}

export async function submittransactions(count: number) {
	// added headers so go's json.NewDecoder(r.Body) works correctly
	const res = await fetch(`${base_url}/submit`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ count })
	});

	return handleresponse(res);
}

export async function resizeworkers(workers: number) {
	const res = await fetch(`${base_url}/resize`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ workers })
	});
	return handleresponse(res);
}

export async function setlatency(ms: number) {
	const res = await fetch(`${base_url}/latency`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ latency: ms })
	});
	return handleresponse(res);
}

export async function resetengine() {
	const res = await fetch(`${base_url}/reset`, {
		method: 'POST'
	});
	return handleresponse(res);
}
