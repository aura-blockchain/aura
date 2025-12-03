package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/compliance/types"
)

// ============================================================================
// AML Profile Update Tests
// ============================================================================

func TestUpdateAMLProfileOnTransaction_NewProfile(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params for risk calculation
	params := types.ComplianceParams{
		VelocityLimit_24H: "1000000", // 1 million threshold
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test1"
	amount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 1000))

	// Update profile (should create new one)
	err = keeper.UpdateAMLProfileOnTransaction(ctx, address, amount)
	require.NoError(t, err)

	// Verify profile was created
	profile, err := keeper.GetAMLProfile(ctx, address)
	require.NoError(t, err)
	require.NotNil(t, profile)
	require.Equal(t, address, profile.Address)
	require.Equal(t, uint64(1), profile.TotalTransactions)
	require.Equal(t, "1000", profile.TotalVolume)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_LOW, profile.RiskLevel)
}

func TestUpdateAMLProfileOnTransaction_ExistingProfile(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params
	params := types.ComplianceParams{
		VelocityLimit_24H: "1000000",
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test1"

	// Create initial profile
	initialProfile := &types.AMLProfile{
		Address:           address,
		RiskLevel:         types.AMLRiskLevel_AML_RISK_LOW,
		TotalTransactions: 5,
		TotalVolume:       "5000",
		LastAssessment:    timestamppb.New(ctx.BlockTime()),
	}
	err = keeper.SetAMLProfile(ctx, initialProfile)
	require.NoError(t, err)

	// Perform transaction
	amount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 2000))
	err = keeper.UpdateAMLProfileOnTransaction(ctx, address, amount)
	require.NoError(t, err)

	// Verify profile was updated
	profile, err := keeper.GetAMLProfile(ctx, address)
	require.NoError(t, err)
	require.Equal(t, uint64(6), profile.TotalTransactions)
	require.Equal(t, "7000", profile.TotalVolume)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_LOW, profile.RiskLevel)
}

func TestUpdateAMLProfileOnTransaction_MultiDenomination(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params
	params := types.ComplianceParams{
		VelocityLimit_24H: "10000000",
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test1"

	// Transaction with multiple denominations
	amount := sdk.NewCoins(
		sdk.NewInt64Coin("uaura", 1000),
		sdk.NewInt64Coin("stake", 2000),
		sdk.NewInt64Coin("token", 3000),
	)

	err = keeper.UpdateAMLProfileOnTransaction(ctx, address, amount)
	require.NoError(t, err)

	// Verify total volume sums all denominations
	profile, err := keeper.GetAMLProfile(ctx, address)
	require.NoError(t, err)
	require.Equal(t, "6000", profile.TotalVolume) // 1000 + 2000 + 3000
}

func TestUpdateAMLProfileOnTransaction_RiskLevelProgression(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params with low thresholds for testing
	params := types.ComplianceParams{
		VelocityLimit_24H: "10000", // Low threshold for testing
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test1"

	// Start with low-risk profile
	err = keeper.UpdateAMLProfileOnTransaction(ctx, address, sdk.NewCoins(sdk.NewInt64Coin("uaura", 1000)))
	require.NoError(t, err)

	profile, err := keeper.GetAMLProfile(ctx, address)
	require.NoError(t, err)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_LOW, profile.RiskLevel)

	// Add enough volume to trigger MEDIUM risk (>50% of threshold = 5000)
	err = keeper.UpdateAMLProfileOnTransaction(ctx, address, sdk.NewCoins(sdk.NewInt64Coin("uaura", 4500)))
	require.NoError(t, err)

	profile, err = keeper.GetAMLProfile(ctx, address)
	require.NoError(t, err)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_MEDIUM, profile.RiskLevel)

	// Add enough volume to trigger HIGH risk (>=threshold = 10000)
	err = keeper.UpdateAMLProfileOnTransaction(ctx, address, sdk.NewCoins(sdk.NewInt64Coin("uaura", 5000)))
	require.NoError(t, err)

	profile, err = keeper.GetAMLProfile(ctx, address)
	require.NoError(t, err)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_HIGH, profile.RiskLevel)
}

func TestUpdateAMLProfileOnTransaction_EventEmission(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params
	params := types.ComplianceParams{
		VelocityLimit_24H: "5000",
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test1"
	amount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 1000))

	// Clear any existing events
	ctx = ctx.WithEventManager(sdk.NewEventManager())

	// Perform transaction
	err = keeper.UpdateAMLProfileOnTransaction(ctx, address, amount)
	require.NoError(t, err)

	// Verify AML profile updated event
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0)

	foundProfileEvent := false
	for _, event := range events {
		if event.Type == types.EventTypeAMLProfileUpdated {
			foundProfileEvent = true

			// Verify event attributes
			attrs := event.Attributes
			require.Greater(t, len(attrs), 0)

			// Check address attribute
			found := false
			for _, attr := range attrs {
				if attr.Key == types.AttributeKeyAddress {
					require.Equal(t, address, attr.Value)
					found = true
					break
				}
			}
			require.True(t, found, "Address attribute not found in event")
		}
	}
	require.True(t, foundProfileEvent, "AML profile updated event not found")
}

