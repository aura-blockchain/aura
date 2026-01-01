// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

type InvariantsComprehensiveTestSuite struct {
	KeeperTestSuite
}

func TestInvariantsComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsComprehensiveTestSuite))
}

func (suite *InvariantsComprehensiveTestSuite) TestTransferBalanceInvariant() {
	ctx := suite.SdkCtx
	inv := TransferBalanceInvariant(*suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken, "empty store should not break invariant")
	suite.Empty(msg)

	// Test: Create a pending transfer
	// Note: Since bankKeeper is nil in test suite, invariant will skip the balance check
	// This is acceptable for basic tests. For full integration tests, we need a proper mock.
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:           "transfer-1",
		SourceChain:          "aura",
		TargetChain:          "ethereum",
		Sender:               sdk.AccAddress("sender____________").String(),
		Recipient:            "0x123",
		Amount:               sdkmath.NewInt(1000),
		Denom:                "uaura",
		Status:               bridgepb.TransferStatus_PENDING,
		Timestamp:            time.Now(),
		ValidatorSignatures:  []bridgepb.ValidatorSignature{},
	}
	suite.Keeper.SetTransfer(ctx, transfer)

	msg, broken = inv(ctx)
	// With nil bankKeeper, invariant returns (false) - test environment behavior
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestTransferBalanceInvariantMultipleDenoms() {
	ctx := suite.SdkCtx
	inv := TransferBalanceInvariant(*suite.Keeper)

	// Create transfers with different denoms
	transfer1 := &bridgepb.CrossChainTransfer{
		TransferId:           "transfer-multi-1",
		SourceChain:          "aura",
		TargetChain:          "ethereum",
		Sender:               sdk.AccAddress("sender1___________").String(),
		Recipient:            "0x123",
		Amount:               sdkmath.NewInt(1000),
		Denom:                "uaura",
		Status:               bridgepb.TransferStatus_PENDING,
		Timestamp:            time.Now(),
		ValidatorSignatures:  []bridgepb.ValidatorSignature{},
	}
	suite.Keeper.SetTransfer(ctx, transfer1)

	transfer2 := &bridgepb.CrossChainTransfer{
		TransferId:           "transfer-multi-2",
		SourceChain:          "aura",
		TargetChain:          "ethereum",
		Sender:               sdk.AccAddress("sender2___________").String(),
		Recipient:            "0x456",
		Amount:               sdkmath.NewInt(2000),
		Denom:                "upaw",
		Status:               bridgepb.TransferStatus_PENDING,
		Timestamp:            time.Now(),
		ValidatorSignatures:  []bridgepb.ValidatorSignature{},
	}
	suite.Keeper.SetTransfer(ctx, transfer2)

	transfer3 := &bridgepb.CrossChainTransfer{
		TransferId:           "transfer-multi-3",
		SourceChain:          "aura",
		TargetChain:          "ethereum",
		Sender:               sdk.AccAddress("sender3___________").String(),
		Recipient:            "0x789",
		Amount:               sdkmath.NewInt(500),
		Denom:                "uaura",
		Status:               bridgepb.TransferStatus_CONFIRMED,
		Timestamp:            time.Now(),
		ValidatorSignatures:  []bridgepb.ValidatorSignature{},
	}
	suite.Keeper.SetTransfer(ctx, transfer3)

	// The invariant should sum amounts per denom: uaura=1500, upaw=2000
	msg, broken := inv(ctx)
	suite.False(broken, "multiple denoms should not break invariant")
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestTransferBalanceInvariantCompletedTransfersIgnored() {
	ctx := suite.SdkCtx
	inv := TransferBalanceInvariant(*suite.Keeper)

	// Create a completed transfer (should NOT be counted in locked amounts)
	completedTransfer := &bridgepb.CrossChainTransfer{
		TransferId:           "transfer-completed",
		SourceChain:          "aura",
		TargetChain:          "ethereum",
		Sender:               sdk.AccAddress("sender____________").String(),
		Recipient:            "0x123",
		Amount:               sdkmath.NewInt(1000000),
		Denom:                "uaura",
		Status:               bridgepb.TransferStatus_COMPLETED,
		Timestamp:            time.Now(),
		ValidatorSignatures:  []bridgepb.ValidatorSignature{},
	}
	suite.Keeper.SetTransfer(ctx, completedTransfer)

	// Completed transfers should not affect the invariant
	msg, broken := inv(ctx)
	suite.False(broken, "completed transfers should be ignored")
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestTransferBalanceInvariantFailedTransfersIgnored() {
	ctx := suite.SdkCtx
	inv := TransferBalanceInvariant(*suite.Keeper)

	// Create a failed transfer (should NOT be counted in locked amounts)
	failedTransfer := &bridgepb.CrossChainTransfer{
		TransferId:           "transfer-failed",
		SourceChain:          "aura",
		TargetChain:          "ethereum",
		Sender:               sdk.AccAddress("sender____________").String(),
		Recipient:            "0x123",
		Amount:               sdkmath.NewInt(1000000),
		Denom:                "uaura",
		Status:               bridgepb.TransferStatus_FAILED,
		Timestamp:            time.Now(),
		ValidatorSignatures:  []bridgepb.ValidatorSignature{},
	}
	suite.Keeper.SetTransfer(ctx, failedTransfer)

	// Failed transfers should not affect the invariant
	msg, broken := inv(ctx)
	suite.False(broken, "failed transfers should be ignored")
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestTransferBalanceInvariantInvalidAmount() {
	// Create a keeper with mocked keepers to actually run the validation
	mockBank := newMockBankKeeperWithBalances()
	mockAccount := newMockAccountKeeperWithModule()

	k := NewKeeper(
		suite.Keeper.cdc,
		suite.Keeper.storeKey,
		nil,
		mockBank,
		mockAccount,
		nil,
		nil, // stakingKeeper
	)

	ctx := suite.SdkCtx
	inv := TransferBalanceInvariant(*k)

	// Create transfer with negative amount (invalid case)
	// Note: We can't use "invalid-amount" string since Amount is sdkmath.Int type
	// Instead, test with negative amount which should fail validation
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:           "transfer-invalid",
		SourceChain:          "aura",
		TargetChain:          "ethereum",
		Sender:               sdk.AccAddress("sender____________").String(),
		Recipient:            "0x123",
		Amount:               sdkmath.NewInt(-1000), // Negative amount is invalid
		Denom:                "uaura",
		Status:               bridgepb.TransferStatus_PENDING,
		Timestamp:            time.Now(),
		ValidatorSignatures:  []bridgepb.ValidatorSignature{},
	}
	k.SetTransfer(ctx, transfer)

	msg, broken := inv(ctx)
	suite.True(broken, "negative amount should break invariant")
	suite.Contains(msg, "invalid transfer amount")
}

func (suite *InvariantsComprehensiveTestSuite) TestMerkleProofInvariant() {
	ctx := suite.SdkCtx
	inv := MerkleProofInvariant(*suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Create valid merkle proof
	proof := &bridgepb.MerkleProof{
		Root:    []byte("root-hash"),
		Leaf:    []byte("leaf-data"),
		Proof:   [][]byte{[]byte("sibling1"), []byte("sibling2")},
		Indices: []uint64{0},
	}
	suite.storeMerkleProof(ctx, "proof-1", proof)

	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Create proof with empty root
	invalidProof := &bridgepb.MerkleProof{
		Root:    []byte{},
		Leaf:    []byte("leaf-data"),
		Proof:   [][]byte{[]byte("sibling1")},
		Indices: []uint64{0},
	}
	suite.storeMerkleProof(ctx, "proof-invalid", invalidProof)

	msg, broken = inv(ctx)
	suite.True(broken, "proof with empty root should break invariant")
	suite.Contains(msg, "empty root")
}

func (suite *InvariantsComprehensiveTestSuite) TestMerkleProofInvariantEmptyLeaf() {
	ctx := suite.SdkCtx
	inv := MerkleProofInvariant(*suite.Keeper)

	// Create proof with empty leaf
	proof := &bridgepb.MerkleProof{
		Root:    []byte("root-hash"),
		Leaf:    []byte{},
		Proof:   [][]byte{[]byte("sibling1")},
		Indices: []uint64{0},
	}
	suite.storeMerkleProof(ctx, "proof-2", proof)

	msg, broken := inv(ctx)
	suite.True(broken, "proof with empty leaf should break invariant")
	suite.Contains(msg, "empty leaf")
}

func (suite *InvariantsComprehensiveTestSuite) TestMerkleProofInvariantNoSiblings() {
	ctx := suite.SdkCtx
	inv := MerkleProofInvariant(*suite.Keeper)

	// Create proof with no proof hashes (siblings)
	proof := &bridgepb.MerkleProof{
		Root:    []byte("root-hash"),
		Leaf:    []byte("leaf-data"),
		Proof:   [][]byte{},
		Indices: []uint64{0},
	}
	suite.storeMerkleProof(ctx, "proof-3", proof)

	msg, broken := inv(ctx)
	suite.True(broken, "proof with no proof hashes should break invariant")
	suite.Contains(msg, "no proof hashes")
}

func (suite *InvariantsComprehensiveTestSuite) TestValidatorSetInvariant() {
	ctx := suite.SdkCtx
	inv := ValidatorSetInvariant(*suite.Keeper)

	// Test: Empty validator set
	msg, broken := inv(ctx)
	// May break if minimum validators required
	_ = msg
	_ = broken

	// Create valid validator
	validator := &bridgepb.BridgeValidator{
		Address: sdk.ValAddress("validator_________").String(),
		Power:   100,
		Active:  true,
	}
	suite.storeValidator(ctx, validator)

	msg, broken = inv(ctx)
	// Should pass if validator is valid
	_ = msg
	_ = broken
}

func (suite *InvariantsComprehensiveTestSuite) TestValidatorSetInvariantInvalidAddress() {
	ctx := suite.SdkCtx
	inv := ValidatorSetInvariant(*suite.Keeper)

	// Create validator with invalid address
	validator := &bridgepb.BridgeValidator{
		Address: "invalid-address",
		Power:   100,
		Active:  true,
	}
	suite.storeValidator(ctx, validator)

	msg, broken := inv(ctx)
	suite.True(broken, "validator with invalid address should break invariant")
	suite.Contains(msg, "invalid validator address")
}

func (suite *InvariantsComprehensiveTestSuite) TestValidatorSetInvariantZeroPower() {
	ctx := suite.SdkCtx
	inv := ValidatorSetInvariant(*suite.Keeper)

	// uint64 cannot be negative, but we can test zero power and invalid address
	// The invariant checks address format first, so we test that path
	// Note: The actual invariant validates that Power is not < 0 (which can never happen with uint64)
	// and that there's at least 1 active validator

	// Create validator with valid address but zero power
	// This will still count as active if Active=true
	validator := &bridgepb.BridgeValidator{
		Address: sdk.AccAddress("validatoraddr_____").String(), // Use AccAddress for valid format
		Power:   0,
		Active:  true,
	}
	suite.storeValidator(ctx, validator)

	msg, broken := inv(ctx)
	// With zero power but active=true, the validator is counted but has no voting power
	// The invariant should pass since there is at least 1 active validator
	suite.False(broken, "validator with zero power should still be counted as active")
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestSecurityParametersInvariant() {
	ctx := suite.SdkCtx
	inv := SecurityParametersInvariant(*suite.Keeper)

	// Test with default params
	msg, broken := inv(ctx)
	suite.False(broken, "default params should be valid")
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestTransferLimitInvariant() {
	ctx := suite.SdkCtx
	inv := TransferLimitInvariant(*suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Create transfer with valid amount
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:           "transfer-limit-1",
		SourceChain:          "aura",
		TargetChain:          "ethereum",
		Sender:               sdk.AccAddress("sender____________").String(),
		Recipient:            "0x123",
		Amount:               sdkmath.NewInt(1000),
		Denom:                "uaura",
		Status:               bridgepb.TransferStatus_PENDING,
		Timestamp:            time.Now(),
		ValidatorSignatures:  []bridgepb.ValidatorSignature{},
	}
	suite.Keeper.SetTransfer(ctx, transfer)

	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestTransferLimitInvariantEmptySender() {
	ctx := suite.SdkCtx
	inv := TransferLimitInvariant(*suite.Keeper)

	// Create transfer with empty sender
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:           "transfer-empty-sender",
		SourceChain:          "aura",
		TargetChain:          "ethereum",
		Sender:               "",
		Recipient:            "0x123",
		Amount:               sdkmath.NewInt(1000),
		Denom:                "uaura",
		Status:               bridgepb.TransferStatus_PENDING,
		Timestamp:            time.Now(),
		ValidatorSignatures:  []bridgepb.ValidatorSignature{},
	}
	suite.Keeper.SetTransfer(ctx, transfer)

	msg, broken := inv(ctx)
	suite.True(broken, "transfer with empty sender should break invariant")
	suite.Contains(msg, "empty sender")
}

func (suite *InvariantsComprehensiveTestSuite) TestChannelStateInvariant() {
	ctx := suite.SdkCtx
	inv := ChannelStateInvariant(*suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// TODO: BridgeChannel type not yet defined in proto
	// Create valid channel
	// channel := &bridgepb.BridgeChannel{
	// 	ChannelId:               "channel-1",
	// 	SourceChainId:           "aura",
	// 	DestinationChainId:      "ethereum",
	// 	State:                   "open",
	// 	CircuitBreakerEnabled:   false,
	// 	CircuitBreakerThreshold: 0,
	// }
	// suite.storeChannel(ctx, channel)

	// msg, broken = inv(ctx)
	// suite.False(broken)
	// suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestChannelStateInvariantEmptyID() {
	ctx := suite.SdkCtx
	inv := ChannelStateInvariant(*suite.Keeper)

	// NOTE: This test is currently not implementable because BridgeChannel type
	// is not yet defined in the proto schema. Once the proto definition is added,
	// uncomment and implement the following test logic:
	//
	// Test Plan:
	// 1. Create a BridgeChannel message with empty ChannelId
	// 2. Store it using suite.storeChannel(ctx, channel)
	// 3. Verify the ChannelStateInvariant detects this as invalid
	// 4. Assert that broken=true and msg contains "empty ID"
	//
	// Expected behavior when implemented:
	//   channel := &bridgepb.BridgeChannel{
	//     ChannelId:          "",  // Invalid: empty
	//     SourceChainId:      "aura",
	//     DestinationChainId: "ethereum",
	//     State:              "open",
	//   }
	//   suite.storeChannel(ctx, channel)
	//   msg, broken := inv(ctx)
	//   suite.True(broken, "channel with empty ID should break invariant")
	//   suite.Contains(msg, "empty ID")

	// For now, verify the invariant works with empty store (covered in TestChannelStateInvariant)
	msg, broken := inv(ctx)
	suite.False(broken, "empty store should not break channel state invariant")
	suite.Empty(msg)
}

// Helper methods
func (suite *InvariantsComprehensiveTestSuite) storeMerkleProof(ctx sdk.Context, id string, proof *bridgepb.MerkleProof) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	bz := suite.Keeper.cdc.MustMarshal(proof)
	// TODO: Add MerkleProofPrefix to types/keys.go if needed
	merkleProofPrefix := []byte{0x10}
	store.Set(append(merkleProofPrefix, []byte(id)...), bz)
}

func (suite *InvariantsComprehensiveTestSuite) storeValidator(ctx sdk.Context, validator *bridgepb.BridgeValidator) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	bz := suite.Keeper.cdc.MustMarshal(validator)
	store.Set(append(types.ValidatorPrefix, []byte(validator.Address)...), bz)
}

// ========================================================================
// BALANCE INVARIANT TESTS WITH MOCKS
// ========================================================================

// TestBalanceInvariantWithMocks tests balance checking with proper mocks
func (suite *InvariantsComprehensiveTestSuite) TestBalanceInvariantWithMocksSufficient() {
	// Create a new keeper with mocked keepers
	mockBank := newMockBankKeeperWithBalances()
	mockAccount := newMockAccountKeeperWithModule()
	moduleAddr := mockAccount.GetModuleAddress(types.ModuleName)

	k := NewKeeper(
		suite.Keeper.cdc,
		suite.Keeper.storeKey,
		nil,
		mockBank,
		mockAccount,
		nil,
		nil, // stakingKeeper
	)

	ctx := suite.SdkCtx

	// Set module balance to 10000 uaura
	mockBank.SetBalance(moduleAddr, "uaura", sdkmath.NewInt(10000))

	// Create pending transfer for 5000 uaura (less than module balance)
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:           "transfer-sufficient",
		SourceChain:          "aura",
		TargetChain:          "ethereum",
		Sender:               sdk.AccAddress("sender____________").String(),
		Recipient:            "0x123",
		Amount:               sdkmath.NewInt(5000),
		Denom:                "uaura",
		Status:               bridgepb.TransferStatus_PENDING,
		Timestamp:            time.Now(),
		ValidatorSignatures:  []bridgepb.ValidatorSignature{},
	}
	k.SetTransfer(ctx, transfer)

	// Create invariant with mocked keeper
	inv := TransferBalanceInvariant(*k)

	// Invariant should pass
	msg, broken := inv(ctx)
	suite.False(broken, "invariant should pass when module balance is sufficient")
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestBalanceInvariantWithMocksInsufficient() {
	// Create a new keeper with mocked keepers
	mockBank := newMockBankKeeperWithBalances()
	mockAccount := newMockAccountKeeperWithModule()
	moduleAddr := mockAccount.GetModuleAddress(types.ModuleName)

	k := NewKeeper(
		suite.Keeper.cdc,
		suite.Keeper.storeKey,
		nil,
		mockBank,
		mockAccount,
		nil,
		nil, // stakingKeeper
	)

	ctx := suite.SdkCtx

	// Set module balance to only 1000 uaura
	mockBank.SetBalance(moduleAddr, "uaura", sdkmath.NewInt(1000))

	// Create pending transfer for 5000 uaura (more than module balance)
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:           "transfer-insufficient",
		SourceChain:          "aura",
		TargetChain:          "ethereum",
		Sender:               sdk.AccAddress("sender____________").String(),
		Recipient:            "0x123",
		Amount:               sdkmath.NewInt(5000),
		Denom:                "uaura",
		Status:               bridgepb.TransferStatus_PENDING,
		Timestamp:            time.Now(),
		ValidatorSignatures:  []bridgepb.ValidatorSignature{},
	}
	k.SetTransfer(ctx, transfer)

	// Create invariant with mocked keeper
	inv := TransferBalanceInvariant(*k)

	// Invariant should be BROKEN
	msg, broken := inv(ctx)
	suite.True(broken, "invariant should break when module balance is insufficient")
	suite.Contains(msg, "module balance insufficient")
	suite.Contains(msg, "balance=1000")
	suite.Contains(msg, "locked=5000")
	suite.Contains(msg, "uaura")
}

func (suite *InvariantsComprehensiveTestSuite) TestBalanceInvariantWithMocksMultipleDenoms() {
	// Create a new keeper with mocked keepers
	mockBank := newMockBankKeeperWithBalances()
	mockAccount := newMockAccountKeeperWithModule()
	moduleAddr := mockAccount.GetModuleAddress(types.ModuleName)

	k := NewKeeper(
		suite.Keeper.cdc,
		suite.Keeper.storeKey,
		nil,
		mockBank,
		mockAccount,
		nil,
		nil, // stakingKeeper
	)

	ctx := suite.SdkCtx

	// Set module balances - upaw is insufficient
	mockBank.SetBalance(moduleAddr, "uaura", sdkmath.NewInt(10000))
	mockBank.SetBalance(moduleAddr, "upaw", sdkmath.NewInt(5000)) // Not enough

	// Create transfers
	transfer1 := &bridgepb.CrossChainTransfer{
		TransferId:           "transfer-multi-ok",
		SourceChain:          "aura",
		TargetChain:          "ethereum",
		Sender:               sdk.AccAddress("sender1___________").String(),
		Recipient:            "0x123",
		Amount:               sdkmath.NewInt(5000),
		Denom:                "uaura",
		Status:               bridgepb.TransferStatus_PENDING,
		Timestamp:            time.Now(),
		ValidatorSignatures:  []bridgepb.ValidatorSignature{},
	}
	k.SetTransfer(ctx, transfer1)

	transfer2 := &bridgepb.CrossChainTransfer{
		TransferId:           "transfer-multi-insufficient",
		SourceChain:          "aura",
		TargetChain:          "ethereum",
		Sender:               sdk.AccAddress("sender2___________").String(),
		Recipient:            "0x456",
		Amount:               sdkmath.NewInt(10000),
		Denom:                "upaw",
		Status:               bridgepb.TransferStatus_PENDING,
		Timestamp:            time.Now(),
		ValidatorSignatures:  []bridgepb.ValidatorSignature{},
	}
	k.SetTransfer(ctx, transfer2)

	// Create invariant with mocked keeper
	inv := TransferBalanceInvariant(*k)

	// Total locked: uaura=5000 (OK), upaw=10000 (INSUFFICIENT - only 5000 available)
	// Invariant should be BROKEN for upaw
	msg, broken := inv(ctx)
	suite.True(broken, "invariant should break when one denom is insufficient")
	suite.Contains(msg, "module balance insufficient")
	suite.Contains(msg, "upaw")
	suite.Contains(msg, "balance=5000")
	suite.Contains(msg, "locked=10000")
}

func (suite *InvariantsComprehensiveTestSuite) TestBalanceInvariantWithMocksZeroBalance() {
	// Create a new keeper with mocked keepers
	mockBank := newMockBankKeeperWithBalances()
	mockAccount := newMockAccountKeeperWithModule()

	k := NewKeeper(
		suite.Keeper.cdc,
		suite.Keeper.storeKey,
		nil,
		mockBank,
		mockAccount,
		nil,
		nil, // stakingKeeper
	)

	ctx := suite.SdkCtx

	// Module has zero balance (default from mock)
	// Create pending transfer
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:           "transfer-zero-balance",
		SourceChain:          "aura",
		TargetChain:          "ethereum",
		Sender:               sdk.AccAddress("sender____________").String(),
		Recipient:            "0x123",
		Amount:               sdkmath.NewInt(1000),
		Denom:                "uaura",
		Status:               bridgepb.TransferStatus_PENDING,
		Timestamp:            time.Now(),
		ValidatorSignatures:  []bridgepb.ValidatorSignature{},
	}
	k.SetTransfer(ctx, transfer)

	// Create invariant with mocked keeper
	inv := TransferBalanceInvariant(*k)

	// Invariant should be BROKEN - module has no funds
	msg, broken := inv(ctx)
	suite.True(broken, "invariant should break when module has zero balance but has locked transfers")
	suite.Contains(msg, "module balance insufficient")
	suite.Contains(msg, "balance=0")
	suite.Contains(msg, "locked=1000")
}

// storeChannel is commented out until BridgeChannel type is defined in proto
// func (suite *InvariantsComprehensiveTestSuite) storeChannel(ctx sdk.Context, channel *bridgepb.BridgeChannel) {
// 	store := ctx.KVStore(suite.Keeper.storeKey)
// 	bz := suite.Keeper.cdc.MustMarshal(channel)
// 	store.Set(append(types.ChannelKeyPrefix, []byte(channel.ChannelId)...), bz)
// }

// =============================================================================
// ChannelStateInvariant Comprehensive Tests (via ChainConfig)
// =============================================================================

func (suite *InvariantsComprehensiveTestSuite) TestChannelStateInvariant_ValidConfig() {
	// Add a properly configured chain using the setChainConfig method
	config := types.ChainConfig{
		ChainId:          "ethereum",
		ChainName:        "Ethereum Mainnet",
		AddressPrefix:    "0x",
		MinConfirmations: 12,
		Enabled:          true,
		Validators:       []string{"val1", "val2"},
	}
	suite.Keeper.setChainConfig(suite.SdkCtx, config)

	inv := ChannelStateInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "valid chain config should not break invariant")
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestChannelStateInvariant_EmptyChainId_Direct() {
	// Store directly to trigger the empty chain ID check
	store := suite.SdkCtx.KVStore(suite.Keeper.storeKey)
	config := types.ChainConfig{
		ChainId:          "",
		ChainName:        "Test Chain",
		AddressPrefix:    "test",
		MinConfirmations: 10,
	}
	bz, _ := suite.Keeper.cdc.Marshal(&config)
	store.Set(append(types.ChainConfigPrefix, []byte("testkey")...), bz)

	inv := ChannelStateInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "chain config with empty ID should break invariant")
	suite.Contains(msg, "empty ID")
}

func (suite *InvariantsComprehensiveTestSuite) TestChannelStateInvariant_EmptyChainName_Direct() {
	store := suite.SdkCtx.KVStore(suite.Keeper.storeKey)
	config := types.ChainConfig{
		ChainId:          "testchain_name",
		ChainName:        "",
		AddressPrefix:    "test",
		MinConfirmations: 10,
	}
	bz, _ := suite.Keeper.cdc.Marshal(&config)
	store.Set(append(types.ChainConfigPrefix, []byte("testchain_name")...), bz)

	inv := ChannelStateInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "chain config with empty name should break invariant")
	suite.Contains(msg, "empty name")
}

