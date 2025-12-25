// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"errors"
	"time"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

func (suite *KeeperTestSuite) TestRegisterSentryNodeInvalidValidator() {
	// Try to register sentry for non-existent validator
	err := suite.keeper.RegisterSentryNode(suite.ctx, "nonexistent", "sentry1", "192.168.1.1", 26656)
	suite.Require().Error(err)
	suite.Require().True(errors.Is(err, types.ErrValidatorNotFound))
}

func (suite *KeeperTestSuite) TestRegisterSentryNodeInvalidIP() {
	validatorAddr := "auravaloper1sentryip"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Try with empty IP
	err = suite.keeper.RegisterSentryNode(suite.ctx, validatorAddr, "sentry1", "", 26656)
	suite.Require().Error(err)
	suite.Require().True(errors.Is(err, types.ErrInvalidSentryNode))
}

func (suite *KeeperTestSuite) TestRegisterSentryNodeInvalidPort() {
	validatorAddr := "auravaloper1sentryport"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Try with invalid port
	err = suite.keeper.RegisterSentryNode(suite.ctx, validatorAddr, "sentry1", "192.168.1.1", 0)
	suite.Require().Error(err)
	suite.Require().True(errors.Is(err, types.ErrInvalidSentryNode))

	// Try with port > 65535
	err = suite.keeper.RegisterSentryNode(suite.ctx, validatorAddr, "sentry1", "192.168.1.1", 70000)
	suite.Require().Error(err)
	suite.Require().True(errors.Is(err, types.ErrInvalidSentryNode))
}

func (suite *KeeperTestSuite) TestRegisterSentryNodeDuplicate() {
	validatorAddr := "auravaloper1sentrydup"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Register sentry
	err = suite.keeper.RegisterSentryNode(suite.ctx, validatorAddr, "sentry1", "192.168.1.1", 26656)
	suite.Require().NoError(err)

	// Register same sentry again - should succeed without error
	err = suite.keeper.RegisterSentryNode(suite.ctx, validatorAddr, "sentry1", "192.168.1.1", 26656)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestGetSentryNodeInfoNotFound() {
	_, err := suite.keeper.GetSentryNodeInfo(suite.ctx, "nonexistent")
	suite.Require().Error(err)
	suite.Require().True(errors.Is(err, types.ErrSentryNodeNotFound))
}

func (suite *KeeperTestSuite) TestUpdateSentryHeartbeatNotFound() {
	err := suite.keeper.UpdateSentryHeartbeat(suite.ctx, "nonexistent")
	suite.Require().Error(err)
	suite.Require().True(errors.Is(err, types.ErrSentryNodeNotFound))
}

func (suite *KeeperTestSuite) TestRecordSentryRequest() {
	validatorAddr := "auravaloper1sentryreq"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	err = suite.keeper.RegisterSentryNode(suite.ctx, validatorAddr, "sentry1", "192.168.1.1", 26656)
	suite.Require().NoError(err)

	// Record normal request
	err = suite.keeper.RecordSentryRequest(suite.ctx, "sentry1", false)
	suite.Require().NoError(err)

	// Record blocked request
	err = suite.keeper.RecordSentryRequest(suite.ctx, "sentry1", true)
	suite.Require().NoError(err)

	// Verify counts
	node, err := suite.keeper.GetSentryNodeInfo(suite.ctx, "sentry1")
	suite.Require().NoError(err)
	suite.Require().Equal(int64(2), node.RequestCount)
	suite.Require().Equal(int64(1), node.BlockedRequests)
}

func (suite *KeeperTestSuite) TestRecordSentryRequestNotFound() {
	err := suite.keeper.RecordSentryRequest(suite.ctx, "nonexistent", false)
	suite.Require().Error(err)
	suite.Require().True(errors.Is(err, types.ErrSentryNodeNotFound))
}

func (suite *KeeperTestSuite) TestDeactivateSentryNode() {
	validatorAddr := "auravaloper1sentrydeact"

	// Setup
	params := types.DefaultParams()
	params.RequireSentryNodes = true
	params.MinSentryNodes = 2
	suite.Require().NoError(suite.keeper.SetParams(suite.ctx, params))

	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Register multiple sentries
	err = suite.keeper.RegisterSentryNode(suite.ctx, validatorAddr, "sentry1", "192.168.1.1", 26656)
	suite.Require().NoError(err)
	err = suite.keeper.RegisterSentryNode(suite.ctx, validatorAddr, "sentry2", "192.168.1.2", 26656)
	suite.Require().NoError(err)

	// Deactivate one
	err = suite.keeper.DeactivateSentryNode(suite.ctx, "sentry1")
	suite.Require().NoError(err)

	// Verify deactivated
	node, err := suite.keeper.GetSentryNodeInfo(suite.ctx, "sentry1")
	suite.Require().NoError(err)
	suite.Require().False(node.IsActive)
}

func (suite *KeeperTestSuite) TestDeactivateSentryNodeNotFound() {
	err := suite.keeper.DeactivateSentryNode(suite.ctx, "nonexistent")
	suite.Require().Error(err)
	suite.Require().True(errors.Is(err, types.ErrSentryNodeNotFound))
}

func (suite *KeeperTestSuite) TestDeactivateSentryNodeBelowMinimum() {
	validatorAddr := "auravaloper1sentrymin"

	// Setup
	params := types.DefaultParams()
	params.RequireSentryNodes = true
	params.MinSentryNodes = 2
	params.EnableAutoFailover = false
	suite.Require().NoError(suite.keeper.SetParams(suite.ctx, params))

	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Register only one sentry
	err = suite.keeper.RegisterSentryNode(suite.ctx, validatorAddr, "sentry1", "192.168.1.1", 26656)
	suite.Require().NoError(err)

	// Deactivate - should create critical alert
	err = suite.keeper.DeactivateSentryNode(suite.ctx, "sentry1")
	suite.Require().NoError(err)

	// Check for alert
	alerts := suite.keeper.GetValidatorAlerts(suite.ctx, validatorAddr)
	suite.Require().Greater(len(alerts), 0)
}

func (suite *KeeperTestSuite) TestTriggerFailover() {
	validatorAddr := "auravaloper1failover"
	backupAddr := "auravaloper1backup"

	// Setup main and backup validators
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, []string{backupAddr})
	suite.Require().NoError(err)
	err = suite.keeper.RegisterValidator(suite.ctx, backupAddr, "hot2", "cold2", "region", "US", 38.0, -122.0, nil)
	suite.Require().NoError(err)

	// Trigger failover
	err = suite.keeper.TriggerFailover(suite.ctx, validatorAddr)
	suite.Require().NoError(err)

	// Verify failover active
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().True(info.FailoverActive)
	suite.Require().Equal(backupAddr, info.ActiveBackup)
}

