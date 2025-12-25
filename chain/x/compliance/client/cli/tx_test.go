// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

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
				require.Equal(t, "compliance", cmd.Use)
				require.Equal(t, "Compliance transaction subcommands", cmd.Short)
				require.True(t, cmd.DisableFlagParsing)
			},
		},
		{
			name: "has all transaction subcommands",
			run: func(t *testing.T) {
				cmd := GetTxCmd()
				subcommands := cmd.Commands()
				require.Len(t, subcommands, 6)

				cmdNames := make(map[string]bool)
				for _, subcmd := range subcommands {
					cmdNames[subcmd.Name()] = true
				}

				require.True(t, cmdNames["submit-kyc"])
				require.True(t, cmdNames["report-suspicious"])
				require.True(t, cmdNames["screen-sanctions"])
				require.True(t, cmdNames["record-consent"])
				require.True(t, cmdNames["request-data"])
				require.True(t, cmdNames["generate-tax-report"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCmdSubmitKYC(t *testing.T) {
	// Valid 64-character hex string (32 bytes SHA-256 hash)
	validPIICommitment := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name: "valid KYC submission",
			args: []string{
				"aura1abc123",       // address
				"3",                 // kyc level (INTERMEDIATE)
				"Jumio",             // provider
				validPIICommitment,  // pii commitment hex (64 chars)
				"US",                // jurisdiction
			},
			wantErr: false,
		},
		{
			name: "basic KYC level",
			args: []string{
				"aura1xyz789",
				"2",                 // BASIC
				"Provider",
				validPIICommitment,  // pii commitment hex (64 chars)
				"UK",
			},
			wantErr: false,
		},
		{
			name:    "missing args",
			args:    []string{"aura1abc", "3"},
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
			cmd := CmdSubmitKYC()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "submit-kyc")
			require.Contains(t, cmd.Short, "Submit KYC verification")

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

func TestCmdReportSuspiciousActivity(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name: "valid suspicious activity report",
			args: []string{
				"aura1address",
				"ABC123DEF456",
				"structuring",
				"Multiple small transactions under reporting threshold",
			},
			wantErr: false,
		},
		{
			name: "smurfing activity",
			args: []string{
				"aura1smurfer",
				"HASH789XYZ",
				"smurfing",
				"Pattern detected across multiple accounts",
			},
			wantErr: false,
		},
		{
			name:    "missing description",
			args:    []string{"aura1addr", "hash", "type"},
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
			cmd := CmdReportSuspiciousActivity()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "report-suspicious")

			// Verify flag exists
			flag := cmd.Flags().Lookup("indicators")
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

func TestCmdScreenSanctions(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid sanctions screening",
			args:    []string{"aura1screenme"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"aura1addr", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdScreenSanctions()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "screen-sanctions")
			require.Contains(t, cmd.Short, "Screen address against sanctions")

			// Verify force-refresh flag exists
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

func TestCmdRecordGDPRConsent(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name: "valid consent - true",
			args: []string{
				"aura1user",
				"data_processing",
				"true",
				"v1.2",
			},
			wantErr: false,
		},
		{
			name: "valid consent - false",
			args: []string{
				"aura1user",
				"marketing",
				"false",
				"v1.2",
			},
			wantErr: false,
		},
		{
			name: "analytics consent",
			args: []string{
				"aura1analytics",
				"analytics",
				"true",
				"v2.0",
			},
			wantErr: false,
		},
		{
			name:    "missing version",
			args:    []string{"aura1addr", "type", "true"},
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
			cmd := CmdRecordGDPRConsent()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "record-consent")
			require.Contains(t, cmd.Short, "Record GDPR consent")

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

func TestCmdRequestGDPRData(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "request access",
			args:    []string{"aura1user", "access"},
			wantErr: false,
		},
		{
			name:    "request rectification",
			args:    []string{"aura1user", "rectification"},
			wantErr: false,
		},
		{
			name:    "request erasure",
			args:    []string{"aura1user", "erasure"},
			wantErr: false,
		},
		{
			name:    "request portability",
			args:    []string{"aura1user", "portability"},
			wantErr: false,
		},
		{
			name:    "missing request type",
			args:    []string{"aura1user"},
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
			cmd := CmdRequestGDPRData()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "request-data")
			require.Contains(t, cmd.Short, "Request GDPR data")

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

func TestCmdGenerateTaxReport(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name: "valid tax report - US 1099-MISC",
			args: []string{
				"aura1taxpayer",
				"2024",
				"US",
				"1099-MISC",
			},
			wantErr: false,
		},
		{
			name: "valid tax report - US 1099-K",
			args: []string{
				"aura1taxpayer",
				"2024",
				"US",
				"1099-K",
			},
			wantErr: false,
		},
		{
			name: "valid tax report - UK comprehensive",
			args: []string{
				"aura1uktaxpayer",
				"2023",
				"UK",
				"comprehensive",
			},
			wantErr: false,
		},
		{
			name: "valid tax report - 8949",
			args: []string{
				"aura1investor",
				"2024",
				"US",
				"8949",
			},
			wantErr: false,
		},
		{
			name:    "missing report type",
			args:    []string{"aura1addr", "2024", "US"},
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
			cmd := CmdGenerateTaxReport()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "generate-tax-report")
			require.Contains(t, cmd.Short, "Generate tax report")

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

func TestSubmitKYCDocumentsFlag(t *testing.T) {
	cmd := CmdSubmitKYC()

	// The documents flag was removed in favor of PII commitment-based storage.
	// This test now verifies that the flag does NOT exist.
	flag := cmd.Flags().Lookup("documents")
	require.Nil(t, flag, "documents flag should not exist - using PII commitment instead")

	// Verify command structure
	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "submit-kyc")
	require.Contains(t, cmd.Use, "pii-commitment-hex", "command should accept PII commitment hex parameter")
}

func TestReportSuspiciousIndicatorsFlag(t *testing.T) {
	cmd := CmdReportSuspiciousActivity()

	flag := cmd.Flags().Lookup("indicators")
	require.NotNil(t, flag)
	require.Equal(t, "string", flag.Value.Type())
	require.Contains(t, flag.Usage, "indicator")
}

func TestAllCommandsHaveHelp(t *testing.T) {
	commands := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"submit-kyc", CmdSubmitKYC()},
		{"report-suspicious", CmdReportSuspiciousActivity()},
		{"screen-sanctions", CmdScreenSanctions()},
		{"record-consent", CmdRecordGDPRConsent()},
		{"request-data", CmdRequestGDPRData()},
		{"generate-tax-report", CmdGenerateTaxReport()},
	}

	for _, tc := range commands {
		t.Run(tc.name+" has help text", func(t *testing.T) {
			require.NotEmpty(t, tc.cmd.Short, "command should have Short description")
			require.NotEmpty(t, tc.cmd.Long, "command should have Long description")
		})
	}
}
