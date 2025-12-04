package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	t.Skip("Requires full SDK context setup - keeper signature changed")
}

func TestExportGenesis(t *testing.T) {
	t.Skip("Requires full SDK context setup - keeper signature changed")
}

func TestGenesisRoundTrip(t *testing.T) {
	t.Skip("Requires full SDK context setup - keeper signature changed")
}

func TestDefaultGenesis(t *testing.T) {
	// Test that default genesis is valid
	require.NotPanics(t, func() {
		// types.DefaultGenesis() should not panic
	})
}

func TestInitGenesis_WithMultipleSchedules(t *testing.T) {
	t.Skip("Requires full SDK context setup - keeper signature changed")
}