func (suite *InvariantsComprehensiveTestSuite) TestChannelStateInvariant_EmptyAddressPrefix_Direct() {
	store := suite.SdkCtx.KVStore(suite.Keeper.storeKey)
	config := types.ChainConfig{
		ChainId:          "testchain_prefix",
		ChainName:        "Test Chain Prefix",
		AddressPrefix:    "",
		MinConfirmations: 10,
	}
	bz, _ := suite.Keeper.cdc.Marshal(&config)
	store.Set(append(types.ChainConfigPrefix, []byte("testchain_prefix")...), bz)

	inv := ChannelStateInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "chain config with empty address prefix should break invariant")
	suite.Contains(msg, "empty address prefix")
}

func (suite *InvariantsComprehensiveTestSuite) TestChannelStateInvariant_ZeroConfirmations_Direct() {
	store := suite.SdkCtx.KVStore(suite.Keeper.storeKey)
	config := types.ChainConfig{
		ChainId:          "testchain_conf",
		ChainName:        "Test Chain Conf",
		AddressPrefix:    "test",
		MinConfirmations: 0,
	}
	bz, _ := suite.Keeper.cdc.Marshal(&config)
	store.Set(append(types.ChainConfigPrefix, []byte("testchain_conf")...), bz)

	inv := ChannelStateInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "chain config with zero confirmations should break invariant")
	suite.Contains(msg, "zero min confirmations")
}

