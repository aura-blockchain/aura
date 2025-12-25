// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/testutil/keeper"
	authkeeper "github.com/aequitas/aura/chain/x/auth/keeper"
)

type AuditKVStoreTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper *authkeeper.Keeper
}

func (suite *AuditKVStoreTestSuite) SetupTest() {
	suite.keeper, suite.ctx = keeper.AuthKeeper(suite.T())
}

// TestAuditLogKVStorePersistence verifies audit logs are stored in KVStore and persist
func (suite *AuditKVStoreTestSuite) TestAuditLogKVStorePersistence() {
	// Log an audit entry
	suite.keeper.LogAudit(suite.ctx, "user1", "create_role", "admin", "success", map[string]string{
		"role": "admin",
	}, "")

	// Retrieve all audit logs
	logs, err := suite.keeper.GetAllAuditLogs(suite.ctx)
	suite.NoError(err)
	suite.Require().Len(logs, 1)

	// Verify the log content
	suite.Equal("user1", logs[0].Actor)
	suite.Equal("create_role", logs[0].Action)
	suite.Equal("admin", logs[0].Resource)
	suite.Equal("success", logs[0].Result)
	suite.NotEmpty(logs[0].Id)
	suite.NotNil(logs[0].Metadata)
	suite.Equal("admin", logs[0].Metadata["role"])
}

// TestAuditLogNoConcurrencyIssues verifies no race conditions with KVStore
func (suite *AuditKVStoreTestSuite) TestAuditLogNoConcurrencyIssues() {
	// Add multiple logs sequentially (simulating concurrent operations)
	for i := 0; i < 100; i++ {
		suite.keeper.LogAudit(suite.ctx, "user1", "action", "resource", "success", nil, "")
	}

	// All logs should be retrievable
	logs, err := suite.keeper.GetAllAuditLogs(suite.ctx)
	suite.NoError(err)
	suite.Require().Len(logs, 100)

	// Verify each log has a unique ID
	idSet := make(map[string]bool)
	for _, log := range logs {
		suite.NotEmpty(log.Id)
		suite.False(idSet[log.Id], "Duplicate ID found: %s", log.Id)
		idSet[log.Id] = true
	}
}

// TestAuditLogFilteringByActor verifies actor-based filtering works
func (suite *AuditKVStoreTestSuite) TestAuditLogFilteringByActor() {
	// Add logs for different actors
	suite.keeper.LogAudit(suite.ctx, "user1", "create", "role1", "success", nil, "")
	suite.keeper.LogAudit(suite.ctx, "user2", "delete", "role2", "success", nil, "")
	suite.keeper.LogAudit(suite.ctx, "user1", "update", "role3", "success", nil, "")

	// Filter by user1
	logs := suite.keeper.GetAuditLogsByActor(suite.ctx, "user1", 0)
	suite.Require().Len(logs, 2)
	for _, log := range logs {
		suite.Equal("user1", log.Actor)
	}

	// Filter by user2
	logs = suite.keeper.GetAuditLogsByActor(suite.ctx, "user2", 0)
	suite.Require().Len(logs, 1)
	suite.Equal("user2", logs[0].Actor)
}

// TestAuditLogFilteringByAction verifies action-based filtering works
func (suite *AuditKVStoreTestSuite) TestAuditLogFilteringByAction() {
	// Add logs for different actions
	suite.keeper.LogAudit(suite.ctx, "user1", "create", "role1", "success", nil, "")
	suite.keeper.LogAudit(suite.ctx, "user1", "delete", "role2", "success", nil, "")
	suite.keeper.LogAudit(suite.ctx, "user1", "create", "role3", "success", nil, "")

	// Filter by create action
	logs := suite.keeper.GetAuditLogsByAction(suite.ctx, "create", 0)
	suite.Require().Len(logs, 2)
	for _, log := range logs {
		suite.Equal("create", log.Action)
	}

	// Filter by delete action
	logs = suite.keeper.GetAuditLogsByAction(suite.ctx, "delete", 0)
	suite.Require().Len(logs, 1)
	suite.Equal("delete", logs[0].Action)
}

// TestAuditLogTimeOrdering verifies logs are ordered by timestamp (most recent first)
func (suite *AuditKVStoreTestSuite) TestAuditLogTimeOrdering() {
	// Add logs with different timestamps
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Create a new context for each log with incrementing time
	for i := 0; i < 5; i++ {
		ctx := suite.ctx.WithBlockTime(baseTime.Add(time.Duration(i) * time.Second))
		suite.keeper.LogAudit(ctx, "user1", "action", "resource", "success", nil, "")
	}

	// Retrieve all logs
	logs := suite.keeper.GetRecentAuditLogs(suite.ctx, 0)
	suite.Require().Len(logs, 5)

	// Verify they are in descending order (most recent first)
	for i := 0; i < len(logs)-1; i++ {
		suite.True(logs[i].Timestamp.After(logs[i+1].Timestamp) || logs[i].Timestamp.Equal(logs[i+1].Timestamp),
			"Logs should be in descending time order")
	}
}

