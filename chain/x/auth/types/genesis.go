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
        Params:               *DefaultParams(),
        Roles:                []authproto.Role{},
        RoleAssignments:      []authproto.RoleAssignment{},
        MultisigWallets:      []authproto.MultisigWallet{},
        MultisigProposals:    []authproto.MultisigProposal{},
        TimeLockedActions:    []authproto.TimeLockedAction{},
        EmergencyAdmins:      []authproto.EmergencyAdmin{},
        ValidatorKeyRotations: []authproto.ValidatorKeyRotation{},
        Sessions:             []authproto.Session{},
        RateLimitConfigs:     []authproto.RateLimitConfig{},
        AuditLogs:            []authproto.AuditLog{},
    }
}

// ValidateGenesis performs lightweight validation of the auth genesis state.
func ValidateGenesis(gen *GenesisState) error {
    if gen == nil {
        return fmt.Errorf("genesis state cannot be nil")
    }

    for _, role := range gen.Roles {
        if err := ValidateRole(role); err != nil {
            return err
        }
    }

    for _, assignment := range gen.RoleAssignments {
        if err := ValidateRoleAssignment(assignment); err != nil {
            return err
        }
    }

    for _, wallet := range gen.MultisigWallets {
        if err := ValidateMultisigWallet(wallet); err != nil {
            return err
        }
    }

    for _, proposal := range gen.MultisigProposals {
        if err := ValidateMultisigProposal(proposal); err != nil {
            return err
        }
    }

    for _, action := range gen.TimeLockedActions {
        if err := ValidateTimeLockedAction(action); err != nil {
            return err
        }
    }

    for _, admin := range gen.EmergencyAdmins {
        if err := ValidateEmergencyAdmin(admin); err != nil {
            return err
        }
    }

    for _, session := range gen.Sessions {
        if err := ValidateSession(session); err != nil {
            return err
        }
    }

    return nil
}
