<script lang="ts">
  export let backgroundGradient: boolean = true;
</script>

<div class="page-layout" class:gradient={backgroundGradient}>
  <slot />
</div>

<style>
  .page-layout {
    min-height: 100vh;
    background: var(--bg);
  }

  .page-layout.gradient {
    background: linear-gradient(to bottom, var(--bg), var(--bg-secondary));
  }
</style>