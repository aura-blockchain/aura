// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestGetQueryCmd(t *testing.T) {
	cmd := GetQueryCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "monitoring", cmd.Use)
	require.Contains(t, cmd.Short, "monitoring module")
	require.True(t, cmd.DisableFlagParsing)

	// Check all subcommands exist
	subcommands := cmd.Commands()
	require.Len(t, subcommands, 10)
}

func TestCmdQueryParams(t *testing.T) {
	cmd := CmdQueryParams()
	require.NotNil(t, cmd)
	require.Equal(t, "params", cmd.Use)

	// No args required
	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestCmdQueryAlerts(t *testing.T) {
	cmd := CmdQueryAlerts()
	require.NotNil(t, cmd)
	require.Equal(t, "alerts", cmd.Use)

	// Check flags exist
	require.NotNil(t, cmd.Flags().Lookup("severity"))
	require.NotNil(t, cmd.Flags().Lookup("unresolved"))

	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestCmdQueryAlert(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryAlert()
			require.NotNil(t, cmd)

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

func TestCmdQueryNetworkHealth(t *testing.T) {
	cmd := CmdQueryNetworkHealth()
	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "network-health")

	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestCmdQueryValidatorUptime(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid validator address",
			args:    []string{"auravaloper1abc123"},
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
			cmd := CmdQueryValidatorUptime()
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

func TestCmdQueryGasPriceTracking(t *testing.T) {
	cmd := CmdQueryGasPriceTracking()
	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "gas-price-tracking")

	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestCmdQueryTVLMonitoring(t *testing.T) {
	cmd := CmdQueryTVLMonitoring()
	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "tvl-monitoring")

	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestCmdQueryTransactionMonitor(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid tx hash",
			args:    []string{"ABCD1234"},
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
			cmd := CmdQueryTransactionMonitor()
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

func TestCmdQueryAnomalies(t *testing.T) {
	cmd := CmdQueryAnomalies()
	require.NotNil(t, cmd)
	require.Equal(t, "anomalies", cmd.Use)

	// Check flags
	require.NotNil(t, cmd.Flags().Lookup("type"))
	require.NotNil(t, cmd.Flags().Lookup("above-threshold"))

	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestCmdQuerySecurityEvents(t *testing.T) {
	cmd := CmdQuerySecurityEvents()
	require.NotNil(t, cmd)
	require.Equal(t, "security-events", cmd.Use)

	// Check flags
	require.NotNil(t, cmd.Flags().Lookup("severity"))

	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestAllQueryCommandsHaveHelp(t *testing.T) {
	commands := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"params", CmdQueryParams()},
		{"alerts", CmdQueryAlerts()},
		{"alert", CmdQueryAlert()},
		{"network-health", CmdQueryNetworkHealth()},
		{"validator-uptime", CmdQueryValidatorUptime()},
		{"gas-price-tracking", CmdQueryGasPriceTracking()},
		{"tvl-monitoring", CmdQueryTVLMonitoring()},
		{"transaction-monitor", CmdQueryTransactionMonitor()},
		{"anomalies", CmdQueryAnomalies()},
		{"security-events", CmdQuerySecurityEvents()},
	}

	for _, tc := range commands {
		t.Run(tc.name+" has help", func(t *testing.T) {
			require.NotEmpty(t, tc.cmd.Short)
		})
	}
}
