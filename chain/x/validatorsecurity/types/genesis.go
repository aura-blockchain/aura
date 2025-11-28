package types

import (
	"fmt"
)

// NewGenesisState creates a new GenesisState instance
func NewGenesisState(params *ValidatorSecurityParams) *GenesisState {
	return &GenesisState{
		Params:              params,
		Validators:          []*ValidatorSecurityInfo{},
		DoubleSignEvidences: []*DoubleSignEvidence{},
		DowntimeInfractions: []*DowntimeInfraction{},
		Alerts:              []*ValidatorAlert{},
		SentryNodes:         []*SentryNodeInfo{},
	}
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams())
}

// ValidateGenesisState performs basic genesis state validation
func ValidateGenesisState(gs *GenesisState) error {
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
	seenAlertIDs := make(map[string]bool)
	for i, alert := range gs.Alerts {
		if err := ValidateValidatorAlert(alert); err != nil {
			return fmt.Errorf("invalid alert at index %d: %w", i, err)
		}
		if seenAlertIDs[alert.Id] {
			return fmt.Errorf("duplicate alert ID: %s", alert.Id)
		}
		seenAlertIDs[alert.Id] = true
	}

	// Validate sentry nodes
	seenSentryNodes := make(map[string]bool)
	for i, node := range gs.SentryNodes {
		if err := ValidateSentryNodeInfo(node); err != nil {
			return fmt.Errorf("invalid sentry node at index %d: %w", i, err)
		}
		// Check for duplicate sentry node addresses
		if seenSentryNodes[node.Address] {
			return fmt.Errorf("duplicate sentry node address: %s", node.Address)
		}
		seenSentryNodes[node.Address] = true
	}

	return nil
}

func ValidateValidatorInfo(info *ValidatorSecurityInfo) error {
	if info == nil {
		return fmt.Errorf("validator info cannot be nil")
	}
	if info.ValidatorAddress == "" {
		return fmt.Errorf("validator address cannot be empty")
	}
	// Validate latitude and longitude if set
	if info.Latitude < -90 || info.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90, got %f", info.Latitude)
	}
	if info.Longitude < -180 || info.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180, got %f", info.Longitude)
	}
	return nil
}

func ValidateDoubleSignEvidence(evidence *DoubleSignEvidence) error {
	if evidence == nil {
		return fmt.Errorf("evidence cannot be nil")
	}
	if evidence.ValidatorAddress == "" {
		return fmt.Errorf("validator address cannot be empty")
	}
	return nil
}

func ValidateDowntimeInfraction(infraction *DowntimeInfraction) error {
	if infraction == nil {
		return fmt.Errorf("infraction cannot be nil")
	}
	if infraction.ValidatorAddress == "" {
		return fmt.Errorf("validator address cannot be empty")
	}
	return nil
}

func ValidateValidatorAlert(alert *ValidatorAlert) error {
	if alert == nil {
		return fmt.Errorf("alert cannot be nil")
	}
	if alert.ValidatorAddress == "" {
		return fmt.Errorf("validator address cannot be empty")
	}
	return nil
}

func ValidateSentryNodeInfo(info *SentryNodeInfo) error {
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
