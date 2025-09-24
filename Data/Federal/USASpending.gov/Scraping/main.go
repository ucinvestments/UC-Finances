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
	InternalID                 int         `json:"internal_id"`
	AwardID                    string      `json:"Award ID"`
	RecipientName              string      `json:"Recipient Name"`
	AwardAmount                interface{} `json:"Award Amount"`
	TotalOutlays               interface{} `json:"Total Outlays"`
	Description                string      `json:"Description"`
	ContractAwardType          string          `json:"Contract Award Type"`
	RecipientUEI               string          `json:"Recipient UEI"`
	RecipientLocation          Location        `json:"Recipient Location"`
	PrimaryPlaceOfPerformance  Location        `json:"Primary Place of Performance"`
	DefCodes                   []string        `json:"def_codes"`
	COVID19Obligations         interface{}     `json:"COVID-19 Obligations"`
	COVID19Outlays             interface{}     `json:"COVID-19 Outlays"`
	InfrastructureObligations  interface{}     `json:"Infrastructure Obligations"`
	InfrastructureOutlays      interface{}     `json:"Infrastructure Outlays"`
	AwardingAgency             string          `json:"Awarding Agency"`
	AwardingSubAgency          string          `json:"Awarding Sub Agency"`
	StartDate                  string          `json:"Start Date"`
	EndDate                    string          `json:"End Date"`
	NAICS                      CodeDescription `json:"NAICS"`
	PSC                        CodeDescription `json:"PSC"`
	RecipientID                string      `json:"recipient_id"`
	PrimeAwardRecipientID      string      `json:"prime_award_recipient_id"`
	GeneratedInternalID        string      `json:"generated_internal_id"`
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

func (s *Scraper) createRequest() APIRequest {
	return APIRequest{
		Filters: Filters{
			Keywords: []string{"University of California"},
			TimePeriod: []TimePeriod{
				{
					StartDate: "2007-10-01",
					EndDate:   "2025-09-30",
				},
			},
			AwardTypeCodes: []string{"A", "B", "C", "D"},
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
		Page:   1,
		Limit:  100,
		Sort:   "Award Amount",
		Order:  "desc",
		AuditTrail: "Results Table - Spending by award search",
		Fields: []string{
			"Award ID",
			"Recipient Name",
			"Award Amount",
			"Total Outlays",
			"Description",
			"Contract Award Type",
			"Recipient UEI",
			"Recipient Location",
			"Primary Place of Performance",
			"def_codes",
			"COVID-19 Obligations",
			"COVID-19 Outlays",
			"Infrastructure Obligations",
			"Infrastructure Outlays",
			"Awarding Agency",
			"Awarding Sub Agency",
			"Start Date",
			"End Date",
			"NAICS",
			"PSC",
			"recipient_id",
			"prime_award_recipient_id",
		},
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

func (s *Scraper) scrapeAllData(ctx context.Context) ([]Award, error) {
	var allAwards []Award
	page := 1

	log.Printf("Starting to scrape University of California data...")

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		log.Printf("Fetching page %d...", page)

		request := s.createRequest()
		request.Page = page

		response, err := s.makeRequest(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("error fetching page %d: %w", page, err)
		}

		allAwards = append(allAwards, response.Results...)

		log.Printf("Page %d: got %d awards, total so far: %d",
			page, len(response.Results), len(allAwards))

		// Check if there are more pages
		if !response.PageMetadata.HasNext || len(response.Results) == 0 {
			log.Printf("No more pages. Total awards collected: %d", len(allAwards))
			break
		}

		page++

		// Be respectful to the API
		time.Sleep(s.delay)
	}

	return allAwards, nil
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

	awards, err := scraper.scrapeAllData(ctx)
	if err != nil {
		log.Fatalf("Error scraping data: %v", err)
	}

	// Save to JSON file
	filename := fmt.Sprintf("uc_awards_%s.json", time.Now().Format("2006-01-02"))
	if err := saveToJSON(awards, filename); err != nil {
		log.Fatalf("Error saving data: %v", err)
	}

	log.Printf("Successfully scraped %d awards and saved to %s", len(awards), filename)

	// Print summary statistics
	log.Printf("\nSummary:")
	log.Printf("- Total awards: %d", len(awards))

	if len(awards) > 0 {
		// Count unique recipients
		recipients := make(map[string]int)
		for _, award := range awards {
			recipients[award.RecipientName]++
		}
		log.Printf("- Unique recipients: %d", len(recipients))

		// Show top 5 recipients by number of awards
		log.Printf("\nTop recipients by number of awards:")
		type recipientCount struct {
			name  string
			count int
		}

		var sortedRecipients []recipientCount
		for name, count := range recipients {
			sortedRecipients = append(sortedRecipients, recipientCount{name, count})
		}

		// Simple sorting by count (descending)
		for i := 0; i < len(sortedRecipients)-1; i++ {
			for j := 0; j < len(sortedRecipients)-i-1; j++ {
				if sortedRecipients[j].count < sortedRecipients[j+1].count {
					sortedRecipients[j], sortedRecipients[j+1] = sortedRecipients[j+1], sortedRecipients[j]
				}
			}
		}

		max := 5
		if len(sortedRecipients) < 5 {
			max = len(sortedRecipients)
		}

		for i := 0; i < max; i++ {
			log.Printf("  %s: %d awards", sortedRecipients[i].name, sortedRecipients[i].count)
		}
	}
}