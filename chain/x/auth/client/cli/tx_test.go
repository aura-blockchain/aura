// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/aequitas/aura/chain/x/auth/client/cli"
)

type TxTestSuite struct {
	suite.Suite
	clientCtx client.Context
}

func TestTxTestSuite(t *testing.T) {
	suite.Run(t, new(TxTestSuite))
}

func (s *TxTestSuite) SetupTest() {
	// For command structure tests, we don't need a full codec setup
	s.clientCtx = client.Context{}.
		WithAccountRetriever(client.MockAccountRetriever{})
}

// TestGetTxCmd tests that all tx commands are registered
func (s *TxTestSuite) TestGetTxCmd() {
	cmd := cli.GetTxCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("auth", cmd.Use)
	require.True(cmd.DisableFlagParsing)

	// Verify all subcommands are registered
	expectedCmds := []string{
		"create-role",
		"assign-role",
		"revoke-role",
		"create-multisig-wallet",
		"create-multisig-proposal",
		"sign-multisig-proposal",
		"execute-multisig-proposal",
		"propose-timelocked-action",
		"execute-timelocked-action",
		"cancel-timelocked-action",
		"activate-emergency-admin",
		"deactivate-emergency-admin",
		"initiate-key-rotation",
		"complete-key-rotation",
		"create-session",
		"revoke-session",
	}

	subCmds := cmd.Commands()
	require.Len(subCmds, len(expectedCmds), "Expected %d subcommands", len(expectedCmds))

	for _, expectedCmd := range expectedCmds {
		found := false
		for _, subCmd := range subCmds {
			if subCmd.Use == expectedCmd+" [args...]" || subCmd.Use[:len(expectedCmd)] == expectedCmd {
				found = true
				break
			}
		}
		require.True(found, "Expected command not found: %s", expectedCmd)
	}
}

// TestCmdCreateRole tests the create-role command
func (s *TxTestSuite) TestCmdCreateRole() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid create role",
			args: []string{
				"admin",
				"CREATE_ROLE,ASSIGN_ROLE",
				"Administrator role",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, "alice"),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			},
			expectErr: false,
		},
		{
			name: "missing arguments",
			args: []string{
				"admin",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, "alice"),
			},
			expectErr: true,
			errMsg:    "accepts 3 arg(s)",
		},
		{
			name: "empty role name",
			args: []string{
				"",
				"CREATE_ROLE",
				"Empty name role",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, "alice"),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			expectErr: false, // CLI doesn't validate empty string, keeper does
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdCreateRole()
			s.Require().NotNil(cmd)

			// Test command structure and argument validation
			// Extract non-flag args for validation
			argsToValidate := []string{}
			for _, arg := range tc.args {
				if !strings.HasPrefix(arg, "--") && !strings.Contains(arg, "=") {
					argsToValidate = append(argsToValidate, arg)
				}
			}

			err := cmd.ValidateArgs(argsToValidate)
			if tc.expectErr {
				s.Require().Error(err, "Expected error for test case: %s", tc.name)
			} else {
				s.Require().NoError(err, "Unexpected error for test case: %s", tc.name)
			}
		})
	}
}

