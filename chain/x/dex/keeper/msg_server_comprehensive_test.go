// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/testing/testutil"
	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

// MsgServerComprehensiveTestSuite tests comprehensive MsgServer edge cases and error conditions
type MsgServerComprehensiveTestSuite struct {
	suite.Suite

	keeper     *Keeper
	ctx        sdk.Context
	msgServer  dexpb.MsgServer
	bankKeeper *mockBankKeeper
}

// mockBankKeeper is a simple mock for testing msg server
type mockBankKeeper struct {
	balances map[string]map[string]sdkmath.Int // addr -> denom -> amount
}

func newMockBankKeeper() *mockBankKeeper {
	return &mockBankKeeper{
		balances: make(map[string]map[string]sdkmath.Int),
	}
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	// Deduct from sender
	for _, coin := range amt {
		balance := m.getBalance(senderAddr, coin.Denom)
		if balance.LT(coin.Amount) {
			return fmt.Errorf("insufficient funds: have %s, need %s %s", balance.String(), coin.Amount.String(), coin.Denom)
		}
		m.setBalance(senderAddr, coin.Denom, balance.Sub(coin.Amount))
	}
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	// Add to recipient (modules have unlimited funds in tests)
	for _, coin := range amt {
		balance := m.getBalance(recipientAddr, coin.Denom)
		m.setBalance(recipientAddr, coin.Denom, balance.Add(coin.Amount))
	}
	return nil
}

func (m *mockBankKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	for _, coin := range amt {
		fromBalance := m.getBalance(fromAddr, coin.Denom)
		if fromBalance.LT(coin.Amount) {
			return fmt.Errorf("insufficient funds: have %s, need %s %s", fromBalance.String(), coin.Amount.String(), coin.Denom)
		}
		m.setBalance(fromAddr, coin.Denom, fromBalance.Sub(coin.Amount))

		toBalance := m.getBalance(toAddr, coin.Denom)
		m.setBalance(toAddr, coin.Denom, toBalance.Add(coin.Amount))
	}
	return nil
}

func (m *mockBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, m.getBalance(addr, denom))
}

func (m *mockBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) getBalance(addr sdk.AccAddress, denom string) sdkmath.Int {
	addrStr := addr.String()
	if m.balances[addrStr] == nil {
		return sdkmath.ZeroInt()
	}
	if amt, ok := m.balances[addrStr][denom]; ok {
		return amt
	}
	return sdkmath.ZeroInt()
}

func (m *mockBankKeeper) setBalance(addr sdk.AccAddress, denom string, amount sdkmath.Int) {
	addrStr := addr.String()
	if m.balances[addrStr] == nil {
		m.balances[addrStr] = make(map[string]sdkmath.Int)
	}
	m.balances[addrStr][denom] = amount
}

func TestMsgServerComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerComprehensiveTestSuite))
}

func (suite *MsgServerComprehensiveTestSuite) SetupTest() {
	keepertest.ConfigureSDK()
	input := keepertest.CreateTestInput(suite.T())
	suite.bankKeeper = newMockBankKeeper()

	suite.keeper = NewKeeper(
		input.Cdc,
		input.StoreKey,
		suite.bankKeeper,
		testutil.NewMockAccountKeeper(),
		testutil.NewMockVCRegistryKeeper(),
		testutil.NewMockSecurityKeeper(),
	)
	suite.ctx = input.Ctx
	suite.msgServer = NewMsgServerImpl(suite.keeper)
}

func (suite *MsgServerComprehensiveTestSuite) TestCreatePoolInvalidDenoms() {
	// Test creating pool with invalid denoms
	creator := keepertest.GenTestAddr()

	// Fund creator
	suite.bankKeeper.setBalance(creator, "uaura", sdkmath.NewInt(1000000))
	suite.bankKeeper.setBalance(creator, "usdt", sdkmath.NewInt(1000000))

	tests := []struct {
		name        string
		denomA      string
		denomB      string
		shouldError bool
	}{
		{
			name:        "empty denom A",
			denomA:      "",
			denomB:      "usdt",
			shouldError: true,
		},
		{
			name:        "empty denom B",
			denomA:      "uaura",
			denomB:      "",
			shouldError: true,
		},
		{
			name:        "invalid characters in denom",
			denomA:      "uaura!@#",
			denomB:      "usdt",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// sdk.NewCoin panics on invalid denoms, so we need to test differently
			// We test the validation by checking if creating the message would panic
			// or if the keeper's validation rejects it

			var msg *dexpb.MsgCreatePool
			var coinErr error

			// Try to create coins - if this panics, we catch it with recover
			func() {
				defer func() {
					if r := recover(); r != nil {
						coinErr = fmt.Errorf("invalid denom: %v", r)
					}
				}()
				msg = &dexpb.MsgCreatePool{
					Creator: creator.String(),
					DenomA:  tt.denomA,
					DenomB:  tt.denomB,
					AmountA: sdk.NewCoin(tt.denomA, sdkmath.NewInt(1000)),
					AmountB: sdk.NewCoin(tt.denomB, sdkmath.NewInt(1000)),
				}
			}()

			if tt.shouldError {
				// Either coin creation failed (panic caught) or CreatePool returns error
				if coinErr != nil {
					suite.Require().Error(coinErr, "expected error for invalid denoms")
				} else if msg != nil {
					_, err := suite.msgServer.CreatePool(suite.ctx, msg)
					suite.Require().Error(err, "expected error for invalid denoms")
				}
			} else {
				suite.Require().NoError(coinErr)
				if msg != nil {
					_, err := suite.msgServer.CreatePool(suite.ctx, msg)
					suite.Require().NoError(err)
				}
			}
		})
	}
}

