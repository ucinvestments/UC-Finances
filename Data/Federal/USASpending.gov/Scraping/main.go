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
	"strings"
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
	Keywords                    []string             `json:"keywords"`
	TimePeriod                  []TimePeriod         `json:"time_period"`
	AwardTypeCodes              []string             `json:"award_type_codes"`
	RecipientTypeNames          []string             `json:"recipient_type_names"`
	PlaceOfPerformanceLocations []PlaceOfPerformance `json:"place_of_performance_locations"`
}

type APIRequest struct {
	Filters       Filters  `json:"filters"`
	Page          int      `json:"page"`
	Limit         int      `json:"limit"`
	Sort          string   `json:"sort"`
	Order         string   `json:"order"`
	AuditTrail    string   `json:"auditTrail"`
	Fields        []string `json:"fields"`
	SpendingLevel string   `json:"spending_level"`
}

// API Response Structures
type PageMetadata struct {
	Page                int    `json:"page"`
	HasNext             bool   `json:"hasNext"`
	LastRecordUniqueID  int    `json:"last_record_unique_id"`
	LastRecordSortValue string `json:"last_record_sort_value"`
}

type Location struct {
	LocationCountryCode string  `json:"location_country_code"`
	CountryName         string  `json:"country_name"`
	StateCode           string  `json:"state_code"`
	StateName           string  `json:"state_name"`
	CityName            string  `json:"city_name"`
	CountyCode          string  `json:"county_code"`
	CountyName          string  `json:"county_name"`
	AddressLine1        string  `json:"address_line1"`
	AddressLine2        *string `json:"address_line2"`
	AddressLine3        *string `json:"address_line3"`
	CongressionalCode   string  `json:"congressional_code"`
	Zip4                string  `json:"zip4"`
	Zip5                string  `json:"zip5"`
	ForeignPostalCode   *string `json:"foreign_postal_code"`
	ForeignProvince     *string `json:"foreign_province"`
}

type CodeDescription struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type Award struct {
	InternalID                int         `json:"internal_id"`
	AwardID                   string      `json:"Award ID"`
	RecipientName             string      `json:"Recipient Name"`
	AwardAmount               interface{} `json:"Award Amount"`
	TotalOutlays              interface{} `json:"Total Outlays"`
	Description               string      `json:"Description"`
	ContractAwardType         string      `json:"Contract Award Type"`
	RecipientUEI              string      `json:"Recipient UEI"`
	RecipientLocation         interface{} `json:"Recipient Location"`           // Can be Location object or string fields
	PrimaryPlaceOfPerformance interface{} `json:"Primary Place of Performance"` // Can be Location object or string fields
	DefCodes                  []string    `json:"def_codes"`
	COVID19Obligations        interface{} `json:"COVID-19 Obligations"`
	COVID19Outlays            interface{} `json:"COVID-19 Outlays"`
	InfrastructureObligations interface{} `json:"Infrastructure Obligations"`
	InfrastructureOutlays     interface{} `json:"Infrastructure Outlays"`
	AwardingAgency            string      `json:"Awarding Agency"`
	AwardingSubAgency         string      `json:"Awarding Sub Agency"`
	StartDate                 string      `json:"Start Date"`
	EndDate                   string      `json:"End Date"`
	NAICS                     interface{} `json:"NAICS"` // Can be CodeDescription or string
	PSC                       interface{} `json:"PSC"`   // Can be CodeDescription or string
	RecipientID               string      `json:"recipient_id"`
	PrimeAwardRecipientID     string      `json:"prime_award_recipient_id"`
	GeneratedInternalID       string      `json:"generated_internal_id"`

	// Missing fields from search response
	AwardingAgencyID int    `json:"awarding_agency_id"`
	AgencySlug       string `json:"agency_slug"`

	// Loan-specific fields
	LoanValue     interface{} `json:"Loan Value"`
	SubsidyCost   interface{} `json:"Subsidy Cost"`
	IssuedDate    string      `json:"Issued Date"`
	FundingAgency string      `json:"Funding Agency"`

	// Location fields for loans (separate from Location object)
	RecipientLocationCityName     string `json:"recipient_location_city_name"`
	RecipientLocationStateCode    string `json:"recipient_location_state_code"`
	RecipientLocationCountryName  string `json:"recipient_location_country_name"`
	RecipientLocationAddressLine1 string `json:"recipient_location_address_line1"`
	POPCityName                   string `json:"pop_city_name"`
	POPStateCode                  string `json:"pop_state_code"`
	POPCountryName                string `json:"pop_country_name"`
}

