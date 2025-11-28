package keeper_test

import (
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/contractregistry/keeper"
	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper *keeper.Keeper
	cdc    codec.BinaryCodec
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx

	registry := codectypes.NewInterfaceRegistry()
	suite.cdc = codec.NewProtoCodec(registry)

	// NewKeeper signature: (storeKey, cdc, authority)
	suite.keeper = keeper.NewKeeper(
		key,
		suite.cdc,
		"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", // authority
	)

	// Initialize with default params
	params := types.DefaultParams()
	suite.NoError(suite.keeper.SetParams(suite.ctx, &params))
}

func (suite *KeeperTestSuite) TestRegisterContract() {
	contractAddr := "cosmos1contract"
	creator := "cosmos1creator"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address: contractAddr,
		CodeId:  1,
		Creator: creator,
		Admin:   admin,
		Label:   "test-contract",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Metadata: &pb.ContractMetadata{
			Name:        "Test Contract",
			Description: "A test contract",
			Version:     "1.0.0",
			Tags:        []string{"test", "demo"},
		},
		SecurityPolicy: &pb.SecurityPolicy{
			AllowPause:  true,
			MaxGasPerTx: 1000000,
			RateLimitPerUser: 100,
		},
		Compliance: &pb.ComplianceRequirements{
			EnforceKyc: false,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}

	// Register contract
	err := suite.keeper.RegisterContract(suite.ctx, info)
	suite.NoError(err)

	// Verify contract is registered
	suite.True(suite.keeper.IsContractRegistered(suite.ctx, contractAddr))

	// Verify contract info
	stored, found := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.True(found)
	suite.Equal(contractAddr, stored.Address)
	suite.Equal(creator, stored.Creator)
	suite.Equal(admin, stored.Admin)

	// Verify creator index
	creatorContracts := suite.keeper.GetCreatorContracts(suite.ctx, creator)
	suite.Len(creatorContracts, 1)
	suite.Equal(contractAddr, creatorContracts[0].Address)

	// Verify tag indexes
	tagContracts := suite.keeper.GetTagContracts(suite.ctx, "test")
	suite.Len(tagContracts, 1)
	suite.Equal(contractAddr, tagContracts[0].Address)

	// Verify metrics initialized
	metrics, found := suite.keeper.GetContractMetrics(suite.ctx, contractAddr)
	suite.True(found)
	suite.Equal(contractAddr, metrics.ContractAddress)
	suite.Equal(uint64(0), metrics.TotalExecutions)
}

func (suite *KeeperTestSuite) TestRegisterContractAlreadyExists() {
	contractAddr := "cosmos1contract"
	creator := "cosmos1creator"

	info := &pb.ContractInfo{
		Address: contractAddr,
		CodeId:  1,
		Creator: creator,
		Admin:   creator,
		Label:   "test-contract",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Metadata: &pb.ContractMetadata{
			Name:        "Test Contract",
			Description: "A test contract",
		},
		SecurityPolicy: &pb.SecurityPolicy{},
		Compliance: &pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}

	// Register contract
	err := suite.keeper.RegisterContract(suite.ctx, info)
	suite.NoError(err)

	// Try to register again
	err = suite.keeper.RegisterContract(suite.ctx, info)
	suite.ErrorIs(err, types.ErrContractAlreadyExists)
}

func (suite *KeeperTestSuite) TestUpdateContractMetadata() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	// Register contract
	info := &pb.ContractInfo{
		Address: contractAddr,
		CodeId:  1,
		Creator: "cosmos1creator",
		Admin:   admin,
		Label:   "test-contract",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Metadata: &pb.ContractMetadata{
			Name:        "Test Contract",
			Description: "Original description",
			Tags:        []string{"old"},
		},
		SecurityPolicy: &pb.SecurityPolicy{},
		Compliance: &pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Update metadata
	newMetadata := &pb.ContractMetadata{
		Name:        "Updated Contract",
		Description: "New description",
		Version:     "2.0.0",
		Tags:        []string{"new", "updated"},
	}
	err := suite.keeper.UpdateContractMetadata(suite.ctx, contractAddr, admin, newMetadata)
	suite.NoError(err)

	// Verify update
	stored, found := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.True(found)
	suite.Equal("Updated Contract", stored.Metadata.Name)
	suite.Equal("New description", stored.Metadata.Description)
	suite.Equal("2.0.0", stored.Metadata.Version)

	// Verify tag indexes updated
	oldTagContracts := suite.keeper.GetTagContracts(suite.ctx, "old")
	suite.Len(oldTagContracts, 0)

	newTagContracts := suite.keeper.GetTagContracts(suite.ctx, "new")
	suite.Len(newTagContracts, 1)
}

