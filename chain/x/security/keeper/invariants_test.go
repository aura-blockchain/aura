// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testutil"
	"github.com/aequitas/aura/chain/x/security/keeper"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// =============================================================================
// RegisterInvariants Tests
// =============================================================================

func TestRegisterInvariants(t *testing.T) {
	_, k := testutil.NewSecurityKeeperForTest(t)

	// Create a mock invariant registry
	registry := &mockInvariantRegistry{routes: make(map[string]map[string]sdk.Invariant)}

	// Register invariants
	keeper.RegisterInvariants(registry, k)

	// Verify all expected invariants were registered
	require.Contains(t, registry.routes, "security")
	require.Contains(t, registry.routes["security"], "params-valid")
	require.Contains(t, registry.routes["security"], "spending-limits-validity")
	require.Contains(t, registry.routes["security"], "audit-log-integrity")
	require.Contains(t, registry.routes["security"], "privacy-data-consistency")
}

// =============================================================================
// AllInvariants Tests
// =============================================================================

func TestAllInvariantsPass(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// With no data, all invariants should pass
	invariant := keeper.AllInvariants(k)
	msg, broken := invariant(ctx)

	require.False(t, broken, "invariant should not be broken with empty state")
	require.Empty(t, msg)
}

// =============================================================================
// ParamsInvariant Tests
// =============================================================================

func TestParamsInvariantPass(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Set default params
	k.SetParams(ctx, k.GetParams(ctx))

	invariant := keeper.ParamsInvariant(k)
	msg, broken := invariant(ctx)

	require.False(t, broken, "params invariant should pass with default params")
	require.Empty(t, msg)
}

// =============================================================================
// SpendingLimitsValidityInvariant Tests
// =============================================================================

func TestSpendingLimitsValidityInvariant_Pass(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()
	resetTime := now.Add(24 * time.Hour)

	// Add a valid spending limit
	limit := &securitypb.SpendingLimit{
		WalletId:            "wallet-1",
		DailyLimit:          "1000000",
		WeeklyLimit:         "5000000",
		MonthlyLimit:        "20000000",
		CurrentDailySpent:   "100000",
		CurrentWeeklySpent:  "500000",
		CurrentMonthlySpent: "2000000",
		DailyResetAt:        &resetTime,
		WeeklyResetAt:       &resetTime,
		MonthlyResetAt:      &resetTime,
		Enabled:             true,
	}
	k.SetSpendingLimit(ctx, limit)

	invariant := keeper.SpendingLimitsValidityInvariant(k)
	msg, broken := invariant(ctx)

	require.False(t, broken, "invariant should pass with valid spending limit")
	require.Empty(t, msg)
}

func TestSpendingLimitsValidityInvariant_EmptyWalletID(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Add a spending limit with empty wallet ID
	limit := &securitypb.SpendingLimit{
		WalletId:   "", // Invalid: empty
		DailyLimit: "1000000",
		Enabled:    false,
	}
	k.SetSpendingLimit(ctx, limit)

	invariant := keeper.SpendingLimitsValidityInvariant(k)
	msg, broken := invariant(ctx)

	require.True(t, broken, "invariant should fail with empty wallet ID")
	require.Contains(t, msg, "empty wallet ID")
}

func TestSpendingLimitsValidityInvariant_InvalidDailyLimit(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Add a spending limit with invalid daily limit
	limit := &securitypb.SpendingLimit{
		WalletId:   "wallet-2",
		DailyLimit: "not-a-number", // Invalid
		Enabled:    false,
	}
	k.SetSpendingLimit(ctx, limit)

	invariant := keeper.SpendingLimitsValidityInvariant(k)
	msg, broken := invariant(ctx)

	require.True(t, broken, "invariant should fail with invalid daily limit")
	require.Contains(t, msg, "invalid daily limit")
}

func TestSpendingLimitsValidityInvariant_NegativeLimit(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Add a spending limit with negative limit
	limit := &securitypb.SpendingLimit{
		WalletId:   "wallet-3",
		DailyLimit: "-1000", // Invalid: negative
		Enabled:    false,
	}
	k.SetSpendingLimit(ctx, limit)

	invariant := keeper.SpendingLimitsValidityInvariant(k)
	msg, broken := invariant(ctx)

	require.True(t, broken, "invariant should fail with negative daily limit")
	require.Contains(t, msg, "invalid daily limit")
}

