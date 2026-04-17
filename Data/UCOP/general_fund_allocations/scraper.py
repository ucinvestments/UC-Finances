#!/usr/bin/env python3
"""UCOP "Summary of State General Fund Allocations" scraper.

Discovers per-year PDFs from the UCOP Operating Budget "Other Resources" page,
downloads them, and parses each into structured JSON + CSV.

Usage (inside the Docker image defined in the sibling Dockerfile):
    docker run --rm -v "$PWD:/data" ucop-gfa --all
    docker run --rm -v "$PWD:/data" ucop-gfa --discover
    docker run --rm -v "$PWD:/data" ucop-gfa --parse-only
    docker run --rm -v "$PWD:/data" ucop-gfa --year 2024-25

The script is also runnable directly (`python scraper.py --all --out .`) if the
requirements in requirements.txt are installed locally.
"""
from __future__ import annotations

import argparse
import csv
import json
import logging
import re
import sys
import time
import unicodedata
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Iterable, Optional
from urllib.parse import urljoin, urlparse

import pdfplumber
import requests
from bs4 import BeautifulSoup

INDEX_URL = "https://www.ucop.edu/operating-budget/budgets-and-reports/other-resources/index.html"
USER_AGENT = "UCOP-Finance-Scraper/1.0 (+https://github.com/ucinvestments/holdings)"

LINK_TEXT_RE = re.compile(r"summary\s*of\s*state\s*general\s*fund\s*allocations", re.I)
YEAR_RE = re.compile(r"(20\d{2})\s*[\-\u2010\u2011\u2012\u2013\u2014]\s*(\d{2}|20\d{2})")

# Amount cell tokens: currency marker, a formatted number, a dash, or parenthesized negatives.
AMOUNT_TOKEN_RE = re.compile(r"^(\$|\(|\)|\-|\u2010|[\d]{1,3}(?:,\d{3})*(?:\.\d+)?|\.\d+|\d+)$")
NUMERIC_TOKEN_RE = re.compile(r"^\(?[\d]{1,3}(?:,\d{3})*(?:\.\d+)?\)?$|^\(?\d+(?:\.\d+)?\)?$")

log = logging.getLogger("ucop-gfa")


# ---------------------------------------------------------------------------
# Models
# ---------------------------------------------------------------------------


@dataclass
class DiscoveredDoc:
    fiscal_year: str  # e.g. "2025-26"
    title: str
    url: str
    filename: str  # basename chosen for local storage


@dataclass
class ParsedDoc:
    fiscal_year: str
    source_url: str
    source_filename: str
    retrieved_at: str
    columns: list[str]
    rows: list[dict]


# ---------------------------------------------------------------------------
# Discovery
# ---------------------------------------------------------------------------


def make_session() -> requests.Session:
    s = requests.Session()
    s.headers.update({"User-Agent": USER_AGENT})
    return s


def _normalize_year(raw: str) -> Optional[str]:
    m = YEAR_RE.search(unicodedata.normalize("NFKC", raw))
    if not m:
        return None
    start, end = m.group(1), m.group(2)
    if len(end) == 2:
        end_full = end
    else:
        end_full = end[-2:]
    return f"{start}-{end_full}"


def discover(session: requests.Session, index_url: str = INDEX_URL) -> list[DiscoveredDoc]:
    log.info("Fetching index page: %s", index_url)
    resp = session.get(index_url, timeout=30)
    resp.raise_for_status()
    soup = BeautifulSoup(resp.text, "lxml")

    docs: dict[str, DiscoveredDoc] = {}
    for a in soup.find_all("a", href=True):
        href = a["href"].strip()
        text = " ".join(a.get_text(" ").split())
        hay = f"{text} {href}"
        if not href.lower().endswith(".pdf"):
            continue
        if not LINK_TEXT_RE.search(hay):
            continue

        year = _normalize_year(text) or _normalize_year(href)
        if not year:
            log.warning("Could not determine fiscal year for: %s (%s)", text, href)
            continue

        full_url = urljoin(index_url, href)
        orig_name = Path(urlparse(full_url).path).name
        local_name = f"{year}.pdf"

        docs.setdefault(
            year,
            DiscoveredDoc(
                fiscal_year=year,
                title=text or orig_name,
                url=full_url,
                filename=local_name,
            ),
        )

    ordered = sorted(docs.values(), key=lambda d: d.fiscal_year, reverse=True)
    log.info("Discovered %d PDFs (%s)", len(ordered), ", ".join(d.fiscal_year for d in ordered))
    return ordered


