package types_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

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
				Params: *types.DefaultParams(),
				Validators: []types.ValidatorSecurityInfo{
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
				DoubleSignEvidences: []types.DoubleSignEvidence{},
				DowntimeInfractions: []types.DowntimeInfraction{},
				Alerts:              []types.ValidatorAlert{},
				SentryNodes:         []types.SentryNodeInfo{},
			},
			expectErr: false,
		},
		{
			name: "invalid params",
			genesis: &types.GenesisState{
				Params: types.ValidatorSecurityParams{
					DoubleSignSlashFraction: sdkmath.LegacyMustNewDecFromStr("1.5"), // Invalid - > 1.0
					DowntimeSlashFraction:   sdkmath.LegacyMustNewDecFromStr("0.01"),
					SignedBlocksWindow:      1000,
					MinSignedPerWindow:      sdkmath.LegacyNewDecWithPrec(5, 1),
					DowntimeJailDuration:    1 * time.Hour,
					MinimumStakeAmount:      sdkmath.NewInt(1000000),
					MonitoringInterval:      1 * time.Minute,
					FailoverTimeout:         5 * time.Minute,
				},
				Validators:          []types.ValidatorSecurityInfo{},
				DoubleSignEvidences: []types.DoubleSignEvidence{},
				DowntimeInfractions: []types.DowntimeInfraction{},
				Alerts:              []types.ValidatorAlert{},
				SentryNodes:         []types.SentryNodeInfo{},
			},
			expectErr: true,
		},
		{
			name: "invalid validator - empty address",
			genesis: &types.GenesisState{
				Params: *types.DefaultParams(),
				Validators: []types.ValidatorSecurityInfo{
					{
						ValidatorAddress: "", // Invalid
						HotKey:           "hot",
						ColdKey:          "cold",
					},
				},
				DoubleSignEvidences: []types.DoubleSignEvidence{},
				DowntimeInfractions: []types.DowntimeInfraction{},
				Alerts:              []types.ValidatorAlert{},
				SentryNodes:         []types.SentryNodeInfo{},
			},
			expectErr: true,
		},
		{
			name: "invalid validator - invalid latitude",
			genesis: &types.GenesisState{
				Params: *types.DefaultParams(),
				Validators: []types.ValidatorSecurityInfo{
					{
						ValidatorAddress: "val1",
						Latitude:         91.0, // Invalid
						Longitude:        0.0,
					},
				},
				DoubleSignEvidences: []types.DoubleSignEvidence{},
				DowntimeInfractions: []types.DowntimeInfraction{},
				Alerts:              []types.ValidatorAlert{},
				SentryNodes:         []types.SentryNodeInfo{},
			},
			expectErr: true,
		},
		{
			name: "duplicate validator address",
			genesis: &types.GenesisState{
				Params: *types.DefaultParams(),
				Validators: []types.ValidatorSecurityInfo{
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
				DoubleSignEvidences: []types.DoubleSignEvidence{},
				DowntimeInfractions: []types.DowntimeInfraction{},
				Alerts:              []types.ValidatorAlert{},
				SentryNodes:         []types.SentryNodeInfo{},
			},
			expectErr: true,
		},
		{
			name: "invalid double sign evidence",
			genesis: &types.GenesisState{
				Params:     *types.DefaultParams(),
				Validators: []types.ValidatorSecurityInfo{},
				DoubleSignEvidences: []types.DoubleSignEvidence{
					{
						ValidatorAddress: "", // Invalid
						Height:           100,
						VoteA:            []byte("vote_a"),
						VoteB:            []byte("vote_b"),
					},
				},
				DowntimeInfractions: []types.DowntimeInfraction{},
				Alerts:              []types.ValidatorAlert{},
				SentryNodes:         []types.SentryNodeInfo{},
			},
			expectErr: true,
		},
		{
			name: "duplicate alert ID",
			genesis: func() *types.GenesisState {
				now := time.Now()
				return &types.GenesisState{
					Params:              *types.DefaultParams(),
					Validators:          []types.ValidatorSecurityInfo{},
					DoubleSignEvidences: []types.DoubleSignEvidence{},
					DowntimeInfractions: []types.DowntimeInfraction{},
					Alerts: []types.ValidatorAlert{
						{
							Id:               "alert1",
							ValidatorAddress: "val1",
							AlertType:        types.ValidatorAlert_DOWNTIME,
							Severity:         types.ValidatorAlert_WARNING,
							Message:          "Test",
							Timestamp:        &now,
						},
						{
							Id:               "alert1", // Duplicate
							ValidatorAddress: "val2",
							AlertType:        types.ValidatorAlert_DOWNTIME,
							Severity:         types.ValidatorAlert_WARNING,
							Message:          "Test",
							Timestamp:        &now,
						},
					},
					SentryNodes: []types.SentryNodeInfo{},
				}
			}(),
			expectErr: true,
		},
		{
			name: "duplicate sentry node",
			genesis: &types.GenesisState{
				Params:              *types.DefaultParams(),
				Validators:          []types.ValidatorSecurityInfo{},
				DoubleSignEvidences: []types.DoubleSignEvidence{},
				DowntimeInfractions: []types.DowntimeInfraction{},
				Alerts:              []types.ValidatorAlert{},
				SentryNodes: []types.SentryNodeInfo{
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
