// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

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
	suite.Require().Equal(uint64(600*1024), params.GetMaxWasmCodeSize())
	suite.Require().Equal(uint64(10_000_000), params.GetMaxGasWasmExecution())
	suite.Require().True(params.GetSecurityAnalysisEnabled())
	suite.Require().True(params.GetRequireAdminForMigrate())

	// Test setting custom params
	newParams := types.Params{
		CodeUploadAccess: types.AccessConfig{
			Permission: types.AccessTypeEverybody,
		},
		InstantiateDefaultPermission: types.AccessTypeEverybody,
		MaxWasmCodeSize:              1024 * 1024, // 1MB
		MaxGasWasmExecution:          3_000_000,
		SecurityAnalysisEnabled:      false,
		RequireAdminForMigrate:       true,
	}

	err := suite.keeper.SetParams(suite.ctx, newParams)
	suite.Require().NoError(err)

	// Verify params were set
	params = suite.keeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams.MaxWasmCodeSize, params.GetMaxWasmCodeSize())
	suite.Require().Equal(newParams.MaxGasWasmExecution, params.GetMaxGasWasmExecution())
	suite.Require().Equal(newParams.SecurityAnalysisEnabled, params.GetSecurityAnalysisEnabled())
	suite.Require().Equal(newParams.RequireAdminForMigrate, params.GetRequireAdminForMigrate())
}

func (suite *KeeperTestSuite) TestParamsValidation() {
	// Test invalid params
	invalidParams := types.Params{
		CodeUploadAccess: types.AccessConfig{
			Permission: types.AccessTypeEverybody,
		},
		InstantiateDefaultPermission: types.AccessTypeEverybody,
		MaxWasmCodeSize:              0, // Invalid - must be positive
		MaxGasWasmExecution:          2_000_000,
		SecurityAnalysisEnabled:      true,
		RequireAdminForMigrate:       false,
	}

	err := suite.keeper.SetParams(suite.ctx, invalidParams)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "max wasm code size must be positive")

	// Test contract size too large (over reasonable limit)
	invalidParams.MaxWasmCodeSize = 11 * 1024 * 1024 // 11MB - may be considered too large
	err = suite.keeper.SetParams(suite.ctx, invalidParams)
	// This may or may not error depending on validation rules
	// For now, just test that the system handles it
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

	// Without explicit authorization, should require being in the authorized list
	suite.Require().False(suite.keeper.IsAuthorizedUploader(suite.ctx, address))

	// Authorize the uploader
	err := suite.keeper.AuthorizeUploader(suite.ctx, address)
	suite.Require().NoError(err)

	// Should now be authorized
	suite.Require().True(suite.keeper.IsAuthorizedUploader(suite.ctx, address))
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
	suite.Require().Equal(uint64(1), stats.GetContractsPaused())

	// Unpause
	err = suite.keeper.UnpauseContract(suite.ctx, contractAddr)
	suite.Require().NoError(err)

	// Should no longer be paused
	suite.Require().False(suite.keeper.IsContractPaused(suite.ctx, contractAddr))

	// Verify stats updated
	stats = suite.keeper.GetSecurityStats(suite.ctx)
	suite.Require().Equal(uint64(0), stats.GetContractsPaused())
}

