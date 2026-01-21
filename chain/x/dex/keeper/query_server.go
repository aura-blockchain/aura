// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"container/heap"
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

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

// Query limits to prevent DoS via unbounded iteration
const (
	// maxOrderbookOrders limits total orders processed per orderbook query
	maxOrderbookOrders = 10000
	// defaultOrderbookPrealloc is the default pre-allocation size for order slices
	defaultOrderbookPrealloc = 256
)

// orderHeap implements a min-heap for efficient top-N selection
type orderHeap struct {
	orders []dexpb.SwapOrder
	less   func(a, b dexpb.SwapOrder) bool
}

func (h orderHeap) Len() int           { return len(h.orders) }
func (h orderHeap) Less(i, j int) bool { return h.less(h.orders[i], h.orders[j]) }
func (h orderHeap) Swap(i, j int)      { h.orders[i], h.orders[j] = h.orders[j], h.orders[i] }
func (h *orderHeap) Push(x any)        { h.orders = append(h.orders, x.(dexpb.SwapOrder)) }
func (h *orderHeap) Pop() any {
	old := h.orders
	n := len(old)
	x := old[n-1]
	h.orders = old[0 : n-1]
	return x
}

// topNOrders returns the top N orders using a heap for O(n log k) complexity
// instead of O(n log n) for full sort. If limit is 0, returns all orders sorted.
func topNOrders(orders []dexpb.SwapOrder, limit int, less func(a, b dexpb.SwapOrder) bool) []dexpb.SwapOrder {
	if len(orders) == 0 {
		return orders
	}

	// If no limit or limit >= len, just sort all
	if limit <= 0 || limit >= len(orders) {
		sort.Slice(orders, func(i, j int) bool { return less(orders[i], orders[j]) })
		return orders
	}

	// Use a min-heap of size k to find top k elements in O(n log k)
	// We invert the comparison to get a min-heap that evicts the smallest
	h := &orderHeap{
		orders: make([]dexpb.SwapOrder, 0, limit),
		less:   func(a, b dexpb.SwapOrder) bool { return !less(a, b) }, // inverted
	}

	for i := range orders {
		if h.Len() < limit {
			heap.Push(h, orders[i])
		} else if less(orders[i], h.orders[0]) {
			// New order is better than worst in heap, replace
			h.orders[0] = orders[i]
			heap.Fix(h, 0)
		}
	}

	// Extract in reverse order and reverse at end
	result := make([]dexpb.SwapOrder, h.Len())
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(dexpb.SwapOrder)
	}
	return result
}

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

	return &dexpb.QueryPoolResponse{Pool: *pool}, nil
}

