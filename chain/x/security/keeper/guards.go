package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/common/security"
	"github.com/aequitas/aura/chain/x/security/types"
)

// SecurityGuards provides centralized security primitives for all modules
// This implements the security guard functionality that all modules should use
type SecurityGuards struct {
	reentrancyGuard *security.ReentrancyGuard
	pauseGuard      *security.PauseGuard
	inputValidator  *security.InputValidator
	safeMath        *security.SafeMath
	gasLimitGuard   *security.GasLimitGuard
	accessControl   *security.AccessControl
}

// NewSecurityGuards creates a new security guards instance
func NewSecurityGuards(admin string, maxGasPerTx uint64, initialAdmins []string) *SecurityGuards {
	return &SecurityGuards{
		reentrancyGuard: security.NewReentrancyGuard(),
		pauseGuard:      security.NewPauseGuard(admin),
		inputValidator:  security.NewInputValidator(),
		safeMath:        security.NewSafeMath(),
		gasLimitGuard:   security.NewGasLimitGuard(maxGasPerTx),
		accessControl:   security.NewAccessControl(initialAdmins),
	}
}

// =============================================================================
// Reentrancy Protection
// =============================================================================

// EnterNoReentrant marks entry into a protected section with a scoped lock
// key: unique identifier for the protected resource (e.g., "swap:pool123", "bridge:transfer456")
// Returns error if that key is already locked (reentrancy detected)
func (k Keeper) EnterNoReentrant(ctx sdk.Context, key string) error {
	store := k.GetMemStore(ctx)
	lockKey := types.GetReentrancyLockKey(key)

	// Check if already locked
	if store.Has(lockKey) {
		return types.ErrReentrancyDetected
	}

	// Set lock
	store.Set(lockKey, []byte{1})

	// Log for auditing
	k.Logger(ctx).Debug("reentrancy guard entered",
		"key", key,
		"height", ctx.BlockHeight(),
	)

	return nil
}

// ExitNoReentrant marks exit from a protected section
func (k Keeper) ExitNoReentrant(ctx sdk.Context, key string) {
	store := k.GetMemStore(ctx)
	lockKey := types.GetReentrancyLockKey(key)

	// Remove lock
	store.Delete(lockKey)

	k.Logger(ctx).Debug("reentrancy guard exited",
		"key", key,
		"height", ctx.BlockHeight(),
	)
}

// WithReentrancyGuard executes a function with reentrancy protection
func (k Keeper) WithReentrancyGuard(ctx sdk.Context, key string, fn func() error) error {
	if err := k.EnterNoReentrant(ctx, key); err != nil {
		return err
	}
	defer k.ExitNoReentrant(ctx, key)
	return fn()
}

// =============================================================================
// Pause Guard - Emergency Circuit Breaker
// =============================================================================

// RequireNotPaused checks if a module is paused and returns an error if it is
// moduleName: name of the module to check (e.g., "dex", "bridge", "governance")
func (k Keeper) RequireNotPaused(ctx sdk.Context, moduleName string) error {
	store := k.GetStore(ctx)
	pauseKey := types.GetPauseStateKey(moduleName)

	if store.Has(pauseKey) {
		bz := store.Get(pauseKey)
		if len(bz) > 0 && bz[0] == 1 {
			return types.ErrSystemPaused.Wrapf("module %s is paused", moduleName)
		}
	}

	return nil
}

// PauseModule pauses a specific module (emergency circuit breaker)
// Only governance or emergency admin can pause
func (k Keeper) PauseModule(ctx sdk.Context, moduleName string, pausedBy string) error {
	// Check authorization
	if pausedBy != k.GetAuthority() {
		return types.ErrUnauthorized.Wrap("only governance can pause modules")
	}

	store := k.GetStore(ctx)
	pauseKey := types.GetPauseStateKey(moduleName)

	// Check if already paused
	if store.Has(pauseKey) {
		bz := store.Get(pauseKey)
		if len(bz) > 0 && bz[0] == 1 {
			return types.ErrInvalidState.Wrap("module already paused")
		}
	}

	// Set pause state
	store.Set(pauseKey, []byte{1})

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSystemPaused,
			sdk.NewAttribute("module", moduleName),
			sdk.NewAttribute("paused_by", pausedBy),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyTimestamp, ctx.BlockTime().Format(time.RFC3339)),
		),
	)

	k.Logger(ctx).Info("module paused",
		"module", moduleName,
		"paused_by", pausedBy,
		"height", ctx.BlockHeight(),
	)

	return nil
}

