package types

import (
	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
)

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *identitypb.GenesisState {
	return &identitypb.GenesisState{
		Params:                   *DefaultParams(),
		Roles:                    []identitypb.Role{},
		RoleAssignments:          []identitypb.RoleAssignment{},
		AuditLogs:                []identitypb.AuditLog{},
		Sessions:                 []identitypb.Session{},
		RateLimits:               []identitypb.RateLimitConfig{},
		MultisigWallets:          []identitypb.MultisigWallet{},
		MultisigProposals:        []identitypb.MultisigProposal{},
		TimeLockedActions:        []identitypb.TimeLockedAction{},
		EmergencyAdmins:          []identitypb.EmergencyAdmin{},
		ValidatorRotations:       []identitypb.ValidatorKeyRotation{},
		IdentityRecords:          []identitypb.IdentityRecord{},
		CredentialRevocations:    []identitypb.CredentialRevocation{},
		DidKeyRotations:          []identitypb.DIDKeyRotation{},
		DidKeyHistories:          []identitypb.DIDKeyHistory{},
		ChangeRequests:           []identitypb.ChangeRequest{},
		ChangeHistory:            []identitypb.ChangeHistory{},
		IdentityChangesSuspended: false,
		NextAuditLogId:           1,
	}
}

// ValidateGenesisState performs basic validation of genesis data
func ValidateGenesisState(gs *identitypb.GenesisState) error {
	// Note: Params.Auth and Params.Change are non-nullable in proto
	// They will never be nil, just validate their contents

	if gs.Params.Auth.MaxRolesPerAccount == 0 {
		return ErrInvalidInput.Wrap("max_roles_per_account must be greater than 0")
	}
	if gs.Params.Auth.DefaultRequestsPerMinute == 0 {
		return ErrInvalidInput.Wrap("default_requests_per_minute must be greater than 0")
	}
	if gs.Params.Change.MaxRequestsPerWalletPerMonth == 0 {
		return ErrInvalidInput.Wrap("max_requests_per_wallet_per_month must be greater than 0")
	}

	// Validate roles
	roleNames := make(map[string]bool)
	for _, role := range gs.Roles {
		if role.Name == "" {
			return ErrInvalidRole.Wrap("role name cannot be empty")
		}
		if roleNames[role.Name] {
			return ErrRoleAlreadyExists.Wrapf("duplicate role: %s", role.Name)
		}
		roleNames[role.Name] = true
	}

	// Validate role assignments
	for _, assignment := range gs.RoleAssignments {
		if assignment.Address == "" {
			return ErrInvalidRoleAssignment.Wrap("address cannot be empty")
		}
		if assignment.RoleName == "" {
			return ErrInvalidRoleAssignment.Wrap("role name cannot be empty")
		}
		if !roleNames[assignment.RoleName] && assignment.RoleName != RoleAdmin && assignment.RoleName != RoleUser {
			return ErrRoleNotFound.Wrapf("role not found: %s", assignment.RoleName)
		}
	}

	// Validate multisig wallets
	walletIDs := make(map[string]bool)
	for _, wallet := range gs.MultisigWallets {
		if wallet.Id == "" {
			return ErrInvalidMultisigWallet.Wrap("wallet ID cannot be empty")
		}
		if walletIDs[wallet.Id] {
			return ErrMultisigWalletExists.Wrapf("duplicate wallet ID: %s", wallet.Id)
		}
		walletIDs[wallet.Id] = true

		if wallet.Threshold == 0 {
			return ErrInvalidMultisigWallet.Wrap("threshold must be greater than 0")
		}
		if uint32(len(wallet.Signers)) < wallet.Threshold {
			return ErrInvalidMultisigWallet.Wrap("threshold cannot exceed number of signers")
		}
	}

	// Validate identity records
	didMap := make(map[string]bool)
	for _, record := range gs.IdentityRecords {
		if record.Did == "" {
			return ErrInvalidDID.Wrap("DID cannot be empty")
		}
		if didMap[record.Did] {
			return ErrIdentityAlreadyExists.Wrapf("duplicate DID: %s", record.Did)
		}
		didMap[record.Did] = true

		if record.Address == "" {
			return ErrInvalidInput.Wrap("identity address cannot be empty")
		}
	}

	// Validate change requests
	requestIDs := make(map[string]bool)
	for _, request := range gs.ChangeRequests {
		if request.Id == "" {
			return ErrInvalidChangeRequest.Wrap("request ID cannot be empty")
		}
		if requestIDs[request.Id] {
			return ErrChangeRequestInvalid.Wrapf("duplicate request ID: %s", request.Id)
		}
		requestIDs[request.Id] = true

		if request.Requester == "" {
			return ErrInvalidChangeRequest.Wrap("requester cannot be empty")
		}
		if request.Did == "" {
			return ErrInvalidDID.Wrap("target DID cannot be empty")
		}
	}

	return nil
}
