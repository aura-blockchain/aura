// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"
)

func nowTimestamp() *types.Timestamp {
	now := time.Now()
	return &types.Timestamp{Seconds: now.Unix(), Nanos: int32(now.Nanosecond())}
}

func pastTimestamp() *types.Timestamp {
	past := time.Now().Add(-24 * time.Hour)
	return &types.Timestamp{Seconds: past.Unix(), Nanos: int32(past.Nanosecond())}
}

func futureTimestamp() *types.Timestamp {
	future := time.Now().Add(24 * time.Hour)
	return &types.Timestamp{Seconds: future.Unix(), Nanos: int32(future.Nanosecond())}
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
		{
			name: "expires_at before created_at",
			state: func() *GenesisState {
				return &GenesisState{
					DataItems: []*DataItem{{
						DataId: "id-1", OwnerAddress: "owner",
						DataType:     DataItemType_DATA_ITEM_TYPE_PHOTO,
						ContentHash:  []byte("hash123"),
						CreatedAt:    nowTimestamp(),
						ExpiresAt:    pastTimestamp(), // before created_at
						AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PRIVATE},
					}},
				}
			},
			wantErr: "expires_at is before created_at",
		},
		{
			name: "nil verification in list",
			state: func() *GenesisState {
				return &GenesisState{
					DataItems: []*DataItem{{
						DataId:       "id-1",
						OwnerAddress: "owner",
						DataType:     DataItemType_DATA_ITEM_TYPE_PHOTO,
						ContentHash:  []byte("hash123"),
						CreatedAt:    nowTimestamp(),
						AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PRIVATE},
						Verifications: []*Verification{
							nil,
						},
					}},
				}
			},
			wantErr: "verification at index 0 for data item id-1 is nil",
		},
		{
			name: "empty verifier in verification",
			state: func() *GenesisState {
				return &GenesisState{
					DataItems: []*DataItem{{
						DataId:       "id-1",
						OwnerAddress: "owner",
						DataType:     DataItemType_DATA_ITEM_TYPE_PHOTO,
						ContentHash:  []byte("hash123"),
						CreatedAt:    nowTimestamp(),
						AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PRIVATE},
						Verifications: []*Verification{
							{VerifierAddress: "", Level: VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED},
						},
					}},
				}
			},
			wantErr: "has empty verifier",
		},
		{
			name: "duplicate verifier",
			state: func() *GenesisState {
				ts := nowTimestamp()
				return &GenesisState{
					DataItems: []*DataItem{{
						DataId:       "id-1",
						OwnerAddress: "owner",
						DataType:     DataItemType_DATA_ITEM_TYPE_PHOTO,
						ContentHash:  []byte("hash123"),
						CreatedAt:    ts,
						AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PRIVATE},
						Verifications: []*Verification{
							{VerifierAddress: "verifier1", Level: VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, VerifiedAt: ts},
							{VerifierAddress: "verifier1", Level: VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, VerifiedAt: ts},
						},
					}},
				}
			},
			wantErr: "duplicate verifier verifier1",
		},
		{
			name: "unspecified verification level",
			state: func() *GenesisState {
				ts := nowTimestamp()
				return &GenesisState{
					DataItems: []*DataItem{{
						DataId:       "id-1",
						OwnerAddress: "owner",
						DataType:     DataItemType_DATA_ITEM_TYPE_PHOTO,
						ContentHash:  []byte("hash123"),
						CreatedAt:    ts,
						AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PRIVATE},
						Verifications: []*Verification{
							{VerifierAddress: "verifier1", Level: VerificationLevel_VERIFICATION_LEVEL_UNSPECIFIED, VerifiedAt: ts},
						},
					}},
				}
			},
			wantErr: "unspecified level",
		},
		{
			name: "nil verified_at timestamp",
			state: func() *GenesisState {
				return &GenesisState{
					DataItems: []*DataItem{{
						DataId:       "id-1",
						OwnerAddress: "owner",
						DataType:     DataItemType_DATA_ITEM_TYPE_PHOTO,
						ContentHash:  []byte("hash123"),
						CreatedAt:    nowTimestamp(),
						AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PRIVATE},
						Verifications: []*Verification{
							{VerifierAddress: "verifier1", Level: VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, VerifiedAt: nil},
						},
					}},
				}
			},
			wantErr: "nil verified_at timestamp",
		},
		{
			name: "empty metadata key",
			state: func() *GenesisState {
				return &GenesisState{
					DataItems: []*DataItem{{
						DataId:       "id-1",
						OwnerAddress: "owner",
						DataType:     DataItemType_DATA_ITEM_TYPE_PHOTO,
						ContentHash:  []byte("hash123"),
						CreatedAt:    nowTimestamp(),
						AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PRIVATE},
						Metadata:     map[string]string{"": "value"},
					}},
				}
			},
			wantErr: "metadata with empty key",
		},
		{
			name: "valid data item with all fields",
			state: func() *GenesisState {
				ts := nowTimestamp()
				return &GenesisState{
					Params: &Params{
						MaxDataItemsPerUser: 100,
						MaxStorageBytes:     1024 * 1024,
						StorageFee:          "100",
						VerificationReward:  10,
						AuthorizedVerifiers: []string{},
					},
					NextDataId: 5,
					DataItems: []*DataItem{{
						DataId:       "id-1",
						OwnerAddress: "aura1owner",
						DataType:     DataItemType_DATA_ITEM_TYPE_PHOTO,
						ContentHash:  []byte("hash123hash123hash123hash123hash1"),
						CreatedAt:    ts,
						ExpiresAt:    futureTimestamp(),
						AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PUBLIC},
						Metadata:     map[string]string{"key1": "value1", "key2": "value2"},
						Verifications: []*Verification{
							{
								VerifierAddress: "verifier1",
								Level:           VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
								ConfidenceScore: 90,
								VerifiedAt:      ts,
							},
							{
								VerifierAddress: "verifier2",
								Level:           VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED,
								ConfidenceScore: 100,
								VerifiedAt:      ts,
							},
						},
					}},
				}
			},
		},
		{
			name: "valid with multiple data items",
			state: func() *GenesisState {
				ts := nowTimestamp()
				return &GenesisState{
					DataItems: []*DataItem{
						{
							DataId:       "item-1",
							OwnerAddress: "owner1",
							DataType:     DataItemType_DATA_ITEM_TYPE_PHOTO,
							ContentHash:  []byte("hash1234567890"),
							CreatedAt:    ts,
							AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PRIVATE},
						},
						{
							DataId:       "item-2",
							OwnerAddress: "owner2",
							DataType:     DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
							ContentHash:  []byte("hash0987654321"),
							CreatedAt:    ts,
							AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_WHITELIST},
						},
					},
				}
			},
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

