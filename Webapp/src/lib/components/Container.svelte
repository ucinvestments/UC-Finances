<script lang="ts">
  export let maxWidth: string = "1200px";
  export let padding: string = "3rem 2rem";
</script>

<div class="container" style="max-width: {maxWidth}; padding: {padding}">
  <slot />
</div>

<style>
  .container {
    margin: 0 auto;
  }

  @media (max-width: 768px) {
    .container {
      padding: 2rem 1rem !important;
    }
  }
</style>