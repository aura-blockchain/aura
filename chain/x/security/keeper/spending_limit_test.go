// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"sync"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testing/testutil"
	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

var setupSDKConfigOnce sync.Once

func newTestSecurityKeeper(t *testing.T) (Keeper, sdk.Context) {
	t.Helper()

	// Configure bech32 prefixes for "aura" chain
	setupSDKConfigOnce.Do(func() {
		cfg := sdk.GetConfig()
		cfg.SetBech32PrefixForAccount("aura", "aurapub")
		cfg.SetBech32PrefixForValidator("auravaloper", "auravaloperpub")
		cfg.SetBech32PrefixForConsensusNode("auravalcons", "auravalconspub")
		cfg.Seal()
	})

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(memKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, cms.LoadLatestVersion())

	header := cmtproto.Header{Height: 1, Time: time.Now()}
	ctx := sdk.NewContext(cms, header, false, log.NewNopLogger())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	keeper := NewKeeper(
		cdc,
		storeKey,
		memKey,
		"aura1v96hg6r0wf5hg72lta047h6lta047h6lxkle3a", // valid bech32 test authority
		testutil.NewMockBankKeeper(),
		nil,
		testutil.NewMockAccountKeeper(),
	)
	keeper.SetParams(ctx, types.DefaultParams())

	return *keeper, ctx
}

func TestCheckSpendingLimit_AllowsWithinLimit(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	keeper.SetSpendingLimit(ctx, &securitypb.SpendingLimit{
		WalletId:          "wallet-1",
		Denom:             "uaura",
		DailyLimit:        sdkmath.NewInt(2_000).String(),
		CurrentDailySpent: sdkmath.NewInt(500).String(),
		Enabled:           true,
	})

	err := keeper.CheckSpendingLimit(ctx, "wallet-1", "uaura", sdkmath.NewInt(1_400).String())
	require.NoError(t, err, "spending within remaining allowance should succeed")
}

func TestCheckSpendingLimit_BlocksWhenExceeding(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	keeper.SetSpendingLimit(ctx, &securitypb.SpendingLimit{
		WalletId:          "wallet-2",
		Denom:             "uaura",
		DailyLimit:        sdkmath.NewInt(1_000).String(),
		CurrentDailySpent: sdkmath.NewInt(400).String(),
		Enabled:           true,
	})

	err := keeper.CheckSpendingLimit(ctx, "wallet-2", "uaura", sdkmath.NewInt(700).String())
	require.Error(t, err, "request that exceeds available headroom must be rejected")
}

func TestCheckSpendingLimit_InvalidInputsRejected(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	keeper.SetSpendingLimit(ctx, &securitypb.SpendingLimit{
		WalletId:          "wallet-3",
		Denom:             "uaura",
		DailyLimit:        "abc", // invalid
		CurrentDailySpent: sdkmath.ZeroInt().String(),
		Enabled:           true,
	})

	err := keeper.CheckSpendingLimit(ctx, "wallet-3", "uaura", sdkmath.NewInt(1).String())
	require.Error(t, err, "invalid configured limits must surface as an error")

	err = keeper.CheckSpendingLimit(ctx, "wallet-3", "uaura", "not-a-number")
	require.Error(t, err, "non-numeric requested amounts must be rejected")
}