type APIResponse struct {
	Results      []Award      `json:"results"`
	PageMetadata PageMetadata `json:"page_metadata"`
}

// Detailed Award Response Structures
type AgencyInfo struct {
	ID           int `json:"id"`
	HasAgencyPage bool `json:"has_agency_page"`
	ToptierAgency struct {
		Name         string `json:"name"`
		Code         string `json:"code"`
		Abbreviation string `json:"abbreviation"`
		Slug         string `json:"slug"`
	} `json:"toptier_agency"`
	SubtierAgency struct {
		Name         string `json:"name"`
		Code         string `json:"code"`
		Abbreviation string `json:"abbreviation"`
	} `json:"subtier_agency"`
	OfficeAgencyName string `json:"office_agency_name"`
}

type PeriodOfPerformance struct {
	StartDate         string `json:"start_date"`
	EndDate           string `json:"end_date"`
	LastModifiedDate  string `json:"last_modified_date"`
	PotentialEndDate  string `json:"potential_end_date"`
}

type RecipientDetail struct {
	RecipientHash         string   `json:"recipient_hash"`
	RecipientName         string   `json:"recipient_name"`
	RecipientUEI          string   `json:"recipient_uei"`
	RecipientUniqueID     *string  `json:"recipient_unique_id"`
	ParentRecipientHash   string   `json:"parent_recipient_hash"`
	ParentRecipientName   string   `json:"parent_recipient_name"`
	ParentRecipientUEI    string   `json:"parent_recipient_uei"`
	ParentRecipientUniqueID *string `json:"parent_recipient_unique_id"`
	BusinessCategories    []string `json:"business_categories"`
	Location              Location `json:"location"`
}

type ExecutiveDetails struct {
	Officers []struct {
		Name   *string `json:"name"`
		Amount *string `json:"amount"`
	} `json:"officers"`
}

type HierarchyCode struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type PSCHierarchy struct {
	ToptierCode  HierarchyCode `json:"toptier_code"`
	MidtierCode  HierarchyCode `json:"midtier_code"`
	SubtierCode  interface{}   `json:"subtier_code"` // Can be empty
	BaseCode     HierarchyCode `json:"base_code"`
}

type NAICSHierarchy struct {
	ToptierCode HierarchyCode `json:"toptier_code"`
	MidtierCode HierarchyCode `json:"midtier_code"`
	BaseCode    HierarchyCode `json:"base_code"`
}

