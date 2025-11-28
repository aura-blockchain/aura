package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// RegisterLegacyAminoCodec registers the necessary x/networksecurity interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register message types here if networksecurity module has messages
}

// RegisterInterfaces registers the x/networksecurity interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// Register message types with the interface registry
	// TODO: Uncomment when message service is defined
	// msgservice.RegisterMsgServiceDesc(registry, &_MsgService_serviceDesc)
}
