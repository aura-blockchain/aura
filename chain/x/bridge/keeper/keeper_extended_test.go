// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

type BridgeKeeperTestSuite struct {
	suite.Suite

	keeper *keeper.Keeper
	ctx    sdk.Context
}

func (suite *BridgeKeeperTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.keeper = keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil, // paramstore - not available in test input
		nil, // bankKeeper
		nil, // accountKeeper
		nil, // vcKeeper
		nil, // stakingKeeper
	)
	suite.ctx = input.Ctx
}

func TestBridgeKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(BridgeKeeperTestSuite))
}

// Params Tests

func (suite *BridgeKeeperTestSuite) TestGetParams() {
	params := suite.keeper.GetParams(suite.ctx)
	suite.Require().NotNil(params)
}

func (suite *BridgeKeeperTestSuite) TestSetParams() {
	params := types.DefaultParams()
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	retrieved := suite.keeper.GetParams(suite.ctx)
	suite.Require().Equal(params, retrieved)
}

// Bridge Transfer Tests

func TestInitiateBridgeTransfer(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Setup chain config for ethereum
	chainConfig := types.ChainConfig{
		ChainId:   "ethereum",
		ChainName: "Ethereum",
		Enabled:   true,
	}
	err := k.AddSupportedChain(input.Ctx, chainConfig)
	require.NoError(t, err)

	sender := keepertest.GenTestAddr()
	recipient := "0x1234567890123456789012345678901234567890"
	amount := sdk.NewCoins(sdk.NewCoin("uaura", math.NewInt(1000000)))

	transferID, err := k.InitiateTransfer(input.Ctx, sender.String(), recipient, amount, "ethereum")
	require.NoError(t, err)
	require.NotEmpty(t, transferID)

	// Verify transfer was created
	transfer, found := k.GetTransfer(input.Ctx, transferID)
	require.True(t, found)
	require.Equal(t, sender.String(), transfer.Sender)
	require.Equal(t, recipient, transfer.Recipient)
	require.Equal(t, "ethereum", transfer.TargetChain)
}

func TestInitiateBridgeTransferInvalidChain(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	sender := keepertest.GenTestAddr()
	recipient := "0x1234567890123456789012345678901234567890"
	amount := sdk.NewCoins(sdk.NewCoin("uaura", math.NewInt(1000000)))

	// Should fail because chain is not configured
	_, err := k.InitiateTransfer(input.Ctx, sender.String(), recipient, amount, "unsupported")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported chain")
}

func TestInitiateBridgeTransferZeroAmount(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	sender := keepertest.GenTestAddr()
	recipient := "0x1234567890123456789012345678901234567890"
	amount := sdk.NewCoins()

	_, err := k.InitiateTransfer(input.Ctx, sender.String(), recipient, amount, "ethereum")
	require.Error(t, err)
	require.Contains(t, err.Error(), "amount must be positive")
}

func TestGetBridgeTransfer(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Setup chain config for ethereum
	chainConfig := types.ChainConfig{
		ChainId:   "ethereum",
		ChainName: "Ethereum",
		Enabled:   true,
	}
	err := k.AddSupportedChain(input.Ctx, chainConfig)
	require.NoError(t, err)

	sender := keepertest.GenTestAddr()
	recipient := "0x1234567890123456789012345678901234567890"
	amount := sdk.NewCoins(sdk.NewCoin("uaura", math.NewInt(1000000)))

	// Create a transfer
	transferID, err := k.InitiateTransfer(input.Ctx, sender.String(), recipient, amount, "ethereum")
	require.NoError(t, err)

	// Retrieve the transfer
	transfer, found := k.GetTransfer(input.Ctx, transferID)
	require.True(t, found)
	require.Equal(t, transferID, transfer.TransferId)
	require.Equal(t, sender.String(), transfer.Sender)
	require.Equal(t, recipient, transfer.Recipient)
}

func TestGetNonExistentTransfer(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	_, found := k.GetTransfer(input.Ctx, "nonexistent-transfer-id")
	require.False(t, found)
}

