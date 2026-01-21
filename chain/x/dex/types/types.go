// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import pb "github.com/aequitas/aura/proto/aura/dex/v1beta1"

// Re-export all proto types
type (
	// Enums
	SwapOrderType   = pb.SwapOrderType
	SwapOrderStatus = pb.SwapOrderStatus

	// Core types - Liquidity Pool
	LiquidityPool      = pb.LiquidityPool
	LiquidityProvider  = pb.LiquidityProvider
	LiquidityLock      = pb.LiquidityLock
	PoolPair           = pb.PoolPair
	MarketPrice        = pb.MarketPrice
	TWAPPrice          = pb.TWAPPrice
	CircuitBreaker     = pb.CircuitBreaker
	MinLiquidityTier   = pb.MinLiquidityTier
	PoolCreationRecord = pb.PoolCreationRecord

	// Core types - Swap
	SwapOrder       = pb.SwapOrder
	Orderbook       = pb.Orderbook
	HTLCData        = pb.HTLCData
	SwapStats       = pb.SwapStats
	TradeHistory    = pb.TradeHistory
	OrderCommitment = pb.OrderCommitment
	QueuedOrder     = pb.QueuedOrder

	// Core types - Security
	OrderManipulationDetection = pb.OrderManipulationDetection
	WashTradeDetection         = pb.WashTradeDetection
	SecurityParams             = pb.SecurityParams

	// Params and Genesis
	Params       = pb.Params
	GenesisState = pb.GenesisState

	// Message types - Liquidity Pool
	MsgCreatePool              = pb.MsgCreatePool
	MsgCreatePoolResponse      = pb.MsgCreatePoolResponse
	MsgAddLiquidity            = pb.MsgAddLiquidity
	MsgAddLiquidityResponse    = pb.MsgAddLiquidityResponse
	MsgRemoveLiquidity         = pb.MsgRemoveLiquidity
	MsgRemoveLiquidityResponse = pb.MsgRemoveLiquidityResponse

	// Message types - Swap
	MsgCreateOrder         = pb.MsgCreateOrder
	MsgCreateOrderResponse = pb.MsgCreateOrderResponse
	MsgCancelOrder         = pb.MsgCancelOrder
	MsgCancelOrderResponse = pb.MsgCancelOrderResponse
	MsgExecuteSwap         = pb.MsgExecuteSwap
	MsgExecuteSwapResponse = pb.MsgExecuteSwapResponse
	MsgSwapExactIn         = pb.MsgSwapExactIn
	MsgSwapExactInResponse = pb.MsgSwapExactInResponse
	MsgCreateHTLC          = pb.MsgCreateHTLC
	MsgCreateHTLCResponse  = pb.MsgCreateHTLCResponse
	MsgClaimHTLC           = pb.MsgClaimHTLC
	MsgClaimHTLCResponse   = pb.MsgClaimHTLCResponse
	MsgRefundHTLC          = pb.MsgRefundHTLC
	MsgRefundHTLCResponse  = pb.MsgRefundHTLCResponse

	// Query types
	QueryPoolRequest            = pb.QueryPoolRequest
	QueryPoolResponse           = pb.QueryPoolResponse
	QueryAllPoolsRequest        = pb.QueryAllPoolsRequest
	QueryAllPoolsResponse       = pb.QueryAllPoolsResponse
	QueryPoolStatsRequest       = pb.QueryPoolStatsRequest
	QueryPoolStatsResponse      = pb.QueryPoolStatsResponse
	QueryGetQuoteRequest        = pb.QueryGetQuoteRequest
	QueryGetQuoteResponse       = pb.QueryGetQuoteResponse
	QueryMarketPriceRequest     = pb.QueryMarketPriceRequest
	QueryMarketPriceResponse    = pb.QueryMarketPriceResponse
	QuerySupportedCoinsRequest  = pb.QuerySupportedCoinsRequest
	QuerySupportedCoinsResponse = pb.QuerySupportedCoinsResponse
	QueryOrderRequest           = pb.QueryOrderRequest
	QueryOrderResponse          = pb.QueryOrderResponse
	QueryUserOrdersRequest      = pb.QueryUserOrdersRequest
	QueryUserOrdersResponse     = pb.QueryUserOrdersResponse
	QueryOrderbookRequest       = pb.QueryOrderbookRequest
	QueryOrderbookResponse      = pb.QueryOrderbookResponse
	QuerySpotPriceRequest       = pb.QuerySpotPriceRequest
	QuerySpotPriceResponse      = pb.QuerySpotPriceResponse
	QueryHTLCRequest            = pb.QueryHTLCRequest
	QueryHTLCResponse           = pb.QueryHTLCResponse
	QueryParamsRequest          = pb.QueryParamsRequest
	QueryParamsResponse         = pb.QueryParamsResponse
)

// Re-export enum values for SwapOrderType
const (
	SwapOrderType_BUY  = pb.SwapOrderType_BUY
	SwapOrderType_SELL = pb.SwapOrderType_SELL
)

// Re-export enum values for SwapOrderStatus
const (
	SwapOrderStatus_PENDING      = pb.SwapOrderStatus_PENDING
	SwapOrderStatus_MATCHED      = pb.SwapOrderStatus_MATCHED
	SwapOrderStatus_HTLC_CREATED = pb.SwapOrderStatus_HTLC_CREATED
	SwapOrderStatus_COMPLETED    = pb.SwapOrderStatus_COMPLETED
	SwapOrderStatus_CANCELLED    = pb.SwapOrderStatus_CANCELLED
	SwapOrderStatus_EXPIRED      = pb.SwapOrderStatus_EXPIRED
)
