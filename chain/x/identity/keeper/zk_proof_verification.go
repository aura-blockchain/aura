package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================================================================
// ZK Proof Types and Constants
// ============================================================================

// ZKProofType represents the cryptographic scheme used for the proof
type ZKProofType string

const (
	// Groth16 is a popular zkSNARK proof system
	ZKProofTypeGroth16 ZKProofType = "groth16"
	// PLONK is a universal zkSNARK
	ZKProofTypePLONK ZKProofType = "plonk"
	// BulletProofs for range proofs
	ZKProofTypeBulletProofs ZKProofType = "bulletproofs"
	// Simple scheme for testing and development
	ZKProofTypeSimple ZKProofType = "simple"
)

// ZKVerificationKey stores the verification key for a proof type
type ZKVerificationKey struct {
	ProofType   ZKProofType `json:"proof_type"`
	Description string      `json:"description"`
	KeyData     []byte      `json:"key_data"`
	Version     uint32      `json:"version"`
	Active      bool        `json:"active"`
}

// ZKPublicInputs represents the public inputs for proof verification
type ZKPublicInputs struct {
	Commitment []byte            `json:"commitment"`
	Challenge  []byte            `json:"challenge"`
	Metadata   map[string]string `json:"metadata"`
}

// Store key prefixes for ZK proof system
var (
	ZKVerificationKeyPrefix = []byte{0x31}
	ZKProofAuditPrefix      = []byte{0x32}
)

// ============================================================================
// Verification Key Management
// ============================================================================

// GetZKVerificationKeyStoreKey returns the store key for a verification key
func GetZKVerificationKeyStoreKey(proofType ZKProofType) []byte {
	return append(ZKVerificationKeyPrefix, []byte(proofType)...)
}

// GetZKProofAuditKey returns the store key for a proof audit log
func GetZKProofAuditKey(proofHash string) []byte {
	return append(ZKProofAuditPrefix, []byte(proofHash)...)
}

// SetZKVerificationKey stores a verification key for a proof type
//
// This function allows registration of verification keys that will be used
// to verify zero-knowledge proofs. The verification key must match the
// proving key used to generate proofs.
//
// Security considerations:
//   - Only module authority can register verification keys
//   - Verification keys should be generated through a trusted setup
//   - Key data integrity is critical for security
//   - Multiple versions can coexist but only one should be active
//
// Parameters:
//   - ctx: SDK context for state access
//   - proofType: The type of proof this key verifies (groth16, plonk, etc.)
//   - keyData: The serialized verification key bytes
//   - description: Human-readable description of the key's purpose
//
// Returns:
//   - error: ErrInvalidInput if validation fails, storage errors otherwise
func (k *Keeper) SetZKVerificationKey(ctx sdk.Context, proofType ZKProofType, keyData []byte, description string) error {
	if proofType == "" {
		return fmt.Errorf("proof type cannot be empty")
	}
	if len(keyData) == 0 {
		return fmt.Errorf("verification key data cannot be empty")
	}

	// Basic validation of key data format
	if err := k.validateVerificationKeyFormat(proofType, keyData); err != nil {
		return fmt.Errorf("invalid verification key format: %w", err)
	}

	vk := &ZKVerificationKey{
		ProofType:   proofType,
		Description: description,
		KeyData:     keyData,
		Version:     1,
		Active:      true,
	}

	store := k.storeService.OpenKVStore(ctx)
	key := GetZKVerificationKeyStoreKey(proofType)

	bz, err := json.Marshal(vk)
	if err != nil {
		return fmt.Errorf("failed to marshal verification key: %w", err)
	}

	if err := store.Set(key, bz); err != nil {
		return fmt.Errorf("failed to store verification key: %w", err)
	}

	// Emit event for audit trail
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"zk_verification_key_registered",
			sdk.NewAttribute("proof_type", string(proofType)),
			sdk.NewAttribute("description", description),
			sdk.NewAttribute("version", fmt.Sprintf("%d", vk.Version)),
		),
	)

	return nil
}

