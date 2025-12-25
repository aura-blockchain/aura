// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"
)

// DefaultParams returns a default set of parameters
func DefaultParams() *Params {
	return &Params{
		MaxVcsPerUser:                   50,
		MaxMintPerDay:                   5,
		MaxMintPerHour:                  2,
		DefaultVcExpiryDays:             365,
		RevocationMerkleUpdateFrequency: 100,
		DidPrefix:                       "did:aura",
		DidNetwork:                      "mainnet",
		MintFee:                         "1000000uaura",
		RevokeFee:                       "0uaura",
		PolicyCreationDeposit:           "10000000uaura",
		RateLimitingEnabled:             true,
	}
}

// ValidateParams performs validation on the Params
func ValidateParams(p *Params) error {
	if p == nil {
		return fmt.Errorf("params cannot be nil")
	}

	if p.MaxVcsPerUser == 0 {
		return fmt.Errorf("max vcs per user must be positive")
	}

	if p.MaxMintPerDay == 0 {
		return fmt.Errorf("max mint per day must be positive")
	}

	if p.MaxMintPerHour == 0 {
		return fmt.Errorf("max mint per hour must be positive")
	}

	if p.MaxMintPerHour > p.MaxMintPerDay {
		return fmt.Errorf("max mint per hour cannot exceed max mint per day")
	}

	if p.DefaultVcExpiryDays == 0 {
		return fmt.Errorf("default vc expiry days must be positive")
	}

	if p.RevocationMerkleUpdateFrequency == 0 {
		return fmt.Errorf("revocation merkle update frequency must be positive")
	}

	if p.DidPrefix == "" {
		return fmt.Errorf("did prefix cannot be empty")
	}

	if p.DidNetwork == "" {
		return fmt.Errorf("did network cannot be empty")
	}

	if p.DidNetwork != "mainnet" && p.DidNetwork != "testnet" {
		return fmt.Errorf("did network must be either 'mainnet' or 'testnet', got %s", p.DidNetwork)
	}

	if p.MintFee == "" {
		return fmt.Errorf("mint fee cannot be empty")
	}

	if p.RevokeFee == "" {
		return fmt.Errorf("revoke fee cannot be empty")
	}

	if p.PolicyCreationDeposit == "" {
		return fmt.Errorf("policy creation deposit cannot be empty")
	}

	return nil
}
