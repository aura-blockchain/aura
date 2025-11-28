package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

func TestDefaultGenesisState(t *testing.T) {
	genesis := types.DefaultGenesisState()
	require.NotNil(t, genesis)
	require.NoError(t, types.ValidateGenesisState(genesis))
}

func TestGenesisValidation(t *testing.T) {
	tests := []struct {
		name      string
		genesis   *types.GenesisState
		expectErr bool
	}{
		{
			name:      "default genesis is valid",
			genesis:   types.DefaultGenesisState(),
			expectErr: false,
		},
		{
			name: "valid genesis with validators",
			genesis: &types.GenesisState{
				Params: types.DefaultParams(),
				Validators: []*types.ValidatorSecurityInfo{
					{
						ValidatorAddress: "val1",
						HotKey:           "hot1",
						ColdKey:          "cold1",
						KeysSeparated:    true,
						Region:           "us-west",
						CountryCode:      "US",
						Latitude:         37.7749,
						Longitude:        -122.4194,
						IsJailed:         false,
						IsTombstoned:     false,
					},
				},
				DoubleSignEvidences: []*types.DoubleSignEvidence{},
				DowntimeInfractions: []*types.DowntimeInfraction{},
				Alerts:              []*types.ValidatorAlert{},
				SentryNodes:         []*types.SentryNodeInfo{},
			},
			expectErr: false,
		},
		{
			name: "invalid params",
			genesis: &types.GenesisState{
				Params: &types.ValidatorSecurityParams{
					DoubleSignSlashFraction: "", // Invalid - empty
					DowntimeSlashFraction:   "0.01",
				},
				Validators:          []*types.ValidatorSecurityInfo{},
				DoubleSignEvidences: []*types.DoubleSignEvidence{},
				DowntimeInfractions: []*types.DowntimeInfraction{},
				Alerts:              []*types.ValidatorAlert{},
				SentryNodes:         []*types.SentryNodeInfo{},
			},
			expectErr: true,
		},
		{
			name: "invalid validator - empty address",
			genesis: &types.GenesisState{
				Params: types.DefaultParams(),
				Validators: []*types.ValidatorSecurityInfo{
					{
						ValidatorAddress: "", // Invalid
						HotKey:           "hot",
						ColdKey:          "cold",
					},
				},
				DoubleSignEvidences: []*types.DoubleSignEvidence{},
				DowntimeInfractions: []*types.DowntimeInfraction{},
				Alerts:              []*types.ValidatorAlert{},
				SentryNodes:         []*types.SentryNodeInfo{},
			},
			expectErr: true,
		},
		{
			name: "invalid validator - invalid latitude",
			genesis: &types.GenesisState{
				Params: types.DefaultParams(),
				Validators: []*types.ValidatorSecurityInfo{
					{
						ValidatorAddress: "val1",
						Latitude:         91.0, // Invalid
						Longitude:        0.0,
					},
				},
				DoubleSignEvidences: []*types.DoubleSignEvidence{},
				DowntimeInfractions: []*types.DowntimeInfraction{},
				Alerts:              []*types.ValidatorAlert{},
				SentryNodes:         []*types.SentryNodeInfo{},
			},
			expectErr: true,
		},
		{
			name: "duplicate validator address",
			genesis: &types.GenesisState{
				Params: types.DefaultParams(),
				Validators: []*types.ValidatorSecurityInfo{
					{
						ValidatorAddress: "val1",
						HotKey:           "hot1",
						ColdKey:          "cold1",
						KeysSeparated:    true,
					},
					{
						ValidatorAddress: "val1", // Duplicate
						HotKey:           "hot2",
						ColdKey:          "cold2",
						KeysSeparated:    true,
					},
				},
				DoubleSignEvidences: []*types.DoubleSignEvidence{},
				DowntimeInfractions: []*types.DowntimeInfraction{},
				Alerts:              []*types.ValidatorAlert{},
				SentryNodes:         []*types.SentryNodeInfo{},
			},
			expectErr: true,
		},
		{
			name: "invalid double sign evidence",
			genesis: &types.GenesisState{
				Params:     types.DefaultParams(),
				Validators: []*types.ValidatorSecurityInfo{},
				DoubleSignEvidences: []*types.DoubleSignEvidence{
					{
						ValidatorAddress: "", // Invalid
						Height:           100,
						VoteA:            []byte("vote_a"),
						VoteB:            []byte("vote_b"),
						SlashFraction:    "0.05",
					},
				},
				DowntimeInfractions: []*types.DowntimeInfraction{},
				Alerts:              []*types.ValidatorAlert{},
				SentryNodes:         []*types.SentryNodeInfo{},
			},
			expectErr: true,
		},
		{
			name: "duplicate alert ID",
			genesis: &types.GenesisState{
				Params:              types.DefaultParams(),
				Validators:          []*types.ValidatorSecurityInfo{},
				DoubleSignEvidences: []*types.DoubleSignEvidence{},
				DowntimeInfractions: []*types.DowntimeInfraction{},
				Alerts: []*types.ValidatorAlert{
					{
						Id:               "alert1",
						ValidatorAddress: "val1",
						AlertType:        types.ValidatorAlert_DOWNTIME,
						Severity:         types.ValidatorAlert_WARNING,
						Message:          "Test",
						Timestamp:        timestamppb.New(time.Now()),
					},
					{
						Id:               "alert1", // Duplicate
						ValidatorAddress: "val2",
						AlertType:        types.ValidatorAlert_DOWNTIME,
						Severity:         types.ValidatorAlert_WARNING,
						Message:          "Test",
						Timestamp:        timestamppb.New(time.Now()),
					},
				},
				SentryNodes: []*types.SentryNodeInfo{},
			},
			expectErr: true,
		},
		{
			name: "duplicate sentry node",
			genesis: &types.GenesisState{
				Params:              types.DefaultParams(),
				Validators:          []*types.ValidatorSecurityInfo{},
				DoubleSignEvidences: []*types.DoubleSignEvidence{},
				DowntimeInfractions: []*types.DowntimeInfraction{},
				Alerts:              []*types.ValidatorAlert{},
				SentryNodes: []*types.SentryNodeInfo{
					{
						Address:          "sentry1",
						ValidatorAddress: "val1",
						IpAddress:        "192.168.1.1",
						Port:             26656,
					},
					{
						Address:          "sentry1", // Duplicate
						ValidatorAddress: "val1",
						IpAddress:        "192.168.1.2",
						Port:             26656,
					},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateGenesisState(tt.genesis)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
