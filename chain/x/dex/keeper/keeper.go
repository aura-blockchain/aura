package keeper

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
)

var (
	ParamsKey = []byte{0x00}
)

const (
	// MinTWAPObservations is the minimum number of TWAP price observations required
	// before using TWAP price. This prevents oracle manipulation attacks by ensuring
	// sufficient historical data exists before trusting the TWAP calculation.
	// With insufficient observations, the governance fallback price is used instead.
	MinTWAPObservations = 100
)

// Keeper of the dex store
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	// Dependencies
	bankKeeper     types.BankKeeper
	accountKeeper  types.AccountKeeper
	vcKeeper       types.VCRegistryKeeper // For IR verification check
	securityKeeper types.SecurityKeeper   // Centralized security primitives
}

// NewKeeper creates a new dex Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	vcKeeper types.VCRegistryKeeper,
	securityKeeper types.SecurityKeeper,
) *Keeper {
	return &Keeper{
		storeKey:       storeKey,
		cdc:            cdc,
		bankKeeper:     bankKeeper,
		accountKeeper:  accountKeeper,
		vcKeeper:       vcKeeper,
		securityKeeper: securityKeeper,
	}
}

// GetStoreKey returns the store key (used for testing)
func (k Keeper) GetStoreKey() storetypes.StoreKey {
	return k.storeKey
}

// AccountKeeper returns the account keeper (used for testing)
func (k Keeper) AccountKeeper() types.AccountKeeper {
	return k.accountKeeper
}

// BankKeeper returns the bank keeper (used for testing)
func (k Keeper) BankKeeper() types.BankKeeper {
	return k.bankKeeper
}

// GetParams returns the total set of dex parameters.
func (k Keeper) GetParams(ctx sdk.Context) *types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ParamsKey)
	if bz == nil {
		defaults := types.DefaultParams()
		return &defaults
	}

	var params types.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		ctx.Logger().Error("failed to unmarshal DEX params, returning defaults",
			"error", err,
			"data_len", len(bz))
		defaults := types.DefaultParams()
		return &defaults
	}
	return &params
}

// SetParams sets the dex parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params *types.Params) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(params)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal params: %v", err)
	}
	store.Set(ParamsKey, bz)
	return nil
}

// ============================================================================
// Dynamic Minimum Liquidity (BRILLIANT FEATURE!)
// ============================================================================

// GetGovernanceFallbackPrice returns the governance-controlled fallback price for AURA.
// This price is used when TWAP data is insufficient (< MinTWAPObservations) to prevent
// oracle manipulation attacks.
//
// SECURITY: The fallback price can only be modified through governance proposals,
// preventing attackers from manipulating it. This ensures that even with insufficient
// TWAP data, minimum liquidity requirements are calculated using a safe, governance-
// controlled price rather than an easily manipulable spot price.
//
// Returns:
//   - Governance fallback price from params, or $0.10 default if not set
func (k Keeper) GetGovernanceFallbackPrice(ctx sdk.Context) sdkmath.LegacyDec {
	params := k.GetParams(ctx)

	// Validate fallback price is positive
	if params.GovernanceFallbackPrice.IsNil() || params.GovernanceFallbackPrice.IsZero() || params.GovernanceFallbackPrice.IsNegative() {
		// Default to $0.10 if not set or invalid
		return sdkmath.LegacyNewDecWithPrec(10, 2) // $0.10
	}

	return params.GovernanceFallbackPrice
}

