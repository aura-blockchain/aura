package keeper

import (
	"crypto/sha256"
	"testing"
	"time"

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

	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	pool := types.LiquidityPool{
		PoolId:        "pool-1",
		DenomA:        "uaura",
		DenomB:        "utoken",
		ReserveA:      sdkmath.NewInt(1000000),
		ReserveB:      sdkmath.NewInt(2000000),
		TotalLpTokens: sdkmath.NewInt(1414213),
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
		PoolId:        "",
		DenomA:        "uaura",
		DenomB:        "utoken",
		ReserveA:      sdkmath.NewInt(1000000),
		ReserveB:      sdkmath.NewInt(2000000),
		TotalLpTokens: sdkmath.NewInt(1414213),
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
		PoolId:        "pool-2",
		DenomA:        "",
		DenomB:        "utoken",
		ReserveA:      sdkmath.NewInt(1000000),
		ReserveB:      sdkmath.NewInt(2000000),
		TotalLpTokens: sdkmath.NewInt(1414213),
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
		PoolId:        "pool-3",
		DenomA:        "uaura",
		DenomB:        "utoken",
		ReserveA:      sdkmath.NewInt(-1000000),
		ReserveB:      sdkmath.NewInt(2000000),
		TotalLpTokens: sdkmath.NewInt(1414213),
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
		PoolId:        "pool-4",
		DenomA:        "uaura",
		DenomB:        "utoken",
		ReserveA:      sdkmath.NewInt(1000000),
		ReserveB:      sdkmath.NewInt(2000000),
		TotalLpTokens: sdkmath.ZeroInt(),
	}
	suite.storePool(ctx, &pool)

	msg, broken := inv(ctx)
	suite.True(broken, "pool with reserves but zero shares should break invariant")
	suite.Contains(msg, "LP tokens")
}

