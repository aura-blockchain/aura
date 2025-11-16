package stress

import (
	"testing"
)

func TestDefaultLoadTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	config := DefaultLoadTestConfig()
	config.TestDuration = 1 * 60 // 1 minute for testing
	config.NumConcurrentUsers = 10
	config.NumTransactionsPerUser = 10

	runner := NewLoadTestRunner(config)
	runner.Run(t)
}

func TestHighLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high load test in short mode")
	}

	RunHighLoadTest(t)
}

func TestSpikeLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping spike test in short mode")
	}

	RunSpikeTest(t)
}

func TestSoakTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping soak test in short mode")
	}

	RunSoakTest(t)
}
