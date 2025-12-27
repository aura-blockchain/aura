// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

// Package types defines the types for the consolidated security module.
// This module combines: networksecurity, validatorsecurity, walletsecurity,
// incidentresponse, cryptography, and privacy into a unified security layer.
package types

const (
	// ModuleName defines the module name for the consolidated security module
	ModuleName = "security"

	// StoreKey defines the primary store key
	StoreKey = ModuleName

	// RouterKey defines the routing key
	RouterKey = ModuleName

	// QuerierRoute defines the query route
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_" + ModuleName
)

// Store key prefixes for different security domains
// These maintain logical separation within the unified store
var (
	// Network security prefixes
	NetworkParamsKey      = []byte{0x01}
	RateLimitKey          = []byte{0x02}
	ReputationKey         = []byte{0x03}
	TrustedPeerKey        = []byte{0x04}
	BlacklistKey          = []byte{0x05}
	GossipFilterKey       = []byte{0x06}
	SybilDetectionKey     = []byte{0x07}
	ForkAlertKey          = []byte{0x08}
	PartitionAlertKey     = []byte{0x09}

	// Validator security prefixes
	ValidatorParamsKey    = []byte{0x10}
	ValidatorInfoKey      = []byte{0x11}
	DoubleSignEvidenceKey = []byte{0x12}
	DowntimeInfractionKey = []byte{0x13}
	ValidatorAlertKey     = []byte{0x14}
	SentryNodeKey         = []byte{0x15}
	JailRecordKey         = []byte{0x16}
	SlashRecordKey        = []byte{0x17}

	// Wallet security prefixes
	WalletParamsKey      = []byte{0x20}
	HardwareWalletKey    = []byte{0x21}
	MultiSigWalletKey    = []byte{0x22}
	PendingMultiSigTxKey = []byte{0x23}
	SocialRecoveryKey    = []byte{0x24}
	RecoveryRequestKey   = []byte{0x25}
	DeviceFingerprintKey = []byte{0x26}
	SessionKey           = []byte{0x27}
	AnomalyDetectionKey  = []byte{0x28}
	WalletAnalyticsKey   = []byte{0x29}
	InsurancePolicyKey   = []byte{0x2A}
	SpendingLimitKey     = []byte{0x2B}
	BiometricAuthKey     = []byte{0x2C}

	// Incident response prefixes
	IncidentParamsKey   = []byte{0x30}
	IncidentKey         = []byte{0x31}
	PauseStateKey       = []byte{0x32}
	WalletLimitKey      = []byte{0x33}
	NextIncidentIDKey   = []byte{0x34}
	CircuitBreakerKey   = []byte{0x35}
	AuditLogKey         = []byte{0x36}

	// Security guard prefixes (memory store)
	ReentrancyLockKey = []byte{0xF0} // Memory-only, cleared on block boundaries

	// Cryptography prefixes
	CryptoParamsKey           = []byte{0x40}
	KeyRotationScheduleKey    = []byte{0x41}
	ThresholdSchemeKey        = []byte{0x42}
	ZKProofConfigKey          = []byte{0x43}
	SecureEnclaveKey          = []byte{0x44}
	QuantumResistantKeyPrefix = []byte{0x45}
	RandomSourceKey           = []byte{0x46}
	KeyStretchingKey          = []byte{0x47}
	CertificatePinKey         = []byte{0x48}

	// Privacy prefixes
	PrivacyParamsKey  = []byte{0x50}
	MixingPoolKey     = []byte{0x51}
	ViewKeyKey        = []byte{0x52}
	RingSignatureKey  = []byte{0x53}
	StealthAddressKey = []byte{0x54}
	ConfidentialTxKey = []byte{0x55}
)

// Legacy store keys for migration compatibility
// These allow reading state from chains that used the old module structure
const (
	LegacyNetworkSecurityStore   = "networksecurity"
	LegacyValidatorSecurityStore = "validatorsecurity"
	LegacyWalletSecurityStore    = "walletsecurity"
	LegacyIncidentResponseStore  = "incidentresponse"
	LegacyCryptographyStore      = "cryptography"
	LegacyPrivacyStore           = "privacy"
)

