// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

// RegisterLegacyAminoCodec registers the necessary x/vcregistry interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// VCRegistry is proto-only, no amino registration needed
}

// RegisterInterfaces registers the x/vcregistry interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &vcregistrypb.Msg_serviceDesc)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&vcregistrypb.MsgMintVC{},
		&vcregistrypb.MsgRevokeVC{},
		&vcregistrypb.MsgAdminRevokeVC{},
		&vcregistrypb.MsgSuspendVC{},
		&vcregistrypb.MsgReactivateVC{},
		&vcregistrypb.MsgCreateVCPolicy{},
		&vcregistrypb.MsgUpdateVCPolicy{},
		&vcregistrypb.MsgDeprecateVCPolicy{},
		&vcregistrypb.MsgRegisterDID{},
		&vcregistrypb.MsgUpdateDIDDocument{},
	)
}
