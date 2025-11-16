package keeper

import (
	"context"

	dextypes "github.com/aequitas/aura/chain/x/dex/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ dextypes.QueryServer = queryServer{}

// queryServer implements the QueryServer interface
type queryServer struct {
	types.UnimplementedQueryServer
	Keeper *Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper *Keeper) dextypes.QueryServer {
	return &queryServer{Keeper: keeper}
}

// Pool queries a liquidity pool by ID
func (qs queryServer) Pool(goCtx context.Context, req *dextypes.QueryPoolRequest) (*dextypes.QueryPoolResponse, error) {
	if req == nil {
		return nil, dextypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pool, err := qs.Keeper.GetPool(ctx, req.PoolId)
	if err != nil {
		return nil, err
	}

	return &dextypes.QueryPoolResponse{Pool: *pool}, nil
}

// AllPools queries all liquidity pools
func (qs queryServer) AllPools(goCtx context.Context, req *dextypes.QueryAllPoolsRequest) (*dextypes.QueryAllPoolsResponse, error) {
	if req == nil {
		return nil, dextypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pools := qs.Keeper.GetAllPools(ctx)

	return &dextypes.QueryAllPoolsResponse{Pools: pools}, nil
}

// GetQuote gets swap quote without executing
func (qs queryServer) GetQuote(goCtx context.Context, req *dextypes.QueryGetQuoteRequest) (*dextypes.QueryGetQuoteResponse, error) {
	if req == nil {
		return nil, dextypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	estimatedOutput, effectivePrice, priceImpact, err := qs.Keeper.GetSwapQuote(ctx, req.PoolId, req.DenomIn, req.AmountIn)
	if err != nil {
		return nil, err
	}

	return &dextypes.QueryGetQuoteResponse{
		EstimatedOutput:    estimatedOutput,
		EffectivePrice:     effectivePrice,
		PriceImpactPercent: priceImpact,
	}, nil
}

// PoolStats queries pool statistics
func (qs queryServer) PoolStats(goCtx context.Context, req *dextypes.QueryPoolStatsRequest) (*dextypes.QueryPoolStatsResponse, error) {
	if req == nil {
		return nil, dextypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	stats, err := qs.Keeper.GetPoolStats(ctx, req.PoolId)
	if err != nil {
		return nil, err
	}

	return &dextypes.QueryPoolStatsResponse{
		TotalLiquidity: stats.TotalLiquidity,
		Volume24H:      stats.Volume24H,
		Fees24H:        stats.Fees24H,
		AprPercent:     stats.AprPercent,
		LpTokenSupply:  stats.LpTokenSupply,
	}, nil
}

// Orderbook queries P2P orderbook
func (qs queryServer) Orderbook(goCtx context.Context, req *dextypes.QueryOrderbookRequest) (*dextypes.QueryOrderbookResponse, error) {
	if req == nil {
		return nil, dextypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	buyOrders, sellOrders := qs.Keeper.GetOrderbook(ctx, req.Pair)

	return &dextypes.QueryOrderbookResponse{
		BuyOrders:  buyOrders,
		SellOrders: sellOrders,
	}, nil
}

// Order queries specific order
func (qs queryServer) Order(goCtx context.Context, req *dextypes.QueryOrderRequest) (*dextypes.QueryOrderResponse, error) {
	if req == nil {
		return nil, dextypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	order, err := qs.Keeper.GetOrder(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}

	return &dextypes.QueryOrderResponse{Order: *order}, nil
}

// UserOrders queries all orders for user
func (qs queryServer) UserOrders(goCtx context.Context, req *dextypes.QueryUserOrdersRequest) (*dextypes.QueryUserOrdersResponse, error) {
	if req == nil {
		return nil, dextypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	orders := qs.Keeper.GetUserOrders(ctx, req.Address)

	return &dextypes.QueryUserOrdersResponse{Orders: orders}, nil
}

// MarketPrice queries current market price
func (qs queryServer) MarketPrice(goCtx context.Context, req *dextypes.QueryMarketPriceRequest) (*dextypes.QueryMarketPriceResponse, error) {
	if req == nil {
		return nil, dextypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	price, err := qs.Keeper.GetMarketPrice(ctx, req.Coin)
	if err != nil {
		return nil, err
	}

	return &dextypes.QueryMarketPriceResponse{
		Price:     price.Price,
		Volume24H: price.Volume24H,
	}, nil
}

// SupportedCoins queries list of supported altcoins
func (qs queryServer) SupportedCoins(goCtx context.Context, req *dextypes.QuerySupportedCoinsRequest) (*dextypes.QuerySupportedCoinsResponse, error) {
	if req == nil {
		return nil, dextypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	coins := qs.Keeper.GetSupportedCoins(ctx)

	return &dextypes.QuerySupportedCoinsResponse{Coins: coins}, nil
}

// HTLC queries Hash Time-Locked Contract
func (qs queryServer) HTLC(goCtx context.Context, req *dextypes.QueryHTLCRequest) (*dextypes.QueryHTLCResponse, error) {
	if req == nil {
		return nil, dextypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	htlc, err := qs.Keeper.GetHTLC(ctx, req.HtlcId)
	if err != nil {
		return nil, err
	}

	return &dextypes.QueryHTLCResponse{Htlc: *htlc}, nil
}
