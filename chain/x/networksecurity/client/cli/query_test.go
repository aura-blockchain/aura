package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type QueryCLITestSuite struct {
	suite.Suite
}

func TestQueryCLITestSuite(t *testing.T) {
	t.Skip("Network security CLI tests require a live node; skipping in unit runs")
	suite.Run(t, new(QueryCLITestSuite))
}

// TestGetQueryCmd tests that GetQueryCmd returns a properly configured command
func (s *QueryCLITestSuite) TestGetQueryCmd() {
	cmd := GetQueryCmd()

	s.Require().NotNil(cmd)
	s.Require().Equal("networksecurity", cmd.Use)
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
	// Will fail without proper context, but validates structure
	s.Require().Error(err)
}

// TestCmdQueryPeerInfo tests peer info query command
func (s *QueryCLITestSuite) TestCmdQueryPeerInfo() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid peer ID",
			args:      []string{"peer123"},
			expectErr: false,
		},
		{
			name:      "missing peer ID",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "too many arguments",
			args:      []string{"peer123", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryPeerInfo()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQueryAllPeers tests all peers query command
func (s *QueryCLITestSuite) TestCmdQueryAllPeers() {
	cmd := CmdQueryAllPeers()

	s.Require().NotNil(cmd)
	s.Require().Equal("peers", cmd.Use)

	// Test with no arguments
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	s.Require().Error(err) // Will fail without context
}

// TestCmdQueryTrustedPeers tests trusted peers query command
func (s *QueryCLITestSuite) TestCmdQueryTrustedPeers() {
	cmd := CmdQueryTrustedPeers()

	s.Require().NotNil(cmd)
	s.Require().Equal("trusted-peers", cmd.Use)

	// Test with no arguments
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	s.Require().Error(err) // Will fail without context
}

// TestCmdQueryPeerReputation tests peer reputation query command
func (s *QueryCLITestSuite) TestCmdQueryPeerReputation() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid peer ID",
			args:      []string{"peer123"},
			expectErr: false,
		},
		{
			name:      "missing peer ID",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryPeerReputation()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQueryRateLimitStatus tests rate limit status query command
func (s *QueryCLITestSuite) TestCmdQueryRateLimitStatus() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid peer ID",
			args:      []string{"peer123"},
			expectErr: false,
		},
		{
			name:      "missing peer ID",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryRateLimitStatus()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQueryMempoolStats tests mempool stats query command
func (s *QueryCLITestSuite) TestCmdQueryMempoolStats() {
	cmd := CmdQueryMempoolStats()

	s.Require().NotNil(cmd)
	s.Require().Equal("mempool-stats", cmd.Use)

	// Test with no arguments
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	s.Require().Error(err) // Will fail without context
}

// TestCmdQueryForkAlerts tests fork alerts query command
func (s *QueryCLITestSuite) TestCmdQueryForkAlerts() {
	tests := []struct {
		name      string
		args      []string
		flags     map[string]string
		expectErr bool
	}{
		{
			name:      "without include-resolved flag",
			args:      []string{},
			expectErr: false,
		},
		{
			name:      "with include-resolved flag",
			args:      []string{},
			flags:     map[string]string{"include-resolved": "true"},
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
			cmd := CmdQueryForkAlerts()
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

// TestCmdQueryPartitionAlerts tests partition alerts query command
func (s *QueryCLITestSuite) TestCmdQueryPartitionAlerts() {
	tests := []struct {
		name      string
		args      []string
		flags     map[string]string
		expectErr bool
	}{
		{
			name:      "without include-resolved flag",
			args:      []string{},
			expectErr: false,
		},
		{
			name:      "with include-resolved flag",
			args:      []string{},
			flags:     map[string]string{"include-resolved": "true"},
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
			cmd := CmdQueryPartitionAlerts()
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

// TestCmdQueryNetworkHealth tests network health query command
func (s *QueryCLITestSuite) TestCmdQueryNetworkHealth() {
	cmd := CmdQueryNetworkHealth()

	s.Require().NotNil(cmd)
	s.Require().Equal("health", cmd.Use)

	// Test with no arguments
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	s.Require().Error(err) // Will fail without context
}
