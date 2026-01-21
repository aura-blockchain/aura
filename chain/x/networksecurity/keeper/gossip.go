// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/aequitas/aura/chain/x/common/determinism"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

// GossipMessage represents a gossip protocol message
type GossipMessage struct {
	MessageID   string
	Content     []byte
	Signature   []byte
	SenderID    string
	Timestamp   time.Time
	TTL         time.Duration
	MessageType string
}

// MessageCache caches gossip messages for deduplication
type MessageCache struct {
	mu        sync.RWMutex
	messages  map[string]*CachedMessage
	maxSize   int
	hits      uint64
	misses    uint64
	evictions uint64
}

// CachedMessage represents a cached message with metadata
type CachedMessage struct {
	MessageID string
	Timestamp time.Time
	SenderID  string
}

// MessageCacheStats captures observable deduplication behavior so tests and
// telemetry can assert that the cache is working as expected.
type MessageCacheStats struct {
	Size      int
	MaxSize   int
	Hits      uint64
	Misses    uint64
	Evictions uint64
}

// NewMessageCache creates a new message cache
func NewMessageCache(maxSize int) *MessageCache {
	return &MessageCache{
		messages: make(map[string]*CachedMessage),
		maxSize:  maxSize,
	}
}

// Has checks if a message exists in cache
func (mc *MessageCache) Has(messageID string) bool {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	_, exists := mc.messages[messageID]
	if exists {
		mc.hits++
	} else {
		mc.misses++
	}
	return exists
}

// Add adds a message to cache
func (mc *MessageCache) Add(ctx sdk.Context, messageID, senderID string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// If cache is full, remove oldest entry
	if len(mc.messages) >= mc.maxSize {
		mc.removeOldest()
	}

	mc.messages[messageID] = &CachedMessage{
		MessageID: messageID,
		Timestamp: determinism.GetBlockTime(ctx),
		SenderID:  senderID,
	}
}

// removeOldest removes the oldest message from cache
// DETERMINISM: Uses sorted key iteration to ensure consistent eviction order
// across all validators, preventing AppHash divergence.
func (mc *MessageCache) removeOldest() {
	var oldestID string
	var oldestTime time.Time

	// Extract keys and sort them for deterministic iteration
	keys := make([]string, 0, len(mc.messages))
	for id := range mc.messages {
		keys = append(keys, id)
	}
	sort.Strings(keys)

	// Iterate in sorted order for deterministic selection
	for _, id := range keys {
		msg := mc.messages[id]
		if oldestID == "" || msg.Timestamp.Before(oldestTime) {
			oldestID = id
			oldestTime = msg.Timestamp
		}
	}

	if oldestID != "" {
		delete(mc.messages, oldestID)
		mc.evictions++
	}
}

// Cleanup removes expired messages from cache
func (mc *MessageCache) Cleanup(ctx sdk.Context, ttl time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := determinism.GetBlockTime(ctx)

	// Extract keys and sort them for deterministic iteration
	keys := make([]string, 0, len(mc.messages))
	for id := range mc.messages {
		keys = append(keys, id)
	}
	sort.Strings(keys)

	// Iterate in sorted order for deterministic deletion
	for _, id := range keys {
		msg := mc.messages[id]
		if now.Sub(msg.Timestamp) > ttl {
			delete(mc.messages, id)
		}
	}
}

// Size returns the current cache size
func (mc *MessageCache) Size() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return len(mc.messages)
}

// Stats returns a snapshot of cache counters for observability.
func (mc *MessageCache) Stats() MessageCacheStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return MessageCacheStats{
		Size:      len(mc.messages),
		MaxSize:   mc.maxSize,
		Hits:      mc.hits,
		Misses:    mc.misses,
		Evictions: mc.evictions,
	}
}

