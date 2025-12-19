package cli_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	aicli "github.com/aequitas/aura/chain/x/aiassistant/client/cli"
)

func TestNewQueryCmdStructure(t *testing.T) {
	cmd := aicli.NewQueryCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "aiassistant", cmd.Use)
	require.True(t, cmd.DisableFlagParsing)

	expected := map[string]bool{
		"assistant [address]":  false,
		"assistants":           false,
		"locale [locale-code]": false,
		"params":               false,
	}

	for _, sub := range cmd.Commands() {
		if _, ok := expected[sub.Use]; ok {
			expected[sub.Use] = true
		}
	}
	for use, seen := range expected {
		require.Truef(t, seen, "expected query subcommand not found: %s", use)
	}
}

func TestQueryAssistantCmdArgs(t *testing.T) {
	cmd := getSubCommand(t, aicli.NewQueryCmd(), "assistant [address]")
	require.Equal(t, "assistant [address]", cmd.Use)

	err := cmd.ValidateArgs([]string{"aura1assistant"})
	require.NoError(t, err)

	err = cmd.ValidateArgs([]string{})
	require.Error(t, err)
}

func TestQueryAssistantsCmdArgs(t *testing.T) {
	cmd := getSubCommand(t, aicli.NewQueryCmd(), "assistants")
	require.Equal(t, "assistants", cmd.Use)

	err := cmd.ValidateArgs([]string{})
	require.NoError(t, err)

	err = cmd.ValidateArgs([]string{"extra"})
	require.Error(t, err)
}

func TestQueryLocaleCmdArgs(t *testing.T) {
	cmd := getSubCommand(t, aicli.NewQueryCmd(), "locale [locale-code]")
	require.Equal(t, "locale [locale-code]", cmd.Use)

	err := cmd.ValidateArgs([]string{"en-us"})
	require.NoError(t, err)

	err = cmd.ValidateArgs([]string{})
	require.Error(t, err)
}

func TestQueryParamsCmdArgs(t *testing.T) {
	cmd := getSubCommand(t, aicli.NewQueryCmd(), "params")
	require.Equal(t, "params", cmd.Use)

	err := cmd.ValidateArgs([]string{})
	require.NoError(t, err)

	err = cmd.ValidateArgs([]string{"extra"})
	require.Error(t, err)
}

func getSubCommand(t *testing.T, root *cobra.Command, use string) *cobra.Command {
	t.Helper()
	for _, cmd := range root.Commands() {
		if cmd.Use == use {
			return cmd
		}
	}
	t.Fatalf("subcommand %s not found", use)
	return nil
}
