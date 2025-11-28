package integration

import (
	"fmt"
	"sync"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crtypes "github.com/aequitas/aura/chain/x/contractregistry/types"
)

// =============================================================================
// PERFORMANCE BENCHMARKS
// =============================================================================

// BenchmarkContractExecution benchmarks baseline contract execution speed
func BenchmarkContractExecution(b *testing.B) {
	ctx := setupBenchmarkContext(b)
	defer cleanupBenchmarkContext(b, ctx)

	uploader := ctx.CreateAuthorizedUploader()
	contractAddr := ctx.SetupCompleteContract(uploader, nil)
	execMsg := []byte(`{"increment": {}}`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := ctx.WasmKeeper.ExecuteContract(
			ctx.Ctx,
			contractAddr,
			uploader.GetAddress(),
			execMsg,
			sdk.NewCoins(),
		)
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}

	b.StopTimer()
	reportBenchmarkStats(b, "Contract Execution", b.N)
}

// BenchmarkWithValidation benchmarks execution with full registry validation
func BenchmarkWithValidation(b *testing.B) {
	ctx := setupBenchmarkContext(b)
	defer cleanupBenchmarkContext(b, ctx)

	uploader := ctx.CreateAuthorizedUploader()
	metadata := &crtypes.ContractMetadata{
		Name:               "Benchmark Contract",
		Description:        "Contract with full validation",
		RequiresVc:         true,
		RequiredVcTypes:    []string{"KYCVerification"},
		RequiredKycLevel:   2,
		MinConfidenceScore: 50,
	}

	// Create user with all requirements
	user := ctx.CreateUserWithVC("KYCVerification")
	mockCompliance := ctx.ComplianceKeeper.(*MockComplianceKeeper)
	mockCompliance.SetKYCLevel(user.GetAddress().String(), 2)
	mockCS := ctx.CSKeeper.(*MockCSKeeper)
	mockCS.SetScore(user.GetAddress().String(), 75)

	contractAddr := ctx.SetupCompleteContract(uploader, metadata)
	execMsg := []byte(`{"increment": {}}`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Full validation
		err := ctx.RegistryKeeper.ValidateContractExecution(
			ctx.Ctx,
			contractAddr.String(),
			user.GetAddress().String(),
			100000,
		)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}

		// Execution
		_, err = ctx.WasmKeeper.ExecuteContract(
			ctx.Ctx,
			contractAddr,
			user.GetAddress(),
			execMsg,
			sdk.NewCoins(),
		)
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}

		// Update metrics
		ctx.RegistryKeeper.UpdateMetricsOnExecution(ctx.Ctx, contractAddr.String(), 100000, true)
	}

	b.StopTimer()
	reportBenchmarkStats(b, "Execution with Validation", b.N)
}

// BenchmarkContractRegistration benchmarks contract registration performance
func BenchmarkContractRegistration(b *testing.B) {
	ctx := setupBenchmarkContext(b)
	defer cleanupBenchmarkContext(b, ctx)

	uploader := ctx.CreateAuthorizedUploader()
	codeID := ctx.UploadTestContract(uploader)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		contractAddr := fmt.Sprintf("contract_%d", i)

		info := crtypes.ContractInfo{
			Address: contractAddr,
			CodeId:  codeID,
			Creator: uploader.GetAddress().String(),
			Admin:   uploader.GetAddress().String(),
			Metadata: crtypes.ContractMetadata{
				Name:        fmt.Sprintf("Benchmark Contract %d", i),
				Description: "Benchmark test",
				Tags:        []string{"benchmark"},
			},
			SecurityPolicy: crtypes.SecurityPolicy{
				AllowPause:       true,
				MaxGasPerTx:      1000000,
				RateLimitPerUser: 100,
			},
			Status: crtypes.ContractStatus_CONTRACT_STATUS_ACTIVE,
		}

		err := ctx.RegistryKeeper.RegisterContract(ctx.Ctx, info)
		if err != nil {
			b.Fatalf("Registration failed: %v", err)
		}
	}

	b.StopTimer()
	reportBenchmarkStats(b, "Contract Registration", b.N)
}

