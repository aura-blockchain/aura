package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

// ============================================================================
// Signature Verification Benchmarks
// ============================================================================

// BenchmarkSignatureVerification benchmarks ECDSA signature verification
func BenchmarkSignatureVerification(b *testing.B) {
	keeper, ctx := setupKeeperForBenchmark(b)

	// Create a test message and signature
	message := []byte("test message for signature verification")
	msgHash := sha256.Sum256(message)

	// Generate a test signature (65 bytes: R=32, S=32, V=1)
	signature := make([]byte, 65)
	for i := range signature[:64] {
		signature[i] = byte(i % 256)
	}
	signature[64] = 0 // recovery ID

	auraAddress := "aura1test"
	pawAddress := "paw1test"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.VerifyPawAddressOwnership(ctx, auraAddress, pawAddress, signature)
	}
}

// BenchmarkSignatureVerification_Batch benchmarks batch signature verification
func BenchmarkSignatureVerification_Batch(b *testing.B) {
	keeper, ctx := setupKeeperForBenchmark(b)

	// Create 10 signatures for batch verification
	signatures := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		sig := make([]byte, 65)
		for j := range sig[:64] {
			sig[j] = byte((i + j) % 256)
		}
		sig[64] = 0
		signatures[i] = sig
	}

	auraAddress := "aura1test"
	pawAddresses := make([]string, 10)
	for i := 0; i < 10; i++ {
		pawAddresses[i] = fmt.Sprintf("paw1test%d", i)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			_ = keeper.VerifyPawAddressOwnership(ctx, auraAddress, pawAddresses[j], signatures[j])
		}
	}
}

// ============================================================================
// Lock/Unlock Token Benchmarks
// ============================================================================

// BenchmarkLockTokens benchmarks the token locking operation
func BenchmarkLockTokens(b *testing.B) {
	keeper, ctx := setupKeeperForBenchmark(b)

	sender := "aura1sender"
	targetChain := "paw"
	targetAddress := "paw1target"
	amount := sdkmath.NewInt(1000000)
	denom := "uaura"

	// Setup chain
	chain := &types.SupportedChain{
		ChainId:          "paw",
		ChainName:        "PAW",
		IsEnabled:        true,
		MinTransferLimit: "1000",
		MaxTransferLimit: "1000000000",
	}
	keeper.setChain(ctx, chain)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Reset state between iterations
		ctx = sdk.NewContext(ctx.MultiStore(), ctx.BlockHeader(), false, keeper.Logger(ctx))
		b.StartTimer()

		_, _ = keeper.lockTokens(ctx, sender, targetChain, targetAddress, amount, denom)
	}
}

// BenchmarkUnlockTokens benchmarks the token unlocking operation
func BenchmarkUnlockTokens(b *testing.B) {
	keeper, ctx := setupKeeperForBenchmark(b)

	transferID := "transfer-1"
	recipient := "aura1recipient"
	amount := sdkmath.NewInt(1000000)
	denom := "uaura"

	// Create a transfer to unlock
	transfer := &types.CrossChainTransfer{
		TransferId:    transferID,
		SourceChain:   "paw",
		TargetChain:   "aura",
		Sender:        "paw1sender",
		Recipient:     recipient,
		Amount:        amount,
		Denom:         denom,
		Status:        types.TransferStatusPending,
		BlockHeight:   ctx.BlockHeight(),
		SourceBlockHeight: 100,
	}
	keeper.setTransfer(ctx, transfer)

	// Create validator signatures
	validators := []string{"val1", "val2", "val3"}
	signatures := [][]byte{
		make([]byte, 65),
		make([]byte, 65),
		make([]byte, 65),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		keeper.setTransfer(ctx, transfer)
		b.StartTimer()

		_ = keeper.unlockTokens(ctx, transferID, recipient, validators, signatures)
	}
}

// ============================================================================
// Batch Transfer Benchmarks
// ============================================================================

