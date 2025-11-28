package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestGetTxCmd(t *testing.T) {
	cmd := GetTxCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "monitoring", cmd.Use)
	require.Contains(t, cmd.Short, "Monitoring transaction")
	require.True(t, cmd.DisableFlagParsing)

	// Check all subcommands exist
	subcommands := cmd.Commands()
	require.Len(t, subcommands, 2)
}

func TestCmdAcknowledgeAlert(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid alert ID",
			args:    []string{"alert-123"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"alert-123", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdAcknowledgeAlert()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "acknowledge-alert")
			require.Contains(t, cmd.Short, "Acknowledge")

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

func TestCmdResolveAlert(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid resolution",
			args:    []string{"alert-456", "Issue fixed by restarting validator"},
			wantErr: false,
		},
		{
			name:    "short resolution note",
			args:    []string{"alert-789", "Fixed"},
			wantErr: false,
		},
		{
			name:    "missing notes",
			args:    []string{"alert-123"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdResolveAlert()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "resolve-alert")
			require.Contains(t, cmd.Short, "Resolve")

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

func TestTxCommandStructure(t *testing.T) {
	cmd := GetTxCmd()

	// Test command properties
	require.Equal(t, "monitoring", cmd.Use)
	require.NotEmpty(t, cmd.Short)
	require.True(t, cmd.DisableFlagParsing)
	require.Equal(t, 2, cmd.SuggestionsMinimumDistance)

	// Verify each subcommand has proper structure
	for _, subcmd := range cmd.Commands() {
		require.NotEmpty(t, subcmd.Use)
		require.NotEmpty(t, subcmd.Short)
		require.NotNil(t, subcmd.RunE)
	}
}

func TestAllTxCommandsHaveHelp(t *testing.T) {
	commands := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"acknowledge-alert", CmdAcknowledgeAlert()},
		{"resolve-alert", CmdResolveAlert()},
	}

	for _, tc := range commands {
		t.Run(tc.name+" has help", func(t *testing.T) {
			require.NotEmpty(t, tc.cmd.Short)
			require.NotEmpty(t, tc.cmd.Long)
		})
	}
}
