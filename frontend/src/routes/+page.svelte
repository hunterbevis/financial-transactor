<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { 
        Activity, 
        Cpu, 
        AlertCircle,
        Info as InfoIcon 
    } from '@lucide/svelte';

    import { metrics, session_processed, txlog, queue_pressure, is_congested } from '$lib/stores';
    import { connectmetrics, connecttxstream } from '$lib/ws';
    import { submittransactions } from '$lib/api';

    import * as HoverCard from '../components/hover-card/index.ts';
    import MetricCard from '../components/metric-card.svelte';
    import VirtualizedList from '../components/virtualized-list/virtualized-list.svelte';

    let disconnectmetrics: (() => void) | undefined;
    let disconnectTx: (() => void) | undefined;

	let lastProcessed = $state(0); // make this state or just a let
    let lastTs = performance.now();
    let tps = $state(0);
    let initialized = false; // track the first data burst

    $effect(() => {
        const current = $metrics.processed;
        const now = performance.now();
        const timediff = (now - lastTs) / 1000;

        // if this is the first time we get data, or the engine was reset
        if (!initialized || current < lastProcessed) {
            lastProcessed = current;
            lastTs = now;
            initialized = true;
            return; // skip calculation for the first frame
        }

        if (timediff > 0) {
            const instanttps = (current - lastProcessed) / timediff;
            // smoothing
            tps = (instanttps * 0.15) + (tps * 0.85);
        }

        lastProcessed = current;
        lastTs = now;
    });

	onMount(() => {
        const disconnectMetrics = connectmetrics();
        const disconnectTx = connecttxstream();

        return () => {
            disconnectMetrics();
            disconnectTx();
        };
    });

    onDestroy(() => {
        disconnectmetrics?.();
        disconnectTx?.();
    });
</script>

