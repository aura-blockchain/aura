// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
)

// RegisterLegacyAminoCodec registers legacy amino types.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	_ = cdc
}

// RegisterInterfaces registers protobuf interfaces.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &identitychangepb.Msg_serviceDesc)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&identitychangepb.MsgRequestIdentityChange{},
		&identitychangepb.MsgSubmitAssistantProof{},
		&identitychangepb.MsgApplyIdentityChange{},
		&identitychangepb.MsgRejectIdentityChange{},
		&identitychangepb.MsgSuspendIdentityChanges{},
	)

	registry.RegisterImplementations((*sdktx.MsgResponse)(nil),
		&identitychangepb.MsgRequestIdentityChangeResponse{},
		&identitychangepb.MsgSubmitAssistantProofResponse{},
		&identitychangepb.MsgApplyIdentityChangeResponse{},
		&identitychangepb.MsgRejectIdentityChangeResponse{},
		&identitychangepb.MsgSuspendIdentityChangesResponse{},
	)
}
