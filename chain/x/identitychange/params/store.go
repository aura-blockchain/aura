// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package params

import (
	"fmt"
	"sync"

	"github.com/aequitas/aura/chain/x/identitychange/types"
)

type Store struct {
	mu     sync.RWMutex
	params types.Params
}

func NewStore(defaults types.Params) *Store {
	if err := types.ValidateParams(&defaults); err != nil {
		panic(fmt.Sprintf("invalid default params: %v", err))
	}
	return &Store{params: defaults}
}

func (s *Store) GetParams() types.Params {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params
}

func (s *Store) SetParams(params types.Params) error {
	if err := types.ValidateParams(&params); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.params = params
	return nil
}
