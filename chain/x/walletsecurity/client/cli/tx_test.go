package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TxCLITestSuite struct {
	suite.Suite
}

func TestTxCLITestSuite(t *testing.T) {
	t.Skip("CLI transaction tests require a live node; skipping in unit runs")
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
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid ledger wallet",
			args:      []string{"LEDGER", "device123", "2.1.0", "m/44'/118'/0'/0/0", "0xabcdef1234567890"},
			expectErr: false,
		},
		{
			name:      "valid trezor wallet",
			args:      []string{"TREZOR", "device456", "1.10.0", "m/44'/118'/0'/0/0", "0x1234567890abcdef"},
			expectErr: false,
		},
		{
			name:      "invalid hardware type",
			args:      []string{"INVALID", "device123", "2.1.0", "m/44'/118'/0'/0/0", "0xabcdef1234567890"},
			expectErr: true,
			errMsg:    "invalid hardware wallet type",
		},
		{
			name:      "invalid signature hex",
			args:      []string{"LEDGER", "device123", "2.1.0", "m/44'/118'/0'/0/0", "invalid-hex"},
			expectErr: true,
			errMsg:    "invalid signature hex",
		},
		{
			name:      "missing arguments",
			args:      []string{"LEDGER", "device123"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdRegisterHardwareWallet()
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

// TestCmdCreateMultiSigWallet tests multi-sig wallet creation command
func (s *TxCLITestSuite) TestCmdCreateMultiSigWallet() {
	tests := []struct {
		name      string
		args      []string
		flags     map[string]string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid multisig",
			args:      []string{"alice,bob,charlie", "2"},
			expectErr: false,
		},
		{
			name:      "with weights",
			args:      []string{"alice,bob,charlie", "2"},
			flags:     map[string]string{"signer-weights": "alice=2,bob=1,charlie=1"},
			expectErr: false,
		},
		{
			name:      "with time-lock",
			args:      []string{"alice,bob", "2"},
			flags:     map[string]string{"time-lock": "3600s"},
			expectErr: false,
		},
		{
			name:      "invalid threshold",
			args:      []string{"alice,bob", "invalid"},
			expectErr: true,
			errMsg:    "invalid threshold",
		},
		{
			name:      "invalid weight",
			args:      []string{"alice,bob", "2"},
			flags:     map[string]string{"signer-weights": "alice=invalid"},
			expectErr: true,
			errMsg:    "invalid weight",
		},
		{
			name:      "invalid time-lock",
			args:      []string{"alice,bob", "2"},
			flags:     map[string]string{"time-lock": "invalid"},
			expectErr: true,
			errMsg:    "invalid time-lock",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdCreateMultiSigWallet()
			cmd.SetArgs(tc.args)

			for k, v := range tc.flags {
				cmd.Flags().Set(k, v)
			}

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

// TestCmdSignMultiSigTransaction tests multi-sig transaction signing command
func (s *TxCLITestSuite) TestCmdSignMultiSigTransaction() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid signature",
			args:      []string{"tx123", "0xabcdef1234567890"},
			expectErr: false,
		},
		{
			name:      "signature without 0x prefix",
			args:      []string{"tx123", "abcdef1234567890"},
			expectErr: false,
		},
		{
			name:      "invalid signature hex",
			args:      []string{"tx123", "invalid-hex"},
			expectErr: true,
			errMsg:    "invalid signature hex",
		},
		{
			name:      "missing arguments",
			args:      []string{"tx123"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdSignMultiSigTransaction()
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

// TestCmdConfigureSocialRecovery tests social recovery configuration command
func (s *TxCLITestSuite) TestCmdConfigureSocialRecovery() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid configuration",
			args:      []string{"wallet123", "guardian1,guardian2,guardian3", "2", "7d"},
			expectErr: false,
		},
		{
			name:      "single guardian",
			args:      []string{"wallet123", "guardian1", "1", "24h"},
			expectErr: false,
		},
		{
			name:      "invalid threshold",
			args:      []string{"wallet123", "guardian1,guardian2", "invalid", "7d"},
			expectErr: true,
			errMsg:    "invalid threshold",
		},
		{
			name:      "invalid delay",
			args:      []string{"wallet123", "guardian1,guardian2", "2", "invalid"},
			expectErr: true,
			errMsg:    "invalid delay",
		},
		{
			name:      "missing arguments",
			args:      []string{"wallet123", "guardian1"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdConfigureSocialRecovery()
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

// TestCmdEnrollBiometric tests biometric enrollment command
func (s *TxCLITestSuite) TestCmdEnrollBiometric() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid fingerprint enrollment",
			args:      []string{"wallet123", "FINGERPRINT", "0xabcdef1234567890"},
			expectErr: false,
		},
		{
			name:      "valid face id enrollment",
			args:      []string{"wallet123", "FACE_ID", "0x1234567890abcdef"},
			expectErr: false,
		},
		{
			name:      "valid iris enrollment",
			args:      []string{"wallet123", "IRIS", "0xfedcba0987654321"},
			expectErr: false,
		},
		{
			name:      "invalid biometric type",
			args:      []string{"wallet123", "INVALID", "0xabcdef1234567890"},
			expectErr: true,
			errMsg:    "invalid biometric type",
		},
		{
			name:      "invalid enrollment data",
			args:      []string{"wallet123", "FINGERPRINT", "invalid-hex"},
			expectErr: true,
			errMsg:    "invalid enrollment data hex",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdEnrollBiometric()
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

// TestCmdStoreInSecureEnclave tests secure enclave storage command
func (s *TxCLITestSuite) TestCmdStoreInSecureEnclave() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid TEE enclave",
			args:      []string{"wallet123", "TEE", "0xabcdef1234567890", "cert123"},
			expectErr: false,
		},
		{
			name:      "valid SGX enclave",
			args:      []string{"wallet123", "SGX", "0x1234567890abcdef", "cert456"},
			expectErr: false,
		},
		{
			name:      "valid TPM enclave",
			args:      []string{"wallet123", "TPM", "0xfedcba0987654321", "cert789"},
			expectErr: false,
		},
		{
			name:      "invalid enclave type",
			args:      []string{"wallet123", "INVALID", "0xabcdef1234567890", "cert123"},
			expectErr: true,
			errMsg:    "invalid enclave type",
		},
		{
			name:      "invalid encrypted key",
			args:      []string{"wallet123", "TEE", "invalid-hex", "cert123"},
			expectErr: true,
			errMsg:    "invalid encrypted key hex",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdStoreInSecureEnclave()
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

// TestCmdCreateEncryptedBackup tests encrypted backup creation command
func (s *TxCLITestSuite) TestCmdCreateEncryptedBackup() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid cloud backup",
			args:      []string{"wallet123", "0xabcdef1234567890", "AES256", "PBKDF2", "0x1234567890abcdef", "10000", "CLOUD"},
			expectErr: false,
		},
		{
			name:      "valid hardware backup",
			args:      []string{"wallet123", "0xabcdef1234567890", "AES256", "PBKDF2", "0x1234567890abcdef", "10000", "HARDWARE"},
			expectErr: false,
		},
		{
			name:      "invalid encrypted seed",
			args:      []string{"wallet123", "invalid-hex", "AES256", "PBKDF2", "0x1234567890abcdef", "10000", "CLOUD"},
			expectErr: true,
			errMsg:    "invalid encrypted seed hex",
		},
		{
			name:      "invalid salt",
			args:      []string{"wallet123", "0xabcdef1234567890", "AES256", "PBKDF2", "invalid-hex", "10000", "CLOUD"},
			expectErr: true,
			errMsg:    "invalid salt hex",
		},
		{
			name:      "invalid iterations",
			args:      []string{"wallet123", "0xabcdef1234567890", "AES256", "PBKDF2", "0x1234567890abcdef", "invalid", "CLOUD"},
			expectErr: true,
			errMsg:    "invalid iterations",
		},
		{
			name:      "invalid location",
			args:      []string{"wallet123", "0xabcdef1234567890", "AES256", "PBKDF2", "0x1234567890abcdef", "10000", "INVALID"},
			expectErr: true,
			errMsg:    "invalid backup location",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdCreateEncryptedBackup()
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

// TestCmdConfigureDustFilter tests dust filter configuration command
func (s *TxCLITestSuite) TestCmdConfigureDustFilter() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid enabled filter",
			args:      []string{"wallet123", "true", "1000", "10", "5"},
			expectErr: false,
		},
		{
			name:      "valid disabled filter",
			args:      []string{"wallet123", "false", "1000", "10", "5"},
			expectErr: false,
		},
		{
			name:      "invalid enabled flag",
			args:      []string{"wallet123", "invalid", "1000", "10", "5"},
			expectErr: true,
			errMsg:    "invalid enabled",
		},
		{
			name:      "invalid max dust tx",
			args:      []string{"wallet123", "true", "1000", "invalid", "5"},
			expectErr: true,
			errMsg:    "invalid max dust tx",
		},
		{
			name:      "invalid threshold",
			args:      []string{"wallet123", "true", "1000", "10", "invalid"},
			expectErr: true,
			errMsg:    "invalid threshold",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdConfigureDustFilter()
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

// TestCmdValidateAddressChecksum tests address checksum validation command
func (s *TxCLITestSuite) TestCmdValidateAddressChecksum() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid EIP55",
			args:      []string{"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed", "EIP55"},
			expectErr: false,
		},
		{
			name:      "valid BECH32",
			args:      []string{"aura1abc123def456", "BECH32"},
			expectErr: false,
		},
		{
			name:      "valid BASE58CHECK",
			args:      []string{"1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "BASE58CHECK"},
			expectErr: false,
		},
		{
			name:      "invalid algorithm",
			args:      []string{"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed", "INVALID"},
			expectErr: true,
			errMsg:    "invalid algorithm",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdValidateAddressChecksum()
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

// TestCmdConfigureSession tests session configuration command
func (s *TxCLITestSuite) TestCmdConfigureSession() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid configuration",
			args:      []string{"wallet123", "30m", "true", "600"},
			expectErr: false,
		},
		{
			name:      "valid with hours",
			args:      []string{"wallet123", "2h", "false", "300"},
			expectErr: false,
		},
		{
			name:      "invalid timeout",
			args:      []string{"wallet123", "invalid", "true", "600"},
			expectErr: true,
			errMsg:    "invalid timeout",
		},
		{
			name:      "invalid auto-lock",
			args:      []string{"wallet123", "30m", "invalid", "600"},
			expectErr: true,
			errMsg:    "invalid auto-lock",
		},
		{
			name:      "invalid inactivity threshold",
			args:      []string{"wallet123", "30m", "true", "invalid"},
			expectErr: true,
			errMsg:    "invalid inactivity threshold",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdConfigureSession()
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
