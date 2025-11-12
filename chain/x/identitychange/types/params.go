package types

import "fmt"

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
