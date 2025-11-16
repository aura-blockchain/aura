package keeper

import (
	"github.com/aequitas/aura/chain/x/dex/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SecureSwapExactIn is a wrapper around SwapExactIn with all security checks
func (k Keeper) SecureSwapExactIn(
	ctx sdk.Context,
	sender string,
	poolID string,
	coinIn sdk.Coin,
	minAmountOut sdk.Int,
	maxSlippageBps uint64,
) (sdk.Int, sdk.Dec, sdk.Dec, error) {
	// SECURITY CHECK 1: Circuit Breaker
	if k.IsCircuitBreakerActive(ctx, poolID) {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), types.ErrCircuitBreakerActive
	}

	// SECURITY CHECK 2: Front-Running Protection
	if err := k.CheckFrontRunningProtection(ctx, sender, poolID); err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), err
	}

	// SECURITY CHECK 3: MEV Protection
	if err := k.CheckMEVProtection(ctx, sender); err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), err
	}

	// SECURITY CHECK 4: Wash Trading Detection
	if err := k.DetectWashTrading(ctx, sender, poolID); err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), err
	}

	// SECURITY CHECK 5: Dust Attack Prevention
	if err := k.CheckDustAttack(ctx, coinIn.Amount); err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), err
	}

	// SECURITY CHECK 6: Maximum Trade Size
	if err := k.CheckMaxTradeSize(ctx, poolID, coinIn.Amount); err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), err
	}

	// Execute the swap
	amountOut, effectivePrice, priceImpact, err := k.SwapExactIn(
		ctx,
		sender,
		poolID,
		coinIn,
		minAmountOut,
		maxSlippageBps,
	)

	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), err
	}

	// SECURITY CHECK 7: Pool-Specific Slippage Limits
	if err := k.CheckPoolSlippageLimit(ctx, poolID, priceImpact); err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), err
	}

	// SECURITY CHECK 8: Price Impact Rejection Threshold
	if err := k.CheckPriceImpactThreshold(ctx, priceImpact); err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), sdk.ZeroDec(), err
	}

	// Record trade for security tracking
	k.RecordTradeBlock(ctx, sender, poolID)

	// Update TWAP oracle
	if err := k.RecordTWAPObservation(ctx, poolID); err != nil {
		// Log error but don't fail the trade
		ctx.Logger().Error("failed to record TWAP observation", "error", err)
	}

	return amountOut, effectivePrice, priceImpact, nil
}

// SecureCreatePool is a wrapper around CreatePool with all security checks
func (k Keeper) SecureCreatePool(
	ctx sdk.Context,
	creator string,
	denomA string,
	denomB string,
	amountA sdk.Coin,
	amountB sdk.Coin,
) (*types.LiquidityPool, sdk.Int, error) {
	// SECURITY CHECK 1: Circuit Breaker
	poolID := k.GeneratePoolID(denomA, denomB)
	if k.IsCircuitBreakerActive(ctx, poolID) {
		return nil, sdk.ZeroInt(), types.ErrCircuitBreakerActive
	}

	// SECURITY CHECK 2: Pool Creation Limits
	if err := k.CheckPoolCreationLimits(ctx, creator, amountA.Amount); err != nil {
		return nil, sdk.ZeroInt(), err
	}

	// Create the pool
	pool, lpTokens, err := k.CreatePool(ctx, creator, denomA, denomB, amountA, amountB)
	if err != nil {
		return nil, sdk.ZeroInt(), err
	}

	// Record pool creation for tracking
	k.RecordPoolCreation(ctx, creator, poolID)

	// SECURITY CHECK 3: Create Liquidity Lock
	if err := k.CreateLiquidityLock(ctx, creator, poolID, lpTokens); err != nil {
		// Log error but don't fail pool creation
		ctx.Logger().Error("failed to create liquidity lock", "error", err)
	}

	// Initialize TWAP oracle
	if err := k.RecordTWAPObservation(ctx, poolID); err != nil {
		ctx.Logger().Error("failed to initialize TWAP oracle", "error", err)
	}

	return pool, lpTokens, nil
}