type ContractData struct {
	IDVTypeDescription                               *string `json:"idv_type_description"`
	TypeOfIDCDescription                             *string `json:"type_of_idc_description"`
	ReferencedIDVAgencyIden                          *string `json:"referenced_idv_agency_iden"`
	ReferencedIDVAgencyDesc                          *string `json:"referenced_idv_agency_desc"`
	SolicitationIdentifier                           *string `json:"solicitation_identifier"`
	SolicitationProcedures                           string  `json:"solicitation_procedures"`
	NumberOfOffersReceived                           string  `json:"number_of_offers_received"`
	ExtentCompeted                                   string  `json:"extent_competed"`
	TypeSetAside                                     string  `json:"type_set_aside"`
	TypeSetAsideDescription                          string  `json:"type_set_aside_description"`
	EvaluatedPreference                              string  `json:"evaluated_preference"`
	FedBizOpps                                       string  `json:"fed_biz_opps"`
	FedBizOppsDescription                            string  `json:"fed_biz_opps_description"`
	SmallBusinessCompetitive                         bool    `json:"small_business_competitive"`
	ProductOrServiceCode                             string  `json:"product_or_service_code"`
	NAICS                                            string  `json:"naics"`
	NAICSDescription                                 string  `json:"naics_description"`
	SeaTransportation                                *string `json:"sea_transportation"`
	ClingerCohenActPlanning                          string  `json:"clinger_cohen_act_planning"`
	LaborStandards                                   string  `json:"labor_standards"`
	CostOrPricingData                                string  `json:"cost_or_pricing_data"`
	DomesticOrForeignEntity                          *string `json:"domestic_or_foreign_entity"`
	ForeignFunding                                   string  `json:"foreign_funding"`
	MajorProgram                                     *string `json:"major_program"`
	ProgramAcronym                                   *string `json:"program_acronym"`
	SubcontractingPlan                               string  `json:"subcontracting_plan"`
	MultiYearContract                                string  `json:"multi_year_contract"`
	ConsolidatedContract                             string  `json:"consolidated_contract"`
	TypeOfContractPricing                            string  `json:"type_of_contract_pricing"`
	NationalInterestAction                           *string `json:"national_interest_action"`
	MultipleOrSingleAwardDescription                 *string `json:"multiple_or_single_award_description"`
	SolicitationProceduresDescription                string  `json:"solicitation_procedures_description"`
	ExtentCompetedDescription                        string  `json:"extent_competed_description"`
	OtherThanFullAndOpen                             *string `json:"other_than_full_and_open"`
	OtherThanFullAndOpenDescription                  *string `json:"other_than_full_and_open_description"`
	CommercialItemAcquisition                        string  `json:"commercial_item_acquisition"`
	CommercialItemAcquisitionDescription             string  `json:"commercial_item_acquisition_description"`
	CommercialItemTestProgram                        string  `json:"commercial_item_test_program"`
	CommercialItemTestProgramDescription             string  `json:"commercial_item_test_program_description"`
	EvaluatedPreferenceDescription                   string  `json:"evaluated_preference_description"`
	FairOpportunityLimited                           *string `json:"fair_opportunity_limited"`
	FairOpportunityLimitedDescription                *string `json:"fair_opportunity_limited_description"`
	ProductOrServiceDescription                      string  `json:"product_or_service_description"`
	DODClaimantProgram                               *string `json:"dod_claimant_program"`
	DODClaimantProgramDescription                    *string `json:"dod_claimant_program_description"`
	DODAcquisitionProgram                            *string `json:"dod_acquisition_program"`
	DODAcquisitionProgramDescription                 *string `json:"dod_acquisition_program_description"`
	InformationTechnologyCommercialItemCategory      *string `json:"information_technology_commercial_item_category"`
	InformationTechnologyCommercialItemCategoryDescription *string `json:"information_technology_commercial_item_category_description"`
	SeaTransportationDescription                     *string `json:"sea_transportation_description"`
	ClingerCohenActPlanningDescription               string  `json:"clinger_cohen_act_planning_description"`
	ConstructionWageRate                             string  `json:"construction_wage_rate"`
	ConstructionWageRateDescription                  string  `json:"construction_wage_rate_description"`
	LaborStandardsDescription                        string  `json:"labor_standards_description"`
	MaterialsSupplies                                string  `json:"materials_supplies"`
	MaterialsSuppliesDescription                     string  `json:"materials_supplies_description"`
	CostOrPricingDataDescription                     string  `json:"cost_or_pricing_data_description"`
	DomesticOrForeignEntityDescription               *string `json:"domestic_or_foreign_entity_description"`
	ForeignFundingDescription                        string  `json:"foreign_funding_description"`
	InteragencyContractingAuthority                  string  `json:"interagency_contracting_authority"`
	InteragencyContractingAuthorityDescription       string  `json:"interagency_contracting_authority_description"`
	PriceEvaluationAdjustment                        string  `json:"price_evaluation_adjustment"`
	SubcontractingPlanDescription                    string  `json:"subcontracting_plan_description"`
	MultiYearContractDescription                     string  `json:"multi_year_contract_description"`
	PurchaseCardAsPaymentMethod                      string  `json:"purchase_card_as_payment_method"`
	PurchaseCardAsPaymentMethodDescription           string  `json:"purchase_card_as_payment_method_description"`
	ConsolidatedContractDescription                  string  `json:"consolidated_contract_description"`
	TypeOfContractPricingDescription                 string  `json:"type_of_contract_pricing_description"`
	NationalInterestActionDescription                *string `json:"national_interest_action_description"`
}

type AccountObligation struct {
	Code   string  `json:"code"`
	Amount float64 `json:"amount"`
}

