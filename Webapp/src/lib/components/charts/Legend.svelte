<script lang="ts">
  interface Item {
    id: string;
    label: string;
    color: string;
  }
  interface Props {
    items: Item[];
    compact?: boolean;
  }
  let { items, compact = false }: Props = $props();
</script>

<ul class="legend" class:compact>
  {#each items as item (item.id)}
    <li>
      <span class="chip" style="background: {item.color};"></span>
      <span class="label">{item.label}</span>
    </li>
  {/each}
</ul>

<style>
  .legend {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem 1rem;
  }
  .legend.compact {
    gap: 0.35rem 0.75rem;
  }
  li {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.82rem;
    color: #334155;
    line-height: 1.2;
  }
  .chip {
    display: inline-block;
    width: 11px;
    height: 11px;
    border-radius: 3px;
    flex: none;
    box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.08);
  }
  .label {
    font-weight: 500;
  }
</style>
