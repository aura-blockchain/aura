// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/contractregistry/keeper"
	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type ContractLifecycleTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper *keeper.Keeper
	cdc    codec.BinaryCodec
}

func TestContractLifecycleTestSuite(t *testing.T) {
	suite.Run(t, new(ContractLifecycleTestSuite))
}

func (suite *ContractLifecycleTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx

	registry := codectypes.NewInterfaceRegistry()
	suite.cdc = codec.NewProtoCodec(registry)

	// NewKeeper signature: (storeKey, cdc, authority)
	suite.keeper = keeper.NewKeeper(
		key,
		suite.cdc,
		"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
	)

	// Initialize with default params
	params := types.DefaultParams()
	suite.NoError(suite.keeper.SetParams(suite.ctx, &params))
}

// ============================
// Full Lifecycle Tests
// ============================

func (suite *ContractLifecycleTestSuite) TestFullLifecycle_Register_Active_Deprecated() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	// Step 1: Register contract
	info := &pb.ContractInfo{
		Address:  contractAddr,
		CodeId:   1,
		Creator:  admin,
		Admin:    admin,
		Label:    "lifecycle-test",
		Metadata: pb.ContractMetadata{
			Name:        "Lifecycle Test Contract",
			Description: "Testing full lifecycle",
			Version:     "1.0.0",
		},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause: true,
		},
		Compliance: pb.ComplianceRequirements{},
		Status:     pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}

	err := suite.keeper.RegisterContract(suite.ctx, info)
	suite.NoError(err)

	// Verify active status
	stored, found := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.True(found)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_ACTIVE, stored.Status)
	suite.True(suite.keeper.IsContractApproved(suite.ctx, contractAddr))

	// Step 2: Deprecate contract
	err = suite.keeper.DeprecateContract(suite.ctx, contractAddr, admin, "end of life", "")
	suite.NoError(err)

	// Verify deprecated status
	stored, _ = suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_DEPRECATED, stored.Status)
	suite.True(suite.keeper.IsContractApproved(suite.ctx, contractAddr)) // Still approved
}

func (suite *ContractLifecycleTestSuite) TestFullLifecycle_Register_Pause_Unpause_Deprecate() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	// Step 1: Register
	info := &pb.ContractInfo{
		Address:  contractAddr,
		CodeId:   1,
		Creator:  admin,
		Admin:    admin,
		Label:    "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause: true,
		},
		Compliance: pb.ComplianceRequirements{},
		Status:     pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Step 2: Pause
	err := suite.keeper.PauseContract(suite.ctx, contractAddr, admin, "maintenance")
	suite.NoError(err)

	stored, _ := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_PAUSED, stored.Status)
	suite.False(suite.keeper.IsContractApproved(suite.ctx, contractAddr))

	// Step 3: Unpause
	err = suite.keeper.UnpauseContract(suite.ctx, contractAddr, admin)
	suite.NoError(err)

	stored, _ = suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_ACTIVE, stored.Status)
	suite.True(suite.keeper.IsContractApproved(suite.ctx, contractAddr))

	// Step 4: Deprecate
	err = suite.keeper.DeprecateContract(suite.ctx, contractAddr, admin, "new version available", "cosmos1newcontract")
	suite.NoError(err)

	stored, _ = suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_DEPRECATED, stored.Status)
}

func (suite *ContractLifecycleTestSuite) TestLifecycle_MultiplePauseCycles() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	// Register
	info := &pb.ContractInfo{
		Address:  contractAddr,
		CodeId:   1,
		Creator:  admin,
		Admin:    admin,
		Label:    "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause: true,
		},
		Compliance: pb.ComplianceRequirements{},
		Status:     pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Multiple pause/unpause cycles
	for i := 0; i < 3; i++ {
		// Pause
		err := suite.keeper.PauseContract(suite.ctx, contractAddr, admin, "cycle")
		suite.NoError(err)

		stored, _ := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
		suite.Equal(pb.ContractStatus_CONTRACT_STATUS_PAUSED, stored.Status)

		// Unpause
		err = suite.keeper.UnpauseContract(suite.ctx, contractAddr, admin)
		suite.NoError(err)

		stored, _ = suite.keeper.GetContractInfo(suite.ctx, contractAddr)
		suite.Equal(pb.ContractStatus_CONTRACT_STATUS_ACTIVE, stored.Status)
	}
}

