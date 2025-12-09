package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

// RegisterLegacyAminoCodec registers legacy codec items.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	_ = cdc
}

// RegisterInterfaces registers the inclusion routines proto interfaces.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &inclusionroutinespb.Msg_serviceDesc)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&inclusionroutinespb.MsgCreateIR{},
		&inclusionroutinespb.MsgUpdateIR{},
		&inclusionroutinespb.MsgDeleteIR{},
		&inclusionroutinespb.MsgSetIRPrerequisites{},
		&inclusionroutinespb.MsgSetIRRateLimit{},
		&inclusionroutinespb.MsgSuspendIR{},
		&inclusionroutinespb.MsgActivateIR{},
	)

	registry.RegisterImplementations((*sdktx.MsgResponse)(nil),
		&inclusionroutinespb.MsgCreateIRResponse{},
		&inclusionroutinespb.MsgUpdateIRResponse{},
		&inclusionroutinespb.MsgDeleteIRResponse{},
		&inclusionroutinespb.MsgSetIRPrerequisitesResponse{},
		&inclusionroutinespb.MsgSetIRRateLimitResponse{},
		&inclusionroutinespb.MsgSuspendIRResponse{},
		&inclusionroutinespb.MsgActivateIRResponse{},
	)
}
