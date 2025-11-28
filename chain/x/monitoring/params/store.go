package params

import (
	"sync"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// Store handles parameter storage for the monitoring module
type Store struct {
	params types.Params
	mu     sync.RWMutex
}

// NewStore creates a new parameter store
func NewStore(params types.Params) *Store {
	return &Store{
		params: params,
	}
}

// GetParams returns the current parameters
func (s *Store) GetParams() types.Params {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params
}

// SetParams updates the parameters
func (s *Store) SetParams(params types.Params) error {
	if err := types.ValidateParams(params); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.params = params
	return nil
}
