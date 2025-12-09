package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type WalletSecurityTxSuite struct {
	suite.Suite
}

func TestWalletSecurityTxSuite(t *testing.T) {
	suite.Run(t, new(WalletSecurityTxSuite))
}

func (s *WalletSecurityTxSuite) TestGetTxCmd() {
	cmd := GetTxCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("walletsecurity", cmd.Use)
	require.True(cmd.DisableFlagParsing)

	expected := []string{
		"register-hw-wallet",
		"create-multisig",
		"sign-multisig-tx",
		"configure-social-recovery",
		"initiate-recovery",
		"approve-recovery",
		"execute-recovery",
		"simulate-tx",
		"verify-domain",
		"set-spending-limit",
		"configure-session",
		"lock-session",
		"unlock-session",
		"enroll-biometric",
		"authenticate-biometric",
		"store-in-enclave",
		"create-backup",
		"configure-dust-filter",
		"validate-address",
	}

	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, name := range expected {
		require.True(names[name], "expected tx subcommand %s", name)
	}
}

func (s *WalletSecurityTxSuite) TestCmdRegisterHardwareWalletArgs() {
	require := s.Require()
	require.NoError(CmdRegisterHardwareWallet().ValidateArgs([]string{"LEDGER", "dev", "1.0.0", "m/44'/118'/0'/0/0", "0xsig"}))
	require.Error(CmdRegisterHardwareWallet().ValidateArgs([]string{"LEDGER"}))
}

func (s *WalletSecurityTxSuite) TestCmdCreateMultiSigWalletArgs() {
	require := s.Require()
	require.NoError(CmdCreateMultiSigWallet().ValidateArgs([]string{"a,b,c", "2"}))
	require.Error(CmdCreateMultiSigWallet().ValidateArgs([]string{"a,b,c"}))
}

func (s *WalletSecurityTxSuite) TestCmdValidateAddressChecksumArgs() {
	require := s.Require()
	require.NoError(CmdValidateAddressChecksum().ValidateArgs([]string{"aura1address", "bech32"}))
	require.Error(CmdValidateAddressChecksum().ValidateArgs([]string{}))
}

func (s *WalletSecurityTxSuite) TestCmdStoreInSecureEnclaveArgs() {
	require := s.Require()
	require.NoError(CmdStoreInSecureEnclave().ValidateArgs([]string{"wallet", "TEE", "0xdeadbeef", "cert"}))
	require.Error(CmdStoreInSecureEnclave().ValidateArgs([]string{"wallet"}))
}
