// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

// mockBankKeeperWithBalances is an enhanced mock that tracks balances for testing
type mockBankKeeperWithBalances struct {
	balances map[string]map[string]math.Int // address -> denom -> amount
}

func newMockBankKeeperWithBalances() *mockBankKeeperWithBalances {
	return &mockBankKeeperWithBalances{
		balances: make(map[string]map[string]math.Int),
	}
}

func (m *mockBankKeeperWithBalances) SetBalance(addr sdk.AccAddress, denom string, amount math.Int) {
	addrStr := addr.String()
	if m.balances[addrStr] == nil {
		m.balances[addrStr] = make(map[string]math.Int)
	}
	m.balances[addrStr][denom] = amount
}

func (m *mockBankKeeperWithBalances) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	addrStr := addr.String()
	if m.balances[addrStr] != nil {
		if amt, ok := m.balances[addrStr][denom]; ok {
			return sdk.NewCoin(denom, amt)
		}
	}
	return sdk.NewCoin(denom, math.ZeroInt())
}

func (m *mockBankKeeperWithBalances) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m *mockBankKeeperWithBalances) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return nil
}

func (m *mockBankKeeperWithBalances) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m *mockBankKeeperWithBalances) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (m *mockBankKeeperWithBalances) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (m *mockBankKeeperWithBalances) GetSupply(ctx sdk.Context, denom string) sdk.Coin {
	return sdk.NewCoin(denom, math.ZeroInt())
}

// mockAccountKeeperWithModule is an enhanced mock that provides module address
type mockAccountKeeperWithModule struct {
	moduleAddr sdk.AccAddress
}

func newMockAccountKeeperWithModule() *mockAccountKeeperWithModule {
	return &mockAccountKeeperWithModule{
		moduleAddr: sdk.AccAddress([]byte("bridge_module_address")),
	}
}

func (m *mockAccountKeeperWithModule) GetModuleAddress(moduleName string) sdk.AccAddress {
	if moduleName == types.ModuleName {
		return m.moduleAddr
	}
	return nil
}

func (m *mockAccountKeeperWithModule) GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	return nil
}

func (m *mockAccountKeeperWithModule) SetAccount(ctx sdk.Context, acc sdk.AccountI) {
}

func (m *mockAccountKeeperWithModule) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	return nil
}
