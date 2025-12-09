package keeper

import (
	"context"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/cosmos/cosmos-sdk/types/query"
	metrics "github.com/hashicorp/go-metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

type queryServer struct {
	keeper *Keeper
	dexpb.UnimplementedQueryServer
}

// NewQueryServerImpl returns a QueryServer backed by the keeper.
func NewQueryServerImpl(k *Keeper) dexpb.QueryServer {
	return &queryServer{keeper: k}
}

var _ dexpb.QueryServer = (*queryServer)(nil)

func (qs queryServer) Pool(ctx context.Context, req *dexpb.QueryPoolRequest) (*dexpb.QueryPoolResponse, error) {
	if req == nil || req.PoolId == "" {
		return nil, status.Error(codes.InvalidArgument, "pool id required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	pool := qs.keeper.GetPool(sdkCtx, req.PoolId)
	if pool == nil {
		return nil, status.Error(codes.NotFound, "pool not found")
	}

	return &dexpb.QueryPoolResponse{Pool: pool}, nil
}

func (qs queryServer) AllPools(ctx context.Context, req *dexpb.QueryAllPoolsRequest) (*dexpb.QueryAllPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)
	poolStore := prefix.NewStore(store, types.PoolPrefix)

	pools := []*dexpb.LiquidityPool{}
	pageRes, err := query.Paginate(poolStore, req.Pagination, func(key []byte, value []byte) error {
		var pool dexpb.LiquidityPool
		if err := qs.keeper.cdc.Unmarshal(value, &pool); err != nil {
			return err
		}
		pools = append(pools, &pool)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &dexpb.QueryAllPoolsResponse{
		Pools:      pools,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) GetQuote(ctx context.Context, req *dexpb.QueryGetQuoteRequest) (*dexpb.QueryGetQuoteResponse, error) {
	if req == nil || req.PoolId == "" {
		return nil, status.Error(codes.InvalidArgument, "pool id required")
	}
	if req.DenomIn == "" || req.AmountIn == "" {
		return nil, status.Error(codes.InvalidArgument, "input denom and amount required")
	}

	amountIn, ok := sdkmath.NewIntFromString(req.AmountIn)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	estimated, effective, impact, fee, err := qs.keeper.GetQuote(sdkCtx, req.PoolId, req.DenomIn, amountIn)
	if err != nil {
		return nil, err
	}

	return &dexpb.QueryGetQuoteResponse{
		EstimatedOutput:    estimated.String(),
		EffectivePrice:     effective.String(),
		PriceImpactPercent: impact.String(),
		FeeAmount:          fee.String(),
	}, nil
}

func (qs queryServer) PoolStats(ctx context.Context, req *dexpb.QueryPoolStatsRequest) (*dexpb.QueryPoolStatsResponse, error) {
	if req == nil || req.PoolId == "" {
		return nil, status.Error(codes.InvalidArgument, "pool id required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	pool := qs.keeper.GetPool(sdkCtx, req.PoolId)
	if pool == nil {
		return nil, status.Error(codes.NotFound, "pool not found")
	}

	reserveA, err := qs.keeper.parseReserve(pool.ReserveA)
	if err != nil {
		return nil, err
	}
	reserveB, err := qs.keeper.parseReserve(pool.ReserveB)
	if err != nil {
		return nil, err
	}
	price := reserveB.ToLegacyDec().Quo(reserveA.ToLegacyDec())

	return &dexpb.QueryPoolStatsResponse{
		PoolId:                  pool.PoolId,
		DenomA:                  pool.DenomA,
		DenomB:                  pool.DenomB,
		ReserveA:                pool.ReserveA,
		ReserveB:                pool.ReserveB,
		TotalLpTokens:           pool.TotalLpTokens,
		LiquidityProvidersCount: uint64(len(pool.Providers)),
		CurrentPrice:            price.String(),
		TotalVolume:             pool.TotalVolume,
		TotalFeesCollected:      pool.TotalFeesCollected,
		SwapCount:               pool.SwapCount,
		TvlA:                    pool.ReserveA,
		TvlB:                    pool.ReserveB,
	}, nil
}

func (qs queryServer) Orderbook(ctx context.Context, req *dexpb.QueryOrderbookRequest) (*dexpb.QueryOrderbookResponse, error) {
	if req == nil || req.Pair == "" {
		return nil, status.Error(codes.InvalidArgument, "pair required")
	}

	base, quote := parsePair(req.Pair)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	orders := qs.keeper.GetOrderbookForPair(sdkCtx, base, quote)

	orderbook := &dexpb.Orderbook{
		Pair:         fmt.Sprintf("%s-%s", base, quote),
		BuyOrders:    []*dexpb.SwapOrder{},
		SellOrders:   []*dexpb.SwapOrder{},
		TotalPending: uint64(len(orders)),
	}

	// Initialize to zero to prevent nil pointer dereference on empty orderbook
	bestBid := sdkmath.LegacyZeroDec()
	bestAsk := sdkmath.LegacyZeroDec()
	firstBid := true
	firstAsk := true

	for _, order := range orders {
		price := orderPriceDec(order)
		order.PricePerAura = price.String()

		if order.OrderType == dexpb.SwapOrderType_BUY {
			orderbook.BuyOrders = append(orderbook.BuyOrders, order)
			if firstBid || price.GT(bestBid) {
				bestBid = price
				firstBid = false
			}
		} else {
			orderbook.SellOrders = append(orderbook.SellOrders, order)
			if firstAsk || price.LT(bestAsk) || bestAsk.IsZero() {
				bestAsk = price
				firstAsk = false
			}
		}
	}

	sort.Slice(orderbook.BuyOrders, func(i, j int) bool {
		left := orderPriceDec(orderbook.BuyOrders[i])
		right := orderPriceDec(orderbook.BuyOrders[j])
		return left.GT(right)
	})
	sort.Slice(orderbook.SellOrders, func(i, j int) bool {
		left := orderPriceDec(orderbook.SellOrders[i])
		right := orderPriceDec(orderbook.SellOrders[j])
		return left.LT(right)
	})

	orderbook.BestBid = bestBid.String()
	orderbook.BestAsk = bestAsk.String()

	if !bestAsk.IsZero() && !bestBid.IsZero() {
		spread := bestAsk.Sub(bestBid).Quo(bestAsk).MulInt64(100)
		orderbook.SpreadPercent = spread.String()
	}

	return &dexpb.QueryOrderbookResponse{Orderbook: orderbook}, nil
}

func (qs queryServer) Order(ctx context.Context, req *dexpb.QueryOrderRequest) (*dexpb.QueryOrderResponse, error) {
	if req == nil || req.OrderId == "" {
		qs.logError(ctx, "order", "order id required")
		return nil, status.Error(codes.InvalidArgument, "order id required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	order := qs.keeper.GetOrder(sdkCtx, req.OrderId)
	if order == nil {
		qs.logError(ctx, "order", "order not found")
		return nil, status.Error(codes.NotFound, "order not found")
	}

	return &dexpb.QueryOrderResponse{Order: order}, nil
}

func (qs queryServer) UserOrders(ctx context.Context, req *dexpb.QueryUserOrdersRequest) (*dexpb.QueryUserOrdersResponse, error) {
	if req == nil || req.Address == "" {
		qs.logError(ctx, "user_orders", "address required")
		return nil, status.Error(codes.InvalidArgument, "address required")
	}
	if _, err := sdk.AccAddressFromBech32(req.Address); err != nil {
		qs.logError(ctx, "user_orders", "invalid address")
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)
	userOrderStore := prefix.NewStore(store, types.UserOrderAddressPrefix(req.Address))

	orders := []*dexpb.SwapOrder{}
	pageRes, err := query.Paginate(userOrderStore, req.Pagination, func(key []byte, value []byte) error {
		// The value stored is just the order ID reference, we need to fetch the actual order
		// Extract order ID from the key (it's after the timestamp bytes)
		if len(key) >= 8 {
			orderID := string(key[8:])
			order := qs.keeper.GetOrder(sdkCtx, orderID)
			if order != nil {
				orders = append(orders, order)
			}
		}
		return nil
	})
	if err != nil {
		qs.logError(ctx, "user_orders", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &dexpb.QueryUserOrdersResponse{
		Orders:     orders,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) MarketPrice(ctx context.Context, req *dexpb.QueryMarketPriceRequest) (*dexpb.QueryMarketPriceResponse, error) {
	if req == nil || req.Coin == "" {
		qs.logError(ctx, "market_price", "coin required")
		return nil, status.Error(codes.InvalidArgument, "coin required")
	}

	coin := strings.ToLower(req.Coin)
	poolID := qs.keeper.GeneratePoolID("uaura", coin)

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if price, found := qs.keeper.GetMarketPrice(sdkCtx, coin); found {
		return &dexpb.QueryMarketPriceResponse{Price: price}, nil
	}

	pool := qs.keeper.GetPool(sdkCtx, poolID)
	if pool == nil {
		qs.logError(ctx, "market_price", "pool not found")
		return nil, status.Error(codes.NotFound, "pool not found")
	}

	reserveA, err := qs.keeper.parseReserve(pool.ReserveA)
	if err != nil {
		return nil, err
	}
	reserveB, err := qs.keeper.parseReserve(pool.ReserveB)
	if err != nil {
		return nil, err
	}

	priceAura := reserveA.ToLegacyDec().Quo(reserveB.ToLegacyDec())
	priceUSD := sdkmath.LegacyZeroDec()
	if coin == "usdt" || coin == "usdc" {
		priceUSD = sdkmath.LegacyOneDec()
	}

	return &dexpb.QueryMarketPriceResponse{
		Price: &dexpb.MarketPrice{
			Coin:       coin,
			PriceUsd:   priceUSD.String(),
			PriceAura:  priceAura.String(),
			UpdatedAt:  timestamppb.New(sdkCtx.BlockTime()),
			SampleSize: pool.SwapCount,
		},
	}, nil
}

func (qs queryServer) SpotPrice(ctx context.Context, req *dexpb.QuerySpotPriceRequest) (*dexpb.QuerySpotPriceResponse, error) {
	if req == nil || req.PoolId == "" || req.BaseDenom == "" || req.QuoteDenom == "" {
		qs.logError(ctx, "spot_price", "pool id, base denom, and quote denom required")
		return nil, status.Error(codes.InvalidArgument, "pool id, base denom, and quote denom required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	price, err := qs.keeper.GetSpotPrice(sdkCtx, req.PoolId, strings.ToLower(req.BaseDenom), strings.ToLower(req.QuoteDenom))
	if err != nil {
		qs.logError(ctx, "spot_price", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &dexpb.QuerySpotPriceResponse{Price: price.String()}, nil
}

func (qs queryServer) SupportedCoins(ctx context.Context, _ *dexpb.QuerySupportedCoinsRequest) (*dexpb.QuerySupportedCoinsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	pools := qs.keeper.GetAllPools(sdkCtx)
	coins := make(map[string]struct{})

	for _, pool := range pools {
		if strings.ToLower(pool.DenomA) != "uaura" {
			coins[strings.ToLower(pool.DenomA)] = struct{}{}
		}
		if strings.ToLower(pool.DenomB) != "uaura" {
			coins[strings.ToLower(pool.DenomB)] = struct{}{}
		}
	}

	list := make([]string, 0, len(coins))
	for denom := range coins {
		list = append(list, denom)
	}

	sort.Strings(list)
	return &dexpb.QuerySupportedCoinsResponse{Coins: list}, nil
}

func (qs queryServer) HTLC(ctx context.Context, req *dexpb.QueryHTLCRequest) (*dexpb.QueryHTLCResponse, error) {
	if req == nil || req.HtlcId == "" {
		qs.logError(ctx, "htlc", "htlc id required")
		return nil, status.Error(codes.InvalidArgument, "htlc id required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	htlc, found := qs.keeper.GetHTLC(sdkCtx, req.HtlcId)
	if !found {
		qs.logError(ctx, "htlc", "htlc not found")
		return nil, status.Error(codes.NotFound, "htlc not found")
	}

	return &dexpb.QueryHTLCResponse{Htlc: htlc}, nil
}

func parsePair(pair string) (string, string) {
	pair = strings.ReplaceAll(pair, "/", "-")
	parts := strings.Split(pair, "-")
	if len(parts) == 2 {
		return strings.ToLower(parts[0]), strings.ToLower(parts[1])
	}
	return "uaura", strings.ToLower(pair)

}

func (qs queryServer) logError(ctx context.Context, op, reason string) {
	if sdkCtx, ok := unwrapSDKContextSafe(ctx); ok {
		sdkCtx.Logger().Error("dex query validation failed", "query", op, "reason", reason)
	}
	telemetry.IncrCounterWithLabels(
		[]string{"dex", "query", "validation_failed"},
		float32(1),
		[]metrics.Label{telemetry.NewLabel("query", op)},
	)
}

func unwrapSDKContextSafe(ctx context.Context) (sdk.Context, bool) {
	var (
		sdkCtx sdk.Context
		ok     = true
	)
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	sdkCtx = sdk.UnwrapSDKContext(ctx)
	return sdkCtx, ok
}
