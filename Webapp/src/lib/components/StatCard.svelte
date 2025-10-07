<script lang="ts">
  import Icon from "@iconify/svelte";
  import { fade } from "svelte/transition";

  export let icon: string;
  export let value: string;
  export let label: string;
  export let delay: number = 0;
</script>

<div class="stat-card" in:fade={{ duration: 600, delay }}>
  <Icon {icon} class="stat-icon" />
  <div class="stat-content">
    <div class="stat-value">{value}</div>
    <div class="stat-label">{label}</div>
  </div>
</div>

<style>
  .stat-card {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    color: white;
  }

  :global(.stat-icon) {
    font-size: 1.125rem;
  }

  .stat-value {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--sec);
    font-family: "Space Grotesk", sans-serif;
  }

  .stat-divider {
    color: rgba(255, 255, 255, 0.4);
    font-size: 1.25rem;
  }

  .stat-label {
    font-size: 0.925rem;
    color: rgba(255, 255, 255, 0.95);
    font-weight: 500;
  }

  @media (max-width: 640px) {
    .stat-card {
      justify-content: center;
    }

    .stat-value {
      font-size: 1.125rem;
    }

    .stat-label {
      font-size: 0.875rem;
    }
  }
</style>