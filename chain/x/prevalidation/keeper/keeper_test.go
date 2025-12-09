package keeper_test

import (
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/prevalidation/keeper"
	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

type PrevalidationKeeperTestSuite struct {
	suite.Suite

	keeper *keeper.Keeper
	ctx    sdk.Context
}

func (suite *PrevalidationKeeperTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.keeper = keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
	)
	suite.ctx = input.Ctx
}

func TestPrevalidationKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(PrevalidationKeeperTestSuite))
}

// Params Tests

func (suite *PrevalidationKeeperTestSuite) TestGetParams() {
	params, err := suite.keeper.GetParams(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().NotNil(params)
}

func (suite *PrevalidationKeeperTestSuite) TestSetParams() {
	params := types.DefaultParams()
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	retrieved, err := suite.keeper.GetParams(suite.ctx)
	suite.Require().NoError(err)

	// Compare key fields directly (gogoproto types have internal cache fields that differ)
	suite.Require().Equal(params.Enabled, retrieved.Enabled)
	suite.Require().Equal(params.MaxCacheSize, retrieved.MaxCacheSize)
	suite.Require().Equal(params.ExpiryHours, retrieved.ExpiryHours)
	suite.Require().Equal(params.EncryptionAlgorithm, retrieved.EncryptionAlgorithm)
	suite.Require().Equal(params.ControlGroupPercentage, retrieved.ControlGroupPercentage)
	suite.Require().Equal(params.MinConfidenceScore, retrieved.MinConfidenceScore)
}

// Transaction Validation Tests

func TestValidateTransaction(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	tx := types.Transaction{
		Sender:    keepertest.GenTestAddr().String(),
		Recipient: keepertest.GenTestAddr().String(),
		Amount:    "1000",
		Nonce:     1,
	}

	valid, err := k.ValidateTransaction(input.Ctx, tx)
	require.NoError(t, err)
	require.True(t, valid)
}

func TestValidateTransactionInvalidSender(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	tx := types.Transaction{
		Sender:    "", // Invalid
		Recipient: keepertest.GenTestAddr().String(),
		Amount:    "1000",
		Nonce:     1,
	}

	valid, err := k.ValidateTransaction(input.Ctx, tx)
	require.Error(t, err)
	require.False(t, valid)
}

func TestValidateTransactionInvalidNonce(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	sender := keepertest.GenTestAddr().String()

	// Set current nonce
	k.SetNonce(input.Ctx, sender, 5)

	tx := types.Transaction{
		Sender:    sender,
		Recipient: keepertest.GenTestAddr().String(),
		Amount:    "1000",
		Nonce:     3, // Old nonce
	}

	valid, err := k.ValidateTransaction(input.Ctx, tx)
	require.Error(t, err)
	require.False(t, valid)
}

// Nonce Management Tests

func TestGetNonce(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	addr := keepertest.GenTestAddr().String()

	nonce := k.GetNonce(input.Ctx, addr)
	require.Equal(t, uint64(0), nonce)
}

func TestSetNonce(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	addr := keepertest.GenTestAddr().String()

	k.SetNonce(input.Ctx, addr, 5)

	nonce := k.GetNonce(input.Ctx, addr)
	require.Equal(t, uint64(5), nonce)
}

func TestIncrementNonce(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	addr := keepertest.GenTestAddr().String()

	k.SetNonce(input.Ctx, addr, 5)
	k.IncrementNonce(input.Ctx, addr)

	nonce := k.GetNonce(input.Ctx, addr)
	require.Equal(t, uint64(6), nonce)
}

// Balance Check Tests

func TestCheckSufficientBalance(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	addr := keepertest.GenTestAddr().String()
	amount := "1000"

	// Assume balance check passes
	sufficient := k.CheckSufficientBalance(input.Ctx, addr, amount)
	require.True(t, sufficient)
}

func TestCheckInsufficientBalance(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	addr := keepertest.GenTestAddr().String()
	largeAmount := "999999999999999"

	sufficient := k.CheckSufficientBalance(input.Ctx, addr, largeAmount)
	require.False(t, sufficient)
}

// Signature Validation Tests

func TestValidateSignature(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	signer := keepertest.GenTestAddr().String()
	message := []byte("test_message")
	signature := []byte("valid_signature")

	valid := k.ValidateSignature(input.Ctx, signer, message, signature)
	require.True(t, valid)
}

func TestValidateInvalidSignature(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	signer := keepertest.GenTestAddr().String()
	message := []byte("test_message")
	invalidSignature := []byte("invalid_signature")

	valid := k.ValidateSignature(input.Ctx, signer, message, invalidSignature)
	require.False(t, valid)
}

// Gas Estimation Tests

func TestEstimateGas(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	tx := types.Transaction{
		Sender:    keepertest.GenTestAddr().String(),
		Recipient: keepertest.GenTestAddr().String(),
		Amount:    "1000",
		Nonce:     1,
	}

	gas := k.EstimateGas(input.Ctx, tx)
	require.Greater(t, gas, uint64(0))
}

func TestEstimateGasComplexTransaction(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	tx := types.Transaction{
		Sender:    keepertest.GenTestAddr().String(),
		Recipient: keepertest.GenTestAddr().String(),
		Amount:    "1000000",
		Data:      []byte("complex_data"),
		Nonce:     1,
	}

	gas := k.EstimateGas(input.Ctx, tx)
	require.Greater(t, gas, uint64(21000)) // Base + data cost
}

// Mempool Validation Tests

func TestAddToMempool(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	tx := types.Transaction{
		Sender:    keepertest.GenTestAddr().String(),
		Recipient: keepertest.GenTestAddr().String(),
		Amount:    "1000",
		Nonce:     1,
	}

	err := k.AddToMempool(input.Ctx, tx)
	require.NoError(t, err)
}

func TestAddDuplicateToMempool(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	tx := types.Transaction{
		Sender:    keepertest.GenTestAddr().String(),
		Recipient: keepertest.GenTestAddr().String(),
		Amount:    "1000",
		Nonce:     1,
	}

	err := k.AddToMempool(input.Ctx, tx)
	require.NoError(t, err)

	// Try to add again
	err = k.AddToMempool(input.Ctx, tx)
	require.Error(t, err, "Should not allow duplicate in mempool")
}

func TestGetMempoolTransactions(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	// Add multiple transactions
	for i := 0; i < 5; i++ {
		tx := types.Transaction{
			Sender:    keepertest.GenTestAddr().String(),
			Recipient: keepertest.GenTestAddr().String(),
			Amount:    "1000",
			Nonce:     uint64(i + 1),
		}
		err := k.AddToMempool(input.Ctx, tx)
		require.NoError(t, err)
	}

	txs := k.GetMempoolTransactions(input.Ctx)
	require.GreaterOrEqual(t, len(txs), 5)
}

func TestRemoveFromMempool(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	tx := types.Transaction{
		Sender:    keepertest.GenTestAddr().String(),
		Recipient: keepertest.GenTestAddr().String(),
		Amount:    "1000",
		Nonce:     1,
	}

	err := k.AddToMempool(input.Ctx, tx)
	require.NoError(t, err)

	txHash := k.GetTransactionHash(tx)
	k.RemoveFromMempool(input.Ctx, txHash)

	txs := k.GetMempoolTransactions(input.Ctx)
	for _, mempoolTx := range txs {
		require.NotEqual(t, txHash, k.GetTransactionHash(mempoolTx))
	}
}

// Priority Validation Tests

func TestCalculatePriority(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	tx := types.Transaction{
		Sender:    keepertest.GenTestAddr().String(),
		Recipient: keepertest.GenTestAddr().String(),
		Amount:    "1000",
		GasPrice:  "100",
		Nonce:     1,
	}

	priority := k.CalculatePriority(input.Ctx, tx)
	require.Greater(t, priority, uint64(0))
}

func TestHighPriorityTransaction(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	highPriorityTx := types.Transaction{
		Sender:    keepertest.GenTestAddr().String(),
		Recipient: keepertest.GenTestAddr().String(),
		Amount:    "1000",
		GasPrice:  "1000", // High gas price
		Nonce:     1,
	}

	lowPriorityTx := types.Transaction{
		Sender:    keepertest.GenTestAddr().String(),
		Recipient: keepertest.GenTestAddr().String(),
		Amount:    "1000",
		GasPrice:  "10", // Low gas price
		Nonce:     1,
	}

	highPriority := k.CalculatePriority(input.Ctx, highPriorityTx)
	lowPriority := k.CalculatePriority(input.Ctx, lowPriorityTx)

	require.Greater(t, highPriority, lowPriority)
}

// Anti-spam Tests

func TestRateLimitCheck(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	addr := keepertest.GenTestAddr().String()

	allowed := k.CheckRateLimit(input.Ctx, addr)
	require.True(t, allowed)
}

func TestRateLimitExceeded(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	addr := keepertest.GenTestAddr().String()

	// Send many transactions rapidly
	for i := 0; i < 100; i++ {
		k.RecordTransaction(input.Ctx, addr)
	}

	allowed := k.CheckRateLimit(input.Ctx, addr)
	require.False(t, allowed, "Rate limit should be exceeded")
}

// Genesis Tests

func TestInitGenesis(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	genesisState := types.DefaultGenesis()
	err := k.InitGenesis(input.Ctx, genesisState)
	require.NoError(t, err)
}

func TestExportGenesis(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	genesisState := types.DefaultGenesis()
	err := k.InitGenesis(input.Ctx, genesisState)
	require.NoError(t, err)

	exported := k.ExportGenesis(input.Ctx)
	require.NotNil(t, exported)

	// Compare key fields directly (gogoproto types have internal cache fields that differ)
	require.Equal(t, genesisState.Params.Enabled, exported.Params.Enabled)
	require.Equal(t, genesisState.Params.MaxCacheSize, exported.Params.MaxCacheSize)
	require.Equal(t, genesisState.Params.ExpiryHours, exported.Params.ExpiryHours)
}
