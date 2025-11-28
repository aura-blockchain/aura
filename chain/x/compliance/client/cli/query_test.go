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
				require.Equal(t, "compliance", cmd.Use)
				require.Equal(t, "Querying commands for the compliance module", cmd.Short)
				require.True(t, cmd.DisableFlagParsing)
			},
		},
		{
			name: "has all subcommands",
			run: func(t *testing.T) {
				cmd := GetQueryCmd()
				subcommands := cmd.Commands()
				require.Len(t, subcommands, 5)

				// Verify each subcommand exists
				cmdNames := make(map[string]bool)
				for _, subcmd := range subcommands {
					cmdNames[subcmd.Use] = true
				}

				require.True(t, cmdNames["kyc-record [address]"])
				require.True(t, cmdNames["aml-profile [address]"])
				require.True(t, cmdNames["sanctions [address]"])
				require.True(t, cmdNames["alerts [address]"])
				require.True(t, cmdNames["tax-report [address] [tax-year] [jurisdiction]"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCmdQueryKYCRecord(t *testing.T) {
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
			cmd := CmdQueryKYCRecord()
			require.NotNil(t, cmd)
			require.Equal(t, "kyc-record [address]", cmd.Use)
			require.Equal(t, "Query KYC record for an address", cmd.Short)

			// Test arg validation
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

func TestCmdQueryAMLProfile(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid address",
			args:    []string{"aura1xyz789"},
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
			cmd := CmdQueryAMLProfile()
			require.NotNil(t, cmd)
			require.Equal(t, "aml-profile [address]", cmd.Use)
			require.Equal(t, "Query AML risk profile for an address", cmd.Short)

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

func TestCmdQuerySanctionsScreening(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		flags   map[string]string
		wantErr bool
	}{
		{
			name:    "basic query",
			args:    []string{"aura1address"},
			wantErr: false,
		},
		{
			name:    "with force refresh flag",
			args:    []string{"aura1address"},
			flags:   map[string]string{"force-refresh": "true"},
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
			cmd := CmdQuerySanctionsScreening()
			require.NotNil(t, cmd)
			require.Equal(t, "sanctions [address]", cmd.Use)
			require.Contains(t, cmd.Short, "Query sanctions screening")

			// Verify flag exists
			flag := cmd.Flags().Lookup("force-refresh")
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

func TestCmdQueryTransactionAlerts(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid query",
			args:    []string{"aura1alertaddr"},
			wantErr: false,
		},
		{
			name:    "missing address",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryTransactionAlerts()
			require.NotNil(t, cmd)
			require.Equal(t, "alerts [address]", cmd.Use)

			// Verify flag exists
			flag := cmd.Flags().Lookup("unreviewed-only")
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

func TestCmdQueryTaxReport(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid tax report query",
			args:    []string{"aura1addr", "2024", "US"},
			wantErr: false,
		},
		{
			name:    "uk jurisdiction",
			args:    []string{"aura1addr", "2023", "UK"},
			wantErr: false,
		},
		{
			name:    "missing jurisdiction",
			args:    []string{"aura1addr", "2024"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"aura1addr", "2024", "US", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryTaxReport()
			require.NotNil(t, cmd)
			require.Equal(t, "tax-report [address] [tax-year] [jurisdiction]", cmd.Use)
			require.Contains(t, cmd.Short, "Query tax report")

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
	cmd := GetQueryCmd()

	// Test command properties
	require.Equal(t, "compliance", cmd.Use)
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
