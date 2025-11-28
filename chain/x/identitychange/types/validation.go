package types

import (
	"fmt"
	"unsafe"

	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
)

// Enum constant aliases for backward compatibility
const (
	IdentityChangeStatusPendingVerification = IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION
	IdentityChangeStatusReadyToApply        = IdentityChangeStatus_IDENTITY_CHANGE_STATUS_READY_TO_APPLY
	IdentityChangeStatusRejected            = IdentityChangeStatus_IDENTITY_CHANGE_STATUS_REJECTED
	IdentityChangeStatusApplied             = IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPLIED
)

// DefaultParams returns default identity change parameters
func DefaultParams() Params {
	return Params{
		MaxRequestsPerWalletPerMonth:  5,
		MinConfidenceAfterChange:      30,
		StalenessHeightThreshold:      100000, // ~1 day at 1 block/sec
		AssistantSlashOnFalsePositive: true,
		StalenessInvestigatorChain:    "aura-mainnet",
	}
}

// DefaultParamsProto returns default params as proto type for compatibility
func DefaultParamsProto() *identitychangepb.Params {
	p := DefaultParams()
	return (*identitychangepb.Params)(unsafe.Pointer(&p))
}

// ParamsFromProto converts proto Params to Params (no-op since we use proto types directly)
func ParamsFromProto(p *identitychangepb.Params) *identitychangepb.Params {
	return p
}

// ValidateParams performs validation on the Params
func ValidateParams(p *Params) error {
	if p == nil {
		return fmt.Errorf("params cannot be nil")
	}

	if p.MaxRequestsPerWalletPerMonth <= 0 {
		return fmt.Errorf("max_requests_per_wallet_per_month must be positive")
	}

	if p.MinConfidenceAfterChange < 0 {
		return fmt.Errorf("min_confidence_after_change cannot be negative")
	}

	if p.MinConfidenceAfterChange > 100 {
		return fmt.Errorf("min_confidence_after_change cannot exceed 100")
	}

	if p.StalenessHeightThreshold <= 0 {
		return fmt.Errorf("staleness_height_threshold must be positive")
	}

	if p.StalenessInvestigatorChain == "" {
		return fmt.Errorf("staleness_investigator_chain cannot be empty")
	}

	return nil
}

