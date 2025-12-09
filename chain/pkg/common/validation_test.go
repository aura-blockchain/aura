package common_test

import (
	"testing"

	"github.com/aequitas/aura/chain/pkg/common"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestValidateAddress(t *testing.T) {
	// Initialize SDK config with aura prefix
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("aura", "aurapub")
	config.Seal()

	tests := []struct {
		name      string
		addr      string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid address",
			addr:    "aura1qyqszqgpqyqszqgpqyqszqgpqyqszqgpawv63j",
			wantErr: false,
		},
		{
			name:      "empty address",
			addr:      "",
			wantErr:   true,
			errSubstr: "cannot be empty",
		},
		{
			name:      "invalid bech32",
			addr:      "invalid_address",
			wantErr:   true,
			errSubstr: "invalid address format",
		},
		{
			name:      "wrong prefix",
			addr:      "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpq4qa56p",
			wantErr:   true,
			errSubstr: "invalid address format",
		},
		{
			name:      "invalid checksum",
			addr:      "aura1qyqszqgpqyqszqgpqyqszqgpqyqszqgpxxxxxx",
			wantErr:   true,
			errSubstr: "invalid address format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := common.ValidateAddress(tt.addr)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errSubstr != "" {
					require.Contains(t, err.Error(), tt.errSubstr)
				}
				require.Nil(t, addr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, addr)
				require.Equal(t, tt.addr, addr.String())
			}
		})
	}
}

func TestValidateAddresses(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("aura", "aurapub")
	config.Seal()

	validAddr1 := "aura1qyqszqgpqyqszqgpqyqszqgpqyqszqgpawv63j"
	validAddr2 := "aura1qyqszqgpqyqszqgpqyqszqgpqyqszqgp3jk5u2"
	invalidAddr := "invalid"

	tests := []struct {
		name      string
		addrs     []string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid addresses",
			addrs:   []string{validAddr1, validAddr2},
			wantErr: false,
		},
		{
			name:      "empty list",
			addrs:     []string{},
			wantErr:   true,
			errSubstr: "cannot be empty",
		},
		{
			name:      "one invalid address",
			addrs:     []string{validAddr1, invalidAddr},
			wantErr:   true,
			errSubstr: "invalid address at index 1",
		},
		{
			name:      "first address invalid",
			addrs:     []string{invalidAddr, validAddr1},
			wantErr:   true,
			errSubstr: "invalid address at index 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addrs, err := common.ValidateAddresses(tt.addrs)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errSubstr != "" {
					require.Contains(t, err.Error(), tt.errSubstr)
				}
				require.Nil(t, addrs)
			} else {
				require.NoError(t, err)
				require.Len(t, addrs, len(tt.addrs))
				for i, addr := range addrs {
					require.Equal(t, tt.addrs[i], addr.String())
				}
			}
		})
	}
}

func TestMustValidateAddress(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("aura", "aurapub")
	config.Seal()

	validAddr := "aura1qyqszqgpqyqszqgpqyqszqgpqyqszqgpawv63j"

	t.Run("valid address", func(t *testing.T) {
		addr := common.MustValidateAddress(validAddr)
		require.NotNil(t, addr)
		require.Equal(t, validAddr, addr.String())
	})

	t.Run("invalid address panics", func(t *testing.T) {
		require.Panics(t, func() {
			common.MustValidateAddress("invalid")
		})
	})

	t.Run("empty address panics", func(t *testing.T) {
		require.Panics(t, func() {
			common.MustValidateAddress("")
		})
	})
}