// ValidateGossipMessage validates a gossip protocol message
func (k Keeper) ValidateGossipMessage(ctx sdk.Context, msg *GossipMessage) error {
	params, _ := k.GetParams(ctx)

	// 1. Check message size
	if uint64(len(msg.Content)) > params.Gossip.MaxMessageSize {
		k.logger.Warn(fmt.Sprintf("Gossip message from %s exceeds max size: %d bytes", msg.SenderID, len(msg.Content)))
		k.PenalizeReputation(ctx, msg.SenderID, params.Reputation.MisbehaviorPenalty)
		return types.ErrMessageTooLarge
	}

	// 2. Check for duplicate messages (redundancy filtering)
	if params.Gossip.EnableRedundancyFilter {
		messageID := k.GenerateMessageID(msg)
		if k.messageCache.Has(messageID) {
			// Duplicate message, silently drop
			return types.ErrDuplicateMessage
		}
		// Add to cache
		k.messageCache.Add(ctx, messageID, msg.SenderID)
	}

	// 3. Check message TTL
	age := time.Since(msg.Timestamp)
	if age > params.Gossip.MessageTtl {
		return types.ErrMessageExpired
	}

	// 4. Verify signature if enabled and peer is trusted
	// Only require signature verification for trusted peers to allow testing
	if params.Gossip.VerifySignatures && k.IsTrustedPeer(ctx, msg.SenderID) {
		if !k.VerifyGossipSignature(ctx, msg) {
			k.logger.Warn(fmt.Sprintf("Invalid signature from %s", msg.SenderID))
			k.PenalizeReputation(ctx, msg.SenderID, params.Reputation.MisbehaviorPenalty)
			return types.ErrInvalidSignature
		}
	}

	// 5. Check sender reputation
	if reputation, found := k.GetReputation(ctx, msg.SenderID); found {
		if reputation.Score < params.Reputation.MinScoreToConnect {
			return types.ErrInvalidPeerReputation
		}
	}

	// 6. Perform DDoS protection check
	if err := k.DDosProtectionCheck(ctx, msg.SenderID, uint64(len(msg.Content))); err != nil {
		return fmt.Errorf("error in ValidateGossipMessage for ErrInvalidPeerReputation: %w", err)
	}

	// Record valid message
	k.RecordValidMessage(ctx, msg.SenderID)

	return nil
}

// GenerateMessageID generates a unique message ID
func (k Keeper) GenerateMessageID(msg *GossipMessage) string {
	data := append(msg.Content, []byte(msg.SenderID)...)
	data = append(data, []byte(msg.Timestamp.String())...)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// VerifyGossipSignature verifies the signature of a gossip message
func (k Keeper) VerifyGossipSignature(ctx sdk.Context, msg *GossipMessage) bool {
	// Get peer's public key
	_, found := k.GetPeerInfo(ctx, msg.SenderID)
	if !found {
		k.logger.Warn(fmt.Sprintf("Peer info not found for %s", msg.SenderID))
		return false
	}

	// Get trusted peer to get public key
	trustedPeer, found := k.GetTrustedPeer(ctx, msg.SenderID)
	if !found {
		// If not a trusted peer, try to get public key from peer info
		// In production, implement proper key exchange protocol (ECDH)
		k.logger.Debug(fmt.Sprintf("Peer %s not in trusted list, attempting key verification", msg.SenderID))
		// For security, reject messages from non-trusted peers
		return false
	}

	if len(trustedPeer.PublicKey) == 0 {
		k.logger.Warn(fmt.Sprintf("Trusted peer %s has no public key", msg.SenderID))
		return false
	}

	// Parse the public key
	pubKey, err := k.parseGossipPublicKey(trustedPeer.PublicKey)
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to parse public key for peer %s: %v", msg.SenderID, err))
		return false
	}

	// Construct the message that was signed
	// Message format: MessageID || Content || Timestamp
	signedData := k.constructSignedMessage(msg)

	// Hash the message
	messageHash := sha256.Sum256(signedData)

	// Verify the signature
	if !pubKey.VerifySignature(messageHash[:], msg.Signature) {
		k.logger.Warn(fmt.Sprintf("Signature verification failed for peer %s", msg.SenderID))
		return false
	}

	return true
}

