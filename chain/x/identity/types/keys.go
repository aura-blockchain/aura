package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "identity"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	// ParamsKey stores module parameters
	ParamsKey = []byte{0x00}

	// Role and Permission prefixes (from auth module)
	RolePrefix            = []byte{0x01}
	RoleAssignmentPrefix  = []byte{0x02}
	PermissionGrantPrefix = []byte{0x03}
	AuditLogPrefix        = []byte{0x04}

	// Account and Session prefixes (from auth module)
	AccountPrefix        = []byte{0x05}
	SessionPrefix        = []byte{0x06}
	UserSessionsPrefix   = []byte{0x07}
	RateLimitConfigPrefix = []byte{0x08}

	// Multisig and Time-Lock prefixes (from auth module)
	MultisigWalletPrefix   = []byte{0x09}
	MultisigProposalPrefix = []byte{0x0a}
	TimeLockedActionPrefix = []byte{0x0b}

	// Emergency Admin prefixes (from auth module)
	EmergencyAdminPrefix  = []byte{0x0c}
	EmergencyActionPrefix = []byte{0x0d}

	// Validator prefixes (from auth module)
	ValidatorRotationPrefix = []byte{0x0e}

	// Identity Change prefixes (from identitychange module)
	IdentityRecordPrefix  = []byte{0x10}
	ChangeRequestPrefix   = []byte{0x11}
	ChangeHistoryPrefix   = []byte{0x12}
	RecoveryRecordPrefix  = []byte{0x13}
	VerificationPrefix    = []byte{0x14}
	DelegationPrefix      = []byte{0x15}
	FederationPrefix      = []byte{0x16}
	CrossChainLinkPrefix  = []byte{0x17}

	// Counter prefixes
	AuditLogCounterPrefix       = []byte{0x20}
	ChangeRequestCounterPrefix  = []byte{0x21}

	// Suspended flag
	SuspendedKey = []byte{0x30}
)

// ============================================================================
// Role and Permission Key Functions
// ============================================================================

// GetRoleKey returns the store key for a role
func GetRoleKey(roleID string) []byte {
	return append(RolePrefix, []byte(roleID)...)
}

// GetRoleAssignmentKey returns the store key for a role assignment
func GetRoleAssignmentKey(address string) []byte {
	return append(RoleAssignmentPrefix, []byte(address)...)
}

// GetPermissionGrantKey returns the store key for a permission grant
func GetPermissionGrantKey(grantee, permission string) []byte {
	key := append(PermissionGrantPrefix, []byte(grantee)...)
	key = append(key, []byte(":")...)
	key = append(key, []byte(permission)...)
	return key
}

// GetAuditLogKey returns the store key for an audit log entry
func GetAuditLogKey(logID uint64) []byte {
	return append(AuditLogPrefix, sdk.Uint64ToBigEndian(logID)...)
}

// ============================================================================
// Account and Session Key Functions
// ============================================================================

// GetAccountKey returns the store key for an account
func GetAccountKey(address string) []byte {
	return append(AccountPrefix, []byte(address)...)
}

// GetSessionKey returns the store key for a session
func GetSessionKey(sessionID string) []byte {
	return append(SessionPrefix, []byte(sessionID)...)
}

// GetUserSessionsKey returns the store key for a user's sessions list
func GetUserSessionsKey(userAddress string) []byte {
	return append(UserSessionsPrefix, []byte(userAddress)...)
}

// GetRateLimitConfigKey returns the store key for rate limit config
func GetRateLimitConfigKey(userAddress string) []byte {
	return append(RateLimitConfigPrefix, []byte(userAddress)...)
}

// ============================================================================
// Multisig and Time-Lock Key Functions
// ============================================================================

// GetMultisigWalletKey returns the store key for a multisig wallet
func GetMultisigWalletKey(walletID string) []byte {
	return append(MultisigWalletPrefix, []byte(walletID)...)
}

// GetMultisigProposalKey returns the store key for a multisig proposal
func GetMultisigProposalKey(proposalID string) []byte {
	return append(MultisigProposalPrefix, []byte(proposalID)...)
}

// GetTimeLockedActionKey returns the store key for a time-locked action
func GetTimeLockedActionKey(actionID string) []byte {
	return append(TimeLockedActionPrefix, []byte(actionID)...)
}

// ============================================================================
// Emergency Admin Key Functions
// ============================================================================

// GetEmergencyAdminKey returns the store key for an emergency admin
func GetEmergencyAdminKey(address string) []byte {
	return append(EmergencyAdminPrefix, []byte(address)...)
}

// GetEmergencyActionKey returns the store key for an emergency action
func GetEmergencyActionKey(actionID string) []byte {
	return append(EmergencyActionPrefix, []byte(actionID)...)
}

// ============================================================================
// Validator Key Functions
// ============================================================================

// GetValidatorRotationKey returns the store key for validator key rotation
func GetValidatorRotationKey(validatorAddress string) []byte {
	return append(ValidatorRotationPrefix, []byte(validatorAddress)...)
}

// ============================================================================
// Identity Change Key Functions
// ============================================================================

// GetIdentityRecordKey returns the store key for an identity record
func GetIdentityRecordKey(did string) []byte {
	return append(IdentityRecordPrefix, []byte(did)...)
}

// GetChangeRequestKey returns the store key for a change request
func GetChangeRequestKey(requestID string) []byte {
	return append(ChangeRequestPrefix, []byte(requestID)...)
}

// GetChangeHistoryKey returns the store key for change history
func GetChangeHistoryKey(did string, height uint64) []byte {
	key := append(ChangeHistoryPrefix, []byte(did)...)
	key = append(key, []byte("/")...)
	key = append(key, sdk.Uint64ToBigEndian(height)...)
	return key
}

// GetRecoveryRecordKey returns the store key for a recovery record
func GetRecoveryRecordKey(did string) []byte {
	return append(RecoveryRecordPrefix, []byte(did)...)
}

// GetVerificationKey returns the store key for verification
func GetVerificationKey(did string) []byte {
	return append(VerificationPrefix, []byte(did)...)
}

// GetDelegationKey returns the store key for delegation
func GetDelegationKey(did string) []byte {
	return append(DelegationPrefix, []byte(did)...)
}

// GetFederationKey returns the store key for federation
func GetFederationKey(did string) []byte {
	return append(FederationPrefix, []byte(did)...)
}

// GetCrossChainLinkKey returns the store key for cross-chain link
func GetCrossChainLinkKey(did string) []byte {
	return append(CrossChainLinkPrefix, []byte(did)...)
}
