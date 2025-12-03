package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

// TestGetParams_WithParamstore tests GetParams when paramstore has KeyTable
func TestGetParams_WithParamstore(t *testing.T) {
	input := keepertest.CreateTestInput(t)

	// Create LegacyAmino codec for paramstore
	legacyAmino := codec.NewLegacyAmino()

	// Create a paramstore with KeyTable
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge")
	ps = ps.WithKeyTable(types.ParamKeyTable())

	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		&ps,
		nil,
		nil,
		nil,
		nil, // stakingKeeper
	)

	// Set params
	params := types.DefaultParams()
	params.BridgeEnabled = false
	err := k.SetParams(input.Ctx, params)
	require.NoError(t, err)

	// Get params should return what we set
	retrieved := k.GetParams(input.Ctx)
	require.Equal(t, params.BridgeEnabled, retrieved.BridgeEnabled)
}

// TestGetParams_WithoutParamstore tests GetParams when paramstore is nil
func TestGetParams_WithoutParamstore(t *testing.T) {
	input := keepertest.CreateTestInput(t)

	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil, // nil paramstore
		nil,
		nil,
		nil,
		nil, // stakingKeeper
	)

	// Get params should return defaults
	retrieved := k.GetParams(input.Ctx)
	require.Equal(t, types.DefaultParams(), retrieved)
}

// TestExecuteWithdrawal_NotFound tests error when withdrawal doesn't exist
func TestExecuteWithdrawal_NotFound(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Try to execute non-existent withdrawal
	err := k.ExecuteWithdrawal(input.Ctx, "non-existent-withdrawal")
	require.Error(t, err)
	require.Equal(t, types.ErrWithdrawalNotFound, err)
}

// TestGetFraudProof_NotFound tests GetFraudProof when no proof exists
func TestGetFraudProof_NotFound(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	proof, found := k.GetFraudProof(input.Ctx, "non-existent-transfer")
	require.False(t, found)
	require.Empty(t, proof.Challenger)
}

// TestGetSupportedChain_UnmarshalError tests error handling in GetSupportedChain
func TestGetSupportedChain_UnmarshalError(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Manually add invalid data to store
	store := input.Ctx.KVStore(input.StoreKey)
	key := []byte("chain-corrupt")
	store.Set(key, []byte{0xFF, 0xFF}) // Invalid protobuf data

	// Should handle unmarshal error gracefully
	chainConfig, found := k.GetSupportedChain(input.Ctx, "corrupt")
	require.False(t, found)
	require.Empty(t, chainConfig.ChainId)
}

// TestDisableChain_NotFound tests error when trying to disable non-existent chain
func TestDisableChain_NotFound(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	err := k.DisableChain(input.Ctx, "non-existent-chain")
	require.Error(t, err)
	require.Equal(t, types.ErrChainNotFound, err)
}

// TestAddCollectedFee_WithExistingFees tests AddCollectedFee with existing fees
func TestAddCollectedFee_WithExistingFees(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Add first fee
	fee1 := sdk.NewCoin("uaura", math.NewInt(1000))
	k.AddCollectedFee(input.Ctx, fee1)

	// Add second fee (should accumulate)
	fee2 := sdk.NewCoin("uaura", math.NewInt(2000))
	k.AddCollectedFee(input.Ctx, fee2)

	// Get collected fees
	collected := k.GetCollectedFees(input.Ctx)
	require.True(t, collected.AmountOf("uaura").Equal(math.NewInt(3000)))
}

// TestAddCollectedFee_MultiDenom tests AddCollectedFee with multiple denominations
func TestAddCollectedFee_MultiDenom(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Add fees for different denoms
	k.AddCollectedFee(input.Ctx, sdk.NewCoin("uaura", math.NewInt(1000)))
	k.AddCollectedFee(input.Ctx, sdk.NewCoin("uatom", math.NewInt(2000)))
	k.AddCollectedFee(input.Ctx, sdk.NewCoin("uaura", math.NewInt(500)))

	// Get collected fees
	collected := k.GetCollectedFees(input.Ctx)
	require.True(t, collected.AmountOf("uaura").Equal(math.NewInt(1500)))
	require.True(t, collected.AmountOf("uatom").Equal(math.NewInt(2000)))
}

// TestGetCollectedFees_Empty tests GetCollectedFees with no fees
func TestGetCollectedFees_Empty(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Get collected fees when empty
	collected := k.GetCollectedFees(input.Ctx)
	require.NotNil(t, collected)
	require.True(t, collected.IsZero())
}

// TestGetCollectedFees_WithCorruptData tests GetCollectedFees with invalid data
func TestGetCollectedFees_WithCorruptData(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	// Manually add corrupt data
	store := input.Ctx.KVStore(input.StoreKey)
	store.Set([]byte("collected-fees-corrupt"), []byte{0xFF, 0xFF})

	// Should skip corrupt data
	collected := k.GetCollectedFees(input.Ctx)
	require.NotNil(t, collected)
}

// TestGetAttestations_Empty tests GetAttestations with no attestations
func TestGetAttestations_Empty(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	attestations := k.GetAttestations(input.Ctx, "transfer-no-attestations")
	require.Empty(t, attestations)
}

