// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
)

// DefaultParams returns default network security parameters
func DefaultParams() *Params {
	return &Params{
		RateLimit: RateLimitConfig{
			MaxRequestsPerSecond:  100,
			BurstSize:             200,
			WindowDuration:        60 * time.Second,
			BanDuration:           3600 * time.Second, // 1 hour
			BandwidthLimitPerPeer: 1048576,            // 1 MB/s
		},
		Connection: ConnectionConfig{
			MaxInboundConnections:  50,
			MaxOutboundConnections: 10,
			MaxConnectionsPerIp:    5,
			ConnectionTimeout:      60 * time.Second,
			TrustedPeersOnly:       false,
			MinPeerDiversity:       3,
		},
		Mempool: MempoolConfig{
			MaxSize:            5000,
			MaxBytes:           10485760,          // 10 MB
			MinPriorityFee:     math.NewInt(1000), // 1000 as minimum fee
			MaxTxsPerAccount:   100,
			EvictionPolicy:     "oldest",
			EnablePriorityFees: true,
		},
		Reputation: ReputationConfig{
			EnableTracking:     true,
			InitialScore:       50,
			MinScoreToConnect:  0,
			DecayRate:          1,
			MaxScore:           100,
			MisbehaviorPenalty: 10,
			GoodBehaviorReward: 5,
		},
		Gossip: GossipConfig{
			VerifySignatures:       true,
			MaxMessageSize:         1048576, // 1 MB
			MessageTtl:             300 * time.Second,
			EnableRedundancyFilter: true,
			MaxFanout:              10,
		},
		ForkDetection: ForkDetectionConfig{
			EnableDetection:      true,
			HeightDiffThreshold:  100,
			EnableAutoResolution: false,
			ConfirmationDepth:    6,
		},
		PartitionDetection: PartitionDetectionConfig{
			EnableDetection:    true,
			MinConnectedPeers:  5,
			CheckInterval:      60 * time.Second,
			PartitionThreshold: 3,
		},
	}
}

// ValidateParams validates network security parameters
func ValidateParams(params *Params) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}

	// Validate RateLimit
	if params.RateLimit.MaxRequestsPerSecond == 0 {
		return fmt.Errorf("max_requests_per_second must be greater than 0")
	}
	if params.RateLimit.BurstSize == 0 {
		return fmt.Errorf("burst_size must be greater than 0")
	}

	// Validate Connection
	if params.Connection.MaxInboundConnections == 0 {
		return fmt.Errorf("max_inbound_connections must be greater than 0")
	}
	if params.Connection.ConnectionTimeout == 0 {
		return fmt.Errorf("connection_timeout must be greater than 0")
	}

	// Validate Mempool
	if params.Mempool.MaxSize == 0 {
		return fmt.Errorf("max_size must be greater than 0")
	}
	if params.Mempool.MaxBytes == 0 {
		return fmt.Errorf("max_bytes must be greater than 0")
	}

	// Validate Reputation
	if params.Reputation.EnableTracking {
		if params.Reputation.MaxScore <= params.Reputation.MinScoreToConnect {
			return fmt.Errorf("max_score must be greater than min_score_to_connect")
		}
		if params.Reputation.InitialScore < params.Reputation.MinScoreToConnect || params.Reputation.InitialScore > params.Reputation.MaxScore {
			return fmt.Errorf("initial_score must be between min_score_to_connect and max_score")
		}
	}

	// Validate Gossip
	if params.Gossip.MaxMessageSize == 0 {
		return fmt.Errorf("max_message_size must be greater than 0")
	}

	// Validate ForkDetection
	if params.ForkDetection.EnableDetection {
		if params.ForkDetection.HeightDiffThreshold == 0 {
			return fmt.Errorf("height_diff_threshold must be greater than 0")
		}
	}

	// Validate PartitionDetection
	if params.PartitionDetection.EnableDetection {
		if params.PartitionDetection.MinConnectedPeers == 0 {
			return fmt.Errorf("min_connected_peers must be greater than 0")
		}
	}

	return nil
}

// NOTE: ValidateGenesisState and DefaultGenesisState are defined in genesis.go
// to avoid duplication.