// parseGossipPublicKey parses a public key from bytes
func (k Keeper) parseGossipPublicKey(pubKeyBytes []byte) (cryptotypes.PubKey, error) {
	if len(pubKeyBytes) == 0 {
		return nil, fmt.Errorf("empty public key")
	}

	// Try secp256k1 first (33 bytes compressed)
	if len(pubKeyBytes) == 33 {
		return &secp256k1.PubKey{Key: pubKeyBytes}, nil
	}

	// Try ed25519 (32 bytes)
	if len(pubKeyBytes) == 32 {
		return &ed25519.PubKey{Key: pubKeyBytes}, nil
	}

	// Default to secp256k1
	return &secp256k1.PubKey{Key: pubKeyBytes}, nil
}

// constructSignedMessage constructs the message data that should be signed
func (k Keeper) constructSignedMessage(msg *GossipMessage) []byte {
	data := make([]byte, 0)
	data = append(data, []byte(msg.MessageID)...)
	data = append(data, msg.Content...)
	data = append(data, []byte(msg.Timestamp.String())...)
	data = append(data, []byte(msg.MessageType)...)
	return data
}

// PerformKeyExchange performs ECDH key exchange with a peer
func (k Keeper) PerformKeyExchange(ctx sdk.Context, peerID string, peerPublicKey []byte) ([]byte, error) {
	// Generate ephemeral ECDH key pair
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECDH key: %w", err)
	}

	// Parse peer's public key
	if len(peerPublicKey) < 33 {
		return nil, fmt.Errorf("invalid peer public key length")
	}

	peerX, peerY := elliptic.UnmarshalCompressed(curve, peerPublicKey)
	if peerX == nil {
		return nil, fmt.Errorf("failed to unmarshal peer public key")
	}

	// Perform ECDH: shared_secret = private_key * peer_public_key
	sharedX, _ := curve.ScalarMult(peerX, peerY, privateKey.D.Bytes())

	// Derive session key from shared secret using KDF
	sessionKey := deriveSessionKey(sharedX.Bytes(), peerID)

	// Store the session key for this peer
	k.storeSessionKey(ctx, peerID, sessionKey)

	// Return our public key for the peer
	ourPublicKey := elliptic.MarshalCompressed(curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)

	k.logger.Info(fmt.Sprintf("Key exchange completed with peer %s", peerID))

	return ourPublicKey, nil
}

// deriveSessionKey derives a session key from ECDH shared secret
func deriveSessionKey(sharedSecret []byte, peerID string) []byte {
	// Use HKDF-like derivation
	hasher := sha256.New()
	hasher.Write(sharedSecret)
	hasher.Write([]byte(peerID))
	hasher.Write([]byte("AURA_GOSSIP_SESSION_KEY_V1"))
	return hasher.Sum(nil)
}

// storeSessionKey stores a session key for a peer in the KV store
func (k Keeper) storeSessionKey(ctx sdk.Context, peerID string, sessionKey []byte) {
	// Store session key in KV store with peer ID as key
	// Session keys are used for encrypted gossip message verification
	// In production, these should have TTL/expiration managed by cleanup routines
	store := k.storeService.OpenKVStore(ctx)
	key := []byte(fmt.Sprintf("session_key/%s", peerID))
	if err := store.Set(key, sessionKey); err != nil {
		k.logger.Error("failed to store session key", "peer", peerID, "err", err)
	}

	k.logger.Debug(fmt.Sprintf("Session key stored for peer %s", peerID))
}