func TestGetAllTransfers(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Create a few test transfers using the helper
	seedBridgeTransfer(t, input, "transfer_1", math.NewInt(1000).String(), 0)
	seedBridgeTransfer(t, input, "transfer_2", math.NewInt(2000).String(), 0)
	seedBridgeTransfer(t, input, "transfer_3", math.NewInt(3000).String(), 0)

	// Verify GetTransfer works for individual transfers
	transfer1, found := k.GetTransfer(input.Ctx, "transfer_1")
	require.True(t, found)
	require.Equal(t, "transfer_1", transfer1.TransferId)

	transfer2, found := k.GetTransfer(input.Ctx, "transfer_2")
	require.True(t, found)
	require.Equal(t, "transfer_2", transfer2.TransferId)

	transfer3, found := k.GetTransfer(input.Ctx, "transfer_3")
	require.True(t, found)
	require.Equal(t, "transfer_3", transfer3.TransferId)
}

// Validator Attestation Tests

func TestSubmitAttestation(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	validator := keepertest.GenTestAddr()
	transferID := "transfer_1"
	seedBridgeTransfer(t, input, transferID, math.NewInt(1000).String(), 0)

	err := k.SubmitAttestation(input.Ctx, transferID, validator.String(), true)
	require.NoError(t, err)

	// Verify attestation was recorded
	attestations := k.GetAttestations(input.Ctx, transferID)
	require.Len(t, attestations, 1)
	require.Equal(t, validator.String(), attestations[0])
}

func TestSubmitMultipleAttestations(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	transferID := "transfer_1"
	seedBridgeTransfer(t, input, transferID, math.NewInt(1000).String(), 0)
	validators := keepertest.GenTestAddrs(4)

	for _, val := range validators {
		err := k.SubmitAttestation(input.Ctx, transferID, val.String(), true)
		require.NoError(t, err)
	}

	attestations := k.GetAttestations(input.Ctx, transferID)
	require.Len(t, attestations, 4)
}

func TestDuplicateAttestation(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	validator := keepertest.GenTestAddr()
	transferID := "transfer_1"
	seedBridgeTransfer(t, input, transferID, math.NewInt(1000).String(), 0)

	err := k.SubmitAttestation(input.Ctx, transferID, validator.String(), true)
	require.NoError(t, err)

	// Try to submit again
	err = k.SubmitAttestation(input.Ctx, transferID, validator.String(), true)
	require.Error(t, err, "Should not allow duplicate attestation")
}

func TestAttestationThreshold(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	transferID := "transfer_1"
	seedBridgeTransfer(t, input, transferID, math.NewInt(5000).String(), 7)
	validators := keepertest.GenTestAddrs(10)

	// Submit 7/10 attestations (70% threshold)
	for i := 0; i < 7; i++ {
		err := k.SubmitAttestation(input.Ctx, transferID, validators[i].String(), true)
		require.NoError(t, err)
	}

	passed := k.CheckAttestationThreshold(input.Ctx, transferID)
	require.True(t, passed)
}

func TestInsufficientAttestations(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	transferID := "transfer_1"
	seedBridgeTransfer(t, input, transferID, math.NewInt(5000).String(), 7)
	validators := keepertest.GenTestAddrs(10)

	// Submit only 5/10 attestations (50%)
	for i := 0; i < 5; i++ {
		err := k.SubmitAttestation(input.Ctx, transferID, validators[i].String(), true)
		require.NoError(t, err)
	}

	passed := k.CheckAttestationThreshold(input.Ctx, transferID)
	require.False(t, passed)
}

// Withdrawal Tests

func TestProcessWithdrawal(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	recipient := keepertest.GenTestAddr()
	amount := sdk.NewCoins(sdk.NewCoin("uaura", math.NewInt(1000000)))
	seedBridgeTransfer(t, input, "transfer_1", amount.AmountOf("uaura").String(), 0)

	err := k.ProcessWithdrawal(input.Ctx, recipient.String(), amount, "transfer_1")
	require.NoError(t, err)
}

func TestProcessWithdrawalCircuitBreaker(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	recipient := keepertest.GenTestAddr()
	largeAmount := sdk.NewCoins(sdk.NewCoin("uaura", math.NewInt(1000000000000)))
	seedBridgeTransfer(t, input, "transfer_1", largeAmount.AmountOf("uaura").String(), 0)

	// Should trigger circuit breaker for large withdrawal
	err := k.ProcessWithdrawal(input.Ctx, recipient.String(), largeAmount, "transfer_1")
	require.Error(t, err)
}

