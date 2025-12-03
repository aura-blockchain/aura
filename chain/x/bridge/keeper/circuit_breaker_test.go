package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/codec"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

// setupKeeperForCircuitBreaker creates a keeper with proper paramstore initialization
func setupKeeperForCircuitBreaker(t *testing.T) (*keeper.Keeper, keepertest.TestInput) {
	input := keepertest.CreateTestInput(t)

	// Create LegacyAmino codec for paramstore
	legacyAmino := codec.NewLegacyAmino()

	// Create a paramstore with KeyTable
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge")

	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		&ps,
		nil, // bankKeeper
		nil, // accountKeeper
		nil, // vcKeeper
		nil, // stakingKeeper
	)

	return k, input
}

// TestCircuitBreaker_GlobalPauseBlocksAllOperations verifies that when the global
// pause flag is set, all bridge operations (lock, unlock, mint) are blocked.
func TestCircuitBreaker_GlobalPauseBlocksAllOperations(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	// Set up chain config
	require.NoError(t, k.AddSupportedChain(ctx, types.ChainConfig{ChainId: "paw", Enabled: true}))

	// Set global pause
	params := k.GetParams(ctx)
	params.Paused = true
	k.SetParams(ctx, params)

	// Test 1: LockTokens should be blocked
	amount := sdk.NewCoin("uaura", sdkmath.NewInt(1000000))
	lockMsg := &types.MsgLockTokens{
		Sender:      keepertest.GenTestAddr().String(),
		Recipient:   "paw1recipient",
		Amount:      &amount,
		TargetChain: "paw",
	}
	_, err := ms.LockTokens(sdk.WrapSDKContext(ctx), lockMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "globally paused")

	// Test 2: MintTokens should be blocked
	mintMsg := &types.MsgMintTokens{
		Validator:    keepertest.GenTestAddr().String(),
		SourceChain:  "paw",
		SourceTxHash: "0x123",
		Recipient:    keepertest.GenTestAddr().String(),
		Amount:       "1000000",
		Denom:        "upaw",
	}
	_, err = ms.MintTokens(sdk.WrapSDKContext(ctx), mintMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "globally paused")

	// Test 3: UnlockTokens should be blocked
	unlockMsg := &types.MsgUnlockTokens{
		Sender:      keepertest.GenTestAddr().String(),
		SourceChain: "paw",
		BurnTxHash:  "0x456",
		Amount:      "1000000",
		Denom:       "uaura",
	}
	_, err = ms.UnlockTokens(sdk.WrapSDKContext(ctx), unlockMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "globally paused")
}

// TestCircuitBreaker_PerChainPauseSelectivelyBlocks verifies that per-chain pause
// only affects operations for the specific paused chain, not other chains.
func TestCircuitBreaker_PerChainPauseSelectivelyBlocks(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx

	// Pause only "paw" chain
	params := k.GetParams(ctx)
	params.PausedChains = []string{"paw"}
	k.SetParams(ctx, params)

	// PAW operations should be blocked
	err := k.RequireNotPaused(ctx, "paw")
	require.Error(t, err)
	require.Contains(t, err.Error(), "paused for chain paw")

	// XAI operations should still work
	err = k.RequireNotPaused(ctx, "xai")
	require.NoError(t, err)

	// Aura operations should still work
	err = k.RequireNotPaused(ctx, "aura")
	require.NoError(t, err)
}

