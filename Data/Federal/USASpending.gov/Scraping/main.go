package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	baseURL        = "https://api.usaspending.gov/api/v2"
	searchEndpoint = "/search/spending_by_award"
	maxWorkers     = 10
	pageSize       = 100
	requestTimeout = 30 * time.Second
	rateLimit      = 100 * time.Millisecond
)

type SearchRequest struct {
	Filters Filters  `json:"filters"`
	Fields  []string `json:"fields,omitempty"`
	Page    int      `json:"page"`
	Limit   int      `json:"limit"`
	Sort    string   `json:"sort"`
	Order   string   `json:"order"`
}

type Filters struct {
	Keywords            []string            `json:"keywords"`
	RecipientLocations  []RecipientLocation `json:"recipient_locations"`
	RecipientTypeNames  []string            `json:"recipient_type_names"`
	TimePeriod          []TimePeriod        `json:"time_period,omitempty"`
	AwardTypeCodesArray []string            `json:"award_type_codes,omitempty"`
}

type RecipientLocation struct {
	Country  string `json:"country"`
	State    string `json:"state"`
	County   string `json:"county,omitempty"`
	District string `json:"district,omitempty"`
	City     string `json:"city,omitempty"`
	Zip      string `json:"zip,omitempty"`
}

type TimePeriod struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type SearchResponse struct {
	Page       int         `json:"page"`
	HasNext    bool        `json:"hasNext"`
	Next       int         `json:"next,omitempty"`
	Previous   int         `json:"previous,omitempty"`
	PageCount  int         `json:"page_count"`
	Limit      int         `json:"limit"`
	TotalCount int         `json:"total"`
	Results    []AwardData `json:"results"`
}

type AwardData struct {
	InternalID               string      `json:"internal_id"`
	AwardID                  interface{} `json:"Award ID"`
	RecipientName            interface{} `json:"Recipient Name"`
	RecipientUEI             interface{} `json:"recipient_uei"`
	StartDate                interface{} `json:"Start Date"`
	EndDate                  interface{} `json:"End Date"`
	AwardAmount              interface{} `json:"Award Amount"`
	TotalObligated           interface{} `json:"Total Outlays"`
	AwardingAgencyName       interface{} `json:"Awarding Agency"`
	AwardingSubAgencyName    interface{} `json:"Awarding Sub Agency"`
	AwardType                interface{} `json:"Award Type"`
	PrimeAwardType           interface{} `json:"prime_award_type"`
	FundingAgencyName        interface{} `json:"Funding Agency"`
	FundingSubAgencyName     interface{} `json:"Funding Sub Agency"`
	Description              interface{} `json:"Description"`
	Piid                     interface{} `json:"piid"`
	Fain                     interface{} `json:"fain"`
	Uri                      interface{} `json:"uri"`
	CFDANumber               interface{} `json:"cfda_number"`
	CFDATitle                interface{} `json:"cfda_title"`
	PscDescription           interface{} `json:"psc_description"`
	NaicsCode                interface{} `json:"naics_code"`
	NaicsDescription         interface{} `json:"naics_description"`
	RecipientCityName        interface{} `json:"recipient_city_name"`
	RecipientCountyName      interface{} `json:"recipient_county_name"`
	RecipientStateCode       interface{} `json:"recipient_state_code"`
	RecipientZip5            interface{} `json:"recipient_zip5"`
	RecipientCongressionalDistrict interface{} `json:"recipient_congressional_district"`
	PrimaryPlaceOfPerformance interface{} `json:"primary_place_of_performance"`
	BusinessCategories       interface{} `json:"business_categories"`
	TypeOfContractPricing    interface{} `json:"type_of_contract_pricing"`
	TypeSetAside             interface{} `json:"type_set_aside"`
	ExtentCompeted           interface{} `json:"extent_competed"`
}

type Scraper struct {
	client     *http.Client
	mu         sync.Mutex
	wg         sync.WaitGroup
	rateLimiter *time.Ticker
}

func NewScraper() *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: requestTimeout,
		},
		rateLimiter: time.NewTicker(rateLimit),
	}
}

func (s *Scraper) fetchAllData() error {
	defer s.rateLimiter.Stop()

	log.Println("Starting to fetch all award data...")

	awardTypeSets := map[string][]string{
		"contracts": {"A", "B", "C", "D"},
		"contract_idvs": {"IDV_A", "IDV_B", "IDV_B_A", "IDV_B_B", "IDV_B_C", "IDV_C", "IDV_D", "IDV_E"},
		"grants": {"02", "03", "04", "05"},
		"direct_payments": {"06", "10"},
		"loans": {"07", "08", "09"},
		"other": {"11", "99"},
	}

	allResults := make(map[string][]AwardData)
	resultsMutex := &sync.Mutex{}

	jobs := make(chan JobData, 100)

	for i := 0; i < maxWorkers; i++ {
		s.wg.Add(1)
		go s.worker(jobs, allResults, resultsMutex)
	}

	currentYear := time.Now().Year()
	startYear := 2008

	for category, codes := range awardTypeSets {
		for year := startYear; year <= currentYear; year++ {
			jobs <- JobData{
				Category: category,
				Codes:    codes,
				Year:     year,
			}
		}
	}
	close(jobs)

	s.wg.Wait()

	for category, results := range allResults {
		if len(results) > 0 {
			if err := s.saveResults(results, category); err != nil {
				log.Printf("Error saving %s results: %v", category, err)
			}
		}
	}

	return nil
}

