package types

import "encoding/binary"

const (
	// ModuleName defines the module name
	ModuleName = "dex"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// KVStore key prefixes
var (
	// Liquidity Pool keys
	PoolPrefix = []byte{0x01}

	// Order keys
	OrderPrefix = []byte{0x02}

	// Orderbook index keys
	OrderbookPrefix = []byte{0x03}

	// HTLC keys
	HTLCPrefix = []byte{0x04}

	// Atomic Swap keys
	AtomicSwapPrefix = []byte{0x05}

	// User order history keys
	UserOrderPrefix = []byte{0x06}

	// Swap stats keys
	SwapStatsPrefix = []byte{0x07}

	// Market price keys
	MarketPricePrefix = []byte{0x08}

	// Order commitment keys (commit-reveal scheme)
	OrderCommitmentPrefix = []byte{0x09}

	// Queued order keys (batch execution)
	QueuedOrderPrefix = []byte{0x0A}
)

// PoolKey returns the store key for a liquidity pool
func PoolKey(poolID string) []byte {
	return append(PoolPrefix, []byte(poolID)...)
}

// OrderKey returns the store key for an order
func OrderKey(orderID string) []byte {
	return append(OrderPrefix, []byte(orderID)...)
}

// OrderbookKey returns the store key for an orderbook entry
func OrderbookKey(pairKey, orderID string) []byte {
	key := append(OrderbookPrefix, []byte(pairKey)...)
	key = append(key, byte(0x00)) // separator
	key = append(key, []byte(orderID)...)
	return key
}

// OrderbookPairPrefix returns the prefix for all orders in a trading pair
func OrderbookPairPrefix(pairKey string) []byte {
	return append(OrderbookPrefix, []byte(pairKey)...)
}

// HTLCKey returns the store key for an HTLC
func HTLCKey(htlcID string) []byte {
	return append(HTLCPrefix, []byte(htlcID)...)
}

// AtomicSwapKey returns the store key for an atomic swap
func AtomicSwapKey(swapID string) []byte {
	return append(AtomicSwapPrefix, []byte(swapID)...)
}

// UserOrderAddressPrefix returns the prefix for all orders created by an address.
func UserOrderAddressPrefix(address string) []byte {
	key := append(UserOrderPrefix, []byte(address)...)
	key = append(key, byte(0x00))
	return key
}

// UserOrderKey creates a sortable key for a user's order (newest first).
func UserOrderKey(address string, timestamp uint64, orderID string) []byte {
	key := UserOrderAddressPrefix(address)

	invertedTs := ^timestamp
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, invertedTs)

	key = append(key, tsBytes...)
	key = append(key, []byte(orderID)...)
	return key
}

// SwapStatsKey returns the key for swap stats associated with a pool.
func SwapStatsKey(poolID string) []byte {
	return append(SwapStatsPrefix, []byte(poolID)...)
}

// MarketPriceKey returns the key for stored market prices.
func MarketPriceKey(poolID string) []byte {
	return append(MarketPricePrefix, []byte(poolID)...)
}

// OrderCommitmentKey returns the store key for an order commitment
func OrderCommitmentKey(commitID string) []byte {
	return append(OrderCommitmentPrefix, []byte(commitID)...)
}

// QueuedOrderKey returns the store key for a queued order
func QueuedOrderKey(orderID string) []byte {
	return append(QueuedOrderPrefix, []byte(orderID)...)
}