// Event types
const (
	// Network security events
	EventTypeRateLimitExceeded      = "rate_limit_exceeded"
	EventTypePeerBlacklisted        = "peer_blacklisted"
	EventTypeForkDetected           = "fork_detected"
	EventTypePartitionDetected      = "partition_detected"
	EventTypeSybilAttackDetected    = "sybil_attack_detected"
	EventTypeTrustedPeerAdded       = "trusted_peer_added"
	EventTypeTrustedPeerRemoved     = "trusted_peer_removed"
	EventTypePeerBanned             = "peer_banned"
	EventTypePeerUnbanned           = "peer_unbanned"
	EventTypePeerReputationUpdated  = "peer_reputation_updated"
	EventTypeForkAlertResolved      = "fork_alert_resolved"
	EventTypePartitionAlertResolved = "partition_alert_resolved"

	// Validator security events
	EventTypeDoubleSignDetected           = "double_sign_detected"
	EventTypeDoubleSignReported           = "double_sign_reported"
	EventTypeDowntimeDetected             = "downtime_detected"
	EventTypeValidatorSlashed             = "validator_slashed"
	EventTypeValidatorJailed              = "validator_jailed"
	EventTypeValidatorUnjailed            = "validator_unjailed"
	EventTypeValidatorSecurityRegistered  = "validator_security_registered"
	EventTypeValidatorSecurityUpdated     = "validator_security_updated"
	EventTypeSentryNodeRegistered         = "sentry_node_registered"
	EventTypeSentryNodeRemoved            = "sentry_node_removed"
	EventTypeValidatorAlertAcknowledged   = "validator_alert_acknowledged"
	EventTypeFailoverTriggered            = "failover_triggered"

	// Wallet security events
	EventTypeWalletCreated           = "wallet_created"
	EventTypeHardwareWalletRegistered = "hardware_wallet_registered"
	EventTypeMultiSigWalletCreated   = "multisig_wallet_created"
	EventTypeMultiSigTxProposed      = "multisig_tx_proposed"
	EventTypeMultiSigTxSigned        = "multisig_tx_signed"
	EventTypeMultiSigTxExecuted      = "multisig_tx_executed"
	EventTypeMultiSigProposed        = "multisig_proposed"
	EventTypeMultiSigApproved        = "multisig_approved"
	EventTypeMultiSigExecuted        = "multisig_executed"
	EventTypeSocialRecoveryConfigured = "social_recovery_configured"
	EventTypeRecoveryInitiated       = "recovery_initiated"
	EventTypeRecoveryApproved        = "recovery_approved"
	EventTypeRecoveryExecuted        = "recovery_executed"
	EventTypeRecoveryCompleted       = "recovery_completed"
	EventTypeAnomalyDetected         = "anomaly_detected"
	EventTypeSpendingLimitSet        = "spending_limit_set"
	EventTypeBiometricRegistered     = "biometric_registered"

	// Incident response events
	EventTypeIncidentCreated        = "incident_created"
	EventTypeIncidentUpdated        = "incident_updated"
	EventTypeIncidentResolved       = "incident_resolved"
	EventTypeResponseActionExecuted = "response_action_executed"
	EventTypeAuditLog               = "audit_log"
	EventTypeSystemPaused           = "system_paused"
	EventTypeSystemResumed          = "system_resumed"

	// Cryptography events
	EventTypeKeyRotated                  = "key_rotated"
	EventTypeKeyRotationScheduleCreated  = "key_rotation_schedule_created"
	EventTypeThresholdSchemeInit         = "threshold_scheme_initialized"
	EventTypeThresholdSchemeCreated      = "threshold_scheme_created"
	EventTypeThresholdShareSubmitted     = "threshold_share_submitted"
	EventTypeZKProofVerified             = "zk_proof_verified"
	EventTypeZKCircuitRegistered         = "zk_circuit_registered"
	EventTypeZKProofSubmitted            = "zk_proof_submitted"
	EventTypeQuantumKeyGenerated         = "quantum_key_generated"
	EventTypeParamsUpdated               = "params_updated"

	// Privacy events
	EventTypeMixingPoolCreated    = "mixing_pool_created"
	EventTypeMixingPoolJoined     = "mixing_pool_joined"
	EventTypeMixingExecuted       = "mixing_executed"
	EventTypeMixingCompleted      = "mixing_completed"
	EventTypeStealthAddressGenerated = "stealth_address_generated"
	EventTypeStealthAddressUsed   = "stealth_address_used"
	EventTypeRingSignatureCreated = "ring_signature_created"
	EventTypeConfidentialTxCreated = "confidential_tx_created"
	EventTypeConfidentialTxSent   = "confidential_tx_sent"
)

