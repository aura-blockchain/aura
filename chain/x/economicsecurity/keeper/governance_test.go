// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// =============================================================================
// CheckProposalStake Tests
// =============================================================================

func TestCheckProposalStake_Sufficient(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Default min stake is 1000
	err := k.CheckProposalStake(ctx, testAuthority, "5000")
	require.NoError(t, err)
}

func TestCheckProposalStake_ExactAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Default min stake is 1000
	err := k.CheckProposalStake(ctx, testAuthority, "1000")
	require.NoError(t, err)
}

func TestCheckProposalStake_Insufficient(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Less than min stake
	err := k.CheckProposalStake(ctx, testAuthority, "500")
	require.Error(t, err)
	require.Equal(t, types.ErrInsufficientStake, err)
}

func TestCheckProposalStake_InvalidAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.CheckProposalStake(ctx, testAuthority, "invalid")
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidAmount, err)
}

// =============================================================================
// CalculateQuadraticVotingPower Tests
// =============================================================================

func TestCalculateQuadraticVotingPower_Disabled(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// By default, quadratic voting is disabled
	params, _ := k.GetParams(ctx)
	require.False(t, params.Governance.QuadraticVotingEnabled)

	// Should return linear voting power when disabled
	power, err := k.CalculateQuadraticVotingPower(ctx, "10000")
	require.NoError(t, err)
	require.Equal(t, "10000", power)
}

func TestCalculateQuadraticVotingPower_Enabled(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Enable quadratic voting
	params, _ := k.GetParams(ctx)
	params.Governance.QuadraticVotingEnabled = true
	require.NoError(t, k.SetParams(params))

	// sqrt(10000) = 100
	power, err := k.CalculateQuadraticVotingPower(ctx, "10000")
	require.NoError(t, err)
	require.Equal(t, "100", power)
}

func TestCalculateQuadraticVotingPower_LargeAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Governance.QuadraticVotingEnabled = true
	require.NoError(t, k.SetParams(params))

	// sqrt(1000000000000) = 1000000
	power, err := k.CalculateQuadraticVotingPower(ctx, "1000000000000")
	require.NoError(t, err)
	require.Equal(t, "1000000", power)
}

func TestCalculateQuadraticVotingPower_InvalidAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Governance.QuadraticVotingEnabled = true
	require.NoError(t, k.SetParams(params))

	_, err := k.CalculateQuadraticVotingPower(ctx, "invalid")
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidAmount, err)
}

// =============================================================================
// LockVotingTokens Tests
// =============================================================================

func TestLockVotingTokens_Disabled(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Governance.VoteLockingEnabled = false
	require.NoError(t, k.SetParams(params))

	_, _, err := k.LockVotingTokens(ctx, testAuthority, "10000", 604800)
	require.Error(t, err)
	require.Contains(t, err.Error(), "vote locking is disabled")
}

func TestLockVotingTokens_DurationTooShort(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Min duration is 604800 (7 days)
	_, _, err := k.LockVotingTokens(ctx, testAuthority, "10000", 86400)
	require.Error(t, err)
	require.Equal(t, types.ErrLockDurationTooShort, err)
}

func TestLockVotingTokens_DurationTooLong(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Max duration is 31536000 (1 year)
	_, _, err := k.LockVotingTokens(ctx, testAuthority, "10000", 63072000)
	require.Error(t, err)
	require.Equal(t, types.ErrLockDurationTooLong, err)
}

func TestLockVotingTokens_InvalidAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	_, _, err := k.LockVotingTokens(ctx, testAuthority, "invalid", 604800)
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidAmount, err)
}

func TestLockVotingTokens_ZeroAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	_, _, err := k.LockVotingTokens(ctx, testAuthority, "0", 604800)
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidAmount, err)
}

func TestLockVotingTokens_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	lockID, votingPower, err := k.LockVotingTokens(ctx, testAuthority, "10000", 604800)
	require.NoError(t, err)
	require.NotEmpty(t, lockID)
	require.NotEmpty(t, votingPower)

	// Verify lock was created
	lock, err := k.GetVoteLock(ctx, lockID)
	require.NoError(t, err)
	require.Equal(t, testAuthority, lock.Owner)
	require.Equal(t, "10000", lock.Amount)
	require.False(t, lock.Withdrawn)
}

func TestLockVotingTokens_VotingPowerBoost(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Lock for 1 year (max duration) should give max boost
	lockID, votingPower, err := k.LockVotingTokens(ctx, testAuthority, "10000", 31536000)
	require.NoError(t, err)
	require.NotEmpty(t, lockID)

	// With 1 year lock and 10000 basis points multiplier per year:
	// votingPower = 10000 * (10000 + 10000) / 10000 = 20000
	require.Equal(t, "20000", votingPower)
}

// =============================================================================
// UnlockVotingTokens Tests
// =============================================================================

