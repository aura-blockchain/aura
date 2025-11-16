package types

import (
	"fmt"

	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

// Params defines the parameters for the inclusionroutines module
type Params struct {
	MaxIRPerLocale       int32  `json:"max_ir_per_locale"`
	DefaultRateLimitHour int32  `json:"default_rate_limit_hour"`
	SuspensionFee        string `json:"suspension_fee"`
	MinGovernanceDeposit string `json:"min_governance_deposit"`
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		MaxIRPerLocale:       50,
		DefaultRateLimitHour: 10,
		SuspensionFee:        "1000000uaura",  // 1 AURA
		MinGovernanceDeposit: "10000000uaura", // 10 AURA
	}
}

// DefaultParamsProto returns a default set of parameters in proto format
func DefaultParamsProto() *inclusionroutinespb.Params {
	defaults := DefaultParams()
	return ParamsToProto(defaults)
}

// ParamsFromProto converts proto Params to internal Params type
func ParamsFromProto(pb *inclusionroutinespb.Params) Params {
	if pb == nil {
		return Params{}
	}
	return Params{
		MaxIRPerLocale:       pb.MaxIrPerLocale,
		DefaultRateLimitHour: pb.DefaultRateLimitHour,
		SuspensionFee:        pb.SuspensionFee,
		MinGovernanceDeposit: pb.MinGovernanceDeposit,
	}
}

// ParamsToProto converts internal Params to proto Params type
func ParamsToProto(p Params) *inclusionroutinespb.Params {
	return &inclusionroutinespb.Params{
		MaxIrPerLocale:       p.MaxIRPerLocale,
		DefaultRateLimitHour: p.DefaultRateLimitHour,
		SuspensionFee:        p.SuspensionFee,
		MinGovernanceDeposit: p.MinGovernanceDeposit,
	}
}

// Validate performs validation on the Params
func (p Params) Validate() error {
	if p.MaxIRPerLocale <= 0 {
		return fmt.Errorf("max ir per locale must be positive, got %d", p.MaxIRPerLocale)
	}

	if p.DefaultRateLimitHour < 0 {
		return fmt.Errorf("default rate limit hour must be non-negative, got %d", p.DefaultRateLimitHour)
	}

	if p.SuspensionFee == "" {
		return fmt.Errorf("suspension fee cannot be empty")
	}

	if p.MinGovernanceDeposit == "" {
		return fmt.Errorf("min governance deposit cannot be empty")
	}

	return nil
}
