<script lang="ts">
  import type { PageData } from "./$types";
  import Container from "$lib/components/Container.svelte";
  import { fmtUSDShort } from "$lib/utils/chart";
  import { reportedSystemTotal } from "$lib/utils/retrendsMerge";
  import { summarizeAllocations } from "$lib/utils/ucopHierarchy";

  let { data }: { data: PageData } = $props();

  const allocation = $derived(summarizeAllocations(data.latestAlloc).reported.system_total);
  const revenue = $derived(reportedSystemTotal(data.latestRetrends, "revenue"));
  const expense = $derived(reportedSystemTotal(data.latestRetrends, "expense"));

  function formatTotal(value: number | undefined): string {
    return value === undefined ? "Unavailable" : fmtUSDShort(value);
  }
</script>

<svelte:head>
  <title>UC Finances Explorer</title>
  <meta
    name="description"
    content="Explore University of California revenues, expenses, and state General Fund allocations."
  />
</svelte:head>

<Container>
  <header class="introduction">
    <h1>UC Finances Explorer</h1>
    <p>
      Follow what the University of California brings in, what it spends, and how
      state General Fund support is allocated.
    </p>
  </header>

  <section class="overview" aria-label="Financial reports">
    <article>
      <div>
        <h2><a href="/revenue-expenses">Revenue and expense trends</a></h2>
        <p>Compare the UC system, individual campuses, and medical centers.</p>
      </div>
      <dl aria-label="System totals for fiscal year {data.latestRetrends.year_key}">
        <div>
          <dt>FY {data.latestRetrends.year_key} revenue</dt>
          <dd>{formatTotal(revenue)}</dd>
        </div>
        <div>
          <dt>FY {data.latestRetrends.year_key} expenses</dt>
          <dd>{formatTotal(expense)}</dd>
        </div>
      </dl>
    </article>

    <article>
      <div>
        <h2><a href="/fund-allocations">State General Fund allocations</a></h2>
        <p>See how state support is distributed by campus and budget category.</p>
      </div>
      <dl>
        <div>
          <dt>FY {data.latestAlloc.fiscal_year} system total</dt>
          <dd>{formatTotal(allocation)}</dd>
        </div>
      </dl>
    </article>
  </section>
  <p class="methodology"><a href="/about">Data sources and methodology</a></p>
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
    max-width: 58ch;
    color: var(--text-secondary);
  }

  .introduction p {
    font-size: clamp(1rem, 2vw, 1.25rem);
  }

  article {
    display: grid;
    grid-template-columns: minmax(0, 1.33fr) minmax(0, 1fr);
    align-items: center;
    gap: 1.5rem;
    padding-block: 1.5rem;
    border-block-start: 1px solid var(--border);
  }

  article:last-child {
    border-block-end: 1px solid var(--border);
  }

  h2 {
    font-family: "Space Grotesk", "Inter", sans-serif;
    font-size: clamp(1.25rem, 2.5vw, 1.75rem);
    line-height: 1.2;
    margin-block-end: 0.375rem;
  }

  a {
    color: var(--pri);
    text-decoration: underline;
    text-underline-offset: 0.18em;
  }

  a:hover {
    text-decoration-thickness: 0.14em;
  }

  dl {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 1rem;
    margin: 0;
  }

  dt {
    color: var(--text-secondary);
    font-size: var(--text-sm);
  }

  dd {
    color: var(--pri);
    font-size: clamp(1.125rem, 3vw, 1.625rem);
    font-weight: var(--font-bold);
    font-variant-numeric: tabular-nums;
  }

  .methodology {
    margin-block-start: 1.5rem;
  }

  @media (max-width: 700px) {
    article {
      grid-template-columns: minmax(0, 1fr);
      gap: 0.875rem;
    }

    dl {
      grid-template-columns: repeat(2, minmax(0, max-content));
      column-gap: 1.5rem;
    }
  }
</style>
