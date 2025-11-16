package types

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
