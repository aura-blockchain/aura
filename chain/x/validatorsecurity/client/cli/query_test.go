package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryCLITestSuite(t *testing.T) {
	// Test command structure without requiring a live node
	TestGetQueryCmd(t)
	TestCmdQueryParamsStructure(t)
	TestCmdQueryValidatorSecurityInfoStructure(t)
	TestCmdQueryAllValidatorsStructure(t)
	TestCmdQueryJailedValidatorsStructure(t)
	TestCmdQueryTombstonedValidatorsStructure(t)
	TestCmdQueryDoubleSignEvidencesStructure(t)
	TestCmdQueryValidatorAlertsStructure(t)
	TestCmdQuerySentryNodesStructure(t)
}

// TestGetQueryCmd tests that GetQueryCmd returns a properly configured command
func TestGetQueryCmd(t *testing.T) {
	cmd := GetQueryCmd()

	require.NotNil(t, cmd)
	require.Equal(t, "validatorsecurity", cmd.Use)
	require.True(t, cmd.DisableFlagParsing)
	require.Greater(t, len(cmd.Commands()), 0)
}

// TestCmdQueryParamsStructure tests params query command structure
func TestCmdQueryParamsStructure(t *testing.T) {
	cmd := CmdQueryParams()

	require.NotNil(t, cmd)
	require.Equal(t, "params", cmd.Use)
	require.NotEmpty(t, cmd.Short)

	// Test with no arguments - will fail without context, which is expected
	cmd.SetArgs([]string{})
}

// TestCmdQueryValidatorSecurityInfoStructure tests validator security info query command structure
func TestCmdQueryValidatorSecurityInfoStructure(t *testing.T) {
	cmd := CmdQueryValidatorSecurityInfo()

	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "validator")
	require.NotEmpty(t, cmd.Short)

	// Test argument validation
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err, "Should require validator address argument")
}

// TestCmdQueryAllValidatorsStructure tests all validators query command structure
func TestCmdQueryAllValidatorsStructure(t *testing.T) {
	cmd := CmdQueryAllValidators()

	require.NotNil(t, cmd)
	require.Equal(t, "validators", cmd.Use)
	require.NotEmpty(t, cmd.Short)

	// Test with no arguments - will fail without context, which is expected
	cmd.SetArgs([]string{})
}

// TestCmdQueryJailedValidatorsStructure tests jailed validators query command structure
func TestCmdQueryJailedValidatorsStructure(t *testing.T) {
	cmd := CmdQueryJailedValidators()

	require.NotNil(t, cmd)
	require.Equal(t, "jailed", cmd.Use)
	require.NotEmpty(t, cmd.Short)

	// Test with no arguments - will fail without context, which is expected
	cmd.SetArgs([]string{})
}

// TestCmdQueryTombstonedValidatorsStructure tests tombstoned validators query command structure
func TestCmdQueryTombstonedValidatorsStructure(t *testing.T) {
	cmd := CmdQueryTombstonedValidators()

	require.NotNil(t, cmd)
	require.Equal(t, "tombstoned", cmd.Use)
	require.NotEmpty(t, cmd.Short)

	// Test with no arguments - will fail without context, which is expected
	cmd.SetArgs([]string{})
}

// TestCmdQueryDoubleSignEvidencesStructure tests double sign evidences query command structure
func TestCmdQueryDoubleSignEvidencesStructure(t *testing.T) {
	cmd := CmdQueryDoubleSignEvidences()

	require.NotNil(t, cmd)
	require.Equal(t, "evidences", cmd.Use)
	require.NotEmpty(t, cmd.Short)

	// Test with no arguments - will fail without context, which is expected
	cmd.SetArgs([]string{})
}

// TestCmdQueryValidatorAlertsStructure tests validator alerts query command structure
func TestCmdQueryValidatorAlertsStructure(t *testing.T) {
	cmd := CmdQueryValidatorAlerts()

	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "alerts")
	require.NotEmpty(t, cmd.Short)

	// Test argument validation
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err, "Should require validator address argument")
}

// TestCmdQuerySentryNodesStructure tests sentry nodes query command structure
func TestCmdQuerySentryNodesStructure(t *testing.T) {
	cmd := CmdQuerySentryNodes()

	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "sentry-nodes")
	require.NotEmpty(t, cmd.Short)

	// Test argument validation
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err, "Should require validator address argument")
}

// TestQueryCommandIntegration tests that all query commands are properly registered
func TestQueryCommandIntegration(t *testing.T) {
	queryCmd := GetQueryCmd()

	// Verify all subcommands are registered
	subCommands := queryCmd.Commands()
	require.GreaterOrEqual(t, len(subCommands), 8, "Should have at least 8 query subcommands")

	// Check that specific commands exist
	commandNames := make(map[string]bool)
	for _, cmd := range subCommands {
		commandNames[cmd.Name()] = true
	}

	require.True(t, commandNames["params"], "Should have params command")
	require.True(t, commandNames["validator"], "Should have validator command")
	require.True(t, commandNames["validators"], "Should have validators command")
	require.True(t, commandNames["jailed"], "Should have jailed command")
	require.True(t, commandNames["tombstoned"], "Should have tombstoned command")
	require.True(t, commandNames["evidences"], "Should have evidences command")
	require.True(t, commandNames["alerts"], "Should have alerts command")
	require.True(t, commandNames["sentry-nodes"], "Should have sentry-nodes command")
}
