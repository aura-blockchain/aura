package types

import (
	"fmt"

	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

// Params defines the parameters for the confidencescore module
type Params struct {
	// Verification thresholds
	VerificationThreshold  uint64 `json:"verification_threshold"`
	HighAssuranceThreshold uint64 `json:"high_assurance_threshold"`
	ArenaFocusThreshold    uint64 `json:"arena_focus_threshold"`

	// Velocity bonuses (time-based completion bonuses)
	VelocityBonusDays        []uint64  `json:"velocity_bonus_days"`
	VelocityBonusMultipliers []float32 `json:"velocity_bonus_multipliers"`

	// Arena multipliers
	ArenaMultipliers map[uint64]float32 `json:"arena_multipliers"`

	// Slashing parameters
	SlashPercentage uint64 `json:"slash_percentage"`
	AppealDeposit   string `json:"appeal_deposit"`

	// Rate limiting
	MaxIRsPerDay  uint64 `json:"max_irs_per_day"`
	MaxIRsPerHour uint64 `json:"max_irs_per_hour"`

	// Jackpot probabilities
	JackpotOdds        []uint64  `json:"jackpot_odds"`
	JackpotMultipliers []float32 `json:"jackpot_multipliers"`

	// Staleness (future)
	StalenessEnabled       bool   `json:"staleness_enabled"`
	DegradationRatePerYear uint64 `json:"degradation_rate_per_year"`

	// PoI (Proof-of-Identity) Rewards (Whitepaper Section 12.0)
	PoIRewardsEnabled      bool   `json:"poi_rewards_enabled"`
	UserRewardSplitPercent uint64 `json:"user_reward_split_percent"` // % of reward to user (rest to node operator)
	VelocityBonusEnabled   bool   `json:"velocity_bonus_enabled"`    // Enable VBT (Velocity Bonus Tier)
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		VerificationThreshold:  10000,
		HighAssuranceThreshold: 15000,
		ArenaFocusThreshold:    5000,

		VelocityBonusDays:        []uint64{7, 30},
		VelocityBonusMultipliers: []float32{1.25, 1.10},

		ArenaMultipliers: map[uint64]float32{
			3000: 1.1,
			4000: 1.2,
			5000: 1.5,
		},

		SlashPercentage: 25,                // 25% default slash
		AppealDeposit:   "1000000000uaura", // 1000 AURA

		MaxIRsPerDay:  10,
		MaxIRsPerHour: 3,

		JackpotOdds:        []uint64{100, 1000},
		JackpotMultipliers: []float32{5.0, 25.0},

		StalenessEnabled:       false,
		DegradationRatePerYear: 0,

		// PoI Rewards (enabled by default)
		PoIRewardsEnabled:      true,
		UserRewardSplitPercent: 50, // 50% to user, 50% to node operator
		VelocityBonusEnabled:   true,
	}
}

// DefaultParamsProto returns a default set of parameters in proto format
func DefaultParamsProto() *confidencescorepb.Params {
	defaults := DefaultParams()
	return ParamsToProto(defaults)
}

// ParamsFromProto converts proto Params to internal Params type
func ParamsFromProto(pb *confidencescorepb.Params) Params {
	if pb == nil {
		return Params{}
	}

	arenaMultipliers := make(map[uint64]float32)
	for threshold, multiplier := range pb.ArenaMultipliers {
		arenaMultipliers[threshold] = multiplier
	}

	return Params{
		VerificationThreshold:    pb.VerificationThreshold,
		HighAssuranceThreshold:   pb.HighAssuranceThreshold,
		ArenaFocusThreshold:      pb.ArenaFocusThreshold,
		VelocityBonusDays:        pb.VelocityBonusDays,
		VelocityBonusMultipliers: pb.VelocityBonusMultipliers,
		ArenaMultipliers:         arenaMultipliers,
		SlashPercentage:          pb.SlashPercentage,
		AppealDeposit:            pb.AppealDeposit,
		MaxIRsPerDay:             pb.MaxIrsPerDay,
		MaxIRsPerHour:            pb.MaxIrsPerHour,
		JackpotOdds:              pb.JackpotOdds,
		JackpotMultipliers:       pb.JackpotMultipliers,
		StalenessEnabled:         pb.StalenessEnabled,
		DegradationRatePerYear:   pb.DegradationRatePerYear,
		PoIRewardsEnabled:        pb.PoiRewardsEnabled,
		UserRewardSplitPercent:   pb.UserRewardSplitPercent,
		VelocityBonusEnabled:     pb.VelocityBonusEnabled,
	}
}

