package cli_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/aequitas/aura/chain/x/auth/client/cli"
)

type QueryTestSuite struct {
	suite.Suite
	clientCtx client.Context
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}

func (s *QueryTestSuite) SetupTest() {
	// For command structure tests, we don't need a full codec setup
	s.clientCtx = client.Context{}
}

// TestGetQueryCmd tests that all query commands are registered
func (s *QueryTestSuite) TestGetQueryCmd() {
	cmd := cli.GetQueryCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("auth", cmd.Use)
	require.True(cmd.DisableFlagParsing)

	// Verify all query subcommands are registered
	expectedCmds := []string{
		"role",
		"roles",
		"role-assignments",
		"has-permission",
		"multisig-wallet",
		"multisig-wallets",
		"multisig-proposal",
		"multisig-proposals",
		"timelocked-action",
		"timelocked-actions",
		"emergency-admin",
		"emergency-admins",
		"validator-key-rotation",
		"session",
		"sessions",
		"rate-limit-status",
		"audit-logs",
		"params",
	}

	subCmds := cmd.Commands()
	require.Len(subCmds, len(expectedCmds), "Expected %d query subcommands", len(expectedCmds))

	for _, expectedCmd := range expectedCmds {
		found := false
		for _, subCmd := range subCmds {
			if subCmd.Use == expectedCmd+" [args...]" || (len(subCmd.Use) >= len(expectedCmd) && subCmd.Use[:len(expectedCmd)] == expectedCmd) {
				found = true
				break
			}
		}
		require.True(found, "Expected query command not found: %s", expectedCmd)
	}
}

// TestCmdQueryRole tests the role query command
func (s *QueryTestSuite) TestCmdQueryRole() {
	cmd := cli.GetCmdQueryRole()
	s.Require().NotNil(cmd)
	s.Require().Equal("role [name]", cmd.Use)

	// Test args validation
	err := cmd.ValidateArgs([]string{"admin"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{})
	s.Require().Error(err)

	err = cmd.ValidateArgs([]string{"admin", "extra"})
	s.Require().Error(err)
}

// TestCmdQueryListRoles tests the list roles query command
func (s *QueryTestSuite) TestCmdQueryListRoles() {
	cmd := cli.GetCmdQueryListRoles()
	s.Require().NotNil(cmd)
	s.Require().Equal("roles", cmd.Use)

	// Should accept no args
	err := cmd.ValidateArgs([]string{})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{"extra"})
	s.Require().Error(err)
}

// TestCmdQueryRoleAssignments tests role assignments query
func (s *QueryTestSuite) TestCmdQueryRoleAssignments() {
	cmd := cli.GetCmdQueryRoleAssignments()
	s.Require().NotNil(cmd)
	s.Require().Equal("role-assignments [address]", cmd.Use)

	err := cmd.ValidateArgs([]string{"aura1abc"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{})
	s.Require().Error(err)
}

// TestCmdQueryHasPermission tests has-permission query
func (s *QueryTestSuite) TestCmdQueryHasPermission() {
	cmd := cli.GetCmdQueryHasPermission()
	s.Require().NotNil(cmd)
	s.Require().Equal("has-permission [address] [permission]", cmd.Use)

	err := cmd.ValidateArgs([]string{"aura1abc", "CREATE_ROLE"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{"aura1abc"})
	s.Require().Error(err)
}

// TestCmdQueryMultisigWallet tests multisig wallet query
func (s *QueryTestSuite) TestCmdQueryMultisigWallet() {
	cmd := cli.GetCmdQueryMultisigWallet()
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "multisig-wallet")

	err := cmd.ValidateArgs([]string{"wallet123"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{})
	s.Require().Error(err)
}

// TestCmdQueryListMultisigWallets tests list multisig wallets query
func (s *QueryTestSuite) TestCmdQueryListMultisigWallets() {
	cmd := cli.GetCmdQueryListMultisigWallets()
	s.Require().NotNil(cmd)
	s.Require().Equal("multisig-wallets", cmd.Use)

	err := cmd.ValidateArgs([]string{})
	s.Require().NoError(err)
}

// TestCmdQueryMultisigProposal tests multisig proposal query
func (s *QueryTestSuite) TestCmdQueryMultisigProposal() {
	cmd := cli.GetCmdQueryMultisigProposal()
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "multisig-proposal")

	err := cmd.ValidateArgs([]string{"proposal123"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{})
	s.Require().Error(err)
}

// TestCmdQueryListMultisigProposals tests list multisig proposals query
func (s *QueryTestSuite) TestCmdQueryListMultisigProposals() {
	cmd := cli.GetCmdQueryListMultisigProposals()
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "multisig-proposals")

	err := cmd.ValidateArgs([]string{"wallet123"})
	s.Require().NoError(err)

	// Test status flag
	flag := cmd.Flags().Lookup("status")
	s.Require().NotNil(flag)
}

// TestCmdQueryTimeLockedAction tests time-locked action query
func (s *QueryTestSuite) TestCmdQueryTimeLockedAction() {
	cmd := cli.GetCmdQueryTimeLockedAction()
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "timelocked-action")

	err := cmd.ValidateArgs([]string{"action123"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{})
	s.Require().Error(err)
}

// TestCmdQueryListTimeLockedActions tests list time-locked actions query
func (s *QueryTestSuite) TestCmdQueryListTimeLockedActions() {
	cmd := cli.GetCmdQueryListTimeLockedActions()
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "timelocked-actions")

	err := cmd.ValidateArgs([]string{})
	s.Require().NoError(err)

	// Test status flag
	flag := cmd.Flags().Lookup("status")
	s.Require().NotNil(flag)
}

