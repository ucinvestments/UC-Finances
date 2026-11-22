<script lang="ts">
  import "../app.css";
  import favicon from "$lib/assets/favicon.svg";
  import Navbar from "$lib/components/Navbar.svelte";
  import Footer from "$lib/components/Footer.svelte";

  let { children } = $props();

  const navLinks = [
    { href: "/", label: "Overview", icon: "mdi:home" },
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
    { href: "/about", label: "Sources & Methods", icon: "mdi:information" },
  ];

  const footerLinks = [
    { href: "/revenue-expenses", label: "Revenue & Expenses" },
    { href: "/fund-allocations", label: "Fund Allocations" },
    { href: "/about", label: "Sources & Methods" },
    {
      href: "https://github.com/ucinvestments/UC-Finances",
      label: "Source code",
      external: true,
    },
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

<a class="skip-link" href="#main-content">Skip to main content</a>
<Navbar appName="UC Finances" logoIcon="mdi:briefcase" links={navLinks} />

<main id="main-content" tabindex="-1">
  {@render children?.()}
</main>

<Footer
  appName="UC Finances"
  description="University of California revenues, expenses, and state General Fund allocations."
  links={footerLinks}
  contactEmail="admin@ucinvestments.info"
  altContactEmail="sdokita@berkeley.edu"
  {cryptoAddresses}
  dataSources="UC Office of the President budget and financial reports."
  affiliationDisclaimer="Not affiliated with or authorized by the University of California."
/>

<style>
  .skip-link {
    position: fixed;
    z-index: var(--z-overlay);
    inset-block-start: 0.75rem;
    inset-inline-start: 0.75rem;
    padding: 0.75rem 1rem;
    background: var(--bg);
    color: var(--pri);
    transform: translateY(calc(-100% - 1rem));
  }

  .skip-link:focus {
    transform: none;
  }

  :global(:focus-visible) {
    outline: 2px solid var(--pri);
    outline-offset: 2px;
  }

  main {
    min-height: calc(100vh - 400px);
    scroll-margin-top: 5rem;
  }
</style>
