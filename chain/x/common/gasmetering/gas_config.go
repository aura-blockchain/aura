// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package gasmetering

// GasConfig defines gas costs for common operations across all modules
type GasConfig struct {
	// Store operations
	StoreWriteCost      uint64
	StoreReadCost       uint64
	StoreDeleteCost     uint64
	StoreIterationCost  uint64
	StoreHasCost        uint64

	// Marshaling costs (per byte)
	MarshalCostPerByte   uint64
	UnmarshalCostPerByte uint64

	// Iteration limits
	MaxIterationResults uint32
	IterationBaseCost   uint64

	// Cryptographic operations
	HashCost                uint64
	SignatureVerifyCost     uint64
	PublicKeyDeriveCost     uint64
	ECDSAVerifyCost         uint64
	ED25519VerifyCost       uint64
	Secp256k1VerifyCost     uint64

	// Complex operations
	MerkleProofVerifyCost   uint64
	RingSignatureVerifyCost uint64
	ZKProofVerifyCost       uint64

	// Validation costs
	AddressValidationCost   uint64
	SignatureCheckCost      uint64
	AmountValidationCost    uint64
	StringValidationCost    uint64
}

// DefaultGasConfig returns the default gas configuration
func DefaultGasConfig() GasConfig {
	return GasConfig{
		// Store operations - based on Cosmos SDK defaults
		StoreWriteCost:      2000,
		StoreReadCost:       1000,
		StoreDeleteCost:     1000,
		StoreIterationCost:  30,
		StoreHasCost:        1000,

		// Marshaling - 1 gas per byte
		MarshalCostPerByte:   1,
		UnmarshalCostPerByte: 1,

		// Iteration
		MaxIterationResults: 1000,
		IterationBaseCost:   10000,

		// Crypto operations - based on computational complexity
		HashCost:                3000,
		SignatureVerifyCost:     6000,
		PublicKeyDeriveCost:     5000,
		ECDSAVerifyCost:         6000,
		ED25519VerifyCost:       5500,
		Secp256k1VerifyCost:     6000,

		// Complex operations
		MerkleProofVerifyCost:   10000,
		RingSignatureVerifyCost: 50000,
		ZKProofVerifyCost:       100000,

		// Validation
		AddressValidationCost:   1000,
		SignatureCheckCost:      5000,
		AmountValidationCost:    500,
		StringValidationCost:    100,
	}
}

// ModuleGasConfig allows modules to override default gas costs
type ModuleGasConfig struct {
	Base   GasConfig
	Module string

	// Module-specific overrides
	CustomOperations map[string]uint64
}

// NewModuleGasConfig creates a gas config for a specific module
func NewModuleGasConfig(module string) ModuleGasConfig {
	return ModuleGasConfig{
		Base:             DefaultGasConfig(),
		Module:           module,
		CustomOperations: make(map[string]uint64),
	}
}

// GetOperationCost returns the gas cost for a custom operation
func (m ModuleGasConfig) GetOperationCost(operation string) uint64 {
	if cost, ok := m.CustomOperations[operation]; ok {
		return cost
	}
	return 1000 // Default cost
}

// SetOperationCost sets the gas cost for a custom operation
func (m *ModuleGasConfig) SetOperationCost(operation string, cost uint64) {
	m.CustomOperations[operation] = cost
}
