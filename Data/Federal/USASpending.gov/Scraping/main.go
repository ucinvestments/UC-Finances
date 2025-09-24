package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// API Request Structures
type TimePeriod struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type PlaceOfPerformance struct {
	Country string `json:"country"`
	State   string `json:"state"`
}

type Filters struct {
	Keywords                       []string             `json:"keywords"`
	TimePeriod                     []TimePeriod         `json:"time_period"`
	AwardTypeCodes                 []string             `json:"award_type_codes"`
	RecipientTypeNames             []string             `json:"recipient_type_names"`
	PlaceOfPerformanceLocations    []PlaceOfPerformance `json:"place_of_performance_locations"`
}

type APIRequest struct {
	Filters      Filters  `json:"filters"`
	Page         int      `json:"page"`
	Limit        int      `json:"limit"`
	Sort         string   `json:"sort"`
	Order        string   `json:"order"`
	AuditTrail   string   `json:"auditTrail"`
	Fields       []string `json:"fields"`
	SpendingLevel string  `json:"spending_level"`
}

// API Response Structures
type PageMetadata struct {
	Page                   int    `json:"page"`
	HasNext                bool   `json:"hasNext"`
	LastRecordUniqueID     int    `json:"last_record_unique_id"`
	LastRecordSortValue    string `json:"last_record_sort_value"`
}

type Location struct {
	LocationCountryCode   string  `json:"location_country_code"`
	CountryName           string  `json:"country_name"`
	StateCode             string  `json:"state_code"`
	StateName             string  `json:"state_name"`
	CityName              string  `json:"city_name"`
	CountyCode            string  `json:"county_code"`
	CountyName            string  `json:"county_name"`
	AddressLine1          string  `json:"address_line1"`
	AddressLine2          *string `json:"address_line2"`
	AddressLine3          *string `json:"address_line3"`
	CongressionalCode     string  `json:"congressional_code"`
	Zip4                  string  `json:"zip4"`
	Zip5                  string  `json:"zip5"`
	ForeignPostalCode     *string `json:"foreign_postal_code"`
	ForeignProvince       *string `json:"foreign_province"`
}

type CodeDescription struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type Award struct {
	InternalID                        int         `json:"internal_id"`
	AwardID                          string      `json:"Award ID"`
	RecipientName                    string      `json:"Recipient Name"`
	AwardAmount                      interface{} `json:"Award Amount"`
	TotalOutlays                     interface{} `json:"Total Outlays"`
	Description                      string      `json:"Description"`
	ContractAwardType                string          `json:"Contract Award Type"`
	RecipientUEI                     string          `json:"Recipient UEI"`
	RecipientLocation                interface{}     `json:"Recipient Location"` // Can be Location object or string fields
	PrimaryPlaceOfPerformance        interface{}     `json:"Primary Place of Performance"` // Can be Location object or string fields
	DefCodes                         []string        `json:"def_codes"`
	COVID19Obligations               interface{}     `json:"COVID-19 Obligations"`
	COVID19Outlays                   interface{}     `json:"COVID-19 Outlays"`
	InfrastructureObligations        interface{}     `json:"Infrastructure Obligations"`
	InfrastructureOutlays            interface{}     `json:"Infrastructure Outlays"`
	AwardingAgency                   string          `json:"Awarding Agency"`
	AwardingSubAgency                string          `json:"Awarding Sub Agency"`
	StartDate                        string          `json:"Start Date"`
	EndDate                          string          `json:"End Date"`
	NAICS                            interface{}     `json:"NAICS"` // Can be CodeDescription or string
	PSC                              interface{}     `json:"PSC"`   // Can be CodeDescription or string
	RecipientID                      string          `json:"recipient_id"`
	PrimeAwardRecipientID            string          `json:"prime_award_recipient_id"`
	GeneratedInternalID              string          `json:"generated_internal_id"`

	// Loan-specific fields
	LoanValue                        interface{} `json:"Loan Value"`
	SubsidyCost                      interface{} `json:"Subsidy Cost"`
	IssuedDate                       string      `json:"Issued Date"`
	FundingAgency                    string      `json:"Funding Agency"`

	// Location fields for loans (separate from Location object)
	RecipientLocationCityName        string `json:"recipient_location_city_name"`
	RecipientLocationStateCode       string `json:"recipient_location_state_code"`
	RecipientLocationCountryName     string `json:"recipient_location_country_name"`
	RecipientLocationAddressLine1    string `json:"recipient_location_address_line1"`
	POPCityName                      string `json:"pop_city_name"`
	POPStateCode                     string `json:"pop_state_code"`
	POPCountryName                   string `json:"pop_country_name"`
}

