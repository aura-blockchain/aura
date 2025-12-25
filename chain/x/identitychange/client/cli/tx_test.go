// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestGetTxCmd(t *testing.T) {
	cmd := GetTxCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "identitychange", cmd.Use)
	require.Contains(t, cmd.Short, "Identity change")

	subcommands := cmd.Commands()
	require.Len(t, subcommands, 5)
}

func TestCmdRequestIdentityChange(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid request",
			args:    []string{"did:aura:abc123", "metadata-hash", "ir-789", "proof-xyz"},
			wantErr: false,
		},
		{
			name:    "missing args",
			args:    []string{"did:aura:abc123", "hash"},
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
			cmd := CmdRequestIdentityChange()
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

func TestCmdSubmitAssistantProof(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid proof - success",
			args:    []string{"req-123", "proof-xyz", "10", "true"},
			wantErr: false,
		},
		{
			name:    "valid proof - failure",
			args:    []string{"req-456", "proof-abc", "-5", "false"},
			wantErr: false,
		},
		{
			name:    "missing args",
			args:    []string{"req-123", "proof"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdSubmitAssistantProof()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "submit-proof")

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

func TestCmdApplyIdentityChange(t *testing.T) {
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
			cmd := CmdApplyIdentityChange()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "apply")

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

func TestCmdRejectIdentityChange(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid rejection",
			args:    []string{"req-123", "Insufficient proof"},
			wantErr: false,
		},
		{
			name:    "missing reason",
			args:    []string{"req-123"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdRejectIdentityChange()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "reject")

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

func TestCmdSuspendIdentityChanges(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid suspension",
			args:    []string{"Security issue detected"},
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
			cmd := CmdSuspendIdentityChanges()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "suspend")

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

func TestAllTxCommandsHaveHelp(t *testing.T) {
	commands := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"request", CmdRequestIdentityChange()},
		{"submit-proof", CmdSubmitAssistantProof()},
		{"apply", CmdApplyIdentityChange()},
		{"reject", CmdRejectIdentityChange()},
		{"suspend", CmdSuspendIdentityChanges()},
	}

	for _, tc := range commands {
		t.Run(tc.name+" has help", func(t *testing.T) {
			require.NotEmpty(t, tc.cmd.Short)
			require.NotEmpty(t, tc.cmd.Long)
		})
	}
}