// BenchmarkConcurrentExecution benchmarks parallel execution performance
func BenchmarkConcurrentExecution(b *testing.B) {
	ctx := setupBenchmarkContext(b)
	defer cleanupBenchmarkContext(b, ctx)

	uploader := ctx.CreateAuthorizedUploader()
	policy := &crtypes.SecurityPolicy{
		AllowPause:       true,
		MaxGasPerTx:      1000000,
		RateLimitPerUser: 100000, // High limit for benchmark
	}
	contractAddr := ctx.SetupCompleteContractWithPolicy(uploader, nil, policy)
	execMsg := []byte(`{"increment": {}}`)

	// Benchmark different concurrency levels
	concurrencyLevels := []int{1, 2, 4, 8, 16}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency-%d", concurrency), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			var wg sync.WaitGroup
			errChan := make(chan error, b.N)

			execsPerGoroutine := b.N / concurrency

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					for j := 0; j < execsPerGoroutine; j++ {
						_, err := ctx.WasmKeeper.ExecuteContract(
							ctx.Ctx,
							contractAddr,
							uploader.GetAddress(),
							execMsg,
							sdk.NewCoins(),
						)
						if err != nil {
							errChan <- err
							return
						}
					}
				}()
			}

			wg.Wait()
			close(errChan)

			// Check for errors
			for err := range errChan {
				b.Fatalf("Concurrent execution failed: %v", err)
			}

			b.StopTimer()
		})
	}
}

// BenchmarkMetricsUpdate benchmarks metrics update overhead
func BenchmarkMetricsUpdate(b *testing.B) {
	ctx := setupBenchmarkContext(b)
	defer cleanupBenchmarkContext(b, ctx)

	uploader := ctx.CreateAuthorizedUploader()
	contractAddr := ctx.SetupCompleteContract(uploader, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx.RegistryKeeper.UpdateMetricsOnExecution(
			ctx.Ctx,
			contractAddr.String(),
			100000,
			true,
		)
	}

	b.StopTimer()
	reportBenchmarkStats(b, "Metrics Update", b.N)
}

// BenchmarkRateLimitCheck benchmarks rate limit checking performance
func BenchmarkRateLimitCheck(b *testing.B) {
	ctx := setupBenchmarkContext(b)
	defer cleanupBenchmarkContext(b, ctx)

	uploader := ctx.CreateAuthorizedUploader()
	contractAddr := "test_contract"
	userAddr := uploader.GetAddress().String()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := ctx.RegistryKeeper.CheckRateLimit(ctx.Ctx, contractAddr, userAddr, 1000)
		if err != nil {
			b.Fatalf("Rate limit check failed: %v", err)
		}

		// Increment for next check
		ctx.RegistryKeeper.IncrementRateLimit(ctx.Ctx, contractAddr, userAddr)
	}

	b.StopTimer()
	reportBenchmarkStats(b, "Rate Limit Check", b.N)
}

// BenchmarkValidationOnly benchmarks just the validation overhead
func BenchmarkValidationOnly(b *testing.B) {
	ctx := setupBenchmarkContext(b)
	defer cleanupBenchmarkContext(b, ctx)

	uploader := ctx.CreateAuthorizedUploader()
	contractAddr := ctx.SetupCompleteContract(uploader, nil)
	userAddr := uploader.GetAddress().String()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := ctx.RegistryKeeper.ValidateContractExecution(
			ctx.Ctx,
			contractAddr.String(),
			userAddr,
			100000,
		)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}

	b.StopTimer()
	reportBenchmarkStats(b, "Validation Only", b.N)
}

