package match

import (
	"fmt"
	"sort"
	"strings"

	"errandShop/internal/core/types"
	"errandShop/pkg/textnorm"
	"github.com/xrash/smetrics"
)

// Matcher handles address-to-zone matching
type Matcher struct {
	zones []types.DeliveryZone
}

// NewMatcher creates a new matcher with delivery zones
func NewMatcher(zones []types.DeliveryZone) *Matcher {
	return &Matcher{
		zones: zones,
	}
}

// MatchAddress matches an address to a delivery zone
func (m *Matcher) MatchAddress(address string) (*types.MatchResult, *types.NoMatchResult) {
	// Step 1: Normalize the input address
	normalizedAddress := textnorm.Normalize(address)
	
	// Step 2: Try exact phrase matching
	if result := m.exactMatch(normalizedAddress); result != nil {
		return result, nil
	}
	
	// Step 3: Try fuzzy matching
	if result := m.fuzzyMatch(normalizedAddress); result != nil {
		return result, nil
	}
	
	// Step 4: No match found, return suggestions
	suggestions := m.generateSuggestions(normalizedAddress)
	noMatch := &types.NoMatchResult{
		MatchedBy:   "none",
		Message:     "No matching delivery zone found for the provided address",
		Suggestions: suggestions,
	}
	
	return nil, noMatch
}

// exactMatch performs exact phrase substring matching
func (m *Matcher) exactMatch(normalizedAddress string) *types.MatchResult {
	var bestMatch *types.MatchResult
	longestKeywordLen := 0
	
	for _, zone := range m.zones {
		for _, location := range zone.Locations {
			normalizedLocation := textnorm.Normalize(location)
			
			// Check if the normalized location is a substring of the normalized address
			if strings.Contains(normalizedAddress, normalizedLocation) {
				// If this is the longest match so far, use it
				if len(normalizedLocation) > longestKeywordLen {
					longestKeywordLen = len(normalizedLocation)
					bestMatch = &types.MatchResult{
						ZoneID:         zone.ZoneID,
						ZoneName:       fmt.Sprintf("Zone %d", zone.ZoneID),
						MatchedKeyword: location,
						MatchedBy:      "exact",
						Confidence:     1.0,
						Price:          zone.Price,
					}
				}
			}
		}
	}
	
	return bestMatch
}

// fuzzyMatch performs fuzzy matching using Jaro-Winkler similarity
func (m *Matcher) fuzzyMatch(normalizedAddress string) *types.MatchResult {
	var bestMatch *types.MatchResult
	bestScore := 0.0
	longestKeywordLen := 0
	const threshold = 0.88
	
	for _, zone := range m.zones {
		for _, location := range zone.Locations {
			normalizedLocation := textnorm.Normalize(location)
			
			// Calculate Jaro-Winkler similarity
			score := smetrics.JaroWinkler(normalizedAddress, normalizedLocation, 0.7, 4)
			
			if score >= threshold {
				// Choose highest score, tie-break by length
				if score > bestScore || (score == bestScore && len(normalizedLocation) > longestKeywordLen) {
					bestScore = score
					longestKeywordLen = len(normalizedLocation)
					bestMatch = &types.MatchResult{
						ZoneID:         zone.ZoneID,
						ZoneName:       fmt.Sprintf("Zone %d", zone.ZoneID),
						MatchedKeyword: location,
						MatchedBy:      "fuzzy",
						Confidence:     score,
						Price:          zone.Price,
					}
				}
			}
		}
	}
	
	return bestMatch
}

// generateSuggestions generates top 3 suggestions based on fuzzy matching
func (m *Matcher) generateSuggestions(normalizedAddress string) []types.MatchSuggestion {
	type scoredSuggestion struct {
		suggestion types.MatchSuggestion
		score      float64
		length     int
	}
	
	var scored []scoredSuggestion
	
	for _, zone := range m.zones {
		for _, location := range zone.Locations {
			normalizedLocation := textnorm.Normalize(location)
			score := smetrics.JaroWinkler(normalizedAddress, normalizedLocation, 0.7, 4)
			
			scored = append(scored, scoredSuggestion{
				suggestion: types.MatchSuggestion{
					ZoneID:     zone.ZoneID,
					Keyword:    location,
					Price:      zone.Price,
					Confidence: score,
				},
				score:  score,
				length: len(normalizedLocation),
			})
		}
	}
	
	// Sort by score (descending), then by length (descending)
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		return scored[i].length > scored[j].length
	})
	
	// Return top 3 suggestions
	var suggestions []types.MatchSuggestion
	for i := 0; i < len(scored) && i < 3; i++ {
		suggestions = append(suggestions, scored[i].suggestion)
	}
	
	return suggestions
}