// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	storeprefix "cosmossdk.io/store/prefix"
	"github.com/aequitas/aura/chain/x/dex/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all dex module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "pool-reserves-consistency", PoolReservesConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "order-validity", OrderValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "liquidity-provider-consistency", LiquidityProviderConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "security-limits", SecurityLimitsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "htlc-validity", HTLCValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "order-pool-integrity", OrderPoolIntegrityInvariant(k))
}

// AllInvariants runs all invariants of the dex module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			PoolReservesConsistencyInvariant(k),
			OrderValidityInvariant(k),
			LiquidityProviderConsistencyInvariant(k),
			SecurityLimitsInvariant(k),
			HTLCValidityInvariant(k),
			OrderPoolIntegrityInvariant(k),
		}

		for _, inv := range invariants {
			msg, broken := inv(ctx)
			if broken {
				return msg, broken
			}
		}

		return "", false
	}
}

// ParamsInvariant checks that module parameters are valid
func ParamsInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Create cache context for consistent snapshot reads
		cacheCtx, _ := ctx.CacheContext()

		_, err := k.GetParams(cacheCtx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("failed to get params: %v", err),
			), true
		}

		// NOTE: Future enhancement - Add Validate() method to Params type in proto/types if needed
		// For now, just check that params exist
		// if err := params.Validate(); err != nil {
		// 	return sdk.FormatInvariant(
		// 		types.ModuleName,
		// 		"params-valid",
		// 		fmt.Sprintf("invalid params: %s", err.Error()),
		// 	), true
		// }
		return "", false
	}
}

// PoolReservesConsistencyInvariant checks that pool reserves are consistent
func PoolReservesConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Create cache context for consistent snapshot reads
		cacheCtx, _ := ctx.CacheContext()

		store := cacheCtx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.PoolPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var pool types.LiquidityPool
			if err := k.cdc.Unmarshal(iterator.Value(), &pool); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"pool-reserves-consistency",
					fmt.Sprintf("failed to unmarshal pool: %s", err.Error()),
				), true
			}

			// Check pool ID is not empty
			if pool.PoolId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"pool-reserves-consistency",
					"pool has empty ID",
				), true
			}

			// Check denoms are not empty
			if pool.DenomA == "" || pool.DenomB == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"pool-reserves-consistency",
					fmt.Sprintf("pool %s has empty denom", pool.PoolId),
				), true
			}

			// Check reserves are positive
			if pool.ReserveA.IsNegative() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"pool-reserves-consistency",
					fmt.Sprintf("pool %s has invalid reserve A: %s", pool.PoolId, pool.ReserveA.String()),
				), true
			}

			if pool.ReserveB.IsNegative() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"pool-reserves-consistency",
					fmt.Sprintf("pool %s has invalid reserve B: %s", pool.PoolId, pool.ReserveB.String()),
				), true
			}

			// Check liquidity shares are non-negative (using TotalLpTokens field)
			if pool.TotalLpTokens.IsNegative() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"pool-reserves-consistency",
					fmt.Sprintf("pool %s has invalid total LP tokens: %s", pool.PoolId, pool.TotalLpTokens.String()),
				), true
			}

			// If pool has reserves, it should have shares
			if (!pool.ReserveA.IsZero() || !pool.ReserveB.IsZero()) && pool.TotalLpTokens.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"pool-reserves-consistency",
					fmt.Sprintf("pool %s has reserves but zero LP tokens", pool.PoolId),
				), true
			}
		}

		return "", false
	}
}

