# USAspending awards

## Awards for verified UC recipients

The verified collector uses the UEI and legacy DUNS numbers published by
UCOP for the ten campuses and the Office of the President. It searches
USAspending by identifier, then requires an exact match on the returned
recipient UEI or DUNS before including an award. Search results without a
confirmed identifier are listed separately by award ID.

Run from the scraper directory, choosing a new output file outside the
historical archive:

```sh
cd Data/Federal/USASpending.gov/Scraping
go run . -verified-output /tmp/uc-verified-awards.json
```

The default window is October 1, 2007 through September 30, 2025. Use
`-start` and `-end` for a shorter collection. The collector checks the number
of fetched rows against the USAspending count endpoint for each award type,
rejects duplicate generated award IDs, and writes the output only after
those checks pass. It records the search window and recipient roster in the
output. The search covers all performance locations and reports prime
awards. Departmental identifiers, distinct affiliates and laboratories,
and subawards require separate scope and verification. Treat the output as
coverage of the listed recipients.

The full historical run reconciled all six award categories against
USAspending. Of 56,724 search matches, 56,722 had a verified recipient ID.
The two remaining contracts used an alternate UC San Francisco registration
and are listed in the export under `unmatched_award_ids`.

[UCOP recipient identifiers](https://www.ucop.edu/research-policy-analysis-coordination/resources-tools/contract-and-grant-manual/chapter10/index.html#10-530)
· [USAspending award categories](https://github.com/fedspendingtransparency/usaspending-api/blob/master/usaspending_api/awards/v2/lookups/lookups.py)

### Additional UC recipient coverage

Two separately registered UC units have identifiers published by their
sponsored-projects offices. Both fall outside the default roster.

| Unit | UEI | Legacy DUNS | Source |
| --- | --- | --- | --- |
| Scripps Institution of Oceanography, non-NIH proposals | `QJ8HMDK7MRM3` | `175104595` | [UC San Diego institutional information](https://blink.ucsd.edu/research/spo/preparing-proposals/institutional-information-codes.html) |
| UC Agriculture and Natural Resources | `K5KAMCPRVED6` | `604591925` | [UC ANR institutional information](https://ucanr.edu/dept/office-contracts-and-grants/institutional-information) |

Check the recorded recipient and reporting entity before expanding the
roster to health systems, laboratory contractors, or affiliated entities.

[First UCSF contract](https://www.usaspending.gov/award/CONT_AWD_FA486118PB060_9700_-NONE-_-NONE-)
· [Second UCSF contract](https://www.usaspending.gov/award/CONT_AWD_VA26117P3176_3600_-NONE-_-NONE-)

## Historical keyword snapshot

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

The previous keyword-based grants search reported exactly 10,000 results. A follow-up using
the USAspending award-count endpoint with the same filters returned 50,047
grant matches. This is a count for the keyword search. The archive's legacy
folders also group direct payments and other assistance differently from
the current API summary. Use the source award type codes when rebuilding the
dataset. Keep obligations, outlays, loan amounts, and IDV values distinct
when calculating totals.

[Open tracking issue](https://github.com/ucinvestments/UC-Finances/issues/1)
· [USAspending award-count API](https://github.com/fedspendingtransparency/usaspending-api/blob/master/usaspending_api/api_contracts/contracts/v2/search/spending_by_award_count.md)