func TestUpdateAMLProfileOnTransaction_RiskLevelChangeEvent(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params with low threshold
	params := types.ComplianceParams{
		VelocityLimit_24H: "5000",
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test1"

	// First transaction (LOW risk)
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	err = keeper.UpdateAMLProfileOnTransaction(ctx, address, sdk.NewCoins(sdk.NewInt64Coin("uaura", 1000)))
	require.NoError(t, err)

	// Second transaction to trigger MEDIUM risk
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	err = keeper.UpdateAMLProfileOnTransaction(ctx, address, sdk.NewCoins(sdk.NewInt64Coin("uaura", 2000)))
	require.NoError(t, err)

	// Verify risk level changed event
	events := ctx.EventManager().Events()
	foundRiskChangeEvent := false
	for _, event := range events {
		if event.Type == types.EventTypeRiskLevelChanged {
			foundRiskChangeEvent = true

			// Verify previous and new risk attributes
			attrs := event.Attributes
			hasPreviousRisk := false
			hasNewRisk := false

			for _, attr := range attrs {
				if attr.Key == types.AttributeKeyPreviousRisk {
					hasPreviousRisk = true
				}
				if attr.Key == types.AttributeKeyNewRisk {
					hasNewRisk = true
				}
			}

			require.True(t, hasPreviousRisk, "Previous risk attribute not found")
			require.True(t, hasNewRisk, "New risk attribute not found")
		}
	}
	require.True(t, foundRiskChangeEvent, "Risk level changed event not found")
}

// ============================================================================
// Risk Level Calculation Tests
// ============================================================================

func TestCalculateRiskLevel_LowRisk(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params
	params := types.ComplianceParams{
		VelocityLimit_24H: "1000000",
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Low volume, low frequency profile
	profile := &types.AMLProfile{
		Address:           "aura1test",
		TotalTransactions: 5,
		TotalVolume:       "10000",
		RiskFactors:       []string{},
		PepStatus:         false,
	}

	riskLevel := keeper.calculateRiskLevel(ctx, profile)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_LOW, riskLevel)
}

func TestCalculateRiskLevel_MediumRiskVolume(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params
	params := types.ComplianceParams{
		VelocityLimit_24H: "1000000", // High threshold: 1 million
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Medium volume (>50% of high threshold)
	profile := &types.AMLProfile{
		Address:           "aura1test",
		TotalTransactions: 10,
		TotalVolume:       "600000", // >500k (50% of 1M)
		RiskFactors:       []string{},
		PepStatus:         false,
	}

	riskLevel := keeper.calculateRiskLevel(ctx, profile)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_MEDIUM, riskLevel)
}

func TestCalculateRiskLevel_MediumRiskFrequency(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params
	params := types.ComplianceParams{
		VelocityLimit_24H: "10000000",
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// High frequency (>50 transactions)
	profile := &types.AMLProfile{
		Address:           "aura1test",
		TotalTransactions: 75, // >50 but <100
		TotalVolume:       "100000",
		RiskFactors:       []string{},
		PepStatus:         false,
	}

	riskLevel := keeper.calculateRiskLevel(ctx, profile)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_MEDIUM, riskLevel)
}

func TestCalculateRiskLevel_MediumRiskFactors(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params
	params := types.ComplianceParams{
		VelocityLimit_24H: "10000000",
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Has risk factors
	profile := &types.AMLProfile{
		Address:           "aura1test",
		TotalTransactions: 10,
		TotalVolume:       "50000",
		RiskFactors:       []string{"unusual_pattern"},
		PepStatus:         false,
	}

	riskLevel := keeper.calculateRiskLevel(ctx, profile)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_MEDIUM, riskLevel)
}

func TestCalculateRiskLevel_HighRiskVolume(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params
	params := types.ComplianceParams{
		VelocityLimit_24H: "1000000",
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// High volume (>=threshold)
	profile := &types.AMLProfile{
		Address:           "aura1test",
		TotalTransactions: 20,
		TotalVolume:       "1500000", // >1M threshold
		RiskFactors:       []string{},
		PepStatus:         false,
	}

	riskLevel := keeper.calculateRiskLevel(ctx, profile)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_HIGH, riskLevel)
}

func TestCalculateRiskLevel_HighRiskFrequency(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params
	params := types.ComplianceParams{
		VelocityLimit_24H: "10000000",
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Very high frequency (>100 transactions)
	profile := &types.AMLProfile{
		Address:           "aura1test",
		TotalTransactions: 150,
		TotalVolume:       "200000",
		RiskFactors:       []string{},
		PepStatus:         false,
	}

	riskLevel := keeper.calculateRiskLevel(ctx, profile)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_HIGH, riskLevel)
}

func TestCalculateRiskLevel_SeverePEPStatus(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params
	params := types.ComplianceParams{
		VelocityLimit_24H: "10000000",
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// PEP status triggers SEVERE
	profile := &types.AMLProfile{
		Address:           "aura1test",
		TotalTransactions: 10,
		TotalVolume:       "50000",
		RiskFactors:       []string{},
		PepStatus:         true, // PEP = SEVERE
	}

	riskLevel := keeper.calculateRiskLevel(ctx, profile)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_SEVERE, riskLevel)
}

func TestCalculateRiskLevel_SevereMultipleRiskFactors(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params
	params := types.ComplianceParams{
		VelocityLimit_24H: "10000000",
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Multiple risk factors trigger SEVERE
	profile := &types.AMLProfile{
		Address:           "aura1test",
		TotalTransactions: 30,
		TotalVolume:       "500000",
		RiskFactors:       []string{"structuring", "layering", "placement"}, // >=3 factors
		PepStatus:         false,
	}

	riskLevel := keeper.calculateRiskLevel(ctx, profile)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_SEVERE, riskLevel)
}

func TestCalculateRiskLevel_DefaultThreshold(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Don't set params (or set invalid threshold)
	params := types.ComplianceParams{
		VelocityLimit_24H: "", // Invalid/empty threshold
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Should use default threshold (1 million)
	profile := &types.AMLProfile{
		Address:           "aura1test",
		TotalTransactions: 10,
		TotalVolume:       "2000000", // >1M default
		RiskFactors:       []string{},
		PepStatus:         false,
	}

	riskLevel := keeper.calculateRiskLevel(ctx, profile)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_HIGH, riskLevel)
}

func TestCalculateRiskLevel_InvalidVolumeString(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params
	params := types.ComplianceParams{
		VelocityLimit_24H: "1000000",
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Invalid volume string should be treated as zero
	profile := &types.AMLProfile{
		Address:           "aura1test",
		TotalTransactions: 5,
		TotalVolume:       "invalid", // Invalid number
		RiskFactors:       []string{},
		PepStatus:         false,
	}

	riskLevel := keeper.calculateRiskLevel(ctx, profile)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_LOW, riskLevel) // Should default to LOW
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestAMLProfileUpdate_CompleteFlow(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params
	params := types.ComplianceParams{
		VelocityLimit_24H: "100000",
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test1"

	// Simulate multiple transactions and track risk progression
	testCases := []struct {
		name              string
		amount            sdk.Coins
		expectedTxCount   uint64
		expectedRiskLevel types.AMLRiskLevel
	}{
		{
			name:              "First transaction - LOW risk",
			amount:            sdk.NewCoins(sdk.NewInt64Coin("uaura", 10000)),
			expectedTxCount:   1,
			expectedRiskLevel: types.AMLRiskLevel_AML_RISK_LOW,
		},
		{
			name:              "Second transaction - still LOW risk",
			amount:            sdk.NewCoins(sdk.NewInt64Coin("uaura", 15000)),
			expectedTxCount:   2,
			expectedRiskLevel: types.AMLRiskLevel_AML_RISK_LOW,
		},
		{
			name:              "Third transaction - MEDIUM risk (>50k)",
			amount:            sdk.NewCoins(sdk.NewInt64Coin("uaura", 30000)),
			expectedTxCount:   3,
			expectedRiskLevel: types.AMLRiskLevel_AML_RISK_MEDIUM,
		},
		{
			name:              "Fourth transaction - HIGH risk (>100k)",
			amount:            sdk.NewCoins(sdk.NewInt64Coin("uaura", 50000)),
			expectedTxCount:   4,
			expectedRiskLevel: types.AMLRiskLevel_AML_RISK_HIGH,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := keeper.UpdateAMLProfileOnTransaction(ctx, address, tc.amount)
			require.NoError(t, err)

			profile, err := keeper.GetAMLProfile(ctx, address)
			require.NoError(t, err)
			require.Equal(t, tc.expectedTxCount, profile.TotalTransactions)
			require.Equal(t, tc.expectedRiskLevel, profile.RiskLevel)
		})
	}
}

func TestAMLProfileUpdate_ConcurrentAddresses(t *testing.T) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	ctx := input.Ctx

	// Set params
	params := types.ComplianceParams{
		VelocityLimit_24H: "1000000",
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Update multiple addresses
	addresses := []string{"aura1test1", "aura1test2", "aura1test3"}

	for _, addr := range addresses {
		err := keeper.UpdateAMLProfileOnTransaction(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("uaura", 5000)))
		require.NoError(t, err)
	}

	// Verify each profile is independent
	for _, addr := range addresses {
		profile, err := keeper.GetAMLProfile(ctx, addr)
		require.NoError(t, err)
		require.Equal(t, addr, profile.Address)
		require.Equal(t, uint64(1), profile.TotalTransactions)
		require.Equal(t, "5000", profile.TotalVolume)
	}
}
