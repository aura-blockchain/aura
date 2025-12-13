package keeper

import (
	storetypes "cosmossdk.io/store/types"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================
// SCORE VERIFICATION PROOFS
// Implements cryptographic proofs for score claims
// ============================

// ScoreProof represents a cryptographic proof of a score claim
type ScoreProof struct {
	WalletAddress string
	ClaimedScore  uint64
	ProofHash     string
	Timestamp     int64
	BlockHeight   uint64
	Signature     []byte
	MerkleRoot    string
	MerkleProof   []string
}

// GenerateScoreProof creates a cryptographic proof for a user's score
func (k *Keeper) GenerateScoreProof(ctx sdk.Context, walletAddr string) (*ScoreProof, error) {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return nil, types.ErrUserRecordNotFound
	}

	// Create proof data
	proofData := fmt.Sprintf("%s:%d:%d:%d",
		walletAddr,
		record.TotalScore,
		ctx.BlockHeight(),
		ctx.BlockTime().Unix(),
	)

	// Generate proof hash
	hash := sha256.Sum256([]byte(proofData))
	proofHash := hex.EncodeToString(hash[:])

	// Generate Merkle proof (simplified - in production use proper Merkle tree)
	merkleRoot, merkleProof := k.generateMerkleProof(ctx, walletAddr, record.TotalScore)

	proof := &ScoreProof{
		WalletAddress: walletAddr,
		ClaimedScore:  record.TotalScore,
		ProofHash:     proofHash,
		Timestamp:     ctx.BlockTime().Unix(),
		BlockHeight:   uint64(ctx.BlockHeight()),
		Signature:     []byte{}, // In production, sign with module key
		MerkleRoot:    merkleRoot,
		MerkleProof:   merkleProof,
	}

	// Store proof hash
	k.storeVerificationProofHash(ctx, walletAddr, proofHash)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"score_proof_generated",
			sdk.NewAttribute("wallet", walletAddr),
			sdk.NewAttribute("score", fmt.Sprintf("%d", record.TotalScore)),
			sdk.NewAttribute("proof_hash", proofHash),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return proof, nil
}

// VerifyScoreProof verifies a score proof
func (k *Keeper) VerifyScoreProof(ctx sdk.Context, proof *ScoreProof) (bool, error) {
	// Get current score
	record, ok := k.GetUserRecord(ctx, proof.WalletAddress)
	if !ok {
		return false, types.ErrUserRecordNotFound
	}

	// Check if proof hash was previously generated
	if !k.hasVerificationProofHash(ctx, proof.WalletAddress, proof.ProofHash) {
		return false, fmt.Errorf("proof hash not found in records")
	}

	// Verify proof data integrity
	expectedProofData := fmt.Sprintf("%s:%d:%d:%d",
		proof.WalletAddress,
		proof.ClaimedScore,
		proof.BlockHeight,
		proof.Timestamp,
	)

	hash := sha256.Sum256([]byte(expectedProofData))
	expectedHash := hex.EncodeToString(hash[:])

	if expectedHash != proof.ProofHash {
		return false, fmt.Errorf("proof hash mismatch")
	}

	// Verify Merkle proof
	if !k.verifyMerkleProof(proof.MerkleRoot, proof.MerkleProof, proof.WalletAddress, proof.ClaimedScore) {
		return false, fmt.Errorf("Merkle proof verification failed")
	}

	// Check if score has changed since proof was generated
	if record.TotalScore != proof.ClaimedScore {
		return false, fmt.Errorf("score has changed since proof generation: claimed=%d, current=%d",
			proof.ClaimedScore, record.TotalScore)
	}

	return true, nil
}

// storeVerificationProofHash stores a proof hash for a user
func (k *Keeper) storeVerificationProofHash(ctx sdk.Context, walletAddr string, proofHash string) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.VerificationProofHashStoreKey(walletAddr, proofHash)
	if err := store.Set([]byte(key), []byte{1}); err != nil {
		ctx.Logger().Error("failed to store verification proof hash", "wallet", walletAddr, "error", err)
	}
}