func TestSpendingLimitsValidityInvariant_EnabledWithoutResetTime(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Add enabled limit without reset time
	limit := &securitypb.SpendingLimit{
		WalletId:     "wallet-4",
		DailyLimit:   "1000000",
		DailyResetAt: nil, // Invalid: enabled but no reset time
		Enabled:      true,
	}
	k.SetSpendingLimit(ctx, limit)

	invariant := keeper.SpendingLimitsValidityInvariant(k)
	msg, broken := invariant(ctx)

	require.True(t, broken, "invariant should fail when enabled without reset time")
	require.Contains(t, msg, "reset time is nil")
}

// =============================================================================
// AuditLogIntegrityInvariant Tests
// =============================================================================

func TestAuditLogIntegrityInvariant_Pass(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()

	// Add valid audit log entry
	entry := &securitypb.AuditLogEntry{
		LogId:     "log-1",
		Timestamp: now,
		EventType: "SECURITY_EVENT",
		Actor:     "admin",
		Action:    "test-action",
	}
	k.SetAuditLogEntry(ctx, entry)

	invariant := keeper.AuditLogIntegrityInvariant(k)
	msg, broken := invariant(ctx)

	require.False(t, broken, "invariant should pass with valid audit log")
	require.Empty(t, msg)
}

func TestAuditLogIntegrityInvariant_EmptyLogID(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()

	entry := &securitypb.AuditLogEntry{
		LogId:     "", // Invalid: empty
		Timestamp: now,
		EventType: "SECURITY_EVENT",
		Actor:     "admin",
	}
	k.SetAuditLogEntry(ctx, entry)

	invariant := keeper.AuditLogIntegrityInvariant(k)
	msg, broken := invariant(ctx)

	require.True(t, broken, "invariant should fail with empty log ID")
	require.Contains(t, msg, "empty ID")
}

func TestAuditLogIntegrityInvariant_ZeroTimestamp(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	entry := &securitypb.AuditLogEntry{
		LogId:     "log-zero-ts",
		Timestamp: time.Time{}, // Zero timestamp
		EventType: "SECURITY_EVENT",
		Actor:     "admin",
	}
	k.SetAuditLogEntry(ctx, entry)

	invariant := keeper.AuditLogIntegrityInvariant(k)
	msg, broken := invariant(ctx)

	require.True(t, broken, "invariant should fail with zero timestamp")
	require.Contains(t, msg, "zero timestamp")
}

func TestAuditLogIntegrityInvariant_EmptyEventType(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()

	entry := &securitypb.AuditLogEntry{
		LogId:     "log-no-event",
		Timestamp: now,
		EventType: "", // Invalid: empty
		Actor:     "admin",
	}
	k.SetAuditLogEntry(ctx, entry)

	invariant := keeper.AuditLogIntegrityInvariant(k)
	msg, broken := invariant(ctx)

	require.True(t, broken, "invariant should fail with empty event type")
	require.Contains(t, msg, "empty event type")
}

func TestAuditLogIntegrityInvariant_EmptyActor(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()

	entry := &securitypb.AuditLogEntry{
		LogId:     "log-no-actor",
		Timestamp: now,
		EventType: "SECURITY_EVENT",
		Actor:     "", // Invalid: empty
	}
	k.SetAuditLogEntry(ctx, entry)

	invariant := keeper.AuditLogIntegrityInvariant(k)
	msg, broken := invariant(ctx)

	require.True(t, broken, "invariant should fail with empty actor")
	require.Contains(t, msg, "empty actor")
}

// =============================================================================
// PrivacyDataConsistencyInvariant Tests
// =============================================================================

func TestPrivacyDataConsistencyInvariant_Pass(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()

	// Add valid stealth address
	stealthAddr := &securitypb.StealthAddress{
		OneTimeAddress:  []byte("one-time-address-bytes"),
		PublicSpendKey:  []byte("spend-key"),
		PublicViewKey:   []byte("view-key"),
		TxPublicKey:     []byte("tx-key"),
		EncryptedAmount: []byte("encrypted"),
		CreatedAt:       &now,
	}
	k.SetStealthAddress(ctx, stealthAddr)

	// Add valid ring signature
	ringSig := &securitypb.RingSignature{
		KeyImage: []byte("key-image-bytes"),
		RingSize: 11,
		RingMembers: [][]byte{
			[]byte("member-1"), []byte("member-2"), []byte("member-3"),
			[]byte("member-4"), []byte("member-5"), []byte("member-6"),
			[]byte("member-7"), []byte("member-8"), []byte("member-9"),
			[]byte("member-10"), []byte("member-11"),
		},
		CreatedAt: &now,
	}
	k.SetRingSignature(ctx, ringSig)

	invariant := keeper.PrivacyDataConsistencyInvariant(k)
	msg, broken := invariant(ctx)

	require.False(t, broken, "invariant should pass with valid privacy data")
	require.Empty(t, msg)
}

