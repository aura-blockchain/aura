package security

import (
	"cosmossdk.io/math"
	"fmt"
	"sync"
	"sync/atomic"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ReentrancyGuard provides protection against reentrancy attacks
type ReentrancyGuard struct {
	entered uint32
}

// NewReentrancyGuard creates a new reentrancy guard
func NewReentrancyGuard() *ReentrancyGuard {
	return &ReentrancyGuard{entered: 0}
}

// Enter marks entry into a protected section
// Returns error if already entered (reentrancy detected)
func (rg *ReentrancyGuard) Enter() error {
	if !atomic.CompareAndSwapUint32(&rg.entered, 0, 1) {
		return ErrReentrancyDetected
	}
	return nil
}

// Exit marks exit from a protected section
func (rg *ReentrancyGuard) Exit() {
	atomic.StoreUint32(&rg.entered, 0)
}

// WithReentrancyGuard executes a function with reentrancy protection
func (rg *ReentrancyGuard) WithReentrancyGuard(fn func() error) error {
	if err := rg.Enter(); err != nil {
		return err
	}
	defer rg.Exit()
	return fn()
}

// PauseGuard provides emergency pause functionality for modules
type PauseGuard struct {
	mu     sync.RWMutex
	paused bool
	admin  string // Admin address that can pause/unpause
}

// NewPauseGuard creates a new pause guard
func NewPauseGuard(admin string) *PauseGuard {
	return &PauseGuard{
		paused: false,
		admin:  admin,
	}
}

// IsPaused returns whether the module is paused
func (pg *PauseGuard) IsPaused() bool {
	pg.mu.RLock()
	defer pg.mu.RUnlock()
	return pg.paused
}

// Pause pauses the module (only admin can pause)
func (pg *PauseGuard) Pause(ctx sdk.Context, caller string) error {
	pg.mu.Lock()
	defer pg.mu.Unlock()

	if pg.admin != "" && caller != pg.admin {
		return ErrUnauthorized
	}

	if pg.paused {
		return ErrAlreadyPaused
	}

	pg.paused = true

	// Emit pause event if EventManager is available
	// Safely check if context and event manager are valid
	if !ctx.IsZero() && ctx.EventManager() != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"module_paused",
				sdk.NewAttribute("paused_by", caller),
				sdk.NewAttribute("height", fmt.Sprintf("%d", ctx.BlockHeight())),
			),
		)
	}

	return nil
}

// Unpause unpauses the module (only admin can unpause)
func (pg *PauseGuard) Unpause(ctx sdk.Context, caller string) error {
	pg.mu.Lock()
	defer pg.mu.Unlock()

	if pg.admin != "" && caller != pg.admin {
		return ErrUnauthorized
	}

	if !pg.paused {
		return ErrNotPaused
	}

	pg.paused = false

	// Emit unpause event if EventManager is available
	// Safely check if context and event manager are valid
	if !ctx.IsZero() && ctx.EventManager() != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"module_unpaused",
				sdk.NewAttribute("unpaused_by", caller),
				sdk.NewAttribute("height", fmt.Sprintf("%d", ctx.BlockHeight())),
			),
		)
	}

	return nil
}

// CheckNotPaused returns an error if the module is paused
func (pg *PauseGuard) CheckNotPaused() error {
	if pg.IsPaused() {
		return ErrModulePaused
	}
	return nil
}

// UpdateAdmin updates the admin address
func (pg *PauseGuard) UpdateAdmin(newAdmin string, caller string) error {
	pg.mu.Lock()
	defer pg.mu.Unlock()

	if pg.admin != "" && caller != pg.admin {
		return ErrUnauthorized
	}

	pg.admin = newAdmin
	return nil
}

// InputValidator provides comprehensive input validation
type InputValidator struct{}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	return &InputValidator{}
}

// ValidateAddress validates a blockchain address
func (iv *InputValidator) ValidateAddress(addr string) error {
	if addr == "" {
		return ErrInvalidAddress
	}
	if len(addr) < 20 {
		return ErrInvalidAddress
	}
	// Additional validation can be added here
	return nil
}

// ValidateAmount validates a token amount
func (iv *InputValidator) ValidateAmount(amount math.Int) error {
	if amount.IsNil() {
		return ErrInvalidAmount
	}
	if amount.IsNegative() {
		return ErrNegativeAmount
	}
	if amount.IsZero() {
		return ErrZeroAmount
	}
	return nil
}

// ValidateNonNegativeAmount validates that amount is non-negative (can be zero)
func (iv *InputValidator) ValidateNonNegativeAmount(amount math.Int) error {
	if amount.IsNil() {
		return ErrInvalidAmount
	}
	if amount.IsNegative() {
		return ErrNegativeAmount
	}
	return nil
}

// ValidateString validates a non-empty string
func (iv *InputValidator) ValidateString(s string, fieldName string) error {
	if s == "" {
		return fmt.Errorf("field %s cannot be empty", fieldName)
	}
	return nil
}

// ValidateStringLength validates string length
func (iv *InputValidator) ValidateStringLength(s string, fieldName string, minLen, maxLen int) error {
	if len(s) < minLen {
		return fmt.Errorf("field %s must be at least %d characters", fieldName, minLen)
	}
	if len(s) > maxLen {
		return fmt.Errorf("field %s must be at most %d characters", fieldName, maxLen)
	}
	return nil
}

