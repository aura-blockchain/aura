package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// ============================================================================
// Session Management
// ============================================================================

// SetSession stores a session in the KVStore
func (k *Keeper) SetSession(ctx sdk.Context, session *types.Session) error {
	if session.Id == "" {
		return types.ErrInvalidSession.Wrap("session ID cannot be empty")
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := types.GetSessionKey(session.Id)
	if err := store.Set(key, bz); err != nil {
		return err
	}

	// Update user sessions index
	return k.addUserSession(ctx, session.Address, session.Id)
}

// GetSession retrieves a session from the KVStore
func (k *Keeper) GetSession(ctx sdk.Context, sessionID string) (types.Session, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetSessionKey(sessionID)
	bz, err := store.Get(key)
	if err != nil {
		return types.Session{}, err
	}
	if bz == nil {
		return types.Session{}, types.ErrSessionNotFound.Wrapf("session not found: %s", sessionID)
	}

	var session types.Session
	if err := json.Unmarshal(bz, &session); err != nil {
		return types.Session{}, fmt.Errorf("failed to unmarshal session: %w", err)
	}
	return session, nil
}

// GetAllSessions retrieves all sessions
func (k *Keeper) GetAllSessions(ctx sdk.Context) ([]*types.Session, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.SessionPrefix, storetypes.PrefixEndBytes(types.SessionPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var sessions []*types.Session
	for ; iterator.Valid(); iterator.Next() {
		var session types.Session
		if err := json.Unmarshal(iterator.Value(), &session); err != nil {
			return nil, fmt.Errorf("failed to unmarshal session: %w", err)
		}
		sessions = append(sessions, &session)
	}
	return sessions, nil
}

// DeleteSession removes a session
func (k *Keeper) DeleteSession(ctx sdk.Context, sessionID string) error {
	session, err := k.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	key := types.GetSessionKey(sessionID)
	if err := store.Delete(key); err != nil {
		return err
	}

	// Update user sessions index
	return k.removeUserSession(ctx, session.Address, sessionID)
}

// addUserSession adds a session ID to a user's session list
func (k *Keeper) addUserSession(ctx sdk.Context, userAddress, sessionID string) error {
	sessions, _ := k.GetUserSessions(ctx, userAddress)
	sessions = append(sessions, sessionID)

	store := k.storeService.OpenKVStore(ctx)
	bz, err := json.Marshal(types.SessionIDList{SessionIDs: sessions})
	if err != nil {
		return err
	}
	key := types.GetUserSessionsKey(userAddress)
	return store.Set(key, bz)
}

// removeUserSession removes a session ID from a user's session list
func (k *Keeper) removeUserSession(ctx sdk.Context, userAddress, sessionID string) error {
	sessions, _ := k.GetUserSessions(ctx, userAddress)
	var filtered []string
	for _, sid := range sessions {
		if sid != sessionID {
			filtered = append(filtered, sid)
		}
	}

	store := k.storeService.OpenKVStore(ctx)
	key := types.GetUserSessionsKey(userAddress)

	if len(filtered) == 0 {
		return store.Delete(key)
	}

	bz, err := json.Marshal(types.SessionIDList{SessionIDs: filtered})
	if err != nil {
		return err
	}
	return store.Set(key, bz)
}

// GetUserSessions retrieves all session IDs for a user
func (k *Keeper) GetUserSessions(ctx sdk.Context, userAddress string) ([]string, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetUserSessionsKey(userAddress)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return []string{}, nil
	}

	var sessionList types.SessionIDList
	if err := json.Unmarshal(bz, &sessionList); err != nil {
		return nil, err
	}
	return sessionList.SessionIDs, nil
}

// CreateSession creates a new session for a user
func (k *Keeper) CreateSession(ctx sdk.Context, userAddress string, expirySeconds uint64) (types.Session, error) {
	// Get metrics instance
	metrics := GetIdentityMetrics()

	params, _ := k.GetParams(ctx)

	// Check max sessions per user (use max roles per account as proxy for now)
	sessions, _ := k.GetUserSessions(ctx, userAddress)
	maxSessions := uint32(10) // default max sessions
	if params != nil && params.Auth.MaxRolesPerAccount > 0 {
		maxSessions = params.Auth.MaxRolesPerAccount
	}
	if uint32(len(sessions)) >= maxSessions {
		return types.Session{}, types.ErrInvalidSession.Wrap("maximum sessions per user exceeded")
	}

	// Generate session ID
	sessionID := fmt.Sprintf("sess-%s-%d", userAddress, ctx.BlockTime().Unix())

	now := ctx.BlockTime()
	expiresAt := now.Add(time.Duration(expirySeconds) * time.Second)

	session := &types.Session{
		Id:           sessionID,
		Address:      userAddress,
		CreatedAt:    now,
		ExpiresAt:    expiresAt,
		LastAccessed: &now,
		IsActive:     true,
	}

	if err := k.SetSession(ctx, session); err != nil {
		return types.Session{}, err
	}

	k.LogAudit(ctx, userAddress, "create_session", sessionID, "success", nil, "")

	// Record metrics
	metrics.SessionsCreated.Inc()
	// Update active sessions count
	allSessions, _ := k.GetAllSessions(ctx)
	activeCount := 0
	for _, s := range allSessions {
		if s.IsActive {
			activeCount++
		}
	}
	metrics.SessionsActive.Set(float64(activeCount))

	return *session, nil
}

// RevokeSession revokes an active session
func (k *Keeper) RevokeSession(ctx sdk.Context, userAddress, sessionID string) error {
	session, err := k.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.Address != userAddress {
		return types.ErrPermissionDenied.Wrap("session does not belong to user")
	}

	if err := k.DeleteSession(ctx, sessionID); err != nil {
		return err
	}

	k.LogAudit(ctx, userAddress, "revoke_session", sessionID, "success", nil, "")

	// Record metrics
	metrics := GetIdentityMetrics()
	metrics.SessionsTerminated.WithLabelValues("revoked").Inc()
	// Update active sessions count
	allSessions, _ := k.GetAllSessions(ctx)
	activeCount := 0
	for _, s := range allSessions {
		if s.IsActive {
			activeCount++
		}
	}
	metrics.SessionsActive.Set(float64(activeCount))

	return nil
}

// ============================================================================
// Rate Limit Configuration
// ============================================================================

// SetRateLimitConfig stores a rate limit config
func (k *Keeper) SetRateLimitConfig(ctx sdk.Context, config *types.RateLimitConfig) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limit config: %w", err)
	}
	key := types.GetRateLimitConfigKey(config.UserAddress)
	return store.Set(key, bz)
}

