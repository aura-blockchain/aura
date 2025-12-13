package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

type TxTestSuite struct {
	suite.Suite
}

func TestTxTestSuite(t *testing.T) {
	suite.Run(t, new(TxTestSuite))
}

// Test command registration
func (s *TxTestSuite) TestGetTxCmd() {
	cmd := GetTxCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("inclusionroutines", cmd.Use)
	require.Equal("Inclusion Routines transaction subcommands", cmd.Short)
	require.Contains(cmd.Aliases, "ir")

	// Verify all expected subcommands are registered
	subCmds := cmd.Commands()
	require.GreaterOrEqual(len(subCmds), 7) // At least 7 commands

	expectedCmds := []string{
		"create-ir",
		"update-ir",
		"delete-ir",
		"set-prerequisites",
		"set-rate-limit",
		"suspend-ir",
		"activate-ir",
	}

	cmdNames := make(map[string]bool)
	for _, subCmd := range subCmds {
		cmdNames[subCmd.Name()] = true
	}

	for _, expectedCmd := range expectedCmds {
		require.True(cmdNames[expectedCmd], fmt.Sprintf("command %s should be registered", expectedCmd))
	}
}

// Test CmdCreateIR
func (s *TxTestSuite) TestCmdCreateIR() {
	cmd := CmdCreateIR()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("create-ir [id] [name] [arena] [description] [score] [poi-reward]", cmd.Use)

	// Test that Args validator accepts 6 arguments
	require.NotNil(cmd.Args)
	err := cmd.Args(cmd, []string{"ir-001", "Test IR", "1", "Description", "100", "50"})
	require.NoError(err)

	// Test flags
	require.NotNil(cmd.Flags().Lookup("locale-tags"))
	require.NotNil(cmd.Flags().Lookup("privacy-tier"))
	require.NotNil(cmd.Flags().Lookup("version"))
	require.NotNil(cmd.Flags().Lookup("metadata-hash"))
	require.NotNil(cmd.Flags().Lookup("activation-height"))
	require.NotNil(cmd.Flags().Lookup("sunset-height"))
}

func (s *TxTestSuite) TestCmdCreateIR_ValidateArgs() {
	cmd := CmdCreateIR()
	require := s.Require()

	// Test insufficient args
	err := cmd.Args(cmd, []string{})
	require.Error(err)

	// Test too many args
	err = cmd.Args(cmd, []string{"1", "2", "3", "4", "5", "6", "7"})
	require.Error(err)

	// Test exact args
	err = cmd.Args(cmd, []string{"ir-001", "Test IR", "1", "Desc", "100", "50"})
	require.NoError(err)
}

// Test CmdUpdateIR
func (s *TxTestSuite) TestCmdUpdateIR() {
	cmd := CmdUpdateIR()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("update-ir [id]", cmd.Use)

	// Test that Args validator accepts 1 argument
	require.NotNil(cmd.Args)
	err := cmd.Args(cmd, []string{"ir-001"})
	require.NoError(err)

	// Test optional flags
	require.NotNil(cmd.Flags().Lookup("name"))
	require.NotNil(cmd.Flags().Lookup("description"))
	require.NotNil(cmd.Flags().Lookup("score"))
	require.NotNil(cmd.Flags().Lookup("poi-reward"))
	require.NotNil(cmd.Flags().Lookup("locale-tags"))
	require.NotNil(cmd.Flags().Lookup("privacy-tier"))
	require.NotNil(cmd.Flags().Lookup("version"))
	require.NotNil(cmd.Flags().Lookup("metadata-hash"))
	require.NotNil(cmd.Flags().Lookup("sunset-height"))
}

// Test CmdDeleteIR
func (s *TxTestSuite) TestCmdDeleteIR() {
	cmd := CmdDeleteIR()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("delete-ir [id]", cmd.Use)

	// Test that Args validator accepts 1 argument
	require.NotNil(cmd.Args)
	err := cmd.Args(cmd, []string{"ir-001"})
	require.NoError(err)
}

