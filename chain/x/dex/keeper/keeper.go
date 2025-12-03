package keeper

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/common/security"
	"github.com/aequitas/aura/chain/x/dex/types"
)

var (
	ParamsKey = []byte{0x00}
)

// Keeper of the dex store
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	// Dependencies
	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper
	vcKeeper      types.VCRegistryKeeper // For IR verification check

	// Security primitives from common security library
	reentrancyGuard *security.ReentrancyGuard
	pauseGuard      *security.PauseGuard
	inputValidator  *security.InputValidator
	safeMath        *security.SafeMath
	gasLimitGuard   *security.GasLimitGuard
	accessControl   *security.AccessControl
}

// NewKeeper creates a new dex Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	vcKeeper types.VCRegistryKeeper,
) *Keeper {
	return &Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		vcKeeper:      vcKeeper,

		// Initialize security primitives
		reentrancyGuard: security.NewReentrancyGuard(),
		pauseGuard:      security.NewPauseGuard(""), // Admin will be set via params
		inputValidator:  security.NewInputValidator(),
		safeMath:        security.NewSafeMath(),
		gasLimitGuard:   security.NewGasLimitGuard(1_000_000), // 1M gas default
		accessControl:   security.NewAccessControl([]string{}), // Admins set via params
	}
}

// GetParams returns the total set of dex parameters.
func (k Keeper) GetParams(ctx sdk.Context) *types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		ctx.Logger().Error("failed to unmarshal DEX params, returning defaults",
			"error", err,
			"data_len", len(bz))
		return types.DefaultParams()
	}
	return &params
}

// SetParams sets the dex parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params *types.Params) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(params)
	store.Set(ParamsKey, bz)
	return nil
}

// ============================================================================
// Dynamic Minimum Liquidity (BRILLIANT FEATURE!)
// ============================================================================

// GetAuraPrice returns current AURA price in USD from USDT pool.
//
// SECURITY: This function now uses TWAP (Time-Weighted Average Price) to prevent
// flash loan attacks and single-block price manipulation. The price is calculated
// from historical observations with the following protections:
//
// 1. TWAP calculation over configurable window (prevents single-block manipulation)
// 2. Price sanity checks reject movements > 10% per block
// 3. Fallback to validated spot price if insufficient TWAP data
//
// Attack Vectors Prevented:
// - Flash loan price manipulation
// - Single-block oracle attacks
// - Sandwich attacks with extreme price movements
func (k Keeper) GetAuraPrice(ctx sdk.Context) sdkmath.LegacyDec {
	poolID := "uaura-usdt"
	params := k.GetSecurityParams(ctx)

	// Try TWAP first (uses configured window from security params)
	twapPrice, err := k.GetTWAPPrice(ctx, poolID, params.TwapWindowBlocks)
	if err == nil && !twapPrice.IsZero() {
		return twapPrice
	}

	// Fallback: get pool and return spot price
	pool := k.GetPoolByDenoms(ctx, "uaura", "usdt")
	if pool == nil {
		// Default to very low price if no pool exists yet
		// This ensures bootstrap phase minimum ($1,000) applies
		return sdkmath.LegacyNewDecWithPrec(10, 2) // $0.10
	}

	// Parse reserves
	reserveA, ok := sdkmath.NewIntFromString(pool.ReserveA)
	if !ok || reserveA.IsZero() {
		return sdkmath.LegacyNewDecWithPrec(10, 2) // $0.10 fallback
	}
	reserveB, ok := sdkmath.NewIntFromString(pool.ReserveB)
	if !ok {
		return sdkmath.LegacyNewDecWithPrec(10, 2) // $0.10 fallback
	}

	// Calculate spot price
	spotPrice := sdkmath.LegacyNewDecFromInt(reserveB).Quo(sdkmath.LegacyNewDecFromInt(reserveA))

	// Apply price sanity check even on fallback
	lastPrice := k.GetLastRecordedPrice(ctx, poolID)
	if !lastPrice.IsZero() {
		// Reject if movement > 10%
		maxChange := lastPrice.Mul(sdkmath.LegacyNewDecWithPrec(10, 2)) // 10%
		if spotPrice.Sub(lastPrice).Abs().GT(maxChange) {
			// Suspicious movement, return last valid price
			return lastPrice
		}
	}

	return spotPrice
}

