// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import "fmt"

// Event types for the governance module
const (
	EventTypeProposalSubmitted = "proposal_submitted"
	EventTypeProposalDeposit   = "proposal_deposit"
	EventTypeVoteCast          = "vote_cast"
	EventTypeProposalPassed    = "proposal_passed"
	EventTypeProposalRejected  = "proposal_rejected"
	EventTypeProposalExpired   = "proposal_expired"
	EventTypeParamsUpdated     = "params_updated"
)

// Event attribute keys
const (
	AttributeKeyProposalID    = "proposal_id"
	AttributeKeyProposer      = "proposer"
	AttributeKeyTitle         = "title"
	AttributeKeyDepositor     = "depositor"
	AttributeKeyDepositAmount = "deposit_amount"
	AttributeKeyTotalDeposit  = "total_deposit"
	AttributeKeyVoter         = "voter"
	AttributeKeyOption        = "option"
	AttributeKeyVotingPower   = "voting_power"
	AttributeKeyYesVotes      = "yes_votes"
	AttributeKeyNoVotes       = "no_votes"
	AttributeKeyAbstainVotes  = "abstain_votes"
	AttributeKeyVetoVotes     = "veto_votes"
	AttributeKeyTotalVotes    = "total_votes"
	AttributeKeyQuorum        = "quorum"
	AttributeKeyThreshold     = "threshold"
	AttributeKeyBlockHeight   = "block_height"
	AttributeKeyBlockTime     = "block_time"
)

// NewProposalSubmittedEvent creates event attributes
func NewProposalSubmittedEvent(proposalID uint64, proposer, title string, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeyProposalID:  formatUint64(proposalID),
		AttributeKeyProposer:    proposer,
		AttributeKeyTitle:       title,
		AttributeKeyBlockHeight: formatInt64(blockHeight),
		AttributeKeyBlockTime:   blockTime,
	}
}

func formatInt64(i int64) string   { return fmt.Sprintf("%d", i) }
func formatUint64(u uint64) string { return fmt.Sprintf("%d", u) }
