package keeper

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/dataregistry/params"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
	pb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestInitGenesis(t *testing.T) {
	tests := []struct {
		name    string
		genesis types.GenesisState
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid genesis with data items",
			genesis: types.GenesisState{
				Params: &pb.Params{
					MaxDataSize:          10485760,
					StorageFeePerKb:      "100",
					VerificationFee:      "50",
					MaxRetentionPeriod:   365 * 24 * 3600,
					EnableIpfs:           true,
					RequireVerification:  false,
					AllowPublicAccess:    true,
					MaxVerifiersPerItem:  10,
					MinVerificationLevel: pb.VerificationLevel_VERIFICATION_LEVEL_BASIC,
				},
				DataItems: []*pb.DataItem{
					{
						DataId:      "data-1",
						Owner:       "aura1owner1",
						DataType:    pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
						ContentHash: []byte("hash1"),
						IpfsCid:     "QmTest1",
						Status:      pb.DataItemStatus_DATA_ITEM_STATUS_ACTIVE,
						CreatedAt:   timestamppb.Now(),
						UpdatedAt:   timestamppb.Now(),
						SizeBytes:   1024,
						AccessPolicy: &pb.AccessPolicy{
							Mode: pb.AccessMode_ACCESS_MODE_PUBLIC,
						},
						Metadata:      map[string]string{"key": "value"},
						Tags:          []string{"photo", "test"},
						Verifications: []*pb.Verification{},
					},
					{
						DataId:      "data-2",
						Owner:       "aura1owner2",
						DataType:    pb.DataItemType_DATA_ITEM_TYPE_VEHICLE_REGISTRATION,
						ContentHash: []byte("hash2"),
						IpfsCid:     "QmTest2",
						Status:      pb.DataItemStatus_DATA_ITEM_STATUS_ACTIVE,
						CreatedAt:   timestamppb.Now(),
						UpdatedAt:   timestamppb.Now(),
						SizeBytes:   2048,
						AccessPolicy: &pb.AccessPolicy{
							Mode: pb.AccessMode_ACCESS_MODE_PRIVATE,
						},
						VehicleRegistrationData: &pb.VehicleRegistrationData{
							Vin:         "VIN123456789",
							Make:        "Tesla",
							Model:       "Model 3",
							Year:        2023,
							LicensePlate: "ABC-123",
						},
						Verifications: []*pb.Verification{
							{
								Verifier:   "aura1verifier",
								Level:      pb.VerificationLevel_VERIFICATION_LEVEL_BASIC,
								VerifiedAt: timestamppb.Now(),
							},
						},
					},
				},
				NextDataId: 3,
			},
			wantErr: false,
		},
		{
			name: "default genesis",
			genesis: types.GenesisState{
				Params:     nil,
				DataItems:  []*pb.DataItem{},
				NextDataId: 1,
			},
			wantErr: false,
		},
		{
			name: "invalid genesis - empty data ID",
			genesis: types.GenesisState{
				Params: nil,
				DataItems: []*pb.DataItem{
					{
						DataId:      "",
						Owner:       "aura1owner",
						DataType:    pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
						ContentHash: []byte("hash"),
						CreatedAt:   timestamppb.Now(),
						UpdatedAt:   timestamppb.Now(),
						AccessPolicy: &pb.AccessPolicy{
							Mode: pb.AccessMode_ACCESS_MODE_PUBLIC,
						},
					},
				},
				NextDataId: 1,
			},
			wantErr: true,
			errMsg:  "empty data_id",
		},
		{
			name: "invalid genesis - duplicate data IDs",
			genesis: types.GenesisState{
				Params: nil,
				DataItems: []*pb.DataItem{
					{
						DataId:      "data-1",
						Owner:       "aura1owner",
						DataType:    pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
						ContentHash: []byte("hash1"),
						CreatedAt:   timestamppb.Now(),
						UpdatedAt:   timestamppb.Now(),
						AccessPolicy: &pb.AccessPolicy{
							Mode: pb.AccessMode_ACCESS_MODE_PUBLIC,
						},
					},
					{
						DataId:      "data-1", // Duplicate
						Owner:       "aura1owner2",
						DataType:    pb.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
						ContentHash: []byte("hash2"),
						CreatedAt:   timestamppb.Now(),
						UpdatedAt:   timestamppb.Now(),
						AccessPolicy: &pb.AccessPolicy{
							Mode: pb.AccessMode_ACCESS_MODE_PUBLIC,
						},
					},
				},
				NextDataId: 2,
			},
			wantErr: true,
			errMsg:  "duplicate data item ID",
		},
		{
			name: "invalid genesis - empty owner",
			genesis: types.GenesisState{
				Params: nil,
				DataItems: []*pb.DataItem{
					{
						DataId:      "data-1",
						Owner:       "",
						DataType:    pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
						ContentHash: []byte("hash"),
						CreatedAt:   timestamppb.Now(),
						UpdatedAt:   timestamppb.Now(),
						AccessPolicy: &pb.AccessPolicy{
							Mode: pb.AccessMode_ACCESS_MODE_PUBLIC,
						},
					},
				},
				NextDataId: 1,
			},
			wantErr: true,
			errMsg:  "empty owner",
		},
		{
			name: "invalid genesis - unspecified data type",
			genesis: types.GenesisState{
				Params: nil,
				DataItems: []*pb.DataItem{
					{
						DataId:      "data-1",
						Owner:       "aura1owner",
						DataType:    pb.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED,
						ContentHash: []byte("hash"),
						CreatedAt:   timestamppb.Now(),
						UpdatedAt:   timestamppb.Now(),
						AccessPolicy: &pb.AccessPolicy{
							Mode: pb.AccessMode_ACCESS_MODE_PUBLIC,
						},
					},
				},
				NextDataId: 1,
			},
			wantErr: true,
			errMsg:  "unspecified data type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paramsStore := params.NewStore(types.DefaultParams())
			keeper := NewKeeper(paramsStore)

			err := keeper.InitGenesis(tt.genesis)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)

				// Verify params were set
				p := keeper.GetParams()
				require.NotNil(t, p)

				// Verify data items were loaded
				if len(tt.genesis.DataItems) > 0 {
					for _, item := range tt.genesis.DataItems {
						retrieved, ok := keeper.GetDataItem(item.DataId)
						require.True(t, ok)
						require.Equal(t, item.Owner, retrieved.Owner)
						require.Equal(t, item.DataType, retrieved.DataType)
					}
				}

				// Verify next data ID
				if tt.genesis.NextDataId > 0 {
					// Next ID should be set correctly
					require.NotZero(t, keeper.nextDataID)
				}
			}
		})
	}
}

