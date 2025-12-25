// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/auth/keeper"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// AuthKeeper creates a test auth keeper with context
func AuthKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	t.Helper()

	input := CreateTestInput(t)

	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
	)

	// Initialize default params
	params := authproto.Params{
		SessionTimeoutSeconds:         3600,
		DefaultTimelockDelaySeconds:   86400,
		DefaultRequestsPerMinute:      60,
		DefaultRequestsPerHour:        3600,
		DefaultRequestsPerDay:         86400,
		MultisigProposalExpirySeconds: 604800,
		AuditLoggingEnabled:           true,
	}
	require.NoError(t, k.SetParams(input.Ctx, &params))

	return k, input.Ctx
}

// AuthKeeperWithCustomParams creates auth keeper with custom params
func AuthKeeperWithCustomParams(t testing.TB, params *authproto.Params) (*keeper.Keeper, sdk.Context) {
	t.Helper()

	k, ctx := AuthKeeper(t)
	require.NoError(t, k.SetParams(ctx, params))

	return k, ctx
}

// SetupAuthTestWithRoles creates auth keeper and sets up test roles
func SetupAuthTestWithRoles(t testing.TB) (*keeper.Keeper, sdk.Context) {
	t.Helper()

	k, ctx := AuthKeeper(t)

	// Create test roles
	roles := []*authproto.Role{
		{
			Name:        "test_admin",
			Permissions: []string{"admin", "create", "read", "update", "delete"},
			Description: "Test admin role",
		},
		{
			Name:        "test_user",
			Permissions: []string{"read"},
			Description: "Test user role",
		},
	}

	for _, role := range roles {
		require.NoError(t, k.SetRole(ctx, role))
	}

	return k, ctx
}
