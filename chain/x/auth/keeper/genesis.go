package keeper

import (
	"context"

	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the auth module state from genesis
func (k *Keeper) InitGenesis(ctx context.Context, data *authproto.GenesisState) error {
	if data == nil {
		return nil
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Set parameters
	if data.Params != nil {
		if err := k.SetParams(sdkCtx, data.Params); err != nil {
			return err
		}
	}

	// Initialize roles
	for _, role := range data.Roles {
		// CreateRole expects (ctx, creator, name, permissions, description)
		if _, err := k.CreateRole(sdkCtx, "genesis", role.Name, role.Permissions, role.Description); err != nil {
			// Log but don't fail on individual role creation errors
			continue
		}
	}

	// TODO: Initialize other genesis data once keeper methods are fully converted to KVStore
	// For now, just initialize params and roles to allow chain to start

	return nil
}

// ExportGenesis exports the auth module state to genesis
func (k *Keeper) ExportGenesis(ctx context.Context) *authproto.GenesisState {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params, _ := k.GetParams(sdkCtx)

	return &authproto.GenesisState{
		Params:                params,
		Roles:                 []*authproto.Role{},
		RoleAssignments:       []*authproto.RoleAssignment{},
		MultisigWallets:       []*authproto.MultisigWallet{},
		MultisigProposals:     []*authproto.MultisigProposal{},
		TimeLockedActions:     []*authproto.TimeLockedAction{},
		EmergencyAdmins:       []*authproto.EmergencyAdmin{},
		ValidatorKeyRotations: []*authproto.ValidatorKeyRotation{},
		Sessions:              []*authproto.Session{},
		RateLimitConfigs:      []*authproto.RateLimitConfig{},
		AuditLogs:             []*authproto.AuditLog{},
	}
}
