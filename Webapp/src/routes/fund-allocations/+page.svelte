<script lang="ts">
  import type { PageData } from "./$types";
  import Hero from "$lib/components/Hero.svelte";
  import Container from "$lib/components/Container.svelte";
  import Section from "$lib/components/Section.svelte";
  import PageLayout from "$lib/components/PageLayout.svelte";
  import SearchInput from "$lib/components/SearchInput.svelte";
  import ChartFrame from "$lib/components/charts/ChartFrame.svelte";
  import AllocationTreemap from "$lib/components/charts/AllocationTreemap.svelte";
  import CampusBarChart from "$lib/components/charts/CampusBarChart.svelte";
  import MultiLineChart from "$lib/components/charts/MultiLineChart.svelte";
  import { buildAllocationTree, isTotalRow } from "$lib/utils/ucopHierarchy";
  import { fmtUSDShort, CHART_COLORS } from "$lib/utils/chart";
  import {
    CAMPUS_KEYS,
    type CampusKey,
    type Series,
    type AllocationRow,
  } from "$lib/types/ucop";
  import { untrack } from "svelte";

  let { data }: { data: PageData } = $props();

  let selected = $state(untrack(() => data.selectedFy));
  let searchQuery = $state("");

  const selectedReport = $derived(
    data.reports.find((r) => r.fiscal_year === selected) ?? data.reports[0],
  );
  const tree = $derived(buildAllocationTree(selectedReport));

  // Top-level category segments across campuses (for stacked bar)
  const topLevelSegments = $derived.by(() => {
    const topLevels = (tree.children ?? []).filter(
      (c) => (c.value ?? 0) > 0,
    );
    return topLevels.slice(0, 8).map((top, i) => {
      // Sum per-campus contribution of this top-level category.
      // Walk the rows that descend from this top-level (by label + indent hint).
      const topIndent = top.indent ?? 0;
      const topLabel = top.name;
      let collecting = false;
      const values: Partial<Record<CampusKey, number>> = {};
      for (const row of selectedReport.rows) {
        if (row.indent === topIndent) {
          collecting = row.label === topLabel;
          continue;
        }
        if (!collecting) continue;
        if (row.indent <= topIndent) {
          collecting = false;
          continue;
        }
        if (isTotalRow(row.label)) continue;
        for (const c of CAMPUS_KEYS) {
          const v = row.amounts?.[c];
          if (typeof v === "number") values[c] = (values[c] ?? 0) + v;
        }
      }
      return {
        id: top.name,
        label: top.name,
        color: CHART_COLORS[i % CHART_COLORS.length],
        values,
      };
    });
  });

  // Top-level category totals across all fiscal years (multi-line chart)
  const categoryTrendSeries = $derived.by<Series[]>(() => {
    // Collect set of top-level category names across all reports
    const allCats = new Set<string>();
    const perYearTotals = new Map<string, Map<string, number>>(); // fy -> (category -> total system_total)
    for (const rep of data.reports) {
      const perCat = new Map<string, number>();
      // Walk rows; maintain current top-level context.
      let currentTop: string | null = null;
      let topIndent = -Infinity;
      for (const row of rep.rows) {
        if (row.indent <= 42) {
          currentTop = isTotalRow(row.label) ? null : row.label;
          topIndent = row.indent;
          continue;
        }
        if (!currentTop) continue;
        if (row.indent <= topIndent) {
          currentTop = null;
          continue;
        }
        if (isTotalRow(row.label)) continue;
        const v = row.amounts?.system_total;
        if (typeof v === "number")
          perCat.set(currentTop, (perCat.get(currentTop) ?? 0) + v);
      }
      perYearTotals.set(rep.fiscal_year, perCat);
      perCat.forEach((_, k) => allCats.add(k));
    }

    const cats = [...allCats].slice(0, 10);
    return cats.map((cat, i) => ({
      id: cat,
      label: cat,
      color: CHART_COLORS[i % CHART_COLORS.length],
      values: data.reports.map((rep) => ({
        year: fyStartYear(rep.fiscal_year),
        value: perYearTotals.get(rep.fiscal_year)?.get(cat) ?? 0,
      })),
    }));
  });

  function fyStartYear(fy: string): number {
    return Number(fy.split("-")[0]);
  }

  // Filtered raw table
  const filteredRows = $derived.by<AllocationRow[]>(() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return selectedReport.rows;
    return selectedReport.rows.filter((r) =>
      r.label.toLowerCase().includes(q),
    );
  });

  function fmtCell(v: number | undefined | null): string {
    if (typeof v !== "number") return "";
    return fmtUSDShort(v);
  }
</script>

<svelte:head>
  <title>Fund Allocations · UC Finances</title>
  <meta
    name="description"
    content="Ten years of state General Fund allocations to University of California campuses."
  />
</svelte:head>

