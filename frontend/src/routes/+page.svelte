<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import {
		Timer,
		Activity,
		Zap,
		ShieldAlert,
		Cpu,
		Lock,
		RotateCcw,
		SlidersHorizontal,
		InfoIcon
	} from '@lucide/svelte';
	import { metrics } from '$lib/stores';
	import { connectMetrics } from '$lib/ws';
	import { submitTransactions, resizeWorkers, resetEngine, setLatency } from '$lib/api';
	import * as HoverCard from '../components/hover-card/index.ts';
	import MetricCard from '../components/metric-card.svelte';

	let canvas: HTMLCanvasElement;
	let workerCount = 8;
	let localLatency = 1; // local state for the slider
	let isDraggingLatency = false;
	let isDraggingWorkers = false; // added this
	let disconnect: (() => void) | undefined;

	const getcolor = (bal: number) => {
		const baseline = 10000;
		if (bal <= 0) return '#000000';
		if (bal === baseline) return '#18181b';
		if (bal < baseline) {
			const diff = baseline - bal;
			if (diff > 8000) return '#450a0a';
			if (diff > 5000) return '#7f1d1d';
			return '#991b1b';
		} else {
			const diff = bal - baseline;
			if (diff >= 10000) return '#6ee7b7';
			if (diff > 5000) return '#10b981';
			return '#064e3b';
		}
	};

	const draw = () => {
		if (!canvas) return;
		const ctx = canvas.getContext('2d');
		if (!ctx) return;
		const cols = 100;
		const rows = 100;
		const cellw = canvas.width / cols;
		const cellh = canvas.height / rows;

		ctx.clearRect(0, 0, canvas.width, canvas.height);
		for (let i = 0; i < 10000; i++) {
			const bal = $metrics.balances[`acct_${i}`] ?? 10000;
			const x = (i % cols) * cellw;
			const y = Math.floor(i / cols) * cellh;
			ctx.fillStyle = getcolor(bal);
			ctx.fillRect(x, y, cellw - 0.5, cellh - 0.5);
		}
	};

	$: if ($metrics.balances) draw();

	// sync latency from backend
	$: if ($metrics.latency_ms !== undefined && !isDraggingLatency) {
		localLatency = $metrics.latency_ms;
	}

	// sync worker count (goroutines slider) from backend
	$: if ($metrics.worker_count !== undefined && !isDraggingWorkers) {
		workerCount = $metrics.worker_count;
	}

	onMount(() => {
		disconnect = connectMetrics();
		const resize = () => {
			if (!canvas) return;
			canvas.width = canvas.offsetWidth;
			canvas.height = canvas.offsetHeight;
			draw();
		};
		window.addEventListener('resize', resize);
		resize();
	});

	onDestroy(() => disconnect?.());
</script>

<div
	class="h-screen w-full overflow-hidden bg-black p-6 font-sans text-zinc-100 selection:bg-emerald-500/30"
