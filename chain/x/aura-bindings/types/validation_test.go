// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/aura-bindings/types"
)

func TestErrorTypes(t *testing.T) {
	errors := []struct {
		name string
		err  error
	}{
		{"InvalidQuery", types.ErrInvalidQuery},
		{"InvalidMessage", types.ErrInvalidMessage},
		{"QueryRateLimitExceeded", types.ErrQueryRateLimitExceeded},
		{"MessageRateLimitExceeded", types.ErrMessageRateLimitExceeded},
		{"Unauthorized", types.ErrUnauthorized},
		{"InvalidParam", types.ErrInvalidParam},
		{"QueryFailed", types.ErrQueryFailed},
		{"MessageFailed", types.ErrMessageFailed},
		{"InvalidAddress", types.ErrInvalidAddress},
		{"NotFound", types.ErrNotFound},
	}

	for _, tc := range errors {
		t.Run(tc.name, func(t *testing.T) {
			require.NotNil(t, tc.err)
			require.Error(t, tc.err)
		})
	}
}

func TestModuleConstants(t *testing.T) {
	require.Equal(t, "aurabindings", types.ModuleName)
	require.Equal(t, "aurabindings", types.StoreKey)
	require.Equal(t, "aurabindings", types.RouterKey)
	require.Equal(t, "aurabindings", types.QuerierRoute)
	require.Equal(t, "mem_aurabindings", types.MemStoreKey)
}

func TestMaxConstants(t *testing.T) {
	require.Equal(t, 1000, types.MaxQueriesPerBlock)
	require.Equal(t, 100, types.MaxMessagesPerBlock)
	require.Greater(t, types.MaxQueriesPerBlock, 0)
	require.Greater(t, types.MaxMessagesPerBlock, 0)
}

func TestKeyPrefixes(t *testing.T) {
	require.NotNil(t, types.QueryStatsPrefix)
	require.NotNil(t, types.MessageStatsPrefix)
	require.NotNil(t, types.RateLimitPrefix)

	require.Len(t, types.QueryStatsPrefix, 1)
	require.Len(t, types.MessageStatsPrefix, 1)
	require.Len(t, types.RateLimitPrefix, 1)

	// Ensure prefixes are unique
	require.NotEqual(t, types.QueryStatsPrefix, types.MessageStatsPrefix)
	require.NotEqual(t, types.QueryStatsPrefix, types.RateLimitPrefix)
	require.NotEqual(t, types.MessageStatsPrefix, types.RateLimitPrefix)
}

func TestEventTypes(t *testing.T) {
	eventTypes := []string{
		types.EventTypeCustomQuery,
		types.EventTypeCustomMessage,
		types.EventTypeRateLimitHit,
		types.EventTypeQueryStats,
		types.EventTypeMessageStats,
	}

	for _, eventType := range eventTypes {
		require.NotEmpty(t, eventType)
	}
}

func TestAttributeKeys(t *testing.T) {
	attributeKeys := []string{
		types.AttributeKeyQueryType,
		types.AttributeKeyMessageType,
		types.AttributeKeyAddress,
		types.AttributeKeySuccess,
		types.AttributeKeyError,
		types.AttributeKeyCount,
		types.AttributeKeyBlockHeight,
		types.AttributeKeyContractAddr,
		types.AttributeKeyQueryCount,
		types.AttributeKeyMessageCount,
		types.AttributeKeyRateLimitType,
	}

	for _, key := range attributeKeys {
		require.NotEmpty(t, key)
	}
}

func TestNewQueryEvent(t *testing.T) {
	event := types.NewQueryEvent("test_query", "aura1test", true, "")
	require.NotNil(t, event)
	require.Equal(t, types.EventTypeCustomQuery, event.Type)
	require.NotEmpty(t, event.Attributes)
}

func TestNewQueryEventWithError(t *testing.T) {
	event := types.NewQueryEvent("test_query", "aura1test", false, "test error")
	require.NotNil(t, event)
	require.Equal(t, types.EventTypeCustomQuery, event.Type)
	require.NotEmpty(t, event.Attributes)

	// Verify error attribute is present
	hasError := false
	for _, attr := range event.Attributes {
		if attr.Key == types.AttributeKeyError {
			hasError = true
			require.Equal(t, "test error", attr.Value)
		}
	}
	require.True(t, hasError)
}

func TestNewMessageEvent(t *testing.T) {
	event := types.NewMessageEvent("test_msg", "aura1test", true, "")
	require.NotNil(t, event)
	require.Equal(t, types.EventTypeCustomMessage, event.Type)
	require.NotEmpty(t, event.Attributes)
}

func TestNewMessageEventWithError(t *testing.T) {
	event := types.NewMessageEvent("test_msg", "aura1test", false, "test error")
	require.NotNil(t, event)
	require.Equal(t, types.EventTypeCustomMessage, event.Type)
	require.NotEmpty(t, event.Attributes)

	// Verify error attribute is present
	hasError := false
	for _, attr := range event.Attributes {
		if attr.Key == types.AttributeKeyError {
			hasError = true
			require.Equal(t, "test error", attr.Value)
		}
	}
	require.True(t, hasError)
}

func TestNewRateLimitEvent(t *testing.T) {
	event := types.NewRateLimitEvent("aura1test", "query", 100)
	require.NotNil(t, event)
	require.Equal(t, types.EventTypeRateLimitHit, event.Type)
	require.NotEmpty(t, event.Attributes)
}
