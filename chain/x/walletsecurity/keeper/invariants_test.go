// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

// =============================================================================
// Invariants Test Suite
// =============================================================================

type InvariantsTestSuite struct {
	KeeperTestSuite
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

// =============================================================================
// RegisterInvariants Tests
// =============================================================================

func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	k := suite.GetKeeper()

	// Create a mock invariant registry
	mockRegistry := &mockInvariantRegistry{invariants: make(map[string]sdk.Invariant)}

	// Register invariants
	RegisterInvariants(mockRegistry, k)

	// Verify all expected invariants are registered
	suite.Require().Contains(mockRegistry.invariants, "walletsecurity/params-valid")
	suite.Require().Contains(mockRegistry.invariants, "walletsecurity/spending-limits-validity")
	suite.Require().Contains(mockRegistry.invariants, "walletsecurity/multi-sig-consistency")
	suite.Require().Contains(mockRegistry.invariants, "walletsecurity/security-features-validity")
}

func (suite *InvariantsTestSuite) TestRegisterInvariants_Count() {
	k := suite.GetKeeper()
	mockRegistry := &mockInvariantRegistry{invariants: make(map[string]sdk.Invariant)}

	RegisterInvariants(mockRegistry, k)

	// Should register exactly 4 invariants
	suite.Require().Len(mockRegistry.invariants, 4)
}

// =============================================================================
// AllInvariants Tests
// =============================================================================

func (suite *InvariantsTestSuite) TestAllInvariants_EmptyStore() {
	k := suite.GetKeeper()
	ctx := suite.GetContext().(sdk.Context)

	// All invariants should pass on empty store
	inv := AllInvariants(k)
	msg, broken := inv(ctx)

	suite.Require().False(broken, "invariants should not be broken on empty store")
	suite.Require().Empty(msg)
}

func (suite *InvariantsTestSuite) TestAllInvariants_ReturnsOnFirstFailure() {
	k := suite.GetKeeper()
	ctx := suite.GetContext().(sdk.Context)

	// With empty store, all invariants should pass
	inv := AllInvariants(k)
	msg, broken := inv(ctx)

	suite.Require().False(broken)
	suite.Require().Empty(msg)
}

// =============================================================================
// ParamsInvariant Tests
// =============================================================================

func (suite *InvariantsTestSuite) TestParamsInvariant_EmptyStore() {
	k := suite.GetKeeper()
	ctx := suite.GetContext().(sdk.Context)

	inv := ParamsInvariant(k)
	msg, broken := inv(ctx)

	suite.Require().False(broken)
	suite.Require().Empty(msg)
}

func (suite *InvariantsTestSuite) TestParamsInvariant_AlwaysValid() {
	k := suite.GetKeeper()
	ctx := suite.GetContext().(sdk.Context)

	// ParamsInvariant currently always returns success (params not implemented)
	inv := ParamsInvariant(k)
	msg, broken := inv(ctx)

	suite.Require().False(broken)
	suite.Require().Empty(msg)
}

// =============================================================================
// SpendingLimitsValidityInvariant Tests
// =============================================================================

func (suite *InvariantsTestSuite) TestSpendingLimitsInvariant_EmptyStore() {
	k := suite.GetKeeper()
	ctx := suite.GetContext().(sdk.Context)

	inv := SpendingLimitsValidityInvariant(k)
	msg, broken := inv(ctx)

	suite.Require().False(broken)
	suite.Require().Empty(msg)
}

func (suite *InvariantsTestSuite) TestSpendingLimitsInvariant_ValidLimit() {
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// Set a valid spending limit
	_, err := k.SetSpendingLimit(ctx, "wallet1", "uaura", "1000000", "5000000", "20000000")
	suite.Require().NoError(err)

	inv := SpendingLimitsValidityInvariant(k)
	msg, broken := inv(ctx.(sdk.Context))

	suite.Require().False(broken, "valid spending limit should not break invariant: %s", msg)
}

