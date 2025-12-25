// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/aequitas/aura/chain/x/confidencescore/params"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

// KeeperBuilder provides a type-safe builder pattern for constructing an immutable Keeper.
// This eliminates circular dependencies by ensuring all dependencies are set BEFORE
// the keeper is built, rather than using post-construction mutation.
//
// Usage:
//
//	keeper := NewKeeperBuilder(storeService, cdc, paramsStore, authority, logger).
//	    WithIRRegistry(irKeeper).
//	    Build()
type KeeperBuilder struct {
	storeService store.KVStoreService
	cdc          codec.BinaryCodec
	paramsStore  *params.Store
	authority    string
	logger       log.Logger
	irRegistry   IRRegistry
}

// NewKeeperBuilder initializes a new KeeperBuilder with required parameters.
func NewKeeperBuilder(
	storeService store.KVStoreService,
	cdc codec.BinaryCodec,
	paramsStore *params.Store,
	authority string,
	logger log.Logger,
) *KeeperBuilder {
	if paramsStore == nil {
		paramsStore = params.NewStore(types.DefaultParams())
	}
	return &KeeperBuilder{
		storeService: storeService,
		cdc:          cdc,
		paramsStore:  paramsStore,
		authority:    authority,
		logger:       logger,
	}
}

// WithIRRegistry sets the inclusion routines registry dependency.
// This is OPTIONAL - can be set later via SetIRRegistry on the keeper.
func (b *KeeperBuilder) WithIRRegistry(irRegistry IRRegistry) *KeeperBuilder {
	b.irRegistry = irRegistry
	return b
}

// Build constructs and returns an immutable Keeper instance.
// All required dependencies must be set before calling this method.
func (b *KeeperBuilder) Build() *Keeper {
	// Create keeper with all dependencies set
	keeper := &Keeper{
		storeService: b.storeService,
		cdc:          b.cdc,
		paramsStore:  b.paramsStore,
		authority:    b.authority,
		logger:       b.logger,
		irRegistry:   b.irRegistry,
	}

	return keeper
}

// Validate checks that all required dependencies have been set.
// Returns nil if valid, otherwise returns an error describing missing dependencies.
func (b *KeeperBuilder) Validate() error {
	if b.storeService == nil {
		return fmt.Errorf("store service is required: %w", types.ErrInvalidRequest)
	}
	if b.cdc == nil {
		return fmt.Errorf("codec is required: %w", types.ErrInvalidRequest)
	}
	if b.logger == nil {
		return fmt.Errorf("logger is required: %w", types.ErrInvalidRequest)
	}
	// irRegistry is optional - can be set later
	return nil
}
