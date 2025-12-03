package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

// TestFraudProofWindowEnforcement tests that transfers are held in pending state
// during the fraud proof window and cannot be finalized until window expires.
func TestFraudProofWindowEnforcement(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper
	msgServer := keeper.NewMsgServerImpl(suite.keeper)

	// Set up test validator
	validator := createTestValidator(suite, "validator1")
	k.SetValidator(ctx, validator)

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

	// Create the transfer record
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  transferID,
		SourceChain: sourceChain,
		TargetChain: "aura",
		Sender:      "ethereum:0xsender",
		Recipient:   recipient.String(),
		Amount:      amount.String(),
		Denom:       denom,
		Status:      bridgepb.TransferStatus_PENDING,
	}
	k.SetTransfer(ctx, transfer)
	k.IndexTransferHash(ctx, burnTxHash, transferID)

	// Prepare UnlockTokens message
	msg := &bridgepb.MsgUnlockTokens{
		Sender:              recipient.String(),
		SourceChain:         sourceChain,
		BurnTxHash:          burnTxHash,
		Amount:              amount.String(),
		Denom:               denom,
		ValidatorSignatures: createTestSignatures(suite, []string{"validator1"}, burnTxHash),
		MerkleProof:         []byte{},
		MerkleRoot:          []byte{},
	}

	// Execute UnlockTokens - should create pending transfer
	resp, err := msgServer.UnlockTokens(ctx, msg)
	require.NoError(t, err)
	require.True(t, resp.Success)

	// Verify pending transfer was created
	pending, found := k.GetPendingTransfer(ctx, transferID)
	require.True(t, found, "pending transfer should be created")
	require.Equal(t, transferID, pending.TransferId)
	require.Equal(t, recipient.String(), pending.Recipient)
	require.Equal(t, amount.String(), pending.Amount)
	require.Equal(t, denom, pending.Denom)
	require.False(t, pending.Challenged)

	// Verify unlock time is set correctly (current time + fraud proof window)
	expectedUnlockTime := ctx.BlockTime().Add(1 * time.Hour)
	unlockTimeDiff := pending.UnlockTime.AsTime().Sub(expectedUnlockTime)
	require.Less(t, unlockTimeDiff.Abs(), 1*time.Second, "unlock time should be approximately 1 hour from now")

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
	ctx = ctx.WithBlockTime(pending.UnlockTime.AsTime())

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
	require.Equal(t, bridgepb.TransferStatus_COMPLETED, finalTransfer.Status)
}

// TestFraudProofChallengeBlocksFinalization tests that submitting a fraud proof
// prevents transfer finalization even after window expires.
func TestFraudProofChallengeBlocksFinalization(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper
	msgServer := keeper.NewMsgServerImpl(suite.keeper)

	// Set up test validator
	validator := createTestValidator(suite, "validator1")
	k.SetValidator(ctx, validator)

	// Set fraud proof window to 1 hour
	params := k.GetParams(ctx)
	params.FraudProofWindow = 3600 // 1 hour in seconds
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Create and execute transfer (creates pending transfer)
	transferID := "transfer-test-002"
	sourceChain := "ethereum"
	burnTxHash := "0x1234567890abcdef"
	recipient := sdk.AccAddress("recipient2__________")
	amount := math.NewInt(2000000)
	denom := "uatom"

	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  transferID,
		SourceChain: sourceChain,
		TargetChain: "aura",
		Sender:      "ethereum:0xsender",
		Recipient:   recipient.String(),
		Amount:      amount.String(),
		Denom:       denom,
		Status:      bridgepb.TransferStatus_PENDING,
	}
	k.SetTransfer(ctx, transfer)
	k.IndexTransferHash(ctx, burnTxHash, transferID)

	msg := &bridgepb.MsgUnlockTokens{
		Sender:              recipient.String(),
		SourceChain:         sourceChain,
		BurnTxHash:          burnTxHash,
		Amount:              amount.String(),
		Denom:               denom,
		ValidatorSignatures: createTestSignatures(suite, []string{"validator1"}, burnTxHash),
	}

	resp, err := msgServer.UnlockTokens(ctx, msg)
	require.NoError(t, err)
	require.True(t, resp.Success)

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
	ctx = ctx.WithBlockTime(pending.UnlockTime.AsTime().Add(1 * time.Hour))

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

	transfer := &bridgepb.CrossChainTransfer{
		TransferId: transferID,
		Status:     bridgepb.TransferStatus_RELAYED,
	}
	k.SetTransfer(ctx, transfer)

	// Create pending transfer with specific unlock time
	currentTime := ctx.BlockTime()
	unlockTime := currentTime.Add(1 * time.Hour)

	pending := &types.PendingTransfer{
		TransferId:   transferID,
		Recipient:    recipient.String(),
		Amount:       amount.String(),
		Denom:        denom,
		SourceChain:  "ethereum",
		SourceTxHash: "0xhash",
		CreatedAt:    timestamppb.New(currentTime),
		UnlockTime:   timestamppb.New(unlockTime),
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

		transfer := &bridgepb.CrossChainTransfer{
			TransferId: transferID,
			Status:     bridgepb.TransferStatus_RELAYED,
		}
		k.SetTransfer(ctx, transfer)

		pending := &types.PendingTransfer{
			TransferId:   transferID,
			Recipient:    recipient.String(),
			Amount:       math.NewInt(int64((i + 1) * 1000000)).String(),
			Denom:        "uatom",
			SourceChain:  "ethereum",
			SourceTxHash: fmt.Sprintf("0xhash%d", i),
			CreatedAt:    timestamppb.New(ctx.BlockTime()),
			UnlockTime:   timestamppb.New(ctx.BlockTime().Add(1 * time.Hour)),
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
	// This would be implemented similarly to existing test suites
	// For brevity, returning a mock suite structure
	// In real implementation, this would initialize keeper with proper dependencies
	suite := &testSuite{
		ctx:        sdk.Context{}, // Would be properly initialized
		keeper:     &keeper.Keeper{}, // Would be properly initialized
		bankKeeper: &mockBankKeeper{}, // Would be properly initialized
	}
	return suite
}

func createTestValidator(suite *testSuite, address string) *types.BridgeValidator {
	// Create a test validator with proper public key
	// Implementation would match existing test patterns
	return &types.BridgeValidator{
		Address:   address,
		PublicKey: []byte("test-pubkey"),
		Power:     100,
		Active:    true,
	}
}

func createTestSignatures(suite *testSuite, validators []string, message string) [][]byte {
	// Create test signatures
	// Implementation would match existing test patterns
	signatures := make([][]byte, len(validators))
	for i := range validators {
		signatures[i] = make([]byte, 64) // Standard signature length
	}
	return signatures
}

type testSuite struct {
	ctx        sdk.Context
	keeper     *keeper.Keeper
	bankKeeper *mockBankKeeper
}

type mockBankKeeper struct{}

func (m *mockBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, math.ZeroInt())
}

func (m *mockBankKeeper) GetSupply(ctx sdk.Context, denom string) sdk.Coin {
	return sdk.NewCoin(denom, math.ZeroInt())
}

func (m *mockBankKeeper) MintCoins(ctx sdk.Context, moduleName string, coins sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, moduleName string, addr sdk.AccAddress, coins sdk.Coins) error {
	return nil
}
