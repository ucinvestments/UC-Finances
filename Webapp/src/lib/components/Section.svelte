<script lang="ts">
  import Icon from "@iconify/svelte";
  import { fly } from "svelte/transition";

  export let title: string;
  export let icon: string = "";
  export let delay: number = 0;
</script>

<section class="section" in:fly={{ y: 30, duration: 600, delay }}>
  {#if title}
    <div class="section-header">
      {#if icon}
        <Icon {icon} class="section-icon" />
      {/if}
      <h2 class="section-title">{title}</h2>
    </div>
  {/if}
  <slot />
</section>

<style>
  .section {
    margin-bottom: 4rem;
  }

  .section-header {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 2rem;
  }

  :global(.section-icon) {
    font-size: 2rem;
    color: var(--founder);
  }

  .section-title {
    font-family: "Space Grotesk", sans-serif;
    font-size: 2rem;
    font-weight: 700;
    color: var(--pri);
    margin: 0;
    letter-spacing: -0.01em;
  }

  @media (max-width: 768px) {
    .section-title {
      font-size: 1.5rem;
    }

    .section-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.5rem;
    }
  }
</style>