package types

import (
	"encoding/binary"
	"fmt"
)

// Security feature key prefixes
var (
	TradeBlockPrefix        = []byte{0x10}
	TWAPPrefix              = []byte{0x11}
	LiquidityBlockPrefix    = []byte{0x12}
	LiquidityLockPrefix     = []byte{0x13}
	TradeHistoryPrefix      = []byte{0x14}
	PoolCreationPrefix      = []byte{0x15}
	CircuitBreakerPrefix    = []byte{0x16}
	WashTradePrefix         = []byte{0x17}
	OrderManipulationPrefix = []byte{0x18}
)

// TradeBlockKey returns the key for tracking last trade block
func TradeBlockKey(address string, poolID string) []byte {
	key := append(TradeBlockPrefix, []byte(address)...)
	key = append(key, byte(0x00)) // separator
	key = append(key, []byte(poolID)...)
	return key
}

// TWAPKey returns the key for a TWAP observation
func TWAPKey(poolID string, blockHeight int64) []byte {
	key := append(TWAPPrefix, []byte(poolID)...)
	key = append(key, byte(0x00)) // separator

	// Append block height as bytes
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, uint64(blockHeight))
	key = append(key, heightBytes...)

	return key
}

// TWAPPrefixKey returns the prefix for all TWAP observations for a pool
func TWAPPrefixKey(poolID string) []byte {
	key := append(TWAPPrefix, []byte(poolID)...)
	key = append(key, byte(0x00)) // separator
	return key
}

// LiquidityBlockKey returns the key for tracking last liquidity operation block
func LiquidityBlockKey(provider string, poolID string) []byte {
	key := append(LiquidityBlockPrefix, []byte(provider)...)
	key = append(key, byte(0x00)) // separator
	key = append(key, []byte(poolID)...)
	return key
}

// LiquidityLockKey returns the key for a liquidity lock
func LiquidityLockKey(provider string, poolID string) []byte {
	key := append(LiquidityLockPrefix, []byte(provider)...)
	key = append(key, byte(0x00)) // separator
	key = append(key, []byte(poolID)...)
	return key
}

// TradeHistoryKey returns the key for trade history
func TradeHistoryKey(address string) []byte {
	return append(TradeHistoryPrefix, []byte(address)...)
}

// PoolCreationKey returns the key for pool creation record
func PoolCreationKey(creator string) []byte {
	return append(PoolCreationPrefix, []byte(creator)...)
}

// CircuitBreakerKey returns the key for circuit breaker state
func CircuitBreakerKey() []byte {
	return CircuitBreakerPrefix
}

// WashTradeKey returns the key for wash trade detection
func WashTradeKey(address string, poolID string) []byte {
	key := append(WashTradePrefix, []byte(address)...)
	key = append(key, byte(0x00)) // separator
	key = append(key, []byte(poolID)...)
	return key
}

// OrderManipulationKey returns the key for order manipulation detection
func OrderManipulationKey(address string, poolID string) []byte {
	key := append(OrderManipulationPrefix, []byte(address)...)
	key = append(key, byte(0x00)) // separator
	key = append(key, []byte(poolID)...)
	return key
}

// FormatSecurityKey formats a security key for logging
func FormatSecurityKey(prefix []byte, parts ...string) string {
	return fmt.Sprintf("security/%x/%s", prefix, parts)
}
