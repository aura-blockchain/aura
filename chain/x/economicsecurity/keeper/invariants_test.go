package keeper

//lint:file-ignore SA1019 -- tests rely on SDK invariant interfaces until crisis module is removed upstream

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

func TestAllInvariants(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Test with default genesis state (should pass)
	invariant := AllInvariants(k)
	msg, broken := invariant(ctx)
	require.False(t, broken, "all invariants should pass with default state: %s", msg)
	require.Empty(t, msg)

	// Test with valid vesting schedule
	validSchedule := &types.VestingSchedule{
		ScheduleId:         "test-schedule-1",
		BeneficiaryAddress: "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
		TotalAmount:        "1000000",
		VestedAmount:       "100000",
		VestingDuration:    31536000, // 1 year in seconds
		CliffDuration:      7776000,  // 90 days in seconds
		StartTime:          time.Now(),
		VestingType:        types.VestingTypeLinear,
	}
	err := k.SetVestingSchedule(ctx, validSchedule)
	require.NoError(t, err)

	// Test with valid vote lock
	validLock := &types.VoteLock{
		LockId:      "test-lock-1",
		Owner:       "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
		Amount:      "500000",
		LockStart:   time.Now(),
		LockEnd:     time.Now().Add(30 * 24 * time.Hour),
		VotingPower: "500000",
	}
	err = k.SetVoteLock(ctx, validLock)
	require.NoError(t, err)

	// Test with valid treasury tx
	// NOTE: Signatures must be non-nil empty slice, not nil
	validTx := &types.PendingTreasuryTx{
		TxId:         "test-tx-1",
		Recipient:    "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
		Amount:       "100000",
		Proposer:     "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
		Signatures:   make([]string, 0), // Empty slice, not nil
		CreatedAt:    time.Now(),
		ExecutableAt: time.Now().Add(24 * time.Hour),
	}
	err = k.SetPendingTreasuryTx(ctx, validTx)
	require.NoError(t, err)

	// Test with valid inflation alert
	validAlert := &types.InflationAlert{
		AlertId:              "test-alert-1",
		AlertType:            types.InflationAlertTypeRapidChange,
		Severity:             types.AlertSeverityCritical,
		CurrentInflationRate: 1000, // 10% in basis points
		TriggeredAt:          time.Now(),
		Message:              "Inflation rate spike detected",
	}
	err = k.SetInflationAlert(ctx, validAlert)
	require.NoError(t, err)

	// Test with valid large tx record
	validRecord := &types.LargeTxRecord{
		TxHash:             "test-hash-1",
		Sender:             "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
		Recipient:          "aura1w3jhxapjta047h6lta047h6lta047h6l42n9lg",
		Amount:             "1000000",
		Timestamp:          time.Now(),
		BlockHeight:        12345,
		PercentageOfSupply: 100, // 1% in basis points
	}
	err = k.SetLargeTxRecord(ctx, validRecord)
	require.NoError(t, err)

	// Test with valid MEV balance
	err = k.SetUserMEVBalance(ctx, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr", "50000")
	require.NoError(t, err)

	// Set total MEV pending to match user balances
	err = k.SetTotalMEVPending(ctx, "50000")
	require.NoError(t, err)

	// All invariants should still pass
	msg, broken = invariant(ctx)
	require.False(t, broken, "all invariants should pass with valid data: %s", msg)
	require.Empty(t, msg)

	// Test with invalid vesting schedule (vested > total)
	invalidSchedule := &types.VestingSchedule{
		ScheduleId:         "invalid-schedule",
		BeneficiaryAddress: "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
		TotalAmount:        "1000000",
		VestedAmount:       "2000000", // More than total - INVALID
		VestingDuration:    31536000,
		CliffDuration:      7776000,
		StartTime:          time.Now(),
		VestingType:        types.VestingTypeLinear,
	}
	err = k.SetVestingSchedule(ctx, invalidSchedule)
	require.NoError(t, err)

	// Now invariant should fail
	msg, broken = invariant(ctx)
	require.True(t, broken, "invariant should be broken with invalid vesting schedule")
	require.Contains(t, msg, "exceeds total")
}

func TestParamsInvariant_CirculatingSupplyExceedsMax(t *testing.T) {
	params := types.DefaultParams()
	params.Tokenomics.MaxSupply = "100"
	params.Tokenomics.CirculatingSupply = "200"

	k, ctx := setupKeeperWithCustomParams(t, params)

	msg, broken := ParamsInvariant(k)(ctx)
	require.True(t, broken, "circulating supply above max must break invariant")
	require.Contains(t, msg, "exceeds max supply")
}

func TestParamsInvariant_WhaleProtectionCooldownTooLong(t *testing.T) {
	params := types.DefaultParams()
	params.WhaleProtection.LargeTxCooldown = 31 * 24 * 60 * 60

	k, ctx := setupKeeperWithCustomParams(t, params)

	msg, broken := ParamsInvariant(k)(ctx)
	require.True(t, broken, "cooldown beyond policy window should break invariant")
	require.Contains(t, msg, "large tx cooldown")
}

func TestMEVBalancesInvariantDetectsMismatch(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	require.NoError(t, k.SetUserMEVBalance(ctx, "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr", "100"))
	require.NoError(t, k.SetUserMEVBalance(ctx, "aura1w3jhxapjta047h6lta047h6lta047h6l42n9lg", "200"))
	require.NoError(t, k.SetTotalMEVPending(ctx, "50")) // mismatch on purpose

	msg, broken := MEVBalancesInvariant(k)(ctx)
	require.True(t, broken, "mismatched MEV totals should break invariant")
	require.Contains(t, msg, "doesn't match sum")
}

func TestRegisterInvariants(t *testing.T) {
	// Test that registering invariants doesn't panic
	require.NotPanics(t, func() {
		k, _ := setupKeeperForTest(t)
		// Mock invariant registry
		registry := &mockInvariantRegistry{}
		RegisterInvariants(registry, k)

		// Verify all 7 invariants were registered
		require.Equal(t, 7, registry.count)
	})
}

// mockInvariantRegistry implements sdk.InvariantRegistry for testing
type mockInvariantRegistry struct {
	count int
}

func (m *mockInvariantRegistry) RegisterRoute(moduleName, route string, invar sdk.Invariant) {
	m.count++
}