// ============================
// Status Transition Tests
// ============================

func (suite *ContractLifecycleTestSuite) TestStatusTransition_ActiveToPaused() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        admin,
		Admin:          admin,
		Label:          "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{AllowPause: true},
		Compliance: pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Transition to paused
	err := suite.keeper.PauseContract(suite.ctx, contractAddr, admin, "test")
	suite.NoError(err)

	stored, _ := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_PAUSED, stored.Status)
}

func (suite *ContractLifecycleTestSuite) TestStatusTransition_PausedToActive() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        admin,
		Admin:          admin,
		Label:          "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{AllowPause: true},
		Compliance: pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))
	suite.NoError(suite.keeper.PauseContract(suite.ctx, contractAddr, admin, "test"))

	// Transition back to active
	err := suite.keeper.UnpauseContract(suite.ctx, contractAddr, admin)
	suite.NoError(err)

	stored, _ := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_ACTIVE, stored.Status)
}

func (suite *ContractLifecycleTestSuite) TestStatusTransition_ActiveToDeprecated() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        admin,
		Admin:          admin,
		Label:          "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{},
		Compliance: pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Transition to deprecated
	err := suite.keeper.DeprecateContract(suite.ctx, contractAddr, admin, "eol", "")
	suite.NoError(err)

	stored, _ := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_DEPRECATED, stored.Status)
}

func (suite *ContractLifecycleTestSuite) TestStatusTransition_InvalidUnpauseFromActive() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        admin,
		Admin:          admin,
		Label:          "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{},
		Compliance: pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Cannot unpause an active contract
	err := suite.keeper.UnpauseContract(suite.ctx, contractAddr, admin)
	suite.Error(err)
}

// ============================
// Metadata Evolution Tests
// ============================

func (suite *ContractLifecycleTestSuite) TestMetadataUpdates_ThroughLifecycle() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	// Register with initial metadata
	info := &pb.ContractInfo{
		Address: contractAddr,
		CodeId:  1,
		Creator: admin,
		Admin:   admin,
		Label:   "test",
		Metadata: pb.ContractMetadata{
			Name:    "Version 1",
			Version: "1.0.0",
			Tags:    []string{"v1"},
		},
		SecurityPolicy: pb.SecurityPolicy{AllowPause: true},
		Compliance: pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Update metadata while active
	newMetadata := &pb.ContractMetadata{
		Name:    "Version 1.1",
		Version: "1.1.0",
		Tags:    []string{"v1", "updated"},
	}
	err := suite.keeper.UpdateContractMetadata(suite.ctx, contractAddr, admin, newMetadata)
	suite.NoError(err)

	// Pause contract
	suite.NoError(suite.keeper.PauseContract(suite.ctx, contractAddr, admin, "update"))

	// Update metadata while paused
	newMetadata2 := &pb.ContractMetadata{
		Name:    "Version 2",
		Version: "2.0.0",
		Tags:    []string{"v2"},
	}
	err = suite.keeper.UpdateContractMetadata(suite.ctx, contractAddr, admin, newMetadata2)
	suite.NoError(err)

	// Verify final metadata
	stored, _ := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal("Version 2", stored.Metadata.Name)
	suite.Equal("2.0.0", stored.Metadata.Version)
}

// ============================
// Security Policy Evolution Tests
// ============================

func (suite *ContractLifecycleTestSuite) TestSecurityPolicy_Updates() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	// Register with initial policy
	info := &pb.ContractInfo{
		Address: contractAddr,
		CodeId:  1,
		Creator: admin,
		Admin:   admin,
		Label:   "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause:       false,
			MaxGasPerTx:      1000000,
			RateLimitPerUser: 100,
		},
		Compliance: pb.ComplianceRequirements{},
		Status:     pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Update to enable pause
	newPolicy := &pb.SecurityPolicy{
		AllowPause:       true,
		MaxGasPerTx:      2000000,
		RateLimitPerUser: 50,
	}
	err := suite.keeper.UpdateSecurityPolicy(suite.ctx, contractAddr, admin, newPolicy)
	suite.NoError(err)

	// Now pause should work
	err = suite.keeper.PauseContract(suite.ctx, contractAddr, admin, "now enabled")
	suite.NoError(err)

	stored, _ := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_PAUSED, stored.Status)
}