func TestDefaultGenesis(t *testing.T) {
	// Create a JSON codec with interface registry
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	// Test DefaultGenesis returns valid raw JSON
	rawJSON := DefaultGenesis(cdc)
	require.NotNil(t, rawJSON)
	require.NotEmpty(t, rawJSON)

	// Verify the JSON is valid by unmarshaling it
	var state GenesisState
	err := cdc.UnmarshalJSON(rawJSON, &state)
	require.NoError(t, err)

	// Verify it matches the default state
	require.NotNil(t, state.Params)
	require.Equal(t, uint64(1000), state.Params.MaxDataItemsPerUser)
	require.Equal(t, uint64(10485760), state.Params.MaxStorageBytes)
	require.Equal(t, "100", state.Params.StorageFee)
	require.NotNil(t, state.DataItems)
	require.Len(t, state.DataItems, 0)
	require.Equal(t, uint64(1), state.NextDataId)
}

func TestValidateGenesisState_ExpiresAtEdgeCases(t *testing.T) {
	// Test expires_at exactly equal to created_at (should be valid)
	ts := nowTimestamp()
	state := &GenesisState{
		DataItems: []*DataItem{{
			DataId:       "id-1",
			OwnerAddress: "owner",
			DataType:     DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:  []byte("hash123"),
			CreatedAt:    ts,
			ExpiresAt:    ts, // exactly equal should be valid
			AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PRIVATE},
		}},
	}
	err := ValidateGenesisState(state)
	require.NoError(t, err)
}

func TestValidateGenesisState_NilMetadata(t *testing.T) {
	// nil metadata should be valid
	ts := nowTimestamp()
	state := &GenesisState{
		DataItems: []*DataItem{{
			DataId:       "id-1",
			OwnerAddress: "owner",
			DataType:     DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:  []byte("hash123"),
			CreatedAt:    ts,
			AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PRIVATE},
			Metadata:     nil, // nil is valid
		}},
	}
	err := ValidateGenesisState(state)
	require.NoError(t, err)
}

func TestValidateGenesisState_EmptyVerifications(t *testing.T) {
	// empty verifications slice should be valid
	ts := nowTimestamp()
	state := &GenesisState{
		DataItems: []*DataItem{{
			DataId:        "id-1",
			OwnerAddress:  "owner",
			DataType:      DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:   []byte("hash123"),
			CreatedAt:     ts,
			AccessPolicy:  &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PRIVATE},
			Verifications: []*Verification{}, // empty is valid
		}},
	}
	err := ValidateGenesisState(state)
	require.NoError(t, err)
}

