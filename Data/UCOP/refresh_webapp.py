"""Validate parsed UCOP reports and synchronize the explorer's data files."""

import argparse
import json
import math
from pathlib import Path
import re
from tempfile import NamedTemporaryFile
from urllib.parse import urlsplit


ROOT = Path(__file__).resolve().parent
DESTINATION = ROOT.parents[1] / "Webapp/static/data/ucop"
SERIES = (
    ("general_fund_allocations", "general_fund_allocations", "parsed", "fiscal_year", "fiscal_years", r"\d{4}-\d{2}", "columns"),
    ("retrends", "reporting_transparency", "parsed/retrends", "year_key", "years", r"\d{4}", "year_columns"),
)


def read_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def validate_report(report: dict, columns_key: str) -> None:
    columns = report[columns_key]
    if not isinstance(columns, list) or not columns or not all(isinstance(column, str) for column in columns):
        raise ValueError(f"invalid {columns_key}")
    rows = report["rows"]
    if not isinstance(rows, list) or not any(row["amounts"] for row in rows):
        raise ValueError("no rows with amounts")
    column_names = set(columns)
    for row in rows:
        if not isinstance(row["label"], str):
            raise ValueError("invalid row label")
        amounts = row["amounts"]
        if not isinstance(amounts, dict) or not amounts.keys() <= column_names:
            raise ValueError(f"invalid columns on row {row['label']!r}")
        if any(type(value) not in (int, float) or not math.isfinite(value) for value in amounts.values()):
            raise ValueError(f"invalid amount on row {row['label']!r}")


def encode_json(document: dict) -> bytes:
    return json.dumps(document, indent=2, allow_nan=False).encode()


def build_distribution(destination: Path) -> dict[Path, bytes]:
    files = {}
    manifest = {}
    manifest_path = destination / "manifest.json"
    previous = read_json(manifest_path) if manifest_path.exists() else {}
    for series, source, parsed, period_key, periods_key, period_pattern, columns_key in SERIES:
        root = ROOT / source
        catalog = {
            entry[period_key]: entry["url"]
            for entry in read_json(root / "catalog.json")["documents"]
            if entry.get("series", series) == series
        }
        periods = []
        for path in sorted((root / parsed).glob("*.json")):
            try:
                period = path.stem
                report = read_json(path)
                if not re.fullmatch(period_pattern, period) or report[period_key] != period:
                    raise ValueError("report period does not match filename")
                validate_report(report, columns_key)
                report["source_url"] = report.get("source_url") or catalog.get(period, "")
                url = urlsplit(report["source_url"])
                if url.scheme != "https" or not url.hostname:
                    raise ValueError("missing HTTPS source URL")
            except (KeyError, TypeError, ValueError) as exc:
                raise ValueError(f"{path}: {exc}") from exc
            periods.append(period)
            files[destination / series / path.name] = encode_json(report)

        missing = (set(catalog) | set(previous.get(series, {}).get(periods_key, []))) - set(periods)
        if missing or not periods:
            raise ValueError(f"{series}: missing parsed reports for {', '.join(sorted(missing)) or 'all periods'}")
        manifest[series] = {"path": f"/data/ucop/{series}", periods_key: periods}
    files[manifest_path] = encode_json(manifest) + b"\n"
    return files


def replace_file(path: Path, contents: bytes) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with NamedTemporaryFile(dir=path.parent, delete=False) as temporary:
        staged = Path(temporary.name)
        try:
            temporary.write(contents)
            temporary.close()
            staged.chmod(0o644)
            staged.replace(path)
        finally:
            staged.unlink(missing_ok=True)


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="fail if site data needs refreshing")
    args = parser.parse_args()

    try:
        files = build_distribution(DESTINATION)
        changed = [path for path, contents in files.items() if not path.exists() or path.read_bytes() != contents]
        if args.check and changed:
            parser.exit(1, "Out of date:\n" + "\n".join(str(path.relative_to(DESTINATION)) for path in changed) + "\n")
        for path in changed:
            replace_file(path, files[path])
    except (OSError, KeyError, TypeError, ValueError) as exc:
        parser.exit(1, f"UCOP refresh failed: {exc}\n")
    print(f"Validated {len(files) - 1} reports. Updated {len(changed)} files.")


if __name__ == "__main__":
    main()