// GetRateLimitConfig retrieves a rate limit config
func (k *Keeper) GetRateLimitConfig(ctx sdk.Context, userAddress string) (types.RateLimitConfig, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetRateLimitConfigKey(userAddress)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		// Return default config if not found
		params, _ := k.GetParams(ctx)
		var defaultPerMinute, defaultPerHour, defaultPerDay uint64 = 60, 3600, 86400
		if params != nil {
			defaultPerMinute = params.Auth.DefaultRequestsPerMinute
			defaultPerHour = params.Auth.DefaultRequestsPerHour
			defaultPerDay = params.Auth.DefaultRequestsPerDay
		}
		return types.RateLimitConfig{
			UserAddress:       userAddress,
			RequestsPerMinute: defaultPerMinute,
			RequestsPerHour:   defaultPerHour,
			RequestsPerDay:    defaultPerDay,
			WindowStart:       ctx.BlockTime(),
		}, nil
	}

	var config types.RateLimitConfig
	if err := json.Unmarshal(bz, &config); err != nil {
		return types.RateLimitConfig{}, fmt.Errorf("failed to unmarshal rate limit config: %w", err)
	}
	return config, nil
}

// GetAllRateLimitConfigs retrieves all rate limit configs
func (k *Keeper) GetAllRateLimitConfigs(ctx sdk.Context) ([]*types.RateLimitConfig, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.RateLimitConfigPrefix, storetypes.PrefixEndBytes(types.RateLimitConfigPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var configs []*types.RateLimitConfig
	for ; iterator.Valid(); iterator.Next() {
		var config types.RateLimitConfig
		if err := json.Unmarshal(iterator.Value(), &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rate limit config: %w", err)
		}
		configs = append(configs, &config)
	}
	return configs, nil
}

// ============================================================================
// Multisig Wallet Management
// ============================================================================

// SetMultisigWallet stores a multisig wallet
func (k *Keeper) SetMultisigWallet(ctx sdk.Context, wallet *types.MultisigWallet) error {
	if wallet.Threshold == 0 {
		return types.ErrInvalidMultisigWallet.Wrap("threshold must be greater than 0")
	}
	if uint32(len(wallet.Signers)) < wallet.Threshold {
		return types.ErrInvalidMultisigWallet.Wrap("threshold cannot exceed number of signers")
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := json.Marshal(wallet)
	if err != nil {
		return fmt.Errorf("failed to marshal multisig wallet: %w", err)
	}
	key := types.GetMultisigWalletKey(wallet.Id)
	return store.Set(key, bz)
}

// GetMultisigWallet retrieves a multisig wallet
func (k *Keeper) GetMultisigWallet(ctx sdk.Context, walletID string) (types.MultisigWallet, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetMultisigWalletKey(walletID)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return types.MultisigWallet{}, types.ErrMultisigWalletNotFound.Wrapf("wallet not found: %s", walletID)
	}

	var wallet types.MultisigWallet
	if err := json.Unmarshal(bz, &wallet); err != nil {
		return types.MultisigWallet{}, fmt.Errorf("failed to unmarshal multisig wallet: %w", err)
	}
	return wallet, nil
}

// GetAllMultisigWallets retrieves all multisig wallets
func (k *Keeper) GetAllMultisigWallets(ctx sdk.Context) ([]*types.MultisigWallet, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.MultisigWalletPrefix, storetypes.PrefixEndBytes(types.MultisigWalletPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var wallets []*types.MultisigWallet
	for ; iterator.Valid(); iterator.Next() {
		var wallet types.MultisigWallet
		if err := json.Unmarshal(iterator.Value(), &wallet); err != nil {
			return nil, fmt.Errorf("failed to unmarshal multisig wallet: %w", err)
		}
		wallets = append(wallets, &wallet)
	}
	return wallets, nil
}

// SetMultisigProposal stores a multisig proposal
func (k *Keeper) SetMultisigProposal(ctx sdk.Context, proposal *types.MultisigProposal) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := json.Marshal(proposal)
	if err != nil {
		return fmt.Errorf("failed to marshal multisig proposal: %w", err)
	}
	key := types.GetMultisigProposalKey(proposal.Id)
	return store.Set(key, bz)
}

// GetMultisigProposal retrieves a multisig proposal
func (k *Keeper) GetMultisigProposal(ctx sdk.Context, proposalID string) (types.MultisigProposal, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetMultisigProposalKey(proposalID)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return types.MultisigProposal{}, types.ErrProposalNotFound.Wrapf("proposal not found: %s", proposalID)
	}

	var proposal types.MultisigProposal
	if err := json.Unmarshal(bz, &proposal); err != nil {
		return types.MultisigProposal{}, fmt.Errorf("failed to unmarshal multisig proposal: %w", err)
	}
	return proposal, nil
}