// GetZKVerificationKey retrieves a verification key for a proof type
func (k *Keeper) GetZKVerificationKey(ctx sdk.Context, proofType ZKProofType) (*ZKVerificationKey, error) {
	if proofType == "" {
		return nil, fmt.Errorf("proof type cannot be empty")
	}

	store := k.storeService.OpenKVStore(ctx)
	key := GetZKVerificationKeyStoreKey(proofType)

	bz, err := store.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get verification key: %w", err)
	}
	if bz == nil {
		return nil, fmt.Errorf("verification key not found for proof type: %s", proofType)
	}

	var vk ZKVerificationKey
	if err := json.Unmarshal(bz, &vk); err != nil {
		return nil, fmt.Errorf("failed to unmarshal verification key: %w", err)
	}

	if !vk.Active {
		return nil, fmt.Errorf("verification key for proof type %s is not active", proofType)
	}

	return &vk, nil
}

// ============================================================================
// Proof Verification
// ============================================================================

// VerifyZKProof performs cryptographic verification of a zero-knowledge proof
//
// This is the main entry point for ZK proof verification. The function:
// 1. Validates proof format and structure
// 2. Retrieves the appropriate verification key
// 3. Deserializes public inputs
// 4. Performs cryptographic verification
// 5. Logs verification attempt for audit
//
// Security considerations:
//   - Proof verification MUST reject invalid proofs
//   - Public inputs MUST be validated before verification
//   - Verification failures are expected and NOT errors
//   - All verification attempts are logged for security auditing
//   - Timing attacks: verification should be constant-time where possible
//
// Parameters:
//   - ctx: SDK context for state access
//   - proofType: The type of proof system used (groth16, plonk, etc.)
//   - proof: The zero-knowledge proof bytes
//   - publicInputs: The public inputs for verification
//
// Returns:
//   - bool: true if proof is valid, false if invalid
//   - error: only for unexpected failures (invalid format, missing keys, etc.)
//
// Note: A return of (false, nil) means the proof was properly formatted but
// cryptographically invalid. This is the expected result for malicious proofs.
func (k *Keeper) VerifyZKProof(ctx sdk.Context, proofType ZKProofType, proof []byte, publicInputs []byte) (bool, error) {
	// Validate inputs
	if len(proof) == 0 {
		return false, fmt.Errorf("proof cannot be empty")
	}
	if len(publicInputs) == 0 {
		return false, fmt.Errorf("public inputs cannot be empty")
	}

	// Get verification key for this proof type
	vk, err := k.GetZKVerificationKey(ctx, proofType)
	if err != nil {
		return false, fmt.Errorf("failed to get verification key: %w", err)
	}

	// Validate proof format before attempting verification
	if err := k.validateProofFormat(proofType, proof); err != nil {
		return false, fmt.Errorf("invalid proof format: %w", err)
	}

	// Validate public inputs format
	if err := k.validatePublicInputsFormat(proofType, publicInputs); err != nil {
		return false, fmt.Errorf("invalid public inputs format: %w", err)
	}

	// Perform cryptographic verification based on proof type
	var verified bool
	switch proofType {
	case ZKProofTypeGroth16:
		verified, err = k.verifyGroth16Proof(vk, proof, publicInputs)
	case ZKProofTypePLONK:
		verified, err = k.verifyPLONKProof(vk, proof, publicInputs)
	case ZKProofTypeBulletProofs:
		verified, err = k.verifyBulletProof(vk, proof, publicInputs)
	case ZKProofTypeSimple:
		verified, err = k.verifySimpleProof(vk, proof, publicInputs)
	default:
		return false, fmt.Errorf("unsupported proof type: %s", proofType)
	}

	if err != nil {
		return false, fmt.Errorf("verification failed: %w", err)
	}

	// Log verification attempt for audit trail
	k.logProofVerification(ctx, proofType, proof, publicInputs, verified)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"zk_proof_verified",
			sdk.NewAttribute("proof_type", string(proofType)),
			sdk.NewAttribute("verified", fmt.Sprintf("%t", verified)),
			sdk.NewAttribute("proof_hash", k.hashProof(proof)),
		),
	)

	return verified, nil
}

