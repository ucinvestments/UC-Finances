import type {
  AllocationColumn,
  AllocationReport,
  RetrendsRow,
  TreeNode,
} from "$lib/types/ucop";

function slug(s: string): string {
  return s
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "_")
    .replace(/^_|_$/g, "");
}

interface HierarchyOptions {
  metric?: AllocationColumn;
  skipEmpty?: boolean;
  /** Drop rows whose label starts with "Total" / "Subtotal" / "Grand Total" to avoid double-counting in sum hierarchies. Defaults true. */
  skipTotals?: boolean;
}

const TOTAL_RE = /^(grand\s+total|total|subtotal|sub-total|sub total)\b/i;
// "2025-26 State General Fund Support (Permanent + One-Time)" is the grand total
// in each report — it sums every other top-level row and must be excluded from
// hierarchy sums or it doubles the root value.
const GRAND_TOTAL_RE =
  /(State General Fund Support.*Permanent.*One.?Time|Permanent.*\+.*One.?Time)/i;

export function isTotalRow(label: string): boolean {
  const t = label.trim();
  return TOTAL_RE.test(t) || GRAND_TOTAL_RE.test(t);
}

export function buildAllocationTree(
  report: AllocationReport,
  opts: HierarchyOptions = {},
): TreeNode {
  const metric = opts.metric ?? "system_total";
  const skipEmpty = opts.skipEmpty ?? true;
  const skipTotals = opts.skipTotals ?? true;

  const root: TreeNode = {
    id: "root",
    name: report.fiscal_year,
    children: [],
    indent: -Infinity,
    depth: 0,
  };
  const stack: TreeNode[] = [root];

  const rows = skipTotals
    ? report.rows.filter((r) => !isTotalRow(r.label))
    : report.rows;

  for (const row of rows) {
    while (stack.length > 1 && (stack[stack.length - 1].indent ?? 0) >= row.indent) {
      stack.pop();
    }
    const parent = stack[stack.length - 1];
    const v = row.amounts?.[metric];
    const node: TreeNode = {
      id: `${parent.id}.${slug(row.label)}`,
      name: row.label,
      indent: row.indent,
      depth: (parent.depth ?? 0) + 1,
      value: typeof v === "number" && v !== 0 ? v : undefined,
      children: [],
    };
    (parent.children ??= []).push(node);
    stack.push(node);
  }

  pruneEmpty(root, skipEmpty);
  return root;
}

export function buildRetrendsTree(rows: RetrendsRow[], year: string): TreeNode {
  const root: TreeNode = {
    id: "root",
    name: year,
    children: [],
    indent: -Infinity,
    depth: 0,
  };
  const stack: TreeNode[] = [root];

  for (const row of rows) {
    while (stack.length > 1 && (stack[stack.length - 1].indent ?? 0) >= row.indent) {
      stack.pop();
    }
    const parent = stack[stack.length - 1];
    const v = row.amounts?.[year];
    const node: TreeNode = {
      id: `${parent.id}.${slug(row.label)}`,
      name: row.label,
      indent: row.indent,
      depth: (parent.depth ?? 0) + 1,
      value: typeof v === "number" ? v * 1000 : undefined,
      children: [],
    };
    (parent.children ??= []).push(node);
    stack.push(node);
  }

  pruneEmpty(root, true);
  return root;
}

function pruneEmpty(node: TreeNode, skipEmpty: boolean): number {
  if (!node.children || node.children.length === 0) {
    return node.value ?? 0;
  }
  let sum = 0;
  const kept: TreeNode[] = [];
  for (const c of node.children) {
    const cv = pruneEmpty(c, skipEmpty);
    const hasChildren = (c.children?.length ?? 0) > 0;
    if (cv > 0 || hasChildren || !skipEmpty) {
      kept.push(c);
      sum += cv;
    }
  }
  node.children = kept;
  if (node.value === undefined && sum > 0) node.value = sum;
  return node.value ?? sum;
}

export function topLevelAllocationRows(report: AllocationReport): AllocationRowSlim[] {
  if (report.rows.length === 0) return [];
  const minIndent = Math.min(...report.rows.map((r) => r.indent));
  const leaves: AllocationRowSlim[] = [];
  for (const row of report.rows) {
    const hasAmount = typeof row.amounts?.system_total === "number";
    if (hasAmount && row.indent > minIndent + 0.01) {
      leaves.push({ label: row.label, indent: row.indent, amounts: row.amounts });
    }
  }
  return leaves;
}

export interface AllocationRowSlim {
  label: string;
  indent: number;
  amounts: Partial<Record<AllocationColumn, number>>;
}

/**
 * Normalize a category label for year-over-year matching:
 *   - lower-cased
 *   - unicode hyphens (U+2010, U+2013, U+2014) collapsed to ASCII "-"
 *   - trailing footnote digits removed
 *   - "set-asides" renamed to "earmarks" (UCOP changed terminology in FY23-24)
 */
export function normalizeCategoryLabel(label: string): string {
  return label
    .replace(/[\u2010\u2011\u2012\u2013\u2014]/g, "-")
    .replace(/\d+$/, "")
    .trim()
    .toLowerCase()
    .replace(/\bset-asides?\b/g, "earmarks");
}

/** Canonical top-level categories used for the 10-year trend chart. */
export const CANONICAL_CATEGORIES = [
  "Earmarks subject to cost adjustments",
  "Earmarks not subject to cost adjustments",
  "Systemwide commitments",
  "Campus-Specific Allocations (Corridors)",
  "Funding Provided on Per-Student Basis",
  "One-Time Allocations",
] as const;

/**
 * For a single AllocationReport, sum system_total of every descendant row
 * beneath each top-level category. Works across years with different indent
 * scales by using the report's own minimum indent as the top-level threshold.
 *
 * Returns a map: canonical category name -> summed system_total.
 */
export function categoryTotalsForReport(
  report: AllocationReport,
): Record<string, number> {
  if (report.rows.length === 0) return {};
  const minIndent = Math.min(...report.rows.map((r) => r.indent));
  const topThreshold = minIndent + 0.5;

  const result: Record<string, number> = {};
  for (const c of CANONICAL_CATEGORIES) result[c] = 0;

  const normToCanonical = new Map<string, string>();
  for (const c of CANONICAL_CATEGORIES) normToCanonical.set(normalizeCategoryLabel(c), c);

  let currentCanonical: string | null = null;
  let currentTopIndent = Infinity;
  for (const row of report.rows) {
    if (row.indent <= topThreshold) {
      // Candidate top-level row.
      if (isTotalRow(row.label)) {
        currentCanonical = null;
        continue;
      }
      const norm = normalizeCategoryLabel(row.label);
      currentCanonical = normToCanonical.get(norm) ?? null;
      currentTopIndent = row.indent;
      continue;
    }
    if (!currentCanonical) continue;
    if (row.indent <= currentTopIndent) {
      currentCanonical = null;
      continue;
    }
    if (isTotalRow(row.label)) continue;
    const v = row.amounts?.system_total;
    if (typeof v === "number") result[currentCanonical] += v;
  }
  return result;
}
