package keeper

import (
	"math/big"
	"strings"
	"time"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	metrics "github.com/hashicorp/go-metrics"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// ============================================================================
// AMM Liquidity Pools - Constant Product Formula (x * y = k)
// Adapted from liquidity_pools.py (541 lines)
// ============================================================================

const (
	// MinimumLiquidity is the minimum liquidity burned on pool creation
	// to prevent first depositor inflation attacks.
	// This amount is permanently locked (not assigned to any provider)
	// to ensure LP token value cannot be manipulated via donation attacks.
	MinimumLiquidity = 1000
)

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
		return nil, sdkmath.ZeroInt(), errors.Wrapf(
			types.ErrPoolAlreadyExists,
			"pool %s already exists",
			poolID,
		)
	}

	// Validate amounts
	if amountA.Amount.IsZero() || amountB.Amount.IsZero() {
		return nil, sdkmath.ZeroInt(), errors.Wrap(
			types.ErrInvalidRequest,
			"amounts must be positive",
		)
	}

	// Check minimum liquidity for NEW pool creators
	if err := k.CheckMinimumLiquidity(ctx, creator, poolID, amountA.Amount); err != nil {
		return nil, sdkmath.ZeroInt(), err
	}

	// Transfer tokens from creator to module
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	if err != nil {
		return nil, sdkmath.ZeroInt(), err
	}

	coins := sdk.NewCoins(amountA, amountB)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		creatorAddr,
		types.ModuleName,
		coins,
	); err != nil {
		return nil, sdkmath.ZeroInt(), err
	}

	// Calculate initial LP tokens using geometric mean (from Python line 78)
	// lp_tokens = (xai_amount * other_amount) ** 0.5
	initialLpTokens := sdkmath.NewIntFromBigInt(new(big.Int).Sqrt(
		amountA.Amount.Mul(amountB.Amount).BigInt(),
	))

	// SECURITY: Burn minimum liquidity to prevent first depositor inflation attack
	// The first depositor attack allows manipulation of LP token value through donation:
	// 1. Attacker creates pool with 1 WEI of each token, receives 1 LP token
	// 2. Attacker donates large amounts directly to pool (not via AddLiquidity)
	// 3. Pool reserves are now imbalanced, but LP supply is still 1
	// 4. Victim adds liquidity, receives 0 LP tokens due to rounding
	// 5. Attacker withdraws, stealing victim's deposit
	//
	// By burning MinimumLiquidity (1000) LP tokens permanently, we ensure:
	// - LP token value stays reasonable (cannot be inflated to extreme values)
	// - Rounding errors are minimized for subsequent depositors
	// - Cost of attack becomes prohibitively expensive
	minimumLiquidity := sdkmath.NewInt(MinimumLiquidity)
	if initialLpTokens.LTE(minimumLiquidity) {
		return nil, sdkmath.ZeroInt(), errors.Wrapf(
			types.ErrInvalidRequest,
			"insufficient initial liquidity: calculated %s LP tokens, need > %d to burn minimum",
			initialLpTokens.String(),
			MinimumLiquidity,
		)
	}

	// Subtract minimum liquidity - this amount is locked forever
	lpTokens := initialLpTokens.Sub(minimumLiquidity)

	// Get parameters
	params := k.GetParams(ctx)

	// Create pool
	pool := &types.LiquidityPool{
		PoolId:                poolID,
		DenomA:                denomA,
		DenomB:                denomB,
		ReserveA:              amountA.Amount.String(),
		ReserveB:              amountB.Amount.String(),
		TotalLpTokens:         initialLpTokens.String(), // Total includes locked liquidity
		FeePercentage:         params.TradingFee,
		ProtocolFeePercentage: params.ProtocolFee,
		TotalVolume:           sdkmath.ZeroInt().String(),
		TotalFeesCollected:    sdkmath.ZeroInt().String(),
		SwapCount:             0,
		ProtocolFeeBalance:    sdkmath.ZeroInt().String(),
		LockedLiquidity:       minimumLiquidity.String(), // Permanently locked
		Providers: []*types.LiquidityProvider{
			{
				Address:  creator,
				LpTokens: lpTokens.String(), // Creator receives total minus locked
			},
		},
		CreatedAt: timestamppb.New(ctx.BlockTime()),
	}

	// Store pool
	k.SetPool(ctx, pool)

	// Emit events
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreatePool,
			sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
			sdk.NewAttribute(types.AttributeKeyCreator, creator),
			sdk.NewAttribute(types.AttributeKeyLPTokens, lpTokens.String()),
		),
		sdk.NewEvent(
			"minimum_liquidity_locked",
			sdk.NewAttribute("pool_id", poolID),
			sdk.NewAttribute("amount", minimumLiquidity.String()),
			sdk.NewAttribute("reason", "first_depositor_attack_prevention"),
		),
	})

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
) (sdkmath.Int, sdkmath.LegacyDec, error) {
	// Get pool
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), errors.Wrapf(
			types.ErrPoolNotFound,
			"pool %s not found",
			poolID,
		)
	}

	// Check minimum liquidity (only for NEW providers, existing are grandfathered!)
	if err := k.CheckMinimumLiquidity(ctx, provider, poolID, amountA.Amount); err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), err
	}

	// Validate amounts
	if amountA.Amount.IsZero() || amountB.Amount.IsZero() {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), errors.Wrap(
			types.ErrInvalidRequest,
			"amounts must be positive",
		)
	}

	// Calculate required ratio (from Python line 106-113)
	// current_ratio = self.other_reserve / self.xai_reserve
	// required_other = xai_amount * current_ratio
	reserveA, reserveB, err := k.getPoolReserves(pool)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), err
	}
	currentRatio := reserveB.ToLegacyDec().Quo(reserveA.ToLegacyDec())
	requiredAmountB := amountA.Amount.ToLegacyDec().Mul(currentRatio)

	// Adjust amounts to match ratio
	actualAmountA := amountA.Amount
	actualAmountB := amountB.Amount

	if amountB.Amount.ToLegacyDec().GT(requiredAmountB) {
		// Have too much B, use less
		actualAmountB = requiredAmountB.TruncateInt()
	} else {
		// Have too little B, adjust A down
		actualAmountA = amountB.Amount.ToLegacyDec().Quo(currentRatio).TruncateInt()
	}

	// Calculate LP tokens proportional to contribution (Python line 116)
	// lp_tokens = (xai_amount / self.xai_reserve) * self.total_liquidity_tokens
	totalLpTokens, err := k.parseLPTokens(pool.TotalLpTokens)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), err
	}
	lpTokens := actualAmountA.ToLegacyDec().
		Quo(reserveA.ToLegacyDec()).
		Mul(totalLpTokens.ToLegacyDec()).
		TruncateInt()

	// SECURITY: Prevent dust/rounding attacks by rejecting deposits that would receive 0 LP tokens
	// This protects against:
	// 1. Donation attacks that inflate pool reserves to cause rounding to zero
	// 2. Dust deposits that waste gas without providing liquidity
	// 3. Economic griefing where deposits are lost to rounding
	if lpTokens.IsZero() {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), errors.Wrapf(
			types.ErrInvalidRequest,
			"liquidity amount too small: would receive 0 LP tokens (amounts: %s %s, %s %s)",
			actualAmountA.String(), pool.DenomA,
			actualAmountB.String(), pool.DenomB,
		)
	}

	// Transfer tokens from provider
	providerAddr, err := sdk.AccAddressFromBech32(provider)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), err
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
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), err
	}

	// Update reserves
	pool.ReserveA = reserveA.Add(actualAmountA).String()
	pool.ReserveB = reserveB.Add(actualAmountB).String()
	pool.TotalLpTokens = totalLpTokens.Add(lpTokens).String()

	// Update or add provider
	found := false
	for i, p := range pool.Providers {
		if p.Address == provider {
			providerLpTokens, err := k.parseLPTokens(p.LpTokens)
			if err != nil {
				return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), err
			}
			pool.Providers[i].LpTokens = providerLpTokens.Add(lpTokens).String()
			found = true
			break
		}
	}
	if !found {
		pool.Providers = append(pool.Providers, &types.LiquidityProvider{
			Address:  provider,
			LpTokens: lpTokens.String(),
		})
	}

	// Calculate pool share
	newTotalLpTokens := totalLpTokens.Add(lpTokens)
	poolShare := lpTokens.ToLegacyDec().Quo(newTotalLpTokens.ToLegacyDec()).MulInt64(100)

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
		return sdk.Coin{}, sdk.Coin{}, errors.Wrapf(
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
		return sdk.Coin{}, sdk.Coin{}, errors.Wrap(
			types.ErrNotLiquidityProvider,
			"not a liquidity provider",
		)
	}

	// Parse provider LP tokens
	providerLpTokens, err := k.parseLPTokens(providerLP.LpTokens)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// Check sufficient LP tokens
	if lpTokens.GT(providerLpTokens) {
		return sdk.Coin{}, sdk.Coin{}, errors.Wrapf(
			types.ErrInsufficientLPTokens,
			"insufficient LP tokens: have %s, requested %s",
			providerLpTokens.String(),
			lpTokens.String(),
		)
	}

	// Parse pool reserves and LP tokens
	reserveA, reserveB, err := k.getPoolReserves(pool)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}
	totalLpTokens, err := k.parseLPTokens(pool.TotalLpTokens)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// Calculate share of pool (Python line 170)
	// share = lp_tokens / self.total_liquidity_tokens
	share := lpTokens.ToLegacyDec().Quo(totalLpTokens.ToLegacyDec())

	// Calculate amounts to return (Python line 172-173)
	// xai_amount = share * self.xai_reserve
	// other_amount = share * self.other_reserve
	amountA := share.MulInt(reserveA).TruncateInt()
	amountB := share.MulInt(reserveB).TruncateInt()

	// Update reserves
	pool.ReserveA = reserveA.Sub(amountA).String()
	pool.ReserveB = reserveB.Sub(amountB).String()
	pool.TotalLpTokens = totalLpTokens.Sub(lpTokens).String()

	// Update provider
	newProviderLpTokens := providerLpTokens.Sub(lpTokens)
	pool.Providers[providerIndex].LpTokens = newProviderLpTokens.String()

	// Remove provider if they have no LP tokens left
	if newProviderLpTokens.IsZero() {
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
) (sdkmath.Int, sdkmath.LegacyDec, sdkmath.LegacyDec, error) {
	// Validate non-zero input
	if coinIn.Amount.IsZero() {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), errors.Wrap(
			types.ErrInvalidRequest,
			"swap amount must be greater than zero",
		)
	}

	// Get pool
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), errors.Wrapf(
			types.ErrPoolNotFound,
			"pool %s not found",
			poolID,
		)
	}

	// Check circuit breaker
	if k.IsCircuitBreakerActive(ctx, poolID) {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), errors.Wrap(
			types.ErrCircuitBreakerActive,
			"circuit breaker: trading paused",
		)
	}

	// Check maximum trade size
	if err := k.CheckMaxTradeSize(ctx, poolID, coinIn.Amount); err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), err
	}

	// Parse reserves
	reserveA, reserveB, err := k.getPoolReserves(pool)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), err
	}

	// Determine which way we're swapping
	var (
		reserveIn  sdkmath.Int
		reserveOut sdkmath.Int
		denomOut   string
		isAtoB     bool
	)

	if coinIn.Denom == pool.DenomA {
		reserveIn = reserveA
		reserveOut = reserveB
		denomOut = pool.DenomB
		isAtoB = true
	} else if coinIn.Denom == pool.DenomB {
		reserveIn = reserveB
		reserveOut = reserveA
		denomOut = pool.DenomA
		isAtoB = false
	} else {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), errors.Wrap(
			types.ErrInvalidRequest,
			"coin denom not in pool",
		)
	}

	// Parse fee percentages
	feePercentage, err := k.parseFeePercentage(pool.FeePercentage)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), err
	}
	protocolFeePercentage, err := k.parseFeePercentage(pool.ProtocolFeePercentage)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), err
	}

	// Apply fees (Python line 215)
	// xai_after_fee = xai_amount * (1 - self.fee_percentage - self.protocol_fee_percentage)
	totalFeeRate := feePercentage.Add(protocolFeePercentage)
	amountAfterFee := coinIn.Amount.ToLegacyDec().
		Mul(sdkmath.LegacyOneDec().Sub(totalFeeRate)).
		TruncateInt()

	// Constant product formula (Python lines 217-224)
	// k = self.xai_reserve * self.other_reserve
	// new_xai_reserve = self.xai_reserve + xai_after_fee
	// new_other_reserve = k / new_xai_reserve
	// other_output = self.other_reserve - new_other_reserve
	k_constant := reserveIn.Mul(reserveOut)
	newReserveIn := reserveIn.Add(amountAfterFee)
	newReserveOut := k_constant.ToLegacyDec().QuoInt(newReserveIn).TruncateInt()
	amountOut := reserveOut.Sub(newReserveOut)

	// Check minimum output
	if amountOut.LT(minAmountOut) {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), errors.Wrapf(
			types.ErrSlippageTooHigh,
			"output amount %s less than minimum %s",
			amountOut.String(),
			minAmountOut.String(),
		)
	}

	// Calculate price impact (Python lines 228-230)
	priceBefore := reserveOut.ToLegacyDec().Quo(reserveIn.ToLegacyDec())
	priceAfter := newReserveOut.ToLegacyDec().Quo(newReserveIn.ToLegacyDec())
	priceImpact := priceBefore.Sub(priceAfter).Quo(priceBefore).Mul(sdkmath.LegacyNewDec(100))

	if priceImpact.IsNegative() {
		priceImpact = priceImpact.Neg()
	}

	// Check slippage limit (Python lines 232-237)
	// maxSlippageBps is in basis points (1 bp = 0.01%)
	// Convert to percentage: bps / 10000 * 100 = bps / 100
	maxSlippagePercent := sdkmath.LegacyNewDec(int64(maxSlippageBps)).QuoInt64(100)
	if maxSlippagePercent.GT(sdkmath.LegacyZeroDec()) && priceImpact.GT(maxSlippagePercent) {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), errors.Wrapf(
			types.ErrSlippageTooHigh,
			"price impact %s%% exceeds maximum %s%%",
			priceImpact.String(),
			maxSlippagePercent.String(),
		)
	}

	// Calculate fees (Python lines 241-242)
	feeAmount := coinIn.Amount.ToLegacyDec().Mul(feePercentage).TruncateInt()
	protocolFee := coinIn.Amount.ToLegacyDec().Mul(protocolFeePercentage).TruncateInt()

	// Parse existing volume and fees
	totalVolume, err := k.parseReserve(pool.TotalVolume)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), err
	}
	totalFeesCollected, err := k.parseReserve(pool.TotalFeesCollected)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), err
	}
	protocolFeeBalance, err := k.parseReserve(pool.ProtocolFeeBalance)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), err
	}

	// Update pool state
	if isAtoB {
		pool.ReserveA = newReserveIn.String()
		pool.ReserveB = newReserveOut.String()
	} else {
		pool.ReserveB = newReserveIn.String()
		pool.ReserveA = newReserveOut.String()
	}

	pool.TotalVolume = totalVolume.Add(coinIn.Amount).String()
	pool.TotalFeesCollected = totalFeesCollected.Add(feeAmount).String()
	pool.ProtocolFeeBalance = protocolFeeBalance.Add(protocolFee).String()
	pool.SwapCount++

	k.SetPool(ctx, pool)

	// Transfer tokens
	senderAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), err
	}

	// Send input to module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		senderAddr,
		types.ModuleName,
		sdk.NewCoins(coinIn),
	); err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), err
	}

	// Send output to sender
	coinOut := sdk.NewCoin(denomOut, amountOut)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		senderAddr,
		sdk.NewCoins(coinOut),
	); err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), err
	}

	// Record swap stats for price tracking
	k.RecordSwapStats(ctx, poolID, coinIn.Amount, amountOut, ctx.BlockTime())

	// Emit event
	effectivePrice := amountOut.ToLegacyDec().Quo(coinIn.Amount.ToLegacyDec())
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
) (sdkmath.Int, sdkmath.LegacyDec, sdkmath.LegacyDec, sdkmath.Int, error) {
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.ZeroInt(),
			errors.Wrapf(types.ErrPoolNotFound, "pool %s not found", poolID)
	}

	// Parse reserves
	reserveA, reserveB, err := k.getPoolReserves(pool)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.ZeroInt(), err
	}

	// Same logic as SwapExactIn but without state changes
	var reserveIn, reserveOut sdkmath.Int

	if denomIn == pool.DenomA {
		reserveIn = reserveA
		reserveOut = reserveB
	} else {
		reserveIn = reserveB
		reserveOut = reserveA
	}

	// Parse fee percentages
	feePercentage, err := k.parseFeePercentage(pool.FeePercentage)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.ZeroInt(), err
	}
	protocolFeePercentage, err := k.parseFeePercentage(pool.ProtocolFeePercentage)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.ZeroInt(), err
	}

	totalFeeRate := feePercentage.Add(protocolFeePercentage)
	amountAfterFee := amountIn.ToLegacyDec().Mul(sdkmath.LegacyOneDec().Sub(totalFeeRate)).TruncateInt()

	k_constant := reserveIn.Mul(reserveOut)
	newReserveIn := reserveIn.Add(amountAfterFee)
	newReserveOut := k_constant.ToLegacyDec().QuoInt(newReserveIn).TruncateInt()
	estimatedOutput := reserveOut.Sub(newReserveOut)

	effectivePrice := estimatedOutput.ToLegacyDec().Quo(amountIn.ToLegacyDec())

	priceBefore := reserveOut.ToLegacyDec().Quo(reserveIn.ToLegacyDec())
	priceAfter := newReserveOut.ToLegacyDec().Quo(newReserveIn.ToLegacyDec())
	priceImpact := priceBefore.Sub(priceAfter).Quo(priceBefore).Mul(sdkmath.LegacyNewDec(100))
	if priceImpact.IsNegative() {
		priceImpact = priceImpact.Neg()
	}

	feeAmount := amountIn.ToLegacyDec().Mul(feePercentage).TruncateInt()

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
	price := sdkmath.LegacyZeroDec()
	if amountIn.IsPositive() {
		price = amountOut.ToLegacyDec().Quo(amountIn.ToLegacyDec())
	}

	stats := &types.SwapStats{
		PoolId:         poolID,
		Timestamp:      timestamppb.New(timestamp),
		AmountIn:       amountIn.String(),
		AmountOut:      amountOut.String(),
		EffectivePrice: price.String(),
	}
	k.setSwapStats(ctx, stats)

	telemetry.IncrCounter(float32(1), "dex", "swap", "recorded")
	telemetry.SetGaugeWithLabels(
		[]string{"dex", "swap", "effective_price"},
		float32(price.MustFloat64()),
		[]metrics.Label{telemetry.NewLabel("pool", poolID)},
	)

	k.updateMarketPrice(ctx, poolID, price, timestamp)
}