// GetCurrentMinimumLiquidity returns minimum liquidity based on current AURA price
func (k Keeper) GetCurrentMinimumLiquidity(ctx sdk.Context) sdkmath.LegacyDec {
	params := k.GetParams(ctx)
	auraPrice := k.GetAuraPrice(ctx)

	// Find appropriate tier
	for _, tier := range params.MinLiquidityTiers {
		maxPrice, err := sdkmath.LegacyNewDecFromStr(tier.MaxAuraPriceUsd)
		if err != nil {
			continue
		}
		minLiq, err := sdkmath.LegacyNewDecFromStr(tier.MinLiquidityUsd)
		if err != nil {
			continue
		}

		// If max_price is 0, it's the highest tier (no maximum)
		if maxPrice.IsZero() {
			return minLiq
		}

		// If current price is below tier maximum, use this tier
		if auraPrice.LT(maxPrice) {
			return minLiq
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
			lpTokens, ok := sdkmath.NewIntFromString(provider.LpTokens)
			if !ok {
				return false
			}
			return !lpTokens.IsZero()
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
	// Query vcregistry keeper for IR status
	irScore := k.vcKeeper.GetIRScore(ctx, address)
	return irScore >= 100
}

// CalculateFeeBoost returns fee boost percentage for user
func (k Keeper) CalculateFeeBoost(ctx sdk.Context, address string) sdkmath.LegacyDec {
	params := k.GetParams(ctx)

	if !params.IrBoostEnabled {
		return sdkmath.LegacyZeroDec()
	}

	if k.IsUserVerified(ctx, address) {
		// Return boost as decimal (40 = 0.40 = 40%)
		return sdkmath.LegacyNewDec(int64(params.IrBoostPercent)).QuoInt64(100)
	}

	return sdkmath.LegacyZeroDec()
}

// CalculateEffectiveFee returns actual fee user receives (base + boost)
func (k Keeper) CalculateEffectiveFee(
	ctx sdk.Context,
	address string,
	baseFee sdkmath.LegacyDec,
) sdkmath.LegacyDec {
	boost := k.CalculateFeeBoost(ctx, address)

	// Effective fee = base_fee × (1 + boost)
	// Example: 0.003 × (1 + 0.40) = 0.0042 (0.42%)
	return baseFee.Mul(sdkmath.LegacyOneDec().Add(boost))
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
	k.cdc.MustUnmarshal(bz, &pool)
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
	feeDec, err := sdkmath.LegacyNewDecFromStr(params.TradingFee)
	if err != nil {
		return sdkmath.ZeroInt(), fmt.Errorf("invalid trading fee: %w", err)
	}

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
func (k Keeper) SetPool(ctx sdk.Context, pool *types.LiquidityPool) {
	store := ctx.KVStore(k.storeKey)
	key := types.PoolKey(pool.PoolId)

	bz := k.cdc.MustMarshal(pool)
	store.Set(key, bz)
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

	var pools []*types.LiquidityPool
	for ; iterator.Valid(); iterator.Next() {
		var pool types.LiquidityPool
		k.cdc.MustUnmarshal(iterator.Value(), &pool)
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

// parseReserve parses a reserve string to math.Int
func (k Keeper) parseReserve(reserveStr string) (sdkmath.Int, error) {
	reserve, ok := sdkmath.NewIntFromString(reserveStr)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("invalid reserve amount: %s", reserveStr)
	}
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

// parseLPTokens parses LP token string to math.Int
func (k Keeper) parseLPTokens(lpTokensStr string) (sdkmath.Int, error) {
	lpTokens, ok := sdkmath.NewIntFromString(lpTokensStr)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("invalid LP tokens amount: %s", lpTokensStr)
	}
	return lpTokens, nil
}

// parseFeePercentage parses fee percentage string to LegacyDec
func (k Keeper) parseFeePercentage(feeStr string) (sdkmath.LegacyDec, error) {
	fee, err := sdkmath.LegacyNewDecFromStr(feeStr)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("invalid fee percentage: %s", feeStr)
	}
	return fee, nil
}
