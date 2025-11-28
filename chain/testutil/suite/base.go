package suite

import (
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/aequitas/aura/chain/testutil/keeper"
	"github.com/aequitas/aura/chain/testutil/testdata"
)

// BaseTestSuite provides common setup for all test suites
type BaseTestSuite struct {
	suite.Suite

	App *keeper.TestApp
	Ctx sdk.Context
}

// SetupTest initializes the test suite before each test
func (s *BaseTestSuite) SetupTest() {
	s.App, s.Ctx = keeper.SetupTestApp(s.T())
}

// TearDownTest cleans up after each test
func (s *BaseTestSuite) TearDownTest() {
	// Cleanup if needed
}

// SetupSuite runs once before all tests in the suite
func (s *BaseTestSuite) SetupSuite() {
	// One-time setup
}

// TearDownSuite runs once after all tests in the suite
func (s *BaseTestSuite) TearDownSuite() {
	// One-time cleanup
}

// Helper methods

// CreateTestAccounts creates n test accounts
func (s *BaseTestSuite) CreateTestAccounts(count int) []sdk.AccAddress {
	return keeper.CreateTestAccounts(s.T(), count)
}

// CreateTestAccountsWithBalances creates accounts with initial balances
func (s *BaseTestSuite) CreateTestAccountsWithBalances(count int, initialBalance sdk.Coins) ([]sdk.AccAddress, map[string]sdk.Coins) {
	return keeper.CreateTestAccountsWithBalances(s.T(), count, initialBalance)
}

// GenTestAddress generates a random test address
func (s *BaseTestSuite) GenTestAddress() sdk.AccAddress {
	return keeper.GenTestAddress()
}

// GenTestValidatorAddress generates a random validator address
func (s *BaseTestSuite) GenTestValidatorAddress() sdk.ValAddress {
	return keeper.GenTestValidatorAddress()
}

// AdvanceBlockHeight advances the block height by n blocks
func (s *BaseTestSuite) AdvanceBlockHeight(n int64) {
	header := s.Ctx.BlockHeader()
	header.Height += n
	s.Ctx = s.Ctx.WithBlockHeader(header)
}

// AdvanceBlockTime advances the block time by duration
func (s *BaseTestSuite) AdvanceBlockTime(seconds int64) {
	header := s.Ctx.BlockHeader()
	header.Time = header.Time.Add(time.Duration(seconds) * time.Second)
	s.Ctx = s.Ctx.WithBlockHeader(header)
}

// GetTestCoins returns standard test coins
func (s *BaseTestSuite) GetTestCoins(amount int64, denom string) sdk.Coins {
	return keeper.CreateTestCoins(amount, denom)
}

// GetMultipleTestCoins returns multiple denominations
func (s *BaseTestSuite) GetMultipleTestCoins(amounts map[string]int64) sdk.Coins {
	return keeper.CreateMultipleTestCoins(amounts)
}

// RequireNoError asserts no error occurred
func (s *BaseTestSuite) RequireNoError(err error, msgAndArgs ...interface{}) {
	s.Require().NoError(err, msgAndArgs...)
}

// RequireError asserts an error occurred
func (s *BaseTestSuite) RequireError(err error, msgAndArgs ...interface{}) {
	s.Require().Error(err, msgAndArgs...)
}

// RequireEqual asserts two values are equal
func (s *BaseTestSuite) RequireEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	s.Require().Equal(expected, actual, msgAndArgs...)
}

// RequireNotEqual asserts two values are not equal
func (s *BaseTestSuite) RequireNotEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	s.Require().NotEqual(expected, actual, msgAndArgs...)
}

// RequireTrue asserts the value is true
func (s *BaseTestSuite) RequireTrue(value bool, msgAndArgs ...interface{}) {
	s.Require().True(value, msgAndArgs...)
}

// RequireFalse asserts the value is false
func (s *BaseTestSuite) RequireFalse(value bool, msgAndArgs ...interface{}) {
	s.Require().False(value, msgAndArgs...)
}

// RequireNil asserts the value is nil
func (s *BaseTestSuite) RequireNil(object interface{}, msgAndArgs ...interface{}) {
	s.Require().Nil(object, msgAndArgs...)
}

// RequireNotNil asserts the value is not nil
func (s *BaseTestSuite) RequireNotNil(object interface{}, msgAndArgs ...interface{}) {
	s.Require().NotNil(object, msgAndArgs...)
}

// GetContext returns the current context
func (s *BaseTestSuite) GetContext() sdk.Context {
	return s.Ctx
}

// UpdateContext updates the context
func (s *BaseTestSuite) UpdateContext(ctx sdk.Context) {
	s.Ctx = ctx
}

// UseTestAddresses returns predefined test addresses
func (s *BaseTestSuite) UseTestAddresses() []sdk.AccAddress {
	return []sdk.AccAddress{
		testdata.TestAddr1,
		testdata.TestAddr2,
		testdata.TestAddr3,
		testdata.TestAddr4,
		testdata.TestAddr5,
	}
}

// UseTestValidatorAddresses returns predefined test validator addresses
func (s *BaseTestSuite) UseTestValidatorAddresses() []sdk.ValAddress {
	return []sdk.ValAddress{
		testdata.TestValAddr1,
		testdata.TestValAddr2,
		testdata.TestValAddr3,
		testdata.TestValAddr4,
	}
}

// ValidatorTestSuite extends BaseTestSuite for validator-related tests
type ValidatorTestSuite struct {
	BaseTestSuite
	Validators []sdk.ValAddress
}

// SetupTest initializes the validator test suite
func (s *ValidatorTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()
	s.Validators = s.UseTestValidatorAddresses()
}

// ModuleTestSuite extends BaseTestSuite for module-specific tests
type ModuleTestSuite struct {
	BaseTestSuite
	ModuleName string
}

// SetupTest initializes the module test suite
func (s *ModuleTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()
}

// IntegrationTestSuite extends BaseTestSuite for integration tests
type IntegrationTestSuite struct {
	BaseTestSuite
	TestAccounts []sdk.AccAddress
}

// SetupTest initializes the integration test suite
func (s *IntegrationTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()
	s.TestAccounts = s.CreateTestAccounts(5)
}

// SetupTestWithAccounts initializes with specific number of accounts
func (s *IntegrationTestSuite) SetupTestWithAccounts(count int) {
	s.BaseTestSuite.SetupTest()
	s.TestAccounts = s.CreateTestAccounts(count)
}
