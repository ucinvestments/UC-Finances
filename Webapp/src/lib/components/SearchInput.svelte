<script lang="ts">
  import Icon from "@iconify/svelte";

  export let value: string = "";
  export let placeholder: string = "Search...";
  export let onInput: ((value: string) => void) | undefined = undefined;
  export let onClear: (() => void) | undefined = undefined;
</script>

<div class="search-wrapper">
  <Icon icon="mdi:magnify" class="search-icon" />
  <input
    type="text"
    {placeholder}
    bind:value
    on:input={() => onInput?.(value)}
    class="search-input"
  />
  {#if value}
    <button
      on:click={() => {
        value = "";
        onClear?.();
      }}
      class="clear-btn"
      aria-label="Clear search"
    >
      <Icon icon="mdi:close" />
    </button>
  {/if}
</div>

<style>
  .search-wrapper {
    position: relative;
    width: 100%;
    max-width: 600px;
  }

  :global(.search-icon) {
    position: absolute;
    left: 1.25rem;
    top: 50%;
    transform: translateY(-50%);
    color: var(--text-secondary);
    font-size: 1.25rem;
  }

  .search-input {
    width: 100%;
    padding: 1rem 3.5rem;
    border: 2px solid var(--border);
    border-radius: 1rem;
    font-size: 1rem;
    transition: all 0.3s ease;
    background: white;
    box-shadow: var(--shadow-sm);
  }

  .search-input:focus {
    outline: none;
    border-color: var(--founder);
    box-shadow: 0 0 0 3px rgba(59, 126, 161, 0.1), var(--shadow-md);
    transform: translateY(-1px);
  }

  .clear-btn {
    position: absolute;
    right: 1.25rem;
    top: 50%;
    transform: translateY(-50%);
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 0.25rem;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 0.5rem;
    transition: all 0.2s ease;
  }

  .clear-btn:hover {
    background: var(--bg-secondary);
    color: var(--text-primary);
  }

  @media (max-width: 768px) {
    .search-input {
      padding: 0.875rem 3rem 0.875rem 2.5rem;
      font-size: 0.925rem;
    }
  }
</style>