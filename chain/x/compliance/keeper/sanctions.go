package keeper

import (
	"fmt"
	"strings"
	"time"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// ScreenSanctions performs sanctions screening on an address
// Feature 3: Sanctions screening against OFAC lists
func (k *Keeper) ScreenSanctions(address string, forceRefresh bool) (*types.SanctionsScreeningResult, error) {
	if !k.params.SanctionsScreeningEnabled {
		return &types.SanctionsScreeningResult{
			Address:    address,
			Status:     types.SanctionsClear,
			ScreenedAt: time.Now(),
		}, nil
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Check cache unless force refresh
	if !forceRefresh {
		if lastScreening, exists := k.sanctionsCache[address]; exists {
			cacheExpiry := time.Duration(k.params.ScreeningCacheHours) * time.Hour
			if time.Since(lastScreening) < cacheExpiry {
				if result, exists := k.sanctionsResults[address]; exists {
					return result, nil
				}
			}
		}
	}

	// Perform sanctions screening
	result := &types.SanctionsScreeningResult{
		Address:              address,
		Status:               types.SanctionsClear,
		Matches:              []*types.SanctionsMatch{},
		ScreenedAt:           time.Now(),
		ScreeningProvider:    "internal",
		RequiresManualReview: false,
	}

	// Check each configured sanctions list
	for _, listName := range k.params.SanctionsList {
		matches := k.checkSanctionsList(address, listName)
		if len(matches) > 0 {
			result.Matches = append(result.Matches, matches...)
		}
	}

	// Determine overall status
	if len(result.Matches) > 0 {
		// Check for high-confidence matches
		hasHighConfidence := false
		for _, match := range result.Matches {
			score := parseMatchScore(match.MatchScore)
			if score >= 0.9 { // 90% or higher
				hasHighConfidence = true
				break
			}
		}

		if hasHighConfidence {
			result.Status = types.SanctionsConfirmed
			result.RequiresManualReview = true
		} else {
			result.Status = types.SanctionsPotentialMatch
			result.RequiresManualReview = true
		}
	}

	// Store result and update cache
	k.sanctionsResults[address] = result
	k.sanctionsCache[address] = time.Now()

	// If using external provider, call it
	for providerName, provider := range k.sanctionsProviders {
		externalResult, err := provider.ScreenAddress(address)
		if err == nil && externalResult != nil {
			result.ScreeningProvider = providerName
			result.Matches = append(result.Matches, externalResult.Matches...)
			if externalResult.Status > result.Status {
				result.Status = externalResult.Status
			}
		}
	}

	return result, nil
}

// checkSanctionsList checks an address against a specific sanctions list
func (k *Keeper) checkSanctionsList(address string, listName string) []*types.SanctionsMatch {
	matches := []*types.SanctionsMatch{}

	// In a real implementation, this would check against actual sanctions databases
	// For now, we'll implement the framework with mock data structure

	switch listName {
	case "OFAC_SDN":
		// Office of Foreign Assets Control - Specially Designated Nationals
		matches = append(matches, k.checkOFAC_SDN(address)...)
	case "EU_SANCTIONS":
		// European Union Sanctions Lists
		matches = append(matches, k.checkEU(address)...)
	case "UN_SANCTIONS":
		// United Nations Security Council Sanctions
		matches = append(matches, k.checkUN(address)...)
	case "OFAC_FSE":
		// OFAC Foreign Sanctions Evaders
		matches = append(matches, k.checkOFAC_FSE(address)...)
	case "OFAC_NONSDN":
		// OFAC Non-SDN lists (various programs)
		matches = append(matches, k.checkOFAC_NonSDN(address)...)
	}

	return matches
}

// checkOFAC_SDN checks against OFAC SDN list
func (k *Keeper) checkOFAC_SDN(address string) []*types.SanctionsMatch {
	// In production, this would query the actual OFAC SDN database
	// Implementation would include fuzzy matching algorithms
	matches := []*types.SanctionsMatch{}

	// Example framework for OFAC SDN checking
	// This would be replaced with actual database queries
	sdnPatterns := k.getSDNPatterns()

	for _, pattern := range sdnPatterns {
		if k.matchesPattern(address, pattern) {
			match := &types.SanctionsMatch{
				ListName:    "OFAC SDN",
				MatchScore:  k.calculateMatchScore(address, pattern),
				MatchedName: pattern.Name,
				MatchedID:   pattern.ID,
				Aliases:     pattern.Aliases,
				Country:     pattern.Country,
				Program:     pattern.Program,
				Remarks:     pattern.Remarks,
			}
			matches = append(matches, match)
		}
	}

	return matches
}

// checkEU checks against EU sanctions lists
func (k *Keeper) checkEU(address string) []*types.SanctionsMatch {
	matches := []*types.SanctionsMatch{}
	// Implementation for EU sanctions screening
	// Would integrate with EU consolidated sanctions list
	return matches
}

// checkUN checks against UN sanctions lists
func (k *Keeper) checkUN(address string) []*types.SanctionsMatch {
	matches := []*types.SanctionsMatch{}
	// Implementation for UN sanctions screening
	// Would integrate with UN Security Council sanctions list
	return matches
}

// checkOFAC_FSE checks against OFAC Foreign Sanctions Evaders
func (k *Keeper) checkOFAC_FSE(address string) []*types.SanctionsMatch {
	matches := []*types.SanctionsMatch{}
	// Implementation for OFAC FSE list
	return matches
}

// checkOFAC_NonSDN checks against OFAC Non-SDN lists
func (k *Keeper) checkOFAC_NonSDN(address string) []*types.SanctionsMatch {
	matches := []*types.SanctionsMatch{}
	// Implementation for various OFAC Non-SDN lists
	// Including: SSI, NS-PLC, etc.
	return matches
}

// SanctionsPattern represents a pattern to match against
type SanctionsPattern struct {
	ID       string
	Name     string
	Aliases  []string
	Country  string
	Program  string
	Remarks  string
	Keywords []string
}

// getSDNPatterns returns patterns for SDN matching
func (k *Keeper) getSDNPatterns() []*SanctionsPattern {
	// In production, this would be loaded from the actual OFAC SDN database
	// For demonstration, we return an empty list
	// Real implementation would include:
	// - Regular database updates from OFAC
	// - Fuzzy matching capabilities
	// - Name variation handling
	return []*SanctionsPattern{}
}

// matchesPattern checks if an address matches a sanctions pattern
func (k *Keeper) matchesPattern(address string, pattern *SanctionsPattern) bool {
	// Simple case-insensitive substring matching
	// Production implementation would use sophisticated fuzzy matching
	addressLower := strings.ToLower(address)

	// Check primary name
	if strings.Contains(addressLower, strings.ToLower(pattern.Name)) {
		return true
	}

	// Check aliases
	for _, alias := range pattern.Aliases {
		if strings.Contains(addressLower, strings.ToLower(alias)) {
			return true
		}
	}

	// Check keywords
	for _, keyword := range pattern.Keywords {
		if strings.Contains(addressLower, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

// calculateMatchScore calculates a confidence score for a match
func (k *Keeper) calculateMatchScore(address string, pattern *SanctionsPattern) string {
	// Simple scoring algorithm
	// Production would use sophisticated fuzzy matching scores
	score := 0.0

	addressLower := strings.ToLower(address)
	nameLower := strings.ToLower(pattern.Name)

	if addressLower == nameLower {
		score = 1.0
	} else if strings.Contains(addressLower, nameLower) {
		score = 0.8
	} else {
		score = 0.5
	}

	return fmt.Sprintf("%.2f", score)
}

// parseMatchScore converts a match score string to float
func parseMatchScore(scoreStr string) float64 {
	var score float64
	fmt.Sscanf(scoreStr, "%f", &score)
	return score
}

// GetSanctionsResult retrieves the cached sanctions screening result

// ReviewSanctionsMatch allows manual review of a sanctions match
func (k *Keeper) ReviewSanctionsMatch(address string, reviewer string, decision string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	result, exists := k.sanctionsResults[address]
	if !exists {
		return fmt.Errorf("no sanctions screening result found for address: %s", address)
	}

	result.ReviewedAt = time.Now()
	result.Reviewer = reviewer
	result.ReviewDecision = decision

	// Update status based on review decision
	switch decision {
	case "false_positive":
		result.Status = types.SanctionsClear
		result.RequiresManualReview = false
	case "confirm_match":
		result.Status = types.SanctionsConfirmed
		result.RequiresManualReview = false
	case "escalate":
		// Keep current status but flag for escalation
		result.RequiresManualReview = true
	}

	return nil
}

// BlockSanctionedAddress blocks an address that has confirmed sanctions match
func (k *Keeper) BlockSanctionedAddress(address string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	result, exists := k.sanctionsResults[address]
	if !exists {
		return fmt.Errorf("no sanctions screening result found for address: %s", address)
	}

	if result.Status != types.SanctionsConfirmed {
		return fmt.Errorf("address not confirmed on sanctions list: %s", address)
	}

	// In a real implementation, this would:
	// 1. Freeze the account
	// 2. Report to authorities
	// 3. Document the action
	// 4. Notify compliance officers

	return nil
}

// GetSanctionedAddresses returns all addresses with confirmed sanctions matches
func (k *Keeper) GetSanctionedAddresses() []string {
	k.mu.RLock()
	defer k.mu.RUnlock()

	sanctioned := []string{}
	for address, result := range k.sanctionsResults {
		if result.Status == types.SanctionsConfirmed {
			sanctioned = append(sanctioned, address)
		}
	}

	return sanctioned
}

// GetPendingReviews returns all addresses requiring manual review
func (k *Keeper) GetPendingReviews() []string {
	k.mu.RLock()
	defer k.mu.RUnlock()

	pending := []string{}
	for address, result := range k.sanctionsResults {
		if result.RequiresManualReview && result.Reviewer == "" {
			pending = append(pending, address)
		}
	}

	return pending
}

// ValidateSanctionsCompliance checks if an address is clear for transactions
func (k *Keeper) ValidateSanctionsCompliance(address string) error {
	if !k.params.SanctionsScreeningEnabled {
		return nil
	}

	// Screen if not already screened or cache expired
	result, err := k.ScreenSanctions(address, false)
	if err != nil {
		return fmt.Errorf("sanctions screening failed: %w", err)
	}

	if result.Status == types.SanctionsConfirmed {
		return fmt.Errorf("address is on sanctions list: %s", address)
	}

	if result.Status == types.SanctionsPotentialMatch && result.RequiresManualReview {
		return fmt.Errorf("address has potential sanctions match pending review: %s", address)
	}

	return nil
}

// UpdateSanctionsLists updates the configured sanctions lists
func (k *Keeper) UpdateSanctionsLists(lists []string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	validLists := map[string]bool{
		"OFAC_SDN":     true,
		"OFAC_FSE":     true,
		"OFAC_NONSDN":  true,
		"EU_SANCTIONS": true,
		"UN_SANCTIONS": true,
		"UK_SANCTIONS": true,
	}

	for _, list := range lists {
		if !validLists[list] {
			return fmt.Errorf("invalid sanctions list: %s", list)
		}
	}

	k.params.SanctionsList = lists
	return nil
}

// GetSanctionsStatistics returns statistics about sanctions screening
func (k *Keeper) GetSanctionsStatistics() map[string]interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	totalScreened := len(k.sanctionsResults)
	clear := 0
	matches := 0
	confirmed := 0
	pendingReview := 0

	for _, result := range k.sanctionsResults {
		switch result.Status {
		case types.SanctionsClear:
			clear++
		case types.SanctionsPotentialMatch:
			matches++
		case types.SanctionsConfirmed:
			confirmed++
		}
		if result.RequiresManualReview && result.Reviewer == "" {
			pendingReview++
		}
	}

	return map[string]interface{}{
		"total_screened": totalScreened,
		"clear":          clear,
		"matches":        matches,
		"confirmed":      confirmed,
		"pending_review": pendingReview,
	}
}
