// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

const (
	// ModuleName defines the module name
	ModuleName = "auth"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

// Store key prefixes
var (
	ParamsKey              = []byte{0x00}
	EmergencyAdminPrefix   = []byte{0x01}
	EmergencyActionPrefix  = []byte{0x02}
	PermissionGrantPrefix  = []byte{0x03}
	RolePrefix             = []byte{0x04}
	RoleAssignmentPrefix   = []byte{0x05}
	MultisigWalletPrefix   = []byte{0x06}
	MultisigProposalPrefix = []byte{0x07}
	TimeLockedActionPrefix = []byte{0x08}
	SessionPrefix          = []byte{0x09}
	RateLimitConfigPrefix  = []byte{0x0a}
	AuditLogPrefix         = []byte{0x0b}
)

// GetEmergencyAdminKey returns the store key for an emergency admin
func GetEmergencyAdminKey(address string) []byte {
	return append(EmergencyAdminPrefix, []byte(address)...)
}

// GetEmergencyActionKey returns the store key for an emergency action
func GetEmergencyActionKey(actionID string) []byte {
	return append(EmergencyActionPrefix, []byte(actionID)...)
}

// GetPermissionGrantKey returns the store key for a permission grant
func GetPermissionGrantKey(grantee, permission string) []byte {
	key := append(PermissionGrantPrefix, []byte(grantee)...)
	key = append(key, []byte(":")...)
	key = append(key, []byte(permission)...)
	return key
}

// GetRoleKey returns the store key for a role
func GetRoleKey(roleID string) []byte {
	return append(RolePrefix, []byte(roleID)...)
}

// GetRoleAssignmentKey returns the store key for a role assignment
func GetRoleAssignmentKey(address, roleID string) []byte {
	key := append(RoleAssignmentPrefix, []byte(address)...)
	key = append(key, []byte(":")...)
	key = append(key, []byte(roleID)...)
	return key
}

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

// GetSessionKey returns the store key for a session
func GetSessionKey(sessionID string) []byte {
	return append(SessionPrefix, []byte(sessionID)...)
}

// GetRateLimitConfigKey returns the store key for a rate limit config
func GetRateLimitConfigKey(identifier string) []byte {
	return append(RateLimitConfigPrefix, []byte(identifier)...)
}

// GetAuditLogKey returns the store key for an audit log entry
func GetAuditLogKey(logID string) []byte {
	return append(AuditLogPrefix, []byte(logID)...)
}
