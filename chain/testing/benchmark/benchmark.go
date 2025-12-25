// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package benchmark

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/testing/testutil"
)

// BenchmarkConfig defines benchmark configuration
type BenchmarkConfig struct {
	Iterations       int
	Concurrency      int
	WarmupIterations int
	Timeout          time.Duration
}

// DefaultBenchmarkConfig returns default benchmark configuration
func DefaultBenchmarkConfig() *BenchmarkConfig {
	return &BenchmarkConfig{
		Iterations:       10000,
		Concurrency:      100,
		WarmupIterations: 100,
		Timeout:          10 * time.Minute,
	}
}

// BenchmarkResult holds benchmark results
type BenchmarkResult struct {
	Name              string
	Iterations        int
	TotalDuration     time.Duration
	AverageDuration   time.Duration
	MinDuration       time.Duration
	MaxDuration       time.Duration
	OperationsPerSec  float64
	MemoryAllocations int64
	BytesAllocated    int64
}

// BenchmarkSuite provides benchmark utilities
type BenchmarkSuite struct {
	config  *BenchmarkConfig
	results []BenchmarkResult
	mu      sync.Mutex
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite(config *BenchmarkConfig) *BenchmarkSuite {
	if config == nil {
		config = DefaultBenchmarkConfig()
	}
	return &BenchmarkSuite{
		config:  config,
		results: make([]BenchmarkResult, 0),
	}
}

// Run executes a benchmark
func (bs *BenchmarkSuite) Run(b *testing.B, name string, fn func(b *testing.B)) {
	b.Run(name, fn)
}

// RecordResult records a benchmark result
func (bs *BenchmarkSuite) RecordResult(result BenchmarkResult) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.results = append(bs.results, result)
}

// GetResults returns all recorded results
func (bs *BenchmarkSuite) GetResults() []BenchmarkResult {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return append([]BenchmarkResult{}, bs.results...)
}

// PrintResults prints all benchmark results
func (bs *BenchmarkSuite) PrintResults() {
	results := bs.GetResults()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("  AURA BLOCKCHAIN BENCHMARK RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	for _, result := range results {
		fmt.Printf("\n%s\n", result.Name)
		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("Iterations:           %d\n", result.Iterations)
		fmt.Printf("Total Duration:       %v\n", result.TotalDuration)
		fmt.Printf("Average Duration:     %v\n", result.AverageDuration)
		fmt.Printf("Min Duration:         %v\n", result.MinDuration)
		fmt.Printf("Max Duration:         %v\n", result.MaxDuration)
		fmt.Printf("Operations/sec:       %.2f\n", result.OperationsPerSec)
		fmt.Printf("Memory Allocations:   %d\n", result.MemoryAllocations)
		fmt.Printf("Bytes Allocated:      %d\n", result.BytesAllocated)
	}
	fmt.Println(strings.Repeat("=", 80))
}

// BenchmarkTransactionProcessing benchmarks transaction processing
func BenchmarkTransactionProcessing(b *testing.B) {
	ctx := testutil.SetupTestContext(b)
	addr := testutil.GenerateTestAddress()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate transaction processing
		_ = processTransaction(ctx.SdkCtx, addr)
	}
}

// BenchmarkStateRead benchmarks state read operations
func BenchmarkStateRead(b *testing.B) {
	ctx := testutil.SetupTestContext(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate state read
		_ = ctx.SdkCtx.BlockHeight()
	}
}

// BenchmarkStateWrite benchmarks state write operations
func BenchmarkStateWrite(b *testing.B) {
	ctx := testutil.SetupTestContext(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate state write
		ctx.SdkCtx.WithBlockHeight(int64(i))
	}
}

// BenchmarkParallelTransactions benchmarks parallel transaction processing
func BenchmarkParallelTransactions(b *testing.B) {
	ctx := testutil.SetupTestContext(b)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		addr := testutil.GenerateTestAddress()
		for pb.Next() {
			_ = processTransaction(ctx.SdkCtx, addr)
		}
	})
}

// BenchmarkBlockProduction benchmarks block production
func BenchmarkBlockProduction(b *testing.B) {
	ctx := testutil.SetupTestContext(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate block production
		produceBlock(ctx.SdkCtx, i)
	}
}

