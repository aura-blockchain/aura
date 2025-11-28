package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/wasm interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgStoreCode{}, "wasm/MsgStoreCode", nil)
	cdc.RegisterConcrete(&MsgInstantiateContract{}, "wasm/MsgInstantiateContract", nil)
	cdc.RegisterConcrete(&MsgExecuteContract{}, "wasm/MsgExecuteContract", nil)
	cdc.RegisterConcrete(&MsgMigrateContract{}, "wasm/MsgMigrateContract", nil)
	cdc.RegisterConcrete(&MsgUpdateAdmin{}, "wasm/MsgUpdateAdmin", nil)
	cdc.RegisterConcrete(&MsgClearAdmin{}, "wasm/MsgClearAdmin", nil)
	cdc.RegisterConcrete(&MsgAuthorizeUploader{}, "wasm/MsgAuthorizeUploader", nil)
	cdc.RegisterConcrete(&MsgRevokeUploader{}, "wasm/MsgRevokeUploader", nil)
	cdc.RegisterConcrete(&MsgPauseContract{}, "wasm/MsgPauseContract", nil)
	cdc.RegisterConcrete(&MsgUnpauseContract{}, "wasm/MsgUnpauseContract", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "wasm/MsgUpdateParams", nil)
}

// RegisterInterfaces registers the x/wasm interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgStoreCode{},
		&MsgInstantiateContract{},
		&MsgExecuteContract{},
		&MsgMigrateContract{},
		&MsgUpdateAdmin{},
		&MsgClearAdmin{},
		&MsgAuthorizeUploader{},
		&MsgRevokeUploader{},
		&MsgPauseContract{},
		&MsgUnpauseContract{},
		&MsgUpdateParams{},
	)

	// msgservice registration is skipped for now since we're using stubs
	// In production this would be: msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	_ = msgservice.RegisterMsgServiceDesc
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

// RegisterMsgServer registers the msg server
func RegisterMsgServer(server interface{}, impl MsgServer) {
	// Stub - in production this would use gRPC registration
}

// RegisterQueryServer registers the query server
func RegisterQueryServer(server interface{}, impl QueryServer) {
	// Stub - in production this would use gRPC registration
}