// SecureAddLiquidity is a wrapper around AddLiquidity with all security checks
func (k Keeper) SecureAddLiquidity(
	ctx sdk.Context,
	provider string,
	poolID string,
	amountA sdk.Coin,
	amountB sdk.Coin,
) (sdk.Int, sdk.Dec, error) {
	// SECURITY CHECK 1: Circuit Breaker
	if k.IsCircuitBreakerActive(ctx, poolID) {
		return sdk.ZeroInt(), sdk.ZeroDec(), types.ErrCircuitBreakerActive
	}

	// SECURITY CHECK 2: Flash Loan Protection
	if err := k.CheckFlashLoanProtection(ctx, provider, poolID, true); err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), err
	}

	// Add liquidity
	lpTokens, poolShare, err := k.AddLiquidity(ctx, provider, poolID, amountA, amountB)
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroDec(), err
	}

	// Record liquidity operation
	k.RecordLiquidityBlock(ctx, provider, poolID)

	// SECURITY CHECK 3: Create Liquidity Lock (for new liquidity)
	if err := k.CreateLiquidityLock(ctx, provider, poolID, lpTokens); err != nil {
		ctx.Logger().Error("failed to create liquidity lock", "error", err)
	}

	// Update TWAP oracle
	if err := k.RecordTWAPObservation(ctx, poolID); err != nil {
		ctx.Logger().Error("failed to record TWAP observation", "error", err)
	}

	return lpTokens, poolShare, nil
}

// SecureRemoveLiquidity is a wrapper around RemoveLiquidity with all security checks
func (k Keeper) SecureRemoveLiquidity(
	ctx sdk.Context,
	provider string,
	poolID string,
	lpTokens sdk.Int,
) (sdk.Coin, sdk.Coin, error) {
	// SECURITY CHECK 1: Circuit Breaker
	if k.IsCircuitBreakerActive(ctx, poolID) {
		return sdk.Coin{}, sdk.Coin{}, types.ErrCircuitBreakerActive
	}

	// SECURITY CHECK 2: Flash Loan Protection
	if err := k.CheckFlashLoanProtection(ctx, provider, poolID, false); err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// SECURITY CHECK 3: Liquidity Lock Check
	if err := k.CheckLiquidityLock(ctx, provider, poolID, lpTokens); err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// Remove liquidity
	coinA, coinB, err := k.RemoveLiquidity(ctx, provider, poolID, lpTokens)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// Record liquidity operation
	k.RecordLiquidityBlock(ctx, provider, poolID)

	// Update TWAP oracle
	if err := k.RecordTWAPObservation(ctx, poolID); err != nil {
		ctx.Logger().Error("failed to record TWAP observation", "error", err)
	}

	return coinA, coinB, nil
}

// ValidateSecurityParams validates security parameters
func (k Keeper) ValidateSecurityParams(params types.SecurityParams) error {
	if params.MinBlockDelay > 100 {
		return types.ErrInvalidParam
	}

	if params.MaxTradeSizePercent.LTE(sdk.ZeroDec()) || params.MaxTradeSizePercent.GT(sdk.OneDec()) {
		return types.ErrInvalidParam
	}

	if params.MaxPriceImpactPercent.LTE(sdk.ZeroDec()) || params.MaxPriceImpactPercent.GT(sdk.NewDec(100)) {
		return types.ErrInvalidParam
	}

	if params.LiquidityLockupSeconds < 0 {
		return types.ErrInvalidParam
	}

	if params.PoolCreationCooldown < 0 {
		return types.ErrInvalidParam
	}

	if params.MaxPoolsPerCreator == 0 {
		return types.ErrInvalidParam
	}

	if params.TwapWindowBlocks == 0 {
		return types.ErrInvalidParam
	}

	if params.MinPoolCreationLiquidity.IsNegative() {
		return types.ErrInvalidParam
	}

	if params.MinLiquidityBlocks > 1000 {
		return types.ErrInvalidParam
	}

	if params.WashTradeMinInterval < 0 {
		return types.ErrInvalidParam
	}

	if params.MinTradeAmount.IsNegative() {
		return types.ErrInvalidParam
	}

	if params.MaxOrderVariance.LTE(sdk.ZeroDec()) {
		return types.ErrInvalidParam
	}

	if params.MaxSwapsPerBlock == 0 || params.MaxSwapsPerBlock > 100 {
		return types.ErrInvalidParam
	}

	return nil
}
