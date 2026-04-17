<script lang="ts">
  import Hero from "$lib/components/Hero.svelte";
  import Container from "$lib/components/Container.svelte";
  import Section from "$lib/components/Section.svelte";
  import PageLayout from "$lib/components/PageLayout.svelte";
  import Card from "$lib/components/Card.svelte";
</script>

<svelte:head>
  <title>About · UC Finances</title>
</svelte:head>

<PageLayout>
  <Hero
    title="About UC Finances"
    subtitle="What's in this explorer, and where the data comes from."
  />

  <Container>
    <Section title="Data sources" icon="mdi:database">
      <div class="grid">
        <Card>
          <h3>State General Fund Allocations</h3>
          <p>
            Ten fiscal years (2016-17 through 2025-26) of state General Fund
            allocations to each UC campus and system-level programme. Parsed from
            the official UCOP summary PDFs published at
            <a
              href="https://www.ucop.edu/operating-budget/budgets-and-reports/other-resources/"
              target="_blank"
              rel="noopener noreferrer"
              >ucop.edu/operating-budget</a
            >.
          </p>
        </Card>
        <Card>
          <h3>Revenue and Expense Trends</h3>
          <p>
            Eleven annual Revenue and Expense Trends reports (2015 through 2025)
            from UC's reporting-transparency portal. Each report covers a rolling
            five-year window; overlapping years are reconciled by keeping the
            value from the most recent report, so restatements carry through.
            Source:
            <a
              href="https://reportingtransparency.universityofcalifornia.edu/documents/finreports/retrends/"
              target="_blank"
              rel="noopener noreferrer"
              >reportingtransparency.universityofcalifornia.edu</a
            >.
          </p>
        </Card>
      </div>
    </Section>

    <Section title="Scope and caveats" icon="mdi:information">
      <Card>
        <p>
          UCOP reports dollar figures in thousands for the Revenue and Expense
          Trends series and in whole dollars for the General Fund Allocations;
          both are normalised to actual dollars before display. Values labelled
          "Total" or "Subtotal" in the source PDFs are excluded from the treemap
          sums to avoid double-counting.
        </p>
        <p style="margin-top: 1rem;">
          This site is not affiliated with the University of California.
        </p>
      </Card>
    </Section>
  </Container>
</PageLayout>

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: 1.5rem;
  }
  h3 {
    margin: 0 0 0.5rem;
    font-size: 1.125rem;
    color: var(--pri);
    font-weight: 600;
  }
  p {
    color: var(--text-secondary);
    line-height: 1.6;
    margin: 0;
  }
  a {
    color: var(--founder);
    text-decoration: underline;
  }
  a:hover {
    color: var(--pri);
  }
</style>
