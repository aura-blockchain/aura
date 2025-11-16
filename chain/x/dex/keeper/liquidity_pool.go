package keeper

import (
	"fmt"
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// ============================================================================
// AMM Liquidity Pools - Constant Product Formula (x * y = k)
// Adapted from liquidity_pools.py (541 lines)
// ============================================================================

// CreatePool creates a new AMM liquidity pool
func (k Keeper) CreatePool(
	ctx sdk.Context,
	creator string,
	denomA string,
	denomB string,
	amountA sdk.Coin,
	amountB sdk.Coin,
) (*types.LiquidityPool, sdkmath.Int, error) {
	// Generate pool ID
	poolID := k.GeneratePoolID(denomA, denomB)

	// Check if pool already exists
	if k.GetPool(ctx, poolID) != nil {
		return nil, sdk.ZeroInt(), sdkerrors.Wrapf(
			types.ErrPoolAlreadyExists,
			"pool %s already exists",
			poolID,
		)
	}

	// Validate amounts
	if amountA.Amount.IsZero() || amountB.Amount.IsZero() {
		return nil, sdk.ZeroInt(), sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			"amounts must be positive",
		)
	}

	// Check minimum liquidity for NEW pool creators
	if err := k.CheckMinimumLiquidity(ctx, creator, poolID, amountA.Amount); err != nil {
		return nil, sdk.ZeroInt(), err
	}

	// Transfer tokens from creator to module
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	if err != nil {
		return nil, sdk.ZeroInt(), err
	}

	coins := sdk.NewCoins(amountA, amountB)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		creatorAddr,
		types.ModuleName,
		coins,
	); err != nil {
		return nil, sdk.ZeroInt(), err
	}

	// Calculate initial LP tokens using geometric mean (from Python line 78)
	// lp_tokens = (xai_amount * other_amount) ** 0.5
	lpTokens := sdk.NewIntFromBigInt(new(big.Int).Sqrt(
		amountA.Amount.Mul(amountB.Amount).BigInt(),
	))

	// Get parameters
	params := k.GetParams(ctx)

	// Create pool
	pool := &types.LiquidityPool{
		PoolId:                poolID,
		DenomA:                denomA,
		DenomB:                denomB,
		ReserveA:              amountA.Amount,
		ReserveB:              amountB.Amount,
		TotalLpTokens:         lpTokens,
		FeePercentage:         params.TradingFee,
		ProtocolFeePercentage: params.ProtocolFee,
		TotalVolume:           sdk.ZeroInt(),
		TotalFeesCollected:    sdk.ZeroInt(),
		SwapCount:             0,
		ProtocolFeeBalance:    sdk.ZeroInt(),
		Providers: []*types.LiquidityProvider{
			{
				Address:  creator,
				LpTokens: lpTokens,
			},
		},
		CreatedAt: ctx.BlockTime(),
	}

	// Store pool
	k.SetPool(ctx, pool)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreatePool,
			sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
			sdk.NewAttribute(types.AttributeKeyCreator, creator),
			sdk.NewAttribute(types.AttributeKeyLPTokens, lpTokens.String()),
		),
	)

	return pool, lpTokens, nil
}

