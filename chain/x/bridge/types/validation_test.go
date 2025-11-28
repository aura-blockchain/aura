package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateBool(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		err   error
	}{
		{
			name:  "valid true",
			input: true,
			err:   nil,
		},
		{
			name:  "valid false",
			input: false,
			err:   nil,
		},
		{
			name:  "invalid string",
			input: "true",
			err:   ErrInvalidParam,
		},
		{
			name:  "invalid int",
			input: 1,
			err:   ErrInvalidParam,
		},
		{
			name:  "invalid nil",
			input: nil,
			err:   ErrInvalidParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBool(tt.input)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()
	require.True(t, params.BridgeEnabled)
}

func TestParamSetPairs(t *testing.T) {
	params := DefaultParams()
	pairs := params.ParamSetPairs()
	require.NotNil(t, pairs)
	require.Len(t, pairs, 5)

	expectedKeys := [][]byte{
		KeyBridgeEnabled,
		KeyMinConfirmations,
		KeyBridgeFeeBasisPoints,
		KeyCoreMaxTransferAmount,
		KeyValidatorThresholdPercentage,
	}

	for _, key := range expectedKeys {
		found := false
		for _, pair := range pairs {
			if string(pair.Key) == string(key) {
				found = true
				break
			}
		}
		require.Truef(t, found, "key %s not found in param set pairs", key)
	}
}

func TestDefaultGenesis(t *testing.T) {
	genesis := DefaultGenesis()
	require.NotNil(t, genesis)
	require.NotNil(t, genesis.Params)
	require.True(t, genesis.Params.Enabled)
}

func TestBridgeParamsStruct(t *testing.T) {
	params := &BridgeParams{
		Enabled:                      true,
		MinConfirmations:             12,
		BridgeFeeBasisPoints:         30,
		MaxTransferAmount:            "1000000",
		ValidatorThresholdPercentage: 67,
	}

	require.True(t, params.Enabled)
	require.Greater(t, params.MinConfirmations, uint64(0))
	require.Greater(t, params.BridgeFeeBasisPoints, uint64(0))
	require.NotEmpty(t, params.MaxTransferAmount)
	require.Greater(t, params.ValidatorThresholdPercentage, uint64(0))
}

func TestCrossChainTransferStruct(t *testing.T) {
	transfer := &CrossChainTransfer{
		TransferId:  "transfer-1",
		SourceChain: "aura",
		TargetChain: "paw",
		Sender:      "aura1sender",
		Recipient:   "paw1recipient",
		Amount:      "1000000",
		Denom:       "uaura",
		Status:      TransferStatus_PENDING,
	}

	require.NotEmpty(t, transfer.TransferId)
	require.NotEmpty(t, transfer.SourceChain)
	require.NotEmpty(t, transfer.TargetChain)
	require.NotEmpty(t, transfer.Sender)
	require.NotEmpty(t, transfer.Recipient)
	require.NotEmpty(t, transfer.Amount)
	require.NotEmpty(t, transfer.Denom)
}

func TestTransferStatusEnums(t *testing.T) {
	// Test that enum constants are properly exported
	require.Equal(t, TransferStatus_PENDING, TransferStatus(0))
	require.NotEqual(t, TransferStatus_PENDING, TransferStatus_CONFIRMED)
	require.NotEqual(t, TransferStatus_CONFIRMED, TransferStatus_COMPLETED)
}

func TestFeeTypeEnums(t *testing.T) {
	// Test that fee type enums are properly exported
	require.NotEqual(t, FeeType_FEE_TRANSFER, FeeType_FEE_MINT_WRAPPED)
	require.NotEqual(t, FeeType_FEE_MINT_WRAPPED, FeeType_FEE_BURN_WRAPPED)
}

func TestCircuitBreakerStatusEnums(t *testing.T) {
	// Test circuit breaker status enums
	require.NotEqual(t, CircuitBreakerStatus_CIRCUIT_CLOSED, CircuitBreakerStatus_CIRCUIT_OPEN)
	require.NotEqual(t, CircuitBreakerStatus_CIRCUIT_OPEN, CircuitBreakerStatus_CIRCUIT_HALF_OPEN)
}
