# UC Reporting Transparency scraper

Discovers, downloads, and parses PDFs from
<https://reportingtransparency.universityofcalifornia.edu/documents/finreports/>.

The primary parsed series is **retrends** (Revenue and Expense Trends —
five-year rolling tables broken out by system, campus, and medical center).
Other finreports series (`debtcapital`, `endowment`, `hwplan`, `mppbond`) are
discovered and downloaded when present, but parsing is only implemented for
retrends because each series uses its own table layout.

## Layout

```
reporting_transparency/
├── Dockerfile
├── requirements.txt
├── scraper.py
├── catalog.json            # written by --discover / --all
├── pdfs/<series>/<year>.pdf
└── parsed/retrends/<year>.{json,csv}
```

## Usage

Build the image once, then run it from this directory so the bind-mount lands
the catalog/PDFs/parsed output here:

```
docker build -t ucop-retrends .

docker run --rm -v "$PWD:/data" ucop-retrends --all
docker run --rm -v "$PWD:/data" ucop-retrends --discover
docker run --rm -v "$PWD:/data" ucop-retrends --series retrends --year 2025
docker run --rm -v "$PWD:/data" ucop-retrends --parse-only --series retrends
```

Direct invocation works too if `requirements.txt` is installed locally:

```
python scraper.py --all --out .
```

## Output schema (retrends)

`parsed/retrends/<year>.json`:

```jsonc
{
  "series": "retrends",
  "year_key": "2025",
  "label": "Revenue and Expense Trends 2025",
  "source_url": "https://.../retrends_2025.pdf",
  "source_filename": "2025.pdf",
  "retrieved_at": "2026-04-17T03:23:35+00:00",
  "year_columns": ["2021", "2022", "2023", "2024", "2025"],
  "rows": [
    {
      "page": 12,
      "table_index": 1,
      "page_title": "University of California System — Revenues by Source ...",
      "section": "REVENUES",
      "indent": 38.6,
      "label": "Student Tuition and Fees",
      "amounts": {"2021": 5100747.0, "2022": 5440530.0, ...}
    }
  ]
}
```

The CSV mirror has columns
`page,table_index,page_title,section,indent,label,<year_columns...>`.

## Known limitations

- Older retrends PDFs (2015–2019) include per-campus pages whose tables use
  campus names as column headers rather than fiscal years. Those pages are
  intentionally skipped — only the year-columned tables (system, per-campus
  five-year summary, medical centers) are extracted.
- The other finreports series (`debtcapital`, `endowment`, `hwplan`,
  `mppbond`) are downloaded only; their PDF layouts aren't year-columned and
  would need their own parsers.
