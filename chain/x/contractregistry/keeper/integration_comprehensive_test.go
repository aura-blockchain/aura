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

type IntegrationTestSuite struct {
	suite.Suite
	ctx       sdk.Context
	keeper    *keeper.Keeper
	msgServer pb.MsgServer
	cdc       codec.BinaryCodec
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx
	registry := codectypes.NewInterfaceRegistry()
	suite.cdc = codec.NewProtoCodec(registry)
	suite.keeper = keeper.NewKeeper(key, suite.cdc, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")
	params := types.DefaultParams()
	suite.NoError(suite.keeper.SetParams(suite.ctx, &params))
	suite.msgServer = keeper.NewMsgServerImpl(*suite.keeper)
}

func (suite *IntegrationTestSuite) TestFullWorkflow_RegisterExecuteDeprecate() {
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	// 1. Register contract
	registerMsg := &pb.MsgRegisterContract{
		Signer:          admin,
		ContractAddress: contractAddr,
		CodeId:          1,
		Creator:         admin,
		Admin:           admin,
		Label:           "integration-test",
		Metadata: pb.ContractMetadata{
			Name:        "Integration Test Contract",
			Description: "Full workflow test",
			Version:     "1.0.0",
			Tags:        []string{"test", "integration"},
		},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause:       true,
			MaxGasPerTx:      1000000,
			RateLimitPerUser: 100,
		},
		Compliance: pb.ComplianceRequirements{},
	}

	resp, err := suite.msgServer.RegisterContract(sdk.WrapSDKContext(suite.ctx), registerMsg)
	suite.NoError(err)
	suite.True(resp.Success)

	// 2. Simulate executions and track metrics
	for i := 0; i < 50; i++ {
		suite.keeper.UpdateMetricsOnExecution(suite.ctx, contractAddr, 50000, true)
	}

	// Verify metrics
	metrics, found := suite.keeper.GetContractMetrics(suite.ctx, contractAddr)
	suite.True(found)
	suite.Equal(uint64(50), metrics.TotalExecutions)

	// 3. Update metadata
	updateMsg := &pb.MsgUpdateContractMetadata{
		Signer:          admin,
		ContractAddress: contractAddr,
		Metadata: pb.ContractMetadata{
			Name:        "Updated Contract",
			Description: "Updated description",
			Version:     "1.1.0",
			Tags:        []string{"test", "updated"},
		},
	}

	_, err = suite.msgServer.UpdateContractMetadata(sdk.WrapSDKContext(suite.ctx), updateMsg)
	suite.NoError(err)

	// 4. Pause for maintenance
	pauseMsg := &pb.MsgPauseContract{
		Signer:          admin,
		ContractAddress: contractAddr,
		Reason:          "maintenance",
	}

	_, err = suite.msgServer.PauseContract(sdk.WrapSDKContext(suite.ctx), pauseMsg)
	suite.NoError(err)

	// 5. Unpause
	unpauseMsg := &pb.MsgUnpauseContract{
		Signer:          admin,
		ContractAddress: contractAddr,
	}

	_, err = suite.msgServer.UnpauseContract(sdk.WrapSDKContext(suite.ctx), unpauseMsg)
	suite.NoError(err)

	// 6. More executions
	for i := 0; i < 30; i++ {
		suite.keeper.UpdateMetricsOnExecution(suite.ctx, contractAddr, 55000, true)
	}

	// 7. Deprecate contract
	deprecateMsg := &pb.MsgDeprecateContract{
		Signer:          admin,
		ContractAddress: contractAddr,
		Reason:          "v2 available",
		MigrationTarget: "cosmos1newcontract",
	}

	_, err = suite.msgServer.DeprecateContract(sdk.WrapSDKContext(suite.ctx), deprecateMsg)
	suite.NoError(err)

	// Verify final state
	info, found := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.True(found)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_DEPRECATED, info.Status)
	suite.Equal("Updated Contract", info.Metadata.Name)

	finalMetrics, _ := suite.keeper.GetContractMetrics(suite.ctx, contractAddr)
	suite.Equal(uint64(80), finalMetrics.TotalExecutions)
}

func (suite *IntegrationTestSuite) TestMultiContractScenario() {
	admin := "cosmos1admin"

	// Register 3 contracts with different configurations
	contracts := []string{"cosmos1contract1", "cosmos1contract2", "cosmos1contract3"}

	for i, addr := range contracts {
		msg := &pb.MsgRegisterContract{
			Signer:          admin,
			ContractAddress: addr,
			CodeId:          uint64(i + 1),
			Creator:         admin,
			Admin:           admin,
			Label:           "test",
			Metadata: pb.ContractMetadata{
				Name: "Contract " + string(rune('1'+i)),
				Tags: []string{"multi-test"},
			},
			SecurityPolicy: pb.SecurityPolicy{},
			Compliance: pb.ComplianceRequirements{},
		}

		resp, err := suite.msgServer.RegisterContract(sdk.WrapSDKContext(suite.ctx), msg)
		suite.NoError(err)
		suite.True(resp.Success)
	}

	// Verify contracts are registered
	for _, addr := range contracts {
		info, found := suite.keeper.GetContractInfo(suite.ctx, addr)
		suite.True(found)
		suite.Equal(pb.ContractStatus_CONTRACT_STATUS_ACTIVE, info.Status)
	}
}
