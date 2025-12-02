package types

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	v1beta1 "github.com/aequitas/aura/proto/aura/validatorsecurity/v1beta1"
)

// RegisterInterfaces wires validatorsecurity messages into the SDK interface registry
// so MsgService registration can resolve type URLs at startup.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &v1beta1.Msg_ServiceDesc)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&v1beta1.MsgRegisterValidator{},
		&v1beta1.MsgUpdateSecurityInfo{},
		&v1beta1.MsgRegisterSentryNode{},
		&v1beta1.MsgReportDoubleSign{},
		&v1beta1.MsgUnjail{},
		&v1beta1.MsgAcknowledgeAlert{},
		&v1beta1.MsgUpdateParams{},
	)
}
