package keeper

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/identitychange/params"
	"github.com/aequitas/aura/chain/x/identitychange/types"
	pb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestInitGenesis(t *testing.T) {
	tests := []struct {
		name    string
		genesis types.GenesisState
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid genesis with all data",
			genesis: types.GenesisState{
				Params: &pb.Params{
					IdentityChangeFee:         "1000",
					MaxChangesPerUser:         10,
					ChangeWaitingPeriodBlocks: 1000,
					EnableTimelock:            true,
					TimelockDuration:          86400,
					RequireApproval:           false,
				},
				ChangeRequests: []*pb.IdentityChangeRequest{
					{
						RequestId:     "req-1",
						TargetDid:     "did:aura:123",
						OldAddress:    "aura1old1",
						NewAddress:    "aura1new1",
						RequestHeight: 1000,
						RequestTime:   timestamppb.Now(),
						Status:        pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING,
						Reason:        "Key rotation",
					},
					{
						RequestId:     "req-2",
						TargetDid:     "did:aura:456",
						OldAddress:    "aura1old2",
						NewAddress:    "aura1new2",
						RequestHeight: 2000,
						RequestTime:   timestamppb.Now(),
						Status:        pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPROVED,
						Reason:        "Lost key",
						ApprovalTxId:  "tx-approval-1",
					},
				},
				OwnershipRecords: []*pb.OwnershipRecord{
					{
						AssetId:      "asset-1",
						AssetType:    "NFT",
						OldOwner:     "aura1old1",
						NewOwner:     "aura1new1",
						TransferTime: timestamppb.Now(),
						TransferTxId: "tx-transfer-1",
					},
				},
				CurrentOwners: map[string]string{
					"asset-1": "aura1new1",
				},
			},
			wantErr: false,
		},
		{
			name: "default genesis",
			genesis: types.GenesisState{
				Params:           nil,
				ChangeRequests:   []*pb.IdentityChangeRequest{},
				OwnershipRecords: []*pb.OwnershipRecord{},
				CurrentOwners:    make(map[string]string),
			},
			wantErr: false,
		},
		{
			name: "invalid genesis - duplicate request IDs",
			genesis: types.GenesisState{
				Params: nil,
				ChangeRequests: []*pb.IdentityChangeRequest{
					{
						RequestId:     "req-1",
						TargetDid:     "did:aura:123",
						OldAddress:    "aura1old1",
						NewAddress:    "aura1new1",
						RequestHeight: 1000,
						RequestTime:   timestamppb.Now(),
						Status:        pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING,
					},
					{
						RequestId:     "req-1", // Duplicate
						TargetDid:     "did:aura:456",
						OldAddress:    "aura1old2",
						NewAddress:    "aura1new2",
						RequestHeight: 2000,
						RequestTime:   timestamppb.Now(),
						Status:        pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING,
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate",
		},
		{
			name: "invalid genesis - empty request ID",
			genesis: types.GenesisState{
				Params: nil,
				ChangeRequests: []*pb.IdentityChangeRequest{
					{
						RequestId:     "",
						TargetDid:     "did:aura:123",
						OldAddress:    "aura1old",
						NewAddress:    "aura1new",
						RequestHeight: 1000,
						RequestTime:   timestamppb.Now(),
						Status:        pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING,
					},
				},
			},
			wantErr: true,
			errMsg:  "missing id",
		},
		{
			name: "invalid genesis - empty target DID",
			genesis: types.GenesisState{
				Params: nil,
				ChangeRequests: []*pb.IdentityChangeRequest{
					{
						RequestId:     "req-1",
						TargetDid:     "",
						OldAddress:    "aura1old",
						NewAddress:    "aura1new",
						RequestHeight: 1000,
						RequestTime:   timestamppb.Now(),
						Status:        pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING,
					},
				},
			},
			wantErr: true,
			errMsg:  "target did required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate genesis state first
			if tt.wantErr {
				err := types.ValidateGenesisState(&tt.genesis)
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			paramsStore := params.NewStore(types.DefaultParams())
			keeper := NewKeeper(paramsStore)

			err := keeper.InitGenesis(tt.genesis)
			require.NoError(t, err)

			// Verify params were set
			p := keeper.GetParams()
			require.NotNil(t, p)

			// Verify change requests were loaded
			if len(tt.genesis.ChangeRequests) > 0 {
				for _, request := range tt.genesis.ChangeRequests {
					retrieved := keeper.GetChangeRequest(request.RequestId)
					require.NotNil(t, retrieved)
					require.Equal(t, request.TargetDid, retrieved.TargetDid)
					require.Equal(t, request.OldAddress, retrieved.OldAddress)
					require.Equal(t, request.NewAddress, retrieved.NewAddress)
				}
			}

			// Verify ownership records were loaded
			if len(tt.genesis.OwnershipRecords) > 0 {
				for _, record := range tt.genesis.OwnershipRecords {
					retrieved := keeper.GetOwnershipHistory(record.AssetId)
					require.NotNil(t, retrieved)
					require.NotEmpty(t, retrieved)
				}
			}

			// Verify current owners were loaded
			if len(tt.genesis.CurrentOwners) > 0 {
				for assetId, expectedOwner := range tt.genesis.CurrentOwners {
					actualOwner := keeper.GetCurrentOwner(assetId)
					require.Equal(t, expectedOwner, actualOwner)
				}
			}
		})
	}
}

func TestExportGenesis(t *testing.T) {
	paramsStore := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(paramsStore)

	// Create test data
	now := time.Now()

	// Submit identity change requests
	reqID1 := keeper.SubmitIdentityChangeRequest(
		"did:aura:test1",
		"aura1old1",
		"aura1new1",
		1000,
		now,
		"Key rotation test",
	)
	require.NotEmpty(t, reqID1)

	reqID2 := keeper.SubmitIdentityChangeRequest(
		"did:aura:test2",
		"aura1old2",
		"aura1new2",
		2000,
		now,
		"Lost key test",
	)
	require.NotEmpty(t, reqID2)

	// Record ownership transfer
	keeper.RecordOwnershipTransfer(
		"asset-test-1",
		"NFT",
		"aura1old1",
		"aura1new1",
		now,
		"tx-test-1",
	)

	// Export genesis
	exported := keeper.ExportGenesis()

	// Verify exported data
	require.NotNil(t, exported.Params)
	require.Greater(t, exported.Params.MaxChangesPerUser, uint64(0))

	require.Len(t, exported.ChangeRequests, 2)

	// Find the exported requests
	var foundReq1, foundReq2 bool
	for _, req := range exported.ChangeRequests {
		if req.TargetDid == "did:aura:test1" {
			foundReq1 = true
			require.Equal(t, "aura1old1", req.OldAddress)
			require.Equal(t, "aura1new1", req.NewAddress)
		}
		if req.TargetDid == "did:aura:test2" {
			foundReq2 = true
			require.Equal(t, "aura1old2", req.OldAddress)
			require.Equal(t, "aura1new2", req.NewAddress)
		}
	}
	require.True(t, foundReq1, "First request should be exported")
	require.True(t, foundReq2, "Second request should be exported")

	require.Len(t, exported.OwnershipRecords, 1)
	require.Equal(t, "asset-test-1", exported.OwnershipRecords[0].AssetId)
	require.Equal(t, "NFT", exported.OwnershipRecords[0].AssetType)

	require.Contains(t, exported.CurrentOwners, "asset-test-1")
	require.Equal(t, "aura1new1", exported.CurrentOwners["asset-test-1"])
}

func TestGenesisRoundTrip(t *testing.T) {
	// Create first keeper with initial state
	paramsStore1 := params.NewStore(types.DefaultParams())
	keeper1 := NewKeeper(paramsStore1)

	now := time.Now()

	// Create comprehensive test data
	reqID1 := keeper1.SubmitIdentityChangeRequest(
		"did:aura:roundtrip1",
		"aura1oldrt1",
		"aura1newrt1",
		5000,
		now,
		"Round trip test 1",
	)
	require.NotEmpty(t, reqID1)

	reqID2 := keeper1.SubmitIdentityChangeRequest(
		"did:aura:roundtrip2",
		"aura1oldrt2",
		"aura1newrt2",
		6000,
		now,
		"Round trip test 2",
	)
	require.NotEmpty(t, reqID2)

	// Approve one request
	keeper1.ApproveIdentityChange(reqID2, "tx-approval-rt")

	// Record ownership transfers
	keeper1.RecordOwnershipTransfer(
		"asset-rt-1",
		"Token",
		"aura1oldrt1",
		"aura1newrt1",
		now,
		"tx-rt-1",
	)
	keeper1.RecordOwnershipTransfer(
		"asset-rt-2",
		"NFT",
		"aura1oldrt2",
		"aura1newrt2",
		now,
		"tx-rt-2",
	)

	// Export genesis from keeper1
	exported := keeper1.ExportGenesis()

	// Create a new keeper and import the exported genesis
	paramsStore2 := params.NewStore(types.DefaultParams())
	keeper2 := NewKeeper(paramsStore2)
	err := keeper2.InitGenesis(exported)
	require.NoError(t, err)

	// Verify all data was preserved
	params1 := keeper1.GetParams()
	params2 := keeper2.GetParams()
	require.Equal(t, params1.MaxChangesPerUser, params2.MaxChangesPerUser)
	require.Equal(t, params1.EnableTimelock, params2.EnableTimelock)

	// Verify first request
	req1 := keeper2.GetChangeRequest(reqID1)
	require.NotNil(t, req1)
	require.Equal(t, "did:aura:roundtrip1", req1.TargetDid)
	require.Equal(t, "aura1oldrt1", req1.OldAddress)
	require.Equal(t, pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING, req1.Status)

	// Verify second request (approved)
	req2 := keeper2.GetChangeRequest(reqID2)
	require.NotNil(t, req2)
	require.Equal(t, "did:aura:roundtrip2", req2.TargetDid)
	require.Equal(t, pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPROVED, req2.Status)
	require.Equal(t, "tx-approval-rt", req2.ApprovalTxId)

	// Verify ownership records
	history1 := keeper2.GetOwnershipHistory("asset-rt-1")
	require.Len(t, history1, 1)
	require.Equal(t, "Token", history1[0].AssetType)

	history2 := keeper2.GetOwnershipHistory("asset-rt-2")
	require.Len(t, history2, 1)
	require.Equal(t, "NFT", history2[0].AssetType)

	// Verify current owners
	owner1 := keeper2.GetCurrentOwner("asset-rt-1")
	require.Equal(t, "aura1newrt1", owner1)

	owner2 := keeper2.GetCurrentOwner("asset-rt-2")
	require.Equal(t, "aura1newrt2", owner2)

	// Export again and verify consistency
	exported2 := keeper2.ExportGenesis()
	require.Equal(t, len(exported.ChangeRequests), len(exported2.ChangeRequests))
	require.Equal(t, len(exported.OwnershipRecords), len(exported2.OwnershipRecords))
	require.Equal(t, len(exported.CurrentOwners), len(exported2.CurrentOwners))
}

func TestDefaultGenesis(t *testing.T) {
	// Test that default genesis is valid
	defaultGen := types.DefaultGenesisState()
	require.NotNil(t, defaultGen)

	// Validate default genesis
	err := types.ValidateGenesisState(defaultGen)
	require.NoError(t, err)

	// Verify default params
	require.NotNil(t, defaultGen.Params)
	require.NotEmpty(t, defaultGen.Params.IdentityChangeFee)
	require.Greater(t, defaultGen.Params.MaxChangesPerUser, uint64(0))

	// Verify default collections are empty
	require.Empty(t, defaultGen.ChangeRequests)
	require.Empty(t, defaultGen.OwnershipRecords)
	require.NotNil(t, defaultGen.CurrentOwners)

	// Test importing default genesis
	paramsStore := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(paramsStore)
	err = keeper.InitGenesis(*defaultGen)
	require.NoError(t, err)

	// Verify keeper state after importing default genesis
	p := keeper.GetParams()
	require.NotNil(t, p)
	require.Greater(t, p.MaxChangesPerUser, uint64(0))
}

func TestInitGenesis_WithMultipleRequests(t *testing.T) {
	now := time.Now()
	genesis := types.GenesisState{
		Params: nil,
		ChangeRequests: []*pb.IdentityChangeRequest{
			{
				RequestId:     "multi-req-1",
				TargetDid:     "did:aura:multi1",
				OldAddress:    "aura1oldm1",
				NewAddress:    "aura1newm1",
				RequestHeight: 1000,
				RequestTime:   timestamppb.New(now),
				Status:        pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING,
				Reason:        "Test 1",
			},
			{
				RequestId:     "multi-req-2",
				TargetDid:     "did:aura:multi2",
				OldAddress:    "aura1oldm2",
				NewAddress:    "aura1newm2",
				RequestHeight: 2000,
				RequestTime:   timestamppb.New(now),
				Status:        pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPROVED,
				Reason:        "Test 2",
				ApprovalTxId:  "tx-approve-multi",
			},
			{
				RequestId:     "multi-req-3",
				TargetDid:     "did:aura:multi3",
				OldAddress:    "aura1oldm3",
				NewAddress:    "aura1newm3",
				RequestHeight: 3000,
				RequestTime:   timestamppb.New(now),
				Status:        pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_EXECUTED,
				Reason:        "Test 3",
				ExecutedAt:    timestamppb.New(now.Add(24 * time.Hour)),
			},
		},
		OwnershipRecords: []*pb.OwnershipRecord{
			{
				AssetId:      "multi-asset-1",
				AssetType:    "Token",
				OldOwner:     "aura1oldm1",
				NewOwner:     "aura1newm1",
				TransferTime: timestamppb.New(now),
				TransferTxId: "tx-multi-1",
			},
			{
				AssetId:      "multi-asset-2",
				AssetType:    "NFT",
				OldOwner:     "aura1oldm2",
				NewOwner:     "aura1newm2",
				TransferTime: timestamppb.New(now),
				TransferTxId: "tx-multi-2",
			},
		},
		CurrentOwners: map[string]string{
			"multi-asset-1": "aura1newm1",
			"multi-asset-2": "aura1newm2",
		},
	}

	paramsStore := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(paramsStore)

	err := keeper.InitGenesis(genesis)
	require.NoError(t, err)

	// Verify all requests were loaded
	req1 := keeper.GetChangeRequest("multi-req-1")
	require.NotNil(t, req1)
	require.Equal(t, pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING, req1.Status)

	req2 := keeper.GetChangeRequest("multi-req-2")
	require.NotNil(t, req2)
	require.Equal(t, pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPROVED, req2.Status)
	require.Equal(t, "tx-approve-multi", req2.ApprovalTxId)

	req3 := keeper.GetChangeRequest("multi-req-3")
	require.NotNil(t, req3)
	require.Equal(t, pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_EXECUTED, req3.Status)

	// Verify ownership records
	history1 := keeper.GetOwnershipHistory("multi-asset-1")
	require.Len(t, history1, 1)
	require.Equal(t, "Token", history1[0].AssetType)

	history2 := keeper.GetOwnershipHistory("multi-asset-2")
	require.Len(t, history2, 1)
	require.Equal(t, "NFT", history2[0].AssetType)

	// Verify current owners
	require.Equal(t, "aura1newm1", keeper.GetCurrentOwner("multi-asset-1"))
	require.Equal(t, "aura1newm2", keeper.GetCurrentOwner("multi-asset-2"))
}
