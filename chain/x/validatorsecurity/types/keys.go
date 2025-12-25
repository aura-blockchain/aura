// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import "fmt"

const (
	// ModuleName defines the module name
	ModuleName = "validatorsecurity"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_validatorsecurity"
)

// Store key prefixes
var (
	ParamsKey                     = []byte{0x01}
	ValidatorSecurityInfoKey      = []byte{0x02}
	DoubleSignEvidenceKey         = []byte{0x03}
	DowntimeInfractionKey         = []byte{0x04}
	ValidatorAlertKey             = []byte{0x05}
	SentryNodeInfoKey             = []byte{0x06}
	SentryNodeKey                 = []byte{0x06} // Alias for SentryNodeInfoKey
	ValidatorSigningInfoKey       = []byte{0x07}
	ValidatorMissedBlockBitKey    = []byte{0x08}
	JailedValidatorsKey           = []byte{0x09}
	TombstonedValidatorsKey       = []byte{0x0a}
	RegionValidatorCountKey       = []byte{0x0b}
	AlertCounterKey               = []byte{0x0c}
	ValidatorMonitoringKeyPrefix  = []byte{0x0d}
	JailingKeyPrefix              = []byte{0x0e}
	SlashingKeyPrefix             = []byte{0x0f}
	SentryNodeKeyPrefix           = []byte{0x10}
)

// GetValidatorSecurityInfoKey creates the key for validator security info
func GetValidatorSecurityInfoKey(validatorAddr string) []byte {
	return append(ValidatorSecurityInfoKey, []byte(validatorAddr)...)
}

// GetDoubleSignEvidenceKey creates the key for double sign evidence
func GetDoubleSignEvidenceKey(validatorAddr string, height int64) []byte {
	key := append(DoubleSignEvidenceKey, []byte(validatorAddr)...)
	return append(key, []byte(fmt.Sprintf("%d", height))...)
}

// GetDowntimeInfractionKey creates the key for downtime infraction
func GetDowntimeInfractionKey(validatorAddr string) []byte {
	return append(DowntimeInfractionKey, []byte(validatorAddr)...)
}

// GetValidatorAlertKey creates the key for validator alert
func GetValidatorAlertKey(alertID string) []byte {
	return append(ValidatorAlertKey, []byte(alertID)...)
}

// GetSentryNodeInfoKey creates the key for sentry node info
func GetSentryNodeInfoKey(sentryAddr string) []byte {
	return append(SentryNodeInfoKey, []byte(sentryAddr)...)
}

// GetValidatorSigningInfoKey creates the key for validator signing info
func GetValidatorSigningInfoKey(validatorAddr string) []byte {
	return append(ValidatorSigningInfoKey, []byte(validatorAddr)...)
}

// GetValidatorMissedBlockBitKey creates the key for validator missed block bit
func GetValidatorMissedBlockBitKey(validatorAddr string, index int64) []byte {
	key := append(ValidatorMissedBlockBitKey, []byte(validatorAddr)...)
	return append(key, []byte(fmt.Sprintf("%d", index))...)
}

// GetRegionValidatorCountKey creates the key for region validator count
func GetRegionValidatorCountKey(region string) []byte {
	return append(RegionValidatorCountKey, []byte(region)...)
}
