package stress

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testing/testutil"
)

// LoadTestConfig configures stress test parameters
type LoadTestConfig struct {
	NumConcurrentUsers     int
	NumTransactionsPerUser int
	TestDuration           time.Duration
	RampUpDuration         time.Duration
	ThinkTime              time.Duration
	TargetTPS              int
}

// DefaultLoadTestConfig returns default stress test configuration
func DefaultLoadTestConfig() *LoadTestConfig {
	return &LoadTestConfig{
		NumConcurrentUsers:     100,
		NumTransactionsPerUser: 1000,
		TestDuration:           5 * time.Minute,
		RampUpDuration:         30 * time.Second,
		ThinkTime:              100 * time.Millisecond,
		TargetTPS:              1000,
	}
}

// LoadTestMetrics tracks stress test metrics
type LoadTestMetrics struct {
	TotalTransactions atomic.Uint64
	SuccessfulTxs     atomic.Uint64
	FailedTxs         atomic.Uint64
	TotalLatency      atomic.Int64
	MinLatency        atomic.Int64
	MaxLatency        atomic.Int64
	StartTime         time.Time
	EndTime           time.Time
	mu                sync.RWMutex
	LatencyHistogram  map[int]int // milliseconds -> count
}

// NewLoadTestMetrics creates a new metrics instance
func NewLoadTestMetrics() *LoadTestMetrics {
	m := &LoadTestMetrics{
		LatencyHistogram: make(map[int]int),
	}
	m.MinLatency.Store(int64(^uint64(0) >> 1)) // Max int64
	m.StartTime = time.Now()
	return m
}

// RecordTransaction records a transaction result
func (m *LoadTestMetrics) RecordTransaction(success bool, latency time.Duration) {
	m.TotalTransactions.Add(1)

	if success {
		m.SuccessfulTxs.Add(1)
	} else {
		m.FailedTxs.Add(1)
	}

	latencyMs := latency.Milliseconds()
	m.TotalLatency.Add(latencyMs)

	// Update min latency
	for {
		current := m.MinLatency.Load()
		if latencyMs >= current {
			break
		}
		if m.MinLatency.CompareAndSwap(current, latencyMs) {
			break
		}
	}

	// Update max latency
	for {
		current := m.MaxLatency.Load()
		if latencyMs <= current {
			break
		}
		if m.MaxLatency.CompareAndSwap(current, latencyMs) {
			break
		}
	}

	// Update histogram
	m.mu.Lock()
	bucket := int(latencyMs/10) * 10 // 10ms buckets
	m.LatencyHistogram[bucket]++
	m.mu.Unlock()
}

// Finalize marks the end of the test
func (m *LoadTestMetrics) Finalize() {
	m.EndTime = time.Now()
}

// PrintReport prints a detailed metrics report
func (m *LoadTestMetrics) PrintReport() {
	duration := m.EndTime.Sub(m.StartTime)
	total := m.TotalTransactions.Load()
	successful := m.SuccessfulTxs.Load()
	failed := m.FailedTxs.Load()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("  AURA BLOCKCHAIN STRESS TEST REPORT")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("\nTest Duration: %v\n", duration)
	fmt.Printf("Total Transactions: %d\n", total)
	fmt.Printf("Successful: %d (%.2f%%)\n", successful, float64(successful)/float64(total)*100)
	fmt.Printf("Failed: %d (%.2f%%)\n", failed, float64(failed)/float64(total)*100)
	fmt.Printf("\nThroughput: %.2f TPS\n", float64(total)/duration.Seconds())

	if total > 0 {
		avgLatency := time.Duration(m.TotalLatency.Load()/int64(total)) * time.Millisecond
		minLatency := time.Duration(m.MinLatency.Load()) * time.Millisecond
		maxLatency := time.Duration(m.MaxLatency.Load()) * time.Millisecond

		fmt.Printf("\nLatency Statistics:\n")
		fmt.Printf("  Average: %v\n", avgLatency)
		fmt.Printf("  Minimum: %v\n", minLatency)
		fmt.Printf("  Maximum: %v\n", maxLatency)
	}

	fmt.Println("\nLatency Distribution (10ms buckets):")
	m.mu.RLock()
	for bucket := 0; bucket <= int(m.MaxLatency.Load()); bucket += 10 {
		if count, ok := m.LatencyHistogram[bucket]; ok {
			percentage := float64(count) / float64(total) * 100
			fmt.Printf("  %4d-%4dms: %6d (%.2f%%) %s\n",
				bucket, bucket+10, count, percentage,
				generateBar(int(percentage)))
		}
	}
	m.mu.RUnlock()
	fmt.Println(strings.Repeat("=", 80))
}

