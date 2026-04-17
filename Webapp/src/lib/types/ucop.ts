export type CampusKey =
  | "berkeley"
  | "davis"
  | "irvine"
  | "los_angeles"
  | "merced"
  | "riverside"
  | "san_diego"
  | "san_francisco"
  | "santa_barbara"
  | "santa_cruz";

export const CAMPUS_KEYS: CampusKey[] = [
  "berkeley",
  "davis",
  "irvine",
  "los_angeles",
  "merced",
  "riverside",
  "san_diego",
  "san_francisco",
  "santa_barbara",
  "santa_cruz",
];

export type AllocationColumn =
  | CampusKey
  | "campus_total"
  | "ucop_to_be_allocated"
  | "capital_debt_service"
  | "off_campus_programs"
  | "system_total";

export interface AllocationRow {
  page: number;
  label: string;
  indent: number;
  amounts: Partial<Record<AllocationColumn, number>>;
}

export interface AllocationReport {
  fiscal_year: string;
  title: string;
  source_url: string;
  source_filename: string;
  retrieved_at: string;
  columns: AllocationColumn[];
  rows: AllocationRow[];
}

export interface RetrendsRow {
  page: number;
  table_index: number;
  page_title: string;
  section: string;
  indent: number;
  label: string;
  amounts: Record<string, number>;
}

export interface RetrendsReport {
  series: "retrends";
  year_key: string;
  label: string;
  source_url: string;
  source_filename: string;
  retrieved_at: string;
  year_columns: string[];
  rows: RetrendsRow[];
}

export interface Manifest {
  general_fund_allocations: { path: string; fiscal_years: string[] };
  retrends: { path: string; years: string[] };
}

export interface YearValue {
  year: number;
  value: number;
}

export interface Series {
  id: string;
  label: string;
  values: YearValue[];
  color?: string;
}

export interface TreeNode {
  id: string;
  name: string;
  value?: number;
  children?: TreeNode[];
  depth?: number;
  indent?: number;
}