// ParamsToProto converts internal Params to proto Params type
func ParamsToProto(p Params) *confidencescorepb.Params {
	arenaMultipliers := make(map[uint64]float32)
	for threshold, multiplier := range p.ArenaMultipliers {
		arenaMultipliers[threshold] = multiplier
	}

	return &confidencescorepb.Params{
		VerificationThreshold:    p.VerificationThreshold,
		HighAssuranceThreshold:   p.HighAssuranceThreshold,
		ArenaFocusThreshold:      p.ArenaFocusThreshold,
		VelocityBonusDays:        p.VelocityBonusDays,
		VelocityBonusMultipliers: p.VelocityBonusMultipliers,
		ArenaMultipliers:         arenaMultipliers,
		SlashPercentage:          p.SlashPercentage,
		AppealDeposit:            p.AppealDeposit,
		MaxIrsPerDay:             p.MaxIRsPerDay,
		MaxIrsPerHour:            p.MaxIRsPerHour,
		JackpotOdds:              p.JackpotOdds,
		JackpotMultipliers:       p.JackpotMultipliers,
		StalenessEnabled:         p.StalenessEnabled,
		DegradationRatePerYear:   p.DegradationRatePerYear,
		PoiRewardsEnabled:        p.PoIRewardsEnabled,
		UserRewardSplitPercent:   p.UserRewardSplitPercent,
		VelocityBonusEnabled:     p.VelocityBonusEnabled,
	}
}

// Validate performs validation on the Params
func (p Params) Validate() error {
	if p.VerificationThreshold == 0 {
		return fmt.Errorf("verification threshold must be positive")
	}

	if p.HighAssuranceThreshold <= p.VerificationThreshold {
		return fmt.Errorf("high assurance threshold must be greater than verification threshold")
	}

	if p.ArenaFocusThreshold == 0 {
		return fmt.Errorf("arena focus threshold must be positive")
	}

	if len(p.VelocityBonusDays) != len(p.VelocityBonusMultipliers) {
		return fmt.Errorf("velocity bonus days and multipliers must have same length")
	}

	for i, multiplier := range p.VelocityBonusMultipliers {
		if multiplier < 1.0 {
			return fmt.Errorf("velocity bonus multiplier[%d] must be >= 1.0, got %f", i, multiplier)
		}
	}

	if len(p.ArenaMultipliers) == 0 {
		return fmt.Errorf("arena multipliers cannot be empty")
	}

	for threshold, multiplier := range p.ArenaMultipliers {
		if threshold == 0 {
			return fmt.Errorf("arena multiplier threshold cannot be 0")
		}
		if multiplier < 1.0 {
			return fmt.Errorf("arena multiplier for threshold %d must be >= 1.0, got %f", threshold, multiplier)
		}
	}

	if p.SlashPercentage > 100 {
		return fmt.Errorf("slash percentage cannot exceed 100, got %d", p.SlashPercentage)
	}

	if p.AppealDeposit == "" {
		return fmt.Errorf("appeal deposit cannot be empty")
	}

	if p.MaxIRsPerHour > p.MaxIRsPerDay {
		return fmt.Errorf("max irs per hour cannot exceed max irs per day")
	}

	if len(p.JackpotOdds) != len(p.JackpotMultipliers) {
		return fmt.Errorf("jackpot odds and multipliers must have same length")
	}

	for i, multiplier := range p.JackpotMultipliers {
		if multiplier < 1.0 {
			return fmt.Errorf("jackpot multiplier[%d] must be >= 1.0, got %f", i, multiplier)
		}
	}

	// Validate PoI reward parameters
	if p.UserRewardSplitPercent > 100 {
		return fmt.Errorf("user reward split percent cannot exceed 100, got %d", p.UserRewardSplitPercent)
	}

	return nil
}
