<script lang="ts">
  import Icon from "@iconify/svelte";
  import { page } from "$app/stores";

  export let appName: string = "UC Data";
  export let logoIcon: string = "mdi:chart-donut";
  export let links: Array<{href: string, label: string, icon: string, external?: boolean}> = [];
</script>

<nav class="navbar">
  <div class="nav-container">
    <a href="/" class="logo-link">
      <div class="logo">
        <Icon icon={logoIcon} class="logo-icon" />
        <span class="logo-text">{appName}</span>
      </div>
    </a>

    <div class="nav-links">
      {#each links as link}
        <a
          href={link.href}
          class="nav-link"
          class:active={$page.url.pathname === link.href}
          class:external={link.external}
          target={link.external ? "_blank" : undefined}
          rel={link.external ? "noopener noreferrer" : undefined}
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
    border-bottom: 1px solid var(--border);
    box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1);
  }

  .nav-container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 1rem 2rem;
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
    font-weight: 700;
    color: var(--pri);
    letter-spacing: -0.01em;
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
    padding: 0.625rem 1.25rem;
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

  @media (max-width: 768px) {
    .nav-container {
      flex-direction: column;
      gap: 1rem;
      padding: 1rem;
    }

    .nav-links {
      width: 100%;
      justify-content: center;
    }

    .nav-link {
      padding: 0.5rem 1rem;
      font-size: 0.875rem;
    }

    :global(.nav-icon) {
      font-size: 1rem;
    }

    .logo-text {
      font-size: 1.125rem;
    }
  }

  @media (max-width: 480px) {
    .nav-links {
      flex-direction: column;
      width: 100%;
    }

    .nav-link {
      width: 100%;
      justify-content: center;
    }
  }
</style>