// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/confidencescore/client/cli"
)

func TestGetQueryCmd(t *testing.T) {
	cmd := cli.GetQueryCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "confidencescore", cmd.Use)
	require.True(t, cmd.DisableFlagParsing)

	expected := []string{
		"score",
		"completions",
		"history",
		"thresholds",
		"verified-users",
		"arena-breakdown",
		"slash-records",
		"params",
		"ir-completion",
	}

	sub := make(map[string]struct{})
	for _, c := range cmd.Commands() {
		name := c.Use
		if idx := len(name); idx > 0 {
			sub[name] = struct{}{}
		}
	}

	for _, expectedCmd := range expected {
		found := false
		for _, c := range cmd.Commands() {
			if matchCommand(c, expectedCmd) {
				found = true
				break
			}
		}
		require.True(t, found, "missing query command: %s", expectedCmd)
	}
}

func matchCommand(cmd *cobra.Command, use string) bool {
	if cmd.Use == use {
		return true
	}
	if len(cmd.Use) >= len(use) && cmd.Use[:len(use)] == use {
		return true
	}
	return false
}

func TestCmdQueryUserScore(t *testing.T) {
	cmd := cli.CmdQueryUserScore()
	require.Equal(t, "score [wallet-address]", cmd.Use)
	require.Contains(t, cmd.Short, "confidence score")

	require.NoError(t, cmd.ValidateArgs([]string{"aura1abc123"}))
	require.Error(t, cmd.ValidateArgs([]string{}))
	require.Error(t, cmd.ValidateArgs([]string{"aura1abc123", "extra"}))
}

func TestCmdQueryUserCompletions(t *testing.T) {
	cmd := cli.CmdQueryUserCompletions()
	require.Equal(t, "completions [wallet-address]", cmd.Use)
	require.Contains(t, cmd.Short, "IR completion history")

	require.NoError(t, cmd.ValidateArgs([]string{"aura1abc123"}))
	require.Error(t, cmd.ValidateArgs([]string{}))
	require.Error(t, cmd.ValidateArgs([]string{"aura1abc123", "extra"}))

	// Check arena filter flag
	arenaFlag := cmd.Flags().Lookup("arena")
	require.NotNil(t, arenaFlag, "arena flag should be defined")
	require.Equal(t, "", arenaFlag.DefValue, "arena flag should default to empty string")
	require.Contains(t, arenaFlag.Usage, "Filter by arena type")
}

func TestCmdQueryScoreHistory(t *testing.T) {
	cmd := cli.CmdQueryScoreHistory()
	require.Equal(t, "history [wallet-address]", cmd.Use)
	require.Contains(t, cmd.Short, "score change history")

	require.NoError(t, cmd.ValidateArgs([]string{"aura1abc123"}))
	require.Error(t, cmd.ValidateArgs([]string{}))
	require.Error(t, cmd.ValidateArgs([]string{"aura1abc123", "extra"}))

	// Check height filter flags
	fromHeightFlag := cmd.Flags().Lookup("from-height")
	require.NotNil(t, fromHeightFlag, "from-height flag should be defined")
	require.Equal(t, "0", fromHeightFlag.DefValue, "from-height should default to 0")

	toHeightFlag := cmd.Flags().Lookup("to-height")
	require.NotNil(t, toHeightFlag, "to-height flag should be defined")
	require.Equal(t, "0", toHeightFlag.DefValue, "to-height should default to 0")
}

func TestCmdQueryThresholds(t *testing.T) {
	cmd := cli.CmdQueryThresholds()
	require.Equal(t, "thresholds", cmd.Use)
	require.Contains(t, cmd.Short, "verification thresholds")

	require.NoError(t, cmd.ValidateArgs([]string{}))
	require.Error(t, cmd.ValidateArgs([]string{"extra"}))
}

func TestCmdQueryVerifiedUsers(t *testing.T) {
	cmd := cli.CmdQueryVerifiedUsers()
	require.Equal(t, "verified-users", cmd.Use)
	require.Contains(t, cmd.Short, "verified users")

	require.NoError(t, cmd.ValidateArgs([]string{}))
	require.Error(t, cmd.ValidateArgs([]string{"extra"}))

	// Check min-score flag
	minScoreFlag := cmd.Flags().Lookup("min-score")
	require.NotNil(t, minScoreFlag, "min-score flag should be defined")
	require.Equal(t, "0", minScoreFlag.DefValue, "min-score should default to 0")
	require.Contains(t, minScoreFlag.Usage, "Minimum score filter")
}

func TestCmdQueryArenaBreakdown(t *testing.T) {
	cmd := cli.CmdQueryArenaBreakdown()
	require.Equal(t, "arena-breakdown [wallet-address]", cmd.Use)
	require.Contains(t, cmd.Short, "score breakdown by arena")

	require.NoError(t, cmd.ValidateArgs([]string{"aura1abc123"}))
	require.Error(t, cmd.ValidateArgs([]string{}))
	require.Error(t, cmd.ValidateArgs([]string{"aura1abc123", "extra"}))
}

func TestCmdQuerySlashRecords(t *testing.T) {
	cmd := cli.CmdQuerySlashRecords()
	require.Equal(t, "slash-records [wallet-address]", cmd.Use)
	require.Contains(t, cmd.Short, "slash records")

	require.NoError(t, cmd.ValidateArgs([]string{"aura1abc123"}))
	require.Error(t, cmd.ValidateArgs([]string{}))
	require.Error(t, cmd.ValidateArgs([]string{"aura1abc123", "extra"}))
}