// TestCircuitBreaker_AutoPauseTriggersOnThresholdExceeded verifies that the
// auto-pause mechanism triggers when hourly minting exceeds the configured threshold.
func TestCircuitBreaker_AutoPauseTriggersOnThresholdExceeded(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx

	// Enable auto-pause with threshold of 5 billion
	params := k.GetParams(ctx)
	params.AutoPauseEnabled = true
	params.AutoPauseThreshold = "5000000000"
	params.Paused = false // Start unpaused
	k.SetParams(ctx, params)

	// Verify bridge is initially operational
	err := k.RequireNotPaused(ctx, "paw")
	require.NoError(t, err)
	require.False(t, params.Paused)

	// Record 4 billion minted (below threshold)
	k.RecordMintedAmount(ctx, "test-denom", sdkmath.NewInt(4_000_000_000))

	// Attempt to mint 2 billion more (total 6 billion > 5 billion threshold)
	// This should trigger auto-pause
	amount := sdkmath.NewInt(2_000_000_000)
	triggered := k.CheckAndTriggerAutoPause(ctx, "test-denom", amount)
	require.True(t, triggered, "auto-pause should have been triggered")

	// Verify bridge is now paused
	params = k.GetParams(ctx)
	require.True(t, params.Paused, "bridge should be paused after threshold exceeded")

	// Verify all operations are blocked after auto-pause
	err = k.RequireNotPaused(ctx, "paw")
	require.Error(t, err)
	require.Contains(t, err.Error(), "globally paused")

	// Verify event was emitted
	events := ctx.EventManager().Events()
	foundAutoPauseEvent := false
	for _, event := range events {
		if event.Type == "bridge_auto_paused" {
			foundAutoPauseEvent = true
			// Verify event attributes
			hasReason := false
			hasThreshold := false
			for _, attr := range event.Attributes {
				if attr.Key == "reason" && attr.Value == "hourly_mint_threshold_exceeded" {
					hasReason = true
				}
				if attr.Key == "threshold" && attr.Value == "5000000000" {
					hasThreshold = true
				}
			}
			require.True(t, hasReason, "event should have reason attribute")
			require.True(t, hasThreshold, "event should have threshold attribute")
		}
	}
	require.True(t, foundAutoPauseEvent, "bridge_auto_paused event should be emitted")
}

// TestCircuitBreaker_AutoPauseDoesNotTriggerBelowThreshold verifies that
// auto-pause does NOT trigger when minting is below the threshold.
func TestCircuitBreaker_AutoPauseDoesNotTriggerBelowThreshold(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx

	// Enable auto-pause with threshold of 5 billion
	params := k.GetParams(ctx)
	params.AutoPauseEnabled = true
	params.AutoPauseThreshold = "5000000000"
	params.Paused = false
	k.SetParams(ctx, params)

	// Record 3 billion minted
	k.RecordMintedAmount(ctx, "test-denom", sdkmath.NewInt(3_000_000_000))

	// Attempt to mint 1 billion more (total 4 billion < 5 billion threshold)
	amount := sdkmath.NewInt(1_000_000_000)
	triggered := k.CheckAndTriggerAutoPause(ctx, "test-denom", amount)
	require.False(t, triggered, "auto-pause should NOT trigger below threshold")

	// Verify bridge remains operational
	params = k.GetParams(ctx)
	require.False(t, params.Paused, "bridge should remain unpaused")

	// Verify operations are still allowed
	err := k.RequireNotPaused(ctx, "paw")
	require.NoError(t, err)
}

// TestCircuitBreaker_AutoPauseDisabledDoesNotTrigger verifies that when
// auto-pause is disabled, no automatic pausing occurs regardless of mint volume.
func TestCircuitBreaker_AutoPauseDisabledDoesNotTrigger(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx

	// Disable auto-pause
	params := k.GetParams(ctx)
	params.AutoPauseEnabled = false
	params.AutoPauseThreshold = "5000000000"
	params.Paused = false
	k.SetParams(ctx, params)

	// Record massive amount minted (well above threshold)
	k.RecordMintedAmount(ctx, "test-denom", sdkmath.NewInt(10_000_000_000))

	// Attempt to mint even more (total far exceeds threshold)
	amount := sdkmath.NewInt(10_000_000_000)
	triggered := k.CheckAndTriggerAutoPause(ctx, "test-denom", amount)
	require.False(t, triggered, "auto-pause should NOT trigger when disabled")

	// Verify bridge remains operational
	params = k.GetParams(ctx)
	require.False(t, params.Paused, "bridge should remain unpaused")
}

// TestCircuitBreaker_UnpauseRestoresFunctionality verifies that when the pause
// flag is cleared, normal bridge operations resume.
func TestCircuitBreaker_UnpauseRestoresFunctionality(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx

	// Start with bridge paused
	params := k.GetParams(ctx)
	params.Paused = true
	k.SetParams(ctx, params)

	// Verify operations are blocked
	err := k.RequireNotPaused(ctx, "paw")
	require.Error(t, err)
	require.Contains(t, err.Error(), "globally paused")

	// Unpause the bridge
	params.Paused = false
	k.SetParams(ctx, params)

	// Verify operations are now allowed
	err = k.RequireNotPaused(ctx, "paw")
	require.NoError(t, err)
}

