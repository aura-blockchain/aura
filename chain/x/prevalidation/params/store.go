package params

import (
	"sync"

	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

// Store manages the parameters for the prevalidation module
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

// GetParams returns the current parameters
func (s *Store) GetParams() types.Params {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params
}

// SetParams updates the parameters
func (s *Store) SetParams(params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.params = params
	return nil
}

// UpdateSchedulerConfig updates just the scheduler configuration
func (s *Store) UpdateSchedulerConfig(config *types.SchedulerConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.params.SchedulerConfig = config
	return nil
}

// UpdateAutoScalingConfig updates just the auto-scaling configuration
func (s *Store) UpdateAutoScalingConfig(config *types.AutoScalingConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.params.AutoScalingConfig = config
	return nil
}

// IsEnabled checks if pre-validation is enabled
func (s *Store) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params.Enabled
}

// GetSchedulerConfig returns the scheduler configuration
func (s *Store) GetSchedulerConfig() *types.SchedulerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params.SchedulerConfig
}

// GetAutoScalingConfig returns the auto-scaling configuration
func (s *Store) GetAutoScalingConfig() *types.AutoScalingConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params.AutoScalingConfig
}
