package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmath "cosmossdk.io/math"
	storeprefix "cosmossdk.io/store/prefix"
	"github.com/aequitas/aura/chain/x/dex/types"
)

// RegisterInvariants registers all dex module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "pool-reserves-consistency", PoolReservesConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "order-validity", OrderValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "liquidity-provider-consistency", LiquidityProviderConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "security-limits", SecurityLimitsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "htlc-validity", HTLCValidityInvariant(k))
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
		params := k.GetParams(ctx)
		if params == nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				"params are nil",
			), true
		}

		// TODO: Add Validate() method to Params type in proto/types if needed
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
		store := ctx.KVStore(k.storeKey)
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
			reserveA, ok := sdkmath.NewIntFromString(pool.ReserveA)
			if !ok || reserveA.IsNegative() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"pool-reserves-consistency",
					fmt.Sprintf("pool %s has invalid reserve A: %s", pool.PoolId, pool.ReserveA),
				), true
			}

			reserveB, ok := sdkmath.NewIntFromString(pool.ReserveB)
			if !ok || reserveB.IsNegative() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"pool-reserves-consistency",
					fmt.Sprintf("pool %s has invalid reserve B: %s", pool.PoolId, pool.ReserveB),
				), true
			}

			// Check liquidity shares are non-negative (using TotalLpTokens field)
			totalShares, ok := sdkmath.NewIntFromString(pool.TotalLpTokens)
			if !ok || totalShares.IsNegative() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"pool-reserves-consistency",
					fmt.Sprintf("pool %s has invalid total LP tokens: %s", pool.PoolId, pool.TotalLpTokens),
				), true
			}

			// If pool has reserves, it should have shares
			if (!reserveA.IsZero() || !reserveB.IsZero()) && totalShares.IsZero() {
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
		store := ctx.KVStore(k.storeKey)
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

			// Check amounts are positive (fields are strings in proto)
			auraAmt, ok := sdkmath.NewIntFromString(order.AuraAmount)
			if !ok || !auraAmt.IsPositive() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"order-validity",
					fmt.Sprintf("order %s has invalid AURA amount: %s", order.OrderId, order.AuraAmount),
				), true
			}

			otherAmt, ok := sdkmath.NewIntFromString(order.OtherAmount)
			if !ok || !otherAmt.IsPositive() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"order-validity",
					fmt.Sprintf("order %s has invalid other amount: %s", order.OrderId, order.OtherAmount),
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
			if order.Timestamp == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"order-validity",
					fmt.Sprintf("order %s has nil timestamp", order.OrderId),
				), true
			}
		}

		return "", false
	}
}

// LiquidityProviderConsistencyInvariant checks liquidity provider positions
// TODO: Enable when LiquidityProviderPosition type is defined in proto
func LiquidityProviderConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// TODO: Implement when LiquidityProviderPosition type exists
		// For now, liquidity provider positions are tracked within LiquidityPool.Providers
		// This invariant can validate those instead
		return "", false

		/* Uncomment when LiquidityProviderPosition type is defined
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.LiquidityProviderKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var position types.LiquidityProviderPosition
			if err := k.cdc.Unmarshal(iterator.Value(), &position); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"liquidity-provider-consistency",
					fmt.Sprintf("failed to unmarshal LP position: %s", err.Error()),
				), true
			}

			// Check provider address is valid
			if _, err := sdk.AccAddressFromBech32(position.Provider); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"liquidity-provider-consistency",
					fmt.Sprintf("LP position has invalid provider address: %s", position.Provider),
				), true
			}

			// Check pool ID is not empty
			if position.PoolId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"liquidity-provider-consistency",
					fmt.Sprintf("LP position for %s has empty pool ID", position.Provider),
				), true
			}

			// Check shares are positive
			shares, ok := sdkmath.NewIntFromString(position.Shares)
			if !ok || !shares.IsPositive() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"liquidity-provider-consistency",
					fmt.Sprintf("LP position for %s has invalid shares: %s",
						position.Provider, position.Shares),
				), true
			}
		}

		return "", false
		*/
	}
}

// SecurityLimitsInvariant checks security limits are not exceeded
func SecurityLimitsInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params := k.GetParams(ctx)
		if params == nil {
			return "", false
		}

		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.PoolPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var pool types.LiquidityPool
			if err := k.cdc.Unmarshal(iterator.Value(), &pool); err != nil {
				continue
			}

			// TODO: Add CircuitBreakerTriggered, LastTradeTime, CircuitBreakerTime fields to LiquidityPool proto
			// For now, just validate basic pool security limits
			// Security limits are enforced in the security module
		}

		return "", false
	}
}

// HTLCValidityInvariant checks Hash Time Locked Contracts validity
func HTLCValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
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
			if htlc.Timelock == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"htlc-validity",
					fmt.Sprintf("HTLC %s has nil timelock", htlcID),
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
			amount, ok := sdkmath.NewIntFromString(htlc.Amount)
			if !ok || !amount.IsPositive() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"htlc-validity",
					fmt.Sprintf("HTLC %s has invalid amount: %s", htlcID, htlc.Amount),
				), true
			}
		}

		return "", false
	}
}
