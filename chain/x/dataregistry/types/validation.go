// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"
)

// RegistryStats represents statistics about the data registry
type RegistryStats struct {
	TotalDataItems     uint64
	TotalVerifiedItems uint64
	TotalVerifications uint64
	TotalStorageBytes  uint64
	ItemsByType        map[string]uint64
}

// DefaultParams returns default dataregistry parameters
func DefaultParams() Params {
	return Params{
		MaxDataItemsPerUser: 1000,
		MaxStorageBytes:     10485760, // 10 MB
		StorageFee:          "100",
		VerificationReward:  10,
		AuthorizedVerifiers: []string{},
	}
}

// ValidateParams performs validation on the Params
func ValidateParams(p *Params) error {
	if p == nil {
		return fmt.Errorf("params cannot be nil")
	}

	if p.MaxDataItemsPerUser == 0 {
		return fmt.Errorf("max_data_items_per_user must be greater than 0")
	}

	if p.MaxStorageBytes == 0 {
		return fmt.Errorf("max_storage_bytes must be greater than 0")
	}

	if p.StorageFee == "" {
		return fmt.Errorf("storage_fee cannot be empty")
	}

	// verification_reward can be 0 (no reward)

	// authorized_verifiers can be empty (anyone can verify)

	return nil
}
