// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected interface needed for token transfers
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx sdk.Context, acc sdk.AccountI)
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI
}

// VCRegistryKeeper defines the expected interface for vcregistry keeper
// Used for IR verification boost feature
type VCRegistryKeeper interface {
	GetIRScore(ctx sdk.Context, address string) uint64
	IsVerified(ctx sdk.Context, address string) bool
}

// SecurityKeeper defines the expected interface for security keeper
// Provides centralized security primitives (reentrancy guards, pause guards, rate limiting)
type SecurityKeeper interface {
	// Reentrancy Protection
	EnterNoReentrant(ctx sdk.Context, key string) error
	ExitNoReentrant(ctx sdk.Context, key string)
	WithReentrancyGuard(ctx sdk.Context, key string, fn func() error) error

	// Emergency Pause (Circuit Breaker)
	RequireNotPaused(ctx sdk.Context, moduleName string) error
	PauseModule(ctx sdk.Context, moduleName string, pausedBy string) error
	UnpauseModule(ctx sdk.Context, moduleName string, unpausedBy string) error
	IsModulePaused(ctx sdk.Context, moduleName string) bool

	// Rate Limiting
	CheckGuardRateLimit(ctx sdk.Context, key string, limit uint64, window time.Duration) error
	IncrementGuardRateLimit(ctx sdk.Context, key string, window time.Duration)

	// Input Validation
	ValidateAddress(address string) error
	ValidateAmount(amount math.Int, min, max math.Int) error
	ValidateNonEmpty(value string, fieldName string) error
	ValidateStringLength(value string, fieldName string, minLen, maxLen int) error

	// Access Control
	CheckAuthorization(ctx sdk.Context, address string, action string) error

	// Audit Logging
	LogSecurityEvent(ctx sdk.Context, eventType string, severity string, actor string, action string, details string)
}
