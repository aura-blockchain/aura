package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
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
	require.Equal(t, "0.000000000000000000", resp.Orderbook.BestBid.String())
	require.Equal(t, "0.000000000000000000", resp.Orderbook.BestAsk.String())

	// Spread may be nil or zero when no orders exist
	if !resp.Orderbook.SpreadPercent.IsNil() {
		require.Equal(t, "0.000000000000000000", resp.Orderbook.SpreadPercent.String())
	}
}

// TestQueryAllPoolsPagination tests that AllPools query supports Cosmos SDK pagination
func TestQueryAllPoolsPagination(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)
	userAddr := keepertest.GenTestAddr()
	user := userAddr.String()

	// Create multiple pools for pagination testing
	poolPairs := []struct {
		denomA string
		denomB string
		amtA   int64
		amtB   int64
	}{
		{"uaura", "usdt", 1_000000, 2_000000},
		{"uaura", "usdc", 1_000000, 2_000000},
		{"uaura", "btc", 1_000000, 100},
	}

	for _, pair := range poolPairs {
		mockBank.SetBalance(userAddr, pair.denomA, math.NewInt(pair.amtA))
		mockBank.SetBalance(userAddr, pair.denomB, math.NewInt(pair.amtB))
		_, _, err := k.CreatePool(ctx, user, pair.denomA, pair.denomB,
			sdk.NewCoin(pair.denomA, math.NewInt(pair.amtA)),
			sdk.NewCoin(pair.denomB, math.NewInt(pair.amtB)))
		require.NoError(t, err)
	}

	server := keeper.NewQueryServerImpl(k)

	// Test 1: Query without pagination (should get default limit)
	resp1, err := server.AllPools(sdk.WrapSDKContext(ctx), &dexpb.QueryAllPoolsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp1)
	require.Len(t, resp1.Pools, 3)
	require.NotNil(t, resp1.Pagination)

	// Test 2: Query with limit=2
	resp2, err := server.AllPools(sdk.WrapSDKContext(ctx), &dexpb.QueryAllPoolsRequest{
		Pagination: &query.PageRequest{
			Limit: 2,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp2.Pools, 2)
	require.NotNil(t, resp2.Pagination)
	require.NotEmpty(t, resp2.Pagination.NextKey)

	// Test 3: Query next page using NextKey
	resp3, err := server.AllPools(sdk.WrapSDKContext(ctx), &dexpb.QueryAllPoolsRequest{
		Pagination: &query.PageRequest{
			Key:   resp2.Pagination.NextKey,
			Limit: 2,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp3.Pools, 1)
	require.NotNil(t, resp3.Pagination)

	// Test 4: Verify no duplicates across pages
	allPoolIDs := make(map[string]bool)
	for _, pool := range resp2.Pools {
		allPoolIDs[pool.PoolId] = true
	}
	for _, pool := range resp3.Pools {
		require.False(t, allPoolIDs[pool.PoolId], "duplicate pool found across pages")
	}
}

// TestQueryUserOrdersPagination tests that UserOrders query supports Cosmos SDK pagination
func TestQueryUserOrdersPagination(t *testing.T) {
	k, ctx, mockBank := setupTestKeeperWithMock(t)

	userAddr := keepertest.GenTestAddr()
	user := userAddr.String()

	// Set up sufficient balance for multiple orders
	mockBank.SetBalance(userAddr, "usdt", math.NewInt(10000))

	// Create multiple orders for pagination testing
	// Order IDs are unique per (user, block height), so advance height between creates
	baseHeight := ctx.BlockHeight()
	for i := 0; i < 5; i++ {
		ctx = ctx.WithBlockHeight(baseHeight + int64(i))
		_, err := k.CreateOrder(ctx, user, types.SwapOrderType_BUY, math.NewInt(100), "usdt", math.NewInt(200), 60)
		require.NoError(t, err)
	}

	server := keeper.NewQueryServerImpl(k)

	// Test 1: Query without pagination (should get default limit)
	resp1, err := server.UserOrders(sdk.WrapSDKContext(ctx), &dexpb.QueryUserOrdersRequest{
		Address: user,
	})
	require.NoError(t, err)
	require.NotNil(t, resp1)
	require.Len(t, resp1.Orders, 5)
	require.NotNil(t, resp1.Pagination)

	// Test 2: Query with limit=3
	resp2, err := server.UserOrders(sdk.WrapSDKContext(ctx), &dexpb.QueryUserOrdersRequest{
		Address: user,
		Pagination: &query.PageRequest{
			Limit: 3,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp2.Orders, 3)
	require.NotNil(t, resp2.Pagination)
	require.NotEmpty(t, resp2.Pagination.NextKey)

	// Test 3: Query next page using NextKey
	resp3, err := server.UserOrders(sdk.WrapSDKContext(ctx), &dexpb.QueryUserOrdersRequest{
		Address: user,
		Pagination: &query.PageRequest{
			Key:   resp2.Pagination.NextKey,
			Limit: 3,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp3.Orders, 2)
	require.NotNil(t, resp3.Pagination)

	// Test 4: Verify all orders belong to the correct user
	for _, order := range resp2.Orders {
		require.Equal(t, user, order.UserAddress)
	}
	for _, order := range resp3.Orders {
		require.Equal(t, user, order.UserAddress)
	}

	// Test 5: Verify no duplicates across pages
	allOrderIDs := make(map[string]bool)
	for _, order := range resp2.Orders {
		allOrderIDs[order.OrderId] = true
	}
	for _, order := range resp3.Orders {
		require.False(t, allOrderIDs[order.OrderId], "duplicate order found across pages")
	}
}
