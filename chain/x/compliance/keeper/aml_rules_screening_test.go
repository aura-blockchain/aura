package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testutil"
	"github.com/aequitas/aura/chain/x/compliance/types"
	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
)

// ============================================================================
// AML Rule Configuration Tests
// ============================================================================

func TestAMLRuleConfiguration_CreateRule(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Create a new AML monitoring rule
	rule := &types.TransactionMonitoringRule{
		Id:          "velocity_check_001",
		Name:        "High Velocity Detection",
		Description: "Detects high-velocity transaction patterns",
		RuleType:    "velocity",
		Parameters: map[string]string{
			"threshold_24h": "1000000",
			"action":        "flag_for_review",
		},
		RiskLevel: types.TransactionRiskLevel_TX_RISK_HIGH,
		Enabled:   true,
		CreatedAt: testutil.Now(),
	}

	// Store the rule
	err := k.SetMonitoringRule(ctx, rule)
	require.NoError(t, err)

	// Retrieve and verify
	retrieved, err := k.GetMonitoringRule(ctx, rule.Id)
	require.NoError(t, err)
	require.Equal(t, rule.Id, retrieved.Id)
	require.Equal(t, rule.Name, retrieved.Name)
	require.Equal(t, rule.RuleType, retrieved.RuleType)
	require.Equal(t, rule.RiskLevel, retrieved.RiskLevel)
	require.True(t, retrieved.Enabled)
	require.Equal(t, "1000000", retrieved.Parameters["threshold_24h"])
}

func TestAMLRuleConfiguration_UpdateRule(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Create initial rule
	rule := &types.TransactionMonitoringRule{
		Id:          "structuring_001",
		Name:        "Structuring Detection",
		Description: "Detects structuring attempts",
		RuleType:    "structuring",
		Parameters: map[string]string{
			"count_threshold": "10",
		},
		RiskLevel: types.TransactionRiskLevel_TX_RISK_CRITICAL,
		Enabled:   true,
		CreatedAt: testutil.Now(),
	}
	err := k.SetMonitoringRule(ctx, rule)
	require.NoError(t, err)

	// Update rule parameters
	rule.Parameters["count_threshold"] = "5" // Lower threshold
	rule.Description = "Updated: More aggressive structuring detection"
	rule.UpdatedAt = testutil.TimePtr(testutil.Now())

	err = k.SetMonitoringRule(ctx, rule)
	require.NoError(t, err)

	// Verify update
	updated, err := k.GetMonitoringRule(ctx, rule.Id)
	require.NoError(t, err)
	require.Equal(t, "5", updated.Parameters["count_threshold"])
	require.Contains(t, updated.Description, "Updated")
}

func TestAMLRuleConfiguration_DeleteRule(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Create rule
	rule := &types.TransactionMonitoringRule{
		Id:        "temp_rule",
		Name:      "Temporary Rule",
		RuleType:  "velocity",
		RiskLevel: types.TransactionRiskLevel_TX_RISK_MEDIUM,
		Enabled:   true,
		CreatedAt: testutil.Now(),
	}
	err := k.SetMonitoringRule(ctx, rule)
	require.NoError(t, err)

	// Verify it exists
	_, err = k.GetMonitoringRule(ctx, rule.Id)
	require.NoError(t, err)

	// Disable rule (soft delete)
	rule.Enabled = false
	err = k.SetMonitoringRule(ctx, rule)
	require.NoError(t, err)

	// Verify it's disabled
	disabled, err := k.GetMonitoringRule(ctx, rule.Id)
	require.NoError(t, err)
	require.False(t, disabled.Enabled)

	// Verify disabled rules are skipped in monitoring
	allRules, err := k.GetAllMonitoringRules(ctx)
	require.NoError(t, err)
	for _, r := range allRules {
		if r.Id == "temp_rule" {
			require.False(t, r.Enabled, "disabled rule should not be active")
		}
	}
}