<div class="h-screen w-full overflow-hidden bg-black p-6 font-sans text-zinc-100 selection:bg-emerald-500/30">
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
                
                <div class="flex items-center gap-2 rounded-full border border-zinc-800 bg-zinc-900/50 px-2.5 py-1">
                    <div class="h-2 w-2 rounded-full {$metrics.isconnected ? 'animate-pulse bg-emerald-500 shadow-[0_0_8px_#10b981]' : 'bg-red-500'}"></div>
                    <span class="text-[9px] font-black tracking-widest text-zinc-400 uppercase">
                        {$metrics.isconnected ? 'system_online' : 'engine_offline'}
                    </span>
                </div>

                <HoverCard.Root openDelay={50} closeDelay={50}>
                    <HoverCard.Trigger class="cursor-pointer transition-colors hover:text-zinc-300">
                        <InfoIcon class="size-3" />
                    </HoverCard.Trigger>
                    <HoverCard.Content class="w-80 border-zinc-800 bg-zinc-950 p-3 shadow-xl">
                        <h4 class="text-xs font-bold uppercase text-white">engine manual</h4>
                        <p class="mt-2 text-[11px] leading-relaxed text-zinc-400">
                            this dashboard monitors a high-performance golang transaction engine. updates occur every 16ms via websockets. ingress buffer tracks the job queue depth.
                        </p>
                    </HoverCard.Content>
                </HoverCard.Root>
            </div>

            <div class="flex gap-2">
                <button 
                    class="rounded border border-zinc-800 bg-zinc-900 px-3 py-1.5 text-[10px] font-bold uppercase hover:bg-zinc-800 transition-colors"
                    onclick={() => submittransactions(10000)}>
                    +10k tx
                </button>
                <button 
                    class="group relative rounded bg-emerald-600 px-4 py-1.5 text-[10px] font-black text-white uppercase shadow-lg shadow-emerald-900/20 hover:bg-emerald-500 active:scale-95 transition-all"
                    onclick={() => submittransactions(5000000)}>
                    flood 5m tx
                </button>
            </div>
        </header>

		<div class="grid shrink-0 grid-cols-2 gap-3 lg:grid-cols-6">
			<MetricCard 
				label="cpu threads" 
				value={$metrics.cpu_threads} 
				title="os threads" 
				description="total logical cores available to the go runtime." 
			/>
			
			<MetricCard 
				label="goroutines" 
				value={$metrics.goroutines} 
				title="active routines" 
				description="total number of active go routines (workers + net/http)." 
			/>
			
			<MetricCard 
				label="worker pool" 
				value={$metrics.worker_pool} 
				title="engine workers" 
				description="dedicated worker goroutines processing the transaction queue." 
			/>
			
			<MetricCard 
				label="velocity" 
				value="{Math.floor(tps).toLocaleString()} tps" 
				title="throughput" 
				description="real-time transactions processed per second."
			/>

			<MetricCard 
				label="processed" 
				value={$session_processed} 
				title="session total" 
				description="cumulative successful transactions since last reset."
			/>

			<MetricCard 
					label="queue" 
					value={$metrics.queue_len} 
					title="backlog" 
					description="transactions currently waiting in the 10m buffer."
			/>

		</div>

        <section class="flex min-h-0 flex-1 flex-col space-y-3 rounded-xl border border-zinc-900 bg-zinc-950 p-4 shadow-inner">
            <div class="flex shrink-0 items-center justify-between">
                <div class="flex items-center gap-2">
                    <Activity size={12} class="text-emerald-500" />
                    <h2 class="text-[10px] font-black tracking-widest text-zinc-500 uppercase">live_transaction_log</h2>
                </div>

                <div class="flex items-center gap-6">
                    {#if $is_congested}
                        <div class="flex items-center gap-1.5 text-[9px] font-black text-red-500 animate-pulse">
                            <AlertCircle size={10} /> BACKPRESSURE_ACTIVE
                        </div>
                    {/if}
                    <div class="flex items-center gap-2 text-[10px] font-bold text-zinc-600 uppercase">
                        active_workers: <span class="font-mono text-emerald-500">{$metrics.worker_pool}</span>
                    </div>
                </div>
            </div>

			<div class="flex-1 overflow-hidden border-t border-zinc-900 pt-2">
				<VirtualizedList itemCount={$txlog.length} itemSize={24} height={400}>
					{#snippet item({ index })}
						{@const tx = $txlog[index]}
						{#if tx}
							<div class="flex items-center gap-3 px-2 font-mono text-[11px] leading-6 tracking-tight text-zinc-400 hover:bg-zinc-900/50 transition-colors">
								<span class="text-zinc-600 w-20">[{new Date(tx.ts).toLocaleTimeString()}]</span>
								
								<span class="rounded bg-zinc-900 px-1 text-emerald-500 min-w-20 text-center border border-zinc-800">
									{tx.submitted_by}
								</span>

								<span class="text-zinc-500">acct_{tx.from}</span>
								<span class="text-zinc-700">→</span>
								<span class="text-zinc-500">acct_{tx.to}</span>
								
								<span class="ml-auto font-bold text-amber-500 tabular-nums">
									${tx.amount.toLocaleString()}
								</span>
							</div>
						{/if}
					{/snippet}
				</VirtualizedList>
			</div>

        </section>

        <div class="grid shrink-0 grid-cols-1 gap-4 lg:grid-cols-2">
            <div class="rounded-xl border border-zinc-900 bg-zinc-950 p-4">
                <div class="mb-3 flex items-center gap-2">
                    <Cpu size={12} class="text-blue-500" />
                    <h3 class="text-[10px] font-black tracking-widest text-zinc-500 uppercase">worker_distribution</h3>
                </div>
                <div class="flex flex-wrap gap-1">
                    {#each Array.from({ length: Math.min($metrics.worker_pool, 300) }) as _}
                        <div class="h-4 w-1 rounded-full bg-blue-500 shadow-[0_0_8px_#3b82f6] opacity-80"></div>
                    {/each}
                </div>
            </div>

            <div class="rounded-xl border border-zinc-900 bg-zinc-950 p-4">
                <div class="mb-3 flex items-center gap-2">
                    <Activity size={12} class="text-zinc-500" />
                    <h3 class="text-[10px] font-black tracking-widest text-zinc-500 uppercase">system_health</h3>
                </div>
                <div class="flex items-center gap-4">
                    <div class="text-[10px] text-zinc-500 uppercase font-bold">
                        shard_count: <span class="text-white font-mono">1024</span>
                    </div>
                    <div class="text-[10px] text-zinc-500 uppercase font-bold">
                        accounts: <span class="text-white font-mono">10,000</span>
                    </div>
                    <div class="ml-auto text-[9px] text-zinc-700 uppercase font-black italic">
                        v2.0_stable
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>