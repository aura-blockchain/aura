// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"cosmossdk.io/log"
	runtime "github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/inclusionroutines/params"
	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
)

// setupInclusionKeeper creates a test context and keeper for inclusionroutines module
func setupInclusionKeeper(t *testing.T) (sdk.Context, *Keeper) {
	return setupInclusionKeeperWithAuthority(t, "authority")
}

// setupInclusionKeeperWithAuthority creates a test context and keeper with custom authority
func setupInclusionKeeperWithAuthority(t *testing.T, authority string) (sdk.Context, *Keeper) {
	t.Helper()

	input := keepertest.CreateTestInputWithStoreKey(t, types.StoreKey)

	paramsStore := params.NewStore(types.DefaultParams())

	k := NewKeeper(
		runtime.NewKVStoreService(input.StoreKey),
		input.Cdc,
		paramsStore,
		authority,
		log.NewNopLogger(),
	)

	return input.Ctx, k
}