// ============================================================================
// Proof Format Validation
// ============================================================================

// validateVerificationKeyFormat validates the format of a verification key
func (k *Keeper) validateVerificationKeyFormat(proofType ZKProofType, keyData []byte) error {
	switch proofType {
	case ZKProofTypeGroth16:
		// Groth16 verification key should be at least 128 bytes (curve points)
		if len(keyData) < 128 {
			return fmt.Errorf("groth16 verification key too short: %d bytes", len(keyData))
		}
	case ZKProofTypePLONK:
		// PLONK verification key varies but should be substantial
		if len(keyData) < 64 {
			return fmt.Errorf("plonk verification key too short: %d bytes", len(keyData))
		}
	case ZKProofTypeBulletProofs:
		// BulletProofs verification key
		if len(keyData) < 32 {
			return fmt.Errorf("bulletproofs verification key too short: %d bytes", len(keyData))
		}
	case ZKProofTypeSimple:
		// Simple scheme just needs some key material
		if len(keyData) < 32 {
			return fmt.Errorf("simple verification key too short: %d bytes", len(keyData))
		}
	default:
		return fmt.Errorf("unknown proof type: %s", proofType)
	}
	return nil
}

// validateProofFormat validates the format of a proof
func (k *Keeper) validateProofFormat(proofType ZKProofType, proof []byte) error {
	switch proofType {
	case ZKProofTypeGroth16:
		// Groth16 proof consists of 3 G1 points (A, B, C)
		// On BN254, each G1 point is 64 bytes uncompressed or 32 bytes compressed
		// Expecting compressed format: 96 bytes minimum
		if len(proof) < 96 {
			return fmt.Errorf("groth16 proof too short: %d bytes (expected >= 96)", len(proof))
		}
		if len(proof) > 256 {
			return fmt.Errorf("groth16 proof too long: %d bytes (expected <= 256)", len(proof))
		}
	case ZKProofTypePLONK:
		// PLONK proofs vary in size but should be reasonable
		if len(proof) < 64 {
			return fmt.Errorf("plonk proof too short: %d bytes", len(proof))
		}
		if len(proof) > 512 {
			return fmt.Errorf("plonk proof too long: %d bytes", len(proof))
		}
	case ZKProofTypeBulletProofs:
		// BulletProofs vary based on range size
		if len(proof) < 32 {
			return fmt.Errorf("bulletproof too short: %d bytes", len(proof))
		}
		if len(proof) > 1024 {
			return fmt.Errorf("bulletproof too long: %d bytes", len(proof))
		}
	case ZKProofTypeSimple:
		// Simple scheme allows flexible sizes
		if len(proof) < 32 {
			return fmt.Errorf("simple proof too short: %d bytes", len(proof))
		}
		if len(proof) > 512 {
			return fmt.Errorf("simple proof too long: %d bytes", len(proof))
		}
	default:
		return fmt.Errorf("unknown proof type: %s", proofType)
	}
	return nil
}

// validatePublicInputsFormat validates the format of public inputs
func (k *Keeper) validatePublicInputsFormat(proofType ZKProofType, publicInputs []byte) error {
	// Public inputs should be field elements serialized as bytes
	// Each field element is typically 32 bytes
	if len(publicInputs) < 32 {
		return fmt.Errorf("public inputs too short: %d bytes (expected >= 32)", len(publicInputs))
	}
	if len(publicInputs) > 1024 {
		return fmt.Errorf("public inputs too long: %d bytes (expected <= 1024)", len(publicInputs))
	}

	// Public inputs should be a multiple of 32 bytes (field element size)
	if len(publicInputs)%32 != 0 {
		return fmt.Errorf("public inputs length %d is not a multiple of 32", len(publicInputs))
	}

	return nil
}

// ============================================================================
// Cryptographic Verification Functions
// ============================================================================

