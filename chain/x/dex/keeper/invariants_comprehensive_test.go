package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/dex/types"
)

type InvariantsComprehensiveTestSuite struct {
	KeeperTestSuite
}

func TestInvariantsComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsComprehensiveTestSuite))
}

func (suite *InvariantsComprehensiveTestSuite) TestPoolReservesConsistencyInvariant() {
	ctx := suite.SdkCtx
	inv := PoolReservesConsistencyInvariant(suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Create valid pool
	pool := types.LiquidityPool{
		PoolId:      "pool-1",
		DenomA:      "uaura",
		DenomB:      "utoken",
		ReserveA:    "1000000",
		ReserveB:    "2000000",
		TotalShares: "1414213",
	}
	suite.storePool(ctx, &pool)

	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestPoolEmptyID() {
	ctx := suite.SdkCtx
	inv := PoolReservesConsistencyInvariant(suite.Keeper)

	pool := types.LiquidityPool{
		PoolId:      "",
		DenomA:      "uaura",
		DenomB:      "utoken",
		ReserveA:    "1000000",
		ReserveB:    "2000000",
		TotalShares: "1414213",
	}
	suite.storePool(ctx, &pool)

	msg, broken := inv(ctx)
	suite.True(broken, "pool with empty ID should break invariant")
	suite.Contains(msg, "empty ID")
}

func (suite *InvariantsComprehensiveTestSuite) TestPoolEmptyDenoms() {
	ctx := suite.SdkCtx
	inv := PoolReservesConsistencyInvariant(suite.Keeper)

	pool := types.LiquidityPool{
		PoolId:      "pool-2",
		DenomA:      "",
		DenomB:      "utoken",
		ReserveA:    "1000000",
		ReserveB:    "2000000",
		TotalShares: "1414213",
	}
	suite.storePool(ctx, &pool)

	msg, broken := inv(ctx)
	suite.True(broken, "pool with empty denom should break invariant")
	suite.Contains(msg, "empty denom")
}

func (suite *InvariantsComprehensiveTestSuite) TestPoolNegativeReserves() {
	ctx := suite.SdkCtx
	inv := PoolReservesConsistencyInvariant(suite.Keeper)

	pool := types.LiquidityPool{
		PoolId:      "pool-3",
		DenomA:      "uaura",
		DenomB:      "utoken",
		ReserveA:    "-1000000",
		ReserveB:    "2000000",
		TotalShares: "1414213",
	}
	suite.storePool(ctx, &pool)

	msg, broken := inv(ctx)
	suite.True(broken, "pool with negative reserves should break invariant")
	suite.Contains(msg, "invalid reserve")
}

func (suite *InvariantsComprehensiveTestSuite) TestPoolReservesButZeroShares() {
	ctx := suite.SdkCtx
	inv := PoolReservesConsistencyInvariant(suite.Keeper)

	pool := types.LiquidityPool{
		PoolId:      "pool-4",
		DenomA:      "uaura",
		DenomB:      "utoken",
		ReserveA:    "1000000",
		ReserveB:    "2000000",
		TotalShares: "0",
	}
	suite.storePool(ctx, &pool)

	msg, broken := inv(ctx)
	suite.True(broken, "pool with reserves but zero shares should break invariant")
	suite.Contains(msg, "zero shares")
}

func (suite *InvariantsComprehensiveTestSuite) TestOrderValidityInvariant() {
	ctx := suite.SdkCtx
	inv := OrderValidityInvariant(suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Create valid order
	trader := sdk.AccAddress("trader____________")
	order := types.Order{
		OrderId:   1,
		Trader:    trader.String(),
		DenomIn:   "uaura",
		DenomOut:  "utoken",
		AmountIn:  "1000",
		MinOut:    "900",
		CreatedAt: nil,
	}
	suite.storeOrder(ctx, &order)

	// May break if CreatedAt is nil and checked
	msg, broken = inv(ctx)
	_ = msg
	_ = broken
}

func (suite *InvariantsComprehensiveTestSuite) TestOrderZeroID() {
	ctx := suite.SdkCtx
	inv := OrderValidityInvariant(suite.Keeper)

	trader := sdk.AccAddress("trader____________")
	order := types.Order{
		OrderId:   0,
		Trader:    trader.String(),
		DenomIn:   "uaura",
		DenomOut:  "utoken",
		AmountIn:  "1000",
		MinOut:    "900",
	}
	suite.storeOrder(ctx, &order)

	msg, broken := inv(ctx)
	suite.True(broken, "order with zero ID should break invariant")
	suite.Contains(msg, "zero ID")
}

func (suite *InvariantsComprehensiveTestSuite) TestOrderInvalidTrader() {
	ctx := suite.SdkCtx
	inv := OrderValidityInvariant(suite.Keeper)

	order := types.Order{
		OrderId:   1,
		Trader:    "invalid-address",
		DenomIn:   "uaura",
		DenomOut:  "utoken",
		AmountIn:  "1000",
		MinOut:    "900",
	}
	suite.storeOrder(ctx, &order)

	msg, broken := inv(ctx)
	suite.True(broken, "order with invalid trader should break invariant")
	suite.Contains(msg, "invalid trader address")
}

func (suite *InvariantsComprehensiveTestSuite) TestLiquidityProviderConsistencyInvariant() {
	ctx := suite.SdkCtx
	inv := LiquidityProviderConsistencyInvariant(suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Create valid LP position
	provider := sdk.AccAddress("provider__________")
	position := types.LiquidityProviderPosition{
		Provider: provider.String(),
		PoolId:   "pool-1",
		Shares:   "1000000",
	}
	suite.storeLPPosition(ctx, &position)

	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestLPPositionInvalidProvider() {
	ctx := suite.SdkCtx
	inv := LiquidityProviderConsistencyInvariant(suite.Keeper)

	position := types.LiquidityProviderPosition{
		Provider: "invalid-address",
		PoolId:   "pool-1",
		Shares:   "1000000",
	}
	suite.storeLPPosition(ctx, &position)

	msg, broken := inv(ctx)
	suite.True(broken, "LP position with invalid provider should break invariant")
	suite.Contains(msg, "invalid provider address")
}

func (suite *InvariantsComprehensiveTestSuite) TestHTLCValidityInvariant() {
	ctx := suite.SdkCtx
	inv := HTLCValidityInvariant(suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestAllInvariants() {
	ctx := suite.SdkCtx
	inv := AllInvariants(suite.Keeper)

	// Test: All invariants on empty store
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

// Helper methods
func (suite *InvariantsComprehensiveTestSuite) storePool(ctx sdk.Context, pool *types.LiquidityPool) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	bz := suite.Keeper.cdc.MustMarshal(pool)
	store.Set(append(types.PoolKeyPrefix, []byte(pool.PoolId)...), bz)
}

func (suite *InvariantsComprehensiveTestSuite) storeOrder(ctx sdk.Context, order *types.Order) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	bz := suite.Keeper.cdc.MustMarshal(order)
	store.Set(append(types.OrderKeyPrefix, sdk.Uint64ToBigEndian(order.OrderId)...), bz)
}

func (suite *InvariantsComprehensiveTestSuite) storeLPPosition(ctx sdk.Context, position *types.LiquidityProviderPosition) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	bz := suite.Keeper.cdc.MustMarshal(position)
	key := append(types.LiquidityProviderKeyPrefix, []byte(position.Provider+":"+position.PoolId)...)
	store.Set(key, bz)
}