func (suite *KeeperTestSuite) TestUpdateContractMetadataUnauthorized() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"
	notAdmin := "cosmos1notadmin"

	// Register contract
	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        "cosmos1creator",
		Admin:          admin,
		Label:          "test-contract",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Metadata:       &pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: &pb.SecurityPolicy{},
		Compliance: &pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Try to update as non-admin
	newMetadata := &pb.ContractMetadata{Name: "Hacked"}
	err := suite.keeper.UpdateContractMetadata(suite.ctx, contractAddr, notAdmin, newMetadata)
	suite.ErrorIs(err, types.ErrNotContractAdmin)
}

func (suite *KeeperTestSuite) TestPauseUnpauseContract() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	// Register contract with pause enabled
	info := &pb.ContractInfo{
		Address: contractAddr,
		CodeId:  1,
		Creator: admin,
		Admin:   admin,
		Label:   "test-contract",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Metadata: &pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: &pb.SecurityPolicy{
			AllowPause: true,
		},
		Compliance: &pb.ComplianceRequirements{},
		Status:     pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Pause contract
	err := suite.keeper.PauseContract(suite.ctx, contractAddr, admin, "maintenance")
	suite.NoError(err)

	// Verify paused
	stored, _ := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_PAUSED, stored.Status)
	suite.False(suite.keeper.IsContractApproved(suite.ctx, contractAddr))

	// Unpause contract
	err = suite.keeper.UnpauseContract(suite.ctx, contractAddr, admin)
	suite.NoError(err)

	// Verify active
	stored, _ = suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_ACTIVE, stored.Status)
	suite.True(suite.keeper.IsContractApproved(suite.ctx, contractAddr))
}

func (suite *KeeperTestSuite) TestDeprecateContract() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"
	migrationTarget := "cosmos1newcontract"

	// Register contract
	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        admin,
		Admin:          admin,
		Label:          "test-contract",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Metadata:       &pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: &pb.SecurityPolicy{},
		Compliance: &pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Deprecate contract
	err := suite.keeper.DeprecateContract(suite.ctx, contractAddr, admin, "v2 available", migrationTarget)
	suite.NoError(err)

	// Verify deprecated
	stored, _ := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_DEPRECATED, stored.Status)
	suite.True(suite.keeper.IsContractApproved(suite.ctx, contractAddr)) // Deprecated still allowed
}

func (suite *KeeperTestSuite) TestRateLimiting() {
	contractAddr := "cosmos1contract"
	userAddr := "cosmos1user"

	// Check rate limit (should pass when count is zero)
	err := suite.keeper.CheckRateLimit(suite.ctx, contractAddr, userAddr, 10)
	suite.NoError(err)

	// Increment counter multiple times
	for i := 0; i < 10; i++ {
		suite.keeper.IncrementRateLimit(suite.ctx, contractAddr, userAddr)
	}

	// Should now exceed limit
	err = suite.keeper.CheckRateLimit(suite.ctx, contractAddr, userAddr, 10)
	suite.ErrorIs(err, types.ErrRateLimitExceeded)

	// Check status
	current, remaining, _ := suite.keeper.GetRateLimitStatus(suite.ctx, contractAddr, userAddr, 10)
	suite.Equal(uint64(10), current)
	suite.Equal(uint64(0), remaining)
}

func (suite *KeeperTestSuite) TestMetricsTracking() {
	contractAddr := "cosmos1contract"

	// Register contract
	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        "cosmos1creator",
		Admin:          "cosmos1admin",
		Label:          "test-contract",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Metadata:       &pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: &pb.SecurityPolicy{},
		Compliance: &pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Update metrics for successful execution
	suite.keeper.UpdateMetricsOnExecution(suite.ctx, contractAddr, 50000, true)

	metrics, found := suite.keeper.GetContractMetrics(suite.ctx, contractAddr)
	suite.True(found)
	suite.Equal(uint64(1), metrics.TotalExecutions)
	suite.Equal(uint64(1), metrics.SuccessfulExecutions)
	suite.Equal(uint64(0), metrics.FailedExecutions)
	suite.Equal(uint64(50000), metrics.TotalGasUsed)
	suite.Equal(uint64(50000), metrics.AvgGasPerExecution)

	// Update metrics for failed execution
	suite.keeper.UpdateMetricsOnExecution(suite.ctx, contractAddr, 30000, false)

	metrics, _ = suite.keeper.GetContractMetrics(suite.ctx, contractAddr)
	suite.Equal(uint64(2), metrics.TotalExecutions)
	suite.Equal(uint64(1), metrics.SuccessfulExecutions)
	suite.Equal(uint64(1), metrics.FailedExecutions)
	suite.Equal(uint64(80000), metrics.TotalGasUsed)
	suite.Equal(uint64(40000), metrics.AvgGasPerExecution)

	// Increment counter metrics
	suite.keeper.IncrementMetricsCounter(suite.ctx, contractAddr, "rate_limit_violation")
	suite.keeper.IncrementMetricsCounter(suite.ctx, contractAddr, "compliance_failure")

	metrics, _ = suite.keeper.GetContractMetrics(suite.ctx, contractAddr)
	suite.Equal(uint64(1), metrics.RateLimitViolations)
	suite.Equal(uint64(1), metrics.ComplianceFailures)
}

