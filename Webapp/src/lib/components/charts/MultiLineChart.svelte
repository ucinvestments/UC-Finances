<script lang="ts">
  import * as d3 from "d3";
  import { onMount, onDestroy, untrack } from "svelte";
  import type { Series } from "$lib/types/ucop";
  import {
    createTooltip,
    fmtUSDShort,
    DEFAULT_MARGIN,
    CHART_COLORS,
    type Tooltip,
  } from "$lib/utils/chart";

  interface Props {
    series: Series[];
    height?: number;
    yLabel?: string;
    yFormat?: (n: number) => string;
    logarithmic?: boolean;
    animate?: boolean;
  }
  let {
    series,
    height = 420,
    yLabel = "",
    yFormat = fmtUSDShort,
    logarithmic = false,
    animate = true,
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
        if (mounted) untrack(() => draw(false));
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
    void logarithmic;
    void height;
    if (mounted && el) draw(animate);
  });

  function draw(doAnimate: boolean) {
    if (!el || !tip) return;
    d3.select(el).selectAll("*").remove();
    const containerWidth = el.clientWidth || 800;
    const m = DEFAULT_MARGIN;
    const w = Math.max(300, containerWidth) - m.left - m.right;
    const h = height - m.top - m.bottom;

    const svg = d3
      .select(el)
      .append("svg")
      .attr("width", containerWidth)
      .attr("height", height);
    const g = svg.append("g").attr("transform", `translate(${m.left},${m.top})`);

    const allPts = series.flatMap((s) => s.values);
    if (allPts.length === 0) return;
    const years = [...new Set(allPts.map((p) => p.year))].sort((a, b) => a - b);
    const x = d3
      .scaleLinear()
      .domain(d3.extent(years) as [number, number])
      .range([0, w]);
    const [lo, hi] = d3.extent(allPts, (p) => p.value) as [number, number];
    const y = logarithmic
      ? d3
          .scaleLog()
          .domain([Math.max(1, lo), Math.max(10, hi)])
          .range([h, 0])
          .nice()
      : d3
          .scaleLinear()
          .domain([Math.min(0, lo), hi])
          .range([h, 0])
          .nice();

    const color = d3
      .scaleOrdinal<string, string>()
      .domain(series.map((s) => s.id))
      .range(CHART_COLORS);

    const line = d3
      .line<{ year: number; value: number }>()
      .defined((d) => Number.isFinite(d.value))
      .x((d) => x(d.year))
      .y((d) => y(logarithmic ? Math.max(1, d.value) : d.value))
      .curve(d3.curveCatmullRom.alpha(0.5));

    // grid + x axis
    g.append("g")
      .attr("class", "x-axis")
      .attr("transform", `translate(0,${h})`)
      .call(d3.axisBottom(x).tickFormat(d3.format("d")).tickSize(-h).tickPadding(8))
      .call((sel) =>
        sel
          .selectAll(".tick line")
          .attr("stroke", "#e5e7eb")
          .attr("stroke-dasharray", "2,2"),
      )
      .call((sel) => sel.select(".domain").attr("stroke", "#cbd5e1"));
    g.append("g")
      .attr("class", "y-axis")
      .call(
        d3
          .axisLeft(y)
          .tickFormat(yFormat as any)
          .tickSize(-w)
          .tickPadding(8),
      )
      .call((sel) =>
        sel
          .selectAll(".tick line")
          .attr("stroke", "#e5e7eb")
          .attr("stroke-dasharray", "2,2"),
      )
      .call((sel) => sel.select(".domain").attr("stroke", "#cbd5e1"));
    g.selectAll(".x-axis text, .y-axis text")
      .style("fill", "#64748b")
      .style("font-size", "12px");

    if (yLabel) {
      g.append("text")
        .attr("transform", "rotate(-90)")
        .attr("x", -h / 2)
        .attr("y", -m.left + 18)
        .attr("text-anchor", "middle")
        .style("font-size", "12px")
        .style("fill", "#475569")
        .text(yLabel);
    }

    series.forEach((s, i) => {
      const vals = s.values.slice().sort((a, b) => a.year - b.year);
      const stroke = s.color ?? color(s.id);
      const path = g
        .append("path")
        .datum(vals)
        .attr("fill", "none")
        .attr("stroke", stroke)
        .attr("stroke-width", 2.5)
        .attr("stroke-linejoin", "round")
        .attr("stroke-linecap", "round")
        .attr("d", line);

      if (doAnimate) {
        const node = path.node() as SVGPathElement | null;
        const len = node ? node.getTotalLength() : 0;
        if (len > 0) {
          path
            .attr("stroke-dasharray", `${len} ${len}`)
            .attr("stroke-dashoffset", len)
            .transition()
            .delay(i * 100)
            .duration(1100)
            .ease(d3.easeCubicInOut)
            .attr("stroke-dashoffset", 0)
            .on("end", function () {
              d3.select(this).attr("stroke-dasharray", null);
            });
        }
      }

      g.append("g")
        .attr("class", "dots")
        .selectAll("circle")
        .data(vals.filter((v) => Number.isFinite(v.value)))
        .enter()
        .append("circle")
        .attr("cx", (d) => x(d.year))
        .attr("cy", (d) => y(logarithmic ? Math.max(1, d.value) : d.value))
        .attr("r", 3.5)
        .attr("fill", stroke)
        .attr("stroke", "white")
        .attr("stroke-width", 1.5)
        .style("cursor", "pointer")
        .on("mouseenter", (e, d) => {
          tip!.show(
            `<strong>${s.label}</strong><br>${d.year}: ${yFormat(d.value)}`,
            e,
          );
          d3.select(e.currentTarget as SVGCircleElement)
            .transition()
            .duration(120)
            .attr("r", 6);
        })
        .on("mousemove", (e) => tip!.move(e))
        .on("mouseleave", (e) => {
          tip!.hide();
          d3.select(e.currentTarget as SVGCircleElement)
            .transition()
            .duration(120)
            .attr("r", 3.5);
        });
    });
  }
</script>

<div class="ml-chart" bind:this={el} style="min-height: {height}px;"></div>

<style>
  .ml-chart {
    width: 100%;
  }
</style>
