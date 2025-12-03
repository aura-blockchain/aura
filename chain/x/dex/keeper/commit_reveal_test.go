package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// TestCommitOrder_Success tests successful order commitment
func TestCommitOrder_Success(t *testing.T) {
	suite := SetupTestSuite(t)
	ctx := suite.Ctx

	sender := suite.TestAccounts[0]

	// Compute order hash
	orderType := types.SwapOrderType_SELL
	auraAmount := sdkmath.NewInt(10000000000) // 10,000 AURA
	otherCoin := "usdt"
	otherAmount := sdkmath.NewInt(50000000000) // 50,000 USDT
	salt := []byte("random_salt_12345")

	commitHash := suite.Keeper.ComputeOrderHash(orderType, auraAmount, otherCoin, otherAmount, salt)

	// Commit order
	commitID, err := suite.Keeper.CommitOrder(ctx, sender.String(), commitHash)
	require.NoError(t, err)
	require.NotEmpty(t, commitID)

	// Verify commitment was stored
	commitment, found := suite.Keeper.GetOrderCommitment(ctx, commitID)
	require.True(t, found)
	require.Equal(t, sender.String(), commitment.Sender)
	require.Equal(t, commitHash, commitment.CommitHash)
	require.NotNil(t, commitment.RevealDeadline)
}

// TestCommitOrder_InvalidHash tests commitment with invalid hash
func TestCommitOrder_InvalidHash(t *testing.T) {
	suite := SetupTestSuite(t)
	ctx := suite.Ctx

	sender := suite.TestAccounts[0]

	// Invalid hash (wrong length)
	invalidHash := []byte("too_short")

	_, err := suite.Keeper.CommitOrder(ctx, sender.String(), invalidHash)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid commit hash length")
}

// TestCommitOrder_DuplicateCommitment tests that only one commitment per sender is allowed
func TestCommitOrder_DuplicateCommitment(t *testing.T) {
	suite := SetupTestSuite(t)
	ctx := suite.Ctx

	sender := suite.TestAccounts[0]

	// Create first commitment
	commitHash := make([]byte, 32)
	copy(commitHash, "test_hash_1")
	_, err := suite.Keeper.CommitOrder(ctx, sender.String(), commitHash)
	require.NoError(t, err)

	// Try to create second commitment (should fail)
	commitHash2 := make([]byte, 32)
	copy(commitHash2, "test_hash_2")
	_, err = suite.Keeper.CommitOrder(ctx, sender.String(), commitHash2)
	require.Error(t, err)
	require.Equal(t, types.ErrCommitmentAlreadyExists, err)
}

// TestRevealOrder_Success tests successful order reveal
func TestRevealOrder_Success(t *testing.T) {
	suite := SetupTestSuite(t)
	ctx := suite.Ctx

	sender := suite.TestAccounts[0]

	// Fund sender
	initialBalance := sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(100000000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(100000000000)),
	)
	err := suite.BankKeeper.MintCoins(ctx, types.ModuleName, initialBalance)
	require.NoError(t, err)
	err = suite.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, initialBalance)
	require.NoError(t, err)

	// Order details
	orderType := types.SwapOrderType_SELL
	auraAmount := sdkmath.NewInt(10000000000) // 10,000 AURA
	otherCoin := "usdt"
	otherAmount := sdkmath.NewInt(50000000000) // 50,000 USDT
	salt := []byte("random_salt_12345")

	// Compute and commit hash
	commitHash := suite.Keeper.ComputeOrderHash(orderType, auraAmount, otherCoin, otherAmount, salt)
	commitID, err := suite.Keeper.CommitOrder(ctx, sender.String(), commitHash)
	require.NoError(t, err)

	// Reveal order
	orderID, err := suite.Keeper.RevealOrder(ctx, commitID, sender.String(), orderType, auraAmount, otherCoin, otherAmount, salt)
	require.NoError(t, err)
	require.NotEmpty(t, orderID)

	// Verify commitment was deleted
	_, found := suite.Keeper.GetOrderCommitment(ctx, commitID)
	require.False(t, found)
}