// BenchmarkBatchTransfers benchmarks processing multiple transfers
func BenchmarkBatchTransfers(b *testing.B) {
	keeper, ctx := setupKeeperForBenchmark(b)

	// Setup chain
	chain := &types.SupportedChain{
		ChainId:          "paw",
		ChainName:        "PAW",
		IsEnabled:        true,
		MinTransferLimit: "1000",
		MaxTransferLimit: "1000000000",
	}
	keeper.setChain(ctx, chain)

	// Create 100 transfers
	transfers := make([]*types.CrossChainTransfer, 100)
	for i := 0; i < 100; i++ {
		transfers[i] = &types.CrossChainTransfer{
			TransferId:    fmt.Sprintf("transfer-%d", i),
			SourceChain:   "aura",
			TargetChain:   "paw",
			Sender:        fmt.Sprintf("aura1sender%d", i),
			Recipient:     fmt.Sprintf("paw1recipient%d", i),
			Amount:        sdkmath.NewInt(1000000),
			Denom:         "uaura",
			Status:        types.TransferStatusPending,
			BlockHeight:   ctx.BlockHeight(),
			SourceBlockHeight: uint64(ctx.BlockHeight()),
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Store all transfers
		for _, transfer := range transfers {
			keeper.setTransfer(ctx, transfer)
		}
		b.StartTimer()

		// Process all transfers (retrieve and validate)
		for _, transfer := range transfers {
			_, _ = keeper.getTransfer(ctx, transfer.TransferId)
		}
	}
}

// ============================================================================
// Merkle Proof Verification Benchmarks
// ============================================================================

// BenchmarkMerkleProofVerification benchmarks Merkle proof verification
func BenchmarkMerkleProofVerification(b *testing.B) {
	keeper, ctx := setupKeeperForBenchmark(b)

	// Create a simple Merkle tree with 8 leaves
	leaves := make([][]byte, 8)
	for i := 0; i < 8; i++ {
		hash := sha256.Sum256([]byte(fmt.Sprintf("leaf-%d", i)))
		leaves[i] = hash[:]
	}

	// Build Merkle root
	merkleRoot := keeper.computeMerkleRoot(leaves)

	// Create proof for leaf 0
	transactionLeaf := leaves[0]
	merkleProof := keeper.buildMerkleProof(leaves, 0)
	merkleProofBytes := serializeMerkleProof(merkleProof)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.VerifyMerkleProofBytes(merkleRoot, transactionLeaf, merkleProofBytes)
	}
}

// BenchmarkMerkleProofVerification_LargeTree benchmarks Merkle proof verification for large tree
func BenchmarkMerkleProofVerification_LargeTree(b *testing.B) {
	keeper, ctx := setupKeeperForBenchmark(b)

	// Create a Merkle tree with 1024 leaves
	leaves := make([][]byte, 1024)
	for i := 0; i < 1024; i++ {
		hash := sha256.Sum256([]byte(fmt.Sprintf("leaf-%d", i)))
		leaves[i] = hash[:]
	}

	// Build Merkle root
	merkleRoot := keeper.computeMerkleRoot(leaves)

	// Create proof for leaf 512 (middle of tree)
	transactionLeaf := leaves[512]
	merkleProof := keeper.buildMerkleProof(leaves, 512)
	merkleProofBytes := serializeMerkleProof(merkleProof)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.VerifyMerkleProofBytes(merkleRoot, transactionLeaf, merkleProofBytes)
	}
}

// ============================================================================
// Transfer ID Generation Benchmarks
// ============================================================================

// BenchmarkTransferIDGeneration benchmarks deterministic transfer ID generation
func BenchmarkTransferIDGeneration(b *testing.B) {
	keeper, ctx := setupKeeperForBenchmark(b)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.nextTransferID(ctx)
	}
}

// ============================================================================
// Validator Consensus Benchmarks
// ============================================================================

// BenchmarkValidatorConsensusCheck benchmarks validator signature validation
func BenchmarkValidatorConsensusCheck(b *testing.B) {
	keeper, ctx := setupKeeperForBenchmark(b)

	// Create test validators
	validators := []string{"val1", "val2", "val3", "val4", "val5"}
	signatures := make([][]byte, 5)
	for i := 0; i < 5; i++ {
		sig := make([]byte, 65)
		for j := range sig[:64] {
			sig[j] = byte((i + j) % 256)
		}
		sig[64] = 0
		signatures[i] = sig
	}

	transferID := "transfer-1"
	message := []byte("test transfer message")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.verifyValidatorSignatures(ctx, transferID, message, validators, signatures)
	}
}

// ============================================================================
// Circuit Breaker Benchmarks
// ============================================================================

