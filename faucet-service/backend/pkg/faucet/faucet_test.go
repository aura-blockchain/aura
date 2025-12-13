package faucet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aura-chain/aura/faucet/pkg/config"
)

func TestValidateAddress(t *testing.T) {
	initSDKConfig()

	cfg := &config.Config{
		NodeRPC:          "http://localhost:26657",
		NodeAPI:          "http://localhost:1317",
		NodeGRPC:         "localhost:9090",
		ChainID:          "test-chain",
		FaucetAddress:    "aura1test",
		AmountPerRequest: 100,
	}

	service := &Service{
		cfg: cfg,
	}

	validAddr := sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String()

	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "valid address",
			address: validAddr,
			wantErr: false,
		},
		{
			name:    "too short",
			address: "aura1short",
			wantErr: true,
		},
		{
			name:    "wrong prefix",
			address: "cosmos1qwertyuiopasdfghjklzxcvbnm123456789test",
			wantErr: true,
		},
		{
			name:    "empty address",
			address: "",
			wantErr: true,
		},
		{
			name:    "too long",
			address: "aura1" + string(make([]byte, 100)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateAddress(tt.address)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	t.Skip("NewService requires live infrastructure and a real database connection")
}
