// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/aiassistant/keeper"
	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

type mockInvariantRegistry struct {
	routes map[string]sdk.Invariant
}

func (m *mockInvariantRegistry) RegisterRoute(moduleName, route string, inv sdk.Invariant) {
	if m.routes == nil {
		m.routes = make(map[string]sdk.Invariant)
	}
	m.routes[fmt.Sprintf("%s/%s", moduleName, route)] = inv
}

func TestRegisterInvariants(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	registry := &mockInvariantRegistry{}

	keeper.RegisterInvariants(registry, *k)
	require.NotEmpty(t, registry.routes)

	genesis := types.DefaultGenesis()
	require.NoError(t, k.InitGenesis(ctx, *genesis))

	for name, inv := range registry.routes {
		msg, broken := inv(ctx)
		require.False(t, broken, "route %s broke: %s", name, msg)
	}
}
