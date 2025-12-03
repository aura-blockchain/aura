package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

// RegisterInterfaces registers the bridge module's interface types
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &bridgepb.Msg_ServiceDesc)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&bridgepb.MsgLockTokens{},
		&bridgepb.MsgMintTokens{},
		&bridgepb.MsgUnlockTokens{},
		&bridgepb.MsgBurnTokens{},
		&bridgepb.MsgLinkAddress{},
		&bridgepb.MsgCrossChainSwap{},
		&bridgepb.MsgRelayTransfer{},
		&bridgepb.MsgFinalizeTransfer{},
		&bridgepb.MsgSubmitFraudProof{},
	)
}
