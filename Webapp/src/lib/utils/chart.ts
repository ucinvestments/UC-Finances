import * as d3 from "d3";

export const RETRENDS_UNIT = 1000;

export const fmtUSD = (n: number): string =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(n);

export const fmtUSDShort = (n: number): string => {
  const abs = Math.abs(n);
  const sign = n < 0 ? "-" : "";
  if (abs >= 1e9) return `${sign}$${(abs / 1e9).toFixed(1)}B`;
  if (abs >= 1e6) return `${sign}$${(abs / 1e6).toFixed(1)}M`;
  if (abs >= 1e3) return `${sign}$${(abs / 1e3).toFixed(0)}K`;
  return `${sign}$${abs.toFixed(0)}`;
};

export const fmtPercent = (n: number, digits = 1): string =>
  `${(n * 100).toFixed(digits)}%`;

export const fmtNumber = (n: number): string =>
  new Intl.NumberFormat("en-US").format(n);

export interface Tooltip {
  show(html: string, evt: MouseEvent): void;
  move(evt: MouseEvent): void;
  hide(): void;
  destroy(): void;
}

export function createTooltip(): Tooltip {
  if (typeof document === "undefined") {
    const noop = () => {};
    return { show: noop, move: noop, hide: noop, destroy: noop };
  }
  const node = d3
    .select("body")
    .append("div")
    .attr("class", "d3-tooltip")
    .style("position", "absolute")
    .style("pointer-events", "none")
    .style("background", "rgba(15,23,42,0.94)")
    .style("color", "#fff")
    .style("padding", "10px 12px")
    .style("border-radius", "8px")
    .style("font-size", "13px")
    .style("line-height", "1.45")
    .style("box-shadow", "0 4px 14px rgba(0,0,0,0.28)")
    .style("opacity", "0")
    .style("transition", "opacity 0.15s")
    .style("z-index", "500")
    .style("max-width", "320px");
  return {
    show(html, evt) {
      node
        .html(html)
        .style("opacity", "1")
        .style("left", evt.pageX + 12 + "px")
        .style("top", evt.pageY - 8 + "px");
    },
    move(evt) {
      node.style("left", evt.pageX + 12 + "px").style("top", evt.pageY - 8 + "px");
    },
    hide() {
      node.style("opacity", "0");
    },
    destroy() {
      node.remove();
    },
  };
}

export interface Margin {
  top: number;
  right: number;
  bottom: number;
  left: number;
}

export const DEFAULT_MARGIN: Margin = { top: 32, right: 32, bottom: 48, left: 72 };

// Berkeley-adjacent categorical palette (from UC-wages)
export const CHART_COLORS = [
  "#003262",
  "#fdb515",
  "#3b7ea1",
  "#d9661f",
  "#00b0da",
  "#859438",
  "#c4820e",
  "#b9d3b6",
  "#ee1f60",
  "#584f29",
  "#46535e",
  "#cfdd45",
];

// Rough CPI multipliers to base year 2025 (US all-items CPI-U, annual average)
// Values only used when the user opts into inflation adjustment.
export const CPI_TO_2025: Record<number, number> = {
  2010: 1.47,
  2011: 1.42,
  2012: 1.39,
  2013: 1.37,
  2014: 1.35,
  2015: 1.35,
  2016: 1.33,
  2017: 1.3,
  2018: 1.27,
  2019: 1.25,
  2020: 1.23,
  2021: 1.18,
  2022: 1.09,
  2023: 1.05,
  2024: 1.03,
  2025: 1.0,
};

export function inflationAdjust(amount: number, fromYear: number): number {
  const m = CPI_TO_2025[fromYear] ?? 1;
  return amount * m;
}
