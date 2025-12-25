// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"github.com/aequitas/aura/chain/testing/testutil"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dex/keeper"
	"github.com/aequitas/aura/chain/x/dex/types"
)

// ============================================================================
// MOCK VC REGISTRY KEEPER FOR IR VERIFICATION TESTING
// ============================================================================

// MockVCRegistryKeeper implements the VCRegistryKeeper interface for testing
type MockVCRegistryKeeper struct {
	// Map of address -> IR score
	irScores map[string]uint64
	// Map of address -> verified status
	verified map[string]bool
}

func NewMockVCRegistryKeeper() *MockVCRegistryKeeper {
	return &MockVCRegistryKeeper{
		irScores: make(map[string]uint64),
		verified: make(map[string]bool),
	}
}

func (m *MockVCRegistryKeeper) GetIRScore(ctx sdk.Context, address string) uint64 {
	if score, ok := m.irScores[address]; ok {
		return score
	}
	return 0
}

func (m *MockVCRegistryKeeper) IsVerified(ctx sdk.Context, address string) bool {
	if verified, ok := m.verified[address]; ok {
		return verified
	}
	return false
}

// SetIRScore sets the IR score for an address (test helper)
func (m *MockVCRegistryKeeper) SetIRScore(address string, score uint64) {
	m.irScores[address] = score
	// Auto-mark as verified if score >= 100
	m.verified[address] = score >= 100
}

// ============================================================================
// TEST SETUP HELPERS
// ============================================================================

// setupTestWithVCKeeper creates a test keeper with mock VC registry keeper
func setupTestWithVCKeeper(t *testing.T) (*keeper.Keeper, sdk.Context, *MockVCRegistryKeeper) {
	input := keepertest.CreateTestInput(t)
	vcKeeper := NewMockVCRegistryKeeper()

	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		testutil.NewMockBankKeeper(),
		testutil.NewMockAccountKeeper(),
		vcKeeper,
		testutil.NewMockSecurityKeeper(),
	)

	return k, input.Ctx, vcKeeper
}

// ============================================================================
// IsUserVerified TESTS
// ============================================================================

func TestIsUserVerified_WithExactly100IRPoints(t *testing.T) {
	k, ctx, vcKeeper := setupTestWithVCKeeper(t)

	addr := keepertest.GenTestAddr()
	address := addr.String()

	// Set exactly 100 IR points (threshold)
	vcKeeper.SetIRScore(address, 100)

	verified := k.IsUserVerified(ctx, address)
	require.True(t, verified, "user with exactly 100 IR points should be verified")
}

func TestIsUserVerified_WithMoreThan100IRPoints(t *testing.T) {
	k, ctx, vcKeeper := setupTestWithVCKeeper(t)

	addr := keepertest.GenTestAddr()
	address := addr.String()

	// Set 150 IR points (above threshold)
	vcKeeper.SetIRScore(address, 150)

	verified := k.IsUserVerified(ctx, address)
	require.True(t, verified, "user with >100 IR points should be verified")
}

func TestIsUserVerified_WithLessThan100IRPoints(t *testing.T) {
	k, ctx, vcKeeper := setupTestWithVCKeeper(t)

	addr := keepertest.GenTestAddr()
	address := addr.String()

	// Set 99 IR points (below threshold)
	vcKeeper.SetIRScore(address, 99)

	verified := k.IsUserVerified(ctx, address)
	require.False(t, verified, "user with <100 IR points should not be verified")
}

func TestIsUserVerified_WithZeroIRPoints(t *testing.T) {
	k, ctx, _ := setupTestWithVCKeeper(t)

	addr := keepertest.GenTestAddr()
	address := addr.String()

	// No IR points set (defaults to 0)
	verified := k.IsUserVerified(ctx, address)
	require.False(t, verified, "user with 0 IR points should not be verified")
}

