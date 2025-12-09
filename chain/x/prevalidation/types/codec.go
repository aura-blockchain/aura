package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	prevalidationpb "github.com/aequitas/aura/proto/aura/prevalidation/v1beta1"
)

// RegisterLegacyAminoCodec registers the necessary x/prevalidation interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register message types here if prevalidation module has messages
}

// RegisterInterfaces registers the x/prevalidation interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// Register message types with the interface registry
	msgservice.RegisterMsgServiceDesc(registry, &prevalidationpb.Msg_serviceDesc)
}