func (suite *InvariantsComprehensiveTestSuite) TestOrderValidityInvariant() {
	ctx := suite.SdkCtx
	inv := OrderValidityInvariant(suite.Keeper)

	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	order := types.SwapOrder{
		OrderId:      "order-1",
		OrderType:    types.SwapOrderType_BUY,
		AuraAmount:   sdkmath.NewInt(1000),
		OtherCoin:    "utoken",
		OtherAmount:  sdkmath.NewInt(900),
		UserAddress:  suite.testAddr("trader"),
		Status:       types.SwapOrderStatus_PENDING,
		Timestamp:    ctx.BlockTime(),
		ExpiresAt:    ctx.BlockTime().Add(time.Hour),
		PricePerAura: sdkmath.LegacyMustNewDecFromStr("0.9"),
	}
	suite.storeOrder(ctx, &order)

	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestOrderZeroID() {
	ctx := suite.SdkCtx
	inv := OrderValidityInvariant(suite.Keeper)

	order := types.SwapOrder{
		OrderId:      "",
		OrderType:    types.SwapOrderType_BUY,
		AuraAmount:   sdkmath.NewInt(1000),
		OtherCoin:    "utoken",
		OtherAmount:  sdkmath.NewInt(900),
		UserAddress:  suite.testAddr("trader-zero"),
		Timestamp:    ctx.BlockTime(),
		ExpiresAt:    ctx.BlockTime().Add(time.Hour),
		PricePerAura: sdkmath.LegacyMustNewDecFromStr("0.9"),
	}
	suite.storeOrder(ctx, &order)

	msg, broken := inv(ctx)
	suite.True(broken, "order with empty ID should break invariant")
	suite.Contains(msg, "empty ID")
}

func (suite *InvariantsComprehensiveTestSuite) TestOrderInvalidTrader() {
	ctx := suite.SdkCtx
	inv := OrderValidityInvariant(suite.Keeper)

	order := types.SwapOrder{
		OrderId:      "order-2",
		OrderType:    types.SwapOrderType_BUY,
		AuraAmount:   sdkmath.NewInt(1000),
		OtherCoin:    "utoken",
		OtherAmount:  sdkmath.NewInt(900),
		UserAddress:  "invalid-address",
		Timestamp:    ctx.BlockTime(),
		ExpiresAt:    ctx.BlockTime().Add(time.Hour),
		PricePerAura: sdkmath.LegacyMustNewDecFromStr("0.9"),
	}
	suite.storeOrder(ctx, &order)

	msg, broken := inv(ctx)
	suite.True(broken, "order with invalid trader should break invariant")
	suite.Contains(msg, "invalid user address")
}

func (suite *InvariantsComprehensiveTestSuite) TestOrderNilTimestamp() {
	ctx := suite.SdkCtx
	inv := OrderValidityInvariant(suite.Keeper)

	order := types.SwapOrder{
		OrderId:      "order-3",
		OrderType:    types.SwapOrderType_BUY,
		AuraAmount:   sdkmath.NewInt(1000),
		OtherCoin:    "utoken",
		OtherAmount:  sdkmath.NewInt(900),
		UserAddress:  suite.testAddr("no-timestamp"),
		Timestamp:    time.Time{},
		ExpiresAt:    ctx.BlockTime().Add(time.Hour),
		PricePerAura: sdkmath.LegacyMustNewDecFromStr("0.9"),
	}
	suite.storeOrder(ctx, &order)

	msg, broken := inv(ctx)
	suite.True(broken, "order with nil timestamp should break invariant")
	suite.Contains(msg, "nil timestamp")
}

func (suite *InvariantsComprehensiveTestSuite) TestLiquidityProviderConsistencyInvariant() {
	ctx := suite.SdkCtx
	inv := LiquidityProviderConsistencyInvariant(suite.Keeper)

	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	lpAddr := suite.testAddr("provider-1")
	pool := types.LiquidityPool{
		PoolId:        "pool-1",
		DenomA:        "uaura",
		DenomB:        "usdt",
		ReserveA:      sdkmath.NewInt(1000),
		ReserveB:      sdkmath.NewInt(2000),
		TotalLpTokens: sdkmath.NewInt(3000),
		Providers: []types.LiquidityProvider{
			{Address: lpAddr, LpTokens: sdkmath.NewInt(3000)},
		},
	}
	suite.storePool(ctx, &pool)

	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestLPProviderInvalidAddress() {
	ctx := suite.SdkCtx
	inv := LiquidityProviderConsistencyInvariant(suite.Keeper)

	pool := types.LiquidityPool{
		PoolId:        "pool-1",
		DenomA:        "uaura",
		DenomB:        "usdt",
		ReserveA:      sdkmath.NewInt(1000),
		ReserveB:      sdkmath.NewInt(2000),
		TotalLpTokens: sdkmath.NewInt(3000),
		Providers: []types.LiquidityProvider{
			{Address: "invalid-address", LpTokens: sdkmath.NewInt(3000)},
		},
	}
	suite.storePool(ctx, &pool)

	msg, broken := inv(ctx)
	suite.True(broken, "invalid provider address should break invariant")
	suite.Contains(msg, "provider address invalid")
}

func (suite *InvariantsComprehensiveTestSuite) TestLPProviderMismatchedTotals() {
	ctx := suite.SdkCtx
	inv := LiquidityProviderConsistencyInvariant(suite.Keeper)

	pool := types.LiquidityPool{
		PoolId:        "pool-2",
		DenomA:        "uaura",
		DenomB:        "usdc",
		ReserveA:      sdkmath.NewInt(5000),
		ReserveB:      sdkmath.NewInt(10000),
		TotalLpTokens: sdkmath.NewInt(5000),
		Providers: []types.LiquidityProvider{
			{Address: suite.testAddr("lp-1"), LpTokens: sdkmath.NewInt(2000)},
			{Address: suite.testAddr("lp-2"), LpTokens: sdkmath.NewInt(2000)},
		},
	}
	suite.storePool(ctx, &pool)

	msg, broken := inv(ctx)
	suite.True(broken, "provider totals should match pool total")
	suite.Contains(msg, "invariant violated")
}

func (suite *InvariantsComprehensiveTestSuite) TestHTLCValidityInvariant() {
	ctx := suite.SdkCtx
	inv := HTLCValidityInvariant(suite.Keeper)

	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestAllInvariants() {
	ctx := suite.SdkCtx
	inv := AllInvariants(suite.Keeper)

	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) storePool(ctx sdk.Context, pool *types.LiquidityPool) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	bz := suite.Keeper.cdc.MustMarshal(pool)
	store.Set(types.PoolKey(pool.PoolId), bz)
}

func (suite *InvariantsComprehensiveTestSuite) storeOrder(ctx sdk.Context, order *types.SwapOrder) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	bz := suite.Keeper.cdc.MustMarshal(order)
	store.Set(types.OrderKey(order.OrderId), bz)
}

func (suite *InvariantsComprehensiveTestSuite) testAddr(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	bech32, err := sdk.Bech32ifyAddressBytes("aura", sum[:20])
	suite.Require().NoError(err)
	return bech32
}
