package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const recipientSource = "https://www.ucop.edu/research-policy-analysis-coordination/resources-tools/contract-and-grant-manual/chapter10/index.html#10-530"
const awardCountURL = "https://api.usaspending.gov/api/v2/search/spending_by_award_count/"

type recipientIdentifiers struct {
	Location string `json:"location"`
	UEI      string `json:"uei"`
	DUNS     string `json:"duns"`
}

var campusRecipients = []recipientIdentifiers{
	{"Office of the President", "PKK5TD16N4H1", "003985512"},
	{"Berkeley", "GS3YEVSS12N6", "124726725"},
	{"Davis", "TX2DAGQPENZ5", "047120084"},
	{"Irvine", "MJC5FCYQTPE6", "046705849"},
	{"Los Angeles", "RN64EPNH8JC6", "092530369"},
	{"Merced", "FFM7VPAG8P92", "113645084"},
	{"Riverside", "MR5QC5FCAVH5", "627797426"},
	{"San Diego", "UYTTZT6G9DT1", "804355790"},
	{"San Francisco", "KMH5K9V7S518", "094878337"},
	{"Santa Barbara", "G9QBQDH39DF4", "094878394"},
	{"Santa Cruz", "VXUFPE4MCZH5", "125084723"},
}

var verifiedAwardGroups = []struct {
	name       string
	legacy     string
	extraCodes []string
}{
	{"contracts", "contracts", nil},
	{"direct_payments", "other_financial_assistance", []string{"F006", "F007"}},
	{"grants", "grants", []string{"F001", "F002"}},
	{"idvs", "idvs", nil},
	{"loans", "loans", []string{"F003", "F004"}},
	{"other", "direct_payments", []string{"-1", "F005", "F008", "F009", "F010"}},
}

type verifiedAward struct {
	Location string `json:"uc_location"`
	Award
}

type verifiedAwardGroup struct {
	Codes          []string        `json:"award_type_codes"`
	APICount       int             `json:"api_count"`
	UnmatchedCount int             `json:"unmatched_count"`
	UnmatchedIDs   []string        `json:"unmatched_award_ids,omitempty"`
	Awards         []verifiedAward `json:"awards"`
}

func (s *Scraper) writeVerifiedAwards(ctx context.Context, output string, period TimePeriod) error {
	start, startErr := time.Parse(time.DateOnly, period.StartDate)
	end, endErr := time.Parse(time.DateOnly, period.EndDate)
	if startErr != nil || endErr != nil || end.Before(start) || period.StartDate < "2007-10-01" {
		return fmt.Errorf("expected a valid date range starting on or after 2007-10-01")
	}
	if _, err := os.Lstat(output); err == nil {
		return fmt.Errorf("output already exists: %s", output)
	} else if !os.IsNotExist(err) {
		return err
	}

	identifiers := make([]string, 0, len(campusRecipients)*2)
	locations := make(map[string]string, cap(identifiers))
	for _, recipient := range campusRecipients {
		identifiers = append(identifiers, recipient.UEI, recipient.DUNS)
		locations[recipient.UEI] = recipient.Location
		locations[recipient.DUNS] = recipient.Location
	}
	filters := Filters{RecipientSearchText: identifiers, TimePeriod: []TimePeriod{period}}
	var counts struct {
		Results map[string]int `json:"results"`
	}
	if err := s.postJSON(ctx, awardCountURL, map[string]Filters{"filters": filters}, &counts); err != nil {
		return fmt.Errorf("count award matches: %w", err)
	}

	groups := make(map[string]verifiedAwardGroup, len(verifiedAwardGroups))
	seen := make(map[string]struct{})
	for _, group := range verifiedAwardGroups {
		codes := append(append([]string(nil), awardTypeGroups[group.legacy]...), group.extraCodes...)
		expected, exists := counts.Results[group.name]
		if !exists {
			return fmt.Errorf("USAspending count is missing %s", group.name)
		}
		request := s.createRequest(group.legacy, codes)
		request.Filters = filters
		request.Filters.AwardTypeCodes = codes
		request.Fields = append(append([]string(nil), request.Fields...), "Recipient DUNS Number")
		request.Sort = "Award ID"
		request.Order = "asc"
		awards, err := s.scrapeAwards(ctx, group.name, request)
		if err != nil {
			return err
		}
		if len(awards) != expected {
			return fmt.Errorf("%s: collected %d results, but USAspending reported %d", group.name, len(awards), expected)
		}

		result := verifiedAwardGroup{Codes: codes, APICount: expected, Awards: []verifiedAward{}}
		for _, award := range awards {
			id := award.GeneratedInternalID
			if _, duplicate := seen[id]; duplicate {
				return fmt.Errorf("award %s appeared in multiple groups", id)
			}
			seen[id] = struct{}{}

			ueiLocation := locations[strings.ToUpper(strings.TrimSpace(award.RecipientUEI))]
			dunsLocation := locations[strings.TrimSpace(award.RecipientDUNS)]
			if ueiLocation != "" && dunsLocation != "" && ueiLocation != dunsLocation {
				return fmt.Errorf("award %s has conflicting UC recipient identifiers", id)
			}
			location := ueiLocation
			if location == "" {
				location = dunsLocation
			}
			if location == "" {
				result.UnmatchedIDs = append(result.UnmatchedIDs, id)
				continue
			}
			result.Awards = append(result.Awards, verifiedAward{Location: location, Award: award})
		}
		result.UnmatchedCount = len(result.UnmatchedIDs)
		groups[group.name] = result
		log.Printf("[%s] %d verified of %d USAspending matches", group.name, len(result.Awards), expected)
	}

	archive := struct {
		CollectedAt     string                        `json:"collected_at"`
		Period          TimePeriod                    `json:"period"`
		RecipientSource string                        `json:"recipient_source"`
		Recipients      []recipientIdentifiers        `json:"recipients"`
		Groups          map[string]verifiedAwardGroup `json:"groups"`
	}{time.Now().UTC().Format(time.RFC3339), period, recipientSource, campusRecipients, groups}
	file, err := os.CreateTemp(filepath.Dir(output), ".verified-awards-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	encodeErr := json.NewEncoder(file).Encode(archive)
	closeErr := file.Close()
	if encodeErr != nil {
		return encodeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Link(file.Name(), output); err != nil {
		return err
	}
	log.Printf("Wrote verified awards to %s", output)
	return nil
}
