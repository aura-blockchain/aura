// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/incidentresponse/keeper"
	"github.com/aequitas/aura/chain/x/incidentresponse/types"
)

// IncidentResponseKeeper creates an incident response keeper for testing
func IncidentResponseKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	t.Helper()

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	require.NoError(t, stateStore.LoadLatestVersion())

	params := types.DefaultParams()
	k := keeper.NewKeeper(params)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	return k, ctx
}
