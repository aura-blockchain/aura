// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

func TestRegisterInterfaces(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()

	// Register interfaces
	RegisterInterfaces(registry)

	// Create a codec with the registry
	cdc := codec.NewProtoCodec(registry)
	require.NotNil(t, cdc)

	// Test that messages can be marshaled
	msgTypes := []sdk.Msg{
		&bridgepb.MsgLockTokens{},
		&bridgepb.MsgMintTokens{},
		&bridgepb.MsgUnlockTokens{},
		&bridgepb.MsgBurnTokens{},
		&bridgepb.MsgLinkAddress{},
		&bridgepb.MsgCrossChainSwap{},
		&bridgepb.MsgRelayTransfer{},
		&bridgepb.MsgFinalizeTransfer{},
		&bridgepb.MsgSubmitFraudProof{},
	}

	for _, msg := range msgTypes {
		t.Run(sdk.MsgTypeURL(msg), func(t *testing.T) {
			// Marshal the message
			bz, err := cdc.Marshal(msg)
			require.NoError(t, err)

			// For properly initialized messages, we should get non-empty bytes
			// Empty messages with defaults may marshal to empty bytes, which is valid
			_ = bz
		})
	}
}

func TestRegisterInterfacesNoDoubleRegistration(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()

	// Register interfaces twice - should not panic
	require.NotPanics(t, func() {
		RegisterInterfaces(registry)
		RegisterInterfaces(registry)
	})
}
