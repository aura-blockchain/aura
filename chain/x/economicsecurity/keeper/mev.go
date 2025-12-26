// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"math/big"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// MEV REDISTRIBUTION (Feature 12)
// ============================

// CaptureMEV captures MEV value for redistribution
func (k *Keeper) CaptureMEV(ctx context.Context, amount string) error {
	params, _ := k.GetParams(ctx)

	if !params.Mev.Enabled {
		return types.ErrMEVRedistributionDisabled
	}

	mevAmt := new(big.Int)
	if _, ok := mevAmt.SetString(amount, 10); !ok {
		return types.ErrInvalidAmount
	}

	// Get current total MEV captured from params
	totalCaptured := new(big.Int)
	if params.Mev.TotalMevCaptured != "" {
		totalCaptured.SetString(params.Mev.TotalMevCaptured, 10)
	}
	totalCaptured.Add(totalCaptured, mevAmt)
	params.Mev.TotalMevCaptured = totalCaptured.String()

	// Get current pending MEV from store
	pendingStr, err := k.GetTotalMEVPending(ctx)
	if err != nil {
		return err
	}

	pending := new(big.Int)
	if pendingStr != "0" {
		pending.SetString(pendingStr, 10)
	}
	pending.Add(pending, mevAmt)

	// Update pending MEV in store
	if err := k.SetTotalMEVPending(ctx, pending.String()); err != nil {
		return err
	}

	return k.SetParams(params)
}

// DistributeMEV distributes captured MEV to users, validators, treasury, and burn
func (k *Keeper) DistributeMEV(
	ctx context.Context,
	activeUsers []string,
	userActivity map[string]uint64,
	userIRScores map[string]uint64,
) (validatorShare string, treasuryShare string, burnShare string, err error) {
	params, _ := k.GetParams(ctx)

	if !params.Mev.Enabled {
		return "0", "0", "0", types.ErrMEVRedistributionDisabled
	}

	// Get pending MEV from store
	pendingStr, err := k.GetTotalMEVPending(ctx)
	if err != nil {
		return "0", "0", "0", err
	}

	totalMEVPending := new(big.Int)
	if pendingStr == "0" || pendingStr == "" {
		return "0", "0", "0", nil
	}
	totalMEVPending.SetString(pendingStr, 10)

	if totalMEVPending.Cmp(big.NewInt(0)) == 0 {
		return "0", "0", "0", nil
	}

	totalMEV := new(big.Int).Set(totalMEVPending)

	// Calculate distribution amounts
	basisPoints := big.NewInt(10000)

	userShareAmt := new(big.Int).Mul(totalMEV, big.NewInt(int64(params.Mev.UserRedistributionPercentage)))
	userShareAmt.Div(userShareAmt, basisPoints)

	validatorShareAmt := new(big.Int).Mul(totalMEV, big.NewInt(int64(params.Mev.ValidatorPercentage)))
	validatorShareAmt.Div(validatorShareAmt, basisPoints)

	treasuryShareAmt := new(big.Int).Mul(totalMEV, big.NewInt(int64(params.Mev.TreasuryPercentage)))
	treasuryShareAmt.Div(treasuryShareAmt, basisPoints)

	burnShareAmt := new(big.Int).Mul(totalMEV, big.NewInt(int64(params.Mev.BurnPercentage)))
	burnShareAmt.Div(burnShareAmt, basisPoints)

	// Distribute to users based on strategy
	if err := k.distributeMEVToUsers(ctx, userShareAmt, activeUsers, userActivity, userIRScores, params.Mev.Strategy); err != nil {
		return "0", "0", "0", err
	}

	// Update total redistributed in params
	totalRedistributed := new(big.Int)
	if params.Mev.TotalMevRedistributed != "" {
		totalRedistributed.SetString(params.Mev.TotalMevRedistributed, 10)
	}
	totalRedistributed.Add(totalRedistributed, totalMEV)
	params.Mev.TotalMevRedistributed = totalRedistributed.String()

	// Update total burned in store
	burnedStr, err := k.GetTotalBurned(ctx)
	if err != nil {
		return "0", "0", "0", err
	}

	totalBurned := new(big.Int)
	if burnedStr != "0" {
		totalBurned.SetString(burnedStr, 10)
	}
	totalBurned.Add(totalBurned, burnShareAmt)

	if err := k.SetTotalBurned(ctx, totalBurned.String()); err != nil {
		return "0", "0", "0", err
	}

	// Reset pending MEV
	if err := k.SetTotalMEVPending(ctx, "0"); err != nil {
		return "0", "0", "0", err
	}

	if err := k.SetParams(params); err != nil {
		return "0", "0", "0", err
	}

	return validatorShareAmt.String(), treasuryShareAmt.String(), burnShareAmt.String(), nil
}

