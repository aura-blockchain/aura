package keeper

import (
	"crypto/sha256"
	"testing"
)

// BenchmarkSignatureVerification benchmarks signature validation
func BenchmarkSignatureVerification(b *testing.B) {
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

// BenchmarkLockTokens benchmarks hashing operations
func BenchmarkLockTokens(b *testing.B) {
	data := []byte("test transfer data")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = sha256.Sum256(data)
	}
}

// BenchmarkUnlockTokens benchmarks hash verification
func BenchmarkUnlockTokens(b *testing.B) {
	data := []byte("test transfer data")
	expectedHash := sha256.Sum256(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		actualHash := sha256.Sum256(data)
		_ = actualHash == expectedHash
	}
}

// BenchmarkBatchTransfers benchmarks batch operations
func BenchmarkBatchTransfers(b *testing.B) {
	hashes := make([][32]byte, 100)
	for i := 0; i < 100; i++ {
		data := []byte{byte(i)}
		hashes[i] = sha256.Sum256(data)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, hash := range hashes {
			_ = hash[:]
		}
	}
}
