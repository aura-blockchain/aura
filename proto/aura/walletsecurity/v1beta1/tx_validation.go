package v1beta1

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/aequitas/aura/proto/common/validation"
)

const (
	// MaxDeviceIDLength is the maximum length for device IDs
	MaxDeviceIDLength = 128
	// MinDeviceIDLength is the minimum length for device IDs
	MinDeviceIDLength = 1
	// MaxFirmwareVersionLength is the maximum length for firmware version strings
	MaxFirmwareVersionLength = 64
	// MaxDerivationPathLength is the maximum length for derivation paths
	MaxDerivationPathLength = 128
	// MaxSignatureSize is the maximum size for signatures
	MaxSignatureSize = 256
	// MinSignatureSize is the minimum size for signatures
	MinSignatureSize = 64
	// MaxSigners is the maximum number of signers in a multi-sig wallet
	MaxSigners = 100
	// MinSigners is the minimum number of signers in a multi-sig wallet
	MinSigners = 1
	// MinThreshold is the minimum threshold for multi-sig wallets
	MinThreshold = int32(1)
	// MaxGuardians is the maximum number of guardians
	MaxGuardians = 20
	// MinGuardians is the minimum number of guardians
	MinGuardians = 1
	// MinRecoveryThreshold is the minimum recovery threshold
	MinRecoveryThreshold = int32(1)
	// MaxTxDataSize is the maximum size for transaction data
	MaxTxDataSize = 1024 * 1024
	// MinTxDataSize is the minimum size for transaction data
	MinTxDataSize = 1
	// MaxDomainLength is the maximum length for domain names
	MaxDomainLength = 253
	// MaxCertHashLength is the maximum length for certificate hashes
	MaxCertHashLength = 128
	// MaxAlgorithmLength is the maximum length for algorithm names
	MaxAlgorithmLength = 64
	// MaxKDFLength is the maximum length for KDF names
	MaxKDFLength = 64
	// MaxSaltSize is the maximum size for salt
	MaxSaltSize = 64
	// MinSaltSize is the minimum size for salt
	MinSaltSize = 16
	// MaxEncryptedSeedSize is the maximum size for encrypted seed
	MaxEncryptedSeedSize = 1024
	// MinEncryptedSeedSize is the minimum size for encrypted seed
	MinEncryptedSeedSize = 32
	// MinIterations is the minimum number of KDF iterations
	MinIterations = int32(10000)
	// MaxIterations is the maximum number of KDF iterations
	MaxIterations = int32(10000000)
	// MaxProofSize is the maximum size for authentication proofs
	MaxProofSize = 4096
	// MinProofSize is the minimum size for authentication proofs
	MinProofSize = 32
	// MaxEnrollmentDataSize is the maximum size for enrollment data
	MaxEnrollmentDataSize = 4096
	// MinEnrollmentDataSize is the minimum size for enrollment data
	MinEnrollmentDataSize = 32
	// MaxKeyMaterialSize is the maximum size for encrypted key material
	MaxKeyMaterialSize = 4096
	// MinKeyMaterialSize is the minimum size for encrypted key material
	MinKeyMaterialSize = 32
	// MaxAttestationLength is the maximum length for attestation certificates
	MaxAttestationLength = 4096
	// MinInactivityThreshold is the minimum inactivity threshold (1 minute)
	MinInactivityThreshold = int32(60)
	// MaxInactivityThreshold is the maximum inactivity threshold (24 hours)
	MaxInactivityThreshold = int32(24 * 60 * 60)
	// MaxDustTransactions is the maximum dust transactions per block
	MaxDustTransactions = int32(1000)
	// MaxSuspiciousThreshold is the maximum suspicious pattern threshold
	MaxSuspiciousThreshold = int32(1000)
)

// parseAndValidatePositiveInt parses a string to Int and validates it's positive
func parseAndValidatePositiveInt(s string, fieldName string) error {
	if s == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	val, ok := sdkmath.NewIntFromString(s)
	if !ok {
		return fmt.Errorf("%s must be a valid integer, got: %s", fieldName, s)
	}
	return validation.ValidatePositiveInt(val, fieldName)
}

