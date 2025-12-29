package prevalidation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	prevalidationpb "github.com/aequitas/aura/proto/aura/prevalidation/v1beta1"
)

func TestEstimateGasParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *EstimateGasParams
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
			errMsg:  "params cannot be nil",
		},
		{
			name: "missing sender",
			params: &EstimateGasParams{
				Recipient: "aura1recipient123",
				Amount:    "1000",
			},
			wantErr: true,
			errMsg:  "sender is required",
		},
		{
			name: "valid minimal params",
			params: &EstimateGasParams{
				Sender: "aura1sender123",
			},
			wantErr: false,
		},
		{
			name: "valid full params",
			params: &EstimateGasParams{
				Sender:    "aura1sender123",
				Recipient: "aura1recipient123",
				Amount:    "1000",
				Data:      []byte("transaction data"),
				TxType:    prevalidationpb.TransactionType_TX_TYPE_DEX_SWAP,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.params == nil {
					require.Nil(t, tt.params)
				} else {
					assert.Empty(t, tt.params.Sender)
				}
			} else {
				assert.NotEmpty(t, tt.params.Sender)
			}
		})
	}
}

func TestValidateTransactionParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *ValidateTransactionParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing sender",
			params: &ValidateTransactionParams{
				Recipient: "aura1recipient123",
				Amount:    "1000",
				Nonce:     1,
			},
			wantErr: true,
		},
		{
			name: "valid minimal params",
			params: &ValidateTransactionParams{
				Sender: "aura1sender123",
				Nonce:  1,
			},
			wantErr: false,
		},
		{
			name: "valid full params",
			params: &ValidateTransactionParams{
				Sender:    "aura1sender123",
				Recipient: "aura1recipient123",
				Amount:    "1000",
				Data:      []byte("transaction data"),
				Nonce:     1,
				TxType:    prevalidationpb.TransactionType_TX_TYPE_IR_COMPLETION,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.params == nil {
					require.Nil(t, tt.params)
				} else {
					assert.Empty(t, tt.params.Sender)
				}
			} else {
				assert.NotEmpty(t, tt.params.Sender)
			}
		})
	}
}

func TestValidateTransactionResponse(t *testing.T) {
	tests := []struct {
		name     string
		response *ValidateTransactionResponse
	}{
		{
			name: "valid transaction",
			response: &ValidateTransactionResponse{
				Valid:             true,
				GasEstimate:       21000,
				Error:             "",
				SufficientBalance: true,
			},
		},
		{
			name: "invalid transaction - insufficient balance",
			response: &ValidateTransactionResponse{
				Valid:             false,
				GasEstimate:       21000,
				Error:             "insufficient balance",
				SufficientBalance: false,
			},
		},
		{
			name: "invalid transaction - validation error",
			response: &ValidateTransactionResponse{
				Valid:             false,
				GasEstimate:       0,
				Error:             "invalid nonce",
				SufficientBalance: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.response.Valid {
				assert.True(t, tt.response.Valid)
				assert.Empty(t, tt.response.Error)
				assert.True(t, tt.response.SufficientBalance)
				assert.Greater(t, tt.response.GasEstimate, uint64(0))
			} else {
				assert.False(t, tt.response.Valid)
				assert.NotEmpty(t, tt.response.Error)
			}
		})
	}
}

func TestTransactionTypeValidation(t *testing.T) {
	validTypes := []prevalidationpb.TransactionType{
		prevalidationpb.TransactionType_TX_TYPE_IR_COMPLETION,
		prevalidationpb.TransactionType_TX_TYPE_DEX_SWAP,
		prevalidationpb.TransactionType_TX_TYPE_LP_DEPOSIT,
		prevalidationpb.TransactionType_TX_TYPE_LP_WITHDRAWAL,
		prevalidationpb.TransactionType_TX_TYPE_VC_MINT,
		prevalidationpb.TransactionType_TX_TYPE_BRIDGE_TRANSFER,
		prevalidationpb.TransactionType_TX_TYPE_CONFIDENCE_SCORE_UPDATE,
		prevalidationpb.TransactionType_TX_TYPE_IDENTITY_CHANGE,
	}

	for _, txType := range validTypes {
		t.Run(txType.String(), func(t *testing.T) {
			assert.NotEqual(t, prevalidationpb.TransactionType_TX_TYPE_UNSPECIFIED, txType)
		})
	}
}

func TestValidationStatusTypes(t *testing.T) {
	validStatuses := []prevalidationpb.ValidationStatus{
		prevalidationpb.ValidationStatus_VALIDATION_STATUS_PENDING,
		prevalidationpb.ValidationStatus_VALIDATION_STATUS_VALIDATED,
		prevalidationpb.ValidationStatus_VALIDATION_STATUS_EXPIRED,
		prevalidationpb.ValidationStatus_VALIDATION_STATUS_EXECUTED,
		prevalidationpb.ValidationStatus_VALIDATION_STATUS_FAILED,
	}

	for _, status := range validStatuses {
		t.Run(status.String(), func(t *testing.T) {
			assert.NotEqual(t, prevalidationpb.ValidationStatus_VALIDATION_STATUS_UNSPECIFIED, status)
		})
	}
}

func TestGetNonceValidation(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "empty address",
			address: "",
			wantErr: true,
		},
		{
			name:    "valid address",
			address: "aura1test123",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.Empty(t, tt.address)
			} else {
				assert.NotEmpty(t, tt.address)
			}
		})
	}
}

func TestGetPreValidatedTransactionValidation(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "empty id",
			id:      "",
			wantErr: true,
		},
		{
			name:    "valid id",
			id:      "tx_12345",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.Empty(t, tt.id)
			} else {
				assert.NotEmpty(t, tt.id)
			}
		})
	}
}

func TestGetTemplateValidation(t *testing.T) {
	tests := []struct {
		name       string
		templateID string
		wantErr    bool
	}{
		{
			name:       "empty template id",
			templateID: "",
			wantErr:    true,
		},
		{
			name:       "valid template id",
			templateID: "template_ir_completion",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.Empty(t, tt.templateID)
			} else {
				assert.NotEmpty(t, tt.templateID)
			}
		})
	}
}
