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
	k, ctx := setupTestKeeper(t)
	user := keepertest.GenTestAddr().String()

	_, err := k.CreateOrder(ctx, user, types.SwapOrderType_BUY, math.NewInt(100), "usdt", math.NewInt(200), 60)
	require.NoError(t, err)

	server := keeper.NewQueryServerImpl(k)
	resp, err := server.UserOrders(sdk.WrapSDKContext(ctx), &dexpb.QueryUserOrdersRequest{Address: user})
	require.NoError(t, err)
	require.Len(t, resp.Orders, 1)
	require.Equal(t, user, resp.Orders[0].UserAddress)
}

func TestQueryMarketPriceUsesStoredValue(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	user := keepertest.GenTestAddr().String()

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
	k, ctx := setupTestKeeper(t)
	user := keepertest.GenTestAddr().String()

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
