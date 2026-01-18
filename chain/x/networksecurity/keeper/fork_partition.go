// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

// ForkDetector detects blockchain forks
type ForkDetector struct {
	keeper              *Keeper
	blockHashCache      map[int64][]string // height -> list of block hashes
	heightDiffThreshold int64
}

// NewForkDetector creates a new fork detector
func NewForkDetector(k *Keeper, heightDiffThreshold int64) *ForkDetector {
	return &ForkDetector{
		keeper:              k,
		blockHashCache:      make(map[int64][]string),
		heightDiffThreshold: heightDiffThreshold,
	}
}

// DetectFork checks for blockchain forks
func (k Keeper) DetectFork(ctx sdk.Context, height int64, blockHash []byte) error {
	params, _ := k.GetParams(ctx)

	if !params.ForkDetection.EnableDetection {
		return nil
	}

	// Store block hash at this height
	hashStr := hex.EncodeToString(blockHash)

	// Check if we've seen a different hash at this height
	store := k.storeService.OpenKVStore(ctx)
	key := append([]byte("block_hash_"), sdk.Uint64ToBigEndian(uint64(height))...)

	existingHashBz, err := store.Get(key)
	if err == nil && existingHashBz != nil {
		existingHash := string(existingHashBz)

		if existingHash != hashStr {
			// Fork detected!
			k.logger.Error(fmt.Sprintf("Fork detected at height %d: existing hash %s != new hash %s",
				height, existingHash, hashStr))

			// Create fork alert
			detectedAt := ctx.BlockTime()
			alert := types.ForkAlert{
				AlertId:     generateAlertID("fork", height, ctx.BlockTime().Unix()),
				BlockHeight: height,
				ChainAHash:  existingHashBz,
				ChainBHash:  blockHash,
				DetectedAt:  detectedAt,
				Resolved:    false,
			}

			if err := k.SetForkAlert(ctx, alert); err != nil {
				return fmt.Errorf("failed to record fork alert: %w", err)
			}

			// If auto-resolution is enabled, attempt to resolve
			if params.ForkDetection.EnableAutoResolution {
				if err := k.ResolveFork(ctx, alert.AlertId); err != nil {
					k.logger.Error("auto-resolve fork failure", "alert_id", alert.AlertId, "error", err)
				}
			}

			return types.ErrForkDetected
		}
	} else {
		// First time seeing this height, store the hash
		if err := store.Set(key, []byte(hashStr)); err != nil {
			return fmt.Errorf("failed to store block hash: %w", err)
		}
	}

	return nil
}

// ResolveFork attempts to resolve a fork automatically
func (k Keeper) ResolveFork(ctx sdk.Context, alertID string) error {
	alert, found := k.GetForkAlert(ctx, alertID)
	if !found {
		return types.ErrAlertNotFound
	}

	if alert.Resolved {
		return types.ErrAlreadyResolved
	}

	params, _ := k.GetParams(ctx)

	// Wait for confirmation depth
	currentHeight := ctx.BlockHeight()
	if currentHeight-alert.BlockHeight < params.ForkDetection.ConfirmationDepth {
		return nil // Not enough confirmations yet
	}

	// In production, implement consensus-based fork resolution
	// For now, mark as resolved
	alert.Resolved = true
	alert.ResolutionDetails = fmt.Sprintf("Auto-resolved after %d confirmations", params.ForkDetection.ConfirmationDepth)

	if err := k.SetForkAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to update fork alert: %w", err)
	}

	k.logger.Info(fmt.Sprintf("Fork alert %s resolved", alertID))

	return nil
}

// SyncAttackDetector detects sync attacks and invalid block data
type SyncAttackDetector struct {
	keeper                  *Keeper
	invalidBlockCount       map[string]uint32 // peer_id -> invalid block count
	suspiciousResponseCount map[string]uint32 // peer_id -> suspicious response count
}

// NewSyncAttackDetector creates a new sync attack detector
func NewSyncAttackDetector(k *Keeper) *SyncAttackDetector {
	return &SyncAttackDetector{
		keeper:                  k,
		invalidBlockCount:       make(map[string]uint32),
		suspiciousResponseCount: make(map[string]uint32),
	}
}

