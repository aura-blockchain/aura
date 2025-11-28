package types

import (
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// Re-export proto types for convenience
// This allows other packages to use types.TypeName instead of securitypb.TypeName

// Network security types (from proto)
type (
	RateLimitEntry = securitypb.RateLimitEntry
	NodeReputation = securitypb.NodeReputation
	TrustedPeer    = securitypb.TrustedPeer
	ForkAlert      = securitypb.ForkAlert
	PartitionAlert = securitypb.PartitionAlert
)

// Validator security types (from proto)
type (
	ValidatorSecurityInfo = securitypb.ValidatorSecurityInfo
	DoubleSignEvidence    = securitypb.DoubleSignEvidence
	DowntimeInfraction    = securitypb.DowntimeInfraction
	ValidatorAlert        = securitypb.ValidatorAlert
	SentryNodeInfo        = securitypb.SentryNodeInfo
)

// Wallet security types (from proto)
type (
	HardwareWalletConfig       = securitypb.HardwareWalletConfig
	MultiSigWallet             = securitypb.MultiSigWallet
	PendingMultiSigTransaction = securitypb.PendingMultiSigTransaction
	SocialRecoveryConfig       = securitypb.SocialRecoveryConfig
	Guardian                   = securitypb.Guardian
	RecoveryRequest            = securitypb.RecoveryRequest
	SpendingLimit              = securitypb.SpendingLimit
	BiometricAuth              = securitypb.BiometricAuth
)

// Incident response types (from proto)
type (
	Incident      = securitypb.Incident
	AuditLogEntry = securitypb.AuditLogEntry
)

// Cryptography types (from proto)
type (
	KeyRotationSchedule = securitypb.KeyRotationSchedule
	ZKProofConfig       = securitypb.ZKProofConfig
	QuantumResistantKey = securitypb.QuantumResistantKey
)

// Privacy types (from proto)
type (
	MixingPool              = securitypb.MixingPool
	StealthAddress          = securitypb.StealthAddress
	RingSignature           = securitypb.RingSignature
	ConfidentialTransaction = securitypb.ConfidentialTransaction
)

// Params and genesis (from proto)
type (
	Params       = securitypb.Params
	GenesisState = securitypb.GenesisState
)

// Message types (from proto)
// Note: Message types should be imported directly from tx.proto when needed

// Note: Some types are defined in supplemental_types.go as they were missing from proto:
// - BlacklistEntry
// - DeviceFingerprint
// - WalletSession
// - AnomalyDetection
// - PauseState
// - WalletLimit
// - ThresholdScheme (alias for ThresholdSignatureScheme)
// - SecureEnclave
// - RandomSource
// - CertificatePin
// - ViewKey
