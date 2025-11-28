package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TxCLITestSuite struct {
	suite.Suite
}

func TestTxCLITestSuite(t *testing.T) {
	t.Skip("Network security CLI tests require a live node; skipping in unit runs")
	suite.Run(t, new(TxCLITestSuite))
}

// TestGetTxCmd tests that GetTxCmd returns a properly configured command
func (s *TxCLITestSuite) TestGetTxCmd() {
	cmd := GetTxCmd()

	s.Require().NotNil(cmd)
	s.Require().Equal("networksecurity", cmd.Use)
	s.Require().True(cmd.DisableFlagParsing)
	s.Require().Greater(len(cmd.Commands()), 0)
}

// TestCmdAddTrustedPeer tests add trusted peer command
func (s *TxCLITestSuite) TestCmdAddTrustedPeer() {
	tests := []struct {
		name      string
		args      []string
		flags     map[string]string
		expectErr bool
	}{
		{
			name:      "valid peer with description",
			args:      []string{"peer123", "node1.aura.network:26656"},
			flags:     map[string]string{"description": "Core validator node"},
			expectErr: false,
		},
		{
			name:      "valid peer without description",
			args:      []string{"peer456", "203.0.113.5:26656"},
			expectErr: false,
		},
		{
			name:      "missing address",
			args:      []string{"peer123"},
			expectErr: true,
		},
		{
			name:      "no arguments",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdAddTrustedPeer()
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

// TestCmdRemoveTrustedPeer tests remove trusted peer command
func (s *TxCLITestSuite) TestCmdRemoveTrustedPeer() {
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
			cmd := CmdRemoveTrustedPeer()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdBanPeer tests ban peer command
func (s *TxCLITestSuite) TestCmdBanPeer() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid ban with duration",
			args:      []string{"peer789", "86400", "Excessive spam"},
			expectErr: false,
		},
		{
			name:      "valid short duration",
			args:      []string{"peer012", "3600", "DDoS attempt"},
			expectErr: false,
		},
		{
			name:      "invalid duration",
			args:      []string{"peer789", "invalid", "Spam"},
			expectErr: true,
			errMsg:    "invalid duration",
		},
		{
			name:      "missing reason",
			args:      []string{"peer789", "86400"},
			expectErr: true,
		},
		{
			name:      "no arguments",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdBanPeer()
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

// TestCmdUnbanPeer tests unban peer command
func (s *TxCLITestSuite) TestCmdUnbanPeer() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid peer ID",
			args:      []string{"peer789"},
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
			cmd := CmdUnbanPeer()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdUpdatePeerReputation tests update peer reputation command
func (s *TxCLITestSuite) TestCmdUpdatePeerReputation() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid high score",
			args:      []string{"peer123", "95", "Consistent uptime and performance"},
			expectErr: false,
		},
		{
			name:      "valid low score",
			args:      []string{"peer456", "20", "Frequent timeout issues"},
			expectErr: false,
		},
		{
			name:      "invalid score",
			args:      []string{"peer123", "invalid", "Test"},
			expectErr: true,
			errMsg:    "invalid score",
		},
		{
			name:      "missing reason",
			args:      []string{"peer123", "95"},
			expectErr: true,
		},
		{
			name:      "no arguments",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdUpdatePeerReputation()
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

// TestCmdResolveForkAlert tests resolve fork alert command
func (s *TxCLITestSuite) TestCmdResolveForkAlert() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid resolution",
			args:      []string{"alert123", "Fork resolved at height 12345"},
			expectErr: false,
		},
		{
			name:      "valid detailed resolution",
			args:      []string{"alert456", "Fork resolved at height 12345, canonical chain confirmed"},
			expectErr: false,
		},
		{
			name:      "missing resolution details",
			args:      []string{"alert123"},
			expectErr: true,
		},
		{
			name:      "no arguments",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdResolveForkAlert()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdResolvePartitionAlert tests resolve partition alert command
func (s *TxCLITestSuite) TestCmdResolvePartitionAlert() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid alert ID",
			args:      []string{"alert456"},
			expectErr: false,
		},
		{
			name:      "missing alert ID",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "too many arguments",
			args:      []string{"alert456", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdResolvePartitionAlert()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}
