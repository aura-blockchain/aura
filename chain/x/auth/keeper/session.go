package keeper

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// CreateSession creates a new API session
func (k *Keeper) CreateSession(ctx context.Context, userAddress, ipAddress string, metadata map[string]string) (*authproto.Session, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	now := time.Now()
	sessionID := k.generateSessionID()
	expiresAt := now.Add(time.Duration(k.params.SessionTimeoutSeconds) * time.Second)

	session := &authproto.Session{
		SessionId:    sessionID,
		UserAddress:  userAddress,
		CreatedAt:    &now,
		ExpiresAt:    &expiresAt,
		LastAccessed: &now,
		IpAddress:    ipAddress,
		IsActive:     true,
		Metadata:     metadata,
	}

	// Validate session
	if err := types.ValidateSession(session); err != nil {
		k.LogAudit(ctx, userAddress, "create_session", sessionID, "failed", nil, err.Error())
		return nil, fmt.Errorf("%w: %v", types.ErrInvalidSession, err)
	}

	k.sessions[sessionID] = session
	k.userSessions[userAddress] = append(k.userSessions[userAddress], sessionID)

	k.LogAudit(ctx, userAddress, "create_session", sessionID, "success", map[string]string{
		"ip_address": ipAddress,
		"expires_at": expiresAt.Format(time.RFC3339),
	}, "")

	return session, nil
}

// GetSession retrieves a session by ID

// ValidateSession checks if a session is valid and active
func (k *Keeper) ValidateSession(ctx context.Context, sessionID string) (*authproto.Session, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	session, ok := k.sessions[sessionID]
	if !ok {
		return nil, types.ErrSessionNotFound
	}

	// Check if session is active
	if !types.IsSessionActive(session) {
		return nil, types.ErrSessionInactive
	}

	// Check if session has expired
	if session.ExpiresAt != nil && time.Now().After(session.ExpiresAt.AsTime()) {
		session.IsActive = false
		k.LogAudit(ctx, session.UserAddress, "session_expired", sessionID, "expired", nil, "")
		return nil, types.ErrSessionExpired
	}

	// Update last accessed time
	now := time.Now()
	session.LastAccessed = &now

	return session, nil
}

// RefreshSession extends the expiry time of a session
func (k *Keeper) RefreshSession(ctx context.Context, sessionID string) (*authproto.Session, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	session, ok := k.sessions[sessionID]
	if !ok {
		return nil, types.ErrSessionNotFound
	}

	if !types.IsSessionActive(session) {
		return nil, types.ErrSessionInactive
	}

	now := time.Now()
	newExpiresAt := now.Add(time.Duration(k.params.SessionTimeoutSeconds) * time.Second)
	session.ExpiresAt = &newExpiresAt
	session.LastAccessed = &now

	k.LogAudit(ctx, session.UserAddress, "refresh_session", sessionID, "success", map[string]string{
		"new_expires_at": newExpiresAt.Format(time.RFC3339),
	}, "")

	return session, nil
}

// RevokeSession revokes an active session
func (k *Keeper) RevokeSession(ctx context.Context, userAddress, sessionID string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	session, ok := k.sessions[sessionID]
	if !ok {
		k.LogAudit(ctx, userAddress, "revoke_session", sessionID, "failed", nil, "session not found")
		return types.ErrSessionNotFound
	}

	// Verify user owns the session or has manage permission
	if session.UserAddress != userAddress {
		if !k.HasPermission(userAddress, types.PermissionManageSession) {
			k.LogAudit(ctx, userAddress, "revoke_session", sessionID, "failed", nil, "insufficient permissions")
			return types.ErrInsufficientPermissions
		}
	}

	session.IsActive = false

	k.LogAudit(ctx, userAddress, "revoke_session", sessionID, "success", nil, "")

	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (k *Keeper) RevokeAllUserSessions(ctx context.Context, revoker, userAddress string) (int, error) {
	// Validate revoker has permission
	if revoker != userAddress {
		if err := k.RequirePermission(ctx, revoker, types.PermissionManageSession); err != nil {
			return 0, err
		}
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	sessionIDs, ok := k.userSessions[userAddress]
	if !ok {
		return 0, nil
	}

	count := 0
	for _, sessionID := range sessionIDs {
		if session, ok := k.sessions[sessionID]; ok {
			if session.IsActive {
				session.IsActive = false
				count++
			}
		}
	}

	k.LogAudit(ctx, revoker, "revoke_all_sessions", userAddress, "success", map[string]string{
		"revoked_count": fmt.Sprintf("%d", count),
	}, "")

	return count, nil
}

// ListSessions returns all sessions for a user
func (k *Keeper) ListSessions(userAddress string) []*authproto.Session {
	k.mu.RLock()
	defer k.mu.RUnlock()

	sessionIDs, ok := k.userSessions[userAddress]
	if !ok {
		return []*authproto.Session{}
	}

	sessions := make([]*authproto.Session, 0)
	for _, sessionID := range sessionIDs {
		if session, ok := k.sessions[sessionID]; ok {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// ListActiveSessions returns all active sessions for a user
func (k *Keeper) ListActiveSessions(userAddress string) []*authproto.Session {
	k.mu.RLock()
	defer k.mu.RUnlock()

	sessionIDs, ok := k.userSessions[userAddress]
	if !ok {
		return []*authproto.Session{}
	}

	sessions := make([]*authproto.Session, 0)
	for _, sessionID := range sessionIDs {
		if session, ok := k.sessions[sessionID]; ok {
			if types.IsSessionActive(session) {
				sessions = append(sessions, session)
			}
		}
	}

	return sessions
}

// generateSessionID generates a cryptographically secure session ID
func (k *Keeper) generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GetSessionByUser retrieves the most recent active session for a user
func (k *Keeper) GetSessionByUser(userAddress string) (*authproto.Session, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	sessionIDs, ok := k.userSessions[userAddress]
	if !ok || len(sessionIDs) == 0 {
		return nil, types.ErrSessionNotFound
	}

	// Find the most recent active session
	var mostRecent *authproto.Session
	for i := len(sessionIDs) - 1; i >= 0; i-- {
		if session, ok := k.sessions[sessionIDs[i]]; ok {
			if types.IsSessionActive(session) {
				if mostRecent == nil || session.CreatedAt.After(*mostRecent.CreatedAt) {
					mostRecent = session
				}
			}
		}
	}

	if mostRecent == nil {
		return nil, types.ErrSessionNotFound
	}

	return mostRecent, nil
}

// UpdateSessionMetadata updates the metadata of a session
func (k *Keeper) UpdateSessionMetadata(ctx context.Context, sessionID string, metadata map[string]string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	session, ok := k.sessions[sessionID]
	if !ok {
		return types.ErrSessionNotFound
	}

	// Merge metadata
	if session.Metadata == nil {
		session.Metadata = make(map[string]string)
	}
	for key, value := range metadata {
		session.Metadata[key] = value
	}

	now := time.Now()
	session.LastAccessed = &now

	k.LogAudit(ctx, session.UserAddress, "update_session_metadata", sessionID, "success", nil, "")

	return nil
}
