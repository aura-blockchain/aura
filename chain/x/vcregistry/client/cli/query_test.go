package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	vcregistryv1beta1 "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
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
	require.Equal("vcregistry", cmd.Use)
	require.Equal("Querying commands for the VC Registry module", cmd.Short)

	// Verify all expected subcommands are registered
	subCmds := cmd.Commands()
	require.GreaterOrEqual(len(subCmds), 12) // At least 12 query commands

	expectedCmds := []string{
		"vc",
		"user-vcs",
		"vc-status",
		"batch-vc-status",
		"policy",
		"policies",
		"revocation-list",
		"check-revocation",
		"resolve-did",
		"did-by-address",
		"validate-mint",
		"stats",
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

// Test CmdQueryVC
func (s *QueryTestSuite) TestCmdQueryVC() {
	cmd := CmdQueryVC()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("vc [vc-id]", cmd.Use)
	require.NoError(cmd.Args(cmd, []string{"vc-123"}))
}

func (s *QueryTestSuite) TestCmdQueryVC_ValidateArgs() {
	cmd := CmdQueryVC()
	require := s.Require()

	// Test insufficient args
	err := cmd.Args(cmd, []string{})
	require.Error(err)

	// Test exact args
	err = cmd.Args(cmd, []string{"vc-123456"})
	require.NoError(err)
}

// Test CmdQueryUserVCs
func (s *QueryTestSuite) TestCmdQueryUserVCs() {
	cmd := CmdQueryUserVCs()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("user-vcs [holder-address]", cmd.Use)
	require.NoError(cmd.Args(cmd, []string{"aura1abc..."}))

	// Test flags
	require.NotNil(cmd.Flags().Lookup("status"))
	require.NotNil(cmd.Flags().Lookup("type"))
}

// Test CmdQueryVCStatus
func (s *QueryTestSuite) TestCmdQueryVCStatus() {
	cmd := CmdQueryVCStatus()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("vc-status [vc-id]", cmd.Use)
	require.NoError(cmd.Args(cmd, []string{"vc-123"}))
}

// Test CmdQueryBatchVCStatus
func (s *QueryTestSuite) TestCmdQueryBatchVCStatus() {
	cmd := CmdQueryBatchVCStatus()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("batch-vc-status [vc-ids]", cmd.Use)
	require.NoError(cmd.Args(cmd, []string{"vc-123,vc-456"}))
}

// Test CmdQueryVCPolicy
func (s *QueryTestSuite) TestCmdQueryVCPolicy() {
	cmd := CmdQueryVCPolicy()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("policy [vc-type-name]", cmd.Use)
	require.NoError(cmd.Args(cmd, []string{"Verified Developer"}))
}

// Test CmdQueryVCPolicies
func (s *QueryTestSuite) TestCmdQueryVCPolicies() {
	cmd := CmdQueryVCPolicies()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("policies", cmd.Use)
	require.NoError(cmd.Args(cmd, []string{}))

	// Test flags
	require.NotNil(cmd.Flags().Lookup("status"))
}

// Test CmdQueryRevocationList
func (s *QueryTestSuite) TestCmdQueryRevocationList() {
	cmd := CmdQueryRevocationList()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("revocation-list", cmd.Use)
	require.NoError(cmd.Args(cmd, []string{}))
}

// Test CmdQueryCheckRevocation
func (s *QueryTestSuite) TestCmdQueryCheckRevocation() {
	cmd := CmdQueryCheckRevocation()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("check-revocation [vc-id]", cmd.Use)
	require.NoError(cmd.Args(cmd, []string{"vc-123"}))
}

// Test CmdQueryResolveDID
func (s *QueryTestSuite) TestCmdQueryResolveDID() {
	cmd := CmdQueryResolveDID()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("resolve-did [did]", cmd.Use)
	require.NoError(cmd.Args(cmd, []string{"did:aura:mainnet:user123"}))
}

// Test CmdQueryDIDByAddress
func (s *QueryTestSuite) TestCmdQueryDIDByAddress() {
	cmd := CmdQueryDIDByAddress()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("did-by-address [controller-address]", cmd.Use)
	require.NoError(cmd.Args(cmd, []string{"aura1abc..."}))
}

// Test CmdQueryValidateMintEligibility
func (s *QueryTestSuite) TestCmdQueryValidateMintEligibility() {
	cmd := CmdQueryValidateMintEligibility()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("validate-mint [holder-address] [vc-type]", cmd.Use)
	require.NoError(cmd.Args(cmd, []string{"aura1abc...", "VC_TYPE_VERIFIED_HUMAN"}))

	// Test flags
	require.NotNil(cmd.Flags().Lookup("custom-type"))
}

// Test CmdQueryStats
func (s *QueryTestSuite) TestCmdQueryStats() {
	cmd := CmdQueryStats()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("stats", cmd.Use)
	require.NoError(cmd.Args(cmd, []string{}))
}

// Test CmdQueryParams
func (s *QueryTestSuite) TestCmdQueryParams() {
	cmd := CmdQueryParams()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("params", cmd.Use)
	require.NoError(cmd.Args(cmd, []string{}))
}

// Test helper function parseVCStatus
func (s *QueryTestSuite) TestParseVCStatus() {
	require := s.Require()

	tests := []struct {
		name     string
		input    string
		expected vcregistryv1beta1.VCStatus
		hasError bool
	}{
		{"Pending", "VC_STATUS_PENDING", vcregistryv1beta1.VCStatus_VC_STATUS_PENDING, false},
		{"Active", "VC_STATUS_ACTIVE", vcregistryv1beta1.VCStatus_VC_STATUS_ACTIVE, false},
		{"Revoked", "VC_STATUS_REVOKED", vcregistryv1beta1.VCStatus_VC_STATUS_REVOKED, false},
		{"Expired", "VC_STATUS_EXPIRED", vcregistryv1beta1.VCStatus_VC_STATUS_EXPIRED, false},
		{"Suspended", "VC_STATUS_SUSPENDED", vcregistryv1beta1.VCStatus_VC_STATUS_SUSPENDED, false},
		{"LowerCase", "vc_status_active", vcregistryv1beta1.VCStatus_VC_STATUS_ACTIVE, false},
		{"Invalid", "INVALID_STATUS", vcregistryv1beta1.VCStatus_VC_STATUS_UNSPECIFIED, true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := parseVCStatus(tt.input)
			if tt.hasError {
				require.Error(err)
				require.Equal(tt.expected, result)
			} else {
				require.NoError(err)
				require.Equal(tt.expected, result)
			}
		})
	}
}

