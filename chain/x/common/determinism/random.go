package determinism

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DeterministicRNG provides deterministic random number generation
// based on block data. This ensures all validators generate the same
// "random" numbers.
type DeterministicRNG struct {
	seed []byte
}

// NewDeterministicRNG creates a new deterministic RNG seeded with block data
func NewDeterministicRNG(ctx context.Context, additionalEntropy ...[]byte) *DeterministicRNG {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Combine multiple sources of entropy
	h := sha256.New()

	// Block height (always different per block)
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, uint64(sdkCtx.BlockHeight()))
	h.Write(heightBytes)

	// Block hash (deterministic per block)
	h.Write(sdkCtx.HeaderHash())

	// Chain ID (ensures different chains have different randomness)
	h.Write([]byte(sdkCtx.ChainID()))

	// Block time (deterministic per block)
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(sdkCtx.BlockTime().Unix()))
	h.Write(timeBytes)

	// Additional entropy (e.g., transaction hash, operation ID)
	for _, entropy := range additionalEntropy {
		h.Write(entropy)
	}

	seed := h.Sum(nil)

	return &DeterministicRNG{
		seed: seed,
	}
}

// Bytes generates n deterministic random bytes
func (rng *DeterministicRNG) Bytes(n int) []byte {
	result := make([]byte, n)
	h := sha256.New()
	h.Write(rng.seed)

	counter := uint64(0)
	offset := 0

	for offset < n {
		// Generate next block of randomness
		counterBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(counterBytes, counter)

		h.Reset()
		h.Write(rng.seed)
		h.Write(counterBytes)
		block := h.Sum(nil)

		// Copy to result
		toCopy := n - offset
		if toCopy > len(block) {
			toCopy = len(block)
		}
		copy(result[offset:], block[:toCopy])

		offset += toCopy
		counter++
	}

	return result
}

// Uint64 generates a deterministic random uint64
func (rng *DeterministicRNG) Uint64() uint64 {
	bytes := rng.Bytes(8)
	return binary.BigEndian.Uint64(bytes)
}

// Int63 generates a deterministic random int63 (positive int64)
func (rng *DeterministicRNG) Int63() int64 {
	return int64(rng.Uint64() & 0x7FFFFFFFFFFFFFFF)
}

// Intn generates a deterministic random integer in [0, n)
func (rng *DeterministicRNG) Intn(n int) int {
	if n <= 0 {
		panic("invalid argument to Intn")
	}
	return int(rng.Int63n(int64(n)))
}

// Int63n generates a deterministic random int64 in [0, n)
func (rng *DeterministicRNG) Int63n(n int64) int64 {
	if n <= 0 {
		panic("invalid argument to Int63n")
	}

	// Use rejection sampling to avoid bias
	max := int64((1 << 63) - 1 - (1<<63)%uint64(n))
	v := rng.Int63()

	for v > max {
		v = rng.Int63()
	}

	return v % n
}

// IntInRange generates a deterministic random integer in [min, max]
func (rng *DeterministicRNG) IntInRange(min, max int64) int64 {
	if min >= max {
		panic("min must be less than max")
	}
	rangeSize := max - min + 1
	return min + rng.Int63n(rangeSize)
}

// Float64 generates a deterministic random float64 in [0.0, 1.0)
// Note: Use with caution - prefer integer math when possible
func (rng *DeterministicRNG) Float64() float64 {
	return float64(rng.Int63n(1<<53)) / (1 << 53)
}

// Bool generates a deterministic random boolean
func (rng *DeterministicRNG) Bool() bool {
	return rng.Uint64()&1 == 1
}

// Shuffle performs a deterministic Fisher-Yates shuffle
// This is safe for use in state machines
func (rng *DeterministicRNG) Shuffle(n int, swap func(i, j int)) {
	for i := n - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		swap(i, j)
	}
}

// Select deterministically selects k unique indices from [0, n)
func (rng *DeterministicRNG) Select(n, k int) []int {
	if k > n {
		panic("k cannot be greater than n")
	}

	// Use Floyd's algorithm for unbiased selection
	selected := make(map[int]bool)
	result := make([]int, 0, k)

	for i := n - k; i < n; i++ {
		j := rng.Intn(i + 1)
		if selected[j] {
			result = append(result, i)
			selected[i] = true
		} else {
			result = append(result, j)
			selected[j] = true
		}
	}

	return result
}

// BigInt generates a deterministic random big.Int in [0, max)
func (rng *DeterministicRNG) BigInt(max *big.Int) *big.Int {
	if max.Sign() <= 0 {
		panic("max must be positive")
	}

	// Calculate number of bytes needed
	byteLen := (max.BitLen() + 7) / 8
	bytes := rng.Bytes(byteLen)

	result := new(big.Int).SetBytes(bytes)

	// Use rejection sampling to avoid bias
	result.Mod(result, max)

	return result
}

// DeterministicShuffleSlice shuffles a slice deterministically
func DeterministicShuffleSlice[T any](ctx context.Context, slice []T, entropy ...[]byte) {
	rng := NewDeterministicRNG(ctx, entropy...)
	rng.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

// DeterministicSelectRandom selects a random element from a slice
func DeterministicSelectRandom[T any](ctx context.Context, slice []T, entropy ...[]byte) (T, error) {
	if len(slice) == 0 {
		var zero T
		return zero, fmt.Errorf("cannot select from empty slice")
	}

	rng := NewDeterministicRNG(ctx, entropy...)
	index := rng.Intn(len(slice))
	return slice[index], nil
}

// DeterministicSelectMultiple selects k random elements from a slice
func DeterministicSelectMultiple[T any](ctx context.Context, slice []T, k int, entropy ...[]byte) ([]T, error) {
	if k > len(slice) {
		return nil, fmt.Errorf("cannot select %d elements from slice of length %d", k, len(slice))
	}

	rng := NewDeterministicRNG(ctx, entropy...)
	indices := rng.Select(len(slice), k)

	result := make([]T, k)
	for i, idx := range indices {
		result[i] = slice[idx]
	}

	return result, nil
}

// GenerateDeterministicID generates a deterministic ID based on context and data
func GenerateDeterministicID(ctx context.Context, prefix string, data ...[]byte) string {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	h := sha256.New()
	h.Write([]byte(prefix))

	// Add block height
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, uint64(sdkCtx.BlockHeight()))
	h.Write(heightBytes)

	// Add all data
	for _, d := range data {
		h.Write(d)
	}

	hash := h.Sum(nil)
	return fmt.Sprintf("%s-%x", prefix, hash[:16])
}

// DeterministicWeightedSelect selects an element based on weights
func DeterministicWeightedSelect[T any](ctx context.Context, items []T, weights []uint64, entropy ...[]byte) (T, error) {
	if len(items) != len(weights) {
		var zero T
		return zero, fmt.Errorf("items and weights must have same length")
	}

	if len(items) == 0 {
		var zero T
		return zero, fmt.Errorf("cannot select from empty slice")
	}

	// Calculate total weight
	totalWeight := uint64(0)
	for _, w := range weights {
		totalWeight += w
	}

	if totalWeight == 0 {
		var zero T
		return zero, fmt.Errorf("total weight is zero")
	}

	// Generate random value in [0, totalWeight)
	rng := NewDeterministicRNG(ctx, entropy...)
	randValue := rng.Uint64() % totalWeight

	// Select item based on weight
	currentWeight := uint64(0)
	for i, w := range weights {
		currentWeight += w
		if randValue < currentWeight {
			return items[i], nil
		}
	}

	// Should never reach here
	return items[len(items)-1], nil
}
