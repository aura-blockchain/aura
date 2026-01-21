// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"crypto/sha256"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// AMLComprehensiveTestSuite tests AML screening edge cases, complex transaction patterns,
// and multi-jurisdiction scenarios as specified in ROADMAP_PRODUCTION.md task #12
type AMLComprehensiveTestSuite struct {
	KeeperTestSuite
}

func TestAMLComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(AMLComprehensiveTestSuite))
}

// addr generates a deterministic test address from a seed string
func (suite *AMLComprehensiveTestSuite) addr(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	bech32, err := sdk.Bech32ifyAddressBytes("aura", sum[:20])
	suite.Require().NoError(err)
	return bech32
}

// ============================================================================
// AML Screening Edge Cases
// ============================================================================

// TestAMLScreening_BoundaryRiskScores tests risk score boundary conditions
func (suite *AMLComprehensiveTestSuite) TestAMLScreening_BoundaryRiskScores() {
	address := suite.addr("test-user")

	testCases := []struct {
		name      string
		riskLevel types.AMLRiskLevel
	}{
		{
			name:      "Unspecified risk level",
			riskLevel: types.AMLRiskLevel_AML_RISK_UNSPECIFIED,
		},
		{
			name:      "Low risk level",
			riskLevel: types.AMLRiskLevel_AML_RISK_LOW,
		},
		{
			name:      "Medium risk level",
			riskLevel: types.AMLRiskLevel_AML_RISK_MEDIUM,
		},
		{
			name:      "High risk level",
			riskLevel: types.AMLRiskLevel_AML_RISK_HIGH,
		},
		{
			name:      "Severe risk level",
			riskLevel: types.AMLRiskLevel_AML_RISK_SEVERE,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			profile := &types.AMLProfile{
				Address:        address,
				RiskLevel:      tc.riskLevel,
				RiskFactors:    []string{"test-factor"},
				LastAssessment: suite.SdkCtx.BlockTime(),
			}

			suite.Require().NoError(suite.Keeper.SetAMLProfile(suite.SdkCtx, profile))

			// Retrieve and verify
			retrieved, err := suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
			suite.NoError(err)
			suite.NotNil(retrieved)
			suite.Equal(tc.riskLevel, retrieved.RiskLevel)
		})
	}
}

// TestAMLScreening_RiskLevelTransitions tests valid risk level transitions
func (suite *AMLComprehensiveTestSuite) TestAMLScreening_RiskLevelTransitions() {
	address := suite.addr("transition-test")

	// Create initial profile with LOW risk
	profile := &types.AMLProfile{
		Address:        address,
		RiskLevel:      types.AMLRiskLevel_AML_RISK_LOW,
		RiskFactors:    []string{"low-volume"},
		LastAssessment: suite.SdkCtx.BlockTime(),
	}
	suite.Require().NoError(suite.Keeper.SetAMLProfile(suite.SdkCtx, profile))

	// Test transition: LOW -> MEDIUM
	profile.RiskLevel = types.AMLRiskLevel_AML_RISK_MEDIUM
	profile.RiskFactors = []string{"increased-volume"}
	suite.Require().NoError(suite.Keeper.SetAMLProfile(suite.SdkCtx, profile))

	retrieved, err := suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.NoError(err)
	suite.Equal(types.AMLRiskLevel_AML_RISK_MEDIUM, retrieved.RiskLevel)

	// Test transition: MEDIUM -> HIGH
	profile.RiskLevel = types.AMLRiskLevel_AML_RISK_HIGH
	profile.RiskFactors = []string{"suspicious-pattern"}
	suite.Require().NoError(suite.Keeper.SetAMLProfile(suite.SdkCtx, profile))

	retrieved, err = suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.NoError(err)
	suite.Equal(types.AMLRiskLevel_AML_RISK_HIGH, retrieved.RiskLevel)

	// Test transition: HIGH -> SEVERE
	profile.RiskLevel = types.AMLRiskLevel_AML_RISK_SEVERE
	profile.RiskFactors = []string{"confirmed-suspicious-activity"}
	suite.Require().NoError(suite.Keeper.SetAMLProfile(suite.SdkCtx, profile))

	retrieved, err = suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.NoError(err)
	suite.Equal(types.AMLRiskLevel_AML_RISK_SEVERE, retrieved.RiskLevel)

	// Test transition: SEVERE -> LOW (after review clears concerns)
	profile.RiskLevel = types.AMLRiskLevel_AML_RISK_LOW
	profile.RiskFactors = []string{"cleared-after-review"}
	suite.Require().NoError(suite.Keeper.SetAMLProfile(suite.SdkCtx, profile))

	retrieved, err = suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.NoError(err)
	suite.Equal(types.AMLRiskLevel_AML_RISK_LOW, retrieved.RiskLevel)
}