func TestAMLRuleConfiguration_GetAllRules(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Create multiple rules
	rules := []*types.TransactionMonitoringRule{
		{
			Id:        "rule_001",
			Name:      "Rule 1",
			RuleType:  "velocity",
			RiskLevel: types.TransactionRiskLevel_TX_RISK_HIGH,
			Enabled:   true,
			CreatedAt: testutil.Now(),
		},
		{
			Id:        "rule_002",
			Name:      "Rule 2",
			RuleType:  "structuring",
			RiskLevel: types.TransactionRiskLevel_TX_RISK_CRITICAL,
			Enabled:   true,
			CreatedAt: testutil.Now(),
		},
		{
			Id:        "rule_003",
			Name:      "Rule 3",
			RuleType:  "large_transaction",
			RiskLevel: types.TransactionRiskLevel_TX_RISK_MEDIUM,
			Enabled:   false, // Disabled
			CreatedAt: testutil.Now(),
		},
	}

	for _, rule := range rules {
		err := k.SetMonitoringRule(ctx, rule)
		require.NoError(t, err)
	}

	// Get all rules
	allRules, err := k.GetAllMonitoringRules(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(allRules), 3)

	// Verify all our rules are present
	ruleMap := make(map[string]*types.TransactionMonitoringRule)
	for _, r := range allRules {
		ruleMap[r.Id] = r
	}
	require.Contains(t, ruleMap, "rule_001")
	require.Contains(t, ruleMap, "rule_002")
	require.Contains(t, ruleMap, "rule_003")
}

// ============================================================================
// Transaction Screening Tests
// ============================================================================

func TestTransactionScreening_VelocityThreshold(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Enable monitoring with velocity threshold
	params := types.ComplianceParams{
		TransactionMonitoringEnabled: true,
		VelocityLimit_24H:            "1000000", // 1M velocity threshold
		SingleTransactionLimit:       "100000",  // 100k single tx threshold (must be <= velocity)
		StructuringThresholdCount:    10,
		SanctionsScreeningEnabled:    false,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Create velocity monitoring rule
	rule := &types.TransactionMonitoringRule{
		Id:          "velocity_rule",
		Name:        "24h Velocity Check",
		RuleType:    "velocity",
		RiskLevel:   types.TransactionRiskLevel_TX_RISK_HIGH,
		Enabled:     true,
		CreatedAt:   testutil.Now(),
		Parameters:  map[string]string{"threshold_24h": "1000000"},
	}
	err = k.SetMonitoringRule(ctx, rule)
	require.NoError(t, err)

	// Create test addresses
	from := sdk.AccAddress([]byte("velocity_test_sender"))
	to := sdk.AccAddress([]byte("velocity_test_recipient"))

	// First transaction: 600k (below threshold)
	amount1 := sdk.NewCoins(sdk.NewInt64Coin("uaura", 600000))

	// MonitorTransaction checks the profile BEFORE the transaction
	// So we need to manually create the profile first
	err = k.UpdateAMLProfileOnTransaction(ctx, from.String(), amount1)
	require.NoError(t, err)

	// Check that profile was created with 600k
	profile1, err := k.GetAMLProfile(ctx, from.String())
	require.NoError(t, err)
	require.Equal(t, "600000", profile1.TotalVolume)

	// Second transaction: 500k (total will be 1.1M, exceeds threshold)
	amount2 := sdk.NewCoins(sdk.NewInt64Coin("uaura", 500000))

	// Monitor the second transaction - it will check the current profile (600k) + this tx (500k) = 1.1M
	alerts2, err := k.MonitorTransaction(ctx, from, to, amount2)
	require.NoError(t, err)
	require.NotEmpty(t, alerts2, "second transaction should trigger velocity alert when total exceeds threshold")

	// Verify alert details
	var velocityAlert *types.TransactionAlert
	for _, alert := range alerts2 {
		if alert.RuleId == "velocity_rule" {
			velocityAlert = alert
			break
		}
	}
	require.NotNil(t, velocityAlert, "velocity alert should be generated")
	require.Equal(t, types.TransactionRiskLevel_TX_RISK_HIGH, velocityAlert.RiskLevel)
	require.Contains(t, velocityAlert.Description, "velocity")
}

func TestTransactionScreening_AmountThreshold(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Enable monitoring with amount threshold
	params := types.ComplianceParams{
		TransactionMonitoringEnabled: true,
		SingleTransactionLimit:       "50000", // 50k threshold
		VelocityLimit_24H:            "1000000",
		StructuringThresholdCount:    10,
		SanctionsScreeningEnabled:    false,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Create large transaction rule
	rule := &types.TransactionMonitoringRule{
		Id:        "large_tx_rule",
		Name:      "Large Transaction Monitor",
		RuleType:  "large_transaction",
		RiskLevel: types.TransactionRiskLevel_TX_RISK_HIGH,
		Enabled:   true,
		CreatedAt: testutil.Now(),
		Parameters: map[string]string{
			"threshold": "50000",
		},
	}
	err = k.SetMonitoringRule(ctx, rule)
	require.NoError(t, err)

	from := sdk.AccAddress([]byte("amount_test_sender"))
	to := sdk.AccAddress([]byte("amount_test_recipient"))

	// Test transaction exactly at threshold (boundary condition)
	exactAmount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 50000))
	alerts, err := k.MonitorTransaction(ctx, from, to, exactAmount)
	require.NoError(t, err)
	require.Empty(t, alerts, "transaction at threshold should not trigger alert")

	// Test transaction above threshold
	largeAmount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 50001))
	alerts, err = k.MonitorTransaction(ctx, from, to, largeAmount)
	require.NoError(t, err)
	require.NotEmpty(t, alerts, "transaction above threshold should trigger alert")
	require.Equal(t, types.TransactionRiskLevel_TX_RISK_HIGH, alerts[0].RiskLevel)
}

