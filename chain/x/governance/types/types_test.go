// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
)

func TestTypeExports(t *testing.T) {
	var _ ProposalStatus
	var _ VoteOption
	var _ ProposalCategory
	var _ Proposal
	var _ TallyResult
	var _ Vote
	var _ WeightedVoteOption
	var _ Deposit
	var _ SnapshotVote
	var _ VoteDelegation
	var _ VetoRequest
	var _ TokenLock
	var _ CategoryParams
	var _ GovernanceParams
}

func TestMessageTypeExports(t *testing.T) {
	var _ MsgSubmitProposal
	var _ MsgVote
	var _ MsgVoteWeighted
	var _ MsgDeposit
	var _ MsgSubmitSnapshotVote
	var _ MsgRevealSecretVote
	var _ MsgDelegateVote
	var _ MsgUndelegateVote
	var _ MsgSubmitVeto
	var _ MsgCosignVeto
	var _ MsgExecuteProposal
}

func TestQueryTypeExports(t *testing.T) {
	var _ QueryProposalRequest
	var _ QueryProposalResponse
	var _ QueryProposalsRequest
	var _ QueryVoteRequest
	var _ QueryVotesRequest
	var _ QueryDepositRequest
	var _ QueryTallyResultRequest
	var _ QueryParamsRequest
}

func TestProposalStatusEnums(t *testing.T) {
	statuses := []ProposalStatus{
		ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
		ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD,
		ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		ProposalStatus_PROPOSAL_STATUS_PASSED,
		ProposalStatus_PROPOSAL_STATUS_REJECTED,
		ProposalStatus_PROPOSAL_STATUS_FAILED,
		ProposalStatus_PROPOSAL_STATUS_VETOED,
		ProposalStatus_PROPOSAL_STATUS_EXECUTION_DELAY,
		ProposalStatus_PROPOSAL_STATUS_READY_FOR_EXECUTION,
		ProposalStatus_PROPOSAL_STATUS_EXECUTED,
	}

	seen := make(map[ProposalStatus]bool)
	for _, status := range statuses {
		if seen[status] {
			t.Errorf("duplicate ProposalStatus value: %v", status)
		}
		seen[status] = true
	}
}

func TestVoteOptionEnums(t *testing.T) {
	options := []VoteOption{
		VoteOption_VOTE_OPTION_UNSPECIFIED,
		VoteOption_VOTE_OPTION_YES,
		VoteOption_VOTE_OPTION_ABSTAIN,
		VoteOption_VOTE_OPTION_NO,
		VoteOption_VOTE_OPTION_NO_WITH_VETO,
	}

	seen := make(map[VoteOption]bool)
	for _, option := range options {
		if seen[option] {
			t.Errorf("duplicate VoteOption value: %v", option)
		}
		seen[option] = true
	}
}

func TestProposalCategoryEnums(t *testing.T) {
	categories := []ProposalCategory{
		ProposalCategory_PROPOSAL_CATEGORY_UNSPECIFIED,
		ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE,
		ProposalCategory_PROPOSAL_CATEGORY_SOFTWARE_UPGRADE,
		ProposalCategory_PROPOSAL_CATEGORY_SPENDING,
		ProposalCategory_PROPOSAL_CATEGORY_EMERGENCY,
		ProposalCategory_PROPOSAL_CATEGORY_CONSTITUTION,
	}

	seen := make(map[ProposalCategory]bool)
	for _, category := range categories {
		if seen[category] {
			t.Errorf("duplicate ProposalCategory value: %v", category)
		}
		seen[category] = true
	}
}