// Event attribute keys
const (
	AttributeKeyPeerId            = "peer_id"
	AttributeKeyReason            = "reason"
	AttributeKeyValidatorAddr     = "validator_address"
	AttributeKeyValidatorAddress  = "validator_address"
	AttributeKeyWalletAddr        = "wallet_address"
	AttributeKeyWalletId          = "wallet_id"
	AttributeKeyIncidentId        = "incident_id"
	AttributeKeySeverity          = "severity"
	AttributeKeyScheduleId        = "schedule_id"
	AttributeKeyRotationTime      = "rotation_time"
	AttributeKeySchemeId          = "scheme_id"
	AttributeKeyProofId           = "proof_id"
	AttributeKeyTxHash            = "tx_hash"
	AttributeKeyAmount            = "amount"
	AttributeKeyDomain            = "domain"
	AttributeKeyBlockHeight       = "block_height"
	AttributeKeyTimestamp         = "timestamp"
	AttributeKeyStatus            = "status"
	AttributeKeyRiskLevel         = "risk_level"
	AttributeKeyThreshold         = "threshold"
	AttributeKeyParticipants      = "participants"
	AttributeKeyAlertId           = "alert_id"
	AttributeKeySentryAddress     = "sentry_address"
	AttributeKeyReporter          = "reporter"
	AttributeKeyAcknowledgedBy    = "acknowledged_by"
	AttributeKeyKeyId             = "key_id"
	AttributeKeyPoolId            = "pool_id"
	AttributeKeyAddress           = "address"
	AttributeKeyCreator           = "creator"
	AttributeKeyRequestId         = "request_id"
	AttributeKeyGuardian          = "guardian"
	AttributeKeyExecutor          = "executor"
	AttributeKeyActionId          = "action_id"
	AttributeKeyLogId             = "log_id"
	AttributeKeyCircuitId         = "circuit_id"
	AttributeKeyAlgorithm         = "algorithm"
	AttributeKeySuccess           = "success"
	AttributeKeyTxId              = "tx_id"
	AttributeKeyProposer          = "proposer"
	AttributeKeySigner            = "signer"
	AttributeKeyInitiator         = "initiator"
	AttributeKeyActionType        = "action_type"
	AttributeKeySubmitter         = "submitter"
	AttributeKeyParticipant       = "participant"
	AttributeKeyOwner             = "owner"
	AttributeKeySignatureId       = "signature_id"
	AttributeKeyRingSize          = "ring_size"
	AttributeKeySender            = "sender"
	AttributeKeyAuthority         = "authority"
)

// =============================================================================
// Store Key Helper Functions
// =============================================================================

// Network security key helpers

// GetRateLimitStoreKey returns the store key for a rate limit entry
func GetRateLimitStoreKey(identifier string) []byte {
	return append(RateLimitKey, []byte(identifier)...)
}

// GetReputationStoreKey returns the store key for a node reputation
func GetReputationStoreKey(nodeId string) []byte {
	return append(ReputationKey, []byte(nodeId)...)
}