// ValidateSliceNotEmpty validates that a slice is not empty
func (iv *InputValidator) ValidateSliceNotEmpty(slice interface{}, fieldName string) error {
	// Use type assertion to check length
	switch v := slice.(type) {
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("field %s cannot be empty", fieldName)
		}
	case []byte:
		if len(v) == 0 {
			return fmt.Errorf("field %s cannot be empty", fieldName)
		}
	default:
		return fmt.Errorf("unsupported slice type for field %s", fieldName)
	}
	return nil
}

// GasLimitGuard provides gas limit enforcement
type GasLimitGuard struct {
	maxGasPerTx uint64
}

// NewGasLimitGuard creates a new gas limit guard
func NewGasLimitGuard(maxGasPerTx uint64) *GasLimitGuard {
	return &GasLimitGuard{
		maxGasPerTx: maxGasPerTx,
	}
}

// ValidateGasLimit validates that gas limit is within acceptable range
func (glg *GasLimitGuard) ValidateGasLimit(gasLimit uint64) error {
	if gasLimit > glg.maxGasPerTx {
		return ErrGasLimitExceeded
	}
	if gasLimit == 0 {
		return ErrZeroGasLimit
	}
	return nil
}

// CheckGasRemaining checks if sufficient gas remains
func (glg *GasLimitGuard) CheckGasRemaining(ctx sdk.Context, minRequired uint64) error {
	gasRemaining := ctx.GasMeter().Limit() - ctx.GasMeter().GasConsumed()
	if gasRemaining < minRequired {
		return ErrInsufficientGas
	}
	return nil
}

// AtomicityGuard ensures atomic operations
type AtomicityGuard struct {
	mu        sync.Mutex
	rollbacks []func() error
}

// NewAtomicityGuard creates a new atomicity guard
func NewAtomicityGuard() *AtomicityGuard {
	return &AtomicityGuard{
		rollbacks: make([]func() error, 0),
	}
}

// AddRollback adds a rollback function
func (ag *AtomicityGuard) AddRollback(fn func() error) {
	ag.mu.Lock()
	defer ag.mu.Unlock()
	ag.rollbacks = append(ag.rollbacks, fn)
}

// Commit clears all rollback functions
func (ag *AtomicityGuard) Commit() {
	ag.mu.Lock()
	defer ag.mu.Unlock()
	ag.rollbacks = make([]func() error, 0)
}

// Rollback executes all rollback functions in reverse order
func (ag *AtomicityGuard) Rollback() error {
	ag.mu.Lock()
	defer ag.mu.Unlock()

	// Execute rollbacks in reverse order
	for i := len(ag.rollbacks) - 1; i >= 0; i-- {
		if err := ag.rollbacks[i](); err != nil {
			// Log error but continue with other rollbacks
			// In production, this should be properly logged
			continue
		}
	}

	ag.rollbacks = make([]func() error, 0)
	return nil
}

// AccessControl provides role-based access control
type AccessControl struct {
	mu     sync.RWMutex
	admins map[string]bool
	roles  map[string]map[string]bool // address -> role -> has_role
}

// NewAccessControl creates a new access control
func NewAccessControl(initialAdmins []string) *AccessControl {
	admins := make(map[string]bool)
	for _, admin := range initialAdmins {
		admins[admin] = true
	}

	return &AccessControl{
		admins: admins,
		roles:  make(map[string]map[string]bool),
	}
}

// IsAdmin checks if address is an admin
func (ac *AccessControl) IsAdmin(addr string) bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.admins[addr]
}

// AddAdmin adds a new admin (only existing admins can add)
func (ac *AccessControl) AddAdmin(newAdmin string, caller string) error {
	if !ac.IsAdmin(caller) {
		return ErrUnauthorized
	}

	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.admins[newAdmin] = true
	return nil
}

// RemoveAdmin removes an admin (only existing admins can remove)
func (ac *AccessControl) RemoveAdmin(admin string, caller string) error {
	if !ac.IsAdmin(caller) {
		return ErrUnauthorized
	}

	ac.mu.Lock()
	defer ac.mu.Unlock()
	delete(ac.admins, admin)
	return nil
}

// HasRole checks if address has a specific role
func (ac *AccessControl) HasRole(addr string, role string) bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	if roles, ok := ac.roles[addr]; ok {
		return roles[role]
	}
	return false
}

// GrantRole grants a role to an address (only admins can grant)
func (ac *AccessControl) GrantRole(addr string, role string, caller string) error {
	if !ac.IsAdmin(caller) {
		return ErrUnauthorized
	}

	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.roles[addr] == nil {
		ac.roles[addr] = make(map[string]bool)
	}
	ac.roles[addr][role] = true
	return nil
}

// RevokeRole revokes a role from an address (only admins can revoke)
func (ac *AccessControl) RevokeRole(addr string, role string, caller string) error {
	if !ac.IsAdmin(caller) {
		return ErrUnauthorized
	}

	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.roles[addr] != nil {
		delete(ac.roles[addr], role)
	}
	return nil
}
