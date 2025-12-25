// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

type QueryTestSuite struct {
	suite.Suite
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}

// Test command registration
func (s *QueryTestSuite) TestGetQueryCmd() {
	cmd := GetQueryCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("inclusionroutines", cmd.Use)
	require.Equal("Querying commands for the inclusionroutines module", cmd.Short)
	require.Contains(cmd.Aliases, "ir")

	// Verify all expected subcommands are registered
	subCmds := cmd.Commands()
	require.GreaterOrEqual(len(subCmds), 5) // At least 5 commands

	expectedCmds := []string{
		"show",
		"list",
		"graph",
		"rate-limit",
		"params",
	}

	cmdNames := make(map[string]bool)
	for _, subCmd := range subCmds {
		cmdNames[subCmd.Name()] = true
	}

	for _, expectedCmd := range expectedCmds {
		require.True(cmdNames[expectedCmd], fmt.Sprintf("command %s should be registered", expectedCmd))
	}
}

// Test GetCmdQueryIR
func (s *QueryTestSuite) TestGetCmdQueryIR() {
	cmd := GetCmdQueryIR()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("show [ir-id]", cmd.Use)
	require.Equal("Query an Inclusion Routine by ID", cmd.Short)

	// Test that Args validator accepts 1 argument
	require.NotNil(cmd.Args)
	err := cmd.Args(cmd, []string{"gov-id-verify"})
	require.NoError(err)
}

func (s *QueryTestSuite) TestGetCmdQueryIR_ValidateArgs() {
	cmd := GetCmdQueryIR()
	require := s.Require()

	// Test insufficient args
	err := cmd.Args(cmd, []string{})
	require.Error(err)

	// Test too many args
	err = cmd.Args(cmd, []string{"ir-001", "extra"})
	require.Error(err)

	// Test exact args
	err = cmd.Args(cmd, []string{"biometric-face"})
	require.NoError(err)
}

// Test GetCmdQueryListIRs
func (s *QueryTestSuite) TestGetCmdQueryListIRs() {
	cmd := GetCmdQueryListIRs()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("list", cmd.Use)
	require.Equal("List all Inclusion Routines with optional filters", cmd.Short)

	// Test that Args validator accepts 0 arguments
	require.NotNil(cmd.Args)
	err := cmd.Args(cmd, []string{})
	require.NoError(err)

	// Test flags exist
	require.NotNil(cmd.Flags().Lookup("status"))
	require.NotNil(cmd.Flags().Lookup("arena"))
	require.NotNil(cmd.Flags().Lookup("locale"))
}

func (s *QueryTestSuite) TestGetCmdQueryListIRs_FlagDefaults() {
	cmd := GetCmdQueryListIRs()
	require := s.Require()

	// Check default values
	status, _ := cmd.Flags().GetInt32("status")
	require.Equal(int32(0), status)

	arena, _ := cmd.Flags().GetInt32("arena")
	require.Equal(int32(0), arena)

	locale, _ := cmd.Flags().GetString("locale")
	require.Equal("", locale)
}

func (s *QueryTestSuite) TestGetCmdQueryListIRs_SetFlags() {
	cmd := GetCmdQueryListIRs()
	require := s.Require()

	// Set various flags
	require.NoError(cmd.Flags().Set("status", "4"))
	require.NoError(cmd.Flags().Set("arena", "2"))
	require.NoError(cmd.Flags().Set("locale", "US"))

	status, _ := cmd.Flags().GetInt32("status")
	require.Equal(int32(4), status)

	arena, _ := cmd.Flags().GetInt32("arena")
	require.Equal(int32(2), arena)

	locale, _ := cmd.Flags().GetString("locale")
	require.Equal("US", locale)
}

// Test GetCmdQueryIRGraph
func (s *QueryTestSuite) TestGetCmdQueryIRGraph() {
	cmd := GetCmdQueryIRGraph()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("graph [ir-id]", cmd.Use)
	require.Equal("Query the prerequisite dependency graph for an IR", cmd.Short)

	// Test that Args validator accepts 1 argument
	require.NotNil(cmd.Args)
	err := cmd.Args(cmd, []string{"advanced-biometric"})
	require.NoError(err)
}

func (s *QueryTestSuite) TestGetCmdQueryIRGraph_ValidateArgs() {
	cmd := GetCmdQueryIRGraph()
	require := s.Require()

	// Test insufficient args
	err := cmd.Args(cmd, []string{})
	require.Error(err)

	// Test exact args
	err = cmd.Args(cmd, []string{"high-assurance-verify"})
	require.NoError(err)
}

// Test GetCmdQueryRateLimit
func (s *QueryTestSuite) TestGetCmdQueryRateLimit() {
	cmd := GetCmdQueryRateLimit()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("rate-limit [ir-id]", cmd.Use)
	require.Equal("Query rate limit settings for an IR", cmd.Short)

	// Test that Args validator accepts 1 argument
	require.NotNil(cmd.Args)
	err := cmd.Args(cmd, []string{"simple-captcha"})
	require.NoError(err)
}

func (s *QueryTestSuite) TestGetCmdQueryRateLimit_ValidateArgs() {
	cmd := GetCmdQueryRateLimit()
	require := s.Require()

	// Test insufficient args
	err := cmd.Args(cmd, []string{})
	require.Error(err)

	// Test exact args
	err = cmd.Args(cmd, []string{"biometric-face"})
	require.NoError(err)
}