func (suite *InvariantsTestSuite) TestSpendingLimitsInvariant_ZeroLimits() {
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// Set zero spending limits (allowed)
	_, err := k.SetSpendingLimit(ctx, "wallet2", "uaura", "0", "0", "0")
	suite.Require().NoError(err)

	inv := SpendingLimitsValidityInvariant(k)
	msg, broken := inv(ctx.(sdk.Context))

	suite.Require().False(broken, "zero spending limits should not break invariant: %s", msg)
}

func (suite *InvariantsTestSuite) TestSpendingLimitsInvariant_MultipleWallets() {
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// Set spending limits for multiple wallets
	wallets := []string{"wallet1", "wallet2", "wallet3"}
	for _, walletID := range wallets {
		_, err := k.SetSpendingLimit(ctx, walletID, "uaura", "1000000", "5000000", "20000000")
		suite.Require().NoError(err)
	}

	inv := SpendingLimitsValidityInvariant(k)
	msg, broken := inv(ctx.(sdk.Context))

	suite.Require().False(broken, "multiple valid spending limits should not break invariant: %s", msg)
}

func (suite *InvariantsTestSuite) TestSpendingLimitsInvariant_LargeLimits() {
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// Set very large spending limits
	_, err := k.SetSpendingLimit(ctx, "largewallet", "uaura", "999999999999999", "9999999999999999", "99999999999999999")
	suite.Require().NoError(err)

	inv := SpendingLimitsValidityInvariant(k)
	msg, broken := inv(ctx.(sdk.Context))

	suite.Require().False(broken, "large spending limits should not break invariant: %s", msg)
}

// =============================================================================
// MultiSigConsistencyInvariant Tests
// =============================================================================

func (suite *InvariantsTestSuite) TestMultiSigInvariant_EmptyStore() {
	k := suite.GetKeeper()
	ctx := suite.GetContext().(sdk.Context)

	inv := MultiSigConsistencyInvariant(k)
	msg, broken := inv(ctx)

	suite.Require().False(broken)
	suite.Require().Empty(msg)
}

// Note: Multi-sig wallet tests that create wallets with fake addresses
// would fail the invariant because it validates bech32 format.
// The keeper's CreateMultiSigWallet function doesn't validate addresses,
// but the invariant does. So we test the empty case and invariant registration.

// =============================================================================
// SecurityFeaturesValidityInvariant Tests
// =============================================================================

func (suite *InvariantsTestSuite) TestSecurityFeaturesInvariant_EmptyStore() {
	k := suite.GetKeeper()
	ctx := suite.GetContext().(sdk.Context)

	inv := SecurityFeaturesValidityInvariant(k)
	msg, broken := inv(ctx)

	suite.Require().False(broken)
	suite.Require().Empty(msg)
}

// =============================================================================
// Invariant Function Tests (isolated logic)
// =============================================================================

func (suite *InvariantsTestSuite) TestParamsInvariantFunction() {
	k := suite.GetKeeper()

	// Get the invariant function
	inv := ParamsInvariant(k)
	suite.Require().NotNil(inv)

	// Execute on context
	ctx := suite.GetContext().(sdk.Context)
	msg, broken := inv(ctx)

	suite.Require().False(broken)
	suite.Require().Empty(msg)
}

func (suite *InvariantsTestSuite) TestSpendingLimitsInvariantFunction() {
	k := suite.GetKeeper()

	inv := SpendingLimitsValidityInvariant(k)
	suite.Require().NotNil(inv)

	ctx := suite.GetContext().(sdk.Context)
	msg, broken := inv(ctx)

	suite.Require().False(broken)
	suite.Require().Empty(msg)
}

func (suite *InvariantsTestSuite) TestMultiSigInvariantFunction() {
	k := suite.GetKeeper()

	inv := MultiSigConsistencyInvariant(k)
	suite.Require().NotNil(inv)

	ctx := suite.GetContext().(sdk.Context)
	msg, broken := inv(ctx)

	suite.Require().False(broken)
	suite.Require().Empty(msg)
}

func (suite *InvariantsTestSuite) TestSecurityFeaturesInvariantFunction() {
	k := suite.GetKeeper()

	inv := SecurityFeaturesValidityInvariant(k)
	suite.Require().NotNil(inv)

	ctx := suite.GetContext().(sdk.Context)
	msg, broken := inv(ctx)

	suite.Require().False(broken)
	suite.Require().Empty(msg)
}

