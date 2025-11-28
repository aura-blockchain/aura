package keeper

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// InitializeRandomSource initializes a cryptographically secure random number source
func (k Keeper) InitializeRandomSource(
	ctx context.Context,
	sourceType cryptoproto.RandomSourceType,
	initialEntropy []byte,
) (string, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return "", err
	}

	// Validate entropy
	entropyBits := len(initialEntropy) * 8
	if entropyBits < int(params.MinEntropyBits) {
		return "", types.ErrInsufficientEntropy
	}

	// Generate source ID using consensus-safe block time
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()
	sourceID := fmt.Sprintf("rng_%s_%d", sourceType.String(), blockTime.Unix())

	// Hash the entropy pool for storage (never store raw entropy)
	h := sha256.New()
	h.Write(initialEntropy)
	entropyPoolHash := h.Sum(nil)

	now := blockTime
	source := &cryptoproto.CryptoRandomSource{
		SourceId:        sourceID,
		SourceType:      sourceType,
		EntropyPoolHash: entropyPoolHash,
		EntropyBits:     int64(entropyBits),
		LastSeeded:      timestamppb.New(now),
		Status:          cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_HEALTHY,
	}

	// Store in KV store
	if err := k.SetRandomSource(ctx, source); err != nil {
		return "", err
	}

	k.Logger(ctx).Info("initialized random source",
		"source_id", sourceID,
		"type", sourceType.String(),
		"entropy_bits", entropyBits,
	)

	return sourceID, nil
}

// Note: GetRandomSource is now implemented in keeper.go using KV store

// GenerateRandomBytesFromSource generates random bytes from a specific source
func (k Keeper) GenerateRandomBytesFromSource(
	ctx context.Context,
	sourceID string,
	length int,
) ([]byte, error) {
	source, err := k.GetRandomSource(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	if source.Status != cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_HEALTHY {
		return nil, types.ErrRandomSourceFailed
	}

	// Generate random bytes using the system's crypto/rand
	randomBytes := make([]byte, length)
	_, err = rand.Read(randomBytes)
	if err != nil {
		k.updateRandomSourceStatus(ctx, sourceID, cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_FAILED)
		return nil, types.ErrRandomSourceFailed
	}

	// Mix with source entropy (in a real implementation)
	// For now, just use the system random

	return randomBytes, nil
}

// GenerateRandomUint64 generates a cryptographically secure random uint64
func (k Keeper) GenerateRandomUint64(ctx context.Context) (uint64, error) {
	randomBytes, err := k.GenerateSecureRandomBytes(8)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint64(randomBytes), nil
}

// GenerateRandomInRange generates a random number in the range [min, max]
func (k Keeper) GenerateRandomInRange(ctx context.Context, min, max int64) (int64, error) {
	if min >= max {
		return 0, fmt.Errorf("invalid range")
	}

	rangeSize := max - min + 1
	randomValue, err := k.GenerateSecureRandomInt(rangeSize)
	if err != nil {
		return 0, err
	}

	return min + randomValue, nil
}

// ReseedRandomSource reseeds a random source with fresh entropy
func (k Keeper) ReseedRandomSource(
	ctx context.Context,
	sourceID string,
	additionalEntropy []byte,
) error {
	source, err := k.GetRandomSource(ctx, sourceID)
	if err != nil {
		return err
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	// Validate entropy
	entropyBits := len(additionalEntropy) * 8
	if entropyBits < int(params.MinEntropyBits) {
		return types.ErrInsufficientEntropy
	}

	// Mix old and new entropy
	h := sha256.New()
	h.Write(source.EntropyPoolHash)
	h.Write(additionalEntropy)
	newEntropyPoolHash := h.Sum(nil)

	reseedCtx := sdk.UnwrapSDKContext(ctx)
	now := reseedCtx.BlockTime()
	source.EntropyPoolHash = newEntropyPoolHash
	source.EntropyBits += int64(entropyBits)
	source.LastSeeded = timestamppb.New(now)
	source.Status = cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_HEALTHY

	// Store updated source
	if err := k.SetRandomSource(ctx, source); err != nil {
		return err
	}

	k.Logger(ctx).Info("reseeded random source",
		"source_id", sourceID,
		"additional_entropy_bits", entropyBits,
	)

	return nil
}

// updateRandomSourceStatus updates the status of a random source
func (k Keeper) updateRandomSourceStatus(
	ctx context.Context,
	sourceID string,
	status cryptoproto.RandomSourceStatus,
) error {
	source, err := k.GetRandomSource(ctx, sourceID)
	if err != nil {
		return err
	}

	source.Status = status

	return k.SetRandomSource(ctx, source)
}

// CheckEntropyHealth checks the health of all random sources
func (k Keeper) CheckEntropyHealth(ctx context.Context) error {
	params, _ := k.GetParams(ctx)

	return k.IterateRandomSources(ctx, func(source *cryptoproto.CryptoRandomSource) bool {
		// Check if source needs reseeding (older than 24 hours)
		if time.Since(source.LastSeeded.AsTime()) > 24*time.Hour {
			k.Logger(ctx).Warn("random source needs reseeding",
				"source_id", source.SourceId,
				"last_seeded", source.LastSeeded,
			)
		}

		// Check entropy level
		if source.EntropyBits < int64(params.MinEntropyBits) {
			k.Logger(ctx).Warn("random source has low entropy",
				"source_id", source.SourceId,
				"entropy_bits", source.EntropyBits,
			)
		}
		return false
	})
}

// GetEntropyStatistics returns statistics about random sources
func (k Keeper) GetEntropyStatistics(ctx context.Context) map[string]interface{} {
	stats := map[string]interface{}{
		"total_sources":       0,
		"healthy_sources":     0,
		"failed_sources":      0,
		"low_entropy_sources": 0,
		"total_entropy_bits":  int64(0),
	}

	params, _ := k.GetParams(ctx)

	_ = k.IterateRandomSources(ctx, func(source *cryptoproto.CryptoRandomSource) bool {
		stats["total_sources"] = stats["total_sources"].(int) + 1
		stats["total_entropy_bits"] = stats["total_entropy_bits"].(int64) + source.EntropyBits

		switch source.Status {
		case cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_HEALTHY:
			stats["healthy_sources"] = stats["healthy_sources"].(int) + 1
		case cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_FAILED:
			stats["failed_sources"] = stats["failed_sources"].(int) + 1
		case cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_LOW_ENTROPY:
			stats["low_entropy_sources"] = stats["low_entropy_sources"].(int) + 1
		}

		if source.EntropyBits < int64(params.MinEntropyBits) {
			stats["low_entropy_sources"] = stats["low_entropy_sources"].(int) + 1
		}
		return false
	})

	return stats
}

// Note: SetRandomSource is now implemented in keeper.go using KV store