// GetTrustedPeerStoreKey returns the store key for a trusted peer
func GetTrustedPeerStoreKey(peerId string) []byte {
	return append(TrustedPeerKey, []byte(peerId)...)
}

// GetBlacklistStoreKey returns the store key for a blacklist entry
func GetBlacklistStoreKey(identifier string) []byte {
	return append(BlacklistKey, []byte(identifier)...)
}

// GetForkAlertStoreKey returns the store key for a fork alert
func GetForkAlertStoreKey(alertId string) []byte {
	return append(ForkAlertKey, []byte(alertId)...)
}

// GetPartitionAlertStoreKey returns the store key for a partition alert
func GetPartitionAlertStoreKey(alertId string) []byte {
	return append(PartitionAlertKey, []byte(alertId)...)
}

// Validator security key helpers

// GetValidatorInfoStoreKey returns the store key for validator security info
func GetValidatorInfoStoreKey(valAddr string) []byte {
	return append(ValidatorInfoKey, []byte(valAddr)...)
}

// GetDoubleSignEvidenceStoreKey returns the store key for double sign evidence
func GetDoubleSignEvidenceStoreKey(evidenceId string) []byte {
	return append(DoubleSignEvidenceKey, []byte(evidenceId)...)
}

// GetDowntimeInfractionStoreKey returns the store key for a downtime infraction
func GetDowntimeInfractionStoreKey(valAddr string) []byte {
	return append(DowntimeInfractionKey, []byte(valAddr)...)
}

// GetValidatorAlertStoreKey returns the store key for a validator alert
func GetValidatorAlertStoreKey(alertId string) []byte {
	return append(ValidatorAlertKey, []byte(alertId)...)
}

// GetSentryNodeStoreKey returns the store key for a sentry node
func GetSentryNodeStoreKey(nodeId string) []byte {
	return append(SentryNodeKey, []byte(nodeId)...)
}

// Wallet security key helpers

// GetHardwareWalletStoreKey returns the store key for hardware wallet config
func GetHardwareWalletStoreKey(walletAddr string) []byte {
	return append(HardwareWalletKey, []byte(walletAddr)...)
}

// GetMultiSigWalletStoreKey returns the store key for a multisig wallet
func GetMultiSigWalletStoreKey(walletAddr string) []byte {
	return append(MultiSigWalletKey, []byte(walletAddr)...)
}

// GetPendingMultiSigTxStoreKey returns the store key for a pending multisig tx
func GetPendingMultiSigTxStoreKey(txId string) []byte {
	return append(PendingMultiSigTxKey, []byte(txId)...)
}

// GetSocialRecoveryStoreKey returns the store key for social recovery config
func GetSocialRecoveryStoreKey(walletAddr string) []byte {
	return append(SocialRecoveryKey, []byte(walletAddr)...)
}

// GetRecoveryRequestStoreKey returns the store key for a recovery request
func GetRecoveryRequestStoreKey(requestId string) []byte {
	return append(RecoveryRequestKey, []byte(requestId)...)
}

// GetDeviceFingerprintStoreKey returns the store key for a device fingerprint
func GetDeviceFingerprintStoreKey(fingerprintId string) []byte {
	return append(DeviceFingerprintKey, []byte(fingerprintId)...)
}

// GetSessionStoreKey returns the store key for a wallet session
func GetSessionStoreKey(sessionId string) []byte {
	return append(SessionKey, []byte(sessionId)...)
}

// GetAnomalyDetectionStoreKey returns the store key for anomaly detection
func GetAnomalyDetectionStoreKey(anomalyId string) []byte {
	return append(AnomalyDetectionKey, []byte(anomalyId)...)
}

// GetSpendingLimitStoreKey returns the store key for a spending limit
func GetSpendingLimitStoreKey(walletId string) []byte {
	return append(SpendingLimitKey, []byte(walletId)...)
}

