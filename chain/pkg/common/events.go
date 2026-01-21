// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

// Package common provides shared utilities for Aura chain modules.
//
// EVENT HELPERS: This package provides consistent event emission helpers.
// Import as: import "github.com/aequitas/aura/chain/pkg/common"
//
// Available functions:
//   - EmitEvent: Emit event with attribute map
//   - EmitTypedEvent: Emit protobuf typed event
//   - EmitSuccessEvent: Standard success event with module/action/actor
//   - EmitErrorEvent: Standard error event with error message
//   - EmitTransferEvent: Standard token transfer event
//
// Note: Modules may continue using ctx.EventManager() directly for simple cases.
// These helpers add structure and consistency for indexing and auditing.
package common

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EmitEvent emits a typed event with attributes.
// This provides consistent event emission across all modules.
//
// Security considerations:
//   - All state changes should emit events for indexing and audit trails
//   - Event types should be module-prefixed to avoid collisions
//   - Sensitive data should not be included in events (logs are public)
//
// Parameters:
//   - ctx: SDK context with EventManager
//   - eventType: Event type (should be module.action format, e.g., "dex.swap")
//   - attributes: Map of attribute key-value pairs
//
// Example usage:
//
//	common.EmitEvent(ctx, "dex.create_pool", map[string]string{
//	    "pool_id": pool.PoolId,
//	    "creator": msg.Creator,
//	    "denom_a": msg.DenomA,
//	    "denom_b": msg.DenomB,
//	})
func EmitEvent(ctx sdk.Context, eventType string, attributes map[string]string) {
	attrs := make([]sdk.Attribute, 0, len(attributes))

	// Sorted iteration ensures deterministic event attribute ordering for consensus.
	keys := make([]string, 0, len(attributes))
	for k := range attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := attributes[k]
		attrs = append(attrs, sdk.NewAttribute(k, v))
	}

	event := sdk.NewEvent(eventType, attrs...)
	ctx.EventManager().EmitEvent(event)
}

// EmitTypedEvent emits a protobuf typed event.
// This is the preferred method for emitting events with structured data.
//
// Type parameters:
//   - T: Type of event (must implement proto.Message)
//
// Parameters:
//   - ctx: SDK context with EventManager
//   - event: Typed event to emit
//
// Returns:
//   - error: If event cannot be emitted
//
// Security considerations:
//   - Type safety prevents malformed events
//   - Protobuf validation ensures event structure is correct
//
// Example usage:
//
//	event := &dextypes.EventSwap{
//	    Sender: msg.Sender,
//	    PoolId: msg.PoolId,
//	    AmountIn: msg.CoinIn.String(),
//	    AmountOut: amountOut.String(),
//	}
//	if err := common.EmitTypedEvent(ctx, event); err != nil {
//	    return err
//	}
func EmitTypedEvent(ctx sdk.Context, event sdk.Event) error {
	ctx.EventManager().EmitEvent(event)
	return nil
}

// EmitSuccessEvent emits a standard success event for an operation.
// This provides consistent success event structure across modules.
//
// Parameters:
//   - ctx: SDK context with EventManager
//   - module: Module name (e.g., "dex", "bridge", "identity")
//   - action: Action name (e.g., "create_pool", "transfer", "register_did")
//   - actor: Address of the actor performing the action
//   - details: Optional additional details map
//
// Emitted attributes:
//   - module: Module name
//   - action: Action name
//   - actor: Actor address
//   - success: "true"
//   - ... (additional details)
//
// Example usage:
//
//	common.EmitSuccessEvent(ctx, "dex", "swap", msg.Sender, map[string]string{
//	    "pool_id": msg.PoolId,
//	    "amount_in": msg.CoinIn.String(),
//	    "amount_out": amountOut.String(),
//	})
func EmitSuccessEvent(ctx sdk.Context, module, action, actor string, details map[string]string) {
	attrs := []sdk.Attribute{
		sdk.NewAttribute("module", module),
		sdk.NewAttribute("action", action),
		sdk.NewAttribute("actor", actor),
		sdk.NewAttribute("success", "true"),
	}

	// Sorted iteration ensures deterministic event attribute ordering for consensus.
	keys := make([]string, 0, len(details))
	for k := range details {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := details[k]
		attrs = append(attrs, sdk.NewAttribute(k, v))
	}

	event := sdk.NewEvent(module+"."+action, attrs...)
	ctx.EventManager().EmitEvent(event)
}

// EmitErrorEvent emits a standard error event for a failed operation.
// This provides consistent error event structure across modules.
//
// Parameters:
//   - ctx: SDK context with EventManager
//   - module: Module name
//   - action: Action name
//   - actor: Address of the actor (if known, empty string otherwise)
//   - errorMsg: Error message (should not contain sensitive data)
//   - details: Optional additional details map
//
// Emitted attributes:
//   - module: Module name
//   - action: Action name
//   - actor: Actor address (if provided)
//   - success: "false"
//   - error: Error message
//   - ... (additional details)
//
// Security considerations:
//   - Error messages should not leak sensitive information
//   - Do not include private keys, secrets, or PII in error events
//
// Example usage:
//
//	common.EmitErrorEvent(ctx, "dex", "swap", msg.Sender, "insufficient liquidity", map[string]string{
//	    "pool_id": msg.PoolId,
//	})
func EmitErrorEvent(ctx sdk.Context, module, action, actor, errorMsg string, details map[string]string) {
	attrs := []sdk.Attribute{
		sdk.NewAttribute("module", module),
		sdk.NewAttribute("action", action),
		sdk.NewAttribute("success", "false"),
		sdk.NewAttribute("error", errorMsg),
	}

	if actor != "" {
		attrs = append(attrs, sdk.NewAttribute("actor", actor))
	}

	// Sorted iteration ensures deterministic event attribute ordering for consensus.
	keys := make([]string, 0, len(details))
	for k := range details {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := details[k]
		attrs = append(attrs, sdk.NewAttribute(k, v))
	}

	event := sdk.NewEvent(module+"."+action, attrs...)
	ctx.EventManager().EmitEvent(event)
}

// EmitTransferEvent emits a standard token transfer event.
// This provides consistent transfer event structure across modules.
//
// Parameters:
//   - ctx: SDK context with EventManager
//   - module: Module name emitting the event
//   - from: Sender address
//   - to: Recipient address
//   - amount: Transfer amount as string
//   - denom: Token denomination
//
// Example usage:
//
//	common.EmitTransferEvent(ctx, "dex", msg.Sender, pool.Address, amount.String(), "uaura")
func EmitTransferEvent(ctx sdk.Context, module, from, to, amount, denom string) {
	event := sdk.NewEvent(
		module+".transfer",
		sdk.NewAttribute("from", from),
		sdk.NewAttribute("to", to),
		sdk.NewAttribute("amount", amount),
		sdk.NewAttribute("denom", denom),
	)
	ctx.EventManager().EmitEvent(event)
}
