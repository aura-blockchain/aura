package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type QueryCLITestSuite struct {
	suite.Suite
}

func TestQueryCLITestSuite(t *testing.T) {
	t.Skip("CLI query tests require a live node; skipping in unit runs")
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
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid wallet ID",
			args:      []string{"wallet123"},
			expectErr: false,
		},
		{
			name:      "missing wallet ID",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "too many arguments",
			args:      []string{"wallet123", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryHardwareWallet()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQueryMultiSigWallet tests multi-sig wallet query command
func (s *QueryCLITestSuite) TestCmdQueryMultiSigWallet() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid wallet ID",
			args:      []string{"wallet456"},
			expectErr: false,
		},
		{
			name:      "missing wallet ID",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryMultiSigWallet()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQueryPendingMultiSigTx tests pending multi-sig transaction query command
func (s *QueryCLITestSuite) TestCmdQueryPendingMultiSigTx() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid tx ID",
			args:      []string{"tx789"},
			expectErr: false,
		},
		{
			name:      "missing tx ID",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryPendingMultiSigTx()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQuerySocialRecoveryConfig tests social recovery config query command
func (s *QueryCLITestSuite) TestCmdQuerySocialRecoveryConfig() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid wallet ID",
			args:      []string{"wallet123"},
			expectErr: false,
		},
		{
			name:      "missing wallet ID",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQuerySocialRecoveryConfig()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQueryRecoveryRequest tests recovery request query command
func (s *QueryCLITestSuite) TestCmdQueryRecoveryRequest() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid request ID",
			args:      []string{"req123"},
			expectErr: false,
		},
		{
			name:      "missing request ID",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryRecoveryRequest()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQuerySpendingLimit tests spending limit query command
func (s *QueryCLITestSuite) TestCmdQuerySpendingLimit() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid wallet and denom",
			args:      []string{"wallet123", "uaura"},
			expectErr: false,
		},
		{
			name:      "missing denom",
			args:      []string{"wallet123"},
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
			cmd := CmdQuerySpendingLimit()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQuerySessionConfig tests session config query command
func (s *QueryCLITestSuite) TestCmdQuerySessionConfig() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid session ID",
			args:      []string{"session456"},
			expectErr: false,
		},
		{
			name:      "missing session ID",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQuerySessionConfig()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQuerySecurityMetrics tests security metrics query command
func (s *QueryCLITestSuite) TestCmdQuerySecurityMetrics() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid wallet ID",
			args:      []string{"wallet123"},
			expectErr: false,
		},
		{
			name:      "missing wallet ID",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQuerySecurityMetrics()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQueryDomainVerification tests domain verification query command
func (s *QueryCLITestSuite) TestCmdQueryDomainVerification() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid domain",
			args:      []string{"app.aura.network"},
			expectErr: false,
		},
		{
			name:      "missing domain",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryDomainVerification()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQueryDustFilter tests dust filter query command
func (s *QueryCLITestSuite) TestCmdQueryDustFilter() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid wallet ID",
			args:      []string{"wallet123"},
			expectErr: false,
		},
		{
			name:      "missing wallet ID",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryDustFilter()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}
