<script lang="ts">
  import * as d3 from "d3";
  import { onMount, onDestroy, untrack } from "svelte";
  import type { TreeNode } from "$lib/types/ucop";
  import { createTooltip, fmtUSD, fmtUSDShort, type Tooltip } from "$lib/utils/chart";

  interface Props {
    tree: TreeNode;
    height?: number;
  }
  let { tree, height = 560 }: Props = $props();

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
    void tree;
    void height;
    if (mounted && el) draw();
  });

  function draw() {
    if (!el || !tip) return;
    d3.select(el).selectAll("*").remove();
    const containerWidth = el.clientWidth || 800;
    const w = Math.max(320, containerWidth);
    const h = height;

    const hierarchyRoot = d3
      .hierarchy<TreeNode>(tree)
      .sum((d) => (d.children && d.children.length > 0 ? 0 : d.value ?? 0))
      .sort((a, b) => (b.value ?? 0) - (a.value ?? 0));

    if ((hierarchyRoot.value ?? 0) === 0) return;

    const layout = d3
      .treemap<TreeNode>()
      .size([w, h])
      .paddingOuter(3)
      .paddingTop(20)
      .paddingInner(2)
      .round(true);
    const root = layout(hierarchyRoot);

    // Top-level palette (Berkeley-Blue family)
    const topKeys = (root.children ?? []).map((c) => c.data.name);
    const palette = [
      "#003262",
      "#3b7ea1",
      "#005581",
      "#2774AE",
      "#1B4685",
      "#0064A4",
      "#002856",
      "#052049",
    ];
    const topColor = d3
      .scaleOrdinal<string, string>()
      .domain(topKeys)
      .range(palette);

    const svg = d3
      .select(el)
      .append("svg")
      .attr("width", w)
      .attr("height", h)
      .style("font-family", "inherit");

    const leaves = root.leaves();
    const totalVal = root.value ?? 1;

    // draw top-level headers first (so leaves sit on top)
    (root.children ?? []).forEach((top) => {
      svg
        .append("rect")
        .attr("x", top.x0)
        .attr("y", top.y0)
        .attr("width", top.x1 - top.x0)
        .attr("height", top.y1 - top.y0)
        .attr("fill", topColor(top.data.name))
        .attr("fill-opacity", 0.15)
        .attr("stroke", topColor(top.data.name))
        .attr("stroke-opacity", 0.4)
        .attr("stroke-width", 1);

      if (top.y1 - top.y0 > 20 && top.x1 - top.x0 > 60) {
        svg
          .append("text")
          .attr("x", top.x0 + 6)
          .attr("y", top.y0 + 14)
          .text(truncate(top.data.name, Math.max(8, Math.floor((top.x1 - top.x0) / 7))))
          .style("font-size", "11px")
          .style("font-weight", "600")
          .style("fill", topColor(top.data.name))
          .style("pointer-events", "none");
      }
    });

    // draw leaves
    const cells = svg
      .selectAll(".leaf")
      .data(leaves)
      .enter()
      .append("g")
      .attr("class", "leaf")
      .attr("transform", (d) => `translate(${d.x0},${d.y0})`);

    cells
      .append("rect")
      .attr("width", (d) => d.x1 - d.x0)
      .attr("height", (d) => d.y1 - d.y0)
      .attr("fill", (d) => {
        const top = d.ancestors().find((a) => a.depth === 1);
        return top ? topColor(top.data.name) : "#64748b";
      })
      .attr("fill-opacity", (d) => 0.55 + Math.min(0.4, (d.depth - 1) * 0.1))
      .attr("stroke", "white")
      .attr("stroke-width", 1)
      .style("cursor", "pointer")
      .on("mouseenter", (e, d) => {
        const path = d
          .ancestors()
          .reverse()
          .slice(1)
          .map((n) => n.data.name)
          .join(" › ");
        const pct = ((d.value ?? 0) / totalVal) * 100;
        tip!.show(
          `<div style="font-weight:600;margin-bottom:4px;">${d.data.name}</div>` +
            `<div style="opacity:0.8;font-size:12px;margin-bottom:4px;">${path}</div>` +
            `<div>${fmtUSD(d.value ?? 0)}</div>` +
            `<div style="opacity:0.8;">(${pct.toFixed(2)}% of total)</div>`,
          e,
        );
        d3.select(e.currentTarget as SVGRectElement).attr("fill-opacity", 0.95);
      })
      .on("mousemove", (e) => tip!.move(e))
      .on("mouseleave", (e, d) => {
        tip!.hide();
        d3.select(e.currentTarget as SVGRectElement).attr(
          "fill-opacity",
          0.55 + Math.min(0.4, (d.depth - 1) * 0.1),
        );
      });

    // leaf labels
    cells.each(function (d) {
      const cw = d.x1 - d.x0;
      const ch = d.y1 - d.y0;
      if (cw < 44 || ch < 22) return;
      const chars = Math.max(6, Math.floor(cw / 6.5));
      const g = d3.select(this);
      g.append("text")
        .attr("x", 6)
        .attr("y", 14)
        .style("font-size", "11px")
        .style("font-weight", "600")
        .style("fill", "white")
        .style("pointer-events", "none")
        .text(truncate(d.data.name, chars));
      if (ch > 38) {
        g.append("text")
          .attr("x", 6)
          .attr("y", 28)
          .style("font-size", "10px")
          .style("fill", "rgba(255,255,255,0.88)")
          .style("pointer-events", "none")
          .text(fmtUSDShort(d.value ?? 0));
      }
    });
  }

  function truncate(s: string, n: number): string {
    return s.length > n ? s.slice(0, Math.max(0, n - 1)) + "…" : s;
  }
</script>

<div class="treemap" bind:this={el} style="min-height: {height}px;"></div>

<style>
  .treemap {
    width: 100%;
  }
</style>
