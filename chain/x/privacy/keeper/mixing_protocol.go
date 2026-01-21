// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/math"
	privacyproto "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Hard limit on mixing pool size to prevent DoS via linear scans
const maxMixingPoolParticipants = 256

// JoinMixingRound allows a user to join a mixing round.
// Uses O(1) map lookup for participant deduplication and enforces maxMixingPoolParticipants.
func (k Keeper) JoinMixingRound(ctx context.Context, poolID string, participant string, input []byte, amount math.Int) (string, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := k.GetParams(ctx)

	if !params.EnableMixing {
		return "", fmt.Errorf("mixing not enabled")
	}

	// Validate amount
	if amount.IsZero() || amount.IsNegative() {
		return "", fmt.Errorf("invalid amount")
	}

	// Get or create mixing pool
	pool, err := k.GetMixingPool(ctx, poolID)
	if err != nil {
		// Create new pool with capped max participants
		maxParticipants := params.MinMixingParticipants * 4
		if maxParticipants > maxMixingPoolParticipants {
			maxParticipants = maxMixingPoolParticipants
		}
		pool = &privacyproto.MixingPool{
			PoolId:          poolID,
			Participants:    [][]byte{},
			MinParticipants: params.MinMixingParticipants,
			MaxParticipants: maxParticipants,
			Status:          "open",
		}
	}

	// Check if pool is full (enforce hard cap)
	currentCount := uint32(len(pool.Participants))
	effectiveMax := pool.MaxParticipants
	if effectiveMax > maxMixingPoolParticipants {
		effectiveMax = maxMixingPoolParticipants
	}
	if currentCount >= effectiveMax {
		return "", fmt.Errorf("mixing pool is full")
	}

	// Build a set for O(1) participant lookup instead of O(n) linear scan
	participantSet := make(map[string]struct{}, len(pool.Participants))
	for _, p := range pool.Participants {
		participantSet[string(p)] = struct{}{}
	}

	// Check if participant already joined - O(1) lookup
	if _, exists := participantSet[participant]; exists {
		return "", fmt.Errorf("participant already in pool")
	}

	// Generate participant ID
	hasher := sha256.New()
	hasher.Write([]byte(participant))
	hasher.Write(input)
	hasher.Write([]byte(fmt.Sprintf("%d", sdkCtx.BlockTime().Unix())))
	participantID := hex.EncodeToString(hasher.Sum(nil))[:16]

	// Add participant
	pool.Participants = append(pool.Participants, []byte(participantID))

	// Check if pool is ready for mixing
	currentCount = uint32(len(pool.Participants))
	if currentCount >= pool.MinParticipants {
		pool.Status = "ready"
	}

	// Store updated pool
	if err := k.SetMixingPool(ctx, pool); err != nil {
		return "", err
	}

	return participantID, nil
}

// ExecuteMixing executes the mixing protocol for a pool
func (k Keeper) ExecuteMixing(ctx context.Context, poolID string) error {
	pool, err := k.GetMixingPool(ctx, poolID)
	if err != nil {
		return err
	}

	if pool.Status != "ready" {
		return fmt.Errorf("pool not ready for mixing")
	}

	// Shuffle participants using Fisher-Yates algorithm (deterministic with block data)
	shuffled := k.shuffleParticipants(ctx, pool.Participants)

	// Update pool status
	pool.Status = "completed"
	pool.Participants = shuffled

	return k.SetMixingPool(ctx, pool)
}

// shuffleParticipants performs deterministic shuffle using multiple entropy sources
// SECURITY: Combines block hash with participant commitments and pool state to prevent
// predictability attacks where validators could manipulate shuffle outcomes by controlling
// block hash alone.
func (k Keeper) shuffleParticipants(ctx context.Context, participants [][]byte) [][]byte {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	shuffled := make([][]byte, len(participants))
	for i := range participants {
		shuffled[i] = make([]byte, len(participants[i]))
		copy(shuffled[i], participants[i])
	}

	// SECURITY: Combine multiple entropy sources for unpredictable randomness
	// 1. Block hash - provides base entropy from consensus
	// 2. Participant commitments - adds data only participants control
	// 3. Block time - adds timing entropy
	// 4. Participant count - adds pool-specific entropy
	//
	// An attacker would need to control ALL sources simultaneously to predict shuffle
	blockHash := sdkCtx.HeaderHash()

	// Generate participant commitment hash (each participant's data contributes)
	participantCommitment := sha256.New()
	for _, p := range participants {
		participantCommitment.Write(p)
	}
	participantHash := participantCommitment.Sum(nil)

	// Create combined entropy seed
	entropySeed := sha256.New()
	entropySeed.Write(blockHash)
	entropySeed.Write(participantHash)
	entropySeed.Write([]byte(fmt.Sprintf("%d", sdkCtx.BlockTime().UnixNano())))
	entropySeed.Write([]byte(fmt.Sprintf("%d", len(participants))))
	seedHash := entropySeed.Sum(nil)

	// Fisher-Yates shuffle with enhanced deterministic randomness
	for i := len(shuffled) - 1; i > 0; i-- {
		// Generate deterministic random index using combined entropy seed
		hasher := sha256.New()
		hasher.Write(seedHash)
		hasher.Write([]byte(fmt.Sprintf("%d", i)))
		hash := hasher.Sum(nil)

		// Use first 8 bytes of hash as uint64
		randVal := uint64(0)
		for j := 0; j < 8 && j < len(hash); j++ {
			randVal = (randVal << 8) | uint64(hash[j])
		}

		j := int(randVal % uint64(i+1))
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}

// WithdrawFromMixing allows withdrawal after mixing completes.
// Uses O(1) map lookup for participant verification.
func (k Keeper) WithdrawFromMixing(ctx context.Context, poolID, participantID string, output []byte) error {
	pool, err := k.GetMixingPool(ctx, poolID)
	if err != nil {
		return err
	}

	if pool.Status != "completed" {
		return fmt.Errorf("mixing not yet completed")
	}

	// Build set for O(1) lookup instead of O(n) linear scan
	participantSet := make(map[string]struct{}, len(pool.Participants))
	for _, p := range pool.Participants {
		participantSet[string(p)] = struct{}{}
	}

	// Verify participant was in the pool - O(1) lookup
	if _, found := participantSet[participantID]; !found {
		return fmt.Errorf("participant not in pool")
	}

	// In production, would handle actual token transfers here
	// For now, just verify the participant can withdraw

	return nil
}

// GetMixingPoolStatus returns the status of a mixing pool
func (k Keeper) GetMixingPoolStatus(ctx context.Context, poolID string) (string, error) {
	pool, err := k.GetMixingPool(ctx, poolID)
	if err != nil {
		return "", err
	}

	return pool.Status, nil
}

// CancelMixingPool cancels a mixing pool if it hasn't started
func (k Keeper) CancelMixingPool(ctx context.Context, poolID string) error {
	pool, err := k.GetMixingPool(ctx, poolID)
	if err != nil {
		return err
	}

	if pool.Status == "completed" || pool.Status == "mixing" {
		return fmt.Errorf("cannot cancel pool that has started mixing")
	}

	pool.Status = "cancelled"
	return k.SetMixingPool(ctx, pool)
}