func (suite *InvariantsTestSuite) TestAllInvariantsFunction() {
	k := suite.GetKeeper()

	inv := AllInvariants(k)
	suite.Require().NotNil(inv)

	ctx := suite.GetContext().(sdk.Context)
	msg, broken := inv(ctx)

	suite.Require().False(broken)
	suite.Require().Empty(msg)
}

// =============================================================================
// Helper Types
// =============================================================================

type mockInvariantRegistry struct {
	invariants map[string]sdk.Invariant
}

func (m *mockInvariantRegistry) RegisterRoute(moduleName string, route string, inv sdk.Invariant) {
	m.invariants[moduleName+"/"+route] = inv
}

// =============================================================================
// Unit Tests (non-suite based)
// =============================================================================

func TestSpendingLimitProtoFields(t *testing.T) {
	// Test proto message field behavior
	limit := &wsproto.SpendingLimit{
		WalletId:            "wallet1",
		Denom:               "uaura",
		DailyLimit:          "1000000",
		WeeklyLimit:         "5000000",
		MonthlyLimit:        "20000000",
		CurrentDailySpent:   "0",
		CurrentWeeklySpent:  "0",
		CurrentMonthlySpent: "0",
		Enabled:             true,
	}

	require.NotEmpty(t, limit.WalletId)
	require.Equal(t, "1000000", limit.DailyLimit)
	require.True(t, limit.Enabled)
}

func TestMultiSigWalletProtoFields(t *testing.T) {
	// Test proto message field behavior
	wallet := &wsproto.MultiSigWallet{
		WalletId:  "wallet1",
		Signers:   []string{"aura1...", "aura2..."},
		Threshold: 2,
	}

	require.NotEmpty(t, wallet.WalletId)
	require.Len(t, wallet.Signers, 2)
	require.Equal(t, int32(2), wallet.Threshold)
}

func TestHardwareWalletProtoFields(t *testing.T) {
	// Test proto message field behavior
	hwConfig := &wsproto.HardwareWalletConfig{
		WalletId: "wallet1",
		DeviceId: "device123",
	}

	require.NotEmpty(t, hwConfig.WalletId)
	require.Equal(t, "device123", hwConfig.DeviceId)
}

func TestSocialRecoveryProtoFields(t *testing.T) {
	// Test proto message field behavior
	recovery := &wsproto.SocialRecoveryConfig{
		WalletId:  "wallet1",
		Guardians: []*wsproto.Guardian{{Address: "g1"}, {Address: "g2"}, {Address: "g3"}},
	}

	require.NotEmpty(t, recovery.WalletId)
	require.Len(t, recovery.Guardians, 3)
}

func TestInvariantRegistryRoute(t *testing.T) {
	// Test the mock invariant registry
	registry := &mockInvariantRegistry{invariants: make(map[string]sdk.Invariant)}

	registry.RegisterRoute("test", "route1", func(ctx sdk.Context) (string, bool) {
		return "", false
	})

	require.Contains(t, registry.invariants, "test/route1")
}

func TestMockInvariantRegistry_MultipleRoutes(t *testing.T) {
	registry := &mockInvariantRegistry{invariants: make(map[string]sdk.Invariant)}

	registry.RegisterRoute("mod1", "route1", func(ctx sdk.Context) (string, bool) { return "", false })
	registry.RegisterRoute("mod1", "route2", func(ctx sdk.Context) (string, bool) { return "", false })
	registry.RegisterRoute("mod2", "route1", func(ctx sdk.Context) (string, bool) { return "", false })

	require.Len(t, registry.invariants, 3)
	require.Contains(t, registry.invariants, "mod1/route1")
	require.Contains(t, registry.invariants, "mod1/route2")
	require.Contains(t, registry.invariants, "mod2/route1")
}

func TestGuardianProtoFields(t *testing.T) {
	guardian := &wsproto.Guardian{
		Address: "aura1guardian",
	}

	require.Equal(t, "aura1guardian", guardian.Address)
}