// TestCmdQueryEmergencyAdmin tests emergency admin query
func (s *QueryTestSuite) TestCmdQueryEmergencyAdmin() {
	cmd := cli.GetCmdQueryEmergencyAdmin()
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "emergency-admin")

	err := cmd.ValidateArgs([]string{"aura1admin"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{})
	s.Require().Error(err)
}

// TestCmdQueryListEmergencyAdmins tests list emergency admins query
func (s *QueryTestSuite) TestCmdQueryListEmergencyAdmins() {
	cmd := cli.GetCmdQueryListEmergencyAdmins()
	s.Require().NotNil(cmd)
	s.Require().Equal("emergency-admins", cmd.Use)

	err := cmd.ValidateArgs([]string{})
	s.Require().NoError(err)
}

// TestCmdQueryValidatorKeyRotation tests validator key rotation query
func (s *QueryTestSuite) TestCmdQueryValidatorKeyRotation() {
	cmd := cli.GetCmdQueryValidatorKeyRotation()
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "validator-key-rotation")

	err := cmd.ValidateArgs([]string{"auravaloper1abc"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{})
	s.Require().Error(err)
}

// TestCmdQuerySession tests session query
func (s *QueryTestSuite) TestCmdQuerySession() {
	cmd := cli.GetCmdQuerySession()
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "session")

	err := cmd.ValidateArgs([]string{"session123"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{})
	s.Require().Error(err)
}

// TestCmdQueryListSessions tests list sessions query
func (s *QueryTestSuite) TestCmdQueryListSessions() {
	cmd := cli.GetCmdQueryListSessions()
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "sessions")

	err := cmd.ValidateArgs([]string{"aura1user"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{})
	s.Require().Error(err)
}

// TestCmdQueryRateLimitStatus tests rate limit status query
func (s *QueryTestSuite) TestCmdQueryRateLimitStatus() {
	cmd := cli.GetCmdQueryRateLimitStatus()
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "rate-limit-status")

	err := cmd.ValidateArgs([]string{"aura1user"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{})
	s.Require().Error(err)
}

// TestCmdQueryAuditLogs tests audit logs query
func (s *QueryTestSuite) TestCmdQueryAuditLogs() {
	cmd := cli.GetCmdQueryAuditLogs()
	s.Require().NotNil(cmd)
	s.Require().Equal("audit-logs", cmd.Use)

	err := cmd.ValidateArgs([]string{})
	s.Require().NoError(err)

	// Test all optional flags exist
	flags := []string{"actor", "action", "start-time", "end-time", "limit"}
	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		s.Require().NotNil(flag, "Expected flag %s not found", flagName)
	}
}

// TestCmdQueryParams tests params query
func (s *QueryTestSuite) TestCmdQueryParams() {
	cmd := cli.GetCmdQueryParams()
	s.Require().NotNil(cmd)
	s.Require().Equal("params", cmd.Use)

	err := cmd.ValidateArgs([]string{})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{"extra"})
	s.Require().Error(err)
}

