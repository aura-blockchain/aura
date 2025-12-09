package keeper

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// SecureSwapExactIn is a wrapper around SwapExactIn with all security checks
func (k Keeper) SecureSwapExactIn(
	ctx sdk.Context,
	sender string,
	poolID string,
	coinIn sdk.Coin,
	minAmountOut math.Int,
	maxSlippageBps uint64,
) (math.Int, math.LegacyDec, math.LegacyDec, error) {
	if err := k.validateSwapInputs(ctx, sender, poolID, coinIn, minAmountOut); err != nil {
		return math.ZeroInt(), math.LegacyZeroDec(), math.LegacyZeroDec(), err
	}

	// SECURITY CHECK 1: Circuit Breaker
	if k.IsCircuitBreakerActive(ctx, poolID) {
		return math.ZeroInt(), math.LegacyZeroDec(), math.LegacyZeroDec(), types.ErrCircuitBreakerActive
	}

	// SECURITY CHECK 2: Front-Running Protection
	if err := k.CheckFrontRunningProtection(ctx, sender, poolID); err != nil {
		return math.ZeroInt(), math.LegacyZeroDec(), math.LegacyZeroDec(), err
	}

	// SECURITY CHECK 3: MEV Protection
	if err := k.CheckMEVProtection(ctx, sender); err != nil {
		return math.ZeroInt(), math.LegacyZeroDec(), math.LegacyZeroDec(), err
	}

	// SECURITY CHECK 4: Wash Trading Detection
	if err := k.DetectWashTrading(ctx, sender, poolID); err != nil {
		return math.ZeroInt(), math.LegacyZeroDec(), math.LegacyZeroDec(), err
	}

	// SECURITY CHECK 5: Dust Attack Prevention
	if err := k.CheckDustAttack(ctx, coinIn.Amount); err != nil {
		return math.ZeroInt(), math.LegacyZeroDec(), math.LegacyZeroDec(), err
	}

	// SECURITY CHECK 6: Maximum Trade Size
	if err := k.CheckMaxTradeSize(ctx, poolID, coinIn.Amount); err != nil {
		return math.ZeroInt(), math.LegacyZeroDec(), math.LegacyZeroDec(), err
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
		return math.ZeroInt(), math.LegacyZeroDec(), math.LegacyZeroDec(), err
	}

	// SECURITY CHECK 7: Pool-Specific Slippage Limits
	if err := k.CheckPoolSlippageLimit(ctx, poolID, priceImpact); err != nil {
		return math.ZeroInt(), math.LegacyZeroDec(), math.LegacyZeroDec(), err
	}

	// SECURITY CHECK 8: Price Impact Rejection Threshold
	if err := k.CheckPriceImpactThreshold(ctx, priceImpact); err != nil {
		return math.ZeroInt(), math.LegacyZeroDec(), math.LegacyZeroDec(), err
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
) (*types.LiquidityPool, math.Int, error) {
	if err := k.validatePoolInputs(ctx, creator, denomA, denomB, amountA, amountB); err != nil {
		return nil, math.ZeroInt(), err
	}

	// SECURITY CHECK 1: Circuit Breaker
	poolID := k.GeneratePoolID(denomA, denomB)
	if k.IsCircuitBreakerActive(ctx, poolID) {
		return nil, math.ZeroInt(), types.ErrCircuitBreakerActive
	}

	// SECURITY CHECK 2: Pool Creation Limits
	if err := k.CheckPoolCreationLimits(ctx, creator, amountA.Amount); err != nil {
		return nil, math.ZeroInt(), err
	}

	// Create the pool
	pool, lpTokens, err := k.CreatePool(ctx, creator, denomA, denomB, amountA, amountB)
	if err != nil {
		return nil, math.ZeroInt(), err
	}

	// Record pool creation for tracking (audit trail for compliance)
	k.RecordPoolCreation(ctx, creator, poolID, denomA, denomB, amountA.Amount, amountB.Amount)

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
) (math.Int, math.LegacyDec, error) {
	if err := k.validateAddLiquidity(ctx, provider, poolID, amountA, amountB); err != nil {
		return math.ZeroInt(), math.LegacyZeroDec(), err
	}

	// SECURITY CHECK 1: Circuit Breaker
	if k.IsCircuitBreakerActive(ctx, poolID) {
		return math.ZeroInt(), math.LegacyZeroDec(), types.ErrCircuitBreakerActive
	}

	// SECURITY CHECK 2: Flash Loan Protection
	if err := k.CheckFlashLoanProtection(ctx, provider, poolID, true); err != nil {
		return math.ZeroInt(), math.LegacyZeroDec(), err
	}

	// Add liquidity
	lpTokens, poolShare, err := k.AddLiquidity(ctx, provider, poolID, amountA, amountB)
	if err != nil {
		return math.ZeroInt(), math.LegacyZeroDec(), err
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
	lpTokens math.Int,
) (sdk.Coin, sdk.Coin, error) {
	if err := k.validateRemoveLiquidity(ctx, provider, poolID, lpTokens); err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

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

	// MaxTradeSizePercent is already LegacyDec
	maxTradeSizePercent := params.MaxTradeSizePercent
	if maxTradeSizePercent.LTE(math.LegacyZeroDec()) || maxTradeSizePercent.GT(math.LegacyOneDec()) {
		return types.ErrInvalidParam
	}

	// MaxPriceImpactPercent is already LegacyDec
	maxPriceImpactPercent := params.MaxPriceImpactPercent
	if maxPriceImpactPercent.LTE(math.LegacyZeroDec()) || maxPriceImpactPercent.GT(math.LegacyNewDec(100)) {
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

	// MinPoolCreationLiquidity is already Int
	minPoolCreationLiquidity := params.MinPoolCreationLiquidity
	if minPoolCreationLiquidity.IsNegative() {
		return types.ErrInvalidParam
	}

	if params.MinLiquidityBlocks > 1000 {
		return types.ErrInvalidParam
	}

	if params.WashTradeMinInterval < 0 {
		return types.ErrInvalidParam
	}

	// MinTradeAmount is already Int
	minTradeAmount := params.MinTradeAmount
	if minTradeAmount.IsNegative() {
		return types.ErrInvalidParam
	}

	// MaxOrderVariance is already LegacyDec
	maxOrderVariance := params.MaxOrderVariance
	if maxOrderVariance.LTE(math.LegacyZeroDec()) {
		return types.ErrInvalidParam
	}

	if params.MaxSwapsPerBlock == 0 || params.MaxSwapsPerBlock > 100 {
		return types.ErrInvalidParam
	}

	return nil
}

func (k Keeper) validateSwapInputs(ctx sdk.Context, sender, poolID string, coinIn sdk.Coin, minAmountOut math.Int) error {
	if strings.TrimSpace(sender) == "" {
		return k.logValidationError(ctx, "swap", "sender required")
	}
	if strings.TrimSpace(poolID) == "" {
		return k.logValidationError(ctx, "swap", "pool id required")
	}
	if err := coinIn.Validate(); err != nil {
		return k.logValidationError(ctx, "swap", fmt.Sprintf("invalid coin in: %v", err))
	}
	if !minAmountOut.IsPositive() {
		return k.logValidationError(ctx, "swap", "min amount out must be positive")
	}
	if pool := k.GetPool(ctx, poolID); pool == nil {
		return k.logValidationError(ctx, "swap", fmt.Sprintf("pool %s not found", poolID))
	}
	return nil
}

func (k Keeper) validatePoolInputs(ctx sdk.Context, creator, denomA, denomB string, amountA, amountB sdk.Coin) error {
	if strings.TrimSpace(creator) == "" {
		return k.logValidationError(ctx, "create_pool", "creator required")
	}
	if denomA == "" || denomB == "" {
		return k.logValidationError(ctx, "create_pool", "denoms required")
	}
	if denomA == denomB {
		return k.logValidationError(ctx, "create_pool", "denoms must differ")
	}
	if err := amountA.Validate(); err != nil {
		return k.logValidationError(ctx, "create_pool", fmt.Sprintf("invalid amountA: %v", err))
	}
	if err := amountB.Validate(); err != nil {
		return k.logValidationError(ctx, "create_pool", fmt.Sprintf("invalid amountB: %v", err))
	}
	if !amountA.Amount.IsPositive() || !amountB.Amount.IsPositive() {
		return k.logValidationError(ctx, "create_pool", "amounts must be positive")
	}
	return nil
}

func (k Keeper) validateAddLiquidity(ctx sdk.Context, provider, poolID string, amountA, amountB sdk.Coin) error {
	if strings.TrimSpace(provider) == "" {
		return k.logValidationError(ctx, "add_liquidity", "provider required")
	}
	if strings.TrimSpace(poolID) == "" {
		return k.logValidationError(ctx, "add_liquidity", "pool id required")
	}
	if err := amountA.Validate(); err != nil {
		return k.logValidationError(ctx, "add_liquidity", fmt.Sprintf("invalid amountA: %v", err))
	}
	if err := amountB.Validate(); err != nil {
		return k.logValidationError(ctx, "add_liquidity", fmt.Sprintf("invalid amountB: %v", err))
	}
	if !amountA.Amount.IsPositive() || !amountB.Amount.IsPositive() {
		return k.logValidationError(ctx, "add_liquidity", "amounts must be positive")
	}
	if pool := k.GetPool(ctx, poolID); pool == nil {
		return k.logValidationError(ctx, "add_liquidity", fmt.Sprintf("pool %s not found", poolID))
	}
	return nil
}

func (k Keeper) validateRemoveLiquidity(ctx sdk.Context, provider, poolID string, lpTokens math.Int) error {
	if strings.TrimSpace(provider) == "" {
		return k.logValidationError(ctx, "remove_liquidity", "provider required")
	}
	if strings.TrimSpace(poolID) == "" {
		return k.logValidationError(ctx, "remove_liquidity", "pool id required")
	}
	if !lpTokens.IsPositive() {
		return k.logValidationError(ctx, "remove_liquidity", "lp tokens must be positive")
	}
	if pool := k.GetPool(ctx, poolID); pool == nil {
		return k.logValidationError(ctx, "remove_liquidity", fmt.Sprintf("pool %s not found", poolID))
	}
	return nil
}

func (k Keeper) logValidationError(ctx sdk.Context, op, reason string) error {
	ctx.Logger().Error("dex validation failed", "operation", op, "reason", reason)
	return fmt.Errorf("%s", reason)
}
