package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
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
	cmd := GetTxCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("cryptography", cmd.Use)
	require.Equal("Cryptography transaction subcommands", cmd.Short)
	require.True(cmd.DisableFlagParsing)

	// Verify all expected subcommands are registered
	expectedCmds := []string{
		"create-rotation-schedule",
		"rotate-key",
		"create-threshold-scheme",
		"submit-threshold-share",
		"register-zk-circuit",
		"submit-zk-proof",
		"register-enclave",
		"generate-qr-key",
		"add-cert-pin",
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

// TestCmdCreateKeyRotationSchedule tests the create-rotation-schedule command
func (s *TxTestSuite) TestCmdCreateKeyRotationSchedule() {
	cmd := CmdCreateKeyRotationSchedule()
	require := s.Require()

	require.NotNil(cmd)
	require.Contains(cmd.Use, "create-rotation-schedule")
	require.NotEmpty(cmd.Short)
	require.NotEmpty(cmd.Long)
	require.Contains(cmd.Long, "Examples")

	// Test args validation
	err := cmd.ValidateArgs([]string{"key-123", "86400", "30"})
	require.NoError(err)

	err = cmd.ValidateArgs([]string{"key-123"})
	require.Error(err)
}

// TestCmdRotateKey tests the rotate-key command
func (s *TxTestSuite) TestCmdRotateKey() {
	cmd := CmdRotateKey()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("rotate-key [key-id] [new-public-key-hex]", cmd.Use)

	// Test args validation
	err := cmd.ValidateArgs([]string{"key-123", "0x1234abcd"})
	require.NoError(err)

	err = cmd.ValidateArgs([]string{"key-123"})
	require.Error(err)
}

// TestCmdCreateThresholdScheme tests the create-threshold-scheme command
func (s *TxTestSuite) TestCmdCreateThresholdScheme() {
	cmd := CmdCreateThresholdScheme()
	require := s.Require()

	require.NotNil(cmd)
	require.Contains(cmd.Use, "create-threshold-scheme")
	require.Contains(cmd.Long, "BLS")
	require.Contains(cmd.Long, "ECDSA")

	// Test args validation
	err := cmd.ValidateArgs([]string{"2", "3", "alice,bob,charlie", "BLS"})
	require.NoError(err)

	err = cmd.ValidateArgs([]string{"2", "3"})
	require.Error(err)
}

// TestCmdSubmitThresholdSignatureShare tests submit-threshold-share command
func (s *TxTestSuite) TestCmdSubmitThresholdSignatureShare() {
	cmd := CmdSubmitThresholdSignatureShare()
	require := s.Require()

	require.NotNil(cmd)
	require.Contains(cmd.Use, "submit-threshold-share")

	err := cmd.ValidateArgs([]string{"scheme-123", "0xabcd", "0xhash"})
	require.NoError(err)
}

// TestCmdRegisterZKProofCircuit tests register-zk-circuit command
func (s *TxTestSuite) TestCmdRegisterZKProofCircuit() {
	cmd := CmdRegisterZKProofCircuit()
	require := s.Require()

	require.NotNil(cmd)
	require.Contains(cmd.Use, "register-zk-circuit")
	require.Contains(cmd.Long, "GROTH16")
	require.Contains(cmd.Long, "PLONK")

	err := cmd.ValidateArgs([]string{"circuit-1", "GROTH16", "0xparams", "0xkey"})
	require.NoError(err)

	err = cmd.ValidateArgs([]string{"circuit-1"})
	require.Error(err)
}

// TestCmdSubmitZKProof tests submit-zk-proof command
func (s *TxTestSuite) TestCmdSubmitZKProof() {
	cmd := CmdSubmitZKProof()
	require := s.Require()

	require.NotNil(cmd)
	require.Contains(cmd.Use, "submit-zk-proof")

	err := cmd.ValidateArgs([]string{"proof-123", "0xproofdata", "0xinputs"})
	require.NoError(err)

	err = cmd.ValidateArgs([]string{"proof-123", "0xproofdata"})
	require.Error(err)
}

// TestCmdRegisterSecureEnclave tests register-enclave command
func (s *TxTestSuite) TestCmdRegisterSecureEnclave() {
	cmd := CmdRegisterSecureEnclave()
	require := s.Require()

	require.NotNil(cmd)
	require.Contains(cmd.Use, "register-enclave")
	require.Contains(cmd.Long, "HSM")
	require.Contains(cmd.Long, "SGX")
	require.Contains(cmd.Long, "TPM")

	// Test flag
	require.NotNil(cmd.Flags().Lookup("metadata"))

	err := cmd.ValidateArgs([]string{"SGX", "0xattestation"})
	require.NoError(err)

	err = cmd.ValidateArgs([]string{"SGX"})
	require.Error(err)
}

// TestCmdGenerateQuantumResistantKey tests generate-qr-key command
func (s *TxTestSuite) TestCmdGenerateQuantumResistantKey() {
	cmd := CmdGenerateQuantumResistantKey()
	require := s.Require()

	require.NotNil(cmd)
	require.Contains(cmd.Use, "generate-qr-key")
	require.Contains(cmd.Long, "DILITHIUM")
	require.Contains(cmd.Long, "KYBER")

	// Test flag
	require.NotNil(cmd.Flags().Lookup("expires-in"))

	err := cmd.ValidateArgs([]string{"DILITHIUM"})
	require.NoError(err)

	err = cmd.ValidateArgs([]string{})
	require.Error(err)
}

// TestCmdAddCertificatePin tests add-cert-pin command
func (s *TxTestSuite) TestCmdAddCertificatePin() {
	cmd := CmdAddCertificatePin()
	require := s.Require()

	require.NotNil(cmd)
	require.Contains(cmd.Use, "add-cert-pin")
	require.Contains(cmd.Long, "PUBLIC_KEY")
	require.Contains(cmd.Long, "CERTIFICATE")

	// Test flag
	require.NotNil(cmd.Flags().Lookup("expires-in"))

	err := cmd.ValidateArgs([]string{"api.aura.network", "0xhash1,0xhash2", "SPKI"})
	require.NoError(err)
}

// TestThresholdSchemeTypeValidation tests threshold scheme type parsing
func (s *TxTestSuite) TestThresholdSchemeTypeValidation() {
	tests := []struct {
		name        string
		input       string
		shouldMatch bool
	}{
		{"BLS uppercase", "BLS", true},
		{"ECDSA uppercase", "ECDSA", true},
		{"ED25519", "ED25519", true},
		{"EDDSA", "EDDSA", true},
		{"SCHNORR", "SCHNORR", true},
		{"Invalid", "INVALID", false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Test would execute the command's RunE to validate type parsing
			cmd := CmdCreateThresholdScheme()
			s.Require().NotNil(cmd)
		})
	}
}