func TestValidateGenesisState_AllDataItemTypes(t *testing.T) {
	ts := nowTimestamp()

	// Test multiple data item types are valid
	dataTypes := []DataItemType{
		DataItemType_DATA_ITEM_TYPE_VEHICLE_REGISTRATION,
		DataItemType_DATA_ITEM_TYPE_VEHICLE_INSURANCE,
		DataItemType_DATA_ITEM_TYPE_PROPERTY_DEED,
		DataItemType_DATA_ITEM_TYPE_LEASE_AGREEMENT,
		DataItemType_DATA_ITEM_TYPE_CONTRACT,
		DataItemType_DATA_ITEM_TYPE_RECEIPT,
		DataItemType_DATA_ITEM_TYPE_WARRANTY,
		DataItemType_DATA_ITEM_TYPE_PHOTO,
		DataItemType_DATA_ITEM_TYPE_VIDEO,
		DataItemType_DATA_ITEM_TYPE_AUDIO,
		DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		DataItemType_DATA_ITEM_TYPE_GOLF_SCORE,
		DataItemType_DATA_ITEM_TYPE_TEST_SCORE,
		DataItemType_DATA_ITEM_TYPE_CERTIFICATION,
		DataItemType_DATA_ITEM_TYPE_ACHIEVEMENT,
		DataItemType_DATA_ITEM_TYPE_NFT,
		DataItemType_DATA_ITEM_TYPE_DIGITAL_ART,
		DataItemType_DATA_ITEM_TYPE_MUSIC_LICENSE,
		DataItemType_DATA_ITEM_TYPE_VACCINATION_RECORD,
		DataItemType_DATA_ITEM_TYPE_MEDICAL_RECORD,
		DataItemType_DATA_ITEM_TYPE_PRESCRIPTION,
		DataItemType_DATA_ITEM_TYPE_CUSTOM,
	}

	for i, dt := range dataTypes {
		state := &GenesisState{
			DataItems: []*DataItem{{
				DataId:       "id-" + string(rune('a'+i)),
				OwnerAddress: "owner",
				DataType:     dt,
				ContentHash:  []byte("hash123"),
				CreatedAt:    ts,
				AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PRIVATE},
			}},
		}
		err := ValidateGenesisState(state)
		require.NoError(t, err, "data type %v should be valid", dt)
	}
}

func TestValidateGenesisState_AllVerificationLevels(t *testing.T) {
	ts := nowTimestamp()

	levels := []VerificationLevel{
		VerificationLevel_VERIFICATION_LEVEL_SELF_ATTESTED,
		VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
		VerificationLevel_VERIFICATION_LEVEL_AI_VERIFIED,
		VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED,
		VerificationLevel_VERIFICATION_LEVEL_BLOCKCHAIN_ANCHORED,
	}

	for i, level := range levels {
		state := &GenesisState{
			DataItems: []*DataItem{{
				DataId:       "id-1",
				OwnerAddress: "owner",
				DataType:     DataItemType_DATA_ITEM_TYPE_PHOTO,
				ContentHash:  []byte("hash123"),
				CreatedAt:    ts,
				AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PRIVATE},
				Verifications: []*Verification{{
					VerifierAddress: "verifier" + string(rune('0'+i)),
					Level:           level,
					VerifiedAt:      ts,
				}},
			}},
		}
		err := ValidateGenesisState(state)
		require.NoError(t, err, "verification level %v should be valid", level)
	}
}

func TestValidateGenesisState_AllAccessModes(t *testing.T) {
	ts := nowTimestamp()

	modes := []AccessMode{
		AccessMode_ACCESS_MODE_PRIVATE,
		AccessMode_ACCESS_MODE_WHITELIST,
		AccessMode_ACCESS_MODE_PUBLIC,
		AccessMode_ACCESS_MODE_VERIFIED_USERS,
	}

	for i, mode := range modes {
		state := &GenesisState{
			DataItems: []*DataItem{{
				DataId:       "id-" + string(rune('a'+i)),
				OwnerAddress: "owner",
				DataType:     DataItemType_DATA_ITEM_TYPE_PHOTO,
				ContentHash:  []byte("hash123"),
				CreatedAt:    ts,
				AccessPolicy: &AccessPolicy{Mode: mode},
			}},
		}
		err := ValidateGenesisState(state)
		require.NoError(t, err, "access mode %v should be valid", mode)
	}
}

func TestValidateGenesisState_AllDataItemStatuses(t *testing.T) {
	ts := nowTimestamp()

	statuses := []DataItemStatus{
		DataItemStatus_DATA_ITEM_STATUS_UNSPECIFIED,
		DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
		DataItemStatus_DATA_ITEM_STATUS_VERIFIED,
		DataItemStatus_DATA_ITEM_STATUS_REJECTED,
		DataItemStatus_DATA_ITEM_STATUS_EXPIRED,
		DataItemStatus_DATA_ITEM_STATUS_REVOKED,
	}

	for i, status := range statuses {
		state := &GenesisState{
			DataItems: []*DataItem{{
				DataId:       "id-" + string(rune('a'+i)),
				OwnerAddress: "owner",
				DataType:     DataItemType_DATA_ITEM_TYPE_PHOTO,
				ContentHash:  []byte("hash123"),
				CreatedAt:    ts,
				AccessPolicy: &AccessPolicy{Mode: AccessMode_ACCESS_MODE_PRIVATE},
				Status:       status,
			}},
		}
		err := ValidateGenesisState(state)
		require.NoError(t, err, "data item status %v should be valid", status)
	}
}