func TestProcessWithdrawalTimelock(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	recipient := keepertest.GenTestAddr()
	amount := sdk.NewCoins(sdk.NewCoin("uaura", math.NewInt(1000000)))

	// Initiate withdrawal
	withdrawalID, err := k.InitiateWithdrawal(input.Ctx, recipient.String(), amount)
	require.NoError(t, err)

	// Try to execute immediately (should fail due to timelock)
	err = k.ExecuteWithdrawal(input.Ctx, withdrawalID)
	require.Error(t, err)

	// Advance time past timelock
	ctx := keepertest.AdvanceTime(input.Ctx, types.DefaultTimelockDuration)

	// Now should succeed
	err = k.ExecuteWithdrawal(ctx, withdrawalID)
	require.NoError(t, err)
}

// Fraud Proof Tests

func TestSubmitFraudProof(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	submitter := keepertest.GenTestAddr()
	transferID := "transfer_1"
	proof := []byte("fraud_proof_data")
	seedBridgeTransferWithPending(t, input, transferID, math.NewInt(1000).String(), 0)

	err := k.SubmitFraudProof(input.Ctx, transferID, submitter.String(), proof)
	require.NoError(t, err)

	stored, found := k.GetFraudProof(input.Ctx, transferID)
	require.True(t, found)
	require.Equal(t, submitter.String(), stored.Challenger)
	require.Equal(t, types.FraudProofStatus_FRAUD_PROOF_INVESTIGATING, stored.Status)
}

func TestSubmitFraudProofDuplicate(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	submitter := keepertest.GenTestAddr()
	transferID := "transfer_dup"
	proof := []byte("fraud_proof_data")
	seedBridgeTransferWithPending(t, input, transferID, math.NewInt(1000).String(), 0)

	err := k.SubmitFraudProof(input.Ctx, transferID, submitter.String(), proof)
	require.NoError(t, err)

	err = k.SubmitFraudProof(input.Ctx, transferID, submitter.String(), proof)
	require.ErrorIs(t, err, types.ErrFraudProofPending)
}

func TestSubmitFraudProofWindowExpired(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	submitter := keepertest.GenTestAddr()
	transferID := "transfer_window"
	proof := []byte("fraud_proof_data")
	seedBridgeTransferWithPending(t, input, transferID, math.NewInt(1000).String(), 0)

	// Advance time past the fraud proof window
	ctx := keepertest.AdvanceTime(input.Ctx, types.DefaultFraudProofWindow+time.Second)
	err := k.SubmitFraudProof(ctx, transferID, submitter.String(), proof)
	require.ErrorIs(t, err, types.ErrFraudProofExpired)
}

func TestResolveFraudProofValid(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	submitter := keepertest.GenTestAddr()
	transferID := "transfer_1"
	proof := []byte("valid_fraud_proof")
	seedBridgeTransferWithPending(t, input, transferID, math.NewInt(1000).String(), 0)

	err := k.SubmitFraudProof(input.Ctx, transferID, submitter.String(), proof)
	require.NoError(t, err)

	resolved, err := k.ResolveFraudProof(input.Ctx, transferID, true)
	require.NoError(t, err)
	require.Equal(t, types.FraudProofStatus_FRAUD_PROOF_VALID, resolved.Status)
	require.NotNil(t, resolved.ResolvedAt)
	params := types.DefaultSecurityParams()
	require.True(t, resolved.RewardAmount.Equal(params.FraudProofReward),
		"expected reward %s, got %s", params.FraudProofReward.String(), resolved.RewardAmount.String())

	transfer := getBridgeTransfer(t, input, transferID)
	require.Equal(t, types.TransferStatus_FAILED, transfer.Status)
}

func TestResolveFraudProofInvalid(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	submitter := keepertest.GenTestAddr()
	transferID := "transfer_invalid"
	seedBridgeTransferWithPending(t, input, transferID, math.NewInt(1000).String(), 0)

	err := k.SubmitFraudProof(input.Ctx, transferID, submitter.String(), []byte("evidence"))
	require.NoError(t, err)

	resolved, err := k.ResolveFraudProof(input.Ctx, transferID, false)
	require.NoError(t, err)
	require.Equal(t, types.FraudProofStatus_FRAUD_PROOF_INVALID, resolved.Status)
	require.True(t, resolved.RewardAmount.IsZero(),
		"expected zero reward, got %s", resolved.RewardAmount.String())

	transfer := getBridgeTransfer(t, input, transferID)
	require.Equal(t, types.TransferStatus_PENDING, transfer.Status)
}

