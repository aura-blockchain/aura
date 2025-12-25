// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/spf13/cobra"
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
	s.Require().Equal("networksecurity", cmd.Use)
	s.Require().True(cmd.DisableFlagParsing)
	s.Require().Greater(len(cmd.Commands()), 0, "should have subcommands")
}

// TestCmdQueryParams tests params query command structure
func (s *QueryCLITestSuite) TestCmdQueryParams() {
	cmd := CmdQueryParams()

	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "params")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.RunE, "should have RunE handler")
}

// TestCmdQueryPeerInfo tests peer info query command structure
func (s *QueryCLITestSuite) TestCmdQueryPeerInfo() {
	cmd := CmdQueryPeerInfo()

	s.Require().NotNil(cmd)
	s.Require().NotEmpty(cmd.Use, "command should have Use field")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.Args, "should validate arguments")
	s.Require().NotNil(cmd.RunE)
}

// TestCmdQueryAllPeers tests all peers query command structure
func (s *QueryCLITestSuite) TestCmdQueryAllPeers() {
	cmd := CmdQueryAllPeers()

	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "peers")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.RunE)
}

// TestCmdQueryTrustedPeers tests trusted peers query command structure
func (s *QueryCLITestSuite) TestCmdQueryTrustedPeers() {
	cmd := CmdQueryTrustedPeers()

	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "trusted-peers")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.RunE)
}

// TestCmdQueryPeerReputation tests peer reputation query command structure
func (s *QueryCLITestSuite) TestCmdQueryPeerReputation() {
	cmd := CmdQueryPeerReputation()

	s.Require().NotNil(cmd)
	s.Require().NotEmpty(cmd.Use, "command should have Use field")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.Args)
	s.Require().NotNil(cmd.RunE)
}

// TestCmdQueryRateLimitStatus tests rate limit status query command structure
func (s *QueryCLITestSuite) TestCmdQueryRateLimitStatus() {
	cmd := CmdQueryRateLimitStatus()

	s.Require().NotNil(cmd)
	s.Require().NotEmpty(cmd.Use, "command should have Use field")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.Args)
	s.Require().NotNil(cmd.RunE)
}

// TestCmdQueryMempoolStats tests mempool stats query command structure
func (s *QueryCLITestSuite) TestCmdQueryMempoolStats() {
	cmd := CmdQueryMempoolStats()

	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "mempool-stats")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.RunE)
}

// TestCmdQueryForkAlerts tests fork alerts query command structure
func (s *QueryCLITestSuite) TestCmdQueryForkAlerts() {
	cmd := CmdQueryForkAlerts()

	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "fork-alerts")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.RunE)

	// Verify include-resolved flag exists
	flag := cmd.Flags().Lookup("include-resolved")
	s.Require().NotNil(flag, "should have include-resolved flag")
}

// TestCmdQueryPartitionAlerts tests partition alerts query command structure
func (s *QueryCLITestSuite) TestCmdQueryPartitionAlerts() {
	cmd := CmdQueryPartitionAlerts()

	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "partition-alerts")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.RunE)

	// Verify include-resolved flag exists
	flag := cmd.Flags().Lookup("include-resolved")
	s.Require().NotNil(flag, "should have include-resolved flag")
}

// TestCmdQueryNetworkHealth tests network health query command structure
func (s *QueryCLITestSuite) TestCmdQueryNetworkHealth() {
	cmd := CmdQueryNetworkHealth()

	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "health")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.RunE)
}

// TestAllQueryCommandsHaveRunE tests that all query commands have RunE handlers
func (s *QueryCLITestSuite) TestAllQueryCommandsHaveRunE() {
	commands := []struct {
		name string
		cmd  func() *cobra.Command
	}{
		{"params", CmdQueryParams},
		{"peer-info", CmdQueryPeerInfo},
		{"peers", CmdQueryAllPeers},
		{"trusted-peers", CmdQueryTrustedPeers},
		{"peer-reputation", CmdQueryPeerReputation},
		{"rate-limit-status", CmdQueryRateLimitStatus},
		{"mempool-stats", CmdQueryMempoolStats},
		{"fork-alerts", CmdQueryForkAlerts},
		{"partition-alerts", CmdQueryPartitionAlerts},
		{"health", CmdQueryNetworkHealth},
	}

	for _, tc := range commands {
		s.Run(tc.name, func() {
			cmd := tc.cmd()
			s.Require().NotNil(cmd.RunE, "command %s should have RunE handler", tc.name)
		})
	}
}
