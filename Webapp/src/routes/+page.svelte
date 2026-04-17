<script lang="ts">
  import type { PageData } from "./$types";
  import Hero from "$lib/components/Hero.svelte";
  import StatCard from "$lib/components/StatCard.svelte";
  import Container from "$lib/components/Container.svelte";
  import Section from "$lib/components/Section.svelte";
  import Button from "$lib/components/Button.svelte";
  import PageLayout from "$lib/components/PageLayout.svelte";
  import ChartFrame from "$lib/components/charts/ChartFrame.svelte";
  import AllocationTreemap from "$lib/components/charts/AllocationTreemap.svelte";
  import StackedAreaChart from "$lib/components/charts/StackedAreaChart.svelte";
  import CampusBarChart from "$lib/components/charts/CampusBarChart.svelte";
  import { buildAllocationTree } from "$lib/utils/ucopHierarchy";
  import { RETRENDS_UNIT, fmtUSDShort } from "$lib/utils/chart";
  import { CAMPUS_KEYS, type CampusKey, type Series } from "$lib/types/ucop";

  let { data }: { data: PageData } = $props();

  const tree = $derived(buildAllocationTree(data.latestAlloc));

  const grandTotalRow = $derived(
    data.latestAlloc.rows.find((r) =>
      /State General Fund Support.*Permanent.*One.?Time/i.test(r.label),
    ),
  );
  const permRow = $derived(
    data.latestAlloc.rows.find((r) => /^Total Permanent State General Funds/i.test(r.label)),
  );
  const oneTimeRow = $derived(
    data.latestAlloc.rows.find((r) => /^Total One-Time Allocations/i.test(r.label)),
  );
  const totalGFAllocation = $derived(
    grandTotalRow?.amounts?.system_total ??
      (permRow?.amounts?.system_total ?? 0) + (oneTimeRow?.amounts?.system_total ?? 0),
  );

  const latestRetrendsYear = $derived(Number(data.latestRetrends.year_key));
  const sysTotals = $derived(
    data.latestRetrends.rows.filter(
      (r) =>
        r.page_title.includes("Revenues by Source and Expenses by Function") &&
        r.indent > 60 &&
        /^Total\s+(Revenues|Expenditures|Expenses)/i.test(r.label),
    ),
  );
  const totalRev = $derived(
    (sysTotals.find((r) => /Revenue/i.test(r.label))?.amounts?.[
      String(latestRetrendsYear)
    ] ?? 0) * RETRENDS_UNIT,
  );
  const totalExp = $derived(
    (sysTotals.find((r) => /Expenditure|Expense/i.test(r.label))?.amounts?.[
      String(latestRetrendsYear)
    ] ?? 0) * RETRENDS_UNIT,
  );

  const revYears = $derived(data.latestRetrends.year_columns.map(Number));
  const revenueSeries = $derived<Series[]>(
    data.latestRetrends.rows
      .filter(
        (r) =>
          r.page_title.includes("Revenues by Source and Expenses by Function") &&
          /^REVENUES/i.test(r.section) &&
          r.indent < 60 &&
          !/^Total/i.test(r.label),
      )
      .map((row) => ({
        id: row.label,
        label: row.label.replace(/\d+$/, "").trim(),
        values: revYears.map((y) => ({
          year: y,
          value: (row.amounts?.[String(y)] ?? 0) * RETRENDS_UNIT,
        })),
      })),
  );

  const campusTotalSegment = $derived([
    {
      id: "total",
      label: `FY ${data.latestAlloc.fiscal_year} allocation`,
      values: CAMPUS_KEYS.reduce<Partial<Record<CampusKey, number>>>((acc, c) => {
        const fromGrand = grandTotalRow?.amounts?.[c];
        if (typeof fromGrand === "number") {
          acc[c] = fromGrand;
        } else {
          acc[c] = (permRow?.amounts?.[c] ?? 0) + (oneTimeRow?.amounts?.[c] ?? 0);
        }
        return acc;
      }, {}),
    },
  ]);
</script>

<svelte:head>
  <title>UC Finances Explorer</title>
  <meta
    name="description"
    content="Visualise University of California general-fund allocations and system-wide revenue and expense trends."
  />
</svelte:head>

<PageLayout>
  <Hero
    title="UC Finances Explorer"
    subtitle="State general-fund allocations, revenues and expenses across the University of California."
  >
    <div class="hero-stats">
      <StatCard
        icon="mdi:cash-multiple"
        value={fmtUSDShort(totalGFAllocation)}
        label="State GF FY {data.latestAlloc.fiscal_year}"
        delay={100}
      />
      <StatCard
        icon="mdi:chart-line"
        value={fmtUSDShort(totalRev)}
        label="System revenues {latestRetrendsYear}"
        delay={200}
      />
      <StatCard
        icon="mdi:chart-line-variant"
        value={fmtUSDShort(totalExp)}
        label="System expenses {latestRetrendsYear}"
        delay={300}
      />
    </div>
  </Hero>

  <Container>
    <Section title="Where the money goes" icon="mdi:chart-tree">
      <p class="section-desc">
        State General Fund allocation to each campus and programme in fiscal year
        {data.latestAlloc.fiscal_year}. Larger rectangles indicate larger allocations;
        rectangles are grouped by top-level budget category.
      </p>
      <ChartFrame
        title="FY {data.latestAlloc.fiscal_year} state General Fund allocations"
        subtitle="System total — click through to explore ten years of history"
        minHeight={560}
      >
        <AllocationTreemap {tree} />
      </ChartFrame>
      <div class="cta">
        <Button href="/fund-allocations" variant="primary" icon="mdi:arrow-right">
          Explore fund allocations
        </Button>
      </div>
    </Section>

    <Section title="Revenues at system scale" icon="mdi:chart-areaspline">
      <p class="section-desc">
        UC system-wide revenues by source, {revYears[0]}–{revYears[revYears.length - 1]},
        drawn from the most recent Revenue and Expense Trends report.
      </p>
      <ChartFrame
        title="UC system revenues by source"
        subtitle="Stacked, in billions — hover to inspect"
      >
        <StackedAreaChart series={revenueSeries} />
      </ChartFrame>
      <div class="cta">
        <Button href="/revenue-expenses" variant="primary" icon="mdi:arrow-right">
          Revenue and expense trends
        </Button>
      </div>
    </Section>

    <Section title="Campus snapshot" icon="mdi:school">
      <p class="section-desc">
        Total state General Fund allocation received by each campus in FY
        {data.latestAlloc.fiscal_year}. Sum of all allocation line items reaching the
        campus.
      </p>
      <ChartFrame
        title="Campus total — FY {data.latestAlloc.fiscal_year}"
        subtitle="Sorted largest to smallest"
      >
        <CampusBarChart segments={campusTotalSegment} sortBy="desc" />
      </ChartFrame>
    </Section>
  </Container>
</PageLayout>

<style>
  .hero-stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 1rem;
    margin-top: 1.5rem;
  }
  .section-desc {
    color: #475569;
    font-size: 1rem;
    line-height: 1.55;
    margin: 0 0 1.5rem;
    max-width: 760px;
  }
  .cta {
    display: flex;
    justify-content: center;
    margin-top: 1rem;
  }
</style>