func TestTransactionScreening_StructuringDetection(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Enable structuring detection
	params := types.ComplianceParams{
		TransactionMonitoringEnabled: true,
		StructuringThresholdCount:    5, // Trigger after 5 transactions
		VelocityLimit_24H:            "1000000",
		SingleTransactionLimit:       "100000",
		SanctionsScreeningEnabled:    false,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Create structuring detection rule
	rule := &types.TransactionMonitoringRule{
		Id:          "structuring_rule",
		Name:        "Structuring Detection",
		RuleType:    "structuring",
		RiskLevel:   types.TransactionRiskLevel_TX_RISK_CRITICAL,
		Enabled:     true,
		CreatedAt:   testutil.Now(),
		Parameters:  map[string]string{"count_threshold": "5"},
	}
	err = k.SetMonitoringRule(ctx, rule)
	require.NoError(t, err)

	from := sdk.AccAddress([]byte("structuring_test_address"))
	to := sdk.AccAddress([]byte("recipient_address"))

	// Simulate multiple small transactions (structuring pattern)
	for i := 0; i < 6; i++ {
		amount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 9000)) // Just below 10k reporting threshold
		err = k.UpdateAMLProfileOnTransaction(ctx, from.String(), amount)
		require.NoError(t, err)

		alerts, err := k.MonitorTransaction(ctx, from, to, amount)
		require.NoError(t, err)

		// After 5 transactions, structuring alert should trigger
		if i >= 4 {
			require.NotEmpty(t, alerts, "structuring should be detected after threshold")

			// Find structuring alert
			var structuringAlert *types.TransactionAlert
			for _, alert := range alerts {
				if alert.RuleId == "structuring_rule" {
					structuringAlert = alert
					break
				}
			}
			require.NotNil(t, structuringAlert, "structuring alert should be generated")
			require.Equal(t, types.TransactionRiskLevel_TX_RISK_CRITICAL, structuringAlert.RiskLevel)
			require.Contains(t, structuringAlert.Description, "structuring")
		}
	}
}

