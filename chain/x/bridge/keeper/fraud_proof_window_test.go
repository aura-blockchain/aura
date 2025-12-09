package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

// TestFraudProofWindowEnforcement tests that transfers are held in pending state
// during the fraud proof window and cannot be finalized until window expires.
//
// NOTE: This test bypasses the full message server signature verification because
// that requires proper cryptographic key setup for validators. Instead, it tests
// the fraud proof window timing logic by directly creating pending transfers.
func TestFraudProofWindowEnforcement(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper

	// Set fraud proof window to 1 hour
	params := k.GetParams(ctx)
	params.FraudProofWindow = 3600 // 1 hour in seconds
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Create a test transfer
	transferID := "transfer-test-001"
	sourceChain := "ethereum"
	burnTxHash := "0xabcdef1234567890"
	recipient := sdk.AccAddress("recipient___________")
	amount := math.NewInt(1000000)
	denom := "uatom"

	// Create the transfer record (would normally be created by UnlockTokens msg)
	transfer := &types.CrossChainTransfer{
		TransferId:  transferID,
		SourceChain: sourceChain,
		TargetChain: "aura",
		Sender:      "ethereum:0xsender",
		Recipient:   recipient.String(),
		Amount:      amount,
		Denom:       denom,
		Status:      types.TransferStatus_RELAYED,
	}
	k.SetTransfer(ctx, transfer)

	// Create pending transfer directly (normally created by UnlockTokens after signature verification)
	// This is the transfer awaiting fraud proof window expiry
	currentTime := ctx.BlockTime()
	unlockTime := currentTime.Add(1 * time.Hour)
	pending := &types.PendingTransfer{
		TransferId:   transferID,
		Recipient:    recipient.String(),
		Amount:       amount,
		Denom:        denom,
		SourceChain:  sourceChain,
		SourceTxHash: burnTxHash,
		CreatedAt:    currentTime,
		UnlockTime:   unlockTime,
		Challenged:   false,
	}
	k.SetPendingTransfer(ctx, pending)

	// Verify pending transfer was created and has correct properties
	retrievedPending, found := k.GetPendingTransfer(ctx, transferID)
	require.True(t, found, "pending transfer should be created")
	require.Equal(t, transferID, retrievedPending.TransferId)
	require.Equal(t, recipient.String(), retrievedPending.Recipient)
	require.True(t, retrievedPending.Amount.Equal(amount))
	require.Equal(t, denom, retrievedPending.Denom)
	require.False(t, retrievedPending.Challenged)
	require.Equal(t, unlockTime, retrievedPending.UnlockTime)

	// TEST 1: Attempt to finalize immediately (should fail - window not expired)
	err = k.FinalizeTransfer(ctx, transferID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "fraud proof window not expired")

	// Verify tokens were NOT minted yet
	balance := suite.bankKeeper.GetBalance(ctx, recipient, denom)
	require.True(t, balance.Amount.IsZero(), "tokens should not be minted yet")

	// TEST 2: Advance time but not enough (30 minutes)
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(30 * time.Minute))

	// Attempt to finalize (should still fail)
	err = k.FinalizeTransfer(ctx, transferID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "fraud proof window not expired")

	// Verify tokens still not minted
	balance = suite.bankKeeper.GetBalance(ctx, recipient, denom)
	require.True(t, balance.Amount.IsZero(), "tokens should not be minted yet")

	// TEST 3: Advance time to exactly when window expires
	ctx = ctx.WithBlockTime(pending.UnlockTime)

	// Now finalization should succeed
	err = k.FinalizeTransfer(ctx, transferID)
	require.NoError(t, err, "finalization should succeed after window expires")

	// Verify tokens were minted and sent to recipient
	balance = suite.bankKeeper.GetBalance(ctx, recipient, denom)
	require.True(t, balance.Amount.Equal(amount), "tokens should be minted after finalization")

	// Verify pending transfer was cleaned up
	_, found = k.GetPendingTransfer(ctx, transferID)
	require.False(t, found, "pending transfer should be deleted after finalization")

	// Verify transfer status updated to COMPLETED
	finalTransfer, found := k.GetTransfer(ctx, transferID)
	require.True(t, found)
	require.Equal(t, types.TransferStatus_COMPLETED, finalTransfer.Status)
}

