package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// TestInitGenesis_Nil tests InitGenesis with nil genesis state
func TestInitGenesis_Nil(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	err := keeper.InitGenesis(ctx, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "genesis state cannot be nil")
}

// TestInitGenesis_Empty tests InitGenesis with empty genesis state
func TestInitGenesis_Empty(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	gs := &types.GenesisState{
		Params: types.DefaultParams(),
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify default roles were created
	roles, err := keeper.GetAllRoles(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, roles, "default roles should be initialized")

	// Verify system roles exist
	systemRoleNames := []string{types.RoleAdmin, types.RoleModerator, types.RoleValidator, types.RoleUser}
	for _, roleName := range systemRoleNames {
		role, err := keeper.GetRole(ctx, roleName)
		require.NoError(t, err)
		require.NotNil(t, role)
		require.True(t, role.IsSystemRole, "system role should be marked as system")
	}
}

// TestInitGenesis_WithRoles tests InitGenesis with custom roles
func TestInitGenesis_WithRoles(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	customRoles := []*types.Role{
		{
			Name:         "custom_role_1",
			Permissions:  []string{types.PermissionAdmin},
			Description:  "Custom role 1",
			CreatedAt:    timestamppb.Now(),
			IsSystemRole: false,
			UpdatedAt:    timestamppb.Now(),
		},
		{
			Name:         "custom_role_2",
			Permissions:  []string{types.PermissionViewAuditLogs, types.PermissionManageIdentity},
			Description:  "Custom role 2",
			CreatedAt:    timestamppb.Now(),
			IsSystemRole: false,
			UpdatedAt:    timestamppb.Now(),
		},
	}

	gs := &types.GenesisState{
		Params: types.DefaultParams(),
		Roles:  customRoles,
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify custom roles were created
	for _, expectedRole := range customRoles {
		role, err := keeper.GetRole(ctx, expectedRole.Name)
		require.NoError(t, err)
		require.NotNil(t, role)
		require.Equal(t, expectedRole.Name, role.Name)
		require.Equal(t, expectedRole.Description, role.Description)
		require.ElementsMatch(t, expectedRole.Permissions, role.Permissions)
	}

	// When custom roles provided, default roles should NOT be created
	allRoles, err := keeper.GetAllRoles(ctx)
	require.NoError(t, err)
	require.Len(t, allRoles, len(customRoles), "only custom roles should exist")
}

// TestInitGenesis_WithRoleAssignments tests InitGenesis with role assignments
func TestInitGenesis_WithRoleAssignments(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	roleAssignments := []*types.RoleAssignment{
		{
			Address:    "aura1user1",
			RoleName:   types.RoleAdmin,
			AssignedAt: timestamppb.Now(),
			AssignedBy: "genesis",
		},
		{
			Address:    "aura1user2",
			RoleName:   types.RoleModerator,
			AssignedAt: timestamppb.Now(),
			AssignedBy: "genesis",
		},
	}

	gs := &types.GenesisState{
		Params:          types.DefaultParams(),
		RoleAssignments: roleAssignments,
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify role assignments
	assignments, err := keeper.GetAllRoleAssignments(ctx)
	require.NoError(t, err)
	require.Len(t, assignments, len(roleAssignments))

	// Create map for easy verification
	assignmentMap := make(map[string]*types.RoleAssignment)
	for _, a := range assignments {
		assignmentMap[a.Address+"_"+a.RoleName] = a
	}

	for _, expected := range roleAssignments {
		key := expected.Address + "_" + expected.RoleName
		assignment, ok := assignmentMap[key]
		require.True(t, ok, "role assignment should exist for %s", key)
		require.Equal(t, expected.Address, assignment.Address)
		require.Equal(t, expected.RoleName, assignment.RoleName)
	}
}

// TestInitGenesis_WithIdentityRecords tests InitGenesis with identity records
func TestInitGenesis_WithIdentityRecords(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	identityRecords := []*types.IdentityRecord{
		{
			Did:       "did:aura:test1",
			Address:   "aura1test1",
			Status:    types.IdentityStatusActive,
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
		},
		{
			Did:       "did:aura:test2",
			Address:   "aura1test2",
			Status:    types.IdentityStatusActive,
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
		},
	}

	gs := &types.GenesisState{
		Params:          types.DefaultParams(),
		IdentityRecords: identityRecords,
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify identity records
	for _, expected := range identityRecords {
		record, err := keeper.GetIdentityRecord(ctx, expected.Did)
		require.NoError(t, err)
		require.NotNil(t, record)
		require.Equal(t, expected.Did, record.Did)
		require.Equal(t, expected.Address, record.Address)
		require.Equal(t, expected.Status, record.Status)
	}
}

// TestInitGenesis_WithSessions tests InitGenesis with sessions
func TestInitGenesis_WithSessions(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	sessions := []*types.Session{
		{
			Id:        "session1",
			Address:   "aura1user1",
			CreatedAt: timestamppb.Now(),
			ExpiresAt: timestamppb.Now(),
			IsActive:  true,
		},
		{
			Id:        "session2",
			Address:   "aura1user2",
			CreatedAt: timestamppb.Now(),
			ExpiresAt: timestamppb.Now(),
			IsActive:  false,
		},
	}

	gs := &types.GenesisState{
		Params:   types.DefaultParams(),
		Sessions: sessions,
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify sessions
	for _, expected := range sessions {
		session, err := keeper.GetSession(ctx, expected.Id)
		require.NoError(t, err)
		require.NotNil(t, session)
		require.Equal(t, expected.Id, session.Id)
		require.Equal(t, expected.Address, session.Address)
		require.Equal(t, expected.IsActive, session.IsActive)
	}
}

// TestInitGenesis_WithMultisigWallets tests InitGenesis with multisig wallets
func TestInitGenesis_WithMultisigWallets(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	multisigWallets := []*types.MultisigWallet{
		{
			Id:        "wallet1",
			Signers:   []string{"aura1owner1", "aura1owner2"},
			Threshold: 2,
			CreatedAt: timestamppb.Now(),
			CreatedBy: "genesis",
		},
		{
			Id:        "wallet2",
			Signers:   []string{"aura1owner3", "aura1owner4", "aura1owner5"},
			Threshold: 2,
			CreatedAt: timestamppb.Now(),
			CreatedBy: "genesis",
		},
	}

	gs := &types.GenesisState{
		Params:          types.DefaultParams(),
		MultisigWallets: multisigWallets,
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify multisig wallets
	for _, expected := range multisigWallets {
		wallet, err := keeper.GetMultisigWallet(ctx, expected.Id)
		require.NoError(t, err)
		require.NotNil(t, wallet)
		require.Equal(t, expected.Id, wallet.Id)
		require.ElementsMatch(t, expected.Signers, wallet.Signers)
		require.Equal(t, expected.Threshold, wallet.Threshold)
	}
}

// TestInitGenesis_WithCredentialRevocations tests InitGenesis with credential revocations
func TestInitGenesis_WithCredentialRevocations(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	credentialRevocations := []*types.CredentialRevocation{
		{
			CredentialId: "cred1",
			RevokedAt:    timestamppb.Now(),
			RevokedBy:    "aura1issuer1",
			Reason:       "Compromised",
		},
		{
			CredentialId: "cred2",
			RevokedAt:    timestamppb.Now(),
			RevokedBy:    "aura1issuer2",
			Reason:       "Expired",
		},
	}

	gs := &types.GenesisState{
		Params:                types.DefaultParams(),
		CredentialRevocations: credentialRevocations,
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify credential revocations
	for _, expected := range credentialRevocations {
		revocation, err := keeper.GetCredentialRevocation(ctx, expected.CredentialId)
		require.NoError(t, err)
		require.NotNil(t, revocation)
		require.Equal(t, expected.CredentialId, revocation.CredentialId)
		require.Equal(t, expected.RevokedBy, revocation.RevokedBy)
		require.Equal(t, expected.Reason, revocation.Reason)
	}
}

// TestInitGenesis_WithChangeRequests tests InitGenesis with change requests
func TestInitGenesis_WithChangeRequests(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	changeRequests := []*types.ChangeRequest{
		{
			Id:          "change1",
			Did:         "did:aura:test1",
			Requester:   "aura1user1",
			ChangeType:  types.ChangeTypeUpdateAttributes,
			Status:      types.ChangeStatusPending,
			RequestedAt: timestamppb.Now(),
		},
		{
			Id:          "change2",
			Did:         "did:aura:test2",
			Requester:   "aura1user2",
			ChangeType:  types.ChangeTypeUpdateMetadata,
			Status:      types.ChangeStatusApproved,
			RequestedAt: timestamppb.Now(),
		},
	}

	gs := &types.GenesisState{
		Params:         types.DefaultParams(),
		ChangeRequests: changeRequests,
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify change requests
	for _, expected := range changeRequests {
		request, err := keeper.GetChangeRequest(ctx, expected.Id)
		require.NoError(t, err)
		require.NotNil(t, request)
		require.Equal(t, expected.Id, request.Id)
		require.Equal(t, expected.Requester, request.Requester)
		require.Equal(t, expected.ChangeType, request.ChangeType)
		require.Equal(t, expected.Status, request.Status)
	}
}

// TestInitGenesis_WithSuspendedFlag tests InitGenesis with suspended flag
func TestInitGenesis_WithSuspendedFlag(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	gs := &types.GenesisState{
		Params:                   types.DefaultParams(),
		IdentityChangesSuspended: true,
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify suspended flag was set
	store := keeper.storeService.OpenKVStore(ctx)
	suspendedBz, err := store.Get(types.SuspendedKey)
	require.NoError(t, err)
	require.NotNil(t, suspendedBz)
	require.Equal(t, byte(0x01), suspendedBz[0], "suspended flag should be set")
}

// TestInitGenesis_WithCounters tests InitGenesis with counters
func TestInitGenesis_WithCounters(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	gs := &types.GenesisState{
		Params:         types.DefaultParams(),
		NextAuditLogId: 42,
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Verify counter was set
	store := keeper.storeService.OpenKVStore(ctx)
	counterBz, err := store.Get(types.AuditLogCounterPrefix)
	require.NoError(t, err)
	require.NotNil(t, counterBz)
	// Note: SDK uses BigEndian encoding for uint64
	require.Len(t, counterBz, 8)
}

// TestExportGenesis_Empty tests ExportGenesis with empty state
func TestExportGenesis_Empty(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Initialize with minimal genesis
	err := keeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.DefaultParams(),
	})
	require.NoError(t, err)

	// Export
	gs, err := keeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NotNil(t, gs)
	require.NotNil(t, gs.Params)
	require.NotEmpty(t, gs.Roles, "default roles should be exported")
}

// TestExportGenesis_RoundTrip tests that init -> export -> init produces same state
func TestExportGenesis_RoundTrip(t *testing.T) {
	keeper1, ctx1 := setupKeeperForTest(t)

	// Create comprehensive genesis state
	customParams := types.DefaultParams()
	customParams.Auth.EnableRbac = true
	customParams.Auth.MaxRolesPerAccount = 5
	customParams.Change.MaxRequestsPerWalletPerMonth = 20

	originalGs := &types.GenesisState{
		Params: customParams,
		Roles: []*types.Role{
			{
				Name:         "custom_role",
				Permissions:  []string{types.PermissionAdmin},
				Description:  "Custom role",
				CreatedAt:    timestamppb.Now(),
				IsSystemRole: false,
				UpdatedAt:    timestamppb.Now(),
			},
		},
		RoleAssignments: []*types.RoleAssignment{
			{
				Address:    "aura1test",
				RoleName:   "custom_role",
				AssignedAt: timestamppb.Now(),
				AssignedBy: "genesis",
			},
		},
		IdentityRecords: []*types.IdentityRecord{
			{
				Did:       "did:aura:test123",
				Address:   "aura1test",
				Status:    types.IdentityStatusActive,
				CreatedAt: timestamppb.Now(),
				UpdatedAt: timestamppb.Now(),
			},
		},
		Sessions: []*types.Session{
			{
				Id:        "session123",
				Address:   "aura1test",
				CreatedAt: timestamppb.Now(),
				ExpiresAt: timestamppb.Now(),
				IsActive:  true,
			},
		},
		IdentityChangesSuspended: true,
		NextAuditLogId:           100,
	}

	// Initialize first keeper
	err := keeper1.InitGenesis(ctx1, originalGs)
	require.NoError(t, err)

	// Export from first keeper
	exportedGs, err := keeper1.ExportGenesis(ctx1)
	require.NoError(t, err)
	require.NotNil(t, exportedGs)

	// Create second keeper and initialize with exported state
	keeper2, ctx2 := setupKeeperForTest(t)
	err = keeper2.InitGenesis(ctx2, exportedGs)
	require.NoError(t, err)

	// Export from second keeper
	reExportedGs, err := keeper2.ExportGenesis(ctx2)
	require.NoError(t, err)
	require.NotNil(t, reExportedGs)

	// Verify key fields match
	require.Equal(t, exportedGs.Params.Auth.EnableRbac, reExportedGs.Params.Auth.EnableRbac)
	require.Equal(t, exportedGs.Params.Auth.MaxRolesPerAccount, reExportedGs.Params.Auth.MaxRolesPerAccount)
	require.Len(t, reExportedGs.Roles, len(exportedGs.Roles))
	require.Len(t, reExportedGs.RoleAssignments, len(exportedGs.RoleAssignments))
	require.Len(t, reExportedGs.IdentityRecords, len(exportedGs.IdentityRecords))
	require.Len(t, reExportedGs.Sessions, len(exportedGs.Sessions))
	require.Equal(t, exportedGs.IdentityChangesSuspended, reExportedGs.IdentityChangesSuspended)
}

// TestExportGenesis_AllDataTypes tests ExportGenesis with all data types
func TestExportGenesis_AllDataTypes(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Initialize with comprehensive data
	gs := &types.GenesisState{
		Params: types.DefaultParams(),
		AuditLogs: []*types.AuditLog{
			{
				Id:        "1",
				Timestamp: timestamppb.Now(),
				Action:    "CREATE_IDENTITY",
				Actor:     "aura1user",
				Target:    "did:aura:test",
				Details:   "Created identity",
				Result:    types.AuditResultSuccess,
			},
		},
		RateLimits: []*types.RateLimitConfig{
			{
				UserAddress:        "aura1user",
				RequestsPerMinute:  60,
				RequestsPerHour:    3600,
				RequestsPerDay:     86400,
				CurrentMinuteCount: 0,
				CurrentHourCount:   0,
				CurrentDayCount:    0,
			},
		},
		MultisigProposals: []*types.MultisigProposal{
			{
				Id:          "proposal1",
				WalletId:    "wallet1",
				Title:       "Test Proposal",
				Description: "Test proposal description",
				Payload:     []byte("test_payload"),
				Signatures:  []string{},
				Status:      types.ProposalStatusPending,
				CreatedAt:   timestamppb.Now(),
				ExpiresAt:   timestamppb.Now(),
			},
		},
		TimeLockedActions: []*types.TimeLockedAction{
			{
				Id:           "action1",
				ActionType:   "REVOKE_CREDENTIAL",
				Payload:      []byte("test_payload"),
				Proposer:     "aura1creator",
				ProposedAt:   timestamppb.Now(),
				ExecutableAt: timestamppb.Now(),
				Status:       types.ActionStatusPending,
				DelaySeconds: 3600,
			},
		},
		EmergencyAdmins: []*types.EmergencyAdmin{
			{
				Address:     "aura1admin",
				Privileges:  []string{types.PermissionAdmin},
				ActivatedAt: timestamppb.Now(),
			},
		},
		ValidatorRotations: []*types.ValidatorKeyRotation{
			{
				ValidatorAddress:   "auravaloper1val",
				OldConsensusPubkey: "oldkey",
				NewConsensusPubkey: "newkey",
				RotationTime:       timestamppb.Now(),
				InitiatedBy:        "aura1initiator",
			},
		},
		DidKeyRotations: []*types.DIDKeyRotation{
			{
				Did:                    "did:aura:test",
				OldVerificationMethod:  "oldkey",
				NewVerificationMethod:  "newkey",
				RotationTime:           timestamppb.Now(),
				InitiatedBy:            "aura1initiator",
			},
		},
		DidKeyHistories: []*types.DIDKeyHistory{
			{
				Did: "did:aura:test",
				Entries: []*types.DIDKeyHistoryEntry{
					{
						VerificationMethod: "key1",
						ActiveFrom:         timestamppb.Now(),
						ActiveUntil:        timestamppb.Now(),
					},
				},
			},
		},
		ChangeHistory: []*types.ChangeHistory{
			{
				RequestId:           "change1",
				TargetDid:           "did:aura:test",
				PrevConfidenceScore: 80,
				NewConfidenceScore:  85,
				TransitionReason:    "Identity verified",
				ChangedHeight:       100,
				ChangedAt:           timestamppb.Now(),
			},
		},
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Export and verify all data types are present
	exportedGs, err := keeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NotNil(t, exportedGs)

	require.Len(t, exportedGs.AuditLogs, 1)
	require.Len(t, exportedGs.RateLimits, 1)
	require.Len(t, exportedGs.MultisigProposals, 1)
	require.Len(t, exportedGs.TimeLockedActions, 1)
	require.Len(t, exportedGs.EmergencyAdmins, 1)
	require.Len(t, exportedGs.ValidatorRotations, 1)
	require.Len(t, exportedGs.DidKeyRotations, 1)
	require.Len(t, exportedGs.DidKeyHistories, 1)
	require.Len(t, exportedGs.ChangeHistory, 1)
}

// TestInitGenesis_InvalidRole tests InitGenesis with invalid role data
func TestInitGenesis_InvalidRole(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Role with empty name (should be validated by SetRole)
	gs := &types.GenesisState{
		Params: types.DefaultParams(),
		Roles: []*types.Role{
			{
				Name:         "", // Invalid: empty name
				Permissions:  []string{types.PermissionAdmin},
				Description:  "Invalid role",
				CreatedAt:    timestamppb.Now(),
				IsSystemRole: false,
				UpdatedAt:    timestamppb.Now(),
			},
		},
	}

	err := keeper.InitGenesis(ctx, gs)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to set role")
}

// TestExportGenesis_PreservesOrder tests that export maintains data consistency
func TestExportGenesis_PreservesOrder(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create multiple identity records
	identities := []*types.IdentityRecord{
		{
			Did:       "did:aura:aaa",
			Address:   "aura1aaa",
			Status:    types.IdentityStatusActive,
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
		},
		{
			Did:       "did:aura:bbb",
			Address:   "aura1bbb",
			Status:    types.IdentityStatusActive,
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
		},
		{
			Did:       "did:aura:ccc",
			Address:   "aura1ccc",
			Status:    types.IdentityStatusActive,
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
		},
	}

	gs := &types.GenesisState{
		Params:          types.DefaultParams(),
		IdentityRecords: identities,
	}

	err := keeper.InitGenesis(ctx, gs)
	require.NoError(t, err)

	// Export
	exportedGs, err := keeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Len(t, exportedGs.IdentityRecords, len(identities))

	// Verify all identities are present (order may vary due to iteration)
	exportedDIDs := make(map[string]bool)
	for _, record := range exportedGs.IdentityRecords {
		exportedDIDs[record.Did] = true
	}

	for _, identity := range identities {
		require.True(t, exportedDIDs[identity.Did], "exported genesis should contain DID: %s", identity.Did)
	}
}
