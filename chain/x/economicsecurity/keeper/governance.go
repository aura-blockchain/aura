// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// GOVERNANCE (Features 7, 8, 9)
// ============================

// CheckProposalStake validates if an address has sufficient stake to create a proposal (Feature 7)
func (k *Keeper) CheckProposalStake(ctx context.Context, proposer, stakeAmount string) error {
	params, _ := k.GetParams(ctx)

	minStake := new(big.Int)
	minStake.SetString(params.Governance.MinProposalStake, 10)

	stake := new(big.Int)
	if _, ok := stake.SetString(stakeAmount, 10); !ok {
		return types.ErrInvalidAmount
	}

	if stake.Cmp(minStake) < 0 {
		return types.ErrInsufficientStake
	}

	return nil
}

// CalculateQuadraticVotingPower calculates voting power using quadratic voting (Feature 8)
func (k *Keeper) CalculateQuadraticVotingPower(ctx context.Context, stakeAmount string) (string, error) {
	params, _ := k.GetParams(ctx)

	if !params.Governance.QuadraticVotingEnabled {
		// If quadratic voting disabled, return linear voting power
		return stakeAmount, nil
	}

	stake := new(big.Int)
	if _, ok := stake.SetString(stakeAmount, 10); !ok {
		return "0", types.ErrInvalidAmount
	}

	// Quadratic voting: voting power = sqrt(stake)
	// We use a simplified integer square root
	votingPower := sqrt(stake)

	return votingPower.String(), nil
}

// LockVotingTokens locks tokens for voting power boost (Feature 9)
func (k *Keeper) LockVotingTokens(ctx context.Context, owner, amount string, lockDuration uint64) (string, string, error) {
	params, _ := k.GetParams(ctx)

	if !params.Governance.VoteLockingEnabled {
		return "", "0", errors.New("vote locking is disabled")
	}

	if lockDuration < params.Governance.MinLockDuration {
		return "", "0", types.ErrLockDurationTooShort
	}

	if lockDuration > params.Governance.MaxLockDuration {
		return "", "0", types.ErrLockDurationTooLong
	}

	lockAmt := new(big.Int)
	if _, ok := lockAmt.SetString(amount, 10); !ok || lockAmt.Cmp(big.NewInt(0)) <= 0 {
		return "", "0", types.ErrInvalidAmount
	}

	// Get current time
	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return "", "0", err
	}

	// Generate lock ID
	lockID := generateLockID(owner, amount, lockDuration, currentTime)

	// Calculate voting power multiplier based on lock duration using integer math
	// multiplier = 1 + (duration / 1 year) * multiplier_per_year
	// Using basis points (10000 = 1.0) for deterministic consensus
	oneYear := uint64(31536000) // seconds in a year
	multiplierBasisPoints := params.Governance.LockMultiplierPerYear

	// Calculate time-based bonus in basis points
	// bonusBps = (lockDuration * multiplierBasisPoints) / oneYear
	bonusBps := (lockDuration * multiplierBasisPoints) / oneYear
	// totalMultiplierBps = 10000 (1.0) + bonusBps
	totalMultiplierBps := types.BasisPoints + bonusBps

	// votingPower = lockAmt * totalMultiplierBps / 10000
	votingPowerInt := new(big.Int).Mul(lockAmt, big.NewInt(int64(totalMultiplierBps)))
	votingPowerInt.Div(votingPowerInt, big.NewInt(int64(types.BasisPoints)))

	lock := &types.VoteLock{
		LockId:      lockID,
		Owner:       owner,
		Amount:      amount,
		LockStart:   time.Unix(currentTime, 0),
		LockEnd:     time.Unix(currentTime+int64(lockDuration), 0),
		VotingPower: votingPowerInt.String(),
		Withdrawn:   false,
	}

	// Store the lock
	if err := k.SetVoteLock(ctx, lock); err != nil {
		return "", "0", err
	}

	// Add to user's vote lock index
	if err := k.AddUserVoteLock(ctx, owner, lockID); err != nil {
		return "", "0", err
	}

	return lockID, votingPowerInt.String(), nil
}

