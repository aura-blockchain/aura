// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

// QuotaTier defines different quota tiers
type QuotaTier string

const (
	TierFree       QuotaTier = "free"
	TierBasic      QuotaTier = "basic"
	TierPremium    QuotaTier = "premium"
	TierEnterprise QuotaTier = "enterprise"
)

// Quota defines usage quota for an address
type Quota struct {
	Address          string
	Tier             QuotaTier
	MonthlyLimit     uint64      // Monthly query limit
	DailyLimit       uint64      // Daily query limit
	MonthlyUsed      uint64      // Queries used this month
	DailyUsed        uint64      // Queries used today
	ComputeLimit     uint64      // Compute units limit
	ComputeUsed      uint64      // Compute units used
	TokenAllowance   sdkmath.Int // Token allowance
	TokenSpent       sdkmath.Int // Tokens spent
	LastReset        time.Time
	NextMonthlyReset time.Time
	CustomLimits     map[string]uint64 // Custom limits per operation type
	IsUnlimited      bool              // Unlimited quota flag
	SponsoredBy      string            // Sponsor address if sponsored
}

// GetDefaultQuota returns default quota for a tier
func GetDefaultQuota(tier QuotaTier) Quota {
	switch tier {
	case TierFree:
		return Quota{
			Tier:           TierFree,
			MonthlyLimit:   1000,
			DailyLimit:     100,
			ComputeLimit:   100000,
			TokenAllowance: sdkmath.NewInt(1000000),
			CustomLimits:   make(map[string]uint64),
			TokenSpent:     sdkmath.ZeroInt(),
		}
	case TierBasic:
		return Quota{
			Tier:           TierBasic,
			MonthlyLimit:   10000,
			DailyLimit:     1000,
			ComputeLimit:   1000000,
			TokenAllowance: sdkmath.NewInt(10000000),
			CustomLimits:   make(map[string]uint64),
			TokenSpent:     sdkmath.ZeroInt(),
		}
	case TierPremium:
		return Quota{
			Tier:           TierPremium,
			MonthlyLimit:   100000,
			DailyLimit:     10000,
			ComputeLimit:   10000000,
			TokenAllowance: sdkmath.NewInt(100000000),
			CustomLimits:   make(map[string]uint64),
			TokenSpent:     sdkmath.ZeroInt(),
		}
	case TierEnterprise:
		return Quota{
			Tier:           TierEnterprise,
			MonthlyLimit:   1000000,
			DailyLimit:     100000,
			ComputeLimit:   100000000,
			TokenAllowance: sdkmath.NewInt(1000000000),
			CustomLimits:   make(map[string]uint64),
			IsUnlimited:    true,
			TokenSpent:     sdkmath.ZeroInt(),
		}
	default:
		return GetDefaultQuota(TierFree)
	}
}

// GetQuota retrieves quota for an address
func (k Keeper) GetQuota(ctx sdk.Context, address string) Quota {
	store := ctx.KVStore(k.storeKey)
	key := types.QuotaKey(address)

	bz := store.Get(key)
	if len(bz) == 0 {
		// Return default free tier quota
		quota := GetDefaultQuota(TierFree)
		quota.Address = address
		quota.LastReset = ctx.BlockTime()
		quota.NextMonthlyReset = ctx.BlockTime().AddDate(0, 1, 0)
		return quota
	}

	var quota Quota
	if err := json.Unmarshal(bz, &quota); err != nil {
		return GetDefaultQuota(TierFree)
	}

	// Reset counters if needed
	now := ctx.BlockTime()
	if now.After(quota.NextMonthlyReset) {
		quota.MonthlyUsed = 0
		quota.ComputeUsed = 0
		quota.TokenSpent = sdkmath.ZeroInt()
		quota.NextMonthlyReset = now.AddDate(0, 1, 0)
	}
	if now.Sub(quota.LastReset) >= 24*time.Hour {
		quota.DailyUsed = 0
		quota.LastReset = now
	}

	return quota
}

// SetQuota stores quota for an address
func (k Keeper) SetQuota(ctx sdk.Context, quota Quota) error {
	store := ctx.KVStore(k.storeKey)
	key := types.QuotaKey(quota.Address)

	bz, err := json.Marshal(&quota)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	store.Set(key, bz)
	return nil
}

