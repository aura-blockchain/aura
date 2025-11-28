package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
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
	s.clientCtx = client.Context{}.
		WithAccountRetriever(client.MockAccountRetriever{})
}

// TestGetQueryCmd tests that all query commands are registered
func (s *QueryTestSuite) TestGetQueryCmd() {
	cmd := GetQueryCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("cryptography", cmd.Use)
	require.Equal("Querying commands for the cryptography module", cmd.Short)
	require.True(cmd.DisableFlagParsing)

	// Verify all expected subcommands are registered
	expectedCmds := []string{
		"params",
		"key-rotation-schedule",
		"threshold-scheme",
		"verify-zk-proof",
		"secure-enclave",
		"quantum-resistant-key",
		"random-source-status",
		"certificate-pin",
	}

	subCmds := cmd.Commands()
	require.GreaterOrEqual(len(subCmds), len(expectedCmds))

	cmdNames := make(map[string]bool)
	for _, subCmd := range subCmds {
		cmdNames[subCmd.Name()] = true
	}

	for _, expectedCmd := range expectedCmds {
		require.True(cmdNames[expectedCmd], "command %s should be registered", expectedCmd)
	}
}

// TestCmdQueryParams tests the params query command
func (s *QueryTestSuite) TestCmdQueryParams() {
	cmd := CmdQueryParams()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("params", cmd.Use)
	require.Contains(cmd.Long, "Key rotation")
	require.Contains(cmd.Long, "Threshold signature")

	// Test args validation - params takes no args
	err := cmd.ValidateArgs([]string{})
	require.NoError(err)

	err = cmd.ValidateArgs([]string{"extra-arg"})
	require.Error(err)
}

// TestCmdQueryKeyRotationSchedule tests the key-rotation-schedule query command
func (s *QueryTestSuite) TestCmdQueryKeyRotationSchedule() {
	cmd := CmdQueryKeyRotationSchedule()
	require := s.Require()

	require.NotNil(cmd)
	require.Contains(cmd.Use, "key-rotation-schedule")
	require.Contains(cmd.Long, "Key ID")
	require.Contains(cmd.Long, "Rotation interval")

	// Test args validation
	err := cmd.ValidateArgs([]string{"schedule-123"})
	require.NoError(err)

	err = cmd.ValidateArgs([]string{})
	require.Error(err)
}

// TestCmdQueryThresholdScheme tests the threshold-scheme query command
func (s *QueryTestSuite) TestCmdQueryThresholdScheme() {
	cmd := CmdQueryThresholdScheme()
	require := s.Require()

	require.NotNil(cmd)
	require.Contains(cmd.Use, "threshold-scheme")
	require.Contains(cmd.Long, "Threshold value")
	require.Contains(cmd.Long, "Participant")

	// Test args validation
	err := cmd.ValidateArgs([]string{"scheme-123"})
	require.NoError(err)

	err = cmd.ValidateArgs([]string{})
	require.Error(err)
}

// TestCmdQueryVerifyZKProof tests the verify-zk-proof query command
func (s *QueryTestSuite) TestCmdQueryVerifyZKProof() {
	cmd := CmdQueryVerifyZKProof()
	require := s.Require()

	require.NotNil(cmd)
	require.Contains(cmd.Use, "verify-zk-proof")
	require.Contains(cmd.Long, "Verification result")

	// Test args validation - needs 3 args
	err := cmd.ValidateArgs([]string{"proof-123", "0xproofdata", "0xinputs"})
	require.NoError(err)

	err = cmd.ValidateArgs([]string{"proof-123"})
	require.Error(err)
}

// TestCmdQuerySecureEnclave tests the secure-enclave query command
func (s *QueryTestSuite) TestCmdQuerySecureEnclave() {
	cmd := CmdQuerySecureEnclave()
	require := s.Require()

	require.NotNil(cmd)
	require.Contains(cmd.Use, "secure-enclave")
	require.Contains(cmd.Long, "Enclave type")
	require.Contains(cmd.Long, "HSM")
	require.Contains(cmd.Long, "SGX")

	// Test args validation
	err := cmd.ValidateArgs([]string{"enclave-123"})
	require.NoError(err)

	err = cmd.ValidateArgs([]string{})
	require.Error(err)
}

// TestCmdQueryQuantumResistantKey tests the quantum-resistant-key query command
func (s *QueryTestSuite) TestCmdQueryQuantumResistantKey() {
	cmd := CmdQueryQuantumResistantKey()
	require := s.Require()

	require.NotNil(cmd)
	require.Contains(cmd.Use, "quantum-resistant-key")
	require.Contains(cmd.Long, "Algorithm")
	require.Contains(cmd.Long, "Dilithium")

	// Test args validation
	err := cmd.ValidateArgs([]string{"key-123"})
	require.NoError(err)

	err = cmd.ValidateArgs([]string{})
	require.Error(err)
}

// TestCmdQueryRandomSourceStatus tests the random-source-status query command
func (s *QueryTestSuite) TestCmdQueryRandomSourceStatus() {
	cmd := CmdQueryRandomSourceStatus()
	require := s.Require()

	require.NotNil(cmd)
	require.Contains(cmd.Use, "random-source-status")
	require.Contains(cmd.Long, "random")
	require.Contains(cmd.Long, "Health")

	// Test args validation - no args required
	err := cmd.ValidateArgs([]string{})
	require.NoError(err)

	err = cmd.ValidateArgs([]string{"extra"})
	require.Error(err)
}

// TestCmdQueryCertificatePin tests the certificate-pin query command
func (s *QueryTestSuite) TestCmdQueryCertificatePin() {
	cmd := CmdQueryCertificatePin()
	require := s.Require()

	require.NotNil(cmd)
	require.Contains(cmd.Use, "certificate-pin")
	require.Contains(cmd.Long, "hostname")
	require.Contains(cmd.Long, "certificate")

	// Test args validation
	err := cmd.ValidateArgs([]string{"api.aura.network"})
	require.NoError(err)

	err = cmd.ValidateArgs([]string{})
	require.Error(err)
}

// TestAllQueryCommandsHaveHelp verifies all query commands have help text
func (s *QueryTestSuite) TestAllQueryCommandsHaveHelp() {
	cmd := GetQueryCmd()
	for _, subCmd := range cmd.Commands() {
		s.Require().NotEmpty(subCmd.Short, "Command %s missing short description", subCmd.Use)
		s.Require().NotEmpty(subCmd.Long, "Command %s missing long description", subCmd.Use)
	}
}

// TestAllQueryCommandsHaveExamples verifies all query commands have examples
func (s *QueryTestSuite) TestAllQueryCommandsHaveExamples() {
	cmd := GetQueryCmd()
	for _, subCmd := range cmd.Commands() {
		// params command might not have examples
		if subCmd.Name() != "params" {
			s.Require().Contains(subCmd.Long, "Example", "Command %s missing examples", subCmd.Use)
		}
	}
}

// Benchmark tests
func BenchmarkQueryParams(b *testing.B) {
	cmd := CmdQueryParams()
	for i := 0; i < b.N; i++ {
		_ = cmd.ValidateArgs([]string{})
	}
}

func BenchmarkQueryKeyRotationSchedule(b *testing.B) {
	cmd := CmdQueryKeyRotationSchedule()
	for i := 0; i < b.N; i++ {
		_ = cmd.ValidateArgs([]string{"schedule-123"})
	}
}
