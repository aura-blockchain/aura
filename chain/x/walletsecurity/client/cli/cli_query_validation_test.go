package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type WalletSecurityQuerySuite struct {
	suite.Suite
}

func TestWalletSecurityQuerySuite(t *testing.T) {
	suite.Run(t, new(WalletSecurityQuerySuite))
}

func (s *WalletSecurityQuerySuite) TestGetQueryCmd() {
	cmd := GetQueryCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("walletsecurity", cmd.Use)
	require.True(cmd.DisableFlagParsing)

	expected := []string{
		"hw-wallet",
		"multisig",
		"pending-multisig-tx",
		"recovery-request",
		"social-recovery",
		"spending-limit",
		"session",
		"security-metrics",
		"domain-verification",
		"dust-filter",
	}

	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, name := range expected {
		require.True(names[name], "expected query subcommand %s", name)
	}
}

func (s *WalletSecurityQuerySuite) TestCommandArgValidation() {
	require := s.Require()

	require.NoError(CmdQueryHardwareWallet().ValidateArgs([]string{"addr"}))
	require.Error(CmdQueryHardwareWallet().ValidateArgs([]string{}))

	require.NoError(CmdQueryMultiSigWallet().ValidateArgs([]string{"wallet"}))
	require.Error(CmdQueryMultiSigWallet().ValidateArgs([]string{}))

	require.NoError(CmdQueryRecoveryRequest().ValidateArgs([]string{"id"}))
	require.Error(CmdQueryRecoveryRequest().ValidateArgs([]string{}))

	require.NoError(CmdQueryPendingMultiSigTx().ValidateArgs([]string{"tx-id"}))
	require.Error(CmdQueryPendingMultiSigTx().ValidateArgs([]string{}))

}
