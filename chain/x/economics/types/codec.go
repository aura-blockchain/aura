package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// RegisterInterfaces registers the economics module interfaces
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &economicspb.Msg_ServiceDesc)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&economicspb.MsgCreateVestingSchedule{},
		&economicspb.MsgReleaseVestedTokens{},
		&economicspb.MsgRevokeVestingSchedule{},
		&economicspb.MsgSubmitProposal{},
		&economicspb.MsgDeposit{},
		&economicspb.MsgVote{},
		&economicspb.MsgVoteWeighted{},
		&economicspb.MsgDelegateVote{},
		&economicspb.MsgUndelegateVote{},
		&economicspb.MsgExecuteProposal{},
		&economicspb.MsgRevealSecretVote{},
		&economicspb.MsgLockVotingTokens{},
		&economicspb.MsgUnlockVotingTokens{},
		&economicspb.MsgProposeTreasurySpend{},
		&economicspb.MsgSignTreasurySpend{},
		&economicspb.MsgExecuteTreasurySpend{},
		&economicspb.MsgUpdateParams{},
		&economicspb.MsgAdjustInflationRate{},
	)

	registry.RegisterImplementations((*sdktx.MsgResponse)(nil),
		&economicspb.MsgCreateVestingScheduleResponse{},
		&economicspb.MsgReleaseVestedTokensResponse{},
		&economicspb.MsgRevokeVestingScheduleResponse{},
		&economicspb.MsgSubmitProposalResponse{},
		&economicspb.MsgDepositResponse{},
		&economicspb.MsgVoteResponse{},
		&economicspb.MsgVoteWeightedResponse{},
		&economicspb.MsgDelegateVoteResponse{},
		&economicspb.MsgUndelegateVoteResponse{},
		&economicspb.MsgExecuteProposalResponse{},
		&economicspb.MsgRevealSecretVoteResponse{},
		&economicspb.MsgLockVotingTokensResponse{},
		&economicspb.MsgUnlockVotingTokensResponse{},
		&economicspb.MsgProposeTreasurySpendResponse{},
		&economicspb.MsgSignTreasurySpendResponse{},
		&economicspb.MsgExecuteTreasurySpendResponse{},
		&economicspb.MsgUpdateParamsResponse{},
		&economicspb.MsgAdjustInflationRateResponse{},
	)

	// Register proto messages for unpacking
	registry.RegisterImplementations((*codectypes.UnpackInterfacesMessage)(nil),
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
