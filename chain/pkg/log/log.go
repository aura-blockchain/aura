// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

// Package log provides structured logging helpers for Aura chain modules.
//
// USAGE: These helpers are available for all modules to provide consistent,
// structured logging. Import as: import auralog "github.com/aequitas/aura/chain/pkg/log"
//
// Available functions:
//   - TxStart, TxSuccess, TxError: Transaction lifecycle logging
//   - SecurityEvent: Security-relevant event logging
//   - StateChange, StateChangeWithAttrs: State modification logging
//   - Debug, Info, Warn: General purpose logging
//
// Note: Modules may continue using ctx.Logger() directly for simple cases.
// These helpers add structure and consistency for security auditing.
package log

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TxStart logs the beginning of a transaction with the message type and sender
func TxStart(ctx sdk.Context, msgType string, sender string) {
	ctx.Logger().Info("transaction started",
		"msg_type", msgType,
		"sender", sender,
	)
}

// TxSuccess logs successful transaction completion with attributes
func TxSuccess(ctx sdk.Context, msgType string, attrs ...interface{}) {
	if len(attrs)%2 != 0 {
		ctx.Logger().Error("TxSuccess called with odd number of attributes")
		return
	}

	logAttrs := []interface{}{"msg_type", msgType, "status", "success"}
	logAttrs = append(logAttrs, attrs...)
	ctx.Logger().Info("transaction completed", logAttrs...)
}

// TxError logs transaction failure with error details and attributes
func TxError(ctx sdk.Context, msgType string, err error, attrs ...interface{}) {
	if len(attrs)%2 != 0 {
		ctx.Logger().Error("TxError called with odd number of attributes")
		return
	}

	logAttrs := []interface{}{"msg_type", msgType, "status", "error", "error", err.Error()}
	logAttrs = append(logAttrs, attrs...)
	ctx.Logger().Error("transaction failed", logAttrs...)
}

// SecurityEvent logs security-relevant events with structured attributes
func SecurityEvent(ctx sdk.Context, event string, attrs ...interface{}) {
	if len(attrs)%2 != 0 {
		ctx.Logger().Error("SecurityEvent called with odd number of attributes")
		return
	}

	logAttrs := []interface{}{"event_type", "security", "event", event}
	logAttrs = append(logAttrs, attrs...)
	ctx.Logger().Info("security event", logAttrs...)
}

// StateChange logs state modifications with entity, action, and identifier
func StateChange(ctx sdk.Context, entity string, action string, id string) {
	ctx.Logger().Info("state change",
		"entity", entity,
		"action", action,
		"id", id,
	)
}

// StateChangeWithAttrs logs state modifications with additional attributes
func StateChangeWithAttrs(ctx sdk.Context, entity string, action string, id string, attrs ...interface{}) {
	if len(attrs)%2 != 0 {
		ctx.Logger().Error("StateChangeWithAttrs called with odd number of attributes")
		return
	}

	logAttrs := []interface{}{"entity", entity, "action", action, "id", id}
	logAttrs = append(logAttrs, attrs...)
	ctx.Logger().Info("state change", logAttrs...)
}

// Debug logs debug-level messages (only shown in debug mode)
func Debug(ctx sdk.Context, msg string, attrs ...interface{}) {
	if len(attrs)%2 != 0 {
		ctx.Logger().Error("Debug called with odd number of attributes")
		return
	}
	ctx.Logger().Debug(msg, attrs...)
}

// Info logs informational messages
func Info(ctx sdk.Context, msg string, attrs ...interface{}) {
	if len(attrs)%2 != 0 {
		ctx.Logger().Error("Info called with odd number of attributes")
		return
	}
	ctx.Logger().Info(msg, attrs...)
}

// Warn logs warning messages
func Warn(ctx sdk.Context, msg string, attrs ...interface{}) {
	if len(attrs)%2 != 0 {
		ctx.Logger().Error("Warn called with odd number of attributes")
		return
	}
	ctx.Logger().Error(fmt.Sprintf("WARNING: %s", msg), attrs...)
}
