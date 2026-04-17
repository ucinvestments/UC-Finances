<script lang="ts">
  import * as d3 from "d3";
  import { onMount, onDestroy, untrack } from "svelte";
  import type { Series } from "$lib/types/ucop";
  import {
    createTooltip,
    fmtUSDShort,
    fmtUSD,
    DEFAULT_MARGIN,
    CHART_COLORS,
    type Tooltip,
  } from "$lib/utils/chart";

  interface Props {
    series: Series[];
    height?: number;
    stackOffset?: "none" | "expand";
    yFormat?: (n: number) => string;
  }
  let {
    series,
    height = 420,
    stackOffset = "none",
    yFormat = fmtUSDShort,
  }: Props = $props();

  let el: HTMLDivElement | undefined = $state();
  let tip: Tooltip | undefined;
  let mounted = $state(false);
  let resizeObs: ResizeObserver | undefined;

  onMount(() => {
    mounted = true;
    tip = createTooltip();
    if (el && typeof ResizeObserver !== "undefined") {
      resizeObs = new ResizeObserver(() => {
        if (mounted) untrack(() => draw());
      });
      resizeObs.observe(el);
    }
  });
  onDestroy(() => {
    tip?.destroy();
    resizeObs?.disconnect();
  });

  $effect(() => {
    void series;
    void stackOffset;
    void height;
    if (mounted && el) draw();
  });

  interface WideRow {
    year: number;
    [key: string]: number;
  }

  function draw() {
    if (!el || !tip) return;
    d3.select(el).selectAll("*").remove();
    const containerWidth = Math.max(280, el.clientWidth || 800);
    const m = { ...DEFAULT_MARGIN, left: el.clientWidth < 420 ? 56 : DEFAULT_MARGIN.left };
    const w = containerWidth - m.left - m.right;
    const h = height - m.top - m.bottom;

    if (series.length === 0) return;

    // Build wide-format rows: one per year, columns = series ids
    const yearSet = new Set<number>();
    series.forEach((s) => s.values.forEach((v) => yearSet.add(v.year)));
    const years = [...yearSet].sort((a, b) => a - b);

    const byId = new Map(series.map((s) => [s.id, new Map(s.values.map((v) => [v.year, v.value]))]));
    const wide: WideRow[] = years.map((year) => {
      const row: WideRow = { year };
      for (const s of series) {
        row[s.id] = byId.get(s.id)?.get(year) ?? 0;
      }
      return row;
    });

    const keys = series.map((s) => s.id);
    const stackGen = d3
      .stack<WideRow>()
      .keys(keys)
      .offset(stackOffset === "expand" ? d3.stackOffsetExpand : d3.stackOffsetNone);
    const stacked = stackGen(wide);

    const svg = d3
      .select(el)
      .append("svg")
      .attr("viewBox", `0 0 ${containerWidth} ${height}`)
      .attr("preserveAspectRatio", "xMidYMid meet")
      .style("width", "100%")
      .style("height", "auto")
      .style("max-width", "100%");
    const g = svg.append("g").attr("transform", `translate(${m.left},${m.top})`);

    const x = d3
      .scaleLinear()
      .domain(d3.extent(years) as [number, number])
      .range([0, w]);

    const yMax = d3.max(stacked, (layer) => d3.max(layer, (d) => d[1])) ?? 0;
    const y = d3
      .scaleLinear()
      .domain([0, yMax])
      .range([h, 0])
      .nice();

    // Resolve colors: prefer per-series `color`, fall back to palette
    const fallback = d3.scaleOrdinal<string, string>().domain(keys).range(CHART_COLORS);
    const colorFor = (id: string): string => {
      const s = series.find((ss) => ss.id === id);
      return s?.color ?? fallback(id);
    };
    const color = (id: string) => colorFor(id);

    const area = d3
      .area<[number, number] & { data: WideRow }>()
      .x((d) => x(d.data.year))
      .y0((d) => y(d[0]))
      .y1((d) => y(d[1]))
      .curve(d3.curveMonotoneX);

    // axes
    g.append("g")
      .attr("class", "x-axis")
      .attr("transform", `translate(0,${h})`)
      .call(d3.axisBottom(x).tickFormat(d3.format("d")).tickSize(-h).tickPadding(8))
      .call((sel) =>
        sel
          .selectAll(".tick line")
          .attr("stroke", "#e5e7eb")
          .attr("stroke-dasharray", "2,2"),
      );
    const yAxisFmt =
      stackOffset === "expand" ? (d: any) => `${Math.round(+d * 100)}%` : yFormat;
    g.append("g")
      .attr("class", "y-axis")
      .call(
        d3
          .axisLeft(y)
          .tickFormat(yAxisFmt as any)
          .tickSize(-w)
          .tickPadding(8),
      )
      .call((sel) =>
        sel
          .selectAll(".tick line")
          .attr("stroke", "#e5e7eb")
          .attr("stroke-dasharray", "2,2"),
      );
    g.selectAll(".x-axis text, .y-axis text")
      .style("fill", "#64748b")
      .style("font-size", "12px");

    // stacked areas
    const layer = g
      .selectAll(".layer")
      .data(stacked)
      .enter()
      .append("g")
      .attr("class", "layer");

    layer
      .append("path")
      .attr("fill", (d) => color(d.key as string))
      .attr("stroke", "white")
      .attr("stroke-width", 0.5)
      .attr("opacity", 0)
      .attr("d", area as any)
      .transition()
      .delay((_, i) => i * 40)
      .duration(700)
      .attr("opacity", 0.92);

    // invisible full-height hover strips per year for tooltip
    const bandW = w / Math.max(1, years.length - 1);
    const hoverG = g.append("g").attr("class", "hover-layer");
    years.forEach((year, i) => {
      const xPos = x(year);
      hoverG
        .append("rect")
        .attr("x", xPos - bandW / 2)
        .attr("y", 0)
        .attr("width", bandW)
        .attr("height", h)
        .attr("fill", "transparent")
        .on("mouseenter", (e) => {
          const row = wide[i];
          const total = keys.reduce((a, k) => a + (row[k] ?? 0), 0);
          const parts = series
            .slice()
            .sort((a, b) => (row[b.id] ?? 0) - (row[a.id] ?? 0))
            .slice(0, 12)
            .map((s) => {
              const v = row[s.id] ?? 0;
              const pct = total > 0 ? ((v / total) * 100).toFixed(1) : "0";
              const sw = `<span style="display:inline-block;width:10px;height:10px;background:${color(s.id)};margin-right:6px;border-radius:2px;"></span>`;
              return `<div>${sw}${s.label}: ${fmtUSD(v)} (${pct}%)</div>`;
            })
            .join("");
          tip!.show(
            `<div style="font-weight:600;margin-bottom:6px;">${year} — Total ${fmtUSD(total)}</div>${parts}`,
            e,
          );
        })
        .on("mousemove", (e) => tip!.move(e))
        .on("mouseleave", () => tip!.hide());
    });
  }
</script>

<div class="sa-chart" bind:this={el} style="min-height: {height}px;"></div>

<style>
  .sa-chart {
    width: 100%;
  }
</style>