type APIResponse struct {
	Results      []Award      `json:"results"`
	PageMetadata PageMetadata `json:"page_metadata"`
}

type Scraper struct {
	client  *http.Client
	baseURL string
	delay   time.Duration
}

func NewScraper() *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.usaspending.gov/api/v2/search/spending_by_award/",
		delay:   1 * time.Second, // Be respectful to the API
	}
}

// Award type groups as defined by the API
var awardTypeGroups = map[string][]string{
	"contracts":                  {"A", "B", "C", "D"},
	"grants":                     {"02", "03", "04", "05"},
	"loans":                      {"07", "08"},
	"idvs":                       {"IDV_A", "IDV_B", "IDV_B_A", "IDV_B_B", "IDV_B_C", "IDV_C", "IDV_D", "IDV_E"},
	"other_financial_assistance": {"06", "10"},
	"direct_payments":            {"09", "11"},
}

// Directory mapping for each award type group - save locally for now due to permissions
var directoryMapping = map[string]string{
	"contracts":                  ".",
	"grants":                     ".",
	"loans":                      ".",
	"idvs":                       ".",
	"other_financial_assistance": ".",
	"direct_payments":            ".",
}

// Award type specific configurations
var awardTypeConfigs = map[string]struct {
	SortField string
	Fields    []string
}{
	"contracts": {
		SortField: "Award Amount",
		Fields: []string{
			"Award ID", "Recipient Name", "Award Amount", "Total Outlays", "Description",
			"Contract Award Type", "Recipient UEI", "Recipient Location", "Primary Place of Performance",
			"def_codes", "COVID-19 Obligations", "COVID-19 Outlays", "Infrastructure Obligations",
			"Infrastructure Outlays", "Awarding Agency", "Awarding Sub Agency", "Start Date", "End Date",
			"NAICS", "PSC", "recipient_id", "prime_award_recipient_id",
		},
	},
	"grants": {
		SortField: "Award Amount",
		Fields: []string{
			"Award ID", "Recipient Name", "Award Amount", "Total Outlays", "Description",
			"Recipient UEI", "Recipient Location", "Primary Place of Performance", "def_codes",
			"COVID-19 Obligations", "COVID-19 Outlays", "Infrastructure Obligations", "Infrastructure Outlays",
			"Awarding Agency", "Awarding Sub Agency", "Start Date", "End Date", "recipient_id", "prime_award_recipient_id",
		},
	},
	"loans": {
		SortField: "Loan Value",
		Fields: []string{
			"Award ID", "Recipient Name", "Loan Value", "Subsidy Cost", "Description", "Recipient UEI",
			"recipient_location_city_name", "recipient_location_state_code", "recipient_location_country_name",
			"recipient_location_address_line1", "pop_city_name", "pop_state_code", "pop_country_name",
			"def_codes", "COVID-19 Obligations", "COVID-19 Outlays", "Infrastructure Obligations", "Infrastructure Outlays",
			"Awarding Agency", "Funding Agency", "Issued Date", "recipient_id", "prime_award_recipient_id",
		},
	},
	"idvs": {
		SortField: "Award Amount",
		Fields: []string{
			"Award ID", "Recipient Name", "Award Amount", "Total Outlays", "Description",
			"Contract Award Type", "Recipient UEI", "Recipient Location", "Primary Place of Performance",
			"def_codes", "COVID-19 Obligations", "COVID-19 Outlays", "Infrastructure Obligations",
			"Infrastructure Outlays", "Awarding Agency", "Awarding Sub Agency", "Start Date", "End Date",
			"NAICS", "PSC", "recipient_id", "prime_award_recipient_id",
		},
	},
	"other_financial_assistance": {
		SortField: "Award Amount",
		Fields: []string{
			"Award ID", "Recipient Name", "Award Amount", "Total Outlays", "Description", "Recipient UEI",
			"Recipient Location", "Primary Place of Performance", "def_codes", "COVID-19 Obligations",
			"COVID-19 Outlays", "Infrastructure Obligations", "Infrastructure Outlays", "Awarding Agency",
			"Awarding Sub Agency", "Start Date", "End Date", "recipient_id", "prime_award_recipient_id",
		},
	},
	"direct_payments": {
		SortField: "Award Amount",
		Fields: []string{
			"Award ID", "Recipient Name", "Award Amount", "Total Outlays", "Description", "Recipient UEI",
			"Recipient Location", "Primary Place of Performance", "def_codes", "COVID-19 Obligations",
			"COVID-19 Outlays", "Infrastructure Obligations", "Infrastructure Outlays", "Awarding Agency",
			"Awarding Sub Agency", "Start Date", "End Date", "recipient_id", "prime_award_recipient_id",
		},
	},
}