func TestIsUserVerified_WithMultipleUsers(t *testing.T) {
	k, ctx, vcKeeper := setupTestWithVCKeeper(t)

	// Create multiple test addresses with unique seeds
	addrs := keepertest.GenTestAddrs(3)
	addr1 := addrs[0]
	addr2 := addrs[1]
	addr3 := addrs[2]

	// Set different IR scores
	vcKeeper.SetIRScore(addr1.String(), 50)   // Not verified
	vcKeeper.SetIRScore(addr2.String(), 100)  // Verified
	vcKeeper.SetIRScore(addr3.String(), 200)  // Verified

	// Verify scores were set correctly
	score1 := vcKeeper.GetIRScore(ctx, addr1.String())
	score2 := vcKeeper.GetIRScore(ctx, addr2.String())
	score3 := vcKeeper.GetIRScore(ctx, addr3.String())

	require.Equal(t, uint64(50), score1, "addr1 should have IR score 50")
	require.Equal(t, uint64(100), score2, "addr2 should have IR score 100")
	require.Equal(t, uint64(200), score3, "addr3 should have IR score 200")

	// Now test verification status
	verified1 := k.IsUserVerified(ctx, addr1.String())
	verified2 := k.IsUserVerified(ctx, addr2.String())
	verified3 := k.IsUserVerified(ctx, addr3.String())

	require.False(t, verified1, "addr1 (score 50) should not be verified, but got %v", verified1)
	require.True(t, verified2, "addr2 (score 100) should be verified, but got %v", verified2)
	require.True(t, verified3, "addr3 (score 200) should be verified, but got %v", verified3)
}

// ============================================================================
// CalculateFeeBoost TESTS
// ============================================================================

func TestCalculateFeeBoost_VerifiedUserWith40Percent(t *testing.T) {
	k, ctx, vcKeeper := setupTestWithVCKeeper(t)

	addr := keepertest.GenTestAddr()
	address := addr.String()

	// Set user as verified
	vcKeeper.SetIRScore(address, 100)

	// Set IR boost params
	params := types.DefaultParams()
	params.IrBoostEnabled = true
	params.IrBoostPercent = 40
	err := k.SetParams(ctx, &params)
	require.NoError(t, err)

	boost := k.CalculateFeeBoost(ctx, address)

	// Should return 0.40 (40%)
	expected := sdkmath.LegacyNewDecWithPrec(40, 2) // 0.40
	require.Equal(t, expected.String(), boost.String(), "boost should be 40% = 0.40")
}

func TestCalculateFeeBoost_UnverifiedUser(t *testing.T) {
	k, ctx, vcKeeper := setupTestWithVCKeeper(t)

	addr := keepertest.GenTestAddr()
	address := addr.String()

	// User has insufficient IR points
	vcKeeper.SetIRScore(address, 50)

	// Set IR boost params
	params := types.DefaultParams()
	params.IrBoostEnabled = true
	params.IrBoostPercent = 40
	err := k.SetParams(ctx, &params)
	require.NoError(t, err)

	boost := k.CalculateFeeBoost(ctx, address)

	// Should return 0.00 (no boost for unverified)
	require.True(t, boost.IsZero(), "unverified user should get 0 boost")
}

func TestCalculateFeeBoost_BoostDisabled(t *testing.T) {
	k, ctx, vcKeeper := setupTestWithVCKeeper(t)

	addr := keepertest.GenTestAddr()
	address := addr.String()

	// Set user as verified
	vcKeeper.SetIRScore(address, 100)

	// Disable IR boost
	params := types.DefaultParams()
	params.IrBoostEnabled = false
	params.IrBoostPercent = 40
	err := k.SetParams(ctx, &params)
	require.NoError(t, err)

	boost := k.CalculateFeeBoost(ctx, address)

	// Should return 0.00 (boost disabled)
	require.True(t, boost.IsZero(), "disabled boost should return 0")
}

func TestCalculateFeeBoost_InvalidBoostPercentageNegative(t *testing.T) {
	k, ctx, mockVC := setupTestWithVCKeeper(t)

	addr := keepertest.GenTestAddr()
	address := addr.String()

	// Set user as verified
	mockVC.SetIRScore(address, 100)

	// Set invalid negative boost (should be rejected by validation)
	params := types.DefaultParams()
	params.IrBoostEnabled = true
	params.IrBoostPercent = ^uint64(0) // Max uint64 (wraps to huge number in int64)
	err := k.SetParams(ctx, &params)
	require.NoError(t, err)

	boost := k.CalculateFeeBoost(ctx, address)

	// Should return 0.00 (invalid params rejected)
	require.True(t, boost.IsZero(), "invalid boost percentage should return 0")
}