// GetAllMultisigProposals retrieves all multisig proposals
func (k *Keeper) GetAllMultisigProposals(ctx sdk.Context) ([]*types.MultisigProposal, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.MultisigProposalPrefix, storetypes.PrefixEndBytes(types.MultisigProposalPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var proposals []*types.MultisigProposal
	for ; iterator.Valid(); iterator.Next() {
		var proposal types.MultisigProposal
		if err := json.Unmarshal(iterator.Value(), &proposal); err != nil {
			return nil, fmt.Errorf("failed to unmarshal multisig proposal: %w", err)
		}
		proposals = append(proposals, &proposal)
	}
	return proposals, nil
}

// ============================================================================
// Time-Locked Action Management
// ============================================================================

// SetTimeLockedAction stores a time-locked action
func (k *Keeper) SetTimeLockedAction(ctx sdk.Context, action *types.TimeLockedAction) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := json.Marshal(action)
	if err != nil {
		return fmt.Errorf("failed to marshal time-locked action: %w", err)
	}
	key := types.GetTimeLockedActionKey(action.Id)
	return store.Set(key, bz)
}

// GetTimeLockedAction retrieves a time-locked action
func (k *Keeper) GetTimeLockedAction(ctx sdk.Context, actionID string) (types.TimeLockedAction, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetTimeLockedActionKey(actionID)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return types.TimeLockedAction{}, types.ErrActionNotFound.Wrapf("action not found: %s", actionID)
	}

	var action types.TimeLockedAction
	if err := json.Unmarshal(bz, &action); err != nil {
		return types.TimeLockedAction{}, fmt.Errorf("failed to unmarshal time-locked action: %w", err)
	}
	return action, nil
}