func TestPrivacyDataConsistencyInvariant_EmptyOneTimeAddress(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()

	stealthAddr := &securitypb.StealthAddress{
		OneTimeAddress: []byte{}, // Invalid: empty
		CreatedAt:      &now,
	}
	k.SetStealthAddress(ctx, stealthAddr)

	invariant := keeper.PrivacyDataConsistencyInvariant(k)
	msg, broken := invariant(ctx)

	require.True(t, broken, "invariant should fail with empty one-time address")
	require.Contains(t, msg, "empty one-time address")
}

func TestPrivacyDataConsistencyInvariant_NilCreatedAtStealth(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	stealthAddr := &securitypb.StealthAddress{
		OneTimeAddress: []byte("one-time-address-bytes"),
		CreatedAt:      nil, // Invalid: nil
	}
	k.SetStealthAddress(ctx, stealthAddr)

	invariant := keeper.PrivacyDataConsistencyInvariant(k)
	msg, broken := invariant(ctx)

	require.True(t, broken, "invariant should fail with nil created_at")
	require.Contains(t, msg, "nil created_at")
}

func TestPrivacyDataConsistencyInvariant_TooSmallRingSize(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()

	ringSig := &securitypb.RingSignature{
		KeyImage:    []byte("key-image"),
		RingSize:    1, // Invalid: too small (min 2)
		RingMembers: [][]byte{[]byte("member-1")},
		CreatedAt:   &now,
	}
	k.SetRingSignature(ctx, ringSig)

	invariant := keeper.PrivacyDataConsistencyInvariant(k)
	msg, broken := invariant(ctx)

	require.True(t, broken, "invariant should fail with too small ring size")
	require.Contains(t, msg, "too small ring size")
}

func TestPrivacyDataConsistencyInvariant_TooLargeRingSize(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()

	ringSig := &securitypb.RingSignature{
		KeyImage:    []byte("key-image"),
		RingSize:    129, // Invalid: too large (max 128)
		RingMembers: make([][]byte, 129),
		CreatedAt:   &now,
	}
	// Fill ring members
	for i := 0; i < 129; i++ {
		ringSig.RingMembers[i] = []byte("member")
	}
	k.SetRingSignature(ctx, ringSig)

	invariant := keeper.PrivacyDataConsistencyInvariant(k)
	msg, broken := invariant(ctx)

	require.True(t, broken, "invariant should fail with too large ring size")
	require.Contains(t, msg, "too large ring size")
}

func TestPrivacyDataConsistencyInvariant_RingMembersMismatch(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()

	ringSig := &securitypb.RingSignature{
		KeyImage:    []byte("key-image"),
		RingSize:    5,
		RingMembers: [][]byte{[]byte("member-1"), []byte("member-2")}, // Only 2 but size says 5
		CreatedAt:   &now,
	}
	k.SetRingSignature(ctx, ringSig)

	invariant := keeper.PrivacyDataConsistencyInvariant(k)
	msg, broken := invariant(ctx)

	require.True(t, broken, "invariant should fail when members don't match ring size")
	require.Contains(t, msg, "doesn't match ring size")
}

// =============================================================================
// NewSecurityGuards Tests
// =============================================================================

func TestNewSecurityGuards(t *testing.T) {
	guards := keeper.NewSecurityGuards("admin", 1000000, []string{"admin1", "admin2"})
	require.NotNil(t, guards)
}

// =============================================================================
// CheckAuthorization Tests
// =============================================================================

func TestCheckAuthorization(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Test with non-authority address - should return error
	err := k.CheckAuthorization(ctx, "aura1someone...", "some-action")
	require.Error(t, err, "should fail with non-authority address")

	// Test with the authority address - should pass
	authority := k.GetAuthority()
	if authority != "" {
		err = k.CheckAuthorization(ctx, authority, "some-action")
		require.NoError(t, err, "should pass with authority address")
	}
}

// =============================================================================
// Helper: Mock Invariant Registry
// =============================================================================

type mockInvariantRegistry struct {
	routes map[string]map[string]sdk.Invariant
}

func (m *mockInvariantRegistry) RegisterRoute(moduleName, route string, invar sdk.Invariant) {
	if m.routes[moduleName] == nil {
		m.routes[moduleName] = make(map[string]sdk.Invariant)
	}
	m.routes[moduleName][route] = invar
}
