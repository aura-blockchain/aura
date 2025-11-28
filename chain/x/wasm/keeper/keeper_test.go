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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.BinaryCodec
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
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
		nil, // wasmd keeper not needed for params/auth tests
		"aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
	)
}

func (suite *KeeperTestSuite) TestParams() {
	// Test default params
	params := suite.keeper.GetParams(suite.ctx)
	suite.Require().Equal(uint64(600*1024), params.MaxContractSize)
	suite.Require().Equal(uint64(2_000_000), params.MaxInstantiateGas)
	suite.Require().Equal(uint64(1_000_000), params.MaxExecuteGas)
	suite.Require().Equal(uint64(100_000), params.MaxQueryGas)
	suite.Require().True(params.RequireAuthorization)
	suite.Require().False(params.EnableMigration)

	// Test setting custom params
	newParams := types.Params{
		MaxContractSize:         1024 * 1024, // 1MB
		MaxInstantiateGas:       3_000_000,
		MaxExecuteGas:           2_000_000,
		MaxQueryGas:             200_000,
		RequireAuthorization:    false,
		EnableMigration:         true,
		MaxContractSizePerBlock: 10 * 1024 * 1024,
	}

	err := suite.keeper.SetParams(suite.ctx, newParams)
	suite.Require().NoError(err)

	// Verify params were set
	params = suite.keeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams.MaxContractSize, params.MaxContractSize)
	suite.Require().Equal(newParams.MaxInstantiateGas, params.MaxInstantiateGas)
	suite.Require().Equal(newParams.MaxExecuteGas, params.MaxExecuteGas)
	suite.Require().Equal(newParams.MaxQueryGas, params.MaxQueryGas)
	suite.Require().Equal(newParams.RequireAuthorization, params.RequireAuthorization)
	suite.Require().Equal(newParams.EnableMigration, params.EnableMigration)
}

func (suite *KeeperTestSuite) TestParamsValidation() {
	// Test invalid params
	invalidParams := types.Params{
		MaxContractSize:         0, // Invalid
		MaxInstantiateGas:       2_000_000,
		MaxExecuteGas:           1_000_000,
		MaxQueryGas:             100_000,
		RequireAuthorization:    true,
		EnableMigration:         false,
		MaxContractSizePerBlock: 5 * 1024 * 1024,
	}

	err := suite.keeper.SetParams(suite.ctx, invalidParams)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "max contract size must be positive")

	// Test contract size too large
	invalidParams.MaxContractSize = 11 * 1024 * 1024 // Over 10MB
	err = suite.keeper.SetParams(suite.ctx, invalidParams)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot exceed 10MB")
}

func (suite *KeeperTestSuite) TestAuthorizeUploader() {
	address := "aura1abc123def456ghi789jkl012mno345pqr678st"

	// Initially not authorized
	suite.Require().False(suite.keeper.IsAuthorizedUploader(suite.ctx, address))

	// Authorize
	err := suite.keeper.AuthorizeUploader(suite.ctx, address)
	suite.Require().NoError(err)

	// Should now be authorized
	suite.Require().True(suite.keeper.IsAuthorizedUploader(suite.ctx, address))

	// Revoke
	err = suite.keeper.RevokeUploader(suite.ctx, address)
	suite.Require().NoError(err)

	// Should no longer be authorized
	suite.Require().False(suite.keeper.IsAuthorizedUploader(suite.ctx, address))
}

func (suite *KeeperTestSuite) TestAuthorizationWithParams() {
	address := "aura1abc123def456ghi789jkl012mno345pqr678st"

	// Set params to not require authorization
	params := suite.keeper.GetParams(suite.ctx)
	params.RequireAuthorization = false
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	// Anyone should be authorized
	suite.Require().True(suite.keeper.IsAuthorizedUploader(suite.ctx, address))

	// Set params to require authorization
	params.RequireAuthorization = true
	err = suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	// Should not be authorized
	suite.Require().False(suite.keeper.IsAuthorizedUploader(suite.ctx, address))
}

func (suite *KeeperTestSuite) TestPauseContract() {
	contractAddr := "aura14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr"

	// Initially not paused
	suite.Require().False(suite.keeper.IsContractPaused(suite.ctx, contractAddr))

	// Pause
	err := suite.keeper.PauseContract(suite.ctx, contractAddr)
	suite.Require().NoError(err)

	// Should now be paused
	suite.Require().True(suite.keeper.IsContractPaused(suite.ctx, contractAddr))

	// Verify stats updated
	stats := suite.keeper.GetSecurityStats(suite.ctx)
	suite.Require().Equal(uint64(1), stats.TotalPausedContracts)

	// Unpause
	err = suite.keeper.UnpauseContract(suite.ctx, contractAddr)
	suite.Require().NoError(err)

	// Should no longer be paused
	suite.Require().False(suite.keeper.IsContractPaused(suite.ctx, contractAddr))

	// Verify stats updated
	stats = suite.keeper.GetSecurityStats(suite.ctx)
	suite.Require().Equal(uint64(0), stats.TotalPausedContracts)
}

