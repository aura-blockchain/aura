package keeper_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/wasm/keeper"
	"github.com/aequitas/aura/chain/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.BinaryCodec
}

func (suite *GenesisTestSuite) SetupTest() {
	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create test context
	testCtx := testutil.DefaultContextWithDB(suite.T(), storeKey, storetypes.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx

	// Create codec
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	suite.cdc = codec.NewProtoCodec(registry)

	// Create keeper (without wasmd keeper for unit tests)
	suite.keeper = keeper.NewKeeper(
		suite.cdc,
		storeKey,
		nil, // wasmd keeper not needed for genesis tests
		"aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
	)
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestInitGenesis() {
	ctx := suite.ctx

	// Test: InitGenesis with default/empty state should not panic
	suite.NotPanics(func() {
		genState := types.DefaultGenesisState()
		err := suite.keeper.InitGenesis(ctx, *genState)
		suite.NoError(err)
	}, "InitGenesis should not panic with empty state")
}

func (suite *GenesisTestSuite) TestExportGenesis() {
	ctx := suite.ctx

	// Test: ExportGenesis should not panic
	suite.NotPanics(func() {
		genState := suite.keeper.ExportGenesis(ctx)
		suite.NotNil(genState)
	}, "ExportGenesis should not panic")
}

func (suite *GenesisTestSuite) TestGenesisRoundTrip() {
	ctx := suite.ctx

	// Set up some state
	params := types.Params{
		CodeUploadAccess: types.AccessConfig{
			Permission: types.AccessTypeEverybody,
		},
		InstantiateDefaultPermission: types.AccessTypeEverybody,
		MaxWasmCodeSize:              1024 * 1024,
		MaxGasWasmExecution:          3_000_000,
		SecurityAnalysisEnabled:      true,
		RequireAdminForMigrate:       true,
	}
	err := suite.keeper.SetParams(ctx, params)
	suite.NoError(err)

	// Export genesis
	exported := suite.keeper.ExportGenesis(ctx)
	suite.NotNil(exported)

	// Create new keeper for import
	newStoreKey := storetypes.NewKVStoreKey("test-new")
	newTestCtx := testutil.DefaultContextWithDB(suite.T(), newStoreKey, storetypes.NewTransientStoreKey("transient_import"))
	newCtx := newTestCtx.Ctx
	newKeeper := keeper.NewKeeper(suite.cdc, newStoreKey, nil, "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")

	// Import genesis
	err = newKeeper.InitGenesis(newCtx, *exported)
	suite.NoError(err)

	// Verify state
	importedParams := newKeeper.GetParams(newCtx)
	suite.Equal(params.MaxWasmCodeSize, importedParams.MaxWasmCodeSize)
}

func (suite *GenesisTestSuite) TestInitGenesisWithValidData() {
	ctx := suite.ctx

	// Create valid genesis state
	genState := &types.GenesisState{
		Params: types.Params{
			CodeUploadAccess: types.AccessConfig{
				Permission: types.AccessTypeEverybody,
			},
			InstantiateDefaultPermission: types.AccessTypeEverybody,
			MaxWasmCodeSize:              600 * 1024,
			MaxGasWasmExecution:          2_000_000,
			SecurityAnalysisEnabled:      true,
			RequireAdminForMigrate:       false,
		},
		AuthorizedUploaders: []string{"aura1abc123", "aura1def456"},
		PausedContracts:     []string{},
		SecurityStats: types.SecurityStats{
			TotalCodesAnalyzed: 0,
			CodesRejected:      0,
			ContractsPaused:    0,
			TotalExecutions:    0,
			FailedExecutions:   0,
			GasConsumedTotal:   0,
			LastSecurityScan:   0,
		},
	}

	err := suite.keeper.InitGenesis(ctx, *genState)
	suite.NoError(err)

	// Verify state was set correctly
	params := suite.keeper.GetParams(ctx)
	suite.Equal(genState.Params.MaxWasmCodeSize, params.MaxWasmCodeSize)
}

func (suite *GenesisTestSuite) TestInitGenesisWithInvalidData() {
	ctx := suite.ctx

	// Create genesis state with invalid params
	genState := &types.GenesisState{
		Params: types.Params{
			CodeUploadAccess: types.AccessConfig{
				Permission: types.AccessTypeEverybody,
			},
			InstantiateDefaultPermission: types.AccessTypeEverybody,
			MaxWasmCodeSize:              0, // Invalid - should be positive
			MaxGasWasmExecution:          2_000_000,
			SecurityAnalysisEnabled:      true,
			RequireAdminForMigrate:       false,
		},
		AuthorizedUploaders: []string{},
		PausedContracts:     []string{},
		SecurityStats:       types.SecurityStats{},
	}

	err := suite.keeper.InitGenesis(ctx, *genState)
	suite.Error(err)
	suite.Contains(err.Error(), "max wasm code size must be positive")
}

func (suite *GenesisTestSuite) TestDefaultGenesis() {
	// Test: Default genesis should be valid
	genState := types.DefaultGenesisState()
	suite.NotNil(genState)

	// Validate it can be used for InitGenesis
	err := suite.keeper.InitGenesis(suite.ctx, *genState)
	suite.NoError(err)

	// Verify default params
	params := suite.keeper.GetParams(suite.ctx)
	suite.Equal(uint64(600*1024), params.MaxWasmCodeSize)
	suite.Equal(uint64(10_000_000), params.MaxGasWasmExecution)
}
