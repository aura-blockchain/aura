package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	govpb "github.com/aequitas/aura/proto/aura/governance/v1beta1"
)

var (
	_ sdk.Msg = (*govpb.MsgSubmitProposal)(nil)
	_ sdk.Msg = (*govpb.MsgDeposit)(nil)
	_ sdk.Msg = (*govpb.MsgVote)(nil)
	_ sdk.Msg = (*govpb.MsgVoteWeighted)(nil)
	_ sdk.Msg = (*govpb.MsgDelegateVote)(nil)
	_ sdk.Msg = (*govpb.MsgUndelegateVote)(nil)
	_ sdk.Msg = (*govpb.MsgSubmitVeto)(nil)
	_ sdk.Msg = (*govpb.MsgCosignVeto)(nil)
	_ sdk.Msg = (*govpb.MsgExecuteProposal)(nil)
	_ sdk.Msg = (*govpb.MsgSubmitSnapshotVote)(nil)
	_ sdk.Msg = (*govpb.MsgRevealSecretVote)(nil)
)

func (m *govpb.MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	proposer, err := sdk.AccAddressFromBech32(m.Proposer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{proposer}
}

func (m *govpb.MsgDeposit) GetSigners() []sdk.AccAddress {
	depositor, err := sdk.AccAddressFromBech32(m.Depositor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{depositor}
}

func (m *govpb.MsgVote) GetSigners() []sdk.AccAddress {
	voter, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voter}
}

func (m *govpb.MsgVoteWeighted) GetSigners() []sdk.AccAddress {
	voter, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voter}
}

func (m *govpb.MsgDelegateVote) GetSigners() []sdk.AccAddress {
	delegator, err := sdk.AccAddressFromBech32(m.Delegator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delegator}
}

func (m *govpb.MsgUndelegateVote) GetSigners() []sdk.AccAddress {
	delegator, err := sdk.AccAddressFromBech32(m.Delegator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delegator}
}

func (m *govpb.MsgSubmitVeto) GetSigners() []sdk.AccAddress {
	vetoer, err := sdk.AccAddressFromBech32(m.Vetoer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{vetoer}
}

func (m *govpb.MsgCosignVeto) GetSigners() []sdk.AccAddress {
	cosigner, err := sdk.AccAddressFromBech32(m.Cosigner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{cosigner}
}

func (m *govpb.MsgExecuteProposal) GetSigners() []sdk.AccAddress {
	executor, err := sdk.AccAddressFromBech32(m.Executor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{executor}
}

func (m *govpb.MsgSubmitSnapshotVote) GetSigners() []sdk.AccAddress {
	voter, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voter}
}

func (m *govpb.MsgRevealSecretVote) GetSigners() []sdk.AccAddress {
	voter, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voter}
}
