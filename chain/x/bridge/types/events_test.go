// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTransferInitiatedEvent(t *testing.T) {
	attrs := NewTransferInitiatedEvent(
		"transfer-123",
		"sender_address",
		"recipient_address",
		"1000",
		"uaura",
		"aura",
		"paw",
		"channel-0",
		12345,
		"2024-01-01T00:00:00Z",
	)

	require.NotNil(t, attrs)
	require.Equal(t, "transfer-123", attrs[AttributeKeyTransferID])
	require.Equal(t, "sender_address", attrs[AttributeKeySender])
	require.Equal(t, "recipient_address", attrs[AttributeKeyRecipient])
	require.Equal(t, "1000", attrs[AttributeKeyAmount])
	require.Equal(t, "uaura", attrs[AttributeKeyDenom])
	require.Equal(t, "aura", attrs[AttributeKeySourceChain])
	require.Equal(t, "paw", attrs[AttributeKeyDestinationChain])
	require.Equal(t, "channel-0", attrs[AttributeKeyChannelID])
	require.Equal(t, "initiated", attrs[AttributeKeyStatus])
	require.Equal(t, "12345", attrs[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-01T00:00:00Z", attrs[AttributeKeyBlockTime])
}

func TestNewTransferCompletedEvent(t *testing.T) {
	attrs := NewTransferCompletedEvent(
		"transfer-456",
		"recipient_address",
		"2000",
		"upaw",
		12,
		67890,
		"2024-01-02T00:00:00Z",
		"0xabcdef123456",
	)

	require.NotNil(t, attrs)
	require.Equal(t, "transfer-456", attrs[AttributeKeyTransferID])
	require.Equal(t, "recipient_address", attrs[AttributeKeyRecipient])
	require.Equal(t, "2000", attrs[AttributeKeyAmount])
	require.Equal(t, "upaw", attrs[AttributeKeyDenom])
	require.Equal(t, "12", attrs[AttributeKeyConfirmations])
	require.Equal(t, "completed", attrs[AttributeKeyStatus])
	require.Equal(t, "67890", attrs[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-02T00:00:00Z", attrs[AttributeKeyBlockTime])
	require.Equal(t, "0xabcdef123456", attrs[AttributeKeyTransactionHash])
}

func TestNewTransferFailedEvent(t *testing.T) {
	attrs := NewTransferFailedEvent(
		"transfer-789",
		"sender_address",
		"insufficient funds",
		11111,
		"2024-01-03T00:00:00Z",
	)

	require.NotNil(t, attrs)
	require.Equal(t, "transfer-789", attrs[AttributeKeyTransferID])
	require.Equal(t, "sender_address", attrs[AttributeKeySender])
	require.Equal(t, "insufficient funds", attrs[AttributeKeyFailureReason])
	require.Equal(t, "failed", attrs[AttributeKeyStatus])
	require.Equal(t, "11111", attrs[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-03T00:00:00Z", attrs[AttributeKeyBlockTime])
}

func TestNewAssetLockedEvent(t *testing.T) {
	attrs := NewAssetLockedEvent(
		"transfer-locked",
		"locker_address",
		"5000",
		"uaura",
		"channel-1",
		22222,
		"2024-01-04T00:00:00Z",
	)

	require.NotNil(t, attrs)
	require.Equal(t, "transfer-locked", attrs[AttributeKeyTransferID])
	require.Equal(t, "locker_address", attrs[AttributeKeySender])
	require.Equal(t, "5000", attrs[AttributeKeyAmount])
	require.Equal(t, "uaura", attrs[AttributeKeyDenom])
	require.Equal(t, "channel-1", attrs[AttributeKeyChannelID])
	require.Equal(t, "22222", attrs[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-04T00:00:00Z", attrs[AttributeKeyBlockTime])
}

func TestNewAssetUnlockedEvent(t *testing.T) {
	attrs := NewAssetUnlockedEvent(
		"transfer-unlocked",
		"recipient_address",
		"3000",
		"upaw",
		"channel-2",
		33333,
		"2024-01-05T00:00:00Z",
	)

	require.NotNil(t, attrs)
	require.Equal(t, "transfer-unlocked", attrs[AttributeKeyTransferID])
	require.Equal(t, "recipient_address", attrs[AttributeKeyRecipient])
	require.Equal(t, "3000", attrs[AttributeKeyAmount])
	require.Equal(t, "upaw", attrs[AttributeKeyDenom])
	require.Equal(t, "channel-2", attrs[AttributeKeyChannelID])
	require.Equal(t, "33333", attrs[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-05T00:00:00Z", attrs[AttributeKeyBlockTime])
}

func TestNewMerkleProofVerifiedEvent(t *testing.T) {
	attrs := NewMerkleProofVerifiedEvent(
		"transfer-merkle",
		"0x1234567890abcdef",
		"0xleafhash",
		8,
		"verifier_address",
		44444,
		"2024-01-06T00:00:00Z",
	)

	require.NotNil(t, attrs)
	require.Equal(t, "transfer-merkle", attrs[AttributeKeyTransferID])
	require.Equal(t, "0x1234567890abcdef", attrs[AttributeKeyMerkleRoot])
	require.Equal(t, "0xleafhash", attrs[AttributeKeyLeafHash])
	require.Equal(t, "8", attrs[AttributeKeyProofDepth])
	require.Equal(t, "verifier_address", attrs[AttributeKeyVerifier])
	require.Equal(t, "44444", attrs[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-06T00:00:00Z", attrs[AttributeKeyBlockTime])
}

func TestNewValidatorAddedEvent(t *testing.T) {
	attrs := NewValidatorAddedEvent(
		"validator_address",
		100,
		5,
		500,
		55555,
		"2024-01-07T00:00:00Z",
	)

	require.NotNil(t, attrs)
	require.Equal(t, "validator_address", attrs[AttributeKeyValidatorAddress])
	require.Equal(t, "100", attrs[AttributeKeyValidatorPower])
	require.Equal(t, "5", attrs[AttributeKeyActiveValidators])
	require.Equal(t, "500", attrs[AttributeKeyTotalPower])
	require.Equal(t, "55555", attrs[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-07T00:00:00Z", attrs[AttributeKeyBlockTime])
}

func TestNewValidatorRemovedEvent(t *testing.T) {
	attrs := NewValidatorRemovedEvent(
		"validator_address",
		4,
		400,
		66666,
		"2024-01-08T00:00:00Z",
	)

	require.NotNil(t, attrs)
	require.Equal(t, "validator_address", attrs[AttributeKeyValidatorAddress])
	require.Equal(t, "4", attrs[AttributeKeyActiveValidators])
	require.Equal(t, "400", attrs[AttributeKeyTotalPower])
	require.Equal(t, "66666", attrs[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-08T00:00:00Z", attrs[AttributeKeyBlockTime])
}

func TestNewValidatorPowerUpdatedEvent(t *testing.T) {
	attrs := NewValidatorPowerUpdatedEvent(
		"validator_address",
		100,
		200,
		77777,
		"2024-01-09T00:00:00Z",
	)

	require.NotNil(t, attrs)
	require.Equal(t, "validator_address", attrs[AttributeKeyValidatorAddress])
	require.Equal(t, "100", attrs["old_power"])
	require.Equal(t, "200", attrs["new_power"])
	require.Equal(t, "77777", attrs[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-09T00:00:00Z", attrs[AttributeKeyBlockTime])
}

func TestNewCircuitBreakerTriggeredEvent(t *testing.T) {
	attrs := NewCircuitBreakerTriggeredEvent(
		"channel-3",
		"excessive failures",
		88888,
		"2024-01-10T00:00:00Z",
	)

	require.NotNil(t, attrs)
	require.Equal(t, "channel-3", attrs[AttributeKeyChannelID])
	require.Equal(t, "excessive failures", attrs[AttributeKeyCircuitBreakerReason])
	require.Equal(t, "88888", attrs[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-10T00:00:00Z", attrs[AttributeKeyBlockTime])
}

func TestNewCircuitBreakerResetEvent(t *testing.T) {
	attrs := NewCircuitBreakerResetEvent(
		"channel-4",
		99999,
		"2024-01-11T00:00:00Z",
	)

	require.NotNil(t, attrs)
	require.Equal(t, "channel-4", attrs[AttributeKeyChannelID])
	require.Equal(t, "99999", attrs[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-11T00:00:00Z", attrs[AttributeKeyBlockTime])
}

func TestNewSecurityViolationDetectedEvent(t *testing.T) {
	attrs := NewSecurityViolationDetectedEvent(
		"double-spend",
		"attempted double spending on transfer-123",
		"transfer-123",
		12345,
		"2024-01-12T00:00:00Z",
	)

	require.NotNil(t, attrs)
	require.Equal(t, "double-spend", attrs[AttributeKeyViolationType])
	require.Equal(t, "attempted double spending on transfer-123", attrs[AttributeKeyViolationDetails])
	require.Equal(t, "transfer-123", attrs[AttributeKeyTransferID])
	require.Equal(t, "12345", attrs[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-12T00:00:00Z", attrs[AttributeKeyBlockTime])
}

func TestNewChannelOpenedEvent(t *testing.T) {
	attrs := NewChannelOpenedEvent(
		"channel-0",
		"aura",
		"paw",
		23456,
		"2024-01-13T00:00:00Z",
	)

	require.NotNil(t, attrs)
	require.Equal(t, "channel-0", attrs[AttributeKeyChannelID])
	require.Equal(t, "aura", attrs[AttributeKeySourceChain])
	require.Equal(t, "paw", attrs[AttributeKeyDestinationChain])
	require.Equal(t, "23456", attrs[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-13T00:00:00Z", attrs[AttributeKeyBlockTime])
}

func TestNewChannelClosedEvent(t *testing.T) {
	attrs := NewChannelClosedEvent(
		"channel-1",
		"manual shutdown",
		34567,
		"2024-01-14T00:00:00Z",
	)

	require.NotNil(t, attrs)
	require.Equal(t, "channel-1", attrs[AttributeKeyChannelID])
	require.Equal(t, "manual shutdown", attrs["close_reason"])
	require.Equal(t, "34567", attrs[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-14T00:00:00Z", attrs[AttributeKeyBlockTime])
}

func TestNewParamsUpdatedEvent(t *testing.T) {
	attrs := NewParamsUpdatedEvent(
		"updater_address",
		12,
		3600,
		45678,
		"2024-01-15T00:00:00Z",
	)

	require.NotNil(t, attrs)
	require.Equal(t, "updater_address", attrs["updated_by"])
	require.Equal(t, "12", attrs["confirmation_depth"])
	require.Equal(t, "3600", attrs["timeout_period"])
	require.Equal(t, "45678", attrs[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-15T00:00:00Z", attrs[AttributeKeyBlockTime])
}

func TestFormatInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"positive", 12345, "12345"},
		{"negative", -6789, "-6789"},
		{"zero", 0, "0"},
		{"large", 9223372036854775807, "9223372036854775807"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatInt64(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatUint32(t *testing.T) {
	tests := []struct {
		name     string
		input    uint32
		expected string
	}{
		{"small", 123, "123"},
		{"zero", 0, "0"},
		{"large", 4294967295, "4294967295"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUint32(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
