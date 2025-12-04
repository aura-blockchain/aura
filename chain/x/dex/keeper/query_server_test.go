package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dex/keeper"
	"github.com/aequitas/aura/chain/x/dex/types"
	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

func TestQueryUserOrders(t *testing.T) {
	k, ctx, mockBank := setupTestKeeperWithMock(t)

	// Set up mock bank keeper with sufficient balance
	userAddr := keepertest.GenTestAddr()
	user := userAddr.String()

	// BUY order needs usdt balance
	mockBank.SetBalance(userAddr, "usdt", math.NewInt(200))

	_, err := k.CreateOrder(ctx, user, types.SwapOrderType_BUY, math.NewInt(100), "usdt", math.NewInt(200), 60)
	require.NoError(t, err)

	server := keeper.NewQueryServerImpl(k)
	resp, err := server.UserOrders(sdk.WrapSDKContext(ctx), &dexpb.QueryUserOrdersRequest{Address: user})
	require.NoError(t, err)
	require.Len(t, resp.Orders, 1)
	require.Equal(t, user, resp.Orders[0].UserAddress)
}

func TestQueryMarketPriceUsesStoredValue(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)
	userAddr := keepertest.GenTestAddr()
	user := userAddr.String()

	// Fund the user account with sufficient balance for pool creation
	mockBank.SetBalance(userAddr, "uaura", math.NewInt(1_000000))
	mockBank.SetBalance(userAddr, "usdt", math.NewInt(2_000000))

	_, _, err := k.CreatePool(ctx, user, "uaura", "usdt", sdk.NewCoin("uaura", math.NewInt(1_000000)), sdk.NewCoin("usdt", math.NewInt(2_000000)))
	require.NoError(t, err)

	k.RecordSwapStats(ctx, "uaura-usdt", math.NewInt(1000), math.NewInt(500), ctx.BlockTime())

	server := keeper.NewQueryServerImpl(k)
	resp, err := server.MarketPrice(sdk.WrapSDKContext(ctx), &dexpb.QueryMarketPriceRequest{Coin: "usdt"})
	require.NoError(t, err)
	require.NotNil(t, resp.Price)
	require.Equal(t, "usdt", resp.Price.Coin)
	require.Equal(t, uint64(1), resp.Price.SampleSize)
}

func TestQuerySpotPrice(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)
	userAddr := keepertest.GenTestAddr()
	user := userAddr.String()

	// Fund the user account with sufficient balance for pool creation
	mockBank.SetBalance(userAddr, "uaura", math.NewInt(1_000000))
	mockBank.SetBalance(userAddr, "usdt", math.NewInt(2_000000))

	_, _, err := k.CreatePool(ctx, user, "uaura", "usdt", sdk.NewCoin("uaura", math.NewInt(1_000000)), sdk.NewCoin("usdt", math.NewInt(2_000000)))
	require.NoError(t, err)

	server := keeper.NewQueryServerImpl(k)
	resp, err := server.SpotPrice(sdk.WrapSDKContext(ctx), &dexpb.QuerySpotPriceRequest{
		PoolId:     "uaura-usdt",
		BaseDenom:  "uaura",
		QuoteDenom: "usdt",
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Price)
}

// TestQueryEmptyOrderbook tests that querying an empty orderbook does not panic
// Regression test for nil pointer dereference bug
func TestQueryEmptyOrderbook(t *testing.T) {
	k, ctx, _ := setupTestKeeper(t)

	server := keeper.NewQueryServerImpl(k)

	// Query orderbook with no orders (empty state)
	resp, err := server.Orderbook(sdk.WrapSDKContext(ctx), &dexpb.QueryOrderbookRequest{
		Pair: "AURA/STAKE",
	})

	// Should not panic and should return empty orderbook
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Orderbook)

	// Verify empty orderbook structure
	require.Equal(t, "aura-stake", resp.Orderbook.Pair)
	require.Empty(t, resp.Orderbook.BuyOrders)
	require.Empty(t, resp.Orderbook.SellOrders)
	require.Equal(t, uint64(0), resp.Orderbook.TotalPending)

	// Verify best bid/ask are zero (not nil)
	require.NotEmpty(t, resp.Orderbook.BestBid)
	require.NotEmpty(t, resp.Orderbook.BestAsk)
	require.Equal(t, "0.000000000000000000", resp.Orderbook.BestBid)
	require.Equal(t, "0.000000000000000000", resp.Orderbook.BestAsk)

	// Spread should be empty or "0" when no orders exist
	require.NotNil(t, resp.Orderbook.SpreadPercent)
}
