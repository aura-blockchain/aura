package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	monitoringpb "github.com/aequitas/aura/proto/aura/monitoring/v1beta1"
)

// RegisterLegacyAminoCodec registers the necessary x/monitoring interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register message types for Amino JSON serialization
	cdc.RegisterConcrete(&monitoringpb.MsgAcknowledgeAlert{}, "aura/monitoring/MsgAcknowledgeAlert", nil)
	cdc.RegisterConcrete(&monitoringpb.MsgResolveAlert{}, "aura/monitoring/MsgResolveAlert", nil)
}

// RegisterInterfaces registers the x/monitoring interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// Register message service
	msgservice.RegisterMsgServiceDesc(registry, &monitoringpb.Msg_serviceDesc)

	// Register message implementations
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&monitoringpb.MsgAcknowledgeAlert{},
		&monitoringpb.MsgResolveAlert{},
	)

	// Register response implementations
	registry.RegisterImplementations(
		(*sdktx.MsgResponse)(nil),
		&monitoringpb.MsgAcknowledgeAlertResponse{},
		&monitoringpb.MsgResolveAlertResponse{},
	)
}
