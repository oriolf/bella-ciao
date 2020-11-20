<script lang="ts">
	import Nav from '../components/Nav.svelte';
	import { setupI18n } from "../i18n.js";

	if (typeof window !== 'undefined') {
		setupI18n(localStorage);
	} else {
		setupI18n(null);
	}

	export let segment: string;
</script>

<style>
	main {
		position: relative;
		max-width: 1024px;
		padding: 5px;
		margin: 0 auto;
		box-sizing: border-box;
	}
</style>

<Nav {segment}/>

<main>
	<slot></slot>
</main>