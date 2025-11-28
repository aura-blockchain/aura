package stress

import (
	"testing"
	"time"
)

func TestDefaultLoadTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	config := DefaultLoadTestConfig()
	config.TestDuration = time.Second
	config.RampUpDuration = time.Second
	config.NumConcurrentUsers = 2
	config.NumTransactionsPerUser = 1
	config.ThinkTime = 0

	runner := NewLoadTestRunner(config)
	runner.Run(t)
}

func TestHighLoad(t *testing.T) {
	t.Skip("High load test is disabled during unit test runs")
}

func TestSpikeLoad(t *testing.T) {
	t.Skip("Spike test is disabled during unit test runs")
}

func TestSoakTest(t *testing.T) {
	t.Skip("Soak test is disabled during unit test runs")
}