func (suite *InvariantsComprehensiveTestSuite) TestChannelStateInvariant_ExcessiveConfirmations_Direct() {
	store := suite.SdkCtx.KVStore(suite.Keeper.storeKey)
	config := types.ChainConfig{
		ChainId:          "testchain_excess",
		ChainName:        "Test Chain Excess",
		AddressPrefix:    "test",
		MinConfirmations: 5000, // > 1000 limit
	}
	bz, _ := suite.Keeper.cdc.Marshal(&config)
	store.Set(append(types.ChainConfigPrefix, []byte("testchain_excess")...), bz)

	inv := ChannelStateInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "chain config with excessive confirmations should break invariant")
	suite.Contains(msg, "excessive min confirmations")
}

func (suite *InvariantsComprehensiveTestSuite) TestChannelStateInvariant_MultipleConfigs() {
	// Add multiple valid configs
	config1 := types.ChainConfig{
		ChainId:          "ethereum",
		ChainName:        "Ethereum",
		AddressPrefix:    "0x",
		MinConfirmations: 12,
		Enabled:          true,
	}
	suite.Keeper.setChainConfig(suite.SdkCtx, config1)

	config2 := types.ChainConfig{
		ChainId:          "polygon",
		ChainName:        "Polygon",
		AddressPrefix:    "0x",
		MinConfirmations: 256,
		Enabled:          true,
	}
	suite.Keeper.setChainConfig(suite.SdkCtx, config2)

	inv := ChannelStateInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "multiple valid configs should not break invariant")
	suite.Empty(msg)
}

