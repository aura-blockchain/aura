package compliance

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/keeper"
	"github.com/aequitas/aura/chain/x/compliance/types"
)

func setupModule(t *testing.T) (AppModule, *keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey("compliance")
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	k := keeper.NewKeeper(cdc, storeKey)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	module := NewAppModule(k)
	return module, k, ctx
}

// ============================================================================
// Module Construction Tests
// ============================================================================

func TestNewAppModule(t *testing.T) {
	module, k, _ := setupModule(t)

	require.NotNil(t, module)
	require.NotNil(t, module.keeper)
	require.Equal(t, k, module.keeper)
}

// ============================================================================
// Module Name Tests
// ============================================================================

func TestAppModule_Name(t *testing.T) {
	module, _, _ := setupModule(t)

	name := module.Name()
	require.Equal(t, "compliance", name)
}

func TestAppModule_Name_NotEmpty(t *testing.T) {
	module, _, _ := setupModule(t)

	name := module.Name()
	require.NotEmpty(t, name)
}

// ============================================================================
// Keeper Access Tests
// ============================================================================

func TestAppModule_GetKeeper(t *testing.T) {
	module, k, _ := setupModule(t)

	retrievedKeeper := module.GetKeeper()
	require.NotNil(t, retrievedKeeper)
	require.Equal(t, k, retrievedKeeper)
}

func TestAppModule_GetKeeper_SameInstance(t *testing.T) {
	module, _, _ := setupModule(t)

	keeper1 := module.GetKeeper()
	keeper2 := module.GetKeeper()
	require.Equal(t, keeper1, keeper2)
}

// ============================================================================
// Genesis Tests
// ============================================================================

func TestAppModule_DefaultGenesis(t *testing.T) {
	module, _, _ := setupModule(t)

	genesis := module.DefaultGenesis()
	require.NotNil(t, genesis)

	// Check default values
	require.False(t, genesis.KycRequired)
	require.Equal(t, types.KYCLevel(0), genesis.MinimumKycLevel)
	require.Equal(t, uint64(365), genesis.KycExpiryDays)
	require.False(t, genesis.TransactionMonitoringEnabled)
	require.Equal(t, "1000000", genesis.VelocityLimit_24H)
	require.Equal(t, "100000", genesis.SingleTransactionLimit)
	require.Equal(t, uint32(10), genesis.StructuringThresholdCount)
	require.False(t, genesis.SanctionsScreeningEnabled)
	require.Empty(t, genesis.SanctionsLists)
	require.Equal(t, uint64(24), genesis.ScreeningCacheHours)
}

func TestAppModule_DefaultGenesis_IsValid(t *testing.T) {
	module, _, _ := setupModule(t)

	genesis := module.DefaultGenesis()
	err := module.ValidateGenesis(genesis)
	require.NoError(t, err)
}

func TestAppModule_ValidateGenesis_Valid(t *testing.T) {
	module, _, _ := setupModule(t)

	params := types.ComplianceParams{
		KycRequired:                  true,
		MinimumKycLevel:              types.KYCLevel_KYC_LEVEL_ADVANCED,
		KycExpiryDays:                180,
		TransactionMonitoringEnabled: true,
		VelocityLimit_24H:            "5000000",
		SingleTransactionLimit:       "500000",
		StructuringThresholdCount:    15,
		SanctionsScreeningEnabled:    true,
		SanctionsLists:               []string{"OFAC", "EU", "UN"},
		ScreeningCacheHours:          48,
		GdprEnabled:                  true,
		DataRetentionDays:            365,
		ProcessingPurposes:           []string{"compliance"},
		TaxReportingEnabled:          true,
		TaxJurisdictions:             []string{"US"},
		TaxYearEnd:                   "12-31",
	}

	err := module.ValidateGenesis(params)
	require.NoError(t, err)
}

func TestAppModule_ValidateGenesis_Empty(t *testing.T) {
	module, _, _ := setupModule(t)

	params := types.ComplianceParams{}
	err := module.ValidateGenesis(params)
	require.NoError(t, err) // All params are optional
}

