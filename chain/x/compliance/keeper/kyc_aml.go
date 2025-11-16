package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// SubmitKYC submits a KYC verification record
// Feature 1: KYC/AML integration capabilities
func (k *Keeper) SubmitKYC(
	address string,
	kycLevel types.KYCLevel,
	provider string,
	verificationID string,
	documents []string,
	jurisdiction string,
) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Validate inputs
	if address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}
	if verificationID == "" {
		return fmt.Errorf("verification ID cannot be empty")
	}

	now := time.Now()
	expiresAt := now.AddDate(0, 0, int(k.params.KYCExpiryDays))

	// Create KYC record
	record := &types.KYCRecord{
		Address:              address,
		KYCLevel:             kycLevel,
		Provider:             provider,
		VerifiedAt:           now,
		ExpiresAt:            expiresAt,
		VerificationID:       verificationID,
		Documents:            documents,
		Jurisdiction:         jurisdiction,
		EnhancedDueDiligence: false,
		RiskScore:            "0",
	}

	// Validate record
	if err := record.Validate(); err != nil {
		return fmt.Errorf("invalid KYC record: %w", err)
	}

	// Check if provider integration exists
	if kycProvider, exists := k.kycProviders[provider]; exists {
		// Update risk score from provider
		riskScore, err := kycProvider.UpdateRiskScore(address)
		if err == nil {
			record.RiskScore = riskScore
		}
	}

	// Store KYC record
	k.kycRecords[address] = record

	// Initialize or update AML profile
	if _, exists := k.amlProfiles[address]; !exists {
		k.amlProfiles[address] = &types.AMLProfile{
			Address:              address,
			RiskLevel:            types.AMLRiskLow,
			RiskFactors:          []string{},
			LastAssessment:       now,
			TotalTransactions:    0,
			TotalVolume:          "0",
			SuspiciousActivities: []*types.SuspiciousActivity{},
			PEPStatus:            false,
			SourceOfFunds:        []string{},
			Occupation:           "",
		}
	}

	return nil
}

// GetKYCRecord retrieves a KYC record for an address

// ValidateKYCLevel checks if an address meets the minimum KYC level requirement
func (k *Keeper) ValidateKYCLevel(address string) error {
	if !k.params.KYCRequired {
		return nil
	}

	record, err := k.GetKYCRecord(address)
	if err != nil {
		return fmt.Errorf("KYC verification required: %w", err)
	}

	if record.IsExpired() {
		return fmt.Errorf("KYC verification expired on %s", record.ExpiresAt.Format("2006-01-02"))
	}

	if record.KYCLevel < k.params.MinimumKYCLevel {
		return fmt.Errorf("insufficient KYC level: required %d, have %d", k.params.MinimumKYCLevel, record.KYCLevel)
	}

	return nil
}

// UpdateAMLRiskScore updates the AML risk assessment for an address
func (k *Keeper) UpdateAMLRiskScore(address string, transactions uint64, volume string, riskFactors []string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	profile, exists := k.amlProfiles[address]
	if !exists {
		profile = &types.AMLProfile{
			Address:              address,
			RiskLevel:            types.AMLRiskLow,
			RiskFactors:          []string{},
			SuspiciousActivities: []*types.SuspiciousActivity{},
			SourceOfFunds:        []string{},
		}
		k.amlProfiles[address] = profile
	}

	profile.TotalTransactions = transactions
	profile.TotalVolume = volume
	profile.RiskFactors = riskFactors
	profile.LastAssessment = time.Now()

	// Calculate risk level based on factors
	riskLevel := k.calculateAMLRiskLevel(profile)
	profile.RiskLevel = riskLevel

	return nil
}

// calculateAMLRiskLevel determines AML risk level based on profile
func (k *Keeper) calculateAMLRiskLevel(profile *types.AMLProfile) types.AMLRiskLevel {
	riskScore := 0

	// PEP status increases risk
	if profile.PEPStatus {
		riskScore += 20
	}

	// Number of suspicious activities
	riskScore += len(profile.SuspiciousActivities) * 15

	// Number of risk factors
	riskScore += len(profile.RiskFactors) * 10

	// Determine risk level
	switch {
	case riskScore >= 70:
		return types.AMLRiskSevere
	case riskScore >= 50:
		return types.AMLRiskHigh
	case riskScore >= 30:
		return types.AMLRiskMedium
	default:
		return types.AMLRiskLow
	}
}

// GetAMLProfile retrieves the AML profile for an address (legacy in-memory version - deprecated)
// Use keeper_kvstore.go version with sdk.Context instead
// func (k *Keeper) GetAMLProfile(address string) (*types.AMLProfile, error) {
// 	k.mu.RLock()
// 	defer k.mu.RUnlock()
//
// 	profile, exists := k.amlProfiles[address]
// 	if !exists {
// 		return nil, fmt.Errorf("AML profile not found for address: %s", address)
// 	}
//
// 	return profile, nil
// }