func (suite *KeeperTestSuite) TestGenesisImportExport() {
	contractAddr := "cosmos1contract"
	creator := "cosmos1creator"

	// Register contract
	info := &pb.ContractInfo{
		Address: contractAddr,
		CodeId:  1,
		Creator: creator,
		Admin:   creator,
		Label:   "test-contract",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Metadata: &pb.ContractMetadata{
			Name:        "Test Contract",
			Description: "Genesis test",
			Tags:        []string{"genesis"},
		},
		SecurityPolicy: &pb.SecurityPolicy{
			MaxGasPerTx: 1000000,
		},
		Compliance: &pb.ComplianceRequirements{},
		Status:     pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Update some metrics
	suite.keeper.UpdateMetricsOnExecution(suite.ctx, contractAddr, 50000, true)

	// Export genesis
	exported := suite.keeper.ExportGenesis(suite.ctx)
	suite.NotNil(exported)
	suite.Len(exported.Contracts, 1)
	suite.Len(exported.Metrics, 1)
	suite.Equal(contractAddr, exported.Contracts[0].Address)

	// Create new keeper and import
	key := storetypes.NewKVStoreKey("test_import")
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_import"))
	importCtx := testCtx.Ctx

	// NewKeeper signature: (storeKey, cdc, authority)
	importKeeper := keeper.NewKeeper(
		key,
		suite.cdc,
		"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
	)

	// Import genesis
	err := importKeeper.InitGenesis(importCtx, exported)
	suite.NoError(err)

	// Verify imported data
	imported, found := importKeeper.GetContractInfo(importCtx, contractAddr)
	suite.True(found)
	suite.Equal(info.Address, imported.Address)
	suite.Equal(info.Metadata.Name, imported.Metadata.Name)

	importedMetrics, found := importKeeper.GetContractMetrics(importCtx, contractAddr)
	suite.True(found)
	suite.Equal(uint64(1), importedMetrics.TotalExecutions)
}

func (suite *KeeperTestSuite) TestMaxContractsPerCreator() {
	creator := "cosmos1creator"

	// Set low limit
	params := suite.keeper.GetParams(suite.ctx)
	params.MaxContractsPerCreator = 2
	suite.NoError(suite.keeper.SetParams(suite.ctx, params))

	// Register first contract
	info1 := &pb.ContractInfo{
		Address:        "cosmos1contract1",
		CodeId:         1,
		Creator:        creator,
		Admin:          creator,
		Label:          "contract1",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Metadata:       &pb.ContractMetadata{Name: "Contract 1"},
		SecurityPolicy: &pb.SecurityPolicy{},
		Compliance: &pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info1))

	// Register second contract
	info2 := &pb.ContractInfo{
		Address:        "cosmos1contract2",
		CodeId:         2,
		Creator:        creator,
		Admin:          creator,
		Label:          "contract2",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Metadata:       &pb.ContractMetadata{Name: "Contract 2"},
		SecurityPolicy: &pb.SecurityPolicy{},
		Compliance: &pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info2))

	// Try to register third contract (should fail)
	info3 := &pb.ContractInfo{
		Address:        "cosmos1contract3",
		CodeId:         3,
		Creator:        creator,
		Admin:          creator,
		Label:          "contract3",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Metadata:       &pb.ContractMetadata{Name: "Contract 3"},
		SecurityPolicy: &pb.SecurityPolicy{},
		Compliance: &pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err := suite.keeper.RegisterContract(suite.ctx, info3)
	suite.ErrorIs(err, types.ErrTooManyContracts)
}

func (suite *KeeperTestSuite) TestCleanupOldRateLimits() {
	contractAddr := "cosmos1contract"
	userAddr := "cosmos1user"

	// Set old timestamp (25 hours ago)
	oldTime := suite.ctx.BlockTime().Add(-25 * time.Hour)
	oldWindow := oldTime.Truncate(time.Hour).Unix()

	// Set old rate limit
	suite.keeper.SetRateLimitCount(suite.ctx, contractAddr, userAddr, oldWindow, 5)

	// Set current rate limit
	currentWindow := suite.ctx.BlockTime().Truncate(time.Hour).Unix()
	suite.keeper.SetRateLimitCount(suite.ctx, contractAddr, userAddr, currentWindow, 3)

	// Cleanup
	suite.keeper.CleanupOldRateLimits(suite.ctx)

	// Old limit should be removed
	oldCount := suite.keeper.GetRateLimitCount(suite.ctx, contractAddr, userAddr, oldWindow)
	suite.Equal(uint64(0), oldCount)

	// Current limit should remain
	currentCount := suite.keeper.GetRateLimitCount(suite.ctx, contractAddr, userAddr, currentWindow)
	suite.Equal(uint64(3), currentCount)
}

// Unit tests
func TestRegisterContractValidation(t *testing.T) {
	// This would test validation logic in isolation
	// Add specific test cases as needed
}

func TestRateLimitCalculation(t *testing.T) {
	// This would test rate limit calculations
	// Add specific test cases as needed
}
