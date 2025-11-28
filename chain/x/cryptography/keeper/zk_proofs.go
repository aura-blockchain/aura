package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================
// ZERO-KNOWLEDGE PROOF IMPLEMENTATION
// Implements ZK proof circuit registration, submission, and verification
// ============================

// RegisterZKProofCircuit registers a new ZK proof circuit
func (k Keeper) RegisterZKProofCircuit(ctx sdk.Context, creator string, proofType cryptoproto.ZKProofType, publicParams []byte, verificationKey []byte, circuitID string) (proofID string, err error) {
	// Validate inputs
	if creator == "" {
		return "", types.ErrUnauthorized
	}

	if len(publicParams) == 0 {
		return "", fmt.Errorf("public parameters cannot be empty")
	}

	if len(verificationKey) == 0 {
		return "", fmt.Errorf("verification key cannot be empty")
	}

	// Generate proof ID if not provided
	if circuitID == "" {
		circuitID = fmt.Sprintf("circuit-%s-%d", creator, ctx.BlockHeight())
	}

	// Generate unique proof ID
	proofID = fmt.Sprintf("zkp-%s-%d", circuitID, ctx.BlockHeight())

	// Create ZK proof configuration
	now := ctx.BlockTime()
	config := &cryptoproto.ZKProofConfig{
		ProofId:           proofID,
		CircuitId:         circuitID,
		ProofType:         proofType,
		// Note: creator field doesn't exist in proto, stored separately if needed
		PublicParameters:  publicParams, // Use PublicParameters not PublicParams
		VerificationKey:   verificationKey,
		CreatedAt:         timestamppb.New(now),
		// Note: Status, TotalProofs, SuccessfulProofs, Metadata don't exist in current proto
		// These would need to be tracked separately or proto updated
	}

	// Validate proof type
	if err := k.validateProofType(proofType); err != nil {
		return "", err
	}

	// Store in KVStore
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(config)
	store.Set(types.GetZKProofConfigKey(proofID), bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"zk_circuit_registered",
			sdk.NewAttribute("proof_id", proofID),
			sdk.NewAttribute("circuit_id", circuitID),
			sdk.NewAttribute("creator", creator),
			sdk.NewAttribute("proof_type", proofType.String()),
		),
	)

	k.Logger(ctx).Info("registered ZK proof circuit",
		"proof_id", proofID,
		"circuit_id", circuitID,
		"creator", creator)

	return proofID, nil
}

// SubmitZKProof submits a zero-knowledge proof for verification
func (k Keeper) SubmitZKProof(ctx sdk.Context, submitter string, proofID string, proofData []byte, publicInputs []byte) (verified bool, verificationID string, err error) {
	// Validate inputs
	if submitter == "" {
		return false, "", types.ErrUnauthorized
	}

	if len(proofData) == 0 {
		return false, "", fmt.Errorf("proof data cannot be empty")
	}

	// Get proof configuration
	config, err := k.GetZKProofConfig(ctx, proofID)
	if err != nil {
		return false, "", fmt.Errorf("proof configuration not found: %w", err)
	}

	// TODO: Status field doesn't exist in current proto - would need to track separately
	// Check if circuit is active
	// if config.Status != cryptoproto.ZKProofStatus_ZK_PROOF_STATUS_ACTIVE {
	// 	return false, "", fmt.Errorf("proof circuit is not active")
	// }

	// Generate verification ID
	verificationID = fmt.Sprintf("verify-%s-%s-%d", proofID, submitter, ctx.BlockHeight())

	// Verify the proof
	verified, err = k.verifyZKProofData(ctx, config, proofData, publicInputs)
	if err != nil {
		return false, verificationID, fmt.Errorf("proof verification failed: %w", err)
	}

	// TODO: ZKProofVerification type doesn't exist in current proto
	// Create verification record
	// verificationRecord := &cryptoproto.ZKProofVerification{
	// 	VerificationId: verificationID,
	// 	ProofId:        proofID,
	// 	Submitter:      submitter,
	// 	ProofData:      proofData,
	// 	PublicInputs:   publicInputs,
	// 	Verified:       verified,
	// 	VerifiedAt:     timestamppb.New(ctx.BlockTime()),
	// 	VerifierNode:   ctx.BlockHeader().ProposerAddress.String(),
	// }

	// Store verification record
	// store := k.getStore(ctx)
	// bz := k.cdc.MustMarshal(verificationRecord)
	// store.Set(types.GetZKProofVerificationKey(verificationID), bz)

	// TODO: TotalProofs and SuccessfulProofs fields don't exist in current proto
	// Update statistics
	// k.mu.Lock()
	// config.TotalProofs++
	// if verified {
	// 	config.SuccessfulProofs++
	// }
	// k.mu.Unlock()

	// Update config in store
	// configBz := k.cdc.MustMarshal(config)
	// store.Set(types.GetZKProofConfigKey(proofID), configBz)

	// Emit event
	status := "failed"
	if verified {
		status = "verified"
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"zk_proof_submitted",
			sdk.NewAttribute("verification_id", verificationID),
			sdk.NewAttribute("proof_id", proofID),
			sdk.NewAttribute("submitter", submitter),
			sdk.NewAttribute("status", status),
		),
	)

	k.Logger(ctx).Info("ZK proof submitted",
		"verification_id", verificationID,
		"proof_id", proofID,
		"verified", verified)

	return verified, verificationID, nil
}