func TestExportGenesis(t *testing.T) {
	paramsStore := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(paramsStore)

	// Store some test data
	dataID1, err := keeper.StoreDataItem(
		"aura1owner1",
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Photo 1",
		"Test photo",
		[]byte("hash1"),
		"QmTest1",
		false,
		nil,
		map[string]string{"category": "test"},
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PUBLIC},
		[]string{"photo", "test"},
	)
	require.NoError(t, err)
	require.NotEmpty(t, dataID1)

	dataID2, err := keeper.StoreDataItem(
		"aura1owner2",
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Document 1",
		"Test document",
		[]byte("hash2"),
		"QmTest2",
		false,
		nil,
		map[string]string{"type": "report"},
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PRIVATE},
		[]string{"document"},
	)
	require.NoError(t, err)
	require.NotEmpty(t, dataID2)

	// Export genesis
	exported := keeper.ExportGenesis()

	// Verify exported data
	require.NotNil(t, exported.Params)
	require.Greater(t, exported.Params.MaxDataSize, uint64(0))

	require.Len(t, exported.DataItems, 2)

	// Find the exported items
	var foundItem1, foundItem2 bool
	for _, item := range exported.DataItems {
		if item.DataId == dataID1 {
			foundItem1 = true
			require.Equal(t, "aura1owner1", item.Owner)
			require.Equal(t, pb.DataItemType_DATA_ITEM_TYPE_PHOTO, item.DataType)
		}
		if item.DataId == dataID2 {
			foundItem2 = true
			require.Equal(t, "aura1owner2", item.Owner)
			require.Equal(t, pb.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF, item.DataType)
		}
	}
	require.True(t, foundItem1, "First data item should be exported")
	require.True(t, foundItem2, "Second data item should be exported")

	require.Greater(t, exported.NextDataId, uint64(2))
}