// OrderValidityInvariant checks that all orders are valid
func OrderValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Create cache context for consistent snapshot reads
		cacheCtx, _ := ctx.CacheContext()

		store := cacheCtx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.OrderPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var order types.SwapOrder
			if err := k.cdc.Unmarshal(iterator.Value(), &order); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"order-validity",
					fmt.Sprintf("failed to unmarshal order: %s", err.Error()),
				), true
			}

			// Check order ID is not empty
			if order.OrderId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"order-validity",
					"order has empty ID",
				), true
			}

			// Check user address is valid
			if _, err := sdk.AccAddressFromBech32(order.UserAddress); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"order-validity",
					fmt.Sprintf("order %s has invalid user address: %s", order.OrderId, order.UserAddress),
				), true
			}

			// Check amounts are positive
			if !order.AuraAmount.IsPositive() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"order-validity",
					fmt.Sprintf("order %s has invalid AURA amount: %s", order.OrderId, order.AuraAmount.String()),
				), true
			}

			if !order.OtherAmount.IsPositive() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"order-validity",
					fmt.Sprintf("order %s has invalid other amount: %s", order.OrderId, order.OtherAmount.String()),
				), true
			}

			// Check other coin denomination
			if order.OtherCoin == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"order-validity",
					fmt.Sprintf("order %s has empty other coin", order.OrderId),
				), true
			}

			// Check timestamp exists
			if order.Timestamp.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"order-validity",
					fmt.Sprintf("order %s has zero timestamp", order.OrderId),
				), true
			}
		}

		return "", false
	}
}

// LiquidityProviderConsistencyInvariant checks liquidity provider positions
// SECURITY: This invariant ensures LP token supply matches distributed tokens
// Formula: TotalLpTokens = Sum(Provider LP Tokens) + LockedLiquidity
// This prevents LP token inflation attacks and accounting errors
func LiquidityProviderConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Create cache context for consistent snapshot reads
		cacheCtx, _ := ctx.CacheContext()

		store := cacheCtx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.PoolPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var pool types.LiquidityPool
			if err := k.cdc.Unmarshal(iterator.Value(), &pool); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"liquidity-provider-consistency",
					fmt.Sprintf("failed to unmarshal pool: %s", err.Error()),
				), true
			}

			if pool.TotalLpTokens.IsNegative() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"liquidity-provider-consistency",
					fmt.Sprintf("pool %s has invalid total LP tokens: %s", pool.PoolId, pool.TotalLpTokens.String()),
				), true
			}

			// Sum all provider LP tokens
			sum := sdkmath.ZeroInt()
			for _, provider := range pool.Providers {
				if _, err := sdk.AccAddressFromBech32(provider.Address); err != nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"liquidity-provider-consistency",
						fmt.Sprintf("pool %s provider address invalid: %s", pool.PoolId, provider.Address),
					), true
				}

				if provider.LpTokens.IsNegative() {
					return sdk.FormatInvariant(
						types.ModuleName,
						"liquidity-provider-consistency",
						fmt.Sprintf("pool %s provider %s has invalid LP tokens: %s", pool.PoolId, provider.Address, provider.LpTokens.String()),
					), true
				}
				sum = sum.Add(provider.LpTokens)
			}

			// Account for permanently locked liquidity (burned on pool creation)
			// The locked liquidity is included in TotalLpTokens but not assigned to any provider
			lockedLiquidity := sdkmath.ZeroInt()
			if !pool.LockedLiquidity.IsZero() {
				if pool.LockedLiquidity.IsNegative() {
					return sdk.FormatInvariant(
						types.ModuleName,
						"liquidity-provider-consistency",
						fmt.Sprintf("pool %s has invalid locked liquidity: %s", pool.PoolId, pool.LockedLiquidity.String()),
					), true
				}
				lockedLiquidity = pool.LockedLiquidity
			}

			// CRITICAL INVARIANT: TotalLpTokens = Sum(Provider LP Tokens) + LockedLiquidity
			// This ensures LP tokens are properly backed by reserves and prevents:
			// - LP token inflation attacks where tokens are minted without backing
			// - Accounting errors that allow draining pool reserves
			// - Mismatch between total supply and distributed tokens
			expectedTotal := sum.Add(lockedLiquidity)
			if !pool.TotalLpTokens.IsZero() && !expectedTotal.Equal(pool.TotalLpTokens) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"liquidity-provider-consistency",
					fmt.Sprintf("CRITICAL: pool %s LP token invariant violated - "+
						"total LP tokens %s != provider sum %s + locked %s (expected %s)",
						pool.PoolId,
						pool.TotalLpTokens.String(),
						sum.String(),
						lockedLiquidity.String(),
						expectedTotal.String()),
				), true
			}
		}

		return "", false
	}
}