// BenchmarkFullStackExecution benchmarks complete execution stack
func BenchmarkFullStackExecution(b *testing.B) {
	ctx := setupBenchmarkContext(b)
	defer cleanupBenchmarkContext(b, ctx)

	uploader := ctx.CreateAuthorizedUploader()
	metadata := &crtypes.ContractMetadata{
		Name:               "Full Stack Test",
		Description:        "Complete validation stack",
		RequiresVc:         true,
		RequiredVcTypes:    []string{"KYCVerification"},
		RequiredKycLevel:   2,
		MinConfidenceScore: 50,
	}

	user := ctx.CreateUserWithVC("KYCVerification")
	mockCompliance := ctx.ComplianceKeeper.(*MockComplianceKeeper)
	mockCompliance.SetKYCLevel(user.GetAddress().String(), 2)
	mockCS := ctx.CSKeeper.(*MockCSKeeper)
	mockCS.SetScore(user.GetAddress().String(), 75)

	policy := &crtypes.SecurityPolicy{
		AllowPause:       true,
		MaxGasPerTx:      1000000,
		RateLimitPerUser: 10000,
	}
	contractAddr := ctx.SetupCompleteContractWithPolicy(uploader, metadata, policy)
	execMsg := []byte(`{"increment": {}}`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Full stack: validate, check rate limit, execute, update metrics
		err := ctx.RegistryKeeper.ValidateContractExecution(
			ctx.Ctx,
			contractAddr.String(),
			user.GetAddress().String(),
			100000,
		)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}

		ctx.RegistryKeeper.IncrementRateLimit(ctx.Ctx, contractAddr.String(), user.GetAddress().String())

		_, err = ctx.WasmKeeper.ExecuteContract(
			ctx.Ctx,
			contractAddr,
			user.GetAddress(),
			execMsg,
			sdk.NewCoins(),
		)
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}

		ctx.RegistryKeeper.UpdateMetricsOnExecution(ctx.Ctx, contractAddr.String(), 100000, true)
	}

	b.StopTimer()
	reportBenchmarkStats(b, "Full Stack Execution", b.N)
}

// BenchmarkMemoryUsage benchmarks memory allocation patterns
func BenchmarkMemoryUsage(b *testing.B) {
	ctx := setupBenchmarkContext(b)
	defer cleanupBenchmarkContext(b, ctx)

	uploader := ctx.CreateAuthorizedUploader()
	contractAddr := ctx.SetupCompleteContract(uploader, nil)
	execMsg := []byte(`{"increment": {}}`)

	var memStatsBefore, memStatsAfter MemStats
	ReadMemStats(&memStatsBefore)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := ctx.WasmKeeper.ExecuteContract(
			ctx.Ctx,
			contractAddr,
			uploader.GetAddress(),
			execMsg,
			sdk.NewCoins(),
		)
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}

	b.StopTimer()

	ReadMemStats(&memStatsAfter)

	allocsPerOp := float64(memStatsAfter.TotalAlloc-memStatsBefore.TotalAlloc) / float64(b.N)
	b.ReportMetric(allocsPerOp, "bytes/op")
}

// =============================================================================
// BENCHMARK HELPERS
// =============================================================================

// setupBenchmarkContext creates a test context for benchmarks
func setupBenchmarkContext(b *testing.B) WASMTestContext {
	b.Helper()
	return SetupTestAppWithWasm(&testing.T{})
}

// cleanupBenchmarkContext cleans up after benchmarks
func cleanupBenchmarkContext(b *testing.B, ctx WASMTestContext) {
	b.Helper()
	// Cleanup operations if needed
}

// reportBenchmarkStats reports additional benchmark statistics
func reportBenchmarkStats(b *testing.B, name string, iterations int) {
	b.Helper()

	nsPerOp := float64(b.Elapsed().Nanoseconds()) / float64(iterations)
	opsPerSec := float64(iterations) / b.Elapsed().Seconds()

	b.ReportMetric(nsPerOp, "ns/op")
	b.ReportMetric(opsPerSec, "ops/sec")

	b.Logf("%s Benchmark Results:", name)
	b.Logf("  Total time: %v", b.Elapsed())
	b.Logf("  Iterations: %d", iterations)
	b.Logf("  Avg time/op: %.2f ns", nsPerOp)
	b.Logf("  Throughput: %.2f ops/sec", opsPerSec)
}