// ValidateSyncData validates sync data from a peer
func (k Keeper) ValidateSyncData(ctx sdk.Context, peerID string, blockHeight int64, blockHash []byte, blockData []byte) error {
	params, _ := k.GetParams(ctx)

	// 1. Check if peer is banned
	if k.IsBanned(ctx, peerID) {
		return types.ErrPeerBanned
	}

	// 2. Validate block data size
	if uint64(len(blockData)) > params.Gossip.MaxMessageSize*10 { // Allow larger size for blocks
		k.logger.Warn(fmt.Sprintf("Oversized block data from %s: %d bytes", peerID, len(blockData)))
		k.PenalizeReputation(ctx, peerID, params.Reputation.MisbehaviorPenalty)
		return types.ErrMessageTooLarge
	}

	// 3. Validate block hash
	if !k.ValidateBlockHash(blockHeight, blockHash, blockData) {
		k.logger.Warn(fmt.Sprintf("Invalid block hash from %s at height %d", peerID, blockHeight))
		k.RecordInvalidBlock(ctx, peerID)
		return types.ErrSyncAttack
	}

	// 4. Check for suspicious patterns (e.g., always sending incorrect data)
	if k.IsSuspiciousSyncPeer(ctx, peerID) {
		k.logger.Warn(fmt.Sprintf("Peer %s showing suspicious sync behavior", peerID))
		banDuration := params.RateLimit.BanDuration
		if err := k.BanPeer(ctx, peerID, int64(banDuration.Seconds())*2, "suspicious sync behavior"); err != nil {
			k.logger.Error("failed to ban suspicious peer", "peer_id", peerID, "error", err)
		}
		return types.ErrSyncAttack
	}

	return nil
}

// ValidateBlockHash validates a block hash against block data
func (k Keeper) ValidateBlockHash(height int64, blockHash []byte, blockData []byte) bool {
	// Calculate hash of block data
	calculatedHash := sha256.Sum256(blockData)

	// Compare with provided hash
	providedHash := hex.EncodeToString(blockHash)
	calculatedHashStr := hex.EncodeToString(calculatedHash[:])

	return providedHash == calculatedHashStr
}

// RecordInvalidBlock records an invalid block from a peer
func (k Keeper) RecordInvalidBlock(ctx sdk.Context, peerID string) {
	params, _ := k.GetParams(ctx)

	// Penalize reputation severely for sending invalid blocks
	k.PenalizeReputation(ctx, peerID, params.Reputation.MisbehaviorPenalty*2)

	// Track invalid block count
	reputation, found := k.GetReputation(ctx, peerID)
	if found {
		reputation.MisbehaviorCount++
		if err := k.SetReputation(ctx, reputation); err != nil {
			k.logger.Error("failed to update reputation after invalid block", "peer_id", peerID, "error", err)
		}

		// If too many invalid blocks, ban the peer
		if reputation.MisbehaviorCount > 5 {
			banDuration := params.RateLimit.BanDuration
			if err := k.BanPeer(ctx, peerID, int64(banDuration.Seconds())*3, "repeated invalid blocks"); err != nil {
				k.logger.Error("failed to ban peer after repeated invalid blocks", "peer_id", peerID, "error", err)
			}
		}
	}
}

// IsSuspiciousSyncPeer checks if a peer exhibits suspicious sync behavior
func (k Keeper) IsSuspiciousSyncPeer(ctx sdk.Context, peerID string) bool {
	reputation, found := k.GetReputation(ctx, peerID)
	if !found {
		return false
	}

	// Check invalid message ratio using basis points (10000 = 100%) for determinism
	if reputation.MessagesReceived > 10 {
		// Calculate ratio in basis points: (invalid * 10000) / total
		invalidRatioBps := (reputation.InvalidMessages * 10000) / reputation.MessagesReceived
		if invalidRatioBps > 5000 { // More than 50% (5000 basis points) invalid messages
			return true
		}
	}

	// Check misbehavior count
	if reputation.MisbehaviorCount > 10 {
		return true
	}

	return false
}

