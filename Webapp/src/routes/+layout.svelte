<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import { fade } from 'svelte/transition';
	import Navbar from '$lib/Navbar.svelte';
	import Footer from '$lib/Footer.svelte';

	let { children } = $props();

	const navLinks = [
		{ href: '/', label: 'Explorer', icon: 'mdi:home' },
		{ href: '/about', label: 'About', icon: 'mdi:information' },
		{ href: '/resources', label: 'Resources', icon: 'mdi:book-open-page-variant' },
		{ href: 'https://github.com/TheArctesian/holdings', label: 'GitHub', icon: 'mdi:github', external: true }
	];

	const footerLinks = [
		{ href: '/', label: 'Explorer' },
		{ href: '/about', label: 'About' },
		{ href: '/resources', label: 'Resources' },
		{ href: 'https://sec.gov/', label: 'SEC Website', external: true }
	];
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<Navbar
	appName="Holdings Explorer"
	logoIcon="mdi:briefcase"
	links={navLinks}
/>

<main in:fade={{ duration: 300 }}>
	{@render children?.()}
</main>

<Footer
	appName="Holdings Data Explorer"
	description="Transparency in institutional holdings and financial data."
	links={footerLinks}
	contactEmail="admin@ucinvestments.info"
	lastUpdated="2024"
	dataAccuracy="SEC Filing Based"
/>

<style>
	:global(html) {
		scroll-behavior: smooth;
	}

	main {
		min-height: calc(100vh - 400px);
	}
</style>