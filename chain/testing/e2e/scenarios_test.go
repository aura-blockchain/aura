package e2e

import (
	"testing"
	"time"
)

func TestE2EScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	// Run all scenarios with 30-minute timeout
	RunScenariosWithTimeout(t, 30*time.Minute)
}

func TestIndividualScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	// Run individual scenarios for debugging
	for _, scenario := range ScenarioRegistry {
		t.Run(scenario.Name, func(t *testing.T) {
			RunScenario(t, scenario)
		})
	}
}
