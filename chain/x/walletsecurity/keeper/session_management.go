// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"
	"time"

	authtypes "github.com/aequitas/aura/chain/x/auth/types"
	"github.com/aequitas/aura/chain/x/common/determinism"
	"github.com/aequitas/aura/chain/x/walletsecurity/types"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
)

// CreateSession creates a new wallet session
func (k Keeper) CreateSession(ctx context.Context, walletID string, duration time.Duration, deviceFingerprint string) (*wsproto.SessionConfig, error) {
	blockTime := determinism.GetBlockTime(ctx)
	sessionID := fmt.Sprintf("session_%s_%d", walletID, blockTime.UnixNano())

	session := &wsproto.SessionConfig{
		SessionId:                  sessionID,
		WalletId:                   walletID,
		DeviceFingerprint:          deviceFingerprint,
		StartedAt:                  timeToGogoTimestamp(blockTime),
		LastActivity:               timeToGogoTimestamp(blockTime),
		TimeoutDuration:            &gogotypes.Duration{Seconds: int64(duration.Seconds()), Nanos: int32(duration.Nanoseconds() % 1e9)},
		ExpiresAt:                  timeToGogoTimestamp(blockTime.Add(duration)),
		AutoLockEnabled:            true,
		InactivityThresholdSeconds: 300, // 5 minutes
		Locked:                     false,
	}

	// Store session
	sessionBytes, err := k.cdc.Marshal(session)
	if err != nil {
		return nil, err
	}

	if err := k.SetSessionConfig(ctx, sessionID, sessionBytes); err != nil {
		return nil, err
	}

	// Emit event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeSessionCreated,
		sdk.NewAttribute(types.AttributeKeySessionID, sessionID),
		sdk.NewAttribute(types.AttributeKeyWalletID, walletID),
		sdk.NewAttribute(types.AttributeKeyDeviceID, deviceFingerprint),
	))

	return session, nil
}

// ValidateSession validates if a session is still active
func (k Keeper) ValidateSession(ctx context.Context, sessionID string) (bool, error) {
	sessionBytes, err := k.GetSessionConfig(ctx, sessionID)
	if err != nil {
		return false, err
	}

	var session wsproto.SessionConfig
	if err := k.cdc.Unmarshal(sessionBytes, &session); err != nil {
		return false, err
	}

	blockTime := determinism.GetBlockTime(ctx)

	// Check if locked
	if session.Locked {
		return false, types.ErrSessionLocked
	}

	// Check expiration
	if session.ExpiresAt != nil && blockTime.After(gogoTimestampToTime(session.ExpiresAt)) {
		if err := k.TerminateSession(ctx, sessionID); err != nil {
			return false, err
		}
		return false, types.ErrSessionExpired
	}

	// Check inactivity timeout
	if session.AutoLockEnabled && session.LastActivity != nil {
		inactiveDuration := blockTime.Sub(gogoTimestampToTime(session.LastActivity))
		thresholdDuration := time.Duration(session.InactivityThresholdSeconds) * time.Second
		if inactiveDuration > thresholdDuration {
			if err := k.LockSessionDueToInactivity(ctx, sessionID); err != nil {
				return false, err
			}
			return false, types.ErrSessionInactive
		}
	}

	return true, nil
}

// UpdateSessionActivity updates the last activity time
func (k Keeper) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	sessionBytes, err := k.GetSessionConfig(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	var session wsproto.SessionConfig
	if err := k.cdc.Unmarshal(sessionBytes, &session); err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	blockTime := determinism.GetBlockTime(ctx)
	session.LastActivity = timeToGogoTimestamp(blockTime)

	updatedBytes, err := k.cdc.Marshal(&session)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	return k.SetSessionConfig(ctx, sessionID, updatedBytes)
}

// LockSessionDueToInactivity locks a session due to inactivity
func (k Keeper) LockSessionDueToInactivity(ctx context.Context, sessionID string) error {
	sessionBytes, err := k.GetSessionConfig(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	var session wsproto.SessionConfig
	if err := k.cdc.Unmarshal(sessionBytes, &session); err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	session.Locked = true

	updatedBytes, err := k.cdc.Marshal(&session)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	if err := k.SetSessionConfig(ctx, sessionID, updatedBytes); err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	// Emit event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeSessionLocked,
		sdk.NewAttribute(types.AttributeKeySessionID, sessionID),
	))

	return nil
}

// UnlockSessionAfterAuth unlocks a locked session after authentication
func (k Keeper) UnlockSessionAfterAuth(ctx context.Context, sessionID string, authProof []byte) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err := k.requireActiveAuthSession(sdkCtx, sessionID); err != nil {
		return fmt.Errorf("failed to UnwrapSDKContext: %w", err)
	}

	sessionBytes, err := k.GetSessionConfig(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	var session wsproto.SessionConfig
	if err := k.cdc.Unmarshal(sessionBytes, &session); err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	// Verify authentication proof (biometric, password, etc.)
	if !k.verifyAuthProof(&session, authProof) {
		return types.ErrUnauthorized
	}

	blockTime := determinism.GetBlockTime(ctx)
	session.Locked = false
	session.LastActivity = timeToGogoTimestamp(blockTime)

	updatedBytes, err := k.cdc.Marshal(&session)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	return k.SetSessionConfig(ctx, sessionID, updatedBytes)
}

// TerminateSession terminates a session
func (k Keeper) TerminateSession(ctx context.Context, sessionID string) error {
	kvStore := k.getStore(ctx)
	key := types.GetSessionConfigKey(sessionID)
	if err := kvStore.Delete(key); err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	// Emit event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeSessionTerminated,
		sdk.NewAttribute(types.AttributeKeySessionID, sessionID),
	))

	return nil
}

func (k Keeper) verifyAuthProof(session *wsproto.SessionConfig, proof []byte) bool {
	// In production, verify biometric or password hash
	return len(proof) > 0
}

func (k Keeper) requireActiveAuthSession(ctx sdk.Context, sessionID string) error {
	if k.authKeeper == nil {
		return nil
	}

	session, err := k.authKeeper.GetSession(ctx, sessionID)
	if err != nil || session == nil {
		return types.ErrInactiveSession
	}

	if !authtypes.IsSessionActive(session, ctx.BlockTime()) {
		return types.ErrInactiveSession
	}

	return nil
}
