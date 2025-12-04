package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTxCmd(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "root command exists",
			run: func(t *testing.T) {
				cmd := GetTxCmd()
				require.NotNil(t, cmd)
				require.Equal(t, "bridge", cmd.Use)
				require.Contains(t, cmd.Aliases, "br")
				require.Contains(t, cmd.Aliases, "xchain")
				require.Equal(t, "Bridge transaction subcommands", cmd.Short)
				require.True(t, cmd.DisableFlagParsing)
				require.Equal(t, 2, cmd.SuggestionsMinimumDistance)
			},
		},
		{
			name: "has all subcommands",
			run: func(t *testing.T) {
				cmd := GetTxCmd()
				subcommands := cmd.Commands()
				require.Len(t, subcommands, 7)

				cmdNames := make(map[string]bool)
				for _, subcmd := range subcommands {
					cmdNames[subcmd.Use] = true
				}

				require.True(t, cmdNames["link-address [aura-address] [paw-address] [xai-address]"])
				require.True(t, cmdNames["lock-tokens [target-chain] [recipient] [amount]"])
				require.True(t, cmdNames["unlock-tokens [source-chain] [burn-tx-hash] [amount] [denom]"])
				require.True(t, cmdNames["mint-tokens [source-chain] [source-tx-hash] [recipient] [amount] [denom]"])
				require.True(t, cmdNames["burn-tokens [target-chain] [recipient] [amount]"])
				require.True(t, cmdNames["cross-chain-swap [source-chain] [input-amount] [target-chain] [target-denom] [min-target-amount]"])
				require.True(t, cmdNames["relay-transfer [transfer-id] [target-tx-hash] [status]"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCmdLinkAddress(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid all three addresses",
			args:    []string{"aura1abc", "paw1def", "xai1ghi"},
			wantErr: false,
		},
		{
			name:    "link only PAW",
			args:    []string{"aura1abc", "paw1def", ""},
			wantErr: false,
		},
		{
			name:    "link only XAI",
			args:    []string{"aura1abc", "", "xai1ghi"},
			wantErr: false,
		},
		{
			name:    "missing args",
			args:    []string{"aura1abc"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"aura1abc", "paw1def", "xai1ghi", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdLinkAddress()
			require.NotNil(t, cmd)
			require.Equal(t, "link-address [aura-address] [paw-address] [xai-address]", cmd.Use)
			require.Contains(t, cmd.Aliases, "link")
			require.Contains(t, cmd.Aliases, "link-addr")
			require.Contains(t, cmd.Short, "Link AURA, PAW, and XAI addresses")

			// Verify flags exist
			pawSigFlag := cmd.Flags().Lookup("paw-signature")
			require.NotNil(t, pawSigFlag)
			xaiSigFlag := cmd.Flags().Lookup("xai-signature")
			require.NotNil(t, xaiSigFlag)

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

func TestCmdLockTokens(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid to paw",
			args:    []string{"paw", "paw1recipient", "100uaura"},
			wantErr: false,
		},
		{
			name:    "valid to xai",
			args:    []string{"xai", "xai1recipient", "5000uaura"},
			wantErr: false,
		},
		{
			name:    "missing args",
			args:    []string{"paw"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"paw", "paw1recipient", "100uaura", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdLockTokens()
			require.NotNil(t, cmd)
			require.Equal(t, "lock-tokens [target-chain] [recipient] [amount]", cmd.Use)
			require.Contains(t, cmd.Short, "Lock tokens on AURA")

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

func TestCmdUnlockTokens(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid unlock from paw",
			args:    []string{"paw", "0xabc123", "1000", "uaura"},
			wantErr: false,
		},
		{
			name:    "valid unlock from xai",
			args:    []string{"xai", "0xdef456", "5000", "uaura"},
			wantErr: false,
		},
		{
			name:    "missing denom",
			args:    []string{"paw", "0xburn", "1000"},
			wantErr: true,
		},
		{
			name:    "missing amount and denom",
			args:    []string{"paw", "0xburn"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"paw", "0xburn", "1000", "uaura", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdUnlockTokens()
			require.NotNil(t, cmd)
			require.Equal(t, "unlock-tokens [source-chain] [burn-tx-hash] [amount] [denom]", cmd.Use)
			require.Contains(t, cmd.Short, "Unlock tokens on AURA")

			// Verify validator-signatures flag exists
			flag := cmd.Flags().Lookup("validator-signatures")
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

func TestCmdMintTokens(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid mint from paw",
			args:    []string{"paw", "0xabc123", "aura1def", "1000", "paw.token"},
			wantErr: false,
		},
		{
			name:    "valid mint from xai",
			args:    []string{"xai", "0xdef456", "aura1ghi", "5000", "xai.coin"},
			wantErr: false,
		},
		{
			name:    "missing denom",
			args:    []string{"paw", "0xabc123", "aura1def", "1000"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"paw", "0xabc", "aura1def", "1000", "paw.token", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdMintTokens()
			require.NotNil(t, cmd)
			require.Equal(t, "mint-tokens [source-chain] [source-tx-hash] [recipient] [amount] [denom]", cmd.Use)
			require.Contains(t, cmd.Short, "Mint wrapped tokens")
			require.Contains(t, cmd.Short, "validator-only")

			// Verify validator-signature flag exists
			flag := cmd.Flags().Lookup("validator-signature")
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

func TestCmdBurnTokens(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid burn to paw",
			args:    []string{"paw", "paw1def", "1000paw.token"},
			wantErr: false,
		},
		{
			name:    "valid burn to xai",
			args:    []string{"xai", "xai1ghi", "5000xai.coin"},
			wantErr: false,
		},
		{
			name:    "missing amount",
			args:    []string{"paw", "paw1def"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"paw", "paw1def", "1000paw.token", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdBurnTokens()
			require.NotNil(t, cmd)
			require.Equal(t, "burn-tokens [target-chain] [recipient] [amount]", cmd.Use)
			require.Contains(t, cmd.Short, "Burn wrapped tokens")

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

func TestCmdCrossChainSwap(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid swap aura to paw",
			args:    []string{"aura", "100uaura", "paw", "paw.token", "90"},
			wantErr: false,
		},
		{
			name:    "valid swap paw to xai",
			args:    []string{"paw", "1000paw.token", "xai", "xai.coin", "900"},
			wantErr: false,
		},
		{
			name:    "missing min target amount",
			args:    []string{"aura", "100uaura", "paw", "paw.token"},
			wantErr: true,
		},
		{
			name:    "incomplete args",
			args:    []string{"aura"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"aura", "100uaura", "paw", "paw.token", "90", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdCrossChainSwap()
			require.NotNil(t, cmd)
			require.Equal(t, "cross-chain-swap [source-chain] [input-amount] [target-chain] [target-denom] [min-target-amount]", cmd.Use)
			require.Contains(t, cmd.Short, "cross-chain swap")

			// Verify flags exist
			recipientFlag := cmd.Flags().Lookup("recipient")
			require.NotNil(t, recipientFlag)
			slippageFlag := cmd.Flags().Lookup("max-slippage")
			require.NotNil(t, slippageFlag)

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

func TestCmdRelayTransfer(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid relay completed",
			args:    []string{"transfer-1", "0xhash", "COMPLETED"},
			wantErr: false,
		},
		{
			name:    "valid relay relayed status",
			args:    []string{"transfer-456", "0xdef456", "RELAYED"},
			wantErr: false,
		},
		{
			name:    "missing status",
			args:    []string{"transfer-1", "0xhash"},
			wantErr: true,
		},
		{
			name:    "incomplete args",
			args:    []string{"transfer-1"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"transfer-1", "0xhash", "COMPLETED", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdRelayTransfer()
			require.NotNil(t, cmd)
			require.Equal(t, "relay-transfer [transfer-id] [target-tx-hash] [status]", cmd.Use)
			require.Contains(t, cmd.Short, "Relay a cross-chain transfer")
			require.Contains(t, cmd.Short, "relayer-only")

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
	require.Equal(t, "bridge", cmd.Use)
	require.Contains(t, cmd.Aliases, "br")
	require.Contains(t, cmd.Aliases, "xchain")
	require.NotEmpty(t, cmd.Short)
	require.True(t, cmd.DisableFlagParsing)
	require.Equal(t, 2, cmd.SuggestionsMinimumDistance)

	// Test all subcommands are registered
	subcommands := cmd.Commands()
	require.NotEmpty(t, subcommands)
	require.Len(t, subcommands, 7)

	// Verify each subcommand has proper structure
	for _, subcmd := range subcommands {
		require.NotEmpty(t, subcmd.Use, "subcommand Use should not be empty")
		require.NotEmpty(t, subcmd.Short, "subcommand Short should not be empty")
		require.NotNil(t, subcmd.RunE, "subcommand RunE should not be nil")
	}
}