// PartitionDetector detects network partitions
type PartitionDetector struct {
	keeper              *Keeper
	expectedPeerCount   uint32
	lastKnownPeerCount  uint32
	consecutiveLowPeers uint32
}

// NewPartitionDetector creates a new partition detector
func NewPartitionDetector(k *Keeper) *PartitionDetector {
	return &PartitionDetector{
		keeper: k,
	}
}

// DetectPartition checks for network partition
func (k Keeper) DetectPartition(ctx sdk.Context) error {
	params, _ := k.GetParams(ctx)

	if !params.PartitionDetection.EnableDetection {
		return nil
	}

	// Get current connected peers
	peers := k.GetAllPeers(ctx)
	currentPeerCount := uint32(len(peers))

	// Check if peer count is at or below threshold
	if currentPeerCount <= params.PartitionDetection.MinConnectedPeers {
		k.logger.Warn(fmt.Sprintf("Low peer count detected: %d (min: %d)",
			currentPeerCount, params.PartitionDetection.MinConnectedPeers))

		// Check if this is a sudden drop using basis points (10000 = 100%) for determinism
		expectedPeers := k.GetExpectedPeerCount(ctx)
		if expectedPeers > 0 && currentPeerCount < expectedPeers {
			// Calculate drop percentage in basis points: ((expected - current) * 10000) / expected
			dropPercentBps := ((expectedPeers - currentPeerCount) * 10000) / expectedPeers

			if dropPercentBps > 5000 { // More than 50% drop (5000 basis points)
				// Potential partition detected
				k.logger.Error(fmt.Sprintf("Potential network partition: %d.%02d%% peer drop", dropPercentBps/100, dropPercentBps%100))

				// Create partition alert
				detectedAt := ctx.BlockTime()
				alert := types.PartitionAlert{
					AlertId:        generateAlertID("partition", ctx.BlockHeight(), ctx.BlockTime().Unix()),
					ConnectedPeers: currentPeerCount,
					ExpectedPeers:  expectedPeers,
					MissingPeerIds: k.GetMissingPeerIDs(ctx),
					DetectedAt:     detectedAt,
					Resolved:       false,
				}

				if err := k.SetPartitionAlert(ctx, alert); err != nil {
					return fmt.Errorf("failed to record partition alert: %w", err)
				}

				return types.ErrPartitionDetected
			}
		}
	} else {
		// Update expected peer count (running average)
		k.UpdateExpectedPeerCount(ctx, currentPeerCount)
	}

	return nil
}

// GetExpectedPeerCount retrieves the expected peer count
func (k Keeper) GetExpectedPeerCount(ctx sdk.Context) uint32 {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte("expected_peer_count"))
	if err != nil || bz == nil {
		return 0
	}

	if len(bz) != 8 {
		return 0
	}
	return uint32(sdk.BigEndianToUint64(bz))
}

// UpdateExpectedPeerCount updates the expected peer count
func (k Keeper) UpdateExpectedPeerCount(ctx sdk.Context, currentCount uint32) {
	expectedCount := k.GetExpectedPeerCount(ctx)

	// Calculate running average (80% old, 20% new) using integer math for determinism
	// Formula: new_expected = (old * 8 + current * 2) / 10
	if expectedCount == 0 {
		expectedCount = currentCount
	} else {
		expectedCount = (expectedCount*8 + currentCount*2) / 10
	}

	store := k.storeService.OpenKVStore(ctx)
	bz := sdk.Uint64ToBigEndian(uint64(expectedCount))
	if err := store.Set([]byte("expected_peer_count"), bz); err != nil {
		k.logger.Error("failed to persist expected peer count", "err", err)
	}
}

// GetMissingPeerIDs identifies peers that were connected but are now missing
func (k Keeper) GetMissingPeerIDs(ctx sdk.Context) []string {
	// Get list of previously known peers
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte("known_peer_list"))
	if err != nil || bz == nil {
		return []string{}
	}

	knownPeers := make([]string, 0, 64)
	if err := json.Unmarshal(bz, &knownPeers); err != nil {
		return []string{}
	}

	// Get currently connected peers
	currentPeers := k.GetAllPeers(ctx)
	currentPeerMap := make(map[string]bool)
	for _, peer := range currentPeers {
		currentPeerMap[peer.PeerId] = true
	}

	// Find missing peers
	missingPeers := make([]string, 0, 64)
	for _, peerID := range knownPeers {
		if !currentPeerMap[peerID] {
			missingPeers = append(missingPeers, peerID)
		}
	}

	return missingPeers
}