// distributeMEVToUsers distributes MEV to users based on strategy
func (k *Keeper) distributeMEVToUsers(
	ctx context.Context,
	totalUserShare *big.Int,
	activeUsers []string,
	userActivity map[string]uint64,
	userIRScores map[string]uint64,
	strategy types.MEVRedistributionStrategy,
) error {
	if len(activeUsers) == 0 {
		return nil
	}

	switch strategy {
	case types.MEVRedistributionStrategy_MEV_STRATEGY_UNSPECIFIED:
		// Default to proportional to activity if unspecified
		return k.distributeMEVToUsers(ctx, totalUserShare, activeUsers, userActivity, userIRScores,
			types.MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_ACTIVITY)

	case types.MEVRedistributionStrategy_MEV_STRATEGY_EQUAL_DISTRIBUTION:
		// Equal distribution among all users
		perUser := new(big.Int).Div(totalUserShare, big.NewInt(int64(len(activeUsers))))
		for _, user := range activeUsers {
			if err := k.addUserMEVBalance(ctx, user, perUser); err != nil {
				return err
			}
		}

	case types.MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_ACTIVITY:
		// Distribution proportional to activity
		totalActivity := uint64(0)
		for _, activity := range userActivity {
			totalActivity += activity
		}

		if totalActivity > 0 {
			for _, user := range activeUsers {
				activity := userActivity[user]
				share := new(big.Int).Mul(totalUserShare, big.NewInt(int64(activity)))
				share.Div(share, big.NewInt(int64(totalActivity)))
				if err := k.addUserMEVBalance(ctx, user, share); err != nil {
					return err
				}
			}
		}

	case types.MEVRedistributionStrategy_MEV_STRATEGY_IR_WEIGHTED:
		// Distribution weighted by IR score
		totalScore := uint64(0)
		for _, score := range userIRScores {
			totalScore += score
		}

		if totalScore > 0 {
			for _, user := range activeUsers {
				score := userIRScores[user]
				share := new(big.Int).Mul(totalUserShare, big.NewInt(int64(score)))
				share.Div(share, big.NewInt(int64(totalScore)))
				if err := k.addUserMEVBalance(ctx, user, share); err != nil {
					return err
				}
			}
		}

	case types.MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_STAKE:
		// For stake-based distribution, we'd need to query staking module
		// For now, fall back to equal distribution
		// In production, integrate with staking keeper to get delegations
		perUser := new(big.Int).Div(totalUserShare, big.NewInt(int64(len(activeUsers))))
		for _, user := range activeUsers {
			if err := k.addUserMEVBalance(ctx, user, perUser); err != nil {
				return err
			}
		}

	default:
		return types.ErrInvalidRedistributionStrategy
	}

	return nil
}

// addUserMEVBalance adds to a user's MEV balance
func (k *Keeper) addUserMEVBalance(ctx context.Context, user string, amount *big.Int) error {
	// Get current balance
	balanceStr, err := k.GetUserMEVBalance(ctx, user)
	if err != nil {
		return err
	}

	balance := new(big.Int)
	if balanceStr != "0" {
		balance.SetString(balanceStr, 10)
	}

	balance.Add(balance, amount)

	return k.SetUserMEVBalance(ctx, user, balance.String())
}

// ClaimMEVRewards allows a user to claim their MEV rewards
func (k *Keeper) ClaimMEVRewards(ctx context.Context, address string) (string, error) {
	balanceStr, err := k.GetUserMEVBalance(ctx, address)
	if err != nil {
		return "0", err
	}

	balance := new(big.Int)
	if balanceStr == "0" || balanceStr == "" {
		return "0", types.ErrInsufficientMEVBalance
	}

	balance.SetString(balanceStr, 10)
	if balance.Cmp(big.NewInt(0)) == 0 {
		return "0", types.ErrInsufficientMEVBalance
	}

	amount := balance.String()

	// Reset balance
	if err := k.SetUserMEVBalance(ctx, address, "0"); err != nil {
		return "0", err
	}

	return amount, nil
}

// GetMEVStats returns MEV redistribution statistics
func (k *Keeper) GetMEVStats(ctx context.Context) (
	enabled bool,
	totalCaptured string,
	totalRedistributed string,
	pending string,
	userPercentage uint64,
	strategy types.MEVRedistributionStrategy,
) {
	params, _ := k.GetParams(ctx)

	if !params.Mev.Enabled {
		return false, "0", "0", "0", 0, types.MEVRedistributionStrategy_MEV_STRATEGY_UNSPECIFIED
	}

	pendingStr, err := k.GetTotalMEVPending(ctx)
	if err != nil {
		pendingStr = "0"
	}

	return true,
		params.Mev.TotalMevCaptured,
		params.Mev.TotalMevRedistributed,
		pendingStr,
		params.Mev.UserRedistributionPercentage,
		params.Mev.Strategy
}

// GetAllUserMEVBalances returns all user MEV balances
func (k *Keeper) GetAllUserMEVBalances(ctx context.Context) (map[string]string, error) {
	balances := make(map[string]string)

	err := k.IterateUserMEVBalances(ctx, func(address string, balance string) bool {
		balances[address] = balance
		return false // Continue iterating
	})

	return balances, err
}
