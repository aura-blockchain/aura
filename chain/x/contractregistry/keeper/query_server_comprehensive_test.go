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
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/suite"
)

type QueryServerComprehensiveTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	keeper      *keeper.Keeper
	queryServer pb.QueryServer
	cdc         codec.BinaryCodec
}

func TestQueryServerComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerComprehensiveTestSuite))
}

func (suite *QueryServerComprehensiveTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx

	registry := codectypes.NewInterfaceRegistry()
	suite.cdc = codec.NewProtoCodec(registry)

	suite.keeper = keeper.NewKeeper(
		key,
		suite.cdc,
		"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
	)

	// Initialize with default params
	params := types.DefaultParams()
	suite.NoError(suite.keeper.SetParams(suite.ctx, &params))

	suite.queryServer = keeper.NewQueryServerImpl(suite.keeper)
}

// ============================
// ContractInfo Query Tests
// ============================

func (suite *QueryServerComprehensiveTestSuite) TestQueryContractInfo_Success() {
	// Register a contract
	contractAddr := "cosmos1contract"
	info := &pb.ContractInfo{
		Address: contractAddr,
		CodeId:  1,
		Creator: "cosmos1creator",
		Admin:   "cosmos1admin",
		Label:   "test-contract",
		Metadata: pb.ContractMetadata{
			Name:        "Test Contract",
			Description: "A test contract",
			Version:     "1.0.0",
			Tags:        []string{"test"},
		},
		SecurityPolicy: pb.SecurityPolicy{
			MaxGasPerTx: 1000000,
		},
		Compliance: pb.ComplianceRequirements{},
		Status:     pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Query contract info
	req := &pb.QueryContractInfoRequest{
		ContractAddress: contractAddr,
	}

	resp, err := suite.queryServer.ContractInfo(sdk.WrapSDKContext(suite.ctx), req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.NotNil(resp.Contract)
	suite.Equal(contractAddr, resp.Contract.Address)
	suite.Equal("Test Contract", resp.Contract.Metadata.Name)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_ACTIVE, resp.Contract.Status)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryContractInfo_NotFound() {
	req := &pb.QueryContractInfoRequest{
		ContractAddress: "cosmos1nonexistent",
	}

	_, err := suite.queryServer.ContractInfo(sdk.WrapSDKContext(suite.ctx), req)
	suite.Error(err)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryContractInfo_NilRequest() {
	_, err := suite.queryServer.ContractInfo(sdk.WrapSDKContext(suite.ctx), nil)
	suite.Error(err)
}

// ============================
// ContractsByCreator Query Tests
// ============================

func (suite *QueryServerComprehensiveTestSuite) TestQueryContractsByCreator_Success() {
	creator := "cosmos1creator"

	// Register multiple contracts
	for i := 1; i <= 3; i++ {
		contractAddr := "cosmos1contract" + string(rune('0'+i))
		info := &pb.ContractInfo{
			Address:        contractAddr,
			CodeId:         uint64(i),
			Creator:        creator,
			Admin:          creator,
			Label:          "test",
			Metadata:       pb.ContractMetadata{Name: "Test " + string(rune('0'+i))},
			SecurityPolicy: pb.SecurityPolicy{},
			Compliance:     pb.ComplianceRequirements{},
			Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
		}
		suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))
	}

	// Query contracts by creator
	req := &pb.QueryContractsByCreatorRequest{
		CreatorAddress: creator,
	}

	resp, err := suite.queryServer.ContractsByCreator(sdk.WrapSDKContext(suite.ctx), req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Len(resp.Contracts, 3)
	suite.Equal(uint64(3), resp.Pagination.Total)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryContractsByCreator_WithPagination() {
	creator := "cosmos1creator"

	// Register 5 contracts
	for i := 1; i <= 5; i++ {
		contractAddr := "cosmos1contract" + string(rune('0'+i))
		info := &pb.ContractInfo{
			Address:        contractAddr,
			CodeId:         uint64(i),
			Creator:        creator,
			Admin:          creator,
			Label:          "test",
			Metadata:       pb.ContractMetadata{Name: "Test"},
			SecurityPolicy: pb.SecurityPolicy{},
			Compliance:     pb.ComplianceRequirements{},
			Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
		}
		suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))
	}

	// Query with pagination (limit 2)
	req := &pb.QueryContractsByCreatorRequest{
		CreatorAddress: creator,
		Pagination: &query.PageRequest{
			Offset: 0,
			Limit:  2,
		},
	}

	resp, err := suite.queryServer.ContractsByCreator(sdk.WrapSDKContext(suite.ctx), req)
	suite.NoError(err)
	suite.Len(resp.Contracts, 2)
	suite.Equal(uint64(5), resp.Pagination.Total)

	// Query second page
	req.Pagination.Offset = 2
	resp, err = suite.queryServer.ContractsByCreator(sdk.WrapSDKContext(suite.ctx), req)
	suite.NoError(err)
	suite.Len(resp.Contracts, 2)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryContractsByCreator_NoContracts() {
	req := &pb.QueryContractsByCreatorRequest{
		CreatorAddress: "cosmos1nobody",
	}

	resp, err := suite.queryServer.ContractsByCreator(sdk.WrapSDKContext(suite.ctx), req)
	suite.NoError(err)
	suite.Len(resp.Contracts, 0)
}

// ============================
// ContractsByTag Query Tests
// ============================

func (suite *QueryServerComprehensiveTestSuite) TestQueryContractsByTag_Success() {
	tag := "defi"

	// Register contracts with tag
	for i := 1; i <= 3; i++ {
		contractAddr := "cosmos1contract" + string(rune('0'+i))
		info := &pb.ContractInfo{
			Address: contractAddr,
			CodeId:  uint64(i),
			Creator: "cosmos1creator",
			Admin:   "cosmos1admin",
			Label:   "test",
			Metadata: pb.ContractMetadata{
				Name: "Test",
				Tags: []string{tag, "other"},
			},
			SecurityPolicy: pb.SecurityPolicy{},
			Compliance:     pb.ComplianceRequirements{},
			Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
		}
		suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))
	}

	// Query contracts by tag
	req := &pb.QueryContractsByTagRequest{
		Tag: tag,
	}

	resp, err := suite.queryServer.ContractsByTag(sdk.WrapSDKContext(suite.ctx), req)
	suite.NoError(err)
	suite.Len(resp.Contracts, 3)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryContractsByTag_WithPagination() {
	tag := "nft"

	// Register 5 contracts with tag
	for i := 1; i <= 5; i++ {
		contractAddr := "cosmos1contract" + string(rune('0'+i))
		info := &pb.ContractInfo{
			Address:        contractAddr,
			CodeId:         uint64(i),
			Creator:        "cosmos1creator",
			Admin:          "cosmos1admin",
			Label:          "test",
			Metadata:       pb.ContractMetadata{Name: "Test", Tags: []string{tag}},
			SecurityPolicy: pb.SecurityPolicy{},
			Compliance:     pb.ComplianceRequirements{},
			Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
		}
		suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))
	}

	// Query with pagination
	req := &pb.QueryContractsByTagRequest{
		Tag: tag,
		Pagination: &query.PageRequest{
			Limit: 3,
		},
	}

	resp, err := suite.queryServer.ContractsByTag(sdk.WrapSDKContext(suite.ctx), req)
	suite.NoError(err)
	suite.Len(resp.Contracts, 3)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryContractsByTag_NoContracts() {
	req := &pb.QueryContractsByTagRequest{
		Tag: "nonexistent",
	}

	resp, err := suite.queryServer.ContractsByTag(sdk.WrapSDKContext(suite.ctx), req)
	suite.NoError(err)
	suite.Len(resp.Contracts, 0)
}