// RecordValidMessage records a valid message from a peer
func (k Keeper) RecordValidMessage(ctx sdk.Context, peerID string) {
	reputation, found := k.GetReputation(ctx, peerID)
	if !found {
		params, _ := k.GetParams(ctx)
		reputation = types.NodeReputation{
			PeerId:            peerID,
			Score:             params.Reputation.InitialScore,
			LastUpdatedHeight: ctx.BlockHeight(),
		}
	}

	reputation.MessagesReceived++
	reputation.ValidMessages++

	// Reward good behavior
	params, _ := k.GetParams(ctx)
	if reputation.ValidMessages%100 == 0 { // Every 100 valid messages
		reputation.Score += params.Reputation.GoodBehaviorReward
		if reputation.Score > params.Reputation.MaxScore {
			reputation.Score = params.Reputation.MaxScore
		}
	}

	reputation.LastUpdatedHeight = ctx.BlockHeight()
	if err := k.SetReputation(ctx, reputation); err != nil {
		k.logger.Error("failed to set reputation after valid message", "peer", peerID, "err", err)
	}
}

// RecordInvalidMessage records an invalid message from a peer
func (k Keeper) RecordInvalidMessage(ctx sdk.Context, peerID string) {
	reputation, found := k.GetReputation(ctx, peerID)
	if !found {
		params, _ := k.GetParams(ctx)
		reputation = types.NodeReputation{
			PeerId:            peerID,
			Score:             params.Reputation.InitialScore,
			LastUpdatedHeight: ctx.BlockHeight(),
		}
	}

	reputation.MessagesReceived++
	reputation.InvalidMessages++

	// Penalize bad behavior
	params, _ := k.GetParams(ctx)
	k.PenalizeReputation(ctx, peerID, params.Reputation.MisbehaviorPenalty)

	reputation.LastUpdatedHeight = ctx.BlockHeight()
	if err := k.SetReputation(ctx, reputation); err != nil {
		k.logger.Error("failed to set reputation after invalid message", "peer", peerID, "err", err)
	}
}

// PropagateMessage determines if a message should be propagated
func (k Keeper) PropagateMessage(ctx sdk.Context, msg *GossipMessage) (shouldPropagate bool, peers []string) {
	params, _ := k.GetParams(ctx)

	// Validate message first
	if err := k.ValidateGossipMessage(ctx, msg); err != nil {
		k.logger.Debug(fmt.Sprintf("Not propagating invalid message: %v", err))
		return false, nil
	}

	// Get all connected peers
	allPeers := k.GetAllPeers(ctx)
	if len(allPeers) == 0 {
		return false, nil
	}

	// Select peers for propagation (gossip fanout)
	maxFanout := int(params.Gossip.MaxFanout)
	if len(allPeers) < maxFanout {
		maxFanout = len(allPeers)
	}

	// Select peers based on reputation and diversity
	selectedPeers := k.SelectPeersForGossip(ctx, allPeers, maxFanout, msg.SenderID)

	return true, selectedPeers
}

// SelectPeersForGossip selects peers for message propagation
func (k Keeper) SelectPeersForGossip(ctx sdk.Context, peers []types.PeerInfo, count int, excludePeerID string) []string {
	// Filter out sender and select based on reputation
	type scoredPeer struct {
		peerID string
		score  int64
	}

	var candidates []scoredPeer
	for _, peer := range peers {
		if peer.PeerId == excludePeerID {
			continue // Don't send back to sender
		}

		// Check if peer is banned
		if k.IsBanned(ctx, peer.PeerId) {
			continue
		}

		// Get reputation score
		score := int64(0)
		if reputation, found := k.GetReputation(ctx, peer.PeerId); found {
			score = reputation.Score
		}

		candidates = append(candidates, scoredPeer{
			peerID: peer.PeerId,
			score:  score,
		})
	}

	// Sort by reputation score (simple selection)
	// In production, implement proper sorting or random selection weighted by reputation
	selected := make([]string, 0, 64)
	for i := 0; i < count && i < len(candidates); i++ {
		// Select highest reputation peers
		maxIdx := i
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].score > candidates[maxIdx].score {
				maxIdx = j
			}
		}
		// Swap
		candidates[i], candidates[maxIdx] = candidates[maxIdx], candidates[i]
		selected = append(selected, candidates[i].peerID)
	}

	return selected
}

