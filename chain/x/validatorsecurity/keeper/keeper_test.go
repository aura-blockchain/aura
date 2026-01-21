// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/validatorsecurity/keeper"
	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx            sdk.Context
	keeper         keeper.Keeper
	cdc            codec.Codec
	stakingKeeper  *MockStakingKeeper
	slashingKeeper *MockSlashingKeeper
	bankKeeper     *MockBankKeeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// MockValidator implements the Validator interface for testing
type MockValidator struct {
	operator sdk.ValAddress
	tokens   sdkmath.Int
	status   int32
}

func (m MockValidator) GetOperator() sdk.ValAddress           { return m.operator }
func (m MockValidator) GetConsPubKey() (interface{}, error)   { return nil, nil }
func (m MockValidator) GetConsAddr() (sdk.ConsAddress, error) { return nil, nil }
func (m MockValidator) GetStatus() int32                      { return m.status }
func (m MockValidator) GetTokens() sdkmath.Int                { return m.tokens }

// MockStakingKeeper implements a mock StakingKeeper for testing
type MockStakingKeeper struct {
	Validators map[string]bool
}

func NewMockStakingKeeper() *MockStakingKeeper {
	return &MockStakingKeeper{
		Validators: make(map[string]bool),
	}
}

func (m *MockStakingKeeper) Validator(ctx context.Context, addr sdk.ValAddress) (keeper.Validator, error) {
	return MockValidator{operator: addr, tokens: sdkmath.NewInt(1000000), status: 3}, nil
}

func (m *MockStakingKeeper) ValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (keeper.Validator, error) {
	return MockValidator{tokens: sdkmath.NewInt(1000000), status: 3}, nil
}

func (m *MockStakingKeeper) Slash(ctx context.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor sdkmath.LegacyDec) (sdkmath.Int, error) {
	return sdkmath.NewInt(100), nil
}

func (m *MockStakingKeeper) Jail(ctx context.Context, consAddr sdk.ConsAddress) error {
	m.Validators[consAddr.String()] = true
	return nil
}

func (m *MockStakingKeeper) Unjail(ctx context.Context, consAddr sdk.ConsAddress) error {
	m.Validators[consAddr.String()] = false
	return nil
}

func (m *MockStakingKeeper) GetAllValidators(ctx context.Context) ([]keeper.Validator, error) {
	return []keeper.Validator{}, nil
}

func (m *MockStakingKeeper) PowerReduction(ctx context.Context) sdkmath.Int {
	return sdkmath.NewInt(1000000)
}

// MockSlashingKeeper implements a mock SlashingKeeper for testing
type MockSlashingKeeper struct {
	Tombstoned map[string]bool
	JailTime   map[string]time.Time
}

func NewMockSlashingKeeper() *MockSlashingKeeper {
	return &MockSlashingKeeper{
		Tombstoned: make(map[string]bool),
		JailTime:   make(map[string]time.Time),
	}
}

func (m *MockSlashingKeeper) IsTombstoned(ctx context.Context, consAddr sdk.ConsAddress) bool {
	return m.Tombstoned[consAddr.String()]
}

func (m *MockSlashingKeeper) Tombstone(ctx context.Context, consAddr sdk.ConsAddress) error {
	m.Tombstoned[consAddr.String()] = true
	return nil
}

func (m *MockSlashingKeeper) JailUntil(ctx context.Context, consAddr sdk.ConsAddress, jailTime time.Time) error {
	m.JailTime[consAddr.String()] = jailTime
	return nil
}

// MockBankKeeper implements a mock BankKeeper for testing
type MockBankKeeper struct {
	Balances map[string]sdk.Coins
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		Balances: make(map[string]sdk.Coins),
	}
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if coins, ok := m.Balances[addr.String()]; ok {
		return sdk.NewCoin(denom, coins.AmountOf(denom))
	}
	return sdk.NewCoin(denom, sdkmath.ZeroInt())
}

func (suite *KeeperTestSuite) SetupTest() {
	// Setup store
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	// Create database and commit multi-store
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	suite.Require().NoError(cms.LoadLatestVersion())

	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create context with proper store
	header := tmproto.Header{
		Height: 1,
		Time:   time.Now(),
	}
	ctx := sdk.NewContext(cms, header, false, log.NewNopLogger())

	// Create mock keepers
	stakingKeeper := NewMockStakingKeeper()
	slashingKeeper := NewMockSlashingKeeper()
	bankKeeper := NewMockBankKeeper()

	// Create keeper with mocks
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		"authority",
		stakingKeeper,  // staking keeper mock
		slashingKeeper, // slashing keeper mock
		bankKeeper,     // bank keeper mock
	)

	suite.ctx = ctx
	suite.keeper = k
	suite.cdc = cdc
	suite.stakingKeeper = stakingKeeper
	suite.slashingKeeper = slashingKeeper
	suite.bankKeeper = bankKeeper
}

