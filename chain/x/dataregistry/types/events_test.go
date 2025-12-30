// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEventTypeConstants(t *testing.T) {
	// Verify all event type constants are unique and non-empty
	eventTypes := []string{
		EventTypeDataItemStored,
		EventTypeDataItemUpdated,
		EventTypeDataItemDeleted,
		EventTypeDataItemVerified,
		EventTypeDataItemRevoked,
	}

	seen := make(map[string]struct{})
	for _, eventType := range eventTypes {
		require.NotEmpty(t, eventType, "event type should not be empty")
		_, exists := seen[eventType]
		require.False(t, exists, "duplicate event type: %s", eventType)
		seen[eventType] = struct{}{}
	}
}

func TestEventTypeValues(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "data item stored",
			constant: EventTypeDataItemStored,
			expected: "data_item_stored",
		},
		{
			name:     "data item updated",
			constant: EventTypeDataItemUpdated,
			expected: "data_item_updated",
		},
		{
			name:     "data item deleted",
			constant: EventTypeDataItemDeleted,
			expected: "data_item_deleted",
		},
		{
			name:     "data item verified",
			constant: EventTypeDataItemVerified,
			expected: "data_item_verified",
		},
		{
			name:     "data item revoked",
			constant: EventTypeDataItemRevoked,
			expected: "data_item_revoked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestAttributeKeyConstants(t *testing.T) {
	// Verify all attribute key constants are unique and non-empty
	attributeKeys := []string{
		AttributeKeyDataID,
		AttributeKeyDataType,
		AttributeKeyOwner,
		AttributeKeyVerifier,
		AttributeKeyVerificationLvl,
		AttributeKeyConfidenceScore,
		AttributeKeyStorageLocation,
		AttributeKeyAuthority,
		AttributeKeyReason,
		AttributeKeyBlockHeight,
		AttributeKeyTimestamp,
	}

	seen := make(map[string]struct{})
	for _, key := range attributeKeys {
		require.NotEmpty(t, key, "attribute key should not be empty")
		_, exists := seen[key]
		require.False(t, exists, "duplicate attribute key: %s", key)
		seen[key] = struct{}{}
	}
}

func TestAttributeKeyValues(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "data ID key",
			constant: AttributeKeyDataID,
			expected: "data_id",
		},
		{
			name:     "data type key",
			constant: AttributeKeyDataType,
			expected: "data_type",
		},
		{
			name:     "owner key",
			constant: AttributeKeyOwner,
			expected: "owner",
		},
		{
			name:     "verifier key",
			constant: AttributeKeyVerifier,
			expected: "verifier",
		},
		{
			name:     "verification level key",
			constant: AttributeKeyVerificationLvl,
			expected: "verification_level",
		},
		{
			name:     "confidence score key",
			constant: AttributeKeyConfidenceScore,
			expected: "confidence_score",
		},
		{
			name:     "storage location key",
			constant: AttributeKeyStorageLocation,
			expected: "storage_location",
		},
		{
			name:     "authority key",
			constant: AttributeKeyAuthority,
			expected: "authority",
		},
		{
			name:     "reason key",
			constant: AttributeKeyReason,
			expected: "reason",
		},
		{
			name:     "block height key",
			constant: AttributeKeyBlockHeight,
			expected: "block_height",
		},
		{
			name:     "timestamp key",
			constant: AttributeKeyTimestamp,
			expected: "timestamp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestEventConstantsFollowConventions(t *testing.T) {
	// Verify event types use snake_case
	eventTypes := []string{
		EventTypeDataItemStored,
		EventTypeDataItemUpdated,
		EventTypeDataItemDeleted,
		EventTypeDataItemVerified,
		EventTypeDataItemRevoked,
	}

	for _, eventType := range eventTypes {
		// Check that it's lowercase with underscores
		for _, c := range eventType {
			isValid := (c >= 'a' && c <= 'z') || c == '_'
			require.True(t, isValid, "event type %s should be snake_case, found char: %c", eventType, c)
		}
	}

	// Verify attribute keys use snake_case
	attributeKeys := []string{
		AttributeKeyDataID,
		AttributeKeyDataType,
		AttributeKeyOwner,
		AttributeKeyVerifier,
		AttributeKeyVerificationLvl,
		AttributeKeyConfidenceScore,
		AttributeKeyStorageLocation,
		AttributeKeyAuthority,
		AttributeKeyReason,
		AttributeKeyBlockHeight,
		AttributeKeyTimestamp,
	}

	for _, key := range attributeKeys {
		for _, c := range key {
			isValid := (c >= 'a' && c <= 'z') || c == '_'
			require.True(t, isValid, "attribute key %s should be snake_case, found char: %c", key, c)
		}
	}
}
