# UC Finances Explorer

The explorer shows University of California State General Fund allocations and
revenue and expense trends. It uses SvelteKit and deploys to Vercel with the
project root set to `Webapp`.

## Local development

Run from this directory with Bun installed.

```sh
bun install --frozen-lockfile
bun run dev
bun run check
bun run build
```

## Refresh the financial reports

Run from the repository root. Install the dependencies from each scraper's
`requirements.txt` before fetching or parsing reports.

```sh
python3 Data/UCOP/general_fund_allocations/scraper.py --out Data/UCOP/general_fund_allocations
python3 Data/UCOP/reporting_transparency/scraper.py --series retrends --out Data/UCOP/reporting_transparency
python3 Data/UCOP/refresh_webapp.py
python3 Data/UCOP/refresh_webapp.py --check
```

For the reports already in the repository, start with `refresh_webapp.py`.
It validates both parsed series before writing, fills missing source URLs
from the UCOP catalogs, and writes the data and manifest under
`Webapp/static/data/ucop`. Review changes to report numbers and coverage
against the source PDFs before committing updated data. The refresh checks
report structure and coverage. Financial reconciliation is a separate review.
