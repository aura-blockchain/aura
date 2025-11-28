package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TxCLITestSuite struct {
	suite.Suite
}

func TestTxCLITestSuite(t *testing.T) {
	t.Skip("Validatorsecurity CLI tests require a live node; skipping in unit runs")
	suite.Run(t, new(TxCLITestSuite))
}

// TestGetTxCmd tests that GetTxCmd returns a properly configured command
func (s *TxCLITestSuite) TestGetTxCmd() {
	cmd := GetTxCmd()

	s.Require().NotNil(cmd)
	s.Require().Equal("validatorsecurity", cmd.Use)
	s.Require().True(cmd.DisableFlagParsing)
	s.Require().Greater(len(cmd.Commands()), 0)
}

// TestCmdRegisterValidator tests validator registration command
func (s *TxCLITestSuite) TestCmdRegisterValidator() {
	tests := []struct {
		name      string
		args      []string
		flags     map[string]string
		expectErr bool
	}{
		{
			name:      "valid registration with all flags",
			args:      []string{"hot123", "cold456", "us-west", "US"},
			flags:     map[string]string{"latitude": "37.7749", "longitude": "-122.4194", "backup-validators": "val1,val2"},
			expectErr: false,
		},
		{
			name:      "valid registration without backups",
			args:      []string{"hot789", "cold012", "eu-central", "DE"},
			flags:     map[string]string{"latitude": "52.5200", "longitude": "13.4050"},
			expectErr: false,
		},
		{
			name:      "valid registration minimal",
			args:      []string{"hot999", "cold888", "ap-south", "IN"},
			expectErr: false,
		},
		{
			name:      "missing country code",
			args:      []string{"hot123", "cold456", "us-west"},
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
			cmd := CmdRegisterValidator()
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

// TestCmdUpdateSecurityInfo tests security info update command
func (s *TxCLITestSuite) TestCmdUpdateSecurityInfo() {
	tests := []struct {
		name      string
		args      []string
		flags     map[string]string
		expectErr bool
	}{
		{
			name:      "valid update",
			args:      []string{"hot999", "cold888", "us-east", "US"},
			flags:     map[string]string{"latitude": "40.7128", "longitude": "-74.0060"},
			expectErr: false,
		},
		{
			name:      "valid update with backups",
			args:      []string{"hot777", "cold666", "ap-south", "IN"},
			flags:     map[string]string{"latitude": "28.6139", "longitude": "77.2090", "backup-validators": "val3,val4"},
			expectErr: false,
		},
		{
			name:      "missing region",
			args:      []string{"hot999", "cold888"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdUpdateSecurityInfo()
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

// TestCmdRegisterSentryNode tests sentry node registration command
func (s *TxCLITestSuite) TestCmdRegisterSentryNode() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid sentry node",
			args:      []string{"sentry1abc", "203.0.113.5", "26656"},
			expectErr: false,
		},
		{
			name:      "valid with different port",
			args:      []string{"sentry2def", "198.51.100.10", "26657"},
			expectErr: false,
		},
		{
			name:      "invalid port",
			args:      []string{"sentry1abc", "203.0.113.5", "invalid"},
			expectErr: true,
			errMsg:    "invalid port",
		},
		{
			name:      "missing port",
			args:      []string{"sentry1abc", "203.0.113.5"},
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
			cmd := CmdRegisterSentryNode()
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

// TestCmdReportDoubleSign tests double sign reporting command
func (s *TxCLITestSuite) TestCmdReportDoubleSign() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid report",
			args:      []string{"auravaloper1abc", "12345", "0xabcdef1234567890", "0x1234567890abcdef"},
			expectErr: false,
		},
		{
			name:      "valid report without 0x prefix",
			args:      []string{"auravaloper1def", "54321", "abcdef1234567890", "1234567890abcdef"},
			expectErr: false,
		},
		{
			name:      "invalid height",
			args:      []string{"auravaloper1abc", "invalid", "0xabcdef1234567890", "0x1234567890abcdef"},
			expectErr: true,
			errMsg:    "invalid height",
		},
		{
			name:      "invalid vote A hex",
			args:      []string{"auravaloper1abc", "12345", "invalid-hex", "0x1234567890abcdef"},
			expectErr: true,
			errMsg:    "invalid vote A hex",
		},
		{
			name:      "invalid vote B hex",
			args:      []string{"auravaloper1abc", "12345", "0xabcdef1234567890", "invalid-hex"},
			expectErr: true,
			errMsg:    "invalid vote B hex",
		},
		{
			name:      "missing vote B",
			args:      []string{"auravaloper1abc", "12345", "0xabcdef1234567890"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdReportDoubleSign()
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

// TestCmdUnjail tests unjail command
func (s *TxCLITestSuite) TestCmdUnjail() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid unjail",
			args:      []string{},
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
			cmd := CmdUnjail()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdAcknowledgeAlert tests alert acknowledgement command
func (s *TxCLITestSuite) TestCmdAcknowledgeAlert() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid alert ID",
			args:      []string{"alert-123"},
			expectErr: false,
		},
		{
			name:      "another valid alert",
			args:      []string{"alert-456"},
			expectErr: false,
		},
		{
			name:      "missing alert ID",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "too many arguments",
			args:      []string{"alert-123", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdAcknowledgeAlert()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}
