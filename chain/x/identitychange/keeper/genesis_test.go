// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/types"
	pb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
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
					MaxRequestsPerWalletPerMonth:   10,
					MinConfidenceAfterChange:       50,
					StalenessHeightThreshold:       1000,
					AssistantSlashOnFalsePositive:  true,
					StalenessInvestigatorChain:     "aura-testnet",
				},
				Records: []*pb.IdentityRecord{
					{
						Did:               "did:aura:123",
						Owner:             "aura1owner",
						ConfidenceScore:   75,
						MetadataHash:      "hash123",
						LatestIrVersion:   "v1",
						LastChangedHeight: 1000,
						Status:            pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_IDLE,
					},
				},
				Requests: []*pb.IdentityChangeRequest{
					{
						RequestId:       "req-1",
						TargetDid:       "did:aura:123",
						Requester:       "aura1requester",
						Assistant:       "",
						IrId:            "ir-1",
						ProofHash:       "proof-hash-1",
						RequestMetaHash: "meta-hash-1",
						Status:          pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
						Reason:          "Key rotation",
						CreatedHeight:   1000,
						VerdictHeight:   0,
					},
				},
				History: []*pb.IdentityChangeHistory{
					{
						RequestId:           "req-1",
						TargetDid:           "did:aura:123",
						PrevConfidenceScore: 70,
						NewConfidenceScore:  75,
						TransitionReason:    "applied",
						ChangedHeight:       1001,
					},
				},
				Suspended: false,
			},
			wantErr: false,
		},
		{
			name: "default genesis",
			genesis: types.GenesisState{
				Params:    nil,
				Records:   []*pb.IdentityRecord{},
				Requests:  []*pb.IdentityChangeRequest{},
				History:   []*pb.IdentityChangeHistory{},
				Suspended: false,
			},
			wantErr: false,
		},
		{
			name: "invalid genesis - duplicate request IDs",
			genesis: types.GenesisState{
				Params: nil,
				Requests: []*pb.IdentityChangeRequest{
					{
						RequestId:       "req-1",
						TargetDid:       "did:aura:123",
						Requester:       "aura1requester",
						IrId:            "ir-1",
						ProofHash:       "proof-1",
						RequestMetaHash: "meta-1",
						Status:          pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
						CreatedHeight:   1000,
					},
					{
						RequestId:       "req-1", // Duplicate
						TargetDid:       "did:aura:456",
						Requester:       "aura1requester2",
						IrId:            "ir-2",
						ProofHash:       "proof-2",
						RequestMetaHash: "meta-2",
						Status:          pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
						CreatedHeight:   2000,
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
				Requests: []*pb.IdentityChangeRequest{
					{
						RequestId:       "",
						TargetDid:       "did:aura:123",
						Requester:       "aura1requester",
						IrId:            "ir-1",
						ProofHash:       "proof-1",
						RequestMetaHash: "meta-1",
						Status:          pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
						CreatedHeight:   1000,
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
				Requests: []*pb.IdentityChangeRequest{
					{
						RequestId:       "req-1",
						TargetDid:       "",
						Requester:       "aura1requester",
						IrId:            "ir-1",
						ProofHash:       "proof-1",
						RequestMetaHash: "meta-1",
						Status:          pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
						CreatedHeight:   1000,
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

			input := keepertest.CreateTestInput(t)
			keeper := NewKeeper(
				keepertest.WrapStoreService(input.StoreKey),
				input.Cdc,
				nil,
				"authority",
				keepertest.Logger(),
			)

			err := keeper.InitGenesis(input.Ctx, tt.genesis)
			require.NoError(t, err)

			// Verify params were set
			p, _ := keeper.GetParams(input.Ctx)
			require.NotNil(t, p)

			// Verify requests were loaded
			if len(tt.genesis.Requests) > 0 {
				for _, request := range tt.genesis.Requests {
					retrieved, found := keeper.GetRequest(input.Ctx, request.RequestId)
					require.True(t, found)
					require.Equal(t, request.TargetDid, retrieved.TargetDid)
					require.Equal(t, request.Requester, retrieved.Requester)
				}
			}

			// Verify records were loaded
			if len(tt.genesis.Records) > 0 {
				for _, record := range tt.genesis.Records {
					retrieved, found := keeper.GetIdentityRecord(input.Ctx, record.Did)
					require.True(t, found)
					require.Equal(t, record.Owner, retrieved.Owner)
					require.Equal(t, record.ConfidenceScore, retrieved.ConfidenceScore)
				}
			}
		})
	}
}

func TestExportGenesis(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	keeper := NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		nil,
		"authority",
		keepertest.Logger(),
	)

	// Set default params
	params := types.DefaultParams()
	err := keeper.SetParams(params)
	require.NoError(t, err)

	// Create test data
	request1 := types.IdentityChangeRequest{
		RequestId:       "req-export-1",
		TargetDid:       "did:aura:export1",
		Requester:       "aura1requester1",
		IrId:            "ir-export-1",
		ProofHash:       "proof-hash-1",
		RequestMetaHash: "meta-hash-1",
		Status:          pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
		CreatedHeight:   1000,
	}
	err = keeper.SetRequest(input.Ctx, request1)
	require.NoError(t, err)

	record1 := types.IdentityRecord{
		Did:               "did:aura:export1",
		Owner:             "aura1owner1",
		ConfidenceScore:   80,
		MetadataHash:      "metadata-hash-1",
		LatestIrVersion:   "v1",
		LastChangedHeight: 1001,
		Status:            pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPLIED,
	}
	err = keeper.SetIdentityRecord(input.Ctx, record1)
	require.NoError(t, err)

	// Export genesis
	exported := keeper.ExportGenesis(input.Ctx)

	// Verify exported data
	require.NotNil(t, exported.Params)
	require.Len(t, exported.Requests, 1)
	require.Len(t, exported.Records, 1)

	// Verify request exported correctly
	require.Equal(t, "req-export-1", exported.Requests[0].RequestId)
	require.Equal(t, "did:aura:export1", exported.Requests[0].TargetDid)

	// Verify record exported correctly
	require.Equal(t, "did:aura:export1", exported.Records[0].Did)
	require.Equal(t, "aura1owner1", exported.Records[0].Owner)
}

func TestGenesisRoundTrip(t *testing.T) {
	// Create first keeper with initial state
	input1 := keepertest.CreateTestInput(t)
	keeper1 := NewKeeper(
		keepertest.WrapStoreService(input1.StoreKey),
		input1.Cdc,
		nil,
		"authority",
		keepertest.Logger(),
	)

	// Set params
	params := types.DefaultParams()
	err := keeper1.SetParams(params)
	require.NoError(t, err)

	// Create comprehensive test data
	request := types.IdentityChangeRequest{
		RequestId:       "req-roundtrip",
		TargetDid:       "did:aura:roundtrip",
		Requester:       "aura1requesterrt",
		IrId:            "ir-rt",
		ProofHash:       "proof-rt",
		RequestMetaHash: "meta-rt",
		Status:          pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
		CreatedHeight:   5000,
	}
	err = keeper1.SetRequest(input1.Ctx, request)
	require.NoError(t, err)

	record := types.IdentityRecord{
		Did:               "did:aura:roundtrip",
		Owner:             "aura1ownerrt",
		ConfidenceScore:   85,
		MetadataHash:      "metadata-rt",
		LatestIrVersion:   "v1",
		LastChangedHeight: 5001,
		Status:            pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPLIED,
	}
	err = keeper1.SetIdentityRecord(input1.Ctx, record)
	require.NoError(t, err)

	// Export genesis from keeper1
	exported := keeper1.ExportGenesis(input1.Ctx)

	// Create a new keeper and import the exported genesis
	input2 := keepertest.CreateTestInput(t)
	keeper2 := NewKeeper(
		keepertest.WrapStoreService(input2.StoreKey),
		input2.Cdc,
		nil,
		"authority",
		keepertest.Logger(),
	)
	err = keeper2.InitGenesis(input2.Ctx, exported)
	require.NoError(t, err)

	// Verify all data was preserved
	params1, _ := keeper1.GetParams(input1.Ctx)
	params2, _ := keeper2.GetParams(input2.Ctx)
	require.Equal(t, params1.MaxRequestsPerWalletPerMonth, params2.MaxRequestsPerWalletPerMonth)
	require.Equal(t, params1.MinConfidenceAfterChange, params2.MinConfidenceAfterChange)

	// Verify request
	req, found := keeper2.GetRequest(input2.Ctx, "req-roundtrip")
	require.True(t, found)
	require.Equal(t, "did:aura:roundtrip", req.TargetDid)
	require.Equal(t, pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION, req.Status)

	// Verify record
	rec, found := keeper2.GetIdentityRecord(input2.Ctx, "did:aura:roundtrip")
	require.True(t, found)
	require.Equal(t, "aura1ownerrt", rec.Owner)
	require.Equal(t, int64(85), rec.ConfidenceScore)

	// Export again and verify consistency
	exported2 := keeper2.ExportGenesis(input2.Ctx)
	require.Equal(t, len(exported.Requests), len(exported2.Requests))
	require.Equal(t, len(exported.Records), len(exported2.Records))
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
	require.Greater(t, defaultGen.Params.MaxRequestsPerWalletPerMonth, int32(0))

	// Verify default collections are empty
	require.Empty(t, defaultGen.Requests)
	require.Empty(t, defaultGen.Records)
	require.Empty(t, defaultGen.History)

	// Test importing default genesis
	input := keepertest.CreateTestInput(t)
	keeper := NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		nil,
		"authority",
		keepertest.Logger(),
	)
	err = keeper.InitGenesis(input.Ctx, *defaultGen)
	require.NoError(t, err)

	// Verify keeper state after importing default genesis
	p, _ := keeper.GetParams(input.Ctx)
	require.NotNil(t, p)
	require.Greater(t, p.MaxRequestsPerWalletPerMonth, int32(0))
}
