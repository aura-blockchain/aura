// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

// RegistryStats contains statistics about the VC registry
type RegistryStats struct {
	TotalVCs      uint64
	ActiveVCs     uint64
	RevokedVCs    uint64
	ExpiredVCs    uint64
	TotalDIDs     uint64
	TotalPolicies uint64
	VCsByType     map[VCType]uint64
}
