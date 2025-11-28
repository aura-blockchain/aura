package types

const (
	// ModuleName defines the module name
	ModuleName = "walletsecurity"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	ParamsKey                = []byte{0x00}
	HardwareWalletPrefix     = []byte{0x01}
	MultiSigWalletPrefix     = []byte{0x02}
	PendingMultiSigTxPrefix  = []byte{0x03}
	SocialRecoveryPrefix     = []byte{0x04}
	RecoveryRequestPrefix    = []byte{0x05}
	SpendingLimitPrefix      = []byte{0x06}
	SessionPrefix            = []byte{0x07}
	SessionConfigPrefix      = []byte{0x07} // Session configuration storage (same as SessionPrefix for backward compatibility)
	BiometricAuthPrefix      = []byte{0x08}
	SecureEnclavePrefix      = []byte{0x09}
	EncryptedBackupPrefix    = []byte{0x0a}
	DustFilterPrefix         = []byte{0x0b}
	DomainVerificationPrefix = []byte{0x0c}
	SecurityMetricsPrefix    = []byte{0x0d}
	DustTransactionPrefix    = []byte{0x0e}
	DeviceFingerprintPrefix  = []byte{0x0f}
	AnomalyPrefix            = []byte{0x10}
	WalletAnalyticsPrefix    = []byte{0x11}
	InsurancePolicyPrefix    = []byte{0x12}
)

// GetHardwareWalletKey returns the store key for a hardware wallet
func GetHardwareWalletKey(walletID string) []byte {
	return append(HardwareWalletPrefix, []byte(walletID)...)
}

// GetMultiSigWalletKey returns the store key for a multi-sig wallet
func GetMultiSigWalletKey(walletID string) []byte {
	return append(MultiSigWalletPrefix, []byte(walletID)...)
}

// GetPendingMultiSigTxKey returns the store key for a pending multi-sig transaction
func GetPendingMultiSigTxKey(txID string) []byte {
	return append(PendingMultiSigTxPrefix, []byte(txID)...)
}

// GetSocialRecoveryKey returns the store key for social recovery config
func GetSocialRecoveryKey(walletID string) []byte {
	return append(SocialRecoveryPrefix, []byte(walletID)...)
}

// GetRecoveryRequestKey returns the store key for a recovery request
func GetRecoveryRequestKey(requestID string) []byte {
	return append(RecoveryRequestPrefix, []byte(requestID)...)
}

// GetSpendingLimitKey returns the store key for spending limits
func GetSpendingLimitKey(walletID, denom string) []byte {
	return append(SpendingLimitPrefix, []byte(walletID+":"+denom)...)
}

// GetSessionKey returns the store key for session
func GetSessionKey(sessionID string) []byte {
	return append(SessionPrefix, []byte(sessionID)...)
}

// GetSessionConfigKey returns the store key for session configuration
func GetSessionConfigKey(sessionID string) []byte {
	return append(SessionPrefix, []byte(sessionID)...)
}

// GetDeviceFingerprintKey returns the store key for device fingerprint
func GetDeviceFingerprintKey(deviceID string) []byte {
	return append(DeviceFingerprintPrefix, []byte(deviceID)...)
}

// GetAnomalyKey returns the store key for anomaly detection
func GetAnomalyKey(anomalyID string) []byte {
	return append(AnomalyPrefix, []byte(anomalyID)...)
}

// GetWalletAnalyticsKey returns the store key for wallet analytics
func GetWalletAnalyticsKey(walletID string) []byte {
	return append(WalletAnalyticsPrefix, []byte(walletID)...)
}

// GetInsurancePolicyKey returns the store key for insurance policy
func GetInsurancePolicyKey(policyID string) []byte {
	return append(InsurancePolicyPrefix, []byte(policyID)...)
}

// GetBiometricAuthKey returns the store key for biometric auth
func GetBiometricAuthKey(walletID string) []byte {
	return append(BiometricAuthPrefix, []byte(walletID)...)
}

// GetSecureEnclaveKey returns the store key for secure enclave config
func GetSecureEnclaveKey(walletID string) []byte {
	return append(SecureEnclavePrefix, []byte(walletID)...)
}

// GetEncryptedBackupKey returns the store key for encrypted backup
func GetEncryptedBackupKey(backupID string) []byte {
	return append(EncryptedBackupPrefix, []byte(backupID)...)
}

// GetDustFilterKey returns the store key for dust filter
func GetDustFilterKey(walletID string) []byte {
	return append(DustFilterPrefix, []byte(walletID)...)
}

// GetDomainVerificationKey returns the store key for domain verification
func GetDomainVerificationKey(domain string) []byte {
	return append(DomainVerificationPrefix, []byte(domain)...)
}

// GetSecurityMetricsKey returns the store key for security metrics
func GetSecurityMetricsKey(walletID string) []byte {
	return append(SecurityMetricsPrefix, []byte(walletID)...)
}

// GetDustTransactionKey returns the store key for a dust transaction
func GetDustTransactionKey(txHash string) []byte {
	return append(DustTransactionPrefix, []byte(txHash)...)
}
