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
  import ChartDataTable from "$lib/components/charts/ChartDataTable.svelte";
  import {
    toSeries,
    type RevExpKind,
    type RetrendsSource,
  } from "$lib/utils/retrendsMerge";
  import { fmtUSD, paletteFor } from "$lib/utils/chart";
  import { getCampusTheme } from "$lib/utils/campusColors";
  import type { Series } from "$lib/types/ucop";

  let { data }: { data: PageData } = $props();

  let stackMode = $state<"none" | "expand">("none");
  let donutKind = $state<RevExpKind>("revenue");
  let donutYear = $state<number>(0);

  const years = $derived(data.merged.years);
  const tables = $derived(data.merged.tables);

  const sourceColumns = [
    { scope: "system", kind: "revenue", label: "System revenue" },
    { scope: "system", kind: "expense", label: "System expenses" },
    { scope: "campus", kind: "revenue", label: "Campus revenue" },
    { scope: "campus", kind: "expense", label: "Campus expenses" },
    { scope: "medical", kind: "revenue", label: "Medical center revenue" },
    { scope: "medical", kind: "expense", label: "Medical center expenses" },
  ] as const;

  const coverage = $derived(
    years.length === 1 ? `FY ${years[0]}` : `FY ${years[0]}–${years.at(-1)}`,
  );
  const systemRev = $derived(paletteFor(toSeries(tables.system.revenue)));
  const systemExp = $derived(paletteFor(toSeries(tables.system.expense)));
  const campusRevSeries = $derived(
    withCampusColors(toSeries(tables.campus.revenue)),
  );
  const campusExpSeries = $derived(
    withCampusColors(toSeries(tables.campus.expense)),
  );
  const medCenterSeries = $derived(
    withCampusColors(toSeries(tables.medical[donutKind])),
  );
  const medicalYears = $derived(
    Object.keys(tables.medical[donutKind].sources).map(Number),
  );
  const activeDonutYear = $derived(
    medicalYears.includes(donutYear) ? donutYear : medicalYears.at(-1),
  );
  const donutValues = $derived(
    medCenterSeries.map(({ label, values }) => ({
      label,
      value: values.find((point) => point.year === activeDonutYear)?.value,
    })),
  );
  const donutSlices = $derived(
    donutValues.flatMap(({ label, value }) =>
      value !== undefined && value > 0
        ? [{ label, value, themeKey: label }]
        : [],
    ),
  );
  const donutTotal = $derived(
    donutSlices.reduce((sum, slice) => sum + slice.value, 0),
  );

  function withCampusColors(series: Series[]): Series[] {
    return series.map((item) => ({
      ...item,
      color: getCampusTheme(item.label).primary,
    }));
  }

  function pageLink(source: RetrendsSource): string {
    return `${source.url.split("#")[0]}#page=${source.page}`;
  }
</script>

<svelte:head>
  <title>Revenue &amp; Expenses · UC Finances</title>
  <meta
    name="description"
    content={`University of California system, campus, and medical center revenues and expenses for ${coverage}.`}
  />
</svelte:head>