// GetAllTimeLockedActions retrieves all time-locked actions
func (k *Keeper) GetAllTimeLockedActions(ctx sdk.Context) ([]*types.TimeLockedAction, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.TimeLockedActionPrefix, storetypes.PrefixEndBytes(types.TimeLockedActionPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var actions []*types.TimeLockedAction
	for ; iterator.Valid(); iterator.Next() {
		var action types.TimeLockedAction
		if err := json.Unmarshal(iterator.Value(), &action); err != nil {
			return nil, fmt.Errorf("failed to unmarshal time-locked action: %w", err)
		}
		actions = append(actions, &action)
	}
	return actions, nil
}

// ============================================================================
// Emergency Admin Management
// ============================================================================

// SetEmergencyAdmin stores an emergency admin
func (k *Keeper) SetEmergencyAdmin(ctx sdk.Context, admin *types.EmergencyAdmin) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := json.Marshal(admin)
	if err != nil {
		return fmt.Errorf("failed to marshal emergency admin: %w", err)
	}
	key := types.GetEmergencyAdminKey(admin.Address)
	return store.Set(key, bz)
}

// GetEmergencyAdmin retrieves an emergency admin
func (k *Keeper) GetEmergencyAdmin(ctx sdk.Context, address string) (types.EmergencyAdmin, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetEmergencyAdminKey(address)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return types.EmergencyAdmin{}, types.ErrEmergencyAdminNotFound.Wrapf("emergency admin not found: %s", address)
	}

	var admin types.EmergencyAdmin
	if err := json.Unmarshal(bz, &admin); err != nil {
		return types.EmergencyAdmin{}, fmt.Errorf("failed to unmarshal emergency admin: %w", err)
	}
	return admin, nil
}

// GetAllEmergencyAdmins retrieves all emergency admins
func (k *Keeper) GetAllEmergencyAdmins(ctx sdk.Context) ([]*types.EmergencyAdmin, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.EmergencyAdminPrefix, storetypes.PrefixEndBytes(types.EmergencyAdminPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var admins []*types.EmergencyAdmin
	for ; iterator.Valid(); iterator.Next() {
		var admin types.EmergencyAdmin
		if err := json.Unmarshal(iterator.Value(), &admin); err != nil {
			return nil, fmt.Errorf("failed to unmarshal emergency admin: %w", err)
		}
		admins = append(admins, &admin)
	}
	return admins, nil
}

// ============================================================================
// Validator Key Rotation Management
// ============================================================================

// SetValidatorRotation stores a validator key rotation
func (k *Keeper) SetValidatorRotation(ctx sdk.Context, rotation *types.ValidatorKeyRotation) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := json.Marshal(rotation)
	if err != nil {
		return fmt.Errorf("failed to marshal validator rotation: %w", err)
	}
	key := types.GetValidatorRotationKey(rotation.ValidatorAddress)
	return store.Set(key, bz)
}

// GetValidatorRotation retrieves a validator key rotation
func (k *Keeper) GetValidatorRotation(ctx sdk.Context, validatorAddress string) (types.ValidatorKeyRotation, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetValidatorRotationKey(validatorAddress)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return types.ValidatorKeyRotation{}, types.ErrRotationNotFound.Wrapf("rotation not found: %s", validatorAddress)
	}

	var rotation types.ValidatorKeyRotation
	if err := json.Unmarshal(bz, &rotation); err != nil {
		return types.ValidatorKeyRotation{}, fmt.Errorf("failed to unmarshal validator rotation: %w", err)
	}
	return rotation, nil
}

// GetAllValidatorRotations retrieves all validator rotations
func (k *Keeper) GetAllValidatorRotations(ctx sdk.Context) ([]*types.ValidatorKeyRotation, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.ValidatorRotationPrefix, storetypes.PrefixEndBytes(types.ValidatorRotationPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var rotations []*types.ValidatorKeyRotation
	for ; iterator.Valid(); iterator.Next() {
		var rotation types.ValidatorKeyRotation
		if err := json.Unmarshal(iterator.Value(), &rotation); err != nil {
			return nil, fmt.Errorf("failed to unmarshal validator rotation: %w", err)
		}
		rotations = append(rotations, &rotation)
	}
	return rotations, nil
}
