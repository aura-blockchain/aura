package types

import (
	"cosmossdk.io/math"
	"fmt"
	networksecuritypb "github.com/aequitas/aura/proto/aura/networksecurity/v1beta1"
)

// DefaultParams returns default parameters
func DefaultParams() *networksecuritypb.Params {
	rateLimit := DefaultRateLimitConfig()
	connection := DefaultConnectionConfig()
	mempool := DefaultMempoolConfig()
	reputation := DefaultReputationConfig()
	gossip := DefaultGossipConfig()
	forkDetection := DefaultForkDetectionConfig()
	partitionDetection := DefaultPartitionDetectionConfig()

	return &networksecuritypb.Params{
		RateLimit:          rateLimit,
		Connection:         connection,
		Mempool:            mempool,
		Reputation:         reputation,
		Gossip:             gossip,
		ForkDetection:      forkDetection,
		PartitionDetection: partitionDetection,
	}
}
func ValidateRateLimitConfig(c *RateLimitConfig) error {
	if c.MaxRequestsPerSecond == 0 {
		return fmt.Errorf("max_requests_per_second must be positive")
	}
	if c.BurstSize < c.MaxRequestsPerSecond {
		return fmt.Errorf("burst_size must be >= max_requests_per_second")
	}
	if c.WindowDuration == nil || c.WindowDuration.AsDuration() <= 0 {
		return fmt.Errorf("window_duration must be positive")
	}
	if c.BanDuration == nil || c.BanDuration.AsDuration() <= 0 {
		return fmt.Errorf("ban_duration must be positive")
	}
	if c.BandwidthLimitPerPeer == 0 {
		return fmt.Errorf("bandwidth_limit_per_peer must be positive")
	}
	return nil
}

// ValidateConnectionConfig validates ConnectionConfig
func ValidateConnectionConfig(c *ConnectionConfig) error {
	if c.MaxInboundConnections == 0 {
		return fmt.Errorf("max_inbound_connections must be positive")
	}
	if c.MaxOutboundConnections == 0 {
		return fmt.Errorf("max_outbound_connections must be positive")
	}
	if c.MaxConnectionsPerIp == 0 {
		return fmt.Errorf("max_connections_per_ip must be positive")
	}
	if c.ConnectionTimeout == nil || c.ConnectionTimeout.AsDuration() <= 0 {
		return fmt.Errorf("connection_timeout must be positive")
	}
	return nil
}

// ValidateMempoolConfig validates MempoolConfig
func ValidateMempoolConfig(c *MempoolConfig) error {
	if c.MaxSize == 0 {
		return fmt.Errorf("max_size must be positive")
	}
	if c.MaxBytes == 0 {
		return fmt.Errorf("max_bytes must be positive")
	}
	// Parse string to math.Int for validation
	fee, ok := math.NewIntFromString(c.MinPriorityFee)
	if !ok {
		return fmt.Errorf("invalid min_priority_fee format")
	}
	if fee.IsNegative() {
		return fmt.Errorf("min_priority_fee cannot be negative")
	}
	if c.MaxTxsPerAccount == 0 {
		return fmt.Errorf("max_txs_per_account must be positive")
	}
	validPolicies := map[string]bool{
		"oldest":     true,
		"lowest_fee": true,
		"random":     true,
	}
	if !validPolicies[c.EvictionPolicy] {
		return fmt.Errorf("invalid eviction_policy: %s", c.EvictionPolicy)
	}
	return nil
}

// ValidateReputationConfig validates ReputationConfig
func ValidateReputationConfig(c *ReputationConfig) error {
	if c.InitialScore < 0 {
		return fmt.Errorf("initial_score cannot be negative")
	}
	if c.MinScoreToConnect < 0 {
		return fmt.Errorf("min_score_to_connect cannot be negative")
	}
	if c.MaxScore <= c.InitialScore {
		return fmt.Errorf("max_score must be greater than initial_score")
	}
	if c.MisbehaviorPenalty < 0 {
		return fmt.Errorf("misbehavior_penalty cannot be negative")
	}
	if c.GoodBehaviorReward < 0 {
		return fmt.Errorf("good_behavior_reward cannot be negative")
	}
	return nil
}

// ValidateGossipConfig validates GossipConfig
func ValidateGossipConfig(c *GossipConfig) error {
	if c.MaxMessageSize == 0 {
		return fmt.Errorf("max_message_size must be positive")
	}
	if c.MessageTtl == nil || c.MessageTtl.AsDuration() <= 0 {
		return fmt.Errorf("message_ttl must be positive")
	}
	if c.MaxFanout == 0 {
		return fmt.Errorf("max_fanout must be positive")
	}
	return nil
}

// ValidateForkDetectionConfig validates ForkDetectionConfig
func ValidateForkDetectionConfig(c *ForkDetectionConfig) error {
	if c.HeightDiffThreshold < 0 {
		return fmt.Errorf("height_diff_threshold cannot be negative")
	}
	if c.ConfirmationDepth < 0 {
		return fmt.Errorf("confirmation_depth cannot be negative")
	}
	return nil
}

// ValidatePartitionDetectionConfig validates PartitionDetectionConfig
func ValidatePartitionDetectionConfig(c *PartitionDetectionConfig) error {
	if c.MinConnectedPeers == 0 {
		return fmt.Errorf("min_connected_peers must be positive")
	}
	if c.CheckInterval == nil || c.CheckInterval.AsDuration() <= 0 {
		return fmt.Errorf("check_interval must be positive")
	}
	if c.PartitionThreshold == 0 {
		return fmt.Errorf("partition_threshold must be positive")
	}
	return nil
}

// ValidateParams validates all network security parameters
func ValidateParams(p *Params) error {
	if p == nil {
		return fmt.Errorf("params cannot be nil")
	}

	if p.RateLimit != nil {
		if err := ValidateRateLimitConfig(p.RateLimit); err != nil {
			return fmt.Errorf("invalid rate limit config: %w", err)
		}
	}

	if p.Connection != nil {
		if err := ValidateConnectionConfig(p.Connection); err != nil {
			return fmt.Errorf("invalid connection config: %w", err)
		}
	}

	if p.Mempool != nil {
		if err := ValidateMempoolConfig(p.Mempool); err != nil {
			return fmt.Errorf("invalid mempool config: %w", err)
		}
	}

	if p.Reputation != nil {
		if err := ValidateReputationConfig(p.Reputation); err != nil {
			return fmt.Errorf("invalid reputation config: %w", err)
		}
	}

	if p.Gossip != nil {
		if err := ValidateGossipConfig(p.Gossip); err != nil {
			return fmt.Errorf("invalid gossip config: %w", err)
		}
	}

	if p.ForkDetection != nil {
		if err := ValidateForkDetectionConfig(p.ForkDetection); err != nil {
			return fmt.Errorf("invalid fork detection config: %w", err)
		}
	}

	if p.PartitionDetection != nil {
		if err := ValidatePartitionDetectionConfig(p.PartitionDetection); err != nil {
			return fmt.Errorf("invalid partition detection config: %w", err)
		}
	}

	return nil
}