// TestCircuitBreaker_PerChainUnpauseRestoresFunctionality verifies that
// removing a chain from the paused list restores operations for that chain.
func TestCircuitBreaker_PerChainUnpauseRestoresFunctionality(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx

	// Pause "paw" chain
	params := k.GetParams(ctx)
	params.PausedChains = []string{"paw", "xai"}
	k.SetParams(ctx, params)

	// Verify PAW is blocked
	err := k.RequireNotPaused(ctx, "paw")
	require.Error(t, err)

	// Verify XAI is blocked
	err = k.RequireNotPaused(ctx, "xai")
	require.Error(t, err)

	// Unpause only PAW (remove from list)
	params.PausedChains = []string{"xai"}
	k.SetParams(ctx, params)

	// Verify PAW is now allowed
	err = k.RequireNotPaused(ctx, "paw")
	require.NoError(t, err)

	// Verify XAI is still blocked
	err = k.RequireNotPaused(ctx, "xai")
	require.Error(t, err)
}

// TestCircuitBreaker_EmergencyPauseAuthorization verifies that only authorized
// addresses can trigger emergency pause.
func TestCircuitBreaker_EmergencyPauseAuthorization(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx

	// Generate test addresses
	addrs := keepertest.GenTestAddrs(3)
	guardian1 := addrs[0].String()
	guardian2 := addrs[1].String()
	unauthorized := addrs[2].String()

	// Set authorized guardians
	params := k.GetParams(ctx)
	params.EmergencyPauseAddresses = []string{guardian1, guardian2}
	k.SetParams(ctx, params)

	// Guardian 1 should be authorized
	require.True(t, k.IsEmergencyPauseAuthorized(ctx, guardian1))

	// Guardian 2 should be authorized
	require.True(t, k.IsEmergencyPauseAuthorized(ctx, guardian2))

	// Unauthorized address should NOT be authorized
	require.False(t, k.IsEmergencyPauseAuthorized(ctx, unauthorized))

	// Empty address should NOT be authorized
	require.False(t, k.IsEmergencyPauseAuthorized(ctx, ""))
}

// TestCircuitBreaker_HourlyMintTracking verifies that hourly minting is
// correctly tracked and accumulated.
func TestCircuitBreaker_HourlyMintTracking(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx

	// Initially, no mints recorded
	hourly := k.GetHourlyMintedAmount(ctx, "test-denom")
	require.True(t, hourly.IsZero())

	// Record first mint
	k.RecordMintedAmount(ctx, "test-denom", sdkmath.NewInt(1_000_000))
	hourly = k.GetHourlyMintedAmount(ctx, "test-denom")
	require.Equal(t, sdkmath.NewInt(1_000_000), hourly)

	// Record second mint (should accumulate)
	k.RecordMintedAmount(ctx, "test-denom", sdkmath.NewInt(2_000_000))
	hourly = k.GetHourlyMintedAmount(ctx, "test-denom")
	require.Equal(t, sdkmath.NewInt(3_000_000), hourly)

	// Record third mint (should accumulate)
	k.RecordMintedAmount(ctx, "test-denom", sdkmath.NewInt(3_000_000))
	hourly = k.GetHourlyMintedAmount(ctx, "test-denom")
	require.Equal(t, sdkmath.NewInt(6_000_000), hourly)
}

// TestCircuitBreaker_MintTokensAutopauses verifies that MintTokens operation
// triggers auto-pause when threshold is exceeded.
func TestCircuitBreaker_MintTokensAutopauses(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	// Enable auto-pause with low threshold
	params := k.GetParams(ctx)
	params.AutoPauseEnabled = true
	params.AutoPauseThreshold = "1000000" // 1 million threshold
	params.Paused = false
	params.BridgeEnabled = true
	k.SetParams(ctx, params)

	// First mint attempt (should trigger auto-pause check and trigger)
	mintMsg := &types.MsgMintTokens{
		Validator:    keepertest.GenTestAddr().String(),
		SourceChain:  "paw",
		SourceTxHash: "0xtest123",
		Recipient:    keepertest.GenTestAddr().String(),
		Amount:       "2000000", // 2 million (exceeds threshold)
		Denom:        "upaw",
	}

	// This should trigger auto-pause because amount (2M) > threshold (1M)
	_, err := ms.MintTokens(sdk.WrapSDKContext(ctx), mintMsg)
	require.Error(t, err, "should fail due to auto-pause trigger")
	require.Contains(t, err.Error(), "auto-pause triggered")

	// Verify bridge is now paused
	params = k.GetParams(ctx)
	require.True(t, params.Paused, "bridge should be paused after auto-pause")
}