# ---------------------------------------------------------------------------
# Download
# ---------------------------------------------------------------------------


def download(
    session: requests.Session,
    doc: DiscoveredDoc,
    pdf_dir: Path,
    force: bool = False,
    sleep: float = 0.5,
) -> Path:
    pdf_dir.mkdir(parents=True, exist_ok=True)
    dest = pdf_dir / doc.filename
    if dest.exists() and not force:
        log.info("Skip %s (exists)", dest.name)
        return dest
    log.info("Downloading %s -> %s", doc.url, dest.name)
    resp = session.get(doc.url, timeout=60)
    resp.raise_for_status()
    dest.write_bytes(resp.content)
    time.sleep(sleep)
    return dest


# ---------------------------------------------------------------------------
# Parsing
# ---------------------------------------------------------------------------


# Canonical column keys in canonical left-to-right order.
CANONICAL_COLUMNS = [
    "berkeley",
    "davis",
    "irvine",
    "los_angeles",
    "merced",
    "riverside",
    "san_diego",
    "san_francisco",
    "santa_barbara",
    "santa_cruz",
    "campus_total",
    "ucop_to_be_allocated",
    "capital_debt_service",
    "off_campus_programs",
    "system_total",
]

# Header anchors used to locate each column's x-position. Each value is a list of
# token sequences; the first sequence to match wins. Matching is case-insensitive
# and restricted to the header y-band to avoid false hits in the body.
COLUMN_ANCHORS: list[tuple[str, list[list[str]]]] = [
    ("berkeley", [["BERKELEY"]]),
    ("davis", [["DAVIS"]]),
    ("irvine", [["IRVINE"]]),
    ("los_angeles", [["LOS", "ANGELES"], ["ANGELES"]]),
    ("merced", [["MERCED"]]),
    ("riverside", [["RIVERSIDE"]]),
    ("san_diego", [["SAN", "DIEGO"], ["DIEGO"]]),
    ("san_francisco", [["SAN", "FRANCISCO"], ["FRANCISCO"]]),
    ("santa_barbara", [["SANTA", "BARBARA"], ["BARBARA"]]),
    ("santa_cruz", [["SANTA", "CRUZ"], ["CRUZ"]]),
    ("campus_total", [["CAMPUS", "TOTAL"]]),
    ("ucop_to_be_allocated", [["TO", "BE", "ALLOCATED"], ["ALLOCATED"]]),
    ("capital_debt_service", [["DEBT", "SERVICE"], ["SERVICE"]]),
    ("off_campus_programs", [["OFF-CAMPUS"], ["OFF", "CAMPUS", "PROGRAMS"], ["PROGRAMS"]]),
    ("system_total", [["SYSTEM", "TOTAL"]]),
]


def _find_phrase_x(words: list[dict], tokens: list[str]) -> Optional[float]:
    """Return the right edge x of a phrase (sequence of tokens) in ``words``.

    Returns ``x1`` of the final token so alignment with right-aligned value
    columns is clean (for long headers like ``SAN FRANCISCO`` the phrase center
    is nowhere near the value column).
    """
    upper_tokens = [t.upper() for t in tokens]
    for i, w in enumerate(words):
        if w["text"].upper() != upper_tokens[0]:
            continue
        matched = [w]
        y_ref = w["top"]
        cur_x1 = w["x1"]
        ok = True
        for t in upper_tokens[1:]:
            nxt = next(
                (
                    ww
                    for ww in words
                    if ww["text"].upper() == t
                    and abs(ww["top"] - y_ref) < 8
                    and ww["x0"] >= cur_x1
                    and ww["x0"] - cur_x1 < 120
                ),
                None,
            )
            if not nxt:
                ok = False
                break
            matched.append(nxt)
            cur_x1 = nxt["x1"]
        if ok:
            return matched[-1]["x1"]
    return None


