package types

import (
	"context"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected bank keeper interface for managing token transfers
type BankKeeper interface {
	// SendCoinsFromAccountToModule transfers coins from an account to a module account
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	// SendCoinsFromModuleToAccount transfers coins from a module account to an account
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error

	// GetBalance returns the balance of a specific denom for an account
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
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
