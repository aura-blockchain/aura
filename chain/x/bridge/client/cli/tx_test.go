package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTxCmd(t *testing.T) {
	cmd := GetTxCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "bridge", cmd.Use)
	require.True(t, cmd.DisableFlagParsing)
	subcommands := cmd.Commands()
	require.NotZero(t, len(subcommands))
}

func TestCmdLockTokensArgs(t *testing.T) {
	cmd := CmdLockTokens()
	require.NotNil(t, cmd)

	cases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"valid", []string{"paw", "paw1recipient", "100uaura"}, false},
		{"missing args", []string{"paw"}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := cmd.Args(cmd, tc.args)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdUnlockTokensArgs(t *testing.T) {
	cmd := CmdUnlockTokens()
	require.NotNil(t, cmd)

	err := cmd.Args(cmd, []string{"paw", "0xburn", "1000", "uaura"})
	require.NoError(t, err)

	err = cmd.Args(cmd, []string{"paw", "0xburn"})
	require.Error(t, err)
}

func TestCmdCrossChainSwapArgs(t *testing.T) {
	cmd := CmdCrossChainSwap()
	require.NotNil(t, cmd)

	err := cmd.Args(cmd, []string{"aura", "100uaura", "paw", "paw.token", "90"})
	require.NoError(t, err)

	err = cmd.Args(cmd, []string{"aura"})
	require.Error(t, err)
}

func TestCmdRelayTransferArgs(t *testing.T) {
	cmd := CmdRelayTransfer()
	require.NotNil(t, cmd)

	err := cmd.Args(cmd, []string{"transfer-1", "0xhash", "COMPLETED"})
	require.NoError(t, err)

	err = cmd.Args(cmd, []string{"transfer-1"})
	require.Error(t, err)
}
