// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	v1beta1 "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

// ValidateGenesis performs comprehensive validation for the DEX genesis state
// to prevent data corruption and invalid state on chain initialization.
func ValidateGenesis(gen *v1beta1.GenesisState) error {
	if gen == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}

	// Validate parameters
	if err := validateParams(&gen.Params); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate liquidity pools
	poolIDs := make(map[string]bool)
	for i, pool := range gen.LiquidityPools {
		if err := validateLiquidityPool(&pool, i); err != nil {
			return fmt.Errorf("invalid liquidity pool at index %d: %w", i, err)
		}
		// Check for duplicate pool IDs
		if poolIDs[pool.PoolId] {
			return fmt.Errorf("duplicate pool ID %s at index %d", pool.PoolId, i)
		}
		poolIDs[pool.PoolId] = true
	}

	// Validate swap orders
	orderIDs := make(map[string]bool)
	for i, order := range gen.SwapOrders {
		if err := validateSwapOrder(&order, i, poolIDs); err != nil {
			return fmt.Errorf("invalid swap order at index %d: %w", i, err)
		}
		// Check for duplicate order IDs
		if order.OrderId != "" {
			if orderIDs[order.OrderId] {
				return fmt.Errorf("duplicate order ID %s at index %d", order.OrderId, i)
			}
			orderIDs[order.OrderId] = true
		}
	}

	// Validate orderbooks
	for i, book := range gen.Orderbooks {
		if err := validateOrderbook(&book, i, poolIDs); err != nil {
			return fmt.Errorf("invalid orderbook at index %d: %w", i, err)
		}
	}

	// Validate market prices
	for i, price := range gen.MarketPrices {
		if err := validateMarketPrice(&price, i, poolIDs); err != nil {
			return fmt.Errorf("invalid market price at index %d: %w", i, err)
		}
	}

	// Validate swap stats
	for i, stats := range gen.SwapStats {
		if err := validateSwapStats(&stats, i, poolIDs); err != nil {
			return fmt.Errorf("invalid swap stats at index %d: %w", i, err)
		}
	}

	// Validate order commitments
	commitmentHashes := make(map[string]bool)
	for i, commitment := range gen.OrderCommitments {
		if err := validateOrderCommitment(&commitment, i); err != nil {
			return fmt.Errorf("invalid order commitment at index %d: %w", i, err)
		}
		// Check for duplicate commitment hashes
		commitHash := string(commitment.CommitHash)
		if commitmentHashes[commitHash] {
			return fmt.Errorf("duplicate commitment hash at index %d", i)
		}
		commitmentHashes[commitHash] = true
	}

	// Validate queued orders
	for i, queuedOrder := range gen.QueuedOrders {
		if err := validateQueuedOrder(&queuedOrder, i, poolIDs); err != nil {
			return fmt.Errorf("invalid queued order at index %d: %w", i, err)
		}
	}

	// Validate pool creation records
	for i, record := range gen.PoolCreationRecords {
		if err := validatePoolCreationRecord(&record, i); err != nil {
			return fmt.Errorf("invalid pool creation record at index %d: %w", i, err)
		}
	}

	return nil
}

// validateParams validates DEX module parameters
func validateParams(params *v1beta1.Params) error {
	if params.TradingFee.IsNil() {
		return fmt.Errorf("trading fee cannot be nil")
	}
	if params.TradingFee.IsNegative() {
		return fmt.Errorf("trading fee cannot be negative")
	}
	if params.TradingFee.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("trading fee cannot exceed 100%%")
	}

	if params.ProtocolFee.IsNil() {
		return fmt.Errorf("protocol fee cannot be nil")
	}
	if params.ProtocolFee.IsNegative() {
		return fmt.Errorf("protocol fee cannot be negative")
	}
	if params.ProtocolFee.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("protocol fee cannot exceed 100%%")
	}

	if params.MaxSlippageBps == 0 {
		return fmt.Errorf("max slippage must be greater than zero")
	}
	if params.MaxSlippageBps > 10000 {
		return fmt.Errorf("max slippage basis points cannot exceed 10000 (100%%)")
	}

	if params.MinSwapAmount.IsNil() {
		return fmt.Errorf("min swap amount cannot be nil")
	}
	if params.MinSwapAmount.IsNegative() {
		return fmt.Errorf("min swap amount cannot be negative")
	}

	return nil
}