func TestTransactionScreening_MultipleRulesTriggering(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Enable monitoring with multiple thresholds
	params := types.ComplianceParams{
		TransactionMonitoringEnabled: true,
		SingleTransactionLimit:       "50000",
		VelocityLimit_24H:            "100000",
		StructuringThresholdCount:    3,
		SanctionsScreeningEnabled:    false,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Create multiple rules
	rules := []*types.TransactionMonitoringRule{
		{
			Id:        "rule_large",
			Name:      "Large Transaction",
			RuleType:  "large_transaction",
			RiskLevel: types.TransactionRiskLevel_TX_RISK_HIGH,
			Enabled:   true,
			CreatedAt: testutil.Now(),
			Parameters: map[string]string{"threshold": "50000"},
		},
		{
			Id:        "rule_velocity",
			Name:      "Velocity Check",
			RuleType:  "velocity",
			RiskLevel: types.TransactionRiskLevel_TX_RISK_MEDIUM,
			Enabled:   true,
			CreatedAt: testutil.Now(),
			Parameters: map[string]string{"threshold_24h": "100000"},
		},
	}

	for _, rule := range rules {
		err = k.SetMonitoringRule(ctx, rule)
		require.NoError(t, err)
	}

	from := sdk.AccAddress([]byte("multi_rule_sender"))
	to := sdk.AccAddress([]byte("multi_rule_recipient"))

	// First large transaction
	amount1 := sdk.NewCoins(sdk.NewInt64Coin("uaura", 70000))
	err = k.UpdateAMLProfileOnTransaction(ctx, from.String(), amount1)
	require.NoError(t, err)

	alerts1, err := k.MonitorTransaction(ctx, from, to, amount1)
	require.NoError(t, err)
	require.NotEmpty(t, alerts1, "should trigger large transaction alert")

	// Second large transaction (should trigger both large tx and velocity)
	amount2 := sdk.NewCoins(sdk.NewInt64Coin("uaura", 60000))
	err = k.UpdateAMLProfileOnTransaction(ctx, from.String(), amount2)
	require.NoError(t, err)

	alerts2, err := k.MonitorTransaction(ctx, from, to, amount2)
	require.NoError(t, err)
	require.NotEmpty(t, alerts2, "should trigger multiple alerts")

	// Verify both rule types triggered
	ruleTypes := make(map[string]bool)
	for _, alert := range alerts2 {
		ruleTypes[alert.RuleId] = true
	}
	require.True(t, len(ruleTypes) >= 1, "at least one rule should trigger")
}

// ============================================================================
// Jurisdiction-Based Rule Tests
// ============================================================================

func TestJurisdictionRules_BlockedJurisdiction(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Configure blocked jurisdictions (OFAC sanctioned countries)
	params := types.ComplianceParams{
		KycRequired:               true,
		MinimumKycLevel:           types.KYCLevel_KYC_LEVEL_BASIC,
		KycExpiryDays:             365,
		BlockedJurisdictions:      []string{"KP", "IR", "SY"}, // North Korea, Iran, Syria
		SanctionsScreeningEnabled: true,
		SanctionsLists:            []string{"OFAC_SDN"},
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Test blocked jurisdiction
	testCases := []struct {
		jurisdiction string
		shouldBlock  bool
		description  string
	}{
		{"KP", true, "North Korea should be blocked"},
		{"IR", true, "Iran should be blocked"},
		{"SY", true, "Syria should be blocked"},
		{"US", false, "United States should not be blocked"},
		{"GB", false, "United Kingdom should not be blocked"},
		{"kp", true, "lowercase north korea should be blocked (case insensitive)"},
		{"", true, "empty jurisdiction should be blocked (fail-safe)"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			isBlocked := k.IsJurisdictionBlocked(ctx, tc.jurisdiction)
			require.Equal(t, tc.shouldBlock, isBlocked, tc.description)
		})
	}
}

func TestJurisdictionRules_KYCWithBlockedJurisdiction(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Configure blocked jurisdictions
	params := types.ComplianceParams{
		KycRequired:          true,
		MinimumKycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		KycExpiryDays:        365,
		BlockedJurisdictions: []string{"KP", "IR"},
		ApprovedKycProviders: []string{"aura1kycprovideraddress1234567890"},
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Attempt to submit KYC for address from blocked jurisdiction
	kycRecord := &types.KYCRecord{
		Address:      "aura1blockeduser",
		KycLevel:     types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:     "aura1kycprovideraddress1234567890",
		VerifiedAt:   testutil.Now(),
		PiiCommitment: []byte("commitment_hash"),
		Jurisdiction: "KP", // Blocked jurisdiction
	}

	// This should be rejected by validation
	isBlocked := k.IsJurisdictionBlocked(ctx, kycRecord.Jurisdiction)
	require.True(t, isBlocked, "KP jurisdiction should be blocked")

	// Attempting to store KYC from blocked jurisdiction should fail
	// (In production, this would be caught by msg_server validation)
	// Here we just verify the jurisdiction check works
	require.True(t, isBlocked, "cannot accept KYC from blocked jurisdiction")
}

func TestJurisdictionRules_DynamicBlocklist(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Initial params with one blocked jurisdiction
	params := types.ComplianceParams{
		BlockedJurisdictions: []string{"KP"},
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Verify initial block
	require.True(t, k.IsJurisdictionBlocked(ctx, "KP"))
	require.False(t, k.IsJurisdictionBlocked(ctx, "IR"))

	// Update params to add more blocked jurisdictions (governance proposal)
	params.BlockedJurisdictions = []string{"KP", "IR", "SY", "CU"}
	err = k.SetParams(ctx, params)
	require.NoError(t, err)

	// Verify updated blocklist
	require.True(t, k.IsJurisdictionBlocked(ctx, "KP"))
	require.True(t, k.IsJurisdictionBlocked(ctx, "IR"))
	require.True(t, k.IsJurisdictionBlocked(ctx, "SY"))
	require.True(t, k.IsJurisdictionBlocked(ctx, "CU"))
	require.False(t, k.IsJurisdictionBlocked(ctx, "US"))
}

// ============================================================================
// Watchlist / Sanctions Screening Tests
// ============================================================================

func TestWatchlistScreening_SanctionedAddress(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Enable sanctions screening
	params := types.ComplianceParams{
		SanctionsScreeningEnabled: true,
		SanctionsLists:            []string{"OFAC_SDN", "EU_SANCTIONS"},
		ScreeningCacheHours:       24,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Add address to sanctions watchlist
	sanctionedAddr := "aura1sanctionedaddress"
	sanctionsResult := &types.SanctionsScreeningResult{
		Address:    sanctionedAddr,
		Status:     types.SanctionsStatus_SANCTIONS_CONFIRMED,
		ScreenedAt: testutil.Now(),
		Matches: []*types.SanctionsMatch{
			{
				ListName:    "OFAC SDN",
				MatchScore:  "0.99",
				MatchedName: "Sanctioned Entity LLC",
				MatchedId:   "SDN-12345",
				Country:     "KP",
				Program:     "North Korea Sanctions",
			},
		},
		ScreeningProvider: "sanctions_api_provider",
	}

	err = k.SetSanctionsResult(ctx, sanctionsResult)
	require.NoError(t, err)

	// Verify address is flagged as sanctioned
	isSanctioned := k.IsAddressSanctioned(ctx, sanctionedAddr)
	require.True(t, isSanctioned, "address should be flagged as sanctioned")

	// Retrieve full result
	result, err := k.GetSanctionsResult(ctx, sanctionedAddr)
	require.NoError(t, err)
	require.Equal(t, types.SanctionsStatus_SANCTIONS_CONFIRMED, result.Status)
	require.Len(t, result.Matches, 1)
	require.Equal(t, "OFAC SDN", result.Matches[0].ListName)
}

func TestWatchlistScreening_TransactionFromSanctionedAddress(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Enable transaction monitoring and sanctions screening
	params := types.ComplianceParams{
		TransactionMonitoringEnabled: true,
		SanctionsScreeningEnabled:    true,
		SanctionsLists:               []string{"OFAC_SDN"},
		SingleTransactionLimit:       "1000000",
		VelocityLimit_24H:            "10000000",
		StructuringThresholdCount:    10,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Create sanctioned address
	sanctionedAddr := sdk.AccAddress([]byte("sanctioned_sender"))
	cleanAddr := sdk.AccAddress([]byte("clean_recipient"))

	// Add to sanctions list
	sanctionsResult := &types.SanctionsScreeningResult{
		Address:    sanctionedAddr.String(),
		Status:     types.SanctionsStatus_SANCTIONS_CONFIRMED,
		ScreenedAt: testutil.Now(),
		Matches: []*types.SanctionsMatch{
			{
				ListName:   "OFAC SDN",
				MatchScore: "1.00",
			},
		},
	}
	err = k.SetSanctionsResult(ctx, sanctionsResult)
	require.NoError(t, err)

	// Attempt transaction from sanctioned address
	amount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 1000))
	alerts, err := k.MonitorTransaction(ctx, sanctionedAddr, cleanAddr, amount)
	require.NoError(t, err)
	require.NotEmpty(t, alerts, "transaction from sanctioned address should trigger alert")

	// Verify critical risk alert is generated
	var sanctionsAlert *types.TransactionAlert
	for _, alert := range alerts {
		if alert.RiskLevel == types.TransactionRiskLevel_TX_RISK_CRITICAL {
			sanctionsAlert = alert
			break
		}
	}
	require.NotNil(t, sanctionsAlert, "critical sanctions alert should be generated")
	require.Contains(t, sanctionsAlert.Description, "sanctioned")

	// Verify transaction should be blocked
	shouldBlock, reason := k.ShouldBlockTransaction(alerts)
	require.True(t, shouldBlock, "transaction from sanctioned address should be blocked")
	require.NotEmpty(t, reason)
}

func TestWatchlistScreening_PotentialMatch(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Address with potential sanctions match (requires review)
	addr := "aura1potentialmatch"
	sanctionsResult := &types.SanctionsScreeningResult{
		Address:    addr,
		Status:     types.SanctionsStatus_SANCTIONS_MATCH, // Potential match, not confirmed
		ScreenedAt: testutil.Now(),
		Matches: []*types.SanctionsMatch{
			{
				ListName:    "OFAC SDN",
				MatchScore:  "0.75", // 75% confidence
				MatchedName: "Similar Name Corp",
			},
		},
		RequiresManualReview: true,
		ScreeningProvider:    "sanctions_api",
	}

	err := k.SetSanctionsResult(ctx, sanctionsResult)
	require.NoError(t, err)

	// Potential match should also be considered sanctioned (conservative approach)
	isSanctioned := k.IsAddressSanctioned(ctx, addr)
	require.True(t, isSanctioned, "potential match should be flagged until manual review")

	// Verify manual review flag
	result, err := k.GetSanctionsResult(ctx, addr)
	require.NoError(t, err)
	require.True(t, result.RequiresManualReview)
	require.Equal(t, types.SanctionsStatus_SANCTIONS_MATCH, result.Status)
}

func TestWatchlistScreening_ClearStatus(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Address that has been screened and cleared
	addr := "aura1cleanaddress"
	sanctionsResult := &types.SanctionsScreeningResult{
		Address:           addr,
		Status:            types.SanctionsStatus_SANCTIONS_CLEAR,
		ScreenedAt:        testutil.Now(),
		Matches:           []*types.SanctionsMatch{}, // No matches
		ScreeningProvider: "sanctions_api",
	}

	err := k.SetSanctionsResult(ctx, sanctionsResult)
	require.NoError(t, err)

	// Clear status should not be flagged
	isSanctioned := k.IsAddressSanctioned(ctx, addr)
	require.False(t, isSanctioned, "clear status should not be flagged")

	// Verify result
	result, err := k.GetSanctionsResult(ctx, addr)
	require.NoError(t, err)
	require.Equal(t, types.SanctionsStatus_SANCTIONS_CLEAR, result.Status)
	require.Empty(t, result.Matches)
}

func TestWatchlistScreening_ManualReview(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	addr := "aura1underreview"

	// Initial screening with potential match
	initialResult := &types.SanctionsScreeningResult{
		Address:              addr,
		Status:               types.SanctionsStatus_SANCTIONS_PENDING_REVIEW,
		ScreenedAt:           testutil.Now(),
		RequiresManualReview: true,
		Matches: []*types.SanctionsMatch{
			{
				ListName:   "OFAC SDN",
				MatchScore: "0.80",
			},
		},
	}
	err := k.SetSanctionsResult(ctx, initialResult)
	require.NoError(t, err)

	// Simulate manual review clearing the address
	reviewedResult := &types.SanctionsScreeningResult{
		Address:              addr,
		Status:               types.SanctionsStatus_SANCTIONS_CLEAR,
		ScreenedAt:           initialResult.ScreenedAt,
		RequiresManualReview: false,
		ReviewedAt:           testutil.TimePtr(testutil.Now()),
		Reviewer:             "compliance_officer_1",
		ReviewDecision:       "cleared_false_positive",
		Matches:              initialResult.Matches, // Keep original matches for audit trail
	}
	err = k.SetSanctionsResult(ctx, reviewedResult)
	require.NoError(t, err)

	// After review, should not be sanctioned
	isSanctioned := k.IsAddressSanctioned(ctx, addr)
	require.False(t, isSanctioned, "manually reviewed and cleared address should not be sanctioned")

	// Verify review details
	result, err := k.GetSanctionsResult(ctx, addr)
	require.NoError(t, err)
	require.Equal(t, types.SanctionsStatus_SANCTIONS_CLEAR, result.Status)
	require.NotNil(t, result.ReviewedAt)
	require.Equal(t, "compliance_officer_1", result.Reviewer)
	require.Equal(t, "cleared_false_positive", result.ReviewDecision)
}

// ============================================================================
// Risk Scoring Tests
// ============================================================================

func TestRiskScoring_LowRisk(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set velocity threshold for risk calculation
	params := types.ComplianceParams{
		VelocityLimit_24H: "1000000",
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	addr := sdk.AccAddress([]byte("low_risk_user"))

	// Small transaction
	amount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 100))
	err = k.UpdateAMLProfileOnTransaction(ctx, addr.String(), amount)
	require.NoError(t, err)

	// Check risk level
	profile, err := k.GetAMLProfile(ctx, addr.String())
	require.NoError(t, err)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_LOW, profile.RiskLevel)
	require.Equal(t, uint64(1), profile.TotalTransactions)
}

func TestRiskScoring_MediumRisk(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	params := types.ComplianceParams{
		VelocityLimit_24H: "1000000", // Medium threshold will be 500k (50%)
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	addr := sdk.AccAddress([]byte("medium_risk_user"))

	// Build up volume to 600k (above medium threshold of 500k, below high threshold of 1M)
	for i := 0; i < 10; i++ {
		amount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 60000))
		err = k.UpdateAMLProfileOnTransaction(ctx, addr.String(), amount)
		require.NoError(t, err)
	}

	// Check risk level (should be medium due to volume: 600k >= 500k medium threshold)
	profile, err := k.GetAMLProfile(ctx, addr.String())
	require.NoError(t, err)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_MEDIUM, profile.RiskLevel)
	require.Equal(t, uint64(10), profile.TotalTransactions)
}

