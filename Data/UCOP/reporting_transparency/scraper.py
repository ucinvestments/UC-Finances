#!/usr/bin/env python3
"""UC Reporting Transparency scraper.

Discovers, downloads, and parses PDFs published at
``reportingtransparency.universityofcalifornia.edu`` under ``/documents/finreports/``.

The primary parsed series is ``retrends`` (Revenue and Expense Trends — five-year
rolling tables broken out by system, campus, and medical center). Additional
finreports series are discovered and downloaded when present; parsing is only
implemented for retrends because each other series uses a different table layout.

Usage (inside the Docker image defined in the sibling Dockerfile):
    docker run --rm -v "$PWD:/data" ucop-retrends --all
    docker run --rm -v "$PWD:/data" ucop-retrends --discover
    docker run --rm -v "$PWD:/data" ucop-retrends --parse-only
    docker run --rm -v "$PWD:/data" ucop-retrends --series retrends --year 2025

Runnable directly (``python scraper.py --all --out .``) if requirements.txt is
installed locally.
"""
from __future__ import annotations

import argparse
import csv
import json
import logging
import re
import sys
import time
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path
from typing import Iterable, Optional

import pdfplumber
import requests

BASE_URL = "https://reportingtransparency.universityofcalifornia.edu"
DOCS_ROOT = f"{BASE_URL}/documents/finreports"
USER_AGENT = "UCOP-Finance-Scraper/1.0 (+https://github.com/ucinvestments/holdings)"

# Known finreports series and the inclusive range of fiscal years that are
# probed during discovery. Ranges were derived from the public landing page
# (https://reportingtransparency.universityofcalifornia.edu/). Years that don't
# resolve (HTTP != 200) are silently skipped.
#
# Year token meaning per series:
#   retrends     -> end fiscal year of the 5-year rolling report
#   debtcapital  -> fiscal year
#   endowment    -> last two digits of end fiscal year (e.g. "11" = 2010-11)
#   hwplan       -> fiscal-year pair (e.g. "2010-2011")
#   mppbond      -> fiscal year
KNOWN_SERIES: dict[str, dict] = {
    "retrends": {
        "path": "retrends/retrends_{year}.pdf",
        "year_format": "yyyy",
        "year_range": (2015, datetime.now().year + 1),
        "label": "Revenue and Expense Trends",
    },
    "debtcapital": {
        "path": "debtcapital/debtcapital_{year}.pdf",
        "year_format": "yyyy",
        "year_range": (2008, 2020),
        "label": "Annual Debt Capital Report",
    },
    "endowment": {
        "path": "endowment/endowment_{year}.pdf",
        "year_format": "yy",
        "year_range": (6, 15),  # 2006-07 .. 2014-15 (site shows through 2010-11)
        "label": "Endowment Report",
    },
    "hwplan": {
        "path": "hwplan/hwplan_{year}.pdf",
        "year_format": "yyyy-yyyy",
        "year_range": (2005, 2015),
        "label": "Health & Welfare Plans",
    },
    "mppbond": {
        "path": "mppbond/mppbond_{year}.pdf",
        "year_format": "yyyy",
        "year_range": (2004, 2012),
        "label": "UC Revenue Bonds / Multiple Purpose Project",
    },
}

log = logging.getLogger("ucop-retrends")


# ---------------------------------------------------------------------------
# Models
# ---------------------------------------------------------------------------


@dataclass
class DiscoveredDoc:
    series: str
    year_key: str           # canonical key used as filename stem, e.g. "2025"
    label: str              # human-readable title (series label + year)
    url: str
    filename: str           # basename chosen for local storage

    def to_dict(self) -> dict:
        return {
            "series": self.series,
            "year_key": self.year_key,
            "label": self.label,
            "url": self.url,
            "filename": self.filename,
        }


# ---------------------------------------------------------------------------
# HTTP
# ---------------------------------------------------------------------------


def make_session() -> requests.Session:
    s = requests.Session()
    s.headers.update({"User-Agent": USER_AGENT})
    return s