func (suite *KeeperTestSuite) TestSecurityStats() {
	// Get initial stats
	stats := suite.keeper.GetSecurityStats(suite.ctx)
	suite.Require().Equal(uint64(0), stats.TotalContractsUploaded)
	suite.Require().Equal(uint64(0), stats.TotalContractsInstantiated)
	suite.Require().Equal(uint64(0), stats.TotalExecutions)

	// Update stats by calling internal methods
	// (In real usage, these are updated by the keeper methods)
	newStats := types.SecurityStats{
		TotalContractsUploaded:    5,
		TotalContractsInstantiated: 10,
		TotalExecutions:           100,
		TotalPausedContracts:      2,
		ReentrancyAttemptsBlocked: 3,
	}
	suite.keeper.SetSecurityStats(suite.ctx, newStats)

	// Verify stats were set
	stats = suite.keeper.GetSecurityStats(suite.ctx)
	suite.Require().Equal(uint64(5), stats.TotalContractsUploaded)
	suite.Require().Equal(uint64(10), stats.TotalContractsInstantiated)
	suite.Require().Equal(uint64(100), stats.TotalExecutions)
	suite.Require().Equal(uint64(2), stats.TotalPausedContracts)
	suite.Require().Equal(uint64(3), stats.ReentrancyAttemptsBlocked)
}

func (suite *KeeperTestSuite) TestValidateContractUpload() {
	address := "aura1abc123def456ghi789jkl012mno345pqr678st"

	// Authorize the uploader
	err := suite.keeper.AuthorizeUploader(suite.ctx, address)
	suite.Require().NoError(err)

	// Valid contract code
	validCode := make([]byte, 1024) // 1KB
	err = suite.keeper.ValidateContractUpload(suite.ctx, address, validCode)
	suite.Require().NoError(err)

	// Empty contract code
	err = suite.keeper.ValidateContractUpload(suite.ctx, address, []byte{})
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot be empty")

	// Contract too large
	params := suite.keeper.GetParams(suite.ctx)
	largeCode := make([]byte, params.MaxContractSize+1)
	err = suite.keeper.ValidateContractUpload(suite.ctx, address, largeCode)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "exceeds maximum")

	// Unauthorized uploader
	unauthorizedAddr := "aura1xyz789abc456def123ghi456jkl789mno012pq"
	err = suite.keeper.ValidateContractUpload(suite.ctx, unauthorizedAddr, validCode)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "not authorized")
}

func (suite *KeeperTestSuite) TestValidateContractExecution() {
	contractAddr := "aura14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr"

	// Should be valid initially
	err := suite.keeper.ValidateContractExecution(suite.ctx, contractAddr)
	suite.Require().NoError(err)

	// Pause the contract
	err = suite.keeper.PauseContract(suite.ctx, contractAddr)
	suite.Require().NoError(err)

	// Should now fail validation
	err = suite.keeper.ValidateContractExecution(suite.ctx, contractAddr)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "is paused")

	// Unpause
	err = suite.keeper.UnpauseContract(suite.ctx, contractAddr)
	suite.Require().NoError(err)

	// Should be valid again
	err = suite.keeper.ValidateContractExecution(suite.ctx, contractAddr)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestGenesisExportImport() {
	// Set up some state
	params := types.Params{
		MaxContractSize:         1024 * 1024,
		MaxInstantiateGas:       3_000_000,
		MaxExecuteGas:           2_000_000,
		MaxQueryGas:             200_000,
		RequireAuthorization:    true,
		EnableMigration:         true,
		MaxContractSizePerBlock: 10 * 1024 * 1024,
	}
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	// Authorize some uploaders
	uploader1 := "aura1abc123"
	uploader2 := "aura1def456"
	err = suite.keeper.AuthorizeUploader(suite.ctx, uploader1)
	suite.Require().NoError(err)
	err = suite.keeper.AuthorizeUploader(suite.ctx, uploader2)
	suite.Require().NoError(err)

	// Pause a contract
	contract := "aura14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr"
	err = suite.keeper.PauseContract(suite.ctx, contract)
	suite.Require().NoError(err)

	// Set some stats
	stats := types.SecurityStats{
		TotalContractsUploaded:    10,
		TotalContractsInstantiated: 8,
		TotalExecutions:           100,
		TotalPausedContracts:      1,
		ReentrancyAttemptsBlocked: 2,
	}
	suite.keeper.SetSecurityStats(suite.ctx, stats)

	// Export genesis
	exported := suite.keeper.ExportGenesis(suite.ctx)
	suite.Require().NotNil(exported)
	suite.Require().Equal(params.MaxContractSize, exported.Params.MaxContractSize)
	suite.Require().Len(exported.AuthorizedUploaders, 2)
	suite.Require().Len(exported.PausedContracts, 1)
	suite.Require().Equal(stats.TotalContractsUploaded, exported.SecurityStats.TotalContractsUploaded)

	// Create new context for import
	newStoreKey := storetypes.NewKVStoreKey("test-new")
	newTestCtx := testutil.DefaultContextWithDB(suite.T(), newStoreKey, storetypes.NewTransientStoreKey("transient_import"))
	newCtx := newTestCtx.Ctx
	newKeeper := keeper.NewKeeper(suite.cdc, newStoreKey, nil, "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")

	// Import genesis
	err = newKeeper.InitGenesis(newCtx, *exported)
	suite.Require().NoError(err)

	// Verify imported state
	importedParams := newKeeper.GetParams(newCtx)
	suite.Require().Equal(params.MaxContractSize, importedParams.MaxContractSize)
	suite.Require().True(newKeeper.IsAuthorizedUploader(newCtx, uploader1))
	suite.Require().True(newKeeper.IsAuthorizedUploader(newCtx, uploader2))
	suite.Require().True(newKeeper.IsContractPaused(newCtx, contract))
	importedStats := newKeeper.GetSecurityStats(newCtx)
	suite.Require().Equal(stats.TotalContractsUploaded, importedStats.TotalContractsUploaded)
}