func TestCalculateFeeBoost_InvalidBoostPercentageOver100(t *testing.T) {
	k, ctx, vcKeeper := setupTestWithVCKeeper(t)

	addr := keepertest.GenTestAddr()
	address := addr.String()

	// Set user as verified
	vcKeeper.SetIRScore(address, 100)

	// Set invalid >100% boost
	params := types.DefaultParams()
	params.IrBoostEnabled = true
	params.IrBoostPercent = 150 // 150% is excessive
	err := k.SetParams(ctx, &params)
	require.NoError(t, err)

	boost := k.CalculateFeeBoost(ctx, address)

	// Should return 0.00 (invalid params rejected)
	require.True(t, boost.IsZero(), "boost >100% should be rejected")
}

func TestCalculateFeeBoost_ValidBoostPercentages(t *testing.T) {
	tests := []struct {
		name           string
		boostPercent   uint64
		expectedBoost  string
	}{
		{"0% boost", 0, "0.000000000000000000"},
		{"10% boost", 10, "0.100000000000000000"},
		{"25% boost", 25, "0.250000000000000000"},
		{"40% boost", 40, "0.400000000000000000"},
		{"50% boost", 50, "0.500000000000000000"},
		{"75% boost", 75, "0.750000000000000000"},
		{"100% boost", 100, "1.000000000000000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx, vcKeeper := setupTestWithVCKeeper(t)

			addr := keepertest.GenTestAddr()
			address := addr.String()

			// Set user as verified
			vcKeeper.SetIRScore(address, 100)

			// Set boost params
			params := types.DefaultParams()
			params.IrBoostEnabled = true
			params.IrBoostPercent = tt.boostPercent
			err := k.SetParams(ctx, &params)
			require.NoError(t, err)

			boost := k.CalculateFeeBoost(ctx, address)
			require.Equal(t, tt.expectedBoost, boost.String())
		})
	}
}

// ============================================================================
// CalculateEffectiveFee TESTS
// ============================================================================

func TestCalculateEffectiveFee_VerifiedUserWith40PercentBoost(t *testing.T) {
	k, ctx, vcKeeper := setupTestWithVCKeeper(t)

	addr := keepertest.GenTestAddr()
	address := addr.String()

	// Set user as verified
	vcKeeper.SetIRScore(address, 100)

	// Set IR boost params
	params := types.DefaultParams()
	params.IrBoostEnabled = true
	params.IrBoostPercent = 40
	err := k.SetParams(ctx, &params)
	require.NoError(t, err)

	// Base fee: 0.003 (0.3%)
	baseFee := sdkmath.LegacyNewDecWithPrec(3, 3)

	effectiveFee := k.CalculateEffectiveFee(ctx, address, baseFee)

	// Effective fee = 0.003 * (1 + 0.40) = 0.003 * 1.40 = 0.0042
	expected := sdkmath.LegacyNewDecWithPrec(42, 4) // 0.0042
	require.Equal(t, expected.String(), effectiveFee.String())
}

func TestCalculateEffectiveFee_UnverifiedUser(t *testing.T) {
	k, ctx, vcKeeper := setupTestWithVCKeeper(t)

	addr := keepertest.GenTestAddr()
	address := addr.String()

	// User has insufficient IR points
	vcKeeper.SetIRScore(address, 50)

	// Set IR boost params
	params := types.DefaultParams()
	params.IrBoostEnabled = true
	params.IrBoostPercent = 40
	err := k.SetParams(ctx, &params)
	require.NoError(t, err)

	// Base fee: 0.003 (0.3%)
	baseFee := sdkmath.LegacyNewDecWithPrec(3, 3)

	effectiveFee := k.CalculateEffectiveFee(ctx, address, baseFee)

	// Effective fee = 0.003 * (1 + 0.00) = 0.003 (no boost)
	require.Equal(t, baseFee.String(), effectiveFee.String())
}

