// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

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
	s.Require().Equal("walletsecurity", cmd.Use)
	s.Require().True(cmd.DisableFlagParsing)
	s.Require().Greater(len(cmd.Commands()), 0)
}

// TestCmdQueryHardwareWallet tests hardware wallet query command
func (s *QueryCLITestSuite) TestCmdQueryHardwareWallet() {
	cmd := CmdQueryHardwareWallet()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "hw-wallet")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 1 arg
	err := cmd.Args(cmd, []string{"wallet123"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{})
	s.Require().Error(err, "should error with missing wallet ID")

	err = cmd.Args(cmd, []string{"wallet123", "extra"})
	s.Require().Error(err, "should error with too many args")
}

// TestCmdQueryMultiSigWallet tests multi-sig wallet query command
func (s *QueryCLITestSuite) TestCmdQueryMultiSigWallet() {
	cmd := CmdQueryMultiSigWallet()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "multisig")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 1 arg
	err := cmd.Args(cmd, []string{"wallet456"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{})
	s.Require().Error(err, "should error with missing wallet ID")
}

// TestCmdQueryPendingMultiSigTx tests pending multi-sig transaction query command
func (s *QueryCLITestSuite) TestCmdQueryPendingMultiSigTx() {
	cmd := CmdQueryPendingMultiSigTx()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "pending-multisig-tx")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 1 arg
	err := cmd.Args(cmd, []string{"tx789"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{})
	s.Require().Error(err, "should error with missing tx ID")
}

// TestCmdQuerySocialRecoveryConfig tests social recovery config query command
func (s *QueryCLITestSuite) TestCmdQuerySocialRecoveryConfig() {
	cmd := CmdQuerySocialRecoveryConfig()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "social-recovery")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 1 arg
	err := cmd.Args(cmd, []string{"wallet123"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{})
	s.Require().Error(err, "should error with missing wallet ID")
}

// TestCmdQueryRecoveryRequest tests recovery request query command
func (s *QueryCLITestSuite) TestCmdQueryRecoveryRequest() {
	cmd := CmdQueryRecoveryRequest()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "recovery-request")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 1 arg
	err := cmd.Args(cmd, []string{"req123"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{})
	s.Require().Error(err, "should error with missing request ID")
}

// TestCmdQuerySpendingLimit tests spending limit query command
func (s *QueryCLITestSuite) TestCmdQuerySpendingLimit() {
	cmd := CmdQuerySpendingLimit()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "spending-limit")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 2 args
	err := cmd.Args(cmd, []string{"wallet123", "uaura"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{"wallet123"})
	s.Require().Error(err, "should error with missing denom")

	err = cmd.Args(cmd, []string{})
	s.Require().Error(err, "should error with no arguments")
}

// TestCmdQuerySessionConfig tests session config query command
func (s *QueryCLITestSuite) TestCmdQuerySessionConfig() {
	cmd := CmdQuerySessionConfig()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "session")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 1 arg
	err := cmd.Args(cmd, []string{"session456"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{})
	s.Require().Error(err, "should error with missing session ID")
}

// TestCmdQuerySecurityMetrics tests security metrics query command
func (s *QueryCLITestSuite) TestCmdQuerySecurityMetrics() {
	cmd := CmdQuerySecurityMetrics()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "security-metrics")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 1 arg
	err := cmd.Args(cmd, []string{"wallet123"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{})
	s.Require().Error(err, "should error with missing wallet ID")
}

// TestCmdQueryDomainVerification tests domain verification query command
func (s *QueryCLITestSuite) TestCmdQueryDomainVerification() {
	cmd := CmdQueryDomainVerification()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "domain-verification")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 1 arg
	err := cmd.Args(cmd, []string{"app.aura.network"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{})
	s.Require().Error(err, "should error with missing domain")
}

// TestCmdQueryDustFilter tests dust filter query command
func (s *QueryCLITestSuite) TestCmdQueryDustFilter() {
	cmd := CmdQueryDustFilter()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "dust-filter")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 1 arg
	err := cmd.Args(cmd, []string{"wallet123"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{})
	s.Require().Error(err, "should error with missing wallet ID")
}

// TestAllQueryCommandsExist ensures all expected query commands are registered
func (s *QueryCLITestSuite) TestAllQueryCommandsExist() {
	cmd := GetQueryCmd()
	subcommands := cmd.Commands()

	// Verify we have the expected number of commands
	s.Require().GreaterOrEqual(len(subcommands), 10, "should have at least 10 query subcommands")

	// Verify specific commands exist by checking their names
	commandNames := make(map[string]bool)
	for _, subcmd := range subcommands {
		commandNames[subcmd.Name()] = true
	}

	expectedCommands := []string{
		"hw-wallet",
		"multisig",
		"pending-multisig-tx",
		"social-recovery",
		"recovery-request",
		"spending-limit",
		"session",
		"security-metrics",
		"domain-verification",
		"dust-filter",
	}

	for _, expected := range expectedCommands {
		s.Require().True(commandNames[expected], "query command %s should exist", expected)
	}
}