// AddLiquidity adds liquidity to existing pool
// Adapted from liquidity_pools.py add_liquidity() (lines 62-143)
func (k Keeper) AddLiquidity(
	ctx sdk.Context,
	provider string,
	poolID string,
	amountA sdk.Coin,
	amountB sdk.Coin,
) (sdkmath.Int, sdk.Dec, error) {
	// Get pool
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdkerrors.Wrapf(
			types.ErrPoolNotFound,
			"pool %s not found",
			poolID,
		)
	}

	// Check minimum liquidity (only for NEW providers, existing are grandfathered!)
	if err := k.CheckMinimumLiquidity(ctx, provider, poolID, amountA.Amount); err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), err
	}

	// Validate amounts
	if amountA.Amount.IsZero() || amountB.Amount.IsZero() {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			"amounts must be positive",
		)
	}

	// Calculate required ratio (from Python line 106-113)
	// current_ratio = self.other_reserve / self.xai_reserve
	// required_other = xai_amount * current_ratio
	currentRatio := pool.ReserveB.ToDec().Quo(pool.ReserveA.ToDec())
	requiredAmountB := amountA.Amount.ToDec().Mul(currentRatio)

	// Adjust amounts to match ratio
	actualAmountA := amountA.Amount
	actualAmountB := amountB.Amount

	if amountB.Amount.ToDec().GT(requiredAmountB) {
		// Have too much B, use less
		actualAmountB = requiredAmountB.TruncateInt()
	} else {
		// Have too little B, adjust A down
		actualAmountA = amountB.Amount.ToDec().Quo(currentRatio).TruncateInt()
	}

	// Calculate LP tokens proportional to contribution (Python line 116)
	// lp_tokens = (xai_amount / self.xai_reserve) * self.total_liquidity_tokens
	lpTokens := actualAmountA.ToDec().
		Quo(pool.ReserveA.ToDec()).
		Mul(pool.TotalLpTokens.ToDec()).
		TruncateInt()

	// Transfer tokens from provider
	providerAddr, err := sdk.AccAddressFromBech32(provider)
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), err
	}

	coins := sdk.NewCoins(
		sdk.NewCoin(pool.DenomA, actualAmountA),
		sdk.NewCoin(pool.DenomB, actualAmountB),
	)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		providerAddr,
		types.ModuleName,
		coins,
	); err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), err
	}

	// Update reserves
	pool.ReserveA = pool.ReserveA.Add(actualAmountA)
	pool.ReserveB = pool.ReserveB.Add(actualAmountB)
	pool.TotalLpTokens = pool.TotalLpTokens.Add(lpTokens)

	// Update or add provider
	found := false
	for i, p := range pool.Providers {
		if p.Address == provider {
			pool.Providers[i].LpTokens = p.LpTokens.Add(lpTokens)
			found = true
			break
		}
	}
	if !found {
		pool.Providers = append(pool.Providers, &types.LiquidityProvider{
			Address:  provider,
			LpTokens: lpTokens,
		})
	}

	// Calculate pool share
	poolShare := lpTokens.ToDec().Quo(pool.TotalLpTokens.ToDec()).MulInt64(100)

	// Save pool
	k.SetPool(ctx, pool)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAddLiquidity,
			sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
			sdk.NewAttribute(types.AttributeKeyProvider, provider),
			sdk.NewAttribute(types.AttributeKeyLPTokens, lpTokens.String()),
			sdk.NewAttribute(types.AttributeKeyPoolShare, poolShare.String()),
		),
	)

	return lpTokens, poolShare, nil
}

// RemoveLiquidity removes liquidity from pool
// Adapted from liquidity_pools.py remove_liquidity() (lines 145-193)
func (k Keeper) RemoveLiquidity(
	ctx sdk.Context,
	provider string,
	poolID string,
	lpTokens sdkmath.Int,
) (sdk.Coin, sdk.Coin, error) {
	// Get pool
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return sdk.Coin{}, sdk.Coin{}, sdkerrors.Wrapf(
			types.ErrPoolNotFound,
			"pool %s not found",
			poolID,
		)
	}

	// Find provider
	var providerLP *types.LiquidityProvider
	var providerIndex int
	for i, p := range pool.Providers {
		if p.Address == provider {
			providerLP = p
			providerIndex = i
			break
		}
	}

	if providerLP == nil {
		return sdk.Coin{}, sdk.Coin{}, sdkerrors.Wrap(
			types.ErrNotLiquidityProvider,
			"not a liquidity provider",
		)
	}

	// Check sufficient LP tokens
	if lpTokens.GT(providerLP.LpTokens) {
		return sdk.Coin{}, sdk.Coin{}, sdkerrors.Wrapf(
			types.ErrInsufficientLPTokens,
			"insufficient LP tokens: have %s, requested %s",
			providerLP.LpTokens.String(),
			lpTokens.String(),
		)
	}

	// Calculate share of pool (Python line 170)
	// share = lp_tokens / self.total_liquidity_tokens
	share := lpTokens.ToDec().Quo(pool.TotalLpTokens.ToDec())

	// Calculate amounts to return (Python line 172-173)
	// xai_amount = share * self.xai_reserve
	// other_amount = share * self.other_reserve
	amountA := share.MulInt(pool.ReserveA).TruncateInt()
	amountB := share.MulInt(pool.ReserveB).TruncateInt()

	// Update reserves
	pool.ReserveA = pool.ReserveA.Sub(amountA)
	pool.ReserveB = pool.ReserveB.Sub(amountB)
	pool.TotalLpTokens = pool.TotalLpTokens.Sub(lpTokens)

	// Update provider
	pool.Providers[providerIndex].LpTokens = providerLP.LpTokens.Sub(lpTokens)

	// Remove provider if they have no LP tokens left
	if pool.Providers[providerIndex].LpTokens.IsZero() {
		pool.Providers = append(
			pool.Providers[:providerIndex],
			pool.Providers[providerIndex+1:]...,
		)
	}

	// Save pool
	k.SetPool(ctx, pool)

	// Transfer tokens back to provider
	providerAddr, err := sdk.AccAddressFromBech32(provider)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	coinA := sdk.NewCoin(pool.DenomA, amountA)
	coinB := sdk.NewCoin(pool.DenomB, amountB)
	coins := sdk.NewCoins(coinA, coinB)

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		providerAddr,
		coins,
	); err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRemoveLiquidity,
			sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
			sdk.NewAttribute(types.AttributeKeyProvider, provider),
			sdk.NewAttribute(types.AttributeKeyLPTokens, lpTokens.String()),
		),
	)

	return coinA, coinB, nil
}

