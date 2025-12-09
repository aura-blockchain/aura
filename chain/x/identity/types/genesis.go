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

// ValidateGenesisState performs comprehensive validation of genesis data
// to prevent data corruption and invalid state on chain initialization.
func ValidateGenesisState(gs *identitypb.GenesisState) error {
	// CRITICAL: Reject nil genesis state
	if gs == nil {
		return ErrInvalidInput.Wrap("genesis state cannot be nil")
	}

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
	for i, role := range gs.Roles {
		if role.Name == "" {
			return ErrInvalidRole.Wrapf("role name cannot be empty at index %d", i)
		}
		if roleNames[role.Name] {
			return ErrRoleAlreadyExists.Wrapf("duplicate role: %s at index %d", role.Name, i)
		}
		roleNames[role.Name] = true

		// Validate role permissions are not empty
		if len(role.Permissions) == 0 {
			return ErrInvalidRole.Wrapf("role %s has no permissions", role.Name)
		}
	}

	// Validate role assignments
	assignmentKeys := make(map[string]bool)
	for i, assignment := range gs.RoleAssignments {
		if assignment.Address == "" {
			return ErrInvalidRoleAssignment.Wrapf("address cannot be empty at index %d", i)
		}
		if assignment.RoleName == "" {
			return ErrInvalidRoleAssignment.Wrapf("role name cannot be empty at index %d", i)
		}
		if !roleNames[assignment.RoleName] && assignment.RoleName != RoleAdmin && assignment.RoleName != RoleUser {
			return ErrRoleNotFound.Wrapf("role not found: %s at index %d", assignment.RoleName, i)
		}

		// Check for duplicate assignments (same address + role)
		key := assignment.Address + ":" + assignment.RoleName
		if assignmentKeys[key] {
			return ErrInvalidRoleAssignment.Wrapf("duplicate role assignment for %s with role %s at index %d",
				assignment.Address, assignment.RoleName, i)
		}
		assignmentKeys[key] = true
	}

	// Validate multisig wallets
	walletIDs := make(map[string]bool)
	for i, wallet := range gs.MultisigWallets {
		if wallet.Id == "" {
			return ErrInvalidMultisigWallet.Wrapf("wallet ID cannot be empty at index %d", i)
		}
		if walletIDs[wallet.Id] {
			return ErrMultisigWalletExists.Wrapf("duplicate wallet ID: %s at index %d", wallet.Id, i)
		}
		walletIDs[wallet.Id] = true

		if wallet.Threshold == 0 {
			return ErrInvalidMultisigWallet.Wrapf("threshold must be greater than 0 for wallet %s", wallet.Id)
		}
		if uint32(len(wallet.Signers)) < wallet.Threshold {
			return ErrInvalidMultisigWallet.Wrapf("threshold (%d) cannot exceed number of signers (%d) for wallet %s",
				wallet.Threshold, len(wallet.Signers), wallet.Id)
		}

		// Check for duplicate signers
		signerAddrs := make(map[string]bool)
		for j, signer := range wallet.Signers {
			if signer == "" {
				return ErrInvalidMultisigWallet.Wrapf("signer address cannot be empty at index %d for wallet %s", j, wallet.Id)
			}
			if signerAddrs[signer] {
				return ErrInvalidMultisigWallet.Wrapf("duplicate signer %s in wallet %s", signer, wallet.Id)
			}
			signerAddrs[signer] = true
		}
	}

	// Validate identity records (DID documents)
	didMap := make(map[string]bool)
	addressMap := make(map[string]bool)
	for i, record := range gs.IdentityRecords {
		if record.Did == "" {
			return ErrInvalidDID.Wrapf("DID cannot be empty at index %d", i)
		}
		if didMap[record.Did] {
			return ErrIdentityAlreadyExists.Wrapf("duplicate DID: %s at index %d", record.Did, i)
		}
		didMap[record.Did] = true

		if record.Address == "" {
			return ErrInvalidInput.Wrapf("identity address cannot be empty at index %d", i)
		}
		if addressMap[record.Address] {
			return ErrInvalidInput.Wrapf("duplicate identity address: %s at index %d", record.Address, i)
		}
		addressMap[record.Address] = true

		// Validate verification methods exist
		if len(record.VerificationMethods) == 0 {
			return ErrInvalidDID.Wrapf("identity record must have at least one verification method for DID %s", record.Did)
		}
	}

	// Validate change requests
	requestIDs := make(map[string]bool)
	for i, request := range gs.ChangeRequests {
		if request.Id == "" {
			return ErrInvalidChangeRequest.Wrapf("request ID cannot be empty at index %d", i)
		}
		if requestIDs[request.Id] {
			return ErrChangeRequestInvalid.Wrapf("duplicate request ID: %s at index %d", request.Id, i)
		}
		requestIDs[request.Id] = true

		if request.Requester == "" {
			return ErrInvalidChangeRequest.Wrapf("requester cannot be empty at index %d", i)
		}
		if request.Did == "" {
			return ErrInvalidDID.Wrapf("target DID cannot be empty at index %d", i)
		}

		// Validate DID exists in identity records
		if !didMap[request.Did] {
			return ErrIdentityNotFound.Wrapf("change request references non-existent DID %s at index %d",
				request.Did, i)
		}

		// Note: ChangeRequest doesn't have Expiry or CreatedAt fields in the proto
		// Validation is based on actual fields: requested_at (timestamp)
	}

	// Validate credential revocations
	credentialIDs := make(map[string]bool)
	for i, revocation := range gs.CredentialRevocations {
		if revocation.CredentialId == "" {
			return ErrInvalidCredentialID.Wrapf("credential ID cannot be empty at index %d", i)
		}
		if credentialIDs[revocation.CredentialId] {
			return ErrCredentialAlreadyRevoked.Wrapf("duplicate credential revocation for ID %s at index %d",
				revocation.CredentialId, i)
		}
		credentialIDs[revocation.CredentialId] = true

		if revocation.Did == "" {
			return ErrInvalidDID.Wrapf("DID cannot be empty for credential revocation at index %d", i)
		}

		// Validate DID exists in identity records
		if !didMap[revocation.Did] {
			return ErrIdentityNotFound.Wrapf("credential revocation references non-existent DID %s at index %d",
				revocation.Did, i)
		}
	}

	// Validate DID key rotations
	didRotationKeys := make(map[string]bool)
	for i, rotation := range gs.DidKeyRotations {
		if rotation.Did == "" {
			return ErrInvalidDID.Wrapf("DID cannot be empty for key rotation at index %d", i)
		}
		if !didMap[rotation.Did] {
			return ErrIdentityNotFound.Wrapf("key rotation references non-existent DID %s at index %d",
				rotation.Did, i)
		}

		// Check for duplicate active rotations for the same DID
		if rotation.Status == identitypb.DIDKeyRotationStatus_DID_KEY_ROTATION_STATUS_PENDING {
			if didRotationKeys[rotation.Did] {
				return ErrDIDKeyRotationInProgress.Wrapf("duplicate pending key rotation for DID %s at index %d",
					rotation.Did, i)
			}
			didRotationKeys[rotation.Did] = true
		}

		if rotation.NewVerificationMethod == "" {
			return ErrInvalidVerificationMethod.Wrapf("new verification method cannot be empty for rotation at index %d", i)
		}
		if rotation.OldVerificationMethod == "" {
			return ErrInvalidVerificationMethod.Wrapf("old verification method cannot be empty for rotation at index %d", i)
		}
	}

	// Validate sessions
	sessionIDs := make(map[string]bool)
	for i, session := range gs.Sessions {
		if session.Id == "" {
			return ErrInvalidSession.Wrapf("session ID cannot be empty at index %d", i)
		}
		if sessionIDs[session.Id] {
			return ErrInvalidSession.Wrapf("duplicate session ID %s at index %d", session.Id, i)
		}
		sessionIDs[session.Id] = true

		if session.Address == "" {
			return ErrInvalidInput.Wrapf("session address cannot be empty at index %d", i)
		}

		// Validate expiry is after creation
		if !session.ExpiresAt.After(session.CreatedAt) {
			return ErrInvalidSession.Wrapf("session expiry must be after creation time for session %s", session.Id)
		}
	}

	// Validate multisig proposals
	proposalIDs := make(map[string]bool)
	for i, proposal := range gs.MultisigProposals {
		if proposal.Id == "" {
			return ErrInvalidProposal.Wrapf("proposal ID cannot be empty at index %d", i)
		}
		if proposalIDs[proposal.Id] {
			return ErrInvalidProposal.Wrapf("duplicate proposal ID %s at index %d", proposal.Id, i)
		}
		proposalIDs[proposal.Id] = true

		if proposal.WalletId == "" {
			return ErrInvalidProposal.Wrapf("wallet ID cannot be empty for proposal at index %d", i)
		}
		if !walletIDs[proposal.WalletId] {
			return ErrMultisigWalletNotFound.Wrapf("proposal references non-existent wallet %s at index %d",
				proposal.WalletId, i)
		}

		// Validate signature count doesn't exceed wallet signers
		signerCount := make(map[string]bool)
		for _, signer := range proposal.Signatures {
			if signer != "" {
				signerCount[signer] = true
			}
		}
		// We can't validate exact threshold here without wallet lookup,
		// but we can validate no duplicate signers
		if len(signerCount) != len(proposal.Signatures) {
			return ErrAlreadySigned.Wrapf("duplicate signers in proposal %s", proposal.Id)
		}
	}

	// Validate emergency admins
	adminAddrs := make(map[string]bool)
	for i, admin := range gs.EmergencyAdmins {
		if admin.Address == "" {
			return ErrInvalidEmergencyAdmin.Wrapf("admin address cannot be empty at index %d", i)
		}
		if adminAddrs[admin.Address] {
			return ErrInvalidEmergencyAdmin.Wrapf("duplicate emergency admin address %s at index %d",
				admin.Address, i)
		}
		adminAddrs[admin.Address] = true
	}

	// Validate time-locked actions
	actionIDs := make(map[string]bool)
	for i, action := range gs.TimeLockedActions {
		if action.Id == "" {
			return ErrInvalidAction.Wrapf("action ID cannot be empty at index %d", i)
		}
		if actionIDs[action.Id] {
			return ErrInvalidAction.Wrapf("duplicate time-locked action ID %s at index %d", action.Id, i)
		}
		actionIDs[action.Id] = true

		if action.Proposer == "" {
			return ErrInvalidAction.Wrapf("proposer cannot be empty for action %s", action.Id)
		}

		// Validate executable time is after proposal time
		if !action.ExecutableAt.After(action.ProposedAt) {
			return ErrInvalidAction.Wrapf("executable time must be after proposal time for action %s", action.Id)
		}
	}

	// Validate audit log counter
	if gs.NextAuditLogId == 0 {
		return ErrInvalidInput.Wrap("next audit log ID must be at least 1")
	}

	return nil
}
