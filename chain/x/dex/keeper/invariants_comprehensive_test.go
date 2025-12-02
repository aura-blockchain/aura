package keeper

import (
	"crypto/sha256"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

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
		ReserveA:      "1000000",
		ReserveB:      "2000000",
		TotalLpTokens: "1414213",
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
		ReserveA:      "1000000",
		ReserveB:      "2000000",
		TotalLpTokens: "1414213",
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
		ReserveA:      "1000000",
		ReserveB:      "2000000",
		TotalLpTokens: "1414213",
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
		ReserveA:      "-1000000",
		ReserveB:      "2000000",
		TotalLpTokens: "1414213",
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
		ReserveA:      "1000000",
		ReserveB:      "2000000",
		TotalLpTokens: "0",
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
		AuraAmount:   "1000",
		OtherCoin:    "utoken",
		OtherAmount:  "900",
		UserAddress:  suite.testAddr("trader"),
		Status:       types.SwapOrderStatus_PENDING,
		Timestamp:    timestamppb.New(ctx.BlockTime()),
		ExpiresAt:    timestamppb.New(ctx.BlockTime().Add(time.Hour)),
		PricePerAura: "0.9",
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
		AuraAmount:   "1000",
		OtherCoin:    "utoken",
		OtherAmount:  "900",
		UserAddress:  suite.testAddr("trader-zero"),
		Timestamp:    timestamppb.New(ctx.BlockTime()),
		ExpiresAt:    timestamppb.New(ctx.BlockTime().Add(time.Hour)),
		PricePerAura: "0.9",
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
		AuraAmount:   "1000",
		OtherCoin:    "utoken",
		OtherAmount:  "900",
		UserAddress:  "invalid-address",
		Timestamp:    timestamppb.New(ctx.BlockTime()),
		ExpiresAt:    timestamppb.New(ctx.BlockTime().Add(time.Hour)),
		PricePerAura: "0.9",
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
		AuraAmount:   "1000",
		OtherCoin:    "utoken",
		OtherAmount:  "900",
		UserAddress:  suite.testAddr("no-timestamp"),
		Timestamp:    nil,
		ExpiresAt:    timestamppb.New(ctx.BlockTime().Add(time.Hour)),
		PricePerAura: "0.9",
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
		ReserveA:      "1000",
		ReserveB:      "2000",
		TotalLpTokens: "3000",
		Providers: []*types.LiquidityProvider{
			{Address: lpAddr, LpTokens: "3000"},
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
		ReserveA:      "1000",
		ReserveB:      "2000",
		TotalLpTokens: "3000",
		Providers: []*types.LiquidityProvider{
			{Address: "invalid-address", LpTokens: "3000"},
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
		ReserveA:      "5000",
		ReserveB:      "10000",
		TotalLpTokens: "5000",
		Providers: []*types.LiquidityProvider{
			{Address: suite.testAddr("lp-1"), LpTokens: "2000"},
			{Address: suite.testAddr("lp-2"), LpTokens: "2000"},
		},
	}
	suite.storePool(ctx, &pool)

	msg, broken := inv(ctx)
	suite.True(broken, "provider totals should match pool total")
	suite.Contains(msg, "do not match provider balances")
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