// ValidateBasic implements the sdk.Msg interface for MsgRegisterHardwareWallet
func (m *MsgRegisterHardwareWallet) ValidateBasic() error {
	// Validate address
	if err := validation.ValidateAccAddress(m.Address); err != nil {
		return fmt.Errorf("address: %w", err)
	}

	// Hardware wallet type enum is validated at protobuf level

	// Validate device ID
	if err := validation.ValidateBoundedString(m.DeviceId, MinDeviceIDLength, MaxDeviceIDLength, "device_id"); err != nil {
		return err
	}

	// Validate firmware version
	if err := validation.ValidateBoundedString(m.FirmwareVersion, 1, MaxFirmwareVersionLength, "firmware_version"); err != nil {
		return err
	}

	// Validate derivation path
	if err := validation.ValidateBoundedString(m.DerivationPath, 1, MaxDerivationPathLength, "derivation_path"); err != nil {
		return err
	}

	// Validate signature
	if err := validation.ValidateBytes(m.Signature, MinSignatureSize, MaxSignatureSize, "signature"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCreateMultiSigWallet
func (m *MsgCreateMultiSigWallet) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Validate signers
	if len(m.Signers) < MinSigners {
		return fmt.Errorf("signers: must have at least %d signer", MinSigners)
	}

	if len(m.Signers) > MaxSigners {
		return fmt.Errorf("signers: cannot exceed %d signers, got %d", MaxSigners, len(m.Signers))
	}

	// Validate each signer address
	for i, signer := range m.Signers {
		if err := validation.ValidateAccAddress(signer); err != nil {
			return fmt.Errorf("signers[%d]: %w", i, err)
		}
	}

	// Validate threshold
	if m.Threshold < MinThreshold {
		return fmt.Errorf("threshold must be >= %d, got %d", MinThreshold, m.Threshold)
	}

	if m.Threshold > int32(len(m.Signers)) {
		return fmt.Errorf("threshold cannot exceed number of signers: %d > %d", m.Threshold, len(m.Signers))
	}

	// If weight threshold is used, validate it
	if m.WeightThreshold > 0 {
		if len(m.SignerWeights) == 0 {
			return fmt.Errorf("signer_weights must be provided when weight_threshold is set")
		}

		// Validate all signers have weights
		for _, signer := range m.Signers {
			if _, exists := m.SignerWeights[signer]; !exists {
				return fmt.Errorf("signer %s missing weight", signer)
			}
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgSignMultiSigTransaction
func (m *MsgSignMultiSigTransaction) ValidateBasic() error {
	// Validate signer address
	if err := validation.ValidateAccAddress(m.Signer); err != nil {
		return fmt.Errorf("signer: %w", err)
	}

	// Validate transaction ID
	if err := validation.ValidateID(m.TxId, "tx_id"); err != nil {
		return err
	}

	// Validate signature
	if err := validation.ValidateBytes(m.Signature, MinSignatureSize, MaxSignatureSize, "signature"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgConfigureSocialRecovery
func (m *MsgConfigureSocialRecovery) ValidateBasic() error {
	// Validate wallet ID
	if err := validation.ValidateID(m.WalletId, "wallet_id"); err != nil {
		return err
	}

	// Validate guardians
	if len(m.Guardians) < MinGuardians {
		return fmt.Errorf("guardians: must have at least %d guardian", MinGuardians)
	}

	if len(m.Guardians) > MaxGuardians {
		return fmt.Errorf("guardians: cannot exceed %d guardians, got %d", MaxGuardians, len(m.Guardians))
	}

	// Guardian validation would be in the Guardian type

	// Validate recovery threshold
	if m.RecoveryThreshold < MinRecoveryThreshold {
		return fmt.Errorf("recovery_threshold must be >= %d, got %d", MinRecoveryThreshold, m.RecoveryThreshold)
	}

	if m.RecoveryThreshold > int32(len(m.Guardians)) {
		return fmt.Errorf("recovery_threshold cannot exceed number of guardians: %d > %d", m.RecoveryThreshold, len(m.Guardians))
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgInitiateRecovery
func (m *MsgInitiateRecovery) ValidateBasic() error {
	// Validate wallet ID
	if err := validation.ValidateID(m.WalletId, "wallet_id"); err != nil {
		return err
	}

	// Validate new address
	if err := validation.ValidateAccAddress(m.NewAddress); err != nil {
		return fmt.Errorf("new_address: %w", err)
	}

	// Validate initiator address
	if err := validation.ValidateAccAddress(m.Initiator); err != nil {
		return fmt.Errorf("initiator: %w", err)
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgApproveRecovery
func (m *MsgApproveRecovery) ValidateBasic() error {
	// Validate request ID
	if err := validation.ValidateID(m.RequestId, "request_id"); err != nil {
		return err
	}

	// Validate guardian address
	if err := validation.ValidateAccAddress(m.Guardian); err != nil {
		return fmt.Errorf("guardian: %w", err)
	}

	// Validate signature
	if err := validation.ValidateBytes(m.Signature, MinSignatureSize, MaxSignatureSize, "signature"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgExecuteRecovery
func (m *MsgExecuteRecovery) ValidateBasic() error {
	// Validate request ID
	if err := validation.ValidateID(m.RequestId, "request_id"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgSimulateTransaction
func (m *MsgSimulateTransaction) ValidateBasic() error {
	// Validate sender address
	if err := validation.ValidateAccAddress(m.Sender); err != nil {
		return fmt.Errorf("sender: %w", err)
	}

	// Validate transaction data
	if err := validation.ValidateBytes(m.TxData, MinTxDataSize, MaxTxDataSize, "tx_data"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgVerifyDomain
func (m *MsgVerifyDomain) ValidateBasic() error {
	// Validate verifier address
	if err := validation.ValidateAccAddress(m.Verifier); err != nil {
		return fmt.Errorf("verifier: %w", err)
	}

	// Validate domain
	if err := validation.ValidateBoundedString(m.Domain, 1, MaxDomainLength, "domain"); err != nil {
		return err
	}

	// Validate certificate hash
	if err := validation.ValidateBoundedString(m.CertificateHash, validation.MinHashLength, MaxCertHashLength, "certificate_hash"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgSetSpendingLimit
func (m *MsgSetSpendingLimit) ValidateBasic() error {
	// Validate wallet ID
	if err := validation.ValidateID(m.WalletId, "wallet_id"); err != nil {
		return err
	}

	// Validate denom
	if err := validation.ValidateDenom(m.Denom); err != nil {
		return fmt.Errorf("denom: %w", err)
	}

	// Validate daily limit (if provided)
	if m.DailyLimit != "" {
		if err := parseAndValidatePositiveInt(m.DailyLimit, "daily_limit"); err != nil {
			return err
		}
	}

	// Validate weekly limit (if provided)
	if m.WeeklyLimit != "" {
		if err := parseAndValidatePositiveInt(m.WeeklyLimit, "weekly_limit"); err != nil {
			return err
		}
	}

	// Validate monthly limit (if provided)
	if m.MonthlyLimit != "" {
		if err := parseAndValidatePositiveInt(m.MonthlyLimit, "monthly_limit"); err != nil {
			return err
		}
	}

	// At least one limit must be provided
	if m.DailyLimit == "" && m.WeeklyLimit == "" && m.MonthlyLimit == "" {
		return fmt.Errorf("at least one limit (daily, weekly, or monthly) must be provided")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgConfigureSession
func (m *MsgConfigureSession) ValidateBasic() error {
	// Validate wallet ID
	if err := validation.ValidateID(m.WalletId, "wallet_id"); err != nil {
		return err
	}

	// Validate inactivity threshold
	if m.InactivityThresholdSeconds < MinInactivityThreshold {
		return fmt.Errorf("inactivity_threshold_seconds must be >= %d, got %d", MinInactivityThreshold, m.InactivityThresholdSeconds)
	}

	if m.InactivityThresholdSeconds > MaxInactivityThreshold {
		return fmt.Errorf("inactivity_threshold_seconds cannot exceed %d, got %d", MaxInactivityThreshold, m.InactivityThresholdSeconds)
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgLockSession
func (m *MsgLockSession) ValidateBasic() error {
	// Validate session ID
	if err := validation.ValidateID(m.SessionId, "session_id"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUnlockSession
func (m *MsgUnlockSession) ValidateBasic() error {
	// Validate session ID
	if err := validation.ValidateID(m.SessionId, "session_id"); err != nil {
		return err
	}

	// Validate authentication proof
	if err := validation.ValidateBytes(m.AuthenticationProof, MinProofSize, MaxProofSize, "authentication_proof"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgEnrollBiometric
func (m *MsgEnrollBiometric) ValidateBasic() error {
	// Validate wallet ID
	if err := validation.ValidateID(m.WalletId, "wallet_id"); err != nil {
		return err
	}

	// Biometric type enum is validated at protobuf level

	// Validate enrollment data
	if err := validation.ValidateBytes(m.EnrollmentData, MinEnrollmentDataSize, MaxEnrollmentDataSize, "enrollment_data"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAuthenticateBiometric
func (m *MsgAuthenticateBiometric) ValidateBasic() error {
	// Validate wallet ID
	if err := validation.ValidateID(m.WalletId, "wallet_id"); err != nil {
		return err
	}

	// Validate biometric proof
	if err := validation.ValidateBytes(m.BiometricProof, MinProofSize, MaxProofSize, "biometric_proof"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgStoreInSecureEnclave
func (m *MsgStoreInSecureEnclave) ValidateBasic() error {
	// Validate wallet ID
	if err := validation.ValidateID(m.WalletId, "wallet_id"); err != nil {
		return err
	}

	// Enclave type enum is validated at protobuf level

	// Validate encrypted key material
	if err := validation.ValidateBytes(m.EncryptedKeyMaterial, MinKeyMaterialSize, MaxKeyMaterialSize, "encrypted_key_material"); err != nil {
		return err
	}

	// Validate attestation certificate
	if err := validation.ValidateBoundedString(m.AttestationCertificate, 1, MaxAttestationLength, "attestation_certificate"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCreateEncryptedBackup
func (m *MsgCreateEncryptedBackup) ValidateBasic() error {
	// Validate wallet ID
	if err := validation.ValidateID(m.WalletId, "wallet_id"); err != nil {
		return err
	}

	// Validate encrypted seed
	if err := validation.ValidateBytes(m.EncryptedSeed, MinEncryptedSeedSize, MaxEncryptedSeedSize, "encrypted_seed"); err != nil {
		return err
	}

	// Validate encryption algorithm
	if err := validation.ValidateBoundedString(m.EncryptionAlgorithm, 1, MaxAlgorithmLength, "encryption_algorithm"); err != nil {
		return err
	}

	// Validate key derivation function
	if err := validation.ValidateBoundedString(m.KeyDerivationFunction, 1, MaxKDFLength, "key_derivation_function"); err != nil {
		return err
	}

	// Validate salt
	if err := validation.ValidateBytes(m.Salt, MinSaltSize, MaxSaltSize, "salt"); err != nil {
		return err
	}

	// Validate iterations
	if m.Iterations < MinIterations {
		return fmt.Errorf("iterations must be >= %d, got %d", MinIterations, m.Iterations)
	}

	if m.Iterations > MaxIterations {
		return fmt.Errorf("iterations cannot exceed %d, got %d", MaxIterations, m.Iterations)
	}

	// Backup location enum is validated at protobuf level

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgConfigureDustFilter
func (m *MsgConfigureDustFilter) ValidateBasic() error {
	// Validate wallet ID
	if err := validation.ValidateID(m.WalletId, "wallet_id"); err != nil {
		return err
	}

	// Enabled is a boolean, no validation needed

	// Validate minimum amount (if provided)
	if m.MinimumAmount != "" {
		if err := parseAndValidatePositiveInt(m.MinimumAmount, "minimum_amount"); err != nil {
			return err
		}
	}

	// Validate max dust transactions
	if m.MaxDustTransactionsPerBlock < 0 {
		return fmt.Errorf("max_dust_transactions_per_block cannot be negative")
	}

	if m.MaxDustTransactionsPerBlock > MaxDustTransactions {
		return fmt.Errorf("max_dust_transactions_per_block cannot exceed %d, got %d", MaxDustTransactions, m.MaxDustTransactionsPerBlock)
	}

	// Validate suspicious pattern threshold
	if m.SuspiciousPatternThreshold < 0 {
		return fmt.Errorf("suspicious_pattern_threshold cannot be negative")
	}

	if m.SuspiciousPatternThreshold > MaxSuspiciousThreshold {
		return fmt.Errorf("suspicious_pattern_threshold cannot exceed %d, got %d", MaxSuspiciousThreshold, m.SuspiciousPatternThreshold)
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgValidateAddressChecksum
func (m *MsgValidateAddressChecksum) ValidateBasic() error {
	// Validate address
	if err := validation.ValidateAccAddress(m.Address); err != nil {
		return fmt.Errorf("address: %w", err)
	}

	// Checksum algorithm enum is validated at protobuf level

	return nil
}