func TestCalculateEffectiveFee_NegativeBaseFeeRejected(t *testing.T) {
	k, ctx, vcKeeper := setupTestWithVCKeeper(t)

	addr := keepertest.GenTestAddr()
	address := addr.String()

	// Set user as verified
	vcKeeper.SetIRScore(address, 100)

	// Negative base fee (invalid)
	baseFee := sdkmath.LegacyNewDec(-1)

	effectiveFee := k.CalculateEffectiveFee(ctx, address, baseFee)

	// Should return 0 (negative fees rejected)
	require.True(t, effectiveFee.IsZero(), "negative base fee should return 0")
}

func TestCalculateEffectiveFee_ZeroBaseFee(t *testing.T) {
	k, ctx, vcKeeper := setupTestWithVCKeeper(t)

	addr := keepertest.GenTestAddr()
	address := addr.String()

	// Set user as verified
	vcKeeper.SetIRScore(address, 100)

	// Zero base fee
	baseFee := sdkmath.LegacyZeroDec()

	effectiveFee := k.CalculateEffectiveFee(ctx, address, baseFee)

	// Should return 0 (0 * 1.40 = 0)
	require.True(t, effectiveFee.IsZero(), "zero base fee should return 0")
}

func TestCalculateEffectiveFee_LargeBaseFee(t *testing.T) {
	k, ctx, vcKeeper := setupTestWithVCKeeper(t)

	addr := keepertest.GenTestAddr()
	address := addr.String()

	// Set user as verified
	vcKeeper.SetIRScore(address, 100)

	// Set IR boost params
	params := types.DefaultParams()
	params.IrBoostEnabled = true
	params.IrBoostPercent = 40
	err := k.SetParams(ctx, &params)
	require.NoError(t, err)

	// Large base fee: 1000000
	baseFee := sdkmath.LegacyNewDec(1_000_000)

	effectiveFee := k.CalculateEffectiveFee(ctx, address, baseFee)

	// Effective fee = 1000000 * 1.40 = 1400000
	expected := sdkmath.LegacyNewDec(1_400_000)
	require.Equal(t, expected.String(), effectiveFee.String())

	// Must not be negative
	require.False(t, effectiveFee.IsNegative())
}

func TestCalculateEffectiveFee_RealisticScenarios(t *testing.T) {
	scenarios := []struct {
		name         string
		baseFee      string
		boostPercent uint64
		verified     bool
		expectedFee  string
	}{
		{
			name:         "0.3% base, 40% boost, verified",
			baseFee:      "0.003",
			boostPercent: 40,
			verified:     true,
			expectedFee:  "0.0042", // 0.003 * 1.40
		},
		{
			name:         "0.3% base, 40% boost, unverified",
			baseFee:      "0.003",
			boostPercent: 40,
			verified:     false,
			expectedFee:  "0.003", // No boost
		},
		{
			name:         "0.5% base, 50% boost, verified",
			baseFee:      "0.005",
			boostPercent: 50,
			verified:     true,
			expectedFee:  "0.0075", // 0.005 * 1.50
		},
		{
			name:         "1% base, 25% boost, verified",
			baseFee:      "0.01",
			boostPercent: 25,
			verified:     true,
			expectedFee:  "0.0125", // 0.01 * 1.25
		},
		{
			name:         "0.1% base, 100% boost, verified",
			baseFee:      "0.001",
			boostPercent: 100,
			verified:     true,
			expectedFee:  "0.002", // 0.001 * 2.00
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			k, ctx, vcKeeper := setupTestWithVCKeeper(t)

			addr := keepertest.GenTestAddr()
			address := addr.String()

			// Set verification status
			if scenario.verified {
				vcKeeper.SetIRScore(address, 100)
			} else {
				vcKeeper.SetIRScore(address, 50)
			}

			// Set boost params
			params := types.DefaultParams()
			params.IrBoostEnabled = true
			params.IrBoostPercent = scenario.boostPercent
			err := k.SetParams(ctx, &params)
			require.NoError(t, err)

			// Parse base fee
			baseFee, err := sdkmath.LegacyNewDecFromStr(scenario.baseFee)
			require.NoError(t, err)

			effectiveFee := k.CalculateEffectiveFee(ctx, address, baseFee)

			// Parse expected fee
			expectedFee, err := sdkmath.LegacyNewDecFromStr(scenario.expectedFee)
			require.NoError(t, err)

			// Allow small rounding differences
			diff := effectiveFee.Sub(expectedFee).Abs()
			maxDiff := sdkmath.LegacyNewDecWithPrec(1, 10) // 1e-10 tolerance

			require.True(t, diff.LTE(maxDiff),
				"expected %s, got %s (diff: %s)",
				expectedFee.String(), effectiveFee.String(), diff.String())
		})
	}
}

