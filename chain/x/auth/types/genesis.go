// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"

	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// GenesisState re-exports the protobuf genesis structure for convenience.
type GenesisState = authproto.GenesisState

// DefaultGenesis returns a conservative default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:                *DefaultParams(),
		Roles:                 []authproto.Role{},
		RoleAssignments:       []authproto.RoleAssignment{},
		MultisigWallets:       []authproto.MultisigWallet{},
		MultisigProposals:     []authproto.MultisigProposal{},
		TimeLockedActions:     []authproto.TimeLockedAction{},
		EmergencyAdmins:       []authproto.EmergencyAdmin{},
		ValidatorKeyRotations: []authproto.ValidatorKeyRotation{},
		Sessions:              []authproto.Session{},
		RateLimitConfigs:      []authproto.RateLimitConfig{},
		AuditLogs:             []authproto.AuditLog{},
	}
}

// ValidateGenesis performs lightweight validation of the auth genesis state.
func ValidateGenesis(gen *GenesisState) error {
	if gen == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}

	for i := range gen.Roles {
		if err := ValidateRole(&gen.Roles[i]); err != nil {
			return fmt.Errorf("ValidateGenesis: invalid role at index %d: %w", i, err)
		}
	}

	for i := range gen.RoleAssignments {
		if err := ValidateRoleAssignment(&gen.RoleAssignments[i]); err != nil {
			return fmt.Errorf("ValidateGenesis: invalid role assignment at index %d: %w", i, err)
		}
	}

	for i := range gen.MultisigWallets {
		if err := ValidateMultisigWallet(&gen.MultisigWallets[i]); err != nil {
			return fmt.Errorf("ValidateGenesis: invalid multisig wallet at index %d: %w", i, err)
		}
	}

	for i := range gen.MultisigProposals {
		if err := ValidateMultisigProposal(&gen.MultisigProposals[i]); err != nil {
			return fmt.Errorf("ValidateGenesis: invalid multisig proposal at index %d: %w", i, err)
		}
	}

	for i := range gen.TimeLockedActions {
		if err := ValidateTimeLockedAction(&gen.TimeLockedActions[i]); err != nil {
			return fmt.Errorf("ValidateGenesis: invalid time-locked action at index %d: %w", i, err)
		}
	}

	for i := range gen.EmergencyAdmins {
		if err := ValidateEmergencyAdmin(&gen.EmergencyAdmins[i]); err != nil {
			return fmt.Errorf("ValidateGenesis: invalid emergency admin at index %d: %w", i, err)
		}
	}

	for i := range gen.Sessions {
		if err := ValidateSession(&gen.Sessions[i]); err != nil {
			return fmt.Errorf("ValidateGenesis: invalid session at index %d: %w", i, err)
		}
	}

	return nil
}