// UnpauseModule unpauses a specific module
// Only governance or emergency admin can unpause
func (k Keeper) UnpauseModule(ctx sdk.Context, moduleName string, unpausedBy string) error {
	// Check authorization
	if unpausedBy != k.GetAuthority() {
		return types.ErrUnauthorized.Wrap("only governance can unpause modules")
	}

	store := k.GetStore(ctx)
	pauseKey := types.GetPauseStateKey(moduleName)

	// Check if actually paused
	if !store.Has(pauseKey) {
		return types.ErrInvalidState.Wrap("module not paused")
	}

	bz := store.Get(pauseKey)
	if len(bz) == 0 || bz[0] != 1 {
		return types.ErrInvalidState.Wrap("module not paused")
	}

	// Remove pause state
	store.Delete(pauseKey)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSystemResumed,
			sdk.NewAttribute("module", moduleName),
			sdk.NewAttribute("unpaused_by", unpausedBy),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyTimestamp, ctx.BlockTime().Format(time.RFC3339)),
		),
	)

	k.Logger(ctx).Info("module unpaused",
		"module", moduleName,
		"unpaused_by", unpausedBy,
		"height", ctx.BlockHeight(),
	)

	return nil
}

// IsModulePaused checks if a module is currently paused
func (k Keeper) IsModulePaused(ctx sdk.Context, moduleName string) bool {
	store := k.GetStore(ctx)
	pauseKey := types.GetPauseStateKey(moduleName)

	if store.Has(pauseKey) {
		bz := store.Get(pauseKey)
		return len(bz) > 0 && bz[0] == 1
	}

	return false
}

// =============================================================================
// Rate Limiting
// =============================================================================

