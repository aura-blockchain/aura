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
	s.Require().Equal("privacy", cmd.Use)
	s.Require().True(cmd.DisableFlagParsing)
	s.Require().Greater(len(cmd.Commands()), 0)
}

// TestCmdSubmitPrivateTransaction tests private transaction submission command
func (s *TxCLITestSuite) TestCmdSubmitPrivateTransaction() {
	cmd := CmdSubmitPrivateTransaction()

	s.Require().NotNil(cmd)
	s.Require().Equal("submit-private-tx [tx-data-file]", cmd.Use)
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotEmpty(cmd.Long)

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
			// Test Args validator
			if cmd.Args != nil {
				err := cmd.Args(cmd, tc.args)
				if tc.expectErr {
					s.Require().Error(err)
				} else {
					s.Require().NoError(err)
				}
			}
		})
	}
}

// TestCmdCreateMixingPool tests mixing pool creation command
func (s *TxCLITestSuite) TestCmdCreateMixingPool() {
	cmd := CmdCreateMixingPool()

	s.Require().NotNil(cmd)
	s.Require().Equal("create-mixing-pool [min-participants] [max-participants] [denomination] [mixing-rounds] [deadline-duration]", cmd.Use)
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotEmpty(cmd.Long)

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid argument count",
			args:      []string{"3", "10", "1000000", "5", "3600"},
			expectErr: false,
		},
		{
			name:      "valid with different params",
			args:      []string{"5", "20", "5000000", "10", "7200"},
			expectErr: false,
		},
		{
			name:      "missing arguments",
			args:      []string{"3", "10"},
			expectErr: true,
		},
		{
			name:      "too many arguments",
			args:      []string{"3", "10", "1000000", "5", "3600", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Test Args validator
			if cmd.Args != nil {
				err := cmd.Args(cmd, tc.args)
				if tc.expectErr {
					s.Require().Error(err)
				} else {
					s.Require().NoError(err)
				}
			}
		})
	}
}

// TestCmdJoinMixingPool tests join mixing pool command
func (s *TxCLITestSuite) TestCmdJoinMixingPool() {
	cmd := CmdJoinMixingPool()

	s.Require().NotNil(cmd)
	s.Require().Equal("join-mixing-pool [pool-id] [commitment]", cmd.Use)
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotEmpty(cmd.Long)

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid argument count",
			args:      []string{"pool-123", "abc123def456"},
			expectErr: false,
		},
		{
			name:      "valid join another pool",
			args:      []string{"pool-456", "789abcdef012"},
			expectErr: false,
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
		{
			name:      "too many arguments",
			args:      []string{"pool-123", "abc123", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Test Args validator
			if cmd.Args != nil {
				err := cmd.Args(cmd, tc.args)
				if tc.expectErr {
					s.Require().Error(err)
				} else {
					s.Require().NoError(err)
				}
			}
		})
	}
}

// TestCmdRegisterViewKey tests view key registration command
func (s *TxCLITestSuite) TestCmdRegisterViewKey() {
	cmd := CmdRegisterViewKey()

	s.Require().NotNil(cmd)
	s.Require().Equal("register-view-key [key-type] [public-view-key] [permissions]", cmd.Use)
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotEmpty(cmd.Long)

	// Verify expiration flag exists
	expirationFlag := cmd.Flags().Lookup("expiration")
	s.Require().NotNil(expirationFlag)

	tests := []struct {
		name      string
		args      []string
		flags     map[string]string
		expectErr bool
	}{
		{
			name:      "valid argument count",
			args:      []string{"INCOMING", "abc123", "view,decrypt"},
			expectErr: false,
		},
		{
			name:      "valid audit view key",
			args:      []string{"AUDIT", "def456", "view,decrypt,audit"},
			expectErr: false,
		},
		{
			name:      "valid with expiration flag",
			args:      []string{"INCOMING", "abc123", "view,decrypt"},
			flags:     map[string]string{"expiration": "3600"},
			expectErr: false,
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
		{
			name:      "too many arguments",
			args:      []string{"INCOMING", "abc123", "view,decrypt", "extra"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Test Args validator
			if cmd.Args != nil {
				err := cmd.Args(cmd, tc.args)
				if tc.expectErr {
					s.Require().Error(err)
				} else {
					s.Require().NoError(err)
				}
			}

			// Test flag parsing
			for k, v := range tc.flags {
				err := cmd.Flags().Set(k, v)
				s.Require().NoError(err)
			}
		})
	}
}

// TestCmdRevokeViewKey tests view key revocation command
func (s *TxCLITestSuite) TestCmdRevokeViewKey() {
	cmd := CmdRevokeViewKey()

	s.Require().NotNil(cmd)
	s.Require().Equal("revoke-view-key [public-view-key]", cmd.Use)
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotEmpty(cmd.Long)

	tests := []struct {
		name      string
		args      []string
		expectErr bool
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
			// Test Args validator
			if cmd.Args != nil {
				err := cmd.Args(cmd, tc.args)
				if tc.expectErr {
					s.Require().Error(err)
				} else {
					s.Require().NoError(err)
				}
			}
		})
	}
}

// TestCmdUpdateNetworkPrivacy tests network privacy update command
func (s *TxCLITestSuite) TestCmdUpdateNetworkPrivacy() {
	cmd := CmdUpdateNetworkPrivacy()

	s.Require().NotNil(cmd)
	s.Require().Equal("update-network-privacy [network-type]", cmd.Use)
	s.Require().NotEmpty(cmd.Short)
	s.Require().NotEmpty(cmd.Long)

	// Verify flags exist
	s.Require().NotNil(cmd.Flags().Lookup("onion-address"))
	s.Require().NotNil(cmd.Flags().Lookup("i2p-destination"))
	s.Require().NotNil(cmd.Flags().Lookup("proxy-enabled"))
	s.Require().NotNil(cmd.Flags().Lookup("circuit-lifetime"))
	s.Require().NotNil(cmd.Flags().Lookup("stream-isolation"))

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
			// Test Args validator
			if cmd.Args != nil {
				err := cmd.Args(cmd, tc.args)
				if tc.expectErr {
					s.Require().Error(err)
				} else {
					s.Require().NoError(err)
				}
			}

			// Test flag parsing
			for k, v := range tc.flags {
				err := cmd.Flags().Set(k, v)
				s.Require().NoError(err)
			}
		})
	}
}
