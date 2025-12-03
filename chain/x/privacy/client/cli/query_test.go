package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type QueryCLITestSuite struct {
	suite.Suite
}

func TestQueryCLITestSuite(t *testing.T) {
	t.Skip("Privacy CLI query tests require a live node; skipping in unit runs")
	suite.Run(t, new(QueryCLITestSuite))
}

// TestGetQueryCmd tests that GetQueryCmd returns a properly configured command
func (s *QueryCLITestSuite) TestGetQueryCmd() {
	cmd := GetQueryCmd()

	s.Require().NotNil(cmd)
	s.Require().Equal("privacy", cmd.Use)
	s.Require().True(cmd.DisableFlagParsing)
	s.Require().Greater(len(cmd.Commands()), 0)
}

// TestCmdQueryParams tests params query command
func (s *QueryCLITestSuite) TestCmdQueryParams() {
	cmd := CmdQueryParams()

	s.Require().NotNil(cmd)
	s.Require().Equal("params", cmd.Use)

	// Test with no arguments
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	s.Require().Error(err) // Will fail without context
}

// TestCmdQueryMixingPool tests mixing pool query command
func (s *QueryCLITestSuite) TestCmdQueryMixingPool() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid pool ID",
			args:      []string{"pool-123"},
			expectErr: false,
		},
		{
			name:      "another valid pool ID",
			args:      []string{"pool-456"},
			expectErr: false,
		},
		{
			name:      "missing pool ID",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "too many arguments",
			args:      []string{"pool-123", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryMixingPool()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQueryMixingPools tests mixing pools query command
func (s *QueryCLITestSuite) TestCmdQueryMixingPools() {
	tests := []struct {
		name      string
		args      []string
		flags     map[string]string
		expectErr bool
	}{
		{
			name:      "without status filter",
			args:      []string{},
			expectErr: false,
		},
		{
			name:      "with OPEN status",
			args:      []string{},
			flags:     map[string]string{"status": "OPEN"},
			expectErr: false,
		},
		{
			name:      "with MIXING status",
			args:      []string{},
			flags:     map[string]string{"status": "MIXING"},
			expectErr: false,
		},
		{
			name:      "with COMPLETED status",
			args:      []string{},
			flags:     map[string]string{"status": "COMPLETED"},
			expectErr: false,
		},
		{
			name:      "too many arguments",
			args:      []string{"extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryMixingPools()
			cmd.SetArgs(tc.args)

			for k, v := range tc.flags {
				cmd.Flags().Set(k, v)
			}

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQueryViewKey tests view key query command
func (s *QueryCLITestSuite) TestCmdQueryViewKey() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid public view key",
			args:      []string{"abc123def456"},
			expectErr: false,
		},
		{
			name:      "another valid key",
			args:      []string{"789abcdef012"},
			expectErr: false,
		},
		{
			name:      "invalid public view key hex",
			args:      []string{"invalid-hex"},
			expectErr: true,
			errMsg:    "invalid public-view-key",
		},
		{
			name:      "missing public view key",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryViewKey()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
				if tc.errMsg != "" {
					s.Require().Contains(err.Error(), tc.errMsg)
				}
			}
		})
	}
}

// TestCmdQueryViewKeys tests view keys query command
func (s *QueryCLITestSuite) TestCmdQueryViewKeys() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid address",
			args:      []string{"aura1abc"},
			expectErr: false,
		},
		{
			name:      "another valid address",
			args:      []string{"aura1def"},
			expectErr: false,
		},
		{
			name:      "missing address",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "too many arguments",
			args:      []string{"aura1abc", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryViewKeys()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQueryVerifyZKProof tests ZK proof verification query command
func (s *QueryCLITestSuite) TestCmdQueryVerifyZKProof() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid proof file",
			args:      []string{"proof.json"},
			expectErr: false,
		},
		{
			name:      "another valid proof file",
			args:      []string{"zk_proof_data.json"},
			expectErr: false,
		},
		{
			name:      "missing proof file",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "too many arguments",
			args:      []string{"proof.json", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryVerifyZKProof()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// SECURITY NOTE: TestCmdQueryDecryptWithViewKey has been removed.
// The DecryptWithViewKey query command was removed because decryption must be performed
// client-side using private keys that never leave the client.
//
// Private view keys must NEVER be:
// - Transmitted to the blockchain
// - Passed as CLI arguments
// - Sent in API requests
// - Stored in blockchain state
//
// Decryption functionality should be tested in client-side libraries, not in the CLI.
func (s *QueryCLITestSuite) TestCmdQueryDecryptWithViewKey_RemovedForSecurity() {
	// This test documents that DecryptWithViewKey was intentionally removed
	// for security reasons. Decryption must be client-side only.
	s.T().Log("DecryptWithViewKey command removed: decryption must be client-side only")
}
