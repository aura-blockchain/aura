package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dex/types"
)

func TestGenesisRoundTrip(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	params := types.DefaultParams()
	params.TradingFee = "0.001"
	require.NoError(t, k.SetParams(ctx, params))

	creator := keepertest.GenTestAddr().String()
	_, _, err := k.CreatePool(ctx, creator, "uaura", "usdt", sdk.NewCoin("uaura", math.NewInt(1_000000)), sdk.NewCoin("usdt", math.NewInt(2_000000)))
	require.NoError(t, err)

	order, err := k.CreateOrder(ctx, creator, types.SwapOrderType_BUY, math.NewInt(100), "usdt", math.NewInt(200), 30)
	require.NoError(t, err)
	require.NotNil(t, order)

	k.RecordSwapStats(ctx, "uaura-usdt", math.NewInt(1000), math.NewInt(500), ctx.BlockTime())

	gen := k.ExportGenesis(ctx)
	require.NotNil(t, gen.Params)
	require.Len(t, gen.LiquidityPools, 1)
	require.Len(t, gen.SwapOrders, 1)
	require.NotEmpty(t, gen.Orderbooks)
	require.NotEmpty(t, gen.MarketPrices)
	require.NotEmpty(t, gen.SwapStats)

	k2, ctx2 := setupTestKeeper(t)
	require.NoError(t, k2.InitGenesis(ctx2, gen))

	reParams := k2.GetParams(ctx2)
	require.Equal(t, params.TradingFee, reParams.TradingFee)

	pools := k2.GetAllPools(ctx2)
	require.Len(t, pools, 1)

	userOrders := k2.GetOrdersByUser(ctx2, creator)
	require.Len(t, userOrders, 1)
	require.Equal(t, types.SwapOrderStatus_PENDING, userOrders[0].Status)

	orderbook := k2.GetOrderbookForPair(ctx2, "uaura", "usdt")
	require.NotEmpty(t, orderbook)

	stats, found := k2.GetSwapStats(ctx2, "uaura-usdt")
	require.True(t, found)
	require.Equal(t, "uaura-usdt", stats.PoolId)

	price, found := k2.GetMarketPrice(ctx2, "usdt")
	require.True(t, found)
	require.Equal(t, "usdt", price.Coin)
}