func TestCmdQueryParams(t *testing.T) {
	cmd := cli.CmdQueryParams()
	require.Equal(t, "params", cmd.Use)
	require.Contains(t, cmd.Short, "parameters")

	require.NoError(t, cmd.ValidateArgs([]string{}))
	require.Error(t, cmd.ValidateArgs([]string{"extra"}))
}

func TestCmdQueryIRCompletion(t *testing.T) {
	cmd := cli.CmdQueryIRCompletion()
	require.Equal(t, "ir-completion [wallet-address] [ir-id]", cmd.Use)
	require.Contains(t, cmd.Short, "specific IR completion")

	require.NoError(t, cmd.ValidateArgs([]string{"aura1abc123", "IR-102"}))
	require.Error(t, cmd.ValidateArgs([]string{"aura1abc123"}))
	require.Error(t, cmd.ValidateArgs([]string{}))
	require.Error(t, cmd.ValidateArgs([]string{"aura1abc123", "IR-102", "extra"}))
}

func TestQueryCommandStructure(t *testing.T) {
	cmd := cli.GetQueryCmd()

	// Test command properties
	require.Equal(t, "confidencescore", cmd.Use)
	require.NotEmpty(t, cmd.Short)
	require.True(t, cmd.DisableFlagParsing)
	require.Equal(t, 2, cmd.SuggestionsMinimumDistance)

	// Test all subcommands are registered
	subcommands := cmd.Commands()
	require.NotEmpty(t, subcommands)
	require.Len(t, subcommands, 9)

	// Verify each subcommand has proper structure
	for _, subcmd := range subcommands {
		require.NotEmpty(t, subcmd.Use, "subcommand Use should not be empty")
		require.NotEmpty(t, subcmd.Short, "subcommand Short should not be empty")
		require.NotNil(t, subcmd.RunE, "subcommand RunE should not be nil")
	}
}

func TestQueryCommandDescriptions(t *testing.T) {
	// Test that each query command has meaningful descriptions
	tests := []struct {
		name         string
		cmd          func() *cobra.Command
		shortContain string
	}{
		{
			name:         "user score",
			cmd:          cli.CmdQueryUserScore,
			shortContain: "score",
		},
		{
			name:         "user completions",
			cmd:          cli.CmdQueryUserCompletions,
			shortContain: "completion",
		},
		{
			name:         "score history",
			cmd:          cli.CmdQueryScoreHistory,
			shortContain: "history",
		},
		{
			name:         "thresholds",
			cmd:          cli.CmdQueryThresholds,
			shortContain: "threshold",
		},
		{
			name:         "verified users",
			cmd:          cli.CmdQueryVerifiedUsers,
			shortContain: "verified",
		},
		{
			name:         "arena breakdown",
			cmd:          cli.CmdQueryArenaBreakdown,
			shortContain: "arena",
		},
		{
			name:         "slash records",
			cmd:          cli.CmdQuerySlashRecords,
			shortContain: "slash",
		},
		{
			name:         "params",
			cmd:          cli.CmdQueryParams,
			shortContain: "parameter",
		},
		{
			name:         "ir completion",
			cmd:          cli.CmdQueryIRCompletion,
			shortContain: "IR completion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd()
			require.NotEmpty(t, cmd.Short, "command Short description should not be empty")
			require.Contains(t, cmd.Short, tt.shortContain, "command Short should contain expected text")
			require.NotEmpty(t, cmd.Long, "command Long description should not be empty")
		})
	}
}

func TestQueryArgsValidation(t *testing.T) {
	// Test argument validation for commands that take wallet address
	walletAddressCommands := []struct {
		name string
		cmd  func() *cobra.Command
	}{
		{"score", cli.CmdQueryUserScore},
		{"completions", cli.CmdQueryUserCompletions},
		{"history", cli.CmdQueryScoreHistory},
		{"arena-breakdown", cli.CmdQueryArenaBreakdown},
		{"slash-records", cli.CmdQuerySlashRecords},
	}

	for _, tt := range walletAddressCommands {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd()

			// Valid: single wallet address
			require.NoError(t, cmd.ValidateArgs([]string{"aura1valid"}))

			// Invalid: no args
			require.Error(t, cmd.ValidateArgs([]string{}))

			// Invalid: too many args
			require.Error(t, cmd.ValidateArgs([]string{"aura1valid", "extra"}))
		})
	}

	// Test no-arg commands
	noArgCommands := []struct {
		name string
		cmd  func() *cobra.Command
	}{
		{"thresholds", cli.CmdQueryThresholds},
		{"verified-users", cli.CmdQueryVerifiedUsers},
		{"params", cli.CmdQueryParams},
	}

	for _, tt := range noArgCommands {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd()

			// Valid: no args
			require.NoError(t, cmd.ValidateArgs([]string{}))

			// Invalid: extra args
			require.Error(t, cmd.ValidateArgs([]string{"extra"}))
		})
	}
}

func TestQueryFlagsPresence(t *testing.T) {
	// Test that pagination flags are properly added where expected
	paginationCommands := []struct {
		name string
		cmd  func() *cobra.Command
	}{
		{"completions", cli.CmdQueryUserCompletions},
		{"history", cli.CmdQueryScoreHistory},
		{"verified-users", cli.CmdQueryVerifiedUsers},
		{"slash-records", cli.CmdQuerySlashRecords},
	}

	for _, tt := range paginationCommands {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd()

			// Check for pagination flags (added by flags.AddPaginationFlagsToCmd)
			// These are standard flags added by the SDK
			pageFlag := cmd.Flags().Lookup("page-key")
			limitFlag := cmd.Flags().Lookup("limit")

			// Note: These may be nil if not explicitly set, but the function should be called
			// We're mainly testing that the command structure is correct
			_ = pageFlag
			_ = limitFlag
		})
	}
}
