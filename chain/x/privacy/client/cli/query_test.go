package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type QueryCLITestSuite struct {
	suite.Suite
}

func TestQueryCLITestSuite(t *testing.T) {
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
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotEmpty(cmd.Long)

	// Verify command expects no arguments
	s.Require().NotNil(cmd.Args)
}

// TestCmdQueryMixingPool tests mixing pool query command
func (s *QueryCLITestSuite) TestCmdQueryMixingPool() {
	cmd := CmdQueryMixingPool()

	s.Require().NotNil(cmd)
	s.Require().Equal("mixing-pool [pool-id]", cmd.Use)
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotEmpty(cmd.Long)

	// Test argument validation
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
			// Test Args validator directly if it exists
			if cmd.Args != nil {
				err := cmd.Args(cmd, tc.args)
				if tc.expectErr {
					s.Require().Error(err)
				} else {
					s.Require().NoError(err)
				}
			}
		})
	}
}

// TestCmdQueryMixingPools tests mixing pools query command
func (s *QueryCLITestSuite) TestCmdQueryMixingPools() {
	cmd := CmdQueryMixingPools()

	s.Require().NotNil(cmd)
	s.Require().Equal("mixing-pools", cmd.Use)
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotEmpty(cmd.Long)

	// Verify status flag exists
	statusFlag := cmd.Flags().Lookup("status")
	s.Require().NotNil(statusFlag)

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
			// Test Args validator directly if it exists
			if cmd.Args != nil {
				err := cmd.Args(cmd, tc.args)
				if tc.expectErr {
					s.Require().Error(err)
				} else {
					s.Require().NoError(err)
				}
			}

			// Test flag parsing
			for k, v := range tc.flags {
				err := cmd.Flags().Set(k, v)
				s.Require().NoError(err)
			}
		})
	}
}

// TestCmdQueryViewKey tests view key query command
func (s *QueryCLITestSuite) TestCmdQueryViewKey() {
	cmd := CmdQueryViewKey()

	s.Require().NotNil(cmd)
	s.Require().Equal("view-key [public-view-key]", cmd.Use)
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotEmpty(cmd.Long)

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid argument count",
			args:      []string{"abc123def456"},
			expectErr: false,
		},
		{
			name:      "another valid argument count",
			args:      []string{"789abcdef012"},
			expectErr: false,
		},
		{
			name:      "missing public view key",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "too many arguments",
			args:      []string{"abc123", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Test Args validator
			if cmd.Args != nil {
				err := cmd.Args(cmd, tc.args)
				if tc.expectErr {
					s.Require().Error(err)
				} else {
					s.Require().NoError(err)
				}
			}
		})
	}
}

// TestCmdQueryViewKeys tests view keys query command
func (s *QueryCLITestSuite) TestCmdQueryViewKeys() {
	cmd := CmdQueryViewKeys()

	s.Require().NotNil(cmd)
	s.Require().Equal("view-keys [address]", cmd.Use)
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotEmpty(cmd.Long)

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
			// Test Args validator
			if cmd.Args != nil {
				err := cmd.Args(cmd, tc.args)
				if tc.expectErr {
					s.Require().Error(err)
				} else {
					s.Require().NoError(err)
				}
			}
		})
	}
}

// TestCmdQueryVerifyZKProof tests ZK proof verification query command
func (s *QueryCLITestSuite) TestCmdQueryVerifyZKProof() {
	cmd := CmdQueryVerifyZKProof()

	s.Require().NotNil(cmd)
	s.Require().Equal("verify-zk-proof [proof-file]", cmd.Use)
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotEmpty(cmd.Long)

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
			// Test Args validator
			if cmd.Args != nil {
				err := cmd.Args(cmd, tc.args)
				if tc.expectErr {
					s.Require().Error(err)
				} else {
					s.Require().NoError(err)
				}
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
