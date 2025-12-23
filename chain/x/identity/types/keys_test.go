package types

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRoleKey(t *testing.T) {
	key := GetRoleKey("admin")
	require.True(t, bytes.HasPrefix(key, RolePrefix))
	require.Contains(t, string(key), "admin")
}

func TestGetRoleAssignmentKey(t *testing.T) {
	key := GetRoleAssignmentKey("aura1xxx")
	require.True(t, bytes.HasPrefix(key, RoleAssignmentPrefix))
	require.Contains(t, string(key), "aura1xxx")
}

func TestGetPermissionGrantKey(t *testing.T) {
	key := GetPermissionGrantKey("grantee", "permission")
	require.True(t, bytes.HasPrefix(key, PermissionGrantPrefix))
	require.Contains(t, string(key), "grantee:permission")
}

func TestGetAuditLogKey(t *testing.T) {
	key := GetAuditLogKey(123)
	require.True(t, bytes.HasPrefix(key, AuditLogPrefix))
	require.Equal(t, len(AuditLogPrefix)+8, len(key)) // 8 bytes for uint64
}

func TestGetAccountKey(t *testing.T) {
	key := GetAccountKey("aura1account")
	require.True(t, bytes.HasPrefix(key, AccountPrefix))
	require.Contains(t, string(key), "aura1account")
}

func TestGetSessionKey(t *testing.T) {
	key := GetSessionKey("session123")
	require.True(t, bytes.HasPrefix(key, SessionPrefix))
	require.Contains(t, string(key), "session123")
}

func TestGetUserSessionsKey(t *testing.T) {
	key := GetUserSessionsKey("aura1user")
	require.True(t, bytes.HasPrefix(key, UserSessionsPrefix))
	require.Contains(t, string(key), "aura1user")
}

func TestGetRateLimitConfigKey(t *testing.T) {
	key := GetRateLimitConfigKey("aura1user")
	require.True(t, bytes.HasPrefix(key, RateLimitConfigPrefix))
	require.Contains(t, string(key), "aura1user")
}

func TestGetMultisigWalletKey(t *testing.T) {
	key := GetMultisigWalletKey("wallet123")
	require.True(t, bytes.HasPrefix(key, MultisigWalletPrefix))
	require.Contains(t, string(key), "wallet123")
}

func TestGetMultisigProposalKey(t *testing.T) {
	key := GetMultisigProposalKey("proposal123")
	require.True(t, bytes.HasPrefix(key, MultisigProposalPrefix))
	require.Contains(t, string(key), "proposal123")
}

func TestGetTimeLockedActionKey(t *testing.T) {
	key := GetTimeLockedActionKey("action123")
	require.True(t, bytes.HasPrefix(key, TimeLockedActionPrefix))
	require.Contains(t, string(key), "action123")
}

func TestGetEmergencyAdminKey(t *testing.T) {
	key := GetEmergencyAdminKey("aura1admin")
	require.True(t, bytes.HasPrefix(key, EmergencyAdminPrefix))
	require.Contains(t, string(key), "aura1admin")
}

func TestGetEmergencyActionKey(t *testing.T) {
	key := GetEmergencyActionKey("emergencyAction1")
	require.True(t, bytes.HasPrefix(key, EmergencyActionPrefix))
	require.Contains(t, string(key), "emergencyAction1")
}

func TestGetValidatorRotationKey(t *testing.T) {
	key := GetValidatorRotationKey("auravaloper1xxx")
	require.True(t, bytes.HasPrefix(key, ValidatorRotationPrefix))
	require.Contains(t, string(key), "auravaloper1xxx")
}

func TestGetDIDKeyRotationKey(t *testing.T) {
	key := GetDIDKeyRotationKey("did:aura:123")
	require.True(t, bytes.HasPrefix(key, DIDKeyRotationPrefix))
	require.Contains(t, string(key), "did:aura:123")
}

func TestGetDIDKeyHistoryKey(t *testing.T) {
	key := GetDIDKeyHistoryKey("did:aura:123")
	require.True(t, bytes.HasPrefix(key, DIDKeyHistoryPrefix))
	require.Contains(t, string(key), "did:aura:123")
}

func TestGetIdentityRecordKey(t *testing.T) {
	key := GetIdentityRecordKey("did:aura:user1")
	require.True(t, bytes.HasPrefix(key, IdentityRecordPrefix))
	require.Contains(t, string(key), "did:aura:user1")
}

func TestGetChangeRequestKey(t *testing.T) {
	key := GetChangeRequestKey("request123")
	require.True(t, bytes.HasPrefix(key, ChangeRequestPrefix))
	require.Contains(t, string(key), "request123")
}

func TestGetChangeHistoryKey(t *testing.T) {
	key := GetChangeHistoryKey("did:aura:1", 1000)
	require.True(t, bytes.HasPrefix(key, ChangeHistoryPrefix))
	require.Contains(t, string(key), "did:aura:1")
	// The height is encoded as big-endian uint64
	require.Equal(t, len(ChangeHistoryPrefix)+len("did:aura:1")+1+8, len(key))
}

func TestGetRecoveryRecordKey(t *testing.T) {
	key := GetRecoveryRecordKey("did:aura:1")
	require.True(t, bytes.HasPrefix(key, RecoveryRecordPrefix))
	require.Contains(t, string(key), "did:aura:1")
}