// TestCircuitBreaker_MultipleOperationsAfterPause verifies that all operation
// types remain blocked after pause, even with multiple attempts.
func TestCircuitBreaker_MultipleOperationsAfterPause(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	// Set up chain config
	require.NoError(t, k.AddSupportedChain(ctx, types.ChainConfig{ChainId: "paw", Enabled: true}))

	// Enable bridge and pause it
	params := k.GetParams(ctx)
	params.BridgeEnabled = true
	params.Paused = true // Start paused
	k.SetParams(ctx, params)

	// Try LockTokens multiple times - all should fail
	for i := 0; i < 3; i++ {
		lockAmount := sdk.NewCoin("uaura", sdkmath.NewInt(1000000))
		lockMsg := &types.MsgLockTokens{
			Sender:      keepertest.GenTestAddr().String(),
			Recipient:   "paw1recipient",
			Amount:      &lockAmount,
			TargetChain: "paw",
		}
		_, err := ms.LockTokens(sdk.WrapSDKContext(ctx), lockMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "globally paused")
	}

	// Try MintTokens multiple times - all should fail
	for i := 0; i < 3; i++ {
		mintMsg := &types.MsgMintTokens{
			Validator:    keepertest.GenTestAddr().String(),
			SourceChain:  "paw",
			SourceTxHash: "0xtest" + string(rune(i)),
			Recipient:    keepertest.GenTestAddr().String(),
			Amount:       "1000000",
			Denom:        "upaw",
		}
		_, err := ms.MintTokens(sdk.WrapSDKContext(ctx), mintMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "globally paused")
	}

	// Try UnlockTokens multiple times - all should fail
	for i := 0; i < 3; i++ {
		unlockMsg := &types.MsgUnlockTokens{
			Sender:      keepertest.GenTestAddr().String(),
			SourceChain: "paw",
			BurnTxHash:  "0xburn" + string(rune(i)),
			Amount:      "1000000",
			Denom:       "uaura",
		}
		_, err := ms.UnlockTokens(sdk.WrapSDKContext(ctx), unlockMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "globally paused")
	}
}

// TestCircuitBreaker_ZeroAmountMintDoesNotTriggerAutoPause verifies that
// minting zero amount doesn't trigger auto-pause or affect hourly tracking.
func TestCircuitBreaker_ZeroAmountMintDoesNotTriggerAutoPause(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx

	// Enable auto-pause with very low threshold
	params := k.GetParams(ctx)
	params.AutoPauseEnabled = true
	params.AutoPauseThreshold = "1" // Even 1 unit would trigger
	params.Paused = false
	k.SetParams(ctx, params)

	// Record zero mint (should be ignored)
	k.RecordMintedAmount(ctx, "test-denom", sdkmath.ZeroInt())

	// Verify nothing was recorded
	hourly := k.GetHourlyMintedAmount(ctx, "test-denom")
	require.True(t, hourly.IsZero())

	// Verify auto-pause check doesn't trigger
	triggered := k.CheckAndTriggerAutoPause(ctx, "test-denom", sdkmath.ZeroInt())
	require.False(t, triggered)

	// Verify bridge remains operational
	params = k.GetParams(ctx)
	require.False(t, params.Paused)
}

// TestCircuitBreaker_InvalidThresholdDoesNotTrigger verifies that with an
// invalid threshold configuration, auto-pause does not trigger.
func TestCircuitBreaker_InvalidThresholdDoesNotTrigger(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx

	// Enable auto-pause with invalid threshold
	params := k.GetParams(ctx)
	params.AutoPauseEnabled = true
	params.AutoPauseThreshold = "not-a-number"
	params.Paused = false
	k.SetParams(ctx, params)

	// Try to trigger with large amount
	amount := sdkmath.NewInt(10_000_000_000)
	triggered := k.CheckAndTriggerAutoPause(ctx, "test-denom", amount)
	require.False(t, triggered, "should not trigger with invalid threshold")

	// Verify bridge remains operational
	params = k.GetParams(ctx)
	require.False(t, params.Paused)
}

