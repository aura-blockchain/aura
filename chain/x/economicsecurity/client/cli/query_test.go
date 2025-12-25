// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/spf13/cobra"
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
				require.Equal(t, "economicsecurity", cmd.Use)
				require.Contains(t, cmd.Short, "economic security")
				require.True(t, cmd.DisableFlagParsing)
			},
		},
		{
			name: "has all query subcommands",
			run: func(t *testing.T) {
				cmd := GetQueryCmd()
				subcommands := cmd.Commands()
				require.Len(t, subcommands, 14)

				cmdNames := make(map[string]bool)
				for _, subcmd := range subcommands {
					cmdNames[subcmd.Name()] = true
				}

				require.True(t, cmdNames["params"])
				require.True(t, cmdNames["vesting-schedule"])
				require.True(t, cmdNames["vesting-schedules"])
				require.True(t, cmdNames["vote-lock"])
				require.True(t, cmdNames["vote-locks"])
				require.True(t, cmdNames["voting-power"])
				require.True(t, cmdNames["pending-treasury-tx"])
				require.True(t, cmdNames["pending-treasury-txs"])
				require.True(t, cmdNames["inflation-metrics"])
				require.True(t, cmdNames["inflation-alerts"])
				require.True(t, cmdNames["liquidity-mining-stats"])
				require.True(t, cmdNames["mev-stats"])
				require.True(t, cmdNames["user-mev-balance"])
				require.True(t, cmdNames["tokenomics-stats"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCmdQueryParams(t *testing.T) {
	cmd := CmdQueryParams()
	require.NotNil(t, cmd)
	require.Equal(t, "params", cmd.Use)
	require.Contains(t, cmd.Short, "parameters")

	// Test no args required
	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)

	// Test too many args
	err = cmd.Args(cmd, []string{"extra"})
	require.Error(t, err)
}

func TestCmdQueryVestingSchedule(t *testing.T) {
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
			cmd := CmdQueryVestingSchedule()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "vesting-schedule")

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

func TestCmdQueryVestingSchedulesByBeneficiary(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid beneficiary address",
			args:    []string{"aura1beneficiary123"},
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
			cmd := CmdQueryVestingSchedulesByBeneficiary()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "vesting-schedules")

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

func TestCmdQueryVoteLock(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid lock ID",
			args:    []string{"lock-456"},
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
			cmd := CmdQueryVoteLock()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "vote-lock")

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

func TestCmdQueryVoteLocksByOwner(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid owner address",
			args:    []string{"aura1owner789"},
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
			cmd := CmdQueryVoteLocksByOwner()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "vote-locks")

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

func TestCmdQueryVotingPower(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid address",
			args:    []string{"aura1voter123"},
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
			cmd := CmdQueryVotingPower()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "voting-power")

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

func TestCmdQueryPendingTreasuryTx(t *testing.T) {
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
			cmd := CmdQueryPendingTreasuryTx()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "pending-treasury-tx")

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

func TestCmdQueryPendingTreasuryTxs(t *testing.T) {
	cmd := CmdQueryPendingTreasuryTxs()
	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "pending-treasury-txs")

	// Test no args required
	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestCmdQueryInflationMetrics(t *testing.T) {
	cmd := CmdQueryInflationMetrics()
	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "inflation-metrics")

	// Test no args required
	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestCmdQueryInflationAlerts(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no limit specified",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "with limit",
			args:    []string{"20"},
			wantErr: false,
		},
		{
			name:    "too many args",
			args:    []string{"10", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryInflationAlerts()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "inflation-alerts")

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

func TestCmdQueryLiquidityMiningStats(t *testing.T) {
	cmd := CmdQueryLiquidityMiningStats()
	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "liquidity-mining-stats")

	// Test no args required
	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestCmdQueryMEVStats(t *testing.T) {
	cmd := CmdQueryMEVStats()
	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "mev-stats")

	// Test no args required
	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestCmdQueryUserMEVBalance(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid address",
			args:    []string{"aura1mevuser456"},
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
			cmd := CmdQueryUserMEVBalance()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "user-mev-balance")

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

func TestCmdQueryTokenomicsStats(t *testing.T) {
	cmd := CmdQueryTokenomicsStats()
	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "tokenomics-stats")

	// Test no args required
	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestQueryCommandStructure(t *testing.T) {
	cmd := GetQueryCmd()

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

func TestAllQueryCommandsHaveHelp(t *testing.T) {
	commands := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"params", CmdQueryParams()},
		{"vesting-schedule", CmdQueryVestingSchedule()},
		{"vesting-schedules", CmdQueryVestingSchedulesByBeneficiary()},
		{"vote-lock", CmdQueryVoteLock()},
		{"vote-locks", CmdQueryVoteLocksByOwner()},
		{"voting-power", CmdQueryVotingPower()},
		{"pending-treasury-tx", CmdQueryPendingTreasuryTx()},
		{"pending-treasury-txs", CmdQueryPendingTreasuryTxs()},
		{"inflation-metrics", CmdQueryInflationMetrics()},
		{"inflation-alerts", CmdQueryInflationAlerts()},
		{"liquidity-mining-stats", CmdQueryLiquidityMiningStats()},
		{"mev-stats", CmdQueryMEVStats()},
		{"user-mev-balance", CmdQueryUserMEVBalance()},
		{"tokenomics-stats", CmdQueryTokenomicsStats()},
	}

	for _, tc := range commands {
		t.Run(tc.name+" has help text", func(t *testing.T) {
			require.NotEmpty(t, tc.cmd.Short, "command should have Short description")
		})
	}
}
