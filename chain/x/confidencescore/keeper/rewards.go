package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the interface for bank module operations
type BankKeeper interface {
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// PriceOracle defines the interface for getting AURA price
type PriceOracle interface {
	GetAuraPrice(ctx sdk.Context) math.LegacyDec
}

// ============================
// PoI REWARD DISTRIBUTION
// Based on whitepaper Section 12.0
// ============================

// RewardTier defines the reward structure for different AURA price levels
type RewardTier struct {
	MaxPrice     math.LegacyDec // Maximum AURA price for this tier (0 = unlimited)
	RewardAmount math.Int       // Fixed AURA amount (for low price tiers)
	UseUSDCap    bool           // Whether to use USD cap instead of fixed amount
	USDCap       math.LegacyDec // Maximum USD value (for high price tiers)
}

// GetRewardTiers returns the PoI reward tiers from whitepaper Section 12.0
func GetRewardTiers() []RewardTier {
	return []RewardTier{
		{
			// Tier 1: AURA < $0.11 → 500 AURA
			MaxPrice:     math.LegacyNewDecWithPrec(11, 2), // $0.11
			RewardAmount: math.NewInt(500_000_000),         // 500 AURA (in uaura: 500 * 1e6)
			UseUSDCap:    false,
		},
		{
			// Tier 2: $0.11 ≤ AURA < $0.30 → 250 AURA
			MaxPrice:     math.LegacyNewDecWithPrec(30, 2), // $0.30
			RewardAmount: math.NewInt(250_000_000),         // 250 AURA
			UseUSDCap:    false,
		},
		{
			// Tier 3: $0.30 ≤ AURA < $0.50 → 100 AURA
			MaxPrice:     math.LegacyNewDecWithPrec(50, 2), // $0.50
			RewardAmount: math.NewInt(100_000_000),         // 100 AURA
			UseUSDCap:    false,
		},
		{
			// Tier 4: AURA ≥ $0.50 → Variable amount not exceeding $50 USD
			MaxPrice:     math.LegacyZeroDec(), // Unlimited
			RewardAmount: math.ZeroInt(),       // Calculated dynamically
			UseUSDCap:    true,
			USDCap:       math.LegacyNewDec(50), // $50 USD max
		},
	}
}

// CalculatePoIReward calculates the total PoI reward based on current AURA price
// Returns the reward amount in uaura (micro-AURA)
func (k *Keeper) CalculatePoIReward(ctx sdk.Context, auraPrice math.LegacyDec) math.Int {
	tiers := GetRewardTiers()

	for _, tier := range tiers {
		// If tier has no max (unlimited), or price is below tier max
		if tier.MaxPrice.IsZero() || auraPrice.LT(tier.MaxPrice) {
			if tier.UseUSDCap {
				// Tier 4: Calculate reward based on USD cap
				// reward = USD_cap / AURA_price
				// Example: $50 / $0.75 = 66.67 AURA
				rewardDec := tier.USDCap.Quo(auraPrice)
				return rewardDec.Mul(math.LegacyNewDec(1_000_000)).TruncateInt() // Convert to uaura
			}
			// Tiers 1-3: Return fixed AURA amount
			return tier.RewardAmount
		}
	}

	// Fallback (should never reach here)
	return math.NewInt(100_000_000) // 100 AURA default
}

// SplitPoIReward splits the PoI reward between User and Node Operator
// Based on whitepaper Section 6.5
// Returns (userReward, nodeOperatorReward)
func (k *Keeper) SplitPoIReward(totalReward math.Int, userSplitPercent uint64) (math.Int, math.Int) {
	if userSplitPercent > 100 {
		userSplitPercent = 50 // Default to 50/50 if invalid
	}

	// User reward = total * (user_percent / 100)
	userReward := totalReward.Mul(math.NewInt(int64(userSplitPercent))).Quo(math.NewInt(100))

	// Node operator reward = total - user_reward
	nodeReward := totalReward.Sub(userReward)

	return userReward, nodeReward
}

// DistributePoIReward mints and distributes PoI rewards for IR completion
// This is called when a user successfully completes an IR
func (k *Keeper) DistributePoIReward(
	ctx sdk.Context,
	userAddress string,
	nodeOperatorAddress string,
	irID string,
	auraPrice math.LegacyDec,
	bankKeeper BankKeeper,
) error {
	params := k.GetParams()

	// Check if rewards are enabled
	if !params.PoIRewardsEnabled {
		return fmt.Errorf("PoI rewards are currently disabled")
	}

	// Calculate total reward based on current AURA price
	totalReward := k.CalculatePoIReward(ctx, auraPrice)

	// Split reward between user and node operator
	userReward, nodeReward := k.SplitPoIReward(totalReward, params.UserRewardSplitPercent)

	// Mint total reward to module account
	totalCoins := sdk.NewCoins(sdk.NewCoin("uaura", totalReward))
	if err := bankKeeper.MintCoins(ctx, types.ModuleName, totalCoins); err != nil {
		return fmt.Errorf("failed to mint PoI rewards: %w", err)
	}

	// Send user portion
	userAddr, err := sdk.AccAddressFromBech32(userAddress)
	if err != nil {
		return fmt.Errorf("invalid user address: %w", err)
	}
	userCoins := sdk.NewCoins(sdk.NewCoin("uaura", userReward))
	if err := bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, userAddr, userCoins); err != nil {
		return fmt.Errorf("failed to send user reward: %w", err)
	}

	// Send node operator portion
	nodeAddr, err := sdk.AccAddressFromBech32(nodeOperatorAddress)
	if err != nil {
		return fmt.Errorf("invalid node operator address: %w", err)
	}
	nodeCoins := sdk.NewCoins(sdk.NewCoin("uaura", nodeReward))
	if err := bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, nodeAddr, nodeCoins); err != nil {
		return fmt.Errorf("failed to send node operator reward: %w", err)
	}

	// Record reward distribution in completion record
	completion, ok := k.GetIRCompletion(userAddress, irID)
	if ok {
		completion.RewardAmount = totalReward.Uint64()
		completion.UserReward = userReward.Uint64()
		completion.NodeOperatorReward = nodeReward.Uint64()
		k.SetIRCompletion(userAddress, completion)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"poi_reward_distributed",
			sdk.NewAttribute("user", userAddress),
			sdk.NewAttribute("node_operator", nodeOperatorAddress),
			sdk.NewAttribute("ir_id", irID),
			sdk.NewAttribute("total_reward", totalReward.String()),
			sdk.NewAttribute("user_reward", userReward.String()),
			sdk.NewAttribute("node_reward", nodeReward.String()),
			sdk.NewAttribute("aura_price", auraPrice.String()),
		),
	)

	return nil
}

