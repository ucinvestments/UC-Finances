<script lang="ts">
  import { page } from "$app/state";
  import type { PageData } from "./$types";
  import Container from "$lib/components/Container.svelte";
  import Section from "$lib/components/Section.svelte";
  import SearchInput from "$lib/components/SearchInput.svelte";
  import ChartFrame from "$lib/components/charts/ChartFrame.svelte";
  import ChartDataTable from "$lib/components/charts/ChartDataTable.svelte";
  import AllocationTreemap from "$lib/components/charts/AllocationTreemap.svelte";
  import CampusBarChart from "$lib/components/charts/CampusBarChart.svelte";
  import MultiLineChart from "$lib/components/charts/MultiLineChart.svelte";
  import {
    allocationTree,
    CANONICAL_CATEGORIES,
    summarizeAllocations,
    type AllocationSummary,
  } from "$lib/utils/ucopHierarchy";
  import { fmtNumber, fmtUSD } from "$lib/utils/chart";
  import { campusDisplayName } from "$lib/utils/campusColors";
  import { CAMPUS_KEYS, type AllocationColumn, type Series } from "$lib/types/ucop";

  let { data }: { data: PageData } = $props();
  let searchQuery = $state("");

  const selectedIndex = $derived.by(() => {
    const index = data.fiscalYears.indexOf(page.url.searchParams.get("fy") ?? "");
    return index === -1 ? data.fiscalYears.length - 1 : index;
  });
  const selected = $derived(data.fiscalYears[selectedIndex]);
  const selectedReport = $derived(data.reports[selectedIndex]);
  const summaries = $derived(data.reports.map(summarizeAllocations));
  const allocation = $derived(summaries[selectedIndex]);
  const tree = $derived(allocationTree(allocation));
  const campusTotal = $derived([
    { id: "reported", label: "Reported allocation", values: allocation.reported },
  ]);
  const trend = $derived<Series[]>([
    {
      id: "reported-total",
      label: "Reported system total",
      values: summaries.flatMap(({ fiscalYear, reported }) =>
        reported.system_total === undefined
          ? []
          : [{ year: Number(fiscalYear.slice(0, 4)), value: reported.system_total }],
      ),
    },
  ]);

  const allColumns: { key: AllocationColumn; label: string }[] = [
    ...CAMPUS_KEYS.map((key) => ({ key, label: campusDisplayName(key) })),
    { key: "campus_total", label: "Campus total" },
    { key: "ucop_to_be_allocated", label: "UCOP to be allocated" },
    { key: "capital_debt_service", label: "Capital debt service" },
    { key: "off_campus_programs", label: "Off-campus programs" },
    { key: "system_total", label: "System total" },
  ];
  const visibleColumns = $derived(
    allColumns.filter(({ key }) => selectedReport.columns.includes(key)),
  );

  const filteredRows = $derived(
    selectedReport.rows.filter((row) =>
      row.label.toLowerCase().includes(searchQuery.trim().toLowerCase()),
    ),
  );

  function formatAllocation(value: number | undefined) {
    return value === undefined ? "—" : fmtUSD(value);
  }

  function tableEntries(summary: AllocationSummary) {
    return [
      ...summary.categories,
      { label: "Difference from reported total", amounts: summary.difference },
      { label: "Reported total", amounts: summary.reported },
    ];
  }

  const historicalRows = $derived([
    ...CANONICAL_CATEGORIES.map((label, index) => ({
      label,
      values: summaries.map((summary) =>
        formatAllocation(summary.categories[index].amounts.system_total),
      ),
    })),
    {
      label: "Difference from reported total",
      values: summaries.map((summary) => formatAllocation(summary.difference.system_total)),
    },
    {
      label: "Reported total",
      values: summaries.map((summary) => formatAllocation(summary.reported.system_total)),
    },
  ]);
</script>

<svelte:head>
  <title>State General Fund allocations · UC Finances</title>
  <meta
    name="description"
    content="Explore California state General Fund allocations to the University of California."
  />
</svelte:head>

