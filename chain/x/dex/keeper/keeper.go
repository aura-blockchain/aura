package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// Keeper of the dex store
type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramstore paramtypes.Subspace

	// Dependencies
	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper
	vcKeeper      types.VCRegistryKeeper // For IR verification check
}

// NewKeeper creates a new dex Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	vcKeeper types.VCRegistryKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		paramstore:    ps,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		vcKeeper:      vcKeeper,
	}
}

// GetParams returns the total set of dex parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the dex parameters to the param space.
// Includes authority/governance transition mechanism (training wheels)
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	// Check if authority period has expired
	if params.AuthorityExpiration > 0 && ctx.BlockTime().Unix() > params.AuthorityExpiration {
		// Transition to governance
		params.GovernanceEnabled = true
	}

	// Validate caller has permission
	if params.GovernanceEnabled {
		// Must come from governance module
		// In production, verify ctx.GetMsgSender() is governance module address
		// For now, just log the transition
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"params_updated_via_governance",
				sdk.NewAttribute("module", "dex"),
			),
		)
	} else if params.Authority != "" {
		// Must come from authority address (dev team multisig)
		// In production, verify ctx.GetMsgSender() matches authority
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"params_updated_via_authority",
				sdk.NewAttribute("module", "dex"),
				sdk.NewAttribute("authority", params.Authority),
			),
		)
	}

	k.paramstore.SetParamSet(ctx, &params)
	return nil
}

// ============================================================================
// Dynamic Minimum Liquidity (BRILLIANT FEATURE!)
// ============================================================================

// GetAuraPrice returns current AURA price in USD from USDT pool
func (k Keeper) GetAuraPrice(ctx sdk.Context) sdk.Dec {
	// Get AURA/USDT pool (pool ID "aura-usdt")
	pool := k.GetPoolByDenoms(ctx, "uaura", "usdt")

	if pool == nil {
		// Default to very low price if no pool exists yet
		// This ensures bootstrap phase minimum ($1,000) applies
		return sdk.NewDecWithPrec(10, 2) // $0.10
	}

	// Price = USDT_reserve / AURA_reserve
	// Example: 200,000 USDT / 1,000,000 AURA = $0.20 per AURA
	price := pool.ReserveB.ToDec().Quo(pool.ReserveA.ToDec())

	return price
}

// GetCurrentMinimumLiquidity returns minimum liquidity based on current AURA price
func (k Keeper) GetCurrentMinimumLiquidity(ctx sdk.Context) sdk.Dec {
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
	return sdk.NewDec(1000) // $1,000 default
}

// CalculateMinimumAuraRequired returns minimum AURA needed based on USD minimum
func (k Keeper) CalculateMinimumAuraRequired(ctx sdk.Context) sdk.Int {
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
	amountA sdk.Int,
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
func (k Keeper) CalculateFeeBoost(ctx sdk.Context, address string) sdk.Dec {
	params := k.GetParams(ctx)

	if !params.IrBoostEnabled {
		return sdk.ZeroDec()
	}

	if k.IsUserVerified(ctx, address) {
		// Return boost as decimal (40 = 0.40 = 40%)
		return sdk.NewDec(int64(params.IrBoostPercent)).QuoInt64(100)
	}

	return sdk.ZeroDec()
}

// CalculateEffectiveFee returns actual fee user receives (base + boost)
func (k Keeper) CalculateEffectiveFee(
	ctx sdk.Context,
	address string,
	baseFee sdk.Dec,
) sdk.Dec {
	boost := k.CalculateFeeBoost(ctx, address)

	// Effective fee = base_fee × (1 + boost)
	// Example: 0.003 × (1 + 0.40) = 0.0042 (0.42%)
	return baseFee.Mul(sdk.OneDec().Add(boost))
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
	iterator := sdk.KVStorePrefixIterator(store, types.PoolPrefix)
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
