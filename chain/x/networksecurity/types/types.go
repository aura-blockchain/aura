package types

import (
	"time"

	"cosmossdk.io/math"
	networksecuritypb "github.com/aequitas/aura/proto/aura/networksecurity/v1beta1"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Type aliases for proto-generated types
type (
	Params                   = networksecuritypb.Params
	GenesisState             = networksecuritypb.GenesisState
	RateLimitConfig          = networksecuritypb.RateLimitConfig
	ConnectionConfig         = networksecuritypb.ConnectionConfig
	MempoolConfig            = networksecuritypb.MempoolConfig
	ReputationConfig         = networksecuritypb.ReputationConfig
	GossipConfig             = networksecuritypb.GossipConfig
	ForkDetectionConfig      = networksecuritypb.ForkDetectionConfig
	PartitionDetectionConfig = networksecuritypb.PartitionDetectionConfig
	PeerInfo                 = networksecuritypb.PeerInfo
	RateLimitEntry           = networksecuritypb.RateLimitEntry
	NodeReputation           = networksecuritypb.NodeReputation
	TrustedPeer              = networksecuritypb.TrustedPeer
	ForkAlert                = networksecuritypb.ForkAlert
	PartitionAlert           = networksecuritypb.PartitionAlert
	MempoolStats             = networksecuritypb.MempoolStats
)

// Helper functions for creating default configs

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		MaxRequestsPerSecond:  100,
		BurstSize:             200,
		WindowDuration:        durationpb.New(time.Minute),
		BanDuration:           durationpb.New(time.Hour),
		BandwidthLimitPerPeer: 10 * 1024 * 1024, // 10 MB/s
	}
}

// DefaultConnectionConfig returns default connection configuration
func DefaultConnectionConfig() *ConnectionConfig {
	return &ConnectionConfig{
		MaxInboundConnections:  100,
		MaxOutboundConnections: 50,
		MaxConnectionsPerIp:    10,
		ConnectionTimeout:      durationpb.New(time.Minute * 5),
		TrustedPeersOnly:       false,
		MinPeerDiversity:       5, // At least 5 different ASNs
	}
}

// DefaultMempoolConfig returns default mempool configuration
func DefaultMempoolConfig() *MempoolConfig {
	return &MempoolConfig{
		MaxSize:            10000,
		MaxBytes:           100 * 1024 * 1024, // 100 MB
		MinPriorityFee:     math.NewInt(1000).String(),
		MaxTxsPerAccount:   100,
		EvictionPolicy:     "lowest_fee",
		EnablePriorityFees: true,
	}
}

// DefaultReputationConfig returns default reputation configuration
func DefaultReputationConfig() *ReputationConfig {
	return &ReputationConfig{
		EnableTracking:     true,
		InitialScore:       100,
		MinScoreToConnect:  0,
		DecayRate:          1,
		MaxScore:           1000,
		MisbehaviorPenalty: 50,
		GoodBehaviorReward: 1,
	}
}

// DefaultGossipConfig returns default gossip configuration
func DefaultGossipConfig() *GossipConfig {
	return &GossipConfig{
		VerifySignatures:       true,
		MaxMessageSize:         1024 * 1024, // 1 MB
		MessageTtl:             durationpb.New(time.Minute * 5),
		EnableRedundancyFilter: true,
		MaxFanout:              8,
	}
}

// DefaultForkDetectionConfig returns default fork detection configuration
func DefaultForkDetectionConfig() *ForkDetectionConfig {
	return &ForkDetectionConfig{
		EnableDetection:      true,
		HeightDiffThreshold:  10,
		EnableAutoResolution: false,
		ConfirmationDepth:    6,
	}
}

// DefaultPartitionDetectionConfig returns default partition detection configuration
func DefaultPartitionDetectionConfig() *PartitionDetectionConfig {
	return &PartitionDetectionConfig{
		EnableDetection:    true,
		MinConnectedPeers:  10,
		CheckInterval:      durationpb.New(time.Minute),
		PartitionThreshold: 5,
	}
}