def find_column_positions(page) -> dict[str, float]:
    """Locate canonical column axes for this page.

    Numbers in these tables are right-aligned, so each column has a stable
    **right edge** (``x1``). We cluster the ``x1`` of every numeric token in the
    body to find those right edges, then pair them in reading order with the
    header labels we can identify. The returned x values are column right edges
    (not centers) — cell assignment uses the same coordinate.
    """
    words = page.extract_words(use_text_flow=False, keep_blank_chars=False)
    anchor = next((w for w in words if w["text"].upper() == "BERKELEY"), None)
    if not anchor:
        return {}
    y_top, y_bot = anchor["top"] - 40, anchor["bottom"] + 40
    header_words = [w for w in words if y_top <= w["top"] <= y_bot]

    # Label positions (for naming only); use phrase center.
    label_xs: list[tuple[str, float]] = []
    for key, variants in COLUMN_ANCHORS:
        for tokens in variants:
            x = _find_phrase_x(header_words, tokens)
            if x is not None:
                label_xs.append((key, x))
                break
    label_xs.sort(key=lambda kv: kv[1])

    # Column right edges from body numeric tokens.
    body_top = anchor["bottom"] + 12
    body_words = _merge_number_fragments(
        [w for w in words if w["top"] > body_top]
    )
    numeric_x1s = sorted(
        w["x1"]
        for w in body_words
        if NUMERIC_TOKEN_RE.match(w["text"]) and not w["text"].startswith(",")
    )
    # Keep only clusters with >=3 samples — filters out footnote markers and
    # other one-off digits that happen to sit in label territory.
    edges = _cluster_values(numeric_x1s, tolerance=6.0, min_count=3)
    if not edges:
        return {}

    # Drop edges that fall clearly left of any known column label — those are
    # artifacts from numbers that appear inside labels (e.g. "2012-13").
    if label_xs:
        min_label_x = label_xs[0][1] - 30.0
        edges = [e for e in edges if e >= min_label_x]

    return _pair_labels_to_edges(label_xs, edges)


def _cluster_values(
    xs: list[float], tolerance: float, min_count: int = 1
) -> list[float]:
    """Cluster a sorted list of values into groups within ``tolerance``; return means.

    Clusters with fewer than ``min_count`` members are discarded.
    """
    if not xs:
        return []
    groups: list[list[float]] = [[xs[0]]]
    for x in xs[1:]:
        if abs(x - groups[-1][-1]) <= tolerance:
            groups[-1].append(x)
        else:
            groups.append([x])
    return [sum(g) / len(g) for g in groups if len(g) >= min_count]


SKIP_LABEL_PENALTY = 1e6  # strongly prefer pairing labels if possible
SKIP_EDGE_PENALTY = 1.0  # mild preference to pair edges


def _pair_labels_to_edges(
    labels: list[tuple[str, float]], edges: list[float]
) -> dict[str, float]:
    """Monotonically align ordered labels with ordered right-edge clusters.

    Needleman-Wunsch-style alignment: at each cell we either pair, skip a
    label, or skip an edge. Pairing cost is the x-distance between the label
    center and the edge; skipping incurs fixed penalties tuned so the
    alignment prefers to pair labels and only drops extras.
    """
    if not labels or not edges:
        return {}
    n, m = len(labels), len(edges)
    INF = float("inf")
    dp = [[INF] * (m + 1) for _ in range(n + 1)]
    back: list[list[Optional[tuple]]] = [[None] * (m + 1) for _ in range(n + 1)]
    dp[0][0] = 0.0
    for j in range(1, m + 1):
        dp[0][j] = dp[0][j - 1] + SKIP_EDGE_PENALTY
        back[0][j] = ("skip_edge",)
    for i in range(1, n + 1):
        dp[i][0] = dp[i - 1][0] + SKIP_LABEL_PENALTY
        back[i][0] = ("skip_label",)
    for i in range(1, n + 1):
        for j in range(1, m + 1):
            pair_cost = dp[i - 1][j - 1] + abs(labels[i - 1][1] - edges[j - 1])
            skip_label = dp[i - 1][j] + SKIP_LABEL_PENALTY
            skip_edge = dp[i][j - 1] + SKIP_EDGE_PENALTY
            best = min(pair_cost, skip_label, skip_edge)
            dp[i][j] = best
            if best == pair_cost:
                back[i][j] = ("pair", j - 1)
            elif best == skip_label:
                back[i][j] = ("skip_label",)
            else:
                back[i][j] = ("skip_edge",)
    assignments: dict[str, float] = {}
    i, j = n, m
    while i > 0 or j > 0:
        op = back[i][j]
        if op is None:
            break
        tag = op[0]
        if tag == "pair":
            assignments[labels[i - 1][0]] = edges[op[1]]
            i -= 1
            j -= 1
        elif tag == "skip_label":
            i -= 1
        else:
            j -= 1
    return assignments