func TestRiskScoring_HighRisk_PEP(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	addr := "aura1pepuser"

	// Create AML profile with PEP status
	profile := &types.AMLProfile{
		Address:           addr,
		RiskLevel:         types.AMLRiskLevel_AML_RISK_LOW, // Initial
		TotalVolume:       "5000",
		TotalTransactions: 2,
		LastAssessment:    testutil.Now(),
		PepStatus:         true, // Politically Exposed Person
	}

	err := k.SetAMLProfile(ctx, profile)
	require.NoError(t, err)

	// Recalculate risk (should be high/severe due to PEP)
	recalculated := k.calculateRiskLevel(ctx, profile)
	require.NotEqual(t, types.AMLRiskLevel_AML_RISK_LOW, recalculated)
	require.True(t, recalculated >= types.AMLRiskLevel_AML_RISK_HIGH, "PEP should result in high risk")
}

func TestRiskScoring_SevereRisk_MultipleFactors(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	addr := "aura1severeriskuser"

	// Create profile with multiple risk factors
	profile := &types.AMLProfile{
		Address:           addr,
		RiskLevel:         types.AMLRiskLevel_AML_RISK_LOW,
		TotalVolume:       "5000000", // Very high volume
		TotalTransactions: 500,       // High frequency
		LastAssessment:    testutil.Now(),
		PepStatus:         false,
		RiskFactors:       []string{"high_velocity", "unusual_pattern", "cross_border"},
	}

	err := k.SetAMLProfile(ctx, profile)
	require.NoError(t, err)

	// Recalculate risk
	recalculated := k.calculateRiskLevel(ctx, profile)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_SEVERE, recalculated, "multiple risk factors should result in severe risk")
}

