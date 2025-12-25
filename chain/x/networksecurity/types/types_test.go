// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()
	require.NotNil(t, params)
	require.NotNil(t, params.RateLimit)
	require.NotNil(t, params.Connection)
	require.NotNil(t, params.Mempool)
	require.NotNil(t, params.Reputation)
	require.NotNil(t, params.Gossip)
	require.NotNil(t, params.ForkDetection)
	require.NotNil(t, params.PartitionDetection)

	// Validate default params
	require.NoError(t, ValidateParams(params))
}

func TestValidateParams_Valid(t *testing.T) {
	params := DefaultParams()
	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_NilParams(t *testing.T) {
	err := ValidateParams(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "params cannot be nil")
}

func TestValidateParams_RateLimit(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*Params)
		wantError string
	}{
		{
			name: "zero max requests per second",
			mutate: func(p *Params) {
				p.RateLimit.MaxRequestsPerSecond = 0
			},
			wantError: "max_requests_per_second must be greater than 0",
		},
		{
			name: "zero burst size",
			mutate: func(p *Params) {
				p.RateLimit.BurstSize = 0
			},
			wantError: "burst_size must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := DefaultParams()
			tt.mutate(params)
			err := ValidateParams(params)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantError)
		})
	}
}

func TestValidateParams_Connection(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*Params)
		wantError string
	}{
		{
			name: "zero max inbound connections",
			mutate: func(p *Params) {
				p.Connection.MaxInboundConnections = 0
			},
			wantError: "max_inbound_connections must be greater than 0",
		},
		{
			name: "zero connection timeout",
			mutate: func(p *Params) {
				p.Connection.ConnectionTimeout = 0
			},
			wantError: "connection_timeout must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := DefaultParams()
			tt.mutate(params)
			err := ValidateParams(params)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantError)
		})
	}
}

func TestValidateParams_Mempool(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*Params)
		wantError string
	}{
		{
			name: "zero max size",
			mutate: func(p *Params) {
				p.Mempool.MaxSize = 0
			},
			wantError: "max_size must be greater than 0",
		},
		{
			name: "zero max bytes",
			mutate: func(p *Params) {
				p.Mempool.MaxBytes = 0
			},
			wantError: "max_bytes must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := DefaultParams()
			tt.mutate(params)
			err := ValidateParams(params)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantError)
		})
	}
}

func TestValidateParams_Reputation(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*Params)
		wantError string
	}{
		{
			name: "max score less than min score",
			mutate: func(p *Params) {
				p.Reputation.EnableTracking = true
				p.Reputation.MaxScore = 10
				p.Reputation.MinScoreToConnect = 20
			},
			wantError: "max_score must be greater than min_score_to_connect",
		},
		{
			name: "initial score below min",
			mutate: func(p *Params) {
				p.Reputation.EnableTracking = true
				p.Reputation.InitialScore = 10
				p.Reputation.MinScoreToConnect = 20
			},
			wantError: "initial_score must be between min_score_to_connect and max_score",
		},
		{
			name: "initial score above max",
			mutate: func(p *Params) {
				p.Reputation.EnableTracking = true
				p.Reputation.InitialScore = 150
				p.Reputation.MaxScore = 100
			},
			wantError: "initial_score must be between min_score_to_connect and max_score",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := DefaultParams()
			tt.mutate(params)
			err := ValidateParams(params)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantError)
		})
	}
}

func TestValidateParams_ReputationDisabled(t *testing.T) {
	params := DefaultParams()
	params.Reputation.EnableTracking = false
	params.Reputation.MaxScore = 10
	params.Reputation.MinScoreToConnect = 20
	// Should not error when tracking is disabled
	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_Gossip(t *testing.T) {
	params := DefaultParams()
	params.Gossip.MaxMessageSize = 0
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "max_message_size must be greater than 0")
}

func TestValidateParams_ForkDetection(t *testing.T) {
	params := DefaultParams()
	params.ForkDetection.EnableDetection = true
	params.ForkDetection.HeightDiffThreshold = 0
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "height_diff_threshold must be greater than 0")
}

func TestValidateParams_ForkDetectionDisabled(t *testing.T) {
	params := DefaultParams()
	params.ForkDetection.EnableDetection = false
	params.ForkDetection.HeightDiffThreshold = 0
	// Should not error when detection is disabled
	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_PartitionDetection(t *testing.T) {
	params := DefaultParams()
	params.PartitionDetection.EnableDetection = true
	params.PartitionDetection.MinConnectedPeers = 0
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "min_connected_peers must be greater than 0")
}

