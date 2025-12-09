package keeper_test

import (
	"github.com/aequitas/aura/chain/testing/testutil"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dex/keeper"
)

// Note: MockBankKeeper is defined in keeper_comprehensive_test.go to avoid duplication

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
		testutil.NewMockAccountKeeper(),
		testutil.NewMockVCRegistryKeeper(),
		testutil.NewMockSecurityKeeper(),
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
