import type { PageLoad } from "./$types";
import type {
  AllocationReport,
  Manifest,
  RetrendsReport,
} from "$lib/types/ucop";

export const load: PageLoad = async ({ fetch }) => {
  const manifest = (await (await fetch("/data/ucop/manifest.json")).json()) as Manifest;

  const latestFy = manifest.general_fund_allocations.fiscal_years.at(-1)!;
  const latestYear = manifest.retrends.years.at(-1)!;

  const [latestAlloc, latestRetrends] = await Promise.all([
    fetch(`/data/ucop/general_fund_allocations/${latestFy}.json`).then(
      (r) => r.json() as Promise<AllocationReport>,
    ),
    fetch(`/data/ucop/retrends/${latestYear}.json`).then(
      (r) => r.json() as Promise<RetrendsReport>,
    ),
  ]);

  return { manifest, latestAlloc, latestRetrends };
};
