<script lang="ts">
  interface Props {
    title?: string;
    subtitle?: string;
    minHeight?: number;
    legend?: import("svelte").Snippet;
    controls?: import("svelte").Snippet;
    children?: import("svelte").Snippet;
  }
  let { title, subtitle, minHeight = 420, legend, controls, children }: Props = $props();
</script>

<div class="chart-frame">
  {#if title || subtitle || controls}
    <header>
      <div class="titles">
        {#if title}<h3>{title}</h3>{/if}
        {#if subtitle}<p>{subtitle}</p>{/if}
      </div>
      {#if controls}<div class="controls">{@render controls()}</div>{/if}
    </header>
  {/if}
  {#if legend}<div class="legend">{@render legend()}</div>{/if}
  <div class="body" style="min-height: {minHeight}px;">
    {@render children?.()}
  </div>
</div>

<style>
  .chart-frame {
    background: #ffffff;
    border-radius: 16px;
    box-shadow: 0 10px 25px -15px rgba(0, 50, 98, 0.25);
    padding: 1.75rem;
    margin: 0 0 2rem;
    border: 1px solid rgba(0, 50, 98, 0.06);
  }
  header {
    display: flex;
    flex-wrap: wrap;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
    margin-bottom: 1rem;
  }
  .titles h3 {
    margin: 0;
    font-size: 1.125rem;
    font-weight: 600;
    color: #0f172a;
    letter-spacing: -0.01em;
  }
  .titles p {
    margin: 0.25rem 0 0;
    color: #64748b;
    font-size: 0.875rem;
  }
  .controls {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }
  .legend {
    margin-bottom: 0.75rem;
  }
  .body {
    width: 100%;
  }
  :global(.chart-frame .body svg) {
    display: block;
    max-width: 100%;
  }
</style>
