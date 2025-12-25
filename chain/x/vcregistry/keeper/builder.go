// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/aequitas/aura/chain/x/vcregistry/params"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
)

// KeeperBuilder provides a type-safe builder pattern for constructing an immutable Keeper.
// This eliminates circular dependencies by ensuring all dependencies are set BEFORE
// the keeper is built, rather than using post-construction mutation.
//
// Usage:
//   keeper := NewKeeperBuilder(paramsStore, authority).
//       WithStore(storeKey, codec).
//       WithConfidenceScoreKeeper(csKeeper).
//       Build()
type KeeperBuilder struct {
	paramsStore *params.Store
	authority   string
	storeKey    storetypes.StoreKey
	codec       codec.BinaryCodec
	csKeeper    ConfidenceScoreKeeper
}

// NewKeeperBuilder initializes a new KeeperBuilder with required parameters.
func NewKeeperBuilder(paramsStore *params.Store, authority string) *KeeperBuilder {
	if paramsStore == nil {
		paramsStore = params.NewStore(*types.DefaultParams())
	}
	return &KeeperBuilder{
		paramsStore: paramsStore,
		authority:   authority,
	}
}

// WithStore sets the KV store key and codec for on-chain persistence.
// This is REQUIRED before calling Build().
func (b *KeeperBuilder) WithStore(storeKey storetypes.StoreKey, codec codec.BinaryCodec) *KeeperBuilder {
	b.storeKey = storeKey
	b.codec = codec
	return b
}

// WithConfidenceScoreKeeper sets the confidence score keeper dependency.
// This is REQUIRED before calling Build().
func (b *KeeperBuilder) WithConfidenceScoreKeeper(csKeeper ConfidenceScoreKeeper) *KeeperBuilder {
	b.csKeeper = csKeeper
	return b
}

// Build constructs and returns an immutable Keeper instance.
// Required: storeKey and codec must be set before calling this method.
// Optional: csKeeper (ConfidenceScoreKeeper) - if not set, a no-op implementation is used.
// Returns panic if required dependencies are missing (fail-fast at initialization).
func (b *KeeperBuilder) Build() *Keeper {
	if b.storeKey == nil {
		panic("vcregistry keeper builder: storeKey is required (call WithStore)")
	}
	if b.codec == nil {
		panic("vcregistry keeper builder: codec is required (call WithStore)")
	}

	// Use no-op confidence score keeper if not provided
	// This allows the node to start without circular dependency issues
	csKeeper := b.csKeeper
	if csKeeper == nil {
		csKeeper = &noOpConfidenceScoreKeeper{}
	}

	// Create the store
	store := NewStore(b.storeKey, b.codec)

	// Return immutable keeper with all dependencies set
	return &Keeper{
		store:       &store,
		paramsStore: b.paramsStore,
		csKeeper:    csKeeper,
		authority:   b.authority,
	}
}

// noOpConfidenceScoreKeeper provides safe default implementations
// when the confidence score module is not available or not yet wired.
// This allows the vcregistry to function independently for basic operations.
type noOpConfidenceScoreKeeper struct{}

func (n *noOpConfidenceScoreKeeper) GetUserScore(walletAddr string) (uint64, bool) {
	return 0, false
}

func (n *noOpConfidenceScoreKeeper) HasCompletedIR(walletAddr, irID string) bool {
	return false
}

func (n *noOpConfidenceScoreKeeper) GetArenaScore(walletAddr, arena string) (uint64, error) {
	return 0, nil
}

func (n *noOpConfidenceScoreKeeper) GetAnchorInfo(walletAddr string) (interface{}, bool) {
	return nil, false
}

func (n *noOpConfidenceScoreKeeper) IsVerified(walletAddr string) bool {
	return false
}

// Validate checks that all required dependencies have been set.
// Returns nil if valid, otherwise returns an error describing missing dependencies.
// Note: csKeeper is optional - a no-op implementation will be used if not provided.
func (b *KeeperBuilder) Validate() error {
	if b.storeKey == nil {
		return fmt.Errorf("%w: storeKey is required", types.ErrInvalidRequest)
	}
	if b.codec == nil {
		return fmt.Errorf("%w: codec is required", types.ErrInvalidRequest)
	}
	// csKeeper is optional - no validation needed
	return nil
}
