package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	espb "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
)

// RegisterLegacyAminoCodec registers the necessary x/economicsecurity interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register message types here if economicsecurity module has messages
}

// RegisterInterfaces registers the x/economicsecurity interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &espb.Msg_serviceDesc)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&espb.MsgCreateVestingSchedule{},
		&espb.MsgReleaseVestedTokens{},
		&espb.MsgRevokeVestingSchedule{},
		&espb.MsgLockVotingTokens{},
		&espb.MsgUnlockVotingTokens{},
		&espb.MsgProposeTreasurySpend{},
		&espb.MsgSignTreasurySpend{},
		&espb.MsgExecuteTreasurySpend{},
		&espb.MsgUpdateParams{},
		&espb.MsgAdjustInflationRate{},
	)
}
