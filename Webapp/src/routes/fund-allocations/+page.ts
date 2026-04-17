import type { PageLoad } from "./$types";
import type {
  AllocationReport,
  Manifest,
} from "$lib/types/ucop";

export const load: PageLoad = async ({ fetch, url }) => {
  const manifest = (await (await fetch("/data/ucop/manifest.json")).json()) as Manifest;
  const fiscalYears = manifest.general_fund_allocations.fiscal_years;
  const requestedFy = url.searchParams.get("fy");
  const selectedFy =
    requestedFy && fiscalYears.includes(requestedFy)
      ? requestedFy
      : (fiscalYears.at(-1) as string);

  const reports: AllocationReport[] = await Promise.all(
    fiscalYears.map((fy) =>
      fetch(`/data/ucop/general_fund_allocations/${fy}.json`).then(
        (r) => r.json() as Promise<AllocationReport>,
      ),
    ),
  );

  return { manifest, reports, selectedFy };
};