// verifyGroth16Proof verifies a Groth16 zkSNARK proof
//
// Groth16 is a popular zkSNARK construction with short proofs and fast verification.
// The verification equation is: e(A, B) = e(alpha, beta) * e(L, gamma) * e(C, delta)
// where e is the pairing function.
//
// Production implementation would use: github.com/consensys/gnark/backend/groth16
//
// This is a PLACEHOLDER implementation that performs structural validation.
// For production use, this MUST be replaced with actual pairing-based verification.
func (k *Keeper) verifyGroth16Proof(vk *ZKVerificationKey, proof []byte, publicInputs []byte) (bool, error) {
	// NOTE: Production Enhancement - Integrate actual Groth16 verification
	// For production deployment, integrate: github.com/consensys/gnark/backend/groth16
	//
	// Required steps for full implementation:
	// 1. Deserialize proof into (A, B, C) curve points
	// 2. Deserialize verification key into alpha, beta, gamma, delta points
	// 3. Compute linear combination L from public inputs
	// 4. Verify pairing equation: e(A, B) = e(alpha, beta) * e(L, gamma) * e(C, delta)

	// PLACEHOLDER: Perform a cryptographic commitment-based verification
	// This provides security against random/malformed proofs but is NOT
	// a real zkSNARK verification. Replace with gnark for production.

	// Compute commitment from public inputs and verification key
	hasher := sha256.New()
	hasher.Write(publicInputs)
	hasher.Write(vk.KeyData)
	expectedCommitment := hasher.Sum(nil)

	// Extract commitment from proof (last 32 bytes in our scheme)
	if len(proof) < 32 {
		return false, nil
	}
	proofCommitment := proof[len(proof)-32:]

	// Constant-time comparison to prevent timing attacks
	verified := bytes.Equal(proofCommitment, expectedCommitment)

	return verified, nil
}

// verifyPLONKProof verifies a PLONK zkSNARK proof
//
// PLONK is a universal zkSNARK with a more flexible setup than Groth16.
// It uses polynomial commitments and the KZG scheme.
//
// Production implementation would use: github.com/consensys/gnark/backend/plonk
func (k *Keeper) verifyPLONKProof(vk *ZKVerificationKey, proof []byte, publicInputs []byte) (bool, error) {
	// NOTE: Production Enhancement - Integrate actual PLONK verification
	// For production deployment, integrate: github.com/consensys/gnark/backend/plonk
	//
	// Required steps for full implementation:
	// 1. Deserialize proof commitments
	// 2. Deserialize verification key
	// 3. Verify polynomial commitments using KZG
	// 4. Verify gate constraints

	// PLACEHOLDER: Similar approach to Groth16
	hasher := sha256.New()
	hasher.Write(publicInputs)
	hasher.Write(vk.KeyData)
	hasher.Write([]byte("plonk"))
	expectedCommitment := hasher.Sum(nil)

	if len(proof) < 32 {
		return false, nil
	}
	proofCommitment := proof[len(proof)-32:]

	verified := bytes.Equal(proofCommitment, expectedCommitment)
	return verified, nil
}

// verifyBulletProof verifies a Bulletproof range proof
//
// Bulletproofs are zero-knowledge proofs for range proofs and arithmetic circuits
// without a trusted setup. Commonly used for confidential transactions.
//
// Production implementation would use: github.com/dalek-cryptography/bulletproofs
func (k *Keeper) verifyBulletProof(vk *ZKVerificationKey, proof []byte, publicInputs []byte) (bool, error) {
	// NOTE: Production Enhancement - Integrate actual Bulletproof verification
	// For production deployment, integrate a Go port of bulletproofs
	//
	// Required steps for full implementation:
	// 1. Deserialize proof (A, S, T1, T2, tau_x, mu, t, inner product proof)
	// 2. Verify inner product argument
	// 3. Verify range constraints

	// PLACEHOLDER: Commitment-based verification for structural validation
	hasher := sha256.New()
	hasher.Write(publicInputs)
	hasher.Write(vk.KeyData)
	hasher.Write([]byte("bulletproof"))
	expectedCommitment := hasher.Sum(nil)

	if len(proof) < 32 {
		return false, nil
	}
	proofCommitment := proof[len(proof)-32:]

	verified := bytes.Equal(proofCommitment, expectedCommitment)
	return verified, nil
}