func (qs queryServer) AllPools(ctx context.Context, req *dexpb.QueryAllPoolsRequest) (*dexpb.QueryAllPoolsResponse, error) {
	if req == nil {
		req = &dexpb.QueryAllPoolsRequest{}
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)
	poolStore := prefix.NewStore(store, types.PoolPrefix)

	var pools []*dexpb.LiquidityPool
	pageRes, err := query.Paginate(poolStore, req.Pagination, func(key []byte, value []byte) error {
		var pool dexpb.LiquidityPool
		if err := qs.keeper.cdc.Unmarshal(value, &pool); err != nil {
			return fmt.Errorf("error in AllPools for LiquidityPool: %w", err)
		}
		pools = append(pools, &pool)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert []*LiquidityPool to []LiquidityPool (value slice)
	poolValues := make([]dexpb.LiquidityPool, len(pools))
	for i, pool := range pools {
		poolValues[i] = *pool
	}

	return &dexpb.QueryAllPoolsResponse{
		Pools:      poolValues,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) GetQuote(ctx context.Context, req *dexpb.QueryGetQuoteRequest) (*dexpb.QueryGetQuoteResponse, error) {
	if req == nil || req.PoolId == "" {
		return nil, status.Error(codes.InvalidArgument, "pool id required")
	}
	if req.DenomIn == "" || req.AmountIn.IsZero() {
		return nil, status.Error(codes.InvalidArgument, "input denom and amount required")
	}

	// AmountIn is already math.Int type (customtype in proto)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	estimated, effective, impact, fee, err := qs.keeper.GetQuote(sdkCtx, req.PoolId, req.DenomIn, req.AmountIn)
	if err != nil {
		return nil, err
	}

	return &dexpb.QueryGetQuoteResponse{
		EstimatedOutput:    estimated,
		EffectivePrice:     effective,
		PriceImpactPercent: impact,
		FeeAmount:          fee,
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
		CurrentPrice:            price,
		TotalVolume:             pool.TotalVolume,
		TotalFeesCollected:      pool.TotalFeesCollected,
		SwapCount:               pool.SwapCount,
		TvlA:                    pool.ReserveA,
		TvlB:                    pool.ReserveB,
	}, nil
}

// Orderbook queries the orderbook for a trading pair.
// Uses bounded iteration and smart pre-allocation to prevent DoS.
func (qs queryServer) Orderbook(ctx context.Context, req *dexpb.QueryOrderbookRequest) (*dexpb.QueryOrderbookResponse, error) {
	if req == nil || req.Pair == "" {
		return nil, status.Error(codes.InvalidArgument, "pair required")
	}

	base, quote := parsePair(req.Pair)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	orders := qs.keeper.GetOrderbookForPair(sdkCtx, base, quote)

	// Get limit from request (0 means default limit for safety)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 100 // Default limit when not specified
	}

	// Hard cap on total orders processed to prevent DoS
	totalOrders := len(orders)
	if totalOrders > maxOrderbookOrders {
		totalOrders = maxOrderbookOrders
	}

	// Smart pre-allocation: use limit if specified, otherwise use bounded default
	prealloc := limit
	if prealloc > defaultOrderbookPrealloc {
		prealloc = defaultOrderbookPrealloc
	}

	// Separate buy/sell orders and calculate best bid/ask in single pass
	buyOrders := make([]dexpb.SwapOrder, 0, prealloc)
	sellOrders := make([]dexpb.SwapOrder, 0, prealloc)
	bestBid := sdkmath.LegacyZeroDec()
	bestAsk := sdkmath.LegacyZeroDec()

	for i, order := range orders {
		// Hard cap on iteration
		if i >= totalOrders {
			break
		}

		price := orderPriceDec(order)
		// Create a copy and update PricePerAura
		orderCopy := *order
		orderCopy.PricePerAura = price

		if order.OrderType == dexpb.SwapOrderType_BUY {
			buyOrders = append(buyOrders, orderCopy)
			// Track best bid (highest price) for buy orders
			if bestBid.IsZero() || price.GT(bestBid) {
				bestBid = price
			}
		} else {
			sellOrders = append(sellOrders, orderCopy)
			// Track best ask (lowest price) for sell orders
			if bestAsk.IsZero() || price.LT(bestAsk) {
				bestAsk = price
			}
		}
	}

	// Use heap-based partial sort when limit is specified for O(n log k) instead of O(n log n)
	// Buy orders: highest price first (best bid at index 0)
	buyOrders = topNOrders(buyOrders, limit, func(a, b dexpb.SwapOrder) bool {
		return a.PricePerAura.GT(b.PricePerAura)
	})
	// Sell orders: lowest price first (best ask at index 0)
	sellOrders = topNOrders(sellOrders, limit, func(a, b dexpb.SwapOrder) bool {
		return a.PricePerAura.LT(b.PricePerAura)
	})

	orderbook := &dexpb.Orderbook{
		Pair:         fmt.Sprintf("%s-%s", base, quote),
		BuyOrders:    buyOrders,
		SellOrders:   sellOrders,
		TotalPending: uint64(len(orders)),
		BestBid:      bestBid,
		BestAsk:      bestAsk,
	}

	if !bestAsk.IsZero() && !bestBid.IsZero() {
		orderbook.SpreadPercent = bestAsk.Sub(bestBid).Quo(bestAsk).MulInt64(100)
	}

	return &dexpb.QueryOrderbookResponse{Orderbook: *orderbook}, nil
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

	// Dereference pointer for value type
	return &dexpb.QueryOrderResponse{Order: *order}, nil
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

	var orders []*dexpb.SwapOrder
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

	// Convert pointer slice to value slice
	orderValues := make([]dexpb.SwapOrder, len(orders))
	for i, order := range orders {
		orderValues[i] = *order
	}

	return &dexpb.QueryUserOrdersResponse{
		Orders:     orderValues,
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
		// Dereference pointer for value type
		return &dexpb.QueryMarketPriceResponse{Price: *price}, nil
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
		Price: dexpb.MarketPrice{
			Coin:       coin,
			PriceUsd:   priceUSD,
			PriceAura:  priceAura,
			UpdatedAt:  sdkCtx.BlockTime(),
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

	return &dexpb.QuerySpotPriceResponse{Price: price}, nil
}

func (qs queryServer) SupportedCoins(ctx context.Context, _ *dexpb.QuerySupportedCoinsRequest) (*dexpb.QuerySupportedCoinsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Use the pre-built index for O(k) lookup where k = number of supported coins
	// instead of O(n) where n = number of pools
	coins := qs.keeper.GetSupportedCoins(sdkCtx)
	return &dexpb.QuerySupportedCoinsResponse{Coins: coins}, nil
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

	// Dereference pointer for value type
	return &dexpb.QueryHTLCResponse{Htlc: *htlc}, nil
}

func (qs queryServer) Params(ctx context.Context, _ *dexpb.QueryParamsRequest) (*dexpb.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params, err := qs.keeper.GetParams(sdkCtx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get params: %s", err.Error())
	}

	return &dexpb.QueryParamsResponse{Params: params}, nil
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
