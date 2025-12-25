// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package common_test

import (
	"testing"

	"github.com/aequitas/aura/chain/pkg/common"
	"github.com/stretchr/testify/require"
)

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name      string
		addr      string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid address",
			addr:    "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
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
			name:      "invalid checksum",
			addr:      "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq999999",
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
	validAddr := "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a"
	invalidAddr := "invalid"

	tests := []struct {
		name      string
		addrs     []string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid addresses",
			addrs:   []string{validAddr, validAddr},
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
			addrs:     []string{validAddr, invalidAddr},
			wantErr:   true,
			errSubstr: "invalid address at index 1",
		},
		{
			name:      "first address invalid",
			addrs:     []string{invalidAddr, validAddr},
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
	validAddr := "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a"

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
