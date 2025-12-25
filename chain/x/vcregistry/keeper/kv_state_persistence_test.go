// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
	"github.com/stretchr/testify/require"
)

func TestKVDIDDocumentUpdatesPersist(t *testing.T) {
	keeper, ctx := setupKeeperWithKVStore(t)
	now := time.Now().Unix()
	keeper.SetCurrentTime(now)

	did := "did:aura:kv"
	controller := "aura1kv"

	require.NoError(t, keeper.RegisterDID(ctx, did, controller, []*types.VerificationMethod{}, "initial"))

	vm := []*types.VerificationMethod{{
		Id:        "key1",
		Type:      "Ed25519VerificationKey2020",
		PublicKey: []byte("pk"),
	}}

	require.NoError(t, keeper.UpdateDIDDocument(ctx, did, vm, "updated"))
	require.NoError(t, keeper.AddCredentialToDID(ctx, did, "vc1"))
	require.NoError(t, keeper.AddCredentialToDID(ctx, did, "vc2"))
	require.NoError(t, keeper.RemoveCredentialFromDID(ctx, did, "vc1"))

	stored, ok := keeper.GetDIDDocument(ctx, did)
	require.True(t, ok)
	require.Equal(t, "updated", stored.MetadataUri)
	require.Len(t, stored.VerificationMethods, 1)
	require.Equal(t, []string{"vc2"}, stored.CredentialIds)
}

func TestGenesisMintCountsKVRoundTrip(t *testing.T) {
	keeper1, ctx1 := setupKeeperWithKVStore(t)
	now := time.Now().Unix()
	keeper1.SetCurrentTime(now)

	addr := "aura1mintkv"
	keeper1.IncrementMintCount(ctx1, addr)
	keeper1.IncrementMintCount(ctx1, addr)
	keeper1.IncrementMintCount(ctx1, addr)

	gs := keeper1.ExportGenesis(ctx1)

	keeper2, ctx2 := setupKeeperWithKVStore(t)
	keeper2.SetCurrentTime(now)
	require.NoError(t, keeper2.InitGenesis(ctx2, gs))

	dayTimestamp := now / 86400
	count, ok := keeper2.store.getMintCount(ctx2, addr, dayTimestamp)
	require.True(t, ok)
	require.Equal(t, uint64(3), count)
}
