// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SecurityKeeper defines the interface that modules should use to access
// centralized security primitives (reentrancy guards, pause guards, rate limiting, etc.)
//
// All modules with critical state-changing operations should inject this interface
// and use it to protect against common attack vectors.
type SecurityKeeper interface {
	// Reentrancy Protection
	// Use these to protect state-changing functions from reentrancy attacks
	EnterNoReentrant(ctx sdk.Context, key string) error
	ExitNoReentrant(ctx sdk.Context, key string)
	WithReentrancyGuard(ctx sdk.Context, key string, fn func() error) error

	// Emergency Pause (Circuit Breaker)
	// Use RequireNotPaused at the start of all critical operations
	RequireNotPaused(ctx sdk.Context, moduleName string) error
	PauseModule(ctx sdk.Context, moduleName string, pausedBy string) error
	UnpauseModule(ctx sdk.Context, moduleName string, unpausedBy string) error
	IsModulePaused(ctx sdk.Context, moduleName string) bool

	// Rate Limiting (Guard-specific, distinct from network rate limiting)
	// Use to prevent spam and DoS attacks
	CheckGuardRateLimit(ctx sdk.Context, key string, limit uint64, window time.Duration) error
	IncrementGuardRateLimit(ctx sdk.Context, key string, window time.Duration)

	// Input Validation
	// Use to validate inputs before processing
	ValidateAddress(address string) error
	ValidateAmount(amount math.Int, min, max math.Int) error
	ValidateNonEmpty(value string, fieldName string) error
	ValidateStringLength(value string, fieldName string, minLen, maxLen int) error

	// Access Control
	// Use to check authorization for privileged operations
	CheckAuthorization(ctx sdk.Context, address string, action string) error

	// Audit Logging
	// Use to log security-relevant events
	LogSecurityEvent(ctx sdk.Context, eventType string, severity string, actor string, action string, details string)
}