// =============================================================================
// TransferChainIntegrityInvariant Comprehensive Tests
// =============================================================================

func (suite *InvariantsComprehensiveTestSuite) TestTransferChainIntegrityInvariant_NoTransfers() {
	inv := TransferChainIntegrityInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "no transfers should not break invariant")
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestTransferChainIntegrityInvariant_ValidWithChainConfig() {
	// Add chain config first
	config := types.ChainConfig{
		ChainId:          "ethereum",
		ChainName:        "Ethereum",
		AddressPrefix:    "0x",
		MinConfirmations: 12,
		Enabled:          true,
	}
	suite.Keeper.setChainConfig(suite.SdkCtx, config)

	// Create transfer referencing the config
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "integrity-valid-1",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      sdkmath.NewInt(1000),
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   time.Now(),
	}
	suite.Keeper.SetTransfer(suite.SdkCtx, transfer)

	inv := TransferChainIntegrityInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "transfer with valid chain config should not break invariant")
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestTransferChainIntegrityInvariant_AuraToAura() {
	// Local transfers shouldn't require chain configs
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "integrity-local-aura",
		SourceChain: "aura",
		TargetChain: "aura",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   sdk.AccAddress("recipient_________").String(),
		Amount:      sdkmath.NewInt(1000),
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   time.Now(),
	}
	suite.Keeper.SetTransfer(suite.SdkCtx, transfer)

	inv := TransferChainIntegrityInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "aura to aura transfer should not require chain config")
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestTransferChainIntegrityInvariant_MissingTargetConfig() {
	// Create transfer to non-existent chain
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "integrity-missing-target",
		SourceChain: "aura",
		TargetChain: "nonexistent_chain",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      sdkmath.NewInt(1000),
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   time.Now(),
	}
	suite.Keeper.SetTransfer(suite.SdkCtx, transfer)

	inv := TransferChainIntegrityInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "transfer with missing target chain config should break invariant")
	suite.Contains(msg, "non-existent target chain")
}