// TestFraudProofChallengeBlocksFinalization tests that submitting a fraud proof
// prevents transfer finalization even after window expires.
//
// NOTE: This test bypasses the full message server signature verification and
// directly creates pending transfers to focus on testing the fraud proof challenge workflow.
func TestFraudProofChallengeBlocksFinalization(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper

	// Set fraud proof window to 1 hour
	params := k.GetParams(ctx)
	params.FraudProofWindow = 3600 // 1 hour in seconds
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Create test transfer
	transferID := "transfer-test-002"
	sourceChain := "ethereum"
	burnTxHash := "0x1234567890abcdef"
	recipient := sdk.AccAddress("recipient2__________")
	amount := math.NewInt(2000000)
	denom := "uatom"

	// Create transfer record
	transfer := &types.CrossChainTransfer{
		TransferId:  transferID,
		SourceChain: sourceChain,
		TargetChain: "aura",
		Sender:      "ethereum:0xsender",
		Recipient:   recipient.String(),
		Amount:      amount,
		Denom:       denom,
		Status:      types.TransferStatus_RELAYED,
	}
	k.SetTransfer(ctx, transfer)

	// Create pending transfer directly
	currentTime := ctx.BlockTime()
	unlockTime := currentTime.Add(1 * time.Hour)
	pending := &types.PendingTransfer{
		TransferId:   transferID,
		Recipient:    recipient.String(),
		Amount:       amount,
		Denom:        denom,
		SourceChain:  sourceChain,
		SourceTxHash: burnTxHash,
		CreatedAt:    currentTime,
		UnlockTime:   unlockTime,
		Challenged:   false,
	}
	k.SetPendingTransfer(ctx, pending)

	// Verify pending transfer was created
	pending, found := k.GetPendingTransfer(ctx, transferID)
	require.True(t, found)
	require.False(t, pending.Challenged)

	// Submit fraud proof during the window
	challenger := sdk.AccAddress("challenger__________")
	fraudEvidence := []byte("evidence of invalid signature")

	err = k.SubmitFraudProof(ctx, transferID, challenger.String(), fraudEvidence)
	require.NoError(t, err, "fraud proof submission should succeed")

	// Verify pending transfer is now marked as challenged
	pending, found = k.GetPendingTransfer(ctx, transferID)
	require.True(t, found)
	require.True(t, pending.Challenged, "pending transfer should be marked as challenged")
	require.NotEmpty(t, pending.FraudProofId)

	// Advance time past the fraud proof window
	ctx = ctx.WithBlockTime(pending.UnlockTime.Add(1 * time.Hour))

	// Attempt to finalize (should fail because transfer is challenged)
	err = k.FinalizeTransfer(ctx, transferID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "challenged with fraud proof")

	// Verify tokens were NOT minted
	balance := suite.bankKeeper.GetBalance(ctx, recipient, denom)
	require.True(t, balance.Amount.IsZero(), "tokens should not be minted when challenged")

	// Verify pending transfer still exists (not finalized)
	_, found = k.GetPendingTransfer(ctx, transferID)
	require.True(t, found, "pending transfer should still exist when challenged")
}

// TestFraudProofSubmissionAfterWindowFails tests that fraud proofs cannot be submitted
// after the window expires.
func TestFraudProofSubmissionAfterWindowFails(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper

	// Set fraud proof window to 1 hour
	params := k.GetParams(ctx)
	params.FraudProofWindow = 3600
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Create a pending transfer
	transferID := "transfer-test-003"
	recipient := sdk.AccAddress("recipient3__________")
	amount := math.NewInt(3000000)
	denom := "uatom"

	transfer := &types.CrossChainTransfer{
		TransferId: transferID,
		Status:     types.TransferStatus_RELAYED,
	}
	k.SetTransfer(ctx, transfer)

	// Create pending transfer with specific unlock time
	currentTime := ctx.BlockTime()
	unlockTime := currentTime.Add(1 * time.Hour)

	pending := &types.PendingTransfer{
		TransferId:   transferID,
		Recipient:    recipient.String(),
		Amount:       amount,
		Denom:        denom,
		SourceChain:  "ethereum",
		SourceTxHash: "0xhash",
		CreatedAt:    currentTime,
		UnlockTime:   unlockTime,
		Challenged:   false,
	}
	// Use internal helper to set pending transfer
	k.SetPendingTransfer(ctx, pending)

	// Advance time past the fraud proof window
	ctx = ctx.WithBlockTime(unlockTime.Add(1 * time.Hour))

	// Attempt to submit fraud proof after window expired
	challenger := sdk.AccAddress("challenger__________")
	fraudEvidence := []byte("too late evidence")

	err = k.SubmitFraudProof(ctx, transferID, challenger.String(), fraudEvidence)
	require.Error(t, err)
	require.Equal(t, types.ErrFraudProofExpired, err)
}

