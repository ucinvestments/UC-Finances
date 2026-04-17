import type { RetrendsReport, RetrendsRow, Series } from "$lib/types/ucop";
import { RETRENDS_UNIT } from "./chart";

export interface MergedSeries {
  page_title: string;
  section: string;
  /** Section with trailing footnote digits stripped (e.g. "REVENUES4" -> "REVENUES"). */
  section_norm: string;
  label: string;
  /** Label with trailing footnote digits stripped (e.g. "Interest Expense2" -> "Interest Expense"). */
  label_norm: string;
  indent: number;
  table_index: number;
  /** year -> value in thousands (raw retrends unit) */
  amounts: Record<number, number>;
  /** year -> report year_key that supplied this value (provenance) */
  sources: Record<number, string>;
}

export interface MergedRetrends {
  years: number[];
  series: MergedSeries[];
}

function stripFootnote(s: string): string {
  return s.replace(/\d+$/, "").trim();
}

function keyOf(r: RetrendsRow): string {
  return `${r.page_title}\u0001${stripFootnote(r.section)}\u0001${stripFootnote(r.label)}`;
}

export function mergeRetrends(reports: RetrendsReport[]): MergedRetrends {
  const sorted = [...reports].sort((a, b) => a.year_key.localeCompare(b.year_key));
  const bucket = new Map<string, MergedSeries>();
  const yearSet = new Set<number>();

  for (const rep of sorted) {
    for (const row of rep.rows) {
      const k = keyOf(row);
      let entry = bucket.get(k);
      if (!entry) {
        entry = {
          page_title: row.page_title,
          section: row.section,
          section_norm: stripFootnote(row.section),
          label: row.label,
          label_norm: stripFootnote(row.label),
          indent: row.indent,
          table_index: row.table_index,
          amounts: {},
          sources: {},
        };
        bucket.set(k, entry);
      }
      for (const [y, v] of Object.entries(row.amounts)) {
        if (typeof v !== "number") continue;
        const yr = Number(y);
        if (!Number.isFinite(yr)) continue;
        yearSet.add(yr);
        entry.amounts[yr] = v;
        entry.sources[yr] = rep.year_key;
      }
    }
  }

  return {
    years: [...yearSet].sort((a, b) => a - b),
    series: [...bucket.values()],
  };
}

export type RevExpKind = "revenue" | "expense";

function sectionMatchesKind(sectionNorm: string, kind: RevExpKind): boolean {
  if (kind === "revenue") return sectionNorm === "REVENUES";
  return sectionNorm === "EXPENDITURES" || sectionNorm === "EXPENSES";
}

/** Leaf (non-total) rows only — total rows have indent > 60. */
function isLeaf(s: MergedSeries): boolean {
  return s.indent < 60 && !/^Total\b/i.test(s.label_norm);
}

/** System-wide revenues by source OR expenses by function. */
export function systemRevExp(m: MergedRetrends, kind: RevExpKind): MergedSeries[] {
  return m.series.filter(
    (s) =>
      s.page_title.includes("Revenues by Source and Expenses by Function") &&
      sectionMatchesKind(s.section_norm, kind) &&
      isLeaf(s),
  );
}

/** Per-campus totals — each campus as a series across all years. */
export function perCampusRevExp(m: MergedRetrends, kind: RevExpKind): MergedSeries[] {
  return m.series.filter(
    (s) =>
      s.page_title.includes("Revenues and Expenses by Campus") &&
      sectionMatchesKind(s.section_norm, kind) &&
      isLeaf(s),
  );
}

/** Medical-center per-center totals. */
export function medCenterRevExp(m: MergedRetrends, kind: RevExpKind): MergedSeries[] {
  return m.series.filter(
    (s) =>
      s.page_title.includes("Revenues and Expenses by Medical Center") &&
      sectionMatchesKind(s.section_norm, kind) &&
      isLeaf(s),
  );
}

/** Convert MergedSeries[] to Series[] (scaled to actual dollars). */
export function toSeries(
  merged: MergedSeries[],
  idFn: (s: MergedSeries) => string = (s) => s.label_norm,
  labelFn: (s: MergedSeries) => string = (s) => s.label_norm,
): Series[] {
  return merged.map((s) => ({
    id: idFn(s),
    label: labelFn(s),
    values: Object.entries(s.amounts)
      .map(([y, v]) => ({ year: Number(y), value: v * RETRENDS_UNIT }))
      .sort((a, b) => a.year - b.year),
  }));
}
