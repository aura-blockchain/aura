package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// Common permissions
const (
	PermissionAdmin              = "admin"
	PermissionCreateRole         = "create_role"
	PermissionAssignRole         = "assign_role"
	PermissionRevokeRole         = "revoke_role"
	PermissionManageMultisig     = "manage_multisig"
	PermissionManageTimeLock     = "manage_timelock"
	PermissionManageEmergency    = "manage_emergency"
	PermissionRotateValidatorKey = "rotate_validator_key"
	PermissionManageSession      = "manage_session"
	PermissionViewAuditLogs      = "view_audit_logs"
)

// Predefined roles
const (
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
	RoleValidator = "validator"
	RoleUser      = "user"
)

// GenerateID generates a unique ID from components
func GenerateID(prefix string, components ...string) string {
	h := sha256.New()
	h.Write([]byte(prefix))
	for _, c := range components {
		h.Write([]byte(c))
	}
	h.Write([]byte(time.Now().String()))
	return prefix + "-" + hex.EncodeToString(h.Sum(nil))[:16]
}

// ValidateRole validates a role
func ValidateRole(role *authproto.Role) error {
	if role.Name == "" {
		return fmt.Errorf("role name cannot be empty")
	}
	if len(role.Permissions) == 0 {
		return fmt.Errorf("role must have at least one permission")
	}
	return nil
}

// ValidateRoleAssignment validates a role assignment
func ValidateRoleAssignment(assignment *authproto.RoleAssignment) error {
	if assignment.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if assignment.RoleName == "" {
		return fmt.Errorf("role name cannot be empty")
	}
	if assignment.AssignedBy == "" {
		return fmt.Errorf("assigned_by cannot be empty")
	}
	return nil
}

// ValidateMultisigWallet validates a multisig wallet
func ValidateMultisigWallet(wallet *authproto.MultisigWallet) error {
	if wallet.Id == "" {
		return fmt.Errorf("wallet ID cannot be empty")
	}
	if len(wallet.Signers) == 0 {
		return fmt.Errorf("wallet must have at least one signer")
	}
	if wallet.Threshold == 0 {
		return fmt.Errorf("threshold must be greater than 0")
	}
	if wallet.Threshold > uint32(len(wallet.Signers)) {
		return fmt.Errorf("threshold cannot be greater than number of signers")
	}

	// Validate wallet type configuration
	switch wallet.WalletType {
	case authproto.WalletType_WALLET_TYPE_3_OF_5:
		if len(wallet.Signers) != 5 || wallet.Threshold != 3 {
			return fmt.Errorf("3-of-5 wallet must have 5 signers and threshold of 3")
		}
	case authproto.WalletType_WALLET_TYPE_5_OF_7:
		if len(wallet.Signers) != 7 || wallet.Threshold != 5 {
			return fmt.Errorf("5-of-7 wallet must have 7 signers and threshold of 5")
		}
	case authproto.WalletType_WALLET_TYPE_CUSTOM:
		// Custom wallets have flexible configuration
	default:
		return fmt.Errorf("invalid wallet type")
	}

	return nil
}

// ValidateMultisigProposal validates a multisig proposal
func ValidateMultisigProposal(proposal *authproto.MultisigProposal) error {
	if proposal.Id == "" {
		return fmt.Errorf("proposal ID cannot be empty")
	}
	if proposal.WalletId == "" {
		return fmt.Errorf("wallet ID cannot be empty")
	}
	if proposal.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if len(proposal.Payload) == 0 {
		return fmt.Errorf("payload cannot be empty")
	}
	return nil
}

// IsProposalApproved checks if a proposal has enough signatures
func IsProposalApproved(proposal *authproto.MultisigProposal, wallet *authproto.MultisigWallet) bool {
	return uint32(len(proposal.Signatures)) >= wallet.Threshold
}

// IsProposalExpired checks if a proposal has expired
func IsProposalExpired(proposal *authproto.MultisigProposal) bool {
	if proposal.ExpiresAt == nil {
		return false
	}
	return time.Now().After(proposal.ExpiresAt.AsTime())
}

