<script lang="ts">
  import type { Series } from "$lib/types/ucop";
  import { fmtUSD } from "$lib/utils/chart";
  import ChartDataTable from "./ChartDataTable.svelte";

  let { series, caption }: { series: Series[]; caption: string } = $props();

  const years = $derived(
    [...new Set(series.flatMap((item) => item.values.map((point) => point.year)))].sort(
      (left, right) => right - left,
    ),
  );
  const rows = $derived(
    series.map((item) => {
      const amounts = new Map(item.values.map((point) => [point.year, point.value]));
      return {
        label: item.label,
        values: years.map((year) => {
          const amount = amounts.get(year);
          return amount === undefined ? undefined : fmtUSD(amount);
        }),
      };
    }),
  );
</script>

<ChartDataTable {caption} rowHeading="Series" columns={years.map(String)} {rows} />
