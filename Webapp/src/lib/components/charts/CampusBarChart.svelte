<script lang="ts">
  import * as d3 from "d3";
  import { onMount, onDestroy, untrack } from "svelte";
  import { CAMPUS_KEYS, type CampusKey } from "$lib/types/ucop";
  import {
    createTooltip,
    fmtUSD,
    fmtUSDShort,
    DEFAULT_MARGIN,
    CHART_COLORS,
    type Tooltip,
  } from "$lib/utils/chart";
  import { campusDisplayName, getCampusTheme } from "$lib/utils/campusColors";

  interface Segment {
    id: string;
    label: string;
    color?: string;
    values: Partial<Record<CampusKey, number>>;
  }

  interface Props {
    segments: Segment[];
    height?: number;
    sortBy?: "desc" | "asc" | "campus";
    title?: string;
  }
  let { segments, height = 440, sortBy = "desc" }: Props = $props();

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
    void segments;
    void sortBy;
    void height;
    if (mounted && el) draw();
  });

  function draw() {
    if (!el || !tip) return;
    d3.select(el).selectAll("*").remove();
    const containerWidth = el.clientWidth || 800;
    const m = { ...DEFAULT_MARGIN, bottom: 72 };
    const w = Math.max(300, containerWidth) - m.left - m.right;
    const h = height - m.top - m.bottom;

    const totalsByCampus = new Map<CampusKey, number>();
    CAMPUS_KEYS.forEach((c) => {
      let t = 0;
      for (const seg of segments) t += seg.values[c] ?? 0;
      totalsByCampus.set(c, t);
    });

    let orderedCampuses = [...CAMPUS_KEYS];
    if (sortBy === "desc") {
      orderedCampuses.sort(
        (a, b) => (totalsByCampus.get(b) ?? 0) - (totalsByCampus.get(a) ?? 0),
      );
    } else if (sortBy === "asc") {
      orderedCampuses.sort(
        (a, b) => (totalsByCampus.get(a) ?? 0) - (totalsByCampus.get(b) ?? 0),
      );
    }

    const svg = d3
      .select(el)
      .append("svg")
      .attr("width", containerWidth)
      .attr("height", height);
    const g = svg.append("g").attr("transform", `translate(${m.left},${m.top})`);

    const x = d3.scaleBand().domain(orderedCampuses).range([0, w]).padding(0.2);
    const yMax = d3.max(orderedCampuses, (c) => totalsByCampus.get(c) ?? 0) ?? 0;
    const y = d3.scaleLinear().domain([0, yMax]).range([h, 0]).nice();

    const color = d3
      .scaleOrdinal<string, string>()
      .domain(segments.map((s) => s.id))
      .range(CHART_COLORS);

    // grid
    g.append("g")
      .attr("class", "y-axis")
      .call(
        d3
          .axisLeft(y)
          .tickFormat(fmtUSDShort as any)
          .tickSize(-w)
          .tickPadding(8),
      )
      .call((sel) =>
        sel
          .selectAll(".tick line")
          .attr("stroke", "#e5e7eb")
          .attr("stroke-dasharray", "2,2"),
      )
      .call((sel) => sel.select(".domain").remove());
    g.append("g")
      .attr("class", "x-axis")
      .attr("transform", `translate(0,${h})`)
      .call(d3.axisBottom(x).tickFormat((c) => campusDisplayName(String(c))))
      .call((sel) =>
        sel
          .selectAll("text")
          .attr("transform", "rotate(-30)")
          .style("text-anchor", "end")
          .attr("dy", "0.35em")
          .attr("dx", "-0.5em"),
      )
      .call((sel) => sel.select(".domain").attr("stroke", "#cbd5e1"));
    g.selectAll(".x-axis text, .y-axis text")
      .style("fill", "#64748b")
      .style("font-size", "12px");

    orderedCampuses.forEach((campus) => {
      let cumulative = 0;
      segments.forEach((seg, i) => {
        const v = seg.values[campus] ?? 0;
        if (v <= 0) return;
        const y0 = y(cumulative);
        const y1 = y(cumulative + v);
        const fill = seg.color ?? color(seg.id);
        const rect = g
          .append("rect")
          .attr("x", x(campus)!)
          .attr("width", x.bandwidth())
          .attr("y", h)
          .attr("height", 0)
          .attr("fill", fill)
          .attr("stroke", "white")
          .attr("stroke-width", 0.5)
          .style("cursor", "pointer");
        rect
          .on("mouseenter", (e) => {
            const total = totalsByCampus.get(campus) ?? 0;
            const pct = total > 0 ? ((v / total) * 100).toFixed(1) : "0";
            const theme = getCampusTheme(campus);
            tip!.show(
              `<div style="font-weight:600;margin-bottom:4px;">${theme.name}</div>` +
                `<div>${seg.label}: ${fmtUSD(v)}</div>` +
                (segments.length > 1
                  ? `<div style="opacity:0.8;">(${pct}% of campus total ${fmtUSDShort(total)})</div>`
                  : ""),
              e,
            );
            d3.select(e.currentTarget as SVGRectElement).attr("opacity", 0.8);
          })
          .on("mousemove", (e) => tip!.move(e))
          .on("mouseleave", (e) => {
            tip!.hide();
            d3.select(e.currentTarget as SVGRectElement).attr("opacity", 1);
          })
          .transition()
          .delay(i * 80)
          .duration(650)
          .ease(d3.easeCubicOut)
          .attr("y", y1)
          .attr("height", y0 - y1);
        cumulative += v;
      });
    });
  }
</script>

<div class="bar-chart" bind:this={el} style="min-height: {height}px;"></div>

<style>
  .bar-chart {
    width: 100%;
  }
</style>
