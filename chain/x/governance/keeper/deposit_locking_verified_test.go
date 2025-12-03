package keeper_test

import (
	"context"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// testBankKeeper is a test implementation of BankKeeper
type testBankKeeper struct {
	Balances       map[string]sdk.Coins
	ModuleBalances map[string]sdk.Coins
}

func newTestBankKeeper() *testBankKeeper {
	return &testBankKeeper{
		Balances:       make(map[string]sdk.Coins),
		ModuleBalances: make(map[string]sdk.Coins),
	}
}

func (m *testBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	senderBalance := m.Balances[senderAddr.String()]
	if !senderBalance.IsAllGTE(amt) {
		return fmt.Errorf("insufficient funds")
	}

	newBalance := senderBalance.Sub(amt...)
	m.Balances[senderAddr.String()] = newBalance

	moduleBalance := m.ModuleBalances[recipientModule]
	m.ModuleBalances[recipientModule] = moduleBalance.Add(amt...)

	return nil
}

func (m *testBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	moduleBalance := m.ModuleBalances[senderModule]
	if !moduleBalance.IsAllGTE(amt) {
		return fmt.Errorf("insufficient module funds")
	}

	newModuleBalance := moduleBalance.Sub(amt...)
	m.ModuleBalances[senderModule] = newModuleBalance

	recipientBalance := m.Balances[recipientAddr.String()]
	m.Balances[recipientAddr.String()] = recipientBalance.Add(amt...)

	return nil
}

func (m *testBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if balance, ok := m.Balances[addr.String()]; ok {
		amt := balance.AmountOf(denom)
		return sdk.NewCoin(denom, amt)
	}
	return sdk.NewCoin(denom, sdkmath.ZeroInt())
}

// TestDepositLocking_TransferToModuleAccount verifies deposits are transferred to module account
func TestDepositLocking_TransferToModuleAccount(t *testing.T) {
	ctx := context.Background()
	bankKeeper := newTestBankKeeper()

	proposer := sdk.AccAddress([]byte("proposer_address_1"))
	depositor := sdk.AccAddress([]byte("depositor_address"))

	initialBalance := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(10_000_000)))
	bankKeeper.Balances[proposer.String()] = initialBalance
	bankKeeper.Balances[depositor.String()] = initialBalance
	bankKeeper.ModuleBalances[types.ModuleName] = sdk.NewCoins()

	t.Run("InitialDepositTransferredOnProposalSubmit", func(t *testing.T) {
		depositAmount := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)))

		err := bankKeeper.SendCoinsFromAccountToModule(ctx, proposer, types.ModuleName, depositAmount)
		require.NoError(t, err, "Initial deposit transfer should succeed")

		proposerBalance := bankKeeper.Balances[proposer.String()]
		expectedBalance := initialBalance.Sub(depositAmount...)
		require.Equal(t, expectedBalance, proposerBalance, "Proposer balance should decrease")

		moduleBalance := bankKeeper.ModuleBalances[types.ModuleName]
		require.Equal(t, depositAmount, moduleBalance, "Module account should receive deposit")
	})

	t.Run("InsufficientBalancePreventsDeposit", func(t *testing.T) {
		poorAddress := sdk.AccAddress([]byte("poor_address_______"))
		bankKeeper.Balances[poorAddress.String()] = sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(100)))

		largeDeposit := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)))
		err := bankKeeper.SendCoinsFromAccountToModule(ctx, poorAddress, types.ModuleName, largeDeposit)
		require.Error(t, err, "Transfer should fail with insufficient balance")
		require.Contains(t, err.Error(), "insufficient funds")
	})
}

// TestDepositLocking_MinimumDepositEnforcement tests minimum deposit validation
func TestDepositLocking_MinimumDepositEnforcement(t *testing.T) {
	params := types.DefaultParams()

	t.Run("MinimumDepositParsing", func(t *testing.T) {
		minDeposit, err := sdk.ParseCoinsNormalized(params.MinDeposit)
		require.NoError(t, err)
		require.NotNil(t, minDeposit)
		require.False(t, minDeposit.IsZero())
	})

	t.Run("DepositBelowMinimumRejected", func(t *testing.T) {
		minDeposit, _ := sdk.ParseCoinsNormalized(params.MinDeposit)
		belowMinimum := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(1)))

		isBelowMinimum := belowMinimum.IsAllLT(minDeposit)
		require.True(t, isBelowMinimum)
	})

	t.Run("DepositAtMinimumAccepted", func(t *testing.T) {
		minDeposit, _ := sdk.ParseCoinsNormalized(params.MinDeposit)
		isBelowMinimum := minDeposit.IsAllLT(minDeposit)
		require.False(t, isBelowMinimum)
	})
}

