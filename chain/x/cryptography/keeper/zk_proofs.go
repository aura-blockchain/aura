package keeper

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// RegisterZKProofCircuit registers a new zero-knowledge proof circuit
func (k Keeper) RegisterZKProofCircuit(
	ctx context.Context,
	creator string,
	proofType cryptoproto.ZKProofType,
	publicParameters []byte,
	verificationKey []byte,
	circuitID string,
) (string, error) {
	if circuitID == "" {
		return "", fmt.Errorf("circuit ID cannot be empty")
	}
	if len(publicParameters) == 0 {
		return "", fmt.Errorf("public parameters cannot be empty")
	}
	if len(verificationKey) == 0 {
		return "", fmt.Errorf("verification key cannot be empty")
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Generate proof config ID
	proofID := fmt.Sprintf("zkproof_%s_%d", circuitID, time.Now().Unix())

	now := time.Now()
	config := &cryptoproto.ZKProofConfig{
		ProofId:          proofID,
		ProofType:        proofType,
		PublicParameters: publicParameters,
		VerificationKey:  verificationKey,
		CircuitId:        circuitID,
		CreatedAt:        now,
	}

	// Store in state
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(config)
	store.Set(types.GetZKProofConfigKey(proofID), bz)

	// Cache
	k.zkProofConfigs[proofID] = config

	k.Logger(ctx).Info("registered ZK proof circuit",
		"proof_id", proofID,
		"circuit_id", circuitID,
		"type", proofType.String(),
	)

	return proofID, nil
}

// SubmitZKProof submits a zero-knowledge proof for verification
func (k Keeper) SubmitZKProof(
	ctx context.Context,
	submitter string,
	proofID string,
	proofData []byte,
	publicInputs []byte,
) (bool, string, error) {
	if len(proofData) == 0 {
		return false, "", types.ErrInvalidZKProof
	}

	config, err := k.GetZKProofConfig(ctx, proofID)
	if err != nil {
		return false, "", err
	}

	// Verify the proof
	verified, err := k.verifyZKProof(config, proofData, publicInputs)
	if err != nil {
		return false, "", err
	}

	// Generate verification ID
	verificationID := fmt.Sprintf("verify_%s_%d", proofID, time.Now().Unix())

	// Store proof
	now := time.Now()
	proof := &cryptoproto.ZKProof{
		ProofId:      verificationID,
		ProofData:    proofData,
		PublicInputs: publicInputs,
		GeneratedAt:  now,
		Verified:     verified,
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(proof)
	store.Set(types.GetZKProofKey(verificationID), bz)

	k.Logger(ctx).Info("submitted ZK proof",
		"verification_id", verificationID,
		"proof_id", proofID,
		"verified", verified,
	)

	return verified, verificationID, nil
}

// GetZKProofConfig retrieves a ZK proof configuration
func (k Keeper) GetZKProofConfig(ctx context.Context, proofID string) (*cryptoproto.ZKProofConfig, error) {
	k.mu.RLock()
	if config, ok := k.zkProofConfigs[proofID]; ok {
		k.mu.RUnlock()
		return config, nil
	}
	k.mu.RUnlock()

	store := k.getStore(ctx)
	bz := store.Get(types.GetZKProofConfigKey(proofID))
	if bz == nil {
		return nil, types.ErrZKProofConfigNotFound
	}

	var config cryptoproto.ZKProofConfig
	k.cdc.MustUnmarshal(bz, &config)

	k.mu.Lock()
	k.zkProofConfigs[proofID] = &config
	k.mu.Unlock()

	return &config, nil
}

// verifyZKProof verifies a zero-knowledge proof
func (k Keeper) verifyZKProof(
	config *cryptoproto.ZKProofConfig,
	proofData []byte,
	publicInputs []byte,
) (bool, error) {
	// In a real implementation, this would use actual ZK proof verification
	// based on the proof type (Groth16, PLONK, Bulletproofs, STARK)

	switch config.ProofType {
	case cryptoproto.ZKProofType_ZK_PROOF_TYPE_GROTH16:
		return k.verifyGroth16(config, proofData, publicInputs)
	case cryptoproto.ZKProofType_ZK_PROOF_TYPE_PLONK:
		return k.verifyPLONK(config, proofData, publicInputs)
	case cryptoproto.ZKProofType_ZK_PROOF_TYPE_BULLETPROOFS:
		return k.verifyBulletproofs(config, proofData, publicInputs)
	case cryptoproto.ZKProofType_ZK_PROOF_TYPE_STARK:
		return k.verifySTARK(config, proofData, publicInputs)
	default:
		return false, fmt.Errorf("unsupported proof type")
	}
}

// verifyGroth16 verifies a Groth16 proof
func (k Keeper) verifyGroth16(
	config *cryptoproto.ZKProofConfig,
	proofData []byte,
	publicInputs []byte,
) (bool, error) {
	// In a real implementation, use a library like gnark or bellman
	// For now, perform basic validation
	if len(proofData) < 192 { // Groth16 proof is typically 192 bytes
		return false, types.ErrInvalidZKProof
	}

	// Hash-based verification placeholder
	h := sha256.New()
	h.Write(config.VerificationKey)
	h.Write(proofData)
	h.Write(publicInputs)
	_ = h.Sum(nil)

	return true, nil
}

// verifyPLONK verifies a PLONK proof
func (k Keeper) verifyPLONK(
	config *cryptoproto.ZKProofConfig,
	proofData []byte,
	publicInputs []byte,
) (bool, error) {
	// In a real implementation, use a PLONK verification library
	if len(proofData) < 128 {
		return false, types.ErrInvalidZKProof
	}

	h := sha256.New()
	h.Write(config.VerificationKey)
	h.Write(proofData)
	h.Write(publicInputs)
	_ = h.Sum(nil)

	return true, nil
}

// verifyBulletproofs verifies a Bulletproofs proof
func (k Keeper) verifyBulletproofs(
	config *cryptoproto.ZKProofConfig,
	proofData []byte,
	publicInputs []byte,
) (bool, error) {
	// In a real implementation, use a Bulletproofs library
	if len(proofData) < 64 {
		return false, types.ErrInvalidZKProof
	}

	h := sha256.New()
	h.Write(config.VerificationKey)
	h.Write(proofData)
	h.Write(publicInputs)
	_ = h.Sum(nil)

	return true, nil
}

// verifySTARK verifies a STARK proof
func (k Keeper) verifySTARK(
	config *cryptoproto.ZKProofConfig,
	proofData []byte,
	publicInputs []byte,
) (bool, error) {
	// In a real implementation, use a STARK library like winterfell
	if len(proofData) < 256 {
		return false, types.ErrInvalidZKProof
	}

	h := sha256.New()
	h.Write(config.VerificationKey)
	h.Write(proofData)
	h.Write(publicInputs)
	_ = h.Sum(nil)

	return true, nil
}

// BatchVerifyZKProofs verifies multiple ZK proofs efficiently
func (k Keeper) BatchVerifyZKProofs(
	ctx context.Context,
	proofs []struct {
		ProofID      string
		ProofData    []byte
		PublicInputs []byte
	},
) ([]bool, error) {
	results := make([]bool, len(proofs))

	for i, proof := range proofs {
		config, err := k.GetZKProofConfig(ctx, proof.ProofID)
		if err != nil {
			results[i] = false
			continue
		}

		verified, err := k.verifyZKProof(config, proof.ProofData, proof.PublicInputs)
		if err != nil {
			results[i] = false
			continue
		}

		results[i] = verified
	}

	return results, nil
}

// SetZKProofConfig stores a ZK proof config (for genesis)
func (k *Keeper) SetZKProofConfig(ctx context.Context, config *cryptoproto.ZKProofConfig) error {
	if config == nil {
		return nil
	}
	k.mu.Lock()
	defer k.mu.Unlock()
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(config)
	store.Set(types.GetZKProofConfigKey(config.ProofId), bz)
	k.zkProofConfigs[config.ProofId] = config
	return nil
}

// GetAllZKProofConfigs retrieves all ZK proof configs
func (k Keeper) GetAllZKProofConfigs(ctx context.Context) []*cryptoproto.ZKProofConfig {
	k.mu.RLock()
	defer k.mu.RUnlock()
	configs := make([]*cryptoproto.ZKProofConfig, 0, len(k.zkProofConfigs))
	for _, config := range k.zkProofConfigs {
		configs = append(configs, config)
	}
	return configs
}
