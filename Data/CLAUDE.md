# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This repository contains data and documentation related to University of California (UC) investments and holdings, including:
- UC Investments Office holdings data (GEP and UCRP)
- SEC filings data (Form 13F, Schedule 13G)
- Federal audit and spending data from USASpending.gov and Federal Audit Clearinghouse
- Campus foundation information
- UC Annual Reports

## Repository Structure

```
Data/
├── UC_Investments_Office/     # UC Investment holdings data
│   ├── GEP/                  # General Endowment Pool holdings
│   └── UCRP/                 # UC Retirement Plan holdings
├── SEC/Data/                 # SEC EDGAR filings and data
├── Federal/                  # Federal government data
│   ├── USASpending.gov/      # Federal contracts, grants, loans data
│   │   ├── Contract_IDVs/
│   │   ├── Contracts/
│   │   ├── Direct_Payments/
│   │   ├── Grants/
│   │   ├── Loans/
│   │   └── Other/
│   └── Federal_Audit_Clearinghouse/  # Annual audit reports by year
├── Campus_Foundation/         # Individual UC campus foundation data
└── UC_Annual_Reports/        # Annual investment reports
```

## Key Data Sources

### SEC Data
- CIK Number for UC Regents: 0000315054
- Form 13F filing URL pattern: `https://www.sec.gov/Archives/edgar/data/315054/{accession_number}/xslForm13F_X02/informationtable.xml`
- Main data JSON: `https://data.sec.gov/submissions/CIK0000315054.json`

### USASpending.gov
- General search: https://www.usaspending.gov/search?hash=5cd0c7a1f68599eff5bb36887e18ad09
- Specific search: https://www.usaspending.gov/search?hash=f5e03f4e69c5cdf58b1da88ee250f04f

### Federal Audit Clearinghouse
- Example 2024 report: https://app.fac.gov/dissemination/summary/2024-06-GSAFAC-0000356436

## Data Processing Notes

- The repository primarily contains raw data downloads and documentation
- Data is organized by source and year where applicable
- Most data files are in Excel (.xlsx), JSON, or CSV format
- No automated processing scripts or build tools are currently present

## Important Context

- UC removed their sponsored projects dashboard in late 2024, eliminating historical contract data
- Form 13F filings show UC reclassified many holdings in Q2 2024, claiming lack of discretionary authority
- Blue & Gold Endowment holdings are NOT disclosed despite policy requiring it
- Private equity/credit holdings are largely exempt from disclosure