func TestGenesisRoundTrip(t *testing.T) {
	// Create first keeper with initial state
	paramsStore1 := params.NewStore(types.DefaultParams())
	keeper1 := NewKeeper(paramsStore1)

	// Store test data
	now := time.Now()
	dataID1, err := keeper1.StoreDataItem(
		"aura1owner1",
		types.DataItemType_DATA_ITEM_TYPE_VEHICLE_REGISTRATION,
		"Vehicle Reg 1",
		"Test vehicle registration",
		[]byte("hash123"),
		"QmVehicle1",
		false,
		nil,
		map[string]string{"vin": "VIN123456789"},
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PRIVATE},
		[]string{"vehicle", "registration"},
	)
	require.NoError(t, err)

	dataID2, err := keeper1.StoreDataItem(
		"aura1owner2",
		types.DataItemType_DATA_ITEM_TYPE_GOLF_SCORE,
		"Golf Score 1",
		"Test golf score",
		[]byte("hash456"),
		"QmGolf1",
		false,
		nil,
		map[string]string{"course": "Pebble Beach"},
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PUBLIC},
		[]string{"golf", "score"},
	)
	require.NoError(t, err)

	// Export genesis from keeper1
	exported := keeper1.ExportGenesis()

	// Create a new keeper and import the exported genesis
	paramsStore2 := params.NewStore(types.DefaultParams())
	keeper2 := NewKeeper(paramsStore2)
	err = keeper2.InitGenesis(exported)
	require.NoError(t, err)

	// Verify all data was preserved
	params1 := keeper1.GetParams()
	params2 := keeper2.GetParams()
	require.Equal(t, params1.MaxDataSize, params2.MaxDataSize)
	require.Equal(t, params1.EnableIpfs, params2.EnableIpfs)

	// Verify first data item
	item1, ok := keeper2.GetDataItem(dataID1)
	require.True(t, ok)
	require.Equal(t, "aura1owner1", item1.Owner)
	require.Equal(t, pb.DataItemType_DATA_ITEM_TYPE_VEHICLE_REGISTRATION, item1.DataType)
	require.Equal(t, "Vehicle Reg 1", item1.Title)
	require.Equal(t, []byte("hash123"), item1.ContentHash)

	// Verify second data item
	item2, ok := keeper2.GetDataItem(dataID2)
	require.True(t, ok)
	require.Equal(t, "aura1owner2", item2.Owner)
	require.Equal(t, pb.DataItemType_DATA_ITEM_TYPE_GOLF_SCORE, item2.DataType)
	require.Equal(t, "Golf Score 1", item2.Title)

	// Verify user index was rebuilt
	userItems1 := keeper2.GetDataItemsByOwner("aura1owner1")
	require.Len(t, userItems1, 1)
	require.Equal(t, dataID1, userItems1[0].DataId)

	userItems2 := keeper2.GetDataItemsByOwner("aura1owner2")
	require.Len(t, userItems2, 1)
	require.Equal(t, dataID2, userItems2[0].DataId)

	// Export again and verify consistency
	exported2 := keeper2.ExportGenesis()
	require.Equal(t, len(exported.DataItems), len(exported2.DataItems))
	require.Equal(t, exported.NextDataId, exported2.NextDataId)

	_ = now
}

func TestDefaultGenesis(t *testing.T) {
	// Test that default genesis is valid
	defaultGen := types.DefaultGenesisState()
	require.NotNil(t, defaultGen)

	// Validate default genesis
	err := types.ValidateGenesisState(defaultGen)
	require.NoError(t, err)

	// Verify default params
	require.NotNil(t, defaultGen.Params)
	require.Greater(t, defaultGen.Params.MaxDataSize, uint64(0))
	require.NotEmpty(t, defaultGen.Params.StorageFeePerKb)

	// Verify default data items is empty
	require.Empty(t, defaultGen.DataItems)

	// Test importing default genesis
	paramsStore := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(paramsStore)
	err = keeper.InitGenesis(*defaultGen)
	require.NoError(t, err)

	// Verify keeper state after importing default genesis
	p := keeper.GetParams()
	require.NotNil(t, p)
	require.Greater(t, p.MaxDataSize, uint64(0))
}

func TestInitGenesis_WithVerifications(t *testing.T) {
	now := timestamppb.Now()
	genesis := types.GenesisState{
		Params: &pb.Params{
			MaxDataSize:          10485760,
			StorageFeePerKb:      "100",
			VerificationFee:      "50",
			MaxRetentionPeriod:   365 * 24 * 3600,
			EnableIpfs:           true,
			RequireVerification:  true,
			AllowPublicAccess:    true,
			MaxVerifiersPerItem:  10,
			MinVerificationLevel: pb.VerificationLevel_VERIFICATION_LEVEL_ENHANCED,
		},
		DataItems: []*pb.DataItem{
			{
				DataId:      "data-verified",
				Owner:       "aura1owner",
				DataType:    pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
				ContentHash: []byte("hash1"),
				IpfsCid:     "QmTest",
				Status:      pb.DataItemStatus_DATA_ITEM_STATUS_ACTIVE,
				CreatedAt:   now,
				UpdatedAt:   now,
				SizeBytes:   1024,
				AccessPolicy: &pb.AccessPolicy{
					Mode: pb.AccessMode_ACCESS_MODE_PUBLIC,
				},
				Verifications: []*pb.Verification{
					{
						Verifier:   "aura1verifier1",
						Level:      pb.VerificationLevel_VERIFICATION_LEVEL_ENHANCED,
						VerifiedAt: now,
						Signature:  []byte("sig1"),
					},
					{
						Verifier:   "aura1verifier2",
						Level:      pb.VerificationLevel_VERIFICATION_LEVEL_PREMIUM,
						VerifiedAt: now,
						Signature:  []byte("sig2"),
					},
				},
			},
		},
		NextDataId: 2,
	}

	paramsStore := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(paramsStore)

	err := keeper.InitGenesis(genesis)
	require.NoError(t, err)

	// Verify data item with verifications was loaded
	item, ok := keeper.GetDataItem("data-verified")
	require.True(t, ok)
	require.Len(t, item.Verifications, 2)
	require.Equal(t, "aura1verifier1", item.Verifications[0].Verifier)
	require.Equal(t, "aura1verifier2", item.Verifications[1].Verifier)
}