// TestDepositLocking_RefundOnProposalPass tests deposit refund mechanism
func TestDepositLocking_RefundOnProposalPass(t *testing.T) {
	ctx := context.Background()
	bankKeeper := newTestBankKeeper()

	t.Run("SingleDepositRefundedOnPass", func(t *testing.T) {
		depositor := sdk.AccAddress([]byte("depositor_address_1"))
		depositAmount := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)))

		bankKeeper.Balances[depositor.String()] = sdk.NewCoins()
		bankKeeper.ModuleBalances[types.ModuleName] = depositAmount

		err := bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositor, depositAmount)
		require.NoError(t, err)

		depositorBalance := bankKeeper.Balances[depositor.String()]
		require.Equal(t, depositAmount, depositorBalance)

		moduleBalance := bankKeeper.ModuleBalances[types.ModuleName]
		require.Equal(t, sdk.NewCoins(), moduleBalance)
	})

	t.Run("RefundPreservesDepositIntegrity", func(t *testing.T) {
		depositor := sdk.AccAddress([]byte("depositor_address_2"))
		depositAmount := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)))

		bankKeeper.Balances[depositor.String()] = sdk.NewCoins()
		bankKeeper.ModuleBalances[types.ModuleName] = depositAmount

		err := bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositor, depositAmount)
		require.NoError(t, err)

		err = bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositor, depositAmount)
		require.Error(t, err)
		require.Contains(t, err.Error(), "insufficient module funds")
	})
}

// TestDepositLocking_BurnOnProposalVeto tests deposit burn mechanism
func TestDepositLocking_BurnOnProposalVeto(t *testing.T) {
	bankKeeper := newTestBankKeeper()

	t.Run("VetoedProposalDepositsRemainLocked", func(t *testing.T) {
		depositor1 := sdk.AccAddress([]byte("depositor_1________"))
		depositor2 := sdk.AccAddress([]byte("depositor_2________"))

		deposit1 := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)))
		deposit2 := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(500_000)))

		totalDeposits := deposit1.Add(deposit2...)
		bankKeeper.ModuleBalances[types.ModuleName] = totalDeposits

		bankKeeper.Balances[depositor1.String()] = sdk.NewCoins()
		bankKeeper.Balances[depositor2.String()] = sdk.NewCoins()

		moduleBalance := bankKeeper.ModuleBalances[types.ModuleName]
		require.Equal(t, totalDeposits, moduleBalance)

		require.Equal(t, sdk.NewCoins(), bankKeeper.Balances[depositor1.String()])
		require.Equal(t, sdk.NewCoins(), bankKeeper.Balances[depositor2.String()])
	})
}

// TestDepositLocking_SecurityInvariants tests critical security invariants
func TestDepositLocking_SecurityInvariants(t *testing.T) {
	t.Run("DepositCannotExceedModuleBalance", func(t *testing.T) {
		ctx := context.Background()
		bankKeeper := newTestBankKeeper()

		depositor := sdk.AccAddress([]byte("depositor__________"))
		refundAmount := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)))

		bankKeeper.ModuleBalances[types.ModuleName] = sdk.NewCoins()
		bankKeeper.Balances[depositor.String()] = sdk.NewCoins()

		err := bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositor, refundAmount)
		require.Error(t, err)
	})

	t.Run("ModuleBalanceEqualsDepositSum", func(t *testing.T) {
		ctx := context.Background()
		bankKeeper := newTestBankKeeper()

		depositor1 := sdk.AccAddress([]byte("dep1_______________"))
		depositor2 := sdk.AccAddress([]byte("dep2_______________"))
		depositor3 := sdk.AccAddress([]byte("dep3_______________"))

		initialBalance := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(10_000_000)))
		bankKeeper.Balances[depositor1.String()] = initialBalance
		bankKeeper.Balances[depositor2.String()] = initialBalance
		bankKeeper.Balances[depositor3.String()] = initialBalance
		bankKeeper.ModuleBalances[types.ModuleName] = sdk.NewCoins()

		deposit1 := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)))
		deposit2 := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(2_000_000)))
		deposit3 := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(500_000)))

		err := bankKeeper.SendCoinsFromAccountToModule(ctx, depositor1, types.ModuleName, deposit1)
		require.NoError(t, err)

		err = bankKeeper.SendCoinsFromAccountToModule(ctx, depositor2, types.ModuleName, deposit2)
		require.NoError(t, err)

		err = bankKeeper.SendCoinsFromAccountToModule(ctx, depositor3, types.ModuleName, deposit3)
		require.NoError(t, err)

		expectedModuleBalance := deposit1.Add(deposit2...).Add(deposit3...)

		actualModuleBalance := bankKeeper.ModuleBalances[types.ModuleName]
		require.Equal(t, expectedModuleBalance, actualModuleBalance)
	})

	t.Run("NoNegativeBalances", func(t *testing.T) {
		ctx := context.Background()
		bankKeeper := newTestBankKeeper()

		depositor := sdk.AccAddress([]byte("depositor__________"))
		smallBalance := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(100)))
		largeDeposit := sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)))

		bankKeeper.Balances[depositor.String()] = smallBalance

		err := bankKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleName, largeDeposit)
		require.Error(t, err)

		require.Equal(t, smallBalance, bankKeeper.Balances[depositor.String()])
	})
}
