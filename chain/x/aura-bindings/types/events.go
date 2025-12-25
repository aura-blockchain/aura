// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Event types for aura-bindings module
const (
	EventTypeCustomQuery    = "custom_query"
	EventTypeCustomMessage  = "custom_message"
	EventTypeRateLimitHit   = "rate_limit_hit"
	EventTypeQueryStats     = "query_stats"
	EventTypeMessageStats   = "message_stats"

	AttributeKeyQueryType     = "query_type"
	AttributeKeyMessageType   = "message_type"
	AttributeKeyAddress       = "address"
	AttributeKeySuccess       = "success"
	AttributeKeyError         = "error"
	AttributeKeyCount         = "count"
	AttributeKeyBlockHeight   = "block_height"
	AttributeKeyContractAddr  = "contract_address"
	AttributeKeyQueryCount    = "query_count"
	AttributeKeyMessageCount  = "message_count"
	AttributeKeyRateLimitType = "rate_limit_type"
)

// NewQueryEvent creates a new query event
func NewQueryEvent(queryType, address string, success bool, errorMsg string) sdk.Event {
	attributes := []sdk.Attribute{
		sdk.NewAttribute(AttributeKeyQueryType, queryType),
		sdk.NewAttribute(AttributeKeyAddress, address),
		sdk.NewAttribute(AttributeKeySuccess, fmt.Sprintf("%t", success)),
	}

	if errorMsg != "" {
		attributes = append(attributes, sdk.NewAttribute(AttributeKeyError, errorMsg))
	}

	return sdk.NewEvent(
		EventTypeCustomQuery,
		attributes...,
	)
}

// NewMessageEvent creates a new message event
func NewMessageEvent(msgType, address string, success bool, errorMsg string) sdk.Event {
	attributes := []sdk.Attribute{
		sdk.NewAttribute(AttributeKeyMessageType, msgType),
		sdk.NewAttribute(AttributeKeyAddress, address),
		sdk.NewAttribute(AttributeKeySuccess, fmt.Sprintf("%t", success)),
	}

	if errorMsg != "" {
		attributes = append(attributes, sdk.NewAttribute(AttributeKeyError, errorMsg))
	}

	return sdk.NewEvent(
		EventTypeCustomMessage,
		attributes...,
	)
}

// NewRateLimitEvent creates a new rate limit hit event
func NewRateLimitEvent(address, limitType string, count int) sdk.Event {
	return sdk.NewEvent(
		EventTypeRateLimitHit,
		sdk.NewAttribute(AttributeKeyAddress, address),
		sdk.NewAttribute(AttributeKeyRateLimitType, limitType),
		sdk.NewAttribute(AttributeKeyCount, fmt.Sprintf("%d", count)),
	)
}
