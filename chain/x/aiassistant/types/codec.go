package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	aiassistantpb "github.com/aequitas/aura/proto/aura/aiassistant/v1beta1"
)

// RegisterLegacyAminoCodec registers legacy amino types.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	_ = cdc
}

// RegisterInterfaces registers proto interfaces for aiassistant.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &aiassistantpb.Msg_serviceDesc)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&aiassistantpb.MsgRegisterAssistant{},
		&aiassistantpb.MsgUpdateLocales{},
		&aiassistantpb.MsgHeartbeat{},
		&aiassistantpb.MsgReportMisbehavior{},
		&aiassistantpb.MsgUpdateParams{},
	)

	registry.RegisterImplementations((*sdktx.MsgResponse)(nil),
		&aiassistantpb.MsgRegisterAssistantResponse{},
		&aiassistantpb.MsgUpdateLocalesResponse{},
		&aiassistantpb.MsgHeartbeatResponse{},
		&aiassistantpb.MsgReportMisbehaviorResponse{},
		&aiassistantpb.MsgUpdateParamsResponse{},
	)
}