func (suite *KeeperTestSuite) TestParams() {
	// Test default params
	params := types.DefaultParams()
	suite.Require().NoError(types.ValidateParams(params))

	// Set params
	suite.Require().NoError(suite.keeper.SetParams(suite.ctx, params))

	// Get params
	retrievedParams := suite.keeper.GetParams(suite.ctx)

	// Compare field by field instead of using Equal (which compares internal proto fields)
	suite.Require().Equal(params.DoubleSignSlashFraction, retrievedParams.DoubleSignSlashFraction)
	suite.Require().Equal(params.DowntimeSlashFraction, retrievedParams.DowntimeSlashFraction)
	suite.Require().Equal(params.SignedBlocksWindow, retrievedParams.SignedBlocksWindow)
	suite.Require().Equal(params.MinSignedPerWindow, retrievedParams.MinSignedPerWindow)
	suite.Require().Equal(params.MinimumStakeAmount, retrievedParams.MinimumStakeAmount)
	suite.Require().Equal(params.EnableGeoDistribution, retrievedParams.EnableGeoDistribution)
	suite.Require().Equal(params.MaxValidatorsPerRegion, retrievedParams.MaxValidatorsPerRegion)
	suite.Require().Equal(params.RequireSentryNodes, retrievedParams.RequireSentryNodes)
	suite.Require().Equal(params.MinSentryNodes, retrievedParams.MinSentryNodes)
	suite.Require().Equal(params.EnableAutoFailover, retrievedParams.EnableAutoFailover)
}

func (suite *KeeperTestSuite) TestRegisterValidator() {
	validatorAddr := "auravaloper1test"

	// Register validator
	err := suite.keeper.RegisterValidator(
		suite.ctx,
		validatorAddr,
		"hot_key_123",
		"cold_key_456",
		"us-west",
		"US",
		37.7749,
		-122.4194,
		[]string{"backup1", "backup2"},
	)
	suite.Require().NoError(err)

	// Verify validator was registered
	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().Equal(validatorAddr, info.ValidatorAddress)
	suite.Require().Equal("hot_key_123", info.HotKey)
	suite.Require().Equal("cold_key_456", info.ColdKey)
	suite.Require().True(info.KeysSeparated)
	suite.Require().Equal("us-west", info.Region)
	suite.Require().Equal("US", info.CountryCode)
	suite.Require().False(info.IsJailed)
	suite.Require().False(info.IsTombstoned)
}

func (suite *KeeperTestSuite) TestRegisterValidatorDuplicate() {
	validatorAddr := "auravaloper1test"

	// Register validator first time
	err := suite.keeper.RegisterValidator(
		suite.ctx,
		validatorAddr,
		"hot_key",
		"cold_key",
		"us-west",
		"US",
		37.0,
		-122.0,
		nil,
	)
	suite.Require().NoError(err)

	// Try to register again
	err = suite.keeper.RegisterValidator(
		suite.ctx,
		validatorAddr,
		"hot_key_2",
		"cold_key_2",
		"us-east",
		"US",
		40.0,
		-74.0,
		nil,
	)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrValidatorAlreadyRegistered, err)
}