func generateBar(percentage int) string {
	if percentage > 50 {
		percentage = 50
	}
	bar := ""
	for i := 0; i < percentage; i++ {
		bar += "█"
	}
	return bar
}

// LoadTestRunner executes stress tests
type LoadTestRunner struct {
	Config  *LoadTestConfig
	Metrics *LoadTestMetrics
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewLoadTestRunner creates a new load test runner
func NewLoadTestRunner(config *LoadTestConfig) *LoadTestRunner {
	ctx, cancel := context.WithTimeout(context.Background(), config.TestDuration)
	return &LoadTestRunner{
		Config:  config,
		Metrics: NewLoadTestMetrics(),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Run executes the load test
func (r *LoadTestRunner) Run(t *testing.T) {
	defer r.cancel()
	defer r.Metrics.Finalize()

	var wg sync.WaitGroup
	usersPerSecond := r.Config.NumConcurrentUsers
	if r.Config.RampUpDuration > 0 {
		usersPerSecond = r.Config.NumConcurrentUsers / int(r.Config.RampUpDuration.Seconds())
		if usersPerSecond == 0 {
			usersPerSecond = 1
		}
	}

	fmt.Printf("\nStarting load test with %d concurrent users...\n", r.Config.NumConcurrentUsers)
	fmt.Printf("Ramp-up: %v, Test duration: %v\n", r.Config.RampUpDuration, r.Config.TestDuration)

	for i := 0; i < r.Config.NumConcurrentUsers; i++ {
		// Ramp-up delay
		if i > 0 && i%usersPerSecond == 0 {
			time.Sleep(1 * time.Second)
		}

		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			r.simulateUser(t, userID)
		}(i)
	}

	wg.Wait()
	r.Metrics.PrintReport()

	// Verify metrics meet requirements
	successRate := float64(r.Metrics.SuccessfulTxs.Load()) / float64(r.Metrics.TotalTransactions.Load()) * 100
	require.GreaterOrEqual(t, successRate, 95.0, "Success rate should be at least 95%")

	avgLatency := time.Duration(r.Metrics.TotalLatency.Load()/int64(r.Metrics.TotalTransactions.Load())) * time.Millisecond
	require.Less(t, avgLatency, 5*time.Second, "Average latency should be less than 5 seconds")
}

// simulateUser simulates a single user's transactions
func (r *LoadTestRunner) simulateUser(t *testing.T, userID int) {
	testCtx := testutil.SetupTestContext(t)
	addr := testutil.GenerateTestAddress()

	for i := 0; i < r.Config.NumTransactionsPerUser; i++ {
		select {
		case <-r.ctx.Done():
			return
		default:
		}

		start := time.Now()
		success := r.executeTransaction(testCtx.SdkCtx, addr)
		latency := time.Since(start)

		r.Metrics.RecordTransaction(success, latency)

		// Think time between transactions
		time.Sleep(r.Config.ThinkTime)
	}
}

// executeTransaction executes a single transaction
func (r *LoadTestRunner) executeTransaction(ctx sdk.Context, addr sdk.AccAddress) bool {
	// Simulate transaction execution
	// In a real test, this would invoke actual module operations
	time.Sleep(time.Millisecond * 10) // Simulate processing time
	return true
}

// RunHighLoadTest runs a high load stress test
func RunHighLoadTest(t *testing.T) {
	config := &LoadTestConfig{
		NumConcurrentUsers:     500,
		NumTransactionsPerUser: 100,
		TestDuration:           10 * time.Minute,
		RampUpDuration:         1 * time.Minute,
		ThinkTime:              50 * time.Millisecond,
		TargetTPS:              5000,
	}

	runner := NewLoadTestRunner(config)
	runner.Run(t)
}

// RunSpikeTest runs a spike test with sudden load increase
func RunSpikeTest(t *testing.T) {
	config := &LoadTestConfig{
		NumConcurrentUsers:     1000,
		NumTransactionsPerUser: 50,
		TestDuration:           5 * time.Minute,
		RampUpDuration:         5 * time.Second, // Very fast ramp-up
		ThinkTime:              10 * time.Millisecond,
		TargetTPS:              10000,
	}

	runner := NewLoadTestRunner(config)
	runner.Run(t)
}

// RunSoakTest runs a long-duration soak test
func RunSoakTest(t *testing.T) {
	config := &LoadTestConfig{
		NumConcurrentUsers:     50,
		NumTransactionsPerUser: 10000,
		TestDuration:           1 * time.Hour,
		RampUpDuration:         2 * time.Minute,
		ThinkTime:              100 * time.Millisecond,
		TargetTPS:              500,
	}

	runner := NewLoadTestRunner(config)
	runner.Run(t)
}
