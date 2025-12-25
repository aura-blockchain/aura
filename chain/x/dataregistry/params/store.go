// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package params

import (
	"sync"

	"github.com/aequitas/aura/chain/x/dataregistry/types"
)

// Store manages parameters for the dataregistry module
type Store struct {
	mu     sync.RWMutex
	params types.Params
}

// NewStore creates a new parameter store
func NewStore(params types.Params) *Store {
	return &Store{
		params: params,
	}
}

// GetParams returns the current module parameters
func (s *Store) GetParams() types.Params {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params
}

// SetParams sets new module parameters
func (s *Store) SetParams(params types.Params) error {
	if err := types.ValidateParams(&params); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.params = params
	return nil
}
