package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupTestContext(t *testing.T) {
	ctx := SetupTestContext(t)
	require.NotNil(t, ctx)
	require.NotNil(t, ctx.Ctx)
	require.NotNil(t, ctx.SdkCtx)
	require.NotNil(t, ctx.DB)
	require.NotNil(t, ctx.Logger)
}

func TestCreateTestCodec(t *testing.T) {
	codec := CreateTestCodec()
	require.NotNil(t, codec)
}

func TestMockTime(t *testing.T) {
	mockTime := MockTime()
	require.Equal(t, 2025, mockTime.Year())
	require.Equal(t, 1, int(mockTime.Month()))
	require.Equal(t, 1, mockTime.Day())
}

func TestGenerateTestAddress(t *testing.T) {
	addr := GenerateTestAddress()
	require.NotNil(t, addr)
	require.NotEmpty(t, addr.String())
}

func TestGenerateTestAddresses(t *testing.T) {
	addrs := GenerateTestAddresses(5)
	require.Len(t, addrs, 5)
	for _, addr := range addrs {
		require.NotNil(t, addr)
	}
}