// Test helper function parseVCPolicyStatus
func (s *QueryTestSuite) TestParseVCPolicyStatus() {
	require := s.Require()

	tests := []struct {
		name     string
		input    string
		expected vcregistryv1beta1.VCPolicyStatus
		hasError bool
	}{
		{"Draft", "VC_POLICY_STATUS_DRAFT", vcregistryv1beta1.VCPolicyStatus_VC_POLICY_STATUS_DRAFT, false},
		{"Active", "VC_POLICY_STATUS_ACTIVE", vcregistryv1beta1.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE, false},
		{"Deprecated", "VC_POLICY_STATUS_DEPRECATED", vcregistryv1beta1.VCPolicyStatus_VC_POLICY_STATUS_DEPRECATED, false},
		{"LowerCase", "vc_policy_status_active", vcregistryv1beta1.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE, false},
		{"Invalid", "INVALID_STATUS", vcregistryv1beta1.VCPolicyStatus_VC_POLICY_STATUS_UNSPECIFIED, true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := parseVCPolicyStatus(tt.input)
			if tt.hasError {
				require.Error(err)
				require.Equal(tt.expected, result)
			} else {
				require.NoError(err)
				require.Equal(tt.expected, result)
			}
		})
	}
}

// Test batch VC ID parsing
func (s *QueryTestSuite) TestBatchVCStatus_IDParsing() {
	require := s.Require()

	testCases := []struct {
		name          string
		input         string
		expectedCount int
	}{
		{
			name:          "single VC ID",
			input:         "vc-123",
			expectedCount: 1,
		},
		{
			name:          "multiple VC IDs",
			input:         "vc-123,vc-456,vc-789",
			expectedCount: 3,
		},
		{
			name:          "IDs with spaces",
			input:         "vc-123 , vc-456 , vc-789",
			expectedCount: 3,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Simulate the parsing logic from the command
			vcIDs := append([]string{}, splitAndTrim(tc.input, ",")...)
			require.Equal(tc.expectedCount, len(vcIDs))
		})
	}
}

// Test filter combinations
func (s *QueryTestSuite) TestUserVCs_FilterCombinations() {
	cmd := CmdQueryUserVCs()
	require := s.Require()

	// Set status filter
	require.NoError(cmd.Flags().Set("status", "VC_STATUS_ACTIVE"))
	status, err := cmd.Flags().GetString("status")
	require.NoError(err)
	require.Equal("VC_STATUS_ACTIVE", status)

	// Set type filter
	require.NoError(cmd.Flags().Set("type", "VC_TYPE_VERIFIED_HUMAN"))
	vcType, err := cmd.Flags().GetString("type")
	require.NoError(err)
	require.Equal("VC_TYPE_VERIFIED_HUMAN", vcType)
}

// Test pagination flags
func (s *QueryTestSuite) TestPaginationFlags() {
	cmd := CmdQueryUserVCs()
	require := s.Require()

	// Pagination flags should be present (added by AddPaginationFlagsToCmd)
	require.NotNil(cmd.Flags().Lookup("page-key"))
	require.NotNil(cmd.Flags().Lookup("offset"))
	require.NotNil(cmd.Flags().Lookup("limit"))
	require.NotNil(cmd.Flags().Lookup("count-total"))
	require.NotNil(cmd.Flags().Lookup("reverse"))
}

// Helper function
func splitAndTrim(s, sep string) []string {
	var result []string
	for _, part := range split(s, sep) {
		trimmed := trim(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func split(s, sep string) []string {
	result := make([]string, 0)
	current := ""
	for _, c := range s {
		if string(c) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" || len(result) == 0 {
		result = append(result, current)
	}
	return result
}

func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// Benchmark tests
func BenchmarkParseVCStatus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := parseVCStatus("VC_STATUS_ACTIVE"); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkParseVCPolicyStatus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := parseVCPolicyStatus("VC_POLICY_STATUS_ACTIVE"); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
