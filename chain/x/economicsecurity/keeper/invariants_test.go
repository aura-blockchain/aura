package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

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
		ScheduleId:          "test-schedule-1",
		BeneficiaryAddress:  "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
		TotalAmount:         "1000000",
		VestedAmount:        "100000",
		VestingDuration:     31536000, // 1 year in seconds
		CliffDuration:       7776000,  // 90 days in seconds
		StartTime:           timestamppb.New(time.Now()),
		VestingType:         types.VestingTypeLinear,
	}
	err := k.SetVestingSchedule(ctx, validSchedule)
	require.NoError(t, err)

	// Test with valid vote lock
	validLock := &types.VoteLock{
		LockId:       "test-lock-1",
		Owner:        "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
		Amount:       "500000",
		LockStart:    timestamppb.New(time.Now()),
		LockEnd:      timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
		VotingPower:  "500000",
	}
	err = k.SetVoteLock(ctx, validLock)
	require.NoError(t, err)

	// Test with valid treasury tx
	validTx := &types.PendingTreasuryTx{
		TxId:          "test-tx-1",
		Recipient:     "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
		Amount:        "100000",
		Proposer:      "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
		Signatures:    []string{},
		CreatedAt:     timestamppb.New(time.Now()),
		ExecutableAt:  timestamppb.New(time.Now().Add(24 * time.Hour)),
	}
	err = k.SetPendingTreasuryTx(ctx, validTx)
	require.NoError(t, err)

	// Test with valid inflation alert
	validAlert := &types.InflationAlert{
		AlertId:              "test-alert-1",
		AlertType:            types.InflationAlertTypeRapidChange,
		Severity:             types.AlertSeverityCritical,
		CurrentInflationRate: 1000, // 10% in basis points
		TriggeredAt:          timestamppb.New(time.Now()),
		Message:              "Inflation rate spike detected",
	}
	err = k.SetInflationAlert(ctx, validAlert)
	require.NoError(t, err)

	// Test with valid large tx record
	validRecord := &types.LargeTxRecord{
		TxHash:             "test-hash-1",
		Sender:             "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
		Recipient:          "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
		Amount:             "1000000",
		Timestamp:          timestamppb.New(time.Now()),
		BlockHeight:        12345,
		PercentageOfSupply: 100, // 1% in basis points
	}
	err = k.SetLargeTxRecord(ctx, validRecord)
	require.NoError(t, err)

	// Test with valid MEV balance
	err = k.SetUserMEVBalance(ctx, "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", "50000")
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
		ScheduleId:          "invalid-schedule",
		BeneficiaryAddress:  "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
		TotalAmount:         "1000000",
		VestedAmount:        "2000000", // More than total - INVALID
		VestingDuration:     31536000,
		CliffDuration:       7776000,
		StartTime:           timestamppb.New(time.Now()),
		VestingType:         types.VestingTypeLinear,
	}
	err = k.SetVestingSchedule(ctx, invalidSchedule)
	require.NoError(t, err)

	// Now invariant should fail
	msg, broken = invariant(ctx)
	require.True(t, broken, "invariant should be broken with invalid vesting schedule")
	require.Contains(t, msg, "exceeds total")
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
