// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	gogotypes "github.com/cosmos/gogoproto/types"
)

// ProposalTemplate represents a template for creating proposals
type ProposalTemplate struct {
	Category            ProposalCategory
	Name                string
	Description         string
	TitleTemplate       string
	DescriptionTemplate string
	ContentTemplate     string
	RequiredFields      []string
	OptionalFields      []string
	Examples            []string
}

// ExecutionResult represents the result of proposal execution
type ExecutionResult struct {
	Success bool
	Message string
	Data    []byte
}

// ProposalExecution represents a proposal execution record
type ProposalExecution struct {
	ProposalId    uint64
	ExecutedAt    *gogotypes.Timestamp
	Success       bool
	ErrorMessage  string
	GasUsed       uint64
	EventsEmitted uint64
	ResultData    string
}

// ExecutionStatistics represents execution statistics
type ExecutionStatistics struct {
	TotalExecuted        uint64
	SuccessfulExecutions uint64
	FailedExecutions     uint64
	PendingExecutions    uint64
	AverageGasUsed       uint64
	TotalGasUsed         uint64
}

// GovernanceAnalytics represents governance analytics data
type GovernanceAnalytics struct {
	TotalProposals          uint64
	ProposalsByStatus       map[string]uint64
	ProposalsByType         map[string]uint64
	AverageVotingPower      string
	ParticipationRate       float64
	AverageProposalDuration uint64
	PassRate                float64
	VetoRate                float64
}

// ProposerAnalytics represents analytics for a specific proposer
type ProposerAnalytics struct {
	Proposer          string
	TotalProposals    uint64
	PassedProposals   uint64
	FailedProposals   uint64
	RejectedProposals uint64
	ExecutedProposals uint64
	SuccessRate       float64
}

// ParameterValidationRules represents parameter validation rules
type ParameterValidationRules struct {
	MinVotingPeriod        uint64
	MaxVotingPeriod        uint64
	MinDepositPeriod       uint64
	MaxDepositPeriod       uint64
	MinQuorum              uint64
	MaxQuorum              uint64
	MinThreshold           uint64
	MaxThreshold           uint64
	MinVetoThreshold       uint64
	MaxVetoThreshold       uint64
	MaxExecutionDelay      uint64
	MaxDelegationsPerUser  uint64
	MinVoteCreditsPerToken uint64
	MaxProposalTitleLength uint64
	MaxProposalDescLength  uint64
}

// Additional helper aliases for status constants
var (
	ProposalStatusDepositPeriod = StatusDepositPeriod
	ProposalStatusVotingPeriod  = StatusVotingPeriod
	ProposalStatusPassed        = StatusPassed
	ProposalStatusRejected      = StatusRejected
	ProposalStatusFailed        = StatusFailed
	ProposalStatusVetoed        = StatusVetoed
	ProposalStatusExecuted      = StatusExecuted
)

// Additional helper aliases for vote options
var (
	VoteOptionYes         = OptionYes
	VoteOptionNo          = OptionNo
	VoteOptionAbstain     = OptionAbstain
	VoteOptionNoWithVeto  = OptionNoWithVeto
)

// Additional error for invalid proposal type
var (
	ErrInvalidProposalType     = ErrInvalidProposal
	ErrAutoExecutionDisabled   = ErrInvalidProposal
	ErrNoScheduledExecution    = ErrInvalidProposal
	ErrNoVotingPower           = ErrInvalidVote
	ErrInvalidVoteOption       = ErrInvalidVote
	ErrUnauthorized            = ErrUnauthorizedVeto
)
