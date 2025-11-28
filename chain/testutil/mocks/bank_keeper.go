package mocks

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MockBankKeeper implements a mock BankKeeper for testing
type MockBankKeeper struct {
	Balances        map[string]sdk.Coins
	Supplies        map[string]sdk.Coin
	SendCoinsError  error
	MintCoinsError  error
	BurnCoinsError  error
	BlockedAddrs    map[string]bool
}

// NewMockBankKeeper creates a new mock bank keeper
func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		Balances:     make(map[string]sdk.Coins),
		Supplies:     make(map[string]sdk.Coin),
		BlockedAddrs: make(map[string]bool),
	}
}

// SendCoins mocks sending coins from one account to another
func (m *MockBankKeeper) SendCoins(ctx sdk.Context, from, to sdk.AccAddress, amt sdk.Coins) error {
	if m.SendCoinsError != nil {
		return m.SendCoinsError
	}

	// Simulate balance changes
	fromKey := from.String()
	toKey := to.String()

	if fromBalance, ok := m.Balances[fromKey]; ok {
		// Check if sender has enough balance
		if !fromBalance.IsAllGTE(amt) {
			return fmt.Errorf("insufficient funds")
		}
		m.Balances[fromKey] = fromBalance.Sub(amt...)
	}

	if toBalance, ok := m.Balances[toKey]; ok {
		m.Balances[toKey] = toBalance.Add(amt...)
	} else {
		m.Balances[toKey] = amt
	}

	return nil
}

// GetBalance returns the balance of a specific denom for an address
func (m *MockBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if coins, ok := m.Balances[addr.String()]; ok {
		return sdk.NewCoin(denom, coins.AmountOf(denom))
	}
	return sdk.NewCoin(denom, math.ZeroInt())
}

// GetAllBalances returns all balances for an address
func (m *MockBankKeeper) GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	if coins, ok := m.Balances[addr.String()]; ok {
		return coins
	}
	return sdk.NewCoins()
}

// SpendableCoins returns the spendable balance for an address
func (m *MockBankKeeper) SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return m.GetAllBalances(ctx, addr)
}

// MintCoins mocks minting coins to a module account
func (m *MockBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	if m.MintCoinsError != nil {
		return m.MintCoinsError
	}

	// Update supplies
	for _, coin := range amt {
		if supply, ok := m.Supplies[coin.Denom]; ok {
			m.Supplies[coin.Denom] = supply.Add(coin)
		} else {
			m.Supplies[coin.Denom] = coin
		}
	}

	return nil
}

// BurnCoins mocks burning coins from a module account
func (m *MockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	if m.BurnCoinsError != nil {
		return m.BurnCoinsError
	}

	// Update supplies
	for _, coin := range amt {
		if supply, ok := m.Supplies[coin.Denom]; ok {
			m.Supplies[coin.Denom] = supply.Sub(coin)
		}
	}

	return nil
}

// SendCoinsFromModuleToAccount mocks sending coins from module to account
func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if m.SendCoinsError != nil {
		return m.SendCoinsError
	}

	recipientKey := recipientAddr.String()
	if balance, ok := m.Balances[recipientKey]; ok {
		m.Balances[recipientKey] = balance.Add(amt...)
	} else {
		m.Balances[recipientKey] = amt
	}

	return nil
}

// SendCoinsFromAccountToModule mocks sending coins from account to module
func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if m.SendCoinsError != nil {
		return m.SendCoinsError
	}

	senderKey := senderAddr.String()
	if balance, ok := m.Balances[senderKey]; ok {
		if !balance.IsAllGTE(amt) {
			return fmt.Errorf("insufficient funds")
		}
		m.Balances[senderKey] = balance.Sub(amt...)
	}

	return nil
}

// BlockedAddr checks if an address is blocked
func (m *MockBankKeeper) BlockedAddr(addr sdk.AccAddress) bool {
	return m.BlockedAddrs[addr.String()]
}

// GetSupply returns the total supply of a denom
func (m *MockBankKeeper) GetSupply(ctx sdk.Context, denom string) sdk.Coin {
	if supply, ok := m.Supplies[denom]; ok {
		return supply
	}
	return sdk.NewCoin(denom, math.ZeroInt())
}

// SetBalance is a test helper to set an account balance
func (m *MockBankKeeper) SetBalance(addr sdk.AccAddress, coins sdk.Coins) {
	m.Balances[addr.String()] = coins
}

// BlockAddress is a test helper to block an address
func (m *MockBankKeeper) BlockAddress(addr sdk.AccAddress) {
	m.BlockedAddrs[addr.String()] = true
}

// UnblockAddress is a test helper to unblock an address
func (m *MockBankKeeper) UnblockAddress(addr sdk.AccAddress) {
	delete(m.BlockedAddrs, addr.String())
}