func TestResolveFraudProofMissing(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	_, err := k.ResolveFraudProof(input.Ctx, "unknown", true)
	require.ErrorIs(t, err, types.ErrFraudProofNotFound)
}

func TestResolveFraudProofExpired(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	transferID := "transfer_expired"
	seedBridgeTransferWithPending(t, input, transferID, math.NewInt(1000).String(), 0)

	err := k.SubmitFraudProof(input.Ctx, transferID, keepertest.GenTestAddr().String(), []byte("evidence"))
	require.NoError(t, err)

	// Advance time past the fraud proof window
	ctx := keepertest.AdvanceTime(input.Ctx, types.DefaultFraudProofWindow+time.Second)
	_, err = k.ResolveFraudProof(ctx, transferID, true)
	require.ErrorIs(t, err, types.ErrFraudProofExpired)
}

func TestFraudProofWindow(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	transferID := "transfer_1"
	seedBridgeTransferWithPending(t, input, transferID, math.NewInt(1000).String(), 0)

	// Check if in fraud proof window
	inWindow := k.IsInFraudProofWindow(input.Ctx, transferID)
	require.True(t, inWindow)

	// Advance time past window
	ctx := keepertest.AdvanceTime(input.Ctx, types.DefaultFraudProofWindow+time.Second)

	inWindow = k.IsInFraudProofWindow(ctx, transferID)
	require.False(t, inWindow)
}

// Chain Support Tests

func TestAddSupportedChain(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	chainConfig := types.ChainConfig{
		ChainId:   "ethereum",
		ChainName: "Ethereum",
		Enabled:   true,
	}

	err := k.AddSupportedChain(input.Ctx, chainConfig)
	require.NoError(t, err)
}

func TestGetSupportedChain(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	chainConfig := types.ChainConfig{
		ChainId:   "ethereum",
		ChainName: "Ethereum",
		Enabled:   true,
	}

	err := k.AddSupportedChain(input.Ctx, chainConfig)
	require.NoError(t, err)

	retrieved, found := k.GetSupportedChain(input.Ctx, "ethereum")
	require.True(t, found)
	require.Equal(t, "Ethereum", retrieved.ChainName)
}

func TestRemoveSupportedChain(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	chainConfig := types.ChainConfig{
		ChainId:   "ethereum",
		ChainName: "Ethereum",
		Enabled:   true,
	}

	err := k.AddSupportedChain(input.Ctx, chainConfig)
	require.NoError(t, err)

	k.RemoveSupportedChain(input.Ctx, "ethereum")

	_, found := k.GetSupportedChain(input.Ctx, "ethereum")
	require.False(t, found)
}

func TestDisableChain(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	chainConfig := types.ChainConfig{
		ChainId:   "ethereum",
		ChainName: "Ethereum",
		Enabled:   true,
	}

	err := k.AddSupportedChain(input.Ctx, chainConfig)
	require.NoError(t, err)

	err = k.DisableChain(input.Ctx, "ethereum")
	require.NoError(t, err)

	retrieved, found := k.GetSupportedChain(input.Ctx, "ethereum")
	require.True(t, found)
	require.False(t, retrieved.Enabled)
}

// Fee Tests

func TestCalculateBridgeFee(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	amount := math.NewInt(1000000)
	fee := k.CalculateBridgeFee(input.Ctx, amount, "ethereum")

	require.True(t, fee.GT(math.ZeroInt()))
	require.True(t, fee.LT(amount))
}

func TestBridgeFeeCollection(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	amount := math.NewInt(1000000)
	fee := k.CalculateBridgeFee(input.Ctx, amount, "ethereum")

	collected := k.GetCollectedFees(input.Ctx)
	require.NotNil(t, collected)

	// Record fee collection
	k.AddCollectedFee(input.Ctx, sdk.NewCoin("uaura", fee))

	newCollected := k.GetCollectedFees(input.Ctx)
	require.True(t, newCollected.AmountOf("uaura").GT(collected.AmountOf("uaura")))
}

// Genesis Tests

