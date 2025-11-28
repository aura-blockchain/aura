package types

import "fmt"

// Event types for the bridge module
const (
	EventTypeTransferInitiated       = "transfer_initiated"
	EventTypeTransferCompleted       = "transfer_completed"
	EventTypeTransferFailed          = "transfer_failed"
	EventTypeAssetLocked             = "asset_locked"
	EventTypeAssetUnlocked           = "asset_unlocked"
	EventTypeMerkleProofVerified     = "merkle_proof_verified"
	EventTypeValidatorAdded          = "validator_added"
	EventTypeValidatorRemoved        = "validator_removed"
	EventTypeValidatorPowerUpdated   = "validator_power_updated"
	EventTypeCircuitBreakerTriggered = "circuit_breaker_triggered"
	EventTypeCircuitBreakerReset     = "circuit_breaker_reset"
	EventTypeSecurityViolationDetected = "security_violation_detected"
	EventTypeChannelOpened           = "channel_opened"
	EventTypeChannelClosed           = "channel_closed"
	EventTypeParamsUpdated           = "params_updated"
)

// Event attribute keys
const (
	AttributeKeyTransferID        = "transfer_id"
	AttributeKeySender            = "sender"
	AttributeKeyRecipient         = "recipient"
	AttributeKeyAmount            = "amount"
	AttributeKeyDenom             = "denom"
	AttributeKeySourceChain       = "source_chain"
	AttributeKeyDestinationChain  = "destination_chain"
	AttributeKeyChannelID         = "channel_id"
	AttributeKeyStatus            = "status"
	AttributeKeyProofHash         = "proof_hash"
	AttributeKeyBlockHeight       = "block_height"
	AttributeKeyBlockHash         = "block_hash"
	AttributeKeyConfirmations     = "confirmations"
	AttributeKeyValidatorAddress  = "validator_address"
	AttributeKeyValidatorPower    = "validator_power"
	AttributeKeyActiveValidators  = "active_validators"
	AttributeKeyTotalPower        = "total_power"
	AttributeKeyCircuitBreakerReason = "circuit_breaker_reason"
	AttributeKeyViolationType     = "violation_type"
	AttributeKeyViolationDetails  = "violation_details"
	AttributeKeyFailureReason     = "failure_reason"
	AttributeKeyTimestamp         = "timestamp"
	AttributeKeyBlockTime         = "block_time"
	AttributeKeyTransactionHash   = "transaction_hash"
	AttributeKeyProofDepth        = "proof_depth"
	AttributeKeyMerkleRoot        = "merkle_root"
	AttributeKeyLeafHash          = "leaf_hash"
	AttributeKeyVerifier          = "verifier"
)

// Helper functions for creating event attributes

// NewTransferInitiatedEvent creates attributes for transfer initiation
func NewTransferInitiatedEvent(
	transferID, sender, recipient, amount, denom string,
	sourceChain, destinationChain, channelID string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyTransferID:       transferID,
		AttributeKeySender:           sender,
		AttributeKeyRecipient:        recipient,
		AttributeKeyAmount:           amount,
		AttributeKeyDenom:            denom,
		AttributeKeySourceChain:      sourceChain,
		AttributeKeyDestinationChain: destinationChain,
		AttributeKeyChannelID:        channelID,
		AttributeKeyStatus:           "initiated",
		AttributeKeyBlockHeight:      formatInt64(blockHeight),
		AttributeKeyBlockTime:        blockTime,
	}
}

// NewTransferCompletedEvent creates attributes for transfer completion
func NewTransferCompletedEvent(
	transferID, recipient, amount, denom string,
	confirmations uint32,
	blockHeight int64, blockTime, txHash string,
) map[string]string {
	return map[string]string{
		AttributeKeyTransferID:      transferID,
		AttributeKeyRecipient:       recipient,
		AttributeKeyAmount:          amount,
		AttributeKeyDenom:           denom,
		AttributeKeyConfirmations:   formatUint32(confirmations),
		AttributeKeyStatus:          "completed",
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
		AttributeKeyTransactionHash: txHash,
	}
}

// NewTransferFailedEvent creates attributes for transfer failure
func NewTransferFailedEvent(
	transferID, sender, reason string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyTransferID:    transferID,
		AttributeKeySender:        sender,
		AttributeKeyFailureReason: reason,
		AttributeKeyStatus:        "failed",
		AttributeKeyBlockHeight:   formatInt64(blockHeight),
		AttributeKeyBlockTime:     blockTime,
	}
}

// NewAssetLockedEvent creates attributes for asset locking
func NewAssetLockedEvent(
	transferID, sender, amount, denom string,
	channelID string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyTransferID:  transferID,
		AttributeKeySender:      sender,
		AttributeKeyAmount:      amount,
		AttributeKeyDenom:       denom,
		AttributeKeyChannelID:   channelID,
		AttributeKeyBlockHeight: formatInt64(blockHeight),
		AttributeKeyBlockTime:   blockTime,
	}
}

