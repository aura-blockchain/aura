// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

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
	s.Require().Equal("walletsecurity", cmd.Use)
	s.Require().True(cmd.DisableFlagParsing)
	s.Require().Greater(len(cmd.Commands()), 0)
}

// TestCmdRegisterHardwareWallet tests hardware wallet registration command
func (s *TxCLITestSuite) TestCmdRegisterHardwareWallet() {
	cmd := CmdRegisterHardwareWallet()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Equal("register-hw-wallet", cmd.Use[:18])
	s.Require().Greater(len(cmd.Short), 0)
	s.Require().Greater(len(cmd.Long), 0)

	// Test argument validation - command expects exactly 5 args
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "correct number of args",
			args:      []string{"LEDGER", "device123", "2.1.0", "m/44'/118'/0'/0/0", "0xabcdef1234567890"},
			expectErr: false,
		},
		{
			name:      "missing arguments",
			args:      []string{"LEDGER", "device123"},
			expectErr: true,
		},
		{
			name:      "too many arguments",
			args:      []string{"LEDGER", "device123", "2.1.0", "m/44'/118'/0'/0/0", "0xabcdef1234567890", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Test argument count validation
			err := cmd.Args(cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// TestCmdCreateMultiSigWallet tests multi-sig wallet creation command
func (s *TxCLITestSuite) TestCmdCreateMultiSigWallet() {
	cmd := CmdCreateMultiSigWallet()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "create-multisig")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 2 args
	err := cmd.Args(cmd, []string{"alice,bob,charlie", "2"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{"alice,bob"})
	s.Require().Error(err, "should error with missing threshold")
}

// TestCmdSignMultiSigTransaction tests multi-sig transaction signing command
func (s *TxCLITestSuite) TestCmdSignMultiSigTransaction() {
	cmd := CmdSignMultiSigTransaction()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "sign-multisig-tx")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation
	err := cmd.Args(cmd, []string{"tx123", "0xabcdef1234567890"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{"tx123"})
	s.Require().Error(err, "should error with missing signature")
}

// TestCmdConfigureSocialRecovery tests social recovery configuration command
func (s *TxCLITestSuite) TestCmdConfigureSocialRecovery() {
	cmd := CmdConfigureSocialRecovery()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "configure-social-recovery")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 4 args
	err := cmd.Args(cmd, []string{"wallet123", "guardian1,guardian2,guardian3", "2", "7d"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{"wallet123", "guardian1"})
	s.Require().Error(err, "should error with missing arguments")
}

// TestCmdEnrollBiometric tests biometric enrollment command
func (s *TxCLITestSuite) TestCmdEnrollBiometric() {
	cmd := CmdEnrollBiometric()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "enroll-biometric")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 3 args
	err := cmd.Args(cmd, []string{"wallet123", "FINGERPRINT", "0xabcdef1234567890"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{"wallet123", "FINGERPRINT"})
	s.Require().Error(err, "should error with missing enrollment data")
}

// TestCmdStoreInSecureEnclave tests secure enclave storage command
func (s *TxCLITestSuite) TestCmdStoreInSecureEnclave() {
	cmd := CmdStoreInSecureEnclave()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "store-in-enclave")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 4 args
	err := cmd.Args(cmd, []string{"wallet123", "TEE", "0xabcdef1234567890", "cert123"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{"wallet123", "TEE"})
	s.Require().Error(err, "should error with missing arguments")
}

// TestCmdCreateEncryptedBackup tests encrypted backup creation command
func (s *TxCLITestSuite) TestCmdCreateEncryptedBackup() {
	cmd := CmdCreateEncryptedBackup()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "create-backup")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 7 args
	err := cmd.Args(cmd, []string{"wallet123", "0xabcdef1234567890", "AES256", "PBKDF2", "0x1234567890abcdef", "10000", "CLOUD"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{"wallet123"})
	s.Require().Error(err, "should error with missing arguments")
}

// TestCmdConfigureDustFilter tests dust filter configuration command
func (s *TxCLITestSuite) TestCmdConfigureDustFilter() {
	cmd := CmdConfigureDustFilter()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "configure-dust-filter")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 5 args
	err := cmd.Args(cmd, []string{"wallet123", "true", "1000", "10", "5"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{"wallet123", "true"})
	s.Require().Error(err, "should error with missing arguments")
}

// TestCmdValidateAddressChecksum tests address checksum validation command
func (s *TxCLITestSuite) TestCmdValidateAddressChecksum() {
	cmd := CmdValidateAddressChecksum()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "validate-address")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 2 args
	err := cmd.Args(cmd, []string{"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed", "EIP55"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"})
	s.Require().Error(err, "should error with missing algorithm")
}

// TestCmdConfigureSession tests session configuration command
func (s *TxCLITestSuite) TestCmdConfigureSession() {
	cmd := CmdConfigureSession()

	// Test command structure
	s.Require().NotNil(cmd)
	s.Require().Contains(cmd.Use, "configure-session")
	s.Require().Greater(len(cmd.Short), 0)

	// Test argument validation - expects exactly 4 args
	err := cmd.Args(cmd, []string{"wallet123", "30m", "true", "600"})
	s.Require().NoError(err)

	err = cmd.Args(cmd, []string{"wallet123", "30m"})
	s.Require().Error(err, "should error with missing arguments")
}

// TestAllCommandsExist ensures all expected commands are registered
func (s *TxCLITestSuite) TestAllCommandsExist() {
	cmd := GetTxCmd()
	subcommands := cmd.Commands()

	// Verify we have the expected number of commands
	s.Require().GreaterOrEqual(len(subcommands), 10, "should have at least 10 subcommands")

	// Verify specific commands exist by checking their names
	commandNames := make(map[string]bool)
	for _, subcmd := range subcommands {
		commandNames[subcmd.Name()] = true
	}

	expectedCommands := []string{
		"register-hw-wallet",
		"create-multisig",
		"sign-multisig-tx",
		"configure-social-recovery",
		"enroll-biometric",
		"store-in-enclave",
		"create-backup",
		"configure-dust-filter",
		"validate-address",
		"configure-session",
	}

	for _, expected := range expectedCommands {
		s.Require().True(commandNames[expected], "command %s should exist", expected)
	}
}