// TestMultiplePendingTransfers tests that multiple transfers can be pending simultaneously.
func TestMultiplePendingTransfers(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper

	// Set fraud proof window
	params := k.GetParams(ctx)
	params.FraudProofWindow = 3600
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Create multiple pending transfers
	numTransfers := 5
	for i := 0; i < numTransfers; i++ {
		transferID := fmt.Sprintf("transfer-%d", i)
		recipient := sdk.AccAddress(fmt.Sprintf("recipient%d_________", i))

		transfer := &types.CrossChainTransfer{
			TransferId: transferID,
			Status:     types.TransferStatus_RELAYED,
		}
		k.SetTransfer(ctx, transfer)

		pending := &types.PendingTransfer{
			TransferId:   transferID,
			Recipient:    recipient.String(),
			Amount:       math.NewInt(int64((i + 1) * 1000000)),
			Denom:        "uatom",
			SourceChain:  "ethereum",
			SourceTxHash: fmt.Sprintf("0xhash%d", i),
			CreatedAt:    ctx.BlockTime(),
			UnlockTime:   ctx.BlockTime().Add(1 * time.Hour),
			Challenged:   false,
		}
		k.SetPendingTransfer(ctx, pending)
	}

	// Retrieve all pending transfers
	allPending := k.GetAllPendingTransfers(ctx)
	require.Len(t, allPending, numTransfers, "should have correct number of pending transfers")

	// Challenge one transfer
	err = k.MarkPendingTransferChallenged(ctx, "transfer-2", "fraud-proof-2")
	require.NoError(t, err)

	// Advance time past window
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Hour))

	// Finalize non-challenged transfers
	for i := 0; i < numTransfers; i++ {
		transferID := fmt.Sprintf("transfer-%d", i)

		if i == 2 {
			// This one is challenged, should fail
			err = k.FinalizeTransfer(ctx, transferID)
			require.Error(t, err)
			require.Contains(t, err.Error(), "challenged")
		} else {
			// Others should succeed
			err = k.FinalizeTransfer(ctx, transferID)
			require.NoError(t, err)
		}
	}

	// Verify only challenged transfer remains pending
	allPending = k.GetAllPendingTransfers(ctx)
	require.Len(t, allPending, 1, "only challenged transfer should remain pending")
	require.Equal(t, "transfer-2", allPending[0].TransferId)
}

// TestFraudProofWindowParameter tests parameter validation and usage.
func TestFraudProofWindowParameter(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper

	// Test 1: Default value
	params := k.GetParams(ctx)
	require.Greater(t, params.FraudProofWindow, int64(0), "default fraud proof window should be > 0")

	// Test 2: Valid window (1 day)
	params.FraudProofWindow = 86400 // 1 day
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Test 3: Too short window (should fail validation)
	params.FraudProofWindow = 1800 // 30 minutes (less than 1 hour minimum)
	err = params.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "at least 1 hour")

	// Test 4: Too long window (should fail validation)
	params.FraudProofWindow = 2592001 // > 30 days
	err = params.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot exceed 30 days")

	// Test 5: Exact boundaries
	params.FraudProofWindow = 3600 // Exactly 1 hour (minimum)
	err = params.Validate()
	require.NoError(t, err)

	params.FraudProofWindow = 2592000 // Exactly 30 days (maximum)
	err = params.Validate()
	require.NoError(t, err)
}

