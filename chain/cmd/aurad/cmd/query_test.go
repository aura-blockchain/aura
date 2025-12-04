package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryCmd(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "root query command exists",
			run: func(t *testing.T) {
				cmd := QueryCmd()
				require.NotNil(t, cmd)
				require.Equal(t, "query", cmd.Use)
				require.Contains(t, cmd.Aliases, "q")
				require.Equal(t, "Query blockchain state", cmd.Short)
				require.Contains(t, cmd.Long, "Query blockchain state, transactions, accounts")
			},
		},
		{
			name: "has all subcommands",
			run: func(t *testing.T) {
				cmd := QueryCmd()
				subcommands := cmd.Commands()
				require.NotEmpty(t, subcommands)
				// Verify at least the core subcommands exist
				require.GreaterOrEqual(t, len(subcommands), 4)

				// Check that commands have been added
				foundAccount := false
				for _, subcmd := range subcommands {
					if subcmd.Use == "account [address]" {
						foundAccount = true
						break
					}
				}
				require.True(t, foundAccount, "should have account query command")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestQueryAccountCmd(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid address",
			args:    []string{"aura1abc123"},
			wantErr: false,
		},
		{
			name:    "valid long address",
			args:    []string{"aura1qypqxpq9qcrsszgszyfpx9q4zct3sxfqzgq2qt"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"aura1abc", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := queryAccountCmd()
			require.NotNil(t, cmd)
			require.Equal(t, "account [address]", cmd.Use)
			require.Equal(t, "Query account information", cmd.Short)

			cmd.SetArgs(tt.args)
			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestQueryCommandStructure(t *testing.T) {
	cmd := QueryCmd()

	// Test command properties
	require.Equal(t, "query", cmd.Use)
	require.Contains(t, cmd.Aliases, "q")
	require.NotEmpty(t, cmd.Short)
	require.NotEmpty(t, cmd.Long)

	// Test all subcommands are registered
	subcommands := cmd.Commands()
	require.NotEmpty(t, subcommands)

	// Verify each subcommand has proper structure
	for _, subcmd := range subcommands {
		require.NotEmpty(t, subcmd.Use, "subcommand Use should not be empty")
		require.NotEmpty(t, subcmd.Short, "subcommand Short should not be empty")
		require.NotNil(t, subcmd.RunE, "subcommand RunE should not be nil")
	}
}