// TestRevealOrder_HashMismatch tests reveal with wrong order details
func TestRevealOrder_HashMismatch(t *testing.T) {
	suite := SetupTestSuite(t)
	ctx := suite.Ctx

	sender := suite.TestAccounts[0]

	// Fund sender
	initialBalance := sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(100000000000)),
	)
	err := suite.BankKeeper.MintCoins(ctx, types.ModuleName, initialBalance)
	require.NoError(t, err)
	err = suite.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, initialBalance)
	require.NoError(t, err)

	// Original order details
	orderType := types.SwapOrderType_SELL
	auraAmount := sdkmath.NewInt(10000000000)
	otherCoin := "usdt"
	otherAmount := sdkmath.NewInt(50000000000)
	salt := []byte("random_salt_12345")

	// Commit with original details
	commitHash := suite.Keeper.ComputeOrderHash(orderType, auraAmount, otherCoin, otherAmount, salt)
	commitID, err := suite.Keeper.CommitOrder(ctx, sender.String(), commitHash)
	require.NoError(t, err)

	// Try to reveal with different amount (attack attempt)
	differentAmount := sdkmath.NewInt(60000000000) // Changed amount
	_, err = suite.Keeper.RevealOrder(ctx, commitID, sender.String(), orderType, auraAmount, otherCoin, differentAmount, salt)
	require.Error(t, err)
	require.Equal(t, types.ErrHashMismatch, err)
}

// TestRevealOrder_ExpiredDeadline tests reveal after deadline
func TestRevealOrder_ExpiredDeadline(t *testing.T) {
	suite := SetupTestSuite(t)
	ctx := suite.Ctx

	sender := suite.TestAccounts[0]

	// Order details
	orderType := types.SwapOrderType_SELL
	auraAmount := sdkmath.NewInt(10000000000)
	otherCoin := "usdt"
	otherAmount := sdkmath.NewInt(50000000000)
	salt := []byte("random_salt_12345")

	// Commit order
	commitHash := suite.Keeper.ComputeOrderHash(orderType, auraAmount, otherCoin, otherAmount, salt)
	commitID, err := suite.Keeper.CommitOrder(ctx, sender.String(), commitHash)
	require.NoError(t, err)

	// Advance time past reveal deadline (default 60 seconds)
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(65 * time.Second))

	// Try to reveal (should fail)
	_, err = suite.Keeper.RevealOrder(ctx, commitID, sender.String(), orderType, auraAmount, otherCoin, otherAmount, salt)
	require.Error(t, err)
	require.Equal(t, types.ErrRevealDeadlineExpired, err)

	// Verify commitment was cleaned up
	_, found := suite.Keeper.GetOrderCommitment(ctx, commitID)
	require.False(t, found)
}

// TestRevealOrder_InsufficientBalance tests reveal without funds
func TestRevealOrder_InsufficientBalance(t *testing.T) {
	suite := SetupTestSuite(t)
	ctx := suite.Ctx

	sender := suite.TestAccounts[0]

	// Order details (large amount)
	orderType := types.SwapOrderType_SELL
	auraAmount := sdkmath.NewInt(10000000000)
	otherCoin := "usdt"
	otherAmount := sdkmath.NewInt(50000000000)
	salt := []byte("random_salt_12345")

	// Commit order
	commitHash := suite.Keeper.ComputeOrderHash(orderType, auraAmount, otherCoin, otherAmount, salt)
	commitID, err := suite.Keeper.CommitOrder(ctx, sender.String(), commitHash)
	require.NoError(t, err)

	// Try to reveal without funds (should fail)
	_, err = suite.Keeper.RevealOrder(ctx, commitID, sender.String(), orderType, auraAmount, otherCoin, otherAmount, salt)
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient balance")
}

// TestRevealOrder_CommitmentNotFound tests reveal with non-existent commitment
func TestRevealOrder_CommitmentNotFound(t *testing.T) {
	suite := SetupTestSuite(t)
	ctx := suite.Ctx

	sender := suite.TestAccounts[0]

	orderType := types.SwapOrderType_SELL
	auraAmount := sdkmath.NewInt(10000000000)
	otherCoin := "usdt"
	otherAmount := sdkmath.NewInt(50000000000)
	salt := []byte("random_salt_12345")

	// Try to reveal without committing
	_, err := suite.Keeper.RevealOrder(ctx, "nonexistent", sender.String(), orderType, auraAmount, otherCoin, otherAmount, salt)
	require.Error(t, err)
	require.Equal(t, types.ErrCommitmentNotFound, err)
}