type DetailedAwardResponse struct {
	ID                      int                 `json:"id"`
	GeneratedUniqueAwardID  string              `json:"generated_unique_award_id"`
	PIID                    string              `json:"piid"`
	Category                string              `json:"category"`
	Type                    string              `json:"type"`
	TypeDescription         string              `json:"type_description"`
	Description             string              `json:"description"`
	TotalObligation         float64             `json:"total_obligation"`
	SubawardCount           int                 `json:"subaward_count"`
	TotalSubawardAmount     *float64            `json:"total_subaward_amount"`
	DateSigned              string              `json:"date_signed"`
	BaseExercisedOptions    float64             `json:"base_exercised_options"`
	BaseAndAllOptions       float64             `json:"base_and_all_options"`
	TotalAccountOutlay      float64             `json:"total_account_outlay"`
	TotalAccountObligation  float64             `json:"total_account_obligation"`
	AccountOutlaysByDEFC    []AccountObligation `json:"account_outlays_by_defc"`
	AccountObligationsByDEFC []AccountObligation `json:"account_obligations_by_defc"`
	ParentAward             interface{}         `json:"parent_award"`
	LatestTransactionContractData *ContractData `json:"latest_transaction_contract_data"`
	FundingAgency           AgencyInfo          `json:"funding_agency"`
	AwardingAgency          AgencyInfo          `json:"awarding_agency"`
	PeriodOfPerformance     PeriodOfPerformance `json:"period_of_performance"`
	Recipient               RecipientDetail     `json:"recipient"`
	ExecutiveDetails        ExecutiveDetails    `json:"executive_details"`
	PlaceOfPerformance      Location            `json:"place_of_performance"`
	PSCHierarchy            PSCHierarchy        `json:"psc_hierarchy"`
	NAICSHierarchy          NAICSHierarchy      `json:"naics_hierarchy"`
	TotalOutlay             float64             `json:"total_outlay"`
}

