# USASpending.gov Enhanced Scraper ✅ ENHANCED VERSION

This Go application scrapes comprehensive federal spending data related to University of California from the USASpending.gov API, including detailed award information and organized file storage.

**Status**: ✅ Enhanced version with detailed data collection and hierarchical organization
**Previous Collection**: 16,760 awards across all federal spending categories (Sept 23, 2025)
**New Features**: Individual detailed award data + organized file structure

## Overview

The enhanced scraper combines two API endpoints:
1. `/api/v2/search/spending_by_award/` - Collects comprehensive award listings
2. `/api/v2/awards/{id}/` - Fetches detailed information for each individual award

## Enhanced Features

- **Two-Phase Data Collection**: 
  - Phase 1: Collect all basic award data (search endpoint)
  - Phase 2: Fetch detailed information for each award (detail endpoint)
- **Complete Field Coverage**: Captures ALL available fields from both API responses
- **Hierarchical Organization**: Organizes awards by `[Type]/[Recipient]/[Year]/[Agency]/[Award_ID].json`
- **Rate Limiting**: Includes 1-second delays between requests to respect API limits
- **Error Resilience**: Continues processing if individual detail fetches fail
- **Comprehensive Details**: Includes contract specifics, agency hierarchies, executive compensation, business categories, and more

## Usage

```bash
# Build the enhanced scraper
go build -o usaspending-enhanced-scraper main.go

# Run the enhanced scraper
./usaspending-enhanced-scraper
```

## Enhanced Output Structure

Instead of single JSON files per award type, awards are now organized hierarchically:

```
../Contracts/
  └── THE_REGENTS_OF_THE_UNIVERSITY_OF_CALIFORNIA/
      └── 2024/
          └── Department_of_Energy/
              └── CONT_AWD_DEAC0205CH11231_8900_-NONE-_-NONE-.json
```

Each JSON file contains both basic search data and detailed award information.

## Data Structure

Each award record includes structured data for:
- **Locations**: Both recipient and performance locations with full address details
- **Classifications**: NAICS and PSC codes with descriptions
- **Financial Data**: Award amounts, total outlays, and special funding categories
- **Metadata**: Internal IDs, generated identifiers, and audit trails

## Search Criteria

The scraper uses the following filters:
- **Keywords**: "University of California"
- **Time Period**: 2007-10-01 to 2025-09-30
- **Award Types**: A, B, C, D (Contracts, Grants, Direct Payments, Loans)
- **Recipient Types**: Various higher education institution types
- **Location**: California (USA)

## API Information

- **Endpoint**: `https://api.usaspending.gov/api/v2/search/spending_by_award/`
- **Method**: POST
- **Rate Limit**: Self-imposed 1 second delay between requests
- **Pagination**: 100 records per page

## Dependencies

Uses only Go standard library:
- `net/http` for API requests
- `encoding/json` for JSON handling
- `context` for request management
- `time` for rate limiting and timestamps

## Notes

- The API limits historical data to 2007-10-01 for this endpoint
- For earlier data, use the bulk download endpoints as suggested by the API
- The scraper includes error handling and retry logic
- All monetary values are returned as numbers from the API