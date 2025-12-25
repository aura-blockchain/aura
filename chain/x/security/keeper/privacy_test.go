// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/aequitas/aura/chain/testutil"
	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

func TestMixingPoolCRUD(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	pool := &securitypb.MixingPool{
		PoolId:          "pool-1",
		MinParticipants: 2,
		MaxParticipants: 4,
		Denomination:    sdkmath.NewInt(1000000),
		MixingRounds:    3,
		Status:          "active",
	}

	k.SetMixingPool(ctx, pool)

	got, found := k.GetMixingPool(ctx, pool.PoolId)
	require.True(t, found)
	require.Equal(t, pool.PoolId, got.PoolId)
	require.Equal(t, pool.Status, got.Status)
	require.Equal(t, pool.MinParticipants, got.MinParticipants)
	require.Equal(t, pool.MaxParticipants, got.MaxParticipants)
	require.Equal(t, pool.Denomination, got.Denomination)

	all := k.GetAllMixingPools(ctx)
	require.Len(t, all, 1)
	require.Equal(t, pool.PoolId, all[0].PoolId)

	k.DeleteMixingPool(ctx, pool.PoolId)
	_, found = k.GetMixingPool(ctx, pool.PoolId)
	require.False(t, found)
}

func TestGetMixingPoolByDenomination(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	active := &securitypb.MixingPool{
		PoolId:          "active-pool",
		MinParticipants: 2,
		MaxParticipants: 4,
		Denomination:    sdkmath.NewInt(10),
		Status:          "active",
	}
	inactive := &securitypb.MixingPool{
		PoolId:          "inactive-pool",
		MinParticipants: 2,
		MaxParticipants: 4,
		Denomination:    sdkmath.NewInt(10),
		Status:          "paused",
	}

	k.SetMixingPool(ctx, inactive)
	k.SetMixingPool(ctx, active)

	foundPool, found := k.GetMixingPoolByDenomination(ctx, "uatest")
	require.True(t, found)
	require.Equal(t, active.PoolId, foundPool.PoolId)

	// When all pools are inactive, no match should be returned.
	k.DeleteMixingPool(ctx, active.PoolId)
	_, found = k.GetMixingPoolByDenomination(ctx, "uatest")
	require.False(t, found)
}

func TestValidateRingSize(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	require.NoError(t, k.ValidateRingSize(ctx, k.GetParams(ctx).Privacy.MinRingSize))
	require.NoError(t, k.ValidateRingSize(ctx, k.GetParams(ctx).Privacy.MaxRingSize))

	err := k.ValidateRingSize(ctx, k.GetParams(ctx).Privacy.MinRingSize-1)
	require.ErrorIs(t, err, types.ErrRingTooSmall)

	err = k.ValidateRingSize(ctx, k.GetParams(ctx).Privacy.MaxRingSize+1)
	require.ErrorIs(t, err, types.ErrRingTooLarge)
}

func TestValidateMixingParticipants(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)
	min := k.GetParams(ctx).Privacy.MinMixingParticipants

	require.NoError(t, k.ValidateMixingParticipants(ctx, min))
	err := k.ValidateMixingParticipants(ctx, min-1)
	require.ErrorIs(t, err, types.ErrInsufficientMixers)
}

