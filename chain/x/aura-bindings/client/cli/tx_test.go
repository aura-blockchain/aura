package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/aura-bindings/types"
)

type TxTestSuite struct {
	suite.Suite
}

func TestTxTestSuite(t *testing.T) {
	suite.Run(t, new(TxTestSuite))
}

// Test command registration
func (s *TxTestSuite) TestGetTxCmd() {
	cmd := GetTxCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal(types.ModuleName, cmd.Use)
	require.Equal("AuraBindings transaction subcommands", cmd.Short)

	// The aura-bindings module doesn't have user-facing transaction commands
	// as it primarily serves as a bridge for CosmWasm custom bindings
	subCmds := cmd.Commands()
	require.Equal(0, len(subCmds), "aura-bindings should have no tx subcommands")
}

// Test parent command validates subcommands
func (s *TxTestSuite) TestTxCmd_ValidateCmd() {
	cmd := GetTxCmd()
	require := s.Require()

	require.NotNil(cmd.RunE)
}

// Test module name is correct
func (s *TxTestSuite) TestModuleName() {
	require := s.Require()
	require.Equal("aurabindings", types.ModuleName)
}

// Benchmark tests
func BenchmarkGetTxCmd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cmd := GetTxCmd()
		_ = cmd.Use
	}
}
