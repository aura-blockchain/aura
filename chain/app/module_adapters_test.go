package app

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type moduleInitNoErr struct {
	initCalled  bool
	beginCalled bool
	endCalled   bool
}

func (m *moduleInitNoErr) InitGenesis(_ sdk.Context, _ codec.JSONCodec, _ json.RawMessage) {
	m.initCalled = true
}

func (m *moduleInitNoErr) BeginBlock(_ sdk.Context) {
	m.beginCalled = true
}

func (m *moduleInitNoErr) EndBlock(_ sdk.Context) {
	m.endCalled = true
}

type moduleInitWithError struct {
	err error
}

func (m moduleInitWithError) InitGenesis(_ sdk.Context, _ codec.JSONCodec, _ json.RawMessage) error {
	return m.err
}

func (m moduleInitWithError) BeginBlock(_ sdk.Context) error {
	return m.err
}

func (m moduleInitWithError) EndBlock(_ sdk.Context) error {
	return m.err
}

type moduleExportCodec struct {
	exported bool
}

func (m *moduleExportCodec) ExportGenesis(_ sdk.Context, _ codec.JSONCodec) json.RawMessage {
	m.exported = true
	return json.RawMessage(`{"ok":true}`)
}

// makeTestCodec provides a JSONCodec for adapter assertions without wiring the full app.
func makeTestCodec() codec.JSONCodec {
	reg := codectypes.NewInterfaceRegistry()
	return codec.NewProtoCodec(reg)
}

func TestAdapterInitGenesisDelegates(t *testing.T) {
	cdc := makeTestCodec()
	module := &moduleInitNoErr{}
	adapter := adapterModule{name: "walletsecurity", module: module}

	_, err := adapter.InitGenesis(sdk.Context{}, cdc, json.RawMessage(`{}`))
	require.NoError(t, err)
	require.True(t, module.initCalled, "InitGenesis should be invoked")
	require.NoError(t, adapter.BeginBlock(sdk.Context{}))
	require.True(t, module.beginCalled, "BeginBlock should be invoked")
	require.NoError(t, adapter.EndBlock(sdk.Context{}))
	require.True(t, module.endCalled, "EndBlock should be invoked")
}

func TestAdapterInitGenesisErrorPropagates(t *testing.T) {
	cdc := makeTestCodec()
	expectedErr := errors.New("boom")
	module := moduleInitWithError{err: expectedErr}
	adapter := adapterModule{name: "failing", module: module}

	_, err := adapter.InitGenesis(sdk.Context{}, cdc, json.RawMessage(`{}`))
	require.EqualError(t, err, expectedErr.Error())
	require.EqualError(t, adapter.BeginBlock(sdk.Context{}), expectedErr.Error())
	require.EqualError(t, adapter.EndBlock(sdk.Context{}), expectedErr.Error())
}

func TestAdapterExportGenesisUsesCodecPath(t *testing.T) {
	cdc := makeTestCodec()
	module := &moduleExportCodec{}
	adapter := adapterModule{name: "exporter", module: module}

	exported, err := adapter.ExportGenesis(sdk.Context{}, cdc)
	require.NoError(t, err)
	require.True(t, module.exported, "ExportGenesis should be invoked")
	require.JSONEq(t, `{"ok":true}`, string(exported))
}
