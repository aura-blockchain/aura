package v1beta1

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/common/validation"
)

const (
	// MaxPayloadSize is the maximum size for proposal payloads
	MaxPayloadSize = 1024 * 1024 // 1MB
	// MinPayloadSize is the minimum size for proposal payloads
	MinPayloadSize = 1
	// MaxThreshold is the maximum threshold for multisig wallets
	MaxThreshold = 100
	// MinThreshold is the minimum threshold for multisig wallets
	MinThreshold = 1
	// MaxSigners is the maximum number of signers in a multisig wallet
	MaxSigners = 100
	// MinSigners is the minimum number of signers in a multisig wallet
	MinSigners = 1
	// MaxDelaySeconds is the maximum delay for time-locked actions (30 days)
	MaxDelaySeconds = uint64(30 * 24 * 60 * 60)
	// MinDelaySeconds is the minimum delay for time-locked actions (1 hour)
	MinDelaySeconds = uint64(60 * 60)
	// MaxPrivileges is the maximum number of privileges
	MaxPrivileges = 50
)

// ValidateBasic implements the sdk.Msg interface for MsgCreateRole
func (m *MsgCreateRole) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Validate role name
	if err := validation.ValidateBoundedString(m.Name, 1, validation.MaxNameLength, "name"); err != nil {
		return err
	}

	// Validate permissions (must have at least one)
	if err := validation.ValidateStringSlice(m.Permissions, "permissions"); err != nil {
		return err
	}

	// Limit the number of permissions
	if len(m.Permissions) > MaxPrivileges {
		return fmt.Errorf("permissions: cannot exceed %d permissions, got %d", MaxPrivileges, len(m.Permissions))
	}

	// Validate description (optional but if present must be valid)
	if m.Description != "" {
		if err := validation.ValidateBoundedString(m.Description, 0, validation.MaxDescriptionLength, "description"); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAssignRole
func (m *MsgAssignRole) ValidateBasic() error {
	// Validate assigner address
	if err := validation.ValidateAccAddress(m.Assigner); err != nil {
		return fmt.Errorf("assigner: %w", err)
	}

	// Validate target address
	if err := validation.ValidateAccAddress(m.Address); err != nil {
		return fmt.Errorf("address: %w", err)
	}

	// Validate role name
	if err := validation.ValidateBoundedString(m.RoleName, 1, validation.MaxNameLength, "role_name"); err != nil {
		return err
	}

	// Validate expiry (if set, must be non-negative)
	if err := validation.ValidateTimestamp(m.ExpiresInSeconds, "expires_in_seconds"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgRevokeRole
func (m *MsgRevokeRole) ValidateBasic() error {
	// Validate revoker address
	if err := validation.ValidateAccAddress(m.Revoker); err != nil {
		return fmt.Errorf("revoker: %w", err)
	}

	// Validate target address
	if err := validation.ValidateAccAddress(m.Address); err != nil {
		return fmt.Errorf("address: %w", err)
	}

	// Validate role name
	if err := validation.ValidateBoundedString(m.RoleName, 1, validation.MaxNameLength, "role_name"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCreateMultisigWallet
func (m *MsgCreateMultisigWallet) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Validate signers (must have at least one)
	if len(m.Signers) < MinSigners {
		return fmt.Errorf("signers: must have at least %d signer", MinSigners)
	}

	if len(m.Signers) > MaxSigners {
		return fmt.Errorf("signers: cannot exceed %d signers, got %d", MaxSigners, len(m.Signers))
	}

	// Validate each signer address
	for i, signer := range m.Signers {
		if err := validation.ValidateAccAddress(signer); err != nil {
			return fmt.Errorf("signers[%d]: %w", i, err)
		}
	}

	// Validate threshold
	if err := validation.ValidateBoundedUint32(m.Threshold, MinThreshold, MaxThreshold, "threshold"); err != nil {
		return err
	}

	// Threshold must not exceed number of signers
	if m.Threshold > uint32(len(m.Signers)) {
		return fmt.Errorf("threshold cannot exceed number of signers: %d > %d", m.Threshold, len(m.Signers))
	}

	// Wallet type is an enum, validated at protobuf level

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCreateMultisigProposal
func (m *MsgCreateMultisigProposal) ValidateBasic() error {
	// Validate proposer address
	if err := validation.ValidateAccAddress(m.Proposer); err != nil {
		return fmt.Errorf("proposer: %w", err)
	}

	// Validate wallet ID
	if err := validation.ValidateID(m.WalletId, "wallet_id"); err != nil {
		return err
	}

	// Validate title
	if err := validation.ValidateBoundedString(m.Title, 1, validation.MaxNameLength, "title"); err != nil {
		return err
	}

	// Validate description
	if err := validation.ValidateBoundedString(m.Description, 1, validation.MaxDescriptionLength, "description"); err != nil {
		return err
	}

	// Validate payload
	if err := validation.ValidateBytes(m.Payload, MinPayloadSize, MaxPayloadSize, "payload"); err != nil {
		return err
	}

	// Validate expiry (must be positive)
	if err := validation.ValidatePositiveTimestamp(m.ExpiresInSeconds, "expires_in_seconds"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgSignMultisigProposal
func (m *MsgSignMultisigProposal) ValidateBasic() error {
	// Validate signer address
	if err := validation.ValidateAccAddress(m.Signer); err != nil {
		return fmt.Errorf("signer: %w", err)
	}

	// Validate proposal ID
	if err := validation.ValidateID(m.ProposalId, "proposal_id"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgExecuteMultisigProposal
func (m *MsgExecuteMultisigProposal) ValidateBasic() error {
	// Validate executor address
	if err := validation.ValidateAccAddress(m.Executor); err != nil {
		return fmt.Errorf("executor: %w", err)
	}

	// Validate proposal ID
	if err := validation.ValidateID(m.ProposalId, "proposal_id"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgProposeTimeLockedAction
func (m *MsgProposeTimeLockedAction) ValidateBasic() error {
	// Validate proposer address
	if err := validation.ValidateAccAddress(m.Proposer); err != nil {
		return fmt.Errorf("proposer: %w", err)
	}

	// Validate action type
	if err := validation.ValidateBoundedString(m.ActionType, 1, 64, "action_type"); err != nil {
		return err
	}

	// Validate payload
	if err := validation.ValidateBytes(m.Payload, MinPayloadSize, MaxPayloadSize, "payload"); err != nil {
		return err
	}

	// Validate delay
	if err := validation.ValidateBoundedUint64(m.DelaySeconds, MinDelaySeconds, MaxDelaySeconds, "delay_seconds"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgExecuteTimeLockedAction
func (m *MsgExecuteTimeLockedAction) ValidateBasic() error {
	// Validate executor address
	if err := validation.ValidateAccAddress(m.Executor); err != nil {
		return fmt.Errorf("executor: %w", err)
	}

	// Validate action ID
	if err := validation.ValidateID(m.ActionId, "action_id"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCancelTimeLockedAction
func (m *MsgCancelTimeLockedAction) ValidateBasic() error {
	// Validate canceller address
	if err := validation.ValidateAccAddress(m.Canceller); err != nil {
		return fmt.Errorf("canceller: %w", err)
	}

	// Validate action ID
	if err := validation.ValidateID(m.ActionId, "action_id"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgActivateEmergencyAdmin
func (m *MsgActivateEmergencyAdmin) ValidateBasic() error {
	// Validate activator address
	if err := validation.ValidateAccAddress(m.Activator); err != nil {
		return fmt.Errorf("activator: %w", err)
	}

	// Validate admin address
	if err := validation.ValidateAccAddress(m.AdminAddress); err != nil {
		return fmt.Errorf("admin_address: %w", err)
	}

	// Validate privileges (must have at least one)
	if err := validation.ValidateStringSlice(m.Privileges, "privileges"); err != nil {
		return err
	}

	// Limit the number of privileges
	if len(m.Privileges) > MaxPrivileges {
		return fmt.Errorf("privileges: cannot exceed %d privileges, got %d", MaxPrivileges, len(m.Privileges))
	}

	// Validate expiry (must be positive)
	if err := validation.ValidatePositiveTimestamp(m.ExpiresInSeconds, "expires_in_seconds"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgDeactivateEmergencyAdmin
func (m *MsgDeactivateEmergencyAdmin) ValidateBasic() error {
	// Validate deactivator address
	if err := validation.ValidateAccAddress(m.Deactivator); err != nil {
		return fmt.Errorf("deactivator: %w", err)
	}

	// Validate admin address
	if err := validation.ValidateAccAddress(m.AdminAddress); err != nil {
		return fmt.Errorf("admin_address: %w", err)
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgInitiateValidatorKeyRotation
func (m *MsgInitiateValidatorKeyRotation) ValidateBasic() error {
	// Validate initiator address
	if err := validation.ValidateAccAddress(m.Initiator); err != nil {
		return fmt.Errorf("initiator: %w", err)
	}

	// Validate validator address
	if err := validation.ValidateAccAddress(m.ValidatorAddress); err != nil {
		return fmt.Errorf("validator_address: %w", err)
	}

	// Validate new consensus pubkey (must be non-empty)
	if err := validation.ValidateBoundedString(m.NewConsensusPubkey, 1, 256, "new_consensus_pubkey"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCompleteValidatorKeyRotation
func (m *MsgCompleteValidatorKeyRotation) ValidateBasic() error {
	// Validate completer address
	if err := validation.ValidateAccAddress(m.Completer); err != nil {
		return fmt.Errorf("completer: %w", err)
	}

	// Validate validator address
	if err := validation.ValidateAccAddress(m.ValidatorAddress); err != nil {
		return fmt.Errorf("validator_address: %w", err)
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCreateSession
func (m *MsgCreateSession) ValidateBasic() error {
	// Validate user address
	if err := validation.ValidateAccAddress(m.UserAddress); err != nil {
		return fmt.Errorf("user_address: %w", err)
	}

	// Validate IP address (optional, but if present must be valid)
	if m.IpAddress != "" {
		if err := validation.ValidateBoundedString(m.IpAddress, 7, 45, "ip_address"); err != nil {
			return err
		}
	}

	// Metadata is optional and map types are validated at runtime

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgRevokeSession
func (m *MsgRevokeSession) ValidateBasic() error {
	// Validate user address
	if err := validation.ValidateAccAddress(m.UserAddress); err != nil {
		return fmt.Errorf("user_address: %w", err)
	}

	// Validate session ID
	if err := validation.ValidateID(m.SessionId, "session_id"); err != nil {
		return err
	}

	return nil
}
















