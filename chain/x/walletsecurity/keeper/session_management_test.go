// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// =============================================================================
// Session Management Tests
// =============================================================================

type SessionManagementTestSuite struct {
	KeeperTestSuite
}

func (suite *SessionManagementTestSuite) TestCreateSession_Success() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	session, err := k.CreateSession(ctx, "wallet-1", 30*time.Minute, "device-fp-123")
	suite.Require().NoError(err)
	suite.Require().NotNil(session)
	suite.Require().NotEmpty(session.SessionId)
	suite.Require().Equal("wallet-1", session.WalletId)
	suite.Require().Equal("device-fp-123", session.DeviceFingerprint)
	suite.Require().True(session.AutoLockEnabled)
	suite.Require().False(session.Locked)
	suite.Require().NotNil(session.StartedAt)
	suite.Require().NotNil(session.ExpiresAt)
}

func (suite *SessionManagementTestSuite) TestCreateSession_DifferentWallets() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Sessions for different wallets should be distinct
	session1, err := k.CreateSession(ctx, "wallet-2a", 10*time.Minute, "fp-1")
	suite.Require().NoError(err)
	suite.Require().NotEmpty(session1.SessionId)

	session2, err := k.CreateSession(ctx, "wallet-2b", 10*time.Minute, "fp-2")
	suite.Require().NoError(err)
	suite.Require().NotEmpty(session2.SessionId)

	// Session IDs include wallet ID, so different wallets get different IDs
	suite.Require().NotEqual(session1.SessionId, session2.SessionId)
}

func (suite *SessionManagementTestSuite) TestValidateSession_Valid() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	session, err := k.CreateSession(ctx, "wallet-3", 30*time.Minute, "fp")
	suite.Require().NoError(err)

	isValid, err := k.ValidateSession(ctx, session.SessionId)
	suite.Require().NoError(err)
	suite.Require().True(isValid)
}

func (suite *SessionManagementTestSuite) TestValidateSession_NotFound() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	isValid, err := k.ValidateSession(ctx, "nonexistent-session")
	suite.Require().Error(err)
	suite.Require().False(isValid)
}

func (suite *SessionManagementTestSuite) TestValidateSession_Locked() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	session, err := k.CreateSession(ctx, "wallet-4", 30*time.Minute, "fp")
	suite.Require().NoError(err)

	// Lock the session
	err = k.LockSession(ctx, session.SessionId)
	suite.Require().NoError(err)

	// Validate should fail for locked session
	isValid, err := k.ValidateSession(ctx, session.SessionId)
	suite.Require().Error(err)
	suite.Require().False(isValid)
	suite.Require().Contains(err.Error(), "locked")
}

func (suite *SessionManagementTestSuite) TestLockSession_Success() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	session, err := k.CreateSession(ctx, "wallet-5", 30*time.Minute, "fp")
	suite.Require().NoError(err)

	err = k.LockSession(ctx, session.SessionId)
	suite.Require().NoError(err)
}

func (suite *SessionManagementTestSuite) TestLockSession_NotFound() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	err := k.LockSession(ctx, "nonexistent-session")
	suite.Require().Error(err)
}

func (suite *SessionManagementTestSuite) TestUnlockSession_Success() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	session, err := k.CreateSession(ctx, "wallet-6", 30*time.Minute, "fp")
	suite.Require().NoError(err)

	// Lock first
	err = k.LockSession(ctx, session.SessionId)
	suite.Require().NoError(err)

	// Unlock with valid auth proof
	err = k.UnlockSession(ctx, session.SessionId, []byte("valid-auth-proof"))
	suite.Require().NoError(err)
}

func (suite *SessionManagementTestSuite) TestUnlockSession_NotLocked() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	session, err := k.CreateSession(ctx, "wallet-7", 30*time.Minute, "fp")
	suite.Require().NoError(err)

	// Try to unlock a session that's not locked
	err = k.UnlockSession(ctx, session.SessionId, []byte("auth-proof"))
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "not locked")
}

func (suite *SessionManagementTestSuite) TestUpdateSessionActivity_Success() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	session, err := k.CreateSession(ctx, "wallet-8", 30*time.Minute, "fp")
	suite.Require().NoError(err)

	err = k.UpdateSessionActivity(ctx, session.SessionId)
	suite.Require().NoError(err)
}

func (suite *SessionManagementTestSuite) TestUpdateSessionActivity_NotFound() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	err := k.UpdateSessionActivity(ctx, "nonexistent-session")
	suite.Require().Error(err)
}

func (suite *SessionManagementTestSuite) TestLockSessionDueToInactivity() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	session, err := k.CreateSession(ctx, "wallet-9", 30*time.Minute, "fp")
	suite.Require().NoError(err)

	err = k.LockSessionDueToInactivity(ctx, session.SessionId)
	suite.Require().NoError(err)

	// Validate should fail now
	isValid, err := k.ValidateSession(ctx, session.SessionId)
	suite.Require().Error(err)
	suite.Require().False(isValid)
}

func (suite *SessionManagementTestSuite) TestTerminateSession_Success() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	session, err := k.CreateSession(ctx, "wallet-10", 30*time.Minute, "fp")
	suite.Require().NoError(err)

	err = k.TerminateSession(ctx, session.SessionId)
	suite.Require().NoError(err)

	// Session should no longer be found
	isValid, err := k.ValidateSession(ctx, session.SessionId)
	suite.Require().Error(err)
	suite.Require().False(isValid)
}

func (suite *SessionManagementTestSuite) TestVerifyAuthProof_ValidProof() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Create a real session
	session, err := k.CreateSession(ctx, "wallet-auth-test-1", 30*time.Minute, "fp")
	suite.Require().NoError(err)

	// verifyAuthProof checks len(proof) > 0
	result := k.verifyAuthProof(session, []byte("valid-proof"))
	suite.Require().True(result)
}

func (suite *SessionManagementTestSuite) TestVerifyAuthProof_EmptyProof() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	session, err := k.CreateSession(ctx, "wallet-auth-test-2", 30*time.Minute, "fp")
	suite.Require().NoError(err)

	result := k.verifyAuthProof(session, []byte{})
	suite.Require().False(result)
}

func (suite *SessionManagementTestSuite) TestVerifyAuthProof_NilProof() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	session, err := k.CreateSession(ctx, "wallet-auth-test-3", 30*time.Minute, "fp")
	suite.Require().NoError(err)

	result := k.verifyAuthProof(session, nil)
	suite.Require().False(result)
}

func TestSessionManagementTestSuite(t *testing.T) {
	suite.Run(t, new(SessionManagementTestSuite))
}
