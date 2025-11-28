package coverage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultCoverageConfig(t *testing.T) {
	config := DefaultCoverageConfig()
	require.NotNil(t, config)
	require.Equal(t, 90.0, config.MinimumCoverage)
	require.NotEmpty(t, config.ExcludePaths)
}

func TestCoverageReport(t *testing.T) {
	report := &CoverageReport{
		Modules: []ModuleCoverage{
			{
				ModuleName:      "testmodule",
				TotalStatements: 100,
				CoveredLines:    90,
				CoveragePercent: 90.0,
			},
		},
		TotalCoverage:   90.0,
		Threshold:       85.0,
		PassesThreshold: true,
	}

	require.True(t, report.PassesThreshold)
	require.Equal(t, 90.0, report.TotalCoverage)
}

func TestCountStatements(t *testing.T) {
	// Test with nil node
	count := countStatements(nil)
	require.Equal(t, 0, count)
}
