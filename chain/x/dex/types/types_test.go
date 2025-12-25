// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
)

func TestSwapOrderTypeConstants(t *testing.T) {
	// Test that SwapOrderType constants are defined and different
	if SwapOrderType_BUY == SwapOrderType_SELL {
		t.Error("SwapOrderType_BUY and SwapOrderType_SELL should be different")
	}
}

func TestSwapOrderStatusConstants(t *testing.T) {
	// Test that all SwapOrderStatus constants are defined and unique
	statuses := []SwapOrderStatus{
		SwapOrderStatus_PENDING,
		SwapOrderStatus_MATCHED,
		SwapOrderStatus_HTLC_CREATED,
		SwapOrderStatus_COMPLETED,
		SwapOrderStatus_CANCELLED,
		SwapOrderStatus_EXPIRED,
	}

	// Check uniqueness
	seen := make(map[SwapOrderStatus]bool)
	for _, status := range statuses {
		if seen[status] {
			t.Errorf("duplicate SwapOrderStatus value: %v", status)
		}
		seen[status] = true
	}

	if len(seen) != len(statuses) {
		t.Error("not all SwapOrderStatus values are unique")
	}
}

func TestTypeExports(t *testing.T) {
	// Test that all type exports compile correctly
	var _ SwapOrderType
	var _ SwapOrderStatus
	var _ LiquidityPool
	var _ LiquidityProvider
	var _ LiquidityLock
	var _ PoolPair
	var _ MarketPrice
	var _ TWAPPrice
	var _ CircuitBreaker
	var _ MinLiquidityTier
	var _ PoolCreationRecord
	var _ SwapOrder
	var _ Orderbook
	var _ HTLCData
	var _ SwapStats
	var _ TradeHistory
	var _ OrderManipulationDetection
	var _ WashTradeDetection
	var _ SecurityParams
	var _ Params
	var _ GenesisState
}

func TestMessageTypeExports(t *testing.T) {
	// Test that all message type exports compile correctly
	var _ MsgCreatePool
	var _ MsgCreatePoolResponse
	var _ MsgAddLiquidity
	var _ MsgAddLiquidityResponse
	var _ MsgRemoveLiquidity
	var _ MsgRemoveLiquidityResponse
	var _ MsgCreateOrder
	var _ MsgCreateOrderResponse
	var _ MsgCancelOrder
	var _ MsgCancelOrderResponse
	var _ MsgExecuteSwap
	var _ MsgExecuteSwapResponse
	var _ MsgSwapExactIn
	var _ MsgSwapExactInResponse
	var _ MsgCreateHTLC
	var _ MsgCreateHTLCResponse
	var _ MsgClaimHTLC
	var _ MsgClaimHTLCResponse
	var _ MsgRefundHTLC
	var _ MsgRefundHTLCResponse
}

func TestQueryTypeExports(t *testing.T) {
	// Test that all query type exports compile correctly
	var _ QueryPoolRequest
	var _ QueryPoolResponse
	var _ QueryAllPoolsRequest
	var _ QueryAllPoolsResponse
	var _ QueryPoolStatsRequest
	var _ QueryPoolStatsResponse
	var _ QueryGetQuoteRequest
	var _ QueryGetQuoteResponse
	var _ QueryMarketPriceRequest
	var _ QueryMarketPriceResponse
	var _ QuerySupportedCoinsRequest
	var _ QuerySupportedCoinsResponse
	var _ QueryOrderRequest
	var _ QueryOrderResponse
	var _ QueryUserOrdersRequest
	var _ QueryUserOrdersResponse
	var _ QueryOrderbookRequest
	var _ QueryOrderbookResponse
	var _ QueryHTLCRequest
	var _ QueryHTLCResponse
}

func TestSwapOrderTypeValues(t *testing.T) {
	// Test that enum values can be compared
	buy := SwapOrderType_BUY
	sell := SwapOrderType_SELL

	if buy == sell {
		t.Error("BUY and SELL should have different values")
	}

	// Test that they are usable in comparisons
	orderType := buy
	if orderType != SwapOrderType_BUY {
		t.Error("should be able to compare orderType with constant")
	}
}

func TestSwapOrderStatusValues(t *testing.T) {
	// Test that status values can be compared
	pending := SwapOrderStatus_PENDING
	completed := SwapOrderStatus_COMPLETED

	if pending == completed {
		t.Error("PENDING and COMPLETED should have different values")
	}

	// Test that they are usable in comparisons
	status := pending
	if status != SwapOrderStatus_PENDING {
		t.Error("should be able to compare status with constant")
	}
}

func TestTypeAliasesWork(t *testing.T) {
	// This test ensures that the type aliases can be used
	// in variable declarations and assignments

	var orderType SwapOrderType = SwapOrderType_BUY
	if orderType != SwapOrderType_BUY {
		t.Error("type alias assignment should work")
	}

	var status SwapOrderStatus = SwapOrderStatus_PENDING
	if status != SwapOrderStatus_PENDING {
		t.Error("type alias assignment should work")
	}
}

func TestProtoTypeCompatibility(t *testing.T) {
	// Test that proto types can be instantiated
	// This ensures the proto import and re-export works correctly

	// Try to create instances of proto types
	var _ *Params
	var _ *GenesisState
	var _ *SecurityParams
	var _ *LiquidityPool
	var _ *SwapOrder
}
