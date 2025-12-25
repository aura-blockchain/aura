// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import pb "github.com/aequitas/aura/proto/aura/governance/v1beta1"

// Re-export all proto types
type (
	// Enums
	ProposalStatus   = pb.ProposalStatus
	VoteOption       = pb.VoteOption
	ProposalCategory = pb.ProposalCategory

	// Core types
	Proposal            = pb.Proposal
	TallyResult         = pb.TallyResult
	Vote                = pb.Vote
	WeightedVoteOption  = pb.WeightedVoteOption
	Deposit             = pb.Deposit
	SnapshotVote        = pb.SnapshotVote
	VoteDelegation      = pb.VoteDelegation
	VetoRequest         = pb.VetoRequest
	TokenLock           = pb.TokenLock
	CategoryParams      = pb.CategoryParams
	GovernanceParams    = pb.GovernanceParams

	// Message types
	MsgSubmitProposal            = pb.MsgSubmitProposal
	MsgSubmitProposalResponse    = pb.MsgSubmitProposalResponse
	MsgVote                      = pb.MsgVote
	MsgVoteResponse              = pb.MsgVoteResponse
	MsgVoteWeighted              = pb.MsgVoteWeighted
	MsgVoteWeightedResponse      = pb.MsgVoteWeightedResponse
	MsgDeposit                   = pb.MsgDeposit
	MsgDepositResponse           = pb.MsgDepositResponse
	MsgSubmitSnapshotVote        = pb.MsgSubmitSnapshotVote
	MsgSubmitSnapshotVoteResponse = pb.MsgSubmitSnapshotVoteResponse
	MsgRevealSecretVote          = pb.MsgRevealSecretVote
	MsgRevealSecretVoteResponse  = pb.MsgRevealSecretVoteResponse
	MsgDelegateVote              = pb.MsgDelegateVote
	MsgDelegateVoteResponse      = pb.MsgDelegateVoteResponse
	MsgUndelegateVote            = pb.MsgUndelegateVote
	MsgUndelegateVoteResponse    = pb.MsgUndelegateVoteResponse
	MsgSubmitVeto                = pb.MsgSubmitVeto
	MsgSubmitVetoResponse        = pb.MsgSubmitVetoResponse
	MsgCosignVeto                = pb.MsgCosignVeto
	MsgCosignVetoResponse        = pb.MsgCosignVetoResponse
	MsgExecuteProposal           = pb.MsgExecuteProposal
	MsgExecuteProposalResponse   = pb.MsgExecuteProposalResponse

	// Query types
	QueryProposalRequest            = pb.QueryProposalRequest
	QueryProposalResponse           = pb.QueryProposalResponse
	QueryProposalsRequest           = pb.QueryProposalsRequest
	QueryProposalsResponse          = pb.QueryProposalsResponse
	QueryVoteRequest                = pb.QueryVoteRequest
	QueryVoteResponse               = pb.QueryVoteResponse
	QueryVotesRequest               = pb.QueryVotesRequest
	QueryVotesResponse              = pb.QueryVotesResponse
	QueryDepositRequest             = pb.QueryDepositRequest
	QueryDepositResponse            = pb.QueryDepositResponse
	QueryDepositsRequest            = pb.QueryDepositsRequest
	QueryDepositsResponse           = pb.QueryDepositsResponse
	QueryTallyResultRequest         = pb.QueryTallyResultRequest
	QueryTallyResultResponse        = pb.QueryTallyResultResponse
	QuerySnapshotVotesRequest       = pb.QuerySnapshotVotesRequest
	QuerySnapshotVotesResponse      = pb.QuerySnapshotVotesResponse
	QueryVoteDelegationsRequest     = pb.QueryVoteDelegationsRequest
	QueryVoteDelegationsResponse    = pb.QueryVoteDelegationsResponse
	QueryVotingPowerRequest         = pb.QueryVotingPowerRequest
	QueryVotingPowerResponse        = pb.QueryVotingPowerResponse
	QueryVetoRequestsRequest        = pb.QueryVetoRequestsRequest
	QueryVetoRequestsResponse       = pb.QueryVetoRequestsResponse
	QueryTokenLocksRequest          = pb.QueryTokenLocksRequest
	QueryTokenLocksResponse         = pb.QueryTokenLocksResponse
	QueryParamsRequest              = pb.QueryParamsRequest
	QueryParamsResponse             = pb.QueryParamsResponse
)