func TestInitGenesis_WithTypeSpecificData(t *testing.T) {
	now := timestamppb.Now()
	genesis := types.GenesisState{
		Params: nil,
		DataItems: []*pb.DataItem{
			{
				DataId:      "vehicle-1",
				Owner:       "aura1owner",
				DataType:    pb.DataItemType_DATA_ITEM_TYPE_VEHICLE_REGISTRATION,
				ContentHash: []byte("hash1"),
				Status:      pb.DataItemStatus_DATA_ITEM_STATUS_ACTIVE,
				CreatedAt:   now,
				UpdatedAt:   now,
				AccessPolicy: &pb.AccessPolicy{
					Mode: pb.AccessMode_ACCESS_MODE_PRIVATE,
				},
				VehicleRegistrationData: &pb.VehicleRegistrationData{
					Vin:          "VIN123456789",
					Make:         "Tesla",
					Model:        "Model S",
					Year:         2023,
					Color:        "Blue",
					LicensePlate: "ABC-123",
					Country:      "US",
					State:        "CA",
				},
			},
			{
				DataId:      "photo-1",
				Owner:       "aura1owner",
				DataType:    pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
				ContentHash: []byte("hash2"),
				Status:      pb.DataItemStatus_DATA_ITEM_STATUS_ACTIVE,
				CreatedAt:   now,
				UpdatedAt:   now,
				AccessPolicy: &pb.AccessPolicy{
					Mode: pb.AccessMode_ACCESS_MODE_PUBLIC,
				},
				PhotoData: &pb.PhotoData{
					CaptureTime:  now,
					GpsLatitude:  37.7749,
					GpsLongitude: -122.4194,
					DeviceModel:  "iPhone 13",
					CameraModel:  "12MP Wide",
					Width:        4032,
					Height:       3024,
				},
			},
			{
				DataId:      "golf-1",
				Owner:       "aura1owner",
				DataType:    pb.DataItemType_DATA_ITEM_TYPE_GOLF_SCORE,
				ContentHash: []byte("hash3"),
				Status:      pb.DataItemStatus_DATA_ITEM_STATUS_ACTIVE,
				CreatedAt:   now,
				UpdatedAt:   now,
				AccessPolicy: &pb.AccessPolicy{
					Mode: pb.AccessMode_ACCESS_MODE_PUBLIC,
				},
				GolfScoreData: &pb.GolfScoreData{
					PlayerName:   "John Doe",
					CourseName:   "Pebble Beach",
					PlayDate:     now,
					TotalScore:   72,
					Handicap:     5,
					TotalHoles:   18,
					Verified:     true,
					VerifiedBy:   "aura1verifier",
				},
			},
		},
		NextDataId: 4,
	}

	paramsStore := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(paramsStore)

	err := keeper.InitGenesis(genesis)
	require.NoError(t, err)

	// Verify vehicle data
	vehicle, ok := keeper.GetDataItem("vehicle-1")
	require.True(t, ok)
	require.NotNil(t, vehicle.VehicleRegistrationData)
	require.Equal(t, "VIN123456789", vehicle.VehicleRegistrationData.Vin)
	require.Equal(t, "Tesla", vehicle.VehicleRegistrationData.Make)

	// Verify photo data
	photo, ok := keeper.GetDataItem("photo-1")
	require.True(t, ok)
	require.NotNil(t, photo.PhotoData)
	require.Equal(t, "iPhone 13", photo.PhotoData.DeviceModel)
	require.Equal(t, int32(4032), photo.PhotoData.Width)

	// Verify golf score data
	golf, ok := keeper.GetDataItem("golf-1")
	require.True(t, ok)
	require.NotNil(t, golf.GolfScoreData)
	require.Equal(t, "Pebble Beach", golf.GolfScoreData.CourseName)
	require.Equal(t, int32(72), golf.GolfScoreData.TotalScore)
}