// ReportSuspiciousActivity files a suspicious activity report (SAR)
func (k *Keeper) ReportSuspiciousActivity(
	reporter string,
	address string,
	transactionHash string,
	activityType string,
	description string,
	amount string,
	indicators []string,
) (string, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	now := time.Now()

	// Generate unique ID for the suspicious activity
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s:%d", address, transactionHash, activityType, now.Unix())))
	id := hex.EncodeToString(hash[:])[:16]

	activity := &types.SuspiciousActivity{
		ID:              id,
		Address:         address,
		TransactionHash: transactionHash,
		ActivityType:    activityType,
		Description:     description,
		Amount:          amount,
		DetectedAt:      now,
		ReportedAt:      now,
		FiledSAR:        false,
		SARReference:    "",
		Indicators:      indicators,
	}

	if err := activity.Validate(); err != nil {
		return "", fmt.Errorf("invalid suspicious activity: %w", err)
	}

	// Add to AML profile
	profile, exists := k.amlProfiles[address]
	if !exists {
		profile = &types.AMLProfile{
			Address:              address,
			RiskLevel:            types.AMLRiskMedium, // Default to medium for suspicious activity
			RiskFactors:          []string{},
			LastAssessment:       now,
			SuspiciousActivities: []*types.SuspiciousActivity{},
			SourceOfFunds:        []string{},
		}
		k.amlProfiles[address] = profile
	}

	profile.SuspiciousActivities = append(profile.SuspiciousActivities, activity)

	// Update risk factors
	riskFactor := fmt.Sprintf("suspicious_activity_%s", activityType)
	profile.RiskFactors = append(profile.RiskFactors, riskFactor)

	// Recalculate risk level
	profile.RiskLevel = k.calculateAMLRiskLevel(profile)
	profile.LastAssessment = now

	return id, nil
}

// FileSAR files an official Suspicious Activity Report with regulatory authorities
func (k *Keeper) FileSAR(activityID string, sarReference string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Find the suspicious activity across all profiles
	for _, profile := range k.amlProfiles {
		for _, activity := range profile.SuspiciousActivities {
			if activity.ID == activityID {
				activity.FiledSAR = true
				activity.SARReference = sarReference
				return nil
			}
		}
	}

	return fmt.Errorf("suspicious activity not found: %s", activityID)
}

// SetPEPStatus sets the Politically Exposed Person status for an address
func (k *Keeper) SetPEPStatus(address string, isPEP bool) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	profile, exists := k.amlProfiles[address]
	if !exists {
		profile = &types.AMLProfile{
			Address:              address,
			RiskLevel:            types.AMLRiskLow,
			RiskFactors:          []string{},
			SuspiciousActivities: []*types.SuspiciousActivity{},
			SourceOfFunds:        []string{},
		}
		k.amlProfiles[address] = profile
	}

	profile.PEPStatus = isPEP

	if isPEP {
		// Add PEP risk factor
		profile.RiskFactors = append(profile.RiskFactors, "politically_exposed_person")
		// Recalculate risk level
		profile.RiskLevel = k.calculateAMLRiskLevel(profile)
		profile.LastAssessment = time.Now()
	}

	return nil
}

// UpdateSourceOfFunds updates the source of funds information
func (k *Keeper) UpdateSourceOfFunds(address string, sources []string, occupation string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	profile, exists := k.amlProfiles[address]
	if !exists {
		profile = &types.AMLProfile{
			Address:              address,
			RiskLevel:            types.AMLRiskLow,
			RiskFactors:          []string{},
			SuspiciousActivities: []*types.SuspiciousActivity{},
			SourceOfFunds:        []string{},
		}
		k.amlProfiles[address] = profile
	}

	profile.SourceOfFunds = sources
	profile.Occupation = occupation
	profile.LastAssessment = time.Now()

	return nil
}

// RequireEnhancedDueDiligence marks an address as requiring enhanced due diligence
func (k *Keeper) RequireEnhancedDueDiligence(address string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	record, exists := k.kycRecords[address]
	if !exists {
		return fmt.Errorf("KYC record not found for address: %s", address)
	}

	record.EnhancedDueDiligence = true
	return nil
}

// GetHighRiskAddresses returns all addresses with high or severe AML risk
func (k *Keeper) GetHighRiskAddresses() []string {
	k.mu.RLock()
	defer k.mu.RUnlock()

	highRiskAddresses := []string{}
	for address, profile := range k.amlProfiles {
		if profile.RiskLevel >= types.AMLRiskHigh {
			highRiskAddresses = append(highRiskAddresses, address)
		}
	}

	return highRiskAddresses
}

// GetExpiredKYCRecords returns all addresses with expired KYC records
func (k *Keeper) GetExpiredKYCRecords() []string {
	k.mu.RLock()
	defer k.mu.RUnlock()

	expiredAddresses := []string{}
	for address, record := range k.kycRecords {
		if record.IsExpired() {
			expiredAddresses = append(expiredAddresses, address)
		}
	}

	return expiredAddresses
}
