<script lang="ts">
  import type { PageData } from "./$types";
  import Hero from "$lib/components/Hero.svelte";
  import Container from "$lib/components/Container.svelte";
  import Section from "$lib/components/Section.svelte";
  import PageLayout from "$lib/components/PageLayout.svelte";
  import ChartFrame from "$lib/components/charts/ChartFrame.svelte";
  import StackedAreaChart from "$lib/components/charts/StackedAreaChart.svelte";
  import MultiLineChart from "$lib/components/charts/MultiLineChart.svelte";
  import MedCenterDonut from "$lib/components/charts/MedCenterDonut.svelte";
  import Legend from "$lib/components/charts/Legend.svelte";
  import {
    systemRevExp,
    perCampusRevExp,
    medCenterRevExp,
    toSeries,
    type RevExpKind,
  } from "$lib/utils/retrendsMerge";
  import { RETRENDS_UNIT, paletteFor } from "$lib/utils/chart";
  import { getCampusTheme } from "$lib/utils/campusColors";
  import type { Series } from "$lib/types/ucop";

  let { data }: { data: PageData } = $props();

  let stackMode = $state<"none" | "expand">("none");
  let donutKind = $state<RevExpKind>("revenue");
  let donutYear = $state<number>(0);

  const years = $derived(data.merged.years);

  // System rev/exp stacked areas
  const systemRev = $derived(paletteFor(toSeries(systemRevExp(data.merged, "revenue"))));
  const systemExp = $derived(paletteFor(toSeries(systemRevExp(data.merged, "expense"))));

  // Per-campus totals — one line per campus
  const CAMPUS_LABELS = [
    "Berkeley",
    "Davis",
    "Irvine",
    "Los Angeles",
    "Merced",
    "Riverside",
    "San Diego",
    "San Francisco",
    "Santa Barbara",
    "Santa Cruz",
  ];
  const campusRevSeries = $derived.by<Series[]>(() => {
    const raw = perCampusRevExp(data.merged, "revenue").filter((s) =>
      CAMPUS_LABELS.includes(s.label_norm),
    );
    return raw.map((s) => ({
      id: s.label_norm,
      label: s.label_norm,
      color: getCampusTheme(s.label_norm).primary,
      values: Object.entries(s.amounts)
        .map(([y, v]) => ({ year: Number(y), value: v * RETRENDS_UNIT }))
        .sort((a, b) => a.year - b.year),
    }));
  });
  const campusExpSeries = $derived.by<Series[]>(() => {
    const raw = perCampusRevExp(data.merged, "expense").filter((s) =>
      CAMPUS_LABELS.includes(s.label_norm),
    );
    return raw.map((s) => ({
      id: s.label_norm,
      label: s.label_norm,
      color: getCampusTheme(s.label_norm).primary,
      values: Object.entries(s.amounts)
        .map(([y, v]) => ({ year: Number(y), value: v * RETRENDS_UNIT }))
        .sort((a, b) => a.year - b.year),
    }));
  });

  // Medical centers (series over years, one per center)
  const MED_LABELS = ["Davis", "Irvine", "Los Angeles", "San Diego", "San Francisco"];
  const medCenterSeries = $derived.by<Series[]>(() => {
    const raw = medCenterRevExp(data.merged, donutKind).filter((s) =>
      MED_LABELS.includes(s.label_norm),
    );
    return raw.map((s) => ({
      id: s.label_norm,
      label: s.label_norm,
      color: getCampusTheme(s.label_norm).primary,
      values: Object.entries(s.amounts)
        .map(([y, v]) => ({ year: Number(y), value: v * RETRENDS_UNIT }))
        .sort((a, b) => a.year - b.year),
    }));
  });

  // Donut slices for a single year
  const donutSlices = $derived.by(() => {
    const year = donutYear || (years.length ? years[years.length - 1] : 0);
    const raw = medCenterRevExp(data.merged, donutKind).filter((s) =>
      MED_LABELS.includes(s.label_norm),
    );
    return raw
      .map((s) => ({
        label: s.label_norm,
        value: (s.amounts[year] ?? 0) * RETRENDS_UNIT,
        themeKey: s.label_norm,
      }))
      .filter((s) => s.value > 0);
  });

  const activeDonutYear = $derived(donutYear || years[years.length - 1]);
</script>

<svelte:head>
  <title>Revenue &amp; Expenses · UC Finances</title>
  <meta
    name="description"
    content="University of California system-wide and per-campus revenues and expenses since 2015, from UC's Revenue and Expense Trends reports."
  />
</svelte:head>

