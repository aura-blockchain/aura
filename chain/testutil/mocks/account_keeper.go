package mocks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// MockAccountKeeper implements a mock AccountKeeper for testing
type MockAccountKeeper struct {
	Accounts       map[string]sdk.AccountI
	NextAccountNum uint64
	ModuleAccounts map[string]bool
}

// NewMockAccountKeeper creates a new mock account keeper
func NewMockAccountKeeper() *MockAccountKeeper {
	return &MockAccountKeeper{
		Accounts:       make(map[string]sdk.AccountI),
		NextAccountNum: 0,
		ModuleAccounts: make(map[string]bool),
	}
}

// GetAccount returns an account by address
func (m *MockAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	if acc, ok := m.Accounts[addr.String()]; ok {
		return acc
	}
	return nil
}

// SetAccount sets an account
func (m *MockAccountKeeper) SetAccount(ctx sdk.Context, acc sdk.AccountI) {
	m.Accounts[acc.GetAddress().String()] = acc
}

// NewAccountWithAddress creates a new account with the given address
func (m *MockAccountKeeper) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	acc := authtypes.NewBaseAccountWithAddress(addr)
	acc.AccountNumber = m.NextAccountNum
	m.NextAccountNum++
	return acc
}

// NewAccount creates a new account
func (m *MockAccountKeeper) NewAccount(ctx sdk.Context, acc sdk.AccountI) sdk.AccountI {
	baseAcc, ok := acc.(*authtypes.BaseAccount)
	if ok {
		baseAcc.AccountNumber = m.NextAccountNum
		m.NextAccountNum++
	}
	return acc
}

// GetModuleAddress returns the address for a module account
func (m *MockAccountKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	return authtypes.NewModuleAddress(moduleName)
}

// GetModuleAccount returns a module account
//
//nolint:staticcheck // ModuleAccountI is deprecated but still used in SDK v0.50
func (m *MockAccountKeeper) GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI {
	addr := m.GetModuleAddress(moduleName)
	if acc, ok := m.Accounts[addr.String()]; ok {
		if modAcc, ok := acc.(authtypes.ModuleAccountI); ok {
			return modAcc
		}
	}

	// Create new module account if it doesn't exist
	modAcc := authtypes.NewEmptyModuleAccount(moduleName)
	m.Accounts[addr.String()] = modAcc
	m.ModuleAccounts[moduleName] = true
	return modAcc
}

// SetModuleAccount sets a module account
//
//nolint:staticcheck // ModuleAccountI is deprecated but still used in SDK v0.50
func (m *MockAccountKeeper) SetModuleAccount(ctx sdk.Context, macc authtypes.ModuleAccountI) {
	m.Accounts[macc.GetAddress().String()] = macc
	m.ModuleAccounts[macc.GetName()] = true
}

// GetParams returns the auth params (mock implementation)
func (m *MockAccountKeeper) GetParams(ctx sdk.Context) authtypes.Params {
	return authtypes.DefaultParams()
}

// SetParams sets the auth params (mock implementation)
func (m *MockAccountKeeper) SetParams(ctx sdk.Context, params authtypes.Params) error {
	return nil
}

// HasAccount checks if an account exists
func (m *MockAccountKeeper) HasAccount(ctx sdk.Context, addr sdk.AccAddress) bool {
	_, ok := m.Accounts[addr.String()]
	return ok
}

// RemoveAccount removes an account (test helper)
func (m *MockAccountKeeper) RemoveAccount(ctx sdk.Context, addr sdk.AccAddress) {
	delete(m.Accounts, addr.String())
}

// GetAllAccounts returns all accounts (test helper)
func (m *MockAccountKeeper) GetAllAccounts(ctx sdk.Context) []sdk.AccountI {
	accounts := make([]sdk.AccountI, 0, len(m.Accounts))
	for _, acc := range m.Accounts {
		accounts = append(accounts, acc)
	}
	return accounts
}