// NewAssetUnlockedEvent creates attributes for asset unlocking
func NewAssetUnlockedEvent(
	transferID, recipient, amount, denom string,
	channelID string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyTransferID:  transferID,
		AttributeKeyRecipient:   recipient,
		AttributeKeyAmount:      amount,
		AttributeKeyDenom:       denom,
		AttributeKeyChannelID:   channelID,
		AttributeKeyBlockHeight: formatInt64(blockHeight),
		AttributeKeyBlockTime:   blockTime,
	}
}

// NewMerkleProofVerifiedEvent creates attributes for merkle proof verification
func NewMerkleProofVerifiedEvent(
	transferID, merkleRoot, leafHash string,
	proofDepth uint32,
	verifier string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyTransferID:  transferID,
		AttributeKeyMerkleRoot:  merkleRoot,
		AttributeKeyLeafHash:    leafHash,
		AttributeKeyProofDepth:  formatUint32(proofDepth),
		AttributeKeyVerifier:    verifier,
		AttributeKeyBlockHeight: formatInt64(blockHeight),
		AttributeKeyBlockTime:   blockTime,
	}
}

// NewValidatorAddedEvent creates attributes for validator addition
func NewValidatorAddedEvent(
	validatorAddress string,
	power int64,
	activeValidators uint32,
	totalPower int64,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyValidatorAddress: validatorAddress,
		AttributeKeyValidatorPower:   formatInt64(power),
		AttributeKeyActiveValidators: formatUint32(activeValidators),
		AttributeKeyTotalPower:       formatInt64(totalPower),
		AttributeKeyBlockHeight:      formatInt64(blockHeight),
		AttributeKeyBlockTime:        blockTime,
	}
}

// NewValidatorRemovedEvent creates attributes for validator removal
func NewValidatorRemovedEvent(
	validatorAddress string,
	activeValidators uint32,
	totalPower int64,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyValidatorAddress: validatorAddress,
		AttributeKeyActiveValidators: formatUint32(activeValidators),
		AttributeKeyTotalPower:       formatInt64(totalPower),
		AttributeKeyBlockHeight:      formatInt64(blockHeight),
		AttributeKeyBlockTime:        blockTime,
	}
}

// NewValidatorPowerUpdatedEvent creates attributes for validator power update
func NewValidatorPowerUpdatedEvent(
	validatorAddress string,
	oldPower, newPower int64,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyValidatorAddress: validatorAddress,
		"old_power":                  formatInt64(oldPower),
		"new_power":                  formatInt64(newPower),
		AttributeKeyBlockHeight:      formatInt64(blockHeight),
		AttributeKeyBlockTime:        blockTime,
	}
}

// NewCircuitBreakerTriggeredEvent creates attributes for circuit breaker trigger
func NewCircuitBreakerTriggeredEvent(
	channelID, reason string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyChannelID:            channelID,
		AttributeKeyCircuitBreakerReason: reason,
		AttributeKeyBlockHeight:          formatInt64(blockHeight),
		AttributeKeyBlockTime:            blockTime,
	}
}

// NewCircuitBreakerResetEvent creates attributes for circuit breaker reset
func NewCircuitBreakerResetEvent(
	channelID string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyChannelID:   channelID,
		AttributeKeyBlockHeight: formatInt64(blockHeight),
		AttributeKeyBlockTime:   blockTime,
	}
}

// NewSecurityViolationDetectedEvent creates attributes for security violation
func NewSecurityViolationDetectedEvent(
	violationType, details string,
	transferID string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyViolationType:    violationType,
		AttributeKeyViolationDetails: details,
		AttributeKeyTransferID:       transferID,
		AttributeKeyBlockHeight:      formatInt64(blockHeight),
		AttributeKeyBlockTime:        blockTime,
	}
}

// NewChannelOpenedEvent creates attributes for channel opening
func NewChannelOpenedEvent(
	channelID, sourceChain, destinationChain string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyChannelID:        channelID,
		AttributeKeySourceChain:      sourceChain,
		AttributeKeyDestinationChain: destinationChain,
		AttributeKeyBlockHeight:      formatInt64(blockHeight),
		AttributeKeyBlockTime:        blockTime,
	}
}

// NewChannelClosedEvent creates attributes for channel closing
func NewChannelClosedEvent(
	channelID, reason string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyChannelID:   channelID,
		"close_reason":          reason,
		AttributeKeyBlockHeight: formatInt64(blockHeight),
		AttributeKeyBlockTime:   blockTime,
	}
}

// NewParamsUpdatedEvent creates attributes for params update
func NewParamsUpdatedEvent(
	updatedBy string,
	confirmationDepth, timeoutPeriod uint32,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		"updated_by":         updatedBy,
		"confirmation_depth": formatUint32(confirmationDepth),
		"timeout_period":     formatUint32(timeoutPeriod),
		AttributeKeyBlockHeight: formatInt64(blockHeight),
		AttributeKeyBlockTime:   blockTime,
	}
}

// Helper formatting functions

func formatInt64(i int64) string {
	return fmt.Sprintf("%d", i)
}

func formatUint32(u uint32) string {
	return fmt.Sprintf("%d", u)
}