// validateLiquidityPool validates a single liquidity pool with AMM invariant checks
func validateLiquidityPool(pool *v1beta1.LiquidityPool, index int) error {
	if pool.PoolId == "" {
		return fmt.Errorf("pool ID cannot be empty")
	}
	if pool.DenomA == "" {
		return fmt.Errorf("denom A cannot be empty")
	}
	if pool.DenomB == "" {
		return fmt.Errorf("denom B cannot be empty")
	}
	if pool.DenomA == pool.DenomB {
		return fmt.Errorf("denom A and denom B must be different")
	}

	// Validate reserves are positive (critical for AMM)
	if pool.ReserveA.IsNil() || pool.ReserveA.LTE(sdkmath.ZeroInt()) {
		return fmt.Errorf("reserve A must be positive, got %s", pool.ReserveA.String())
	}
	if pool.ReserveB.IsNil() || pool.ReserveB.LTE(sdkmath.ZeroInt()) {
		return fmt.Errorf("reserve B must be positive, got %s", pool.ReserveB.String())
	}

	// Validate k = x * y invariant for AMM pools
	// The product of reserves must be valid (no overflow, reasonable values)
	k := pool.ReserveA.Mul(pool.ReserveB)
	if k.IsNil() || k.LTE(sdkmath.ZeroInt()) {
		return fmt.Errorf("AMM invariant k = x * y must be positive")
	}

	// Validate total LP tokens
	if pool.TotalLpTokens.IsNil() || pool.TotalLpTokens.IsNegative() {
		return fmt.Errorf("total LP tokens cannot be nil or negative")
	}

	// Validate fee percentages
	if pool.FeePercentage.IsNil() || pool.FeePercentage.IsNegative() {
		return fmt.Errorf("fee percentage cannot be nil or negative")
	}
	if pool.FeePercentage.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("fee percentage cannot exceed 100%%")
	}

	if pool.ProtocolFeePercentage.IsNil() || pool.ProtocolFeePercentage.IsNegative() {
		return fmt.Errorf("protocol fee percentage cannot be nil or negative")
	}
	if pool.ProtocolFeePercentage.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("protocol fee percentage cannot exceed 100%%")
	}

	// Validate volume and fees (must not be negative)
	if pool.TotalVolume.IsNil() || pool.TotalVolume.IsNegative() {
		return fmt.Errorf("total volume cannot be nil or negative")
	}
	if pool.TotalFeesCollected.IsNil() || pool.TotalFeesCollected.IsNegative() {
		return fmt.Errorf("total fees collected cannot be nil or negative")
	}
	if pool.ProtocolFeeBalance.IsNil() || pool.ProtocolFeeBalance.IsNegative() {
		return fmt.Errorf("protocol fee balance cannot be nil or negative")
	}
	if pool.LockedLiquidity.IsNil() || pool.LockedLiquidity.IsNegative() {
		return fmt.Errorf("locked liquidity cannot be nil or negative")
	}

	// Validate liquidity providers
	providerAddrs := make(map[string]bool)
	totalLPTokens := sdkmath.ZeroInt()
	for j, provider := range pool.Providers {
		if provider.Address == "" {
			return fmt.Errorf("provider address cannot be empty at index %d", j)
		}
		if providerAddrs[provider.Address] {
			return fmt.Errorf("duplicate provider address %s at index %d", provider.Address, j)
		}
		providerAddrs[provider.Address] = true

		if provider.LpTokens.IsNil() || provider.LpTokens.LTE(sdkmath.ZeroInt()) {
			return fmt.Errorf("provider LP tokens must be positive at index %d", j)
		}
		totalLPTokens = totalLPTokens.Add(provider.LpTokens)
	}

	// Verify that sum of provider LP tokens + locked liquidity matches total
	// Invariant: TotalLpTokens = Sum(Provider LP Tokens) + LockedLiquidity
	expectedTotal := totalLPTokens.Add(pool.LockedLiquidity)
	if len(pool.Providers) > 0 && !expectedTotal.Equal(pool.TotalLpTokens) {
		return fmt.Errorf("sum of provider LP tokens (%s) + locked liquidity (%s) does not match total LP tokens (%s)",
			totalLPTokens.String(), pool.LockedLiquidity.String(), pool.TotalLpTokens.String())
	}

	return nil
}