func (suite *KeeperTestSuite) TestSecurityStats() {
	// Get initial stats
	stats := suite.keeper.GetSecurityStats(suite.ctx)
	suite.Require().Equal(uint64(0), stats.GetTotalCodesAnalyzed())
	suite.Require().Equal(uint64(0), stats.GetTotalExecutions())

	// Update stats by calling internal methods
	// (In real usage, these are updated by the keeper methods)
	newStats := types.SecurityStats{
		TotalCodesAnalyzed: 5,
		CodesRejected:      1,
		ContractsPaused:    2,
		TotalExecutions:    100,
		FailedExecutions:   3,
		GasConsumedTotal:   1000000,
		LastSecurityScan:   12345,
	}
	suite.keeper.SetSecurityStats(suite.ctx, newStats)

	// Verify stats were set
	stats = suite.keeper.GetSecurityStats(suite.ctx)
	suite.Require().Equal(uint64(5), stats.GetTotalCodesAnalyzed())
	suite.Require().Equal(uint64(1), stats.GetCodesRejected())
	suite.Require().Equal(uint64(2), stats.GetContractsPaused())
	suite.Require().Equal(uint64(100), stats.GetTotalExecutions())
	suite.Require().Equal(uint64(3), stats.GetFailedExecutions())
	suite.Require().Equal(uint64(1000000), stats.GetGasConsumedTotal())
	suite.Require().Equal(uint64(12345), stats.GetLastSecurityScan())
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
	largeCode := make([]byte, params.GetMaxWasmCodeSize()+1)
	err = suite.keeper.ValidateContractUpload(suite.ctx, address, largeCode)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "exceeds maximum")

	// Unauthorized uploader
	// First, set params to restrict uploads to only authorized uploaders
	params.CodeUploadAccess = types.AccessConfig{
		Permission: types.AccessTypeOnlyAddress,
		Address:    address, // Only the authorized address can upload
	}
	err = suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

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
		CodeUploadAccess: types.AccessConfig{
			Permission: types.AccessTypeEverybody,
		},
		InstantiateDefaultPermission: types.AccessTypeEverybody,
		MaxWasmCodeSize:              1024 * 1024,
		MaxGasWasmExecution:          3_000_000,
		SecurityAnalysisEnabled:      true,
		RequireAdminForMigrate:       true,
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
		TotalCodesAnalyzed: 10,
		CodesRejected:      2,
		TotalExecutions:    100,
		ContractsPaused:    1,
		FailedExecutions:   5,
	}
	suite.keeper.SetSecurityStats(suite.ctx, stats)

	// Export genesis
	exported := suite.keeper.ExportGenesis(suite.ctx)
	suite.Require().NotNil(exported)
	suite.Require().Equal(params.GetMaxWasmCodeSize(), exported.Params.MaxWasmCodeSize)
	suite.Require().Len(exported.GetAuthorizedUploaders(), 2)
	suite.Require().Len(exported.GetPausedContracts(), 1)
	suite.Require().Equal(stats.GetTotalCodesAnalyzed(), exported.SecurityStats.TotalCodesAnalyzed)

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
	suite.Require().Equal(params.GetMaxWasmCodeSize(), importedParams.GetMaxWasmCodeSize())
	suite.Require().True(newKeeper.IsAuthorizedUploader(newCtx, uploader1))
	suite.Require().True(newKeeper.IsAuthorizedUploader(newCtx, uploader2))
	suite.Require().True(newKeeper.IsContractPaused(newCtx, contract))
	importedStats := newKeeper.GetSecurityStats(newCtx)
	suite.Require().Equal(stats.GetTotalCodesAnalyzed(), importedStats.GetTotalCodesAnalyzed())
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
			params:    *types.DefaultParams(),
			expectErr: false,
		},
		{
			name: "zero max wasm code size",
			params: types.Params{
				CodeUploadAccess: types.AccessConfig{
					Permission: types.AccessTypeEverybody,
				},
				InstantiateDefaultPermission: types.AccessTypeEverybody,
				MaxWasmCodeSize:              0,
				MaxGasWasmExecution:          2_000_000,
				SecurityAnalysisEnabled:      true,
				RequireAdminForMigrate:       false,
			},
			expectErr: true,
			errMsg:    "max wasm code size must be positive",
		},
		{
			name: "contract size too large",
			params: types.Params{
				CodeUploadAccess: types.AccessConfig{
					Permission: types.AccessTypeEverybody,
				},
				InstantiateDefaultPermission: types.AccessTypeEverybody,
				MaxWasmCodeSize:              11 * 1024 * 1024,
				MaxGasWasmExecution:          2_000_000,
				SecurityAnalysisEnabled:      true,
				RequireAdminForMigrate:       false,
			},
			expectErr: true,
			errMsg:    "cannot exceed 10MB",
		},
		{
			name: "zero max gas",
			params: types.Params{
				CodeUploadAccess: types.AccessConfig{
					Permission: types.AccessTypeEverybody,
				},
				InstantiateDefaultPermission: types.AccessTypeEverybody,
				MaxWasmCodeSize:              600 * 1024,
				MaxGasWasmExecution:          0,
				SecurityAnalysisEnabled:      true,
				RequireAdminForMigrate:       false,
			},
			expectErr: true,
			errMsg:    "max gas wasm execution must be positive",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// params.Validate() is not available in the generated proto
			// Instead we use SetParams which validates
			storeKey := storetypes.NewKVStoreKey("test")
			testCtx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test"))
			ctx := testCtx.Ctx

			registry := codectypes.NewInterfaceRegistry()
			types.RegisterInterfaces(registry)
			cdc := codec.NewProtoCodec(registry)

			k := keeper.NewKeeper(cdc, storeKey, nil, "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")

			err := k.SetParams(ctx, tc.params)
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
