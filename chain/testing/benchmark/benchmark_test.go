package benchmark

import (
	"testing"
)

// Run all benchmarks
func TestBenchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmarks in short mode")
	}

	// This is just a marker test
	// Actual benchmarks are run with: go test -bench=.
}

// Baseline benchmark tests
func TestBaselineBenchmarks(t *testing.T) {
	benchmarks := []struct {
		name string
		fn   func(b *testing.B)
	}{
		{"TransactionProcessing", BenchmarkTransactionProcessing},
		{"StateRead", BenchmarkStateRead},
		{"StateWrite", BenchmarkStateWrite},
		{"BlockProduction", BenchmarkBlockProduction},
		{"SignatureVerification", BenchmarkSignatureVerification},
		{"HashComputation", BenchmarkHashComputation},
		{"VCIssuance", BenchmarkVCIssuance},
		{"VCVerification", BenchmarkVCVerification},
		{"DEXSwap", BenchmarkDEXSwap},
	}

	for _, bench := range benchmarks {
		t.Run(bench.name, func(t *testing.T) {
			result := testing.Benchmark(bench.fn)
			t.Logf("%s: %d ns/op, %d B/op, %d allocs/op",
				bench.name,
				result.NsPerOp(),
				result.AllocedBytesPerOp(),
				result.AllocsPerOp())
		})
	}
}