func (s *Scraper) createRequest(groupName string, awardTypeCodes []string) APIRequest {
	config := awardTypeConfigs[groupName]

	return APIRequest{
		Filters: Filters{
			Keywords: []string{"University of California"},
			TimePeriod: []TimePeriod{
				{
					StartDate: "2007-10-01",
					EndDate:   "2025-09-30",
				},
			},
			AwardTypeCodes: awardTypeCodes,
			RecipientTypeNames: []string{
				"higher_education",
				"public_institution_of_higher_education",
				"private_institution_of_higher_education",
				"minority_serving_institution_of_higher_education",
				"school_of_forestry",
				"veterinary_college",
				"government",
			},
			PlaceOfPerformanceLocations: []PlaceOfPerformance{
				{
					Country: "USA",
					State:   "CA",
				},
			},
		},
		Page:          1,
		Limit:         100,
		Sort:          config.SortField,
		Order:         "desc",
		AuditTrail:    "Results Table - Spending by award search",
		Fields:        config.Fields,
		SpendingLevel: "awards",
	}
}

func (s *Scraper) makeRequest(ctx context.Context, request APIRequest) (*APIResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "UC-Holdings-Scraper/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &apiResponse, nil
}

func (s *Scraper) scrapeGroupData(ctx context.Context, groupName string, awardTypeCodes []string) ([]Award, error) {
	var groupAwards []Award
	page := 1

	log.Printf("Starting to scrape %s data (codes: %v)...", groupName, awardTypeCodes)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		log.Printf("[%s] Fetching page %d...", groupName, page)

		request := s.createRequest(groupName, awardTypeCodes)
		request.Page = page

		response, err := s.makeRequest(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("error fetching %s page %d: %w", groupName, page, err)
		}

		groupAwards = append(groupAwards, response.Results...)

		log.Printf("[%s] Page %d: got %d awards, total so far: %d",
			groupName, page, len(response.Results), len(groupAwards))

		// Check if there are more pages
		if !response.PageMetadata.HasNext || len(response.Results) == 0 {
			log.Printf("[%s] No more pages. Total awards collected: %d", groupName, len(groupAwards))
			break
		}

		page++

		// Be respectful to the API
		time.Sleep(s.delay)
	}

	return groupAwards, nil
}

func (s *Scraper) scrapeAndSaveAllData(ctx context.Context) (int, error) {
	totalAwards := 0
	timestamp := time.Now().Format("2006-01-02")

	log.Printf("Starting to scrape University of California data for all award types...")

	for groupName, awardTypeCodes := range awardTypeGroups {
		groupAwards, err := s.scrapeGroupData(ctx, groupName, awardTypeCodes)
		if err != nil {
			return 0, fmt.Errorf("error scraping %s: %w", groupName, err)
		}

		// Save each group to its respective directory
		if len(groupAwards) > 0 {
			directory := directoryMapping[groupName]
			filename := fmt.Sprintf("%s/uc_%s_%s.json", directory, groupName, timestamp)

			if err := saveToJSON(groupAwards, filename); err != nil {
				return 0, fmt.Errorf("error saving %s data: %w", groupName, err)
			}

			log.Printf("Saved %s data to %s", groupName, filename)
		}

		totalAwards += len(groupAwards)
		log.Printf("Completed %s: %d awards. Running total: %d", groupName, len(groupAwards), totalAwards)

		// Small delay between groups
		time.Sleep(s.delay)
	}

	log.Printf("Completed all award types. Total awards collected: %d", totalAwards)
	return totalAwards, nil
}

func saveToJSON(data []Award, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	return nil
}

func main() {
	ctx := context.Background()

	scraper := NewScraper()

	totalAwards, err := scraper.scrapeAndSaveAllData(ctx)
	if err != nil {
		log.Fatalf("Error scraping data: %v", err)
	}

	log.Printf("Successfully scraped %d awards across all award types", totalAwards)
	log.Printf("Data saved to respective directories:")
	for groupName, directory := range directoryMapping {
		log.Printf("  %s -> %s/", groupName, directory)
	}
}