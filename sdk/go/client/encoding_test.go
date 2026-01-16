package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMakeEncodingConfig(t *testing.T) {
	cfg := MakeEncodingConfig()

	require.NotNil(t, cfg.InterfaceRegistry)
	require.NotNil(t, cfg.Codec)
	require.NotNil(t, cfg.TxConfig)
	require.NotNil(t, cfg.Amino)

	// TxConfig should support default sign modes
	require.NotEmpty(t, cfg.TxConfig.SignModeHandler().DefaultMode())
}
