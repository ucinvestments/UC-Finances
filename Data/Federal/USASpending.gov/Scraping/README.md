# USASpending.gov Scraper

This Go application scrapes federal spending data related to University of California from the USASpending.gov API.

## Overview

The scraper targets the `/api/v2/search/spending_by_award/` endpoint to collect comprehensive award data for University of California institutions from 2007 to present.

## Features

- **Complete Data Collection**: Automatically paginates through all results
- **Rate Limiting**: Includes 1-second delays between requests to be respectful to the API
- **Comprehensive Fields**: Collects all available award metadata including:
  - Award details (ID, amount, type, description)
  - Recipient information (name, UEI, location)
  - Performance location data
  - Agency and sub-agency information
  - NAICS and PSC classification codes
  - COVID-19 and Infrastructure funding flags
  - Timeline information (start/end dates)

## Usage

```bash
# Build the scraper
go build -o usaspending-scraper main.go

# Run the scraper
./usaspending-scraper
```

## Output

The scraper creates a JSON file named `uc_awards_YYYY-MM-DD.json` containing all collected award data.

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