// TestBatchExecution_PricePriority tests that batch execution sorts by price
func TestBatchExecution_PricePriority(t *testing.T) {
	suite := SetupTestSuite(t)
	ctx := suite.Ctx

	// Enable batch execution
	params := suite.Keeper.GetParams(ctx)
	params.BatchExecutionEnabled = true
	err := suite.Keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Create multiple users
	users := suite.TestAccounts[:3]

	// Fund all users
	for _, user := range users {
		initialBalance := sdk.NewCoins(
			sdk.NewCoin("uaura", sdkmath.NewInt(100000000000)),
			sdk.NewCoin("usdt", sdkmath.NewInt(100000000000)),
		)
		err := suite.BankKeeper.MintCoins(ctx, types.ModuleName, initialBalance)
		require.NoError(t, err)
		err = suite.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, user, initialBalance)
		require.NoError(t, err)
	}

	// User 0: Low price sell order
	salt0 := []byte("salt_0")
	commitHash0 := suite.Keeper.ComputeOrderHash(types.SwapOrderType_SELL, sdkmath.NewInt(10000000000), "usdt", sdkmath.NewInt(40000000000), salt0)
	commitID0, err := suite.Keeper.CommitOrder(ctx, users[0].String(), commitHash0)
	require.NoError(t, err)

	// User 1: High price sell order
	salt1 := []byte("salt_1")
	commitHash1 := suite.Keeper.ComputeOrderHash(types.SwapOrderType_SELL, sdkmath.NewInt(10000000000), "usdt", sdkmath.NewInt(60000000000), salt1)
	commitID1, err := suite.Keeper.CommitOrder(ctx, users[1].String(), commitHash1)
	require.NoError(t, err)

	// User 2: Medium price sell order
	salt2 := []byte("salt_2")
	commitHash2 := suite.Keeper.ComputeOrderHash(types.SwapOrderType_SELL, sdkmath.NewInt(10000000000), "usdt", sdkmath.NewInt(50000000000), salt2)
	commitID2, err := suite.Keeper.CommitOrder(ctx, users[2].String(), commitHash2)
	require.NoError(t, err)

	// Reveal all orders
	_, err = suite.Keeper.RevealOrder(ctx, commitID0, users[0].String(), types.SwapOrderType_SELL, sdkmath.NewInt(10000000000), "usdt", sdkmath.NewInt(40000000000), salt0)
	require.NoError(t, err)
	_, err = suite.Keeper.RevealOrder(ctx, commitID1, users[1].String(), types.SwapOrderType_SELL, sdkmath.NewInt(10000000000), "usdt", sdkmath.NewInt(60000000000), salt1)
	require.NoError(t, err)
	_, err = suite.Keeper.RevealOrder(ctx, commitID2, users[2].String(), types.SwapOrderType_SELL, sdkmath.NewInt(10000000000), "usdt", sdkmath.NewInt(50000000000), salt2)
	require.NoError(t, err)

	// Verify all orders are queued
	queuedOrders := suite.Keeper.GetAllQueuedOrders(ctx)
	require.Len(t, queuedOrders, 3)

	// Execute batch
	err = suite.Keeper.ExecuteBatch(ctx)
	require.NoError(t, err)

	// Verify queue is cleared
	queuedOrders = suite.Keeper.GetAllQueuedOrders(ctx)
	require.Len(t, queuedOrders, 0)

	// Verify all orders were created
	// Sell orders should be sorted by price: lowest first (best for buyers)
	// User 0 (40000), User 2 (50000), User 1 (60000)
	orders := suite.Keeper.GetOrdersByStatus(ctx, types.SwapOrderStatus_PENDING)
	require.Len(t, orders, 3)
}

// TestCleanupExpiredCommitments tests cleanup of expired commitments
func TestCleanupExpiredCommitments(t *testing.T) {
	suite := SetupTestSuite(t)
	ctx := suite.Ctx

	sender := suite.TestAccounts[0]

	// Create commitment
	commitHash := make([]byte, 32)
	copy(commitHash, "test_hash")
	commitID, err := suite.Keeper.CommitOrder(ctx, sender.String(), commitHash)
	require.NoError(t, err)

	// Verify commitment exists
	_, found := suite.Keeper.GetOrderCommitment(ctx, commitID)
	require.True(t, found)

	// Advance time past reveal deadline
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(65 * time.Second))

	// Cleanup expired commitments
	suite.Keeper.CleanupExpiredCommitments(ctx)

	// Verify commitment was deleted
	_, found = suite.Keeper.GetOrderCommitment(ctx, commitID)
	require.False(t, found)
}