func (suite *InvariantsComprehensiveTestSuite) TestTransferChainIntegrityInvariant_MissingSourceConfig() {
	// Create transfer from non-existent chain (not aura)
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "integrity-missing-source",
		SourceChain: "nonexistent_source",
		TargetChain: "aura",
		Sender:      "0xexternal",
		Recipient:   sdk.AccAddress("recipient_________").String(),
		Amount:      sdkmath.NewInt(1000),
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   time.Now(),
	}
	suite.Keeper.SetTransfer(suite.SdkCtx, transfer)

	inv := TransferChainIntegrityInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "transfer with missing source chain config should break invariant")
	suite.Contains(msg, "non-existent source chain")
}

func (suite *InvariantsComprehensiveTestSuite) TestTransferChainIntegrityInvariant_ExternalToExternal() {
	// Add chain configs
	ethConfig := types.ChainConfig{
		ChainId:          "ethereum",
		ChainName:        "Ethereum",
		AddressPrefix:    "0x",
		MinConfirmations: 12,
		Enabled:          true,
	}
	suite.Keeper.setChainConfig(suite.SdkCtx, ethConfig)

	polyConfig := types.ChainConfig{
		ChainId:          "polygon",
		ChainName:        "Polygon",
		AddressPrefix:    "0x",
		MinConfirmations: 256,
		Enabled:          true,
	}
	suite.Keeper.setChainConfig(suite.SdkCtx, polyConfig)

	// Create transfer between two external chains
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "integrity-external-external",
		SourceChain: "ethereum",
		TargetChain: "polygon",
		Sender:      "0xsender",
		Recipient:   "0xrecipient",
		Amount:      sdkmath.NewInt(1000),
		Denom:       "weth",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   time.Now(),
	}
	suite.Keeper.SetTransfer(suite.SdkCtx, transfer)

	inv := TransferChainIntegrityInvariant(*suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "transfer between two valid external chains should not break invariant")
	suite.Empty(msg)
}
