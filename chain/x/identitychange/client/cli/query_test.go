package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestGetQueryCmd(t *testing.T) {
	cmd := GetQueryCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "identitychange", cmd.Use)
	require.Contains(t, cmd.Short, "identitychange")

	subcommands := cmd.Commands()
	require.Len(t, subcommands, 4) // record, request, history, params
}

func TestCmdQueryIdentityRecord(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid DID",
			args:    []string{"did:aura:abc123"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryIdentityRecord()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "record")

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

func TestCmdQueryIdentityChangeRequest(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid request ID",
			args:    []string{"req-123"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryIdentityChangeRequest()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "request")

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

func TestCmdQueryIdentityChangeHistory(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid DID",
			args:    []string{"did:aura:def456"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryIdentityChangeHistory()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "history")

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

func TestAllQueryCommandsHaveHelp(t *testing.T) {
	commands := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"record", CmdQueryIdentityRecord()},
		{"request", CmdQueryIdentityChangeRequest()},
		{"history", CmdQueryIdentityChangeHistory()},
		{"params", CmdQueryParams()},
	}

	for _, tc := range commands {
		t.Run(tc.name+" has help", func(t *testing.T) {
			require.NotEmpty(t, tc.cmd.Short)
			require.NotEmpty(t, tc.cmd.Long)
		})
	}
}