func TestParamsValidation(t *testing.T) {
	testCases := []struct {
		name      string
		params    types.Params
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid params",
			params:    types.DefaultParams(),
			expectErr: false,
		},
		{
			name: "zero max contract size",
			params: types.Params{
				MaxContractSize:         0,
				MaxInstantiateGas:       2_000_000,
				MaxExecuteGas:           1_000_000,
				MaxQueryGas:             100_000,
				RequireAuthorization:    true,
				EnableMigration:         false,
				MaxContractSizePerBlock: 5 * 1024 * 1024,
			},
			expectErr: true,
			errMsg:    "max contract size must be positive",
		},
		{
			name: "contract size too large",
			params: types.Params{
				MaxContractSize:         11 * 1024 * 1024,
				MaxInstantiateGas:       2_000_000,
				MaxExecuteGas:           1_000_000,
				MaxQueryGas:             100_000,
				RequireAuthorization:    true,
				EnableMigration:         false,
				MaxContractSizePerBlock: 5 * 1024 * 1024,
			},
			expectErr: true,
			errMsg:    "cannot exceed 10MB",
		},
		{
			name: "zero instantiate gas",
			params: types.Params{
				MaxContractSize:         600 * 1024,
				MaxInstantiateGas:       0,
				MaxExecuteGas:           1_000_000,
				MaxQueryGas:             100_000,
				RequireAuthorization:    true,
				EnableMigration:         false,
				MaxContractSizePerBlock: 5 * 1024 * 1024,
			},
			expectErr: true,
			errMsg:    "max instantiate gas must be positive",
		},
		{
			name: "zero execute gas",
			params: types.Params{
				MaxContractSize:         600 * 1024,
				MaxInstantiateGas:       2_000_000,
				MaxExecuteGas:           0,
				MaxQueryGas:             100_000,
				RequireAuthorization:    true,
				EnableMigration:         false,
				MaxContractSizePerBlock: 5 * 1024 * 1024,
			},
			expectErr: true,
			errMsg:    "max execute gas must be positive",
		},
		{
			name: "zero query gas",
			params: types.Params{
				MaxContractSize:         600 * 1024,
				MaxInstantiateGas:       2_000_000,
				MaxExecuteGas:           1_000_000,
				MaxQueryGas:             0,
				RequireAuthorization:    true,
				EnableMigration:         false,
				MaxContractSizePerBlock: 5 * 1024 * 1024,
			},
			expectErr: true,
			errMsg:    "max query gas must be positive",
		},
		{
			name: "zero max contract size per block",
			params: types.Params{
				MaxContractSize:         600 * 1024,
				MaxInstantiateGas:       2_000_000,
				MaxExecuteGas:           1_000_000,
				MaxQueryGas:             100_000,
				RequireAuthorization:    true,
				EnableMigration:         false,
				MaxContractSizePerBlock: 0,
			},
			expectErr: true,
			errMsg:    "max contract size per block must be positive",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
