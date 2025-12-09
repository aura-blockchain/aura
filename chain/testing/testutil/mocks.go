package testutil

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// MockBankKeeper is a mock implementation of the bank keeper
type MockBankKeeper struct {
	Balances map[string]sdk.Coins
}

// NewMockBankKeeper creates a new mock bank keeper
func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		Balances: make(map[string]sdk.Coins),
	}
}

// GetBalance returns the balance for an address and denom
func (m *MockBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if coins, ok := m.Balances[addr.String()]; ok {
		return sdk.NewCoin(denom, coins.AmountOf(denom))
	}
	return sdk.NewCoin(denom, math.ZeroInt())
}

// SpendableCoins returns spendable coins for an address
func (m *MockBankKeeper) SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	if coins, ok := m.Balances[addr.String()]; ok {
		return coins
	}
	return sdk.NewCoins()
}

// SendCoinsFromAccountToModule sends coins from account to module
func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	senderBalance := m.Balances[senderAddr.String()]
	if !senderBalance.IsAllGTE(amt) {
		return fmt.Errorf("insufficient funds")
	}
	m.Balances[senderAddr.String()] = senderBalance.Sub(amt...)
	return nil
}

// SendCoinsFromModuleToAccount sends coins from module to account
func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	current := m.Balances[recipientAddr.String()]
	m.Balances[recipientAddr.String()] = current.Add(amt...)
	return nil
}

// SendCoins is a passthrough for completeness
func (m *MockBankKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	currentFrom := m.Balances[fromAddr.String()]
	if !currentFrom.IsAllGTE(amt) {
		return fmt.Errorf("insufficient funds")
	}
	m.Balances[fromAddr.String()] = currentFrom.Sub(amt...)
	currentTo := m.Balances[toAddr.String()]
	m.Balances[toAddr.String()] = currentTo.Add(amt...)
	return nil
}

func (m *MockBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	// Track minting against the module name for completeness
	m.Balances[moduleName] = m.Balances[moduleName].Add(amt...)
	return nil
}

func (m *MockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	current := m.Balances[moduleName]
	if !current.IsAllGTE(amt) {
		return fmt.Errorf("insufficient module balance")
	}
	m.Balances[moduleName] = current.Sub(amt...)
	return nil
}

// MockAccountKeeper is a mock implementation of the account keeper
type MockAccountKeeper struct {
	Accounts map[string]sdk.AccountI
}

// NewMockAccountKeeper creates a new mock account keeper
func NewMockAccountKeeper() *MockAccountKeeper {
	return &MockAccountKeeper{
		Accounts: make(map[string]sdk.AccountI),
	}
}

// GetAccount returns an account
func (m *MockAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	return m.Accounts[addr.String()]
}

// SetAccount sets an account
func (m *MockAccountKeeper) SetAccount(ctx sdk.Context, acc sdk.AccountI) {
	m.Accounts[acc.GetAddress().String()] = acc
}

// NewAccountWithAddress creates a basic account for an address
func (m *MockAccountKeeper) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	acc := authtypes.NewBaseAccountWithAddress(addr)
	m.SetAccount(ctx, acc)
	return acc
}

// MockStakingKeeper is a mock implementation of the staking keeper
type MockStakingKeeper struct {
	Validators map[string]interface{}
}

// NewMockStakingKeeper creates a new mock staking keeper
func NewMockStakingKeeper() *MockStakingKeeper {
	return &MockStakingKeeper{
		Validators: make(map[string]interface{}),
	}
}

// GetValidator returns a validator
func (m *MockStakingKeeper) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (interface{}, error) {
	if val, ok := m.Validators[addr.String()]; ok {
		return val, nil
	}
	return nil, nil
}

// MockSlashingKeeper is a mock implementation of the slashing keeper
type MockSlashingKeeper struct {
	SlashedValidators map[string]bool
}

// NewMockSlashingKeeper creates a new mock slashing keeper
func NewMockSlashingKeeper() *MockSlashingKeeper {
	return &MockSlashingKeeper{
		SlashedValidators: make(map[string]bool),
	}
}

// Slash marks a validator as slashed
func (m *MockSlashingKeeper) Slash(ctx sdk.Context, addr sdk.ValAddress) error {
	m.SlashedValidators[addr.String()] = true
	return nil
}

// MockGovernanceKeeper is a mock implementation of the governance keeper
type MockGovernanceKeeper struct {
	Proposals map[uint64]interface{}
}

// NewMockGovernanceKeeper creates a new mock governance keeper
func NewMockGovernanceKeeper() *MockGovernanceKeeper {
	return &MockGovernanceKeeper{
		Proposals: make(map[uint64]interface{}),
	}
}

// GetProposal returns a proposal
func (m *MockGovernanceKeeper) GetProposal(ctx sdk.Context, proposalID uint64) (interface{}, bool) {
	prop, ok := m.Proposals[proposalID]
	return prop, ok
}

// SetProposal sets a proposal
func (m *MockGovernanceKeeper) SetProposal(ctx sdk.Context, proposalID uint64, proposal interface{}) {
	m.Proposals[proposalID] = proposal
}

// MockVCRegistryKeeper is a mock implementation of the VC registry keeper
type MockVCRegistryKeeper struct {
	Credentials map[string]interface{}
	IRScores    map[string]uint64
	Verified    map[string]bool
}

// NewMockVCRegistryKeeper creates a new mock VC registry keeper
func NewMockVCRegistryKeeper() *MockVCRegistryKeeper {
	return &MockVCRegistryKeeper{
		Credentials: make(map[string]interface{}),
		IRScores:    make(map[string]uint64),
		Verified:    make(map[string]bool),
	}
}

