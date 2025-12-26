// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
	"time"

	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"
)

func nowTimestamp() *types.Timestamp {
	now := time.Now()
	return &types.Timestamp{Seconds: now.Unix(), Nanos: int32(now.Nanosecond())}
}

func TestDefaultGenesisState(t *testing.T) {
	state := DefaultGenesisState()
	require.NotNil(t, state)
	require.NotNil(t, state.Params)
	require.NotNil(t, state.DataItems)
	require.Len(t, state.DataItems, 0)
	require.Equal(t, uint64(1), state.NextDataId)
}

func TestValidateGenesisState(t *testing.T) {
	tests := []struct {
		name    string
		state   func() *GenesisState
		wantErr string
	}{
		{
			name:  "valid default",
			state: DefaultGenesisState,
		},
		{
			name:    "nil state",
			state:   func() *GenesisState { return nil },
			wantErr: "genesis state cannot be nil",
		},
		{
			name: "nil params allowed",
			state: func() *GenesisState {
				return &GenesisState{Params: nil, DataItems: []*DataItem{}}
			},
		},
		{
			name: "invalid params",
			state: func() *GenesisState {
				return &GenesisState{
					Params: &Params{MaxDataItemsPerUser: 0, MaxStorageBytes: 100, StorageFee: "1"},
				}
			},
			wantErr: "max_data_items_per_user",
		},
		{
			name: "nil data item",
			state: func() *GenesisState {
				return &GenesisState{Params: nil, DataItems: []*DataItem{nil}}
			},
			wantErr: "data item at index 0 is nil",
		},
		{
			name: "empty data_id",
			state: func() *GenesisState {
				return &GenesisState{
					Params:    nil,
					DataItems: []*DataItem{{DataId: "", OwnerAddress: "owner"}},
				}
			},
			wantErr: "empty data_id",
		},
		{
			name: "duplicate data_id",
			state: func() *GenesisState {
				ts := nowTimestamp()
				item := &DataItem{
					DataId:       "dup-1",
					OwnerAddress: "owner",
					DataType:     DataItemType_DATA_ITEM_TYPE_PHOTO,
					ContentHash:  []byte("hash"),
					CreatedAt:    ts,
					AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PRIVATE},
				}
				return &GenesisState{DataItems: []*DataItem{item, item}}
			},
			wantErr: "duplicate data item ID",
		},
		{
			name: "empty owner",
			state: func() *GenesisState {
				return &GenesisState{
					DataItems: []*DataItem{{DataId: "id-1", OwnerAddress: ""}},
				}
			},
			wantErr: "empty owner",
		},
		{
			name: "unspecified data type",
			state: func() *GenesisState {
				return &GenesisState{
					DataItems: []*DataItem{{
						DataId: "id-1", OwnerAddress: "owner",
						DataType: DataItemType_DATA_ITEM_TYPE_UNSPECIFIED,
					}},
				}
			},
			wantErr: "unspecified data type",
		},
		{
			name: "empty content hash",
			state: func() *GenesisState {
				return &GenesisState{
					DataItems: []*DataItem{{
						DataId: "id-1", OwnerAddress: "owner",
						DataType: DataItemType_DATA_ITEM_TYPE_PHOTO, ContentHash: []byte{},
					}},
				}
			},
			wantErr: "empty content hash",
		},
		{
			name: "nil created_at",
			state: func() *GenesisState {
				return &GenesisState{
					DataItems: []*DataItem{{
						DataId: "id-1", OwnerAddress: "owner",
						DataType: DataItemType_DATA_ITEM_TYPE_PHOTO, ContentHash: []byte("h"),
						CreatedAt: nil,
					}},
				}
			},
			wantErr: "nil created_at",
		},
		{
			name: "nil access policy",
			state: func() *GenesisState {
				return &GenesisState{
					DataItems: []*DataItem{{
						DataId: "id-1", OwnerAddress: "owner",
						DataType: DataItemType_DATA_ITEM_TYPE_PHOTO, ContentHash: []byte("h"),
						CreatedAt: nowTimestamp(), AccessPolicy: nil,
					}},
				}
			},
			wantErr: "nil access policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGenesisState(tt.state())
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateGenesis_Alias(t *testing.T) {
	require.NoError(t, ValidateGenesis(DefaultGenesisState()))
	require.Error(t, ValidateGenesis(nil))
}