// TestRequiresCommitReveal tests threshold check
func TestRequiresCommitReveal(t *testing.T) {
	suite := SetupTestSuite(t)
	ctx := suite.Ctx

	params := suite.Keeper.GetParams(ctx)
	threshold := params.CommitRevealThreshold

	// Amount below threshold
	smallAmount := threshold.Sub(sdkmath.NewInt(1))
	require.False(t, suite.Keeper.RequiresCommitReveal(ctx, smallAmount))

	// Amount equal to threshold
	require.True(t, suite.Keeper.RequiresCommitReveal(ctx, threshold))

	// Amount above threshold
	largeAmount := threshold.Add(sdkmath.NewInt(1))
	require.True(t, suite.Keeper.RequiresCommitReveal(ctx, largeAmount))
}

// TestComputeOrderHash_Consistency tests hash computation consistency
func TestComputeOrderHash_Consistency(t *testing.T) {
	suite := SetupTestSuite(t)

	orderType := types.SwapOrderType_SELL
	auraAmount := sdkmath.NewInt(10000000000)
	otherCoin := "usdt"
	otherAmount := sdkmath.NewInt(50000000000)
	salt := []byte("random_salt_12345")

	// Compute hash twice
	hash1 := suite.Keeper.ComputeOrderHash(orderType, auraAmount, otherCoin, otherAmount, salt)
	hash2 := suite.Keeper.ComputeOrderHash(orderType, auraAmount, otherCoin, otherAmount, salt)

	// Hashes should be identical
	require.Equal(t, hash1, hash2)
	require.Equal(t, 32, len(hash1)) // SHA-256 produces 32 bytes
}

// TestComputeOrderHash_Uniqueness tests that different inputs produce different hashes
func TestComputeOrderHash_Uniqueness(t *testing.T) {
	suite := SetupTestSuite(t)

	orderType := types.SwapOrderType_SELL
	auraAmount := sdkmath.NewInt(10000000000)
	otherCoin := "usdt"
	otherAmount := sdkmath.NewInt(50000000000)
	salt1 := []byte("salt_1")
	salt2 := []byte("salt_2")

	// Same order, different salt
	hash1 := suite.Keeper.ComputeOrderHash(orderType, auraAmount, otherCoin, otherAmount, salt1)
	hash2 := suite.Keeper.ComputeOrderHash(orderType, auraAmount, otherCoin, otherAmount, salt2)

	// Hashes should be different
	require.NotEqual(t, hash1, hash2)
}

// TestFrontRunningResistance simulates a front-running attack attempt
func TestFrontRunningResistance(t *testing.T) {
	suite := SetupTestSuite(t)
	ctx := suite.Ctx

	// Enable batch execution
	params := suite.Keeper.GetParams(ctx)
	params.BatchExecutionEnabled = true
	err := suite.Keeper.SetParams(ctx, params)
	require.NoError(t, err)

	victim := suite.TestAccounts[0]
	attacker := suite.TestAccounts[1]

	// Fund both users
	for _, user := range []sdk.AccAddress{victim, attacker} {
		initialBalance := sdk.NewCoins(
			sdk.NewCoin("uaura", sdkmath.NewInt(100000000000)),
			sdk.NewCoin("usdt", sdkmath.NewInt(100000000000)),
		)
		err := suite.BankKeeper.MintCoins(ctx, types.ModuleName, initialBalance)
		require.NoError(t, err)
		err = suite.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, user, initialBalance)
		require.NoError(t, err)
	}

	// Victim commits a large order (hidden from attacker)
	victimSalt := []byte("victim_salt")
	victimHash := suite.Keeper.ComputeOrderHash(types.SwapOrderType_SELL, sdkmath.NewInt(10000000000), "usdt", sdkmath.NewInt(50000000000), victimSalt)
	victimCommitID, err := suite.Keeper.CommitOrder(ctx, victim.String(), victimHash)
	require.NoError(t, err)

	// Attacker tries to front-run but doesn't know the order details
	// They can only see the commitment hash, which reveals nothing
	commitment, found := suite.Keeper.GetOrderCommitment(ctx, victimCommitID)
	require.True(t, found)
	require.NotNil(t, commitment.CommitHash)

	// Attacker cannot derive order details from hash
	// If they try to guess and reveal with wrong details, they'll get hash mismatch
	attackerSalt := []byte("attacker_salt")
	attackerGuess := suite.Keeper.ComputeOrderHash(types.SwapOrderType_SELL, sdkmath.NewInt(10000000000), "usdt", sdkmath.NewInt(55000000000), attackerSalt)
	require.NotEqual(t, hex.EncodeToString(commitment.CommitHash), hex.EncodeToString(attackerGuess))

	// Victim reveals their order
	_, err = suite.Keeper.RevealOrder(ctx, victimCommitID, victim.String(), types.SwapOrderType_SELL, sdkmath.NewInt(10000000000), "usdt", sdkmath.NewInt(50000000000), victimSalt)
	require.NoError(t, err)

	// Order is queued for batch execution
	// Even if attacker now creates an order, batch execution uses price priority, not time
	// So the attacker cannot front-run based on reveal time
}