// BenchmarkSignatureVerification benchmarks signature verification
func BenchmarkSignatureVerification(b *testing.B) {
	// Setup signature and message
	message := []byte("test message for signature verification")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate signature verification
		verifySignature(message)
	}
}

// BenchmarkHashComputation benchmarks hash computation
func BenchmarkHashComputation(b *testing.B) {
	data := make([]byte, 1024) // 1KB data

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate hash computation
		computeHash(data)
	}
}

// BenchmarkMerkleProofVerification benchmarks Merkle proof verification
func BenchmarkMerkleProofVerification(b *testing.B) {
	// Setup Merkle tree and proof
	leaves := make([][]byte, 100)
	for i := range leaves {
		leaves[i] = []byte(fmt.Sprintf("leaf-%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate Merkle proof verification
		verifyMerkleProof(leaves)
	}
}

// BenchmarkQueryExecution benchmarks query execution
func BenchmarkQueryExecution(b *testing.B) {
	ctx := testutil.SetupTestContext(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate query execution
		executeQuery(ctx.SdkCtx)
	}
}

// BenchmarkGasCalculation benchmarks gas calculation
func BenchmarkGasCalculation(b *testing.B) {
	ctx := testutil.SetupTestContext(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate gas calculation
		calculateGas(ctx.SdkCtx)
	}
}

// Module-specific benchmarks

// BenchmarkVCIssuance benchmarks VC issuance
func BenchmarkVCIssuance(b *testing.B) {
	ctx := testutil.SetupTestContext(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate VC issuance
		if err := issueVC(ctx.SdkCtx); err != nil {
			b.Fatalf("issueVC failed: %v", err)
		}
	}
}

// BenchmarkVCVerification benchmarks VC verification
func BenchmarkVCVerification(b *testing.B) {
	ctx := testutil.SetupTestContext(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate VC verification
		verifyVC(ctx.SdkCtx)
	}
}

// BenchmarkDataRegistryStore benchmarks data registry storage
func BenchmarkDataRegistryStore(b *testing.B) {
	ctx := testutil.SetupTestContext(b)
	data := make([]byte, 10240) // 10KB data

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate data storage
		if err := storeData(ctx.SdkCtx, data); err != nil {
			b.Fatalf("storeData failed: %v", err)
		}
	}
}

// BenchmarkDataRegistryRetrieve benchmarks data retrieval
func BenchmarkDataRegistryRetrieve(b *testing.B) {
	ctx := testutil.SetupTestContext(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate data retrieval
		retrieveData(ctx.SdkCtx)
	}
}

// BenchmarkDEXSwap benchmarks DEX swap operations
func BenchmarkDEXSwap(b *testing.B) {
	ctx := testutil.SetupTestContext(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate DEX swap
		if err := executeSwap(ctx.SdkCtx); err != nil {
			b.Fatalf("executeSwap failed: %v", err)
		}
	}
}

// BenchmarkDEXAddLiquidity benchmarks adding liquidity to DEX
func BenchmarkDEXAddLiquidity(b *testing.B) {
	ctx := testutil.SetupTestContext(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate adding liquidity
		if err := addLiquidity(ctx.SdkCtx); err != nil {
			b.Fatalf("addLiquidity failed: %v", err)
		}
	}
}

// BenchmarkBridgeTransfer benchmarks cross-chain bridge transfer
func BenchmarkBridgeTransfer(b *testing.B) {
	ctx := testutil.SetupTestContext(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate bridge transfer
		if err := executeBridgeTransfer(ctx.SdkCtx); err != nil {
			b.Fatalf("executeBridgeTransfer failed: %v", err)
		}
	}
}

// BenchmarkConfidenceScore benchmarks confidence score calculation
func BenchmarkConfidenceScore(b *testing.B) {
	ctx := testutil.SetupTestContext(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate confidence score calculation
		calculateConfidenceScore(ctx.SdkCtx)
	}
}

// BenchmarkInclusionRoutine benchmarks inclusion routine processing
func BenchmarkInclusionRoutine(b *testing.B) {
	ctx := testutil.SetupTestContext(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate inclusion routine processing
		if err := processInclusionRoutine(ctx.SdkCtx); err != nil {
			b.Fatalf("processInclusionRoutine failed: %v", err)
		}
	}
}

// Helper functions (stubs for actual implementations)

func processTransaction(ctx sdk.Context, addr sdk.AccAddress) error {
	// Simulate transaction processing
	time.Sleep(time.Microsecond * 10)
	return nil
}

func produceBlock(ctx sdk.Context, height int) {
	// Simulate block production
	time.Sleep(time.Microsecond * 50)
}

func verifySignature(message []byte) bool {
	// Simulate signature verification
	time.Sleep(time.Microsecond * 20)
	return true
}

func computeHash(data []byte) []byte {
	// Simulate hash computation
	time.Sleep(time.Microsecond * 5)
	return data[:32]
}

func verifyMerkleProof(leaves [][]byte) bool {
	// Simulate Merkle proof verification
	time.Sleep(time.Microsecond * 30)
	return true
}

func executeQuery(ctx sdk.Context) interface{} {
	// Simulate query execution
	time.Sleep(time.Microsecond * 15)
	return nil
}

func calculateGas(ctx sdk.Context) uint64 {
	// Simulate gas calculation
	time.Sleep(time.Microsecond * 5)
	return 100000
}

func issueVC(ctx sdk.Context) error {
	// Simulate VC issuance
	time.Sleep(time.Microsecond * 100)
	return nil
}

func verifyVC(ctx sdk.Context) bool {
	// Simulate VC verification
	time.Sleep(time.Microsecond * 50)
	return true
}

func storeData(ctx sdk.Context, data []byte) error {
	// Simulate data storage
	time.Sleep(time.Microsecond * 200)
	return nil
}

func retrieveData(ctx sdk.Context) []byte {
	// Simulate data retrieval
	time.Sleep(time.Microsecond * 100)
	return make([]byte, 10240)
}

func executeSwap(ctx sdk.Context) error {
	// Simulate DEX swap
	time.Sleep(time.Microsecond * 150)
	return nil
}

func addLiquidity(ctx sdk.Context) error {
	// Simulate adding liquidity
	time.Sleep(time.Microsecond * 120)
	return nil
}

func executeBridgeTransfer(ctx sdk.Context) error {
	// Simulate bridge transfer
	time.Sleep(time.Microsecond * 300)
	return nil
}

func calculateConfidenceScore(ctx sdk.Context) int {
	// Simulate confidence score calculation
	time.Sleep(time.Microsecond * 80)
	return 85
}

func processInclusionRoutine(ctx sdk.Context) error {
	// Simulate inclusion routine processing
	time.Sleep(time.Microsecond * 250)
	return nil
}

// Baseline metrics for comparison
var BaselineMetrics = map[string]BenchmarkResult{
	"TransactionProcessing": {
		Name:             "TransactionProcessing",
		OperationsPerSec: 10000,
		AverageDuration:  100 * time.Microsecond,
	},
	"BlockProduction": {
		Name:             "BlockProduction",
		OperationsPerSec: 200,
		AverageDuration:  5 * time.Millisecond,
	},
	"VCIssuance": {
		Name:             "VCIssuance",
		OperationsPerSec: 1000,
		AverageDuration:  1 * time.Millisecond,
	},
	"DEXSwap": {
		Name:             "DEXSwap",
		OperationsPerSec: 500,
		AverageDuration:  2 * time.Millisecond,
	},
}

// CompareWithBaseline compares current results with baseline
func CompareWithBaseline(current BenchmarkResult) (float64, string) {
	baseline, exists := BaselineMetrics[current.Name]
	if !exists {
		return 0, "No baseline available"
	}

	diff := (float64(current.AverageDuration) - float64(baseline.AverageDuration)) / float64(baseline.AverageDuration) * 100

	status := "✓ IMPROVED"
	if diff > 10 {
		status = "⚠ REGRESSED"
	} else if diff > 5 {
		status = "~ SIMILAR"
	}

	return diff, status
}
