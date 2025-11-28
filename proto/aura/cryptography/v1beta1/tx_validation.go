package v1beta1

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/common/validation"
)

const (
	// MaxKeyIDLength is the maximum length for key identifiers
	MaxKeyIDLength = 128
	// MinKeyIDLength is the minimum length for key identifiers
	MinKeyIDLength = 1
	// MaxPublicKeySize is the maximum size for a public key (8KB)
	MaxPublicKeySize = 8192
	// MinPublicKeySize is the minimum size for a public key (32 bytes)
	MinPublicKeySize = 32
	// MaxSignatureShareSize is the maximum size for a signature share
	MaxSignatureShareSize = 512
	// MinSignatureShareSize is the minimum size for a signature share
	MinSignatureShareSize = 32
	// MaxMessageHashSize is the maximum size for a message hash
	MaxMessageHashSize = 128
	// MinMessageHashSize is the minimum size for a message hash (32 bytes for SHA-256)
	MinMessageHashSize = 32
	// MinRotationInterval is the minimum key rotation interval (1 day in seconds)
	MinRotationInterval = int64(24 * 60 * 60)
	// MaxRotationInterval is the maximum key rotation interval (365 days in seconds)
	MaxRotationInterval = int64(365 * 24 * 60 * 60)
	// MinThreshold is the minimum threshold for threshold schemes
	MinThreshold = int32(1)
	// MaxParticipants is the maximum number of participants in threshold scheme
	MaxParticipants = int32(100)
	// MaxCircuitSize is the maximum size for ZK proof circuits (10MB)
	MaxCircuitSize = 10 * 1024 * 1024
	// MaxProofSize is the maximum size for ZK proofs (1MB)
	MaxProofSize = 1024 * 1024
	// MinProofSize is the minimum size for ZK proofs
	MinProofSize = 32
	// MaxEnclaveIDLength is the maximum length for enclave identifiers
	MaxEnclaveIDLength = 256
	// MaxAttestationSize is the maximum size for attestation data
	MaxAttestationSize = 4096
	// MaxHostnameLength is the maximum length for hostnames
	MaxHostnameLength = 253
	// MaxCertificatePinSize is the maximum size for certificate pins
	MaxCertificatePinSize = 256
)