// BenchmarkCircuitBreakerCheck benchmarks circuit breaker validation
func BenchmarkCircuitBreakerCheck(b *testing.B) {
	keeper, ctx := setupKeeperForBenchmark(b)

	chainID := "paw"
	amount := sdkmath.NewInt(1000000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.checkCircuitBreaker(ctx, chainID, amount)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// setupKeeperForBenchmark creates a keeper instance for benchmarking
func setupKeeperForBenchmark(b *testing.B) (*Keeper, sdk.Context) {
	b.Helper()

	keeper, ctx, _, _, _, _ := SetupKeeperForTest(b)
	return keeper, ctx
}

// computeMerkleRoot computes the Merkle root of a set of leaves
func (k Keeper) computeMerkleRoot(leaves [][]byte) []byte {
	if len(leaves) == 0 {
		return nil
	}
	if len(leaves) == 1 {
		return leaves[0]
	}

	// Build tree level by level
	level := leaves
	for len(level) > 1 {
		nextLevel := make([][]byte, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			if i+1 < len(level) {
				combined := append(level[i], level[i+1]...)
				hash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, hash[:])
			} else {
				nextLevel = append(nextLevel, level[i])
			}
		}
		level = nextLevel
	}

	return level[0]
}

// buildMerkleProof builds a Merkle proof for a leaf at the given index
func (k Keeper) buildMerkleProof(leaves [][]byte, index int) [][]byte {
	if index >= len(leaves) {
		return nil
	}

	proof := [][]byte{}
	level := leaves
	idx := index

	for len(level) > 1 {
		if idx%2 == 0 {
			// Left node, include right sibling
			if idx+1 < len(level) {
				proof = append(proof, level[idx+1])
			}
		} else {
			// Right node, include left sibling
			proof = append(proof, level[idx-1])
		}

		// Move to next level
		nextLevel := make([][]byte, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			if i+1 < len(level) {
				combined := append(level[i], level[i+1]...)
				hash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, hash[:])
			} else {
				nextLevel = append(nextLevel, level[i])
			}
		}
		level = nextLevel
		idx = idx / 2
	}

	return proof
}

// serializeMerkleProof serializes a Merkle proof to bytes
func serializeMerkleProof(proof [][]byte) []byte {
	var result []byte
	for _, node := range proof {
		result = append(result, node...)
	}
	return result
}

// Helper function to create test transfer
func createTestTransfer(id string, height int64) *types.CrossChainTransfer {
	return &types.CrossChainTransfer{
		TransferId:    id,
		SourceChain:   "aura",
		TargetChain:   "paw",
		Sender:        "aura1sender",
		Recipient:     "paw1recipient",
		Amount:        sdkmath.NewInt(1000000),
		Denom:         "uaura",
		Status:        types.TransferStatusPending,
		BlockHeight:   height,
		SourceBlockHeight: uint64(height),
	}
}

// Helper function to create test signature
func createTestSignature(seed int) []byte {
	sig := make([]byte, 65)
	for i := range sig[:64] {
		sig[i] = byte((seed + i) % 256)
	}
	sig[64] = 0
	return sig
}

// Helper function to create test chain
func createTestChain(chainID string) *types.SupportedChain {
	return &types.SupportedChain{
		ChainId:          chainID,
		ChainName:        chainID,
		IsEnabled:        true,
		MinTransferLimit: "1000",
		MaxTransferLimit: "1000000000",
	}
}

// BenchmarkChainValidation benchmarks chain validation
func BenchmarkChainValidation(b *testing.B) {
	keeper, ctx := setupKeeperForBenchmark(b)

	chain := createTestChain("paw")
	keeper.setChain(ctx, chain)

	amount := sdkmath.NewInt(1000000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.validateChain(ctx, "paw", amount)
	}
}

// BenchmarkTransferStatusUpdate benchmarks transfer status updates
func BenchmarkTransferStatusUpdate(b *testing.B) {
	keeper, ctx := setupKeeperForBenchmark(b)

	transfer := createTestTransfer("transfer-1", ctx.BlockHeight())
	keeper.setTransfer(ctx, transfer)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		transfer.Status = types.TransferStatusPending
		keeper.setTransfer(ctx, transfer)
		b.StartTimer()

		transfer.Status = types.TransferStatusCompleted
		_ = keeper.setTransfer(ctx, transfer)
	}
}