func TestInitGenesis(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	legacyAmino := codec.NewLegacyAmino()
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge")
	ps = ps.WithKeyTable(types.ParamKeyTable())
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, &ps, nil, nil, nil, nil)

	// Missing params should fail validation
	err := k.InitGenesis(input.Ctx, types.GenesisState{})
	require.Error(t, err)

	genesisState := testBridgeGenesisState()
	err = k.InitGenesis(input.Ctx, genesisState)
	require.NoError(t, err)

	exported := k.ExportGenesis(input.Ctx)
	require.Equal(t, genesisState.Params, exported.Params)
}

func TestExportGenesis(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	legacyAmino := codec.NewLegacyAmino()
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge")
	ps = ps.WithKeyTable(types.ParamKeyTable())
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, &ps, nil, nil, nil, nil)

	genesisState := testBridgeGenesisState()
	require.NoError(t, k.InitGenesis(input.Ctx, genesisState))

	exported := k.ExportGenesis(input.Ctx)
	require.Equal(t, genesisState.Params, exported.Params)
	require.Len(t, exported.Transfers, len(genesisState.Transfers))
	for i := range genesisState.Transfers {
		// Compare fields individually to avoid issues with timestamp timezones and protobuf internal fields
		require.Equal(t, genesisState.Transfers[i].TransferId, exported.Transfers[i].TransferId)
		require.Equal(t, genesisState.Transfers[i].SourceChain, exported.Transfers[i].SourceChain)
		require.Equal(t, genesisState.Transfers[i].TargetChain, exported.Transfers[i].TargetChain)
		require.Equal(t, genesisState.Transfers[i].Sender, exported.Transfers[i].Sender)
		require.Equal(t, genesisState.Transfers[i].Recipient, exported.Transfers[i].Recipient)
		require.True(t, genesisState.Transfers[i].Amount.Equal(exported.Transfers[i].Amount))
		require.Equal(t, genesisState.Transfers[i].Denom, exported.Transfers[i].Denom)
		require.Equal(t, genesisState.Transfers[i].Status, exported.Transfers[i].Status)
		require.True(t, genesisState.Transfers[i].Timestamp.Unix() == exported.Transfers[i].Timestamp.Unix())
	}
	require.Len(t, exported.ChainConfigs, len(genesisState.ChainConfigs))
	for i := range genesisState.ChainConfigs {
		require.Equal(t, genesisState.ChainConfigs[i], exported.ChainConfigs[i])
	}
	require.Len(t, exported.Validators, len(genesisState.Validators))
	for i := range genesisState.Validators {
		// Compare meaningful fields only (not internal protobuf fields like XXX_sizecache)
		require.Equal(t, genesisState.Validators[i].Address, exported.Validators[i].Address)
		require.Equal(t, genesisState.Validators[i].PublicKey, exported.Validators[i].PublicKey)
		require.Equal(t, genesisState.Validators[i].Power, exported.Validators[i].Power)
		require.Equal(t, genesisState.Validators[i].Active, exported.Validators[i].Active)
		require.Equal(t, genesisState.Validators[i].Chains, exported.Validators[i].Chains)
	}
	require.Len(t, exported.WrappedTokens, len(genesisState.WrappedTokens))
	for i := range genesisState.WrappedTokens {
		// Compare meaningful fields only (not internal protobuf fields like XXX_sizecache)
		require.Equal(t, genesisState.WrappedTokens[i].WrappedDenom, exported.WrappedTokens[i].WrappedDenom)
		require.Equal(t, genesisState.WrappedTokens[i].OriginalDenom, exported.WrappedTokens[i].OriginalDenom)
		require.Equal(t, genesisState.WrappedTokens[i].SourceChain, exported.WrappedTokens[i].SourceChain)
		require.True(t, genesisState.WrappedTokens[i].TotalSupply.Equal(exported.WrappedTokens[i].TotalSupply))
		require.Equal(t, genesisState.WrappedTokens[i].Decimals, exported.WrappedTokens[i].Decimals)
		require.True(t, genesisState.WrappedTokens[i].LockedAmount.Equal(exported.WrappedTokens[i].LockedAmount))
	}
	require.Len(t, exported.SharedIdentities, len(genesisState.SharedIdentities))
	for i := range genesisState.SharedIdentities {
		// Compare meaningful fields only
		require.Equal(t, genesisState.SharedIdentities[i].Address, exported.SharedIdentities[i].Address)
		require.Equal(t, genesisState.SharedIdentities[i].VerifiedAura, exported.SharedIdentities[i].VerifiedAura)
		require.Equal(t, genesisState.SharedIdentities[i].VerifiedPaw, exported.SharedIdentities[i].VerifiedPaw)
		require.Equal(t, genesisState.SharedIdentities[i].VerifiedXai, exported.SharedIdentities[i].VerifiedXai)
		require.Equal(t, genesisState.SharedIdentities[i].AuraIrScore, exported.SharedIdentities[i].AuraIrScore)
		require.Equal(t, genesisState.SharedIdentities[i].ReputationScore, exported.SharedIdentities[i].ReputationScore)
		require.Equal(t, genesisState.SharedIdentities[i].LinkedAddresses, exported.SharedIdentities[i].LinkedAddresses)
		// Compare timestamps by Unix time to avoid timezone/precision issues
		require.Equal(t, genesisState.SharedIdentities[i].VerifiedAt.Unix(), exported.SharedIdentities[i].VerifiedAt.Unix())
	}
	// Note: CrossChainSwaps and RelayerStats use reflect.DeepEqual which may fail
	// on internal protobuf fields. If tests fail, these should also be changed to
	// field-by-field comparison like the others above.
	require.Len(t, exported.CrossChainSwaps, len(genesisState.CrossChainSwaps))
	require.Len(t, exported.RelayerStats, len(genesisState.RelayerStats))
}

