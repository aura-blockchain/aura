// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import "fmt"

// Event types for the contractregistry module
const (
	EventTypeContractRegistered = "contract_registered"
	EventTypeContractUpdated    = "contract_updated"
	EventTypeContractVerified   = "contract_verified"
	EventTypeContractDeprecated = "contract_deprecated"
	EventTypeContractRemoved    = "contract_removed"
	EventTypeParamsUpdated      = "params_updated"
)

// Event attribute keys
const (
	AttributeKeyContractAddress = "contract_address"
	AttributeKeyContractName    = "contract_name"
	AttributeKeyContractVersion = "contract_version"
	AttributeKeyCreator         = "creator"
	AttributeKeyCodeHash        = "code_hash"
	AttributeKeyCodeID          = "code_id"
	AttributeKeyVerified        = "verified"
	AttributeKeyVerifier        = "verifier"
	AttributeKeyOldVersion      = "old_version"
	AttributeKeyNewVersion      = "new_version"
	AttributeKeyDeprecated      = "deprecated"
	AttributeKeyReason          = "reason"
	AttributeKeyReplacementAddr = "replacement_address"
	AttributeKeyBlockHeight     = "block_height"
	AttributeKeyBlockTime       = "block_time"
	AttributeKeyTimestamp       = "timestamp"
)

// Helper functions for creating event attributes

// NewContractRegisteredEvent creates attributes for contract registration
func NewContractRegisteredEvent(
	contractAddress, contractName, contractVersion, creator string,
	codeHash string,
	codeID uint64,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyContractAddress: contractAddress,
		AttributeKeyContractName:    contractName,
		AttributeKeyContractVersion: contractVersion,
		AttributeKeyCreator:         creator,
		AttributeKeyCodeHash:        codeHash,
		AttributeKeyCodeID:          formatUint64(codeID),
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewContractUpdatedEvent creates attributes for contract update
func NewContractUpdatedEvent(
	contractAddress, contractName string,
	oldVersion, newVersion string,
	codeHash string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyContractAddress: contractAddress,
		AttributeKeyContractName:    contractName,
		AttributeKeyOldVersion:      oldVersion,
		AttributeKeyNewVersion:      newVersion,
		AttributeKeyCodeHash:        codeHash,
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewContractVerifiedEvent creates attributes for contract verification
func NewContractVerifiedEvent(
	contractAddress, contractName, contractVersion string,
	verifier string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyContractAddress: contractAddress,
		AttributeKeyContractName:    contractName,
		AttributeKeyContractVersion: contractVersion,
		AttributeKeyVerified:        "true",
		AttributeKeyVerifier:        verifier,
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewContractDeprecatedEvent creates attributes for contract deprecation
func NewContractDeprecatedEvent(
	contractAddress, contractName, reason string,
	replacementAddress string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyContractAddress: contractAddress,
		AttributeKeyContractName:    contractName,
		AttributeKeyDeprecated:      "true",
		AttributeKeyReason:          reason,
		AttributeKeyReplacementAddr: replacementAddress,
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewContractRemovedEvent creates attributes for contract removal
func NewContractRemovedEvent(
	contractAddress, contractName, reason string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyContractAddress: contractAddress,
		AttributeKeyContractName:    contractName,
		AttributeKeyReason:          reason,
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewParamsUpdatedEvent creates attributes for params update
func NewParamsUpdatedEvent(
	updatedBy string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		"updated_by":            updatedBy,
		AttributeKeyBlockHeight: formatInt64(blockHeight),
		AttributeKeyBlockTime:   blockTime,
	}
}

// Helper formatting functions

func formatInt64(i int64) string {
	return fmt.Sprintf("%d", i)
}

func formatUint64(u uint64) string {
	return fmt.Sprintf("%d", u)
}