// GetAuraPrice returns current AURA price in USD from USDT pool.
//
// SECURITY: This function implements comprehensive oracle manipulation protection:
//
// 1. TWAP Requirement: Requires minimum MinTWAPObservations (100) before using TWAP.
//    This prevents manipulation during bootstrap when observation count is low.
//
// 2. Governance Fallback: When TWAP has insufficient observations, uses governance-
//    controlled fallback price instead of manipulable spot price. This price can only
//    be updated through governance proposals, preventing attacker manipulation.
//
// 3. No Spot Price Fallback: NEVER uses spot price from pools, as these can be
//    manipulated via flash loans, large trades, or low liquidity attacks.
//
// Attack Vectors Prevented:
// - Flash loan price manipulation (eliminated by TWAP + governance fallback)
// - Bootstrap/low-observation manipulation (min observation requirement)
// - Governance-based manipulation (requires full governance process)
// - Spot price manipulation (spot price never used)
//
// Price Selection Logic:
// - IF TWAP observations >= MinTWAPObservations (100): Use TWAP price ✓ (secure)
// - ELSE: Use governance fallback price ✓ (secure, governance-controlled)
// - NEVER: Use spot price ✗ (manipulable, forbidden)
func (k Keeper) GetAuraPrice(ctx sdk.Context) sdkmath.LegacyDec {
	poolID := "uaura-usdt"
	params := k.GetSecurityParams(ctx)

	// Try TWAP with observation count check
	twapPrice, observationCount, err := k.GetTWAPPriceWithCount(ctx, poolID, params.TwapWindowBlocks)
	if err == nil && !twapPrice.IsZero() && observationCount >= MinTWAPObservations {
		// Sufficient observations (>= 100), TWAP is trustworthy
		ctx.Logger().Debug("using TWAP price for AURA",
			"price", twapPrice.String(),
			"observations", observationCount,
			"pool_id", poolID)
		return twapPrice
	}

	// Insufficient TWAP data: use governance fallback price instead of spot price
	// This prevents oracle manipulation attacks that exploit low observation counts
	fallbackPrice := k.GetGovernanceFallbackPrice(ctx)

	ctx.Logger().Info("using governance fallback price for AURA",
		"fallback_price", fallbackPrice.String(),
		"reason", "insufficient_twap_observations",
		"observation_count", observationCount,
		"min_required", MinTWAPObservations,
		"pool_id", poolID)

	// Emit event for monitoring/alerting
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"aura_price_fallback",
			sdk.NewAttribute("pool_id", poolID),
			sdk.NewAttribute("fallback_price", fallbackPrice.String()),
			sdk.NewAttribute("observation_count", fmt.Sprintf("%d", observationCount)),
			sdk.NewAttribute("min_required", fmt.Sprintf("%d", MinTWAPObservations)),
			sdk.NewAttribute("reason", "insufficient_twap_observations"),
		),
	)

	return fallbackPrice
}

// GetCurrentMinimumLiquidity returns minimum liquidity based on current AURA price
func (k Keeper) GetCurrentMinimumLiquidity(ctx sdk.Context) sdkmath.LegacyDec {
	params := k.GetParams(ctx)
	auraPrice := k.GetAuraPrice(ctx)

	// Find appropriate tier
	for _, tier := range params.MinLiquidityTiers {
		// If max_price is 0, it's the highest tier (no maximum)
		if tier.MaxAuraPriceUsd.IsZero() {
			return tier.MinLiquidityUsd
		}

		// If current price is below tier maximum, use this tier
		if auraPrice.LT(tier.MaxAuraPriceUsd) {
			return tier.MinLiquidityUsd
		}
	}

	// Fallback (should never reach here if tiers configured properly)
	return sdkmath.LegacyNewDec(1000) // $1,000 default
}

// CalculateMinimumAuraRequired returns minimum AURA needed based on USD minimum
func (k Keeper) CalculateMinimumAuraRequired(ctx sdk.Context) sdkmath.Int {
	minUSD := k.GetCurrentMinimumLiquidity(ctx)
	auraPrice := k.GetAuraPrice(ctx)

	// Minimum AURA = Minimum USD / Price
	// Example: $1,000 / $0.20 = 5,000 AURA
	minAura := minUSD.Quo(auraPrice)

	return minAura.TruncateInt()
}

// IsExistingLP checks if address is already a liquidity provider in pool
func (k Keeper) IsExistingLP(ctx sdk.Context, address string, poolID string) bool {
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return false
	}

	// Check if address has LP tokens
	for _, provider := range pool.Providers {
		if provider.Address == address {
			return !provider.LpTokens.IsZero()
		}
	}

	return false
}