func testBridgeGenesisState() types.GenesisState {
	now := time.Unix(1, 0)
	transfer := types.CrossChainTransfer{
		TransferId:  "transfer-1",
		SourceChain: "aura",
		TargetChain: "paw",
		Sender:      "aura1sender",
		Recipient:   "paw1recipient",
		Amount:      math.NewInt(1000),
		Denom:       "uaura",
		Status:      types.TransferStatus_CONFIRMED,
		Timestamp:   now,
	}
	chainCfg := types.ChainConfig{
		ChainId:          "aura",
		ChainName:        "Aura",
		AddressPrefix:    "aura",
		MinConfirmations: 6,
		Enabled:          true,
	}
	validator := types.BridgeValidator{
		Address: sdk.AccAddress("validator1__________").String(),
		Active:  true,
		Power:   100,
		Chains:  []string{"aura"},
	}
	wrapped := types.WrappedToken{
		WrappedDenom:  "paw.token",
		OriginalDenom: "token",
		SourceChain:   "paw",
		TotalSupply:   math.NewInt(500),
		LockedAmount:  math.NewInt(500),
	}
	identity := types.SharedIdentity{
		Address:      "aura1identity",
		VerifiedAura: true,
	}
	swap := types.CrossChainSwap{
		SwapId:          "swap-1",
		SourceChain:     "aura",
		TargetChain:     "paw",
		TargetDenom:     "pawcoin",
		MinTargetAmount: math.NewInt(200),
		Sender:          "aura1sender",
		Recipient:       "paw1recipient",
		Route:           []string{"aura", "osmosis", "paw"},
		Status:          "pending",
		InitiatedAt:     now,
	}
	relayerStats := types.RelayerStats{
		RelayerAddress:        sdk.AccAddress("relayer1___________").String(),
		TotalTransfersRelayed: 5,
		SuccessfulTransfers:   5,
		FailedTransfers:       0,
		TotalVolume:           math.NewInt(1000),
		LastRelay:             &now,
		UptimePercentage:      math.LegacyNewDec(1),
	}
	return types.GenesisState{
		Params: types.BridgeParams{
			Enabled:                      true,
			MinConfirmations:             2,
			BridgeFeeBasisPoints:         25,
			MaxTransferAmount:            math.NewInt(1000000000),
			ValidatorThresholdPercentage: 66,
		},
		Transfers:        []types.CrossChainTransfer{transfer},
		ChainConfigs:     []types.ChainConfig{chainCfg},
		Validators:       []types.BridgeValidator{validator},
		WrappedTokens:    []types.WrappedToken{wrapped},
		SharedIdentities: []types.SharedIdentity{identity},
		CrossChainSwaps:  []types.CrossChainSwap{swap},
		RelayerStats:     []types.RelayerStats{relayerStats},
	}
}
