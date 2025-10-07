# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This repository contains SEC filing data for analyzing institutional holdings, specifically tracking the University of California Regents (CIK: 0000315054). It's a data repository focused on 13F-HR filings and related SEC documents.

## Data Structure

### Key Files
- `SEC/Data/CIK0000315054.json` - Complete SEC submission data containing 25+ years of filings
- `SEC/Data/README.md` - Data source documentation

### SEC Data Format
The JSON file contains comprehensive filing metadata including:
- 13F-HR quarterly holdings reports (primary focus)
- Schedule 13G/13G-A ownership disclosures
- Form 4 insider transactions
- Historical data from 1999 to present

### Data Access Patterns
- SEC source: `https://data.sec.gov/submissions/CIK0000315054.json`
- 13F filing URLs: `https://www.sec.gov/Archives/edgar/data/{CIK}/{accession-number}/xslForm13F_X02/informationtable.xml`
- Quarterly reporting cycle (Q1: Mar 31, Q2: Jun 30, Q3: Sep 30, Q4: Dec 31)

## Repository Architecture

This is a data-only repository with no build system, dependencies, or executable code. The structure is:
```
SEC/Data/         # SEC filing data and metadata
LICENSE          # MIT license
```

No package managers, test frameworks, or build tools are present.