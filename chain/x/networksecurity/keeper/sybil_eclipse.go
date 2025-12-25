// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

// SybilDetector detects potential Sybil attacks
type SybilDetector struct {
	// Track IP subnets to detect multiple nodes from same subnet
	subnetCounts map[string]uint32

	// Track ASN (Autonomous System Number) to detect concentration
	asnCounts map[uint32]uint32

	// Track geographic regions
	regionCounts map[string]uint32
}

// NewSybilDetector creates a new Sybil detector
func NewSybilDetector() *SybilDetector {
	return &SybilDetector{
		subnetCounts: make(map[string]uint32),
		asnCounts:    make(map[uint32]uint32),
		regionCounts: make(map[string]uint32),
	}
}

// AnalyzePeerDistribution analyzes peer distribution for Sybil attacks
func (sd *SybilDetector) AnalyzePeerDistribution(peers []types.PeerInfo) (isSybil bool, reason string) {
	sd.subnetCounts = make(map[string]uint32)
	sd.asnCounts = make(map[uint32]uint32)
	sd.regionCounts = make(map[string]uint32)

	totalPeers := len(peers)
	if totalPeers == 0 {
		return false, ""
	}

	// Don't apply concentration checks with very few peers (< 5)
	// Single peer or small peer sets aren't statistically significant
	if totalPeers < 5 {
		return false, ""
	}

	// Analyze distribution
	for _, peer := range peers {
		// Get subnet from IP
		subnet := getSubnet(peer.IpAddress)
		sd.subnetCounts[subnet]++

		// Track ASN
		if peer.Asn > 0 {
			sd.asnCounts[peer.Asn]++
		}

		// Track region
		if peer.Region != "" {
			sd.regionCounts[peer.Region]++
		}
	}

	// 1. Insufficient diversity check (less than 3 unique ASNs with >=5 peers)
	// This is checked first as it's a more fundamental issue
	uniqueASNs := len(sd.asnCounts)
	if uniqueASNs < 3 {
		return true, fmt.Sprintf("insufficient ASN diversity: only %d unique ASNs with %d peers", uniqueASNs, totalPeers)
	}

	// 2. Too many peers from same subnet (>30% from single /24 subnet)
	// Sort subnets for deterministic iteration order
	subnets := make([]string, 0, len(sd.subnetCounts))
	for subnet := range sd.subnetCounts {
		subnets = append(subnets, subnet)
	}
	sort.Strings(subnets)
	for _, subnet := range subnets {
		count := sd.subnetCounts[subnet]
		percentage := float64(count) / float64(totalPeers) * 100
		if percentage > 30.0 {
			return true, fmt.Sprintf("suspicious peer concentration: %.1f%% from subnet %s", percentage, subnet)
		}
	}

	// 3. Too many peers from same ASN (>40% from single ASN)
	// Sort ASNs for deterministic iteration order
	asns := make([]uint32, 0, len(sd.asnCounts))
	for asn := range sd.asnCounts {
		asns = append(asns, asn)
	}
	sort.Slice(asns, func(i, j int) bool { return asns[i] < asns[j] })
	for _, asn := range asns {
		count := sd.asnCounts[asn]
		percentage := float64(count) / float64(totalPeers) * 100
		if percentage > 40.0 {
			return true, fmt.Sprintf("suspicious peer concentration: %.1f%% from ASN %d", percentage, asn)
		}
	}

	// 4. Too many peers from same region (>50% from single region)
	// Sort regions for deterministic iteration order
	regions := make([]string, 0, len(sd.regionCounts))
	for region := range sd.regionCounts {
		regions = append(regions, region)
	}
	sort.Strings(regions)
	for _, region := range regions {
		count := sd.regionCounts[region]
		percentage := float64(count) / float64(totalPeers) * 100
		if percentage > 50.0 {
			return true, fmt.Sprintf("suspicious peer concentration: %.1f%% from region %s", percentage, region)
		}
	}

	return false, ""
}

// CheckSybilResistance performs Sybil resistance checks
func (k Keeper) CheckSybilResistance(ctx sdk.Context) error {
	peers := k.GetAllPeers(ctx)
	if len(peers) == 0 {
		return nil
	}

	detector := NewSybilDetector()
	isSybil, reason := detector.AnalyzePeerDistribution(peers)

	if isSybil {
		k.logger.Warn(fmt.Sprintf("Potential Sybil attack detected: %s", reason))

		// Create alert (could be stored as a custom alert type)
		// For now, log and potentially disconnect suspicious peers

		return types.ErrSybilDetected
	}

	return nil
}