func TestViewKeyLifecycleAndPermissions(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)
	now := time.Now().UTC()
	ctx = ctx.WithBlockTime(now)

	validUntil := now.Add(1 * time.Hour)
	viewKey := &types.ViewKey{
		Id:            "vk-1",
		WalletAddress: "aura1walletxyz",
		ViewKey:       "vk-data",
		RegisteredAt:  &now,
		ValidUntil:    &validUntil,
		Permissions:   []string{"balance", "history"},
	}

	k.SetRegisteredViewKey(ctx, viewKey)

	got, found := k.GetRegisteredViewKey(ctx, viewKey.Id)
	require.True(t, found)
	require.Equal(t, viewKey.Id, got.Id)
	require.Equal(t, viewKey.WalletAddress, got.WalletAddress)
	require.ElementsMatch(t, viewKey.Permissions, got.Permissions)
	require.NotNil(t, got.ValidUntil)
	require.True(t, got.ValidUntil.After(ctx.BlockTime()), "valid until must be in the future relative to block time")

	all := k.GetAllRegisteredViewKeys(ctx)
	require.Len(t, all, 1)

	walletKeys := k.GetViewKeysForWallet(ctx, viewKey.WalletAddress)
	require.Len(t, walletKeys, 1)
	require.Equal(t, viewKey.Id, walletKeys[0].Id)

	require.True(t, k.CheckViewKeyPermission(ctx, viewKey.Id, "balance"))
	require.False(t, k.CheckViewKeyPermission(ctx, viewKey.Id, "all"))
	require.False(t, k.CheckViewKeyPermission(ctx, viewKey.Id, "nonexistent"))

	// Update to wildcard permission and verify it grants all actions.
	viewKey.Permissions = []string{"all"}
	k.SetRegisteredViewKey(ctx, viewKey)
	require.True(t, k.CheckViewKeyPermission(ctx, viewKey.Id, "balance"))
	require.True(t, k.CheckViewKeyPermission(ctx, viewKey.Id, "arbitrary"))

	// Expired keys should fail permission checks.
	expired := now.Add(-1 * time.Hour)
	viewKey.ValidUntil = &expired
	k.SetRegisteredViewKey(ctx, viewKey)
	require.False(t, k.CheckViewKeyPermission(ctx, viewKey.Id, "balance"))
}

func TestJoinMixingPoolValidation(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Unknown pool should return not found.
	err := k.JoinMixingPool(ctx, "missing")
	require.ErrorIs(t, err, types.ErrMixingPoolNotFound)

	// Inactive pool should be rejected.
	inactive := &securitypb.MixingPool{
		PoolId:          "inactive",
		Status:          "paused",
		MinParticipants: 2,
		MaxParticipants: 4,
		Denomination:    sdkmath.NewInt(5),
	}
	k.SetMixingPool(ctx, inactive)
	err = k.JoinMixingPool(ctx, inactive.PoolId)
	require.ErrorIs(t, err, types.ErrInvalidState)

	// Active pool should not error (participants are tracked via state writes).
	active := &securitypb.MixingPool{
		PoolId:          "active",
		Status:          "active",
		MinParticipants: 1,
		MaxParticipants: 2,
		Denomination:    sdkmath.NewInt(5),
	}
	k.SetMixingPool(ctx, active)
	err = k.JoinMixingPool(ctx, active.PoolId)
	require.NoError(t, err)
}

func TestViewKeyDeletion(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)
	now := time.Now().UTC()
	vk := &types.ViewKey{
		Id:            "vk-delete",
		WalletAddress: "aura1walletxyz",
		ViewKey:       "vk",
		RegisteredAt:  &now,
		Permissions:   []string{"all"},
	}

	k.SetRegisteredViewKey(ctx, vk)
	_, found := k.GetRegisteredViewKey(ctx, vk.Id)
	require.True(t, found)

	k.DeleteRegisteredViewKey(ctx, vk.Id)
	_, found = k.GetRegisteredViewKey(ctx, vk.Id)
	require.False(t, found)

	// Wallet lookup should be empty after deletion.
	require.Empty(t, k.GetViewKeysForWallet(ctx, vk.WalletAddress))
}

func TestStealthAndRingStorage(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	stealth := &securitypb.StealthAddress{OneTimeAddress: []byte{0x01, 0x02, 0x03}}
	k.SetStealthAddress(ctx, stealth)
	allStealth := k.GetAllStealthAddresses(ctx)
	require.Len(t, allStealth, 1)
	require.Equal(t, stealth.OneTimeAddress, allStealth[0].OneTimeAddress)

	ring := &securitypb.RingSignature{
		KeyImage:      []byte{0xaa},
		RingMembers:   [][]byte{{0x01}, {0x02}},
		SignatureData: []byte("sig"),
		RingSize:      2,
		MixinLevel:    1,
	}
	k.SetRingSignature(ctx, ring)
	allRings := k.GetAllRingSignatures(ctx)
	require.Len(t, allRings, 1)
	require.Equal(t, ring.RingSize, allRings[0].RingSize)
}

func TestViewKeyPermissionUnknown(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	require.False(t, k.CheckViewKeyPermission(ctx, "unknown", "any"))
}
