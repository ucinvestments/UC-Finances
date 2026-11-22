"""Check archived award identities and inventory recipients for UC verification."""

from collections import Counter
import json
from pathlib import Path


def main() -> int:
    archive = Path(__file__).resolve().parents[1]
    award_paths = {}
    categories = Counter()
    recipients = Counter()
    errors = []

    for path in sorted(archive.glob("*/*/*/*/*.json")):
        try:
            record = json.loads(path.read_text(encoding="utf-8"))
            basic = record["basic_data"]
            detail = record["detailed_data"]
            award_id = basic["generated_internal_id"]
            if not award_id or award_id != detail["generated_unique_award_id"]:
                raise ValueError("missing or inconsistent generated award ID")
            recipient = basic["Recipient Name"]
            if not recipient:
                raise ValueError("missing recipient name")
            if award_id in award_paths:
                raise ValueError(f"duplicate award ID already recorded in {award_paths[award_id]}")
            relative = path.relative_to(archive)
            award_paths[award_id] = str(relative)
            categories[relative.parts[0]] += 1
            recipients[recipient, basic.get("Recipient UEI") or ""] += 1
        except (OSError, KeyError, TypeError, ValueError) as exc:
            errors.append({"file": str(path.relative_to(archive)), "error": str(exc)})

    if not award_paths:
        errors.append({"error": "no valid archived awards"})
    print(json.dumps({
        "unique_awards": len(award_paths),
        "categories": dict(sorted(categories.items())),
        "recipients": [
            {"name": name, "uei": uei or None, "awards": count}
            for (name, uei), count in sorted(recipients.items())
        ],
        "errors": errors,
    }, indent=2))
    return int(bool(errors))


if __name__ == "__main__":
    raise SystemExit(main())