def _url_exists(session: requests.Session, url: str, timeout: float = 15.0) -> bool:
    try:
        r = session.head(url, timeout=timeout, allow_redirects=True)
    except requests.RequestException as e:
        log.debug("HEAD %s failed: %s", url, e)
        return False
    if r.status_code == 200:
        return True
    # Some hosts return 405 for HEAD — fall back to a ranged GET.
    if r.status_code in (403, 405):
        try:
            r2 = session.get(url, headers={"Range": "bytes=0-0"}, timeout=timeout, stream=True)
        except requests.RequestException:
            return False
        try:
            return r2.status_code in (200, 206)
        finally:
            r2.close()
    return False


def _enumerate_years(spec: dict) -> list[str]:
    lo, hi = spec["year_range"]
    fmt = spec["year_format"]
    years: list[str] = []
    if fmt == "yyyy":
        for y in range(lo, hi + 1):
            years.append(f"{y:04d}")
    elif fmt == "yy":
        for y in range(lo, hi + 1):
            years.append(f"{y:02d}")
    elif fmt == "yyyy-yyyy":
        for y in range(lo, hi):
            years.append(f"{y}-{y + 1}")
    else:
        raise ValueError(f"unknown year_format: {fmt}")
    return years


# ---------------------------------------------------------------------------
# Discovery
# ---------------------------------------------------------------------------


def discover(
    session: requests.Session,
    series: Optional[Iterable[str]] = None,
    probe_delay: float = 0.15,
) -> list[DiscoveredDoc]:
    wanted = set(series) if series else set(KNOWN_SERIES.keys())
    found: list[DiscoveredDoc] = []
    for name, spec in KNOWN_SERIES.items():
        if name not in wanted:
            continue
        log.info("Probing series %s", name)
        for y in _enumerate_years(spec):
            url = f"{DOCS_ROOT}/{spec['path'].format(year=y)}"
            if _url_exists(session, url):
                found.append(
                    DiscoveredDoc(
                        series=name,
                        year_key=y,
                        label=f"{spec['label']} {y}",
                        url=url,
                        filename=f"{y}.pdf",
                    )
                )
                log.debug("  FOUND %s", url)
            time.sleep(probe_delay)
    found.sort(key=lambda d: (d.series, d.year_key))
    log.info("Discovered %d PDFs across %d series", len(found), len({d.series for d in found}))
    return found


# ---------------------------------------------------------------------------
# Download
# ---------------------------------------------------------------------------


def download(
    session: requests.Session,
    doc: DiscoveredDoc,
    pdf_root: Path,
    force: bool = False,
    sleep: float = 0.3,
) -> Path:
    series_dir = pdf_root / doc.series
    series_dir.mkdir(parents=True, exist_ok=True)
    dest = series_dir / doc.filename
    if dest.exists() and not force:
        log.info("Skip %s/%s (exists)", doc.series, dest.name)
        return dest
    log.info("Downloading %s -> %s/%s", doc.url, doc.series, dest.name)
    resp = session.get(doc.url, timeout=60)
    resp.raise_for_status()
    dest.write_bytes(resp.content)
    time.sleep(sleep)
    return dest


# ---------------------------------------------------------------------------
# Retrends PDF parser
# ---------------------------------------------------------------------------
#
# Each retrends PDF contains many independent tables but they all share the
# same shape: a header row of four-digit fiscal-year tokens (e.g.
# "2021 2022 2023 2024 2025"), one or more section headers underneath
# ("REVENUES", "EXPENSES", "OTHER CHANGES IN NET POSITION"), and then
# right-aligned numeric rows with a leading label.
#
# Strategy:
#   1. Locate the year header on each page — it defines the right-edge
#      positions of every value column.
#   2. Treat all words above it on the page (minus the boilerplate
#      "(in thousands of dollars)" line) as the table title.
#   3. Walk rows below the year header. All-caps label-only rows become the
#      current section context. Rows with numbers are emitted with their
#      label, section, and {year → amount} mapping.
#   4. Footer hints ("Prepared by the UC Office of the President", "Page ...",
#      "Page intentionally left blank") are ignored.