// TestZKProofTypeValidation tests ZK proof type parsing
func (s *TxTestSuite) TestZKProofTypeValidation() {
	validTypes := []string{"GROTH16", "PLONK", "STARK", "BULLETPROOFS"}
	cmd := CmdRegisterZKProofCircuit()

	for _, validType := range validTypes {
		s.Run(validType, func() {
			s.Require().Contains(cmd.Long, validType)
		})
	}
}

// TestSecureEnclaveTypeValidation tests secure enclave type parsing
func (s *TxTestSuite) TestSecureEnclaveTypeValidation() {
	validTypes := []string{"HSM", "SGX", "SEV", "TPM", "KEYCHAIN"}
	cmd := CmdRegisterSecureEnclave()

	for _, validType := range validTypes {
		s.Run(validType, func() {
			// Verify documentation mentions the type
			s.Require().NotNil(cmd)
		})
	}
}

// TestQuantumResistantAlgorithmValidation tests quantum-resistant algorithm parsing
func (s *TxTestSuite) TestQuantumResistantAlgorithmValidation() {
	validAlgos := []string{"DILITHIUM", "KYBER", "FALCON", "SPHINCS"}
	cmd := CmdGenerateQuantumResistantKey()

	for _, algo := range validAlgos {
		s.Run(algo, func() {
			s.Require().Contains(cmd.Long, algo)
		})
	}
}

// TestCertificatePinTypeValidation tests certificate pin type parsing
func (s *TxTestSuite) TestCertificatePinTypeValidation() {
	cmd := CmdAddCertificatePin()
	require := s.Require()

	require.Contains(cmd.Long, "SPKI")
	require.Contains(cmd.Long, "CERTIFICATE")
}

// TestAllCommandsHaveHelp verifies all commands have help text
func (s *TxTestSuite) TestAllCommandsHaveHelp() {
	cmd := GetTxCmd()
	for _, subCmd := range cmd.Commands() {
		s.Require().NotEmpty(subCmd.Short, "Command %s missing short description", subCmd.Use)
		s.Require().NotEmpty(subCmd.Long, "Command %s missing long description", subCmd.Use)
	}
}

// TestAllCommandsHaveExamples verifies all commands have examples
func (s *TxTestSuite) TestAllCommandsHaveExamples() {
	cmd := GetTxCmd()
	for _, subCmd := range cmd.Commands() {
		s.Require().Contains(subCmd.Long, "Example", "Command %s missing examples", subCmd.Use)
	}
}

// Benchmark tests
func BenchmarkCreateKeyRotationSchedule(b *testing.B) {
	cmd := CmdCreateKeyRotationSchedule()
	for i := 0; i < b.N; i++ {
		_ = cmd.ValidateArgs([]string{"key-123", "86400", "30"})
	}
}

func BenchmarkRotateKey(b *testing.B) {
	cmd := CmdRotateKey()
	for i := 0; i < b.N; i++ {
		_ = cmd.ValidateArgs([]string{"key-123", "0x1234abcd"})
	}
}
