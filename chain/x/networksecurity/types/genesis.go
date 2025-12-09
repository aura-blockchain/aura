package types

import (
	"fmt"
)

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:          *DefaultParams(),
		TrustedPeers:    []TrustedPeer{},
		Reputations:     []NodeReputation{},
		RateLimits:      []RateLimitEntry{},
		ForkAlerts:      []ForkAlert{},
		PartitionAlerts: []PartitionAlert{},
	}
}

// ValidateGenesisState performs basic validation of genesis data
func ValidateGenesisState(gs *GenesisState) error {
	if err := ValidateParams(&gs.Params); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate trusted peers
	trustedPeerIDs := make(map[string]bool)
	for _, peer := range gs.TrustedPeers {
		if peer.PeerId == "" {
			return fmt.Errorf("trusted peer has empty peer_id")
		}
		if peer.Address == "" {
			return fmt.Errorf("trusted peer %s has empty address", peer.PeerId)
		}
		if trustedPeerIDs[peer.PeerId] {
			return fmt.Errorf("duplicate trusted peer_id: %s", peer.PeerId)
		}
		trustedPeerIDs[peer.PeerId] = true
	}

	// Validate reputations
	reputationPeerIDs := make(map[string]bool)
	for _, rep := range gs.Reputations {
		if rep.PeerId == "" {
			return fmt.Errorf("reputation has empty peer_id")
		}
		if reputationPeerIDs[rep.PeerId] {
			return fmt.Errorf("duplicate reputation peer_id: %s", rep.PeerId)
		}
		if rep.Score < 0 || rep.Score > gs.Params.Reputation.MaxScore {
			return fmt.Errorf("reputation score for %s out of valid range", rep.PeerId)
		}
		reputationPeerIDs[rep.PeerId] = true
	}

	// Validate rate limits
	rateLimitPeerIDs := make(map[string]bool)
	for _, rl := range gs.RateLimits {
		if rl.PeerId == "" {
			return fmt.Errorf("rate limit has empty peer_id")
		}
		if rateLimitPeerIDs[rl.PeerId] {
			return fmt.Errorf("duplicate rate limit peer_id: %s", rl.PeerId)
		}
		rateLimitPeerIDs[rl.PeerId] = true
	}

	// Validate fork alerts
	forkAlertIDs := make(map[string]bool)
	for _, alert := range gs.ForkAlerts {
		if alert.AlertId == "" {
			return fmt.Errorf("fork alert has empty alert_id")
		}
		if forkAlertIDs[alert.AlertId] {
			return fmt.Errorf("duplicate fork alert_id: %s", alert.AlertId)
		}
		if alert.BlockHeight < 0 {
			return fmt.Errorf("fork alert %s has negative block height", alert.AlertId)
		}
		forkAlertIDs[alert.AlertId] = true
	}

	// Validate partition alerts
	partitionAlertIDs := make(map[string]bool)
	for _, alert := range gs.PartitionAlerts {
		if alert.AlertId == "" {
			return fmt.Errorf("partition alert has empty alert_id")
		}
		if partitionAlertIDs[alert.AlertId] {
			return fmt.Errorf("duplicate partition alert_id: %s", alert.AlertId)
		}
		partitionAlertIDs[alert.AlertId] = true
	}

	return nil
}
