package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dex/keeper"
)

// MockBankKeeper is a mock implementation of BankKeeper for testing
type MockBankKeeper struct {
	balances map[string]map[string]sdkmath.Int // address -> denom -> amount
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		balances: make(map[string]map[string]sdkmath.Int),
	}
}

func (m *MockBankKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	for _, coin := range amt {
		m.addBalance(toAddr, coin.Denom, coin.Amount)
	}
	return nil
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	addrStr := senderAddr.String()
	_, addressKnown := m.balances[addrStr]

	if addressKnown {
		for _, coin := range amt {
			balance := m.GetBalance(ctx, senderAddr, coin.Denom)
			if balance.Amount.LT(coin.Amount) {
				return fmt.Errorf("insufficient balance: have %s, need %s %s",
					balance.Amount.String(), coin.Amount.String(), coin.Denom)
			}
			newBalance := balance.Amount.Sub(coin.Amount)
			m.SetBalance(senderAddr, coin.Denom, newBalance)
		}
		return nil
	}

	for _, coin := range amt {
		if coin.Amount.GT(sdkmath.ZeroInt()) {
			return fmt.Errorf("insufficient balance: have 0, need %s %s",
				coin.Amount.String(), coin.Denom)
		}
	}
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	for _, coin := range amt {
		m.addBalance(recipientAddr, coin.Denom, coin.Amount)
	}
	return nil
}

func (m *MockBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	addrStr := addr.String()
	if m.balances[addrStr] == nil {
		return sdk.NewCoin(denom, sdkmath.ZeroInt())
	}
	amount := m.balances[addrStr][denom]
	if amount.IsNil() {
		return sdk.NewCoin(denom, sdkmath.ZeroInt())
	}
	return sdk.NewCoin(denom, amount)
}

func (m *MockBankKeeper) SetBalance(addr sdk.AccAddress, denom string, amount sdkmath.Int) {
	addrStr := addr.String()
	if m.balances[addrStr] == nil {
		m.balances[addrStr] = make(map[string]sdkmath.Int)
	}
	m.balances[addrStr][denom] = amount
}

func (m *MockBankKeeper) addBalance(addr sdk.AccAddress, denom string, amount sdkmath.Int) {
	current := m.GetBalance(nil, addr, denom)
	m.SetBalance(addr, denom, current.Amount.Add(amount))
}

// KeeperTestSuite provides common test setup for DEX keeper tests
type KeeperTestSuite struct {
	DexKeeper     *keeper.Keeper
	Keeper        *keeper.Keeper // Alias for DexKeeper for compatibility
	BankKeeper    *MockBankKeeper
	Ctx           sdk.Context
	TestAccs      []sdk.AccAddress
	TestAccounts  []sdk.AccAddress // Alias for TestAccs for compatibility
}

// SetupKeeperTestSuite creates a test suite with keeper and mock dependencies
func SetupKeeperTestSuite(t *testing.T) *KeeperTestSuite {
	// Configure SDK with Aura-specific prefixes
	keepertest.ConfigureSDK()

	input := keepertest.CreateTestInput(t)
	bankKeeper := NewMockBankKeeper()

	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		bankKeeper,
		nil, // accountKeeper
		nil, // vcKeeper
		nil, // securityKeeper
	)

	// Generate test accounts
	testAccs := keepertest.GenTestAddrs(10)

	return &KeeperTestSuite{
		DexKeeper:    k,
		Keeper:       k, // Set both for compatibility
		BankKeeper:   bankKeeper,
		Ctx:          input.Ctx,
		TestAccs:     testAccs,
		TestAccounts: testAccs, // Set both for compatibility
	}
}

// FundAccount funds an account with coins using the mock bank keeper
func (suite *KeeperTestSuite) FundAccount(bankKeeper *MockBankKeeper, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	for _, coin := range coins {
		bankKeeper.SetBalance(addr, coin.Denom, coin.Amount)
	}
}
