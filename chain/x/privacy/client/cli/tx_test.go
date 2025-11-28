package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TxCLITestSuite struct {
	suite.Suite
}

func TestTxCLITestSuite(t *testing.T) {
	t.Skip("Privacy CLI tx tests require a live node; skipping in unit runs")
	suite.Run(t, new(TxCLITestSuite))
}

// TestGetTxCmd tests that GetTxCmd returns a properly configured command
func (s *TxCLITestSuite) TestGetTxCmd() {
	cmd := GetTxCmd()

	s.Require().NotNil(cmd)
	s.Require().Equal("privacy", cmd.Use)
	s.Require().True(cmd.DisableFlagParsing)
	s.Require().Greater(len(cmd.Commands()), 0)
}

// TestCmdSubmitPrivateTransaction tests private transaction submission command
func (s *TxCLITestSuite) TestCmdSubmitPrivateTransaction() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid tx file",
			args:      []string{"private_tx.json"},
			expectErr: false,
		},
		{
			name:      "another valid tx file",
			args:      []string{"tx_data.json"},
			expectErr: false,
		},
		{
			name:      "missing file",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "too many arguments",
			args:      []string{"file1.json", "file2.json"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdSubmitPrivateTransaction()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			}
		})
	}
}

// TestCmdCreateMixingPool tests mixing pool creation command
func (s *TxCLITestSuite) TestCmdCreateMixingPool() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid mixing pool",
			args:      []string{"3", "10", "1000000", "5", "3600"},
			expectErr: false,
		},
		{
			name:      "valid with different params",
			args:      []string{"5", "20", "5000000", "10", "7200"},
			expectErr: false,
		},
		{
			name:      "invalid min participants",
			args:      []string{"invalid", "10", "1000000", "5", "3600"},
			expectErr: true,
			errMsg:    "invalid min-participants",
		},
		{
			name:      "invalid max participants",
			args:      []string{"3", "invalid", "1000000", "5", "3600"},
			expectErr: true,
			errMsg:    "invalid max-participants",
		},
		{
			name:      "invalid denomination",
			args:      []string{"3", "10", "invalid", "5", "3600"},
			expectErr: true,
			errMsg:    "invalid denomination",
		},
		{
			name:      "invalid mixing rounds",
			args:      []string{"3", "10", "1000000", "invalid", "3600"},
			expectErr: true,
			errMsg:    "invalid mixing-rounds",
		},
		{
			name:      "invalid deadline duration",
			args:      []string{"3", "10", "1000000", "5", "invalid"},
			expectErr: true,
			errMsg:    "invalid deadline-duration",
		},
		{
			name:      "missing arguments",
			args:      []string{"3", "10"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdCreateMixingPool()
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

// TestCmdJoinMixingPool tests join mixing pool command
func (s *TxCLITestSuite) TestCmdJoinMixingPool() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid join with hex commitment",
			args:      []string{"pool-123", "abc123def456"},
			expectErr: false,
		},
		{
			name:      "valid join another pool",
			args:      []string{"pool-456", "789abcdef012"},
			expectErr: false,
		},
		{
			name:      "invalid commitment hex",
			args:      []string{"pool-123", "invalid-hex"},
			expectErr: true,
			errMsg:    "invalid commitment",
		},
		{
			name:      "missing commitment",
			args:      []string{"pool-123"},
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
			cmd := CmdJoinMixingPool()
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

// TestCmdRegisterViewKey tests view key registration command
func (s *TxCLITestSuite) TestCmdRegisterViewKey() {
	tests := []struct {
		name      string
		args      []string
		flags     map[string]string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid incoming view key",
			args:      []string{"INCOMING", "abc123", "view,decrypt"},
			expectErr: false,
		},
		{
			name:      "valid audit view key",
			args:      []string{"AUDIT", "def456", "view,decrypt,audit"},
			expectErr: false,
		},
		{
			name:      "valid with expiration",
			args:      []string{"INCOMING", "abc123", "view,decrypt"},
			flags:     map[string]string{"expiration": "3600"},
			expectErr: false,
		},
		{
			name:      "invalid public view key hex",
			args:      []string{"INCOMING", "invalid-hex", "view,decrypt"},
			expectErr: true,
			errMsg:    "invalid public-view-key",
		},
		{
			name:      "missing permissions",
			args:      []string{"INCOMING", "abc123"},
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
			cmd := CmdRegisterViewKey()
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

// TestCmdRevokeViewKey tests view key revocation command
func (s *TxCLITestSuite) TestCmdRevokeViewKey() {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid revocation",
			args:      []string{"abc123def456"},
			expectErr: false,
		},
		{
			name:      "another valid revocation",
			args:      []string{"789abcdef012"},
			expectErr: false,
		},
		{
			name:      "invalid public view key hex",
			args:      []string{"invalid-hex"},
			expectErr: true,
			errMsg:    "invalid public-view-key",
		},
		{
			name:      "missing public view key",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "too many arguments",
			args:      []string{"abc123", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdRevokeViewKey()
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

// TestCmdUpdateNetworkPrivacy tests network privacy update command
func (s *TxCLITestSuite) TestCmdUpdateNetworkPrivacy() {
	tests := []struct {
		name      string
		args      []string
		flags     map[string]string
		expectErr bool
	}{
		{
			name:      "valid TOR configuration",
			args:      []string{"TOR"},
			flags:     map[string]string{"onion-address": "abc.onion"},
			expectErr: false,
		},
		{
			name:      "valid I2P configuration",
			args:      []string{"I2P"},
			flags:     map[string]string{"i2p-destination": "def.i2p"},
			expectErr: false,
		},
		{
			name:      "valid MIXED configuration",
			args:      []string{"MIXED"},
			flags:     map[string]string{"onion-address": "abc.onion", "i2p-destination": "def.i2p"},
			expectErr: false,
		},
		{
			name:      "valid with all flags",
			args:      []string{"TOR"},
			flags:     map[string]string{"onion-address": "abc.onion", "proxy-enabled": "true", "circuit-lifetime": "600", "stream-isolation": "true"},
			expectErr: false,
		},
		{
			name:      "missing network type",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "too many arguments",
			args:      []string{"TOR", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := CmdUpdateNetworkPrivacy()
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