// Combined structure that holds both basic and detailed award data
type EnhancedAward struct {
	BasicData    Award                 `json:"basic_data"`
	DetailedData *DetailedAwardResponse `json:"detailed_data,omitempty"`
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

// Directory mapping for each award type group
var directoryMapping = map[string]string{
	"contracts":                  "../Contracts",
	"grants":                     "../Grants",
	"loans":                      "../Loans",
	"idvs":                       "../Contract_IDVs",
	"other_financial_assistance": "../Other_Financial_Assistance",
	"direct_payments":            "../Direct_Payments",
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

func (s *Scraper) fetchDetailedAward(ctx context.Context, generatedInternalID string) (*DetailedAwardResponse, error) {
	detailURL := fmt.Sprintf("https://api.usaspending.gov/api/v2/awards/%s/", generatedInternalID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", detailURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating detail request: %w", err)
	}

	req.Header.Set("User-Agent", "UC-Holdings-Scraper/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making detail request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("detail API returned status %d: %s", resp.StatusCode, string(body))
	}

	var detailResponse DetailedAwardResponse
	if err := json.NewDecoder(resp.Body).Decode(&detailResponse); err != nil {
		return nil, fmt.Errorf("error decoding detail response: %w", err)
	}

	return &detailResponse, nil
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

func (s *Scraper) scrapeAndSaveEnhancedData(ctx context.Context) (int, error) {
	totalAwards := 0
	totalEnhanced := 0

	log.Printf("Starting enhanced scraping: collecting basic data and detailed information...")

	for groupName, awardTypeCodes := range awardTypeGroups {
		log.Printf("Processing %s awards...", groupName)
		
		// Step 1: Collect basic award data
		groupAwards, err := s.scrapeGroupData(ctx, groupName, awardTypeCodes)
		if err != nil {
			return 0, fmt.Errorf("error scraping %s: %w", groupName, err)
		}
		
		totalAwards += len(groupAwards)
		log.Printf("Collected %d basic %s awards. Now fetching detailed data...", len(groupAwards), groupName)

		// Step 2: Fetch detailed data and save organized files
		for i, award := range groupAwards {
			select {
			case <-ctx.Done():
				return totalEnhanced, ctx.Err()
			default:
			}

			if award.GeneratedInternalID == "" {
				log.Printf("Skipping award %d/%d in %s: missing generated_internal_id", i+1, len(groupAwards), groupName)
				continue
			}

			log.Printf("Fetching details for award %d/%d in %s: %s", i+1, len(groupAwards), groupName, award.GeneratedInternalID)

			// Fetch detailed award data
			detailedData, err := s.fetchDetailedAward(ctx, award.GeneratedInternalID)
			if err != nil {
				log.Printf("Warning: Failed to fetch details for %s: %v", award.GeneratedInternalID, err)
				// Continue with basic data only
				detailedData = nil
			}

			// Create enhanced award structure
			enhancedAward := EnhancedAward{
				BasicData:    award,
				DetailedData: detailedData,
			}

			// Determine file organization
			year := extractYearFromDate(award.StartDate)
			if year == "unknown" && detailedData != nil {
				year = extractYearFromDate(detailedData.DateSigned)
			}

			recipientName := award.RecipientName
			if recipientName == "" {
				recipientName = "Unknown_Recipient"
			}

			awardingAgency := award.AwardingAgency
			if awardingAgency == "" && detailedData != nil {
				awardingAgency = detailedData.AwardingAgency.ToptierAgency.Name
			}
			if awardingAgency == "" {
				awardingAgency = "Unknown_Agency"
			}

			// Create organized file path
			dirPath := createDirectoryPath(groupName, recipientName, year, awardingAgency)
			fileName := fmt.Sprintf("%s.json", sanitizeFileName(award.GeneratedInternalID))
			filePath := fmt.Sprintf("%s/%s", dirPath, fileName)

			// Save enhanced award data
			if err := saveEnhancedAwardToJSON(enhancedAward, filePath); err != nil {
				log.Printf("Error saving award %s: %v", award.GeneratedInternalID, err)
				continue
			}

			totalEnhanced++
			
			// Rate limiting between detail requests
			time.Sleep(s.delay)
		}

		log.Printf("Completed %s: saved %d enhanced awards", groupName, len(groupAwards))

		// Small delay between groups
		time.Sleep(s.delay)
	}

	log.Printf("Enhanced scraping completed!")
	log.Printf("Total basic awards collected: %d", totalAwards)
	log.Printf("Total enhanced awards saved: %d", totalEnhanced)
	return totalEnhanced, nil
}

// Utility functions for directory and file organization
func sanitizeFileName(name string) string {
	// Replace spaces and special characters with underscores
	sanitized := strings.ReplaceAll(name, " ", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")
	sanitized = strings.ReplaceAll(sanitized, ":", "_")
	sanitized = strings.ReplaceAll(sanitized, "*", "_")
	sanitized = strings.ReplaceAll(sanitized, "?", "_")
	sanitized = strings.ReplaceAll(sanitized, "\"", "_")
	sanitized = strings.ReplaceAll(sanitized, "<", "_")
	sanitized = strings.ReplaceAll(sanitized, ">", "_")
	sanitized = strings.ReplaceAll(sanitized, "|", "_")
	return sanitized
}

func extractYearFromDate(dateStr string) string {
	if dateStr == "" {
		return "unknown"
	}
	// Handle various date formats
	if len(dateStr) >= 4 {
		return dateStr[:4]
	}
	return "unknown"
}

func createDirectoryPath(groupName, recipientName, year, awardingAgency string) string {
	baseDir := directoryMapping[groupName]
	if baseDir == "" {
		baseDir = "../Other"
	}
	
	sanitizedRecipient := sanitizeFileName(recipientName)
	sanitizedAgency := sanitizeFileName(awardingAgency)
	
	return fmt.Sprintf("%s/%s/%s/%s", baseDir, sanitizedRecipient, year, sanitizedAgency)
}

func ensureDirectoryExists(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

func saveEnhancedAwardToJSON(award EnhancedAward, filepath string) error {
	if err := ensureDirectoryExists(filepath[:strings.LastIndex(filepath, "/")]); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(award); err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	return nil
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

	log.Printf("Starting USASpending.gov Enhanced Scraper")
	log.Printf("This will collect basic award data and detailed information for each award")
	log.Printf("Awards will be organized by: [Award Type]/[Recipient]/[Year]/[Agency]/[Award ID].json")
	
	totalAwards, err := scraper.scrapeAndSaveEnhancedData(ctx)
	if err != nil {
		log.Fatalf("Error scraping enhanced data: %v", err)
	}

	log.Printf("Successfully scraped and saved %d enhanced awards", totalAwards)
	log.Printf("Data organized in hierarchical directory structure:")
	for groupName, directory := range directoryMapping {
		log.Printf("  %s -> %s/[Recipient]/[Year]/[Agency]/", groupName, directory)
	}
}
