// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package testing

import (
	stderrors "errors"
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
	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	dexkeeper "github.com/aequitas/aura/chain/x/dex/keeper"
	dextypes "github.com/aequitas/aura/chain/x/dex/types"
	securitykeeper "github.com/aequitas/aura/chain/x/security/keeper"
	securitytypes "github.com/aequitas/aura/chain/x/security/types"
	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

func setupDexWithRealSecurityKeeper(t *testing.T) (sdk.Context, *dexkeeper.Keeper, *securitykeeper.Keeper, *testutil.MockBankKeeper) {
	t.Helper()
	keepertest.ConfigureSDK()

	dexStoreKey := storetypes.NewKVStoreKey(dextypes.StoreKey)
	securityStoreKey := storetypes.NewKVStoreKey(securitytypes.StoreKey)
	securityMemKey := storetypes.NewMemoryStoreKey(securitytypes.MemStoreKey)

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(dexStoreKey, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(securityStoreKey, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(securityMemKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, cms.LoadLatestVersion())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	ctx := sdk.NewContext(
		cms,
		cmtproto.Header{Height: 1, Time: time.Now().UTC()},
		false,
		log.NewNopLogger(),
	)

	bankKeeper := testutil.NewMockBankKeeper()
	accountKeeper := testutil.NewMockAccountKeeper()

	authority := keepertest.GenTestAddr().String()
	secKeeper := securitykeeper.NewKeeper(
		cdc,
		securityStoreKey,
		securityMemKey,
		authority,
		bankKeeper,
		nil,
		accountKeeper,
	)
	secKeeper.SetParams(ctx, securitytypes.DefaultParams())

	dexKeeper := dexkeeper.NewKeeper(
		cdc,
		dexStoreKey,
		bankKeeper,
		accountKeeper,
		testutil.NewMockVCRegistryKeeper(),
		secKeeper,
	)
	params := dextypes.DefaultParams()
	require.NoError(t, dexKeeper.SetParams(ctx, &params))

	return ctx, dexKeeper, secKeeper, bankKeeper
}

// TestDexMsgFlowRespectsSecurityPause exercises the actual gRPC Msg path to ensure
// the DEX <-> Security pause boundary holds (not just direct keeper calls).
func TestDexMsgFlowRespectsSecurityPause(t *testing.T) {
	ctx, dexKeeper, secKeeper, bankKeeper := setupDexWithRealSecurityKeeper(t)

	creator := keepertest.GenTestAddr()
	amountA := sdk.NewCoin("uaura", sdkmath.NewInt(2_000_000_000))
	amountB := sdk.NewCoin("uusdc", sdkmath.NewInt(2_000_000_000))
	bankKeeper.Balances[creator.String()] = sdk.NewCoins(amountA, amountB)

	msgServer := dexkeeper.NewMsgServerImpl(dexKeeper)
	msg := &dexpb.MsgCreatePool{
		Creator: creator.String(),
		DenomA:  "uaura",
		DenomB:  "uusdc",
		AmountA: amountA,
		AmountB: amountB,
	}

	// Pause DEX via the real security keeper authority; msg server should now reject CreatePool.
	require.NoError(t, secKeeper.PauseModule(ctx, dextypes.ModuleName, secKeeper.GetAuthority()))
	_, err := msgServer.CreatePool(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)
	require.True(t, stderrors.Is(err, securitytypes.ErrSystemPaused), "CreatePool must be blocked when module is paused")
}

// TestDexMsgFlowReentrancyGuardIsEnforcedAndReleased verifies the DEX CreatePool msg
// enforces the reentrancy lock in the security keeper memstore and does not leak locks.
func TestDexMsgFlowReentrancyGuardIsEnforcedAndReleased(t *testing.T) {
	ctx, dexKeeper, secKeeper, bankKeeper := setupDexWithRealSecurityKeeper(t)

	creator := keepertest.GenTestAddr()
	amountA := sdk.NewCoin("uaura", sdkmath.NewInt(2_000_000_000))
	amountB := sdk.NewCoin("uusdc", sdkmath.NewInt(2_000_000_000))
	bankKeeper.Balances[creator.String()] = sdk.NewCoins(amountA, amountB)

	msgServer := dexkeeper.NewMsgServerImpl(dexKeeper)
	msg := &dexpb.MsgCreatePool{
		Creator: creator.String(),
		DenomA:  "uaura",
		DenomB:  "uusdc",
		AmountA: amountA,
		AmountB: amountB,
	}

	// Successful call must not leave reentrancy lock behind.
	_, err := msgServer.CreatePool(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)

	lockKey := securitytypes.GetReentrancyLockKey("dex:CreatePool")
	require.False(t, secKeeper.GetMemStore(ctx).Has(lockKey), "DEX CreatePool must release reentrancy lock on success")

	// If a lock already exists, msg server must fail.
	require.NoError(t, secKeeper.EnterNoReentrant(ctx, "dex:CreatePool"))
	_, err = msgServer.CreatePool(sdk.WrapSDKContext(ctx), &dexpb.MsgCreatePool{
		Creator: keepertest.GenTestAddr().String(),
		DenomA:  "uatom",
		DenomB:  "uusdc",
		AmountA: sdk.NewCoin("uatom", sdkmath.NewInt(2_000_000_000)),
		AmountB: sdk.NewCoin("uusdc", sdkmath.NewInt(2_000_000_000)),
	})
	require.Error(t, err)
	require.True(t, stderrors.Is(err, securitytypes.ErrReentrancyDetected), "CreatePool must be blocked under active reentrancy lock")
}
