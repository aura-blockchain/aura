// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
	"github.com/stretchr/testify/require"
	gogotypes "github.com/cosmos/gogoproto/types"
)

func TestCreateAttributeVC_ValidationAndSingleton(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	now := time.Now().Unix()
	keeper.SetCurrentTime(now)

	// Missing encrypted value/hash should fail
	avcMissing := types.AttributeVC{
		AttributeVcId:     keeper.GenerateAttributeVCID(ctx, "addr1", types.AttributeType_ATTRIBUTE_TYPE_AGE),
		HolderAddress:     "addr1",
		AttributeType:     types.AttributeType_ATTRIBUTE_TYPE_AGE,
		ExpiresAt:         &gogotypes.Timestamp{Seconds: now+3600, Nanos: 0},
		VerificationLevel: 10,
	}
	err := keeper.CreateAttributeVC(ctx, avcMissing)
	require.Error(t, err)

	// Valid attribute VC
	avc := types.AttributeVC{
		AttributeVcId:     keeper.GenerateAttributeVCID(ctx, "addr1", types.AttributeType_ATTRIBUTE_TYPE_AGE),
		HolderAddress:     "addr1",
		AttributeType:     types.AttributeType_ATTRIBUTE_TYPE_AGE,
		EncryptedValue:    []byte("ciphertext"),
		ExpiresAt:         &gogotypes.Timestamp{Seconds: now+3600, Nanos: 0},
		Issuer:            "issuer1",
		VerificationLevel: 50,
	}
	err = keeper.CreateAttributeVC(ctx, avc)
	require.NoError(t, err)

	// Duplicate type for same holder (active) should fail
	avc2 := avc
	avc2.AttributeVcId = keeper.GenerateAttributeVCID(ctx, "addr1", avc.AttributeType)
	err = keeper.CreateAttributeVC(ctx, avc2)
	require.Error(t, err)

	// Different type should succeed
	avc3 := avc
	avc3.AttributeType = types.AttributeType_ATTRIBUTE_TYPE_EMAIL
	avc3.AttributeVcId = keeper.GenerateAttributeVCID(ctx, "addr1", avc3.AttributeType)
	err = keeper.CreateAttributeVC(ctx, avc3)
	require.NoError(t, err)

	// Filter by type should return only matching VC
	filtered := keeper.ListAttributeVCs(ctx, "addr1", []types.AttributeType{types.AttributeType_ATTRIBUTE_TYPE_EMAIL})
	require.Len(t, filtered, 1)
	require.Equal(t, types.AttributeType_ATTRIBUTE_TYPE_EMAIL, filtered[0].AttributeType)

	// Revoke and verify status
	err = keeper.RevokeAttributeVC(ctx, avc.AttributeVcId, "reason")
	require.NoError(t, err)
	stored, ok := keeper.GetAttributeVC(ctx, avc.AttributeVcId)
	require.True(t, ok)
	require.Equal(t, types.VCStatus_VC_STATUS_REVOKED, stored.Status)
}

func TestDisclosureRequestResponseFlow(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	now := time.Now().Unix()
	keeper.SetCurrentTime(now)

	holder := "addr1"
	req := types.DisclosureRequest{
		RequestId:           "req1",
		VerifierAddress:     "verifier1",
		RequestedAttributes: []types.AttributeType{types.AttributeType_ATTRIBUTE_TYPE_AGE},
		Purpose:             "age check",
		RequestedAt:         &gogotypes.Timestamp{Seconds: now, Nanos: 0},
		ExpiresInSeconds:    600,
	}

	err := keeper.CreateDisclosureRequest(ctx, holder, req)
	require.NoError(t, err)

	// Verify request was created via genesis export
	genesis := keeper.ExportGenesis(ctx)
	require.Contains(t, genesis.PendingDisclosureIndex[holder].Ids, "req1")

	// Disclosing unrequested attribute should fail
	badResp := types.DisclosureResponse{
		RequestId:     "req1",
		HolderAddress: holder,
		Approved:      true,
		DisclosedAttributes: []*types.AttributeDisclosure{
			{AttributeType: types.AttributeType_ATTRIBUTE_TYPE_EMAIL},
		},
	}
	err = keeper.RespondToDisclosureRequest(ctx, badResp)
	require.Error(t, err)

	// Valid response
	resp := types.DisclosureResponse{
		RequestId:     "req1",
		HolderAddress: holder,
		Approved:      true,
		DisclosedAttributes: []*types.AttributeDisclosure{
			{AttributeType: types.AttributeType_ATTRIBUTE_TYPE_AGE},
		},
	}
	err = keeper.RespondToDisclosureRequest(ctx, resp)
	require.NoError(t, err)

	// Verify pending entry removed and response stored via store methods
	genesis2 := keeper.ExportGenesis(ctx)
	if idx := genesis2.PendingDisclosureIndex[holder]; idx != nil {
		require.NotContains(t, idx.Ids, "req1")
	}
	stored, ok := keeper.GetDisclosureResponse(ctx, "req1")
	require.True(t, ok)
	require.True(t, stored.Approved)
}

