package v1beta1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (m *MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	proposer, err := sdk.AccAddressFromBech32(m.Proposer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{proposer}
}

func (m *MsgDeposit) GetSigners() []sdk.AccAddress {
	depositor, err := sdk.AccAddressFromBech32(m.Depositor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{depositor}
}

func (m *MsgVote) GetSigners() []sdk.AccAddress {
	voter, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voter}
}

func (m *MsgVoteWeighted) GetSigners() []sdk.AccAddress {
	voter, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voter}
}

func (m *MsgDelegateVote) GetSigners() []sdk.AccAddress {
	delegator, err := sdk.AccAddressFromBech32(m.Delegator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delegator}
}

func (m *MsgUndelegateVote) GetSigners() []sdk.AccAddress {
	delegator, err := sdk.AccAddressFromBech32(m.Delegator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delegator}
}

func (m *MsgSubmitVeto) GetSigners() []sdk.AccAddress {
	vetoer, err := sdk.AccAddressFromBech32(m.Vetoer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{vetoer}
}

func (m *MsgCosignVeto) GetSigners() []sdk.AccAddress {
	cosigner, err := sdk.AccAddressFromBech32(m.Cosigner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{cosigner}
}

func (m *MsgExecuteProposal) GetSigners() []sdk.AccAddress {
	executor, err := sdk.AccAddressFromBech32(m.Executor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{executor}
}

func (m *MsgSubmitSnapshotVote) GetSigners() []sdk.AccAddress {
	voter, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voter}
}

func (m *MsgRevealSecretVote) GetSigners() []sdk.AccAddress {
	voter, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voter}
}
