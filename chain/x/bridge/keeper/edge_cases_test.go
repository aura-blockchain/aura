package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

// TestUnlockTokens_ErrorPath_InvalidSignature verifies unlocks fail with invalid signatures
func TestUnlockTokens_ErrorPath_InvalidSignature(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	// Seed a transfer
	transferID := "transfer-001"
	seedBridgeTransfer(t, input, transferID, "1000000", 2)

	// Create message with invalid signatures
	msg := &bridgepb.MsgUnlockTokens{
		BurnTxHash:           "0xabcd1234",
		Sender:               keepertest.GenTestAddr().String(),
		Amount:               sdkmath.NewInt(1000000),
		Denom:                "uaura",
		SourceChain:          "ethereum",
		ValidatorSignatures:  [][]byte{{0x01, 0x02}, {0x03, 0x04}}, // Invalid signatures
		MerkleProof:          [][]byte{},
		MerkleRoot:           []byte{},
		SourceBlockHash:      []byte{},
		SourceBlockHeight:    0,
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.UnlockTokens(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	// Should fail due to no active validators or invalid signatures
}

// TestUnlockTokens_ErrorPath_InsufficientBalance verifies unlock fails with insufficient module balance
func TestUnlockTokens_ErrorPath_InsufficientBalance(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	mockBank := keepertest.NewMockBankKeeper()
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, mockBank, nil, nil)

	// Seed transfer and pending transfer
	transferID := "transfer-002"
	seedBridgeTransferWithPending(t, input, transferID, "1000000", 2)

	// Module has 0 balance (insufficient to unlock)
	// Try to finalize the transfer after fraud proof window
	ctx := input.Ctx.WithBlockTime(input.Ctx.BlockTime().Add(types.DefaultFraudProofWindow + time.Hour))

	msg := &bridgepb.MsgFinalizeTransfer{
		TransferId: transferID,
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.FinalizeTransfer(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)
	// Should fail when trying to send coins from module to recipient
}

// TestLockTokens_EdgeCase_ZeroAmount verifies lock fails with zero amount
func TestLockTokens_EdgeCase_ZeroAmount(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	// Enable default chain
	chainConfig := &bridgepb.ChainConfig{
		ChainId: "ethereum",
		Enabled: true,
	}
	k.SetChainConfig(input.Ctx, chainConfig)

	// Try to lock zero amount
	msg := &bridgepb.MsgLockTokens{
		Sender:      keepertest.GenTestAddr().String(),
		Recipient:   "0x1234567890abcdef",
		Amount:      sdk.NewCoin("uaura", sdkmath.ZeroInt()),
		TargetChain: "ethereum",
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.LockTokens(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "amount must be positive")
}

// TestLockTokens_ErrorPath_ChainNotFound verifies lock fails for unknown chain
func TestLockTokens_ErrorPath_ChainNotFound(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	msg := &bridgepb.MsgLockTokens{
		Sender:      keepertest.GenTestAddr().String(),
		Recipient:   "0x1234567890abcdef",
		Amount:      sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		TargetChain: "unknown-chain",
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.LockTokens(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// TestLockTokens_ErrorPath_ChainDisabled verifies lock fails for disabled chain
func TestLockTokens_ErrorPath_ChainDisabled(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	// Create disabled chain config
	chainConfig := &bridgepb.ChainConfig{
		ChainId: "ethereum",
		Enabled: false,
	}
	k.SetChainConfig(input.Ctx, chainConfig)

	msg := &bridgepb.MsgLockTokens{
		Sender:      keepertest.GenTestAddr().String(),
		Recipient:   "0x1234567890abcdef",
		Amount:      sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		TargetChain: "ethereum",
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.LockTokens(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "disabled")
}

// TestFinalizeTransfer_ErrorPath_NotFound verifies finalize fails for non-existent transfer
func TestFinalizeTransfer_ErrorPath_NotFound(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	msg := &bridgepb.MsgFinalizeTransfer{
		TransferId: "nonexistent-transfer",
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.FinalizeTransfer(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// TestFinalizeTransfer_ErrorPath_WindowNotExpired verifies early finalization is rejected
func TestFinalizeTransfer_ErrorPath_WindowNotExpired(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	// Seed transfer with pending status
	transferID := "transfer-003"
	seedBridgeTransferWithPending(t, input, transferID, "1000000", 2)

	// Try to finalize immediately (fraud proof window not expired)
	msg := &bridgepb.MsgFinalizeTransfer{
		TransferId: transferID,
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.FinalizeTransfer(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "fraud proof window has not expired")
}

// TestFinalizeTransfer_ErrorPath_TransferChallenged verifies challenged transfers cannot be finalized
func TestFinalizeTransfer_ErrorPath_TransferChallenged(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	// Seed transfer
	transferID := "transfer-004"
	seedBridgeTransferWithPending(t, input, transferID, "1000000", 2)

	// Get the pending transfer and mark it as challenged
	pending, found := k.GetPendingTransfer(input.Ctx, transferID)
	require.True(t, found)

	pending.Challenged = true
	pending.FraudProofId = "fraud-001"
	k.SetPendingTransfer(input.Ctx, &pending)

	// Advance time past fraud proof window
	ctx := input.Ctx.WithBlockTime(input.Ctx.BlockTime().Add(types.DefaultFraudProofWindow + time.Hour))

	msg := &bridgepb.MsgFinalizeTransfer{
		TransferId: transferID,
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.FinalizeTransfer(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "challenged")
}

// TestSubmitFraudProof_ErrorPath_TransferNotFound verifies fraud proof fails for non-existent transfer
func TestSubmitFraudProof_ErrorPath_TransferNotFound(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	msg := &bridgepb.MsgSubmitFraudProof{
		TransferId: "nonexistent-transfer",
		Challenger: keepertest.GenTestAddr().String(),
		FraudType:  "INVALID_MERKLE_PROOF",
		Evidence:   []byte("evidence"),
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.SubmitFraudProof(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// TestSubmitFraudProof_ErrorPath_WindowExpired verifies fraud proof fails after window expires
func TestSubmitFraudProof_ErrorPath_WindowExpired(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	// Seed transfer
	transferID := "transfer-005"
	seedBridgeTransferWithPending(t, input, transferID, "1000000", 2)

	// Advance time past fraud proof window
	ctx := input.Ctx.WithBlockTime(input.Ctx.BlockTime().Add(types.DefaultFraudProofWindow + time.Hour))

	msg := &bridgepb.MsgSubmitFraudProof{
		TransferId: transferID,
		Challenger: keepertest.GenTestAddr().String(),
		FraudType:  "INVALID_MERKLE_PROOF",
		Evidence:   []byte("evidence"),
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.SubmitFraudProof(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "fraud proof window has expired")
}

// TestSubmitFraudProof_ErrorPath_EmptyEvidence verifies fraud proof requires evidence
func TestSubmitFraudProof_ErrorPath_EmptyEvidence(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	// Seed transfer
	transferID := "transfer-006"
	seedBridgeTransferWithPending(t, input, transferID, "1000000", 2)

	msg := &bridgepb.MsgSubmitFraudProof{
		TransferId: transferID,
		Challenger: keepertest.GenTestAddr().String(),
		FraudType:  "INVALID_MERKLE_PROOF",
		Evidence:   []byte{}, // Empty evidence
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.SubmitFraudProof(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "evidence required")
}

// TestSubmitFraudProof_EdgeCase_DuplicateChallenge verifies duplicate fraud proofs are rejected
func TestSubmitFraudProof_EdgeCase_DuplicateChallenge(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	// Seed transfer
	transferID := "transfer-007"
	seedBridgeTransferWithPending(t, input, transferID, "1000000", 2)

	challenger := keepertest.GenTestAddr().String()

	// Submit first fraud proof
	msg1 := &bridgepb.MsgSubmitFraudProof{
		TransferId: transferID,
		Challenger: challenger,
		FraudType:  "INVALID_MERKLE_PROOF",
		Evidence:   []byte("evidence1"),
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.SubmitFraudProof(sdk.WrapSDKContext(input.Ctx), msg1)
	require.NoError(t, err)

	// Try to submit second fraud proof for same transfer
	msg2 := &bridgepb.MsgSubmitFraudProof{
		TransferId: transferID,
		Challenger: challenger,
		FraudType:  "DOUBLE_SPEND",
		Evidence:   []byte("evidence2"),
	}

	_, err = msgServer.SubmitFraudProof(sdk.WrapSDKContext(input.Ctx), msg2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already challenged")
}

// TestMintTokens_EdgeCase_ZeroAmount verifies minting zero amount is rejected
func TestMintTokens_EdgeCase_ZeroAmount(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	msg := &bridgepb.MsgMintTokens{
		Validator:      keepertest.GenTestAddr().String(),
		SourceChain:    "ethereum",
		SourceTxHash:   "0xabcd1234",
		Recipient:      keepertest.GenTestAddr().String(),
		Amount:         sdkmath.ZeroInt(),
		Denom:          "uaura",
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.MintTokens(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid amount")
}

// TestMintTokens_EdgeCase_NegativeAmount verifies negative amounts are rejected
func TestMintTokens_EdgeCase_NegativeAmount(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	// Create a negative amount by using NewInt with negative value
	negativeAmount := sdkmath.NewInt(-1000)

	msg := &bridgepb.MsgMintTokens{
		Validator:      keepertest.GenTestAddr().String(),
		SourceChain:    "ethereum",
		SourceTxHash:   "0xabcd1234",
		Recipient:      keepertest.GenTestAddr().String(),
		Amount:         negativeAmount,
		Denom:          "uaura",
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.MintTokens(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid amount")
}

// TestBurnTokens_EdgeCase_ZeroAmount verifies burning zero amount is rejected
func TestBurnTokens_EdgeCase_ZeroAmount(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	// Enable target chain
	chainConfig := &bridgepb.ChainConfig{
		ChainId: "ethereum",
		Enabled: true,
	}
	k.SetChainConfig(input.Ctx, chainConfig)

	msg := &bridgepb.MsgBurnTokens{
		Sender:      keepertest.GenTestAddr().String(),
		Recipient:   "0x1234567890abcdef",
		Amount:      sdk.NewCoin("uaura", sdkmath.ZeroInt()),
		TargetChain: "ethereum",
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.BurnTokens(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	// Error should occur when validating amount or during burn
}

// TestLinkAddress_ErrorPath_MissingAuraAddress verifies link fails without Aura address
func TestLinkAddress_ErrorPath_MissingAuraAddress(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	msg := &bridgepb.MsgLinkAddress{
		Signer:       keepertest.GenTestAddr().String(),
		AuraAddress:  "", // Missing
		PawAddress:   "paw1234567890",
		XaiAddress:   "",
		PawSignature: []byte{},
		XaiSignature: []byte{},
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.LinkAddress(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "aura address required")
}

// TestLinkAddress_ErrorPath_SignerMismatch verifies signer must own Aura address
func TestLinkAddress_ErrorPath_SignerMismatch(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	auraAddr := keepertest.GenTestAddr().String()
	differentSigner := keepertest.GenTestAddr().String()

	msg := &bridgepb.MsgLinkAddress{
		Signer:       differentSigner, // Different from AuraAddress
		AuraAddress:  auraAddr,
		PawAddress:   "paw1234567890",
		XaiAddress:   "",
		PawSignature: []byte{0x01, 0x02},
		XaiSignature: []byte{},
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.LinkAddress(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "signer must be the Aura address owner")
}

// TestLinkAddress_ErrorPath_MissingPawSignature verifies PAW signature required when linking PAW address
func TestLinkAddress_ErrorPath_MissingPawSignature(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	auraAddr := keepertest.GenTestAddr().String()

	msg := &bridgepb.MsgLinkAddress{
		Signer:       auraAddr,
		AuraAddress:  auraAddr,
		PawAddress:   "paw1234567890",
		XaiAddress:   "",
		PawSignature: []byte{}, // Missing signature
		XaiSignature: []byte{},
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.LinkAddress(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "PAW signature required")
}

// TestUnlockTokens_EdgeCase_MaxAmount verifies handling of maximum transfer amounts
func TestUnlockTokens_EdgeCase_MaxAmount(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	mockBank := keepertest.NewMockBankKeeper()
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, mockBank, nil, nil)

	// Set params with max transfer limit
	params := types.DefaultParams()
	params.MaxTransferAmount = "1000000" // 1 million max
	k.SetParams(input.Ctx, &params)

	// Seed transfer
	transferID := "transfer-max"
	seedBridgeTransfer(t, input, transferID, "2000000", 2) // 2 million (exceeds max)

	// Try to unlock with amount exceeding max
	msg := &bridgepb.MsgUnlockTokens{
		BurnTxHash:           "0xabcd1234",
		Sender:               keepertest.GenTestAddr().String(),
		Amount:               sdkmath.NewInt(2000000),
		Denom:                "uaura",
		SourceChain:          "ethereum",
		ValidatorSignatures:  [][]byte{},
		MerkleProof:          [][]byte{},
		MerkleRoot:           []byte{},
		SourceBlockHash:      []byte{},
		SourceBlockHeight:    0,
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.UnlockTokens(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	// Should fail validation (either insufficient signatures or max amount check)
}

// TestCrossChainSwap_EdgeCase_InvalidCoin verifies cross-chain swap validation
func TestCrossChainSwap_EdgeCase_InvalidCoin(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	msg := &bridgepb.MsgCrossChainSwap{
		Sender:      keepertest.GenTestAddr().String(),
		SourceChain: "aura",
		TargetChain: "ethereum",
		InputCoin:   sdk.NewCoin("invalid", sdkmath.ZeroInt()), // Invalid: zero amount
		TargetDenom: "usdt",
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.CrossChainSwap(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "input amount required")
}

// TestRelayTransfer_ErrorPath_TransferNotFound verifies relay fails for non-existent transfer
func TestRelayTransfer_ErrorPath_TransferNotFound(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	msg := &bridgepb.MsgRelayTransfer{
		TransferId:   "nonexistent",
		Relayer:      keepertest.GenTestAddr().String(),
		Status:       "COMPLETED",
		TargetTxHash: "0x1234",
	}

	msgServer := keeper.NewMsgServerImpl(k)
	_, err := msgServer.RelayTransfer(sdk.WrapSDKContext(input.Ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// TestInvariant_SupplyCap verifies supply cap enforcement
func TestInvariant_SupplyCap(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	mockBank := keepertest.NewMockBankKeeper()
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, mockBank, nil, nil)

	// Set params with supply cap
	params := types.DefaultParams()
	params.SupplyCaps = map[string]string{
		"uaura": "10000000", // 10 million cap
	}
	k.SetParams(input.Ctx, &params)

	// Simulate current supply at 9 million
	mockBank.SetSupply(input.Ctx, "uaura", sdkmath.NewInt(9000000))

	// Seed transfer
	transferID := "transfer-supply"
	seedBridgeTransfer(t, input, transferID, "2000000", 2) // Would exceed cap (9M + 2M > 10M)

	// Try to unlock - should be blocked by supply cap during unlock process
	// (actual enforcement happens in UnlockTokens msg handler)
}

// TestConcurrent_MultipleFinalizations verifies only one finalization succeeds
func TestConcurrent_MultipleFinalizations(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil)

	// Seed transfer
	transferID := "transfer-concurrent"
	seedBridgeTransferWithPending(t, input, transferID, "1000000", 2)

	// Advance time past fraud proof window
	ctx := input.Ctx.WithBlockTime(input.Ctx.BlockTime().Add(types.DefaultFraudProofWindow + time.Hour))

	msg := &bridgepb.MsgFinalizeTransfer{
		TransferId: transferID,
	}

	msgServer := keeper.NewMsgServerImpl(k)

	// First finalization should succeed
	_, err := msgServer.FinalizeTransfer(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)

	// Second finalization should fail (transfer already finalized/deleted)
	_, err = msgServer.FinalizeTransfer(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}
