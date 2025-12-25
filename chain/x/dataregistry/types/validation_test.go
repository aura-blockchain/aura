// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	require.Equal(t, uint64(1000), params.MaxDataItemsPerUser)
	require.Equal(t, uint64(10485760), params.MaxStorageBytes) // 10 MB
	require.Equal(t, "100", params.StorageFee)
	require.Equal(t, uint64(10), params.VerificationReward)
	require.NotNil(t, params.AuthorizedVerifiers)
	require.Len(t, params.AuthorizedVerifiers, 0)
}

func TestValidateParams(t *testing.T) {
	tests := []struct {
		name    string
		params  *Params
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid params",
			params: &Params{
				MaxDataItemsPerUser: 1000,
				MaxStorageBytes:     10485760,
				StorageFee:          "100",
				VerificationReward:  10,
				AuthorizedVerifiers: []string{},
			},
			wantErr: false,
		},
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
			errMsg:  "params cannot be nil",
		},
		{
			name: "zero max data items",
			params: &Params{
				MaxDataItemsPerUser: 0,
				MaxStorageBytes:     10485760,
				StorageFee:          "100",
				VerificationReward:  10,
			},
			wantErr: true,
			errMsg:  "max_data_items_per_user must be greater than 0",
		},
		{
			name: "zero max storage",
			params: &Params{
				MaxDataItemsPerUser: 1000,
				MaxStorageBytes:     0,
				StorageFee:          "100",
				VerificationReward:  10,
			},
			wantErr: true,
			errMsg:  "max_storage_bytes must be greater than 0",
		},
		{
			name: "empty storage fee",
			params: &Params{
				MaxDataItemsPerUser: 1000,
				MaxStorageBytes:     10485760,
				StorageFee:          "",
				VerificationReward:  10,
			},
			wantErr: true,
			errMsg:  "storage_fee cannot be empty",
		},
		{
			name: "zero verification reward is allowed",
			params: &Params{
				MaxDataItemsPerUser: 1000,
				MaxStorageBytes:     10485760,
				StorageFee:          "100",
				VerificationReward:  0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateParams(tt.params)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRegistryStats(t *testing.T) {
	stats := &RegistryStats{
		TotalDataItems:     100,
		TotalVerifiedItems: 80,
		TotalVerifications: 150,
		TotalStorageBytes:  1024,
		ItemsByType: map[string]uint64{
			"photo":    50,
			"document": 30,
			"video":    20,
		},
	}

	require.Equal(t, uint64(100), stats.TotalDataItems)
	require.Equal(t, uint64(80), stats.TotalVerifiedItems)
	require.Equal(t, uint64(150), stats.TotalVerifications)
	require.Equal(t, uint64(1024), stats.TotalStorageBytes)
	require.NotNil(t, stats.ItemsByType)
	require.Len(t, stats.ItemsByType, 3)
	require.Equal(t, uint64(50), stats.ItemsByType["photo"])
}