// validateSwapOrder validates a swap order
func validateSwapOrder(order *v1beta1.SwapOrder, index int, poolIDs map[string]bool) error {
	if order.OrderId == "" {
		return fmt.Errorf("order ID cannot be empty")
	}
	// Note: SwapOrder uses OtherCoin field, not PoolId
	// Validation updated to match actual protobuf structure
	if order.UserAddress == "" {
		return fmt.Errorf("user address cannot be empty")
	}
	if order.AuraAmount.IsNil() || order.AuraAmount.LTE(sdkmath.ZeroInt()) {
		return fmt.Errorf("AURA amount must be positive")
	}
	if order.OtherAmount.IsNil() || order.OtherAmount.IsNegative() {
		return fmt.Errorf("other amount cannot be nil or negative")
	}

	return nil
}

// validateOrderbook validates an orderbook
func validateOrderbook(book *v1beta1.Orderbook, index int, poolIDs map[string]bool) error {
	// Note: Orderbook uses Pair field, not PoolId
	// Validation updated to match actual protobuf structure
	if book.Pair == "" {
		return fmt.Errorf("trading pair cannot be empty")
	}

	// Validate buy orders
	for j, order := range book.BuyOrders {
		if err := validateSwapOrder(&order, j, poolIDs); err != nil {
			return fmt.Errorf("invalid buy order at index %d: %w", j, err)
		}
	}

	// Validate sell orders
	for j, order := range book.SellOrders {
		if err := validateSwapOrder(&order, j, poolIDs); err != nil {
			return fmt.Errorf("invalid sell order at index %d: %w", j, err)
		}
	}

	return nil
}

// validateMarketPrice validates a market price entry
func validateMarketPrice(price *v1beta1.MarketPrice, index int, poolIDs map[string]bool) error {
	if price.Coin == "" {
		return fmt.Errorf("coin symbol cannot be empty")
	}
	if price.PriceUsd.IsNil() || price.PriceUsd.LTE(sdkmath.LegacyZeroDec()) {
		return fmt.Errorf("price USD must be positive")
	}
	if price.PriceAura.IsNil() || price.PriceAura.LTE(sdkmath.LegacyZeroDec()) {
		return fmt.Errorf("price AURA must be positive")
	}

	return nil
}

// validateSwapStats validates swap statistics
func validateSwapStats(stats *v1beta1.SwapStats, index int, poolIDs map[string]bool) error {
	if stats.PoolId == "" {
		return fmt.Errorf("pool ID cannot be empty")
	}
	if !poolIDs[stats.PoolId] {
		return fmt.Errorf("pool ID %s does not exist in liquidity pools", stats.PoolId)
	}
	if stats.AmountIn.IsNil() || stats.AmountIn.LTE(sdkmath.ZeroInt()) {
		return fmt.Errorf("amount in must be positive")
	}
	if stats.AmountOut.IsNil() || stats.AmountOut.LTE(sdkmath.ZeroInt()) {
		return fmt.Errorf("amount out must be positive")
	}
	if stats.EffectivePrice.IsNil() || stats.EffectivePrice.LTE(sdkmath.LegacyZeroDec()) {
		return fmt.Errorf("effective price must be positive")
	}

	return nil
}

// validateOrderCommitment validates an order commitment
func validateOrderCommitment(commitment *v1beta1.OrderCommitment, index int) error {
	if commitment.CommitId == "" {
		return fmt.Errorf("commit ID cannot be empty")
	}
	if len(commitment.CommitHash) == 0 {
		return fmt.Errorf("commit hash cannot be empty")
	}
	if commitment.Sender == "" {
		return fmt.Errorf("sender address cannot be empty")
	}

	return nil
}

// validateQueuedOrder validates a queued order
func validateQueuedOrder(queuedOrder *v1beta1.QueuedOrder, index int, poolIDs map[string]bool) error {
	if err := validateSwapOrder(&queuedOrder.Order, index, poolIDs); err != nil {
		return fmt.Errorf("invalid order in queued order: %w", err)
	}
	if len(queuedOrder.Salt) == 0 {
		return fmt.Errorf("salt cannot be empty")
	}

	return nil
}

// validatePoolCreationRecord validates a pool creation record
func validatePoolCreationRecord(record *v1beta1.PoolCreationRecord, index int) error {
	if len(record.PoolIds) == 0 {
		return fmt.Errorf("pool IDs list cannot be empty")
	}
	if record.Creator == "" {
		return fmt.Errorf("creator address cannot be empty")
	}

	return nil
}
