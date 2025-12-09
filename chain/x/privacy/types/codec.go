package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	privacypb "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

// RegisterLegacyAminoCodec registers the necessary x/privacy interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register message types here if privacy module has messages
}

// RegisterInterfaces registers the x/privacy interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &privacypb.Msg_serviceDesc)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&privacypb.MsgSubmitPrivateTransaction{},
		&privacypb.MsgCreateMixingPool{},
		&privacypb.MsgJoinMixingPool{},
		&privacypb.MsgRegisterViewKey{},
		&privacypb.MsgRevokeViewKey{},
		&privacypb.MsgUpdateNetworkPrivacy{},
		&privacypb.MsgUpdateParams{},
	)

	registry.RegisterImplementations(
		(*sdktx.MsgResponse)(nil),
		&privacypb.MsgSubmitPrivateTransactionResponse{},
		&privacypb.MsgCreateMixingPoolResponse{},
		&privacypb.MsgJoinMixingPoolResponse{},
		&privacypb.MsgRegisterViewKeyResponse{},
		&privacypb.MsgRevokeViewKeyResponse{},
		&privacypb.MsgUpdateNetworkPrivacyResponse{},
		&privacypb.MsgUpdateParamsResponse{},
	)
}