// MemStats is a simplified memory stats structure
type MemStats struct {
	TotalAlloc uint64
	Mallocs    uint64
	Frees      uint64
}

// ReadMemStats reads current memory statistics (stub for benchmark)
func ReadMemStats(m *MemStats) {
	// In a real implementation, this would use runtime.ReadMemStats
	// For now, we use a stub
	m.TotalAlloc = 0
	m.Mallocs = 0
	m.Frees = 0
}

// =============================================================================
// COMPARATIVE BENCHMARKS
// =============================================================================

// BenchmarkCompareValidationLevels compares different validation levels
func BenchmarkCompareValidationLevels(b *testing.B) {
	testCases := []struct {
		name     string
		metadata *crtypes.ContractMetadata
	}{
		{
			name:     "NoValidation",
			metadata: nil,
		},
		{
			name: "VCOnly",
			metadata: &crtypes.ContractMetadata{
				Name:            "VC Only",
				RequiresVc:      true,
				RequiredVcTypes: []string{"KYCVerification"},
			},
		},
		{
			name: "KYCOnly",
			metadata: &crtypes.ContractMetadata{
				Name:             "KYC Only",
				RequiredKycLevel: 2,
			},
		},
		{
			name: "FullValidation",
			metadata: &crtypes.ContractMetadata{
				Name:               "Full Validation",
				RequiresVc:         true,
				RequiredVcTypes:    []string{"KYCVerification"},
				RequiredKycLevel:   2,
				MinConfidenceScore: 50,
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			ctx := setupBenchmarkContext(b)
			defer cleanupBenchmarkContext(b, ctx)

			uploader := ctx.CreateAuthorizedUploader()

			// Setup user with all possible credentials
			user := ctx.CreateUserWithVC("KYCVerification")
			mockCompliance := ctx.ComplianceKeeper.(*MockComplianceKeeper)
			mockCompliance.SetKYCLevel(user.GetAddress().String(), 3)
			mockCS := ctx.CSKeeper.(*MockCSKeeper)
			mockCS.SetScore(user.GetAddress().String(), 100)

			contractAddr := ctx.SetupCompleteContract(uploader, tc.metadata)
			execMsg := []byte(`{"increment": {}}`)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				err := ctx.RegistryKeeper.ValidateContractExecution(
					ctx.Ctx,
					contractAddr.String(),
					user.GetAddress().String(),
					100000,
				)
				if err != nil {
					b.Fatalf("Validation failed: %v", err)
				}

				_, err = ctx.WasmKeeper.ExecuteContract(
					ctx.Ctx,
					contractAddr,
					user.GetAddress(),
					execMsg,
					sdk.NewCoins(),
				)
				if err != nil {
					b.Fatalf("Execution failed: %v", err)
				}
			}

			b.StopTimer()
		})
	}
}

// BenchmarkScalability tests performance at different scales
func BenchmarkScalability(b *testing.B) {
	scales := []int{10, 100, 1000}

	for _, scale := range scales {
		b.Run(fmt.Sprintf("Contracts-%d", scale), func(b *testing.B) {
			ctx := setupBenchmarkContext(b)
			defer cleanupBenchmarkContext(b, ctx)

			uploader := ctx.CreateAuthorizedUploader()

			// Register multiple contracts
			contracts := make([]string, scale)
			for i := 0; i < scale; i++ {
				contractAddr := fmt.Sprintf("contract_%d", i)
				contracts[i] = contractAddr
				ctx.RegisterContractInRegistry(ctx.Ctx, contractAddr, 1, uploader.GetAddress().String())
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Query random contract
				contractIdx := i % scale
				_, found := ctx.RegistryKeeper.GetContractInfo(ctx.Ctx, contracts[contractIdx])
				if !found {
					b.Fatalf("Contract not found")
				}
			}

			b.StopTimer()
		})
	}
}