func (suite *KeeperTestSuite) TestValidatorKeySeparation() {
	validatorAddr := "auravaloper1test"

	// Register with same hot and cold key (not separated)
	err := suite.keeper.RegisterValidator(
		suite.ctx,
		validatorAddr,
		"same_key",
		"same_key",
		"us-west",
		"US",
		37.0,
		-122.0,
		nil,
	)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestGeographicValidation() {
	// Test invalid latitude
	err := suite.keeper.RegisterValidator(
		suite.ctx,
		"val1",
		"hot",
		"cold",
		"region",
		"US",
		91.0, // Invalid
		0.0,
		nil,
	)
	suite.Require().Error(err)

	// Test invalid longitude
	err = suite.keeper.RegisterValidator(
		suite.ctx,
		"val2",
		"hot",
		"cold",
		"region",
		"US",
		0.0,
		181.0, // Invalid
		nil,
	)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestRegisterSentryNode() {
	validatorAddr := "auravaloper1test"

	// Register validator first
	err := suite.keeper.RegisterValidator(
		suite.ctx,
		validatorAddr,
		"hot",
		"cold",
		"region",
		"US",
		37.0,
		-122.0,
		nil,
	)
	suite.Require().NoError(err)

	// Register sentry node
	err = suite.keeper.RegisterSentryNode(
		suite.ctx,
		validatorAddr,
		"sentry1",
		"192.168.1.1",
		26656,
	)
	suite.Require().NoError(err)

	// Verify sentry node was registered
	nodes := suite.keeper.GetValidatorSentryNodes(suite.ctx, validatorAddr)
	suite.Require().Len(nodes, 1)
	suite.Require().Equal("sentry1", nodes[0].Address)
	suite.Require().Equal("192.168.1.1", nodes[0].IpAddress)
	suite.Require().Equal(int32(26656), nodes[0].Port)
	suite.Require().True(nodes[0].IsActive)
}

func (suite *KeeperTestSuite) TestDoubleSignEvidence() {
	validatorAddr := "auravaloper1test"
	now := time.Now()

	evidence := types.DoubleSignEvidence{
		ValidatorAddress: validatorAddr,
		Height:           100,
		Time:             &now,
		VoteA:            []byte("vote_a"),
		VoteB:            []byte("vote_b"),
		SlashFraction:    sdkmath.LegacyMustNewDecFromStr("0.05"),
	}

	suite.keeper.SetDoubleSignEvidence(suite.ctx, evidence)

	// Retrieve evidence
	retrieved, err := suite.keeper.GetDoubleSignEvidence(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().Equal(validatorAddr, retrieved.ValidatorAddress)
	suite.Require().Equal(int64(100), retrieved.Height)
}

func (suite *KeeperTestSuite) TestDowntimeInfraction() {
	validatorAddr := "auravaloper1test"
	now := time.Now()

	infraction := types.DowntimeInfraction{
		ValidatorAddress: validatorAddr,
		MissedBlocks:     100,
		WindowSize:       1000,
		DetectedAt:       &now,
		SlashFraction:    sdkmath.LegacyMustNewDecFromStr("0.0001"),
	}

	suite.keeper.SetDowntimeInfraction(suite.ctx, infraction)

	// Retrieve infraction
	retrieved, err := suite.keeper.GetDowntimeInfraction(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().Equal(validatorAddr, retrieved.ValidatorAddress)
	suite.Require().Equal(int64(100), retrieved.MissedBlocks)
}

func (suite *KeeperTestSuite) TestCreateAlert() {
	now := time.Now()
	alert := types.ValidatorAlert{
		Id:               "test-alert-1",
		ValidatorAddress: "auravaloper1test",
		AlertType:        types.ValidatorAlert_DOWNTIME,
		Severity:         types.ValidatorAlert_WARNING,
		Message:          "Test alert message",
		Timestamp:        &now,
		Acknowledged:     false,
	}

	suite.keeper.CreateAlert(suite.ctx, alert)

	// Retrieve alerts
	alerts := suite.keeper.GetValidatorAlerts(suite.ctx, "auravaloper1test")
	suite.Require().Len(alerts, 1)
	suite.Require().Equal("test-alert-1", alerts[0].Id)
	suite.Require().Equal(types.ValidatorAlert_DOWNTIME, alerts[0].AlertType)
}

func (suite *KeeperTestSuite) TestAcknowledgeAlert() {
	now := time.Now()
	alert := types.ValidatorAlert{
		Id:               "alert-1",
		ValidatorAddress: "val1",
		AlertType:        types.ValidatorAlert_DOWNTIME,
		Severity:         types.ValidatorAlert_WARNING,
		Message:          "Test",
		Timestamp:        &now,
		Acknowledged:     false,
	}

	suite.keeper.CreateAlert(suite.ctx, alert)

	// Acknowledge alert
	err := suite.keeper.AcknowledgeAlert(suite.ctx, "alert-1", "acknowledger1")
	suite.Require().NoError(err)

	// Verify acknowledged
	alerts := suite.keeper.GetValidatorAlerts(suite.ctx, "val1")
	suite.Require().Len(alerts, 1)
	suite.Require().True(alerts[0].Acknowledged)
	suite.Require().Equal("acknowledger1", alerts[0].AcknowledgedBy)
}

func (suite *KeeperTestSuite) TestSentryHeartbeat() {
	validatorAddr := "auravaloper1test"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	err = suite.keeper.RegisterSentryNode(suite.ctx, validatorAddr, "sentry1", "192.168.1.1", 26656)
	suite.Require().NoError(err)

	// Update heartbeat
	err = suite.keeper.UpdateSentryHeartbeat(suite.ctx, "sentry1")
	suite.Require().NoError(err)

	// Verify
	node, err := suite.keeper.GetSentryNodeInfo(suite.ctx, "sentry1")
	suite.Require().NoError(err)
	suite.Require().True(node.IsActive)
	suite.Require().NotNil(node.LastHeartbeat)
}

func (suite *KeeperTestSuite) TestBlockSignTracking() {
	validatorAddr := "auravaloper1test"

	// Setup
	err := suite.keeper.RegisterValidator(suite.ctx, validatorAddr, "hot", "cold", "region", "US", 37.0, -122.0, nil)
	suite.Require().NoError(err)

	// Track signed block
	err = suite.keeper.TrackBlockSign(suite.ctx, validatorAddr, true)
	suite.Require().NoError(err)

	info, err := suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(1), info.IndexOffset)

	// Track missed block
	err = suite.keeper.TrackBlockSign(suite.ctx, validatorAddr, false)
	suite.Require().NoError(err)

	info, err = suite.keeper.GetValidatorSecurityInfo(suite.ctx, validatorAddr)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(2), info.IndexOffset)
	suite.Require().Equal(int64(1), info.MissedBlocksCounter)
}