// SwapExactIn executes swap with exact input amount
// Adapted from liquidity_pools.py swap_xai_for_other() (lines 195-268)
func (k Keeper) SwapExactIn(
	ctx sdk.Context,
	sender string,
	poolID string,
	coinIn sdk.Coin,
	minAmountOut sdkmath.Int,
	maxSlippageBps uint64,
) (sdkmath.Int, sdk.Dec, sdk.Dec, error) {
	// Get pool
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), sdkerrors.Wrapf(
			types.ErrPoolNotFound,
			"pool %s not found",
			poolID,
		)
	}

	// Determine which way we're swapping
	var (
		reserveIn  sdkmath.Int
		reserveOut sdkmath.Int
		denomOut   string
		isAtoB     bool
	)

	if coinIn.Denom == pool.DenomA {
		reserveIn = pool.ReserveA
		reserveOut = pool.ReserveB
		denomOut = pool.DenomB
		isAtoB = true
	} else if coinIn.Denom == pool.DenomB {
		reserveIn = pool.ReserveB
		reserveOut = pool.ReserveA
		denomOut = pool.DenomA
		isAtoB = false
	} else {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			"coin denom not in pool",
		)
	}

	// Apply fees (Python line 215)
	// xai_after_fee = xai_amount * (1 - self.fee_percentage - self.protocol_fee_percentage)
	totalFeeRate := pool.FeePercentage.Add(pool.ProtocolFeePercentage)
	amountAfterFee := coinIn.Amount.ToDec().
		Mul(sdk.OneDec().Sub(totalFeeRate)).
		TruncateInt()

	// Constant product formula (Python lines 217-224)
	// k = self.xai_reserve * self.other_reserve
	// new_xai_reserve = self.xai_reserve + xai_after_fee
	// new_other_reserve = k / new_xai_reserve
	// other_output = self.other_reserve - new_other_reserve
	k_constant := reserveIn.Mul(reserveOut)
	newReserveIn := reserveIn.Add(amountAfterFee)
	newReserveOut := k_constant.ToDec().QuoInt(newReserveIn).TruncateInt()
	amountOut := reserveOut.Sub(newReserveOut)

	// Check minimum output
	if amountOut.LT(minAmountOut) {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), sdkerrors.Wrapf(
			types.ErrSlippageTooHigh,
			"output amount %s less than minimum %s",
			amountOut.String(),
			minAmountOut.String(),
		)
	}

	// Calculate price impact (Python lines 228-230)
	priceBefore := reserveOut.ToDec().Quo(reserveIn.ToDec())
	priceAfter := newReserveOut.ToDec().Quo(newReserveIn.ToDec())
	priceImpact := priceBefore.Sub(priceAfter).Quo(priceBefore).Mul(sdk.NewDec(100))

	if priceImpact.IsNegative() {
		priceImpact = priceImpact.Neg()
	}

	// Check slippage limit (Python lines 232-237)
	maxSlippage := sdk.NewDec(int64(maxSlippageBps)).QuoInt64(10000)
	if priceImpact.GT(maxSlippage.Mul(sdk.NewDec(100))) {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), sdkerrors.Wrapf(
			types.ErrSlippageTooHigh,
			"price impact %s%% exceeds maximum %s%%",
			priceImpact.String(),
			maxSlippage.Mul(sdk.NewDec(100)).String(),
		)
	}

	// Calculate fees (Python lines 241-242)
	feeAmount := coinIn.Amount.ToDec().Mul(pool.FeePercentage).TruncateInt()
	protocolFee := coinIn.Amount.ToDec().Mul(pool.ProtocolFeePercentage).TruncateInt()

	// Update pool state
	if isAtoB {
		pool.ReserveA = newReserveIn
		pool.ReserveB = newReserveOut
	} else {
		pool.ReserveB = newReserveIn
		pool.ReserveA = newReserveOut
	}

	pool.TotalVolume = pool.TotalVolume.Add(coinIn.Amount)
	pool.TotalFeesCollected = pool.TotalFeesCollected.Add(feeAmount)
	pool.ProtocolFeeBalance = pool.ProtocolFeeBalance.Add(protocolFee)
	pool.SwapCount++

	k.SetPool(ctx, pool)

	// Transfer tokens
	senderAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), err
	}

	// Send input to module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		senderAddr,
		types.ModuleName,
		sdk.NewCoins(coinIn),
	); err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), err
	}

	// Send output to sender
	coinOut := sdk.NewCoin(denomOut, amountOut)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		senderAddr,
		sdk.NewCoins(coinOut),
	); err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), err
	}

	// Record swap stats for price tracking
	k.RecordSwapStats(ctx, poolID, coinIn.Amount, amountOut, time.Now())

	// Emit event
	effectivePrice := amountOut.ToDec().Quo(coinIn.Amount.ToDec())
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwap,
			sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
			sdk.NewAttribute(types.AttributeKeySender, sender),
			sdk.NewAttribute(types.AttributeKeyAmountIn, coinIn.String()),
			sdk.NewAttribute(types.AttributeKeyAmountOut, coinOut.String()),
			sdk.NewAttribute(types.AttributeKeyPriceImpact, priceImpact.String()),
		),
	)

	return amountOut, effectivePrice, priceImpact, nil
}

