// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

func TestRateLimitOperations(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	// No rate limit initially
	rl, found := keeper.GetRateLimit(ctx, "peer1")
	require.False(t, found)
	require.Nil(t, rl)

	// Set rate limit
	keeper.SetRateLimit(ctx, &securitypb.RateLimitEntry{
		PeerId:       "peer1",
		RequestCount: 10,
		IsBanned:     false,
	})

	rl, found = keeper.GetRateLimit(ctx, "peer1")
	require.True(t, found)
	require.Equal(t, uint64(10), rl.RequestCount)

	// Get all
	all := keeper.GetAllRateLimits(ctx)
	require.Len(t, all, 1)

	// Delete
	keeper.DeleteRateLimit(ctx, "peer1")
	_, found = keeper.GetRateLimit(ctx, "peer1")
	require.False(t, found)
}

func TestPeerReputationOperations(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	keeper.SetPeerReputation(ctx, &securitypb.NodeReputation{
		PeerId: "peer1",
		Score:  85,
	})

	rep, found := keeper.GetPeerReputation(ctx, "peer1")
	require.True(t, found)
	require.Equal(t, int64(85), rep.Score)

	all := keeper.GetAllPeerReputations(ctx)
	require.Len(t, all, 1)
}

func TestTrustedPeerOperations(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	keeper.SetTrustedPeer(ctx, &securitypb.TrustedPeer{
		PeerId:      "peer1",
		Description: "test peer",
	})

	peer, found := keeper.GetTrustedPeer(ctx, "peer1")
	require.True(t, found)
	require.Equal(t, "peer1", peer.PeerId)

	all := keeper.GetAllTrustedPeers(ctx)
	require.Len(t, all, 1)

	keeper.DeleteTrustedPeer(ctx, "peer1")
	_, found = keeper.GetTrustedPeer(ctx, "peer1")
	require.False(t, found)
}

func TestBlacklistOperations(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	// Not blacklisted initially
	require.False(t, keeper.IsBlacklisted(ctx, "badactor"))

	// Add permanent blacklist
	keeper.SetBlacklistEntry(ctx, &types.BlacklistEntry{
		Identifier: "badactor",
		Permanent:  true,
	})
	require.True(t, keeper.IsBlacklisted(ctx, "badactor"))

	// Add temporary blacklist with future expiry
	future := time.Now().Add(1 * time.Hour)
	keeper.SetBlacklistEntry(ctx, &types.BlacklistEntry{
		Identifier: "tempbad",
		Permanent:  false,
		ExpiresAt:  &future,
	})
	require.True(t, keeper.IsBlacklisted(ctx, "tempbad"))

	// Get all
	all := keeper.GetAllBlacklistEntries(ctx)
	require.GreaterOrEqual(t, len(all), 2)

	// Delete
	keeper.DeleteBlacklistEntry(ctx, "badactor")
	require.False(t, keeper.IsBlacklisted(ctx, "badactor"))
}

func TestForkAlertOperations(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	keeper.SetForkAlert(ctx, &securitypb.ForkAlert{
		AlertId:     "fork-1",
		BlockHeight: 100,
	})

	alerts := keeper.GetAllForkAlerts(ctx)
	require.Len(t, alerts, 1)
	require.Equal(t, "fork-1", alerts[0].AlertId)
}

func TestPartitionAlertOperations(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	keeper.SetPartitionAlert(ctx, &securitypb.PartitionAlert{
		AlertId:        "partition-1",
		ConnectedPeers: 3,
		ExpectedPeers:  10,
	})

	alerts := keeper.GetAllPartitionAlerts(ctx)
	require.Len(t, alerts, 1)
	require.Equal(t, "partition-1", alerts[0].AlertId)
}

func TestCheckRateLimit(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	// First call creates entry
	err := keeper.CheckRateLimit(ctx, "newpeer")
	require.NoError(t, err)

	rl, found := keeper.GetRateLimit(ctx, "newpeer")
	require.True(t, found)
	require.Equal(t, uint64(1), rl.RequestCount)
}
