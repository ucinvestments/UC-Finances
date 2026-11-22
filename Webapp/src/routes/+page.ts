import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";
import type {
  AllocationReport,
  Manifest,
  RetrendsReport,
} from "$lib/types/ucop";

export const load: PageLoad = async ({ fetch }) => {
  const manifestResponse = await fetch("/data/ucop/manifest.json");
  if (!manifestResponse.ok) error(503, "Financial reports are temporarily unavailable.");
  const manifest: Manifest = await manifestResponse.json();

  const allocations = manifest.general_fund_allocations;
  const retrends = manifest.retrends;
  const latestFy = allocations.fiscal_years.at(-1);
  const latestYear = retrends.years.at(-1);
  if (!latestFy || !latestYear) error(503, "No financial reports are available.");

  const [allocationResponse, retrendsResponse] = await Promise.all([
    fetch(`${allocations.path}/${latestFy}.json`),
    fetch(`${retrends.path}/${latestYear}.json`),
  ]);
  if (!allocationResponse.ok || !retrendsResponse.ok) {
    error(503, "The latest financial reports are temporarily unavailable.");
  }
  const [latestAlloc, latestRetrends]: [AllocationReport, RetrendsReport] = await Promise.all([
    allocationResponse.json(),
    retrendsResponse.json(),
  ]);

  return { manifest, latestAlloc, latestRetrends };
};