func TestGetVerificationKey(t *testing.T) {
	key := GetVerificationKey("did:aura:1")
	require.True(t, bytes.HasPrefix(key, VerificationPrefix))
	require.Contains(t, string(key), "did:aura:1")
}

func TestGetDelegationKey(t *testing.T) {
	key := GetDelegationKey("did:aura:1")
	require.True(t, bytes.HasPrefix(key, DelegationPrefix))
	require.Contains(t, string(key), "did:aura:1")
}

func TestGetFederationKey(t *testing.T) {
	key := GetFederationKey("did:aura:1")
	require.True(t, bytes.HasPrefix(key, FederationPrefix))
	require.Contains(t, string(key), "did:aura:1")
}

func TestGetCrossChainLinkKey(t *testing.T) {
	key := GetCrossChainLinkKey("did:aura:1")
	require.True(t, bytes.HasPrefix(key, CrossChainLinkPrefix))
	require.Contains(t, string(key), "did:aura:1")
}

func TestGetCredentialRevocationKey(t *testing.T) {
	key := GetCredentialRevocationKey("cred123")
	require.True(t, bytes.HasPrefix(key, CredentialRevocationPrefix))
	require.Contains(t, string(key), "cred123")
}

func TestGetAttributePermissionKey(t *testing.T) {
	key := GetAttributePermissionKey("owner1", "attr1", "grantee1")
	require.True(t, bytes.HasPrefix(key, AttributePermissionPrefix))
	require.Contains(t, string(key), "owner1/attr1/grantee1")
}

func TestGetAttributePermissionsByOwnerPrefix(t *testing.T) {
	prefix := GetAttributePermissionsByOwnerPrefix("owner1")
	require.True(t, bytes.HasPrefix(prefix, AttributePermissionPrefix))
	require.Contains(t, string(prefix), "owner1/")
}

func TestGetAttributePermissionsByAttributePrefix(t *testing.T) {
	prefix := GetAttributePermissionsByAttributePrefix("owner1", "attr1")
	require.True(t, bytes.HasPrefix(prefix, AttributePermissionPrefix))
	require.Contains(t, string(prefix), "owner1/attr1/")
}

func TestGetAttributeAccessLogKey(t *testing.T) {
	key := GetAttributeAccessLogKey(456)
	require.True(t, bytes.HasPrefix(key, AttributeAccessLogPrefix))
	require.Equal(t, len(AttributeAccessLogPrefix)+8, len(key))
}

func TestGetAttributeAccessLogByOwnerPrefix(t *testing.T) {
	prefix := GetAttributeAccessLogByOwnerPrefix("owner1")
	require.True(t, bytes.HasPrefix(prefix, AttributeAccessLogPrefix))
	require.Contains(t, string(prefix), "owner:owner1/")
}

func TestGetAttributeConsentKey(t *testing.T) {
	key := GetAttributeConsentKey("did:aura:1", "email", "requester1")
	require.True(t, bytes.HasPrefix(key, AttributeConsentPrefix))
	require.Contains(t, string(key), "did:aura:1/email/requester1")
}

func TestGetAttributeConsentByDIDPrefix(t *testing.T) {
	prefix := GetAttributeConsentByDIDPrefix("did:aura:1")
	require.True(t, bytes.HasPrefix(prefix, AttributeConsentPrefix))
	require.Contains(t, string(prefix), "did:aura:1/")
}

// Test key uniqueness and no collisions
func TestKeyPrefixesUnique(t *testing.T) {
	prefixes := [][]byte{
		ParamsKey,
		RolePrefix,
		RoleAssignmentPrefix,
		PermissionGrantPrefix,
		AuditLogPrefix,
		AccountPrefix,
		SessionPrefix,
		UserSessionsPrefix,
		RateLimitConfigPrefix,
		MultisigWalletPrefix,
		MultisigProposalPrefix,
		TimeLockedActionPrefix,
		EmergencyAdminPrefix,
		EmergencyActionPrefix,
		ValidatorRotationPrefix,
		DIDKeyRotationPrefix,
		DIDKeyHistoryPrefix,
		IdentityRecordPrefix,
		ChangeRequestPrefix,
		ChangeHistoryPrefix,
		RecoveryRecordPrefix,
		VerificationPrefix,
		DelegationPrefix,
		FederationPrefix,
		CrossChainLinkPrefix,
		CredentialRevocationPrefix,
		AttributePermissionPrefix,
		AttributeAccessLogPrefix,
		AttributeConsentPrefix,
		AuditLogCounterPrefix,
		ChangeRequestCounterPrefix,
		AttributeAccessLogCounterPrefix,
		SuspendedKey,
	}

	seen := make(map[string]int)
	for i, prefix := range prefixes {
		key := string(prefix)
		if prevIdx, exists := seen[key]; exists {
			t.Errorf("Prefix collision: index %d and %d have same prefix %v", prevIdx, i, prefix)
		}
		seen[key] = i
	}
}

func TestModuleConstants(t *testing.T) {
	require.Equal(t, "identity", ModuleName)
	require.Equal(t, ModuleName, StoreKey)
	require.Equal(t, ModuleName, RouterKey)
	require.Equal(t, ModuleName, QuerierRoute)
}