<PageLayout>
  <Hero
    title="Revenue and Expense Trends"
    subtitle="Eleven years of UC system-wide, per-campus, and medical-centre revenues and expenses — merged from UC's annual Revenue &amp; Expense Trends reports."
  />

  <Container>
    <Section title="System revenues and expenses" icon="mdi:chart-areaspline">
      <p class="desc">
        UC system-wide revenues by source and expenses by function, with overlapping
        years reconciled by keeping the value from the most recent report (capturing
        restatements).
      </p>
      <div class="toggle-row">
        <span class="toggle-label">Stack mode:</span>
        <button
          class="toggle"
          class:active={stackMode === "none"}
          onclick={() => (stackMode = "none")}
        >
          Dollars
        </button>
        <button
          class="toggle"
          class:active={stackMode === "expand"}
          onclick={() => (stackMode = "expand")}
        >
          100%
        </button>
      </div>
      <ChartFrame
        title="Revenues by source"
        subtitle="Hover any year to see the breakdown"
      >
        {#snippet legend()}
          <Legend items={systemRev.map((s) => ({ id: s.id, label: s.label, color: s.color! }))} />
        {/snippet}
        <StackedAreaChart series={systemRev} stackOffset={stackMode} />
      </ChartFrame>
      <ChartFrame
        title="Expenses by function"
        subtitle="Hover any year to see the breakdown"
      >
        {#snippet legend()}
          <Legend items={systemExp.map((s) => ({ id: s.id, label: s.label, color: s.color! }))} />
        {/snippet}
        <StackedAreaChart series={systemExp} stackOffset={stackMode} />
      </ChartFrame>
    </Section>

    <Section title="Campus revenue and expense over time" icon="mdi:school">
      <p class="desc">
        Total revenue and total expense for each campus, covering the full archive. Each
        campus uses its own brand colour.
      </p>
      <ChartFrame
        title="Total campus revenue"
        subtitle="One line per campus, all years"
      >
        {#snippet legend()}
          <Legend items={campusRevSeries.map((s) => ({ id: s.id, label: s.label, color: s.color! }))} />
        {/snippet}
        <MultiLineChart series={campusRevSeries} yLabel="Total revenue" />
      </ChartFrame>
      <ChartFrame
        title="Total campus expense"
        subtitle="One line per campus, all years"
      >
        {#snippet legend()}
          <Legend items={campusExpSeries.map((s) => ({ id: s.id, label: s.label, color: s.color! }))} />
        {/snippet}
        <MultiLineChart series={campusExpSeries} yLabel="Total expense" />
      </ChartFrame>
    </Section>

    <Section title="Medical centres" icon="mdi:hospital-building">
      <p class="desc">
        UC operates medical centres at Davis, Irvine, Los Angeles, San Diego and San
        Francisco. Medical-centre revenue is now one of the largest components of total
        UC revenue.
      </p>
      <div class="toggle-row">
        <span class="toggle-label">Show:</span>
        <button
          class="toggle"
          class:active={donutKind === "revenue"}
          onclick={() => (donutKind = "revenue")}
        >
          Revenue
        </button>
        <button
          class="toggle"
          class:active={donutKind === "expense"}
          onclick={() => (donutKind = "expense")}
        >
          Expense
        </button>
        <span class="toggle-label" style="margin-left:1.25rem;">Year:</span>
        <select
          bind:value={donutYear}
          class="year-select"
          aria-label="Select year for donut chart"
        >
          <option value={0}>Latest ({years.at(-1) ?? ""})</option>
          {#each years.slice().reverse() as y}
            <option value={y}>{y}</option>
          {/each}
        </select>
      </div>
      <div class="med-grid">
        <ChartFrame
          title="Medical centre {donutKind} share — {activeDonutYear}"
          subtitle="Click a slice to inspect"
          minHeight={480}
        >
          <MedCenterDonut
            slices={donutSlices}
            centerLabel="Total {donutKind}"
          />
        </ChartFrame>
        <ChartFrame
          title="Medical centre {donutKind} over time"
          subtitle="Lines coloured by centre"
          minHeight={480}
        >
          {#snippet legend()}
            <Legend items={medCenterSeries.map((s) => ({ id: s.id, label: s.label, color: s.color! }))} />
          {/snippet}
          <MultiLineChart series={medCenterSeries} />
        </ChartFrame>
      </div>
    </Section>
  </Container>
</PageLayout>

<style>
  .desc {
    color: #475569;
    font-size: 1rem;
    line-height: 1.55;
    margin: 0 0 1.25rem;
    max-width: 820px;
  }
  .toggle-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-wrap: wrap;
    margin-bottom: 1.25rem;
  }
  .toggle-label {
    color: var(--text-secondary);
    font-weight: 500;
    font-size: 0.92rem;
  }
  .toggle {
    padding: 0.45rem 0.95rem;
    border-radius: 0.55rem;
    border: 1px solid var(--border);
    background: white;
    color: var(--text-secondary);
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s ease;
    font-family: inherit;
    font-size: 0.92rem;
  }
  .toggle:hover {
    background: var(--bg-secondary);
    color: var(--pri);
  }
  .toggle.active {
    background: linear-gradient(135deg, var(--founder), var(--pri));
    color: white;
    border-color: transparent;
  }
  .year-select {
    padding: 0.45rem 0.75rem;
    border-radius: 0.55rem;
    border: 1px solid var(--border);
    background: white;
    color: var(--text-primary);
    font-family: inherit;
    font-size: 0.92rem;
    cursor: pointer;
  }
  .med-grid {
    display: grid;
    grid-template-columns: minmax(320px, 1fr) minmax(320px, 1.25fr);
    gap: 1.5rem;
  }
  @media (max-width: 900px) {
    .med-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