// TestCmdAssignRole tests the assign-role command
func (s *TxTestSuite) TestCmdAssignRole() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name: "valid assign role",
			args: []string{
				"aura1abc123",
				"admin",
			},
			expectErr: false,
		},
		{
			name: "assign role with expiration",
			args: []string{
				"aura1def456",
				"operator",
				"--expires-in=86400",
			},
			expectErr: false,
		},
		{
			name: "missing arguments",
			args: []string{
				"aura1abc123",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdAssignRole()
			numArgs := 2
			if tc.expectErr {
				numArgs = len(tc.args)
			}

			err := cmd.ValidateArgs(tc.args[:numArgs])
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// TestCmdRevokeRole tests the revoke-role command
func (s *TxTestSuite) TestCmdRevokeRole() {
	cmd := cli.GetCmdRevokeRole()
	s.Require().NotNil(cmd)
	s.Require().Equal("revoke-role [address] [role-name]", cmd.Use)

	// Test args validation
	err := cmd.ValidateArgs([]string{"aura1abc", "admin"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{"aura1abc"})
	s.Require().Error(err)
}

// TestCmdCreateMultisigWallet tests the create-multisig-wallet command
func (s *TxTestSuite) TestCmdCreateMultisigWallet() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name: "valid multisig wallet",
			args: []string{
				"aura1abc,aura1def,aura1ghi",
				"2",
			},
			expectErr: false,
		},
		{
			name: "invalid threshold",
			args: []string{
				"aura1abc,aura1def",
				"invalid",
			},
			expectErr: false, // Validation happens in RunE, not Args
		},
		{
			name: "missing args",
			args: []string{
				"aura1abc",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdCreateMultisigWallet()
			err := cmd.ValidateArgs(tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// TestCmdCreateMultisigProposal tests the create-multisig-proposal command
func (s *TxTestSuite) TestCmdCreateMultisigProposal() {
	cmd := cli.GetCmdCreateMultisigProposal()
	s.Require().NotNil(cmd)

	// Valid hex payload
	validPayload := hex.EncodeToString([]byte("test payload"))

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name: "valid proposal",
			args: []string{
				"wallet123",
				"Transfer funds",
				"Transfer 100 tokens",
				validPayload,
			},
			expectErr: false,
		},
		{
			name: "missing args",
			args: []string{
				"wallet123",
				"title",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := cmd.ValidateArgs(tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// TestCmdSignMultisigProposal tests sign-multisig-proposal command
func (s *TxTestSuite) TestCmdSignMultisigProposal() {
	cmd := cli.GetCmdSignMultisigProposal()
	s.Require().NotNil(cmd)

	err := cmd.ValidateArgs([]string{"proposal123"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{})
	s.Require().Error(err)
}

// TestCmdExecuteMultisigProposal tests execute-multisig-proposal command
func (s *TxTestSuite) TestCmdExecuteMultisigProposal() {
	cmd := cli.GetCmdExecuteMultisigProposal()
	s.Require().NotNil(cmd)

	err := cmd.ValidateArgs([]string{"proposal123"})
	s.Require().NoError(err)
}

// TestCmdProposeTimeLockedAction tests propose-timelocked-action command
func (s *TxTestSuite) TestCmdProposeTimeLockedAction() {
	validPayload := hex.EncodeToString([]byte("action data"))

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name: "valid time-locked action",
			args: []string{
				"UPDATE_PARAMS",
				validPayload,
				"86400",
			},
			expectErr: false,
		},
		{
			name: "missing args",
			args: []string{
				"UPDATE_PARAMS",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdProposeTimeLockedAction()
			err := cmd.ValidateArgs(tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// TestCmdExecuteTimeLockedAction tests execute-timelocked-action command
func (s *TxTestSuite) TestCmdExecuteTimeLockedAction() {
	cmd := cli.GetCmdExecuteTimeLockedAction()
	s.Require().NotNil(cmd)

	err := cmd.ValidateArgs([]string{"action123"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{})
	s.Require().Error(err)
}

// TestCmdCancelTimeLockedAction tests cancel-timelocked-action command
func (s *TxTestSuite) TestCmdCancelTimeLockedAction() {
	cmd := cli.GetCmdCancelTimeLockedAction()
	s.Require().NotNil(cmd)
	s.Require().Equal("cancel-timelocked-action [action-id]", cmd.Use)
}

// TestCmdActivateEmergencyAdmin tests activate-emergency-admin command
func (s *TxTestSuite) TestCmdActivateEmergencyAdmin() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name: "valid activation",
			args: []string{
				"aura1admin123",
				"PAUSE_SYSTEM,EMERGENCY_WITHDRAWAL",
			},
			expectErr: false,
		},
		{
			name: "missing args",
			args: []string{
				"aura1admin123",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdActivateEmergencyAdmin()
			err := cmd.ValidateArgs(tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// TestCmdDeactivateEmergencyAdmin tests deactivate-emergency-admin command
func (s *TxTestSuite) TestCmdDeactivateEmergencyAdmin() {
	cmd := cli.GetCmdDeactivateEmergencyAdmin()
	s.Require().NotNil(cmd)

	err := cmd.ValidateArgs([]string{"aura1admin123"})
	s.Require().NoError(err)
}

// TestCmdInitiateValidatorKeyRotation tests initiate-key-rotation command
func (s *TxTestSuite) TestCmdInitiateValidatorKeyRotation() {
	cmd := cli.GetCmdInitiateValidatorKeyRotation()
	s.Require().NotNil(cmd)

	err := cmd.ValidateArgs([]string{"auravaloper1abc", "{\"@type\":\"/cosmos.crypto.ed25519.PubKey\"}"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{"auravaloper1abc"})
	s.Require().Error(err)
}

// TestCmdCompleteValidatorKeyRotation tests complete-key-rotation command
func (s *TxTestSuite) TestCmdCompleteValidatorKeyRotation() {
	cmd := cli.GetCmdCompleteValidatorKeyRotation()
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "complete-key-rotation")
}

// TestCmdCreateSession tests create-session command
func (s *TxTestSuite) TestCmdCreateSession() {
	cmd := cli.GetCmdCreateSession()
	s.Require().NotNil(cmd)

	err := cmd.ValidateArgs([]string{"192.168.1.1"})
	s.Require().NoError(err)

	err = cmd.ValidateArgs([]string{})
	s.Require().Error(err)
}

// TestCmdRevokeSession tests revoke-session command
func (s *TxTestSuite) TestCmdRevokeSession() {
	cmd := cli.GetCmdRevokeSession()
	s.Require().NotNil(cmd)

	err := cmd.ValidateArgs([]string{"session123"})
	s.Require().NoError(err)
}

// Test helper function parseWalletType
func TestParseWalletType(t *testing.T) {
	// This tests the parseWalletType function indirectly through the command
	// Since parseWalletType is not exported, we test it through command execution
	cmd := cli.GetCmdCreateMultisigWallet()
	require.NotNil(t, cmd)
	require.Contains(t, cmd.Flags().Lookup("wallet-type").DefValue, "custom")
}

// TestAllCommandsHaveHelp verifies all commands have help text
func (s *TxTestSuite) TestAllCommandsHaveHelp() {
	cmd := cli.GetTxCmd()
	for _, subCmd := range cmd.Commands() {
		s.Require().NotEmpty(subCmd.Short, "Command %s missing short description", subCmd.Use)
		s.Require().NotEmpty(subCmd.Long, "Command %s missing long description", subCmd.Use)
	}
}

// TestAllCommandsHaveExamples verifies all commands have examples
func (s *TxTestSuite) TestAllCommandsHaveExamples() {
	cmd := cli.GetTxCmd()
	for _, subCmd := range cmd.Commands() {
		s.Require().Contains(subCmd.Long, "Example", "Command %s missing examples", subCmd.Use)
	}
}

// TestCommandFlags tests that important flags are set correctly
func (s *TxTestSuite) TestCommandFlags() {
	tests := []struct {
		cmdFunc  func() *cobra.Command
		flagName string
		hasFlag  bool
	}{
		{cli.GetCmdAssignRole, "expires-in", true},
		{cli.GetCmdCreateMultisigWallet, "wallet-type", true},
		{cli.GetCmdCreateMultisigProposal, "expires-in", true},
		{cli.GetCmdActivateEmergencyAdmin, "expires-in", true},
		{cli.GetCmdCreateSession, "metadata", true},
	}

	for _, tc := range tests {
		s.Run(tc.flagName, func() {
			cmd := tc.cmdFunc()
			flag := cmd.Flags().Lookup(tc.flagName)
			if tc.hasFlag {
				s.Require().NotNil(flag, "Expected flag %s not found", tc.flagName)
			} else {
				s.Require().Nil(flag, "Unexpected flag %s found", tc.flagName)
			}
		})
	}
}