def _row_tolerance(font_size: float) -> float:
    return max(1.5, font_size * 0.5)


_NUM_CONT_RE = re.compile(r"^[\d,]+(?:\.\d+)?$")


def _merge_number_fragments(words: list[dict]) -> list[dict]:
    """Merge split numeric tokens that were emitted as two adjacent words.

    pdfplumber occasionally splits a formatted number such as ``1,202,000`` into
    ``1`` / ``,202,000`` or ``16,128,000`` into ``1`` / ``6,128,000``. We merge
    any two tokens that are on the same line, touching in x, and whose texts
    are both numeric continuations (digits / commas, optional decimal).
    """
    if not words:
        return words
    out: list[dict] = []
    for w in words:
        text = w["text"]
        if out and abs(w["top"] - out[-1]["top"]) < 2 and (w["x0"] - out[-1]["x1"]) < 2:
            prev_text = out[-1]["text"]
            if _NUM_CONT_RE.match(prev_text) and _NUM_CONT_RE.match(text):
                merged = dict(out[-1])
                merged["text"] = prev_text + text
                merged["x1"] = w["x1"]
                out[-1] = merged
                continue
        out.append(w)
    return out


def _cluster_rows(words: list[dict]) -> list[list[dict]]:
    """Cluster words into rows by `top` coordinate, tolerant of small drift."""
    if not words:
        return []
    words = sorted(words, key=lambda w: (round(w["top"], 1), w["x0"]))
    rows: list[list[dict]] = []
    current: list[dict] = []
    current_top: Optional[float] = None
    for w in words:
        size = w.get("height", 10) or 10
        tol = _row_tolerance(size)
        if current_top is None or abs(w["top"] - current_top) <= tol:
            current.append(w)
            current_top = w["top"] if current_top is None else current_top
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
    neg = False
    if t.startswith("(") and t.endswith(")"):
        neg = True
        t = t[1:-1]
    t = t.replace("$", "").replace(",", "").strip()
    if not t or t == "-":
        return 0.0
    try:
        v = float(t)
    except ValueError:
        return None
    return -v if neg else v


def _group_row_cells(
    row_words: list[dict], columns: dict[str, float]
) -> tuple[str, dict[str, float]]:
    """Split a row into (label, {column: amount})."""
    row_words = sorted(row_words, key=lambda w: w["x0"])
    if not row_words:
        return "", {}

    label_tokens: list[str] = []
    amount_tokens: list[dict] = []
    for w in row_words:
        text = w["text"]
        looks_amount = bool(AMOUNT_TOKEN_RE.match(text))
        if not amount_tokens and not looks_amount:
            label_tokens.append(text)
        else:
            amount_tokens.append(w)

    label = " ".join(label_tokens).strip()

    cells = _extract_cells(amount_tokens)

    amounts: dict[str, float] = {}
    assigned_x: dict[str, float] = {}
    for x_right, text in cells:
        if not text:
            continue
        val = _parse_amount(text)
        if val is None:
            continue
        best_col = min(columns.keys(), key=lambda c: abs(columns[c] - x_right))
        prev_x = assigned_x.get(best_col)
        if prev_x is None or abs(x_right - columns[best_col]) < abs(prev_x - columns[best_col]):
            amounts[best_col] = val
            assigned_x[best_col] = x_right

    return label, amounts