type JobData struct {
	Category string
	Codes    []string
	Year     int
}

func (s *Scraper) worker(jobs <-chan JobData, allResults map[string][]AwardData, resultsMutex *sync.Mutex) {
	defer s.wg.Done()

	for job := range jobs {
		<-s.rateLimiter.C

		log.Printf("Fetching %s for year %d", job.Category, job.Year)

		startDate := fmt.Sprintf("%d-01-01", job.Year)
		endDate := fmt.Sprintf("%d-12-31", job.Year)

		req := SearchRequest{
			Filters: Filters{
				Keywords: []string{"University of California"},
				RecipientLocations: []RecipientLocation{
					{
						Country: "USA",
						State:   "CA",
					},
				},
				RecipientTypeNames: []string{"higher_education"},
				TimePeriod: []TimePeriod{
					{StartDate: startDate, EndDate: endDate},
				},
				AwardTypeCodesArray: job.Codes,
			},
			Fields: []string{
				"Award ID",
				"Recipient Name",
				"recipient_uei",
				"Start Date",
				"End Date",
				"Award Amount",
				"Total Outlays",
				"Awarding Agency",
				"Awarding Sub Agency",
				"Award Type",
				"prime_award_type",
				"Funding Agency",
				"Funding Sub Agency",
				"Description",
				"piid",
				"fain",
				"uri",
				"cfda_number",
				"cfda_title",
				"psc_description",
				"naics_code",
				"naics_description",
				"recipient_city_name",
				"recipient_county_name",
				"recipient_state_code",
				"recipient_zip5",
				"recipient_congressional_district",
				"primary_place_of_performance",
				"business_categories",
				"type_of_contract_pricing",
				"type_set_aside",
				"extent_competed",
			},
			Page:  1,
			Limit: pageSize,
			Sort:  "Award Amount",
			Order: "desc",
		}

		yearResults := []AwardData{}

		for {
			results, hasNext, err := s.makeAPIRequest(req)
			if err != nil {
				log.Printf("Error fetching %s page %d for year %d: %v", job.Category, req.Page, job.Year, err)
				break
			}

			yearResults = append(yearResults, results...)

			log.Printf("%s year %d, page %d: fetched %d records (total: %d)",
				job.Category, job.Year, req.Page, len(results), len(yearResults))

			if !hasNext {
				break
			}
			req.Page++
		}

		if len(yearResults) > 0 {
			resultsMutex.Lock()
			allResults[job.Category] = append(allResults[job.Category], yearResults...)
			resultsMutex.Unlock()
			log.Printf("Completed %s for year %d: %d total records", job.Category, job.Year, len(yearResults))
		}
	}
}

func (s *Scraper) makeAPIRequest(req SearchRequest) ([]AwardData, bool, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, false, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", baseURL+searchEndpoint+"/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, false, fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, false, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, false, fmt.Errorf("unmarshaling response: %w", err)
	}

	return searchResp.Results, searchResp.HasNext, nil
}

func mapCategoryToDirectory(category string) string {
	switch category {
	case "contracts":
		return "Contracts"
	case "contract_idvs":
		return "Contract_IDVs"
	case "grants":
		return "Grants"
	case "direct_payments":
		return "Direct_Payments"
	case "loans":
		return "Loans"
	case "other":
		return "Other"
	default:
		return "Other"
	}
}

func (s *Scraper) saveResults(results []AwardData, category string) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	dirName := mapCategoryToDirectory(category)
	baseDir := filepath.Join("..", dirName)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	filename := filepath.Join(baseDir, fmt.Sprintf("UC_%s_%s.json", strings.ToLower(dirName), timestamp))

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	data := map[string]interface{}{
		"source":      "USASpending.gov API",
		"query": map[string]interface{}{
			"keywords":           "University of California",
			"recipient_location": "California",
			"recipient_type":     "higher_education",
		},
		"award_category": category,
		"timestamp":      timestamp,
		"total_count":    len(results),
		"results":        results,
	}

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("encoding data: %w", err)
	}

	log.Printf("✓ Saved %d %s records to %s", len(results), category, filename)
	return nil
}

func main() {
	log.Println("=== USASpending.gov Data Scraper ===")
	log.Println("Query Parameters:")
	log.Println("  Keywords: University of California")
	log.Println("  Location: California")
	log.Println("  Recipient Type: Higher Education")
	log.Println("  Years: 2008-" + fmt.Sprintf("%d", time.Now().Year()))
	log.Println("=====================================")

	scraper := NewScraper()

	if err := scraper.fetchAllData(); err != nil {
		log.Fatal("Error during scraping: ", err)
	}

	log.Println("✓ Scraping completed successfully!")
}