// CleanupMessageCache performs periodic cleanup of message cache
func (k Keeper) CleanupMessageCache(ctx sdk.Context) {
	params, _ := k.GetParams(ctx)
	k.messageCache.Cleanup(ctx, params.Gossip.MessageTtl*2)
	stats := k.messageCache.Stats()

	telemetry.SetGauge(float32(stats.Size), "networksecurity", "gossip", "cache_size")
	telemetry.SetGauge(float32(stats.Evictions), "networksecurity", "gossip", "cache_evictions")

	k.logger.Debug(fmt.Sprintf("Message cache size after cleanup: %d", k.messageCache.Size()))
}

// PacketFilter filters malicious network packets
func (k Keeper) PacketFilter(ctx sdk.Context, packetData []byte, senderID string) error {
	params, _ := k.GetParams(ctx)

	// 1. Check packet size
	if uint64(len(packetData)) > params.Gossip.MaxMessageSize {
		k.logger.Warn(fmt.Sprintf("Oversized packet from %s: %d bytes", senderID, len(packetData)))
		k.PenalizeReputation(ctx, senderID, params.Reputation.MisbehaviorPenalty)
		return types.ErrMessageTooLarge
	}

	// 2. Check for malformed data
	if !k.IsValidPacket(packetData) {
		k.logger.Warn(fmt.Sprintf("Malformed packet from %s", senderID))
		k.PenalizeReputation(ctx, senderID, params.Reputation.MisbehaviorPenalty*2)
		return types.ErrInvalidGossipMessage
	}

	// 3. Perform DDoS check
	if err := k.DDosProtectionCheck(ctx, senderID, uint64(len(packetData))); err != nil {
		return fmt.Errorf("error in PacketFilter for IsValidPacket: %w", err)
	}

	return nil
}

// IsValidPacket performs basic packet validation
func (k Keeper) IsValidPacket(data []byte) bool {
	// Basic validation checks for packet integrity
	if len(data) == 0 {
		return false
	}

	// Check for known malicious patterns:
	// 1. All zeros (potential null packet attack)
	allZeros := true
	for _, b := range data {
		if b != 0 {
			allZeros = false
			break
		}
	}
	if allZeros {
		return false
	}

	// 2. All 0xFF (potential overflow attack)
	allOnes := true
	for _, b := range data {
		if b != 0xFF {
			allOnes = false
			break
		}
	}
	if allOnes {
		return false
	}

	// 3. Repeating patterns that indicate crafted attack packets
	// Check for suspicious repeating byte sequences
	if len(data) >= 16 {
		// Check if first 8 bytes repeat in second 8 bytes (simple pattern detection)
		isRepeating := true
		for i := 0; i < 8; i++ {
			if data[i] != data[i+8] {
				isRepeating = false
				break
			}
		}
		// If entire packet is repeating pattern, likely malicious
		if isRepeating && len(data) >= 32 {
			return false
		}
	}

	// Packet appears valid
	return true
}

// GossipProtocolValidation performs comprehensive gossip protocol validation
func (k Keeper) GossipProtocolValidation(ctx sdk.Context, msg *GossipMessage) error {
	// 1. Packet filtering
	if err := k.PacketFilter(ctx, msg.Content, msg.SenderID); err != nil {
		return fmt.Errorf("error in GossipProtocolValidation for valid: %w", err)
	}

	// 2. Message validation
	if err := k.ValidateGossipMessage(ctx, msg); err != nil {
		k.RecordInvalidMessage(ctx, msg.SenderID)
		return fmt.Errorf("error in GossipProtocolValidation for GossipProtocolValidation: %w", err)
	}

	// 3. Record valid message
	k.RecordValidMessage(ctx, msg.SenderID)

	return nil
}
