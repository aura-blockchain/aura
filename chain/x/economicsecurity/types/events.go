package types

import "fmt"

// Event types for the economicsecurity module
const (
	EventTypeFeeAdjusted             = "fee_adjusted"
	EventTypeMEVDetected             = "mev_detected"
	EventTypeMEVPrevented            = "mev_prevented"
	EventTypeWhaleLimitTriggered     = "whale_limit_triggered"
	EventTypeCircuitBreakerTriggered = "circuit_breaker_triggered"
	EventTypeCircuitBreakerReset     = "circuit_breaker_reset"
	EventTypeCongestionDetected      = "congestion_detected"
	EventTypeParamsUpdated           = "params_updated"
	EventTypeInflationAdjusted       = "inflation_adjusted"
)

// Event attribute keys
const (
	AttributeKeyModuleName        = "module_name"
	AttributeKeyAddress           = "address"
	AttributeKeyOldFee            = "old_fee"
	AttributeKeyNewFee            = "new_fee"
	AttributeKeyFeeMultiplier     = "fee_multiplier"
	AttributeKeyBaseFee           = "base_fee"
	AttributeKeyCongestionLevel   = "congestion_level"
	AttributeKeyTxHash            = "tx_hash"
	AttributeKeyMEVType           = "mev_type"
	AttributeKeyProtectionMethod  = "protection_method"
	AttributeKeyTransactionAmount = "transaction_amount"
	AttributeKeyLimit             = "limit"
	AttributeKeyTimeWindow        = "time_window_seconds"
	AttributeKeyCurrentUsage      = "current_usage"
	AttributeKeyThreshold         = "threshold"
	AttributeKeyReason            = "reason"
	AttributeKeyCooldownSeconds   = "cooldown_seconds"
	AttributeKeyBlockHeight       = "block_height"
	AttributeKeyBlockTime         = "block_time"
	AttributeKeyTimestamp         = "timestamp"
	AttributeKeyOldRate           = "old_rate"
	AttributeKeyNewRate           = "new_rate"
	AttributeKeyAuthority         = "authority"
)

// Helper functions for creating event attributes

// NewFeeAdjustedEvent creates attributes for fee adjustment
func NewFeeAdjustedEvent(
	moduleName string,
	oldFee, newFee string,
	feeMultiplier string,
	congestionLevel uint32,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyModuleName:      moduleName,
		AttributeKeyOldFee:          oldFee,
		AttributeKeyNewFee:          newFee,
		AttributeKeyFeeMultiplier:   feeMultiplier,
		AttributeKeyCongestionLevel: formatUint32(congestionLevel),
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewMEVDetectedEvent creates attributes for MEV detection
func NewMEVDetectedEvent(
	txHash, mevType string,
	address string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyTxHash:      txHash,
		AttributeKeyMEVType:     mevType,
		AttributeKeyAddress:     address,
		AttributeKeyBlockHeight: formatInt64(blockHeight),
		AttributeKeyBlockTime:   blockTime,
	}
}

// NewMEVPreventedEvent creates attributes for MEV prevention
func NewMEVPreventedEvent(
	txHash, mevType, protectionMethod string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyTxHash:           txHash,
		AttributeKeyMEVType:          mevType,
		AttributeKeyProtectionMethod: protectionMethod,
		AttributeKeyBlockHeight:      formatInt64(blockHeight),
		AttributeKeyBlockTime:        blockTime,
	}
}

// NewWhaleLimitTriggeredEvent creates attributes for whale limit trigger
func NewWhaleLimitTriggeredEvent(
	address, transactionAmount, limit string,
	currentUsage string,
	timeWindow int64,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAddress:           address,
		AttributeKeyTransactionAmount: transactionAmount,
		AttributeKeyLimit:             limit,
		AttributeKeyCurrentUsage:      currentUsage,
		AttributeKeyTimeWindow:        formatInt64(timeWindow),
		AttributeKeyBlockHeight:       formatInt64(blockHeight),
		AttributeKeyBlockTime:         blockTime,
	}
}

// NewCircuitBreakerTriggeredEvent creates attributes for circuit breaker trigger
func NewCircuitBreakerTriggeredEvent(
	moduleName, reason string,
	threshold uint64,
	cooldownSeconds int64,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyModuleName:      moduleName,
		AttributeKeyReason:          reason,
		AttributeKeyThreshold:       formatUint64(threshold),
		AttributeKeyCooldownSeconds: formatInt64(cooldownSeconds),
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewCircuitBreakerResetEvent creates attributes for circuit breaker reset
func NewCircuitBreakerResetEvent(
	moduleName string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyModuleName:  moduleName,
		AttributeKeyBlockHeight: formatInt64(blockHeight),
		AttributeKeyBlockTime:   blockTime,
	}
}

// NewCongestionDetectedEvent creates attributes for congestion detection
func NewCongestionDetectedEvent(
	congestionLevel uint32,
	newFeeMultiplier string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyCongestionLevel: formatUint32(congestionLevel),
		AttributeKeyFeeMultiplier:   newFeeMultiplier,
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// Helper formatting functions

func formatInt64(i int64) string {
	return fmt.Sprintf("%d", i)
}

func formatUint32(u uint32) string {
	return fmt.Sprintf("%d", u)
}

func formatUint64(u uint64) string {
	return fmt.Sprintf("%d", u)
}
