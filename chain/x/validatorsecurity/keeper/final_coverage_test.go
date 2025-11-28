package keeper_test

import (
	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

func (suite *KeeperTestSuite) TestRegisterValidatorNoBackups() {
	validatorAddr := "auravaloper1nobackup"

	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().Empty(info.BackupValidatorAddresses)
}

func (suite *KeeperTestSuite) TestRegisterValidatorEmptyKeys() {
	validatorAddr := "auravaloper1emptykeys"

	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "", "", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().False(info.KeysSeparated)
}

func (suite *KeeperTestSuite) TestTrackBlockSignMultipleTimes() {
	validatorAddr := "auravaloper1multitrack"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Track multiple blocks
	for i := 0; i < 10; i++ {
		signed := (i%2 == 0) // Alternate between signed and missed
		err = suite.keeper.TrackBlockSign(suite.ctx, validatorAddr, signed)
		suite.Require().NoError(err)
	}

	// Verify state updated
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(10), info.IndexOffset)
}

func (suite *KeeperTestSuite) TestCreateMultipleAlerts() {
	validatorAddr := "auravaloper1manyalerts"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Create multiple alert types
	alertTypes := []types.ValidatorAlert_AlertType{
		types.ValidatorAlert_DOWNTIME,
		types.ValidatorAlert_LOW_STAKE,
		types.ValidatorAlert_SENTRY_NODE_OFFLINE,
	}

	for _, alertType := range alertTypes {
		alert := types.ValidatorAlert{
			ValidatorAddress: validatorAddr,
			AlertType:        alertType,
			Severity:         types.ValidatorAlert_WARNING,
			Message:          "Test alert",
		}
		suite.keeper.CreateAlert(suite.ctx, alert)
	}

	alerts := suite.keeper.GetValidatorAlerts(suite.ctx, validatorAddr)
	suite.Require().GreaterOrEqual(len(alerts), 3)
}

func (suite *KeeperTestSuite) TestMultipleSentryNodes() {
	validatorAddr := "auravaloper1multisentries"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Register multiple sentry nodes
	for i := 0; i < 5; i++ {
		sentryAddr := "sentry" + string(rune('a'+i))
		err = suite.keeper.RegisterSentryNode(suite.ctx, validatorAddr, sentryAddr, "192.168.1."+string(rune('1'+i)), 26656)
		suite.Require().NoError(err)
	}

	nodes := suite.keeper.GetValidatorSentryNodes(suite.ctx, validatorAddr)
	suite.Require().GreaterOrEqual(len(nodes), 5)
}

func (suite *KeeperTestSuite) TestRecordMultipleSentryRequests() {
	validatorAddr := "auravaloper1sentryreqs"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	err = suite.keeper.RegisterSentryNode(suite.ctx, validatorAddr, "sentry_req", "192.168.1.1", 26656)
	suite.Require().NoError(err)

	// Record multiple requests
	for i := 0; i < 10; i++ {
		blocked := (i%3 == 0)
		err = suite.keeper.RecordSentryRequest(suite.ctx, "sentry_req", blocked)
		suite.Require().NoError(err)
	}

	node, err := suite.keeper.GetSentryNodeInfo(suite.ctx, "sentry_req")
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(10), node.RequestCount)
	suite.Require().Greater(node.BlockedRequests, uint64(0))
}

func (suite *KeeperTestSuite) TestGetParamsDefault() {
	params := suite.keeper.GetParams(suite.ctx)
	suite.Require().NotNil(params)
	suite.Require().NotEmpty(params.DoubleSignSlashFraction)
}

func (suite *KeeperTestSuite) TestMultipleValidatorRegistrations() {
	// Register many validators to test iteration
	for i := 0; i < 10; i++ {
		addr := "auravaloper1bulk" + string(rune('a'+i))
		err := suite.keeper.RegisterValidator(suite.ctx, addr, "hot", "cold", "region"+string(rune('0'+i)), "US", 37.0+float64(i), -122.0, nil)
		if err == nil || err == types.ErrValidatorAlreadyRegistered {
			// Success or already registered
		}
	}

	validators := suite.keeper.GetAllValidators(suite.ctx)
	suite.Require().GreaterOrEqual(len(validators), 10)
}

func (suite *KeeperTestSuite) TestAlertAcknowledgmentMultiple() {
	validatorAddr := "auravaloper1acktest"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Create and acknowledge multiple alerts
	for i := 0; i < 5; i++ {
		alertId := "ack-alert-" + string(rune('a'+i))
		alert := types.ValidatorAlert{
			Id:               alertId,
			ValidatorAddress: validatorAddr,
			AlertType:        types.ValidatorAlert_DOWNTIME,
			Severity:         types.ValidatorAlert_INFO,
			Message:          "Test",
		}
		suite.keeper.CreateAlert(suite.ctx, alert)

		err = suite.keeper.AcknowledgeAlert(suite.ctx, alertId, "acknowledger")
		suite.Require().NoError(err)
	}

	alerts := suite.keeper.GetValidatorAlerts(suite.ctx, validatorAddr)
	acknowledgedCount := 0
	for _, alert := range alerts {
		if alert.Acknowledged {
			acknowledgedCount++
		}
	}
	suite.Require().GreaterOrEqual(acknowledgedCount, 5)
}
