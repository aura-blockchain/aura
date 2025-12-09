package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/wasm/types"
)

type WasmQuerySuite struct {
	suite.Suite
}

func TestWasmQuerySuite(t *testing.T) {
	suite.Run(t, new(WasmQuerySuite))
}

func (s *WasmQuerySuite) TestGetQueryCmd() {
	cmd := GetQueryCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal(types.ModuleName, cmd.Use)
	require.True(cmd.DisableFlagParsing)

	expected := []string{
		"params",
		"code",
		"list-code",
		"contract",
		"contract-state-all",
		"contract-history",
		"query-smart",
		"query-raw",
		"security-stats",
		"authorized-uploaders",
		"paused-contracts",
		"is-authorized",
		"is-paused",
	}

	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, name := range expected {
		require.True(names[name], "expected query command %s", name)
	}
}

func (s *WasmQuerySuite) TestCommandArgValidation() {
	require := s.Require()

	require.NoError(GetCmdQueryParams().ValidateArgs([]string{}))
	require.Error(GetCmdQueryParams().ValidateArgs([]string{"extra"}))

	require.NoError(GetCmdQueryCode().ValidateArgs([]string{"1"}))
	require.Error(GetCmdQueryCode().ValidateArgs([]string{}))

	require.NoError(GetCmdListCode().ValidateArgs([]string{}))
	require.Error(GetCmdListCode().ValidateArgs([]string{"extra"}))

	require.NoError(GetCmdQueryContractInfo().ValidateArgs([]string{"aura1contract"}))
	require.Error(GetCmdQueryContractInfo().ValidateArgs([]string{}))
}
