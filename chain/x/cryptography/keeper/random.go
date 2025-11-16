package keeper

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

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

	k.mu.Lock()
	defer k.mu.Unlock()

	// Generate source ID
	sourceID := fmt.Sprintf("rng_%s_%d", sourceType.String(), time.Now().Unix())

	// Hash the entropy pool for storage (never store raw entropy)
	h := sha256.New()
	h.Write(initialEntropy)
	entropyPoolHash := h.Sum(nil)

	now := time.Now()
	source := &cryptoproto.CryptoRandomSource{
		SourceId:        sourceID,
		SourceType:      sourceType,
		EntropyPoolHash: entropyPoolHash,
		EntropyBits:     int64(entropyBits),
		LastSeeded:      now,
		Status:          cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_HEALTHY,
	}

	// Store in state
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(source)
	store.Set(types.GetCryptoRandomSourceKey(sourceID), bz)

	// Cache
	k.randomSources[sourceID] = source

	k.Logger(ctx).Info("initialized random source",
		"source_id", sourceID,
		"type", sourceType.String(),
		"entropy_bits", entropyBits,
	)

	return sourceID, nil
}

// GetRandomSource retrieves a random source
func (k Keeper) GetRandomSource(ctx context.Context, sourceID string) (*cryptoproto.CryptoRandomSource, error) {
	k.mu.RLock()
	if source, ok := k.randomSources[sourceID]; ok {
		k.mu.RUnlock()
		return source, nil
	}
	k.mu.RUnlock()

	store := k.getStore(ctx)
	bz := store.Get(types.GetCryptoRandomSourceKey(sourceID))
	if bz == nil {
		return nil, types.ErrRandomSourceFailed
	}

	var source cryptoproto.CryptoRandomSource
	k.cdc.MustUnmarshal(bz, &source)

	k.mu.Lock()
	k.randomSources[sourceID] = &source
	k.mu.Unlock()

	return &source, nil
}

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

	k.mu.Lock()
	defer k.mu.Unlock()

	// Mix old and new entropy
	h := sha256.New()
	h.Write(source.EntropyPoolHash)
	h.Write(additionalEntropy)
	newEntropyPoolHash := h.Sum(nil)

	now := time.Now()
	source.EntropyPoolHash = newEntropyPoolHash
	source.EntropyBits += int64(entropyBits)
	source.LastSeeded = now
	source.Status = cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_HEALTHY

	// Store updated source
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(source)
	store.Set(types.GetCryptoRandomSourceKey(sourceID), bz)

	k.randomSources[sourceID] = source

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

	k.mu.Lock()
	defer k.mu.Unlock()

	source.Status = status

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(source)
	store.Set(types.GetCryptoRandomSourceKey(sourceID), bz)

	k.randomSources[sourceID] = source

	return nil
}

// CheckEntropyHealth checks the health of all random sources
func (k Keeper) CheckEntropyHealth(ctx context.Context) error {
	k.mu.RLock()
	defer k.mu.RUnlock()

	for sourceID, source := range k.randomSources {
		// Check if source needs reseeding (older than 24 hours)
		if time.Since(source.LastSeeded) > 24*time.Hour {
			k.Logger(ctx).Warn("random source needs reseeding",
				"source_id", sourceID,
				"last_seeded", source.LastSeeded,
			)
		}

		// Check entropy level
		params, _ := k.GetParams(ctx)
		if source.EntropyBits < int64(params.MinEntropyBits) {
			k.Logger(ctx).Warn("random source has low entropy",
				"source_id", sourceID,
				"entropy_bits", source.EntropyBits,
			)
		}
	}

	return nil
}

// GetEntropyStatistics returns statistics about random sources
func (k Keeper) GetEntropyStatistics(ctx context.Context) map[string]interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	stats := map[string]interface{}{
		"total_sources":       len(k.randomSources),
		"healthy_sources":     0,
		"failed_sources":      0,
		"low_entropy_sources": 0,
		"total_entropy_bits":  int64(0),
	}

	params, _ := k.GetParams(ctx)

	for _, source := range k.randomSources {
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
	}

	return stats
}

// SetRandomSource stores a random source (for genesis)
func (k *Keeper) SetRandomSource(ctx context.Context, source *cryptoproto.CryptoRandomSource) error {
	if source == nil {
		return nil
	}
	k.mu.Lock()
	defer k.mu.Unlock()
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(source)
	store.Set(types.GetRandomSourceKey(source.SourceId), bz)
	k.randomSources[source.SourceId] = source
	return nil
}