func TestAppModule_ValidateGenesis_Partial(t *testing.T) {
	module, _, _ := setupModule(t)

	params := types.ComplianceParams{
		KycRequired:     true,
		MinimumKycLevel: types.KYCLevel_KYC_LEVEL_BASIC,
	}

	err := module.ValidateGenesis(params)
	require.NoError(t, err)
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestAppModule_FullLifecycle(t *testing.T) {
	module, k, ctx := setupModule(t)

	// 1. Get module name
	name := module.Name()
	require.Equal(t, "compliance", name)

	// 2. Get default genesis
	genesis := module.DefaultGenesis()
	require.NotNil(t, genesis)

	// 3. Validate genesis
	err := module.ValidateGenesis(genesis)
	require.NoError(t, err)

	// 4. Set params in keeper
	err = k.SetParams(ctx, genesis)
	require.NoError(t, err)

	// 5. Get params from keeper
	params := k.GetParams(ctx)
	require.NotNil(t, params)
}

func TestAppModule_CustomGenesis(t *testing.T) {
	module, k, ctx := setupModule(t)

	// Create custom genesis state
	customGenesis := types.ComplianceParams{
		KycRequired:                  true,
		MinimumKycLevel:              types.KYCLevel_KYC_LEVEL_ADVANCED,
		KycExpiryDays:                180,
		TransactionMonitoringEnabled: true,
		VelocityLimit_24H:            "10000000",
		SingleTransactionLimit:       "1000000",
		StructuringThresholdCount:    20,
		SanctionsScreeningEnabled:    true,
		SanctionsLists:               []string{"OFAC", "EU", "UN", "UK"},
		ScreeningCacheHours:          12,
		GdprEnabled:                  true,
		DataRetentionDays:            365,
		ProcessingPurposes:           []string{"compliance", "analytics"},
		TaxReportingEnabled:          true,
		TaxJurisdictions:             []string{"US", "EU", "UK"},
		TaxYearEnd:                   "12-31",
	}

	// Validate custom genesis
	err := module.ValidateGenesis(customGenesis)
	require.NoError(t, err)

	// Set custom genesis in keeper
	err = k.SetParams(ctx, customGenesis)
	require.NoError(t, err)

	// Retrieve and verify
	params := k.GetParams(ctx)
	require.Equal(t, customGenesis.KycRequired, params.KycRequired)
	require.Equal(t, customGenesis.MinimumKycLevel, params.MinimumKycLevel)
	require.Equal(t, customGenesis.SanctionsLists, params.SanctionsLists)
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestAppModule_MultipleModuleInstances(t *testing.T) {
	module1, k1, _ := setupModule(t)
	module2, k2, _ := setupModule(t)

	// Both should have the same name
	require.Equal(t, module1.Name(), module2.Name())

	// But different keeper instances (compare pointers)
	require.NotSame(t, k1, k2)
	require.NotSame(t, module1.GetKeeper(), module2.GetKeeper())
}

func TestAppModule_GenesisWithAllKYCLevels(t *testing.T) {
	module, _, _ := setupModule(t)

	levels := []types.KYCLevel{
		types.KYCLevel_KYC_LEVEL_UNSPECIFIED,
		types.KYCLevel_KYC_LEVEL_NONE,
		types.KYCLevel_KYC_LEVEL_BASIC,
		types.KYCLevel_KYC_LEVEL_INTERMEDIATE,
		types.KYCLevel_KYC_LEVEL_ADVANCED,
	}

	for _, level := range levels {
		params := types.ComplianceParams{
			MinimumKycLevel: level,
		}
		err := module.ValidateGenesis(params)
		require.NoError(t, err, "level %v should be valid", level)
	}
}

func TestAppModule_GenesisWithEmptySlices(t *testing.T) {
	module, _, _ := setupModule(t)

	params := types.ComplianceParams{
		SanctionsLists:     []string{},
		ProcessingPurposes: []string{},
		TaxJurisdictions:   []string{},
	}

	err := module.ValidateGenesis(params)
	require.NoError(t, err)
}

func TestAppModule_GenesisWithNilSlices(t *testing.T) {
	module, _, _ := setupModule(t)

	params := types.ComplianceParams{
		SanctionsLists:     nil,
		ProcessingPurposes: nil,
		TaxJurisdictions:   nil,
	}

	err := module.ValidateGenesis(params)
	require.NoError(t, err)
}

func TestAppModule_GenesisWithMaxValues(t *testing.T) {
	module, _, _ := setupModule(t)

	params := types.ComplianceParams{
		KycExpiryDays:             18446744073709551615, // max uint64
		StructuringThresholdCount: 4294967295,           // max uint32
		ScreeningCacheHours:       18446744073709551615, // max uint64
		DataRetentionDays:         18446744073709551615, // max uint64
	}

	err := module.ValidateGenesis(params)
	require.NoError(t, err)
}

func TestAppModule_GenesisWithLargeCollections(t *testing.T) {
	module, _, _ := setupModule(t)

	// Create large lists
	largeSanctionsList := make([]string, 100)
	for i := 0; i < 100; i++ {
		largeSanctionsList[i] = "LIST_" + string(rune('0'+i%10))
	}

	largeJurisdictions := make([]string, 100)
	for i := 0; i < 100; i++ {
		largeJurisdictions[i] = "JURISDICTION_" + string(rune('0'+i%10))
	}

	params := types.ComplianceParams{
		SanctionsLists:   largeSanctionsList,
		TaxJurisdictions: largeJurisdictions,
	}

	err := module.ValidateGenesis(params)
	require.NoError(t, err)
}

func TestAppModule_KeeperOperations(t *testing.T) {
	module, k, ctx := setupModule(t)

	// Test that keeper from module works correctly
	moduleKeeper := module.GetKeeper()

	// Set params via module keeper
	params := types.DefaultParams()
	err := moduleKeeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Get params via original keeper
	retrievedParams := k.GetParams(ctx)
	require.Equal(t, params.KycExpiryDays, retrievedParams.KycExpiryDays)
}

// ============================================================================
// Validation Tests
// ============================================================================

func TestAppModule_ValidateGenesis_MultipleScenarios(t *testing.T) {
	module, _, _ := setupModule(t)

	testCases := []struct {
		name   string
		params types.ComplianceParams
		valid  bool
	}{
		{
			name:   "default params",
			params: types.DefaultParams(),
			valid:  true,
		},
		{
			name: "all enabled",
			params: types.ComplianceParams{
				KycRequired:                  true,
				MinimumKycLevel:              types.KYCLevel_KYC_LEVEL_ADVANCED,
				TransactionMonitoringEnabled: true,
				SanctionsScreeningEnabled:    true,
				GdprEnabled:                  true,
				TaxReportingEnabled:          true,
			},
			valid: true,
		},
		{
			name: "all disabled",
			params: types.ComplianceParams{
				KycRequired:                  false,
				TransactionMonitoringEnabled: false,
				SanctionsScreeningEnabled:    false,
				GdprEnabled:                  false,
				TaxReportingEnabled:          false,
			},
			valid: true,
		},
		{
			name: "only KYC",
			params: types.ComplianceParams{
				KycRequired:     true,
				MinimumKycLevel: types.KYCLevel_KYC_LEVEL_BASIC,
				KycExpiryDays:   365,
			},
			valid: true,
		},
		{
			name: "only sanctions",
			params: types.ComplianceParams{
				SanctionsScreeningEnabled: true,
				SanctionsLists:            []string{"OFAC"},
				ScreeningCacheHours:       24,
			},
			valid: true,
		},
		{
			name: "only GDPR",
			params: types.ComplianceParams{
				GdprEnabled:        true,
				DataRetentionDays:  730,
				ProcessingPurposes: []string{"compliance"},
			},
			valid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := module.ValidateGenesis(tc.params)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAppModule_DefaultGenesis_ImmutableAcrossCalls(t *testing.T) {
	module, _, _ := setupModule(t)

	genesis1 := module.DefaultGenesis()
	genesis2 := module.DefaultGenesis()

	// Should return same values
	require.Equal(t, genesis1.KycRequired, genesis2.KycRequired)
	require.Equal(t, genesis1.KycExpiryDays, genesis2.KycExpiryDays)
	require.Equal(t, genesis1.VelocityLimit_24H, genesis2.VelocityLimit_24H)
}
