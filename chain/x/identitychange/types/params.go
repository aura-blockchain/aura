package types

import (
	"fmt"

	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
)

type Params struct {
	MaxRequestsPerWalletPerMonth  int32  `json:"max_requests_per_wallet_per_month"`
	MinConfidenceAfterChange      int64  `json:"min_confidence_after_change"`
	StalenessHeightThreshold      int64  `json:"staleness_height_threshold"`
	AssistantSlashOnFalsePositive bool   `json:"assistant_slash_on_false_positive"`
	StalenessInvestigatorChain    string `json:"staleness_investigator_chain"`
}

func DefaultParams() Params {
	return Params{
		MaxRequestsPerWalletPerMonth:  2,
		MinConfidenceAfterChange:      1000,
		StalenessHeightThreshold:      10000,
		AssistantSlashOnFalsePositive: true,
		StalenessInvestigatorChain:    "",
	}
}

func DefaultParamsProto() *identitychangepb.Params {
	defaults := DefaultParams()
	return ParamsToProto(defaults)
}

func ParamsFromProto(pb *identitychangepb.Params) Params {
	if pb == nil {
		return Params{}
	}
	return Params{
		MaxRequestsPerWalletPerMonth:  pb.MaxRequestsPerWalletPerMonth,
		MinConfidenceAfterChange:      int64(pb.MinConfidenceAfterChange),
		StalenessHeightThreshold:      pb.StalenessHeightThreshold,
		AssistantSlashOnFalsePositive: pb.AssistantSlashOnFalsePositive,
		StalenessInvestigatorChain:    pb.StalenessInvestigatorChain,
	}
}

func ParamsToProto(p Params) *identitychangepb.Params {
	return &identitychangepb.Params{
		MaxRequestsPerWalletPerMonth:  p.MaxRequestsPerWalletPerMonth,
		MinConfidenceAfterChange:      int32(p.MinConfidenceAfterChange),
		StalenessHeightThreshold:      p.StalenessHeightThreshold,
		AssistantSlashOnFalsePositive: p.AssistantSlashOnFalsePositive,
		StalenessInvestigatorChain:    p.StalenessInvestigatorChain,
	}
}

func (p Params) Validate() error {
	if p.MaxRequestsPerWalletPerMonth <= 0 {
		return fmt.Errorf("max requests per wallet must be positive")
	}
	if p.MinConfidenceAfterChange <= 0 {
		return fmt.Errorf("min confidence after change must be positive")
	}
	if p.StalenessHeightThreshold <= 0 {
		return fmt.Errorf("staleness height threshold must be positive")
	}
	return nil
}
