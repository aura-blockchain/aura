package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultGenesisState(t *testing.T) {
	gs := DefaultGenesisState()
	require.NotNil(t, gs)
	require.NotNil(t, gs.Params)
	require.NotNil(t, gs.TrustedPeers)
	require.NotNil(t, gs.Reputations)
	require.NotNil(t, gs.RateLimits)
	require.NotNil(t, gs.ForkAlerts)
	require.NotNil(t, gs.PartitionAlerts)
	require.Empty(t, gs.TrustedPeers)
	require.Empty(t, gs.Reputations)
	require.Empty(t, gs.RateLimits)
	require.Empty(t, gs.ForkAlerts)
	require.Empty(t, gs.PartitionAlerts)

	// Validate default genesis
	require.NoError(t, ValidateGenesisState(gs))
}

func TestValidateGenesisState_Valid(t *testing.T) {
	gs := DefaultGenesisState()
	err := ValidateGenesisState(gs)
	require.NoError(t, err)
}

func TestValidateGenesisState_InvalidParams(t *testing.T) {
	gs := DefaultGenesisState()
	gs.Params.RateLimit.MaxRequestsPerSecond = 0
	err := ValidateGenesisState(gs)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid params")
}

func TestValidateGenesisState_TrustedPeers(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*GenesisState)
		wantError string
	}{
		{
			name: "empty peer_id",
			setup: func(gs *GenesisState) {
				gs.TrustedPeers = []TrustedPeer{
					{
						PeerId:  "",
						Address: "192.168.1.1:26656",
					},
				}
			},
			wantError: "trusted peer has empty peer_id",
		},
		{
			name: "empty address",
			setup: func(gs *GenesisState) {
				gs.TrustedPeers = []TrustedPeer{
					{
						PeerId:  "peer1",
						Address: "",
					},
				}
			},
			wantError: "has empty address",
		},
		{
			name: "duplicate peer_id",
			setup: func(gs *GenesisState) {
				gs.TrustedPeers = []TrustedPeer{
					{
						PeerId:  "peer1",
						Address: "192.168.1.1:26656",
					},
					{
						PeerId:  "peer1",
						Address: "192.168.1.2:26656",
					},
				}
			},
			wantError: "duplicate trusted peer_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := DefaultGenesisState()
			tt.setup(gs)
			err := ValidateGenesisState(gs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantError)
		})
	}
}

func TestValidateGenesisState_TrustedPeers_Valid(t *testing.T) {
	gs := DefaultGenesisState()
	now := time.Now()
	gs.TrustedPeers = []TrustedPeer{
		{
			PeerId:  "peer1",
			Address: "192.168.1.1:26656",
			AddedAt: now,
		},
		{
			PeerId:  "peer2",
			Address: "192.168.1.2:26656",
			AddedAt: now,
		},
	}
	err := ValidateGenesisState(gs)
	require.NoError(t, err)
}

func TestValidateGenesisState_Reputations(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*GenesisState)
		wantError string
	}{
		{
			name: "empty peer_id",
			setup: func(gs *GenesisState) {
				gs.Reputations = []NodeReputation{
					{
						PeerId: "",
						Score:  50,
					},
				}
			},
			wantError: "reputation has empty peer_id",
		},
		{
			name: "duplicate peer_id",
			setup: func(gs *GenesisState) {
				gs.Reputations = []NodeReputation{
					{
						PeerId: "peer1",
						Score:  50,
					},
					{
						PeerId: "peer1",
						Score:  60,
					},
				}
			},
			wantError: "duplicate reputation peer_id",
		},
		{
			name: "score below minimum",
			setup: func(gs *GenesisState) {
				gs.Reputations = []NodeReputation{
					{
						PeerId: "peer1",
						Score:  -10,
					},
				}
			},
			wantError: "reputation score for peer1 out of valid range",
		},
		{
			name: "score above maximum",
			setup: func(gs *GenesisState) {
				gs.Reputations = []NodeReputation{
					{
						PeerId: "peer1",
						Score:  150,
					},
				}
			},
			wantError: "reputation score for peer1 out of valid range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := DefaultGenesisState()
			tt.setup(gs)
			err := ValidateGenesisState(gs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantError)
		})
	}
}

