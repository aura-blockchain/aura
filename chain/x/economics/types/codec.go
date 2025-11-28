package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"

	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// RegisterInterfaces registers the economics module interfaces
func RegisterInterfaces(registry types.InterfaceRegistry) {
	// Register proto messages
	registry.RegisterImplementations((*types.UnpackInterfacesMessage)(nil),
		&economicspb.Params{},
		&economicspb.GenesisState{},
		&economicspb.VestingSchedule{},
		&economicspb.VoteLock{},
		&economicspb.Proposal{},
		&economicspb.Vote{},
		&economicspb.Deposit{},
		&economicspb.VoteDelegation{},
		&economicspb.PendingTreasuryTx{},
	)
}

// RegisterLegacyAminoCodec registers the necessary types for Amino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register helper types
	cdc.RegisterConcrete(&StringList{}, "aura/economics/StringList", nil)
}
