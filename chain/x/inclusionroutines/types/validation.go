// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"
)

// DefaultParams returns default inclusion routines parameters
func DefaultParams() Params {
	return Params{
		MaxIrPerLocale:       100,
		DefaultRateLimitHour: 5,
		SuspensionFee:        "1000000uaura",
		MinGovernanceDeposit: "10000000uaura",
	}
}

// ValidateParams performs validation on the Params (package-level function)
// Works with both *inclusionroutinespb.Params and *Params (local type)
func ValidateParams(p *Params) error {
	if p == nil {
		return fmt.Errorf("params cannot be nil")
	}

	if p.MaxIrPerLocale <= 0 {
		return fmt.Errorf("max_ir_per_locale must be positive")
	}

	if p.DefaultRateLimitHour < 0 {
		return fmt.Errorf("default_rate_limit_hour cannot be negative")
	}

	if p.SuspensionFee == "" {
		return fmt.Errorf("suspension_fee cannot be empty")
	}

	if p.MinGovernanceDeposit == "" {
		return fmt.Errorf("min_governance_deposit cannot be empty")
	}

	return nil
}