func TestValidateGenesisState_Reputations_Valid(t *testing.T) {
	gs := DefaultGenesisState()
	gs.Reputations = []NodeReputation{
		{
			PeerId: "peer1",
			Score:  50,
		},
		{
			PeerId: "peer2",
			Score:  75,
		},
		{
			PeerId: "peer3",
			Score:  100,
		},
	}
	err := ValidateGenesisState(gs)
	require.NoError(t, err)
}

func TestValidateGenesisState_RateLimits(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*GenesisState)
		wantError string
	}{
		{
			name: "empty peer_id",
			setup: func(gs *GenesisState) {
				gs.RateLimits = []RateLimitEntry{
					{
						PeerId:       "",
						RequestCount: 10,
					},
				}
			},
			wantError: "rate limit has empty peer_id",
		},
		{
			name: "duplicate peer_id",
			setup: func(gs *GenesisState) {
				gs.RateLimits = []RateLimitEntry{
					{
						PeerId:       "peer1",
						RequestCount: 10,
					},
					{
						PeerId:       "peer1",
						RequestCount: 20,
					},
				}
			},
			wantError: "duplicate rate limit peer_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := DefaultGenesisState()
			tt.setup(gs)
			err := ValidateGenesisState(gs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantError)
		})
	}
}

func TestValidateGenesisState_ForkAlerts(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*GenesisState)
		wantError string
	}{
		{
			name: "empty alert_id",
			setup: func(gs *GenesisState) {
				gs.ForkAlerts = []ForkAlert{
					{
						AlertId:     "",
						BlockHeight: 1000,
					},
				}
			},
			wantError: "fork alert has empty alert_id",
		},
		{
			name: "duplicate alert_id",
			setup: func(gs *GenesisState) {
				gs.ForkAlerts = []ForkAlert{
					{
						AlertId:     "alert1",
						BlockHeight: 1000,
					},
					{
						AlertId:     "alert1",
						BlockHeight: 2000,
					},
				}
			},
			wantError: "duplicate fork alert_id",
		},
		{
			name: "negative block height",
			setup: func(gs *GenesisState) {
				gs.ForkAlerts = []ForkAlert{
					{
						AlertId:     "alert1",
						BlockHeight: -100,
					},
				}
			},
			wantError: "has negative block height",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := DefaultGenesisState()
			tt.setup(gs)
			err := ValidateGenesisState(gs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantError)
		})
	}
}

func TestValidateGenesisState_PartitionAlerts(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*GenesisState)
		wantError string
	}{
		{
			name: "empty alert_id",
			setup: func(gs *GenesisState) {
				gs.PartitionAlerts = []PartitionAlert{
					{
						AlertId: "",
					},
				}
			},
			wantError: "partition alert has empty alert_id",
		},
		{
			name: "duplicate alert_id",
			setup: func(gs *GenesisState) {
				gs.PartitionAlerts = []PartitionAlert{
					{
						AlertId: "alert1",
					},
					{
						AlertId: "alert1",
					},
				}
			},
			wantError: "duplicate partition alert_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := DefaultGenesisState()
			tt.setup(gs)
			err := ValidateGenesisState(gs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantError)
		})
	}
}

func TestValidateGenesisState_CompleteValid(t *testing.T) {
	params := DefaultParams()
	now := time.Now()
	gs := &GenesisState{
		Params: *params,
		TrustedPeers: []TrustedPeer{
			{
				PeerId:  "peer1",
				Address: "192.168.1.1:26656",
				AddedAt: now,
			},
		},
		Reputations: []NodeReputation{
			{
				PeerId: "peer1",
				Score:  75,
			},
		},
		RateLimits: []RateLimitEntry{
			{
				PeerId:       "peer1",
				RequestCount: 50,
			},
		},
		ForkAlerts: []ForkAlert{
			{
				AlertId:     "fork1",
				BlockHeight: 1000,
				DetectedAt:  now,
			},
		},
		PartitionAlerts: []PartitionAlert{
			{
				AlertId:    "partition1",
				DetectedAt: now,
			},
		},
	}

	err := ValidateGenesisState(gs)
	require.NoError(t, err)
}