// Helper functions for test setup
func setupTestSuite(t *testing.T) *testSuite {
	// Create proper test input with initialized store
	input := keepertest.CreateTestInput(t)

	// Create mock bank keeper for token operations
	bankKeeper := &mockBankKeeper{
		balances: make(map[string]sdk.Coins),
		supplies: make(map[string]math.Int),
	}

	// Initialize keeper with proper dependencies
	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil, // paramstore not needed for these tests
		bankKeeper,
		nil, // accountKeeper not needed
		nil, // vcKeeper not needed
		nil, // stakingKeeper not needed
	)

	return &testSuite{
		ctx:        input.Ctx,
		keeper:     k,
		bankKeeper: bankKeeper,
	}
}

func createTestValidator(suite *testSuite, address string) *types.BridgeValidator {
	return &types.BridgeValidator{
		Address:   address,
		PublicKey: []byte("test-pubkey"),
		Power:     100,
		Active:    true,
		Chains:    []string{"aura", "ethereum"},
	}
}

func createTestSignatures(suite *testSuite, validators []string, message string) [][]byte {
	signatures := make([][]byte, len(validators))
	for i := range validators {
		// Create 64-byte signatures (standard secp256k1 signature length)
		signatures[i] = make([]byte, 64)
		// Fill with deterministic data for testing
		for j := range signatures[i] {
			signatures[i][j] = byte(i + j)
		}
	}
	return signatures
}

type testSuite struct {
	ctx        sdk.Context
	keeper     *keeper.Keeper
	bankKeeper *mockBankKeeper
}

type mockBankKeeper struct {
	balances map[string]sdk.Coins
	supplies map[string]math.Int
}

func (m *mockBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if coins, ok := m.balances[addr.String()]; ok {
		return sdk.NewCoin(denom, coins.AmountOf(denom))
	}
	return sdk.NewCoin(denom, math.ZeroInt())
}

func (m *mockBankKeeper) GetSupply(ctx sdk.Context, denom string) sdk.Coin {
	if supply, ok := m.supplies[denom]; ok {
		return sdk.NewCoin(denom, supply)
	}
	return sdk.NewCoin(denom, math.ZeroInt())
}

func (m *mockBankKeeper) MintCoins(ctx sdk.Context, moduleName string, coins sdk.Coins) error {
	// Update supplies
	for _, coin := range coins {
		if supply, ok := m.supplies[coin.Denom]; ok {
			m.supplies[coin.Denom] = supply.Add(coin.Amount)
		} else {
			m.supplies[coin.Denom] = coin.Amount
		}
	}
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, moduleName string, addr sdk.AccAddress, coins sdk.Coins) error {
	// Update recipient balance
	key := addr.String()
	if balance, ok := m.balances[key]; ok {
		m.balances[key] = balance.Add(coins...)
	} else {
		m.balances[key] = coins
	}
	return nil
}

func (m *mockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, coins sdk.Coins) error {
	// Update supplies (reduce)
	for _, coin := range coins {
		if supply, ok := m.supplies[coin.Denom]; ok {
			m.supplies[coin.Denom] = supply.Sub(coin.Amount)
		}
	}
	return nil
}

func (m *mockBankKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	fromKey := fromAddr.String()
	toKey := toAddr.String()

	// Check sender balance
	if fromBalance, ok := m.balances[fromKey]; ok {
		if !fromBalance.IsAllGTE(amt) {
			return fmt.Errorf("insufficient funds")
		}
		m.balances[fromKey] = fromBalance.Sub(amt...)
	} else {
		return fmt.Errorf("insufficient funds")
	}

	// Update recipient balance
	if toBalance, ok := m.balances[toKey]; ok {
		m.balances[toKey] = toBalance.Add(amt...)
	} else {
		m.balances[toKey] = amt
	}

	return nil
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	senderKey := senderAddr.String()

	// Check and reduce sender balance
	if balance, ok := m.balances[senderKey]; ok {
		if !balance.IsAllGTE(amt) {
			return fmt.Errorf("insufficient funds")
		}
		m.balances[senderKey] = balance.Sub(amt...)
	} else {
		return fmt.Errorf("insufficient funds")
	}

	return nil
}
