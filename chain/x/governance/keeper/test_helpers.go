package keeper

import (
	"context"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MockStakingKeeper for tests - shared across all test files
type MockStakingKeeper struct {
	delegatorBonded map[string]sdkmath.Int
}

// NewMockStakingKeeper creates a properly initialized MockStakingKeeper
func NewMockStakingKeeper() *MockStakingKeeper {
	return &MockStakingKeeper{
		delegatorBonded: make(map[string]sdkmath.Int),
	}
}

// SetDelegatorBonded sets the bonded amount for a delegator (test helper)
func (m *MockStakingKeeper) SetDelegatorBonded(address string, amount sdkmath.Int) {
	if m.delegatorBonded == nil {
		m.delegatorBonded = make(map[string]sdkmath.Int)
	}
	m.delegatorBonded[address] = amount
}

func (m *MockStakingKeeper) GetDelegatorBonded(ctx context.Context, delegator sdk.AccAddress) (sdkmath.Int, error) {
	if amount, ok := m.delegatorBonded[delegator.String()]; ok {
		return amount, nil
	}
	return sdkmath.ZeroInt(), nil
}

func (m *MockStakingKeeper) TotalBondedTokens(ctx context.Context) (sdkmath.Int, error) {
	total := sdkmath.ZeroInt()
	for _, amount := range m.delegatorBonded {
		total = total.Add(amount)
	}
	return total, nil
}

// MockBankKeeper for tests - shared across all test files
type MockBankKeeper struct {
	balances       map[string]sdk.Coins
	moduleBalances map[string]sdk.Coins
	sendErrors     map[string]error
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if err, exists := m.sendErrors[senderAddr.String()]; exists {
		return err
	}

	// Check if sender has enough balance
	senderBalance := m.balances[senderAddr.String()]
	if !senderBalance.IsAllGTE(amt) {
		return fmt.Errorf("insufficient funds")
	}

	// Deduct from sender
	newBalance := senderBalance.Sub(amt...)
	m.balances[senderAddr.String()] = newBalance

	// Add to module
	moduleBalance := m.moduleBalances[recipientModule]
	m.moduleBalances[recipientModule] = moduleBalance.Add(amt...)

	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	// Check if module has enough balance
	moduleBalance := m.moduleBalances[senderModule]
	if !moduleBalance.IsAllGTE(amt) {
		return fmt.Errorf("insufficient module funds")
	}

	// Deduct from module
	newModuleBalance := moduleBalance.Sub(amt...)
	m.moduleBalances[senderModule] = newModuleBalance

	// Add to recipient
	recipientBalance := m.balances[recipientAddr.String()]
	m.balances[recipientAddr.String()] = recipientBalance.Add(amt...)

	return nil
}

func (m *MockBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if balance, ok := m.balances[addr.String()]; ok {
		amt := balance.AmountOf(denom)
		return sdk.NewCoin(denom, amt)
	}
	return sdk.NewCoin(denom, sdkmath.ZeroInt())
}

// MockSecurityKeeper is a mock security keeper for testing - shared across all test files
type MockSecurityKeeper struct{}

func (m *MockSecurityKeeper) EnterNoReentrant(ctx sdk.Context, key string) error {
	return nil
}

func (m *MockSecurityKeeper) ExitNoReentrant(ctx sdk.Context, key string) {}

func (m *MockSecurityKeeper) WithReentrancyGuard(ctx sdk.Context, key string, fn func() error) error {
	return fn()
}

func (m *MockSecurityKeeper) RequireNotPaused(ctx sdk.Context, moduleName string) error {
	return nil
}

func (m *MockSecurityKeeper) PauseModule(ctx sdk.Context, moduleName string, pausedBy string) error {
	return nil
}

func (m *MockSecurityKeeper) UnpauseModule(ctx sdk.Context, moduleName string, unpausedBy string) error {
	return nil
}

func (m *MockSecurityKeeper) IsModulePaused(ctx sdk.Context, moduleName string) bool {
	return false
}

func (m *MockSecurityKeeper) CheckGuardRateLimit(ctx sdk.Context, key string, limit uint64, window time.Duration) error {
	return nil
}

func (m *MockSecurityKeeper) IncrementGuardRateLimit(ctx sdk.Context, key string, window time.Duration) {}

func (m *MockSecurityKeeper) ValidateAddress(address string) error {
	return nil
}

func (m *MockSecurityKeeper) ValidateAmount(amount sdkmath.Int, min, max sdkmath.Int) error {
	return nil
}

func (m *MockSecurityKeeper) ValidateNonEmpty(value string, fieldName string) error {
	return nil
}

func (m *MockSecurityKeeper) ValidateStringLength(value string, fieldName string, minLen, maxLen int) error {
	return nil
}

func (m *MockSecurityKeeper) CheckAuthorization(ctx sdk.Context, address string, action string) error {
	return nil
}

func (m *MockSecurityKeeper) LogSecurityEvent(ctx sdk.Context, eventType string, severity string, actor string, action string, details string) {
}
