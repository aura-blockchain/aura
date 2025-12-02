package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

// RegisterLegacyAminoCodec registers DEX module messages on the legacy Amino codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	_ = cdc
}

// RegisterInterfaces registers the DEX module proto interfaces.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &dexpb.Msg_ServiceDesc)

	// Register all DEX message types as sdk.Msg implementations
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&dexpb.MsgCreatePool{},
		&dexpb.MsgAddLiquidity{},
		&dexpb.MsgRemoveLiquidity{},
		&dexpb.MsgSwapExactIn{},
		&dexpb.MsgCreateOrder{},
		&dexpb.MsgCancelOrder{},
		&dexpb.MsgExecuteSwap{},
		&dexpb.MsgCreateHTLC{},
		&dexpb.MsgClaimHTLC{},
		&dexpb.MsgRefundHTLC{},
	)

	// Register response types
	registry.RegisterImplementations((*sdktx.MsgResponse)(nil),
		&dexpb.MsgCreatePoolResponse{},
		&dexpb.MsgAddLiquidityResponse{},
		&dexpb.MsgRemoveLiquidityResponse{},
		&dexpb.MsgSwapExactInResponse{},
		&dexpb.MsgCreateOrderResponse{},
		&dexpb.MsgCancelOrderResponse{},
		&dexpb.MsgExecuteSwapResponse{},
		&dexpb.MsgCreateHTLCResponse{},
		&dexpb.MsgClaimHTLCResponse{},
		&dexpb.MsgRefundHTLCResponse{},
	)
}

// Note: GetSigners implementations are added directly to the proto-generated
// tx.pb.go file after generation. See proto/aura/dex/v1beta1/tx.pb.go
