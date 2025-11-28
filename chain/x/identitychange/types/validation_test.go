package types

import (
	"testing"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	if params.MaxRequestsPerWalletPerMonth != 5 {
		t.Errorf("expected MaxRequestsPerWalletPerMonth to be 5, got %d", params.MaxRequestsPerWalletPerMonth)
	}

	if params.MinConfidenceAfterChange != 30 {
		t.Errorf("expected MinConfidenceAfterChange to be 30, got %d", params.MinConfidenceAfterChange)
	}

	if params.StalenessHeightThreshold != 100000 {
		t.Errorf("expected StalenessHeightThreshold to be 100000, got %d", params.StalenessHeightThreshold)
	}

	if !params.AssistantSlashOnFalsePositive {
		t.Error("expected AssistantSlashOnFalsePositive to be true")
	}

	if params.StalenessInvestigatorChain != "aura-mainnet" {
		t.Errorf("expected StalenessInvestigatorChain to be 'aura-mainnet', got %s", params.StalenessInvestigatorChain)
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

func TestValidateParams_NegativeMaxRequests(t *testing.T) {
	params := DefaultParams()
	params.MaxRequestsPerWalletPerMonth = 0

	err := ValidateParams(&params)
	if err == nil {
		t.Error("ValidateParams should return error for zero max requests")
	}
}

func TestValidateParams_NegativeConfidence(t *testing.T) {
	params := DefaultParams()
	params.MinConfidenceAfterChange = -1

	err := ValidateParams(&params)
	if err == nil {
		t.Error("ValidateParams should return error for negative confidence")
	}
}

func TestValidateParams_ConfidenceExceeds100(t *testing.T) {
	params := DefaultParams()
	params.MinConfidenceAfterChange = 101

	err := ValidateParams(&params)
	if err == nil {
		t.Error("ValidateParams should return error for confidence > 100")
	}
}

func TestValidateParams_ZeroStalenessThreshold(t *testing.T) {
	params := DefaultParams()
	params.StalenessHeightThreshold = 0

	err := ValidateParams(&params)
	if err == nil {
		t.Error("ValidateParams should return error for zero staleness threshold")
	}
}

func TestValidateParams_EmptyChain(t *testing.T) {
	params := DefaultParams()
	params.StalenessInvestigatorChain = ""

	err := ValidateParams(&params)
	if err == nil {
		t.Error("ValidateParams should return error for empty chain")
	}
}

func TestModuleName(t *testing.T) {
	if ModuleName != "identitychange" {
		t.Errorf("expected ModuleName to be 'identitychange', got %s", ModuleName)
	}
}

func TestEnumAliases(t *testing.T) {
	if IdentityChangeStatusPendingVerification != IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION {
		t.Error("IdentityChangeStatusPendingVerification alias mismatch")
	}

	if IdentityChangeStatusReadyToApply != IdentityChangeStatus_IDENTITY_CHANGE_STATUS_READY_TO_APPLY {
		t.Error("IdentityChangeStatusReadyToApply alias mismatch")
	}

	if IdentityChangeStatusRejected != IdentityChangeStatus_IDENTITY_CHANGE_STATUS_REJECTED {
		t.Error("IdentityChangeStatusRejected alias mismatch")
	}

	if IdentityChangeStatusApplied != IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPLIED {
		t.Error("IdentityChangeStatusApplied alias mismatch")
	}
}
