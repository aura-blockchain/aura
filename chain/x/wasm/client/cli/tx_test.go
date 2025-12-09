package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/wasm/types"
)

type WasmTxSuite struct {
	suite.Suite
}

func TestWasmTxSuite(t *testing.T) {
	suite.Run(t, new(WasmTxSuite))
}

func (s *WasmTxSuite) TestGetTxCmd() {
	cmd := GetTxCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal(types.ModuleName, cmd.Use)
	require.True(cmd.DisableFlagParsing)

	expected := []string{
		"store",
		"instantiate",
		"execute",
		"migrate",
		"set-admin",
		"clear-admin",
		"authorize-uploader",
		"revoke-uploader",
		"pause-contract",
		"unpause-contract",
		"update-params",
	}

	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, name := range expected {
		require.True(names[name], "expected wasm tx subcommand %s", name)
	}
}

func (s *WasmTxSuite) TestCommandArgValidation() {
	require := s.Require()

	require.NoError(GetCmdStoreCode().ValidateArgs([]string{"contract.wasm"}))
	require.Error(GetCmdStoreCode().ValidateArgs([]string{}))

	require.NoError(GetCmdInstantiateContract().ValidateArgs([]string{"1", "{}"}))
	require.Error(GetCmdInstantiateContract().ValidateArgs([]string{"1"}))

	require.NoError(GetCmdExecuteContract().ValidateArgs([]string{"aura1contract", "{}"}))
	require.Error(GetCmdExecuteContract().ValidateArgs([]string{"aura1contract"}))

	require.NoError(GetCmdMigrateContract().ValidateArgs([]string{"aura1contract", "2", "{}"}))
	require.Error(GetCmdMigrateContract().ValidateArgs([]string{"aura1contract", "2"}))

	require.NoError(GetCmdUpdateAdmin().ValidateArgs([]string{"aura1contract", "aura1newadmin"}))
	require.Error(GetCmdUpdateAdmin().ValidateArgs([]string{"aura1contract"}))

	require.NoError(GetCmdClearAdmin().ValidateArgs([]string{"aura1contract"}))
	require.Error(GetCmdClearAdmin().ValidateArgs([]string{}))
}
