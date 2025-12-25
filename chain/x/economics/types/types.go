// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

// Economics module types - helper types not defined in proto
// All main types are re-exported from proto via aliases.go

// ============================
// Helper Types for Local Use
// ============================

// CategoryParams represents governance parameters for a specific proposal category
// (Used for local configuration, may be extended beyond proto)
type CategoryParams struct {
	Category             ProposalCategory `json:"category"`
	MinDeposit           string           `json:"min_deposit"`
	VotingPeriod         int64            `json:"voting_period"`
	QuorumThreshold      string           `json:"quorum_threshold"`
	PassThreshold        string           `json:"pass_threshold"`
	VetoThreshold        string           `json:"veto_threshold"`
	ExecutionDelay       int64            `json:"execution_delay"`
}

// WeightedVoteOption represents a weighted vote option for split voting
type WeightedVoteOption struct {
	Option VoteOption `json:"option"`
	Weight string     `json:"weight"`
}

// VetoRequest represents a veto request on a proposal (helper for complex veto logic)
type VetoRequest struct {
	ProposalId    uint64   `json:"proposal_id"`
	Initiator     string   `json:"initiator"`
	Cosigners     []string `json:"cosigners"`
	Reason        string   `json:"reason"`
	SubmittedAt   int64    `json:"submitted_at"`
}

// SnapshotVote represents a snapshot vote for off-chain voting
type SnapshotVote struct {
	ProposalId  uint64     `json:"proposal_id"`
	Voter       string     `json:"voter"`
	Option      VoteOption `json:"option"`
	VotingPower string     `json:"voting_power"`
	SnapshotAt  int64      `json:"snapshot_at"`
}

// VoteCommitment represents a commitment for secret ballot voting
type VoteCommitment struct {
	ProposalId  uint64 `json:"proposal_id"`
	Voter       string `json:"voter"`
	VoteHash    string `json:"vote_hash"`
	CommittedAt int64  `json:"committed_at"`
	Revealed    bool   `json:"revealed"`
}

// TokenLock represents locked tokens for governance participation
type TokenLock struct {
	Owner      string `json:"owner"`
	LockId     string `json:"lock_id"`
	Amount     string `json:"amount"`
	LockedAt   int64  `json:"locked_at"`
	UnlockTime int64  `json:"unlock_time"`
	Purpose    string `json:"purpose"`
}

// ============================
// Alert and Monitoring Types
// ============================

// AlertSeverity represents the severity of an alert
type AlertSeverity int32

const (
	AlertSeverityUnspecified AlertSeverity = 0
	AlertSeverityInfo        AlertSeverity = 1
	AlertSeverityWarning     AlertSeverity = 2
	AlertSeverityCritical    AlertSeverity = 3
	AlertSeverityEmergency   AlertSeverity = 4
)

// InflationAlertType represents different types of inflation alerts
type InflationAlertType int32

const (
	InflationAlertTypeUnspecified InflationAlertType = 0
	InflationAlertTypeAboveTarget InflationAlertType = 1
	InflationAlertTypeBelowTarget InflationAlertType = 2
	InflationAlertTypeAboveMax    InflationAlertType = 3
	InflationAlertTypeBelowMin    InflationAlertType = 4
	InflationAlertTypeRapidChange InflationAlertType = 5
)

// InflationAlert represents an inflation monitoring alert
type InflationAlert struct {
	AlertId       string             `json:"alert_id"`
	AlertType     InflationAlertType `json:"alert_type"`
	Severity      AlertSeverity      `json:"severity"`
	CurrentRate   uint64             `json:"current_rate"`
	TargetRate    uint64             `json:"target_rate"`
	Message       string             `json:"message"`
	DetectedAt    int64              `json:"detected_at"`
	Acknowledged  bool               `json:"acknowledged"`
}

// LargeTxRecord represents a record of a large transaction
type LargeTxRecord struct {
	TxHash        string `json:"tx_hash"`
	Sender        string `json:"sender"`
	Recipient     string `json:"recipient"`
	Amount        string `json:"amount"`
	Timestamp     int64  `json:"timestamp"`
	BlockHeight   uint64 `json:"block_height"`
	Flagged       bool   `json:"flagged"`
}

// TreasuryMultisig represents the multisig configuration for treasury
type TreasuryMultisig struct {
	TreasuryAddress string   `json:"treasury_address"`
	Signers         []string `json:"signers"`
	Threshold       uint32   `json:"threshold"`
	TimelockDelay   int64    `json:"timelock_delay"`
}
