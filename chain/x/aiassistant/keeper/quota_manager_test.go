package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/aiassistant/keeper"
)

func TestQuotaDefaultsAndResets(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	addr := "aura1quotaaddr"

	initial := k.GetQuota(ctx, addr)
	require.Equal(t, keeper.TierFree, initial.Tier)
	require.Equal(t, addr, initial.Address)
	require.Equal(t, uint64(0), initial.DailyUsed)
	require.True(t, initial.NextMonthlyReset.After(ctx.BlockTime()))

	// Persist a quota at its limits to make sure resets trigger
	initial.DailyUsed = initial.DailyLimit
	initial.MonthlyUsed = initial.MonthlyLimit
	initial.ComputeUsed = initial.ComputeLimit
	initial.TokenSpent = initial.TokenAllowance
	initial.LastReset = ctx.BlockTime().Add(-25 * time.Hour)
	initial.NextMonthlyReset = ctx.BlockTime().Add(-time.Hour)
	require.NoError(t, k.SetQuota(ctx, initial))

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(48 * time.Hour))
	refreshed := k.GetQuota(ctx, addr)

	require.Zero(t, refreshed.DailyUsed, "daily usage should reset after 24h")
	require.Zero(t, refreshed.MonthlyUsed, "monthly usage should reset after monthly boundary")
	require.Zero(t, refreshed.ComputeUsed)
	require.True(t, refreshed.TokenSpent.IsZero())
	require.True(t, refreshed.NextMonthlyReset.After(ctx.BlockTime()))
}

func TestQuotaCheckEnforcesLimits(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	addr := "aura1limits"

	quota := keeper.GetDefaultQuota(keeper.TierBasic)
	quota.Address = addr
	quota.LastReset = ctx.BlockTime()
	quota.NextMonthlyReset = ctx.BlockTime().AddDate(0, 1, 0)
	quota.DailyUsed = quota.DailyLimit
	require.NoError(t, k.SetQuota(ctx, quota))
	err := k.CheckQuota(ctx, addr, keeper.CostEstimate{ComputeUnits: 1, TokenCost: sdkmath.NewInt(1)})
	require.ErrorContains(t, err, "daily quota exceeded")

	quota = keeper.GetDefaultQuota(keeper.TierBasic)
	quota.Address = addr
	quota.LastReset = ctx.BlockTime()
	quota.NextMonthlyReset = ctx.BlockTime().AddDate(0, 1, 0)
	quota.MonthlyUsed = quota.MonthlyLimit
	require.NoError(t, k.SetQuota(ctx, quota))
	err = k.CheckQuota(ctx, addr, keeper.CostEstimate{ComputeUnits: 1, TokenCost: sdkmath.NewInt(1)})
	require.ErrorContains(t, err, "monthly quota exceeded")

	quota = keeper.GetDefaultQuota(keeper.TierBasic)
	quota.Address = addr
	quota.LastReset = ctx.BlockTime()
	quota.NextMonthlyReset = ctx.BlockTime().AddDate(0, 1, 0)
	quota.ComputeLimit = 10
	quota.ComputeUsed = 9
	require.NoError(t, k.SetQuota(ctx, quota))
	err = k.CheckQuota(ctx, addr, keeper.CostEstimate{ComputeUnits: 5, TokenCost: sdkmath.NewInt(1)})
	require.ErrorContains(t, err, "compute quota exceeded")

	quota = keeper.GetDefaultQuota(keeper.TierBasic)
	quota.Address = addr
	quota.LastReset = ctx.BlockTime()
	quota.NextMonthlyReset = ctx.BlockTime().AddDate(0, 1, 0)
	quota.TokenAllowance = sdkmath.NewInt(100)
	quota.TokenSpent = sdkmath.NewInt(90)
	require.NoError(t, k.SetQuota(ctx, quota))
	err = k.CheckQuota(ctx, addr, keeper.CostEstimate{ComputeUnits: 1, TokenCost: sdkmath.NewInt(20)})
	require.ErrorContains(t, err, "token allowance exceeded")
}

func TestConsumeQuotaPersistsUsage(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	addr := "aura1consume"
	estimate := keeper.CostEstimate{
		TokenCost:    sdkmath.NewInt(500),
		ComputeUnits: 250,
	}

	require.NoError(t, k.ConsumeQuota(ctx, addr, estimate))
	stored := k.GetQuota(ctx, addr)

	require.Equal(t, uint64(1), stored.DailyUsed)
	require.Equal(t, uint64(1), stored.MonthlyUsed)
	require.Equal(t, uint64(250), stored.ComputeUsed)
	require.True(t, stored.TokenSpent.Equal(estimate.TokenCost))
}

func TestUpgradeQuotaTier(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	addr := "aura1tier"
	quota := keeper.GetDefaultQuota(keeper.TierFree)
	quota.Address = addr
	quota.LastReset = ctx.BlockTime()
	quota.NextMonthlyReset = ctx.BlockTime().AddDate(0, 1, 0)
	quota.DailyUsed = 5
	require.NoError(t, k.SetQuota(ctx, quota))

	require.NoError(t, k.UpgradeQuotaTier(ctx, addr, keeper.TierPremium))
	upgraded := k.GetQuota(ctx, addr)
	require.Equal(t, keeper.TierPremium, upgraded.Tier)
	require.Equal(t, uint64(5), upgraded.DailyUsed, "usage counters should carry forward to new tier")
	require.Greater(t, upgraded.MonthlyLimit, quota.MonthlyLimit)

	err := k.UpgradeQuotaTier(ctx, addr, keeper.TierBasic)
	require.Error(t, err, "downgrades should be rejected")
}

func TestSetSponsoredQuota(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	user := "aura1user"
	sponsor := "aura1sponsor"

	require.NoError(t, k.SetSponsoredQuota(ctx, user, sponsor, keeper.TierPremium))
	sponsored := k.GetQuota(ctx, user)

	require.Equal(t, sponsor, sponsored.SponsoredBy)
	require.Equal(t, keeper.TierPremium, sponsored.Tier)
	require.False(t, sponsored.LastReset.IsZero())
	require.True(t, sponsored.NextMonthlyReset.After(ctx.BlockTime()))
}

func TestQuotaStatsAndList(t *testing.T) {
	k, ctx, _ := setupKeeper(t)

	entries := []struct {
		addr string
		tier keeper.QuotaTier
	}{
		{"aura1basic", keeper.TierBasic},
		{"aura1premium", keeper.TierPremium},
		{"aura1enterprise", keeper.TierEnterprise},
	}

	for _, entry := range entries {
		q := keeper.GetDefaultQuota(entry.tier)
		q.Address = entry.addr
		q.LastReset = ctx.BlockTime()
		q.NextMonthlyReset = ctx.BlockTime().AddDate(0, 1, 0)
		require.NoError(t, k.SetQuota(ctx, q))
	}

	stats := k.GetQuotaStats(ctx)
	require.Equal(t, uint64(1), stats[keeper.TierBasic])
	require.Equal(t, uint64(1), stats[keeper.TierPremium])
	require.Equal(t, uint64(1), stats[keeper.TierEnterprise])

	all := k.ListQuotas(ctx)
	require.Len(t, all, len(entries))

	addresses := make(map[string]struct{})
	for _, q := range all {
		addresses[q.Address] = struct{}{}
	}
	for _, entry := range entries {
		_, ok := addresses[entry.addr]
		require.True(t, ok, "missing quota for %s", entry.addr)
	}
}
