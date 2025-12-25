<script lang="ts">
	import { InfoIcon } from '@lucide/svelte';
	import * as HoverCard from '../components/hover-card/index.ts';

	export let label: string;
	export let value: number | string;
	export let title: string = ''; // optional title for the tooltip
	export let description: string = ''; // providing this makes the hovercard appear
</script>

<div class="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4 shadow-sm">
	<div
		class="flex items-center justify-between text-[10px] font-black tracking-widest text-zinc-500 uppercase"
	>
		{label}

		{#if description}
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
							{#if title}
								<h4 class="text-sm font-bold text-white">{title}</h4>
							{/if}
							<p class="text-xs leading-relaxed text-zinc-400">
								{description}
							</p>
						</div>
					</div>
				</HoverCard.Content>
			</HoverCard.Root>
		{/if}
	</div>
	<div class="mt-2 font-mono text-2xl font-semibold text-white">
		{typeof value === 'number' ? value.toLocaleString() : value}
	</div>
</div>