// ValidateTimeLockedAction validates a time-locked action
func ValidateTimeLockedAction(action *authproto.TimeLockedAction) error {
	if action.Id == "" {
		return fmt.Errorf("action ID cannot be empty")
	}
	if action.ActionType == "" {
		return fmt.Errorf("action type cannot be empty")
	}
	if len(action.Payload) == 0 {
		return fmt.Errorf("payload cannot be empty")
	}
	if action.Proposer == "" {
		return fmt.Errorf("proposer cannot be empty")
	}
	if action.DelaySeconds == 0 {
		return fmt.Errorf("delay must be greater than 0")
	}
	return nil
}

// IsActionReady checks if a time-locked action is ready for execution
func IsActionReady(action *authproto.TimeLockedAction) bool {
	if action.ExecutableAt == nil {
		return false
	}
	execTime := action.ExecutableAt.AsTime()
	return time.Now().After(execTime) || time.Now().Equal(execTime)
}

// ValidateEmergencyAdmin validates an emergency admin
func ValidateEmergencyAdmin(admin *authproto.EmergencyAdmin) error {
	if admin.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if len(admin.Privileges) == 0 {
		return fmt.Errorf("emergency admin must have at least one privilege")
	}
	if admin.ActivatedBy == "" {
		return fmt.Errorf("activated_by cannot be empty")
	}
	return nil
}

// IsEmergencyAdminActive checks if an emergency admin is currently active
func IsEmergencyAdminActive(admin *authproto.EmergencyAdmin) bool {
	if !admin.IsActive {
		return false
	}
	if admin.ExpiresAt == nil {
		return true
	}
	return time.Now().Before(admin.ExpiresAt.AsTime())
}

// ValidateSession validates a session
func ValidateSession(session *authproto.Session) error {
	if session.SessionId == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	if session.UserAddress == "" {
		return fmt.Errorf("user address cannot be empty")
	}
	return nil
}

// IsSessionActive checks if a session is currently active
func IsSessionActive(session *authproto.Session) bool {
	if !session.IsActive {
		return false
	}
	if session.ExpiresAt == nil {
		return false
	}
	return time.Now().Before(session.ExpiresAt.AsTime())
}

// ValidateRateLimitConfig validates a rate limit config
func ValidateRateLimitConfig(config *authproto.RateLimitConfig) error {
	if config.UserAddress == "" {
		return fmt.Errorf("user address cannot be empty")
	}
	return nil
}

// IsRateLimited checks if a user has exceeded their rate limit
func IsRateLimited(config *authproto.RateLimitConfig) bool {
	now := time.Now()
	windowStart := config.WindowStart

	// Check minute limit
	if windowStart != nil && now.Sub(windowStart.AsTime()) < time.Minute {
		if config.CurrentMinuteCount >= config.RequestsPerMinute {
			return true
		}
	}

	// Check hour limit
	if windowStart != nil && now.Sub(windowStart.AsTime()) < time.Hour {
		if config.CurrentHourCount >= config.RequestsPerHour {
			return true
		}
	}

	// Check day limit
	if windowStart != nil && now.Sub(windowStart.AsTime()) < 24*time.Hour {
		if config.CurrentDayCount >= config.RequestsPerDay {
			return true
		}
	}

	return false
}

// DefaultParams returns default parameters for the auth module
func DefaultParams() *authproto.Params {
	return &authproto.Params{
		SessionTimeoutSeconds:         3600,   // 1 hour
		DefaultTimelockDelaySeconds:   86400,  // 24 hours
		DefaultRequestsPerMinute:      60,     // 60 requests per minute
		DefaultRequestsPerHour:        3600,   // 3600 requests per hour
		DefaultRequestsPerDay:         86400,  // 86400 requests per day
		MultisigProposalExpirySeconds: 604800, // 7 days
		AuditLoggingEnabled:           true,
	}
}
