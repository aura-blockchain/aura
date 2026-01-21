// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/dex/types"
	v1beta1 "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

// validParams returns valid params for testing
func validParams() v1beta1.Params {
	return v1beta1.Params{
		TradingFee:     sdkmath.LegacyMustNewDecFromStr("0.003"),
		ProtocolFee:    sdkmath.LegacyMustNewDecFromStr("0.0005"),
		MaxSlippageBps: 500,
		MinSwapAmount:  sdkmath.NewInt(1000000),
	}
}

// validPool returns a valid liquidity pool for testing
func validPool(poolID string) v1beta1.LiquidityPool {
	return v1beta1.LiquidityPool{
		PoolId:                poolID,
		DenomA:                "uaura",
		DenomB:                "uusdt",
		ReserveA:              sdkmath.NewInt(1000000000),
		ReserveB:              sdkmath.NewInt(1000000000),
		TotalLpTokens:         sdkmath.NewInt(1000000000),
		FeePercentage:         sdkmath.LegacyMustNewDecFromStr("0.003"),
		ProtocolFeePercentage: sdkmath.LegacyMustNewDecFromStr("0.0005"),
		TotalVolume:           sdkmath.ZeroInt(),
		TotalFeesCollected:    sdkmath.ZeroInt(),
		ProtocolFeeBalance:    sdkmath.ZeroInt(),
		LockedLiquidity:       sdkmath.ZeroInt(),
	}
}

