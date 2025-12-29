<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Activity, Cpu, AlertCircle, Info as InfoIcon, User } from '@lucide/svelte';

	import { metrics, session_processed, txlog, is_congested } from '$lib/stores';
	import { connectmetrics, connecttxstream } from '$lib/ws';
	import { submittransactions } from '$lib/api';

	import * as HoverCard from '../components/hover-card/index.ts';
	import MetricCard from '../components/metric-card.svelte';
	import VirtualizedList from '../components/virtualized-list/virtualized-list.svelte';

	let cleanup: (() => void)[] = [];

	let lastProcessed = $state(0);
	let lastTs = performance.now();
	let tps = $state(0);
	let initialized = false;

	function getflag(submitted_by: string) {
		const parts = submitted_by.split('-');
		const code = parts[0]?.toUpperCase();
		const suffix = parts[1]?.toUpperCase();

		if (suffix === 'LOCAL' || !code || code === 'XX') {
			return 'LOCAL';
		}

		if (code.length === 2) {
			return code;
		}

		return 'GL';
	}

	function getdisplayname(submitted_by: string) {
		const parts = submitted_by.split('-');
		return parts.length > 1 ? parts[1] : submitted_by;
	}

	$effect(() => {
		const current = $metrics.processed;
		const now = performance.now();
		const timediff = (now - lastTs) / 1000;

		if (!initialized || current < lastProcessed) {
			lastProcessed = current;
			lastTs = now;
			initialized = true;
			return;
		}

		if (timediff > 0) {
			const instanttps = (current - lastProcessed) / timediff;
			tps = instanttps * 0.15 + tps * 0.85;
		}

		lastProcessed = current;
		lastTs = now;
	});

	let userName = $state('');

	onMount(() => {
		userName = localStorage.getItem('tx_alias') || '';
		cleanup.push(connectmetrics());
		cleanup.push(connecttxstream());
	});

	onDestroy(() => {
		cleanup.forEach((fn) => fn());
	});

	// watch for changes and save to localStorage
	$effect(() => {
		localStorage.setItem('tx_alias', userName);
	});
</script>

<div
	class="h-svh w-full overflow-hidden bg-black p-4 font-sans text-zinc-100 selection:bg-emerald-500/30 md:p-6"
