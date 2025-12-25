# DEX GetOrdersByStatus Performance Optimization

## Summary
Fixed O(n) performance issue in `GetOrdersByStatus` by adding status-based composite index for O(k) lookup.

## Problem
`GetOrdersByStatus` loaded ALL orders into memory then filtered by status - O(n) full table scan.

## Solution
Added composite index: `OrderStatusPrefix + status (1 byte) + orderID`

## Changes

### 1. chain/x/dex/types/keys.go
- Added `OrderStatusPrefix = []byte{0x0F}` constant
- Added `OrderStatusKey(status, orderID)` helper
- Added `OrderStatusPrefixByStatus(status)` helper

### 2. chain/x/dex/keeper/orderbook.go
- **SetOrder**: Maintains status index on create/update
  - Removes old index entry when status changes
  - Adds new index entry
- **DeleteOrder**: Removes status index entry before deletion
- **GetOrdersByStatus**: Rewritten to use prefix iterator - O(k) where k = orders with status
- Added helpers: `addOrderToStatusIndex`, `removeOrderFromStatusIndex`

### 3. chain/x/dex/keeper/orderbook_status_index_test.go
- `TestGetOrdersByStatus_WithStatusIndex`: Verifies O(k) lookup
- `TestStatusIndexUpdate_OnStatusChange`: Verifies index updates
- `TestStatusIndexCleanup_OnDelete`: Verifies cleanup

## Performance Impact
- **Before**: O(n) - scans all orders
- **After**: O(k) - only reads orders with target status
- **Example**: 10,000 total orders, 50 PENDING → 200x faster

## Test Results
All tests pass including existing reentrancy and cleanup tests.