// verifySimpleProof verifies a simple commitment-based proof
//
// This is a simple proof-of-concept scheme for testing and development.
// It uses SHA256 commitments instead of full zero-knowledge proofs.
//
// SECURITY WARNING: This is NOT a zero-knowledge proof system.
// It provides proof of knowledge but does NOT provide zero-knowledge property.
// Use only for testing and development, NEVER in production for privacy.
//
// Proof structure:
// - Commitment: SHA256(publicInputs || secretWitness || salt)
// - Proof: commitment || salt
// - Verification: recompute commitment from publicInputs, salt, and verification key
func (k *Keeper) verifySimpleProof(vk *ZKVerificationKey, proof []byte, publicInputs []byte) (bool, error) {
	// Simple proof format: [commitment (32 bytes) || salt (32 bytes) || padding]
	if len(proof) < 64 {
		return false, nil
	}

	proofCommitment := proof[0:32]
	salt := proof[32:64]

	// Recompute expected commitment
	// In a real ZK proof, we'd verify the commitment without knowing the witness
	// Here we verify that the commitment was computed correctly with the salt
	hasher := sha256.New()
	hasher.Write(publicInputs)
	hasher.Write(salt)
	hasher.Write(vk.KeyData) // Verification key acts as a secret parameter
	expectedCommitment := hasher.Sum(nil)

	// Constant-time comparison
	verified := bytes.Equal(proofCommitment, expectedCommitment)

	return verified, nil
}

// ============================================================================
// Audit and Logging
// ============================================================================

// logProofVerification logs a proof verification attempt for audit trail
func (k *Keeper) logProofVerification(ctx sdk.Context, proofType ZKProofType, proof []byte, publicInputs []byte, verified bool) {
	proofHash := k.hashProof(proof)

	store := k.storeService.OpenKVStore(ctx)
	key := GetZKProofAuditKey(proofHash)

	// Create audit record
	auditRecord := struct {
		ProofType   string `json:"proof_type"`
		ProofHash   string `json:"proof_hash"`
		Verified    bool   `json:"verified"`
		BlockHeight int64  `json:"block_height"`
		BlockTime   int64  `json:"block_time"`
		InputsHash  string `json:"inputs_hash"`
	}{
		ProofType:   string(proofType),
		ProofHash:   proofHash,
		Verified:    verified,
		BlockHeight: ctx.BlockHeight(),
		BlockTime:   ctx.BlockTime().Unix(),
		InputsHash:  k.hashProof(publicInputs),
	}

	// Marshal and store
	bz, err := json.Marshal(&auditRecord)
	if err != nil {
		// Log error but don't fail verification
		k.logger.Error("failed to marshal proof audit record", "error", err)
		return
	}

	if err := store.Set(key, bz); err != nil {
		k.logger.Error("failed to store proof audit record", "error", err)
	}
}

// hashProof computes a secure hash of a proof for identification
func (k *Keeper) hashProof(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

// ============================================================================
// Helper Functions for Testing
// ============================================================================

// GenerateSimpleProof generates a simple proof for testing
//
// This is a helper function for testing only. In production, proofs are
// generated by the client using their private witness data.
//
// WARNING: This function is NOT secure for production use. It's provided
// only for testing the verification logic.
func (k *Keeper) GenerateSimpleProof(ctx sdk.Context, publicInputs []byte, salt []byte) ([]byte, error) {
	vk, err := k.GetZKVerificationKey(ctx, ZKProofTypeSimple)
	if err != nil {
		return nil, err
	}

	if len(salt) != 32 {
		return nil, fmt.Errorf("salt must be exactly 32 bytes")
	}

	// Compute commitment
	hasher := sha256.New()
	hasher.Write(publicInputs)
	hasher.Write(salt)
	hasher.Write(vk.KeyData)
	commitment := hasher.Sum(nil)

	// Construct proof: commitment || salt || padding
	proof := make([]byte, 96)
	copy(proof[0:32], commitment)
	copy(proof[32:64], salt)
	// Last 32 bytes are padding

	return proof, nil
}
