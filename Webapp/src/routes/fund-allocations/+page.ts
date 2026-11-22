import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";
import type { AllocationReport, Manifest } from "$lib/types/ucop";

export const load: PageLoad = async ({ fetch }) => {
  const manifestResponse = await fetch("/data/ucop/manifest.json");
  if (!manifestResponse.ok) error(503, "Allocation reports are temporarily unavailable.");
  const manifest: Manifest = await manifestResponse.json();
  const { path, fiscal_years: fiscalYears } = manifest.general_fund_allocations;
  if (!fiscalYears.length) error(503, "No allocation reports are available.");

  const reports = await Promise.all(
    fiscalYears.map(async (year) => {
      const response = await fetch(`${path}/${year}.json`);
      if (!response.ok) error(503, `The FY ${year} allocation report is temporarily unavailable.`);
      return (await response.json()) as AllocationReport;
    }),
  );

  return { fiscalYears, reports };
};
