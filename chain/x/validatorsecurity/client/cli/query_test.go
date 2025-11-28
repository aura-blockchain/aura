package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type QueryCLITestSuite struct {
	suite.Suite
}

func TestQueryCLITestSuite(t *testing.T) {
	t.Skip("Validatorsecurity CLI tests require a live node; skipping in unit runs")
	suite.Run(t, new(QueryCLITestSuite))
}

// TestGetQueryCmd tests that GetQueryCmd returns a properly configured command
func (s *QueryCLITestSuite) TestGetQueryCmd() {
	cmd := GetQueryCmd()

	s.Require().NotNil(cmd)
	s.Require().Equal("validatorsecurity", cmd.Use)
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

// TestCmdQueryValidatorSecurityInfo tests validator security info query command
func (s *QueryCLITestSuite) TestCmdQueryValidatorSecurityInfo() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid validator address",
			args:      []string{"auravaloper1abc"},
			expectErr: false,
		},
		{
			name:      "another valid address",
			args:      []string{"auravaloper1def"},
			expectErr: false,
		},
		{
			name:      "missing validator address",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "too many arguments",
			args:      []string{"auravaloper1abc", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryValidatorSecurityInfo()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQueryAllValidators tests all validators query command
func (s *QueryCLITestSuite) TestCmdQueryAllValidators() {
	cmd := CmdQueryAllValidators()

	s.Require().NotNil(cmd)
	s.Require().Equal("validators", cmd.Use)

	// Test with no arguments
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	s.Require().Error(err) // Will fail without context
}

// TestCmdQueryJailedValidators tests jailed validators query command
func (s *QueryCLITestSuite) TestCmdQueryJailedValidators() {
	cmd := CmdQueryJailedValidators()

	s.Require().NotNil(cmd)
	s.Require().Equal("jailed", cmd.Use)

	// Test with no arguments
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	s.Require().Error(err) // Will fail without context
}

// TestCmdQueryTombstonedValidators tests tombstoned validators query command
func (s *QueryCLITestSuite) TestCmdQueryTombstonedValidators() {
	cmd := CmdQueryTombstonedValidators()

	s.Require().NotNil(cmd)
	s.Require().Equal("tombstoned", cmd.Use)

	// Test with no arguments
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	s.Require().Error(err) // Will fail without context
}

// TestCmdQueryDoubleSignEvidences tests double sign evidences query command
func (s *QueryCLITestSuite) TestCmdQueryDoubleSignEvidences() {
	cmd := CmdQueryDoubleSignEvidences()

	s.Require().NotNil(cmd)
	s.Require().Equal("evidences", cmd.Use)

	// Test with no arguments
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	s.Require().Error(err) // Will fail without context
}

// TestCmdQueryValidatorAlerts tests validator alerts query command
func (s *QueryCLITestSuite) TestCmdQueryValidatorAlerts() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid validator address",
			args:      []string{"auravaloper1abc"},
			expectErr: false,
		},
		{
			name:      "another valid address",
			args:      []string{"auravaloper1def"},
			expectErr: false,
		},
		{
			name:      "missing validator address",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQueryValidatorAlerts()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdQuerySentryNodes tests sentry nodes query command
func (s *QueryCLITestSuite) TestCmdQuerySentryNodes() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid validator address",
			args:      []string{"auravaloper1abc"},
			expectErr: false,
		},
		{
			name:      "another valid address",
			args:      []string{"auravaloper1def"},
			expectErr: false,
		},
		{
			name:      "missing validator address",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "too many arguments",
			args:      []string{"auravaloper1abc", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdQuerySentryNodes()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}
