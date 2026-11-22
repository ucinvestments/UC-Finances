import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import type { Manifest, RetrendsReport } from '$lib/types/ucop';
import { mergeRetrends } from '$lib/utils/retrendsMerge';

export const load: PageLoad = async ({ fetch }) => {
	const manifestResponse = await fetch('/data/ucop/manifest.json');
	if (!manifestResponse.ok) error(503, 'Financial reports are temporarily unavailable.');
	const manifest: Manifest = await manifestResponse.json();

	const reports = await Promise.all(
		manifest.retrends.years.map(async (year) => {
			const response = await fetch(`${manifest.retrends.path}/${year}.json`);
			if (!response.ok) error(503, `The ${year} financial report is temporarily unavailable.`);
			return (await response.json()) as RetrendsReport;
		})
	);

	const merged = mergeRetrends(reports);
	if (!merged.years.length) error(503, 'No validated financial reports are available.');

	return { merged };
};
