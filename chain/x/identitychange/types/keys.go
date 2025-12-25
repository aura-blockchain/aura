// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

const (
	// ModuleName defines the module name
	ModuleName = "identitychange"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// Store key prefixes
const (
	// RecordStoreKeyPrefix is the prefix for identity records
	RecordStoreKeyPrefix = "record/"

	// RequestStoreKeyPrefix is the prefix for identity change requests
	RequestStoreKeyPrefix = "request/"

	// HistoryStoreKeyPrefix is the prefix for identity change history
	HistoryStoreKeyPrefix = "history/"

	// RecoveryStoreKeyPrefix is the prefix for identity recovery records
	RecoveryStoreKeyPrefix = "recovery/"

	// VerificationStoreKeyPrefix is the prefix for identity verification records
	VerificationStoreKeyPrefix = "verification/"

	// DelegationStoreKeyPrefix is the prefix for identity delegation records
	DelegationStoreKeyPrefix = "delegation/"

	// FederationStoreKeyPrefix is the prefix for identity federation records
	FederationStoreKeyPrefix = "federation/"

	// CrossChainLinkStoreKeyPrefix is the prefix for cross-chain identity links
	CrossChainLinkStoreKeyPrefix = "crosschain/"

	// SuspendedKey is the key for the suspended flag
	SuspendedKey = "suspended"
)

// RecordStoreKey returns the store key for an identity record
func RecordStoreKey(did string) string {
	return RecordStoreKeyPrefix + did
}

// RequestStoreKey returns the store key for an identity change request
func RequestStoreKey(requestID string) string {
	return RequestStoreKeyPrefix + requestID
}

// HistoryStoreKey returns the store key for identity change history
// Format: history/{did}/{height}
func HistoryStoreKey(did string, height uint64) string {
	return HistoryStoreKeyPrefix + did + "/" + string(rune(height))
}

// RecoveryStoreKey returns the store key for identity recovery
func RecoveryStoreKey(did string) string {
	return RecoveryStoreKeyPrefix + did
}

// VerificationStoreKey returns the store key for identity verification
func VerificationStoreKey(did string) string {
	return VerificationStoreKeyPrefix + did
}

// DelegationStoreKey returns the store key for identity delegation
func DelegationStoreKey(did string) string {
	return DelegationStoreKeyPrefix + did
}

// FederationStoreKey returns the store key for identity federation
func FederationStoreKey(did string) string {
	return FederationStoreKeyPrefix + did
}

// CrossChainLinkStoreKey returns the store key for cross-chain identity link
func CrossChainLinkStoreKey(did string) string {
	return CrossChainLinkStoreKeyPrefix + did
}
