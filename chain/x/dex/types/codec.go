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

// GetSigners implementations for each message type
// These implement the sdk.Msg interface requirement

func (msg *dexpb.MsgCreatePool) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *dexpb.MsgAddLiquidity) GetSigners() []sdk.AccAddress {
	provider, err := sdk.AccAddressFromBech32(msg.Provider)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{provider}
}

func (msg *dexpb.MsgRemoveLiquidity) GetSigners() []sdk.AccAddress {
	provider, err := sdk.AccAddressFromBech32(msg.Provider)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{provider}
}

func (msg *dexpb.MsgSwapExactIn) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

func (msg *dexpb.MsgCreateOrder) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *dexpb.MsgCancelOrder) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *dexpb.MsgExecuteSwap) GetSigners() []sdk.AccAddress {
	initiator, err := sdk.AccAddressFromBech32(msg.Initiator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{initiator}
}

func (msg *dexpb.MsgCreateHTLC) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

func (msg *dexpb.MsgClaimHTLC) GetSigners() []sdk.AccAddress {
	recipient, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{recipient}
}

func (msg *dexpb.MsgRefundHTLC) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