// CheckRateLimit checks if an operation is within rate limits
// key: unique identifier for rate limiting (e.g., "swap:user_address", "bridge:chain_id")
// limit: maximum number of operations allowed
// window: time window for the limit (e.g., 1 minute, 1 hour)
func (k Keeper) CheckRateLimit(ctx sdk.Context, key string, limit uint64, window time.Duration) error {
	store := k.GetMemStore(ctx)
	rateLimitKey := types.GetRateLimitStoreKey(fmt.Sprintf("guard:%s", key))

	bz := store.Get(rateLimitKey)
	if bz == nil {
		// First operation, no limit yet
		return nil
	}

	// Unmarshal rate limit data
	var rl types.RateLimit
	if err := k.cdc.Unmarshal(bz, &rl); err != nil {
		// If unmarshal fails, allow operation (fail-open for availability)
		k.Logger(ctx).Error("failed to unmarshal rate limit", "error", err)
		return nil
	}

	currentTime := ctx.BlockTime()
	windowStart := currentTime.Add(-window)

	// Clean up old timestamps outside the window
	validTimestamps := make([]time.Time, 0)
	for _, ts := range rl.Timestamps {
		if ts.After(windowStart) {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check if limit exceeded
	if uint64(len(validTimestamps)) >= limit {
		return types.ErrRateLimitExceeded.Wrapf(
			"rate limit exceeded for %s: %d operations in %v (limit: %d)",
			key, len(validTimestamps), window, limit,
		)
	}

	return nil
}

// IncrementRateLimit increments the rate limit counter for a key
func (k Keeper) IncrementRateLimit(ctx sdk.Context, key string, window time.Duration) {
	store := k.GetMemStore(ctx)
	rateLimitKey := types.GetRateLimitStoreKey(fmt.Sprintf("guard:%s", key))

	var rl types.RateLimit
	bz := store.Get(rateLimitKey)

	currentTime := ctx.BlockTime()
	windowStart := currentTime.Add(-window)

	if bz != nil {
		// Unmarshal existing rate limit
		if err := k.cdc.Unmarshal(bz, &rl); err != nil {
			// If unmarshal fails, reset rate limit
			rl = types.RateLimit{
				Identifier: key,
				Timestamps: []time.Time{},
			}
		}
	} else {
		// Create new rate limit entry
		rl = types.RateLimit{
			Identifier: key,
			Timestamps: []time.Time{},
		}
	}

	// Clean up old timestamps and add current one
	validTimestamps := make([]time.Time, 0)
	for _, ts := range rl.Timestamps {
		if ts.After(windowStart) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	validTimestamps = append(validTimestamps, currentTime)
	rl.Timestamps = validTimestamps

	// Marshal and store
	bz, err := k.cdc.Marshal(&rl)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal rate limit", "error", err)
		return
	}
	store.Set(rateLimitKey, bz)
}

// =============================================================================
// Input Validation
// =============================================================================

// ValidateAddress validates a blockchain address
func (k Keeper) ValidateAddress(address string) error {
	if address == "" {
		return types.ErrInvalidRequest.Wrap("address cannot be empty")
	}

	// Try to parse as SDK address
	_, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return types.ErrInvalidRequest.Wrapf("invalid address format: %v", err)
	}

	return nil
}

// ValidateAmount validates that an amount is within acceptable bounds
func (k Keeper) ValidateAmount(amount math.Int, min, max math.Int) error {
	if amount.IsNil() {
		return types.ErrInvalidRequest.Wrap("amount cannot be nil")
	}

	if amount.IsNegative() {
		return types.ErrInvalidRequest.Wrap("amount cannot be negative")
	}

	if amount.LT(min) {
		return types.ErrInvalidRequest.Wrapf("amount %s below minimum %s", amount.String(), min.String())
	}

	if !max.IsZero() && amount.GT(max) {
		return types.ErrInvalidRequest.Wrapf("amount %s exceeds maximum %s", amount.String(), max.String())
	}

	return nil
}

// ValidateNonEmpty validates that a string is not empty
func (k Keeper) ValidateNonEmpty(value string, fieldName string) error {
	if value == "" {
		return types.ErrInvalidRequest.Wrapf("%s cannot be empty", fieldName)
	}
	return nil
}

// ValidateStringLength validates string length bounds
func (k Keeper) ValidateStringLength(value string, fieldName string, minLen, maxLen int) error {
	if len(value) < minLen {
		return types.ErrInvalidRequest.Wrapf(
			"%s must be at least %d characters (got %d)",
			fieldName, minLen, len(value),
		)
	}

	if maxLen > 0 && len(value) > maxLen {
		return types.ErrInvalidRequest.Wrapf(
			"%s must be at most %d characters (got %d)",
			fieldName, maxLen, len(value),
		)
	}

	return nil
}

// =============================================================================
// Access Control
// =============================================================================

// CheckAuthorization checks if an address is authorized for a specific action
// This can be extended to support role-based access control
func (k Keeper) CheckAuthorization(ctx sdk.Context, address string, action string) error {
	// Check if address is governance authority
	if address == k.GetAuthority() {
		return nil
	}

	// Check module-specific permissions (can be extended)
	// For now, only governance authority can perform privileged actions
	return types.ErrUnauthorized.Wrapf("address %s not authorized for action %s", address, action)
}

// =============================================================================
// Audit Logging
// =============================================================================

// LogSecurityEvent logs a security-relevant event for auditing
func (k Keeper) LogSecurityEvent(
	ctx sdk.Context,
	eventType string,
	severity string,
	actor string,
	action string,
	details string,
) {
	// Create audit log entry
	logID := fmt.Sprintf("audit_%d_%s", ctx.BlockHeight(), ctx.TxIndex())

	entry := &types.AuditLogEntry{
		LogId:       logID,
		Timestamp:   ctx.BlockTime().Unix(),
		Severity:    severity,
		EventType:   eventType,
		Actor:       actor,
		Action:      action,
		Details:     details,
		BlockHeight: ctx.BlockHeight(),
	}

	k.SetAuditLogEntry(ctx, entry)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"security_audit",
			sdk.NewAttribute("log_id", logID),
			sdk.NewAttribute("event_type", eventType),
			sdk.NewAttribute("severity", severity),
			sdk.NewAttribute("actor", actor),
			sdk.NewAttribute("action", action),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	k.Logger(ctx).Info("security event logged",
		"log_id", logID,
		"event_type", eventType,
		"severity", severity,
		"actor", actor,
	)
}
