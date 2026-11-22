import {
  CAMPUS_KEYS,
  type AllocationColumn,
  type AllocationReport,
  type AllocationRow,
  type TreeNode,
} from "$lib/types/ucop";

const columns = [
  ...CAMPUS_KEYS,
  "campus_total",
  "ucop_to_be_allocated",
  "capital_debt_service",
  "off_campus_programs",
  "system_total",
] satisfies AllocationColumn[];

type Amounts = AllocationRow["amounts"];

export const CANONICAL_CATEGORIES = [
  "Earmarks subject to cost adjustments",
  "Earmarks not subject to cost adjustments",
  "Systemwide commitments",
  "Campus-specific allocations",
  "Per-student funding",
  "One-time allocations",
  "Other adjustments",
] as const;

export interface AllocationCategory {
  label: (typeof CANONICAL_CATEGORIES)[number];
  amounts: Amounts;
}

export interface AllocationSummary {
  fiscalYear: string;
  categories: AllocationCategory[];
  difference: Amounts;
  reported: Amounts;
}

function normalizedLabel(row: AllocationRow): string {
  return row.label.toLowerCase().replace(/[‐‑–—]/g, "-").replace(/\s+/g, "");
}

function sumRows(rows: AllocationRow[]): Amounts {
  const amounts: Amounts = {};
  for (const row of rows) {
    for (const column of columns) {
      const value = row.amounts[column];
      if (value !== undefined) amounts[column] = (amounts[column] ?? 0) + value;
    }
  }
  return amounts;
}

export function summarizeAllocations(report: AllocationReport): AllocationSummary {
  const { rows } = report;
  const find = (pattern: RegExp) => rows.find((row) => pattern.test(normalizedLabel(row)));
  const reported = find(/stategeneralfundsupport.*permanent.*one-time/)?.amounts ?? {};
  const permanent = find(/^totalpermanentstategeneralfunds/)?.amounts ?? {};
  const oneTime = { ...find(/^totalone-timeallocations/)?.amounts };
  const otherRow = find(/^totalotheradjustments/);
  const other = { ...otherRow?.amounts };

  // When one component is absent from the PDF extraction, recover it only
  // when the other components of the published net total are known.
  for (const column of columns) {
    const total = reported[column];
    const permanentValue = permanent[column];
    if (total === undefined || permanentValue === undefined) continue;

    const oneTimeValue = oneTime[column];
    if (oneTimeValue === undefined && (!otherRow || other[column] !== undefined)) {
      oneTime[column] = total - permanentValue - (other[column] ?? 0);
    } else if (otherRow && other[column] === undefined && oneTimeValue !== undefined) {
      other[column] = total - permanentValue - oneTimeValue;
    }
  }

  const commitmentsStart = rows.findIndex((row) => /^systemwidecommitments/.test(normalizedLabel(row)));
  const corridorStart = rows.findIndex((row) => /^(campus-specificallocations|merced$)/.test(normalizedLabel(row)));
  const commitmentRows = commitmentsStart >= 0 && corridorStart > commitmentsStart
    ? rows.slice(commitmentsStart + 1, corridorStart)
    : [];
  // The corridor section has two campus rows. Some extracts lack the heading,
  // and the oldest report prints both campus amounts on a single row.
  const corridorRows = [find(/^merced$/), find(/^sanfrancisco$/)].filter(
    (row): row is AllocationRow => row !== undefined,
  );
  const categoryAmounts = [
    find(/^subtotal-*(?:(?:earmarks|set-asides))?subjecttocostadjustment/)?.amounts ?? {},
    find(/^subtotal-*(?:(?:earmarks|set-asides))?notsubjecttocostadjustment/)?.amounts ?? {},
    sumRows(commitmentRows),
    sumRows(corridorRows),
    (find(/^netstatefundsforenrollment$/) ?? find(/^fundingprovided$/))?.amounts ?? {},
    oneTime,
    other,
  ];
  const categories = CANONICAL_CATEGORIES.map((label, index) => ({
    label,
    amounts: categoryAmounts[index],
  }));

  const difference: Amounts = {};
  for (const column of columns) {
    const total = reported[column];
    if (total !== undefined) {
      difference[column] = total - categories.reduce(
        (sum, category) => sum + (category.amounts[column] ?? 0),
        0,
      );
    }
  }
  return { fiscalYear: report.fiscal_year, categories, difference, reported };
}

export function allocationTree(summary: AllocationSummary): TreeNode {
  const children = summary.categories.flatMap(({ label, amounts }) => {
    const value = amounts.system_total;
    return value !== undefined && value > 0 ? [{ id: label, name: label, value }] : [];
  });
  return { id: "root", name: summary.fiscalYear, children };
}
