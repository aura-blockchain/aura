// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetSigners returns the signer addresses for MsgCreateVestingSchedule
func (m *MsgCreateVestingSchedule) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSigners returns the signer addresses for MsgReleaseVestedTokens
func (m *MsgReleaseVestedTokens) GetSigners() []sdk.AccAddress {
	beneficiary, err := sdk.AccAddressFromBech32(m.Beneficiary)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{beneficiary}
}

// GetSigners returns the signer addresses for MsgRevokeVestingSchedule
func (m *MsgRevokeVestingSchedule) GetSigners() []sdk.AccAddress {
	revoker, err := sdk.AccAddressFromBech32(m.Revoker)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{revoker}
}

// GetSigners returns the signer addresses for MsgSubmitProposal
func (m *MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	proposer, err := sdk.AccAddressFromBech32(m.Proposer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{proposer}
}

// GetSigners returns the signer addresses for MsgDeposit
func (m *MsgDeposit) GetSigners() []sdk.AccAddress {
	depositor, err := sdk.AccAddressFromBech32(m.Depositor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{depositor}
}

// GetSigners returns the signer addresses for MsgVote
func (m *MsgVote) GetSigners() []sdk.AccAddress {
	voter, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voter}
}

// GetSigners returns the signer addresses for MsgVoteWeighted
func (m *MsgVoteWeighted) GetSigners() []sdk.AccAddress {
	voter, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voter}
}

// GetSigners returns the signer addresses for MsgDelegateVote
func (m *MsgDelegateVote) GetSigners() []sdk.AccAddress {
	delegator, err := sdk.AccAddressFromBech32(m.Delegator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delegator}
}

// GetSigners returns the signer addresses for MsgUndelegateVote
func (m *MsgUndelegateVote) GetSigners() []sdk.AccAddress {
	delegator, err := sdk.AccAddressFromBech32(m.Delegator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delegator}
}

// GetSigners returns the signer addresses for MsgExecuteProposal
func (m *MsgExecuteProposal) GetSigners() []sdk.AccAddress {
	executor, err := sdk.AccAddressFromBech32(m.Executor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{executor}
}

// GetSigners returns the signer addresses for MsgRevealSecretVote
func (m *MsgRevealSecretVote) GetSigners() []sdk.AccAddress {
	voter, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voter}
}

// GetSigners returns the signer addresses for MsgLockVotingTokens
func (m *MsgLockVotingTokens) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

// GetSigners returns the signer addresses for MsgUnlockVotingTokens
func (m *MsgUnlockVotingTokens) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

// GetSigners returns the signer addresses for MsgProposeTreasurySpend
func (m *MsgProposeTreasurySpend) GetSigners() []sdk.AccAddress {
	proposer, err := sdk.AccAddressFromBech32(m.Proposer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{proposer}
}

// GetSigners returns the signer addresses for MsgSignTreasurySpend
func (m *MsgSignTreasurySpend) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// GetSigners returns the signer addresses for MsgExecuteTreasurySpend
func (m *MsgExecuteTreasurySpend) GetSigners() []sdk.AccAddress {
	executor, err := sdk.AccAddressFromBech32(m.Executor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{executor}
}

// GetSigners returns the signer addresses for MsgUpdateParams
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// GetSigners returns the signer addresses for MsgAdjustInflationRate
func (m *MsgAdjustInflationRate) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}