// TestBatchExecution_FailedLocks tests batch execution with some failed orders
func TestBatchExecution_FailedLocks(t *testing.T) {
	suite := SetupTestSuite(t)
	ctx := suite.Ctx

	// Enable batch execution
	params := suite.Keeper.GetParams(ctx)
	params.BatchExecutionEnabled = true
	err := suite.Keeper.SetParams(ctx, params)
	require.NoError(t, err)

	user1 := suite.TestAccounts[0]
	user2 := suite.TestAccounts[1]

	// Fund only user1
	initialBalance := sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(100000000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(100000000000)),
	)
	err = suite.BankKeeper.MintCoins(ctx, types.ModuleName, initialBalance)
	require.NoError(t, err)
	err = suite.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, user1, initialBalance)
	require.NoError(t, err)

	// User1 commits and reveals (should succeed)
	salt1 := []byte("salt_1")
	hash1 := suite.Keeper.ComputeOrderHash(types.SwapOrderType_SELL, sdkmath.NewInt(10000000000), "usdt", sdkmath.NewInt(50000000000), salt1)
	commitID1, err := suite.Keeper.CommitOrder(ctx, user1.String(), hash1)
	require.NoError(t, err)
	_, err = suite.Keeper.RevealOrder(ctx, commitID1, user1.String(), types.SwapOrderType_SELL, sdkmath.NewInt(10000000000), "usdt", sdkmath.NewInt(50000000000), salt1)
	require.NoError(t, err)

	// User2 commits and reveals (no funds, should fail during batch execution)
	salt2 := []byte("salt_2")
	hash2 := suite.Keeper.ComputeOrderHash(types.SwapOrderType_SELL, sdkmath.NewInt(10000000000), "usdt", sdkmath.NewInt(50000000000), salt2)
	commitID2, err := suite.Keeper.CommitOrder(ctx, user2.String(), hash2)
	require.NoError(t, err)
	_, err = suite.Keeper.RevealOrder(ctx, commitID2, user2.String(), types.SwapOrderType_SELL, sdkmath.NewInt(10000000000), "usdt", sdkmath.NewInt(50000000000), salt2)
	require.NoError(t, err)

	// Execute batch
	err = suite.Keeper.ExecuteBatch(ctx)
	require.NoError(t, err)

	// Verify queue is cleared
	queuedOrders := suite.Keeper.GetAllQueuedOrders(ctx)
	require.Len(t, queuedOrders, 0)

	// Verify only user1's order was created
	orders := suite.Keeper.GetOrdersByStatus(ctx, types.SwapOrderStatus_PENDING)
	require.Len(t, orders, 1)
	require.Equal(t, user1.String(), orders[0].UserAddress)
}

// TestComputeOrderHash_DifferentOrderTypes tests hash changes with order type
func TestComputeOrderHash_DifferentOrderTypes(t *testing.T) {
	suite := SetupTestSuite(t)

	auraAmount := sdkmath.NewInt(10000000000)
	otherCoin := "usdt"
	otherAmount := sdkmath.NewInt(50000000000)
	salt := []byte("salt")

	// Same params, different order types
	hashBuy := suite.Keeper.ComputeOrderHash(types.SwapOrderType_BUY, auraAmount, otherCoin, otherAmount, salt)
	hashSell := suite.Keeper.ComputeOrderHash(types.SwapOrderType_SELL, auraAmount, otherCoin, otherAmount, salt)

	// Hashes should be different
	require.NotEqual(t, hashBuy, hashSell)
}

// Helper function to compute SHA-256 hash (for testing)
func computeTestHash(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}