// VerifyCredential verifies a credential
func (m *MockVCRegistryKeeper) VerifyCredential(ctx sdk.Context, credentialID string) (bool, error) {
	_, exists := m.Credentials[credentialID]
	return exists, nil
}

// GetIRScore returns the IR score for an address
func (m *MockVCRegistryKeeper) GetIRScore(ctx sdk.Context, address string) uint64 {
	if score, ok := m.IRScores[address]; ok {
		return score
	}
	return 0
}

// IsVerified checks if an address is verified
func (m *MockVCRegistryKeeper) IsVerified(ctx sdk.Context, address string) bool {
	if verified, ok := m.Verified[address]; ok {
		return verified
	}
	return false
}

// MockSecurityKeeper is a mock implementation of the security keeper
type MockSecurityKeeper struct {
	SecurityEvents   map[string]interface{}
	PausedModules    map[string]bool
	ReentrantKeys    map[string]bool
	Authorizations   map[string]map[string]bool // address -> action -> authorized
}

// NewMockSecurityKeeper creates a new mock security keeper
func NewMockSecurityKeeper() *MockSecurityKeeper {
	return &MockSecurityKeeper{
		SecurityEvents: make(map[string]interface{}),
		PausedModules:  make(map[string]bool),
		ReentrantKeys:  make(map[string]bool),
		Authorizations: make(map[string]map[string]bool),
	}
}

// RecordSecurityEvent records a security event
func (m *MockSecurityKeeper) RecordSecurityEvent(ctx sdk.Context, eventID string, event interface{}) error {
	m.SecurityEvents[eventID] = event
	return nil
}

// EnterNoReentrant marks a key as entered
func (m *MockSecurityKeeper) EnterNoReentrant(ctx sdk.Context, key string) error {
	if m.ReentrantKeys[key] {
		return fmt.Errorf("reentrant call detected for key: %s", key)
	}
	m.ReentrantKeys[key] = true
	return nil
}

// ExitNoReentrant marks a key as exited
func (m *MockSecurityKeeper) ExitNoReentrant(ctx sdk.Context, key string) {
	delete(m.ReentrantKeys, key)
}

// WithReentrancyGuard executes a function with reentrancy protection
func (m *MockSecurityKeeper) WithReentrancyGuard(ctx sdk.Context, key string, fn func() error) error {
	if err := m.EnterNoReentrant(ctx, key); err != nil {
		return err
	}
	defer m.ExitNoReentrant(ctx, key)
	return fn()
}

// RequireNotPaused checks if a module is not paused
func (m *MockSecurityKeeper) RequireNotPaused(ctx sdk.Context, moduleName string) error {
	if m.PausedModules[moduleName] {
		return fmt.Errorf("module %s is paused", moduleName)
	}
	return nil
}

// PauseModule pauses a module
func (m *MockSecurityKeeper) PauseModule(ctx sdk.Context, moduleName string, pausedBy string) error {
	m.PausedModules[moduleName] = true
	return nil
}

// UnpauseModule unpauses a module
func (m *MockSecurityKeeper) UnpauseModule(ctx sdk.Context, moduleName string, unpausedBy string) error {
	delete(m.PausedModules, moduleName)
	return nil
}

// IsModulePaused checks if a module is paused
func (m *MockSecurityKeeper) IsModulePaused(ctx sdk.Context, moduleName string) bool {
	return m.PausedModules[moduleName]
}

// CheckGuardRateLimit checks rate limit (mock always allows)
func (m *MockSecurityKeeper) CheckGuardRateLimit(ctx sdk.Context, key string, limit uint64, window time.Duration) error {
	return nil
}

// IncrementGuardRateLimit increments rate limit counter (mock does nothing)
func (m *MockSecurityKeeper) IncrementGuardRateLimit(ctx sdk.Context, key string, window time.Duration) {
}

// ValidateAddress validates an address format (mock always succeeds)
func (m *MockSecurityKeeper) ValidateAddress(address string) error {
	if address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	return nil
}

// ValidateAmount validates an amount is within bounds
func (m *MockSecurityKeeper) ValidateAmount(amount math.Int, min, max math.Int) error {
	if amount.LT(min) {
		return fmt.Errorf("amount %s is less than minimum %s", amount, min)
	}
	if !max.IsZero() && amount.GT(max) {
		return fmt.Errorf("amount %s is greater than maximum %s", amount, max)
	}
	return nil
}

// ValidateNonEmpty validates a string is not empty
func (m *MockSecurityKeeper) ValidateNonEmpty(value string, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	return nil
}

// ValidateStringLength validates string length
func (m *MockSecurityKeeper) ValidateStringLength(value string, fieldName string, minLen, maxLen int) error {
	length := len(value)
	if length < minLen {
		return fmt.Errorf("%s length %d is less than minimum %d", fieldName, length, minLen)
	}
	if maxLen > 0 && length > maxLen {
		return fmt.Errorf("%s length %d is greater than maximum %d", fieldName, length, maxLen)
	}
	return nil
}

// CheckAuthorization checks if an address is authorized for an action
func (m *MockSecurityKeeper) CheckAuthorization(ctx sdk.Context, address string, action string) error {
	if actions, ok := m.Authorizations[address]; ok {
		if authorized, exists := actions[action]; exists && authorized {
			return nil
		}
	}
	// Mock allows all by default
	return nil
}

// LogSecurityEvent logs a security event
func (m *MockSecurityKeeper) LogSecurityEvent(ctx sdk.Context, eventType string, severity string, actor string, action string, details string) {
	// Mock implementation - just store the event
	m.SecurityEvents[eventType] = map[string]string{
		"severity": severity,
		"actor":    actor,
		"action":   action,
		"details":  details,
	}
}
