// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"
)

// DefaultParams returns default parameters for the wasm module.
// This creates a Params proto type with sensible defaults.
func DefaultParams() *Params {
	return &Params{
		CodeUploadAccess: AccessConfig{
			Permission: AccessTypeEverybody,
		},
		InstantiateDefaultPermission: AccessTypeEverybody,
		MaxWasmCodeSize:              600 * 1024,    // 600KB
		MaxGasWasmExecution:          10_000_000,    // 10M gas
		SecurityAnalysisEnabled:      true,
		RequireAdminForMigrate:       true,
	}
}

// DefaultGenesisState returns the default genesis state for the wasm module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:              *DefaultParams(),
		Codes:               []Code{},
		Contracts:           []Contract{},
		Sequences:           []Sequence{},
		AuthorizedUploaders: []string{},
		PausedContracts:     []string{},
		SecurityStats:       *DefaultSecurityStats(),
	}
}

// NewGenesisState creates a new GenesisState object.
func NewGenesisState(
	params Params,
	codes []Code,
	contracts []Contract,
	sequences []Sequence,
	authorizedUploaders []string,
	pausedContracts []string,
	securityStats SecurityStats,
) *GenesisState {
	return &GenesisState{
		Params:              params,
		Codes:               codes,
		Contracts:           contracts,
		Sequences:           sequences,
		AuthorizedUploaders: authorizedUploaders,
		PausedContracts:     pausedContracts,
		SecurityStats:       securityStats,
	}
}

// Validate performs basic validation of genesis data.
func ValidateGenesis(gs *GenesisState) error {
	if err := ValidateParams(&gs.Params); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate authorized uploaders
	for i, uploader := range gs.AuthorizedUploaders {
		if len(uploader) == 0 {
			return fmt.Errorf("empty authorized uploader address at index %d", i)
		}
	}

	// Validate paused contracts
	for i, contract := range gs.PausedContracts {
		if len(contract) == 0 {
			return fmt.Errorf("empty paused contract address at index %d", i)
		}
	}

	return nil
}

// ValidateParams validates the module parameters.
func ValidateParams(p *Params) error {
	if p.MaxWasmCodeSize == 0 {
		return fmt.Errorf("max wasm code size must be positive")
	}

	if p.MaxWasmCodeSize > 10*1024*1024 { // 10MB max
		return fmt.Errorf("max wasm code size cannot exceed 10MB")
	}

	if p.MaxGasWasmExecution == 0 {
		return fmt.Errorf("max gas wasm execution must be positive")
	}

	return nil
}

// GetCodeUploadAccess returns the code upload access config.
func GetCodeUploadAccess(p *Params) *AccessConfig {
	return &p.CodeUploadAccess
}

// DefaultSecurityStats returns default security statistics.
func DefaultSecurityStats() *SecurityStats {
	return &SecurityStats{
		TotalCodesAnalyzed:  0,
		CodesRejected:       0,
		ContractsPaused:     0,
		TotalExecutions:     0,
		FailedExecutions:    0,
		GasConsumedTotal:    0,
		LastSecurityScan:    0,
	}
}
