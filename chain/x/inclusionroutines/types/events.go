// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

// Event types for the inclusionroutines module
const (
	EventTypeIRCreated              = "ir_created"
	EventTypeIRUpdated              = "ir_updated"
	EventTypeIRDeleted              = "ir_deleted"
	EventTypeIRActivated            = "ir_activated"
	EventTypeIRDeactivated          = "ir_deactivated"
	EventTypeIRCompleted            = "ir_completed"
	EventTypeArenaAssigned          = "arena_assigned"
	EventTypeParamsUpdated          = "params_updated"
	EventTypeIRRateLimitUpdated     = "ir_rate_limit_updated"
	EventTypeIRPrerequisitesUpdated = "ir_prerequisites_updated"
)

// Event attribute keys
const (
	AttributeKeyIRID        = "ir_id"
	AttributeKeyName        = "name"
	AttributeKeyCreator     = "creator"
	AttributeKeyScorePoints = "score_points"
	AttributeKeyArenaID     = "arena_id"
	AttributeKeyWallet      = "wallet_address"
	AttributeKeyBlockHeight = "block_height"
	AttributeKeyBlockTime   = "block_time"
)
