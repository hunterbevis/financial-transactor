import { writable } from 'svelte/store';

export type Metrics = {
	goroutines: number;
	worker_threads: number; // physical/max thread capacity
	worker_count: number; // active worker pool size from slider
	total_tx: number;
	processed_tx: number;
	failed_tx: number;
	in_flight_tx: number;
	balances: Record<string, number>;
	last_batch_ms: number;
	latency_ms: number;
	worker_status: boolean[];
	active_shards: number[];
	isConnected: boolean;
};

export const metrics = writable<Metrics>({
	goroutines: 0,
	worker_threads: 0,
	worker_count: 0,
	total_tx: 0,
	processed_tx: 0,
	failed_tx: 0,
	in_flight_tx: 0,
	balances: {},
	last_batch_ms: 0,
	latency_ms: 1,
	worker_status: [],
	active_shards: [],
	isConnected: false
});
