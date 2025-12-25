// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// TestGetOrdersByStatus_WithStatusIndex verifies that GetOrdersByStatus correctly
// uses the status index for efficient O(k) lookups instead of O(n) full table scans
func TestGetOrdersByStatus_WithStatusIndex(t *testing.T) {
	f := SetupKeeperTest(t)
	k, ctx := f.Keeper, f.Ctx

	// Create multiple orders with different statuses
	orders := []*types.SwapOrder{
		{
			OrderId:      "order-1",
			OrderType:    types.SwapOrderType_SELL,
			AuraAmount:   sdkmath.NewInt(1000),
			OtherCoin:    "usdt",
			OtherAmount:  sdkmath.NewInt(100),
			UserAddress:  "aura1test1",
			Status:       types.SwapOrderStatus_PENDING,
			Timestamp:    ctx.BlockTime(),
			ExpiresAt:    ctx.BlockTime().Add(3600),
			PricePerAura: sdkmath.LegacyNewDec(1),
		},
		{
			OrderId:      "order-2",
			OrderType:    types.SwapOrderType_BUY,
			AuraAmount:   sdkmath.NewInt(2000),
			OtherCoin:    "usdt",
			OtherAmount:  sdkmath.NewInt(200),
			UserAddress:  "aura1test2",
			Status:       types.SwapOrderStatus_PENDING,
			Timestamp:    ctx.BlockTime(),
			ExpiresAt:    ctx.BlockTime().Add(3600),
			PricePerAura: sdkmath.LegacyNewDec(1),
		},
		{
			OrderId:      "order-3",
			OrderType:    types.SwapOrderType_SELL,
			AuraAmount:   sdkmath.NewInt(3000),
			OtherCoin:    "usdt",
			OtherAmount:  sdkmath.NewInt(300),
			UserAddress:  "aura1test3",
			Status:       types.SwapOrderStatus_COMPLETED,
			Timestamp:    ctx.BlockTime(),
			ExpiresAt:    ctx.BlockTime().Add(3600),
			PricePerAura: sdkmath.LegacyNewDec(1),
		},
		{
			OrderId:      "order-4",
			OrderType:    types.SwapOrderType_BUY,
			AuraAmount:   sdkmath.NewInt(4000),
			OtherCoin:    "usdt",
			OtherAmount:  sdkmath.NewInt(400),
			UserAddress:  "aura1test4",
			Status:       types.SwapOrderStatus_CANCELLED,
			Timestamp:    ctx.BlockTime(),
			ExpiresAt:    ctx.BlockTime().Add(3600),
			PricePerAura: sdkmath.LegacyNewDec(1),
		},
		{
			OrderId:      "order-5",
			OrderType:    types.SwapOrderType_SELL,
			AuraAmount:   sdkmath.NewInt(5000),
			OtherCoin:    "usdt",
			OtherAmount:  sdkmath.NewInt(500),
			UserAddress:  "aura1test5",
			Status:       types.SwapOrderStatus_PENDING,
			Timestamp:    ctx.BlockTime(),
			ExpiresAt:    ctx.BlockTime().Add(3600),
			PricePerAura: sdkmath.LegacyNewDec(1),
		},
	}

	// Store all orders
	for _, order := range orders {
		err := k.SetOrder(ctx, order)
		require.NoError(t, err)
	}

	// Test 1: Get pending orders - should return 3 orders (order-1, order-2, order-5)
	pendingOrders := k.GetOrdersByStatus(ctx, types.SwapOrderStatus_PENDING)
	require.Len(t, pendingOrders, 3)

	// Verify we got the right orders
	pendingIDs := make(map[string]bool)
	for _, order := range pendingOrders {
		pendingIDs[order.OrderId] = true
		require.Equal(t, types.SwapOrderStatus_PENDING, order.Status)
	}
	require.True(t, pendingIDs["order-1"])
	require.True(t, pendingIDs["order-2"])
	require.True(t, pendingIDs["order-5"])

	// Test 2: Get completed orders - should return 1 order (order-3)
	completedOrders := k.GetOrdersByStatus(ctx, types.SwapOrderStatus_COMPLETED)
	require.Len(t, completedOrders, 1)
	require.Equal(t, "order-3", completedOrders[0].OrderId)
	require.Equal(t, types.SwapOrderStatus_COMPLETED, completedOrders[0].Status)

	// Test 3: Get cancelled orders - should return 1 order (order-4)
	cancelledOrders := k.GetOrdersByStatus(ctx, types.SwapOrderStatus_CANCELLED)
	require.Len(t, cancelledOrders, 1)
	require.Equal(t, "order-4", cancelledOrders[0].OrderId)
	require.Equal(t, types.SwapOrderStatus_CANCELLED, cancelledOrders[0].Status)

	// Test 4: Get matched orders - should return 0 orders
	matchedOrders := k.GetOrdersByStatus(ctx, types.SwapOrderStatus_MATCHED)
	require.Len(t, matchedOrders, 0)
}

