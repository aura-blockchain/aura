// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

const (
	// ModuleName is the name of the network security module
	ModuleName = "networksecurity"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the network security module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the network security module
	QuerierRoute = ModuleName
)

// KVStore key prefixes
var (
	ParamsKey                = []byte{0x01}
	PeerInfoPrefix           = []byte{0x02}
	TrustedPeerPrefix        = []byte{0x03}
	RateLimitPrefix          = []byte{0x04}
	ReputationPrefix         = []byte{0x05}
	ForkAlertPrefix          = []byte{0x06}
	PartitionAlertPrefix     = []byte{0x07}
	BannedPeerPrefix         = []byte{0x08}
	MempoolStatsKey          = []byte{0x09}
	ConnectionCountPrefix    = []byte{0x0a}
	GossipMessageCachePrefix = []byte{0x0b}
)

// GetPeerInfoKey returns the key for peer info
func GetPeerInfoKey(peerID string) []byte {
	return append(PeerInfoPrefix, []byte(peerID)...)
}

// GetTrustedPeerKey returns the key for trusted peer
func GetTrustedPeerKey(peerID string) []byte {
	return append(TrustedPeerPrefix, []byte(peerID)...)
}

// GetRateLimitKey returns the key for rate limit entry
func GetRateLimitKey(peerID string) []byte {
	return append(RateLimitPrefix, []byte(peerID)...)
}

// GetReputationKey returns the key for reputation
func GetReputationKey(peerID string) []byte {
	return append(ReputationPrefix, []byte(peerID)...)
}

// GetForkAlertKey returns the key for fork alert
func GetForkAlertKey(alertID string) []byte {
	return append(ForkAlertPrefix, []byte(alertID)...)
}

// GetPartitionAlertKey returns the key for partition alert
func GetPartitionAlertKey(alertID string) []byte {
	return append(PartitionAlertPrefix, []byte(alertID)...)
}

// GetBannedPeerKey returns the key for banned peer
func GetBannedPeerKey(peerID string) []byte {
	return append(BannedPeerPrefix, []byte(peerID)...)
}

// GetConnectionCountKey returns the key for connection count by IP
func GetConnectionCountKey(ipAddress string) []byte {
	return append(ConnectionCountPrefix, []byte(ipAddress)...)
}

// GetGossipMessageCacheKey returns the key for gossip message cache
func GetGossipMessageCacheKey(messageHash string) []byte {
	return append(GossipMessageCachePrefix, []byte(messageHash)...)
}

// Batch processing cursor keys
var (
	ThreatUpdateCursorKey      = []byte{0x0c}
	SecurityAlertCursorKey     = []byte{0x0d}
	ReputationRefreshCursorKey = []byte{0x0e}
)
