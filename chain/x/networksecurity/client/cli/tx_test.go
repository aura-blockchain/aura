package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

type TxCLITestSuite struct {
	suite.Suite
}

func TestTxCLITestSuite(t *testing.T) {
	suite.Run(t, new(TxCLITestSuite))
}

// TestGetTxCmd tests that GetTxCmd returns a properly configured command
func (s *TxCLITestSuite) TestGetTxCmd() {
	cmd := GetTxCmd()

	s.Require().NotNil(cmd)
	s.Require().Equal("networksecurity", cmd.Use)
	s.Require().True(cmd.DisableFlagParsing)
	s.Require().Greater(len(cmd.Commands()), 0, "should have subcommands")
}

// TestCmdAddTrustedPeer tests add trusted peer command structure
func (s *TxCLITestSuite) TestCmdAddTrustedPeer() {
	cmd := CmdAddTrustedPeer()

	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "add-trusted-peer")
	s.Require().NotEmpty(cmd.Short, "should have short description")
	s.Require().NotNil(cmd.Args, "should validate arguments")

	// Verify flags exist
	descFlag := cmd.Flags().Lookup("description")
	s.Require().NotNil(descFlag, "description flag should exist")
}

// TestCmdRemoveTrustedPeer tests remove trusted peer command structure
func (s *TxCLITestSuite) TestCmdRemoveTrustedPeer() {
	cmd := CmdRemoveTrustedPeer()

	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "remove-trusted-peer")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.Args)
}

// TestCmdBanPeer tests ban peer command structure
func (s *TxCLITestSuite) TestCmdBanPeer() {
	cmd := CmdBanPeer()

	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "ban-peer")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.Args)
}

// TestCmdUnbanPeer tests unban peer command structure
func (s *TxCLITestSuite) TestCmdUnbanPeer() {
	cmd := CmdUnbanPeer()

	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "unban-peer")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.Args)
}

// TestCmdUpdatePeerReputation tests update peer reputation command structure
func (s *TxCLITestSuite) TestCmdUpdatePeerReputation() {
	cmd := CmdUpdatePeerReputation()

	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "update-peer-reputation")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.Args)
}

// TestCmdResolveForkAlert tests resolve fork alert command structure
func (s *TxCLITestSuite) TestCmdResolveForkAlert() {
	cmd := CmdResolveForkAlert()

	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "resolve-fork-alert")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.Args)
}

// TestCmdResolvePartitionAlert tests resolve partition alert command structure
func (s *TxCLITestSuite) TestCmdResolvePartitionAlert() {
	cmd := CmdResolvePartitionAlert()

	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "resolve-partition-alert")
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotNil(cmd.Args)
}

// TestAllCommandsHaveRunE tests that all commands have RunE handlers
func (s *TxCLITestSuite) TestAllCommandsHaveRunE() {
	commands := []struct {
		name string
		cmd  func() *cobra.Command
	}{
		{"add-trusted-peer", CmdAddTrustedPeer},
		{"remove-trusted-peer", CmdRemoveTrustedPeer},
		{"ban-peer", CmdBanPeer},
		{"unban-peer", CmdUnbanPeer},
		{"update-peer-reputation", CmdUpdatePeerReputation},
		{"resolve-fork-alert", CmdResolveForkAlert},
		{"resolve-partition-alert", CmdResolvePartitionAlert},
	}

	for _, tc := range commands {
		s.Run(tc.name, func() {
			cmd := tc.cmd()
			s.Require().NotNil(cmd.RunE, "command %s should have RunE handler", tc.name)
		})
	}
}