// TestAllQueryCommandsHaveHelp verifies all query commands have help text
func (s *QueryTestSuite) TestAllQueryCommandsHaveHelp() {
	cmd := cli.GetQueryCmd()
	for _, subCmd := range cmd.Commands() {
		s.Require().NotEmpty(subCmd.Short, "Query command %s missing short description", subCmd.Use)
		s.Require().NotEmpty(subCmd.Long, "Query command %s missing long description", subCmd.Use)
	}
}

// TestAllQueryCommandsHaveExamples verifies all query commands have examples
func (s *QueryTestSuite) TestAllQueryCommandsHaveExamples() {
	cmd := cli.GetQueryCmd()
	for _, subCmd := range cmd.Commands() {
		s.Require().Contains(subCmd.Long, "Example", "Query command %s missing examples", subCmd.Use)
	}
}

// TestParseProposalStatus tests the parseProposalStatus helper function
func TestParseProposalStatus(t *testing.T) {
	// Test through GetCmdQueryListMultisigProposals which uses parseProposalStatus
	cmd := cli.GetCmdQueryListMultisigProposals()
	require.NotNil(t, cmd)

	// Verify status flag exists and has valid values in description
	statusFlag := cmd.Flags().Lookup("status")
	require.NotNil(t, statusFlag)
	require.Contains(t, statusFlag.Usage, "pending")
	require.Contains(t, statusFlag.Usage, "approved")
	require.Contains(t, statusFlag.Usage, "executed")
}

// TestParseActionStatus tests the parseActionStatus helper function
func TestParseActionStatus(t *testing.T) {
	// Test through GetCmdQueryListTimeLockedActions which uses parseActionStatus
	cmd := cli.GetCmdQueryListTimeLockedActions()
	require.NotNil(t, cmd)

	// Verify status flag exists and has valid values in description
	statusFlag := cmd.Flags().Lookup("status")
	require.NotNil(t, statusFlag)
	require.Contains(t, statusFlag.Usage, "pending")
	require.Contains(t, statusFlag.Usage, "ready")
	require.Contains(t, statusFlag.Usage, "executed")
}

// TestQueryFlagsConsistency tests that all query commands have query flags
func (s *QueryTestSuite) TestQueryFlagsConsistency() {
	cmd := cli.GetQueryCmd()
	for _, subCmd := range cmd.Commands() {
		// All query commands should have the standard query flags
		heightFlag := subCmd.Flags().Lookup(flags.FlagHeight)
		s.Require().NotNil(heightFlag, "Command %s missing height flag", subCmd.Use)

		outputFlag := subCmd.Flags().Lookup(flags.FlagOutput)
		s.Require().NotNil(outputFlag, "Command %s missing output flag", subCmd.Use)
	}
}

// Table-driven test for argument validation
func (s *QueryTestSuite) TestQueryCommandArgsValidation() {
	tests := []struct {
		name      string
		cmdFunc   func() *cobra.Command
		validArgs []string
		invalid   []string
	}{
		{
			name:      "role",
			cmdFunc:   cli.GetCmdQueryRole,
			validArgs: []string{"admin"},
			invalid:   []string{},
		},
		{
			name:      "has-permission",
			cmdFunc:   cli.GetCmdQueryHasPermission,
			validArgs: []string{"aura1abc", "CREATE_ROLE"},
			invalid:   []string{"aura1abc"},
		},
		{
			name:      "multisig-wallet",
			cmdFunc:   cli.GetCmdQueryMultisigWallet,
			validArgs: []string{"wallet123"},
			invalid:   []string{},
		},
		{
			name:      "session",
			cmdFunc:   cli.GetCmdQuerySession,
			validArgs: []string{"session123"},
			invalid:   []string{},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := tc.cmdFunc()

			// Valid args should not error
			err := cmd.ValidateArgs(tc.validArgs)
			s.Require().NoError(err, "Valid args failed for %s", tc.name)

			// Invalid args should error
			if len(tc.invalid) > 0 {
				err = cmd.ValidateArgs(tc.invalid)
				s.Require().Error(err, "Invalid args should fail for %s", tc.name)
			}
		})
	}
}
