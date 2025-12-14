package cli

import (
	"testing"

	"github.com/spf13/cobra"
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
				require.Equal(t, "confidencescore", cmd.Use)
				require.Contains(t, cmd.Aliases, "cs")
				require.Contains(t, cmd.Aliases, "score")
				require.Equal(t, "Confidence score transaction subcommands", cmd.Short)
				require.True(t, cmd.DisableFlagParsing)
				require.Equal(t, 2, cmd.SuggestionsMinimumDistance)
			},
		},
		{
			name: "has all subcommands",
			run: func(t *testing.T) {
				cmd := GetTxCmd()
				subcommands := cmd.Commands()
				require.Len(t, subcommands, 5)

				cmdNames := make(map[string]bool)
				for _, subcmd := range subcommands {
					cmdNames[subcmd.Use] = true
				}

				// All 5 transaction commands
				require.True(t, cmdNames["record-completion [wallet-address] [ir-id] [proof-hash] [verifier-hash]"])
				require.True(t, cmdNames["recalculate-score [wallet-address]"])
				require.True(t, cmdNames["slash [wallet-address] [ir-id] [slash-amount] [reason]"])
				require.True(t, cmdNames["appeal [slash-tx-hash] [deposit]"])
				require.True(t, cmdNames["resolve-appeal [wallet-address] [slash-tx-hash] [restore]"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCmdRecordIRCompletion(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid ir completion",
			args:    []string{"aura1abc123", "IR-102", "a1b2c3d4e5f6", "9f8e7d6c5b4a"},
			wantErr: false,
		},
		{
			name:    "valid anchor completion",
			args:    []string{"aura1def456", "IR-000", "deadbeef1234", "cafebabe5678"},
			wantErr: false,
		},
		{
			name:    "missing verifier hash",
			args:    []string{"aura1abc123", "IR-102", "a1b2c3d4e5f6"},
			wantErr: true,
		},
		{
			name:    "missing proof and verifier",
			args:    []string{"aura1abc123", "IR-102"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"aura1abc123", "IR-102", "hash1", "hash2", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdRecordIRCompletion()
			require.NotNil(t, cmd)
			require.Equal(t, "record-completion [wallet-address] [ir-id] [proof-hash] [verifier-hash]", cmd.Use)
			require.Contains(t, cmd.Short, "Record an Inclusion Routine completion")

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

func TestCmdRecalculateScore(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid recalculation",
			args:    []string{"aura1abc123"},
			wantErr: false,
		},
		{
			name:    "valid different address",
			args:    []string{"aura1xyz789"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"aura1abc123", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdRecalculateScore()
			require.NotNil(t, cmd)
			require.Equal(t, "recalculate-score [wallet-address]", cmd.Use)
			require.Contains(t, cmd.Short, "Recalculate a user's confidence score")

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

func TestCmdSlashScore(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid slash",
			args:    []string{"aura1abc123", "IR-305", "5000", "fraud_detected"},
			wantErr: false,
		},
		{
			name:    "valid false attestation",
			args:    []string{"aura1def456", "IR-102", "3000", "false_attestation"},
			wantErr: false,
		},
		{
			name:    "valid collusion",
			args:    []string{"aura1xyz789", "IR-200", "10000", "collusion"},
			wantErr: false,
		},
		{
			name:    "valid duplicate completion",
			args:    []string{"aura1ghi012", "IR-150", "2500", "duplicate_completion"},
			wantErr: false,
		},
		{
			name:    "missing reason",
			args:    []string{"aura1abc123", "IR-305", "5000"},
			wantErr: true,
		},
		{
			name:    "missing slash amount and reason",
			args:    []string{"aura1abc123", "IR-305"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"aura1abc123", "IR-305", "5000", "fraud_detected", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdSlashScore()
			require.NotNil(t, cmd)
			require.Equal(t, "slash [wallet-address] [ir-id] [slash-amount] [reason]", cmd.Use)
			require.Contains(t, cmd.Short, "Slash a user's confidence score")

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

func TestCmdSlashScore_Flags(t *testing.T) {
	cmd := CmdSlashScore()
	require.NotNil(t, cmd)

	// Check evidence flag exists
	evidenceFlag := cmd.Flags().Lookup("evidence")
	require.NotNil(t, evidenceFlag, "evidence flag should be defined")
	require.Equal(t, "", evidenceFlag.DefValue, "evidence flag should default to empty string")
	require.Contains(t, evidenceFlag.Usage, "IPFS hash or URL to evidence")
}

func TestCmdAppealSlash(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid appeal",
			args:    []string{"A1B2C3D4E5F6", "1000aeq"},
			wantErr: false,
		},
		{
			name:    "valid appeal different deposit",
			args:    []string{"DEF456GHI789", "5000aeq"},
			wantErr: false,
		},
		{
			name:    "valid appeal with uaura",
			args:    []string{"XYZ123ABC456", "2000uaura"},
			wantErr: false,
		},
		{
			name:    "missing deposit",
			args:    []string{"A1B2C3D4E5F6"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"A1B2C3D4E5F6", "1000aeq", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdAppealSlash()
			require.NotNil(t, cmd)
			require.Equal(t, "appeal [slash-tx-hash] [deposit]", cmd.Use)
			require.Contains(t, cmd.Short, "Appeal a confidence score slash")

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

func TestCmdAppealSlash_Flags(t *testing.T) {
	cmd := CmdAppealSlash()
	require.NotNil(t, cmd)

	// Check evidence flag exists
	evidenceFlag := cmd.Flags().Lookup("evidence")
	require.NotNil(t, evidenceFlag, "evidence flag should be defined")
	require.Equal(t, "", evidenceFlag.DefValue, "evidence flag should default to empty string")
	require.Contains(t, evidenceFlag.Usage, "IPFS hash or URL to evidence")
}

func TestCmdResolveAppeal(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid resolve - restore true",
			args:    []string{"aura1abc123", "A1B2C3D4E5F6", "true"},
			wantErr: false,
		},
		{
			name:    "valid resolve - restore false",
			args:    []string{"aura1def456", "DEF456GHI789", "false"},
			wantErr: false,
		},
		{
			name:    "valid resolve - different case true",
			args:    []string{"aura1xyz789", "XYZ123ABC456", "TRUE"},
			wantErr: false,
		},
		{
			name:    "valid resolve - different case false",
			args:    []string{"aura1ghi012", "GHI789JKL012", "FALSE"},
			wantErr: false,
		},
		{
			name:    "missing restore",
			args:    []string{"aura1abc123", "A1B2C3D4E5F6"},
			wantErr: true,
		},
		{
			name:    "missing slash hash and restore",
			args:    []string{"aura1abc123"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"aura1abc123", "A1B2C3D4E5F6", "true", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdResolveAppeal()
			require.NotNil(t, cmd)
			require.Equal(t, "resolve-appeal [wallet-address] [slash-tx-hash] [restore]", cmd.Use)
			require.Contains(t, cmd.Short, "Resolve a slash appeal")

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

func TestCmdResolveAppeal_Flags(t *testing.T) {
	cmd := CmdResolveAppeal()
	require.NotNil(t, cmd)

	// Check notes flag exists
	notesFlag := cmd.Flags().Lookup("notes")
	require.NotNil(t, notesFlag, "notes flag should be defined")
	require.Equal(t, "", notesFlag.DefValue, "notes flag should default to empty string")
	require.Contains(t, notesFlag.Usage, "Resolution notes")
}

func TestTxCommandStructure(t *testing.T) {
	cmd := GetTxCmd()

	// Test command properties
	require.Equal(t, "confidencescore", cmd.Use)
	require.Contains(t, cmd.Aliases, "cs")
	require.Contains(t, cmd.Aliases, "score")
	require.NotEmpty(t, cmd.Short)
	require.True(t, cmd.DisableFlagParsing)
	require.Equal(t, 2, cmd.SuggestionsMinimumDistance)

	// Test all subcommands are registered
	subcommands := cmd.Commands()
	require.NotEmpty(t, subcommands)
	require.Len(t, subcommands, 5)

	// Verify each subcommand has proper structure
	for _, subcmd := range subcommands {
		require.NotEmpty(t, subcmd.Use, "subcommand Use should not be empty")
		require.NotEmpty(t, subcmd.Short, "subcommand Short should not be empty")
		require.NotNil(t, subcmd.RunE, "subcommand RunE should not be nil")
	}
}

func TestTxCommandDescriptions(t *testing.T) {
	// Test that each command has a meaningful short description
	tests := []struct {
		name        string
		cmd         func() *cobra.Command
		shortContain string
	}{
		{
			name:        "record completion",
			cmd:         CmdRecordIRCompletion,
			shortContain: "Record",
		},
		{
			name:        "recalculate score",
			cmd:         CmdRecalculateScore,
			shortContain: "Recalculate",
		},
		{
			name:        "slash score",
			cmd:         CmdSlashScore,
			shortContain: "Slash",
		},
		{
			name:        "appeal slash",
			cmd:         CmdAppealSlash,
			shortContain: "Appeal",
		},
		{
			name:        "resolve appeal",
			cmd:         CmdResolveAppeal,
			shortContain: "Resolve",
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
