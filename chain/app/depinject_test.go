package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeeperDependencyGraph(t *testing.T) {
	graph := KeeperDependencyGraph()

	// Verify graph is not empty
	require.NotEmpty(t, graph)

	// Verify core dependencies
	require.Contains(t, graph, "bank")
	require.Contains(t, graph["bank"], "auth", "bank should depend on auth")

	require.Contains(t, graph, "staking")
	require.Contains(t, graph["staking"], "auth", "staking should depend on auth")
	require.Contains(t, graph["staking"], "bank", "staking should depend on bank")

	// Verify AURA module dependencies
	require.Contains(t, graph, "confidencescore")
	require.Contains(t, graph["confidencescore"], "inclusionroutines")

	require.Contains(t, graph, "vcregistry")
	require.Contains(t, graph["vcregistry"], "confidencescore")

	require.Contains(t, graph, "contractregistry")
	require.Contains(t, graph["contractregistry"], "vcregistry")
	require.Contains(t, graph["contractregistry"], "compliance")
	require.Contains(t, graph["contractregistry"], "confidencescore")
}

func TestTopologicalSort(t *testing.T) {
	tests := []struct {
		name         string
		dependencies map[string][]string
		expectError  bool
	}{
		{
			name: "simple linear dependencies",
			dependencies: map[string][]string{
				"c": {"b"},
				"b": {"a"},
				"a": {},
			},
			expectError: false,
		},
		{
			name: "circular dependency",
			dependencies: map[string][]string{
				"a": {"b"},
				"b": {"c"},
				"c": {"a"},
			},
			expectError: true,
		},
		{
			name: "complex valid graph",
			dependencies: map[string][]string{
				"bank":    {"auth"},
				"staking": {"auth", "bank"},
				"dex":     {"bank"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TopologicalSort(tt.dependencies)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				// Verify all modules are in the result
				require.Equal(t, len(tt.dependencies), len(result))

				// Verify dependencies are satisfied
				position := make(map[string]int)
				for i, module := range result {
					position[module] = i
				}

				for module, deps := range tt.dependencies {
					modulePos := position[module]
					for _, dep := range deps {
						depPos, exists := position[dep]
						require.True(t, exists, "dependency %s should be in result", dep)
						require.Less(t, depPos, modulePos, "%s should come before %s", dep, module)
					}
				}
			}
		})
	}
}

func TestValidateKeeperInitializationOrder(t *testing.T) {
	tests := []struct {
		name        string
		order       []string
		expectError bool
	}{
		{
			name: "correct order",
			order: []string{
				"auth", "bank", "staking", "slashing", "distribution",
				"compliance", "cryptography", "walletsecurity", "governance",
				"identitychange", "inclusionroutines",
				"confidencescore",
				"vcregistry", "dataregistry",
				"contractregistry", "bridge", "dex",
				"wasm", "wasmsecurity", "validatorsecurity",
			},
			expectError: false,
		},
		{
			name: "wrong order - bank before auth",
			order: []string{
				"bank", "auth", "staking",
			},
			expectError: true,
		},
		{
			name: "missing dependency",
			order: []string{
				"auth", "staking", // missing bank
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKeeperInitializationOrder(tt.order)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetKeeperInitializationOrder(t *testing.T) {
	order := GetKeeperInitializationOrder()

	// Verify order is not empty
	require.NotEmpty(t, order)

	// Verify it's a valid order
	err := ValidateKeeperInitializationOrder(order)
	require.NoError(t, err, "recommended initialization order should be valid")

	// Verify core modules are present
	require.Contains(t, order, "auth")
	require.Contains(t, order, "bank")
	require.Contains(t, order, "staking")

	// Verify AURA modules are present
	require.Contains(t, order, "vcregistry")
	require.Contains(t, order, "confidencescore")
	require.Contains(t, order, "contractregistry")
}
