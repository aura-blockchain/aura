package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

type InvariantsTestSuite struct {
	KeeperTestSuite
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

func (suite *InvariantsTestSuite) TestRoleConsistencyInvariant() {
	ctx := suite.SdkCtx

	// Test: Empty store should not break invariant
	inv := RoleConsistencyInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "empty store should not break invariant")
	suite.Empty(msg)

	// Create valid role
	now := time.Now()
	role := &authproto.Role{
		Name:        "test-role",
		Permissions: []string{types.PermissionAdmin, types.PermissionCreateRole},
		Description: "Test role",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Store role
	suite.storeRole(ctx, role)

	// Test: Valid role should not break invariant
	msg, broken = inv(ctx)
	suite.False(broken, "valid role should not break invariant")
	suite.Empty(msg)

	// Test: Role with invalid permission
	invalidRole := &authproto.Role{
		Name:        "invalid-role",
		Permissions: []string{"invalid-permission"},
		Description: "Invalid role",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	suite.storeRole(ctx, invalidRole)

	msg, broken = inv(ctx)
	suite.True(broken, "role with invalid permission should break invariant")
	suite.Contains(msg, "invalid permission")

	// Clean up
	suite.deleteRole(ctx, "invalid-role")

	// Test: Role with zero created_at
	nilCreatedRole := &authproto.Role{
		Name:        "nil-created",
		Permissions: []string{types.PermissionAdmin},
		Description: "Role with zero created_at",
		CreatedAt:   time.Time{},
		UpdatedAt:   now,
	}
	suite.storeRole(ctx, nilCreatedRole)

	msg, broken = inv(ctx)
	suite.True(broken, "role with nil created_at should break invariant")
	suite.Contains(msg, "nil created_at")

	// Clean up
	suite.deleteRole(ctx, "nil-created")

	// Test: Role with zero updated_at
	nilUpdatedRole := &authproto.Role{
		Name:        "nil-updated",
		Permissions: []string{types.PermissionAdmin},
		Description: "Role with zero updated_at",
		CreatedAt:   now,
		UpdatedAt:   time.Time{},
	}
	suite.storeRole(ctx, nilUpdatedRole)

	msg, broken = inv(ctx)
	suite.True(broken, "role with nil updated_at should break invariant")
	suite.Contains(msg, "nil updated_at")
}

func (suite *InvariantsTestSuite) TestRoleAssignmentConsistencyInvariant() {
	ctx := suite.SdkCtx

	// Create a valid role first
	now := time.Now()
	role := &authproto.Role{
		Name:        "test-role",
		Permissions: []string{types.PermissionAdmin},
		Description: "Test role",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	suite.storeRole(ctx, role)

	inv := RoleAssignmentConsistencyInvariant(suite.Keeper)

	// Test: Empty assignments should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Create valid assignment
	addr := sdk.AccAddress("test_address______")
	assignment := &authproto.RoleAssignment{
		Address:    addr.String(),
		RoleName:   "test-role",
		AssignedAt: now,
	}
	suite.storeRoleAssignment(ctx, assignment)

	// Test: Valid assignment should not break invariant
	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Test: Assignment with invalid address
	invalidAddrAssignment := &authproto.RoleAssignment{
		Address:    "invalid-address",
		RoleName:   "test-role",
		AssignedAt: now,
	}
	suite.storeRoleAssignment(ctx, invalidAddrAssignment)

	msg, broken = inv(ctx)
	suite.True(broken, "assignment with invalid address should break invariant")
	suite.Contains(msg, "invalid address")

	// Clean up
	suite.deleteRoleAssignment(ctx, invalidAddrAssignment)

	// Test: Assignment referencing non-existent role
	nonExistentRoleAssignment := &authproto.RoleAssignment{
		Address:    addr.String(),
		RoleName:   "non-existent-role",
		AssignedAt: now,
	}
	suite.storeRoleAssignment(ctx, nonExistentRoleAssignment)

	msg, broken = inv(ctx)
	suite.True(broken, "assignment with non-existent role should break invariant")
	suite.Contains(msg, "non-existent role")

	// Clean up
	suite.deleteRoleAssignment(ctx, nonExistentRoleAssignment)

	// Test: Assignment with zero timestamp
	nilTimestampAssignment := &authproto.RoleAssignment{
		Address:    addr.String(),
		RoleName:   "test-role",
		AssignedAt: time.Time{},
	}
	suite.storeRoleAssignment(ctx, nilTimestampAssignment)

	msg, broken = inv(ctx)
	suite.True(broken, "assignment with nil timestamp should break invariant")
	suite.Contains(msg, "nil assigned_at")
}

func (suite *InvariantsTestSuite) TestMultisigQuorumInvariant() {
	ctx := suite.SdkCtx
	inv := MultisigQuorumInvariant(suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Create valid multisig wallet
	addr1 := sdk.AccAddress("test_address1_____")
	addr2 := sdk.AccAddress("test_address2_____")
	now := time.Now()
	wallet := &authproto.MultisigWallet{
		Id:        sdk.AccAddress("wallet_address____").String(),
		Signers:   []string{addr1.String(), addr2.String()},
		Threshold: 2,
		CreatedAt: now,
		CreatedBy: addr1.String(),
	}
	suite.storeMultisigWallet(ctx, wallet)

	// Test: Valid multisig wallet should not break invariant
	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Test: Wallet with threshold exceeding signers
	invalidThresholdWallet := &authproto.MultisigWallet{
		Id:        sdk.AccAddress("invalid_wallet___").String(),
		Signers:   []string{addr1.String()},
		Threshold: 3,
		CreatedAt: now,
		CreatedBy: addr1.String(),
	}
	suite.storeMultisigWallet(ctx, invalidThresholdWallet)

	msg, broken = inv(ctx)
	suite.True(broken, "wallet with threshold > signers should break invariant")
	suite.Contains(msg, "quorum")
	suite.Contains(msg, "exceeds signers count")

	// Clean up
	suite.deleteMultisigWallet(ctx, invalidThresholdWallet.Id)

	// Test: Wallet with zero threshold
	zeroThresholdWallet := &authproto.MultisigWallet{
		Id:        sdk.AccAddress("zero_quorum______").String(),
		Signers:   []string{addr1.String(), addr2.String()},
		Threshold: 0,
		CreatedAt: now,
		CreatedBy: addr1.String(),
	}
	suite.storeMultisigWallet(ctx, zeroThresholdWallet)

	msg, broken = inv(ctx)
	suite.True(broken, "wallet with zero threshold should break invariant")
	suite.Contains(msg, "zero threshold")

	// Clean up
	suite.deleteMultisigWallet(ctx, zeroThresholdWallet.Id)

	// Test: Wallet with invalid signer address
	invalidSignerWallet := &authproto.MultisigWallet{
		Id:        sdk.AccAddress("invalid_signer___").String(),
		Signers:   []string{"invalid-address", addr1.String()},
		Threshold: 1,
		CreatedAt: now,
		CreatedBy: addr1.String(),
	}
	suite.storeMultisigWallet(ctx, invalidSignerWallet)

	msg, broken = inv(ctx)
	suite.True(broken, "wallet with invalid signer should break invariant")
	suite.Contains(msg, "invalid signer address")
}

func (suite *InvariantsTestSuite) TestTimeLockInvariant() {
	ctx := suite.SdkCtx
	inv := TimeLockInvariant(suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Create valid time-locked action
	now := time.Now()
	later := now.Add(24 * time.Hour)
	proposer := sdk.AccAddress("proposer_address__")

	action := &authproto.TimeLockedAction{
		Id:           "action-1",
		ActionType:   "test-action",
		Payload:      []byte("test data"),
		Proposer:     proposer.String(),
		ProposedAt:   now,
		ExecutableAt: later,
		Status:       authproto.ActionStatus_ACTION_STATUS_PENDING,
		DelaySeconds: 86400, // 24 hours
	}
	suite.storeTimeLockedAction(ctx, action)

	// Test: Valid action should not break invariant
	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Test: Action with execution time before creation time
	invalidTimeAction := &authproto.TimeLockedAction{
		Id:           "action-2",
		ActionType:   "invalid-time-action",
		Payload:      []byte("test data"),
		Proposer:     proposer.String(),
		ProposedAt:   later,
		ExecutableAt: now,
		Status:       authproto.ActionStatus_ACTION_STATUS_PENDING,
		DelaySeconds: 86400,
	}
	suite.storeTimeLockedAction(ctx, invalidTimeAction)

	msg, broken = inv(ctx)
	suite.True(broken, "action with execution time before creation should break invariant")
	suite.Contains(msg, "executable time before proposed time")

	// Clean up
	suite.deleteTimeLockedAction(ctx, "action-2")

	// Test: Action with nil timestamps - not possible with non-nullable proto fields
	// Skip this test as the protobuf definition has (gogoproto.nullable) = false

	// Test: Action with invalid proposer address
	invalidProposerAction := &authproto.TimeLockedAction{
		Id:           "action-3",
		ActionType:   "invalid-proposer-action",
		Payload:      []byte("test data"),
		Proposer:     "invalid-address",
		ProposedAt:   now,
		ExecutableAt: later,
		Status:       authproto.ActionStatus_ACTION_STATUS_PENDING,
		DelaySeconds: 86400,
	}
	suite.storeTimeLockedAction(ctx, invalidProposerAction)

	msg, broken = inv(ctx)
	suite.True(broken, "action with invalid proposer should break invariant")
	suite.Contains(msg, "invalid proposer address")
}

func (suite *InvariantsTestSuite) TestSessionValidityInvariant() {
	ctx := suite.SdkCtx
	inv := SessionValidityInvariant(suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Create valid session
	now := time.Now()
	later := now.Add(1 * time.Hour)
	userAddr := sdk.AccAddress("user_address______")

	session := &authproto.Session{
		SessionId:    "session-1",
		UserAddress:  userAddr.String(),
		CreatedAt:    now,
		ExpiresAt:    later,
		LastAccessed: now,
		IsActive:     true,
	}
	suite.storeSession(ctx, session)

	// Test: Valid session should not break invariant
	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Test: Session with expiration before creation
	invalidExpirySession := &authproto.Session{
		SessionId:    "session-2",
		UserAddress:  userAddr.String(),
		CreatedAt:    later,
		ExpiresAt:    now,
		LastAccessed: later,
		IsActive:     true,
	}
	suite.storeSession(ctx, invalidExpirySession)

	msg, broken = inv(ctx)
	suite.True(broken, "session expiring before creation should break invariant")
	suite.Contains(msg, "expires before creation")

	// Clean up
	suite.deleteSession(ctx, "session-2")

	// Test: Session with invalid user address
	invalidAddrSession := &authproto.Session{
		SessionId:    "session-3",
		UserAddress:  "invalid-address",
		CreatedAt:    now,
		ExpiresAt:    later,
		LastAccessed: now,
		IsActive:     true,
	}
	suite.storeSession(ctx, invalidAddrSession)

	msg, broken = inv(ctx)
	suite.True(broken, "session with invalid address should break invariant")
	suite.Contains(msg, "invalid user address")
}

func (suite *InvariantsTestSuite) TestRateLimitInvariant() {
	ctx := suite.SdkCtx
	inv := RateLimitInvariant(suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Create valid rate limit config
	userAddr := sdk.AccAddress("user_address______")
	rateLimit := &authproto.RateLimitConfig{
		UserAddress:         userAddr.String(),
		RequestsPerMinute:   60,
		RequestsPerHour:     3600,
		RequestsPerDay:      86400,
		CurrentMinuteCount:  30,
		CurrentHourCount:    1800,
		CurrentDayCount:     43200,
		WindowStart:         time.Now(),
	}
	suite.storeRateLimit(ctx, rateLimit)

	// Test: Valid rate limit should not break invariant
	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Test: Rate limit with invalid address
	invalidAddrRate := &authproto.RateLimitConfig{
		UserAddress:        "invalid-address",
		RequestsPerMinute:  60,
		CurrentMinuteCount: 30,
		WindowStart:        time.Now(),
	}
	suite.storeRateLimit(ctx, invalidAddrRate)

	msg, broken = inv(ctx)
	suite.True(broken, "rate limit with invalid address should break invariant")
	suite.Contains(msg, "invalid user address")

	// Clean up
	suite.deleteRateLimit(ctx, "invalid-address")

	// Test: Rate limit with all zero limits
	userAddr2 := sdk.AccAddress("user_address2_____")
	zeroLimitsRate := &authproto.RateLimitConfig{
		UserAddress:        userAddr2.String(),
		RequestsPerMinute:  0,
		RequestsPerHour:    0,
		RequestsPerDay:     0,
		CurrentMinuteCount: 0,
		WindowStart:        time.Now(),
	}
	suite.storeRateLimit(ctx, zeroLimitsRate)

	msg, broken = inv(ctx)
	suite.True(broken, "rate limit with all zero limits should break invariant")
	suite.Contains(msg, "all zero limits")

	// Clean up
	suite.deleteRateLimit(ctx, userAddr2.String())

	// Test: Rate limit with current count exceeding limit
	userAddr3 := sdk.AccAddress("user_address3_____")
	exceededRate := &authproto.RateLimitConfig{
		UserAddress:        userAddr3.String(),
		RequestsPerMinute:  60,
		CurrentMinuteCount: 100, // Exceeds limit
		WindowStart:        time.Now(),
	}
	suite.storeRateLimit(ctx, exceededRate)

	msg, broken = inv(ctx)
	suite.True(broken, "rate limit with count > limit should break invariant")
	suite.Contains(msg, "exceeds limit")
}

func (suite *InvariantsTestSuite) TestAuditLogIntegrityInvariant() {
	ctx := suite.SdkCtx
	inv := AuditLogIntegrityInvariant(suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Create valid audit log
	actor := sdk.AccAddress("actor_address_____")
	now := time.Now()

	log1 := &authproto.AuditLog{
		Id:        "log-1",
		Actor:     actor.String(),
		Action:    "create-role",
		Resource:  "role:test-role",
		Timestamp: now,
		Result:    "success",
	}
	suite.storeAuditLog(ctx, log1)

	// Test: Valid audit log should not break invariant
	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Test: Audit log with invalid actor address
	invalidActorLog := &authproto.AuditLog{
		Id:        "log-2",
		Actor:     "invalid-address",
		Action:    "test-action",
		Resource:  "test-resource",
		Timestamp: now.Add(1 * time.Second),
		Result:    "failure",
	}
	suite.storeAuditLog(ctx, invalidActorLog)

	msg, broken = inv(ctx)
	suite.True(broken, "audit log with invalid actor should break invariant")
	suite.Contains(msg, "invalid actor address")
}

func (suite *InvariantsTestSuite) TestAllInvariants() {
	ctx := suite.SdkCtx

	// Test: All invariants on empty store
	inv := AllInvariants(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "all invariants should pass on empty store")
	suite.Empty(msg)

	// Create valid data
	now := time.Now()
	role := &authproto.Role{
		Name:        "admin",
		Permissions: []string{types.PermissionAdmin},
		Description: "Admin role",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	suite.storeRole(ctx, role)

	// Test: All invariants with valid data
	msg, broken = inv(ctx)
	suite.False(broken, "all invariants should pass with valid data")
	suite.Empty(msg)
}

// Helper functions for storing test data
func (suite *InvariantsTestSuite) storeRole(ctx sdk.Context, role *authproto.Role) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	bz := suite.Keeper.cdc.MustMarshal(role)
	store.Set(append(RolesKeyPrefix, []byte(role.Name)...), bz)
}

func (suite *InvariantsTestSuite) deleteRole(ctx sdk.Context, name string) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	store.Delete(append(RolesKeyPrefix, []byte(name)...))
}

func (suite *InvariantsTestSuite) storeRoleAssignment(ctx sdk.Context, assignment *authproto.RoleAssignment) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	key := append(RoleAssignmentsKeyPrefix, []byte(assignment.Address+":"+assignment.RoleName)...)
	bz := suite.Keeper.cdc.MustMarshal(assignment)
	store.Set(key, bz)
}

func (suite *InvariantsTestSuite) deleteRoleAssignment(ctx sdk.Context, assignment *authproto.RoleAssignment) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	key := append(RoleAssignmentsKeyPrefix, []byte(assignment.Address+":"+assignment.RoleName)...)
	store.Delete(key)
}

func (suite *InvariantsTestSuite) storeMultisigWallet(ctx sdk.Context, wallet *authproto.MultisigWallet) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	bz := suite.Keeper.cdc.MustMarshal(wallet)
	store.Set(append(MultisigWalletsKeyPrefix, []byte(wallet.Id)...), bz)
}

func (suite *InvariantsTestSuite) deleteMultisigWallet(ctx sdk.Context, id string) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	store.Delete(append(MultisigWalletsKeyPrefix, []byte(id)...))
}

func (suite *InvariantsTestSuite) storeTimeLockedAction(ctx sdk.Context, action *authproto.TimeLockedAction) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	bz := suite.Keeper.cdc.MustMarshal(action)
	key := append(TimeLockedActionsKeyPrefix, []byte(action.Id)...)
	store.Set(key, bz)
}

func (suite *InvariantsTestSuite) deleteTimeLockedAction(ctx sdk.Context, actionID string) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	key := append(TimeLockedActionsKeyPrefix, []byte(actionID)...)
	store.Delete(key)
}

func (suite *InvariantsTestSuite) storeSession(ctx sdk.Context, session *authproto.Session) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	bz := suite.Keeper.cdc.MustMarshal(session)
	store.Set(append(SessionsKeyPrefix, []byte(session.SessionId)...), bz)
}

func (suite *InvariantsTestSuite) deleteSession(ctx sdk.Context, sessionID string) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	store.Delete(append(SessionsKeyPrefix, []byte(sessionID)...))
}

func (suite *InvariantsTestSuite) storeRateLimit(ctx sdk.Context, rateLimit *authproto.RateLimitConfig) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	bz := suite.Keeper.cdc.MustMarshal(rateLimit)
	store.Set(append(RateLimitsKeyPrefix, []byte(rateLimit.UserAddress)...), bz)
}

func (suite *InvariantsTestSuite) deleteRateLimit(ctx sdk.Context, key string) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	store.Delete(append(RateLimitsKeyPrefix, []byte(key)...))
}

func (suite *InvariantsTestSuite) storeAuditLog(ctx sdk.Context, log *authproto.AuditLog) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	bz := suite.Keeper.cdc.MustMarshal(log)
	key := append(AuditLogsKeyPrefix, []byte(log.Id)...)
	store.Set(key, bz)
}
