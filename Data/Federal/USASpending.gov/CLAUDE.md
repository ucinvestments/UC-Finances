# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This directory contains a Go application for scraping federal spending data related to University of California from the USASpending.gov API. The scraper automatically collects comprehensive award data across multiple federal spending categories.

## Development Commands

### Enhanced Go Application (Scraping/)
```bash
# Navigate to scraper directory
cd Scraping/

# Build the enhanced scraper
go build -o usaspending-enhanced-scraper main.go

# Run the enhanced scraper
./usaspending-enhanced-scraper

# Alternatively, run directly
go run main.go

# Legacy: Build original scraper (for reference)
# Note: main.go now contains enhanced version
```

## Architecture Overview

### Enhanced Scraper Design
The enhanced Go application uses a two-phase approach to collect comprehensive data:

**Phase 1: Basic Data Collection**
- **Award Type Groups**: Contracts, Grants, Loans, IDVs, Direct Payments, Other Financial Assistance
- **Rate Limiting**: 1-second delays between API requests to be respectful
- **Pagination**: Automatically handles API pagination (100 records per page)
- **Complete Field Capture**: Now captures ALL fields from search response including `awarding_agency_id`, `agency_slug`, etc.

**Phase 2: Detailed Data Enhancement**
- **Individual Award Details**: Fetches comprehensive data for each award using detail endpoint
- **Error Resilience**: Continues processing if individual detail fetches fail
- **Hierarchical Organization**: Organizes files by recipient, year, and agency

### API Integration
- **Primary Endpoint**: `https://api.usaspending.gov/api/v2/search/spending_by_award/` (POST)
- **Detail Endpoint**: `https://api.usaspending.gov/api/v2/awards/{generated_internal_id}/` (GET)
- **Search Criteria**: "University of California" keywords, 2007-2025 time range, California location
- **Field Configuration**: Comprehensive field mapping for both basic and detailed responses

### Enhanced Data Structure
Each award now includes both basic and detailed information:
- **Basic Data**: All fields from search endpoint (award IDs, amounts, basic recipient info)
- **Detailed Data**: Contract specifics, agency hierarchies, executive compensation, business categories
- **Combined Structure**: JSON structure with `basic_data` and `detailed_data` sections

### Hierarchical Output Organization
Awards are organized in searchable directory structure:
```
[Award Type]/[Recipient Name]/[Year]/[Awarding Agency]/[Award ID].json
```

Example structure:
- `../Contracts/THE_REGENTS_OF_THE_UNIVERSITY_OF_CALIFORNIA/2024/Department_of_Energy/`
- `../Grants/THE_REGENTS_OF_THE_UNIVERSITY_OF_CALIFORNIA/2023/National_Science_Foundation/`

### Dependencies
Uses only Go standard library:
- `net/http` for API requests
- `encoding/json` for JSON handling
- `context` for request management
- `time` for rate limiting and timestamps

## Important Notes

- The scraper includes error handling and retry logic
- API data is limited to 2007-10-01 onwards for this endpoint
- Monetary values are returned as numbers from the API
- Different award types have different field configurations and sort parameters
- The application respects USASpending.gov API rate limits with built-in delays