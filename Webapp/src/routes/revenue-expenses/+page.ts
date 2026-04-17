import type { PageLoad } from "./$types";
import type { Manifest, RetrendsReport } from "$lib/types/ucop";
import { mergeRetrends } from "$lib/utils/retrendsMerge";

export const load: PageLoad = async ({ fetch }) => {
  const manifest = (await (await fetch("/data/ucop/manifest.json")).json()) as Manifest;

  const reports: RetrendsReport[] = await Promise.all(
    manifest.retrends.years.map((y) =>
      fetch(`/data/ucop/retrends/${y}.json`).then(
        (r) => r.json() as Promise<RetrendsReport>,
      ),
    ),
  );

  const merged = mergeRetrends(reports);

  return { manifest, merged };
};
