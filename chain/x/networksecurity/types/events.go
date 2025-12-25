// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

// Event types for the networksecurity module
const (
	EventTypePeerBanned        = "peer_banned"
	EventTypePeerUnbanned      = "peer_unbanned"
	EventTypeSybilDetected     = "sybil_detected"
	EventTypeRateLimitExceeded = "rate_limit_exceeded"
	EventTypeDDoSDetected      = "ddos_detected"
	EventTypeParamsUpdated     = "params_updated"
)

// Event attribute keys
const (
	AttributeKeyPeerID        = "peer_id"
	AttributeKeyReason        = "reason"
	AttributeKeyReputationScore = "reputation_score"
	AttributeKeyConfidence    = "confidence_score"
	AttributeKeyResourceType  = "resource_type"
	AttributeKeyBlockHeight   = "block_height"
	AttributeKeyBlockTime     = "block_time"
)