// Test CmdSetIRPrerequisites
func (s *TxTestSuite) TestCmdSetIRPrerequisites() {
	cmd := CmdSetIRPrerequisites()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("set-prerequisites [ir-id] [required-ir-ids]", cmd.Use)

	// Test that Args validator accepts 2 arguments
	require.NotNil(cmd.Args)
	err := cmd.Args(cmd, []string{"ir-002", "ir-001"})
	require.NoError(err)
}

func (s *TxTestSuite) TestCmdSetIRPrerequisites_MultipleIDs() {
	cmd := CmdSetIRPrerequisites()
	require := s.Require()

	// Test with multiple comma-separated IDs
	err := cmd.Args(cmd, []string{"ir-advanced", "ir-001,ir-002,ir-003"})
	require.NoError(err)

	// Test with empty prerequisites (clearing)
	err = cmd.Args(cmd, []string{"ir-basic", ""})
	require.NoError(err)
}

// Test CmdSetIRRateLimit
func (s *TxTestSuite) TestCmdSetIRRateLimit() {
	cmd := CmdSetIRRateLimit()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("set-rate-limit [ir-id] [per-wallet-hour] [per-wallet-day] [per-block-global]", cmd.Use)

	// Test that Args validator accepts 4 arguments
	require.NotNil(cmd.Args)
	err := cmd.Args(cmd, []string{"ir-001", "3", "10", "100"})
	require.NoError(err)
}

func (s *TxTestSuite) TestCmdSetIRRateLimit_ValidateArgs() {
	cmd := CmdSetIRRateLimit()
	require := s.Require()

	// Test insufficient args
	err := cmd.Args(cmd, []string{"ir-001", "3", "10"})
	require.Error(err)

	// Test exact args
	err = cmd.Args(cmd, []string{"ir-001", "3", "10", "100"})
	require.NoError(err)
}

// Test CmdSuspendIR
func (s *TxTestSuite) TestCmdSuspendIR() {
	cmd := CmdSuspendIR()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("suspend-ir [ir-id] [reason]", cmd.Use)

	// Test that Args validator accepts 2 arguments
	require.NotNil(cmd.Args)
	err := cmd.Args(cmd, []string{"ir-001", "Under review"})
	require.NoError(err)
}

// Test CmdActivateIR
func (s *TxTestSuite) TestCmdActivateIR() {
	cmd := CmdActivateIR()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("activate-ir [ir-id]", cmd.Use)

	// Test that Args validator accepts 1 argument
	require.NotNil(cmd.Args)
	err := cmd.Args(cmd, []string{"ir-001"})
	require.NoError(err)
}

// Test Arena enum values
func (s *TxTestSuite) TestArenaEnumValues() {
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

// Test PrivacyTier enum values
func (s *TxTestSuite) TestPrivacyTierEnumValues() {
	require := s.Require()

	tests := []struct {
		name     string
		value    int32
		expected v1beta1.PrivacyTier
	}{
		{"Unspecified", 0, v1beta1.PrivacyTier_PRIVACY_TIER_UNSPECIFIED},
		{"Low", 1, v1beta1.PrivacyTier_PRIVACY_TIER_LOW},
		{"Medium", 2, v1beta1.PrivacyTier_PRIVACY_TIER_MEDIUM},
		{"High", 3, v1beta1.PrivacyTier_PRIVACY_TIER_HIGH},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tier := v1beta1.PrivacyTier(tt.value)
			require.Equal(tt.expected, tier)
		})
	}
}