// ============================
// Execution and Metrics Tests
// ============================

func (suite *ContractLifecycleTestSuite) TestMetrics_ThroughLifecycle() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	// Register contract
	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        admin,
		Admin:          admin,
		Label:          "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{AllowPause: true},
		Compliance: pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Execute while active
	suite.keeper.UpdateMetricsOnExecution(suite.ctx, contractAddr, 50000, true)
	suite.keeper.UpdateMetricsOnExecution(suite.ctx, contractAddr, 60000, true)

	metrics, _ := suite.keeper.GetContractMetrics(suite.ctx, contractAddr)
	suite.Equal(uint64(2), metrics.TotalExecutions)

	// Pause
	suite.NoError(suite.keeper.PauseContract(suite.ctx, contractAddr, admin, "test"))

	// Metrics should persist
	metrics, _ = suite.keeper.GetContractMetrics(suite.ctx, contractAddr)
	suite.Equal(uint64(2), metrics.TotalExecutions)

	// Unpause and continue
	suite.NoError(suite.keeper.UnpauseContract(suite.ctx, contractAddr, admin))
	suite.keeper.UpdateMetricsOnExecution(suite.ctx, contractAddr, 55000, true)

	metrics, _ = suite.keeper.GetContractMetrics(suite.ctx, contractAddr)
	suite.Equal(uint64(3), metrics.TotalExecutions)
}

// ============================
// Authorization Tests
// ============================

func (suite *ContractLifecycleTestSuite) TestAuthorization_GovernanceOverride() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"
	governance := suite.keeper.GetAuthority()

	// Register contract
	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        admin,
		Admin:          admin,
		Label:          "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{AllowPause: true},
		Compliance: pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Governance can pause even if not admin
	err := suite.keeper.PauseContract(suite.ctx, contractAddr, governance, "governance action")
	suite.NoError(err)

	stored, _ := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_PAUSED, stored.Status)

	// Governance can unpause
	err = suite.keeper.UnpauseContract(suite.ctx, contractAddr, governance)
	suite.NoError(err)

	stored, _ = suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_ACTIVE, stored.Status)
}

func (suite *ContractLifecycleTestSuite) TestAuthorization_OnlyAdminOrGovernance() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"
	attacker := "cosmos1attacker"

	// Register contract
	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        admin,
		Admin:          admin,
		Label:          "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{AllowPause: true},
		Compliance: pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Attacker cannot pause
	err := suite.keeper.PauseContract(suite.ctx, contractAddr, attacker, "attack")
	suite.Error(err)
	suite.ErrorIs(err, types.ErrUnauthorized)

	// Admin can pause
	err = suite.keeper.PauseContract(suite.ctx, contractAddr, admin, "legitimate")
	suite.NoError(err)

	// Attacker cannot unpause
	err = suite.keeper.UnpauseContract(suite.ctx, contractAddr, attacker)
	suite.Error(err)
	suite.ErrorIs(err, types.ErrUnauthorized)
}

// ============================
// Edge Cases
// ============================

func (suite *ContractLifecycleTestSuite) TestEdgeCase_DeprecateAlreadyDeprecated() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	// Register and deprecate
	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        admin,
		Admin:          admin,
		Label:          "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{},
		Compliance: pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))
	suite.NoError(suite.keeper.DeprecateContract(suite.ctx, contractAddr, admin, "first", ""))

	// Deprecate again (should succeed)
	err := suite.keeper.DeprecateContract(suite.ctx, contractAddr, admin, "second", "")
	suite.NoError(err)
}

func (suite *ContractLifecycleTestSuite) TestEdgeCase_PauseWithoutAllowPauseFlag() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	// Register with pause disabled
	info := &pb.ContractInfo{
		Address:  contractAddr,
		CodeId:   1,
		Creator:  admin,
		Admin:    admin,
		Label:    "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause: false,
		},
		Compliance: pb.ComplianceRequirements{},
		Status:     pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Attempt to pause should fail
	err := suite.keeper.PauseContract(suite.ctx, contractAddr, admin, "test")
	suite.Error(err)
}