# Year tokens may carry a trailing footnote marker like ``2015*`` (used in
# older retrends PDFs to flag restated figures); strip it during normalization.
YEAR_TOKEN_RE = re.compile(r"^(19|20)\d{2}\*?$")
# Numeric value tokens. ``$`` may be glued to the number (``$3,784,046``).
NUMERIC_TOKEN_RE = re.compile(
    r"^\$?\(?[\d]{1,3}(?:,\d{3})*(?:\.\d+)?\)?$|^\$?\(?\d+(?:\.\d+)?\)?$"
)
# Single-cell tokens that may stand alone in an amount column.
AMOUNT_TOKEN_RE = re.compile(r"^(\$|\(|\)|\-|\u2010|[\d]{1,3}(?:,\d{3})*(?:\.\d+)?|\.\d+|\d+)$")
BOILERPLATE_RE = re.compile(
    r"^\s*\(?in\s+thousands\s+of\s+dollars\)?\s*$",
    re.I,
)
FOOTER_HINTS = (
    "prepared by the uc office of the president",
    "page intentionally left blank",
)
SECTION_KEYWORDS = (
    "REVENUES",
    "EXPENSES",
    "EXPENDITURES",
    "OTHER CHANGES IN NET POSITION",
    "INCOME",
    "LOSS",
    "NET POSITION",
)


@dataclass
class ParsedRow:
    page: int
    table_index: int
    page_title: str
    section: str
    label: str
    indent: float
    amounts: dict[str, float] = field(default_factory=dict)


def _row_tolerance(size: float) -> float:
    return max(1.5, (size or 10) * 0.5)


_NUM_CONT_RE = re.compile(r"^[\d,]+(?:\.\d+)?$")


def _merge_number_fragments(words: list[dict]) -> list[dict]:
    """Merge split numeric tokens emitted as two adjacent words.

    pdfplumber occasionally splits e.g. ``16,128,000`` into ``1`` and
    ``6,128,000``. Merge any two same-line adjacent tokens whose texts both
    look like numeric continuations.
    """
    if not words:
        return words
    out: list[dict] = []
    for w in words:
        text = w["text"]
        if out and abs(w["top"] - out[-1]["top"]) < 2 and (w["x0"] - out[-1]["x1"]) < 2:
            prev = out[-1]["text"]
            if _NUM_CONT_RE.match(prev) and _NUM_CONT_RE.match(text):
                merged = dict(out[-1])
                merged["text"] = prev + text
                merged["x1"] = w["x1"]
                out[-1] = merged
                continue
        out.append(w)
    return out


def _cluster_rows(words: list[dict]) -> list[list[dict]]:
    if not words:
        return []
    words = sorted(words, key=lambda w: (round(w["top"], 1), w["x0"]))
    rows: list[list[dict]] = []
    current: list[dict] = []
    current_top: Optional[float] = None
    for w in words:
        size = w.get("height", 10)
        tol = _row_tolerance(size)
        if current_top is None or abs(w["top"] - current_top) <= tol:
            current.append(w)
            if current_top is None:
                current_top = w["top"]
        else:
            rows.append(current)
            current = [w]
            current_top = w["top"]
    if current:
        rows.append(current)
    return rows


def _parse_amount(text: str) -> Optional[float]:
    t = text.strip()
    if not t or t in {"-", "\u2010", "\u2013", "\u2014"}:
        return 0.0
    # Strip currency markers first so a glued token like "$(262,984)" still
    # gets recognized as a parenthesized negative.
    t = t.replace("$", "").strip()
    neg = False
    if t.startswith("(") and t.endswith(")"):
        neg = True
        t = t[1:-1]
    t = t.replace(",", "").strip()
    if not t or t == "-":
        return 0.0
    try:
        v = float(t)
    except ValueError:
        return None
    return -v if neg else v


