package keeper

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MockStakingKeeper for tests - shared across all test files
type MockStakingKeeper struct {
	delegatorBonded map[string]sdkmath.Int
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
