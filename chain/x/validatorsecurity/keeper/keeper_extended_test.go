package keeper_test

import (
	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

func (suite *KeeperTestSuite) TestHasValidatorSecurityInfo() {
	validatorAddr := "auravaloper1has"

	// Check non-existent validator
	has := suite.keeper.HasValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().False(has)

	// Register validator
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Check again
	has = suite.keeper.HasValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().True(has)
}

func (suite *KeeperTestSuite) TestSetValidatorSecurityInfo() {
	validatorAddr := "auravaloper1set"

	// Create info
	info := types.ValidatorSecurityInfo{
		ValidatorAddress: validatorAddr,
		HotKey:           "hot",
		ColdKey:          "cold",
		KeysSeparated:    true,
		Region:           "region",
		CountryCode:      "US",
		Latitude:         37.0,
		Longitude:        -122.0,
		IsJailed:         false,
		IsTombstoned:     false,
	}

	// Set directly
	suite.keeper.SetValidatorSecurityInfo(suite.ctx, info)

	// Retrieve
	retrieved, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().Equal(validatorAddr, retrieved.ValidatorAddress)
	suite.Require().Equal("hot", retrieved.HotKey)
}

func (suite *KeeperTestSuite) TestSetValidatorSecurityInfoJailedIndex() {
	validatorAddr := "auravaloper1jailedidx"

	// Create jailed info
	info := types.ValidatorSecurityInfo{
		ValidatorAddress: validatorAddr,
		HotKey:           "hot",
		ColdKey:          "cold",
		IsJailed:         true,
		IsTombstoned:     false,
	}

	// Set
	suite.keeper.SetValidatorSecurityInfo(suite.ctx, info)

	// Retrieve
	retrieved, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().True(retrieved.IsJailed)
}

func (suite *KeeperTestSuite) TestSetValidatorSecurityInfoTombstonedIndex() {
	validatorAddr := "auravaloper1tombidx"

	// Create tombstoned info
	info := types.ValidatorSecurityInfo{
		ValidatorAddress: validatorAddr,
		HotKey:           "hot",
		ColdKey:          "cold",
		IsJailed:         true,
		IsTombstoned:     true,
	}

	// Set
	suite.keeper.SetValidatorSecurityInfo(suite.ctx, info)

	// Retrieve
	retrieved, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().True(retrieved.IsTombstoned)
}

func (suite *KeeperTestSuite) TestGetValidatorSecurityInfoNotFound() {
	_, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, "nonexistent")
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrValidatorNotFound, err)
}

func (suite *KeeperTestSuite) TestSetSentryNodeInfo() {
	validatorAddr := "auravaloper1sentryinfo"

	// Setup validator
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Create sentry node info
	node := types.SentryNodeInfo{
		Address:          "sentry1",
		ValidatorAddress: validatorAddr,
		IpAddress:        "192.168.1.1",
		Port:             26656,
		IsActive:         true,
	}

	// Set directly
	suite.keeper.SetSentryNodeInfo(suite.ctx, node)

	// Retrieve
	retrieved, err := suite.keeper.GetSentryNodeInfo(suite.ctx, "sentry1")
	suite.Require().NoError(err)
	suite.Require().Equal("sentry1", retrieved.Address)
	suite.Require().Equal(validatorAddr, retrieved.ValidatorAddress)
}

func (suite *KeeperTestSuite) TestGetValidatorAlertsEmpty() {
	validatorAddr := "auravaloper1noalerts"

	// Setup validator
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Get alerts
	alerts := suite.keeper.GetValidatorAlerts(suite.ctx, validatorAddr)
	suite.Require().Empty(alerts)
}

func (suite *KeeperTestSuite) TestGetValidatorSentryNodesEmpty() {
	validatorAddr := "auravaloper1nosentries"

	// Setup validator
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Get sentry nodes
	nodes := suite.keeper.GetValidatorSentryNodes(suite.ctx, validatorAddr)
	suite.Require().Empty(nodes)
}

func (suite *KeeperTestSuite) TestRegisterValidatorWithBackups() {
	validatorAddr := "auravaloper1withbackup"
	backups := []string{"backup1", "backup2"}

	// Register with backups
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, backups)
	suite.Require().NoError(err)

	// Verify backups
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().Equal(backups, info.BackupValidatorAddresses)
}

func (suite *KeeperTestSuite) TestRegisterValidatorGeoDistribution() {
	params := types.DefaultParams()
	params.EnableGeoDistribution = true
	params.MaxValidatorsPerRegion = 2
	suite.Require().NoError(suite.keeper.SetParams(suite.ctx, params))

	// Register first validator
	err := suite.keeper.RegisterValidator(suite.ctx, "val1", "hot1", "cold1", "testregion", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Register second validator in same region
	err = suite.keeper.RegisterValidator(suite.ctx, "val2", "hot2", "cold2", "testregion", "US", 38.0, -122.0, nil)
	suite.Require().NoError(err)

	// Try to register third validator - should fail due to capacity
	err = suite.keeper.RegisterValidator(suite.ctx, "val3", "hot3", "cold3", "testregion", "US", 39.0, -122.0, nil)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrRegionCapacityExceeded, err)
}

func (suite *KeeperTestSuite) TestTrackBlockSignMissed() {
	validatorAddr := "auravaloper1trackmiss"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Track missed block
	err = suite.keeper.TrackBlockSign(suite.ctx, validatorAddr, false)
	suite.Require().NoError(err)

	// Verify
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(1), info.MissedBlocksCounter)
}

func (suite *KeeperTestSuite) TestTrackBlockSignTombstoned() {
	validatorAddr := newValAddr()

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Tombstone
	err = suite.keeper.TombstoneValidator(suite.ctx, validatorAddr)
	suite.Require().NoError(err)

	// Track block - should not error for tombstoned validator
	err = suite.keeper.TrackBlockSign(suite.ctx, validatorAddr, false)
	suite.Require().NoError(err)
}
