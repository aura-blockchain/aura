// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetQueryCmd(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "root command exists",
			run: func(t *testing.T) {
				cmd := GetQueryCmd()
				require.NotNil(t, cmd)
				require.Equal(t, "bridge", cmd.Use)
				require.Equal(t, "Querying commands for the bridge module", cmd.Short)
				require.True(t, cmd.DisableFlagParsing)
				require.Equal(t, 2, cmd.SuggestionsMinimumDistance)
			},
		},
		{
			name: "has all subcommands",
			run: func(t *testing.T) {
				cmd := GetQueryCmd()
				subcommands := cmd.Commands()
				require.GreaterOrEqual(t, len(subcommands), 12)

				// Verify each subcommand exists
				cmdNames := make(map[string]bool)
				for _, subcmd := range subcommands {
					cmdNames[subcmd.Use] = true
				}

				require.True(t, cmdNames["transfer [transfer-id]"])
				require.True(t, cmdNames["transfers"])
				require.True(t, cmdNames["user-transfers [address]"])
				require.True(t, cmdNames["chain-config [chain-id]"])
				require.True(t, cmdNames["chains"])
				require.True(t, cmdNames["wrapped-token [wrapped-denom]"])
				require.True(t, cmdNames["wrapped-tokens"])
				require.True(t, cmdNames["shared-identity [address]"])
				require.True(t, cmdNames["cross-chain-swap [swap-id]"])
				require.True(t, cmdNames["stats"])
				require.True(t, cmdNames["validators"])
				require.True(t, cmdNames["relayer-stats [relayer-address]"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCmdQueryTransfer(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid transfer ID",
			args:    []string{"transfer-123"},
			wantErr: false,
		},
		{
			name:    "valid hex transfer ID",
			args:    []string{"0x1a2b3c4d5e6f"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"transfer-1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryTransfer()
			require.NotNil(t, cmd)
			require.Equal(t, "transfer [transfer-id]", cmd.Use)
			require.Contains(t, cmd.Short, "Query a cross-chain transfer")

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

func TestCmdQueryAllTransfers(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		flags   map[string]string
		wantErr bool
	}{
		{
			name:    "basic query",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "with status filter",
			args:    []string{},
			flags:   map[string]string{"status": "PENDING"},
			wantErr: false,
		},
		{
			name:    "with completed status",
			args:    []string{},
			flags:   map[string]string{"status": "COMPLETED"},
			wantErr: false,
		},
		{
			name:    "extra args - should fail",
			args:    []string{"extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryAllTransfers()
			require.NotNil(t, cmd)
			require.Equal(t, "transfers", cmd.Use)
			require.Contains(t, cmd.Short, "Query all cross-chain transfers")

			// Verify status flag exists
			flag := cmd.Flags().Lookup("status")
			require.NotNil(t, flag)

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

func TestCmdQueryUserTransfers(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid aura address",
			args:    []string{"aura1abc123"},
			wantErr: false,
		},
		{
			name:    "valid paw address",
			args:    []string{"paw1def456"},
			wantErr: false,
		},
		{
			name:    "valid xai address",
			args:    []string{"xai1ghi789"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"aura1abc", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryUserTransfers()
			require.NotNil(t, cmd)
			require.Equal(t, "user-transfers [address]", cmd.Use)
			require.Contains(t, cmd.Short, "Query transfers for a specific user")

			// Verify chain flag exists
			flag := cmd.Flags().Lookup("chain")
			require.NotNil(t, flag)

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

func TestCmdQueryChainConfig(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "query aura config",
			args:    []string{"aura"},
			wantErr: false,
		},
		{
			name:    "query paw config",
			args:    []string{"paw"},
			wantErr: false,
		},
		{
			name:    "query xai config",
			args:    []string{"xai"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"aura", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryChainConfig()
			require.NotNil(t, cmd)
			require.Equal(t, "chain-config [chain-id]", cmd.Use)
			require.Contains(t, cmd.Short, "Query configuration for a specific chain")

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

func TestCmdQueryAllChains(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "basic query",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "extra args - should fail",
			args:    []string{"extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryAllChains()
			require.NotNil(t, cmd)
			require.Equal(t, "chains", cmd.Use)
			require.Contains(t, cmd.Short, "Query all connected chains")

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

func TestCmdQueryWrappedToken(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid paw token",
			args:    []string{"paw.token"},
			wantErr: false,
		},
		{
			name:    "valid xai token",
			args:    []string{"xai.coin"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"paw.token", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryWrappedToken()
			require.NotNil(t, cmd)
			require.Equal(t, "wrapped-token [wrapped-denom]", cmd.Use)
			require.Contains(t, cmd.Short, "Query information about a wrapped token")

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

func TestCmdQueryAllWrappedTokens(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "basic query",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "extra args - should fail",
			args:    []string{"extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryAllWrappedTokens()
			require.NotNil(t, cmd)
			require.Equal(t, "wrapped-tokens", cmd.Use)
			require.Contains(t, cmd.Short, "Query all wrapped tokens")

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

func TestCmdQuerySharedIdentity(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid aura address",
			args:    []string{"aura1abc123"},
			wantErr: false,
		},
		{
			name:    "valid paw address",
			args:    []string{"paw1def456"},
			wantErr: false,
		},
		{
			name:    "valid xai address",
			args:    []string{"xai1ghi789"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"aura1abc", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQuerySharedIdentity()
			require.NotNil(t, cmd)
			require.Equal(t, "shared-identity [address]", cmd.Use)
			require.Contains(t, cmd.Short, "Query shared identity")

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

func TestCmdQueryCrossChainSwap(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid swap ID",
			args:    []string{"swap-123"},
			wantErr: false,
		},
		{
			name:    "valid hex swap ID",
			args:    []string{"0x1a2b3c4d"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"swap-1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryCrossChainSwap()
			require.NotNil(t, cmd)
			require.Equal(t, "cross-chain-swap [swap-id]", cmd.Use)
			require.Contains(t, cmd.Short, "Query cross-chain swap")

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

func TestCmdQueryBridgeStats(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "basic query",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "extra args - should fail",
			args:    []string{"extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryBridgeStats()
			require.NotNil(t, cmd)
			require.Equal(t, "stats", cmd.Use)
			require.Contains(t, cmd.Short, "Query bridge statistics")

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

func TestCmdQueryValidators(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "basic query",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "extra args - should fail",
			args:    []string{"extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryValidators()
			require.NotNil(t, cmd)
			require.Equal(t, "validators", cmd.Use)
			require.Contains(t, cmd.Short, "Query active bridge validators")

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

func TestCmdQueryRelayerStats(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid relayer address",
			args:    []string{"aura1relayer123"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"aura1relayer", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryRelayerStats()
			require.NotNil(t, cmd)
			require.Equal(t, "relayer-stats [relayer-address]", cmd.Use)
			require.Contains(t, cmd.Short, "Query relayer performance")

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

func TestFormatTransferStatus(t *testing.T) {
	tests := []struct {
		status   int32
		expected string
	}{
		{0, "PENDING"},
		{1, "CONFIRMED"},
		{2, "RELAYED"},
		{3, "COMPLETED"},
		{4, "FAILED"},
		{5, "REFUNDED"},
		{99, "UNKNOWN(99)"},
		{-1, "UNKNOWN(-1)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatTransferStatus(tt.status)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatChainList(t *testing.T) {
	tests := []struct {
		name     string
		chains   []string
		expected string
	}{
		{
			name:     "empty list",
			chains:   []string{},
			expected: "none",
		},
		{
			name:     "single chain",
			chains:   []string{"aura"},
			expected: "aura",
		},
		{
			name:     "multiple chains",
			chains:   []string{"aura", "paw", "xai"},
			expected: "aura, paw, xai",
		},
		{
			name:     "two chains",
			chains:   []string{"paw", "xai"},
			expected: "paw, xai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatChainList(tt.chains)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryCommandStructure(t *testing.T) {
	cmd := GetQueryCmd()

	// Test command properties
	require.Equal(t, "bridge", cmd.Use)
	require.NotEmpty(t, cmd.Short)
	require.True(t, cmd.DisableFlagParsing)
	require.Equal(t, 2, cmd.SuggestionsMinimumDistance)

	// Test all subcommands are registered
	subcommands := cmd.Commands()
	require.NotEmpty(t, subcommands)
	require.GreaterOrEqual(t, len(subcommands), 12)

	// Verify each subcommand has proper structure
	for _, subcmd := range subcommands {
		require.NotEmpty(t, subcmd.Use, "subcommand Use should not be empty")
		require.NotEmpty(t, subcmd.Short, "subcommand Short should not be empty")
		require.NotNil(t, subcmd.RunE, "subcommand RunE should not be nil")
	}
}