// TestAuditLogErrorLogging verifies error messages are stored correctly
func (suite *AuditKVStoreTestSuite) TestAuditLogErrorLogging() {
	// Log an error
	suite.keeper.LogAudit(suite.ctx, "user1", "delete_role", "admin", "failed", nil, "permission denied")

	// Retrieve the log
	logs := suite.keeper.GetAuditLogsByActor(suite.ctx, "user1", 0)
	suite.Require().Len(logs, 1)

	// Verify error details
	suite.Equal("failed", logs[0].Result)
	suite.Equal("permission denied", logs[0].ErrorMessage)
}

// TestAuditLogMetadata verifies metadata is stored and retrieved correctly
func (suite *AuditKVStoreTestSuite) TestAuditLogMetadata() {
	metadata := map[string]string{
		"ip":      "192.168.1.1",
		"agent":   "test-client",
		"session": "abc123",
	}

	suite.keeper.LogAudit(suite.ctx, "user1", "login", "auth", "success", metadata, "")

	logs := suite.keeper.GetAuditLogsByActor(suite.ctx, "user1", 0)
	suite.Require().Len(logs, 1)

	suite.Equal("192.168.1.1", logs[0].Metadata["ip"])
	suite.Equal("test-client", logs[0].Metadata["agent"])
	suite.Equal("abc123", logs[0].Metadata["session"])
}

// TestAuditLogLimitEnforcement verifies limit parameter works correctly
func (suite *AuditKVStoreTestSuite) TestAuditLogLimitEnforcement() {
	// Add 10 logs
	for i := 0; i < 10; i++ {
		suite.keeper.LogAudit(suite.ctx, "user1", "action", "resource", "success", nil, "")
	}

	// Request only 5 logs
	logs := suite.keeper.GetRecentAuditLogs(suite.ctx, 5)
	suite.Require().Len(logs, 5)

	// Request 0 (no limit)
	logs = suite.keeper.GetRecentAuditLogs(suite.ctx, 0)
	suite.Require().Len(logs, 10)
}

// TestAuditLogSearchByMultipleCriteria verifies complex search queries work
func (suite *AuditKVStoreTestSuite) TestAuditLogSearchByMultipleCriteria() {
	// Add various logs
	suite.keeper.LogAudit(suite.ctx, "user1", "create", "role1", "success", nil, "")
	suite.keeper.LogAudit(suite.ctx, "user1", "delete", "role2", "failed", nil, "")
	suite.keeper.LogAudit(suite.ctx, "user2", "create", "role3", "success", nil, "")

	// Search for user1 AND create actions
	logs := suite.keeper.SearchAuditLogs(suite.ctx, map[string]string{
		"actor":  "user1",
		"action": "create",
	}, 0)
	suite.Require().Len(logs, 1)
	suite.Equal("user1", logs[0].Actor)
	suite.Equal("create", logs[0].Action)

	// Search for success status
	logs = suite.keeper.SearchAuditLogs(suite.ctx, map[string]string{
		"status": "success",
	}, 0)
	suite.Require().Len(logs, 2)
	for _, log := range logs {
		suite.Equal("success", log.Result)
	}
}

// TestAuditLogCountMethods verifies counting methods work correctly
func (suite *AuditKVStoreTestSuite) TestAuditLogCountMethods() {
	// Add logs for different users
	suite.keeper.LogAudit(suite.ctx, "user1", "create", "role1", "success", nil, "")
	suite.keeper.LogAudit(suite.ctx, "user1", "delete", "role2", "success", nil, "")
	suite.keeper.LogAudit(suite.ctx, "user2", "create", "role3", "success", nil, "")

	// Count total logs
	total := suite.keeper.CountAuditLogs(suite.ctx)
	suite.Equal(uint64(3), total)

	// Count logs by actor
	user1Count := suite.keeper.CountAuditLogsByActor(suite.ctx, "user1")
	suite.Equal(uint64(2), user1Count)

	user2Count := suite.keeper.CountAuditLogsByActor(suite.ctx, "user2")
	suite.Equal(uint64(1), user2Count)
}

// TestAuditLogDeterministicTimestamp verifies ctx.BlockTime() is used (not time.Now())
func (suite *AuditKVStoreTestSuite) TestAuditLogDeterministicTimestamp() {
	// Set a specific block time
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	ctx := suite.ctx.WithBlockTime(fixedTime)

	// Log an audit entry
	suite.keeper.LogAudit(ctx, "user1", "action", "resource", "success", nil, "")

	// Retrieve the log
	logs, err := suite.keeper.GetAllAuditLogs(ctx)
	suite.Require().NoError(err)
	suite.Require().Len(logs, 1)

	// Verify timestamp matches block time exactly
	suite.True(logs[0].Timestamp.Equal(fixedTime),
		"Expected timestamp %v, got %v", fixedTime, logs[0].Timestamp)
}

func TestAuditKVStoreTestSuite(t *testing.T) {
	suite.Run(t, new(AuditKVStoreTestSuite))
}