>
	<div class="mx-auto flex h-full max-w-7xl flex-col space-y-4">
		<header
			class="flex shrink-0 flex-col gap-4 border-b border-zinc-900 pb-4 sm:flex-row sm:items-center sm:justify-between"
		>
			<div class="flex items-center gap-4">
				<div class="space-y-0.5">
					<h1 class="text-lg font-black tracking-tighter text-white uppercase italic md:text-xl">
						financial transactor engine
					</h1>
					<p
						class="font-mono text-[9px] leading-none tracking-tight text-white uppercase md:text-[11px]"
					>
						high-concurrency atomic ledger
					</p>
				</div>

				<div
					class="flex items-center gap-2 rounded-full border border-zinc-800 bg-zinc-900/50 px-2 py-0.5 md:px-2.5 md:py-1"
				>
					<div
						class="h-2 w-2 rounded-full {$metrics.isconnected
							? 'animate-pulse bg-emerald-500 shadow-[0_0_8px_#10b981]'
							: 'bg-red-500'}"
					></div>
					<span class="text-[8px] font-black tracking-widest text-zinc-400 uppercase md:text-[9px]">
						{$metrics.isconnected ? 'online' : 'offline'}
					</span>
				</div>

				<div class="relative hidden sm:block">
					<div class="pointer-events-none absolute inset-y-0 left-2 flex items-center">
						<User size={10} class={userName ? 'text-emerald-500' : 'text-zinc-600'} />
					</div>
					<input
						type="text"
						bind:value={userName}
						onblur={() => (userName = userName.trim())}
						placeholder="SET ALIAS.."
						maxlength="12"
						class="w-28 rounded border border-zinc-800 bg-zinc-950 py-1 pr-2 pl-6 font-normal text-[11px] text-emerald-500 uppercase transition-all placeholder:text-zinc-300 focus:border-emerald-500/50 focus:outline-none"
					/>
					{#if userName.length > 0}
						<span class="absolute right-0 -bottom-4 font-mono text-[11px] text-white">
							{userName.length}/12
						</span>
					{/if}
				</div>

				<HoverCard.Root openDelay={50} closeDelay={50}>
					<HoverCard.Trigger class="cursor-pointer transition-colors hover:text-zinc-300">
						<InfoIcon class="size-3" />
					</HoverCard.Trigger>
					<HoverCard.Content class="w-72 border-zinc-800 bg-zinc-950 p-3 shadow-xl md:w-80">
						<h4 class="text-xs font-bold text-white uppercase">Engine Manual</h4>
						<p class="mt-2 text-[11px] leading-relaxed text-zinc-400">
							Enter a name to submit randomized transactions to the server.
						</p>
					</HoverCard.Content>
				</HoverCard.Root>
			</div>

			<div class="flex gap-2">
				<HoverCard.Root openDelay={50} closeDelay={50}>
					<HoverCard.Trigger
						class="flex items-center text-white justify-center rounded border border-zinc-800 bg-zinc-900 px-4 py-2 text-[9px] font-bold uppercase transition-colors hover:bg-zinc-800 md:text-[11px]"
					>
						Info
					</HoverCard.Trigger>
					<HoverCard.Content
						class="w-80 border-zinc-800 bg-zinc-950 p-4 shadow-2xl md:w-96"
						align="end"
						side="bottom"
						sideOffset={10}
					>
						<div class="space-y-3">
							<h4 class="font-mono text-xs font-black tracking-tighter text-emerald-500 uppercase">
								System Documentation - v1.0.0
							</h4>
							
							<section>
								<h5 class="text-[13px] font-bold text-white uppercase tracking-widest">Usage</h5>
								<p class="mt-1 text-[12px] leading-relaxed text-zinc-400">
									Set a custom alias in the input field and select a batch size (+10k or +10m) to broadcast transactions. The system enforces a strict 10M global queue limit to maintain stability.
								</p>
							</section>

							<section>
								<h5 class="text-[12px] font-bold text-white uppercase tracking-widest">Backend Architecture</h5>
								<p class="mt-1 text-[12px] leading-relaxed text-zinc-400">
									Built in <span class="text-zinc-200">Go</span>, the engine utilizes a <span class="text-zinc-200">1,024-shard ledger</span> with deterministic lock-ordering to process transfers at massive scale without data corruption or deadlocks.
								</p>
							</section>

							<section class="rounded border border-emerald-500/20 bg-emerald-500/5 p-2">
								<h5 class="text-[12px] font-bold text-emerald-400 uppercase tracking-widest">Global Network</h5>
								<p class="mt-1 text-[12px] leading-relaxed text-emerald-500/80 italic">
									All activity is public. You are viewing a live, global stream of transactions submitted by users worldwide in real-time via WebSockets.
								</p>
							</section>
						</div>
					</HoverCard.Content>
				</HoverCard.Root>
				<button
					class="flex-1 rounded border border-zinc-800 bg-zinc-900 px-3 py-2 text-[9px] font-bold uppercase transition-colors hover:bg-zinc-800 sm:flex-none md:text-[11px]"
					onclick={() => submittransactions(10000, userName)}
				>
					+10k tx
				</button>
				<button
					class="group relative flex-1 rounded bg-emerald-600 px-4 py-2 text-[9px] font-black text-white uppercase shadow-lg shadow-emerald-900/20 transition-all hover:bg-emerald-500 active:scale-95 sm:flex-none md:text-[11px]"
					onclick={() => submittransactions(1000000, userName)}
				>
					flood +1m tx
				</button>
			</div>
		</header>

		<div class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-6">
			<MetricCard 
				label="cpu threads" 
				value={$metrics.cpu_threads} 
				title="Runtime Parallelism"
				description="Total logical cores available to the Go runtime (GOMAXPROCS). This determines the maximum hardware concurrency for parallel sharded processing."
			/>
			
			<MetricCard 
				label="goroutines" 
				value={$metrics.goroutines} 
				title="Active Goroutines"
				description="Lightweight execution units managed by the Go scheduler. This includes workers, and system processes goroutines."
			/>
			
			<MetricCard 
				label="workers" 
				value={$metrics.worker_pool} 
				title="Transaction Processors"
				description="The dedicated pool of concurrent workers pulling from the job channel. It is optimized at 4x the physical core count to balance throughput and context switching."
			/>
			
			<MetricCard 
				label="velocity" 
				value="{Math.floor(tps).toLocaleString()} TPS" 
				title="Engine Throughput"
				description="Real-time transactions per second. This represents the current speed of the engine's distributed ledger updates and lock acquisition cycles."
			/>
			
			<MetricCard 
				label="processed" 
				value={$session_processed.toLocaleString()} 
				title="Session Total"
				description="The accumulated count of successfully committed transactions since you have been connected to the engine."
			/>
			
			<MetricCard 
				label="queue" 
				value={$metrics.queue_len.toLocaleString()} 
				title="Buffer Depth"
				description="The number of transactions currently residing in the buffered channel (10M capacity). High values indicate the engine is under peak pressure."
			/>
		</div>

		<section
			class="flex min-h-0 flex-1 flex-col space-y-3 rounded-xl border border-zinc-900 bg-zinc-950 p-3 shadow-inner md:p-4"
		>
			<div class="flex shrink-0 items-center justify-between">
				<div class="flex items-center gap-2">
					<Activity size={12} class="text-emerald-500" />
					<h2 class="text-[9px] font-black tracking-widest text-zinc-500 uppercase md:text-[11px]">
						live_transaction_log
					</h2>
				</div>

				{#if $is_congested}
					<div
						class="flex animate-pulse items-center gap-1.5 text-[8px] font-black text-red-500 md:text-[9px]"
					>
						<AlertCircle size={10} /> BACKPRESSURE
					</div>
				{/if}
			</div>

			<div class="flex-1 overflow-hidden border-t border-zinc-900 pt-2">
				<VirtualizedList itemCount={$txlog.length} itemSize={28} height={500}>
					{#snippet item({ index })}
						{@const tx = $txlog[index]}
						{#if tx}
							<div
								class="flex items-center gap-2 px-1 font-mono text-[11px] leading-7 tracking-tighter text-zinc-400 transition-colors hover:bg-zinc-900/50 md:gap-3 md:px-2 md:text-[11px] md:tracking-tight"
							>
								<span class="hidden w-16 text-zinc-600 sm:inline"
									>[{new Date(tx.ts).toLocaleTimeString([], { hour12: false })}]</span
								>

								<span
									class="flex items-center gap-1 rounded border border-zinc-800 bg-zinc-900 px-1.5 text-emerald-500 md:min-w-25"
								>
									<span class="text-[12px] md:text-[14px]">{getflag(tx.submitted_by)}</span>
									<span class="max-w-20 truncate font-bold uppercase md:max-w-none">
										{getdisplayname(tx.submitted_by)}
									</span>
								</span>

								<div class="flex items-center gap-1">
									<span class="text-zinc-500">_{tx.from}</span>
									<span class="text-zinc-700">→</span>
									<span class="text-zinc-500">_{tx.to}</span>
								</div>

								<span class="ml-auto font-bold text-amber-500 tabular-nums">
									${tx.amount}
								</span>
							</div>
						{/if}
					{/snippet}
				</VirtualizedList>
			</div>
		</section>

		<div class="grid shrink-0 grid-cols-1 gap-3 sm:grid-cols-2">
			<div class="hidden rounded-xl border border-zinc-900 bg-zinc-950 p-3 sm:block md:p-4">
				<div class="mb-2 flex items-center gap-2">
					<Cpu size={12} class="text-blue-500" />
					<h3 class="text-[9px] font-black tracking-widest text-zinc-500 uppercase">
						worker_distribution
					</h3>
				</div>
				<div class="flex flex-wrap gap-1">
					{#each Array.from({ length: Math.min($metrics.worker_pool, 120) }) as _}
						<div
							class="h-3 w-0.5 rounded-full bg-blue-500 opacity-60 shadow-[0_0_4px_#3b82f6]"
						></div>
					{/each}
				</div>
			</div>

			<div class="rounded-xl border border-zinc-900 bg-zinc-950 p-3 md:p-4">
				<div class="mb-2 flex items-center gap-2">
					<Activity size={12} class="text-zinc-500" />
					<h3 class="text-[9px] font-black tracking-widest text-zinc-500 uppercase">
						system_health
					</h3>
				</div>
				<div class="flex items-center justify-between gap-2">
					<div class="flex gap-3">
						<div class="text-[9px] font-bold text-zinc-500 uppercase">
							shards: <span class="text-white">1024</span>
						</div>
						<div class="text-[9px] font-bold text-zinc-500 uppercase">
							accts: <span class="text-white">10k</span>
						</div>
					</div>
					<div class="text-[8px] font-black text-zinc-100 uppercase italic">v1</div>
				</div>
			</div>
		</div>
	</div>
</div>