// UpdateKnownPeerList updates the list of known peers
func (k Keeper) UpdateKnownPeerList(ctx sdk.Context) {
	peers := k.GetAllPeers(ctx)

	peerIDs := make([]string, 0, 64)
	for _, peer := range peers {
		peerIDs = append(peerIDs, peer.PeerId)
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := json.Marshal(peerIDs)
	if err != nil {
		k.logger.Error("failed to marshal known peers", "error", err)
		return
	}
	if err := store.Set([]byte("known_peer_list"), bz); err != nil {
		k.logger.Error("failed to persist known peers", "error", err)
	}
}

// ResolvePartitionAlert resolves a partition alert
func (k Keeper) ResolvePartitionAlert(ctx sdk.Context, alertID string) error {
	alert, found := k.GetPartitionAlert(ctx, alertID)
	if !found {
		return types.ErrAlertNotFound
	}

	if alert.Resolved {
		return types.ErrAlreadyResolved
	}

	alert.Resolved = true
	if err := k.SetPartitionAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to update partition alert: %w", err)
	}

	k.logger.Info(fmt.Sprintf("Partition alert %s resolved", alertID))

	return nil
}

// PerformNetworkHealthCheck performs comprehensive network health check
func (k Keeper) PerformNetworkHealthCheck(ctx sdk.Context) error {
	// 1. Check for forks
	// This would be called when receiving new blocks

	// 2. Check for network partition
	if err := k.DetectPartition(ctx); err != nil {
		return fmt.Errorf("error in PerformNetworkHealthCheck: %w", err)
	}

	// 3. Check Sybil resistance
	if err := k.CheckSybilResistance(ctx); err != nil {
		return fmt.Errorf("error in PerformNetworkHealthCheck: %w", err)
	}

	// 4. Check Eclipse attack
	if err := k.CheckEclipseAttack(ctx); err != nil {
		return fmt.Errorf("error in PerformNetworkHealthCheck: %w", err)
	}

	// 5. Check peer diversity
	if err := k.PerformPeerDiversityCheck(ctx); err != nil {
		return fmt.Errorf("error in PerformNetworkHealthCheck: %w", err)
	}

	// 6. Update known peer list
	k.UpdateKnownPeerList(ctx)

	return nil
}

// generateAlertID generates a unique alert ID
func generateAlertID(alertType string, height int64, timestamp int64) string {
	data := []byte(fmt.Sprintf("%s_%d_%d", alertType, height, timestamp))
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:16])
}

// CleanupResolvedAlerts removes old resolved alerts
func (k Keeper) CleanupResolvedAlerts(ctx sdk.Context) {
	// Cleanup fork alerts older than 1000 blocks
	forkAlerts := k.GetAllForkAlerts(ctx, true)
	for _, alert := range forkAlerts {
		if alert.Resolved && ctx.BlockHeight()-alert.BlockHeight > 1000 {
			store := k.storeService.OpenKVStore(ctx)
			if err := store.Delete(types.GetForkAlertKey(alert.AlertId)); err != nil {
				k.logger.Error("failed to cleanup fork alert", "alert_id", alert.AlertId, "error", err)
			}
		}
	}

	// Cleanup partition alerts older than 1 hour
	partitionAlerts := k.GetAllPartitionAlerts(ctx, true)
	for _, alert := range partitionAlerts {
		if alert.Resolved && !alert.DetectedAt.IsZero() && ctx.BlockTime().Sub(alert.DetectedAt) > time.Hour {
			store := k.storeService.OpenKVStore(ctx)
			if err := store.Delete(types.GetPartitionAlertKey(alert.AlertId)); err != nil {
				k.logger.Error("failed to cleanup partition alert", "alert_id", alert.AlertId, "error", err)
			}
		}
	}
}