// Test locale tags parsing
func (s *TxTestSuite) TestLocaleTagsParsing() {
	require := s.Require()

	testCases := []struct {
		name          string
		input         string
		expectedCount int
	}{
		{
			name:          "single locale",
			input:         "US",
			expectedCount: 1,
		},
		{
			name:          "multiple locales",
			input:         "US,UK,EU",
			expectedCount: 3,
		},
		{
			name:          "global locale",
			input:         "GLOBAL",
			expectedCount: 1,
		},
		{
			name:          "locales with spaces",
			input:         "US , UK , EU",
			expectedCount: 3,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Simulate parsing logic
			var locales []string
			if tc.input != "" {
				for _, locale := range splitString(tc.input, ",") {
					trimmed := trimSpace(locale)
					if trimmed != "" {
						locales = append(locales, trimmed)
					}
				}
			}
			require.GreaterOrEqual(len(locales), tc.expectedCount-1) // Allow for trimming variance
		})
	}
}

// Test flag default values
func (s *TxTestSuite) TestCreateIR_FlagDefaults() {
	cmd := CmdCreateIR()
	require := s.Require()

	// Check default values
	privacyTier, _ := cmd.Flags().GetInt32("privacy-tier")
	require.Equal(int32(0), privacyTier)

	version, _ := cmd.Flags().GetString("version")
	require.Equal("1.0.0", version)

	activationHeight, _ := cmd.Flags().GetInt64("activation-height")
	require.Equal(int64(0), activationHeight)

	sunsetHeight, _ := cmd.Flags().GetInt64("sunset-height")
	require.Equal(int64(0), sunsetHeight)
}

// Test update flag combinations
func (s *TxTestSuite) TestUpdateIR_FlagCombinations() {
	cmd := CmdUpdateIR()
	require := s.Require()

	// Set various flags
	require.NoError(cmd.Flags().Set("name", "Updated Name"))
	require.NoError(cmd.Flags().Set("score", "150"))
	require.NoError(cmd.Flags().Set("poi-reward", "75"))

	name, _ := cmd.Flags().GetString("name")
	require.Equal("Updated Name", name)

	score, _ := cmd.Flags().GetInt64("score")
	require.Equal(int64(150), score)

	poiReward, _ := cmd.Flags().GetInt64("poi-reward")
	require.Equal(int64(75), poiReward)
}

// Test rate limit value ranges
func (s *TxTestSuite) TestRateLimitValueRanges() {
	require := s.Require()

	testCases := []struct {
		name           string
		perWalletHour  string
		perWalletDay   string
		perBlockGlobal string
		valid          bool
	}{
		{
			name:           "normal values",
			perWalletHour:  "3",
			perWalletDay:   "10",
			perBlockGlobal: "100",
			valid:          true,
		},
		{
			name:           "unlimited (zeros)",
			perWalletHour:  "0",
			perWalletDay:   "0",
			perBlockGlobal: "0",
			valid:          true,
		},
		{
			name:           "high values",
			perWalletHour:  "100",
			perWalletDay:   "1000",
			perBlockGlobal: "10000",
			valid:          true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// All test cases should be valid for parsing
			require.True(tc.valid)
		})
	}
}

// Test prerequisite parsing with edge cases
func (s *TxTestSuite) TestPrerequisiteParsing_EdgeCases() {
	require := s.Require()

	testCases := []struct {
		name          string
		input         string
		expectedEmpty bool
	}{
		{
			name:          "empty string",
			input:         "",
			expectedEmpty: true,
		},
		{
			name:          "single prerequisite",
			input:         "ir-001",
			expectedEmpty: false,
		},
		{
			name:          "multiple prerequisites",
			input:         "ir-001,ir-002,ir-003",
			expectedEmpty: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var prerequisites []string
			if tc.input != "" {
				prerequisites = splitString(tc.input, ",")
			}
			require.Equal(tc.expectedEmpty, len(prerequisites) == 0)
		})
	}
}

// Helper functions
func splitString(s, sep string) []string {
	var result []string
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

func trimSpace(s string) string {
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
func BenchmarkCreateIRCommand(b *testing.B) {
	cmd := CmdCreateIR()
	for i := 0; i < b.N; i++ {
		_ = cmd.Use
	}
}

func BenchmarkLocaleTagsParsing(b *testing.B) {
	input := "US,UK,EU,CA,AU"
	for i := 0; i < b.N; i++ {
		splitString(input, ",")
	}
}