def _is_year_header_row(row: list[dict]) -> Optional[list[tuple[str, float]]]:
    """If ``row`` is a year-header row, return [(year_str, x1), ...] sorted by x1.

    A year header has ≥3 four-digit year tokens, sharing a y-band, that form a
    run of consecutive integers. The consecutive-run check rejects cover lines
    like "FISCAL YEARS 2021 THROUGH 2025" (only two tokens, gap of 4).
    """
    candidates = [
        (w["text"].rstrip("*"), w["x1"], w["top"])
        for w in row
        if YEAR_TOKEN_RE.match(w["text"])
    ]
    if len(candidates) < 3:
        return None
    if max(t for _, _, t in candidates) - min(t for _, _, t in candidates) > 4:
        return None
    candidates.sort(key=lambda p: p[1])
    seen = set()
    deduped: list[tuple[str, float]] = []
    for y, x, _ in candidates:
        if y in seen:
            continue
        seen.add(y)
        deduped.append((y, x))
    if len(deduped) < 3:
        return None
    # Pick the longest consecutive run. Older retrends PDFs display the value
    # columns alongside an "Annual Percentage Change" block that repeats some
    # year labels; the value block is the longest +1-step run.
    ints = [int(y) for y, _ in deduped]
    best_start, best_len = 0, 1
    cur_start = 0
    for i in range(1, len(ints)):
        if ints[i] - ints[i - 1] == 1:
            cur_len = i - cur_start + 1
            if cur_len > best_len:
                best_start, best_len = cur_start, cur_len
        else:
            cur_start = i
    if best_len < 3:
        return None
    return deduped[best_start : best_start + best_len]


def _is_footer(text: str) -> bool:
    low = text.lower().strip()
    if not low:
        return False
    if re.match(r"^page\s+\d", low):
        return True
    return any(h in low for h in FOOTER_HINTS)


def _row_text(row: list[dict]) -> str:
    return " ".join(w["text"] for w in sorted(row, key=lambda w: w["x0"]))


def _looks_like_section_header(label: str, has_amounts: bool) -> bool:
    if has_amounts:
        return False
    stripped = label.strip()
    if not stripped or len(stripped) > 80:
        return False
    letters = [c for c in stripped if c.isalpha()]
    if not letters:
        return False
    upper_ratio = sum(1 for c in letters if c.isupper()) / len(letters)
    if upper_ratio < 0.8:
        return False
    upper = stripped.upper()
    return any(kw in upper for kw in SECTION_KEYWORDS)


def _split_row(
    row: list[dict], year_edges: list[tuple[str, float]]
) -> tuple[str, dict[str, float]]:
    """Split a row into (label, {year → amount}) using year-column right edges."""
    row = sorted(row, key=lambda w: w["x0"])
    first_edge_x = year_edges[0][1] if year_edges else float("inf")
    # Label tolerance: everything whose right edge is noticeably left of the
    # first year column, and which isn't itself a numeric token.
    label_tokens: list[str] = []
    value_tokens: list[dict] = []
    for w in row:
        text = w["text"]
        is_numeric_like = bool(
            AMOUNT_TOKEN_RE.match(text) or NUMERIC_TOKEN_RE.match(text)
        )
        if not value_tokens and not is_numeric_like and w["x1"] < first_edge_x - 10:
            label_tokens.append(text)
        elif not value_tokens and not is_numeric_like:
            # Trailing label token that sits under the first year column — treat
            # as still-label (e.g. footnote markers "REVENUES4").
            label_tokens.append(text)
        else:
            value_tokens.append(w)

    label = " ".join(label_tokens).strip()

    # Assemble cells from value_tokens; each numeric token / dash becomes a cell.
    cells: list[tuple[float, str]] = []
    pending_prefix = ""
    for w in value_tokens:
        t = w["text"]
        x1 = w["x1"]
        if t == "$":
            pending_prefix = "$"
            continue
        if t in {"-", "\u2010", "\u2013", "\u2014"}:
            cells.append((x1, "-"))
            pending_prefix = ""
            continue
        if NUMERIC_TOKEN_RE.match(t):
            cells.append((x1, pending_prefix + t))
            pending_prefix = ""
            continue
        pending_prefix = ""

    amounts: dict[str, float] = {}
    best_dist: dict[str, float] = {}
    for x_right, text in cells:
        val = _parse_amount(text)
        if val is None:
            continue
        year, dist = None, float("inf")
        for y, ex in year_edges:
            d = abs(ex - x_right)
            if d < dist:
                year, dist = y, d
        if year is None:
            continue
        # Only accept if within a reasonable horizontal window.
        if dist > 40:
            continue
        if year not in amounts or dist < best_dist[year]:
            amounts[year] = val
            best_dist[year] = dist

    return label, amounts