// TestStatusIndexUpdate_OnStatusChange verifies that the status index is correctly
// updated when an order's status changes
func TestStatusIndexUpdate_OnStatusChange(t *testing.T) {
	f := SetupKeeperTest(t)
	k, ctx := f.Keeper, f.Ctx

	// Create an order with PENDING status
	order := &types.SwapOrder{
		OrderId:      "order-test",
		OrderType:    types.SwapOrderType_SELL,
		AuraAmount:   sdkmath.NewInt(1000),
		OtherCoin:    "usdt",
		OtherAmount:  sdkmath.NewInt(100),
		UserAddress:  "aura1test",
		Status:       types.SwapOrderStatus_PENDING,
		Timestamp:    ctx.BlockTime(),
		ExpiresAt:    ctx.BlockTime().Add(3600),
		PricePerAura: sdkmath.LegacyNewDec(1),
	}

	// Store order
	err := k.SetOrder(ctx, order)
	require.NoError(t, err)

	// Verify it appears in PENDING status
	pendingOrders := k.GetOrdersByStatus(ctx, types.SwapOrderStatus_PENDING)
	require.Len(t, pendingOrders, 1)
	require.Equal(t, "order-test", pendingOrders[0].OrderId)

	// Should not appear in COMPLETED status
	completedOrders := k.GetOrdersByStatus(ctx, types.SwapOrderStatus_COMPLETED)
	require.Len(t, completedOrders, 0)

	// Update order status to COMPLETED
	order.Status = types.SwapOrderStatus_COMPLETED
	err = k.SetOrder(ctx, order)
	require.NoError(t, err)

	// Verify it no longer appears in PENDING status
	pendingOrders = k.GetOrdersByStatus(ctx, types.SwapOrderStatus_PENDING)
	require.Len(t, pendingOrders, 0)

	// Verify it now appears in COMPLETED status
	completedOrders = k.GetOrdersByStatus(ctx, types.SwapOrderStatus_COMPLETED)
	require.Len(t, completedOrders, 1)
	require.Equal(t, "order-test", completedOrders[0].OrderId)

	// Update status again to CANCELLED
	order.Status = types.SwapOrderStatus_CANCELLED
	err = k.SetOrder(ctx, order)
	require.NoError(t, err)

	// Verify it no longer appears in COMPLETED status
	completedOrders = k.GetOrdersByStatus(ctx, types.SwapOrderStatus_COMPLETED)
	require.Len(t, completedOrders, 0)

	// Verify it now appears in CANCELLED status
	cancelledOrders := k.GetOrdersByStatus(ctx, types.SwapOrderStatus_CANCELLED)
	require.Len(t, cancelledOrders, 1)
	require.Equal(t, "order-test", cancelledOrders[0].OrderId)
}

// TestStatusIndexCleanup_OnDelete verifies that the status index entry is removed
// when an order is deleted
func TestStatusIndexCleanup_OnDelete(t *testing.T) {
	f := SetupKeeperTest(t)
	k, ctx := f.Keeper, f.Ctx

	// Create an order
	order := &types.SwapOrder{
		OrderId:      "order-delete",
		OrderType:    types.SwapOrderType_SELL,
		AuraAmount:   sdkmath.NewInt(1000),
		OtherCoin:    "usdt",
		OtherAmount:  sdkmath.NewInt(100),
		UserAddress:  "aura1test",
		Status:       types.SwapOrderStatus_PENDING,
		Timestamp:    ctx.BlockTime(),
		ExpiresAt:    ctx.BlockTime().Add(3600),
		PricePerAura: sdkmath.LegacyNewDec(1),
	}

	// Store order
	err := k.SetOrder(ctx, order)
	require.NoError(t, err)

	// Verify it appears in PENDING status
	pendingOrders := k.GetOrdersByStatus(ctx, types.SwapOrderStatus_PENDING)
	require.Len(t, pendingOrders, 1)

	// Delete the order
	k.DeleteOrder(ctx, "order-delete")

	// Verify it no longer appears in PENDING status
	pendingOrders = k.GetOrdersByStatus(ctx, types.SwapOrderStatus_PENDING)
	require.Len(t, pendingOrders, 0)

	// Verify the order is actually deleted
	deletedOrder := k.GetOrder(ctx, "order-delete")
	require.Nil(t, deletedOrder)
}
