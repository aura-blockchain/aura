package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	sdkerrors "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// AccountMigration represents an account migration request
type AccountMigration struct {
	ID                 string
	OldAddress         string
	NewAddress         string
	Reason             string
	Status             MigrationStatus
	ApprovedBy         []string
	RequiredApprovals  uint32
	CreatedAt          time.Time
	CompletedAt        *time.Time
	TransferredBalance bool
	TransferredRoles   bool
	TransferredData    bool
	Metadata           map[string]string
}

// MigrationStatus defines migration status
type MigrationStatus string

const (
	MigrationStatusPending   MigrationStatus = "pending"
	MigrationStatusApproved  MigrationStatus = "approved"
	MigrationStatusCompleted MigrationStatus = "completed"
	MigrationStatusRejected  MigrationStatus = "rejected"
)

var (
	AccountMigrationKeyPrefix = []byte{0x20}
)

// InitiateAccountMigration starts account migration process
func (k *Keeper) InitiateAccountMigration(ctx sdk.Context, oldAddr, newAddr, initiator, reason string) (*AccountMigration, error) {
	// Validate addresses
	if _, err := sdk.AccAddressFromBech32(oldAddr); err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidAddress, "invalid old address")
	}
	if _, err := sdk.AccAddressFromBech32(newAddr); err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidAddress, "invalid new address")
	}

	// Check permission
	if err := k.RequirePermission(ctx, initiator, types.PermissionAdmin); err != nil {
		return nil, err
	}

	// Create migration request
	// Use ctx.BlockTime() for deterministic ID generation
	migration := &AccountMigration{
		ID:                types.GenerateID("migration", ctx.BlockTime(), oldAddr, newAddr),
		OldAddress:        oldAddr,
		NewAddress:        newAddr,
		Reason:            reason,
		Status:            MigrationStatusPending,
		ApprovedBy:        []string{initiator},
		RequiredApprovals: 3, // Require 3 admin approvals
		CreatedAt:         ctx.BlockTime(),
		Metadata:          make(map[string]string),
	}

	// Store migration
	k.setAccountMigration(ctx, migration)

	// Log audit (uses ctx.BlockTime() internally for determinism)
	k.LogAudit(ctx, initiator, "account_migration", "initiate", "success", map[string]string{
		"old_address": oldAddr,
		"new_address": newAddr,
		"migration_id": migration.ID,
	}, "")

	return migration, nil
}

// ApproveAccountMigration approves a migration request
func (k *Keeper) ApproveAccountMigration(ctx sdk.Context, migrationID, approver string) error {
	// Check permission
	if err := k.RequirePermission(ctx, approver, types.PermissionAdmin); err != nil {
		return fmt.Errorf("error in ApproveAccountMigration for migration_id: %w", err)
	}

	migration, err := k.getAccountMigration(ctx, migrationID)
	if err != nil {
		return fmt.Errorf("failed to getAccountMigration: %w", err)
	}

	if migration.Status != MigrationStatusPending {
		return fmt.Errorf("migration is not in pending status")
	}

	// Check if already approved by this address
	for _, addr := range migration.ApprovedBy {
		if addr == approver {
			return fmt.Errorf("already approved by this address")
		}
	}

	// Add approval
	migration.ApprovedBy = append(migration.ApprovedBy, approver)

	// Check if enough approvals
	if uint32(len(migration.ApprovedBy)) >= migration.RequiredApprovals {
		migration.Status = MigrationStatusApproved
	}

	k.setAccountMigration(ctx, migration)

	// Log audit (uses ctx.BlockTime() internally for determinism)
	k.LogAudit(ctx, approver, "account_migration", "approve", "success", map[string]string{
		"migration_id": migrationID,
		"approvals": fmt.Sprintf("%d/%d", len(migration.ApprovedBy), migration.RequiredApprovals),
	}, "")

	return nil
}

