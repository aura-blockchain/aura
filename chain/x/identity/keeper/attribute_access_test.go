package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/identity/types"
	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
)

// setupKeeper creates a keeper for testing
func setupKeeper(t *testing.T) (*Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	storeService := runtime.NewKVStoreService(storeKey)
	keeper := NewKeeper(storeService, cdc, "authority", log.NewNopLogger())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{Time: time.Now()}, false, log.NewNopLogger())

	return keeper, ctx
}

// ============================================================================
// GrantAttributeAccess Tests
// ============================================================================

func TestGrantAttributeAccess_Success(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	grantee := "cosmos1grantee"
	level := identitypb.AccessLevel_ACCESS_LEVEL_READ
	expiry := ctx.BlockTime().Add(24 * time.Hour)
	purpose := "Service verification"

	err := keeper.GrantAttributeAccess(ctx, owner, attribute, grantee, level, expiry, purpose)
	require.NoError(t, err)

	// Verify permission was stored
	permissions, err := keeper.GetAttributePermissions(ctx, owner, attribute)
	require.NoError(t, err)
	require.Len(t, permissions, 1)
	require.Equal(t, attribute, permissions[0].AttributeName)
	require.Equal(t, grantee, permissions[0].GrantedTo)
	require.Equal(t, level, permissions[0].AccessLevel)
	require.Equal(t, owner, permissions[0].GrantedBy)

	// Verify consent was recorded
	consent, err := keeper.GetAttributeConsent(ctx, owner, attribute, grantee)
	require.NoError(t, err)
	require.Equal(t, purpose, consent.Purpose)
	require.False(t, consent.Revoked)
}

func TestGrantAttributeAccess_PublicAccess(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "age"
	grantee := "*" // Public access
	level := identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY

	err := keeper.GrantAttributeAccess(ctx, owner, attribute, grantee, level, time.Time{}, "Public age verification")
	require.NoError(t, err)

	// Verify anyone can access with VERIFY_ONLY level
	accessLevel, err := keeper.CanAccessAttribute(ctx, owner, attribute, "cosmos1anyone")
	require.NoError(t, err)
	require.Equal(t, identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY, accessLevel)
}