func TestValidateParams_PartitionDetectionDisabled(t *testing.T) {
	params := DefaultParams()
	params.PartitionDetection.EnableDetection = false
	params.PartitionDetection.MinConnectedPeers = 0
	// Should not error when detection is disabled
	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestDefaultParams_Values(t *testing.T) {
	params := DefaultParams()

	// RateLimit
	require.Equal(t, uint64(100), params.RateLimit.MaxRequestsPerSecond)
	require.Equal(t, uint64(200), params.RateLimit.BurstSize)
	require.Equal(t, 60*time.Second, params.RateLimit.WindowDuration)
	require.Equal(t, 3600*time.Second, params.RateLimit.BanDuration)
	require.Equal(t, uint64(1048576), params.RateLimit.BandwidthLimitPerPeer)

	// Connection
	require.Equal(t, uint32(50), params.Connection.MaxInboundConnections)
	require.Equal(t, uint32(10), params.Connection.MaxOutboundConnections)
	require.Equal(t, uint32(5), params.Connection.MaxConnectionsPerIp)
	require.Equal(t, 60*time.Second, params.Connection.ConnectionTimeout)
	require.False(t, params.Connection.TrustedPeersOnly)
	require.Equal(t, uint32(3), params.Connection.MinPeerDiversity)

	// Mempool
	require.Equal(t, uint64(5000), params.Mempool.MaxSize)
	require.Equal(t, uint64(10485760), params.Mempool.MaxBytes)
	require.Equal(t, math.NewInt(1000), params.Mempool.MinPriorityFee)
	require.Equal(t, uint32(100), params.Mempool.MaxTxsPerAccount)
	require.Equal(t, "oldest", params.Mempool.EvictionPolicy)
	require.True(t, params.Mempool.EnablePriorityFees)

	// Reputation
	require.True(t, params.Reputation.EnableTracking)
	require.Equal(t, int64(50), params.Reputation.InitialScore)
	require.Equal(t, int64(0), params.Reputation.MinScoreToConnect)
	require.Equal(t, int64(1), params.Reputation.DecayRate)
	require.Equal(t, int64(100), params.Reputation.MaxScore)
	require.Equal(t, int64(10), params.Reputation.MisbehaviorPenalty)
	require.Equal(t, int64(5), params.Reputation.GoodBehaviorReward)

	// Gossip
	require.True(t, params.Gossip.VerifySignatures)
	require.Equal(t, uint64(1048576), params.Gossip.MaxMessageSize)
	require.Equal(t, 300*time.Second, params.Gossip.MessageTtl)
	require.True(t, params.Gossip.EnableRedundancyFilter)
	require.Equal(t, uint32(10), params.Gossip.MaxFanout)

	// ForkDetection
	require.True(t, params.ForkDetection.EnableDetection)
	require.Equal(t, int64(100), params.ForkDetection.HeightDiffThreshold)
	require.False(t, params.ForkDetection.EnableAutoResolution)
	require.Equal(t, int64(6), params.ForkDetection.ConfirmationDepth)

	// PartitionDetection
	require.True(t, params.PartitionDetection.EnableDetection)
	require.Equal(t, uint32(5), params.PartitionDetection.MinConnectedPeers)
	require.Equal(t, 60*time.Second, params.PartitionDetection.CheckInterval)
	require.Equal(t, uint32(3), params.PartitionDetection.PartitionThreshold)
}

func TestParams_CustomValues(t *testing.T) {
	params := &Params{
		RateLimit: RateLimitConfig{
			MaxRequestsPerSecond:  200,
			BurstSize:             400,
			WindowDuration:        120 * time.Second,
			BanDuration:           7200 * time.Second,
			BandwidthLimitPerPeer: 2097152,
		},
		Connection: ConnectionConfig{
			MaxInboundConnections:  100,
			MaxOutboundConnections: 20,
			MaxConnectionsPerIp:    10,
			ConnectionTimeout:      120 * time.Second,
			TrustedPeersOnly:       true,
			MinPeerDiversity:       5,
		},
		Mempool: MempoolConfig{
			MaxSize:            10000,
			MaxBytes:           20971520,
			MinPriorityFee:     math.NewInt(2000),
			MaxTxsPerAccount:   200,
			EvictionPolicy:     "priority",
			EnablePriorityFees: true,
		},
		Reputation: ReputationConfig{
			EnableTracking:     true,
			InitialScore:       75,
			MinScoreToConnect:  25,
			DecayRate:          2,
			MaxScore:           150,
			MisbehaviorPenalty: 15,
			GoodBehaviorReward: 10,
		},
		Gossip: GossipConfig{
			VerifySignatures:       true,
			MaxMessageSize:         2097152,
			MessageTtl:             600 * time.Second,
			EnableRedundancyFilter: true,
			MaxFanout:              15,
		},
		ForkDetection: ForkDetectionConfig{
			EnableDetection:      true,
			HeightDiffThreshold:  200,
			EnableAutoResolution: true,
			ConfirmationDepth:    12,
		},
		PartitionDetection: PartitionDetectionConfig{
			EnableDetection:    true,
			MinConnectedPeers:  10,
			CheckInterval:      120 * time.Second,
			PartitionThreshold: 5,
		},
	}

	err := ValidateParams(params)
	require.NoError(t, err)
}
