package types

import (
	"time"

	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
	"google.golang.org/protobuf/types/known/durationpb"
)

// DefaultParams returns default security parameters
func DefaultParams() securitypb.Params {
	return securitypb.Params{
		Network: &securitypb.NetworkSecurityParams{
			RateLimit: &securitypb.RateLimitConfig{
				MaxRequestsPerSecond: 100,
				BurstSize:            200,
				WindowDuration:       durationpb.New(1 * time.Second),
				BanDuration:          durationpb.New(1 * time.Hour),
			},
			Connection: &securitypb.ConnectionConfig{
				MaxInboundConnections:  100,
				MaxOutboundConnections: 50,
				MaxConnectionsPerIp:    10,
				ConnectionTimeout:      durationpb.New(30 * time.Second),
			},
			Mempool: &securitypb.MempoolConfig{
				MaxSize:        5000,
				MaxBytes:       104857600,
				MinPriorityFee: "1000",
			},
			Reputation: &securitypb.ReputationConfig{
				EnableTracking: true,
				InitialScore:   100,
			},
			Gossip:             &securitypb.GossipConfig{},
			ForkDetection:      &securitypb.ForkDetectionConfig{},
			PartitionDetection: &securitypb.PartitionDetectionConfig{},
		},
		Validator: &securitypb.ValidatorSecurityParams{
			DoubleSignSlashFraction: "0.050000000000000000",
			DowntimeSlashFraction:   "0.010000000000000000",
			SignedBlocksWindow:      10000,
			MinSignedPerWindow:      "0.500000000000000000",
		},
		Wallet:   &securitypb.WalletSecurityParams{},
		Incident: &securitypb.IncidentResponseParams{},
		Crypto: &securitypb.CryptographyParams{
			MinThresholdParticipants: 3,
		},
		Privacy: &securitypb.PrivacyParams{
			MinRingSize:           3,
			MaxRingSize:           11,
			MinMixingParticipants: 2,
			MixingFee:             "100",
		},
	}
}
