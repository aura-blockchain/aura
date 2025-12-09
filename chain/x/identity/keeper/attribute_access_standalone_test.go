package keeper

// Standalone test file for attribute access control
// This file demonstrates that the attribute access control implementation compiles and works correctly

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

// setupTestKeeper creates a minimal keeper for testing attribute access
func setupTestKeeper(t *testing.T) (*Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	storeService := runtime.NewKVStoreService(storeKey)
	keeper := NewKeeper(storeService, storeKey, cdc, "authority", log.NewNopLogger())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{Time: time.Now()}, false, log.NewNopLogger())

	return keeper, ctx
}

func TestAttributeAccessControlBasic(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	grantee := "cosmos1grantee"

	// Test granting access
	err := keeper.GrantAttributeAccess(
		ctx,
		owner,
		attribute,
		grantee,
		identitypb.AccessLevel_ACCESS_LEVEL_READ,
		time.Time{}, // No expiry
		"Email verification",
	)
	require.NoError(t, err, "Failed to grant attribute access")

	// Test checking access
	level, err := keeper.CanAccessAttribute(ctx, owner, attribute, grantee)
	require.NoError(t, err, "Failed to check access permission")
	require.Equal(t, identitypb.AccessLevel_ACCESS_LEVEL_READ, level, "Incorrect access level")

	// Test revoking access
	err = keeper.RevokeAttributeAccess(ctx, owner, attribute, grantee, "User request")
	require.NoError(t, err, "Failed to revoke access")

	// Verify access denied after revocation
	_, err = keeper.CanAccessAttribute(ctx, owner, attribute, grantee)
	require.Error(t, err, "Access should be denied after revocation")
	require.Contains(t, err.Error(), "access denied", "Error should indicate access denied")
}

func TestAttributeAccessControlSelectiveDisclosure(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	owner := "cosmos1owner"
	attribute := "age"
	verifier := "cosmos1verifier"
	reader := "cosmos1reader"

	// Grant VERIFY_ONLY access to verifier
	err := keeper.GrantAttributeAccess(
		ctx,
		owner,
		attribute,
		verifier,
		identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY,
		time.Time{},
		"Age verification only",
	)
	require.NoError(t, err)

	// Grant READ access to reader
	err = keeper.GrantAttributeAccess(
		ctx,
		owner,
		attribute,
		reader,
		identitypb.AccessLevel_ACCESS_LEVEL_READ,
		time.Time{},
		"Full age data access",
	)
	require.NoError(t, err)

	// Verify different access levels
	verifierLevel, err := keeper.CanAccessAttribute(ctx, owner, attribute, verifier)
	require.NoError(t, err)
	require.Equal(t, identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY, verifierLevel)

	readerLevel, err := keeper.CanAccessAttribute(ctx, owner, attribute, reader)
	require.NoError(t, err)
	require.Equal(t, identitypb.AccessLevel_ACCESS_LEVEL_READ, readerLevel)
}

func TestAttributeAccessControlExpiry(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	owner := "cosmos1owner"
	attribute := "ssn"
	grantee := "cosmos1temp"

	// Grant access with expiry in the past
	pastExpiry := ctx.BlockTime().Add(-1 * time.Hour)
	err := keeper.GrantAttributeAccess(
		ctx,
		owner,
		attribute,
		grantee,
		identitypb.AccessLevel_ACCESS_LEVEL_READ,
		pastExpiry,
		"Temporary access",
	)
	require.NoError(t, err)

	// Access should be denied due to expiry
	_, err = keeper.CanAccessAttribute(ctx, owner, attribute, grantee)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

func TestAttributeAccessControlOwnerAlwaysHasAccess(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	owner := "cosmos1owner"
	attribute := "private_data"

	// Owner should always have READ access without explicit grant
	level, err := keeper.CanAccessAttribute(ctx, owner, attribute, owner)
	require.NoError(t, err)
	require.Equal(t, identitypb.AccessLevel_ACCESS_LEVEL_READ, level)
}

func TestAttributeAccessControlPublicAccess(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	owner := "cosmos1owner"
	attribute := "public_bio"

	// Grant public access
	err := keeper.GrantAttributeAccess(
		ctx,
		owner,
		attribute,
		"*", // Public wildcard
		identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY,
		time.Time{},
		"Public biography verification",
	)
	require.NoError(t, err)

	// Any address should have VERIFY_ONLY access
	randomAddr := "cosmos1random"
	level, err := keeper.CanAccessAttribute(ctx, owner, attribute, randomAddr)
	require.NoError(t, err)
	require.Equal(t, identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY, level)
}

func TestAttributeAccessControlConsentTracking(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	grantee := "cosmos1service"
	purpose := "Newsletter subscription"

	// Grant access
	err := keeper.GrantAttributeAccess(
		ctx,
		owner,
		attribute,
		grantee,
		identitypb.AccessLevel_ACCESS_LEVEL_READ,
		time.Time{},
		purpose,
	)
	require.NoError(t, err)

	// Verify consent was recorded
	consent, err := keeper.GetAttributeConsent(ctx, owner, attribute, grantee)
	require.NoError(t, err)
	require.Equal(t, owner, consent.Did)
	require.Equal(t, attribute, consent.AttributeName)
	require.Equal(t, grantee, consent.Grantee)
	require.Equal(t, purpose, consent.Purpose)
	require.False(t, consent.Revoked)

	// Revoke and check consent updated
	revocationReason := "Unsubscribed from newsletter"
	err = keeper.RevokeAttributeAccess(ctx, owner, attribute, grantee, revocationReason)
	require.NoError(t, err)

	consent, err = keeper.GetAttributeConsent(ctx, owner, attribute, grantee)
	require.NoError(t, err)
	require.True(t, consent.Revoked)
	require.Equal(t, revocationReason, consent.RevocationReason)
	require.NotNil(t, consent.RevokedAt)
}

func TestAttributeAccessControlWithIdentityRecord(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"
	grantee := "cosmos1verifier"

	// Create identity record
	commitment := []byte("email_commitment_hash")
	record := &identitypb.IdentityRecord{
		Did:           owner,
		Address:       owner,
		Status:        identitypb.IdentityStatus_IDENTITY_STATUS_ACTIVE,
		CreatedAt:     ctx.BlockTime(),
		PiiCommitment: commitment,
		MetadataHash:  "metadata_hash",
	}
	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Grant VERIFY_ONLY access
	err = keeper.GrantAttributeAccess(
		ctx,
		owner,
		attribute,
		grantee,
		identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY,
		time.Time{},
		"Email verification",
	)
	require.NoError(t, err)

	// Get attribute with access control should return commitment
	value, err := keeper.GetAttributeWithAccessControl(ctx, owner, attribute, grantee)
	require.NoError(t, err)
	require.Equal(t, commitment, value, "Should return commitment for VERIFY_ONLY access")
}

func TestAttributeAccessControlPermissionQueries(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	owner := "cosmos1owner"
	attribute := "email"

	// Grant multiple permissions
	err := keeper.GrantAttributeAccess(ctx, owner, attribute, "cosmos1grantee1", identitypb.AccessLevel_ACCESS_LEVEL_READ, time.Time{}, "Purpose 1")
	require.NoError(t, err)

	err = keeper.GrantAttributeAccess(ctx, owner, attribute, "cosmos1grantee2", identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY, time.Time{}, "Purpose 2")
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
		require.Equal(t, owner, perm.GrantedBy)
	}
}
