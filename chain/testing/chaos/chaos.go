package chaos

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ChaosConfig defines chaos engineering configuration
type ChaosConfig struct {
	Enabled              bool
	FailureRate          float64 // 0.0 to 1.0
	MaxLatencyMs         int
	NetworkPartitionProb float64
	CrashProb            float64
	DataCorruptionProb   float64
	SlowdownProb         float64
	RandomSeed           int64
}

// DefaultChaosConfig returns default chaos configuration
func DefaultChaosConfig() *ChaosConfig {
	return &ChaosConfig{
		Enabled:              true,
		FailureRate:          0.05, // 5% failure rate
		MaxLatencyMs:         1000,
		NetworkPartitionProb: 0.02,
		CrashProb:            0.01,
		DataCorruptionProb:   0.01,
		SlowdownProb:         0.10,
		RandomSeed:           time.Now().UnixNano(),
	}
}

// ChaosEngine manages chaos injection
type ChaosEngine struct {
	config     *ChaosConfig
	random     *rand.Rand
	mu         sync.RWMutex
	injections []ChaosInjection
	active     bool
}

// ChaosInjection represents a single chaos injection
type ChaosInjection struct {
	Type      ChaosType
	Timestamp time.Time
	Target    string
	Details   string
}

// ChaosType defines types of chaos to inject
type ChaosType string

const (
	ChaosTypeNetworkPartition ChaosType = "network_partition"
	ChaosTypeCrash            ChaosType = "crash"
	ChaosTypeDataCorruption   ChaosType = "data_corruption"
	ChaosTypeSlowdown         ChaosType = "slowdown"
	ChaosTypeLatency          ChaosType = "latency"
	ChaosTypeRandomFailure    ChaosType = "random_failure"
)

// NewChaosEngine creates a new chaos engine
func NewChaosEngine(config *ChaosConfig) *ChaosEngine {
	return &ChaosEngine{
		config:     config,
		random:     rand.New(rand.NewSource(config.RandomSeed)),
		injections: make([]ChaosInjection, 0),
		active:     true,
	}
}

// Start starts the chaos engine
func (ce *ChaosEngine) Start(ctx context.Context) {
	if !ce.config.Enabled {
		return
	}

	go ce.backgroundChaosInjection(ctx)
}

// Stop stops the chaos engine
func (ce *ChaosEngine) Stop() {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.active = false
}

// backgroundChaosInjection runs continuous chaos injection
func (ce *ChaosEngine) backgroundChaosInjection(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ce.injectRandomChaos()
		}
	}
}

// injectRandomChaos randomly injects chaos
func (ce *ChaosEngine) injectRandomChaos() {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	if !ce.active {
		return
	}

	roll := ce.random.Float64()

	if roll < ce.config.NetworkPartitionProb {
		ce.recordInjection(ChaosTypeNetworkPartition, "random_node", "Simulated network partition")
	} else if roll < ce.config.NetworkPartitionProb+ce.config.CrashProb {
		ce.recordInjection(ChaosTypeCrash, "random_node", "Simulated node crash")
	} else if roll < ce.config.NetworkPartitionProb+ce.config.CrashProb+ce.config.DataCorruptionProb {
		ce.recordInjection(ChaosTypeDataCorruption, "random_data", "Simulated data corruption")
	} else if roll < ce.config.NetworkPartitionProb+ce.config.CrashProb+ce.config.DataCorruptionProb+ce.config.SlowdownProb {
		ce.recordInjection(ChaosTypeSlowdown, "random_operation", "Simulated slowdown")
	}
}

// recordInjection records a chaos injection
func (ce *ChaosEngine) recordInjection(chaosType ChaosType, target, details string) {
	injection := ChaosInjection{
		Type:      chaosType,
		Timestamp: time.Now(),
		Target:    target,
		Details:   details,
	}
	ce.injections = append(ce.injections, injection)
}

// ShouldFail determines if an operation should fail
func (ce *ChaosEngine) ShouldFail() bool {
	if !ce.config.Enabled {
		return false
	}

	ce.mu.RLock()
	defer ce.mu.RUnlock()

	if !ce.active {
		return false
	}

	return ce.random.Float64() < ce.config.FailureRate
}

