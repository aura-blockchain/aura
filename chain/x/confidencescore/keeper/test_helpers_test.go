// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	runtime "github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/confidencescore/params"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

// setupConfKeeper creates a test context and keeper for confidencescore module
func setupConfKeeper(t *testing.T) (sdk.Context, *Keeper) {
	t.Helper()

	input := keepertest.CreateTestInputWithStoreKey(t, types.StoreKey)

	paramsStore := params.NewStore(types.DefaultParams())

	k := NewKeeper(
		runtime.NewKVStoreService(input.StoreKey),
		input.Cdc,
		paramsStore,
		"authority",
		log.NewNopLogger(),
	)

	return input.Ctx, k
}

// setupConfKeeperWithTime creates a test context with current time and keeper
func setupConfKeeperWithTime(t *testing.T) (sdk.Context, *Keeper) {
	ctx, k := setupConfKeeper(t)
	return ctx.WithBlockTime(time.Now()), k
}
