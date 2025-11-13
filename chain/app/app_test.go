package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAppRegistersServices(t *testing.T) {
	app := NewApp()
	info := app.GRPCServer().GetServiceInfo()

	require.Contains(t, info, "aura.identitychange.v1beta1.Msg")
	require.Contains(t, info, "aura.identitychange.v1beta1.Query")
}

func TestNewCosmosAppExposesBaseApp(t *testing.T) {
	cApp := NewCosmosApp(nil)
	require.NotNil(t, cApp.BaseApp)
	require.NotNil(t, cApp.Encoding().InterfaceRegistry)
	info := cApp.GRPCServer().GetServiceInfo()
	require.Contains(t, info, "aura.identitychange.v1beta1.Msg")
	require.Contains(t, info, "aura.identitychange.v1beta1.Query")
}
