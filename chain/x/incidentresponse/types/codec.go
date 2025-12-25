// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	incidentresponsepb "github.com/aequitas/aura/proto/aura/incidentresponse/v1beta1"
)

// RegisterLegacyAminoCodec registers the necessary x/incidentresponse interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register message types here if incidentresponse module has messages
}

// RegisterInterfaces registers the x/incidentresponse interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// Register message types with the interface registry
	msgservice.RegisterMsgServiceDesc(registry, &incidentresponsepb.Msg_serviceDesc)
}
