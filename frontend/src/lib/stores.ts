import { derived, writable } from 'svelte/store';

export type metricsdata = {
	cpu_threads: number;
	goroutines: number;
	worker_pool: number;
	processed: number;
	session_start_value: number | null;
	failed: number;
	queue_len: number;
	queue_cap: number;
	isconnected: boolean;
};

export type txevent = {
	id: number;
	from: number;
	to: number;
	amount: number;
	submitted_by: string;
	ts: number;
};

export const metrics = writable<metricsdata>({
	cpu_threads: 0,
	goroutines: 0,
	worker_pool: 0,
	processed: 0,
	failed: 0,
	session_start_value: null,
	queue_len: 0,
	queue_cap: 10000000,
	isconnected: false
});

export const queue_pressure = derived(metrics, ($m) => {
	if (!$m.queue_cap) return 0;
	return ($m.queue_len / $m.queue_cap) * 100;
});

export const is_congested = derived(queue_pressure, ($p) => $p > 80);

export const max_tx_log = 100;
export const txlog = writable<txevent[]>([]);

export function pushtx(tx: txevent) {
	txlog.update((arr) => {
		arr.push(tx);
		if (arr.length > max_tx_log) {
			arr.shift();
		}
		return arr;
	});
}

export const session_processed = derived(metrics, ($m) => {
	if ($m.session_start_value === null) return 0;
	const diff = $m.processed - $m.session_start_value;
	return diff < 0 ? 0 : diff;
});