// hasVerificationProofHash checks if a proof hash exists for a user
func (k *Keeper) hasVerificationProofHash(ctx sdk.Context, walletAddr string, proofHash string) bool {
	store := k.storeService.OpenKVStore(ctx)
	key := types.VerificationProofHashStoreKey(walletAddr, proofHash)
	bz, err := store.Get([]byte(key))
	return err == nil && len(bz) > 0
}

// generateMerkleProof generates a Merkle proof for a score claim
// Simplified implementation - in production use proper Merkle tree library
func (k *Keeper) generateMerkleProof(ctx sdk.Context, walletAddr string, score uint64) (root string, proof []string) {
	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.UserRecordStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return "", []string{}
	}
	defer iterator.Close()

	// Collect all user scores for Merkle tree
	leaves := []string{}
	for ; iterator.Valid(); iterator.Next() {
		var record types.UserConfidenceRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}

		leaf := fmt.Sprintf("%s:%d", record.WalletAddress, record.TotalScore)
		hash := sha256.Sum256([]byte(leaf))
		leaves = append(leaves, hex.EncodeToString(hash[:]))
	}

	if len(leaves) == 0 {
		return "", []string{}
	}

	// Build simple Merkle tree (in production use optimized library)
	// For now, just return a mock root and proof
	allLeaves := ""
	for _, leaf := range leaves {
		allLeaves += leaf
	}
	rootHash := sha256.Sum256([]byte(allLeaves))
	root = hex.EncodeToString(rootHash[:])

	// Mock proof path (in production, generate actual path)
	proof = []string{leaves[0]} // Simplified

	return root, proof
}

// verifyMerkleProof verifies a Merkle proof
// Simplified implementation - in production use proper Merkle tree verification
func (k *Keeper) verifyMerkleProof(root string, proof []string, walletAddr string, score uint64) bool {
	// In production, implement proper Merkle proof verification
	// For now, just check that proof is non-empty
	return len(proof) > 0 && root != ""
}

// GetProofHistory returns all proof hashes for a user
func (k *Keeper) GetProofHistory(ctx sdk.Context, walletAddr string) []string {
	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.VerificationProofHashStoreKeyPrefix + walletAddr + "/")
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return []string{}
	}
	defer iterator.Close()

	proofs := []string{}
	for ; iterator.Valid(); iterator.Next() {
		// Extract proof hash from key
		key := string(iterator.Key())
		parts := splitKey(key, "/")
		if len(parts) > 1 {
			proofs = append(proofs, parts[len(parts)-1])
		}
	}

	return proofs
}

// GenerateBatchProofs generates proofs for multiple users
func (k *Keeper) GenerateBatchProofs(ctx sdk.Context, walletAddrs []string) (map[string]*ScoreProof, error) {
	proofs := make(map[string]*ScoreProof)

	for _, addr := range walletAddrs {
		proof, err := k.GenerateScoreProof(ctx, addr)
		if err != nil {
			continue // Skip users with errors
		}
		proofs[addr] = proof
	}

	return proofs, nil
}

// ExportVerifiableScores exports all scores with cryptographic proofs
// Useful for audits and third-party verification
func (k *Keeper) ExportVerifiableScores(ctx sdk.Context) (map[string]interface{}, error) {
	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.UserRecordStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	export := map[string]interface{}{
		"export_height": ctx.BlockHeight(),
		"export_time":   ctx.BlockTime().Unix(),
		"scores":        []map[string]interface{}{},
	}

	scores := []map[string]interface{}{}

	for ; iterator.Valid(); iterator.Next() {
		var record types.UserConfidenceRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}

		proof, err := k.GenerateScoreProof(ctx, record.WalletAddress)
		if err != nil {
			continue
		}

		scores = append(scores, map[string]interface{}{
			"wallet":      record.WalletAddress,
			"score":       record.TotalScore,
			"status":      record.Status,
			"proof_hash":  proof.ProofHash,
			"merkle_root": proof.MerkleRoot,
		})
	}

	export["scores"] = scores
	export["total_users"] = len(scores)

	return export, nil
}

// Helper function to split key by delimiter
func splitKey(s, delim string) []string {
	var result []string
	var current string
	for _, c := range s {
		if string(c) == delim {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
