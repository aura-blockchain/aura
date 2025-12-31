// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEventTypeConstants(t *testing.T) {
	require.Equal(t, "fee_adjusted", EventTypeFeeAdjusted)
	require.Equal(t, "mev_detected", EventTypeMEVDetected)
	require.Equal(t, "mev_prevented", EventTypeMEVPrevented)
	require.Equal(t, "whale_limit_triggered", EventTypeWhaleLimitTriggered)
	require.Equal(t, "circuit_breaker_triggered", EventTypeCircuitBreakerTriggered)
	require.Equal(t, "circuit_breaker_reset", EventTypeCircuitBreakerReset)
	require.Equal(t, "congestion_detected", EventTypeCongestionDetected)
	require.Equal(t, "params_updated", EventTypeParamsUpdated)
	require.Equal(t, "inflation_adjusted", EventTypeInflationAdjusted)
}

func TestAttributeKeyConstants(t *testing.T) {
	require.NotEmpty(t, AttributeKeyModuleName)
	require.NotEmpty(t, AttributeKeyAddress)
	require.NotEmpty(t, AttributeKeyOldFee)
	require.NotEmpty(t, AttributeKeyNewFee)
	require.NotEmpty(t, AttributeKeyFeeMultiplier)
	require.NotEmpty(t, AttributeKeyBaseFee)
	require.NotEmpty(t, AttributeKeyCongestionLevel)
	require.NotEmpty(t, AttributeKeyTxHash)
	require.NotEmpty(t, AttributeKeyMEVType)
	require.NotEmpty(t, AttributeKeyProtectionMethod)
	require.NotEmpty(t, AttributeKeyTransactionAmount)
	require.NotEmpty(t, AttributeKeyLimit)
	require.NotEmpty(t, AttributeKeyTimeWindow)
	require.NotEmpty(t, AttributeKeyCurrentUsage)
	require.NotEmpty(t, AttributeKeyThreshold)
	require.NotEmpty(t, AttributeKeyReason)
	require.NotEmpty(t, AttributeKeyCooldownSeconds)
	require.NotEmpty(t, AttributeKeyBlockHeight)
	require.NotEmpty(t, AttributeKeyBlockTime)
}

func TestNewFeeAdjustedEvent(t *testing.T) {
	event := NewFeeAdjustedEvent(
		"economicsecurity",
		"1000", "1500",
		"1.5",
		3,
		12345, "2024-01-01T00:00:00Z",
	)

	require.Equal(t, "economicsecurity", event[AttributeKeyModuleName])
	require.Equal(t, "1000", event[AttributeKeyOldFee])
	require.Equal(t, "1500", event[AttributeKeyNewFee])
	require.Equal(t, "1.5", event[AttributeKeyFeeMultiplier])
	require.Equal(t, "3", event[AttributeKeyCongestionLevel])
	require.Equal(t, "12345", event[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-01T00:00:00Z", event[AttributeKeyBlockTime])
}

func TestNewFeeAdjustedEvent_LargeValues(t *testing.T) {
	event := NewFeeAdjustedEvent(
		"test_module",
		"999999999999", "1000000000000",
		"100.5",
		4294967295, // max uint32
		9223372036854775807, "2099-12-31T23:59:59Z",
	)

	require.Equal(t, "4294967295", event[AttributeKeyCongestionLevel])
	require.Equal(t, "9223372036854775807", event[AttributeKeyBlockHeight])
}

func TestNewMEVDetectedEvent(t *testing.T) {
	event := NewMEVDetectedEvent(
		"0xabc123",
		"sandwich",
		"aura1victim",
		100, "2024-01-01T00:00:00Z",
	)

	require.Equal(t, "0xabc123", event[AttributeKeyTxHash])
	require.Equal(t, "sandwich", event[AttributeKeyMEVType])
	require.Equal(t, "aura1victim", event[AttributeKeyAddress])
	require.Equal(t, "100", event[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-01T00:00:00Z", event[AttributeKeyBlockTime])
}

func TestNewMEVPreventedEvent(t *testing.T) {
	event := NewMEVPreventedEvent(
		"0xdef456",
		"frontrun",
		"private_mempool",
		200, "2024-01-02T00:00:00Z",
	)

	require.Equal(t, "0xdef456", event[AttributeKeyTxHash])
	require.Equal(t, "frontrun", event[AttributeKeyMEVType])
	require.Equal(t, "private_mempool", event[AttributeKeyProtectionMethod])
	require.Equal(t, "200", event[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-02T00:00:00Z", event[AttributeKeyBlockTime])
}

func TestNewWhaleLimitTriggeredEvent(t *testing.T) {
	event := NewWhaleLimitTriggeredEvent(
		"aura1whale",
		"1000000000",
		"500000000",
		"750000000",
		86400,
		300, "2024-01-03T00:00:00Z",
	)

	require.Equal(t, "aura1whale", event[AttributeKeyAddress])
	require.Equal(t, "1000000000", event[AttributeKeyTransactionAmount])
	require.Equal(t, "500000000", event[AttributeKeyLimit])
	require.Equal(t, "750000000", event[AttributeKeyCurrentUsage])
	require.Equal(t, "86400", event[AttributeKeyTimeWindow])
	require.Equal(t, "300", event[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-03T00:00:00Z", event[AttributeKeyBlockTime])
}

func TestNewCircuitBreakerTriggeredEvent(t *testing.T) {
	event := NewCircuitBreakerTriggeredEvent(
		"dex",
		"high_volatility",
		1000000,
		3600,
		400, "2024-01-04T00:00:00Z",
	)

	require.Equal(t, "dex", event[AttributeKeyModuleName])
	require.Equal(t, "high_volatility", event[AttributeKeyReason])
	require.Equal(t, "1000000", event[AttributeKeyThreshold])
	require.Equal(t, "3600", event[AttributeKeyCooldownSeconds])
	require.Equal(t, "400", event[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-04T00:00:00Z", event[AttributeKeyBlockTime])
}

func TestNewCircuitBreakerResetEvent(t *testing.T) {
	event := NewCircuitBreakerResetEvent(
		"dex",
		500, "2024-01-05T00:00:00Z",
	)

	require.Equal(t, "dex", event[AttributeKeyModuleName])
	require.Equal(t, "500", event[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-05T00:00:00Z", event[AttributeKeyBlockTime])
}

func TestNewCongestionDetectedEvent(t *testing.T) {
	event := NewCongestionDetectedEvent(
		5,
		"2.5",
		600, "2024-01-06T00:00:00Z",
	)

	require.Equal(t, "5", event[AttributeKeyCongestionLevel])
	require.Equal(t, "2.5", event[AttributeKeyFeeMultiplier])
	require.Equal(t, "600", event[AttributeKeyBlockHeight])
	require.Equal(t, "2024-01-06T00:00:00Z", event[AttributeKeyBlockTime])
}

func TestFormatInt64(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "-1"},
		{9223372036854775807, "9223372036854775807"},
		{-9223372036854775808, "-9223372036854775808"},
	}

	for _, tt := range tests {
		result := formatInt64(tt.input)
		require.Equal(t, tt.expected, result)
	}
}

func TestFormatUint32(t *testing.T) {
	tests := []struct {
		input    uint32
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{4294967295, "4294967295"},
	}

	for _, tt := range tests {
		result := formatUint32(tt.input)
		require.Equal(t, tt.expected, result)
	}
}

func TestFormatUint64(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{18446744073709551615, "18446744073709551615"},
	}

	for _, tt := range tests {
		result := formatUint64(tt.input)
		require.Equal(t, tt.expected, result)
	}
}
