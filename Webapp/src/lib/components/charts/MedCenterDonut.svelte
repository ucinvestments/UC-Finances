<script lang="ts">
  import * as d3 from "d3";
  import { onMount, onDestroy, untrack } from "svelte";
  import { createTooltip, fmtUSD, fmtUSDShort, type Tooltip } from "$lib/utils/chart";
  import { getCampusTheme } from "$lib/utils/campusColors";

  interface Slice {
    label: string;
    value: number;
    themeKey?: string;
  }
  interface Props {
    slices: Slice[];
    height?: number;
    centerLabel?: string;
  }
  let { slices, height = 440, centerLabel = "" }: Props = $props();

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
    void slices;
    void height;
    void centerLabel;
    if (mounted && el) draw();
  });

  function draw() {
    if (!el || !tip) return;
    d3.select(el).selectAll("*").remove();
    const containerWidth = Math.max(280, el.clientWidth || 800);
    const w = containerWidth;
    const h = height;
    const r = Math.min(w, h) / 2 - 16;
    const innerR = r * 0.58;

    const total = slices.reduce((a, s) => a + s.value, 0);
    if (total <= 0) return;

    const svg = d3
      .select(el)
      .append("svg")
      .attr("viewBox", `0 0 ${w} ${h}`)
      .attr("preserveAspectRatio", "xMidYMid meet")
      .style("width", "100%")
      .style("height", "auto")
      .style("max-width", "100%");
    const g = svg.append("g").attr("transform", `translate(${w / 2},${h / 2})`);

    const pie = d3.pie<Slice>().value((d) => d.value).sort(null);
    const arcs = pie(slices);

    const arc = d3.arc<d3.PieArcDatum<Slice>>().innerRadius(innerR).outerRadius(r);
    const arcHover = d3
      .arc<d3.PieArcDatum<Slice>>()
      .innerRadius(innerR * 0.94)
      .outerRadius(r * 1.04);

    g.selectAll("path")
      .data(arcs)
      .enter()
      .append("path")
      .attr("d", arc)
      .attr("fill", (d) => {
        if (d.data.themeKey) return getCampusTheme(d.data.themeKey).primary;
        return "#1B4685";
      })
      .attr("stroke", "white")
      .attr("stroke-width", 2)
      .style("cursor", "pointer")
      .on("mouseenter", function (e, d) {
        const pct = ((d.data.value / total) * 100).toFixed(1);
        tip!.show(
          `<div style="font-weight:600;margin-bottom:4px;">${d.data.label}</div>` +
            `<div>${fmtUSD(d.data.value)}</div>` +
            `<div style="opacity:0.85;">(${pct}% of total)</div>`,
          e,
        );
        d3.select(this).transition().duration(150).attr("d", arcHover as any);
      })
      .on("mousemove", (e) => tip!.move(e))
      .on("mouseleave", function () {
        tip!.hide();
        d3.select(this).transition().duration(150).attr("d", arc as any);
      });

    // labels on slices large enough
    g.selectAll("text.slice-label")
      .data(arcs)
      .enter()
      .append("text")
      .attr("class", "slice-label")
      .attr("transform", (d) => `translate(${arc.centroid(d)})`)
      .attr("text-anchor", "middle")
      .attr("dy", "0.35em")
      .style("fill", "white")
      .style("font-size", "11px")
      .style("font-weight", "600")
      .style("pointer-events", "none")
      .text((d) => (d.endAngle - d.startAngle > 0.2 ? d.data.label : ""));

    // center labels
    g.append("text")
      .attr("text-anchor", "middle")
      .attr("dy", "-0.3em")
      .style("font-size", "13px")
      .style("fill", "#64748b")
      .text(centerLabel);
    g.append("text")
      .attr("text-anchor", "middle")
      .attr("dy", "1.1em")
      .style("font-size", "18px")
      .style("font-weight", "700")
      .style("fill", "#0f172a")
      .text(fmtUSDShort(total));
  }
</script>

<div class="donut" bind:this={el} style="min-height: {height}px;"></div>

<style>
  .donut {
    width: 100%;
  }
</style>
