package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	govpb "github.com/aequitas/aura/proto/aura/governance/v1beta1"
)

// RegisterLegacyAminoCodec registers governance module messages on the legacy Amino codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	_ = cdc
}

// RegisterInterfaces registers the governance module proto interfaces.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &govpb.Msg_ServiceDesc)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&govpb.MsgSubmitProposal{},
		&govpb.MsgDeposit{},
		&govpb.MsgVote{},
		&govpb.MsgVoteWeighted{},
		&govpb.MsgDelegateVote{},
		&govpb.MsgUndelegateVote{},
		&govpb.MsgSubmitVeto{},
		&govpb.MsgCosignVeto{},
		&govpb.MsgExecuteProposal{},
		&govpb.MsgSubmitSnapshotVote{},
		&govpb.MsgRevealSecretVote{},
	)

	registry.RegisterImplementations((*sdktx.MsgResponse)(nil),
		&govpb.MsgSubmitProposalResponse{},
		&govpb.MsgDepositResponse{},
		&govpb.MsgVoteResponse{},
		&govpb.MsgVoteWeightedResponse{},
		&govpb.MsgDelegateVoteResponse{},
		&govpb.MsgUndelegateVoteResponse{},
		&govpb.MsgSubmitVetoResponse{},
		&govpb.MsgCosignVetoResponse{},
		&govpb.MsgExecuteProposalResponse{},
		&govpb.MsgSubmitSnapshotVoteResponse{},
		&govpb.MsgRevealSecretVoteResponse{},
	)
}
