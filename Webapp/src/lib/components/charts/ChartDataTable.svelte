<script lang="ts">
  let {
    caption,
    rowHeading,
    columns,
    rows,
    toggleLabel = "View chart data",
  }: {
    caption: string;
    rowHeading: string;
    columns: string[];
    rows: { label: string; values: (string | undefined)[] }[];
    toggleLabel?: string;
  } = $props();

  function scrollTable(event: KeyboardEvent & { currentTarget: HTMLDivElement }) {
    if (event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) return;

    const scroller = event.currentTarget;
    if (scroller.scrollWidth <= scroller.clientWidth) return;

    if (event.key === "ArrowLeft" || event.key === "ArrowRight") {
      event.preventDefault();
      scroller.scrollBy(event.key === "ArrowRight" ? 80 : -80, 0);
    }
  }
</script>

{#if rows.length > 0 && columns.length > 0}
  <details>
    <summary>{toggleLabel}</summary>
    <!-- svelte-ignore a11y_no_noninteractive_tabindex, a11y_no_noninteractive_element_interactions (Keyboard users need to scroll wide tables.) -->
    <div
      class="table-scroll"
      role="region"
      aria-label={caption}
      tabindex="0"
      onkeydown={scrollTable}
    >
      <table>
        <caption>{caption}</caption>
        <thead>
          <tr>
            <th scope="col">{rowHeading}</th>
            {#each columns as column}
              <th scope="col">{column}</th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each rows as row}
            <tr>
              <th scope="row">{row.label}</th>
              {#each row.values as value}
                <td>{value ?? "Unavailable"}</td>
              {/each}
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </details>
{/if}

<style>
  details {
    border-block-start: 1px solid var(--border);
    margin-block-start: 0.75rem;
  }

  summary {
    min-height: 2.75rem;
    padding-block: 0.625rem;
    color: var(--pri);
    cursor: pointer;
    font-weight: var(--font-semibold);
  }

  .table-scroll {
    overflow-x: auto;
    max-width: 100%;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: var(--text-sm);
    font-variant-numeric: tabular-nums;
  }

  caption {
    color: var(--text-secondary);
    text-align: start;
    padding-block: 0.25rem 0.5rem;
  }

  th,
  td {
    padding: 0.625rem 0.75rem;
    border-block-end: 1px solid var(--border);
    text-align: end;
  }

  th:first-child {
    text-align: start;
    min-width: 9rem;
  }

  td {
    white-space: nowrap;
  }

  thead {
    background: var(--bg-secondary);
  }
</style>
