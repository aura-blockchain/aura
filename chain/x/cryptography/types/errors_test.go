package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	// Test error registration
	tests := []struct {
		name string
		err  error
	}{
		{"ErrInvalidInput", ErrInvalidInput},
		{"ErrInvalidKeyID", ErrInvalidKeyID},
		{"ErrKeyRotationScheduleNotFound", ErrKeyRotationScheduleNotFound},
		{"ErrInvalidRotationInterval", ErrInvalidRotationInterval},
		{"ErrKeyRotationInProgress", ErrKeyRotationInProgress},
		{"ErrInvalidThreshold", ErrInvalidThreshold},
		{"ErrInvalidParticipantCount", ErrInvalidParticipantCount},
		{"ErrThresholdSchemeNotFound", ErrThresholdSchemeNotFound},
		{"ErrInvalidSignatureShare", ErrInvalidSignatureShare},
		{"ErrThresholdNotReached", ErrThresholdNotReached},
		{"ErrInvalidZKProof", ErrInvalidZKProof},
		{"ErrZKProofConfigNotFound", ErrZKProofConfigNotFound},
		{"ErrSecureEnclaveNotFound", ErrSecureEnclaveNotFound},
		{"ErrEnclaveAttestationFailed", ErrEnclaveAttestationFailed},
		{"ErrQuantumKeyNotFound", ErrQuantumKeyNotFound},
		{"ErrInvalidQuantumAlgorithm", ErrInvalidQuantumAlgorithm},
		{"ErrInsufficientEntropy", ErrInsufficientEntropy},
		{"ErrRandomSourceFailed", ErrRandomSourceFailed},
		{"ErrInvalidSaltLength", ErrInvalidSaltLength},
		{"ErrInvalidHashAlgorithm", ErrInvalidHashAlgorithm},
		{"ErrInvalidKeyStretchingConfig", ErrInvalidKeyStretchingConfig},
		{"ErrInvalidIterationCount", ErrInvalidIterationCount},
		{"ErrCertificatePinNotFound", ErrCertificatePinNotFound},
		{"ErrCertificateVerificationFailed", ErrCertificateVerificationFailed},
		{"ErrInvalidCertificateHash", ErrInvalidCertificateHash},
		{"ErrInvalidDerivationPath", ErrInvalidDerivationPath},
		{"ErrHDKeyDerivationFailed", ErrHDKeyDerivationFailed},
		{"ErrInvalidSeed", ErrInvalidSeed},
		{"ErrKeyExpired", ErrKeyExpired},
		{"ErrUnauthorized", ErrUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.err)
			require.NotEmpty(t, tt.err.Error())
		})
	}
}

func TestErrorMessages(t *testing.T) {
	require.Contains(t, ErrInvalidInput.Error(), "invalid input")
	require.Contains(t, ErrInvalidKeyID.Error(), "invalid key ID")
	require.Contains(t, ErrKeyRotationScheduleNotFound.Error(), "key rotation schedule not found")
	require.Contains(t, ErrInvalidRotationInterval.Error(), "invalid rotation interval")
	require.Contains(t, ErrInvalidThreshold.Error(), "invalid threshold")
	require.Contains(t, ErrInsufficientEntropy.Error(), "insufficient entropy")
}