// TestCheckAttestationThreshold_BelowThreshold tests threshold not met
func TestCheckAttestationThreshold_BelowThreshold(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	transferID := "transfer_below_threshold"
	seedBridgeTransfer(t, input, transferID, math.NewInt(1000).String(), 7)
	validators := keepertest.GenTestAddrs(3)

	// Submit only 3 attestations (need 7)
	for _, val := range validators {
		err := k.SubmitAttestation(input.Ctx, transferID, val.String(), true)
		require.NoError(t, err)
	}

	passed := k.CheckAttestationThreshold(input.Ctx, transferID)
	require.False(t, passed)
}

// TestCheckAttestationThreshold_ExactThreshold tests threshold exactly met
func TestCheckAttestationThreshold_ExactThreshold(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	transferID := "transfer_exact_threshold"
	seedBridgeTransfer(t, input, transferID, math.NewInt(1000).String(), 7)
	validators := keepertest.GenTestAddrs(7)

	// Submit exactly 7 attestations
	for _, val := range validators {
		err := k.SubmitAttestation(input.Ctx, transferID, val.String(), true)
		require.NoError(t, err)
	}

	passed := k.CheckAttestationThreshold(input.Ctx, transferID)
	require.True(t, passed)
}

// TestProcessWithdrawal_MaxAllowed tests ProcessWithdrawal with max allowed amount
func TestProcessWithdrawal_MaxAllowed(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	recipient := keepertest.GenTestAddr()
	// Just below default circuit breaker threshold (1,000,000,000 uaura)
	amount := sdk.NewCoins(sdk.NewCoin("uaura", math.NewInt(900000000)))
	seedBridgeTransfer(t, input, "transfer_max", amount.AmountOf("uaura").String(), 0)

	err := k.ProcessWithdrawal(input.Ctx, recipient.String(), amount, "transfer_max")
	require.NoError(t, err)
}

// TestIsInFraudProofWindow_NewTransfer tests fraud proof window for new transfer
func TestIsInFraudProofWindow_NewTransfer(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	transferID := "new-transfer"
	seedBridgeTransfer(t, input, transferID, math.NewInt(5000).String(), 0)

	// First call should record the time and return true
	inWindow := k.IsInFraudProofWindow(input.Ctx, transferID)
	require.True(t, inWindow)

	// Same block time should still be in window
	inWindow = k.IsInFraudProofWindow(input.Ctx, transferID)
	require.True(t, inWindow)
}

// TestAddSupportedChain_MarshalError tests error handling in AddSupportedChain
func TestAddSupportedChain_Success(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	chainConfig := types.ChainConfig{
		ChainId:   "polygon",
		ChainName: "Polygon",
		Enabled:   true,
	}

	err := k.AddSupportedChain(input.Ctx, chainConfig)
	require.NoError(t, err)

	// Verify it was stored
	retrieved, found := k.GetSupportedChain(input.Ctx, "polygon")
	require.True(t, found)
	require.Equal(t, "Polygon", retrieved.ChainName)
	require.True(t, retrieved.Enabled)
}

// TestCalculateBridgeFee_ZeroAmount tests fee calculation with zero amount
func TestCalculateBridgeFee_ZeroAmount(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	amount := math.ZeroInt()
	fee := k.CalculateBridgeFee(input.Ctx, amount, "ethereum")

	require.True(t, fee.Equal(math.ZeroInt()))
}

// TestCalculateBridgeFee_LargeAmount tests fee calculation with large amount
func TestCalculateBridgeFee_LargeAmount(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)

	amount := math.NewInt(1000000000000) // 1 trillion
	fee := k.CalculateBridgeFee(input.Ctx, amount, "ethereum")

	// With default 30 bps fee the charge should be 3 billion
	expectedFee := math.NewInt(3000000000)
	require.True(t, fee.Equal(expectedFee))
}

// TestNewKeeper_WithNilParamstore tests NewKeeper with nil paramstore
func TestNewKeeper_WithNilParamstore(t *testing.T) {
	input := keepertest.CreateTestInput(t)

	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil, // nil paramstore
		nil,
		nil,
		nil,
		nil, // stakingKeeper
	)

	require.NotNil(t, k)

	// Should still be able to get default params
	params := k.GetParams(input.Ctx)
	require.NotNil(t, params)
}

// TestNewKeeper_WithParamstoreNoKeyTable tests NewKeeper with paramstore without KeyTable
func TestNewKeeper_WithParamstoreNoKeyTable(t *testing.T) {
	input := keepertest.CreateTestInput(t)

	// Create LegacyAmino codec for paramstore
	legacyAmino := codec.NewLegacyAmino()

	// Create paramstore without KeyTable
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge")

	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		&ps,
		nil,
		nil,
		nil,
		nil, // stakingKeeper
	)

	require.NotNil(t, k)

	// Should be able to get params after KeyTable is set
	params := k.GetParams(input.Ctx)
	require.NotNil(t, params)
}

// TestNewKeeper_WithParamstoreWithKeyTable tests NewKeeper with paramstore that has KeyTable
func TestNewKeeper_WithParamstoreWithKeyTable(t *testing.T) {
	input := keepertest.CreateTestInput(t)

	// Create LegacyAmino codec for paramstore
	legacyAmino := codec.NewLegacyAmino()

	// Create paramstore with KeyTable
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge")
	ps = ps.WithKeyTable(types.ParamKeyTable())

	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		&ps,
		nil,
		nil,
		nil,
		nil, // stakingKeeper
	)

	require.NotNil(t, k)
}
