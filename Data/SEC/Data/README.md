Data from the SEC via this link: https://data.sec.gov/submissions/CIK0000315054.json

A 13F Filing takes the form of https://www.sec.gov/Archives/edgar/data/315054/000114036124024512/xslForm13F_X02/informationtable.xml

315054 is our CIK or the Univesty of Cali Regents
000114036124024512 is the filing.assesionnumber



Form 13F filings: Quarterly reports of direct stock holdings
Schedule 13G filings: For significant ownership stakes (5%+ ownership)
Note: Recent filings show UC has reclassified many holdings, claiming lack of "discretionary authority"

## Historical Form 13F import

The historical import is proposed in [pull request #2](https://github.com/ucinvestments/UC-Finances/pull/2).
Its 48 CSVs omit two early 2013 filings listed in the checked-in submission
index. The parser uses an absolute input path and prints legacy table
sections without exporting them. The CSV values are in the reporting units
of each individual filing. The SEC switched Form 13F from thousands of
dollars to whole dollars on January 3, 2023.

Before integrating the import, give the parser a caller-supplied input
and output path, record the reporting unit with each filing, and reconcile
coverage, line counts, and value totals against EDGAR. Check for reported
rounding differences in the cover totals. The November 2025 sample has
seven CSV rows that match the embedded XML and sum to one dollar below its
filing cover total.

[SEC Form 13F reporting guidance](https://www.sec.gov/rules-regulations/staff-guidance/frequently-asked-questions-about-form-13f)