def parse_retrends_pdf(pdf_path: Path) -> list[ParsedRow]:
    out: list[ParsedRow] = []
    with pdfplumber.open(str(pdf_path)) as pdf:
        table_index = 0
        current_page_title = ""
        current_years: list[tuple[str, float]] = []
        current_section = ""

        for page_idx, page in enumerate(pdf.pages, start=1):
            words = page.extract_words(use_text_flow=False, keep_blank_chars=False)
            if not words:
                continue
            words = _merge_number_fragments(words)
            rows = _cluster_rows(words)

            # Each page may start a new table (new year-header row). If it
            # doesn't, it's a continuation and we keep current context.
            header_row_idx: Optional[int] = None
            header_edges: Optional[list[tuple[str, float]]] = None
            for i, row in enumerate(rows):
                edges = _is_year_header_row(row)
                if edges:
                    header_row_idx = i
                    header_edges = edges
                    break

            if header_row_idx is not None and header_edges:
                # Collect title rows above the year header (skip boilerplate).
                title_bits: list[str] = []
                for row in rows[:header_row_idx]:
                    text = _row_text(row)
                    if _is_footer(text) or BOILERPLATE_RE.match(text):
                        continue
                    title_bits.append(text.strip())
                if title_bits:
                    current_page_title = " — ".join(t for t in title_bits if t)
                current_years = header_edges
                current_section = ""
                table_index += 1
                body_rows = rows[header_row_idx + 1:]
            else:
                body_rows = rows

            if not current_years:
                continue

            for row in body_rows:
                text = _row_text(row)
                if not text or _is_footer(text) or BOILERPLATE_RE.match(text):
                    continue
                label, amounts = _split_row(row, current_years)
                if not label and not amounts:
                    continue
                # Skip stray footnote rows like "1 Certain revenues..."
                if not amounts and label and re.match(r"^\d+\s", label):
                    continue
                if _looks_like_section_header(label, bool(amounts)):
                    current_section = label
                    continue
                indent = min((w["x0"] for w in row), default=0.0)
                out.append(
                    ParsedRow(
                        page=page_idx,
                        table_index=table_index,
                        page_title=current_page_title,
                        section=current_section,
                        label=label,
                        indent=round(indent, 2),
                        amounts=amounts,
                    )
                )
    return out


# ---------------------------------------------------------------------------
# Output
# ---------------------------------------------------------------------------