func TestCalculateEffectiveFee_NoOverflow(t *testing.T) {
	k, ctx, vcKeeper := setupTestWithVCKeeper(t)

	addr := keepertest.GenTestAddr()
	address := addr.String()

	// Set user as verified with max boost
	vcKeeper.SetIRScore(address, 100)

	params := types.DefaultParams()
	params.IrBoostEnabled = true
	params.IrBoostPercent = 100 // 100% boost (2x multiplier)
	err := k.SetParams(ctx, &params)
	require.NoError(t, err)

	// Very large base fee (but should not overflow with 2x multiplier)
	baseFee := sdkmath.LegacyNewDec(1_000_000_000_000)

	effectiveFee := k.CalculateEffectiveFee(ctx, address, baseFee)

	// Should complete without panic
	require.False(t, effectiveFee.IsNegative(), "effective fee should not be negative")
	require.True(t, effectiveFee.GTE(baseFee), "boosted fee should be >= base fee")

	// Effective fee = 1000000000000 * 2.00 = 2000000000000
	expected := sdkmath.LegacyNewDec(2_000_000_000_000)
	require.Equal(t, expected.String(), effectiveFee.String())
}

// ============================================================================
// DeletePool TESTS
// ============================================================================

func TestDeletePool_Success(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	// Create a pool first
	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uusdt", sdkmath.NewInt(1000000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(1000000)))
	require.NoError(t, err)
	require.NotNil(t, pool)

	poolID := pool.PoolId

	// Verify pool exists
	retrievedPool := k.GetPool(ctx, poolID)
	require.NotNil(t, retrievedPool)

	// Delete the pool
	k.DeletePool(ctx, poolID)

	// Verify pool is deleted
	deletedPool := k.GetPool(ctx, poolID)
	require.Nil(t, deletedPool, "pool should be deleted")
}

func TestDeletePool_NonExistentPool(t *testing.T) {
	k, ctx, _ := setupTestKeeper(t)

	// Try to delete a pool that doesn't exist
	// This should not panic or error - DeletePool is idempotent
	k.DeletePool(ctx, "nonexistent-pool")

	// Verify still no pool
	pool := k.GetPool(ctx, "nonexistent-pool")
	require.Nil(t, pool)
}

func TestDeletePool_DeleteTwice(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	// Create a pool
	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uusdt", sdkmath.NewInt(1000000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(1000000)))
	require.NoError(t, err)

	poolID := pool.PoolId

	// Delete the pool
	k.DeletePool(ctx, poolID)

	// Delete again (should be idempotent)
	k.DeletePool(ctx, poolID)

	// Verify still deleted
	deletedPool := k.GetPool(ctx, poolID)
	require.Nil(t, deletedPool)
}

func TestDeletePool_MultiplePoolsIndependence(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	// Create multiple pools
	creatorAddr := keepertest.GenTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(3000000))
	mockBank.SetBalance(creatorAddr, "uusdt", sdkmath.NewInt(1000000))
	mockBank.SetBalance(creatorAddr, "uatom", sdkmath.NewInt(1000000))

	pool1, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(1000000)))
	require.NoError(t, err)

	// Advance time for pool creation cooldown
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(3601 * 1000000000)) // 1 hour + 1 second in nanoseconds

	pool2, _, err := k.CreatePool(ctx, creator, "uaura", "uatom",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("uatom", sdkmath.NewInt(1000000)))
	require.NoError(t, err)

	// Delete pool1
	k.DeletePool(ctx, pool1.PoolId)

	// Verify pool1 is deleted
	deletedPool := k.GetPool(ctx, pool1.PoolId)
	require.Nil(t, deletedPool)

	// Verify pool2 still exists
	remainingPool := k.GetPool(ctx, pool2.PoolId)
	require.NotNil(t, remainingPool)
	require.Equal(t, pool2.PoolId, remainingPool.PoolId)
}