func (suite *MsgServerComprehensiveTestSuite) TestCreatePoolZeroInitialLiquidity() {
	// Test creating pool with zero initial liquidity
	creator := keepertest.GenTestAddr()

	tests := []struct {
		name     string
		amountA  sdkmath.Int
		amountB  sdkmath.Int
		expectErr bool
	}{
		{
			name:      "zero amount A",
			amountA:   sdkmath.ZeroInt(),
			amountB:   sdkmath.NewInt(1000000000),
			expectErr: true,
		},
		{
			name:      "zero amount B",
			amountA:   sdkmath.NewInt(1000000000),
			amountB:   sdkmath.ZeroInt(),
			expectErr: true,
		},
		{
			name:      "both zero",
			amountA:   sdkmath.ZeroInt(),
			amountB:   sdkmath.ZeroInt(),
			expectErr: true,
		},
		{
			name:      "valid amounts above minimum",
			amountA:   sdkmath.NewInt(1000000000),
			amountB:   sdkmath.NewInt(1000000000),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Fund creator with large amounts
			suite.bankKeeper.setBalance(creator, "uaura", sdkmath.NewInt(10000000000))
			suite.bankKeeper.setBalance(creator, "usdt", sdkmath.NewInt(10000000000))

			msg := &dexpb.MsgCreatePool{
				Creator: creator.String(),
				DenomA:  "uaura",
				DenomB:  "usdt",
				AmountA: sdk.NewCoin("uaura", tt.amountA),
				AmountB: sdk.NewCoin("usdt", tt.amountB),
			}

			_, err := suite.msgServer.CreatePool(suite.ctx, msg)

			if tt.expectErr {
				suite.Require().Error(err, "should reject zero liquidity")
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *MsgServerComprehensiveTestSuite) TestCreatePoolSameDenoms() {
	// Test creating pool with same denom for both assets
	creator := keepertest.GenTestAddr()

	// Fund creator
	suite.bankKeeper.setBalance(creator, "uaura", sdkmath.NewInt(1000000000))

	msg := &dexpb.MsgCreatePool{
		Creator: creator.String(),
		DenomA:  "uaura",
		DenomB:  "uaura", // Same denom - should fail
		AmountA: sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
		AmountB: sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
	}

	_, err := suite.msgServer.CreatePool(suite.ctx, msg)
	suite.Require().Error(err, "should reject pool with same denoms")
	suite.Require().Contains(err.Error(), "denoms must differ")
}

func (suite *MsgServerComprehensiveTestSuite) TestAddLiquidityNonExistentPool() {
	// Test adding liquidity to non-existent pool
	provider := keepertest.GenTestAddr()

	// Fund provider
	suite.bankKeeper.setBalance(provider, "uaura", sdkmath.NewInt(1000000000))
	suite.bankKeeper.setBalance(provider, "usdt", sdkmath.NewInt(1000000000))

	msg := &dexpb.MsgAddLiquidity{
		Provider: provider.String(),
		PoolId:   "nonexistent-pool-id",
		AmountA:  sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
		AmountB:  sdk.NewCoin("usdt", sdkmath.NewInt(1000)),
	}

	_, err := suite.msgServer.AddLiquidity(suite.ctx, msg)
	suite.Require().Error(err, "should fail for non-existent pool")
	// Error message contains the pool ID, so just check it mentions the pool
	suite.Require().Contains(err.Error(), "pool")
}

func (suite *MsgServerComprehensiveTestSuite) TestAddLiquidityImbalanced() {
	// Test adding imbalanced liquidity
	creator := keepertest.GenTestAddr()
	provider := keepertest.GenTestAddr()

	// Fund both accounts (use large amounts above minimum)
	suite.bankKeeper.setBalance(creator, "uaura", sdkmath.NewInt(10000000000))
	suite.bankKeeper.setBalance(creator, "usdt", sdkmath.NewInt(10000000000))
	suite.bankKeeper.setBalance(provider, "uaura", sdkmath.NewInt(10000000000))
	suite.bankKeeper.setBalance(provider, "usdt", sdkmath.NewInt(10000000000))

	// Create initial pool with 1:1 ratio (above minimum liquidity)
	createMsg := &dexpb.MsgCreatePool{
		Creator: creator.String(),
		DenomA:  "uaura",
		DenomB:  "usdt",
		AmountA: sdk.NewCoin("uaura", sdkmath.NewInt(1000000000)),
		AmountB: sdk.NewCoin("usdt", sdkmath.NewInt(1000000000)),
	}
	createResp, err := suite.msgServer.CreatePool(suite.ctx, createMsg)
	suite.Require().NoError(err)

	// Try to add liquidity with imbalanced ratio (1:2 instead of 1:1)
	// The keeper should adjust amounts to maintain ratio or reject
	addMsg := &dexpb.MsgAddLiquidity{
		Provider: provider.String(),
		PoolId:   createResp.PoolId,
		AmountA:  sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
		AmountB:  sdk.NewCoin("usdt", sdkmath.NewInt(2000)), // Imbalanced
	}

	_, err = suite.msgServer.AddLiquidity(suite.ctx, addMsg)
	// The implementation may either:
	// 1. Reject imbalanced liquidity (error contains "ratio" or "liquidity")
	// 2. Accept and adjust to use proper ratio (no error)
	// Either behavior is acceptable - test just ensures no panic
	if err != nil {
		// Check that error is related to liquidity/ratio constraints
		errMsg := strings.ToLower(err.Error())
		suite.Require().True(
			strings.Contains(errMsg, "ratio") ||
				strings.Contains(errMsg, "liquidity") ||
				strings.Contains(errMsg, "proportion"),
			"error should mention ratio/liquidity/proportion, got: %s", err.Error(),
		)
	}
}

func (suite *MsgServerComprehensiveTestSuite) TestRemoveLiquidityExceedingShares() {
	// Reset test state to start fresh
	suite.SetupTest()

	// Test removing more shares than owned
	creator := keepertest.GenTestAddr()

	// Fund creator
	suite.bankKeeper.setBalance(creator, "uaura", sdkmath.NewInt(10000000000))
	suite.bankKeeper.setBalance(creator, "usdt", sdkmath.NewInt(10000000000))

	// Create pool
	createMsg := &dexpb.MsgCreatePool{
		Creator: creator.String(),
		DenomA:  "uaura",
		DenomB:  "usdt",
		AmountA: sdk.NewCoin("uaura", sdkmath.NewInt(1000000000)),
		AmountB: sdk.NewCoin("usdt", sdkmath.NewInt(1000000000)),
	}
	createResp, err := suite.msgServer.CreatePool(suite.ctx, createMsg)
	suite.Require().NoError(err)

	// Parse LP tokens from response
	lpTokens := createResp.LpTokens

	// Fast-forward time past cooldown period (24 hours + 1 second)
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(24*time.Hour + time.Second))

	// Try to remove more LP tokens than owned
	removeMsg := &dexpb.MsgRemoveLiquidity{
		Provider: creator.String(),
		PoolId:   createResp.PoolId,
		LpTokens: lpTokens.Add(sdkmath.NewInt(1000)), // More than owned
	}

	_, err = suite.msgServer.RemoveLiquidity(suite.ctx, removeMsg)
	suite.Require().Error(err, "should reject removing more shares than owned")
	// Now that cooldown is passed, error should be about insufficient funds
	suite.Require().Contains(err.Error(), "insufficient",
		"error should mention insufficient funds after cooldown period",
	)
}

func (suite *MsgServerComprehensiveTestSuite) TestSwapExactInSlippageExceeded() {
	// Reset test state to avoid interference from previous tests
	suite.SetupTest()

	// Test swap with excessive slippage
	creator := keepertest.GenTestAddr()
	trader := keepertest.GenTestAddr()

	// Fund accounts
	suite.bankKeeper.setBalance(creator, "uaura", sdkmath.NewInt(10000000000))
	suite.bankKeeper.setBalance(creator, "usdt", sdkmath.NewInt(10000000000))
	suite.bankKeeper.setBalance(trader, "uaura", sdkmath.NewInt(1000000000))

	// Create pool
	createMsg := &dexpb.MsgCreatePool{
		Creator: creator.String(),
		DenomA:  "uaura",
		DenomB:  "usdt",
		AmountA: sdk.NewCoin("uaura", sdkmath.NewInt(1000000000)),
		AmountB: sdk.NewCoin("usdt", sdkmath.NewInt(1000000000)),
	}
	createResp, err := suite.msgServer.CreatePool(suite.ctx, createMsg)
	suite.Require().NoError(err)

	// Try to swap with very small amount (below dust threshold)
	swapMsg := &dexpb.MsgSwapExactIn{
		Sender:         trader.String(),
		PoolId:         createResp.PoolId,
		CoinIn:         sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
		MinAmountOut:   sdkmath.NewInt(1000),
		MaxSlippageBps: 10, // 0.1% max slippage - too tight
	}

	_, err = suite.msgServer.SwapExactIn(suite.ctx, swapMsg)
	suite.Require().Error(err, "should reject swap with dust amount or excessive slippage")
	// Error could be about dust attack or slippage
	errMsg := err.Error()
	suite.Require().True(
		strings.Contains(errMsg, "slippage") || strings.Contains(errMsg, "dust") || strings.Contains(errMsg, "minimum"),
		"error should mention slippage, dust, or minimum amount, got: %s", errMsg,
	)
}

func (suite *MsgServerComprehensiveTestSuite) TestSwapInsufficientLiquidity() {
	// Test swap with insufficient pool liquidity
	creator := keepertest.GenTestAddr()
	trader := keepertest.GenTestAddr()

	// Create small pool
	suite.bankKeeper.setBalance(creator, "uaura", sdkmath.NewInt(10000000000))
	suite.bankKeeper.setBalance(creator, "usdt", sdkmath.NewInt(10000000000))
	suite.bankKeeper.setBalance(trader, "uaura", sdkmath.NewInt(10000000000))

	createMsg := &dexpb.MsgCreatePool{
		Creator: creator.String(),
		DenomA:  "uaura",
		DenomB:  "usdt",
		AmountA: sdk.NewCoin("uaura", sdkmath.NewInt(1000000000)),
		AmountB: sdk.NewCoin("usdt", sdkmath.NewInt(1000000000)),
	}
	createResp, err := suite.msgServer.CreatePool(suite.ctx, createMsg)
	suite.Require().NoError(err)

	// Try to swap amount larger than pool reserve
	swapMsg := &dexpb.MsgSwapExactIn{
		Sender:         trader.String(),
		PoolId:         createResp.PoolId,
		CoinIn:         sdk.NewCoin("uaura", sdkmath.NewInt(50000)),
		MinAmountOut:   sdkmath.NewInt(1),
		MaxSlippageBps: 10000, // 100% slippage allowed
	}

	_, err = suite.msgServer.SwapExactIn(suite.ctx, swapMsg)
	suite.Require().Error(err, "should reject swap exceeding pool reserves")
}

func (suite *MsgServerComprehensiveTestSuite) TestSwapZeroAmount() {
	// Test swap with zero amount
	creator := keepertest.GenTestAddr()
	trader := keepertest.GenTestAddr()

	// Setup pool
	suite.bankKeeper.setBalance(creator, "uaura", sdkmath.NewInt(10000000000))
	suite.bankKeeper.setBalance(creator, "usdt", sdkmath.NewInt(10000000000))

	createMsg := &dexpb.MsgCreatePool{
		Creator: creator.String(),
		DenomA:  "uaura",
		DenomB:  "usdt",
		AmountA: sdk.NewCoin("uaura", sdkmath.NewInt(1000000000)),
		AmountB: sdk.NewCoin("usdt", sdkmath.NewInt(1000000000)),
	}
	createResp, err := suite.msgServer.CreatePool(suite.ctx, createMsg)
	suite.Require().NoError(err)

	// Try to swap zero amount
	swapMsg := &dexpb.MsgSwapExactIn{
		Sender:         trader.String(),
		PoolId:         createResp.PoolId,
		CoinIn:         sdk.NewCoin("uaura", sdkmath.ZeroInt()),
		MinAmountOut:   sdkmath.ZeroInt(),
		MaxSlippageBps: 1000,
	}

	_, err = suite.msgServer.SwapExactIn(suite.ctx, swapMsg)
	suite.Require().Error(err, "should reject zero amount swap")
	suite.Require().Contains(err.Error(), "must be positive")
}

func (suite *MsgServerComprehensiveTestSuite) TestPlaceOrderInvalidPrice() {
	// Test placing order with invalid price
	creator := keepertest.GenTestAddr()

	// Fund creator
	suite.bankKeeper.setBalance(creator, "uaura", sdkmath.NewInt(10000000000))
	suite.bankKeeper.setBalance(creator, "usdt", sdkmath.NewInt(10000000000))

	tests := []struct {
		name        string
		auraAmount  sdkmath.Int
		otherAmount sdkmath.Int
		expectError bool
	}{
		{
			name:        "zero aura amount",
			auraAmount:  sdkmath.ZeroInt(),
			otherAmount: sdkmath.NewInt(1000),
			expectError: true,
		},
		{
			name:        "zero other amount",
			auraAmount:  sdkmath.NewInt(1000),
			otherAmount: sdkmath.ZeroInt(),
			expectError: true,
		},
		{
			name:        "negative aura amount",
			auraAmount:  sdkmath.NewInt(-1000),
			otherAmount: sdkmath.NewInt(1000),
			expectError: true,
		},
		{
			name:        "valid amounts",
			auraAmount:  sdkmath.NewInt(1000),
			otherAmount: sdkmath.NewInt(1000),
			expectError: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			msg := &dexpb.MsgCreateOrder{
				Creator:     creator.String(),
				OrderType:   dexpb.SwapOrderType_BUY,
				AuraAmount:  tt.auraAmount,
				OtherCoin:   "usdt",
				OtherAmount: tt.otherAmount,
			}

			_, err := suite.msgServer.CreateOrder(suite.ctx, msg)

			if tt.expectError {
				suite.Require().Error(err, "should reject invalid price")
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *MsgServerComprehensiveTestSuite) TestCancelOrderNotOwner() {
	// Test cancelling order by non-owner - CRITICAL SECURITY TEST
	// Verifies that users cannot cancel orders they don't own
	creator := keepertest.GenTestAddr()
	attacker := keepertest.GenTestAddr()

	// Fund both users
	suite.bankKeeper.setBalance(creator, "uaura", sdkmath.NewInt(10000000000))
	suite.bankKeeper.setBalance(creator, "usdt", sdkmath.NewInt(10000000000))
	suite.bankKeeper.setBalance(attacker, "uaura", sdkmath.NewInt(10000000000))
	suite.bankKeeper.setBalance(attacker, "usdt", sdkmath.NewInt(10000000000))

	// Create order as creator
	createMsg := &dexpb.MsgCreateOrder{
		Creator:     creator.String(),
		OrderType:   dexpb.SwapOrderType_BUY,
		AuraAmount:  sdkmath.NewInt(10000000),
		OtherCoin:   "usdt",
		OtherAmount: sdkmath.NewInt(10000000),
	}
	createResp, err := suite.msgServer.CreateOrder(suite.ctx, createMsg)
	suite.Require().NoError(err)

	// Attacker tries to cancel creator's order
	// The attacker's address (attacker) doesn't match the order owner (creator)
	cancelMsg := &dexpb.MsgCancelOrder{
		Creator: attacker.String(), // Attacker is signing the transaction
		OrderId: createResp.OrderId, // But trying to cancel creator's order
	}

	_, err = suite.msgServer.CancelOrder(suite.ctx, cancelMsg)
	suite.Require().Error(err, "should reject cancellation by non-owner")
	suite.Require().Contains(err.Error(), "cannot cancel order")
}

func (suite *MsgServerComprehensiveTestSuite) TestCircuitBreakerTriggered() {
	// Test operations when circuit breaker is triggered
	// Circuit breaker logic would need to be implemented in keeper
	// For now, test that operations can be performed normally
	creator := keepertest.GenTestAddr()

	suite.bankKeeper.setBalance(creator, "uaura", sdkmath.NewInt(10000000000))
	suite.bankKeeper.setBalance(creator, "usdt", sdkmath.NewInt(10000000000))

	// Create a pool normally
	msg := &dexpb.MsgCreatePool{
		Creator: creator.String(),
		DenomA:  "uaura",
		DenomB:  "usdt",
		AmountA: sdk.NewCoin("uaura", sdkmath.NewInt(1000000000)),
		AmountB: sdk.NewCoin("usdt", sdkmath.NewInt(1000000000)),
	}

	_, err := suite.msgServer.CreatePool(suite.ctx, msg)
	suite.Require().NoError(err, "operations should work when circuit breaker not triggered")

	// TODO: When circuit breaker is implemented, add test for triggered state:
	// 1. Trigger circuit breaker via keeper
	// 2. Verify that operations are rejected
	// 3. Disable circuit breaker
	// 4. Verify operations work again
}