def write_retrends_outputs(
    doc: DiscoveredDoc,
    pdf_path: Path,
    rows: list[ParsedRow],
    parsed_root: Path,
) -> tuple[Path, Path]:
    out_dir = parsed_root / doc.series
    out_dir.mkdir(parents=True, exist_ok=True)

    # Canonical year columns for this PDF (sorted across all rows).
    years_seen: set[str] = set()
    for r in rows:
        years_seen.update(r.amounts.keys())
    years = sorted(years_seen)

    json_path = out_dir / f"{doc.year_key}.json"
    json_path.write_text(
        json.dumps(
            {
                "series": doc.series,
                "year_key": doc.year_key,
                "label": doc.label,
                "source_url": doc.url,
                "source_filename": pdf_path.name,
                "retrieved_at": datetime.now(timezone.utc).isoformat(timespec="seconds"),
                "year_columns": years,
                "rows": [
                    {
                        "page": r.page,
                        "table_index": r.table_index,
                        "page_title": r.page_title,
                        "section": r.section,
                        "indent": r.indent,
                        "label": r.label,
                        "amounts": r.amounts,
                    }
                    for r in rows
                ],
            },
            indent=2,
            sort_keys=False,
        )
    )

    csv_path = out_dir / f"{doc.year_key}.csv"
    with csv_path.open("w", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(
            ["page", "table_index", "page_title", "section", "indent", "label", *years]
        )
        for r in rows:
            writer.writerow(
                [
                    r.page,
                    r.table_index,
                    r.page_title,
                    r.section,
                    r.indent,
                    r.label,
                    *[r.amounts.get(y, "") for y in years],
                ]
            )
    return json_path, csv_path


def write_catalog(docs: Iterable[DiscoveredDoc], out_path: Path) -> None:
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(
        json.dumps(
            {
                "retrieved_at": datetime.now(timezone.utc).isoformat(timespec="seconds"),
                "base_url": BASE_URL,
                "documents": [d.to_dict() for d in docs],
            },
            indent=2,
        )
    )


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------


def parse_args(argv: list[str]) -> argparse.Namespace:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--out", type=Path, default=Path("/data"), help="Output base directory")
    p.add_argument(
        "--series",
        action="append",
        choices=sorted(KNOWN_SERIES.keys()),
        help="Restrict to specific series (default: all known series)",
    )
    p.add_argument("--year", action="append", help="Restrict to specific year key(s) (e.g. 2025)")
    p.add_argument("--discover", action="store_true", help="Write the discovery catalog and exit")
    p.add_argument("--download-only", action="store_true", help="Download PDFs but skip parsing")
    p.add_argument("--parse-only", action="store_true", help="Parse already-downloaded PDFs")
    p.add_argument("--all", action="store_true", help="Discover, download, and parse (default)")
    p.add_argument("--force", action="store_true", help="Re-download existing PDFs")
    p.add_argument("-v", "--verbose", action="store_true")
    return p.parse_args(argv)


def _docs_from_local_pdfs(pdf_root: Path, series_filter: Optional[set[str]]) -> list[DiscoveredDoc]:
    docs: list[DiscoveredDoc] = []
    for series_dir in sorted(p for p in pdf_root.iterdir() if p.is_dir()):
        name = series_dir.name
        if series_filter and name not in series_filter:
            continue
        spec = KNOWN_SERIES.get(name)
        label = spec["label"] if spec else name
        for pdf in sorted(series_dir.glob("*.pdf")):
            year = pdf.stem
            docs.append(
                DiscoveredDoc(
                    series=name,
                    year_key=year,
                    label=f"{label} {year}",
                    url="",
                    filename=pdf.name,
                )
            )
    return docs


def main(argv: Optional[list[str]] = None) -> int:
    args = parse_args(list(sys.argv[1:] if argv is None else argv))
    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(asctime)s %(levelname)s %(message)s",
    )

    out_dir: Path = args.out
    pdf_root = out_dir / "pdfs"
    parsed_root = out_dir / "parsed"

    do_download = not args.parse_only
    do_parse = not args.discover and not args.download_only

    session = make_session()
    series_filter = set(args.series) if args.series else None

    if args.parse_only:
        docs = _docs_from_local_pdfs(pdf_root, series_filter)
    else:
        docs = discover(session, series=series_filter)
        write_catalog(docs, out_dir / "catalog.json")

    if args.year:
        wanted = set(args.year)
        docs = [d for d in docs if d.year_key in wanted]
        log.info("Filtered to years: %s", ", ".join(sorted(wanted)))

    if args.discover:
        return 0

    if do_download:
        for doc in docs:
            try:
                download(session, doc, pdf_root, force=args.force)
            except requests.HTTPError as e:
                log.error("Failed to download %s: %s", doc.url, e)

    if do_parse:
        for doc in docs:
            pdf_path = pdf_root / doc.series / doc.filename
            if not pdf_path.exists():
                log.warning("Missing PDF for %s/%s: %s", doc.series, doc.year_key, pdf_path)
                continue
            if doc.series != "retrends":
                log.info("Skipping parse for %s/%s (parser not implemented)", doc.series, doc.year_key)
                continue
            try:
                rows = parse_retrends_pdf(pdf_path)
            except Exception:
                log.exception("Parse failed for %s", pdf_path.name)
                continue
            json_path, csv_path = write_retrends_outputs(doc, pdf_path, rows, parsed_root)
            with_amounts = sum(1 for r in rows if r.amounts)
            log.info(
                "%s/%s parsed: %d rows (%d with amounts) -> %s, %s",
                doc.series,
                doc.year_key,
                len(rows),
                with_amounts,
                json_path.relative_to(out_dir),
                csv_path.relative_to(out_dir),
            )

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