// InjectLatency injects random latency
func (ce *ChaosEngine) InjectLatency() time.Duration {
	if !ce.config.Enabled {
		return 0
	}

	ce.mu.RLock()
	defer ce.mu.RUnlock()

	if !ce.active {
		return 0
	}

	latencyMs := ce.random.Intn(ce.config.MaxLatencyMs)
	if latencyMs > 0 {
		ce.recordInjection(ChaosTypeLatency, "latency_injection", fmt.Sprintf("%dms delay", latencyMs))
	}
	return time.Duration(latencyMs) * time.Millisecond
}

// InjectNetworkPartition simulates a network partition
func (ce *ChaosEngine) InjectNetworkPartition(duration time.Duration) error {
	if !ce.config.Enabled {
		return nil
	}

	ce.mu.Lock()
	ce.recordInjection(ChaosTypeNetworkPartition, "network", fmt.Sprintf("Partition for %v", duration))
	ce.mu.Unlock()

	// Simulate partition by blocking operations
	time.Sleep(duration)
	return nil
}

// InjectCrash simulates a node crash
func (ce *ChaosEngine) InjectCrash() error {
	if !ce.config.Enabled {
		return nil
	}

	ce.mu.Lock()
	defer ce.mu.Unlock()

	ce.recordInjection(ChaosTypeCrash, "node", "Simulated crash")
	return fmt.Errorf("chaos injection: simulated crash")
}

// InjectDataCorruption simulates data corruption
func (ce *ChaosEngine) InjectDataCorruption(data []byte) []byte {
	if !ce.config.Enabled || len(data) == 0 {
		return data
	}

	ce.mu.Lock()
	defer ce.mu.Unlock()

	if ce.random.Float64() < ce.config.DataCorruptionProb {
		// Corrupt random byte
		pos := ce.random.Intn(len(data))
		corrupted := make([]byte, len(data))
		copy(corrupted, data)
		corrupted[pos] ^= 0xFF // Flip all bits

		ce.recordInjection(ChaosTypeDataCorruption, "data", fmt.Sprintf("Corrupted byte at position %d", pos))
		return corrupted
	}

	return data
}

// InjectSlowdown simulates operation slowdown
func (ce *ChaosEngine) InjectSlowdown() {
	if !ce.config.Enabled {
		return
	}

	ce.mu.RLock()
	defer ce.mu.RUnlock()

	if !ce.active {
		return
	}

	if ce.random.Float64() < ce.config.SlowdownProb {
		delay := time.Duration(ce.random.Intn(ce.config.MaxLatencyMs)) * time.Millisecond
		time.Sleep(delay)
	}
}

// GetInjections returns all recorded chaos injections
func (ce *ChaosEngine) GetInjections() []ChaosInjection {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	result := make([]ChaosInjection, len(ce.injections))
	copy(result, ce.injections)
	return result
}

// PrintReport prints a chaos injection report
func (ce *ChaosEngine) PrintReport() {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("  CHAOS ENGINEERING REPORT")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("\nTotal Injections: %d\n", len(ce.injections))

	typeCount := make(map[ChaosType]int)
	for _, injection := range ce.injections {
		typeCount[injection.Type]++
	}

	fmt.Println("\nInjection Breakdown:")
	for chaosType, count := range typeCount {
		fmt.Printf("  %-25s: %d\n", chaosType, count)
	}

	fmt.Println("\nRecent Injections (last 10):")
	start := len(ce.injections) - 10
	if start < 0 {
		start = 0
	}
	for i := start; i < len(ce.injections); i++ {
		inj := ce.injections[i]
		fmt.Printf("  [%s] %s on %s: %s\n",
			inj.Timestamp.Format("15:04:05"),
			inj.Type,
			inj.Target,
			inj.Details)
	}
	fmt.Println(strings.Repeat("=", 80))
}

// ChaosWrapper wraps a function with chaos injection
func (ce *ChaosEngine) ChaosWrapper(ctx sdk.Context, fn func() error) error {
	// Inject latency
	if latency := ce.InjectLatency(); latency > 0 {
		time.Sleep(latency)
	}

	// Inject slowdown
	ce.InjectSlowdown()

	// Check if should fail
	if ce.ShouldFail() {
		ce.mu.Lock()
		ce.recordInjection(ChaosTypeRandomFailure, "operation", "Random failure injection")
		ce.mu.Unlock()
		return fmt.Errorf("chaos injection: random failure")
	}

	// Execute function
	return fn()
}

// ChaosMiddleware provides middleware for chaos injection
func ChaosMiddleware(ce *ChaosEngine) func(sdk.Context, func() error) error {
	return func(ctx sdk.Context, next func() error) error {
		return ce.ChaosWrapper(ctx, next)
	}
}