// GetBiometricAuthStoreKey returns the store key for biometric auth
func GetBiometricAuthStoreKey(walletAddr string) []byte {
	return append(BiometricAuthKey, []byte(walletAddr)...)
}

// Incident response key helpers

// GetIncidentStoreKey returns the store key for an incident
func GetIncidentStoreKey(incidentId string) []byte {
	return append(IncidentKey, []byte(incidentId)...)
}

// GetWalletLimitStoreKey returns the store key for wallet limits
func GetWalletLimitStoreKey(walletAddr string) []byte {
	return append(WalletLimitKey, []byte(walletAddr)...)
}

// GetAuditLogStoreKey returns the store key for an audit log entry
func GetAuditLogStoreKey(logId string) []byte {
	return append(AuditLogKey, []byte(logId)...)
}

// Cryptography key helpers

// GetKeyRotationScheduleStoreKey returns the store key for a key rotation schedule
func GetKeyRotationScheduleStoreKey(scheduleId string) []byte {
	return append(KeyRotationScheduleKey, []byte(scheduleId)...)
}

// GetThresholdSchemeStoreKey returns the store key for a threshold scheme
func GetThresholdSchemeStoreKey(schemeId string) []byte {
	return append(ThresholdSchemeKey, []byte(schemeId)...)
}

// GetZKProofConfigStoreKey returns the store key for a ZK proof config
func GetZKProofConfigStoreKey(proofId string) []byte {
	return append(ZKProofConfigKey, []byte(proofId)...)
}

// GetSecureEnclaveStoreKey returns the store key for a secure enclave
func GetSecureEnclaveStoreKey(enclaveId string) []byte {
	return append(SecureEnclaveKey, []byte(enclaveId)...)
}

// GetQuantumResistantKeyStoreKey returns the store key for a quantum-resistant key
func GetQuantumResistantKeyStoreKey(keyId string) []byte {
	return append(QuantumResistantKeyPrefix, []byte(keyId)...)
}

// GetRandomSourceStoreKey returns the store key for a random source
func GetRandomSourceStoreKey(sourceId string) []byte {
	return append(RandomSourceKey, []byte(sourceId)...)
}

// GetCertificatePinStoreKey returns the store key for a certificate pin
func GetCertificatePinStoreKey(pinId string) []byte {
	return append(CertificatePinKey, []byte(pinId)...)
}

// Privacy key helpers

// GetMixingPoolStoreKey returns the store key for a mixing pool
func GetMixingPoolStoreKey(poolId string) []byte {
	return append(MixingPoolKey, []byte(poolId)...)
}

// GetViewKeyStoreKey returns the store key for a view key
func GetViewKeyStoreKey(keyId string) []byte {
	return append(ViewKeyKey, []byte(keyId)...)
}

// GetRingSignatureStoreKey returns the store key for a ring signature
// Uses hex-encoded KeyImage as the key since RingSignature has no id field
func GetRingSignatureStoreKey(keyImageHex string) []byte {
	return append(RingSignatureKey, []byte(keyImageHex)...)
}

// GetStealthAddressStoreKey returns the store key for a stealth address
// Uses hex-encoded OneTimeAddress as the key since StealthAddress has no id field
func GetStealthAddressStoreKey(oneTimeAddressHex string) []byte {
	return append(StealthAddressKey, []byte(oneTimeAddressHex)...)
}

// GetConfidentialTxStoreKey returns the store key for a confidential transaction
func GetConfidentialTxStoreKey(txId string) []byte {
	return append(ConfidentialTxKey, []byte(txId)...)
}

// Security guard key helpers

// GetReentrancyLockKey returns the store key for a reentrancy lock
// This is stored in memory store and cleared on block boundaries
func GetReentrancyLockKey(lockId string) []byte {
	return append(ReentrancyLockKey, []byte(lockId)...)
}

// GetPauseStateKey returns the store key for module pause state
func GetPauseStateKey(moduleName string) []byte {
	return append(PauseStateKey, []byte(moduleName)...)
}
