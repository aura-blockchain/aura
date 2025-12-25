// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

const (
	// ModuleName defines the governance module name used across the app shell.
	ModuleName = "governance"

	// StoreKey provides the primary KV store key for governance state.
	StoreKey = ModuleName

	// RouterKey is reserved for message routing once gRPC/CLI wiring lands.
	RouterKey = ModuleName

	// QuerierRoute defines the query service route.
	QuerierRoute = ModuleName

	// MemStoreKey exposes an in-memory store key placeholder for future usage.
	MemStoreKey = "mem_" + ModuleName
)

// Event types
const (
	EventTypeSubmitProposal  = "submit_proposal"
	EventTypeDeposit         = "proposal_deposit"
	EventTypeVote            = "proposal_vote"
	EventTypeDelegateVote    = "delegate_vote"
	EventTypeUndelegateVote  = "undelegate_vote"
	EventTypeVeto            = "proposal_veto"
	EventTypeExecuteProposal = "execute_proposal"
	EventTypeSnapshotVote    = "snapshot_vote"
	EventTypeRevealVote      = "reveal_vote"
)

// Attribute keys
const (
	AttributeKeyDelegator = "delegator"
	AttributeKeyDelegate  = "delegate"
	AttributeKeyVetoer    = "vetoer"
	AttributeKeyExecutor  = "executor"
)
