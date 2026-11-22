import type { RetrendsReport, RetrendsRow, Series } from '$lib/types/ucop';
import { RETRENDS_UNIT } from '$lib/utils/chart';

export type RevExpKind = 'revenue' | 'expense';
type Scope = 'system' | 'campus' | 'medical';

interface MergedSeries {
	label: string;
	/** Amounts in the report's original unit of thousands of dollars. */
	amounts: Record<number, number>;
}

export interface RetrendsSource {
	reportYear: string;
	page: number;
	url: string;
	total: number;
}

interface MergedTable {
	series: MergedSeries[];
	sources: Partial<Record<number, RetrendsSource>>;
}

interface MergedRetrends {
	years: number[];
	tables: Record<Scope, Record<RevExpKind, MergedTable>>;
}

const CAMPUS_LABELS = [
	'Berkeley',
	'Davis',
	'Irvine',
	'Los Angeles',
	'Merced',
	'Riverside',
	'San Diego',
	'San Francisco',
	'Santa Barbara',
	'Santa Cruz'
];

const MEDICAL_LABELS = ['Davis', 'Irvine', 'Los Angeles', 'San Diego', 'San Francisco'];

function cleanLabel(value: string): string {
	return value
		.trim()
		.replace(/\d+(?:[,.]\s*[A-Z])?[,\s*]*$/, '')
		.trim()
		.replace(/\s+/g, ' ');
}

function categoryLabel(value: string): string {
	const label = cleanLabel(value);
	if (label === 'DOE Laboratories') return 'Department of Energy Laboratories';
	if (label === 'Interest') return 'Interest Expense';
	return label;
}

function scopeFor(title: string): Scope | undefined {
	if (/Revenues and Expenses by Medical Center/i.test(title)) return 'medical';
	if (/Revenues and Expenses by Campus/i.test(title)) return 'campus';
	if (/Revenues by Sources? and Expenses by Function/i.test(title)) return 'system';
}

function kindFor(section: string): RevExpKind | undefined {
	const heading = section.replace(/[\s\d,*]+$/, '').toUpperCase();
	if (heading === 'REVENUES') return 'revenue';
	if (heading === 'EXPENSES' || heading === 'EXPENDITURES') return 'expense';
}

const totalLabels = {
	revenue: /^Total Revenues$/i,
	expense: /^Total (?:Expenses|Expenditures)$/i
};

function reportTables(report: RetrendsReport) {
	const sections = new Map<string, { scope: Scope; kind: RevExpKind; rows: RetrendsRow[] }>();
	for (const row of report.rows) {
		const scope = scopeFor(row.page_title);
		const kind = kindFor(row.section);
		if (!scope || !kind) continue;

		const key = `${row.page}:${row.table_index}:${scope}:${kind}`;
		const section = sections.get(key) ?? { scope, kind, rows: [] };
		section.rows.push(row);
		sections.set(key, section);
	}
	// Some extracted facing pages repeat the following page’s financial table.
	return [...sections.values()].sort((a, b) => b.rows[0].page - a.rows[0].page);
}

function validatedColumn(rows: RetrendsRow[], year: string, requiredLabels: string[]) {
	const values = new Map<string, number>();
	let sum = 0;
	for (const row of rows) {
		const value = row.amounts[year];
		if (value === undefined) continue;
		if (!Number.isFinite(value)) return;

		const label = categoryLabel(row.label);
		if (values.has(label)) return;
		values.set(label, value);
		sum += value;
	}
	if (values.size === 0 || !requiredLabels.every((label) => values.has(label))) return;
	return { values, sum };
}

export function mergeRetrends(reports: RetrendsReport[]): MergedRetrends {
	const tables = {
		system: { revenue: emptyTable(), expense: emptyTable() },
		campus: { revenue: emptyTable(), expense: emptyTable() },
		medical: { revenue: emptyTable(), expense: emptyTable() }
	};
	const years = new Set<number>();

	// Select an entire table column from one report so totals, categories,
	// and any restatements always use the same source.
	for (const report of [...reports].sort((a, b) => b.year_key.localeCompare(a.year_key))) {
		for (const { scope, kind, rows } of reportTables(report)) {
			const totalIndex = rows.findIndex((row) => totalLabels[kind].test(cleanLabel(row.label)));
			if (totalIndex <= 0) continue;

			const table = tables[scope][kind];
			const totalRow = rows[totalIndex];
			const detailRows = rows.slice(0, totalIndex);
			const requiredLabels =
				scope === 'medical' ? MEDICAL_LABELS : scope === 'campus' ? CAMPUS_LABELS : [];
			for (const year of report.year_columns) {
				const fiscalYear = Number(year);
				if (!/^\d{4}$/.test(year) || table.sources[fiscalYear]) continue;

				const total = totalRow.amounts[year];
				if (!Number.isFinite(total) || total <= 0) continue;
				const column = validatedColumn(detailRows, year, requiredLabels);
				if (!column || column.sum !== total || (scope === 'system' && column.values.has('')))
					continue;

				table.sources[fiscalYear] = {
					reportYear: report.year_key,
					page: totalRow.page,
					url: report.source_url,
					total: total * RETRENDS_UNIT
				};
				years.add(fiscalYear);

				for (const [label, amount] of column.values) {
					if (scope !== 'system' && !requiredLabels.includes(label)) continue;
					let series = table.series.find((item) => item.label === label);
					if (!series) {
						series = { label, amounts: {} };
						table.series.push(series);
					}
					series.amounts[fiscalYear] = amount;
				}
			}
		}
	}

	return { years: [...years].sort((a, b) => a - b), tables };
}

function emptyTable(): MergedTable {
	return { series: [], sources: {} };
}

export function reportedSystemTotal(
	report: RetrendsReport,
	kind: RevExpKind,
	year = Number(report.year_key)
): number | undefined {
	return mergeRetrends([report]).tables.system[kind].sources[year]?.total;
}

export function toSeries(table: MergedTable): Series[] {
	return table.series.map(({ label, amounts }) => ({
		id: label,
		label,
		values: Object.entries(amounts)
			.map(([year, value]) => ({ year: Number(year), value: value * RETRENDS_UNIT }))
			.sort((a, b) => a.year - b.year)
	}));
}
