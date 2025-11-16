package keeper

import (
	"math/big"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// MEV REDISTRIBUTION (Feature 12)
// ============================

// CaptureMEV captures MEV value for redistribution
func (k *Keeper) CaptureMEV(amount string) error {
	params := k.GetParams()

	if !params.Mev.Enabled {
		return types.ErrMEVRedistributionDisabled
	}

	mevAmt := new(big.Int)
	if _, ok := mevAmt.SetString(amount, 10); !ok {
		return types.ErrInvalidAmount
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Update total MEV captured
	totalCaptured := new(big.Int)
	totalCaptured.SetString(params.Mev.TotalMevCaptured, 10)
	totalCaptured.Add(totalCaptured, mevAmt)

	params.Mev.TotalMevCaptured = totalCaptured.String()

	// Add to pending redistribution
	k.totalMEVPending.Add(k.totalMEVPending, mevAmt)

	return k.SetParams(params)
}

// DistributeMEV distributes captured MEV to users, validators, treasury, and burn
func (k *Keeper) DistributeMEV(activeUsers []string, userActivity map[string]uint64, userIRScores map[string]uint64) (string, string, string, error) {
	params := k.GetParams()

	if !params.Mev.Enabled {
		return "0", "0", "0", types.ErrMEVRedistributionDisabled
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	if k.totalMEVPending.Cmp(big.NewInt(0)) == 0 {
		return "0", "0", "0", nil
	}

	totalMEV := new(big.Int).Set(k.totalMEVPending)

	// Calculate distribution amounts
	userShare := new(big.Int).Mul(totalMEV, big.NewInt(int64(params.Mev.UserRedistributionPercentage)))
	userShare.Div(userShare, big.NewInt(types.BasisPoints))

	validatorShare := new(big.Int).Mul(totalMEV, big.NewInt(int64(params.Mev.ValidatorPercentage)))
	validatorShare.Div(validatorShare, big.NewInt(types.BasisPoints))

	treasuryShare := new(big.Int).Mul(totalMEV, big.NewInt(int64(params.Mev.TreasuryPercentage)))
	treasuryShare.Div(treasuryShare, big.NewInt(types.BasisPoints))

	burnShare := new(big.Int).Mul(totalMEV, big.NewInt(int64(params.Mev.BurnPercentage)))
	burnShare.Div(burnShare, big.NewInt(types.BasisPoints))

	// Distribute to users based on strategy
	if err := k.distributeMEVToUsers(userShare, activeUsers, userActivity, userIRScores, params.Mev.Strategy); err != nil {
		return "0", "0", "0", err
	}

	// Update total redistributed
	totalRedistributed := new(big.Int)
	totalRedistributed.SetString(params.Mev.TotalMevRedistributed, 10)
	totalRedistributed.Add(totalRedistributed, totalMEV)

	params.Mev.TotalMevRedistributed = totalRedistributed.String()

	// Update burned amount
	k.totalBurned.Add(k.totalBurned, burnShare)

	// Reset pending
	k.totalMEVPending.SetInt64(0)

	if err := k.SetParams(params); err != nil {
		return "0", "0", "0", err
	}

	return validatorShare.String(), treasuryShare.String(), burnShare.String(), nil
}

// distributeMEVToUsers distributes MEV to users based on strategy
func (k *Keeper) distributeMEVToUsers(
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
	case types.MEVStrategyEqualDistribution:
		// Equal distribution among all users
		perUser := new(big.Int).Div(totalUserShare, big.NewInt(int64(len(activeUsers))))
		for _, user := range activeUsers {
			k.addUserMEVBalance(user, perUser)
		}

	case types.MEVStrategyProportionalToActivity:
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
				k.addUserMEVBalance(user, share)
			}
		}

	case types.MEVStrategyIRWeighted:
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
				k.addUserMEVBalance(user, share)
			}
		}

	case types.MEVStrategyProportionalToStake:
		// Would require stake data - not implemented in this simplified version
		// Fall back to equal distribution
		perUser := new(big.Int).Div(totalUserShare, big.NewInt(int64(len(activeUsers))))
		for _, user := range activeUsers {
			k.addUserMEVBalance(user, perUser)
		}

	default:
		return types.ErrInvalidRedistributionStrategy
	}

	return nil
}

// addUserMEVBalance adds to a user's MEV balance
func (k *Keeper) addUserMEVBalance(user string, amount *big.Int) {
	if k.userMEVBalances[user] == nil {
		k.userMEVBalances[user] = big.NewInt(0)
	}
	k.userMEVBalances[user].Add(k.userMEVBalances[user], amount)
}

// GetUserMEVBalance returns a user's MEV redistribution balance
func (k *Keeper) GetUserMEVBalance(address string) string {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if balance, ok := k.userMEVBalances[address]; ok {
		return balance.String()
	}
	return "0"
}

// ClaimMEVRewards allows a user to claim their MEV rewards
func (k *Keeper) ClaimMEVRewards(address string) (string, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	balance, ok := k.userMEVBalances[address]
	if !ok || balance.Cmp(big.NewInt(0)) == 0 {
		return "0", types.ErrInsufficientMEVBalance
	}

	amount := balance.String()

	// Reset balance
	k.userMEVBalances[address] = big.NewInt(0)

	return amount, nil
}

// GetMEVStats returns MEV redistribution statistics
func (k *Keeper) GetMEVStats() (bool, string, string, string, uint64, types.MEVRedistributionStrategy) {
	params := k.GetParams()

	if !params.Mev.Enabled {
		return false, "0", "0", "0", 0, types.MEVStrategyUnspecified
	}

	k.mu.RLock()
	pending := k.totalMEVPending.String()
	k.mu.RUnlock()

	return true,
		params.Mev.TotalMevCaptured,
		params.Mev.TotalMevRedistributed,
		pending,
		params.Mev.UserRedistributionPercentage,
		params.Mev.Strategy
}