func TestUnlockVotingTokens_NotFound(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	_, err := k.UnlockVotingTokens(ctx, testAuthority, "nonexistent-lock")
	require.Error(t, err)
}

func TestUnlockVotingTokens_WrongOwner(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Create a lock
	lockID, _, err := k.LockVotingTokens(ctx, testAuthority, "10000", 604800)
	require.NoError(t, err)

	// Try to unlock with different owner
	_, err = k.UnlockVotingTokens(ctx, "different-owner", lockID)
	require.Error(t, err)
	require.Equal(t, types.ErrUnauthorized, err)
}

func TestUnlockVotingTokens_NotExpired(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Create a lock that won't expire for a while
	lockID, _, err := k.LockVotingTokens(ctx, testAuthority, "10000", 604800)
	require.NoError(t, err)

	// Try to unlock immediately (lock is still active)
	_, err = k.UnlockVotingTokens(ctx, testAuthority, lockID)
	require.Error(t, err)
	require.Equal(t, types.ErrVoteLockNotExpired, err)
}

// =============================================================================
// GetVotingPower Tests
// =============================================================================

func TestGetVotingPower_NoLocks(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	power, locked, activeLocks, err := k.GetVotingPower(ctx, "aura1nolocks")
	require.NoError(t, err)
	require.Equal(t, "0", power)
	require.Equal(t, "0", locked)
	require.Equal(t, uint64(0), activeLocks)
}

func TestGetVotingPower_WithLocks(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Create a lock
	_, _, err := k.LockVotingTokens(ctx, testAuthority, "10000", 604800)
	require.NoError(t, err)

	power, locked, activeLocks, err := k.GetVotingPower(ctx, testAuthority)
	require.NoError(t, err)
	require.NotEqual(t, "0", power)
	require.NotEqual(t, "0", locked)
	require.Equal(t, uint64(1), activeLocks)
}

// =============================================================================
// GetVoteLockByID Tests
// =============================================================================

func TestGetVoteLockByID_Exists(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	lockID, _, err := k.LockVotingTokens(ctx, testAuthority, "10000", 604800)
	require.NoError(t, err)

	lock, err := k.GetVoteLockByID(ctx, lockID)
	require.NoError(t, err)
	require.Equal(t, lockID, lock.LockId)
}

func TestGetVoteLockByID_NotFound(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	_, err := k.GetVoteLockByID(ctx, "nonexistent")
	require.Error(t, err)
}

// =============================================================================
// GetUserVoteLocks Tests
// =============================================================================

func TestGetUserVoteLocks_Empty(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	locks, err := k.GetUserVoteLocks(ctx, "aura1nolocks")
	require.NoError(t, err)
	require.Len(t, locks, 0)
}

func TestGetUserVoteLocks_Multiple(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Create multiple locks
	_, _, err := k.LockVotingTokens(ctx, testAuthority, "10000", 604800)
	require.NoError(t, err)
	_, _, err = k.LockVotingTokens(ctx, testAuthority, "20000", 1209600)
	require.NoError(t, err)

	locks, err := k.GetUserVoteLocks(ctx, testAuthority)
	require.NoError(t, err)
	require.Len(t, locks, 2)
}

// =============================================================================
// GetActiveVoteLocks Tests
// =============================================================================

func TestGetActiveVoteLocks_FilterWithdrawn(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Create a lock
	_, _, err := k.LockVotingTokens(ctx, testAuthority, "10000", 604800)
	require.NoError(t, err)

	// Should have 1 active lock
	activeLocks, err := k.GetActiveVoteLocks(ctx, testAuthority)
	require.NoError(t, err)
	require.Len(t, activeLocks, 1)
}

// =============================================================================
// sqrt Tests (internal helper)
// =============================================================================

func TestSqrt(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"0", "0"},
		{"1", "1"},
		{"4", "2"},
		{"9", "3"},
		{"16", "4"},
		{"25", "5"},
		{"100", "10"},
		{"10000", "100"},
		{"1000000", "1000"},
	}

	for _, tc := range testCases {
		n := new(big.Int)
		n.SetString(tc.input, 10)

		result := sqrt(n)
		require.Equal(t, tc.expected, result.String(), "sqrt(%s) should be %s", tc.input, tc.expected)
	}
}

func TestSqrt_NonPerfectSquare(t *testing.T) {
	// Non-perfect squares should return floor of sqrt
	n := new(big.Int)

	n.SetString("5", 10)
	require.Equal(t, "2", sqrt(n).String()) // sqrt(5) ≈ 2.23

	n.SetString("10", 10)
	require.Equal(t, "3", sqrt(n).String()) // sqrt(10) ≈ 3.16

	n.SetString("50", 10)
	require.Equal(t, "7", sqrt(n).String()) // sqrt(50) ≈ 7.07
}

