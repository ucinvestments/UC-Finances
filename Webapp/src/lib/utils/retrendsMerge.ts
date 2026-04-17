import type { RetrendsReport, RetrendsRow, Series } from "$lib/types/ucop";
import { RETRENDS_UNIT } from "./chart";

export interface MergedSeries {
  page_title: string;
  /** Page title with the "— For Fiscal Years XXXX through YYYY" suffix stripped,
   *  so the same logical table merges across reports. */
  page_title_norm: string;
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

/** Strip trailing footnote markers: digits, commas, whitespace, asterisks. */
function stripFootnote(s: string): string {
  return s.replace(/[\s\d,*]+$/, "").trim();
}

/** Normalize label text for fuzzy matching against canonical category names. */
function normalizeLabel(s: string): string {
  return stripFootnote(s).toLowerCase().replace(/\s+/g, " ").trim();
}

/** Collapse page-title variants across reports so the same logical table
 *  merges. Removes:
 *   - year-range suffixes ("— For Fiscal Years XXXX through YYYY")
 *   - PDF page-number prefixes/infixes ("24 University of California …", "— 20 —")
 *   - "% changed …" suffix
 *   - Plural/singular drift ("Revenues by Sources" -> "Revenues by Source")
 *   - Dash variants (hyphen vs em-dash)
 */
function normalizePageTitle(s: string): string {
  return s
    .replace(/\s*—\s*For Fiscal Years\s+\d{4}\s+through\s+\d{4}.*$/i, "")
    .replace(/\s+—\s+%\s+changed.*$/i, "")
    .replace(/\s+\d{4}\s+through\s+\d{4}.*$/i, "")
    .replace(/^\s*\d+\s+/, "")
    .replace(/\s+—\s+\d+(?=\s|$)/g, "")
    .replace(/\bRevenues by Sources\b/gi, "Revenues by Source")
    .replace(/[\u2010\u2011\u2012\u2013\u2014]/g, "—")
    .replace(/\s*-\s*/g, " — ")
    .replace(/\s+/g, " ")
    .trim();
}

function keyOf(r: RetrendsRow): string {
  return `${normalizePageTitle(r.page_title)}\u0001${stripFootnote(r.section)}\u0001${stripFootnote(r.label)}`;
}

export function mergeRetrends(reports: RetrendsReport[]): MergedRetrends {
  // Report-year proximity wins: for a given (key, year), the report whose
  // year_key is closest to that year provides the most authoritative value.
  // Guards against PDF scraper errors in the 2020/2021 reports where older
  // years were mis-extracted as single-digit values.
  const sorted = [...reports].sort((a, b) => a.year_key.localeCompare(b.year_key));
  const bucket = new Map<string, MergedSeries>();
  const yearSet = new Set<number>();
  const distance = new Map<string, Record<number, number>>();

  for (const rep of sorted) {
    const repYear = Number(rep.year_key);
    for (const row of rep.rows) {
      const k = keyOf(row);
      let entry = bucket.get(k);
      if (!entry) {
        entry = {
          page_title: row.page_title,
          page_title_norm: normalizePageTitle(row.page_title),
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
        distance.set(k, {});
      }
      const dmap = distance.get(k)!;
      // Suspicious values are those < $1M (the smallest legitimate category
      // on any retrends page is a few million). PDF-scraper mis-extractions
      // typically land at single-digit or tens-of-dollars.
      const SUSPICIOUS = 1000; // thousands → $1M
      for (const [y, v] of Object.entries(row.amounts)) {
        if (typeof v !== "number" || !Number.isFinite(v)) continue;
        const yr = Number(y);
        if (!Number.isFinite(yr)) continue;
        yearSet.add(yr);
        const d = Math.abs(repYear - yr);
        const prevD = dmap[yr] ?? Infinity;
        const prevV = entry.amounts[yr];
        const prevIsSuspect =
          prevV !== undefined && Math.abs(prevV) < SUSPICIOUS;
        const newIsSuspect = Math.abs(v) < SUSPICIOUS;

        // Rule: strictly closer report wins; on tie, prefer larger magnitude.
        // BUT a healthy value always beats a scraper-garbage suspect value,
        // even if it comes from a farther report.
        const closer = d < prevD;
        const tieWins =
          d === prevD && (prevV === undefined || Math.abs(v) > Math.abs(prevV));
        const rescuesSuspect = prevIsSuspect && !newIsSuspect;
        const keepHealthy = !prevIsSuspect && newIsSuspect && prevV !== undefined;

        if (keepHealthy) continue;
        if (closer || tieWins || rescuesSuspect) {
          entry.amounts[yr] = v;
          entry.sources[yr] = rep.year_key;
          dmap[yr] = d;
        }
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

/** Canonical leaf categories on the system-wide "Revenues by Source / Expenses by Function" page. */
export const REVENUE_CATEGORIES = [
  "Student Tuition and Fees",
  "Grants and Contracts",
  "Medical Centers",
  "Educational Activities",
  "Auxiliary Enterprises",
  "Department of Energy Laboratories",
  "State Educational Appropriations",
  "Direct Government Grants",
  "State Hospital Fee Grants",
  "Private Gifts",
  "Investment Income",
  "Other Revenues",
] as const;

export const EXPENSE_CATEGORIES = [
  "Instruction",
  "Research",
  "Public Service",
  "Academic Support",
  "Student Services",
  "Institutional Support",
  "Operation and Maintenance of Plant",
  "Student Financial Aid",
  "Medical Centers",
  "Auxiliary Enterprises",
  "Depreciation and Amortization",
  "Interest Expense",
  "Department of Energy Laboratories",
  "Other Expenditures",
] as const;

const CAMPUS_LABELS_SET = new Set([
  "berkeley",
  "davis",
  "irvine",
  "los angeles",
  "merced",
  "riverside",
  "san diego",
  "san francisco",
  "santa barbara",
  "santa cruz",
]);

function matchesCanonical(labelNorm: string, canonicals: readonly string[]): string | null {
  const needle = normalizeLabel(labelNorm);
  for (const c of canonicals) {
    if (normalizeLabel(c) === needle) return c;
  }
  return null;
}

/** Keep one series per label (canonical), merging amounts across dupes. */
function dedupeByLabel(series: MergedSeries[]): MergedSeries[] {
  const byKey = new Map<string, MergedSeries>();
  for (const s of series) {
    const existing = byKey.get(s.label_norm);
    if (!existing) {
      byKey.set(s.label_norm, { ...s, amounts: { ...s.amounts }, sources: { ...s.sources } });
      continue;
    }
    // Merge: prefer non-zero values, keep later year_key as source.
    for (const [y, v] of Object.entries(s.amounts)) {
      const yr = Number(y);
      const prev = existing.amounts[yr];
      if (prev === undefined || (v !== 0 && (prev === 0 || !Number.isFinite(prev)))) {
        existing.amounts[yr] = v;
        existing.sources[yr] = s.sources[yr];
      }
    }
  }
  return [...byKey.values()];
}

/** System-wide revenues by source OR expenses by function — canonical category whitelist. */
export function systemRevExp(m: MergedRetrends, kind: RevExpKind): MergedSeries[] {
  const canonicals = kind === "revenue" ? REVENUE_CATEGORIES : EXPENSE_CATEGORIES;
  const matched = m.series
    .filter(
      (s) =>
        s.page_title_norm.includes("Revenues by Source and Expenses by Function") &&
        sectionMatchesKind(s.section_norm, kind),
    )
    .map((s) => {
      const canonical = matchesCanonical(s.label_norm, canonicals);
      return canonical ? { ...s, label_norm: canonical } : null;
    })
    .filter((s): s is MergedSeries => s !== null);
  // Preserve canonical order for consistent stacking + legend
  const deduped = dedupeByLabel(matched);
  const order = new Map<string, number>(canonicals.map((c, i) => [c as string, i]));
  return deduped.sort((a, b) => (order.get(a.label_norm) ?? 99) - (order.get(b.label_norm) ?? 99));
}

/** Per-campus totals — each row's label is a campus name. */
export function perCampusRevExp(m: MergedRetrends, kind: RevExpKind): MergedSeries[] {
  const matched = m.series.filter(
    (s) =>
      s.page_title_norm.includes("Revenues and Expenses by Campus") &&
      sectionMatchesKind(s.section_norm, kind) &&
      CAMPUS_LABELS_SET.has(normalizeLabel(s.label_norm)),
  );
  return dedupeByLabel(matched);
}

/** Medical-center per-center totals. */
export function medCenterRevExp(m: MergedRetrends, kind: RevExpKind): MergedSeries[] {
  const matched = m.series.filter(
    (s) =>
      s.page_title_norm.includes("Revenues and Expenses by Medical Center") &&
      sectionMatchesKind(s.section_norm, kind) &&
      CAMPUS_LABELS_SET.has(normalizeLabel(s.label_norm)),
  );
  return dedupeByLabel(matched);
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