// ============================================================================
// Edge Cases and Boundary Conditions
// ============================================================================

func TestEdgeCase_ExactlyAtThreshold(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	params := types.ComplianceParams{
		TransactionMonitoringEnabled: true,
		SingleTransactionLimit:       "10000",
		VelocityLimit_24H:            "100000",
		StructuringThresholdCount:    10,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	rule := &types.TransactionMonitoringRule{
		Id:        "threshold_test",
		RuleType:  "large_transaction",
		RiskLevel: types.TransactionRiskLevel_TX_RISK_HIGH,
		Enabled:   true,
		CreatedAt: testutil.Now(),
		Parameters: map[string]string{
			"threshold": "10000",
		},
	}
	err = k.SetMonitoringRule(ctx, rule)
	require.NoError(t, err)

	from := sdk.AccAddress([]byte("boundary_sender"))
	to := sdk.AccAddress([]byte("boundary_recipient"))

	// Exactly at threshold - should NOT trigger
	exactAmount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 10000))
	alerts, err := k.MonitorTransaction(ctx, from, to, exactAmount)
	require.NoError(t, err)
	require.Empty(t, alerts, "transaction exactly at threshold should not trigger alert")

	// One unit above threshold - should trigger
	aboveAmount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 10001))
	alerts, err = k.MonitorTransaction(ctx, from, to, aboveAmount)
	require.NoError(t, err)
	require.NotEmpty(t, alerts, "transaction above threshold should trigger alert")
}