// ValidateBasic implements the sdk.Msg interface for MsgCreateKeyRotationSchedule
func (m *MsgCreateKeyRotationSchedule) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Validate key ID
	if err := validation.ValidateBoundedString(m.KeyId, MinKeyIDLength, MaxKeyIDLength, "key_id"); err != nil {
		return err
	}

	// Validate rotation interval
	if m.RotationIntervalSeconds < MinRotationInterval {
		return fmt.Errorf("rotation_interval_seconds must be >= %d, got %d", MinRotationInterval, m.RotationIntervalSeconds)
	}
	if m.RotationIntervalSeconds > MaxRotationInterval {
		return fmt.Errorf("rotation_interval_seconds must be <= %d, got %d", MaxRotationInterval, m.RotationIntervalSeconds)
	}

	// Policy enum is validated at protobuf level

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgRotateKey
func (m *MsgRotateKey) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Validate key ID
	if err := validation.ValidateBoundedString(m.KeyId, MinKeyIDLength, MaxKeyIDLength, "key_id"); err != nil {
		return err
	}

	// Validate new public key
	if err := validation.ValidateBytes(m.NewPublicKey, MinPublicKeySize, MaxPublicKeySize, "new_public_key"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCreateThresholdScheme
func (m *MsgCreateThresholdScheme) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Validate threshold
	if m.Threshold < MinThreshold {
		return fmt.Errorf("threshold must be >= %d, got %d", MinThreshold, m.Threshold)
	}

	// Validate total participants
	if m.TotalParticipants < m.Threshold {
		return fmt.Errorf("total_participants must be >= threshold, got %d < %d", m.TotalParticipants, m.Threshold)
	}

	if m.TotalParticipants > MaxParticipants {
		return fmt.Errorf("total_participants cannot exceed %d, got %d", MaxParticipants, m.TotalParticipants)
	}

	// Validate participant IDs
	if len(m.ParticipantIds) != int(m.TotalParticipants) {
		return fmt.Errorf("participant_ids length must match total_participants, got %d != %d", len(m.ParticipantIds), m.TotalParticipants)
	}

	// Validate each participant ID
	for i, id := range m.ParticipantIds {
		if err := validation.ValidateBoundedString(id, 1, MaxKeyIDLength, fmt.Sprintf("participant_ids[%d]", i)); err != nil {
			return err
		}
	}

	// Scheme type enum is validated at protobuf level

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgSubmitThresholdSignatureShare
func (m *MsgSubmitThresholdSignatureShare) ValidateBasic() error {
	// Validate submitter address
	if err := validation.ValidateAccAddress(m.Submitter); err != nil {
		return fmt.Errorf("submitter: %w", err)
	}

	// Validate scheme ID
	if err := validation.ValidateBoundedString(m.SchemeId, MinKeyIDLength, MaxKeyIDLength, "scheme_id"); err != nil {
		return err
	}

	// Validate signature share
	if err := validation.ValidateBytes(m.SignatureShare, MinSignatureShareSize, MaxSignatureShareSize, "signature_share"); err != nil {
		return err
	}

	// Validate message hash
	if err := validation.ValidateBytes(m.MessageHash, MinMessageHashSize, MaxMessageHashSize, "message_hash"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgRegisterZKProofCircuit
func (m *MsgRegisterZKProofCircuit) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Validate circuit ID
	if err := validation.ValidateBoundedString(m.CircuitId, MinKeyIDLength, MaxKeyIDLength, "circuit_id"); err != nil {
		return err
	}

	// ProofType enum is validated at protobuf level

	// Validate public parameters (required for ZK circuit verification)
	if err := validation.ValidateBytes(m.PublicParameters, 1, MaxCircuitSize, "public_parameters"); err != nil {
		return err
	}

	// Validate verification key (required for proof verification)
	if err := validation.ValidateBytes(m.VerificationKey, MinPublicKeySize, MaxPublicKeySize, "verification_key"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgSubmitZKProof
func (m *MsgSubmitZKProof) ValidateBasic() error {
	// Validate submitter address
	if err := validation.ValidateAccAddress(m.Submitter); err != nil {
		return fmt.Errorf("submitter: %w", err)
	}

	// Validate proof ID (references registered circuit)
	if err := validation.ValidateBoundedString(m.ProofId, MinKeyIDLength, MaxKeyIDLength, "proof_id"); err != nil {
		return err
	}

	// Validate proof data
	if err := validation.ValidateBytes(m.ProofData, MinProofSize, MaxProofSize, "proof_data"); err != nil {
		return err
	}

	// Validate public inputs (optional)
	if len(m.PublicInputs) > 0 {
		if err := validation.ValidateBytes(m.PublicInputs, 1, MaxProofSize, "public_inputs"); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgRegisterSecureEnclave
func (m *MsgRegisterSecureEnclave) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// EnclaveType enum is validated at protobuf level

	// Validate attestation data (hardware-backed cryptographic proof)
	if err := validation.ValidateBytes(m.AttestationData, 1, MaxAttestationSize, "attestation_data"); err != nil {
		return err
	}

	// enclave_metadata is optional map, validated by protobuf
	// Individual metadata values could be validated by keeper if needed

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgGenerateQuantumResistantKey
func (m *MsgGenerateQuantumResistantKey) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Algorithm enum is validated at protobuf level

	// expires_at is optional timestamp, validated by protobuf if present

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAddCertificatePin
func (m *MsgAddCertificatePin) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Validate hostname (FQDN format for certificate pinning)
	if err := validation.ValidateBoundedString(m.Hostname, 1, MaxHostnameLength, "hostname"); err != nil {
		return err
	}

	// Validate certificate hashes (at least one required for pinning)
	if len(m.CertificateHashes) == 0 {
		return fmt.Errorf("certificate_hashes: at least one certificate hash is required")
	}

	// Validate each certificate hash
	for i, hash := range m.CertificateHashes {
		if err := validation.ValidateBytes(hash, validation.MinHashLength, MaxCertificatePinSize, fmt.Sprintf("certificate_hashes[%d]", i)); err != nil {
			return err
		}
	}

	// PinType enum is validated at protobuf level

	// expires_at is optional timestamp, validated by protobuf if present

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateParams
func (m *MsgUpdateParams) ValidateBasic() error {
	// Validate authority address
	if err := validation.ValidateAccAddress(m.Authority); err != nil {
		return fmt.Errorf("authority: %w", err)
	}

	// Params validation would be done by the keeper
	// Here we just ensure authority is valid

	return nil
}