// TestCircuitBreaker_NegativeThresholdDoesNotTrigger verifies that with a
// negative threshold, auto-pause does not trigger.
func TestCircuitBreaker_NegativeThresholdDoesNotTrigger(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx

	// Enable auto-pause with negative threshold
	params := k.GetParams(ctx)
	params.AutoPauseEnabled = true
	params.AutoPauseThreshold = "-1000000"
	params.Paused = false
	k.SetParams(ctx, params)

	// Try to trigger with any amount
	amount := sdkmath.NewInt(1_000_000)
	triggered := k.CheckAndTriggerAutoPause(ctx, "test-denom", amount)
	require.False(t, triggered, "should not trigger with negative threshold")

	// Verify bridge remains operational
	params = k.GetParams(ctx)
	require.False(t, params.Paused)
}

// TestCircuitBreaker_IntegrationWithLockTokens verifies that the circuit breaker
// properly integrates with LockTokens to prevent operations when paused.
func TestCircuitBreaker_IntegrationWithLockTokens(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	// Set up chain config
	require.NoError(t, k.AddSupportedChain(ctx, types.ChainConfig{ChainId: "paw", Enabled: true}))

	// Test with bridge operational
	params := k.GetParams(ctx)
	params.Paused = false
	k.SetParams(ctx, params)

	amount := sdk.NewCoin("uaura", sdkmath.NewInt(1000))
	msg := &types.MsgLockTokens{
		Sender:      keepertest.GenTestAddr().String(),
		TargetChain: "paw",
		Recipient:   "paw1recipient",
		Amount:      &amount,
	}

	// Should succeed when not paused
	resp, err := ms.LockTokens(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)
	require.NotEmpty(t, resp.TransferId)

	// Now pause the bridge
	params.Paused = true
	k.SetParams(ctx, params)

	// Same message should now fail
	_, err = ms.LockTokens(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "globally paused")
}

// TestCircuitBreaker_IntegrationWithMintTokens verifies that the circuit breaker
// properly integrates with MintTokens to prevent operations when paused.
func TestCircuitBreaker_IntegrationWithMintTokens(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	msg := &types.MsgMintTokens{
		Validator:    keepertest.GenTestAddr().String(),
		SourceChain:  "paw",
		SourceTxHash: "0xabc",
		Recipient:    keepertest.GenTestAddr().String(),
		Amount:       "1000",
		Denom:        "paw.token",
	}

	// Should succeed when not paused
	resp, err := ms.MintTokens(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)
	require.True(t, resp.Success)

	// Now pause the bridge
	params := k.GetParams(ctx)
	params.Paused = true
	k.SetParams(ctx, params)

	// New mint with different hash should fail when paused
	msg2 := &types.MsgMintTokens{
		Validator:    keepertest.GenTestAddr().String(),
		SourceChain:  "paw",
		SourceTxHash: "0xdef",
		Recipient:    keepertest.GenTestAddr().String(),
		Amount:       "1000",
		Denom:        "paw.token",
	}
	_, err = ms.MintTokens(sdk.WrapSDKContext(ctx), msg2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "globally paused")
}

// TestCircuitBreaker_IntegrationWithUnlockTokens verifies that the circuit breaker
// properly integrates with UnlockTokens to prevent operations when paused.
func TestCircuitBreaker_IntegrationWithUnlockTokens(t *testing.T) {
	k, input := setupKeeperForCircuitBreaker(t)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	// Pause the bridge
	params := k.GetParams(ctx)
	params.Paused = true
	k.SetParams(ctx, params)

	msg := &types.MsgUnlockTokens{
		Sender:      keepertest.GenTestAddr().String(),
		SourceChain: "paw",
		BurnTxHash:  "0x123",
		Amount:      "1000",
		Denom:       "uaura",
	}

	// Should fail when paused
	_, err := ms.UnlockTokens(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "globally paused")

	// Unpause the bridge
	params.Paused = false
	k.SetParams(ctx, params)

	// Still might fail for other reasons (e.g., no transfer record), but should NOT fail for pause
	_, err = ms.UnlockTokens(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		require.NotContains(t, err.Error(), "globally paused")
	}
}