func TestGrantAttributeAccess_InvalidInputs(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	testCases := []struct {
		name      string
		owner     string
		attribute string
		grantee   string
		level     identitypb.AccessLevel
		expectErr bool
	}{
		{
			name:      "empty owner",
			owner:     "",
			attribute: "email",
			grantee:   "cosmos1grantee",
			level:     identitypb.AccessLevel_ACCESS_LEVEL_READ,
			expectErr: true,
		},
		{
			name:      "empty attribute",
			owner:     "cosmos1owner",
			attribute: "",
			grantee:   "cosmos1grantee",
			level:     identitypb.AccessLevel_ACCESS_LEVEL_READ,
			expectErr: true,
		},
		{
			name:      "empty grantee",
			owner:     "cosmos1owner",
			attribute: "email",
			grantee:   "",
			level:     identitypb.AccessLevel_ACCESS_LEVEL_READ,
			expectErr: true,
		},
		{
			name:      "unspecified access level",
			owner:     "cosmos1owner",
			attribute: "email",
			grantee:   "cosmos1grantee",
			level:     identitypb.AccessLevel_ACCESS_LEVEL_UNSPECIFIED,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := keeper.GrantAttributeAccess(ctx, tc.owner, tc.attribute, tc.grantee, tc.level, time.Time{}, "Test")
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// RevokeAttributeAccess Tests
// ============================================================================

func TestRevokeAttributeAccess_Success(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	grantee := "cosmos1grantee"

	// Grant access first
	err := keeper.GrantAttributeAccess(ctx, owner, attribute, grantee, identitypb.AccessLevel_ACCESS_LEVEL_READ, time.Time{}, "Test")
	require.NoError(t, err)

	// Verify access granted
	level, err := keeper.CanAccessAttribute(ctx, owner, attribute, grantee)
	require.NoError(t, err)
	require.Equal(t, identitypb.AccessLevel_ACCESS_LEVEL_READ, level)

	// Revoke access
	reason := "User request"
	err = keeper.RevokeAttributeAccess(ctx, owner, attribute, grantee, reason)
	require.NoError(t, err)

	// Verify access denied
	_, err = keeper.CanAccessAttribute(ctx, owner, attribute, grantee)
	require.Error(t, err)

	// Verify consent marked as revoked
	consent, err := keeper.GetAttributeConsent(ctx, owner, attribute, grantee)
	require.NoError(t, err)
	require.True(t, consent.Revoked)
	require.Equal(t, reason, consent.RevocationReason)
	require.NotNil(t, consent.RevokedAt)
}

func TestRevokeAttributeAccess_NonExistent(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	grantee := "cosmos1grantee"

	// Revoke non-existent permission (should not error)
	err := keeper.RevokeAttributeAccess(ctx, owner, attribute, grantee, "Test")
	require.NoError(t, err)
}

// ============================================================================
// CanAccessAttribute Tests
// ============================================================================

func TestCanAccessAttribute_OwnerAccess(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"

	// Owner should always have READ access
	level, err := keeper.CanAccessAttribute(ctx, owner, attribute, owner)
	require.NoError(t, err)
	require.Equal(t, identitypb.AccessLevel_ACCESS_LEVEL_READ, level)
}

func TestCanAccessAttribute_SpecificGrant(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	grantee := "cosmos1grantee"
	level := identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY

	// Grant specific access
	err := keeper.GrantAttributeAccess(ctx, owner, attribute, grantee, level, time.Time{}, "Test")
	require.NoError(t, err)

	// Check access
	accessLevel, err := keeper.CanAccessAttribute(ctx, owner, attribute, grantee)
	require.NoError(t, err)
	require.Equal(t, level, accessLevel)
}

func TestCanAccessAttribute_PublicGrant(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "age"

	// Grant public access
	err := keeper.GrantAttributeAccess(ctx, owner, attribute, "*", identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY, time.Time{}, "Public")
	require.NoError(t, err)

	// Any user should have VERIFY_ONLY access
	accessLevel, err := keeper.CanAccessAttribute(ctx, owner, attribute, "cosmos1random")
	require.NoError(t, err)
	require.Equal(t, identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY, accessLevel)
}

func TestCanAccessAttribute_SpecificOverridesPublic(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	grantee := "cosmos1special"

	// Grant public VERIFY_ONLY access
	err := keeper.GrantAttributeAccess(ctx, owner, attribute, "*", identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY, time.Time{}, "Public")
	require.NoError(t, err)

	// Grant specific user READ access
	err = keeper.GrantAttributeAccess(ctx, owner, attribute, grantee, identitypb.AccessLevel_ACCESS_LEVEL_READ, time.Time{}, "Special")
	require.NoError(t, err)

	// Specific grant should take precedence
	accessLevel, err := keeper.CanAccessAttribute(ctx, owner, attribute, grantee)
	require.NoError(t, err)
	require.Equal(t, identitypb.AccessLevel_ACCESS_LEVEL_READ, accessLevel)

	// Others should still have VERIFY_ONLY
	accessLevel, err = keeper.CanAccessAttribute(ctx, owner, attribute, "cosmos1other")
	require.NoError(t, err)
	require.Equal(t, identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY, accessLevel)
}

func TestCanAccessAttribute_Expired(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	grantee := "cosmos1grantee"

	// Grant access with expiry in the past
	expiry := ctx.BlockTime().Add(-1 * time.Hour)
	err := keeper.GrantAttributeAccess(ctx, owner, attribute, grantee, identitypb.AccessLevel_ACCESS_LEVEL_READ, expiry, "Test")
	require.NoError(t, err)

	// Access should be denied due to expiry
	_, err = keeper.CanAccessAttribute(ctx, owner, attribute, grantee)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

func TestCanAccessAttribute_NotExpired(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	grantee := "cosmos1grantee"

	// Grant access with future expiry
	expiry := ctx.BlockTime().Add(24 * time.Hour)
	err := keeper.GrantAttributeAccess(ctx, owner, attribute, grantee, identitypb.AccessLevel_ACCESS_LEVEL_READ, expiry, "Test")
	require.NoError(t, err)

	// Access should be allowed
	accessLevel, err := keeper.CanAccessAttribute(ctx, owner, attribute, grantee)
	require.NoError(t, err)
	require.Equal(t, identitypb.AccessLevel_ACCESS_LEVEL_READ, accessLevel)
}

func TestCanAccessAttribute_NoAccess(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	requester := "cosmos1other"

	// No access granted
	_, err := keeper.CanAccessAttribute(ctx, owner, attribute, requester)
	require.Error(t, err)
	require.Contains(t, err.Error(), "access denied")
}

// ============================================================================
// GetAttributeWithAccessControl Tests
// ============================================================================

func TestGetAttributeWithAccessControl_Success(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	grantee := "cosmos1grantee"

	// Create identity record first
	record := &identitypb.IdentityRecord{
		Did:           owner,
		Address:       owner,
		Status:        identitypb.IdentityStatus_IDENTITY_STATUS_ACTIVE,
		CreatedAt:     timestamppb.New(ctx.BlockTime()),
		PiiCommitment: []byte("commitment"),
		MetadataHash:  "metadata_hash",
	}
	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Grant READ access
	err = keeper.GrantAttributeAccess(ctx, owner, attribute, grantee, identitypb.AccessLevel_ACCESS_LEVEL_READ, time.Time{}, "Test")
	require.NoError(t, err)

	// Get attribute with access control
	value, err := keeper.GetAttributeWithAccessControl(ctx, owner, attribute, grantee)
	require.NoError(t, err)
	require.NotNil(t, value)
}

func TestGetAttributeWithAccessControl_VerifyOnlyReturnsCommitment(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "age"
	grantee := "cosmos1verifier"

	// Create identity record
	commitment := []byte("age_commitment")
	record := &identitypb.IdentityRecord{
		Did:           owner,
		Address:       owner,
		Status:        identitypb.IdentityStatus_IDENTITY_STATUS_ACTIVE,
		CreatedAt:     timestamppb.New(ctx.BlockTime()),
		PiiCommitment: commitment,
		MetadataHash:  "metadata",
	}
	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Grant VERIFY_ONLY access
	err = keeper.GrantAttributeAccess(ctx, owner, attribute, grantee, identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY, time.Time{}, "Verify age")
	require.NoError(t, err)

	// Get attribute should return commitment only
	value, err := keeper.GetAttributeWithAccessControl(ctx, owner, attribute, grantee)
	require.NoError(t, err)
	require.Equal(t, commitment, value)
}

func TestGetAttributeWithAccessControl_AccessDenied(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	requester := "cosmos1unauthorized"

	// Create identity record
	record := &identitypb.IdentityRecord{
		Did:       owner,
		Address:   owner,
		Status:    identitypb.IdentityStatus_IDENTITY_STATUS_ACTIVE,
		CreatedAt: timestamppb.New(ctx.BlockTime()),
	}
	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Try to get attribute without permission
	_, err = keeper.GetAttributeWithAccessControl(ctx, owner, attribute, requester)
	require.Error(t, err)
	require.Contains(t, err.Error(), "access denied")
}

func TestGetAttributeWithAccessControl_IdentityNotFound(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	grantee := "cosmos1grantee"

	// Grant access but no identity record exists
	err := keeper.GrantAttributeAccess(ctx, owner, attribute, grantee, identitypb.AccessLevel_ACCESS_LEVEL_READ, time.Time{}, "Test")
	require.NoError(t, err)

	// Get attribute should fail
	_, err = keeper.GetAttributeWithAccessControl(ctx, owner, attribute, grantee)
	require.Error(t, err)
	require.Contains(t, err.Error(), "identity not found")
}

// ============================================================================
// Access Logging Tests
// ============================================================================

func TestAttributeAccessLogging_Success(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	grantee := "cosmos1grantee"

	// Create identity record
	record := &identitypb.IdentityRecord{
		Did:       owner,
		Address:   owner,
		Status:    identitypb.IdentityStatus_IDENTITY_STATUS_ACTIVE,
		CreatedAt: timestamppb.New(ctx.BlockTime()),
	}
	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Grant access and attempt to retrieve attribute
	err = keeper.GrantAttributeAccess(ctx, owner, attribute, grantee, identitypb.AccessLevel_ACCESS_LEVEL_READ, time.Time{}, "Test")
	require.NoError(t, err)

	_, err = keeper.GetAttributeWithAccessControl(ctx, owner, attribute, grantee)
	require.NoError(t, err)

	// Check access logs
	logs, total, err := keeper.GetAttributeAccessLogs(ctx, owner, 10, 0)
	require.NoError(t, err)
	require.Greater(t, total, uint64(0))
	require.NotEmpty(t, logs)

	// Verify log entry
	found := false
	for _, log := range logs {
		if log.AttributeName == attribute && log.Requester == grantee {
			require.True(t, log.Success)
			require.Equal(t, identitypb.AccessLevel_ACCESS_LEVEL_READ, log.AccessLevel)
			found = true
			break
		}
	}
	require.True(t, found, "Access log entry not found")
}

func TestAttributeAccessLogging_Failed(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	requester := "cosmos1unauthorized"

	// Create identity record
	record := &identitypb.IdentityRecord{
		Did:       owner,
		Address:   owner,
		Status:    identitypb.IdentityStatus_IDENTITY_STATUS_ACTIVE,
		CreatedAt: timestamppb.New(ctx.BlockTime()),
	}
	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Attempt unauthorized access
	_, err = keeper.GetAttributeWithAccessControl(ctx, owner, attribute, requester)
	require.Error(t, err)

	// Check access logs
	logs, total, err := keeper.GetAttributeAccessLogs(ctx, owner, 10, 0)
	require.NoError(t, err)
	require.Greater(t, total, uint64(0))

	// Verify failed log entry
	found := false
	for _, log := range logs {
		if log.AttributeName == attribute && log.Requester == requester {
			require.False(t, log.Success)
			require.NotEmpty(t, log.ErrorMessage)
			found = true
			break
		}
	}
	require.True(t, found, "Failed access log entry not found")
}

// ============================================================================
// Consent Tracking Tests
// ============================================================================

func TestConsentTracking_Grant(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	grantee := "cosmos1grantee"
	purpose := "Email verification for service"

	err := keeper.GrantAttributeAccess(ctx, owner, attribute, grantee, identitypb.AccessLevel_ACCESS_LEVEL_READ, time.Time{}, purpose)
	require.NoError(t, err)

	// Verify consent record
	consent, err := keeper.GetAttributeConsent(ctx, owner, attribute, grantee)
	require.NoError(t, err)
	require.Equal(t, owner, consent.Did)
	require.Equal(t, attribute, consent.AttributeName)
	require.Equal(t, grantee, consent.Grantee)
	require.Equal(t, purpose, consent.Purpose)
	require.False(t, consent.Revoked)
	require.Equal(t, identitypb.AccessLevel_ACCESS_LEVEL_READ, consent.AccessLevel)
}

func TestConsentTracking_Revoke(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	grantee := "cosmos1grantee"

	// Grant access
	err := keeper.GrantAttributeAccess(ctx, owner, attribute, grantee, identitypb.AccessLevel_ACCESS_LEVEL_READ, time.Time{}, "Test")
	require.NoError(t, err)

	// Revoke access
	reason := "User withdrew consent"
	err = keeper.RevokeAttributeAccess(ctx, owner, attribute, grantee, reason)
	require.NoError(t, err)

	// Verify consent marked as revoked
	consent, err := keeper.GetAttributeConsent(ctx, owner, attribute, grantee)
	require.NoError(t, err)
	require.True(t, consent.Revoked)
	require.Equal(t, reason, consent.RevocationReason)
	require.NotNil(t, consent.RevokedAt)
}

func TestGetAllAttributeConsents(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"

	// Grant multiple consents
	err := keeper.GrantAttributeAccess(ctx, owner, "email", "cosmos1grantee1", identitypb.AccessLevel_ACCESS_LEVEL_READ, time.Time{}, "Purpose 1")
	require.NoError(t, err)

	err = keeper.GrantAttributeAccess(ctx, owner, "age", "cosmos1grantee2", identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY, time.Time{}, "Purpose 2")
	require.NoError(t, err)

	err = keeper.GrantAttributeAccess(ctx, owner, "address", "*", identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY, time.Time{}, "Public")
	require.NoError(t, err)

	// Get all consents
	consents, err := keeper.GetAllAttributeConsents(ctx, owner)
	require.NoError(t, err)
	require.Len(t, consents, 3)

	// Verify all consents are for the correct owner
	for _, consent := range consents {
		require.Equal(t, owner, consent.Did)
		require.False(t, consent.Revoked)
	}
}

// ============================================================================
// Query Functions Tests
// ============================================================================

func TestGetAttributePermissions(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"

	// Grant permissions to multiple grantees
	err := keeper.GrantAttributeAccess(ctx, owner, attribute, "cosmos1grantee1", identitypb.AccessLevel_ACCESS_LEVEL_READ, time.Time{}, "Test")
	require.NoError(t, err)

	err = keeper.GrantAttributeAccess(ctx, owner, attribute, "cosmos1grantee2", identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY, time.Time{}, "Test")
	require.NoError(t, err)

	err = keeper.GrantAttributeAccess(ctx, owner, attribute, "*", identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY, time.Time{}, "Public")
	require.NoError(t, err)

	// Get all permissions for this attribute
	permissions, err := keeper.GetAttributePermissions(ctx, owner, attribute)
	require.NoError(t, err)
	require.Len(t, permissions, 3)

	// Verify all permissions are for the correct attribute
	for _, perm := range permissions {
		require.Equal(t, attribute, perm.AttributeName)
	}
}

func TestGetAllAttributePermissions(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := "cosmos1owner"

	// Grant permissions for multiple attributes
	err := keeper.GrantAttributeAccess(ctx, owner, "email", "cosmos1grantee1", identitypb.AccessLevel_ACCESS_LEVEL_READ, time.Time{}, "Test")
	require.NoError(t, err)

	err = keeper.GrantAttributeAccess(ctx, owner, "age", "cosmos1grantee2", identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY, time.Time{}, "Test")
	require.NoError(t, err)

	err = keeper.GrantAttributeAccess(ctx, owner, "address", "cosmos1grantee3", identitypb.AccessLevel_ACCESS_LEVEL_READ, time.Time{}, "Test")
	require.NoError(t, err)

	// Get all permissions for owner
	permissions, err := keeper.GetAllAttributePermissions(ctx, owner)
	require.NoError(t, err)
	require.Len(t, permissions, 3)

	// Verify all permissions are granted by the owner
	for _, perm := range permissions {
		require.Equal(t, owner, perm.GrantedBy)
	}
}
