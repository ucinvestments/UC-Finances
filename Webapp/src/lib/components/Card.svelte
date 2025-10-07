<script lang="ts">
  import { fly } from "svelte/transition";

  export let hover: boolean = true;
  export let padding: string = "2.5rem";
  export let delay: number = 0;
</script>

<div
  class="card"
  class:hoverable={hover}
  style="padding: {padding}"
  in:fly={{ y: 30, duration: 600, delay }}
>
  <slot />
</div>

<style>
  .card {
    background: rgba(255, 255, 255, 0.9);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border-radius: 1.5rem;
    border: 1px solid var(--border);
    box-shadow: var(--shadow-md);
    transition: all 0.3s ease;
  }

  .card.hoverable:hover {
    transform: translateY(-4px);
    box-shadow: var(--shadow-lg);
  }

  @media (max-width: 768px) {
    .card {
      padding: 1.5rem !important;
    }
  }
</style>