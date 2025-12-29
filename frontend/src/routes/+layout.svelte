<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import animaginary from '$lib/assets/animaginary.svg'

	let { children } = $props();
</script>

<svelte:head><link rel="icon" href={animaginary} /></svelte:head>
{@render children()}
