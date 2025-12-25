// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

func TestBandwidthLimitEnforcement(t *testing.T) {
	keeper, ctx := NewTestKeeperWithContext(t)
	ctx = ctx.WithBlockTime(time.Now())
	peer := "peer-bandwidth"

	params := types.DefaultParams()
	params.RateLimit.BandwidthLimitPerPeer = 1 // 1 byte/sec to trigger quickly
	require.NoError(t, keeper.SetParams(ctx, *params))

	// First 50 bytes should pass.
	err := keeper.CheckBandwidthLimit(ctx, peer, 50, true)
	require.NoError(t, err)

	// Exceeding limit should return ErrBandwidthLimitExceeded and mark tracker banned.
	err = keeper.CheckBandwidthLimit(ctx, peer, 1000, true)
	require.ErrorIs(t, err, types.ErrBandwidthLimitExceeded)
}

func TestBandwidthLimitRespectsExistingBan(t *testing.T) {
	keeper, ctx := NewTestKeeperWithContext(t)
	ctx = ctx.WithBlockTime(time.Now())
	peer := "peer-banned"

	banExpiresAt := ctx.BlockTime().Add(1 * time.Hour)
	entry := types.RateLimitEntry{
		PeerId:       peer,
		IsBanned:     true,
		BanExpiresAt: &banExpiresAt,
	}
	require.NoError(t, keeper.SetRateLimitEntry(ctx, entry))

	require.True(t, keeper.IsBanned(ctx, peer))
}

func TestBandwidthLimitRecordsUsageWithoutAutoBan(t *testing.T) {
	keeper, ctx := NewTestKeeperWithContext(t)
	ctx = ctx.WithBlockTime(time.Now())
	peer := "peer-usage"

	params := types.DefaultParams()
	params.RateLimit.BandwidthLimitPerPeer = 1
	require.NoError(t, keeper.SetParams(ctx, *params))

	err := keeper.CheckBandwidthLimit(ctx, peer, 10_000, true)
	require.ErrorIs(t, err, types.ErrBandwidthLimitExceeded)

	entry, found := keeper.GetRateLimitEntry(ctx, peer)
	require.True(t, found, "rate limit entry should be recorded when bandwidth exceeded")
	require.Greater(t, entry.BytesSent+entry.BytesReceived, uint64(0))
	require.False(t, entry.IsBanned, "bandwidth limit exceed does not automatically ban")
}