<Container>
  <header class="introduction">
    <h1>State General Fund allocations</h1>
    <p>Annual California state support for UC campuses and systemwide programs.</p>
  </header>

  <Section title="Fiscal year">
    <nav class="year-picker" aria-label="Fiscal year">
      {#each data.fiscalYears as fy}
        <a class="year-btn" aria-current={selected === fy ? "true" : undefined} href="?fy={fy}">
          {fy}
        </a>
      {/each}
    </nav>
    <div class="report-summary">
      <p><strong>{formatAllocation(allocation.reported.system_total)}</strong> reported system total for FY {selected}</p>
      <a
        href={selectedReport.source_url || "https://www.ucop.edu/operating-budget/budgets-and-reports/other-resources/"}
      >
        {selectedReport.source_url ? `UCOP allocation report for FY ${selected} (PDF)` : "UCOP allocation reports"}
      </a>
    </div>
  </Section>

  <Section title="System allocation by category">
    <p class="desc">
      The treemap shows positive category amounts. The table includes reductions and
      any difference between category amounts and UCOP’s reported total.
    </p>
    {#if tree.children?.length}
      <ChartFrame title="System allocations for FY {selected}">
        <AllocationTreemap {tree} height={320} />
        <ChartDataTable
          caption="System allocations for FY {selected}"
          rowHeading="Category"
          toggleLabel="View system category breakdown"
          columns={["System total"]}
          rows={tableEntries(allocation).map(({ label, amounts }) => ({
            label,
            values: [formatAllocation(amounts.system_total)],
          }))}
        />
      </ChartFrame>
    {:else}
      <p>Category values are unavailable for this fiscal year.</p>
    {/if}
  </Section>

  <Section title="Allocations by campus">
    <p class="desc">
      Bars show the final allocation reported for each campus.
      The category table includes campus-level adjustments.
    </p>
    {#if selected === "2017-18" || selected === "2018-19"}
      <p class="desc">
        UCOP’s published campus totals exclude the listed one-time amounts in FY {selected}.
        The table shows the resulting difference. The published system total includes those amounts.
      </p>
    {/if}
    <ChartFrame title="Reported campus totals for FY {selected}">
      <CampusBarChart segments={campusTotal} sortBy="desc" />
    </ChartFrame>
    <ChartDataTable
      caption="Allocations by category and campus for FY {selected}"
      rowHeading="Category"
      toggleLabel="View campus category breakdown"
      columns={CAMPUS_KEYS.map(campusDisplayName)}
      rows={tableEntries(allocation).map(({ label, amounts }) => ({
        label,
        values: CAMPUS_KEYS.map((campus) => formatAllocation(amounts[campus])),
      }))}
    />
  </Section>

  <Section title="Allocations over time">
    <p class="desc">
      The chart shows total state General Fund support reported by UCOP for each fiscal year.
      Year labels show the start of the fiscal year.
    </p>
    <ChartFrame title="Reported system total by fiscal year">
      <MultiLineChart series={trend} animate={false} />
    </ChartFrame>
    <ChartDataTable
      caption="System allocations by category and fiscal year"
      rowHeading="Category"
      toggleLabel="View category history"
      columns={data.fiscalYears}
      rows={historicalRows}
    />
  </Section>

  <Section title="Source table">
    <p class="desc" id="allocation-source-description">
      Values are shown as extracted from the FY {selected} report.
      The table also contains enrollment counts and per-student figures.
    </p>
    <div class="search-wrap">
      <SearchInput bind:value={searchQuery} placeholder="Filter allocation line items" />
    </div>
    {#if filteredRows.length}
      <!-- svelte-ignore a11y_no_noninteractive_tabindex (Keyboard users need to scroll wide tables.) -->
      <div class="table-wrap" role="region" aria-label="Allocation report rows" tabindex="0">
        <table aria-describedby="allocation-source-description">
          <thead>
            <tr>
              <th scope="col" class="sticky-left">Line item</th>
              {#each visibleColumns as { label }}
                <th scope="col" class="num">{label}</th>
              {/each}
            </tr>
          </thead>
          <tbody>
            {#each filteredRows as row}
              <tr class:header={Object.keys(row.amounts).length === 0}>
                <th scope="row" class="lbl sticky-left">{row.label || "Unlabelled in extract"}</th>
                {#each visibleColumns as { key }}
                  {@const value = row.amounts[key]}
                  <td class="num">{value === undefined ? "—" : fmtNumber(value)}</td>
                {/each}
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <p>No line items match “{searchQuery}”.</p>
    {/if}
  </Section>
</Container>

<style>
  .introduction {
    border-inline-start: 0.25rem solid var(--sec);
    padding-inline-start: clamp(1rem, 3vw, 2rem);
    margin-block: clamp(0.5rem, 3vw, 2rem) clamp(2rem, 5vw, 4rem);
  }

  h1 {
    color: var(--pri);
    font-family: "Space Grotesk", "Inter", sans-serif;
    font-size: clamp(2rem, 5.5vw, 4rem);
    line-height: 1.06;
    letter-spacing: -0.045em;
    margin-block-end: 0.75rem;
  }

  p {
    color: var(--text-secondary);
  }

  .desc {
    max-width: 75ch;
    margin-block-end: 1rem;
  }

  .year-picker {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-block-end: 1rem;
  }

  .year-btn {
    min-height: 2.75rem;
    padding: 0.55rem 0.8rem;
    border-radius: 0.4rem;
    border: 1px solid var(--border);
    color: var(--pri);
    font-weight: 500;
  }

  .year-btn:hover {
    background: var(--bg-secondary);
  }

  .year-btn[aria-current] {
    background: var(--pri);
    color: white;
    border-color: var(--pri);
  }

  .report-summary {
    line-height: 1.6;
  }

  .report-summary strong,
  .report-summary a {
    color: var(--pri);
  }

  .report-summary a {
    text-decoration: underline;
    text-underline-offset: 0.18em;
  }

  .search-wrap {
    max-width: 32rem;
    margin-block-end: 1rem;
  }

  .table-wrap {
    overflow-x: auto;
    border: 1px solid var(--border);
    border-radius: 0.5rem;
    background: white;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: var(--text-sm);
  }

  th,
  td {
    padding: 0.5rem 0.75rem;
    text-align: start;
    border-block-end: 1px solid var(--border);
  }

  thead th,
  .header .sticky-left {
    background: var(--bg-secondary);
  }

  .lbl {
    width: 24rem;
    min-width: 11rem;
    font-weight: 500;
  }

  .num {
    text-align: end;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }

  .sticky-left {
    position: sticky;
    inset-inline-start: 0;
    background: white;
    border-inline-end: 1px solid var(--border);
  }
</style>
