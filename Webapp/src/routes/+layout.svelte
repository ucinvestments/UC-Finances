<script lang="ts">
  import "../app.css";
  import favicon from "$lib/assets/favicon.svg";
  import { fade } from "svelte/transition";
  import Navbar from "$lib/components/Navbar.svelte";
  import Footer from "$lib/components/Footer.svelte";

  let { children } = $props();

  const navLinks = [
    { href: "/", label: "Explorer", icon: "mdi:home" },
    {
      href: "/revenue-expenses",
      label: "Revenue & Expenses",
      icon: "mdi:chart-areaspline",
    },
    {
      href: "/fund-allocations",
      label: "Fund Allocations",
      icon: "mdi:file-tree",
    },
    { href: "/about", label: "About", icon: "mdi:information" },
    {
      href: "https://github.com/ucinvestments/UC-Finances",
      label: "GitHub",
      icon: "mdi:github",
      external: true,
    },
  ];

  const footerLinks = [
    { href: "/", label: "Explorer" },
    { href: "/about", label: "About" },
    { href: "https://sec.gov/", label: "SEC Website", external: true },
  ];

  const cryptoAddresses = [
    { label: "ETH", address: "0x623c7559ddC51BAf15Cc81bf5bc13c0B0EA14c01" },
    {
      label: "XMR",
      address:
        "44bvXALNkxUgSkGChKQPnj79v6JwkeYEkGijgKyp2zRq3EiuL6oewAv5u2c7FN7jbN1z7uj1rrPfL77bbsJ3cC8U2ADFoTj",
    },
  ];
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
</svelte:head>

<Navbar appName="UC Finances" logoIcon="mdi:briefcase" links={navLinks} />

<main in:fade={{ duration: 300 }}>
  {@render children?.()}
</main>

<Footer
  appName="UC Finances"
  description="Transparency in institutional holdings and financial data."
  links={footerLinks}
  contactEmail="admin@ucinvestments.info"
  altContactEmail="sdokita@berkeley.edu"
  {cryptoAddresses}
  dataSources={"SEC 13F-HR filings<br />Source: data.sec.gov"}
  affiliationDisclaimer="Not affiliated with or authorized by the University of California."
/>

<style>
  :global(html) {
    scroll-behavior: smooth;
  }

  main {
    min-height: calc(100vh - 400px);
  }
</style>
