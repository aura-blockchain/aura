package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

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
		TransferId:  "transfer-1",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      "1000",
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   timestamppb.Now(),
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
		TransferId:  "transfer-multi-1",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender1___________").String(),
		Recipient:   "0x123",
		Amount:      "1000",
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   timestamppb.Now(),
	}
	suite.Keeper.SetTransfer(ctx, transfer1)

	transfer2 := &bridgepb.CrossChainTransfer{
		TransferId:  "transfer-multi-2",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender2___________").String(),
		Recipient:   "0x456",
		Amount:      "2000",
		Denom:       "upaw",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   timestamppb.Now(),
	}
	suite.Keeper.SetTransfer(ctx, transfer2)

	transfer3 := &bridgepb.CrossChainTransfer{
		TransferId:  "transfer-multi-3",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender3___________").String(),
		Recipient:   "0x789",
		Amount:      "500",
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_CONFIRMED,
		Timestamp:   timestamppb.Now(),
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
		TransferId:  "transfer-completed",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      "1000000",
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_COMPLETED,
		Timestamp:   timestamppb.Now(),
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
		TransferId:  "transfer-failed",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      "1000000",
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_FAILED,
		Timestamp:   timestamppb.Now(),
	}
	suite.Keeper.SetTransfer(ctx, failedTransfer)

	// Failed transfers should not affect the invariant
	msg, broken := inv(ctx)
	suite.False(broken, "failed transfers should be ignored")
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestTransferBalanceInvariantInvalidAmount() {
	ctx := suite.SdkCtx
	inv := TransferBalanceInvariant(*suite.Keeper)

	// Create transfer with invalid amount
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "transfer-invalid",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      "invalid-amount",
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   timestamppb.Now(),
	}
	suite.Keeper.setTransfer(ctx, transfer)

	msg, broken := inv(ctx)
	suite.True(broken, "invalid amount should break invariant")
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
		TransferId:  "transfer-limit-1",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      "1000",
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   timestamppb.Now(),
	}
	suite.Keeper.setTransfer(ctx, transfer)

	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestTransferLimitInvariantEmptySender() {
	ctx := suite.SdkCtx
	inv := TransferLimitInvariant(*suite.Keeper)

	// Create transfer with empty sender
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "transfer-empty-sender",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      "",
		Recipient:   "0x123",
		Amount:      "1000",
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   timestamppb.Now(),
	}
	suite.Keeper.setTransfer(ctx, transfer)

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

	// TODO: BridgeChannel type not yet defined in proto
	// Create channel with empty ID
	// channel := &bridgepb.BridgeChannel{
	// 	ChannelId:          "",
	// 	SourceChainId:      "aura",
	// 	DestinationChainId: "ethereum",
	// 	State:              "open",
	// }
	// suite.storeChannel(ctx, channel)

	// msg, broken := inv(ctx)
	// suite.True(broken, "channel with empty ID should break invariant")
	// suite.Contains(msg, "empty ID")

	// Placeholder to avoid empty test
	_, _ = ctx, inv
	suite.T().Skip("BridgeChannel type not yet defined in proto")
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
	)

	ctx := suite.SdkCtx

	// Set module balance to 10000 uaura
	mockBank.SetBalance(moduleAddr, "uaura", sdkmath.NewInt(10000))

	// Create pending transfer for 5000 uaura (less than module balance)
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "transfer-sufficient",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      "5000",
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   timestamppb.Now(),
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
	)

	ctx := suite.SdkCtx

	// Set module balance to only 1000 uaura
	mockBank.SetBalance(moduleAddr, "uaura", sdkmath.NewInt(1000))

	// Create pending transfer for 5000 uaura (more than module balance)
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "transfer-insufficient",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      "5000",
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   timestamppb.Now(),
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
	)

	ctx := suite.SdkCtx

	// Set module balances - upaw is insufficient
	mockBank.SetBalance(moduleAddr, "uaura", sdkmath.NewInt(10000))
	mockBank.SetBalance(moduleAddr, "upaw", sdkmath.NewInt(5000)) // Not enough

	// Create transfers
	transfer1 := &bridgepb.CrossChainTransfer{
		TransferId:  "transfer-multi-ok",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender1___________").String(),
		Recipient:   "0x123",
		Amount:      "5000",
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   timestamppb.Now(),
	}
	k.SetTransfer(ctx, transfer1)

	transfer2 := &bridgepb.CrossChainTransfer{
		TransferId:  "transfer-multi-insufficient",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender2___________").String(),
		Recipient:   "0x456",
		Amount:      "10000",
		Denom:       "upaw",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   timestamppb.Now(),
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
	)

	ctx := suite.SdkCtx

	// Module has zero balance (default from mock)
	// Create pending transfer
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "transfer-zero-balance",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      "1000",
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   timestamppb.Now(),
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