// ExecuteAccountMigration completes the migration
func (k *Keeper) ExecuteAccountMigration(ctx sdk.Context, migrationID, executor string) error {
	// Check permission
	if err := k.RequirePermission(ctx, executor, types.PermissionAdmin); err != nil {
		return fmt.Errorf("error in ExecuteAccountMigration: %w", err)
	}

	migration, err := k.getAccountMigration(ctx, migrationID)
	if err != nil {
		return fmt.Errorf("failed to getAccountMigration: %w", err)
	}

	if migration.Status != MigrationStatusApproved {
		return fmt.Errorf("migration must be approved before execution")
	}

	// Migrate roles
	if err := k.migrateRoles(ctx, migration); err != nil {
		return fmt.Errorf("failed to getAccountMigration: %w", err)
	}
	migration.TransferredRoles = true

	// Migrate sessions
	if err := k.migrateSessions(ctx, migration); err != nil {
		return fmt.Errorf("error in ExecuteAccountMigration: %w", err)
	}
	migration.TransferredData = true

	// Update status
	now := ctx.BlockTime()
	migration.Status = MigrationStatusCompleted
	migration.CompletedAt = &now

	k.setAccountMigration(ctx, migration)

	// Log audit (uses ctx.BlockTime() internally for determinism)
	k.LogAudit(ctx, executor, "account_migration", "execute", "success", map[string]string{
		"migration_id": migrationID,
		"old_address": migration.OldAddress,
		"new_address": migration.NewAddress,
	}, "")

	return nil
}

// migrateRoles transfers role assignments to new address
func (k *Keeper) migrateRoles(ctx sdk.Context, migration *AccountMigration) error {
	assignments, err := k.GetRoleAssignmentsForAddress(ctx, migration.OldAddress)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	for _, assignment := range assignments {
		// Use ctx.BlockTime() for determinism - NEVER use time.Now() in consensus code
		newAssignment := &authproto.RoleAssignment{
			Address:    migration.NewAddress,
			RoleName:   assignment.RoleName,
			AssignedBy: "migration:" + migration.ID,
			AssignedAt: ctx.BlockTime(),
			ExpiresAt:  assignment.ExpiresAt,
		}
		if err := k.SetRoleAssignment(ctx, newAssignment); err != nil {
			return fmt.Errorf("error in migrateRoles: %w", err)
		}
	}

	return nil
}

// migrateSessions transfers sessions to new address
func (k *Keeper) migrateSessions(ctx sdk.Context, migration *AccountMigration) error {
	sessionIDs, err := k.GetUserSessions(ctx, migration.OldAddress)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	for _, sessionID := range sessionIDs {
		session, err := k.GetSession(ctx, sessionID)
		if err != nil {
			continue
		}

		// Update session to new address
		session.UserAddress = migration.NewAddress
		if err := k.SetSession(ctx, session); err != nil {
			return fmt.Errorf("failed to get: %w", err)
		}
	}

	return nil
}

// setAccountMigration stores a migration
func (k *Keeper) setAccountMigration(ctx sdk.Context, migration *AccountMigration) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(migration)
	if err != nil {
		return
	}
	key := append(AccountMigrationKeyPrefix, []byte(migration.ID)...)
	store.Set(key, bz)
}

// getAccountMigration retrieves a migration
func (k *Keeper) getAccountMigration(ctx sdk.Context, migrationID string) (*AccountMigration, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(AccountMigrationKeyPrefix, []byte(migrationID)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("migration not found: %s", migrationID)
	}

	var migration AccountMigration
	if err := json.Unmarshal(bz, &migration); err != nil {
		return nil, err
	}

	return &migration, nil
}

// GetAccountMigrations returns all migrations
func (k *Keeper) GetAccountMigrations(ctx sdk.Context) ([]*AccountMigration, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, AccountMigrationKeyPrefix)
	defer iterator.Close()

	var migrations []*AccountMigration
	for ; iterator.Valid(); iterator.Next() {
		var migration AccountMigration
		if err := json.Unmarshal(iterator.Value(), &migration); err != nil {
			continue
		}
		migrations = append(migrations, &migration)
	}

	return migrations, nil
}