>
	<div class="mx-auto flex h-full max-w-7xl flex-col space-y-4">
		<header class="flex shrink-0 items-center justify-between border-b border-zinc-900 pb-4">
			<div class="flex items-center gap-4">
				<div class="space-y-0.5">
					<h1 class="text-xl font-black tracking-tighter text-white uppercase italic">
						financial transactor engine
					</h1>
					<p class="font-mono text-[10px] leading-none tracking-tight text-zinc-600 uppercase">
						high-concurrency atomic ledger
					</p>
				</div>
				<div
					class="flex items-center gap-2 rounded-full border border-zinc-800 bg-zinc-900/50 px-2.5 py-1"
				>
					<div
						class="h-2 w-2 rounded-full {$metrics.isConnected
							? 'animate-pulse bg-emerald-500 shadow-[0_0_8px_#10b981]'
							: 'bg-red-500'}"
					></div>
					<span class="text-[9px] font-black tracking-widest text-zinc-400 uppercase"
						>{$metrics.isConnected ? 'system_online' : 'engine_offline'}</span
					>
				</div>
				<HoverCard.Root openDelay={50} closeDelay={50}>
					<HoverCard.Trigger
						rel="noreferrer noopener"
						class="cursor-pointer transition-colors hover:text-zinc-300"
					>
						<InfoIcon class="size-3" />
					</HoverCard.Trigger>
					<HoverCard.Content
						class="w-80 border-zinc-800 bg-zinc-950 p-3 shadow-xl"
						align="start"
						side="bottom"
						sideOffset={10}
					>
						<div class="flex items-start space-x-2">
							<InfoIcon class="mt-0.5 size-4 shrink-0 text-emerald-500" />
							<div class="space-y-1">
								<h4 class="text-sm font-bold tracking-tight text-white uppercase">system manual</h4>
								<div class="text-xs leading-relaxed text-zinc-400">
									<p>
										Press the <span class="font-bold text-zinc-200">10k</span> or
										<span class="font-bold text-emerald-400">100k TX</span> buttons to flood the go engine.
									</p>
									<ul class="mt-2 space-y-1.5 border-t border-zinc-900 pt-2">
										<li class="flex items-start gap-1.5">
											<span class="text-emerald-500">•</span>
											<span
												>adjust <span class="text-zinc-200">worker_delay</span> to simulate expensive
												I/O or database latency.</span
											>
										</li>
										<li class="flex items-start gap-1.5">
											<span class="text-emerald-500">•</span>
											<span
												>scale <span class="text-zinc-200">goroutines</span> to observe how go handles
												concurrent lock contention.</span
											>
										</li>
									</ul>
								</div>
							</div>
						</div>
					</HoverCard.Content>
				</HoverCard.Root>
			</div>

			<div class="flex gap-2">
				<button
					class="flex items-center gap-2 rounded border border-zinc-800 bg-zinc-900 px-3 py-1.5 text-[10px] font-bold uppercase transition-all hover:border-red-900/50 hover:bg-red-950/30 hover:text-red-400"
					onclick={() => resetEngine()}
				>
					<RotateCcw size={10} /> reset balances
				</button>
				<button
					class="rounded border border-zinc-800 bg-zinc-900 px-3 py-1.5 text-[10px] font-bold uppercase hover:bg-zinc-800"
					onclick={() => submitTransactions(10000)}>+10k tx</button
				>
				<button
					class="rounded bg-emerald-600 px-4 py-1.5 text-[10px] font-black text-white uppercase shadow-lg shadow-emerald-900/20 hover:bg-emerald-500 active:scale-95"
					onclick={() => submitTransactions(100000)}>flood 100k tx</button
				>
			</div>
		</header>

		<div class="grid shrink-0 grid-cols-2 gap-3 lg:grid-cols-7">
			<MetricCard
				label="cpu threads"
				value={$metrics.worker_threads}
				title="OS Threads"
				description="the maximum number of operating system threads the go runtime can use simultaneously (GOMAXPROCS)."
			/>
			<MetricCard
				label="goroutines"
				value={$metrics.goroutines}
				title="Active Routines"
				description="the total number of living goroutines. includes workers, websocket handlers, and background runtime tasks."
			/>
			<MetricCard
				label="worker pool"
				value={$metrics.worker_count}
				title="Active Worker Pool"
				description="the number of dedicated worker goroutines currently spawned to process the transaction queue. this is controlled by the goroutines slider."
			/>
			<MetricCard
				label="failed"
				value={$metrics.failed_tx ?? 0}
				title="transaction failures"
				description="transactions that failed due to logic errors, insufficient funds, or state conflicts."
			/>
			<MetricCard
				label="processed"
				value={$metrics.processed_tx}
				title="Total Processed"
				description="the cumulative count of all transactions successfully processed by the engine since the last reset."
			/>
			<div class="rounded-lg border border-zinc-800 bg-zinc-900/40 p-3">
				<div
					class="mb-1 flex items-center gap-2 text-[10px] font-black tracking-wider text-zinc-500 uppercase"
				>
					<Zap size={10} class="text-amber-500" /> in-flight
				</div>
				<div class="font-mono text-xl text-white">{$metrics.in_flight_tx.toLocaleString()}</div>
			</div>
			<div class="rounded-lg border border-zinc-800 bg-zinc-900/40 p-3">
				<div
					class="mb-1 flex items-center gap-2 text-[10px] font-black tracking-wider text-zinc-500 uppercase"
				>
					<Timer size={10} class="text-emerald-500" /> batch time
				</div>
				<div class="font-mono text-xl text-emerald-400">{$metrics.last_batch_ms}ms</div>
			</div>
		</div>

		<section
			class="flex min-h-0 flex-2 flex-col space-y-3 rounded-xl border border-zinc-900 bg-zinc-950 p-4 shadow-inner"
		>
			<div class="flex shrink-0 items-center justify-between">
				<div class="flex items-center gap-2">
					<Activity size={12} class="text-emerald-500" />
					<h2 class="text-[10px] font-black tracking-widest text-zinc-500 uppercase">
						ledger_flux_heatmap
					</h2>
					<HoverCard.Root openDelay={50} closeDelay={50}>
						<HoverCard.Trigger
							rel="noreferrer noopener"
							class="cursor-pointer transition-colors hover:text-zinc-300"
						>
							<InfoIcon class="size-3" />
						</HoverCard.Trigger>
						<HoverCard.Content
							class="w-80 border-zinc-800 bg-zinc-950 p-3 shadow-xl"
							align="end"
							side="top"
							sideOffset={10}
						>
							<div class="flex items-start space-x-2">
								<InfoIcon class="mt-0.5 size-4 shrink-0 text-emerald-500" />
								<div class="space-y-1">
									<h4 class="text-sm font-bold text-white">ledger flux heatmap</h4>
									<p class="text-xs leading-relaxed text-zinc-400">
										each square represents an account balance (10,000 total). as accounts drop below
										$10k they shift <span class="text-red-500">red</span>, and as they grow they
										shift <span class="text-emerald-500">green</span>.
									</p>
								</div>
							</div>
						</HoverCard.Content>
					</HoverCard.Root>
				</div>

				<div class="flex items-center gap-6">
					<label class="flex items-center gap-3 text-[10px] font-bold text-zinc-600 uppercase">
						worker_delay: <span class="font-mono text-amber-500">{localLatency}ms</span>
						<input
							type="range"
							min="0"
							max="50"
							step="1"
							bind:value={localLatency}
							onpointerdown={() => (isDraggingLatency = true)}
							onpointerup={() => {
								isDraggingLatency = false;
								setLatency(localLatency);
							}}
							onchange={() => setLatency(localLatency)}
							class="h-1 w-24 cursor-pointer appearance-none rounded-lg bg-zinc-800 accent-amber-500"
						/>
					</label>

					<label class="flex items-center gap-3 text-[10px] font-bold text-zinc-600 uppercase">
						goroutines: <span class="font-mono text-emerald-500">{workerCount}</span>
						<input
							type="range"
							min="1"
							max="256"
							bind:value={workerCount}
							onpointerdown={() => (isDraggingWorkers = true)}
							onpointerup={() => {
								isDraggingWorkers = false;
								resizeWorkers(workerCount);
							}}
							onchange={() => resizeWorkers(workerCount)}
							class="h-1 w-24 cursor-pointer appearance-none rounded-lg bg-zinc-800 accent-emerald-500"
						/>
					</label>
				</div>
			</div>
			<div
				class="relative min-h-0 flex-1 overflow-hidden rounded-sm border border-zinc-900 bg-black"
			>
				<canvas bind:this={canvas} class="h-full w-full"></canvas>
			</div>
		</section>

		<div class="grid shrink-0 grid-cols-1 gap-4 lg:grid-cols-2">
			<div class="rounded-xl border border-zinc-900 bg-zinc-950 p-4">
				<div class="mb-3 flex items-center gap-2">
					<Cpu size={12} class="text-blue-500" />
					<h3 class="text-[10px] font-black tracking-widest text-zinc-500 uppercase">
						worker_thread_pool
					</h3>
				</div>
				<div class="flex flex-wrap gap-1">
					{#each $metrics.worker_status as isBusy}
						<div
							class="h-4 w-1 rounded-full transition-all duration-75 {isBusy
								? 'bg-blue-500 shadow-[0_0_8px_#3b82f6]'
								: 'bg-zinc-800'}"
						></div>
					{/each}
				</div>
			</div>

			<div class="rounded-xl border border-zinc-900 bg-zinc-950 p-4">
				<div class="mb-3 flex items-center gap-2">
					<Lock size={12} class="text-amber-500" />
					<h3 class="text-[10px] font-black tracking-widest text-zinc-500 uppercase">
						shard_lock_contention
					</h3>
				</div>
				<div class="grid grid-cols-32 gap-0.5">
					{#each Array(1024) as _, i}
						<div
							class="h-1.5 w-1.5 rounded-sm transition-colors duration-75 {$metrics.active_shards?.includes(
								i
							)
								? 'bg-amber-500 shadow-[0_0_5px_#f59e0b]'
								: 'bg-zinc-900'}"
						></div>
					{/each}
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.grid-cols-32 {
		grid-template-columns: repeat(32, minmax(0, 1fr));
	}
</style>