// EclipseDetector detects potential Eclipse attacks
type EclipseDetector struct {
	// Expected minimum diverse connections
	minDiverseASNs uint32

	// Maximum percentage of connections from single entity
	maxConcentration float64
}

// NewEclipseDetector creates a new Eclipse detector
func NewEclipseDetector(minDiverseASNs uint32, maxConcentration float64) *EclipseDetector {
	return &EclipseDetector{
		minDiverseASNs:   minDiverseASNs,
		maxConcentration: maxConcentration,
	}
}

// DetectEclipse checks for Eclipse attack patterns
func (ed *EclipseDetector) DetectEclipse(peers []types.PeerInfo, trustedPeers []types.TrustedPeer) (isEclipse bool, reason string) {
	if len(peers) == 0 {
		return false, ""
	}

	// 1. Check if we have connections to any trusted peers
	hasTrustedConnection := false
	for _, peer := range peers {
		for _, trusted := range trustedPeers {
			if peer.PeerId == trusted.PeerId {
				hasTrustedConnection = true
				break
			}
		}
		if hasTrustedConnection {
			break
		}
	}

	// If we have trusted peers configured but none connected, suspicious
	if len(trustedPeers) > 0 && !hasTrustedConnection {
		return true, "no connections to any trusted peers"
	}

	// 2. Check ASN diversity
	asnCounts := make(map[uint32]uint32)
	for _, peer := range peers {
		if peer.Asn > 0 {
			asnCounts[peer.Asn]++
		}
	}

	uniqueASNs := uint32(len(asnCounts))
	if uniqueASNs < ed.minDiverseASNs {
		return true, fmt.Sprintf("insufficient ASN diversity: %d (min required: %d)", uniqueASNs, ed.minDiverseASNs)
	}

	// 3. Check for concentration from single ASN
	totalPeers := len(peers)
	// Sort ASNs for deterministic iteration order
	asns := make([]uint32, 0, len(asnCounts))
	for asn := range asnCounts {
		asns = append(asns, asn)
	}
	sort.Slice(asns, func(i, j int) bool { return asns[i] < asns[j] })
	for _, asn := range asns {
		count := asnCounts[asn]
		concentration := float64(count) / float64(totalPeers) * 100
		if concentration > ed.maxConcentration {
			return true, fmt.Sprintf("excessive concentration from ASN %d: %.1f%%", asn, concentration)
		}
	}

	// 4. Check for IP address concentration
	ipCounts := make(map[string]uint32)
	for _, peer := range peers {
		subnet := getSubnet(peer.IpAddress)
		if subnet != "" { // Only count valid subnets
			ipCounts[subnet]++
		}
	}

	// Only check subnet concentration if we have valid IP data
	if len(ipCounts) > 0 {
		// Sort subnets for deterministic iteration order
		subnets := make([]string, 0, len(ipCounts))
		for subnet := range ipCounts {
			subnets = append(subnets, subnet)
		}
		sort.Strings(subnets)
		for _, subnet := range subnets {
			count := ipCounts[subnet]
			concentration := float64(count) / float64(totalPeers) * 100
			if concentration > 30.0 { // Max 30% from single /24 subnet
				return true, fmt.Sprintf("excessive concentration from subnet %s: %.1f%%", subnet, concentration)
			}
		}
	}

	// 5. Check if all peers are outbound (could indicate we're isolated)
	outboundCount := 0
	for _, peer := range peers {
		if peer.ConnectionType == "outbound" {
			outboundCount++
		}
	}

	if outboundCount == totalPeers && totalPeers > 5 {
		return true, "all connections are outbound, no inbound connections accepted"
	}

	return false, ""
}

