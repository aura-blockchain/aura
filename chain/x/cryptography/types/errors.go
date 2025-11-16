package types

import (
	"fmt"
)

// Cryptography module errors
var (
	ErrInvalidKeyID                  = fmt.Errorf("invalid key ID")
	ErrKeyRotationScheduleNotFound   = fmt.Errorf("key rotation schedule not found")
	ErrInvalidRotationInterval       = fmt.Errorf("invalid rotation interval")
	ErrKeyRotationInProgress         = fmt.Errorf("key rotation already in progress")
	ErrInvalidThreshold              = fmt.Errorf("invalid threshold value")
	ErrInvalidParticipantCount       = fmt.Errorf("invalid participant count")
	ErrThresholdSchemeNotFound       = fmt.Errorf("threshold scheme not found")
	ErrInvalidSignatureShare         = fmt.Errorf("invalid signature share")
	ErrThresholdNotReached           = fmt.Errorf("threshold not reached")
	ErrInvalidZKProof                = fmt.Errorf("invalid zero-knowledge proof")
	ErrZKProofConfigNotFound         = fmt.Errorf("ZK proof configuration not found")
	ErrSecureEnclaveNotFound         = fmt.Errorf("secure enclave not found")
	ErrEnclaveAttestationFailed      = fmt.Errorf("enclave attestation failed")
	ErrQuantumKeyNotFound            = fmt.Errorf("quantum-resistant key not found")
	ErrInvalidQuantumAlgorithm       = fmt.Errorf("invalid quantum-resistant algorithm")
	ErrInsufficientEntropy           = fmt.Errorf("insufficient entropy")
	ErrRandomSourceFailed            = fmt.Errorf("random source failed")
	ErrInvalidSaltLength             = fmt.Errorf("invalid salt length")
	ErrInvalidHashAlgorithm          = fmt.Errorf("invalid hash algorithm")
	ErrInvalidKeyStretchingConfig    = fmt.Errorf("invalid key stretching configuration")
	ErrInvalidIterationCount         = fmt.Errorf("invalid iteration count")
	ErrCertificatePinNotFound        = fmt.Errorf("certificate pin not found")
	ErrCertificateVerificationFailed = fmt.Errorf("certificate verification failed")
	ErrInvalidCertificateHash        = fmt.Errorf("invalid certificate hash")
	ErrInvalidDerivationPath         = fmt.Errorf("invalid derivation path")
	ErrHDKeyDerivationFailed         = fmt.Errorf("HD key derivation failed")
	ErrInvalidSeed                   = fmt.Errorf("invalid seed")
	ErrKeyExpired                    = fmt.Errorf("key has expired")
	ErrUnauthorized                  = fmt.Errorf("unauthorized operation")
)