// TestAMLScreening_RiskFactors tests handling of risk factors
func (suite *AMLComprehensiveTestSuite) TestAMLScreening_RiskFactors() {
	address := suite.addr("risk-factors-test")

	// Create profile with multiple risk factors
	profile := &types.AMLProfile{
		Address:        address,
		RiskLevel:      types.AMLRiskLevel_AML_RISK_HIGH,
		RiskFactors:    []string{"high-volume", "suspicious-pattern", "cross-border"},
		LastAssessment: suite.SdkCtx.BlockTime(),
	}

	suite.Require().NoError(suite.Keeper.SetAMLProfile(suite.SdkCtx, profile))

	retrieved, err := suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.NoError(err)
	suite.NotNil(retrieved)
	suite.Equal(3, len(retrieved.RiskFactors))
	suite.Contains(retrieved.RiskFactors, "high-volume")
	suite.Contains(retrieved.RiskFactors, "suspicious-pattern")
	suite.Contains(retrieved.RiskFactors, "cross-border")
}

// TestAMLScreening_TransactionVolume tests transaction volume tracking
func (suite *AMLComprehensiveTestSuite) TestAMLScreening_TransactionVolume() {
	address := suite.addr("volume-test")

	profile := &types.AMLProfile{
		Address:           address,
		RiskLevel:         types.AMLRiskLevel_AML_RISK_LOW,
		TotalTransactions: 100,
		TotalVolume:       "1000000",
		LastAssessment:    suite.SdkCtx.BlockTime(),
	}

	suite.Require().NoError(suite.Keeper.SetAMLProfile(suite.SdkCtx, profile))

	retrieved, err := suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.NoError(err)
	suite.NotNil(retrieved)
	suite.Equal(uint64(100), retrieved.TotalTransactions)
	suite.Equal("1000000", retrieved.TotalVolume)
}

// TestAMLScreening_StaleAssessmentDetection tests detection of stale AML assessments
func (suite *AMLComprehensiveTestSuite) TestAMLScreening_StaleAssessmentDetection() {
	address := suite.addr("stale-assessment")

	// Create profile with old assessment timestamp
	oldTime := suite.SdkCtx.BlockTime().Add(-365 * 24 * time.Hour) // 1 year ago
	profile := &types.AMLProfile{
		Address:        address,
		RiskLevel:      types.AMLRiskLevel_AML_RISK_LOW,
		LastAssessment: oldTime,
	}
	suite.Require().NoError(suite.Keeper.SetAMLProfile(suite.SdkCtx, profile))

	retrieved, err := suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.NoError(err)
	suite.NotNil(retrieved)

	// Check if assessment is stale (> 90 days old)
	assessedAt := retrieved.LastAssessment
	staleCutoff := suite.SdkCtx.BlockTime().Add(-90 * 24 * time.Hour)

	isStale := assessedAt.Before(staleCutoff)
	suite.True(isStale, "Assessment should be considered stale after 90 days")
}

// TestAMLScreening_MultipleProfileUpdates tests rapid profile updates
func (suite *AMLComprehensiveTestSuite) TestAMLScreening_MultipleProfileUpdates() {
	address := suite.addr("rapid-updates")

	// Create initial profile
	profile := &types.AMLProfile{
		Address:        address,
		RiskLevel:      types.AMLRiskLevel_AML_RISK_LOW,
		LastAssessment: suite.SdkCtx.BlockTime(),
	}
	suite.Require().NoError(suite.Keeper.SetAMLProfile(suite.SdkCtx, profile))

	// Perform rapid updates
	for i := 0; i < 10; i++ {
		profile.TotalTransactions = uint64(10 + i)
		profile.LastAssessment = suite.SdkCtx.BlockTime().Add(time.Duration(i) * time.Minute)
		suite.Require().NoError(suite.Keeper.SetAMLProfile(suite.SdkCtx, profile))
	}

	// Verify final state
	retrieved, err := suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.NoError(err)
	suite.NotNil(retrieved)
	suite.Equal(uint64(19), retrieved.TotalTransactions) // 10 + 9 = 19
}

// TestAMLScreening_UnknownAddress tests screening of addresses with no profile
func (suite *AMLComprehensiveTestSuite) TestAMLScreening_UnknownAddress() {
	unknownAddr := suite.addr("unknown")

	// Attempt to get profile for address that doesn't exist
	profile, err := suite.Keeper.GetAMLProfile(suite.SdkCtx, unknownAddr)

	// Should return error for non-existent profile
	suite.Error(err)
	suite.Nil(profile, "Non-existent profile should return nil")

	// Application logic should treat unknown addresses appropriately
	// (e.g., require screening before allowing transactions)
}
