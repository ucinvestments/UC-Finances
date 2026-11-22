# UC Finances

The SvelteKit explorer in `Webapp/` covers the University of California’s
State General Fund allocations and revenue and expense trends. Use Bun to
develop, check, and build the app. The existing deployment uses the
SvelteKit Vercel adapter. See [Webapp/README.md](Webapp/README.md) for commands.

`Data/UCOP/` holds the source PDFs, scraper code, catalogs, and parsed
financial reports. Keep the web copies and manifest in sync with the parsed
reports using `python3 Data/UCOP/refresh_webapp.py`.

The federal award snapshot and SEC filings are separate research datasets.
Their scope and outstanding validation are documented in
[the USAspending notes](Data/Federal/USASpending.gov/Readme.md) and
[the SEC notes](Data/SEC/Data/README.md).