// CheckMinimumLiquidity validates if liquidity meets minimum (only for new LPs)
func (k Keeper) CheckMinimumLiquidity(
	ctx sdk.Context,
	provider string,
	poolID string,
	amountA sdkmath.Int,
) error {
	// Existing LPs are grandfathered - they can add ANY amount!
	if k.IsExistingLP(ctx, provider, poolID) {
		return nil
	}

	// New LPs must meet current minimum
	minRequired := k.CalculateMinimumAuraRequired(ctx)

	if amountA.LT(minRequired) {
		auraPrice := k.GetAuraPrice(ctx)
		minUSD := k.GetCurrentMinimumLiquidity(ctx)

		return fmt.Errorf(
			"minimum liquidity requirement not met: need %s AURA (approximately $%s USD at current price of $%s), provided %s AURA",
			minRequired.String(),
			minUSD.String(),
			auraPrice.String(),
			amountA.String(),
		)
	}

	return nil
}

// ============================================================================
// IR Boost Feature (Verified Users Earn 40% More!)
// ============================================================================

// IsUserVerified checks if user has completed 100 IR points
func (k Keeper) IsUserVerified(ctx sdk.Context, address string) bool {
	// If vcKeeper is not set, treat all users as unverified
	if k.vcKeeper == nil {
		return false
	}

	// Query vcregistry keeper for IR status
	irScore := k.vcKeeper.GetIRScore(ctx, address)
	return irScore >= 100
}

// CalculateFeeBoost returns fee boost percentage for user
//
// SECURITY: This function validates IR boost parameters to prevent
// manipulation or configuration errors that could lead to:
// - Excessive fee boosts (>100%)
// - Negative boosts
// - Integer overflow when applied to amounts
func (k Keeper) CalculateFeeBoost(ctx sdk.Context, address string) sdkmath.LegacyDec {
	params := k.GetParams(ctx)

	if !params.IrBoostEnabled {
		return sdkmath.LegacyZeroDec()
	}

	if k.IsUserVerified(ctx, address) {
		// Validate boost percentage is reasonable (0-100%)
		if params.IrBoostPercent > 100 {
			ctx.Logger().Error("invalid IR boost percentage, using 0",
				"boost_percent", params.IrBoostPercent)
			return sdkmath.LegacyZeroDec()
		}

		// Return boost as decimal (40 = 0.40 = 40%)
		return sdkmath.LegacyNewDec(int64(params.IrBoostPercent)).QuoInt64(100)
	}

	return sdkmath.LegacyZeroDec()
}

// CalculateEffectiveFee returns actual fee user receives (base + boost)
//
// SECURITY: Uses validated decimal multiplication to calculate boosted fees.
// While LegacyDec multiplication is inherently safer than Int multiplication,
// we still validate inputs to prevent edge cases.
func (k Keeper) CalculateEffectiveFee(
	ctx sdk.Context,
	address string,
	baseFee sdkmath.LegacyDec,
) sdkmath.LegacyDec {
	// Validate base fee is non-negative
	if baseFee.IsNegative() {
		ctx.Logger().Error("negative base fee provided, using zero",
			"base_fee", baseFee.String())
		return sdkmath.LegacyZeroDec()
	}

	boost := k.CalculateFeeBoost(ctx, address)

	// Effective fee = base_fee × (1 + boost)
	// Example: 0.003 × (1 + 0.40) = 0.0042 (0.42%)
	effectiveFee := baseFee.Mul(sdkmath.LegacyOneDec().Add(boost))

	// Sanity check: effective fee should never be negative
	if effectiveFee.IsNegative() {
		ctx.Logger().Error("effective fee calculation resulted in negative value",
			"base_fee", baseFee.String(),
			"boost", boost.String())
		return baseFee // Fallback to base fee without boost
	}

	return effectiveFee
}

// ============================================================================
// Pool Management
// ============================================================================

