<script lang="ts">
  import Icon from "@iconify/svelte";
  import { page } from "$app/stores";

  interface NavLink {
    href: string;
    label: string;
    icon: string;
    external?: boolean;
  }

  let {
    appName = "UC Data",
    logoIcon = "mdi:chart-donut",
    subtitle = "",
    links = [] as NavLink[],
  }: {
    appName?: string;
    logoIcon?: string;
    subtitle?: string;
    links?: NavLink[];
  } = $props();

  let mobileOpen = $state(false);

  $effect(() => {
    if ($page.url) mobileOpen = false;
  });

  function toggle() {
    mobileOpen = !mobileOpen;
  }

  function close() {
    mobileOpen = false;
  }
</script>

<nav class="navbar">
  <div class="nav-container">
    <a href="/" class="logo-link" onclick={close}>
      <div class="logo">
        <Icon icon={logoIcon} class="logo-icon" />
        <span class="logo-text">
          <strong>{appName}</strong>
          {#if subtitle}<small class="small-subtitle">{subtitle}</small>{/if}
        </span>
      </div>
    </a>

    <button class="burger-btn" onclick={toggle} aria-label="Toggle menu">
      <Icon icon={mobileOpen ? "mdi:close" : "mdi:menu"} class="burger-icon" />
    </button>

    <div class="nav-links" class:mobile-open={mobileOpen}>
      {#each links as link}
        <a
          href={link.href}
          class="nav-link"
          class:active={$page.url.pathname === link.href}
          class:external={link.external}
          target={link.external ? "_blank" : undefined}
          rel={link.external ? "noopener noreferrer" : undefined}
          onclick={close}
        >
          <Icon icon={link.icon} class="nav-icon" />
          {link.label}
          {#if link.external}
            <Icon icon="mdi:open-in-new" class="external-icon" />
          {/if}
        </a>
      {/each}
    </div>
  </div>
</nav>

<style>
  .navbar {
    position: sticky;
    top: 0;
    z-index: 100;
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border-bottom: 1px solid var(--border);
    box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1);
  }

  .nav-container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 0.5rem 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .logo-link {
    text-decoration: none;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    transition: transform 0.3s ease;
  }

  .logo:hover {
    transform: scale(1.05);
  }

  :global(.logo-icon) {
    font-size: 2rem;
    color: var(--founder);
  }

  .logo-text {
    font-family: "Space Grotesk", sans-serif;
    font-size: 1.25rem;
    font-weight: 900;
    color: var(--pri);
    letter-spacing: -0.01em;
  }

  .small-subtitle {
    font-weight: 400;
    font-size: 0.9rem;
  }

  .nav-links {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .nav-link {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.5rem 1rem;
    color: var(--text-secondary);
    text-decoration: none;
    font-weight: 500;
    border-radius: 0.75rem;
    transition: all 0.2s ease;
    position: relative;
  }

  :global(.nav-icon) {
    font-size: 1.125rem;
  }

  .nav-link:hover {
    background: var(--bg-secondary);
    color: var(--pri);
    transform: translateY(-1px);
  }

  .nav-link.active {
    background: linear-gradient(135deg, var(--founder), var(--pri));
    color: white;
  }

  .nav-link.external {
    border: 2px solid var(--border);
  }

  :global(.external-icon) {
    font-size: 0.875rem;
    margin-left: -0.125rem;
  }

  .burger-btn {
    display: none;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.5rem;
    border-radius: 0.5rem;
    transition: background 0.2s ease;
  }

  .burger-btn:hover {
    background: var(--bg-secondary);
  }

  :global(.burger-icon) {
    font-size: 1.5rem;
    color: var(--pri);
  }

  @media (max-width: 768px) {
    .burger-btn {
      display: block;
    }

    .nav-links {
      display: none;
      position: absolute;
      top: 100%;
      left: 0;
      right: 0;
      background: rgba(255, 255, 255, 0.98);
      backdrop-filter: blur(20px);
      flex-direction: column;
      padding: 0.75rem 1rem;
      border-bottom: 1px solid var(--border);
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
      z-index: 99;
    }

    .nav-links.mobile-open {
      display: flex;
    }

    .nav-container {
      position: relative;
      flex-direction: row;
      padding: 0.75rem 1rem;
    }

    .nav-link {
      padding: 0.75rem 1rem;
      font-size: 0.925rem;
      width: 100%;
      justify-content: flex-start;
      border-radius: 0.5rem;
    }

    :global(.nav-icon) {
      font-size: 1rem;
    }

    .logo-text {
      font-size: 1.125rem;
    }
  }
</style>
