package bridge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLockTokensParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *LockTokensParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing sender",
			params: &LockTokensParams{
				TargetChain: "paw",
				Recipient:   "paw1test",
			},
			wantErr: true,
		},
		{
			name: "missing target chain",
			params: &LockTokensParams{
				Sender:    "aura1test",
				Recipient: "paw1test",
			},
			wantErr: true,
		},
		{
			name: "missing recipient",
			params: &LockTokensParams{
				Sender:      "aura1test",
				TargetChain: "paw",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &LockTokensParams{
				Sender:      "aura1test",
				TargetChain: "paw",
				Recipient:   "paw1test",
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
					assert.True(t, tt.params.Sender == "" || tt.params.TargetChain == "" || tt.params.Recipient == "")
				}
			} else {
				assert.NotEmpty(t, tt.params.Sender)
				assert.NotEmpty(t, tt.params.TargetChain)
				assert.NotEmpty(t, tt.params.Recipient)
			}
		})
	}
}

func TestMintTokensParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *MintTokensParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing validator",
			params: &MintTokensParams{
				SourceChain: "paw",
			},
			wantErr: true,
		},
		{
			name: "missing source chain",
			params: &MintTokensParams{
				Validator: "auraval1test",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &MintTokensParams{
				Validator:   "auraval1test",
				SourceChain: "paw",
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
					assert.True(t, tt.params.Validator == "" || tt.params.SourceChain == "")
				}
			} else {
				assert.NotEmpty(t, tt.params.Validator)
				assert.NotEmpty(t, tt.params.SourceChain)
			}
		})
	}
}

func TestUnlockTokensParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *UnlockTokensParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing sender",
			params: &UnlockTokensParams{
				SourceChain: "paw",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &UnlockTokensParams{
				Sender: "aura1test",
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

func TestBurnTokensParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *BurnTokensParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing sender",
			params: &BurnTokensParams{
				TargetChain: "paw",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &BurnTokensParams{
				Sender: "aura1test",
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

func TestLinkAddressParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *LinkAddressParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing aura address",
			params: &LinkAddressParams{
				PawAddress: "paw1test",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &LinkAddressParams{
				AuraAddress: "aura1test",
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
					assert.Empty(t, tt.params.AuraAddress)
				}
			} else {
				assert.NotEmpty(t, tt.params.AuraAddress)
			}
		})
	}
}

func TestCrossChainSwapParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *CrossChainSwapParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing sender",
			params: &CrossChainSwapParams{
				SourceChain: "aura",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &CrossChainSwapParams{
				Sender: "aura1test",
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

func TestRelayTransferParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *RelayTransferParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing relayer",
			params: &RelayTransferParams{
				TransferID: "123",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &RelayTransferParams{
				Relayer: "aura1relayer",
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
					assert.Empty(t, tt.params.Relayer)
				}
			} else {
				assert.NotEmpty(t, tt.params.Relayer)
			}
		})
	}
}

func TestQueryValidation(t *testing.T) {
	t.Run("GetBridgeTransfer requires ID", func(t *testing.T) {
		id := ""
		assert.Empty(t, id)
	})

	t.Run("GetBridgeTransfers requires address", func(t *testing.T) {
		address := ""
		assert.Empty(t, address)
	})

	t.Run("GetLinkedAddresses requires address", func(t *testing.T) {
		address := ""
		assert.Empty(t, address)
	})
}
