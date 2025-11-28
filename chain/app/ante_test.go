package app

import (
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/require"
)

func TestNewAnteHandler(t *testing.T) {
	tests := []struct {
		name        string
		options     HandlerOptions
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid options",
			options: HandlerOptions{
				AccountKeeper:   mockAccountKeeper{},
				BankKeeper:      mockBankKeeper{},
				SignModeHandler: mockSignModeHandler{},
			},
			expectError: false,
		},
		{
			name: "missing account keeper",
			options: HandlerOptions{
				BankKeeper:      mockBankKeeper{},
				SignModeHandler: mockSignModeHandler{},
			},
			expectError: true,
			errorMsg:    "account keeper is required",
		},
		{
			name: "missing bank keeper",
			options: HandlerOptions{
				AccountKeeper:   mockAccountKeeper{},
				SignModeHandler: mockSignModeHandler{},
			},
			expectError: true,
			errorMsg:    "bank keeper is required",
		},
		{
			name: "missing sign mode handler",
			options: HandlerOptions{
				AccountKeeper: mockAccountKeeper{},
				BankKeeper:    mockBankKeeper{},
			},
			expectError: true,
			errorMsg:    "sign mode handler is required",
		},
		{
			name: "wasm keeper without config",
			options: HandlerOptions{
				AccountKeeper:   mockAccountKeeper{},
				BankKeeper:      mockBankKeeper{},
				SignModeHandler: mockSignModeHandler{},
				WasmKeeper:      &mockWasmKeeper{},
			},
			expectError: true,
			errorMsg:    "wasm config is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewAnteHandler(tt.options)

			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMsg)
				require.Nil(t, handler)
			} else {
				require.NoError(t, err)
				require.NotNil(t, handler)
			}
		})
	}
}

func TestValidateAnteHandlerOptions(t *testing.T) {
	tests := []struct {
		name        string
		options     HandlerOptions
		expectError bool
	}{
		{
			name: "valid basic options",
			options: HandlerOptions{
				AccountKeeper:   mockAccountKeeper{},
				BankKeeper:      mockBankKeeper{},
				SignModeHandler: mockSignModeHandler{},
			},
			expectError: false,
		},
		{
			name: "valid with wasm",
			options: HandlerOptions{
				AccountKeeper:     mockAccountKeeper{},
				BankKeeper:        mockBankKeeper{},
				SignModeHandler:   mockSignModeHandler{},
				WasmKeeper:        &mockWasmKeeper{},
				WasmConfig:        wasmtypes.DefaultWasmConfig(),
				TXCounterStoreKey: storetypes.NewKVStoreKey("wasm"),
			},
			expectError: false,
		},
		{
			name: "missing account keeper",
			options: HandlerOptions{
				BankKeeper:      mockBankKeeper{},
				SignModeHandler: mockSignModeHandler{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAnteHandlerOptions(tt.options)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Mock types for testing
type mockAccountKeeper struct{}
type mockBankKeeper struct{}
type mockSignModeHandler struct{}
type mockWasmKeeper struct{}