// Re-export enum values for ProposalStatus
const (
	ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED         = pb.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED
	ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD      = pb.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD
	ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD       = pb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	ProposalStatus_PROPOSAL_STATUS_PASSED              = pb.ProposalStatus_PROPOSAL_STATUS_PASSED
	ProposalStatus_PROPOSAL_STATUS_REJECTED            = pb.ProposalStatus_PROPOSAL_STATUS_REJECTED
	ProposalStatus_PROPOSAL_STATUS_FAILED              = pb.ProposalStatus_PROPOSAL_STATUS_FAILED
	ProposalStatus_PROPOSAL_STATUS_VETOED              = pb.ProposalStatus_PROPOSAL_STATUS_VETOED
	ProposalStatus_PROPOSAL_STATUS_EXECUTION_DELAY     = pb.ProposalStatus_PROPOSAL_STATUS_EXECUTION_DELAY
	ProposalStatus_PROPOSAL_STATUS_READY_FOR_EXECUTION = pb.ProposalStatus_PROPOSAL_STATUS_READY_FOR_EXECUTION
	ProposalStatus_PROPOSAL_STATUS_EXECUTED            = pb.ProposalStatus_PROPOSAL_STATUS_EXECUTED
)

// Re-export enum values for VoteOption
const (
	VoteOption_VOTE_OPTION_UNSPECIFIED  = pb.VoteOption_VOTE_OPTION_UNSPECIFIED
	VoteOption_VOTE_OPTION_YES          = pb.VoteOption_VOTE_OPTION_YES
	VoteOption_VOTE_OPTION_ABSTAIN      = pb.VoteOption_VOTE_OPTION_ABSTAIN
	VoteOption_VOTE_OPTION_NO           = pb.VoteOption_VOTE_OPTION_NO
	VoteOption_VOTE_OPTION_NO_WITH_VETO = pb.VoteOption_VOTE_OPTION_NO_WITH_VETO
)

// Re-export enum values for ProposalCategory
const (
	ProposalCategory_PROPOSAL_CATEGORY_UNSPECIFIED      = pb.ProposalCategory_PROPOSAL_CATEGORY_UNSPECIFIED
	ProposalCategory_PROPOSAL_CATEGORY_TEXT             = pb.ProposalCategory_PROPOSAL_CATEGORY_TEXT
	ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE = pb.ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE
	ProposalCategory_PROPOSAL_CATEGORY_SOFTWARE_UPGRADE = pb.ProposalCategory_PROPOSAL_CATEGORY_SOFTWARE_UPGRADE
	ProposalCategory_PROPOSAL_CATEGORY_SPENDING         = pb.ProposalCategory_PROPOSAL_CATEGORY_SPENDING
	ProposalCategory_PROPOSAL_CATEGORY_EMERGENCY        = pb.ProposalCategory_PROPOSAL_CATEGORY_EMERGENCY
	ProposalCategory_PROPOSAL_CATEGORY_CONSTITUTION     = pb.ProposalCategory_PROPOSAL_CATEGORY_CONSTITUTION
)

// VoteCommitment represents a commitment to a vote for secret ballot voting
type VoteCommitment struct {
	ProposalId  uint64
	Voter       string
	VoteHash    string
	CommittedAt interface{} // timestamppb.Timestamp
	Revealed    bool
}

// QuadraticTallyResult represents tally results for quadratic voting
type QuadraticTallyResult struct {
	YesPower          uint64
	NoPower           uint64
	AbstainPower      uint64
	VetoPower         uint64
	TotalCreditsSpent uint64
	UniqueVoters      uint64
}

// QuadraticVotingStats represents statistics for quadratic voting
type QuadraticVotingStats struct {
	ProposalId             uint64
	TotalVotingPower       uint64
	TotalCreditsSpent      uint64
	UniqueVoters           uint64
	AverageCreditsPerVoter uint64
	CostEfficiency         float64
	QuadraticAdvantage     float64
}
