// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

const (
	// EventTypeDataItemStored is emitted when a data item is stored
	EventTypeDataItemStored = "data_item_stored"

	// EventTypeDataItemUpdated is emitted when a data item is updated
	EventTypeDataItemUpdated = "data_item_updated"

	// EventTypeDataItemDeleted is emitted when a data item is deleted
	EventTypeDataItemDeleted = "data_item_deleted"

	// EventTypeDataItemVerified is emitted when a data item is verified
	EventTypeDataItemVerified = "data_item_verified"

	// EventTypeDataItemRevoked is emitted when a data item is revoked
	EventTypeDataItemRevoked = "data_item_revoked"

	// Attribute keys
	AttributeKeyDataID          = "data_id"
	AttributeKeyDataType        = "data_type"
	AttributeKeyOwner           = "owner"
	AttributeKeyVerifier        = "verifier"
	AttributeKeyVerificationLvl = "verification_level"
	AttributeKeyConfidenceScore = "confidence_score"
	AttributeKeyStorageLocation = "storage_location"
	AttributeKeyAuthority       = "authority"
	AttributeKeyReason          = "reason"
	AttributeKeyBlockHeight     = "block_height"
	AttributeKeyTimestamp       = "timestamp"
)