func TestGenesisRoundTripSelectiveDisclosure(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	now := time.Now().Unix()
	keeper.SetCurrentTime(now)

	// Attribute VC
	avc := types.AttributeVC{
		AttributeVcId:  keeper.GenerateAttributeVCID(ctx, "addr1", types.AttributeType_ATTRIBUTE_TYPE_AGE),
		HolderAddress:  "addr1",
		AttributeType:  types.AttributeType_ATTRIBUTE_TYPE_AGE,
		EncryptedValue: []byte("cipher"),
		ExpiresAt:      &gogotypes.Timestamp{Seconds: now+3600, Nanos: 0},
	}
	require.NoError(t, keeper.CreateAttributeVC(ctx, avc))

	// Disclosure policy
	pol := types.DisclosurePolicy{
		HolderAddress: "addr1",
		DefaultMode:   types.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_DENY,
		Rules: []*types.AttributeDisclosureRule{
			{
				AttributeType: types.AttributeType_ATTRIBUTE_TYPE_AGE,
				Mode:          types.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ALLOW,
			},
		},
	}
	require.NoError(t, keeper.SetDisclosurePolicy(ctx, pol))

	// Disclosure request (pending)
	req := types.DisclosureRequest{
		RequestId:           "req-genesis",
		VerifierAddress:     "verifier1",
		RequestedAttributes: []types.AttributeType{types.AttributeType_ATTRIBUTE_TYPE_AGE},
		RequestedAt:         &gogotypes.Timestamp{Seconds: now, Nanos: 0},
		ExpiresInSeconds:    600,
	}
	require.NoError(t, keeper.CreateDisclosureRequest(ctx, "addr1", req))

	// Export and re-import
	gs := keeper.ExportGenesis(ctx)

	keeper2, ctx2 := setupKeeperForTest(t)
	require.NoError(t, keeper2.InitGenesis(ctx2, gs))

	// Attribute restored
	avcOut, ok := keeper2.GetAttributeVC(ctx2, avc.AttributeVcId)
	require.True(t, ok)
	require.Equal(t, avc.AttributeType, avcOut.AttributeType)

	// Disclosure policy restored
	polOut, ok := keeper2.GetDisclosurePolicy(ctx2, pol.HolderAddress)
	require.True(t, ok)
	require.Equal(t, pol.DefaultMode, polOut.DefaultMode)

	// Pending request restored
	reqOut, ok := keeper2.GetDisclosureRequest(ctx2, req.RequestId)
	require.True(t, ok)
	require.Equal(t, req.VerifierAddress, reqOut.VerifierAddress)

	// Verify pending index via genesis export
	genesis2 := keeper2.ExportGenesis(ctx2)
	require.Contains(t, genesis2.PendingDisclosureIndex["addr1"].Ids, req.RequestId)
}

func TestPresentationPersistenceAndGenesisMapFallback(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	now := time.Now().Unix()
	keeper.SetCurrentTime(now)

	holder := "holder1"
	vc := types.VCRecord{
		VcId:            "vc-map-1",
		HolderAddress:   holder,
		HolderDid:       "did:aura:holder1",
		IssuerAssistant: "issuer1",
		Status:          types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:        &gogotypes.Timestamp{Seconds: now, Nanos: 0},
		ExpiresAt:       &gogotypes.Timestamp{Seconds: now+600, Nanos: 0},
		VcType:          vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
	}
	require.NoError(t, keeper.SetVCRecord(ctx, vc))

	pres, qr, err := keeper.CreatePresentation(ctx, holder, []string{vc.VcId}, nil, 180)
	require.NoError(t, err)
	require.NotEmpty(t, qr)

	storedPres, ok := keeper.store.getPresentation(ctx, pres.PresentationId)
	require.True(t, ok)
	require.Equal(t, pres.PresentationId, storedPres.PresentationId)
	require.Contains(t, keeper.store.listUserPresentations(ctx, holder), pres.PresentationId)

	gs := keeper.ExportGenesis(ctx)
	restore, ctx2 := setupKeeperForTest(t)
	restore.SetCurrentTime(now)
	require.NoError(t, restore.InitGenesis(ctx2, gs))

	restoredPres, ok := restore.store.getPresentation(ctx2, pres.PresentationId)
	require.True(t, ok)
	require.Equal(t, pres.PresentationId, restoredPres.PresentationId)
	require.Contains(t, restore.store.listUserPresentations(ctx2, holder), pres.PresentationId)
}
