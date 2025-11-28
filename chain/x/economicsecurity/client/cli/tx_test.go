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
				require.Equal(t, "economicsecurity", cmd.Use)
				require.Contains(t, cmd.Short, "Economic security transaction")
				require.True(t, cmd.DisableFlagParsing)
			},
		},
		{
			name: "has all transaction subcommands",
			run: func(t *testing.T) {
				cmd := GetTxCmd()
				subcommands := cmd.Commands()
				require.Len(t, subcommands, 8)

				cmdNames := make(map[string]bool)
				for _, subcmd := range subcommands {
					cmdNames[subcmd.Name()] = true
				}

				require.True(t, cmdNames["create-vesting"])
				require.True(t, cmdNames["release-vested"])
				require.True(t, cmdNames["revoke-vesting"])
				require.True(t, cmdNames["lock-voting"])
				require.True(t, cmdNames["unlock-voting"])
				require.True(t, cmdNames["propose-treasury-spend"])
				require.True(t, cmdNames["sign-treasury-spend"])
				require.True(t, cmdNames["execute-treasury-spend"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCmdCreateVestingSchedule(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name: "valid linear vesting for team",
			args: []string{
				"aura1beneficiary",
				"1000000uaura",
				"15552000", // 180 days cliff
				"31104000", // 360 days vesting
				"0",        // LINEAR
				"0",        // TEAM
			},
			wantErr: false,
		},
		{
			name: "valid cliff vesting for investor",
			args: []string{
				"aura1investor",
				"5000000uaura",
				"31536000", // 1 year cliff
				"63072000", // 2 years vesting
				"1",        // CLIFF
				"1",        // INVESTOR
			},
			wantErr: false,
		},
		{
			name: "valid exponential vesting for advisor",
			args: []string{
				"aura1advisor",
				"500000uaura",
				"7776000",  // 90 days cliff
				"31104000", // 360 days vesting
				"2",        // EXPONENTIAL
				"2",        // ADVISOR
			},
			wantErr: false,
		},
		{
			name:    "missing args",
			args:    []string{"aura1addr", "1000uaura"},
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
			cmd := CmdCreateVestingSchedule()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "create-vesting")
			require.Contains(t, cmd.Long, "LINEAR")
			require.Contains(t, cmd.Long, "TEAM")

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

func TestCmdReleaseVestedTokens(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid schedule ID",
			args:    []string{"schedule-123"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"schedule-123", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdReleaseVestedTokens()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "release-vested")

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

func TestCmdRevokeVestingSchedule(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid revocation with reason",
			args:    []string{"schedule-456", "Terminated employment"},
			wantErr: false,
		},
		{
			name:    "revoke for breach",
			args:    []string{"schedule-789", "Contract breach"},
			wantErr: false,
		},
		{
			name:    "missing reason",
			args:    []string{"schedule-123"},
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
			cmd := CmdRevokeVestingSchedule()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "revoke-vesting")

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

func TestCmdLockVotingTokens(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "lock for 1 year",
			args:    []string{"10000uaura", "31536000"},
			wantErr: false,
		},
		{
			name:    "lock for 6 months",
			args:    []string{"5000uaura", "15768000"},
			wantErr: false,
		},
		{
			name:    "lock for 2 years",
			args:    []string{"20000uaura", "63072000"},
			wantErr: false,
		},
		{
			name:    "missing duration",
			args:    []string{"1000uaura"},
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
			cmd := CmdLockVotingTokens()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "lock-voting")
			require.Contains(t, cmd.Long, "voting power")

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

func TestCmdUnlockVotingTokens(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid lock ID",
			args:    []string{"lock-123"},
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
			cmd := CmdUnlockVotingTokens()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "unlock-voting")

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

func TestCmdProposeTreasurySpend(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid proposal",
			args:    []string{"aura1recipient", "100000uaura", "Development grant Q1 2024"},
			wantErr: false,
		},
		{
			name:    "marketing proposal",
			args:    []string{"aura1marketing", "50000uaura", "Marketing campaign"},
			wantErr: false,
		},
		{
			name:    "missing description",
			args:    []string{"aura1addr", "1000uaura"},
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
			cmd := CmdProposeTreasurySpend()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "propose-treasury-spend")

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

func TestCmdSignTreasurySpend(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid transaction ID",
			args:    []string{"tx-123"},
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
			cmd := CmdSignTreasurySpend()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "sign-treasury-spend")

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

func TestCmdExecuteTreasurySpend(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid transaction ID",
			args:    []string{"tx-456"},
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
			cmd := CmdExecuteTreasurySpend()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "execute-treasury-spend")

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
	require.Equal(t, "economicsecurity", cmd.Use)
	require.NotEmpty(t, cmd.Short)
	require.True(t, cmd.DisableFlagParsing)
	require.Equal(t, 2, cmd.SuggestionsMinimumDistance)

	// Test all subcommands are registered
	subcommands := cmd.Commands()
	require.NotEmpty(t, subcommands)

	// Verify each subcommand has proper structure
	for _, subcmd := range subcommands {
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
		{"create-vesting", CmdCreateVestingSchedule()},
		{"release-vested", CmdReleaseVestedTokens()},
		{"revoke-vesting", CmdRevokeVestingSchedule()},
		{"lock-voting", CmdLockVotingTokens()},
		{"unlock-voting", CmdUnlockVotingTokens()},
		{"propose-treasury-spend", CmdProposeTreasurySpend()},
		{"sign-treasury-spend", CmdSignTreasurySpend()},
		{"execute-treasury-spend", CmdExecuteTreasurySpend()},
	}

	for _, tc := range commands {
		t.Run(tc.name+" has help text", func(t *testing.T) {
			require.NotEmpty(t, tc.cmd.Short, "command should have Short description")
			require.NotEmpty(t, tc.cmd.Long, "command should have Long description")
		})
	}
}

func TestVestingTypesDocumentation(t *testing.T) {
	cmd := CmdCreateVestingSchedule()
	require.Contains(t, cmd.Long, "LINEAR")
	require.Contains(t, cmd.Long, "CLIFF")
	require.Contains(t, cmd.Long, "EXPONENTIAL")
}

func TestScheduleTypesDocumentation(t *testing.T) {
	cmd := CmdCreateVestingSchedule()
	require.Contains(t, cmd.Long, "TEAM")
	require.Contains(t, cmd.Long, "INVESTOR")
	require.Contains(t, cmd.Long, "ADVISOR")
	require.Contains(t, cmd.Long, "ECOSYSTEM")
}