// ============================
// RegisteredContracts Query Tests
// ============================

func (suite *QueryServerComprehensiveTestSuite) TestQueryRegisteredContracts_Success() {
	// Register multiple contracts
	for i := 1; i <= 3; i++ {
		contractAddr := "cosmos1contract" + string(rune('0'+i))
		info := &pb.ContractInfo{
			Address:        contractAddr,
			CodeId:         uint64(i),
			Creator:        "cosmos1creator",
			Admin:          "cosmos1admin",
			Label:          "test",
			Metadata:       pb.ContractMetadata{Name: "Test"},
			SecurityPolicy: pb.SecurityPolicy{},
			Compliance:     pb.ComplianceRequirements{},
			Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
		}
		suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))
	}

	// Query all contracts
	req := &pb.QueryRegisteredContractsRequest{}

	resp, err := suite.queryServer.RegisteredContracts(sdk.WrapSDKContext(suite.ctx), req)
	suite.NoError(err)
	suite.Len(resp.Contracts, 3)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryRegisteredContracts_WithStatusFilter() {
	// Register contracts with different statuses
	for i := 1; i <= 3; i++ {
		contractAddr := "cosmos1contract" + string(rune('0'+i))
		status := pb.ContractStatus_CONTRACT_STATUS_ACTIVE
		if i == 2 {
			status = pb.ContractStatus_CONTRACT_STATUS_PAUSED
		}

		info := &pb.ContractInfo{
			Address:        contractAddr,
			CodeId:         uint64(i),
			Creator:        "cosmos1creator",
			Admin:          "cosmos1admin",
			Label:          "test",
			Metadata:       pb.ContractMetadata{Name: "Test"},
			SecurityPolicy: pb.SecurityPolicy{AllowPause: true},
			Compliance:     pb.ComplianceRequirements{},
			Status:         status,
		}
		suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

		if i == 2 {
			suite.NoError(suite.keeper.PauseContract(suite.ctx, contractAddr, "cosmos1admin", "test"))
		}
	}

	// Query only active contracts
	req := &pb.QueryRegisteredContractsRequest{
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}

	resp, err := suite.queryServer.RegisteredContracts(sdk.WrapSDKContext(suite.ctx), req)
	suite.NoError(err)
	suite.Len(resp.Contracts, 2)

	// Verify all are active
	for _, contract := range resp.Contracts {
		suite.Equal(pb.ContractStatus_CONTRACT_STATUS_ACTIVE, contract.Status)
	}
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryRegisteredContracts_Empty() {
	req := &pb.QueryRegisteredContractsRequest{}

	resp, err := suite.queryServer.RegisteredContracts(sdk.WrapSDKContext(suite.ctx), req)
	suite.NoError(err)
	suite.Len(resp.Contracts, 0)
}

// ============================
// ContractMetrics Query Tests
// ============================

func (suite *QueryServerComprehensiveTestSuite) TestQueryContractMetrics_Success() {
	// Register contract
	contractAddr := "cosmos1contract"
	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        "cosmos1creator",
		Admin:          "cosmos1admin",
		Label:          "test",
		Metadata:       pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{},
		Compliance:     pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Update metrics
	suite.keeper.UpdateMetricsOnExecution(suite.ctx, contractAddr, 50000, true)
	suite.keeper.UpdateMetricsOnExecution(suite.ctx, contractAddr, 60000, true)
	suite.keeper.UpdateMetricsOnExecution(suite.ctx, contractAddr, 40000, false)

	// Query metrics
	req := &pb.QueryContractMetricsRequest{
		ContractAddress: contractAddr,
	}

	resp, err := suite.queryServer.ContractMetrics(sdk.WrapSDKContext(suite.ctx), req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.NotNil(resp.Metrics)
	suite.Equal(uint64(3), resp.Metrics.TotalExecutions)
	suite.Equal(uint64(2), resp.Metrics.SuccessfulExecutions)
	suite.Equal(uint64(1), resp.Metrics.FailedExecutions)
	suite.Equal(uint64(150000), resp.Metrics.TotalGasUsed)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryContractMetrics_NoMetrics() {
	// Query metrics for non-existent contract
	req := &pb.QueryContractMetricsRequest{
		ContractAddress: "cosmos1nonexistent",
	}

	resp, err := suite.queryServer.ContractMetrics(sdk.WrapSDKContext(suite.ctx), req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.NotNil(resp.Metrics)
	// Should return zero metrics
	suite.Equal(uint64(0), resp.Metrics.TotalExecutions)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryContractMetrics_NilRequest() {
	_, err := suite.queryServer.ContractMetrics(sdk.WrapSDKContext(suite.ctx), nil)
	suite.Error(err)
}

// ============================
// Edge Cases
// ============================

func (suite *QueryServerComprehensiveTestSuite) TestQueryContracts_MultipleStatuses() {
	// Register contracts with various statuses
	statuses := []pb.ContractStatus{
		pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
		pb.ContractStatus_CONTRACT_STATUS_PAUSED,
		pb.ContractStatus_CONTRACT_STATUS_DEPRECATED,
	}

	for i, status := range statuses {
		contractAddr := "cosmos1contract" + string(rune('0'+i+1))
		info := &pb.ContractInfo{
			Address:        contractAddr,
			CodeId:         uint64(i + 1),
			Creator:        "cosmos1creator",
			Admin:          "cosmos1admin",
			Label:          "test",
			Metadata:       pb.ContractMetadata{Name: "Test"},
			SecurityPolicy: pb.SecurityPolicy{AllowPause: true},
			Compliance:     pb.ComplianceRequirements{},
			Status:         status,
		}
		suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))
	}

	// Query all contracts
	req := &pb.QueryRegisteredContractsRequest{}
	resp, err := suite.queryServer.RegisteredContracts(sdk.WrapSDKContext(suite.ctx), req)
	suite.NoError(err)
	suite.Len(resp.Contracts, 3)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryContractsByCreator_MultipleTags() {
	creator := "cosmos1creator"

	// Register contract with multiple tags
	info := &pb.ContractInfo{
		Address: "cosmos1contract",
		CodeId:  1,
		Creator: creator,
		Admin:   creator,
		Label:   "test",
		Metadata: pb.ContractMetadata{
			Name: "Multi-tag Contract",
			Tags: []string{"defi", "nft", "gaming"},
		},
		SecurityPolicy: pb.SecurityPolicy{},
		Compliance:     pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Query by creator should find it
	req := &pb.QueryContractsByCreatorRequest{
		CreatorAddress: creator,
	}
	resp, err := suite.queryServer.ContractsByCreator(sdk.WrapSDKContext(suite.ctx), req)
	suite.NoError(err)
	suite.Len(resp.Contracts, 1)

	// Each tag should also find it
	for _, tag := range []string{"defi", "nft", "gaming"} {
		tagReq := &pb.QueryContractsByTagRequest{Tag: tag}
		tagResp, err := suite.queryServer.ContractsByTag(sdk.WrapSDKContext(suite.ctx), tagReq)
		suite.NoError(err)
		suite.Len(tagResp.Contracts, 1)
	}
}