// CheckEclipseAttack performs Eclipse attack detection
func (k Keeper) CheckEclipseAttack(ctx sdk.Context) error {
	params, _ := k.GetParams(ctx)
	peers := k.GetAllPeers(ctx)
	trustedPeers := k.GetAllTrustedPeers(ctx)

	detector := NewEclipseDetector(
		params.Connection.MinPeerDiversity,
		40.0, // Max 40% from single ASN
	)

	isEclipse, reason := detector.DetectEclipse(peers, trustedPeers)

	if isEclipse {
		k.logger.Error(fmt.Sprintf("Potential Eclipse attack detected: %s", reason))

		// Could trigger alerts, attempt to connect to trusted peers, etc.
		return types.ErrEclipseDetected
	}

	return nil
}

// ValidateNewConnection validates a new peer connection against Sybil/Eclipse protection
func (k Keeper) ValidateNewConnection(ctx sdk.Context, peerInfo types.PeerInfo) error {
	params, _ := k.GetParams(ctx)

	// 1. Check if peer is banned
	if k.IsBanned(ctx, peerInfo.PeerId) {
		return types.ErrPeerBanned
	}

	// 2. If trusted-peers-only mode is enabled, check if peer is trusted
	if params.Connection.TrustedPeersOnly {
		if !k.IsTrustedPeer(ctx, peerInfo.PeerId) {
			return types.ErrNotTrustedPeer
		}
	}

	// 3. Check connection limits
	currentPeers := k.GetAllPeers(ctx)

	inboundCount := uint32(0)
	outboundCount := uint32(0)
	for _, peer := range currentPeers {
		if peer.ConnectionType == "inbound" {
			inboundCount++
		} else if peer.ConnectionType == "outbound" {
			outboundCount++
		}
	}

	if peerInfo.ConnectionType == "inbound" && inboundCount >= params.Connection.MaxInboundConnections {
		return types.ErrConnectionLimitExceeded
	}

	if peerInfo.ConnectionType == "outbound" && outboundCount >= params.Connection.MaxOutboundConnections {
		return types.ErrConnectionLimitExceeded
	}

	// 4. Check per-IP connection limits
	ipConnCount := k.GetConnectionCount(ctx, peerInfo.IpAddress)
	if ipConnCount >= params.Connection.MaxConnectionsPerIp {
		k.logger.Warn(fmt.Sprintf("Connection limit exceeded for IP %s: %d connections", peerInfo.IpAddress, ipConnCount))
		return types.ErrConnectionLimitExceeded
	}

	// 5. Check reputation (if exists)
	if reputation, found := k.GetReputation(ctx, peerInfo.PeerId); found {
		if reputation.Score < params.Reputation.MinScoreToConnect {
			k.logger.Warn(fmt.Sprintf("Peer %s has low reputation: %d", peerInfo.PeerId, reputation.Score))
			return types.ErrInvalidPeerReputation
		}
	}

	// 6. Validate IP address
	if !isValidIP(peerInfo.IpAddress) {
		return types.ErrInvalidIPAddress
	}

	// 7. Check for private/reserved IP ranges (should not accept from internet)
	// Skip this check if not enforcing strict IP validation (e.g., in test/dev environments)
	if params.Connection.TrustedPeersOnly {
		ip := net.ParseIP(peerInfo.IpAddress)
		if ip != nil && ip.IsLoopback() {
			k.logger.Warn(fmt.Sprintf("Rejecting connection from loopback IP: %s", peerInfo.IpAddress))
			return types.ErrInvalidIPAddress
		}
		// Allow private IPs when not in strict mode for testing/dev purposes
	}

	return nil
}

// AcceptConnection accepts a new peer connection after validation
func (k Keeper) AcceptConnection(ctx sdk.Context, peerInfo types.PeerInfo) error {
	// Validate connection
	if err := k.ValidateNewConnection(ctx, peerInfo); err != nil {
		return fmt.Errorf("error in AcceptConnection for validation: %w", err)
	}

	// Initialize reputation if not exists
	if _, found := k.GetReputation(ctx, peerInfo.PeerId); !found {
		params, _ := k.GetParams(ctx)
		reputation := types.NodeReputation{
			PeerId:            peerInfo.PeerId,
			Score:             params.Reputation.InitialScore,
			LastUpdatedHeight: ctx.BlockHeight(),
		}
		if err := k.SetReputation(ctx, reputation); err != nil {
			return fmt.Errorf("failed to initialize reputation: %w", err)
		}
	}

	// Store peer info
	peerInfo.ConnectedAt = ctx.BlockTime()
	if err := k.SetPeerInfo(ctx, peerInfo); err != nil {
		return fmt.Errorf("failed to store peer info: %w", err)
	}

	// Increment connection count for IP
	if err := k.IncrementConnectionCount(ctx, peerInfo.IpAddress); err != nil {
		return fmt.Errorf("failed to increment connection count: %w", err)
	}

	k.logger.Info(fmt.Sprintf("Accepted connection from peer %s (%s)", peerInfo.PeerId, peerInfo.IpAddress))

	return nil
}