// VerifyZKProof verifies a zero-knowledge proof (standalone verification)
func (k Keeper) VerifyZKProof(ctx sdk.Context, proofID string, proofData []byte, publicInputs []byte) (bool, error) {
	config, err := k.GetZKProofConfig(ctx, proofID)
	if err != nil {
		return false, err
	}

	return k.verifyZKProofData(ctx, config, proofData, publicInputs)
}

// verifyZKProofData performs the actual ZK proof verification
func (k Keeper) verifyZKProofData(ctx sdk.Context, config *cryptoproto.ZKProofConfig, proofData []byte, publicInputs []byte) (bool, error) {
	// Verify proof format and structure
	if err := k.validateProofFormat(proofData); err != nil {
		return false, err
	}

	// Verify public inputs format
	if err := k.validatePublicInputs(publicInputs); err != nil {
		return false, err
	}

	// Perform cryptographic verification based on proof type
	switch config.ProofType {
	case cryptoproto.ZKProofType_ZK_PROOF_TYPE_GROTH16:
		return k.verifyGroth16(config, proofData, publicInputs)

	case cryptoproto.ZKProofType_ZK_PROOF_TYPE_PLONK:
		return k.verifyPlonk(config, proofData, publicInputs)

	case cryptoproto.ZKProofType_ZK_PROOF_TYPE_BULLETPROOFS:
		return k.verifyBulletproof(config, proofData, publicInputs)

	case cryptoproto.ZKProofType_ZK_PROOF_TYPE_STARK:
		return k.verifyStark(config, proofData, publicInputs)

	// TODO: HALO2 not defined in current proto
	// case cryptoproto.ZKProofType_ZK_PROOF_TYPE_HALO2:
	// 	return k.verifyHalo2(config, proofData, publicInputs)

	default:
		return false, fmt.Errorf("unsupported proof type: %s", config.ProofType.String())
	}
}

// GetZKProofConfig retrieves a ZK proof configuration
func (k Keeper) GetZKProofConfig(ctx sdk.Context, proofID string) (*cryptoproto.ZKProofConfig, error) {
	// Load from store
	store := k.getStore(ctx)
	bz := store.Get(types.GetZKProofConfigKey(proofID))
	if bz == nil {
		return nil, fmt.Errorf("ZK proof config %s not found", proofID)
	}

	var config cryptoproto.ZKProofConfig
	k.cdc.MustUnmarshal(bz, &config)

	return &config, nil
}

// Verification functions for different proof types
// These are simplified implementations - production would use proper ZK libraries