// CheckQuota validates if operation can proceed within quota
func (k Keeper) CheckQuota(ctx sdk.Context, address string, estimate CostEstimate) error {
	quota := k.GetQuota(ctx, address)

	// Skip checks for unlimited quota
	if quota.IsUnlimited {
		return nil
	}

	// Check daily limit
	if quota.DailyUsed >= quota.DailyLimit {
		return sdkerrors.Wrap(types.ErrInvalidParams,
			fmt.Sprintf("daily quota exceeded: %d/%d", quota.DailyUsed, quota.DailyLimit))
	}

	// Check monthly limit
	if quota.MonthlyUsed >= quota.MonthlyLimit {
		return sdkerrors.Wrap(types.ErrInvalidParams,
			fmt.Sprintf("monthly quota exceeded: %d/%d", quota.MonthlyUsed, quota.MonthlyLimit))
	}

	// Check compute limit
	if quota.ComputeUsed+estimate.ComputeUnits > quota.ComputeLimit {
		return sdkerrors.Wrap(types.ErrInvalidParams,
			fmt.Sprintf("compute quota exceeded: %d/%d", quota.ComputeUsed, quota.ComputeLimit))
	}

	// Check token allowance
	if quota.TokenSpent.Add(estimate.TokenCost).GT(quota.TokenAllowance) {
		return sdkerrors.Wrap(types.ErrInvalidParams,
			fmt.Sprintf("token allowance exceeded: %s/%s", quota.TokenSpent, quota.TokenAllowance))
	}

	return nil
}

// ConsumeQuota deducts from quota after operation
func (k Keeper) ConsumeQuota(ctx sdk.Context, address string, estimate CostEstimate) error {
	quota := k.GetQuota(ctx, address)

	quota.DailyUsed++
	quota.MonthlyUsed++
	quota.ComputeUsed += estimate.ComputeUnits
	quota.TokenSpent = quota.TokenSpent.Add(estimate.TokenCost)

	return k.SetQuota(ctx, quota)
}

// UpgradeQuotaTier upgrades user to a higher tier
func (k Keeper) UpgradeQuotaTier(ctx sdk.Context, address string, newTier QuotaTier) error {
	quota := k.GetQuota(ctx, address)

	// Validate upgrade path
	tierOrder := map[QuotaTier]int{
		TierFree:       0,
		TierBasic:      1,
		TierPremium:    2,
		TierEnterprise: 3,
	}

	if tierOrder[newTier] <= tierOrder[quota.Tier] {
		return fmt.Errorf("cannot downgrade from %s to %s", quota.Tier, newTier)
	}

	// Get new tier limits
	newQuota := GetDefaultQuota(newTier)
	newQuota.Address = address
	newQuota.MonthlyUsed = quota.MonthlyUsed
	newQuota.DailyUsed = quota.DailyUsed
	newQuota.ComputeUsed = quota.ComputeUsed
	newQuota.TokenSpent = quota.TokenSpent
	newQuota.LastReset = quota.LastReset
	newQuota.NextMonthlyReset = quota.NextMonthlyReset

	return k.SetQuota(ctx, newQuota)
}

// SetSponsoredQuota sets sponsored quota for an address
func (k Keeper) SetSponsoredQuota(ctx sdk.Context, userAddr, sponsorAddr string, tier QuotaTier) error {
	quota := GetDefaultQuota(tier)
	quota.Address = userAddr
	quota.SponsoredBy = sponsorAddr
	quota.LastReset = ctx.BlockTime()
	quota.NextMonthlyReset = ctx.BlockTime().AddDate(0, 1, 0)

	return k.SetQuota(ctx, quota)
}

// GetQuotaStats returns quota statistics for analytics
func (k Keeper) GetQuotaStats(ctx sdk.Context) map[QuotaTier]uint64 {
	stats := make(map[QuotaTier]uint64)

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.QuotaKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var quota Quota
		if err := json.Unmarshal(iterator.Value(), &quota); err != nil {
			continue
		}
		stats[quota.Tier]++
	}

	return stats
}

// ListQuotas returns all quota records
func (k Keeper) ListQuotas(ctx sdk.Context) []Quota {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.QuotaKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var quotas []Quota
	for ; iterator.Valid(); iterator.Next() {
		var quota Quota
		if err := json.Unmarshal(iterator.Value(), &quota); err != nil {
			continue
		}
		quotas = append(quotas, quota)
	}

	return quotas
}