// DisconnectPeer handles peer disconnection
func (k Keeper) DisconnectPeer(ctx sdk.Context, peerID string) error {
	peerInfo, found := k.GetPeerInfo(ctx, peerID)
	if !found {
		return types.ErrPeerNotFound
	}

	// Decrement connection count
	if err := k.DecrementConnectionCount(ctx, peerInfo.IpAddress); err != nil {
		return fmt.Errorf("failed to decrement connection count: %w", err)
	}

	// Remove peer info
	store := k.storeService.OpenKVStore(ctx)
	if err := store.Delete(types.GetPeerInfoKey(peerID)); err != nil {
		return fmt.Errorf("failed to OpenKVStore: %w", err)
	}

	// Clean up rate limiter and bandwidth tracker
	delete(k.rateLimiters, peerID)
	delete(k.bandwidthTrackers, peerID)

	k.logger.Info(fmt.Sprintf("Disconnected peer %s", peerID))

	return nil
}

// getSubnet extracts /24 subnet from IP address
func getSubnet(ipAddr string) string {
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return ""
	}

	// For IPv4, return /24 subnet
	if ip.To4() != nil {
		mask := net.CIDRMask(24, 32)
		network := ip.Mask(mask)
		return network.String() + "/24"
	}

	// For IPv6, return /48 subnet
	mask := net.CIDRMask(48, 128)
	network := ip.Mask(mask)
	return network.String() + "/48"
}

// isValidIP validates an IP address
func isValidIP(ipAddr string) bool {
	ip := net.ParseIP(ipAddr)
	return ip != nil
}

// CalculatePeerDiversity calculates the diversity score of connected peers
func (k Keeper) CalculatePeerDiversity(ctx sdk.Context) (diversityScore float64) {
	peers := k.GetAllPeers(ctx)
	if len(peers) == 0 {
		return 0
	}

	asnSet := make(map[uint32]bool)
	regionSet := make(map[string]bool)
	subnetSet := make(map[string]bool)

	for _, peer := range peers {
		if peer.Asn > 0 {
			asnSet[peer.Asn] = true
		}
		if peer.Region != "" {
			regionSet[peer.Region] = true
		}
		subnet := getSubnet(peer.IpAddress)
		if subnet != "" {
			subnetSet[subnet] = true
		}
	}

	totalPeers := float64(len(peers))
	asnDiversity := float64(len(asnSet)) / totalPeers
	regionDiversity := float64(len(regionSet)) / totalPeers
	subnetDiversity := float64(len(subnetSet)) / totalPeers

	// Weighted average of diversity metrics
	diversityScore = (asnDiversity*0.5 + regionDiversity*0.3 + subnetDiversity*0.2) * 100

	return diversityScore
}

// PerformPeerDiversityCheck checks if peer diversity is sufficient
func (k Keeper) PerformPeerDiversityCheck(ctx sdk.Context) error {
	params, _ := k.GetParams(ctx)
	peers := k.GetAllPeers(ctx)

	if len(peers) < int(params.Connection.MinPeerDiversity) {
		return nil // Not enough peers to check diversity
	}

	asnSet := make(map[uint32]bool)
	for _, peer := range peers {
		if peer.Asn > 0 {
			asnSet[peer.Asn] = true
		}
	}

	if uint32(len(asnSet)) < params.Connection.MinPeerDiversity {
		k.logger.Warn(fmt.Sprintf("Insufficient peer diversity: %d unique ASNs (min: %d)",
			len(asnSet), params.Connection.MinPeerDiversity))
		return types.ErrInsufficientPeerDiversity
	}

	return nil
}

// GeneratePeerID generates a deterministic peer ID from public key
func GeneratePeerID(publicKey []byte) string {
	hash := sha256.Sum256(publicKey)
	return hex.EncodeToString(hash[:])
}