// Test GetCmdQueryParams
func (s *QueryTestSuite) TestGetCmdQueryParams() {
	cmd := GetCmdQueryParams()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("params", cmd.Use)
	require.Equal("Query inclusionroutines module parameters", cmd.Short)

	// Test that Args validator accepts 0 arguments
	require.NotNil(cmd.Args)
	err := cmd.Args(cmd, []string{})
	require.NoError(err)
}

func (s *QueryTestSuite) TestGetCmdQueryParams_RejectsArgs() {
	cmd := GetCmdQueryParams()
	require := s.Require()

	// Test that extra args are rejected
	err := cmd.Args(cmd, []string{"extra"})
	require.Error(err)
}

// Test IRStatus enum values
func (s *QueryTestSuite) TestIRStatusEnumValues() {
	require := s.Require()

	tests := []struct {
		name     string
		value    int32
		expected v1beta1.IRStatus
	}{
		{"Unspecified", 0, v1beta1.IRStatus_IR_STATUS_UNSPECIFIED},
		{"Draft", 1, v1beta1.IRStatus_IR_STATUS_DRAFT},
		{"Reviewing", 2, v1beta1.IRStatus_IR_STATUS_REVIEWING},
		{"Approved", 3, v1beta1.IRStatus_IR_STATUS_APPROVED},
		{"Active", 4, v1beta1.IRStatus_IR_STATUS_ACTIVE},
		{"Suspended", 5, v1beta1.IRStatus_IR_STATUS_SUSPENDED},
		{"Deprecated", 6, v1beta1.IRStatus_IR_STATUS_DEPRECATED},
		{"Retired", 7, v1beta1.IRStatus_IR_STATUS_RETIRED},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			status := v1beta1.IRStatus(tt.value)
			require.Equal(tt.expected, status)
		})
	}
}

// Test Arena enum values for query filtering
func (s *QueryTestSuite) TestArenaEnumValuesForFiltering() {
	require := s.Require()

	tests := []struct {
		name     string
		value    int32
		expected v1beta1.Arena
	}{
		{"Unspecified", 0, v1beta1.Arena_ARENA_UNSPECIFIED},
		{"Anchor", 1, v1beta1.Arena_ARENA_ANCHOR},
		{"Biometric", 2, v1beta1.Arena_ARENA_BIOMETRIC},
		{"Possession", 3, v1beta1.Arena_ARENA_POSSESSION},
		{"Knowledge", 4, v1beta1.Arena_ARENA_KNOWLEDGE},
		{"Social", 5, v1beta1.Arena_ARENA_SOCIAL},
		{"GeoLocation", 6, v1beta1.Arena_ARENA_GEOLOCATION},
		{"HighAssurance", 7, v1beta1.Arena_ARENA_HIGH_ASSURANCE},
		{"Persistence", 8, v1beta1.Arena_ARENA_PERSISTENCE},
		{"Specialized", 9, v1beta1.Arena_ARENA_SPECIALIZED},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			arena := v1beta1.Arena(tt.value)
			require.Equal(tt.expected, arena)
		})
	}
}

// Test command Long descriptions contain examples
func (s *QueryTestSuite) TestCommandsHaveExamples() {
	require := s.Require()

	commands := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"show", GetCmdQueryIR()},
		{"list", GetCmdQueryListIRs()},
		{"graph", GetCmdQueryIRGraph()},
		{"rate-limit", GetCmdQueryRateLimit()},
		{"params", GetCmdQueryParams()},
	}

	for _, tc := range commands {
		s.Run(tc.name, func() {
			require.NotEmpty(tc.cmd.Long, "command %s should have a Long description", tc.name)
			require.Contains(tc.cmd.Long, "Examples:", "command %s should have Examples in Long description", tc.name)
		})
	}
}

// Test filter flag combinations
func (s *QueryTestSuite) TestListIRs_FilterCombinations() {
	require := s.Require()

	testCases := []struct {
		name   string
		status int32
		arena  int32
		locale string
	}{
		{
			name:   "active biometric in US",
			status: 4,
			arena:  2,
			locale: "US",
		},
		{
			name:   "all statuses in GLOBAL",
			status: 0,
			arena:  0,
			locale: "GLOBAL",
		},
		{
			name:   "anchor arena only",
			status: 0,
			arena:  1,
			locale: "",
		},
		{
			name:   "deprecated IRs",
			status: 6,
			arena:  0,
			locale: "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := GetCmdQueryListIRs()
			require.NoError(cmd.Flags().Set("status", fmt.Sprintf("%d", tc.status)))
			require.NoError(cmd.Flags().Set("arena", fmt.Sprintf("%d", tc.arena)))
			if tc.locale != "" {
				require.NoError(cmd.Flags().Set("locale", tc.locale))
			}

			status, _ := cmd.Flags().GetInt32("status")
			arena, _ := cmd.Flags().GetInt32("arena")
			locale, _ := cmd.Flags().GetString("locale")

			require.Equal(tc.status, status)
			require.Equal(tc.arena, arena)
			require.Equal(tc.locale, locale)
		})
	}
}

// Benchmark tests
func BenchmarkQueryIRCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cmd := GetCmdQueryIR()
		_ = cmd.Use
	}
}

func BenchmarkListIRsCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cmd := GetCmdQueryListIRs()
		_ = cmd.Use
	}
}
