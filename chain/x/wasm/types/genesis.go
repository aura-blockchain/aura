package types

import (
	"fmt"
)

// GenesisState represents the genesis state for the wasm module
type GenesisState struct {
	Params              Params              `json:"params"`
	AuthorizedUploaders []string            `json:"authorized_uploaders"`
	PausedContracts     []string            `json:"paused_contracts"`
	SecurityStats       SecurityStats       `json:"security_stats"`
}

// Params represents the wasm module parameters
type Params struct {
	// MaxContractSize is the maximum size of contract code in bytes (600KB default)
	MaxContractSize uint64 `json:"max_contract_size"`

	// MaxInstantiateGas is the maximum gas for contract instantiation
	MaxInstantiateGas uint64 `json:"max_instantiate_gas"`

	// MaxExecuteGas is the maximum gas for contract execution
	MaxExecuteGas uint64 `json:"max_execute_gas"`

	// MaxQueryGas is the maximum gas for contract queries
	MaxQueryGas uint64 `json:"max_query_gas"`

	// RequireAuthorization determines if contract uploads require authorization
	RequireAuthorization bool `json:"require_authorization"`

	// EnableMigration determines if contract migration is allowed
	EnableMigration bool `json:"enable_migration"`

	// MaxContractSizePerBlock is the maximum total size of contracts that can be uploaded per block
	MaxContractSizePerBlock uint64 `json:"max_contract_size_per_block"`
}

// SecurityStats tracks security-related statistics
type SecurityStats struct {
	TotalContractsUploaded   uint64 `json:"total_contracts_uploaded"`
	TotalContractsInstantiated uint64 `json:"total_contracts_instantiated"`
	TotalExecutions          uint64 `json:"total_executions"`
	TotalPausedContracts     uint64 `json:"total_paused_contracts"`
	ReentrancyAttemptsBlocked uint64 `json:"reentrancy_attempts_blocked"`
}

// DefaultParams returns default parameters for the wasm module
func DefaultParams() Params {
	return Params{
		MaxContractSize:         600 * 1024,        // 600KB
		MaxInstantiateGas:       2_000_000,         // 2M gas
		MaxExecuteGas:           1_000_000,         // 1M gas
		MaxQueryGas:             100_000,           // 100K gas
		RequireAuthorization:    true,              // Require authorization initially
		EnableMigration:         false,             // Disable migration initially
		MaxContractSizePerBlock: 5 * 1024 * 1024,   // 5MB per block
	}
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params Params,
	authorizedUploaders []string,
	pausedContracts []string,
	securityStats SecurityStats,
) *GenesisState {
	return &GenesisState{
		Params:              params,
		AuthorizedUploaders: authorizedUploaders,
		PausedContracts:     pausedContracts,
		SecurityStats:       securityStats,
	}
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:              DefaultParams(),
		AuthorizedUploaders: []string{},
		PausedContracts:     []string{},
		SecurityStats:       SecurityStats{},
	}
}

// Validate performs basic validation of genesis data
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate authorized uploaders
	for _, uploader := range gs.AuthorizedUploaders {
		if len(uploader) == 0 {
			return fmt.Errorf("empty authorized uploader address")
		}
	}

	// Validate paused contracts
	for _, contract := range gs.PausedContracts {
		if len(contract) == 0 {
			return fmt.Errorf("empty paused contract address")
		}
	}

	return nil
}

// Validate validates the module parameters
func (p Params) Validate() error {
	if p.MaxContractSize == 0 {
		return fmt.Errorf("max contract size must be positive")
	}

	if p.MaxContractSize > 10*1024*1024 { // 10MB max
		return fmt.Errorf("max contract size cannot exceed 10MB")
	}

	if p.MaxInstantiateGas == 0 {
		return fmt.Errorf("max instantiate gas must be positive")
	}

	if p.MaxExecuteGas == 0 {
		return fmt.Errorf("max execute gas must be positive")
	}

	if p.MaxQueryGas == 0 {
		return fmt.Errorf("max query gas must be positive")
	}

	if p.MaxContractSizePerBlock == 0 {
		return fmt.Errorf("max contract size per block must be positive")
	}

	return nil
}
