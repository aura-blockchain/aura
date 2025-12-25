// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

type noopInvariantRegistry struct{}

func (noopInvariantRegistry) RegisterRoute(_ string, _ string, _ sdk.Invariant) {}

func TestInvariantsRun(t *testing.T) {
	ctx, k := setupConfKeeper(t)

	inv := AllInvariants(k)
	msg, broken := inv(ctx)
	require.False(t, broken)
	require.Empty(t, msg)
}

func TestRegisterInvariants(t *testing.T) {
	_, k := setupConfKeeper(t)
	reg := noopInvariantRegistry{}

	require.NotPanics(t, func() {
		RegisterInvariants(reg, k)
	})
}
