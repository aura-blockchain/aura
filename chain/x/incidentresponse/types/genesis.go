// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import "fmt"

// GenesisState represents the genesis state for the incidentresponse module
type GenesisState struct {
	Params         *IncidentResponseParams `json:"params"`
	Incidents      []*Incident             `json:"incidents"`
	PauseState     *ChainPauseState        `json:"pause_state"`
	WalletLimits   []*WalletLimits         `json:"wallet_limits"`
	NextIncidentID uint64                  `json:"next_incident_id"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	defaultParams := DefaultParams()
	return &GenesisState{
		Params:    &defaultParams,
		Incidents: []*Incident{},
		PauseState: &ChainPauseState{
			IsPaused:   false,
			PauseLevel: PauseLevelNone,
		},
		WalletLimits:   []*WalletLimits{},
		NextIncidentID: 1,
	}
}

// Validate performs basic validation of genesis data
func (gs GenesisState) Validate() error {
	if gs.Params != nil {
		if err := gs.Params.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid params: %w", err)
		}
	}

	// Validate incidents
	incidentIDs := make(map[string]bool)
	for i, incident := range gs.Incidents {
		if incident == nil {
			return fmt.Errorf("incident %d is nil", i)
		}
		if incident.ID == "" {
			return fmt.Errorf("incident %d has empty ID", i)
		}
		if incidentIDs[incident.ID] {
			return fmt.Errorf("duplicate incident ID: %s", incident.ID)
		}
		incidentIDs[incident.ID] = true

		if incident.Title == "" {
			return fmt.Errorf("incident %s has empty title", incident.ID)
		}
		if incident.ReportedBy == "" {
			return fmt.Errorf("incident %s has empty reported_by", incident.ID)
		}
	}

	// Validate pause state
	if gs.PauseState != nil {
		if gs.PauseState.IsPaused {
			if gs.PauseState.PausedBy == "" {
				return fmt.Errorf("paused chain must have paused_by set")
			}
			if gs.PauseState.PauseLevel == PauseLevelNone {
				return fmt.Errorf("paused chain must have valid pause level")
			}
		}
	}

	// Validate wallet limits
	walletAddresses := make(map[string]bool)
	for i, limit := range gs.WalletLimits {
		if limit == nil {
			return fmt.Errorf("wallet limit %d is nil", i)
		}
		if limit.Address == "" {
			return fmt.Errorf("wallet limit %d has empty address", i)
		}
		if walletAddresses[limit.Address] {
			return fmt.Errorf("duplicate wallet limit for address: %s", limit.Address)
		}
		walletAddresses[limit.Address] = true
	}

	// Validate next incident ID
	if gs.NextIncidentID == 0 {
		return fmt.Errorf("next_incident_id must be greater than 0")
	}

	return nil
}
