package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	wasmpb "github.com/aequitas/aura/proto/aura/wasm/v1beta1"
)

// RegisterLegacyAminoCodec registers the necessary x/wasm interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&wasmpb.MsgStoreCode{}, "wasm/MsgStoreCode", nil)
	cdc.RegisterConcrete(&wasmpb.MsgInstantiateContract{}, "wasm/MsgInstantiateContract", nil)
	cdc.RegisterConcrete(&wasmpb.MsgExecuteContract{}, "wasm/MsgExecuteContract", nil)
	cdc.RegisterConcrete(&wasmpb.MsgMigrateContract{}, "wasm/MsgMigrateContract", nil)
	cdc.RegisterConcrete(&wasmpb.MsgUpdateAdmin{}, "wasm/MsgUpdateAdmin", nil)
	cdc.RegisterConcrete(&wasmpb.MsgClearAdmin{}, "wasm/MsgClearAdmin", nil)
	cdc.RegisterConcrete(&wasmpb.MsgAuthorizeUploader{}, "wasm/MsgAuthorizeUploader", nil)
	cdc.RegisterConcrete(&wasmpb.MsgRevokeUploader{}, "wasm/MsgRevokeUploader", nil)
	cdc.RegisterConcrete(&wasmpb.MsgPauseContract{}, "wasm/MsgPauseContract", nil)
	cdc.RegisterConcrete(&wasmpb.MsgUnpauseContract{}, "wasm/MsgUnpauseContract", nil)
	cdc.RegisterConcrete(&wasmpb.MsgUpdateParams{}, "wasm/MsgUpdateParams", nil)
}

// RegisterInterfaces registers the x/wasm interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&wasmpb.MsgStoreCode{},
		&wasmpb.MsgInstantiateContract{},
		&wasmpb.MsgExecuteContract{},
		&wasmpb.MsgMigrateContract{},
		&wasmpb.MsgUpdateAdmin{},
		&wasmpb.MsgClearAdmin{},
		&wasmpb.MsgAuthorizeUploader{},
		&wasmpb.MsgRevokeUploader{},
		&wasmpb.MsgPauseContract{},
		&wasmpb.MsgUnpauseContract{},
		&wasmpb.MsgUpdateParams{},
	)

	// Register the Msg service descriptor
	msgservice.RegisterMsgServiceDesc(registry, &wasmpb.Msg_ServiceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}
