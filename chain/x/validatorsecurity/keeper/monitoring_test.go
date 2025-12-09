package keeper_test

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

func (suite *KeeperTestSuite) TestMonitorValidator() {
	validatorAddr := "auravaloper1monitor"

	// Setup
	params := types.DefaultParams()
	params.RequireSentryNodes = true
	params.MinSentryNodes = 2
	suite.Require().NoError(suite.keeper.SetParams(suite.ctx, params))

	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Monitor validator
	err = suite.keeper.MonitorValidator(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestMonitorAllValidators() {
	// Setup multiple validators
	for i := 0; i < 3; i++ {
		addr := fmt.Sprintf("auravaloper1mon%d", i)
		err := suite.keeper.RegisterValidator(suite.ctx, addr, fmt.Sprintf("hot%d", i), fmt.Sprintf("cold%d", i), "region", "US", 37.0+float64(i), -122.0, nil)
		suite.Require().NoError(err)
	}

	// Monitor all
	suite.keeper.MonitorAllValidators(suite.ctx)

	// Verify alerts were created (if any)
	alerts := suite.keeper.GetAllAlerts(suite.ctx)
	suite.Require().NotNil(alerts)
}

func (suite *KeeperTestSuite) TestGetAllAlerts() {
	validatorAddr := "auravaloper1alerts"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Create multiple alerts
	for i := 0; i < 3; i++ {
		now := time.Now()
		alert := types.ValidatorAlert{
			Id:               fmt.Sprintf("alert-%d", i),
			ValidatorAddress: validatorAddr,
			AlertType:        types.ValidatorAlert_DOWNTIME,
			Severity:         types.ValidatorAlert_WARNING,
			Message:          fmt.Sprintf("Test alert %d", i),
			Timestamp:        &now,
			Acknowledged:     false,
		}
		suite.keeper.CreateAlert(suite.ctx, alert)
	}

	// Get all alerts
	alerts := suite.keeper.GetAllAlerts(suite.ctx)
	suite.Require().GreaterOrEqual(len(alerts), 3)
}

func (suite *KeeperTestSuite) TestMonitorValidatorInactive() {
	validatorAddr := "auravaloper1inactive"

	// Setup
	params := types.DefaultParams()
	// MonitoringInterval is a time.Duration with (gogoproto.stdduration) = true
	params.MonitoringInterval = 60 * time.Second
	suite.Require().NoError(suite.keeper.SetParams(suite.ctx, params))

	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Set last seen to old time
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	oldTime := time.Now().Add(-5 * time.Minute)
	info.LastSeen = &oldTime
	suite.keeper.SetValidatorSecurityInfo(suite.ctx, info)

	// Monitor - should create alert
	err = suite.keeper.MonitorValidator(suite.ctx, validatorAddr)
	suite.Require().NoError(err)

	// Check for alert
	alerts := suite.keeper.GetValidatorAlerts(suite.ctx, validatorAddr)
	suite.Require().Greater(len(alerts), 0)
}

func (suite *KeeperTestSuite) TestMonitorValidatorSentryNodes() {
	validatorAddr := "auravaloper1sentry"

	// Setup
	params := types.DefaultParams()
	params.RequireSentryNodes = true
	params.MinSentryNodes = 2
	params.MonitoringInterval = 60 * time.Second
	suite.Require().NoError(suite.keeper.SetParams(suite.ctx, params))

	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Register sentry nodes
	err = suite.keeper.RegisterSentryNode(suite.ctx, validatorAddr, "sentry1", "192.168.1.1", 26656)
	suite.Require().NoError(err)
	err = suite.keeper.RegisterSentryNode(suite.ctx, validatorAddr, "sentry2", "192.168.1.2", 26656)
	suite.Require().NoError(err)

	// Make one sentry offline by setting old heartbeat
	node, err := suite.keeper.GetSentryNodeInfo(suite.ctx, "sentry1")
	suite.Require().NoError(err)
	oldTime := time.Now().Add(-5 * time.Minute)
	node.LastHeartbeat = &oldTime
	suite.keeper.SetSentryNodeInfo(suite.ctx, node)

	// Monitor - should create alert about offline sentry
	err = suite.keeper.MonitorValidator(suite.ctx, validatorAddr)
	suite.Require().NoError(err)

	alerts := suite.keeper.GetValidatorAlerts(suite.ctx, validatorAddr)
	suite.Require().Greater(len(alerts), 0)
}

func (suite *KeeperTestSuite) TestMonitorValidatorGeoDistribution() {
	_ = "auravaloper1geo" // will use val1 instead

	// Setup with geo distribution enabled
	params := types.DefaultParams()
	params.EnableGeoDistribution = true
	params.MaxValidatorsPerRegion = 1
	suite.Require().NoError(suite.keeper.SetParams(suite.ctx, params))

	// Register two validators in same region
	err := suite.keeper.RegisterValidator(suite.ctx, "val1", "hot1", "cold1", "region1", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Monitor - may create geo alert if region is over capacity
	err = suite.keeper.MonitorValidator(suite.ctx, "val1")
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestMonitorValidatorFailoverActive() {
	validatorAddr := "auravaloper1failover"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, []string{"backup1"})
	suite.Require().NoError(err)

	// Activate failover
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	info.FailoverActive = true
	info.ActiveBackup = "backup1"
	suite.keeper.SetValidatorSecurityInfo(suite.ctx, info)

	// Monitor - should create failover alert
	err = suite.keeper.MonitorValidator(suite.ctx, validatorAddr)
	suite.Require().NoError(err)

	alerts := suite.keeper.GetValidatorAlerts(suite.ctx, validatorAddr)
	suite.Require().Greater(len(alerts), 0)
}

func (suite *KeeperTestSuite) TestCreateAlertAutoID() {
	validatorAddr := "auravaloper1autoid"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Create alert without ID
	now := time.Now()
	alert := types.ValidatorAlert{
		Id:               "", // Empty ID
		ValidatorAddress: validatorAddr,
		AlertType:        types.ValidatorAlert_DOWNTIME,
		Severity:         types.ValidatorAlert_WARNING,
		Message:          "Test auto ID",
		Timestamp:        &now,
		Acknowledged:     false,
	}
	suite.keeper.CreateAlert(suite.ctx, alert)

	// Verify alert was created with auto-generated ID
	alerts := suite.keeper.GetValidatorAlerts(suite.ctx, validatorAddr)
	suite.Require().Greater(len(alerts), 0)
	suite.Require().NotEmpty(alerts[0].Id)
}

func (suite *KeeperTestSuite) TestAcknowledgeAlertNotFound() {
	err := suite.keeper.AcknowledgeAlert(suite.ctx, "nonexistent-alert", "acknowledger")
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrAlertNotFound, err)
}

func (suite *KeeperTestSuite) TestGetAllValidators() {
	// Setup multiple validators
	for i := 0; i < 3; i++ {
		addr := fmt.Sprintf("auravaloper1all%d", i)
		err := suite.keeper.RegisterValidator(suite.ctx, addr, fmt.Sprintf("hot%d", i), fmt.Sprintf("cold%d", i), "region", "US", 37.0+float64(i), -122.0, nil)
		suite.Require().NoError(err)
	}

	// Get all validators
	validators := suite.keeper.GetAllValidators(suite.ctx)
	suite.Require().GreaterOrEqual(len(validators), 3)
}

func (suite *KeeperTestSuite) TestGetAuthority() {
	authority := suite.keeper.GetAuthority()
	suite.Require().Equal("authority", authority)
}