// GetQuote calculates swap output without executing
// Adapted from liquidity_pools.py get_quote() (lines 342-372)
func (k Keeper) GetQuote(
	ctx sdk.Context,
	poolID string,
	denomIn string,
	amountIn sdkmath.Int,
) (sdkmath.Int, sdk.Dec, sdk.Dec, sdkmath.Int, error) {
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroInt(),
			sdkerrors.Wrapf(types.ErrPoolNotFound, "pool %s not found", poolID)
	}

	// Same logic as SwapExactIn but without state changes
	var reserveIn, reserveOut sdkmath.Int

	if denomIn == pool.DenomA {
		reserveIn = pool.ReserveA
		reserveOut = pool.ReserveB
	} else {
		reserveIn = pool.ReserveB
		reserveOut = pool.ReserveA
	}

	totalFeeRate := pool.FeePercentage.Add(pool.ProtocolFeePercentage)
	amountAfterFee := amountIn.ToDec().Mul(sdk.OneDec().Sub(totalFeeRate)).TruncateInt()

	k_constant := reserveIn.Mul(reserveOut)
	newReserveIn := reserveIn.Add(amountAfterFee)
	newReserveOut := k_constant.ToDec().QuoInt(newReserveIn).TruncateInt()
	estimatedOutput := reserveOut.Sub(newReserveOut)

	effectivePrice := estimatedOutput.ToDec().Quo(amountIn.ToDec())

	priceBefore := reserveOut.ToDec().Quo(reserveIn.ToDec())
	priceAfter := newReserveOut.ToDec().Quo(newReserveIn.ToDec())
	priceImpact := priceBefore.Sub(priceAfter).Quo(priceBefore).Mul(sdk.NewDec(100))
	if priceImpact.IsNegative() {
		priceImpact = priceImpact.Neg()
	}

	feeAmount := amountIn.ToDec().Mul(pool.FeePercentage).TruncateInt()

	return estimatedOutput, effectivePrice, priceImpact, feeAmount, nil
}

// RecordSwapStats records swap for price tracking
func (k Keeper) RecordSwapStats(
	ctx sdk.Context,
	poolID string,
	amountIn sdkmath.Int,
	amountOut sdkmath.Int,
	timestamp time.Time,
) {
	// Store swap stats for market price tracking
	// This would be stored in a separate stats store
	// Implementation details depend on price oracle requirements
}
