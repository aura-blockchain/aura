// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	if params.MaxIrPerLocale != 100 {
		t.Errorf("expected MaxIrPerLocale to be 100, got %d", params.MaxIrPerLocale)
	}

	if params.DefaultRateLimitHour != 5 {
		t.Errorf("expected DefaultRateLimitHour to be 5, got %d", params.DefaultRateLimitHour)
	}

	if params.SuspensionFee != "1000000uaura" {
		t.Errorf("expected SuspensionFee to be 1000000uaura, got %s", params.SuspensionFee)
	}

	if params.MinGovernanceDeposit != "10000000uaura" {
		t.Errorf("expected MinGovernanceDeposit to be 10000000uaura, got %s", params.MinGovernanceDeposit)
	}
}

func TestValidateParams_Valid(t *testing.T) {
	params := DefaultParams()

	err := ValidateParams(&params)
	if err != nil {
		t.Errorf("ValidateParams should not return error for default params: %v", err)
	}
}

func TestValidateParams_Nil(t *testing.T) {
	err := ValidateParams(nil)
	if err == nil {
		t.Error("ValidateParams should return error for nil params")
	}
}

func TestValidateParams_ZeroMaxIrPerLocale(t *testing.T) {
	params := DefaultParams()
	params.MaxIrPerLocale = 0

	err := ValidateParams(&params)
	if err == nil {
		t.Error("ValidateParams should return error for zero max_ir_per_locale")
	}
}

func TestValidateParams_NegativeRateLimit(t *testing.T) {
	params := DefaultParams()
	params.DefaultRateLimitHour = -1

	err := ValidateParams(&params)
	if err == nil {
		t.Error("ValidateParams should return error for negative rate limit")
	}
}

func TestValidateParams_EmptySuspensionFee(t *testing.T) {
	params := DefaultParams()
	params.SuspensionFee = ""

	err := ValidateParams(&params)
	if err == nil {
		t.Error("ValidateParams should return error for empty suspension_fee")
	}
}

func TestValidateParams_EmptyGovernanceDeposit(t *testing.T) {
	params := DefaultParams()
	params.MinGovernanceDeposit = ""

	err := ValidateParams(&params)
	if err == nil {
		t.Error("ValidateParams should return error for empty min_governance_deposit")
	}
}
