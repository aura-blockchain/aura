package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	privacyproto "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

// JoinMixingRound allows a user to join a mixing round
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
		// Create new pool
		pool = &privacyproto.MixingPool{
			PoolId:          poolID,
			Participants:    [][]byte{},
			MinParticipants: params.MinMixingParticipants,
			MaxParticipants: params.MinMixingParticipants * 4, // Max is 4x min
			Status:          "open",
		}
	}

	// Check if pool is full
	currentCount := uint32(len(pool.Participants))
	if currentCount >= pool.MaxParticipants {
		return "", fmt.Errorf("mixing pool is full")
	}

	// Check if participant already joined
	for _, p := range pool.Participants {
		if string(p) == participant {
			return "", fmt.Errorf("participant already in pool")
		}
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

// shuffleParticipants performs deterministic shuffle using block hash as seed
func (k Keeper) shuffleParticipants(ctx context.Context, participants [][]byte) [][]byte {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	shuffled := make([][]byte, len(participants))
	for i := range participants {
		shuffled[i] = make([]byte, len(participants[i]))
		copy(shuffled[i], participants[i])
	}

	// Use block hash for deterministic randomness
	blockHash := sdkCtx.HeaderHash()

	// Fisher-Yates shuffle with deterministic randomness
	for i := len(shuffled) - 1; i > 0; i-- {
		// Generate deterministic random index using block hash
		hasher := sha256.New()
		hasher.Write(blockHash)
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

// WithdrawFromMixing allows withdrawal after mixing completes
func (k Keeper) WithdrawFromMixing(ctx context.Context, poolID, participantID string, output []byte) error {
	pool, err := k.GetMixingPool(ctx, poolID)
	if err != nil {
		return err
	}

	if pool.Status != "completed" {
		return fmt.Errorf("mixing not yet completed")
	}

	// Verify participant was in the pool
	found := false
	for _, p := range pool.Participants {
		if string(p) == participantID {
			found = true
			break
		}
	}

	if !found {
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