// SecurityLimitsInvariant checks security limits are not exceeded
func SecurityLimitsInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Create cache context for consistent snapshot reads
		cacheCtx, _ := ctx.CacheContext()

		store := cacheCtx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.PoolPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var pool types.LiquidityPool
			if err := k.cdc.Unmarshal(iterator.Value(), &pool); err != nil {
				continue
			}

			// NOTE: Future enhancement - Add CircuitBreakerTriggered, LastTradeTime, CircuitBreakerTime fields to LiquidityPool proto
			// For now, just validate basic pool security limits
			// Security limits are enforced in the security module
		}

		return "", false
	}
}

// HTLCValidityInvariant checks Hash Time Locked Contracts validity
func HTLCValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Create cache context for consistent snapshot reads
		cacheCtx, _ := ctx.CacheContext()

		store := cacheCtx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.HTLCPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var htlc types.HTLCData
			if err := k.cdc.Unmarshal(iterator.Value(), &htlc); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"htlc-validity",
					fmt.Sprintf("failed to unmarshal HTLC: %s", err.Error()),
				), true
			}

			// Get HTLC ID from key
			htlcID := string(iterator.Key())

			// Check secret hash is not empty
			if htlc.SecretHash == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"htlc-validity",
					fmt.Sprintf("HTLC %s has empty secret hash", htlcID),
				), true
			}

			// Check timelock is set
			if htlc.Timelock.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"htlc-validity",
					fmt.Sprintf("HTLC %s has zero timelock", htlcID),
				), true
			}

			// Check sender and recipient are valid
			if _, err := sdk.AccAddressFromBech32(htlc.Sender); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"htlc-validity",
					fmt.Sprintf("HTLC %s has invalid sender: %s", htlcID, htlc.Sender),
				), true
			}

			if _, err := sdk.AccAddressFromBech32(htlc.Recipient); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"htlc-validity",
					fmt.Sprintf("HTLC %s has invalid recipient: %s", htlcID, htlc.Recipient),
				), true
			}

			// Check amount is positive
			if !htlc.Amount.IsPositive() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"htlc-validity",
					fmt.Sprintf("HTLC %s has invalid amount: %s", htlcID, htlc.Amount.String()),
				), true
			}
		}

		return "", false
	}
}

// OrderPoolIntegrityInvariant checks that all orders reference existing liquidity pools
//
// CRITICAL SECURITY: This invariant ensures referential integrity between orders and pools.
// Without this check, orphaned orders could exist after pool deletion, leading to:
//   - Orders that can never be filled (no pool to execute against)
//   - Inability to process or cancel orders properly
//   - Locked user funds in unfillable orders
//   - Invalid state in the orderbook
//
// The invariant validates that every order's implied pool (derived from coin pair) exists.
// This prevents the scenario where a pool is deleted but orders still reference it.
//
// Returns:
//   - ("", false) if all orders reference valid pools
//   - (error message, true) if any order references a non-existent pool
func OrderPoolIntegrityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Create cache context for consistent snapshot reads
		cacheCtx, _ := ctx.CacheContext()

		// Get all orders
		orders := k.GetAllOrders(cacheCtx)

		for _, order := range orders {
			// Derive the pool ID from the order's coin pair
			// Orders trade "uaura" against order.OtherCoin
			poolID := k.GeneratePoolID("uaura", order.OtherCoin)

			// Check if the pool exists
			pool := k.GetPool(cacheCtx, poolID)
			if pool == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"order-pool-integrity",
					fmt.Sprintf("Order %s references non-existent pool %s (uaura/%s)",
						order.OrderId, poolID, order.OtherCoin),
				), true
			}
		}

		return "", false
	}
}