// validSwapOrder returns a valid swap order for testing
func validSwapOrder(orderID string) v1beta1.SwapOrder {
	return v1beta1.SwapOrder{
		OrderId:     orderID,
		UserAddress: "aura1abc123xyz",
		AuraAmount:  sdkmath.NewInt(1000000),
		OtherAmount: sdkmath.NewInt(1000000),
		OtherCoin:   "uusdt",
		Timestamp:   time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
}

// TestValidateGenesis_NilState tests that nil genesis state returns error
func TestValidateGenesis_NilState(t *testing.T) {
	err := types.ValidateGenesis(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "genesis state cannot be nil")
}

// TestValidateGenesis_DefaultGenesis tests that default genesis is valid
func TestValidateGenesis_DefaultGenesis(t *testing.T) {
	genesis := types.DefaultGenesis()
	err := types.ValidateGenesis(genesis)
	require.NoError(t, err)
}

// TestValidateGenesis_InvalidParams tests genesis validation with invalid params
func TestValidateGenesis_InvalidParams(t *testing.T) {
	tests := []struct {
		name        string
		modifyParam func(*v1beta1.Params)
		errContains string
	}{
		{
			name: "negative trading fee",
			modifyParam: func(p *v1beta1.Params) {
				p.TradingFee = sdkmath.LegacyMustNewDecFromStr("-0.001")
			},
			errContains: "trading fee cannot be negative",
		},
		{
			name: "trading fee exceeds 100%",
			modifyParam: func(p *v1beta1.Params) {
				p.TradingFee = sdkmath.LegacyMustNewDecFromStr("1.5")
			},
			errContains: "trading fee cannot exceed 100%",
		},
		{
			name: "negative protocol fee",
			modifyParam: func(p *v1beta1.Params) {
				p.ProtocolFee = sdkmath.LegacyMustNewDecFromStr("-0.001")
			},
			errContains: "protocol fee cannot be negative",
		},
		{
			name: "protocol fee exceeds 100%",
			modifyParam: func(p *v1beta1.Params) {
				p.ProtocolFee = sdkmath.LegacyMustNewDecFromStr("2.0")
			},
			errContains: "protocol fee cannot exceed 100%",
		},
		{
			name: "zero max slippage",
			modifyParam: func(p *v1beta1.Params) {
				p.MaxSlippageBps = 0
			},
			errContains: "max slippage must be greater than zero",
		},
		{
			name: "max slippage exceeds 10000 bps",
			modifyParam: func(p *v1beta1.Params) {
				p.MaxSlippageBps = 15000
			},
			errContains: "max slippage basis points cannot exceed 10000",
		},
		{
			name: "negative min swap amount",
			modifyParam: func(p *v1beta1.Params) {
				p.MinSwapAmount = sdkmath.NewInt(-100)
			},
			errContains: "min swap amount cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := types.DefaultGenesis()
			tt.modifyParam(&genesis.Params)
			err := types.ValidateGenesis(genesis)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// TestValidateGenesis_ValidLiquidityPools tests valid pool configurations
func TestValidateGenesis_ValidLiquidityPools(t *testing.T) {
	genesis := types.DefaultGenesis()
	genesis.LiquidityPools = []v1beta1.LiquidityPool{
		validPool("pool1"),
		validPool("pool2"),
	}
	err := types.ValidateGenesis(genesis)
	require.NoError(t, err)
}

// TestValidateGenesis_InvalidLiquidityPool tests various invalid pool configurations
func TestValidateGenesis_InvalidLiquidityPool(t *testing.T) {
	tests := []struct {
		name        string
		modifyPool  func(*v1beta1.LiquidityPool)
		errContains string
	}{
		{
			name: "empty pool ID",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.PoolId = ""
			},
			errContains: "pool ID cannot be empty",
		},
		{
			name: "empty denom A",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.DenomA = ""
			},
			errContains: "denom A cannot be empty",
		},
		{
			name: "empty denom B",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.DenomB = ""
			},
			errContains: "denom B cannot be empty",
		},
		{
			name: "same denom A and B",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.DenomA = "uaura"
				p.DenomB = "uaura"
			},
			errContains: "denom A and denom B must be different",
		},
		{
			name: "zero reserve A",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.ReserveA = sdkmath.ZeroInt()
			},
			errContains: "reserve A must be positive",
		},
		{
			name: "negative reserve A",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.ReserveA = sdkmath.NewInt(-100)
			},
			errContains: "reserve A must be positive",
		},
		{
			name: "zero reserve B",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.ReserveB = sdkmath.ZeroInt()
			},
			errContains: "reserve B must be positive",
		},
		{
			name: "negative total LP tokens",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.TotalLpTokens = sdkmath.NewInt(-100)
			},
			errContains: "total LP tokens cannot be nil or negative",
		},
		{
			name: "negative fee percentage",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.FeePercentage = sdkmath.LegacyMustNewDecFromStr("-0.01")
			},
			errContains: "fee percentage cannot be nil or negative",
		},
		{
			name: "fee percentage exceeds 100%",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.FeePercentage = sdkmath.LegacyMustNewDecFromStr("1.5")
			},
			errContains: "fee percentage cannot exceed 100%",
		},
		{
			name: "negative protocol fee percentage",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.ProtocolFeePercentage = sdkmath.LegacyMustNewDecFromStr("-0.01")
			},
			errContains: "protocol fee percentage cannot be nil or negative",
		},
		{
			name: "protocol fee percentage exceeds 100%",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.ProtocolFeePercentage = sdkmath.LegacyMustNewDecFromStr("1.5")
			},
			errContains: "protocol fee percentage cannot exceed 100%",
		},
		{
			name: "negative total volume",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.TotalVolume = sdkmath.NewInt(-100)
			},
			errContains: "total volume cannot be nil or negative",
		},
		{
			name: "negative total fees collected",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.TotalFeesCollected = sdkmath.NewInt(-100)
			},
			errContains: "total fees collected cannot be nil or negative",
		},
		{
			name: "negative protocol fee balance",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.ProtocolFeeBalance = sdkmath.NewInt(-100)
			},
			errContains: "protocol fee balance cannot be nil or negative",
		},
		{
			name: "negative locked liquidity",
			modifyPool: func(p *v1beta1.LiquidityPool) {
				p.LockedLiquidity = sdkmath.NewInt(-100)
			},
			errContains: "locked liquidity cannot be nil or negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := types.DefaultGenesis()
			pool := validPool("pool1")
			tt.modifyPool(&pool)
			genesis.LiquidityPools = []v1beta1.LiquidityPool{pool}
			err := types.ValidateGenesis(genesis)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// TestValidateGenesis_DuplicatePoolIDs tests duplicate pool ID detection
func TestValidateGenesis_DuplicatePoolIDs(t *testing.T) {
	genesis := types.DefaultGenesis()
	genesis.LiquidityPools = []v1beta1.LiquidityPool{
		validPool("same-pool-id"),
		validPool("same-pool-id"),
	}
	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate pool ID")
}

// TestValidateGenesis_ProviderValidation tests liquidity provider validation
func TestValidateGenesis_ProviderValidation(t *testing.T) {
	tests := []struct {
		name        string
		providers   []v1beta1.LiquidityProvider
		errContains string
	}{
		{
			name: "empty provider address",
			providers: []v1beta1.LiquidityProvider{
				{Address: "", LpTokens: sdkmath.NewInt(1000)},
			},
			errContains: "provider address cannot be empty",
		},
		{
			name: "zero provider LP tokens",
			providers: []v1beta1.LiquidityProvider{
				{Address: "aura1abc", LpTokens: sdkmath.ZeroInt()},
			},
			errContains: "provider LP tokens must be positive",
		},
		{
			name: "negative provider LP tokens",
			providers: []v1beta1.LiquidityProvider{
				{Address: "aura1abc", LpTokens: sdkmath.NewInt(-100)},
			},
			errContains: "provider LP tokens must be positive",
		},
		{
			name: "duplicate provider addresses",
			providers: []v1beta1.LiquidityProvider{
				{Address: "aura1same", LpTokens: sdkmath.NewInt(1000)},
				{Address: "aura1same", LpTokens: sdkmath.NewInt(2000)},
			},
			errContains: "duplicate provider address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := types.DefaultGenesis()
			pool := validPool("pool1")
			pool.Providers = tt.providers
			genesis.LiquidityPools = []v1beta1.LiquidityPool{pool}
			err := types.ValidateGenesis(genesis)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// TestValidateGenesis_ProviderLPTokensMismatch tests LP tokens sum validation
func TestValidateGenesis_ProviderLPTokensMismatch(t *testing.T) {
	genesis := types.DefaultGenesis()
	pool := validPool("pool1")
	pool.TotalLpTokens = sdkmath.NewInt(3000)
	pool.LockedLiquidity = sdkmath.ZeroInt()
	pool.Providers = []v1beta1.LiquidityProvider{
		{Address: "aura1abc", LpTokens: sdkmath.NewInt(1000)},
		{Address: "aura1def", LpTokens: sdkmath.NewInt(1000)},
	}
	genesis.LiquidityPools = []v1beta1.LiquidityPool{pool}
	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not match total LP tokens")
}

// TestValidateGenesis_ValidSwapOrders tests valid swap order configurations
func TestValidateGenesis_ValidSwapOrders(t *testing.T) {
	genesis := types.DefaultGenesis()
	genesis.SwapOrders = []v1beta1.SwapOrder{
		validSwapOrder("order1"),
		validSwapOrder("order2"),
	}
	err := types.ValidateGenesis(genesis)
	require.NoError(t, err)
}

// TestValidateGenesis_InvalidSwapOrder tests invalid swap order configurations
func TestValidateGenesis_InvalidSwapOrder(t *testing.T) {
	tests := []struct {
		name        string
		modifyOrder func(*v1beta1.SwapOrder)
		errContains string
	}{
		{
			name: "empty order ID",
			modifyOrder: func(o *v1beta1.SwapOrder) {
				o.OrderId = ""
			},
			errContains: "order ID cannot be empty",
		},
		{
			name: "empty user address",
			modifyOrder: func(o *v1beta1.SwapOrder) {
				o.UserAddress = ""
			},
			errContains: "user address cannot be empty",
		},
		{
			name: "zero aura amount",
			modifyOrder: func(o *v1beta1.SwapOrder) {
				o.AuraAmount = sdkmath.ZeroInt()
			},
			errContains: "AURA amount must be positive",
		},
		{
			name: "negative aura amount",
			modifyOrder: func(o *v1beta1.SwapOrder) {
				o.AuraAmount = sdkmath.NewInt(-100)
			},
			errContains: "AURA amount must be positive",
		},
		{
			name: "negative other amount",
			modifyOrder: func(o *v1beta1.SwapOrder) {
				o.OtherAmount = sdkmath.NewInt(-100)
			},
			errContains: "other amount cannot be nil or negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := types.DefaultGenesis()
			order := validSwapOrder("order1")
			tt.modifyOrder(&order)
			genesis.SwapOrders = []v1beta1.SwapOrder{order}
			err := types.ValidateGenesis(genesis)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// TestValidateGenesis_DuplicateOrderIDs tests duplicate order ID detection
func TestValidateGenesis_DuplicateOrderIDs(t *testing.T) {
	genesis := types.DefaultGenesis()
	genesis.SwapOrders = []v1beta1.SwapOrder{
		validSwapOrder("same-order-id"),
		validSwapOrder("same-order-id"),
	}
	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate order ID")
}

// TestValidateGenesis_Orderbook tests orderbook validation
func TestValidateGenesis_Orderbook(t *testing.T) {
	tests := []struct {
		name        string
		orderbook   v1beta1.Orderbook
		errContains string
	}{
		{
			name: "empty pair",
			orderbook: v1beta1.Orderbook{
				Pair:          "",
				BuyOrders:     []v1beta1.SwapOrder{},
				SellOrders:    []v1beta1.SwapOrder{},
				BestBid:       sdkmath.LegacyZeroDec(),
				BestAsk:       sdkmath.LegacyZeroDec(),
				SpreadPercent: sdkmath.LegacyZeroDec(),
			},
			errContains: "trading pair cannot be empty",
		},
		{
			name: "invalid buy order in orderbook",
			orderbook: v1beta1.Orderbook{
				Pair: "AURA/USDT",
				BuyOrders: []v1beta1.SwapOrder{
					{
						OrderId:     "",
						UserAddress: "aura1abc",
						AuraAmount:  sdkmath.NewInt(1000),
						OtherAmount: sdkmath.NewInt(1000),
						Timestamp:   time.Now(),
						ExpiresAt:   time.Now().Add(24 * time.Hour),
					},
				},
				SellOrders:    []v1beta1.SwapOrder{},
				BestBid:       sdkmath.LegacyZeroDec(),
				BestAsk:       sdkmath.LegacyZeroDec(),
				SpreadPercent: sdkmath.LegacyZeroDec(),
			},
			errContains: "invalid buy order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := types.DefaultGenesis()
			genesis.Orderbooks = []v1beta1.Orderbook{tt.orderbook}
			err := types.ValidateGenesis(genesis)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// TestValidateGenesis_MarketPrice tests market price validation
func TestValidateGenesis_MarketPrice(t *testing.T) {
	tests := []struct {
		name        string
		price       v1beta1.MarketPrice
		errContains string
	}{
		{
			name: "empty coin",
			price: v1beta1.MarketPrice{
				Coin:      "",
				PriceUsd:  sdkmath.LegacyOneDec(),
				PriceAura: sdkmath.LegacyOneDec(),
				UpdatedAt: time.Now(),
			},
			errContains: "coin symbol cannot be empty",
		},
		{
			name: "zero price USD",
			price: v1beta1.MarketPrice{
				Coin:      "uusdt",
				PriceUsd:  sdkmath.LegacyZeroDec(),
				PriceAura: sdkmath.LegacyOneDec(),
				UpdatedAt: time.Now(),
			},
			errContains: "price USD must be positive",
		},
		{
			name: "negative price USD",
			price: v1beta1.MarketPrice{
				Coin:      "uusdt",
				PriceUsd:  sdkmath.LegacyMustNewDecFromStr("-1.0"),
				PriceAura: sdkmath.LegacyOneDec(),
				UpdatedAt: time.Now(),
			},
			errContains: "price USD must be positive",
		},
		{
			name: "zero price AURA",
			price: v1beta1.MarketPrice{
				Coin:      "uusdt",
				PriceUsd:  sdkmath.LegacyOneDec(),
				PriceAura: sdkmath.LegacyZeroDec(),
				UpdatedAt: time.Now(),
			},
			errContains: "price AURA must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := types.DefaultGenesis()
			genesis.MarketPrices = []v1beta1.MarketPrice{tt.price}
			err := types.ValidateGenesis(genesis)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// TestValidateGenesis_SwapStats tests swap stats validation
func TestValidateGenesis_SwapStats(t *testing.T) {
	genesis := types.DefaultGenesis()
	pool := validPool("pool1")
	genesis.LiquidityPools = []v1beta1.LiquidityPool{pool}

	tests := []struct {
		name        string
		stats       v1beta1.SwapStats
		errContains string
	}{
		{
			name: "empty pool ID",
			stats: v1beta1.SwapStats{
				PoolId:         "",
				AmountIn:       sdkmath.NewInt(1000),
				AmountOut:      sdkmath.NewInt(1000),
				EffectivePrice: sdkmath.LegacyOneDec(),
				Timestamp:      time.Now(),
			},
			errContains: "pool ID cannot be empty",
		},
		{
			name: "pool ID not in pools",
			stats: v1beta1.SwapStats{
				PoolId:         "nonexistent-pool",
				AmountIn:       sdkmath.NewInt(1000),
				AmountOut:      sdkmath.NewInt(1000),
				EffectivePrice: sdkmath.LegacyOneDec(),
				Timestamp:      time.Now(),
			},
			errContains: "pool ID nonexistent-pool does not exist",
		},
		{
			name: "zero amount in",
			stats: v1beta1.SwapStats{
				PoolId:         "pool1",
				AmountIn:       sdkmath.ZeroInt(),
				AmountOut:      sdkmath.NewInt(1000),
				EffectivePrice: sdkmath.LegacyOneDec(),
				Timestamp:      time.Now(),
			},
			errContains: "amount in must be positive",
		},
		{
			name: "zero amount out",
			stats: v1beta1.SwapStats{
				PoolId:         "pool1",
				AmountIn:       sdkmath.NewInt(1000),
				AmountOut:      sdkmath.ZeroInt(),
				EffectivePrice: sdkmath.LegacyOneDec(),
				Timestamp:      time.Now(),
			},
			errContains: "amount out must be positive",
		},
		{
			name: "zero effective price",
			stats: v1beta1.SwapStats{
				PoolId:         "pool1",
				AmountIn:       sdkmath.NewInt(1000),
				AmountOut:      sdkmath.NewInt(1000),
				EffectivePrice: sdkmath.LegacyZeroDec(),
				Timestamp:      time.Now(),
			},
			errContains: "effective price must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testGenesis := types.DefaultGenesis()
			testGenesis.LiquidityPools = genesis.LiquidityPools
			testGenesis.SwapStats = []v1beta1.SwapStats{tt.stats}
			err := types.ValidateGenesis(testGenesis)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// TestValidateGenesis_OrderCommitment tests order commitment validation
func TestValidateGenesis_OrderCommitment(t *testing.T) {
	tests := []struct {
		name        string
		commitment  v1beta1.OrderCommitment
		errContains string
	}{
		{
			name: "empty commit ID",
			commitment: v1beta1.OrderCommitment{
				CommitId:       "",
				CommitHash:     []byte("hash"),
				Sender:         "aura1abc",
				CommittedAt:    time.Now(),
				RevealDeadline: time.Now().Add(time.Hour),
			},
			errContains: "commit ID cannot be empty",
		},
		{
			name: "empty commit hash",
			commitment: v1beta1.OrderCommitment{
				CommitId:       "commit1",
				CommitHash:     []byte{},
				Sender:         "aura1abc",
				CommittedAt:    time.Now(),
				RevealDeadline: time.Now().Add(time.Hour),
			},
			errContains: "commit hash cannot be empty",
		},
		{
			name: "empty sender",
			commitment: v1beta1.OrderCommitment{
				CommitId:       "commit1",
				CommitHash:     []byte("hash"),
				Sender:         "",
				CommittedAt:    time.Now(),
				RevealDeadline: time.Now().Add(time.Hour),
			},
			errContains: "sender address cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := types.DefaultGenesis()
			genesis.OrderCommitments = []v1beta1.OrderCommitment{tt.commitment}
			err := types.ValidateGenesis(genesis)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// TestValidateGenesis_DuplicateCommitHash tests duplicate commitment hash detection
func TestValidateGenesis_DuplicateCommitHash(t *testing.T) {
	genesis := types.DefaultGenesis()
	hash := []byte("same-hash")
	genesis.OrderCommitments = []v1beta1.OrderCommitment{
		{
			CommitId:       "commit1",
			CommitHash:     hash,
			Sender:         "aura1abc",
			CommittedAt:    time.Now(),
			RevealDeadline: time.Now().Add(time.Hour),
		},
		{
			CommitId:       "commit2",
			CommitHash:     hash,
			Sender:         "aura1def",
			CommittedAt:    time.Now(),
			RevealDeadline: time.Now().Add(time.Hour),
		},
	}
	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate commitment hash")
}

// TestValidateGenesis_QueuedOrder tests queued order validation
func TestValidateGenesis_QueuedOrder(t *testing.T) {
	tests := []struct {
		name        string
		queuedOrder v1beta1.QueuedOrder
		errContains string
	}{
		{
			name: "empty salt",
			queuedOrder: v1beta1.QueuedOrder{
				Order:    validSwapOrder("order1"),
				Salt:     []byte{},
				QueuedAt: time.Now(),
			},
			errContains: "salt cannot be empty",
		},
		{
			name: "invalid order in queued order",
			queuedOrder: v1beta1.QueuedOrder{
				Order: v1beta1.SwapOrder{
					OrderId:     "",
					UserAddress: "aura1abc",
					AuraAmount:  sdkmath.NewInt(1000),
					OtherAmount: sdkmath.NewInt(1000),
					Timestamp:   time.Now(),
					ExpiresAt:   time.Now().Add(24 * time.Hour),
				},
				Salt:     []byte("salt"),
				QueuedAt: time.Now(),
			},
			errContains: "invalid order in queued order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := types.DefaultGenesis()
			genesis.QueuedOrders = []v1beta1.QueuedOrder{tt.queuedOrder}
			err := types.ValidateGenesis(genesis)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// TestValidateGenesis_PoolCreationRecord tests pool creation record validation
func TestValidateGenesis_PoolCreationRecord(t *testing.T) {
	tests := []struct {
		name        string
		record      v1beta1.PoolCreationRecord
		errContains string
	}{
		{
			name: "empty pool IDs",
			record: v1beta1.PoolCreationRecord{
				Creator:          "aura1abc",
				PoolIds:          []string{},
				LastCreationTime: time.Now(),
			},
			errContains: "pool IDs list cannot be empty",
		},
		{
			name: "empty creator",
			record: v1beta1.PoolCreationRecord{
				Creator:          "",
				PoolIds:          []string{"pool1"},
				LastCreationTime: time.Now(),
			},
			errContains: "creator address cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := types.DefaultGenesis()
			genesis.PoolCreationRecords = []v1beta1.PoolCreationRecord{tt.record}
			err := types.ValidateGenesis(genesis)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// TestValidateGenesis_CompleteValidState tests a complete valid genesis state
func TestValidateGenesis_CompleteValidState(t *testing.T) {
	genesis := types.DefaultGenesis()

	// Add valid pool
	pool := validPool("pool1")
	pool.Providers = []v1beta1.LiquidityProvider{
		{Address: "aura1abc", LpTokens: sdkmath.NewInt(500000000)},
		{Address: "aura1def", LpTokens: sdkmath.NewInt(500000000)},
	}
	pool.TotalLpTokens = sdkmath.NewInt(1000000000)
	pool.LockedLiquidity = sdkmath.ZeroInt()
	genesis.LiquidityPools = []v1beta1.LiquidityPool{pool}

	// Add valid orders
	genesis.SwapOrders = []v1beta1.SwapOrder{
		validSwapOrder("order1"),
		validSwapOrder("order2"),
	}

	// Add valid orderbook
	genesis.Orderbooks = []v1beta1.Orderbook{
		{
			Pair:          "AURA/USDT",
			BuyOrders:     []v1beta1.SwapOrder{validSwapOrder("buy1")},
			SellOrders:    []v1beta1.SwapOrder{validSwapOrder("sell1")},
			BestBid:       sdkmath.LegacyOneDec(),
			BestAsk:       sdkmath.LegacyOneDec(),
			SpreadPercent: sdkmath.LegacyZeroDec(),
		},
	}

	// Add valid market price
	genesis.MarketPrices = []v1beta1.MarketPrice{
		{
			Coin:       "uusdt",
			PriceUsd:   sdkmath.LegacyOneDec(),
			PriceAura:  sdkmath.LegacyOneDec(),
			UpdatedAt:  time.Now(),
			SampleSize: 10,
		},
	}

	// Add valid swap stats
	genesis.SwapStats = []v1beta1.SwapStats{
		{
			PoolId:         "pool1",
			AmountIn:       sdkmath.NewInt(1000),
			AmountOut:      sdkmath.NewInt(1000),
			EffectivePrice: sdkmath.LegacyOneDec(),
			Timestamp:      time.Now(),
		},
	}

	// Add valid order commitments
	genesis.OrderCommitments = []v1beta1.OrderCommitment{
		{
			CommitId:       "commit1",
			CommitHash:     []byte("hash1"),
			Sender:         "aura1abc",
			CommittedAt:    time.Now(),
			RevealDeadline: time.Now().Add(time.Hour),
		},
	}

	// Add valid queued orders
	genesis.QueuedOrders = []v1beta1.QueuedOrder{
		{
			Order:    validSwapOrder("queued1"),
			Salt:     []byte("salt1"),
			QueuedAt: time.Now(),
		},
	}

	// Add valid pool creation records
	genesis.PoolCreationRecords = []v1beta1.PoolCreationRecord{
		{
			Creator:          "aura1abc",
			PoolIds:          []string{"pool1"},
			LastCreationTime: time.Now(),
		},
	}

	err := types.ValidateGenesis(genesis)
	require.NoError(t, err)
}
