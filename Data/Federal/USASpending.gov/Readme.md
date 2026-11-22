# USAspending award snapshot

The September 2025 collection used the keyword “University of California” and
filtered awards by California performance location. The archive includes
awards to recipients outside the UC system, including Dominican University
of California and San Diego State University.

Run the snapshot audit from the repository root:

```sh
python3 Data/Federal/USASpending.gov/Scraping/audit_snapshot.py
```

The audit checks for missing, conflicting, and duplicate generated award IDs.
It also reports award counts by category and inventories recipient names and
UEIs for review. The checked-in snapshot has 16,207 unique awards, including
9,422 grants. These are counts of the available award files.

To produce a UC dataset, verify recipient and parent identifiers against UC
entities, then query the approved recipients with cursor pagination. Verify
coverage and decide how to account for awards with multiple locations. The
previous grants search reported exactly 10,000 results. A follow-up using
the USAspending award-count endpoint with the same filters returned 50,047
grant matches. This is a count for the keyword search. The archive's legacy
folders also group direct payments and other assistance differently from
the current API summary. Use the source award type codes when rebuilding the
dataset. Keep obligations, outlays, and loan amounts distinct when
calculating totals.

[Open tracking issue](https://github.com/ucinvestments/UC-Finances/issues/1)
· [USAspending award-count API](https://github.com/fedspendingtransparency/usaspending-api/blob/master/usaspending_api/api_contracts/contracts/v2/search/spending_by_award_count.md)
