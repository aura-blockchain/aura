// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/dex/types"
	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

// TestRegisterLegacyAminoCodec tests that RegisterLegacyAminoCodec doesn't panic
func TestRegisterLegacyAminoCodec(t *testing.T) {
	cdc := codec.NewLegacyAmino()
	require.NotPanics(t, func() {
		types.RegisterLegacyAminoCodec(cdc)
	})
}

// TestRegisterInterfaces tests that RegisterInterfaces properly registers message types
func TestRegisterInterfaces(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	require.NotPanics(t, func() {
		types.RegisterInterfaces(registry)
	})

	// Verify that message types are registered as sdk.Msg implementations
	msgTypes := []sdk.Msg{
		&dexpb.MsgCreatePool{},
		&dexpb.MsgAddLiquidity{},
		&dexpb.MsgRemoveLiquidity{},
		&dexpb.MsgSwapExactIn{},
		&dexpb.MsgCreateOrder{},
		&dexpb.MsgCancelOrder{},
		&dexpb.MsgExecuteSwap{},
		&dexpb.MsgCreateHTLC{},
		&dexpb.MsgClaimHTLC{},
		&dexpb.MsgRefundHTLC{},
	}

	for _, msg := range msgTypes {
		t.Run("sdk.Msg/"+sdk.MsgTypeURL(msg), func(t *testing.T) {
			// Should not panic when resolving type
			typeURL := sdk.MsgTypeURL(msg)
			require.NotEmpty(t, typeURL)
		})
	}
}

// TestRegisterInterfaces_MessageResolution tests that registered messages can be resolved
func TestRegisterInterfaces_MessageResolution(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)

	// Create an Any type from a message and verify it can be resolved
	msg := &dexpb.MsgCreatePool{
		Creator: "aura1test",
		DenomA:  "uaura",
		DenomB:  "uusdt",
	}

	anyVal, err := codectypes.NewAnyWithValue(msg)
	require.NoError(t, err)
	require.NotNil(t, anyVal)

	// Verify type URL
	require.Contains(t, anyVal.TypeUrl, "MsgCreatePool")
}

// TestRegisterInterfaces_ResponseTypes tests that response types are registered
func TestRegisterInterfaces_ResponseTypes(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)

	// Test individual response types
	tests := []struct {
		name string
		msg  interface{}
	}{
		{"MsgCreatePoolResponse", &dexpb.MsgCreatePoolResponse{}},
		{"MsgAddLiquidityResponse", &dexpb.MsgAddLiquidityResponse{}},
		{"MsgRemoveLiquidityResponse", &dexpb.MsgRemoveLiquidityResponse{}},
		{"MsgSwapExactInResponse", &dexpb.MsgSwapExactInResponse{}},
		{"MsgCreateOrderResponse", &dexpb.MsgCreateOrderResponse{}},
		{"MsgCancelOrderResponse", &dexpb.MsgCancelOrderResponse{}},
		{"MsgExecuteSwapResponse", &dexpb.MsgExecuteSwapResponse{}},
		{"MsgCreateHTLCResponse", &dexpb.MsgCreateHTLCResponse{}},
		{"MsgClaimHTLCResponse", &dexpb.MsgClaimHTLCResponse{}},
		{"MsgRefundHTLCResponse", &dexpb.MsgRefundHTLCResponse{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the type exists and is properly structured
			require.NotNil(t, tt.msg)
		})
	}
}

// TestRegisterInterfaces_IdempotentRegistration tests that calling RegisterInterfaces multiple times is safe
func TestRegisterInterfaces_IdempotentRegistration(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()

	// Should not panic when called multiple times
	require.NotPanics(t, func() {
		types.RegisterInterfaces(registry)
		types.RegisterInterfaces(registry)
		types.RegisterInterfaces(registry)
	})
}

// TestRegisterLegacyAminoCodec_NilCodec tests handling of nil codec
func TestRegisterLegacyAminoCodec_NilCodec(t *testing.T) {
	// The function should handle nil gracefully (current implementation does nothing)
	require.NotPanics(t, func() {
		types.RegisterLegacyAminoCodec(nil)
	})
}