def _extract_cells(amount_tokens: list[dict]) -> list[tuple[float, str]]:
    """Convert raw amount tokens into (x_right, text) cells.

    Each numeric token (or dash) becomes its own cell keyed by its right edge
    ``x1`` — columns are right-aligned, so ``x1`` is the stable column axis. A
    leading ``$`` attaches to the next numeric token.
    """
    cells: list[tuple[float, str]] = []
    pending_prefix = ""
    for w in amount_tokens:
        text = w["text"]
        x1 = w["x1"]
        if text == "$":
            pending_prefix = "$"
            continue
        if text in {"-", "\u2010", "\u2013", "\u2014"}:
            cells.append((x1, "-"))
            pending_prefix = ""
            continue
        if NUMERIC_TOKEN_RE.match(text):
            cells.append((x1, pending_prefix + text))
            pending_prefix = ""
            continue
        pending_prefix = ""
    return cells


def parse_pdf(pdf_path: Path) -> list[dict]:
    """Parse all pages of a PDF into structured rows."""
    out_rows: list[dict] = []
    with pdfplumber.open(str(pdf_path)) as pdf:
        # Calibrate column positions using every page's body tokens combined —
        # single-page calibration misses columns that only have values on the
        # other page (e.g. Off-Campus Programs is sparse).
        cols = find_column_positions_multipage(pdf.pages)
        if not cols:
            log.warning("%s: no column headers detected", pdf_path.name)
            return out_rows

        for page_idx, page in enumerate(pdf.pages, start=1):
            words = page.extract_words(use_text_flow=False, keep_blank_chars=False)
            anchor = next((w for w in words if w["text"].upper() == "BERKELEY"), None)
            if not anchor:
                continue
            body_top = anchor["bottom"] + 12
            body_words = [w for w in words if w["top"] > body_top]
            body_words = _merge_number_fragments(body_words)
            rows = _cluster_rows(body_words)

            for row_words in rows:
                label, amounts = _group_row_cells(row_words, cols)
                if not label and not amounts:
                    continue
                if _looks_like_footer(label):
                    continue
                indent = min((w["x0"] for w in row_words if w["text"]), default=0.0)
                out_rows.append(
                    {
                        "page": page_idx,
                        "label": label,
                        "indent": round(indent, 2),
                        "amounts": amounts,
                    }
                )
    return out_rows


def find_column_positions_multipage(pages) -> dict[str, float]:
    all_numeric_x1s: list[float] = []
    label_xs: list[tuple[str, float]] = []
    seen_labels: set[str] = set()
    for page in pages:
        words = page.extract_words(use_text_flow=False, keep_blank_chars=False)
        anchor = next((w for w in words if w["text"].upper() == "BERKELEY"), None)
        if not anchor:
            continue
        y_top, y_bot = anchor["top"] - 40, anchor["bottom"] + 40
        header_words = [w for w in words if y_top <= w["top"] <= y_bot]
        for key, variants in COLUMN_ANCHORS:
            if key in seen_labels:
                continue
            for tokens in variants:
                x = _find_phrase_x(header_words, tokens)
                if x is not None:
                    label_xs.append((key, x))
                    seen_labels.add(key)
                    break

        body_top = anchor["bottom"] + 12
        body_words = _merge_number_fragments(
            [w for w in words if w["top"] > body_top]
        )
        for w in body_words:
            if NUMERIC_TOKEN_RE.match(w["text"]) and not w["text"].startswith(","):
                all_numeric_x1s.append(w["x1"])

    if not label_xs or not all_numeric_x1s:
        return {}
    label_xs.sort(key=lambda kv: kv[1])
    all_numeric_x1s.sort()
    edges = _cluster_values(all_numeric_x1s, tolerance=6.0, min_count=3)
    if label_xs:
        min_label_x = label_xs[0][1] - 30.0
        edges = [e for e in edges if e >= min_label_x]
    return _pair_labels_to_edges(label_xs, edges)


_FOOTER_HINTS = (
    "ucop budget analysis",
    "rev.",
    "university of california",
    "summary of state general fund allocations",
    "page ",
)


def _looks_like_footer(label: str) -> bool:
    low = label.lower().strip()
    if not low:
        return False
    return any(low.startswith(h) or h in low for h in _FOOTER_HINTS)


# ---------------------------------------------------------------------------
# Output
# ---------------------------------------------------------------------------


