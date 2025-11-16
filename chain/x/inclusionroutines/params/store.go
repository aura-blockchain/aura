package params

import (
	"fmt"
	"sync"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
)

// Store manages the module parameters
type Store struct {
	mu     sync.RWMutex
	params types.Params
}

// NewStore creates a new params store with default parameters
func NewStore(defaults types.Params) *Store {
	if err := defaults.Validate(); err != nil {
		panic(fmt.Sprintf("invalid default params: %v", err))
	}
	return &Store{params: defaults}
}

// GetParams returns the current parameters
func (s *Store) GetParams() types.Params {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params
}

// SetParams sets new parameters after validation
func (s *Store) SetParams(params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.params = params
	return nil
}
