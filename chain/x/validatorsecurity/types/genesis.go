package types

import (
	"fmt"
	pb "github.com/aequitas/aura/proto/aura/validatorsecurity/v1beta1"
)

// NewGenesisState creates a new GenesisState instance
func NewGenesisState(params *pb.ValidatorSecurityParams) *pb.GenesisState {
	return &pb.GenesisState{
		Params:              params,
		Validators:          []*pb.ValidatorSecurityInfo{},
		DoubleSignEvidences: []*pb.DoubleSignEvidence{},
		DowntimeInfractions: []*pb.DowntimeInfraction{},
		Alerts:              []*pb.ValidatorAlert{},
		SentryNodes:         []*pb.SentryNodeInfo{},
	}
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *pb.GenesisState {
	return NewGenesisState(DefaultParams())
}

// ValidateGenesisState performs basic genesis state validation
func ValidateGenesisState(gs *pb.GenesisState) error {
	if gs.Params == nil {
		return fmt.Errorf("params cannot be nil")
	}
	if err := ValidateParams(gs.Params); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate validators
	seenValidators := make(map[string]bool)
	for i, validator := range gs.Validators {
		if err := ValidateValidatorInfo(validator); err != nil {
			return fmt.Errorf("invalid validator at index %d: %w", i, err)
		}
		if seenValidators[validator.ValidatorAddress] {
			return fmt.Errorf("duplicate validator address: %s", validator.ValidatorAddress)
		}
		seenValidators[validator.ValidatorAddress] = true
	}

	// Validate double sign evidences
	for i, evidence := range gs.DoubleSignEvidences {
		if err := ValidateDoubleSignEvidence(evidence); err != nil {
			return fmt.Errorf("invalid double sign evidence at index %d: %w", i, err)
		}
	}

	// Validate downtime infractions
	for i, infraction := range gs.DowntimeInfractions {
		if err := ValidateDowntimeInfraction(infraction); err != nil {
			return fmt.Errorf("invalid downtime infraction at index %d: %w", i, err)
		}
	}

	// Validate alerts
	for i, alert := range gs.Alerts {
		if err := ValidateValidatorAlert(alert); err != nil {
			return fmt.Errorf("invalid alert at index %d: %w", i, err)
		}
	}

	// Validate sentry nodes
	seenSentryNodes := make(map[string]bool)
	for i, node := range gs.SentryNodes {
		if err := ValidateSentryNodeInfo(node); err != nil {
			return fmt.Errorf("invalid sentry node at index %d: %w", i, err)
		}
		key := fmt.Sprintf("%s:%s", node.ValidatorAddress, node.Address)
		if seenSentryNodes[key] {
			return fmt.Errorf("duplicate sentry node: %s", key)
		}
		seenSentryNodes[key] = true
	}

	return nil
}

func ValidateValidatorInfo(info *pb.ValidatorSecurityInfo) error {
	if info == nil {
		return fmt.Errorf("validator info cannot be nil")
	}
	if info.ValidatorAddress == "" {
		return fmt.Errorf("validator address cannot be empty")
	}
	return nil
}

func ValidateDoubleSignEvidence(evidence *pb.DoubleSignEvidence) error {
	if evidence == nil {
		return fmt.Errorf("evidence cannot be nil")
	}
	if evidence.ValidatorAddress == "" {
		return fmt.Errorf("validator address cannot be empty")
	}
	return nil
}

func ValidateDowntimeInfraction(infraction *pb.DowntimeInfraction) error {
	if infraction == nil {
		return fmt.Errorf("infraction cannot be nil")
	}
	if infraction.ValidatorAddress == "" {
		return fmt.Errorf("validator address cannot be empty")
	}
	return nil
}

func ValidateValidatorAlert(alert *pb.ValidatorAlert) error {
	if alert == nil {
		return fmt.Errorf("alert cannot be nil")
	}
	if alert.ValidatorAddress == "" {
		return fmt.Errorf("validator address cannot be empty")
	}
	return nil
}

func ValidateSentryNodeInfo(info *pb.SentryNodeInfo) error {
	if info == nil {
		return fmt.Errorf("sentry node info cannot be nil")
	}
	if info.ValidatorAddress == "" {
		return fmt.Errorf("validator address cannot be empty")
	}
	if info.Address == "" {
		return fmt.Errorf("node address cannot be empty")
	}
	return nil
}
