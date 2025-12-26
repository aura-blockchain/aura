// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/security/types"
)

func TestCreateIncident(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	incident := keeper.CreateIncident(ctx, "exploit", "critical", "Test exploit", "reporter1")
	require.NotNil(t, incident)
	require.Equal(t, "INC-1", incident.IncidentId)
	require.Equal(t, "exploit", incident.Title)

	// Create another - ID should increment
	incident2 := keeper.CreateIncident(ctx, "attack", "high", "Another issue", "reporter2")
	require.Equal(t, "INC-2", incident2.IncidentId)
}

func TestSetGetIncident(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	incident := keeper.CreateIncident(ctx, "test", "low", "desc", "reporter")
	retrieved, found := keeper.GetIncident(ctx, incident.IncidentId)
	require.True(t, found)
	require.Equal(t, incident.IncidentId, retrieved.IncidentId)

	_, found = keeper.GetIncident(ctx, "nonexistent")
	require.False(t, found)
}

func TestResolveIncident(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	incident := keeper.CreateIncident(ctx, "test", "medium", "desc", "reporter")
	err := keeper.ResolveIncident(ctx, incident.IncidentId, []string{"action1"})
	require.NoError(t, err)

	resolved, _ := keeper.GetIncident(ctx, incident.IncidentId)
	require.NotNil(t, resolved.ResolvedAt)

	// Resolve nonexistent
	err = keeper.ResolveIncident(ctx, "bad-id", nil)
	require.Error(t, err)
}

func TestPauseResumeSystem(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	require.False(t, keeper.IsSystemPaused(ctx))
	require.Equal(t, uint32(0), keeper.GetPauseLevel(ctx))

	err := keeper.PauseSystem(ctx, 2, "emergency", "admin")
	require.NoError(t, err)
	require.True(t, keeper.IsSystemPaused(ctx))
	require.Equal(t, uint32(2), keeper.GetPauseLevel(ctx))

	// Double pause fails
	err = keeper.PauseSystem(ctx, 1, "again", "admin")
	require.Error(t, err)

	err = keeper.ResumeSystem(ctx)
	require.NoError(t, err)
	require.False(t, keeper.IsSystemPaused(ctx))
}

func TestCheckTransactionAllowed(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	coins := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(100)))

	// Normal state - allowed
	err := keeper.CheckTransactionAllowed(ctx, "sender1", coins)
	require.NoError(t, err)

	// Pause at level 2 - blocked
	_ = keeper.PauseSystem(ctx, 2, "test", "admin")
	err = keeper.CheckTransactionAllowed(ctx, "sender1", coins)
	require.Error(t, err)

	_ = keeper.ResumeSystem(ctx)

	// Pause at level 1 - allowed (only transfers blocked, checked elsewhere)
	_ = keeper.PauseSystem(ctx, 1, "test", "admin")
	err = keeper.CheckTransactionAllowed(ctx, "sender1", coins)
	require.NoError(t, err)
}

func TestWalletLimits(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	limit, found := keeper.GetWalletLimit(ctx, "wallet1")
	require.False(t, found)
	require.Nil(t, limit)

	keeper.SetWalletLimit(ctx, &types.WalletLimit{
		WalletAddress: "wallet1",
		MaxTxAmount:   "500",
	})

	limit, found = keeper.GetWalletLimit(ctx, "wallet1")
	require.True(t, found)
	require.Equal(t, "500", limit.MaxTxAmount)

	limits := keeper.GetAllWalletLimits(ctx)
	require.Len(t, limits, 1)

	keeper.DeleteWalletLimit(ctx, "wallet1")
	_, found = keeper.GetWalletLimit(ctx, "wallet1")
	require.False(t, found)
}