func (k Keeper) updateMarketPrice(
	ctx sdk.Context,
	poolID string,
	effectivePrice sdkmath.LegacyDec,
	timestamp time.Time,
) {
	coin := strings.ToLower(k.coinFromPoolID(poolID))
	priceEntry, _ := k.GetMarketPrice(ctx, coin)
	if priceEntry == nil {
		priceEntry = &types.MarketPrice{Coin: coin}
	}

	priceEntry.PriceAura = effectivePrice.String()
	if isUSDStable(coin) {
		priceEntry.PriceUsd = sdkmath.LegacyOneDec().String()
	} else {
		priceEntry.PriceUsd = sdkmath.LegacyZeroDec().String()
	}
	priceEntry.UpdatedAt = timestamppb.New(timestamp)
	priceEntry.SampleSize++

	k.setMarketPrice(ctx, priceEntry)

	telemetry.SetGaugeWithLabels(
		[]string{"dex", "market_price", "price_aura"},
		float32(effectivePrice.MustFloat64()),
		[]metrics.Label{telemetry.NewLabel("coin", coin)},
	)
	telemetry.SetGaugeWithLabels(
		[]string{"dex", "market_price", "sample_size"},
		float32(priceEntry.SampleSize),
		[]metrics.Label{telemetry.NewLabel("coin", coin)},
	)
}

func (k Keeper) coinFromPoolID(poolID string) string {
	parts := strings.Split(poolID, "-")
	if len(parts) == 2 {
		if strings.EqualFold(parts[0], "uaura") {
			return parts[1]
		}
		if strings.EqualFold(parts[1], "uaura") {
			return parts[0]
		}
	}
	return poolID
}

func isUSDStable(denom string) bool {
	switch strings.ToLower(denom) {
	case "usdt", "usdc", "dai", "uusd", "usd":
		return true
	default:
		return false
	}
}