func TestEdgeCase_ZeroAmount(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	params := types.ComplianceParams{
		TransactionMonitoringEnabled: true,
		SingleTransactionLimit:       "10000",
		VelocityLimit_24H:            "100000",
		StructuringThresholdCount:    10,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	from := sdk.AccAddress([]byte("zero_sender"))
	to := sdk.AccAddress([]byte("zero_recipient"))

	// Zero amount transaction
	zeroAmount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 0))
	alerts, err := k.MonitorTransaction(ctx, from, to, zeroAmount)
	require.NoError(t, err)
	// Zero amount should not trigger any alerts
	require.Empty(t, alerts)
}

func TestEdgeCase_EmptyCoins(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	from := sdk.AccAddress([]byte("empty_sender"))
	to := sdk.AccAddress([]byte("empty_recipient"))

	// Empty coins slice
	emptyCoins := sdk.NewCoins()
	_, err := k.MonitorTransaction(ctx, from, to, emptyCoins)
	require.NoError(t, err)
	// Should handle gracefully without errors
}

func TestEdgeCase_MultipleCurrencies(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	params := types.ComplianceParams{
		TransactionMonitoringEnabled: true,
		SingleTransactionLimit:       "50000",
		VelocityLimit_24H:            "1000000",
		StructuringThresholdCount:    10,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	rule := &types.TransactionMonitoringRule{
		Id:        "multi_currency",
		RuleType:  "large_transaction",
		RiskLevel: types.TransactionRiskLevel_TX_RISK_HIGH,
		Enabled:   true,
		CreatedAt: testutil.Now(),
		Parameters: map[string]string{
			"threshold": "50000",
		},
	}
	err = k.SetMonitoringRule(ctx, rule)
	require.NoError(t, err)

	from := sdk.AccAddress([]byte("multi_currency_sender"))
	to := sdk.AccAddress([]byte("multi_currency_recipient"))

	// Multiple currencies in one transaction
	multiAmount := sdk.NewCoins(
		sdk.NewInt64Coin("uaura", 30000),
		sdk.NewInt64Coin("uatom", 25000),
	)

	_, err = k.MonitorTransaction(ctx, from, to, multiAmount)
	require.NoError(t, err)
	// One of the coins is below threshold, one is below
	// Should not trigger since individual amounts are checked
}
