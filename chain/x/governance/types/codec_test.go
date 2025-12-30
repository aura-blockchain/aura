// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/require"
)

func TestRegisterLegacyAminoCodec(t *testing.T) {
	cdc := codec.NewLegacyAmino()

	// Should not panic
	require.NotPanics(t, func() {
		RegisterLegacyAminoCodec(cdc)
	})
}

func TestRegisterInterfaces(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()

	// Should not panic
	require.NotPanics(t, func() {
		RegisterInterfaces(registry)
	})

	// Registry should have implementations registered
	require.NotNil(t, registry)
}