// CalculateVBTBoost calculates Velocity Bonus Tier multiplier
// Based on whitepaper Section 7.6
// Returns multiplier as decimal (1.0 = no boost, 1.5 = 50% boost, etc.)
func (k *Keeper) CalculateVBTBoost(completionTimeSeconds int64, irID string) math.LegacyDec {
	params := k.GetParams()

	if !params.VelocityBonusEnabled {
		return math.LegacyOneDec() // No boost
	}

	// Get IR expected completion time
	// This would come from IR definition - using placeholder for now
	expectedTime := int64(3600) // 1 hour default

	// Calculate completion speed ratio
	// ratio = expected_time / actual_time
	if completionTimeSeconds <= 0 {
		return math.LegacyOneDec()
	}

	ratio := math.LegacyNewDec(expectedTime).Quo(math.LegacyNewDec(completionTimeSeconds))

	// Apply VBT tiers (from whitepaper Section 7.6)
	// Tier 1: < 50% expected time → 2.0x
	// Tier 2: < 75% expected time → 1.5x
	// Tier 3: < 100% expected time → 1.25x
	// Tier 4: >= 100% expected time → 1.0x

	if ratio.GTE(math.LegacyNewDec(2)) { // Completed in ≤50% of expected time
		return math.LegacyNewDecWithPrec(20, 1) // 2.0x
	} else if ratio.GTE(math.LegacyNewDecWithPrec(133, 2)) { // ≤75% of expected time
		return math.LegacyNewDecWithPrec(15, 1) // 1.5x
	} else if ratio.GTE(math.LegacyOneDec()) { // ≤100% of expected time
		return math.LegacyNewDecWithPrec(125, 2) // 1.25x
	}

	return math.LegacyOneDec() // 1.0x (no boost)
}

// ApplyVBTBoost applies velocity bonus to a reward
func (k *Keeper) ApplyVBTBoost(baseReward math.Int, vbtMultiplier math.LegacyDec) math.Int {
	boostedReward := math.LegacyNewDecFromInt(baseReward).Mul(vbtMultiplier)
	return boostedReward.TruncateInt()
}

// ============================
// REWARD QUERIES
// ============================

// GetCurrentRewardAmount returns the current PoI reward amount for new completions
func (k *Keeper) GetCurrentRewardAmount(auraPrice math.LegacyDec) math.Int {
	return k.CalculatePoIReward(sdk.Context{}, auraPrice)
}

// GetRewardTierInfo returns information about the current reward tier
func (k *Keeper) GetRewardTierInfo(auraPrice math.LegacyDec) (tierName string, rewardAmount math.Int) {
	if auraPrice.LT(math.LegacyNewDecWithPrec(11, 2)) {
		return "Bootstrap (<$0.11)", math.NewInt(500_000_000)
	} else if auraPrice.LT(math.LegacyNewDecWithPrec(30, 2)) {
		return "Early Growth ($0.11-$0.30)", math.NewInt(250_000_000)
	} else if auraPrice.LT(math.LegacyNewDecWithPrec(50, 2)) {
		return "Growth ($0.30-$0.50)", math.NewInt(100_000_000)
	} else {
		// Calculate dynamic reward
		reward := k.CalculatePoIReward(sdk.Context{}, auraPrice)
		return "Established (≥$0.50)", reward
	}
}

// GetTotalRewardsDistributed returns total PoI rewards distributed to a user
func (k *Keeper) GetTotalRewardsDistributed(walletAddr string) uint64 {
	k.mu.RLock()
	defer k.mu.RUnlock()

	userCompletions, ok := k.completions[walletAddr]
	if !ok {
		return 0
	}

	total := uint64(0)
	for _, completion := range userCompletions {
		total += completion.RewardAmount
	}

	return total
}
