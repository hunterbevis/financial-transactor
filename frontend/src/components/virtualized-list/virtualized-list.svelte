<script lang="ts">
    import { type Snippet } from 'svelte';

    let { itemCount, itemSize, height, item }: {
        itemCount: number;
        itemSize: number;
        height: number;
        item: Snippet<[{ index: number, style: string }]>;
    } = $props();

    let container: HTMLDivElement;
    let scrollTop = $state(0);
    let isAtBottom = $state(true);

    function handleScroll(e: Event) {
        const target = e.currentTarget as HTMLDivElement;
        scrollTop = target.scrollTop;
        const threshold = 50; 
        isAtBottom = target.scrollHeight - target.scrollTop - target.clientHeight < threshold;
    }

    $effect(() => {
        if (itemCount && container && isAtBottom) {
            container.scrollTo({
                top: container.scrollHeight,
                behavior: 'instant'
            });
        }
    });

    const start = $derived(Math.max(0, Math.floor(scrollTop / itemSize)));
    const end = $derived(Math.min(itemCount - 1, start + Math.ceil(height / itemSize) + 1));

    const visibleIndices = $derived.by(() => {
        const indices = [];
        for (let i = start; i <= end; i++) indices.push(i);
        return indices;
    });
</script>

<div 
    bind:this={container}
    class="scrollbar-hide relative overflow-y-auto [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden" 
    style:height="{height}px"
    onscroll={handleScroll}
>
    <div class="pointer-events-none absolute inset-0 z-0 opacity-5" 
         style:background-image="linear-gradient(to bottom, #fff 1px, transparent 1px)"
         style:background-size="100% {itemSize}px">
    </div>

    <div style:height="{itemCount * itemSize}px" class="relative w-full z-10">
        {#each visibleIndices as index (index)}
            <div 
                class="absolute left-0 w-full transition-opacity duration-300"
                style:height="{itemSize}px"
                style:transform="translateY({index * itemSize}px)"
            >
                {@render item({ index, style: "" })}
            </div>
        {/each}
    </div>
</div>