<PageLayout>
  <Hero
    title="State General Fund Allocations"
    subtitle="Ten years of annual allocations from the California state General Fund to UC campuses and system-level programmes."
  />

  <Container>
    <Section title="Select fiscal year" icon="mdi:calendar">
      <div class="year-picker">
        {#each data.manifest.general_fund_allocations.fiscal_years as fy}
          <button
            class="year-btn"
            class:active={selected === fy}
            onclick={() => (selected = fy)}
          >
            FY {fy}
          </button>
        {/each}
      </div>
    </Section>

    <Section title="Allocation breakdown" icon="mdi:file-tree">
      <p class="desc">
        Every state General Fund line item in FY {selected}, grouped by top-level budget
        category. Larger rectangles = larger allocations.
      </p>
      <ChartFrame
        title="Treemap — FY {selected}"
        subtitle="Hover for breadcrumb path and share of system total"
        minHeight={560}
      >
        <AllocationTreemap {tree} />
      </ChartFrame>
    </Section>

    <Section title="Campus composition" icon="mdi:school">
      <p class="desc">
        Each campus's allocation, broken down by top-level category. Stacked so total
        column height = the campus's total state General Fund receipt.
      </p>
      <ChartFrame
        title="Per-campus stacked by category — FY {selected}"
        subtitle="Hover a segment for dollar amount and share"
        minHeight={480}
      >
        <CampusBarChart segments={topLevelSegments} sortBy="desc" />
      </ChartFrame>
    </Section>

    <Section title="Ten-year trend by category" icon="mdi:chart-line">
      <p class="desc">
        System-wide totals for each top-level allocation category across every fiscal
        year in the archive.
      </p>
      <ChartFrame
        title="Category totals {fyStartYear(data.reports[0].fiscal_year)}–{fyStartYear(
          data.reports[data.reports.length - 1].fiscal_year,
        )}"
        subtitle="One line per top-level category"
      >
        <MultiLineChart series={categoryTrendSeries} />
      </ChartFrame>
    </Section>

    <Section title="Raw data" icon="mdi:table">
      <p class="desc">Search the underlying rows for FY {selected}.</p>
      <div class="search-wrap">
        <SearchInput
          bind:value={searchQuery}
          placeholder="Filter labels (e.g. 'Agricultural', 'Labor Center')"
        />
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Label</th>
              <th class="num">Berkeley</th>
              <th class="num">Davis</th>
              <th class="num">Irvine</th>
              <th class="num">UCLA</th>
              <th class="num">Merced</th>
              <th class="num">Riverside</th>
              <th class="num">San Diego</th>
              <th class="num">UCSF</th>
              <th class="num">S.Barbara</th>
              <th class="num">S.Cruz</th>
              <th class="num">System total</th>
            </tr>
          </thead>
          <tbody>
            {#each filteredRows as row}
              <tr class:header={!row.amounts || Object.keys(row.amounts).length === 0}>
                <td class="lbl" style="padding-left: {Math.max(0, (row.indent - 40) * 0.5)}rem;">
                  {row.label}
                </td>
                <td class="num">{fmtCell(row.amounts?.berkeley)}</td>
                <td class="num">{fmtCell(row.amounts?.davis)}</td>
                <td class="num">{fmtCell(row.amounts?.irvine)}</td>
                <td class="num">{fmtCell(row.amounts?.los_angeles)}</td>
                <td class="num">{fmtCell(row.amounts?.merced)}</td>
                <td class="num">{fmtCell(row.amounts?.riverside)}</td>
                <td class="num">{fmtCell(row.amounts?.san_diego)}</td>
                <td class="num">{fmtCell(row.amounts?.san_francisco)}</td>
                <td class="num">{fmtCell(row.amounts?.santa_barbara)}</td>
                <td class="num">{fmtCell(row.amounts?.santa_cruz)}</td>
                <td class="num strong">{fmtCell(row.amounts?.system_total)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </Section>
  </Container>
</PageLayout>

<style>
  .desc {
    color: #475569;
    font-size: 1rem;
    line-height: 1.55;
    margin: 0 0 1.5rem;
    max-width: 820px;
  }
  .year-picker {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }
  .year-btn {
    padding: 0.55rem 1rem;
    border-radius: 0.6rem;
    border: 1px solid var(--border);
    background: white;
    color: var(--text-secondary);
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s ease;
    font-family: inherit;
    font-size: 0.95rem;
  }
  .year-btn:hover {
    background: var(--bg-secondary);
    color: var(--pri);
    transform: translateY(-1px);
  }
  .year-btn.active {
    background: linear-gradient(135deg, var(--founder), var(--pri));
    color: white;
    border-color: transparent;
  }
  .search-wrap {
    max-width: 520px;
    margin-bottom: 1rem;
  }
  .table-wrap {
    overflow-x: auto;
    background: white;
    border-radius: 12px;
    box-shadow: 0 10px 25px -15px rgba(0, 50, 98, 0.25);
    border: 1px solid rgba(0, 50, 98, 0.06);
  }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.85rem;
  }
  thead tr {
    background: rgba(0, 50, 98, 0.04);
  }
  th,
  td {
    padding: 0.5rem 0.75rem;
    text-align: left;
    border-bottom: 1px solid rgba(15, 23, 42, 0.05);
  }
  th {
    font-weight: 600;
    color: #475569;
  }
  td.lbl {
    max-width: 320px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  td.num,
  th.num {
    text-align: right;
    font-variant-numeric: tabular-nums;
    color: #0f172a;
    white-space: nowrap;
  }
  td.num.strong {
    font-weight: 600;
    color: var(--pri);
  }
  tr.header td {
    background: rgba(0, 50, 98, 0.03);
    font-weight: 600;
    color: var(--pri);
  }
</style>