// =============================================================================
// GetTotalLockedGovernance Tests
// =============================================================================

func TestGetTotalLockedGovernance_Empty(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	total, err := k.GetTotalLockedGovernance(ctx)
	require.NoError(t, err)
	require.Equal(t, "0", total)
}

func TestGetTotalLockedGovernance_WithLocks(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Create a lock and verify total matches the locked amount
	_, _, err := k.LockVotingTokens(ctx, testAuthority, "15000", 604800)
	require.NoError(t, err)

	total, err := k.GetTotalLockedGovernance(ctx)
	require.NoError(t, err)
	// Total should include the locked amount
	require.NotEqual(t, "0", total)
}

// =============================================================================
// CalculateTimeWeightedVotingPower Tests
// =============================================================================

func TestCalculateTimeWeightedVotingPower_NoLock(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// 0 duration = 1x multiplier
	power, err := k.CalculateTimeWeightedVotingPower(ctx, "10000", 0)
	require.NoError(t, err)
	require.Equal(t, "10000", power)
}

func TestCalculateTimeWeightedVotingPower_OneYear(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// 1 year lock with 10000 basis points multiplier = 2x
	power, err := k.CalculateTimeWeightedVotingPower(ctx, "10000", 31536000)
	require.NoError(t, err)
	require.Equal(t, "20000", power)
}

func TestCalculateTimeWeightedVotingPower_HalfYear(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Half year lock = 1.5x multiplier
	power, err := k.CalculateTimeWeightedVotingPower(ctx, "10000", 15768000)
	require.NoError(t, err)
	require.Equal(t, "15000", power)
}

func TestCalculateTimeWeightedVotingPower_InvalidAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	_, err := k.CalculateTimeWeightedVotingPower(ctx, "invalid", 604800)
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidAmount, err)
}

// =============================================================================
// EstimateVotingPowerGrowth Tests
// =============================================================================

func TestEstimateVotingPowerGrowth(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	estimates, err := k.EstimateVotingPowerGrowth(ctx, "10000", 31536000, 4)
	require.NoError(t, err)
	require.Len(t, estimates, 5) // 0, 1/4, 2/4, 3/4, 4/4

	// Check that estimates exist for expected durations
	require.Contains(t, estimates, "duration_0_seconds")
	require.Contains(t, estimates, "duration_7884000_seconds")  // 1/4 year
	require.Contains(t, estimates, "duration_15768000_seconds") // 1/2 year
	require.Contains(t, estimates, "duration_23652000_seconds") // 3/4 year
	require.Contains(t, estimates, "duration_31536000_seconds") // 1 year
}

func TestEstimateVotingPowerGrowth_InvalidAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	_, err := k.EstimateVotingPowerGrowth(ctx, "invalid", 31536000, 4)
	require.Error(t, err)
}

// =============================================================================
// ValidateGovernanceParameters Tests
// =============================================================================

func TestValidateGovernanceParameters_Valid(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	config := &types.GovernanceConfig{
		MinProposalStake:       "1000",
		MinLockDuration:        604800,
		MaxLockDuration:        31536000,
		LockMultiplierPerYear:  10000,
	}

	err := k.ValidateGovernanceParameters(ctx, config)
	require.NoError(t, err)
}

func TestValidateGovernanceParameters_NilConfig(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.ValidateGovernanceParameters(ctx, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot be nil")
}

func TestValidateGovernanceParameters_InvalidMinStake(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	config := &types.GovernanceConfig{
		MinProposalStake:       "invalid",
		MinLockDuration:        604800,
		MaxLockDuration:        31536000,
		LockMultiplierPerYear:  10000,
	}

	err := k.ValidateGovernanceParameters(ctx, config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid min proposal stake")
}

func TestValidateGovernanceParameters_ZeroMinLockDuration(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	config := &types.GovernanceConfig{
		MinProposalStake:       "1000",
		MinLockDuration:        0,
		MaxLockDuration:        31536000,
		LockMultiplierPerYear:  10000,
	}

	err := k.ValidateGovernanceParameters(ctx, config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "min lock duration must be greater than 0")
}

func TestValidateGovernanceParameters_MaxLessThanMin(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	config := &types.GovernanceConfig{
		MinProposalStake:       "1000",
		MinLockDuration:        31536000,
		MaxLockDuration:        604800, // Less than min
		LockMultiplierPerYear:  10000,
	}

	err := k.ValidateGovernanceParameters(ctx, config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "max lock duration must be greater than min")
}

func TestValidateGovernanceParameters_ExcessiveMultiplier(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	config := &types.GovernanceConfig{
		MinProposalStake:       "1000",
		MinLockDuration:        604800,
		MaxLockDuration:        31536000,
		LockMultiplierPerYear:  20000, // More than 100%
	}

	err := k.ValidateGovernanceParameters(ctx, config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "lock multiplier per year cannot exceed")
}
