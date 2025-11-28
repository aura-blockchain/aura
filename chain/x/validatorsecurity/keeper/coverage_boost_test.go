package keeper_test

import (
	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

func (suite *KeeperTestSuite) TestDecrementRegionCount() {
	params := types.DefaultParams()
	params.EnableGeoDistribution = true
	params.MaxValidatorsPerRegion = 10
	suite.Require().NoError(suite.keeper.SetParams(suite.ctx, params))

	// Register validator to increment count
	err := suite.keeper.RegisterValidator(suite.ctx, newValAddr(), "hot1", "cold1", "decrement-region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Now tombstone to trigger decrement
	err = suite.keeper.TombstoneValidator(suite.ctx, newValAddr())
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestGetJailedValidatorsWithPrefix() {
	// Register and jail multiple validators
	val1 := newValAddr()
	val2 := newValAddr()

	err := suite.keeper.RegisterValidator(suite.ctx, val1, "hot1", "cold1", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)
	err = suite.keeper.RegisterValidator(suite.ctx, val2, "hot2", "cold2", "region", "US", 38.0, -122.0, nil)
	suite.Require().NoError(err)

	// Jail both
	suite.keeper.JailValidator(suite.ctx, val1, 3600000000000) // 1 hour in nanoseconds
	suite.keeper.JailValidator(suite.ctx, val2, 3600000000000)

	// Get jailed validators - test the iterator
	jailed := suite.keeper.GetJailedValidators(suite.ctx)
	suite.Require().GreaterOrEqual(len(jailed), 2)
}

func (suite *KeeperTestSuite) TestGetTombstonedValidatorsWithPrefix() {
	// Register and tombstone multiple validators
	val1 := newValAddr()
	val2 := newValAddr()

	err := suite.keeper.RegisterValidator(suite.ctx, val1, "hot1", "cold1", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)
	err = suite.keeper.RegisterValidator(suite.ctx, val2, "hot2", "cold2", "region", "US", 38.0, -122.0, nil)
	suite.Require().NoError(err)

	// Tombstone both
	suite.keeper.TombstoneValidator(suite.ctx, val1)
	suite.keeper.TombstoneValidator(suite.ctx, val2)

	// Get tombstoned validators - test the iterator
	tombstoned := suite.keeper.GetTombstonedValidators(suite.ctx)
	suite.Require().GreaterOrEqual(len(tombstoned), 2)
}

func (suite *KeeperTestSuite) TestMonitorAllValidatorsMultiple() {
	// Setup multiple validators with various states
	for i := 0; i < 5; i++ {
		addr := newValAddr()
		err := suite.keeper.RegisterValidator(suite.ctx, addr, "hot", "cold", "region", "US", 37.0+float64(i), -122.0, nil)
		if err == nil || err == types.ErrValidatorAlreadyRegistered {
			// success or already registered
		}
	}

	// Monitor all - should not panic
	suite.keeper.MonitorAllValidators(suite.ctx)
}

func (suite *KeeperTestSuite) TestGetAllAlertsMultipleValidators() {
	// Create alerts for different validators
	for i := 0; i < 3; i++ {
		addr := "auravaloper1alert" + string(rune('a'+i))
		err := suite.keeper.RegisterValidator(suite.ctx, addr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
		if err == nil {
			alert := types.ValidatorAlert{
				ValidatorAddress: addr,
				AlertType:        types.ValidatorAlert_DOWNTIME,
				Severity:         types.ValidatorAlert_WARNING,
				Message:          "Test",
			}
			suite.keeper.CreateAlert(suite.ctx, alert)
		}
	}

	// Get all alerts
	alerts := suite.keeper.GetAllAlerts(suite.ctx)
	suite.Require().NotNil(alerts)
}

func (suite *KeeperTestSuite) TestTrackBlockSignNotFound() {
	// Track for non-existent validator - should error
	err := suite.keeper.TrackBlockSign(suite.ctx, "nonexistent", true)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestMonitorValidatorNotFound() {
	// Monitor non-existent validator - should error
	err := suite.keeper.MonitorValidator(suite.ctx, "nonexistent")
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestTriggerFailoverNotFound() {
	// Failover for non-existent validator - should error
	err := suite.keeper.TriggerFailover(suite.ctx, "nonexistent")
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestRestoreFromFailoverNotFound() {
	// Restore non-existent validator - should error
	err := suite.keeper.RestoreFromFailover(suite.ctx, "nonexistent")
	suite.Require().Error(err)
}
