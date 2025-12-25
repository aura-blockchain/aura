// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTxCLITestSuite(t *testing.T) {
	// Test command structure without requiring a live node
	TestGetTxCmd(t)
	TestCmdRegisterValidatorStructure(t)
	TestCmdUpdateSecurityInfoStructure(t)
	TestCmdRegisterSentryNodeStructure(t)
	TestCmdReportDoubleSignStructure(t)
	TestCmdUnjailStructure(t)
	TestCmdAcknowledgeAlertStructure(t)
}

// TestGetTxCmd tests that GetTxCmd returns a properly configured command
func TestGetTxCmd(t *testing.T) {
	cmd := GetTxCmd()

	require.NotNil(t, cmd)
	require.Equal(t, "validatorsecurity", cmd.Use)
	require.True(t, cmd.DisableFlagParsing)
	require.Greater(t, len(cmd.Commands()), 0)
}

// TestCmdRegisterValidatorStructure tests validator registration command structure
func TestCmdRegisterValidatorStructure(t *testing.T) {
	cmd := CmdRegisterValidator()

	require.NotNil(t, cmd)
	require.Equal(t, "register-validator [hot-key] [cold-key] [region] [country-code]", cmd.Use)
	require.NotEmpty(t, cmd.Short)
	require.NotEmpty(t, cmd.Long)

	// Verify flags exist
	require.NotNil(t, cmd.Flags().Lookup("latitude"))
	require.NotNil(t, cmd.Flags().Lookup("longitude"))
	require.NotNil(t, cmd.Flags().Lookup("backup-validators"))

	// Test argument validation
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err, "Should require arguments")
}

// TestCmdUpdateSecurityInfoStructure tests security info update command structure
func TestCmdUpdateSecurityInfoStructure(t *testing.T) {
	cmd := CmdUpdateSecurityInfo()

	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "update-security-info")
	require.NotEmpty(t, cmd.Short)

	// Verify flags exist
	require.NotNil(t, cmd.Flags().Lookup("latitude"))
	require.NotNil(t, cmd.Flags().Lookup("longitude"))
	require.NotNil(t, cmd.Flags().Lookup("backup-validators"))

	// Test argument validation
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err, "Should require arguments")
}

// TestCmdRegisterSentryNodeStructure tests sentry node registration command structure
func TestCmdRegisterSentryNodeStructure(t *testing.T) {
	cmd := CmdRegisterSentryNode()

	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "register-sentry")
	require.NotEmpty(t, cmd.Short)

	// Test argument validation
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err, "Should require arguments")
}

// TestCmdReportDoubleSignStructure tests double sign reporting command structure
func TestCmdReportDoubleSignStructure(t *testing.T) {
	cmd := CmdReportDoubleSign()

	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "report-double-sign")
	require.NotEmpty(t, cmd.Short)

	// Test argument validation
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err, "Should require arguments")
}

// TestCmdUnjailStructure tests unjail command structure
func TestCmdUnjailStructure(t *testing.T) {
	cmd := CmdUnjail()

	require.NotNil(t, cmd)
	require.Equal(t, "unjail", cmd.Use)
	require.NotEmpty(t, cmd.Short)

	// Unjail should accept no arguments
	cmd.SetArgs([]string{})
	// Will fail without client context, but that's expected
}

// TestCmdAcknowledgeAlertStructure tests alert acknowledgement command structure
func TestCmdAcknowledgeAlertStructure(t *testing.T) {
	cmd := CmdAcknowledgeAlert()

	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "acknowledge-alert")
	require.NotEmpty(t, cmd.Short)

	// Test argument validation
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err, "Should require alert ID argument")
}

// TestCommandIntegration tests that all commands are properly registered
func TestCommandIntegration(t *testing.T) {
	txCmd := GetTxCmd()

	// Verify all subcommands are registered
	subCommands := txCmd.Commands()
	require.GreaterOrEqual(t, len(subCommands), 6, "Should have at least 6 subcommands")

	// Check that specific commands exist
	commandNames := make(map[string]bool)
	for _, cmd := range subCommands {
		commandNames[cmd.Name()] = true
	}

	require.True(t, commandNames["register-validator"], "Should have register-validator command")
	require.True(t, commandNames["update-security-info"], "Should have update-security-info command")
	require.True(t, commandNames["register-sentry"], "Should have register-sentry command")
	require.True(t, commandNames["report-double-sign"], "Should have report-double-sign command")
	require.True(t, commandNames["unjail"], "Should have unjail command")
	require.True(t, commandNames["acknowledge-alert"], "Should have acknowledge-alert command")
}