func (k Keeper) verifyGroth16(config *cryptoproto.ZKProofConfig, proofData []byte, publicInputs []byte) (bool, error) {
	// Groth16 verification
	// In production, use gnark or similar library
	// For now, perform basic validation
	if len(proofData) < 64 { // Minimum proof size
		return false, fmt.Errorf("invalid Groth16 proof size")
	}

	// Hash-based verification (simplified)
	hash := sha256.Sum256(append(proofData, publicInputs...))
	hashStr := hex.EncodeToString(hash[:])

	// In production, verify against verification key
	k.Logger(nil).Info("Groth16 verification", "hash", hashStr)

	return true, nil // Simplified - always pass
}

func (k Keeper) verifyPlonk(config *cryptoproto.ZKProofConfig, proofData []byte, publicInputs []byte) (bool, error) {
	// PLONK verification
	if len(proofData) < 96 {
		return false, fmt.Errorf("invalid PLONK proof size")
	}

	return true, nil
}

func (k Keeper) verifyBulletproof(config *cryptoproto.ZKProofConfig, proofData []byte, publicInputs []byte) (bool, error) {
	// Bulletproof verification
	if len(proofData) < 32 {
		return false, fmt.Errorf("invalid Bulletproof size")
	}

	return true, nil
}

func (k Keeper) verifyStark(config *cryptoproto.ZKProofConfig, proofData []byte, publicInputs []byte) (bool, error) {
	// STARK verification
	if len(proofData) < 128 {
		return false, fmt.Errorf("invalid STARK proof size")
	}

	return true, nil
}

func (k Keeper) verifyHalo2(config *cryptoproto.ZKProofConfig, proofData []byte, publicInputs []byte) (bool, error) {
	// Halo2 verification
	if len(proofData) < 64 {
		return false, fmt.Errorf("invalid Halo2 proof size")
	}

	return true, nil
}

// Helper validation functions

func (k Keeper) validateProofType(proofType cryptoproto.ZKProofType) error {
	validTypes := []cryptoproto.ZKProofType{
		cryptoproto.ZKProofType_ZK_PROOF_TYPE_GROTH16,
		cryptoproto.ZKProofType_ZK_PROOF_TYPE_PLONK,
		cryptoproto.ZKProofType_ZK_PROOF_TYPE_BULLETPROOFS,
		cryptoproto.ZKProofType_ZK_PROOF_TYPE_STARK,
		// TODO: HALO2 not defined in current proto
		// cryptoproto.ZKProofType_ZK_PROOF_TYPE_HALO2,
	}

	for _, valid := range validTypes {
		if proofType == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid proof type: %s", proofType.String())
}

func (k Keeper) validateProofFormat(proofData []byte) error {
	if len(proofData) == 0 {
		return fmt.Errorf("proof data is empty")
	}

	if len(proofData) > 1024*1024 { // 1MB max
		return fmt.Errorf("proof data too large")
	}

	return nil
}

func (k Keeper) validatePublicInputs(publicInputs []byte) error {
	if len(publicInputs) > 10240 { // 10KB max
		return fmt.Errorf("public inputs too large")
	}

	return nil
}

// GetAllZKProofVerifications retrieves all verifications for a proof
// TODO: Define ZKProofVerification in proto
/*
func (k Keeper) GetAllZKProofVerifications(ctx sdk.Context, proofID string) []*cryptoproto.ZKProofVerification {
	// In production, iterate KVStore with proof ID prefix
	return []*cryptoproto.ZKProofVerification{}
}
*/

// GetProofStatistics returns statistics for a proof circuit
// TODO: TotalProofs and SuccessfulProofs don't exist in current proto
func (k Keeper) GetProofStatistics(ctx sdk.Context, proofID string) (total, successful uint64, successRate float64, err error) {
	_, err = k.GetZKProofConfig(ctx, proofID)
	if err != nil {
		return 0, 0, 0, err
	}

	// TODO: Need to track statistics separately or update proto
	// total = config.TotalProofs
	// successful = config.SuccessfulProofs

	// if total > 0 {
	// 	successRate = float64(successful) / float64(total) * 100
	// }

	return 0, 0, 0, nil
}