// GetPool returns a pool by ID
func (k Keeper) GetPool(ctx sdk.Context, poolID string) *types.LiquidityPool {
	store := ctx.KVStore(k.storeKey)
	key := types.PoolKey(poolID)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var pool types.LiquidityPool
	if err := k.cdc.Unmarshal(bz, &pool); err != nil {
		ctx.Logger().Error("failed to unmarshal liquidity pool",
			"pool_id", poolID,
			"error", err,
			"data_len", len(bz))
		return nil
	}
	return &pool
}

// GetPoolByDenoms returns a pool by token pair
func (k Keeper) GetPoolByDenoms(ctx sdk.Context, denomA, denomB string) *types.LiquidityPool {
	// Normalize order (alphabetical)
	if denomA > denomB {
		denomA, denomB = denomB, denomA
	}

	poolID := fmt.Sprintf("%s-%s", denomA, denomB)
	return k.GetPool(ctx, poolID)
}

// GetPoolPrice returns the price of baseDenom in the given pool.
func (k Keeper) GetPoolPrice(ctx sdk.Context, poolID, baseDenom string) (sdkmath.LegacyDec, error) {
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return sdkmath.LegacyZeroDec(), fmt.Errorf("pool %s not found", poolID)
	}

	base := strings.ToLower(baseDenom)
	denomA := strings.ToLower(pool.DenomA)
	denomB := strings.ToLower(pool.DenomB)

	var baseReserve, quoteReserve sdkmath.Int
	switch {
	case base == denomA:
		var err error
		baseReserve, err = k.parseReserve(pool.ReserveA)
		if err != nil {
			return sdkmath.LegacyZeroDec(), err
		}
		quoteReserve, err = k.parseReserve(pool.ReserveB)
		if err != nil {
			return sdkmath.LegacyZeroDec(), err
		}
	case base == denomB:
		var err error
		baseReserve, err = k.parseReserve(pool.ReserveB)
		if err != nil {
			return sdkmath.LegacyZeroDec(), err
		}
		quoteReserve, err = k.parseReserve(pool.ReserveA)
		if err != nil {
			return sdkmath.LegacyZeroDec(), err
		}
	default:
		return sdkmath.LegacyZeroDec(), fmt.Errorf("denom %s not part of pool %s", baseDenom, poolID)
	}

	if baseReserve.IsZero() {
		return sdkmath.LegacyZeroDec(), fmt.Errorf("base reserve is zero for pool %s", poolID)
	}

	return quoteReserve.ToLegacyDec().Quo(baseReserve.ToLegacyDec()), nil
}

// GetSpotPrice returns the instantaneous price between two denoms in a pool.
func (k Keeper) GetSpotPrice(ctx sdk.Context, poolID, baseDenom, quoteDenom string) (sdkmath.LegacyDec, error) {
	price, err := k.GetPoolPrice(ctx, poolID, baseDenom)
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}

	baseLower := strings.ToLower(baseDenom)
	quoteLower := strings.ToLower(quoteDenom)
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return sdkmath.LegacyZeroDec(), fmt.Errorf("pool %s not found", poolID)
	}

	if (baseLower == strings.ToLower(pool.DenomA) && quoteLower == strings.ToLower(pool.DenomB)) ||
		(baseLower == strings.ToLower(pool.DenomB) && quoteLower == strings.ToLower(pool.DenomA)) {
		return price, nil
	}

	return sdkmath.LegacyZeroDec(), fmt.Errorf("pair %s/%s not supported by pool %s", baseDenom, quoteDenom, poolID)
}

// CalculateSwapFee returns the fee amount based on current params.
//
// SECURITY: This function uses SafeMulDec to prevent integer overflow attacks
// where user-controlled amounts could be crafted to:
// - Cause overflow resulting in zero/negative fees (protocol revenue loss)
// - Wrap around to produce incorrect fee amounts (theft)
//
// All fee calculations MUST use overflow-safe arithmetic.
func (k Keeper) CalculateSwapFee(ctx sdk.Context, amount sdkmath.Int) (sdkmath.Int, error) {
	// Validate input is positive
	if err := types.CheckPositive(amount, "swap amount"); err != nil {
		return sdkmath.ZeroInt(), err
	}

	params := k.GetParams(ctx)
	feeDec := params.TradingFee

	// Validate fee rate is non-negative
	if feeDec.IsNegative() {
		return sdkmath.ZeroInt(), fmt.Errorf("fee rate cannot be negative: %s", feeDec.String())
	}

	// Use safe multiplication to prevent overflow
	fee, err := types.SafeMulDec(amount, feeDec)
	if err != nil {
		return sdkmath.ZeroInt(), fmt.Errorf("fee calculation overflow: %w", err)
	}

	// Ensure minimum fee of 1 if non-zero
	if fee.IsZero() && !feeDec.IsZero() {
		return sdkmath.NewInt(1), nil
	}

	return fee, nil
}