func (suite *KeeperTestSuite) TestTriggerFailoverAlreadyActive() {
	validatorAddr := "auravaloper1failactive"
	backupAddr := "auravaloper1backup2"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, []string{backupAddr})
	suite.Require().NoError(err)
	err = suite.keeper.RegisterValidator(suite.ctx, backupAddr, "hot2", "cold2", "region", "US", 38.0, -122.0, nil)
	suite.Require().NoError(err)

	// Activate failover
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	info.FailoverActive = true
	info.ActiveBackup = backupAddr
	suite.keeper.SetValidatorSecurityInfo(suite.ctx, info)

	// Try to trigger again - should succeed without error
	err = suite.keeper.TriggerFailover(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestTriggerFailoverNoBackups() {
	validatorAddr := "auravaloper1nobackup"

	// Setup without backups
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Try failover
	err = suite.keeper.TriggerFailover(suite.ctx, validatorAddr)
	suite.Require().Error(err)
	suite.Require().True(errors.Is(err, types.ErrNoBackupValidators))
}

func (suite *KeeperTestSuite) TestTriggerFailoverInvalidBackup() {
	validatorAddr := newValAddr()
	backupAddr := newValAddr()

	// Setup with jailed backup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, []string{backupAddr})
	suite.Require().NoError(err)
	err = suite.keeper.RegisterValidator(suite.ctx, backupAddr, "hot2", "cold2", "region", "US", 38.0, -122.0, nil)
	suite.Require().NoError(err)

	// Jail the backup
	err = suite.keeper.JailValidator(suite.ctx, backupAddr, time.Hour)
	suite.Require().NoError(err)

	// Try failover - should fail
	err = suite.keeper.TriggerFailover(suite.ctx, validatorAddr)
	suite.Require().Error(err)
	suite.Require().True(errors.Is(err, types.ErrInvalidBackupValidator))
}

func (suite *KeeperTestSuite) TestRestoreFromFailover() {
	validatorAddr := "auravaloper1restore"
	backupAddr := "auravaloper1backup3"

	// Setup
	params := types.DefaultParams()
	params.RequireSentryNodes = true
	params.MinSentryNodes = 1
	suite.Require().NoError(suite.keeper.SetParams(suite.ctx, params))

	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, []string{backupAddr})
	suite.Require().NoError(err)
	err = suite.keeper.RegisterValidator(suite.ctx, backupAddr, "hot2", "cold2", "region", "US", 38.0, -122.0, nil)
	suite.Require().NoError(err)

	// Register sentry
	err = suite.keeper.RegisterSentryNode(suite.ctx, validatorAddr, "sentry1", "192.168.1.1", 26656)
	suite.Require().NoError(err)

	// Activate failover
	err = suite.keeper.TriggerFailover(suite.ctx, validatorAddr)
	suite.Require().NoError(err)

	// Restore
	err = suite.keeper.RestoreFromFailover(suite.ctx, validatorAddr)
	suite.Require().NoError(err)

	// Verify restored
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().False(info.FailoverActive)
	suite.Require().Empty(info.ActiveBackup)
}

func (suite *KeeperTestSuite) TestRestoreFromFailoverNotActive() {
	validatorAddr := "auravaloper1notfailover"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Try to restore when not in failover - should succeed without error
	err = suite.keeper.RestoreFromFailover(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestRestoreFromFailoverInsufficientSentries() {
	validatorAddr := "auravaloper1restorefail"
	backupAddr := "auravaloper1backup4"

	// Setup
	params := types.DefaultParams()
	params.RequireSentryNodes = true
	params.MinSentryNodes = 2
	suite.Require().NoError(suite.keeper.SetParams(suite.ctx, params))

	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, []string{backupAddr})
	suite.Require().NoError(err)
	err = suite.keeper.RegisterValidator(suite.ctx, backupAddr, "hot2", "cold2", "region", "US", 38.0, -122.0, nil)
	suite.Require().NoError(err)

	// Activate failover
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	info.FailoverActive = true
	info.ActiveBackup = backupAddr
	suite.keeper.SetValidatorSecurityInfo(suite.ctx, info)

	// Try to restore without enough sentries - should fail
	err = suite.keeper.RestoreFromFailover(suite.ctx, validatorAddr)
	suite.Require().Error(err)
	suite.Require().True(errors.Is(err, types.ErrInsufficientSentryNodes))
}
