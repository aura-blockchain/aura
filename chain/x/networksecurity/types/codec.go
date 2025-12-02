package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	networksecuritypb "github.com/aequitas/aura/proto/aura/networksecurity/v1beta1"
)

// RegisterLegacyAminoCodec registers the necessary x/networksecurity interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register message types here if networksecurity module has messages
}

// RegisterInterfaces registers the x/networksecurity interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &networksecuritypb.Msg_ServiceDesc)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&networksecuritypb.MsgUpdateParams{},
		&networksecuritypb.MsgAddTrustedPeer{},
		&networksecuritypb.MsgRemoveTrustedPeer{},
		&networksecuritypb.MsgBanPeer{},
		&networksecuritypb.MsgUnbanPeer{},
	)
}
