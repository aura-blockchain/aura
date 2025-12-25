// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"

	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// RegisterCodec registers the necessary concrete types on the provided LegacyAmino codec.
func RegisterCodec(cdc *codec.LegacyAmino) {
	// Register proto types
	// Message registration will be done through proto's generated registration
}

// RegisterInterfaces registers the interface types with the InterfaceRegistry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	// Register proto types
	// Message registration will be done through proto's generated registration

	// Register custom types as Any
	registry.RegisterInterface(
		"aura.security.v1beta1.Params",
		(*securitypb.Params)(nil),
	)
	registry.RegisterInterface(
		"aura.security.v1beta1.GenesisState",
		(*securitypb.GenesisState)(nil),
	)
}
