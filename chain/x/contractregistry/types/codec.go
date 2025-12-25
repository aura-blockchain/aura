// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	contractregistrypb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
)

// RegisterLegacyAminoCodec registers the module messages on the legacy Amino codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// The module is protobuf-native; no amino messages needed.
	_ = cdc
}

// RegisterInterfaces registers the module's protobuf interfaces.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &contractregistrypb.Msg_serviceDesc)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&contractregistrypb.MsgRegisterContract{},
		&contractregistrypb.MsgUpdateContractMetadata{},
		&contractregistrypb.MsgUpdateSecurityPolicy{},
		&contractregistrypb.MsgPauseContract{},
		&contractregistrypb.MsgUnpauseContract{},
		&contractregistrypb.MsgDeprecateContract{},
	)

	registry.RegisterImplementations((*sdktx.MsgResponse)(nil),
		&contractregistrypb.MsgRegisterContractResponse{},
		&contractregistrypb.MsgUpdateContractMetadataResponse{},
		&contractregistrypb.MsgUpdateSecurityPolicyResponse{},
		&contractregistrypb.MsgPauseContractResponse{},
		&contractregistrypb.MsgUnpauseContractResponse{},
		&contractregistrypb.MsgDeprecateContractResponse{},
	)
}
