package app

import (
	"testing"

	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/stretchr/testify/require"
)

// TestValidateAnteHandlerOptions tests the validation logic for ante handler options
func TestValidateAnteHandlerOptions(t *testing.T) {
	tests := []struct {
		name        string
		options     HandlerOptions
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing sign mode handler",
			options: HandlerOptions{
				SignModeHandler: nil,
			},
			expectError: true,
			errorMsg:    "sign mode handler is required",
		},
		{
			name: "valid without wasm",
			options: HandlerOptions{
				SignModeHandler: &txsigning.HandlerMap{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