// UnlockVotingTokens unlocks voting tokens after lock period
func (k *Keeper) UnlockVotingTokens(ctx context.Context, owner, lockID string) (string, error) {
	lock, err := k.GetVoteLock(ctx, lockID)
	if err != nil {
		return "0", err
	}

	if lock.Owner != owner {
		return "0", types.ErrUnauthorized
	}

	if lock.Withdrawn {
		return "0", types.ErrVoteLockAlreadyWithdrawn
	}

	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return "0", err
	}

	lockEnd := lock.LockEnd.Unix()
	if currentTime < lockEnd {
		return "0", types.ErrVoteLockNotExpired
	}

	// Mark as withdrawn
	lock.Withdrawn = true
	if err := k.SetVoteLock(ctx, lock); err != nil {
		return "0", err
	}

	return lock.Amount, nil
}

// GetVotingPower calculates total voting power for an address
func (k *Keeper) GetVotingPower(ctx context.Context, address string) (string, string, uint64, error) {
	lockIDs, err := k.GetUserVoteLockIndex(ctx, address)
	if err != nil {
		return "0", "0", 0, err
	}

	totalVotingPower := big.NewInt(0)
	totalLocked := big.NewInt(0)
	activeLocks := uint64(0)

	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return "0", "0", 0, err
	}

	for _, lockID := range lockIDs {
		lock, err := k.GetVoteLock(ctx, lockID)
		if err != nil || lock.Withdrawn {
			continue
		}

		// Check if lock is still active
		lockEnd := lock.LockEnd.Unix()
		if currentTime >= lockEnd {
			continue
		}

		activeLocks++

		votingPower := new(big.Int)
		votingPower.SetString(lock.VotingPower, 10)
		totalVotingPower.Add(totalVotingPower, votingPower)

		locked := new(big.Int)
		locked.SetString(lock.Amount, 10)
		totalLocked.Add(totalLocked, locked)
	}

	return totalVotingPower.String(), totalLocked.String(), activeLocks, nil
}

// GetVoteLockByID retrieves a vote lock by ID
func (k *Keeper) GetVoteLockByID(ctx context.Context, lockID string) (*types.VoteLock, error) {
	return k.GetVoteLock(ctx, lockID)
}

// GetUserVoteLocks returns all vote locks for a user
func (k *Keeper) GetUserVoteLocks(ctx context.Context, owner string) ([]*types.VoteLock, error) {
	lockIDs, err := k.GetUserVoteLockIndex(ctx, owner)
	if err != nil {
		return nil, err
	}

	locks := make([]*types.VoteLock, 0, len(lockIDs))

	for _, lockID := range lockIDs {
		lock, err := k.GetVoteLock(ctx, lockID)
		if err == nil {
			locks = append(locks, lock)
		}
	}

	return locks, nil
}

// GetActiveVoteLocks returns only active (not withdrawn, not expired) vote locks for a user
func (k *Keeper) GetActiveVoteLocks(ctx context.Context, owner string) ([]*types.VoteLock, error) {
	allLocks, err := k.GetUserVoteLocks(ctx, owner)
	if err != nil {
		return nil, err
	}

	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return nil, err
	}

	activeLocks := make([]*types.VoteLock, 0)
	for _, lock := range allLocks {
		if lock.Withdrawn {
			continue
		}

		lockEnd := lock.LockEnd.Unix()
		if currentTime >= lockEnd {
			continue
		}

		activeLocks = append(activeLocks, lock)
	}

	return activeLocks, nil
}

func generateLockID(owner, amount string, duration uint64, currentTime int64) string {
	h := sha256.New()
	h.Write([]byte(owner))
	h.Write([]byte(amount))
	h.Write([]byte(fmt.Sprintf("%d", duration)))
	h.Write([]byte(fmt.Sprintf("%d", currentTime)))
	return "vl:" + hex.EncodeToString(h.Sum(nil))[:32]
}

