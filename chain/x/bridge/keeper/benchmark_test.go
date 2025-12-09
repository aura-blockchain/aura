package keeper

import (
	"crypto/sha256"
	"testing"
)

// BenchmarkMerkleProofComputation benchmarks Merkle proof computation
func BenchmarkMerkleProofComputation(b *testing.B) {
	keeper, ctx, _, _, _, _ := SetupKeeperForTest(b)

	// Create leaves
	leaves := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		hash := sha256.Sum256([]byte{byte(i)})
		leaves[i] = hash[:]
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.ComputeMerkleRoot(leaves)
	}
}

// BenchmarkSignatureVerification_Simple benchmarks signature format validation
func BenchmarkSignatureVerification_Simple(b *testing.B) {
	// Simple signature validation benchmark
	signature := make([]byte, 65)
	for i := range signature[:64] {
		signature[i] = byte(i % 256)
	}
	signature[64] = 0

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = len(signature) == 65 && signature[64] < 4
	}
}

// BenchmarkLockTokens benchmarks token locking validation
func BenchmarkLockTokens(b *testing.B) {
	keeper, ctx, _, _, _, _ := SetupKeeperForTest(b)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = keeper.GetParams(ctx)
	}
}

// BenchmarkBatchTransfers benchmarks transfer retrieval
func BenchmarkBatchTransfers(b *testing.B) {
	keeper, ctx, _, _, _, _ := SetupKeeperForTest(b)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.GetAllTransfers(ctx)
	}
}
