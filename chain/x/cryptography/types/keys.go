package types

const (
	// ModuleName defines the module name
	ModuleName = "cryptography"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// KVStore keys
var (
	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x01}

	// KeyRotationSchedulePrefix is the prefix for key rotation schedules
	KeyRotationSchedulePrefix = []byte{0x02}

	// HDKeyDerivationPrefix is the prefix for HD key derivations
	HDKeyDerivationPrefix = []byte{0x03}

	// ThresholdSchemePrefix is the prefix for threshold signature schemes
	ThresholdSchemePrefix = []byte{0x04}

	// ThresholdSignatureSharePrefix is the prefix for threshold signature shares
	ThresholdSignatureSharePrefix = []byte{0x05}

	// ZKProofConfigPrefix is the prefix for ZK proof configurations
	ZKProofConfigPrefix = []byte{0x06}

	// ZKProofPrefix is the prefix for ZK proofs
	ZKProofPrefix = []byte{0x07}

	// ZKProofVerificationPrefix is the prefix for ZK proof verifications
	ZKProofVerificationPrefix = []byte{0x0e}

	// SecureEnclavePrefix is the prefix for secure enclave configurations
	SecureEnclavePrefix = []byte{0x08}

	// QuantumResistantKeyPrefix is the prefix for quantum-resistant keys
	QuantumResistantKeyPrefix = []byte{0x09}

	// CryptoRandomSourcePrefix is the prefix for crypto random sources
	CryptoRandomSourcePrefix = []byte{0x0a}

	// SaltedHashPrefix is the prefix for salted hashes
	SaltedHashPrefix = []byte{0x0b}

	// KeyStretchingConfigPrefix is the prefix for key stretching configs
	KeyStretchingConfigPrefix = []byte{0x0c}

	// CertificatePinPrefix is the prefix for certificate pins
	CertificatePinPrefix = []byte{0x0d}
)

// GetKeyRotationScheduleKey returns the store key for a key rotation schedule
func GetKeyRotationScheduleKey(id string) []byte {
	return append(KeyRotationSchedulePrefix, []byte(id)...)
}

// GetHDKeyDerivationKey returns the store key for an HD key derivation
func GetHDKeyDerivationKey(masterKeyID string) []byte {
	return append(HDKeyDerivationPrefix, []byte(masterKeyID)...)
}

// GetThresholdSchemeKey returns the store key for a threshold scheme
func GetThresholdSchemeKey(schemeID string) []byte {
	return append(ThresholdSchemePrefix, []byte(schemeID)...)
}

// GetThresholdSignatureShareKey returns the store key for a threshold signature share
func GetThresholdSignatureShareKey(schemeID, participantID string) []byte {
	key := append(ThresholdSignatureSharePrefix, []byte(schemeID)...)
	key = append(key, []byte("/")...)
	return append(key, []byte(participantID)...)
}

// GetZKProofConfigKey returns the store key for a ZK proof configuration
func GetZKProofConfigKey(proofID string) []byte {
	return append(ZKProofConfigPrefix, []byte(proofID)...)
}

// GetZKProofKey returns the store key for a ZK proof
func GetZKProofKey(proofID string) []byte {
	return append(ZKProofPrefix, []byte(proofID)...)
}

// GetZKProofVerificationKey returns the store key for a ZK proof verification
func GetZKProofVerificationKey(verificationID string) []byte {
	return append(ZKProofVerificationPrefix, []byte(verificationID)...)
}

// GetSecureEnclaveKey returns the store key for a secure enclave
func GetSecureEnclaveKey(enclaveID string) []byte {
	return append(SecureEnclavePrefix, []byte(enclaveID)...)
}

// GetQuantumResistantKeyKey returns the store key for a quantum-resistant key
func GetQuantumResistantKeyKey(keyID string) []byte {
	return append(QuantumResistantKeyPrefix, []byte(keyID)...)
}

// GetCryptoRandomSourceKey returns the store key for a crypto random source
func GetCryptoRandomSourceKey(sourceID string) []byte {
	return append(CryptoRandomSourcePrefix, []byte(sourceID)...)
}

// GetSaltedHashKey returns the store key for a salted hash
func GetSaltedHashKey(hashID string) []byte {
	return append(SaltedHashPrefix, []byte(hashID)...)
}

// GetKeyStretchingConfigKey returns the store key for a key stretching config
func GetKeyStretchingConfigKey(configID string) []byte {
	return append(KeyStretchingConfigPrefix, []byte(configID)...)
}

// GetCertificatePinKey returns the store key for a certificate pin
func GetCertificatePinKey(hostname string) []byte {
	return append(CertificatePinPrefix, []byte(hostname)...)
}