// GetCollectedFees returns total fees collected for a pool.
func (k Keeper) GetCollectedFees(ctx sdk.Context, poolID string) (sdkmath.Int, error) {
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return sdkmath.ZeroInt(), fmt.Errorf("pool %s not found", poolID)
	}

	fees, err := k.parseReserve(pool.TotalFeesCollected)
	if err != nil {
		return sdkmath.ZeroInt(), err
	}
	return fees, nil
}

// SetPool stores a pool
func (k Keeper) SetPool(ctx sdk.Context, pool *types.LiquidityPool) error {
	store := ctx.KVStore(k.storeKey)
	key := types.PoolKey(pool.PoolId)

	bz, err := k.cdc.Marshal(pool)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal pool %s: %v", pool.PoolId, err)
	}
	store.Set(key, bz)
	return nil
}

// DeletePool removes a pool
func (k Keeper) DeletePool(ctx sdk.Context, poolID string) {
	store := ctx.KVStore(k.storeKey)
	key := types.PoolKey(poolID)
	store.Delete(key)
}

// GetAllPools returns all pools
func (k Keeper) GetAllPools(ctx sdk.Context) []*types.LiquidityPool {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PoolPrefix)
	defer iterator.Close()

	pools := make([]*types.LiquidityPool, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var pool types.LiquidityPool
		if err := k.cdc.Unmarshal(iterator.Value(), &pool); err != nil {
			ctx.Logger().Error("failed to unmarshal pool in GetAllPools, skipping",
				"error", err,
				"data_len", len(iterator.Value()))
			continue
		}
		pools = append(pools, &pool)
	}

	return pools
}

// GeneratePoolID generates unique pool ID from denoms
func (k Keeper) GeneratePoolID(denomA, denomB string) string {
	// Normalize order (alphabetical)
	if denomA > denomB {
		denomA, denomB = denomB, denomA
	}
	return fmt.Sprintf("%s-%s", denomA, denomB)
}

// ============================================================================
// Reserve Parsing Helpers
// ============================================================================

// parseReserve returns the reserve value directly (customtype in proto means it's already math.Int)
// This function is kept for backward compatibility but now just returns the value
func (k Keeper) parseReserve(reserve sdkmath.Int) (sdkmath.Int, error) {
	return reserve, nil
}

// getPoolReserves returns both reserves as math.Int
func (k Keeper) getPoolReserves(pool *types.LiquidityPool) (sdkmath.Int, sdkmath.Int, error) {
	reserveA, err := k.parseReserve(pool.ReserveA)
	if err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, err
	}
	reserveB, err := k.parseReserve(pool.ReserveB)
	if err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, err
	}
	return reserveA, reserveB, nil
}

// parseLPTokens parses LP token string to math.Int.
// nolint:unused // retained for potential CLI/admin usage.
func (k Keeper) parseLPTokens(lpTokensStr string) (sdkmath.Int, error) {
	lpTokens, ok := sdkmath.NewIntFromString(lpTokensStr)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("invalid LP tokens amount: %s", lpTokensStr)
	}
	return lpTokens, nil
}

// parseFeePercentage parses fee percentage string to LegacyDec.
// nolint:unused // retained for potential CLI/admin usage.
func (k Keeper) parseFeePercentage(feeStr string) (sdkmath.LegacyDec, error) {
	fee, err := sdkmath.LegacyNewDecFromStr(feeStr)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("invalid fee percentage: %s", feeStr)
	}
	return fee, nil
}
