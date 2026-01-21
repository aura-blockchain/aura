// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

// Params defines the parameters for the privacy module
type Params struct {
	EnableZkProofs                 bool   `json:"enable_zk_proofs"`
	EnableStealthAddresses         bool   `json:"enable_stealth_addresses"`
	EnableRingSignatures           bool   `json:"enable_ring_signatures"`
	EnableConfidentialTransactions bool   `json:"enable_confidential_transactions"`
	EnableNetworkPrivacy           bool   `json:"enable_network_privacy"`
	EnableMixing                   bool   `json:"enable_mixing"`
	MinRingSize                    uint32 `json:"min_ring_size"`
	MaxRingSize                    uint32 `json:"max_ring_size"`
	MinMixingParticipants          uint32 `json:"min_mixing_participants"`
	MixingFee                      string `json:"mixing_fee"`
	ZkProofVerificationCost        uint64 `json:"zk_proof_verification_cost"`
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		EnableZkProofs:                 true,
		EnableStealthAddresses:         true,
		EnableRingSignatures:           true,
		EnableConfidentialTransactions: true,
		EnableNetworkPrivacy:           true,
		EnableMixing:                   true,
		MinRingSize:                    3,
		MaxRingSize:                    16,
		MinMixingParticipants:          5,
		MixingFee:                      "1000",
		ZkProofVerificationCost:        10000,
	}
}

// ValidateParams validates the parameters
func ValidateParams(p Params) error {
	if p.MinRingSize < 2 {
		return ErrInvalidRingSize
	}

	if p.MaxRingSize < p.MinRingSize {
		return ErrInvalidRingSize
	}

	if p.MinMixingParticipants < 2 {
		return ErrInvalidMixingParams
	}

	return nil
}
