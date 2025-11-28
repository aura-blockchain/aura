package params

import (
	"sync"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// Store manages parameter storage for the economic security module
type Store struct {
	mu     sync.RWMutex
	params types.Params
}

// NewStore creates a new parameter store
func NewStore(initialParams types.Params) *Store {
	return &Store{
		params: initialParams,
	}
}

// GetParams returns the current parameters
func (s *Store) GetParams() types.Params {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params
}

// SetParams sets new parameters
func (s *Store) SetParams(params types.Params) error {
	if err := types.ValidateParams(&params); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.params = params
	return nil
}

// GetTokenomicsConfig returns tokenomics configuration
func (s *Store) GetTokenomicsConfig() *types.TokenomicsConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params.Tokenomics
}

// GetWhaleProtection returns whale protection configuration
func (s *Store) GetWhaleProtection() *types.WhaleProtection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params.WhaleProtection
}

// GetTransferTax returns transfer tax configuration
func (s *Store) GetTransferTax() *types.TransferTaxConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params.TransferTax
}

// GetLiquidityMining returns liquidity mining configuration
func (s *Store) GetLiquidityMining() *types.LiquidityMiningConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params.LiquidityMining
}

// GetGovernance returns governance configuration
func (s *Store) GetGovernance() *types.GovernanceConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params.Governance
}

// GetTreasuryMultisig returns treasury multisig configuration
func (s *Store) GetTreasuryMultisig() *types.TreasuryMultisig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params.TreasuryMultisig
}

// GetDynamicFees returns dynamic fee configuration
func (s *Store) GetDynamicFees() *types.DynamicFeeConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params.DynamicFees
}

// GetMEV returns MEV configuration
func (s *Store) GetMEV() *types.MEVConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params.Mev
}