<PageLayout>
  <Hero
    title="Revenue and Expense Trends"
    subtitle={`University of California system, campus, and medical center finances for ${coverage}.`}
  />

  <Container>
    <Section title="System revenues and expenses" icon="mdi:chart-areaspline">
      <p class="desc">
        Each fiscal year uses the newest UC report whose extracted breakdown
        reconciles to the published total. <a href="#source-reports"
          >See source reports.</a
        >
      </p>
      <div class="toggle-row">
        <span class="toggle-label">Chart scale</span>
        <button
          class="toggle"
          class:active={stackMode === "none"}
          aria-pressed={stackMode === "none"}
          onclick={() => (stackMode = "none")}
        >
          Dollars
        </button>
        <button
          class="toggle"
          class:active={stackMode === "expand"}
          aria-pressed={stackMode === "expand"}
          onclick={() => (stackMode = "expand")}
        >
          100%
        </button>
      </div>
      <ChartFrame title="Revenues by source">
        {#snippet legend()}
          <Legend
            items={systemRev.map((s) => ({
              id: s.id,
              label: s.label,
              color: s.color!,
            }))}
          />
        {/snippet}
        <StackedAreaChart series={systemRev} stackOffset={stackMode} />
      </ChartFrame>
      <ChartFrame title="Expenses by function">
        {#snippet legend()}
          <Legend
            items={systemExp.map((s) => ({
              id: s.id,
              label: s.label,
              color: s.color!,
            }))}
          />
        {/snippet}
        <StackedAreaChart series={systemExp} stackOffset={stackMode} />
      </ChartFrame>
    </Section>

    <Section title="Campus revenue and expense over time" icon="mdi:school">
      <p class="desc">
        Campus totals are shown as reported by UC. The system totals also
        include the Department of Energy laboratories and systemwide activities.
      </p>
      <ChartFrame title="Total campus revenue">
        {#snippet legend()}
          <Legend
            items={campusRevSeries.map((s) => ({
              id: s.id,
              label: s.label,
              color: s.color!,
            }))}
          />
        {/snippet}
        <MultiLineChart series={campusRevSeries} yLabel="Total revenue" />
      </ChartFrame>
      <ChartFrame title="Total campus expense">
        {#snippet legend()}
          <Legend
            items={campusExpSeries.map((s) => ({
              id: s.id,
              label: s.label,
              color: s.color!,
            }))}
          />
        {/snippet}
        <MultiLineChart series={campusExpSeries} yLabel="Total expense" />
      </ChartFrame>
    </Section>

    <Section title="Medical centers" icon="mdi:hospital-building">
      <p class="desc">
        UC operates medical centers at Davis, Irvine, Los Angeles, San Diego and
        San Francisco. Medical center revenue is now one of the largest
        components of total UC revenue.
      </p>
      <div class="toggle-row">
        <span class="toggle-label">Measure</span>
        <button
          class="toggle"
          class:active={donutKind === "revenue"}
          aria-pressed={donutKind === "revenue"}
          onclick={() => (donutKind = "revenue")}
        >
          Revenue
        </button>
        <button
          class="toggle"
          class:active={donutKind === "expense"}
          aria-pressed={donutKind === "expense"}
          onclick={() => (donutKind = "expense")}
        >
          Expense
        </button>
        <label class="toggle-label year-label" for="medical-year"
          >Fiscal year</label
        >
        <select
          id="medical-year"
          value={activeDonutYear}
          onchange={(event) => (donutYear = Number(event.currentTarget.value))}
          class="year-select"
          disabled={!medicalYears.length}
        >
          {#each medicalYears.slice().reverse() as year (year)}
            <option value={year}>{year}</option>
          {/each}
        </select>
      </div>
      <div class="med-grid">
        <ChartFrame
          title={`Medical center ${donutKind} share${activeDonutYear ? ` · FY ${activeDonutYear}` : ""}`}
          minHeight={480}
        >
          {#if donutSlices.length}
            <MedCenterDonut
              slices={donutSlices}
              centerLabel="Total {donutKind}"
            />
            <ChartDataTable
              caption={`Medical center ${donutKind}, FY ${activeDonutYear}`}
              rowHeading="Medical center"
              columns={["Amount", "Share"]}
              rows={donutValues.map(({ label, value }) => ({
                label,
                values:
                  value === undefined
                    ? [undefined, undefined]
                    : [
                        fmtUSD(value),
                        `${((value / donutTotal) * 100).toFixed(1)}%`,
                      ],
              }))}
            />
          {:else}
            <p>No validated figures for this period.</p>
          {/if}
        </ChartFrame>
        <ChartFrame
          title="Medical center {donutKind} over time"
          minHeight={480}
        >
          {#snippet legend()}
            <Legend
              items={medCenterSeries.map((s) => ({
                id: s.id,
                label: s.label,
                color: s.color!,
              }))}
            />
          {/snippet}
          <MultiLineChart series={medCenterSeries} />
        </ChartFrame>
      </div>
    </Section>
    <div id="source-reports">
      <Section title="Source reports" icon="mdi:file-document-outline">
        <p class="desc">
          UC's annual reports cover five fiscal years at a time. If a newer
          extraction fails to reconcile, the explorer uses an earlier validated
          report. Campus and medical center breakdowns can use different reports
          for the same year.
        </p>
        <details>
          <summary>View reports by fiscal year</summary>
          <!-- svelte-ignore a11y_no_noninteractive_tabindex (Keyboard users need to scroll wide tables.) -->
          <div
            class="reports-scroll"
            role="region"
            aria-label="Reports by fiscal year"
            tabindex="0"
          >
            <table class="reports-table">
              <caption
                >Report edition and PDF page used for each breakdown</caption
              >
              <thead>
                <tr>
                  <th scope="col">Fiscal year</th>
                  {#each sourceColumns as { label }}
                    <th scope="col">{label}</th>
                  {/each}
                </tr>
              </thead>
              <tbody>
                {#each years.slice().reverse() as year (year)}
                  <tr>
                    <th scope="row">{year}</th>
                    {#each sourceColumns as { scope, kind }}
                      {@const source = tables[scope][kind].sources[year]}
                      <td>
                        {#if source?.url}
                          <a href={pageLink(source)}
                            >{source.reportYear} · p. {source.page}</a
                          >
                        {:else if source}
                          {source.reportYear} · p. {source.page}
                        {:else}
                          Unavailable
                        {/if}
                      </td>
                    {/each}
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </details>
      </Section>
    </div>
  </Container>
</PageLayout>

<style>
  .desc {
    color: var(--text-secondary);
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
  .year-label {
    margin-inline-start: 0.75rem;
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
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.25fr);
    gap: 1.5rem;
  }
  summary {
    cursor: pointer;
    font-weight: 600;
  }
  .reports-scroll {
    overflow-x: auto;
    margin-top: 1rem;
  }
  .reports-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.9rem;
  }
  .reports-table caption {
    text-align: left;
    padding: 0.5rem;
  }
  .reports-table th,
  .reports-table td {
    text-align: left;
    border-bottom: 1px solid var(--border);
    padding: 0.75rem 0.5rem;
  }
  .reports-table td a {
    white-space: nowrap;
  }
  a {
    text-decoration: underline;
  }
  @media (max-width: 900px) {
    .med-grid {
      grid-template-columns: minmax(0, 1fr);
    }
  }
</style>