def write_outputs(
    doc: DiscoveredDoc,
    pdf_path: Path,
    rows: list[dict],
    parsed_dir: Path,
) -> tuple[Path, Path]:
    parsed_dir.mkdir(parents=True, exist_ok=True)
    payload = {
        "fiscal_year": doc.fiscal_year,
        "title": doc.title,
        "source_url": doc.url,
        "source_filename": pdf_path.name,
        "retrieved_at": datetime.now(timezone.utc).isoformat(timespec="seconds"),
        "columns": CANONICAL_COLUMNS,
        "rows": rows,
    }
    json_path = parsed_dir / f"{doc.fiscal_year}.json"
    json_path.write_text(json.dumps(payload, indent=2, sort_keys=False))

    csv_path = parsed_dir / f"{doc.fiscal_year}.csv"
    with csv_path.open("w", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(["page", "indent", "label", *CANONICAL_COLUMNS])
        for r in rows:
            amt = r.get("amounts", {})
            writer.writerow(
                [r["page"], r["indent"], r["label"], *[amt.get(c, "") for c in CANONICAL_COLUMNS]]
            )
    return json_path, csv_path


def write_catalog(docs: Iterable[DiscoveredDoc], out_path: Path) -> None:
    out_path.parent.mkdir(parents=True, exist_ok=True)
    payload = {
        "retrieved_at": datetime.now(timezone.utc).isoformat(timespec="seconds"),
        "index_url": INDEX_URL,
        "documents": [d.__dict__ for d in docs],
    }
    out_path.write_text(json.dumps(payload, indent=2))


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------


def parse_args(argv: list[str]) -> argparse.Namespace:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--out", type=Path, default=Path("/data"), help="Output base directory")
    p.add_argument("--year", action="append", help="Restrict to specific fiscal year(s) (e.g. 2024-25)")
    p.add_argument("--discover", action="store_true", help="Write the discovery catalog and exit")
    p.add_argument("--download-only", action="store_true", help="Download PDFs but skip parsing")
    p.add_argument("--parse-only", action="store_true", help="Parse already-downloaded PDFs")
    p.add_argument("--all", action="store_true", help="Discover, download, and parse (default)")
    p.add_argument("--force", action="store_true", help="Re-download existing PDFs")
    p.add_argument("--index-url", default=INDEX_URL)
    p.add_argument("-v", "--verbose", action="store_true")
    return p.parse_args(argv)


def main(argv: Optional[list[str]] = None) -> int:
    args = parse_args(list(sys.argv[1:] if argv is None else argv))
    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(asctime)s %(levelname)s %(message)s",
    )

    out_dir: Path = args.out
    pdf_dir = out_dir / "pdfs"
    parsed_dir = out_dir / "parsed"

    # Default behaviour: run all phases.
    do_discover = True
    do_download = not args.parse_only
    do_parse = not args.discover and not args.download_only

    session = make_session()
    docs: list[DiscoveredDoc] = []

    if args.parse_only:
        # Build docs from existing PDF filenames.
        for pdf in sorted(pdf_dir.glob("*.pdf")):
            year = pdf.stem
            docs.append(
                DiscoveredDoc(
                    fiscal_year=year,
                    title=f"{year} Summary of State General Fund Allocations",
                    url="",
                    filename=pdf.name,
                )
            )
    else:
        docs = discover(session, args.index_url)
        write_catalog(docs, out_dir / "catalog.json")

    if args.year:
        wanted = set(args.year)
        docs = [d for d in docs if d.fiscal_year in wanted]
        log.info("Filtered to: %s", ", ".join(d.fiscal_year for d in docs))

    if args.discover:
        return 0

    if do_download:
        for doc in docs:
            try:
                download(session, doc, pdf_dir, force=args.force)
            except requests.HTTPError as e:
                log.error("Failed to download %s: %s", doc.url, e)

    if do_parse:
        for doc in docs:
            pdf_path = pdf_dir / doc.filename
            if not pdf_path.exists():
                log.warning("Missing PDF for %s: %s", doc.fiscal_year, pdf_path)
                continue
            try:
                rows = parse_pdf(pdf_path)
            except Exception:
                log.exception("Parse failed for %s", pdf_path.name)
                continue
            json_path, csv_path = write_outputs(doc, pdf_path, rows, parsed_dir)
            amounts_rows = [r for r in rows if r.get("amounts")]
            log.info(
                "%s parsed: %d rows (%d with amounts) -> %s, %s",
                doc.fiscal_year,
                len(rows),
                len(amounts_rows),
                json_path.name,
                csv_path.name,
            )

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