// sqrt calculates integer square root using Newton's method (deterministic integer math)
func sqrt(n *big.Int) *big.Int {
	if n.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}
	if n.Cmp(big.NewInt(1)) == 0 {
		return big.NewInt(1)
	}

	// Newton's method for integer square root
	// x_{n+1} = (x_n + n/x_n) / 2
	x := new(big.Int).Set(n)
	x.Div(x, big.NewInt(2)) // Start with n/2

	one := big.NewInt(1)
	for {
		// x1 = (x + n/x) / 2
		x1 := new(big.Int).Div(n, x)
		x1.Add(x1, x)
		x1.Div(x1, big.NewInt(2))

		// Check for convergence
		diff := new(big.Int).Sub(x, x1)
		diff.Abs(diff)
		if diff.Cmp(one) <= 0 {
			// Return the smaller of x and x1 to ensure floor(sqrt(n))
			if x.Cmp(x1) < 0 {
				return x
			}
			return x1
		}
		x.Set(x1)
	}
}

// GetTotalLockedGovernance returns total amount locked for governance
func (k *Keeper) GetTotalLockedGovernance(ctx context.Context) (string, error) {
	total := big.NewInt(0)

	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return "0", err
	}

	err = k.IterateVoteLocks(ctx, func(lock *types.VoteLock) bool {
		if lock.Withdrawn {
			return true
		}

		lockEnd := lock.LockEnd.Unix()
		if currentTime >= lockEnd {
			return true
		}

		amount := new(big.Int)
		amount.SetString(lock.Amount, 10)
		total.Add(total, amount)

		return true
	})

	if err != nil {
		return "0", err
	}

	return total.String(), nil
}

// CalculateTimeWeightedVotingPower calculates voting power with time weighting
// Longer locks get more voting power per token
// Uses integer math for deterministic consensus
func (k *Keeper) CalculateTimeWeightedVotingPower(
	ctx context.Context,
	amount string,
	lockDuration uint64,
) (string, error) {
	params, _ := k.GetParams(ctx)

	amt := new(big.Int)
	if _, ok := amt.SetString(amount, 10); !ok {
		return "0", types.ErrInvalidAmount
	}

	// Calculate time weight multiplier using integer math
	// multiplier = 1 + (duration / 1 year) * multiplier_per_year
	// Using basis points (10000 = 1.0) for deterministic consensus
	oneYear := uint64(31536000) // seconds in a year
	multiplierBasisPoints := params.Governance.LockMultiplierPerYear

	// bonusBps = (lockDuration * multiplierBasisPoints) / oneYear
	bonusBps := (lockDuration * multiplierBasisPoints) / oneYear
	// totalMultiplierBps = 10000 (1.0) + bonusBps
	totalMultiplierBps := types.BasisPoints + bonusBps

	// votingPower = amt * totalMultiplierBps / 10000
	result := new(big.Int).Mul(amt, big.NewInt(int64(totalMultiplierBps)))
	result.Div(result, big.NewInt(int64(types.BasisPoints)))

	return result.String(), nil
}

// EstimateVotingPowerGrowth estimates how voting power will grow over lock period
func (k *Keeper) EstimateVotingPowerGrowth(
	ctx context.Context,
	amount string,
	maxLockDuration uint64,
	steps uint64,
) (map[string]string, error) {
	estimates := make(map[string]string)

	stepSize := maxLockDuration / steps
	for i := uint64(0); i <= steps; i++ {
		duration := i * stepSize
		if duration > maxLockDuration {
			duration = maxLockDuration
		}

		power, err := k.CalculateTimeWeightedVotingPower(ctx, amount, duration)
		if err != nil {
			return nil, err
		}

		key := fmt.Sprintf("duration_%d_seconds", duration)
		estimates[key] = power
	}

	return estimates, nil
}

// ValidateGovernanceParameters validates governance-related parameters
func (k *Keeper) ValidateGovernanceParameters(ctx context.Context, params *types.GovernanceConfig) error {
	if params == nil {
		return errors.New("governance config cannot be nil")
	}

	// Validate min proposal stake
	minStake := new(big.Int)
	if _, ok := minStake.SetString(params.MinProposalStake, 10); !ok || minStake.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("invalid min proposal stake: %s", params.MinProposalStake)
	}

	// Validate lock durations
	if params.MinLockDuration == 0 {
		return errors.New("min lock duration must be greater than 0")
	}

	if params.MaxLockDuration <= params.MinLockDuration {
		return errors.New("max lock duration must be greater than min lock duration")
	}

	// Validate lock multiplier
	if params.LockMultiplierPerYear > 10000 { // Max 100% increase per year
		return errors.New("lock multiplier per year cannot exceed 10000 basis points (100%)")
	}

	return nil
}
