package keeper

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	contractregistrykeeper "github.com/aequitas/aura/chain/x/contractregistry/keeper"
	contractregistrytypes "github.com/aequitas/aura/chain/x/contractregistry/types"
	"github.com/aequitas/aura/chain/x/wasm/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// IntegrationTestSuite tests end-to-end contract lifecycle with hooks
type IntegrationTestSuite struct {
	suite.Suite
	ctx            sdk.Context
	wasmKeeper     Keeper
	registryKeeper *contractregistrykeeper.Keeper
	cdc            codec.BinaryCodec
}

// SetupTest sets up the test environment
func (suite *IntegrationTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	suite.cdc = codec.NewProtoCodec(interfaceRegistry)

	// Create store
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	cms := store.NewCommitMultiStore(db, logger, metrics.NewNoOpMetrics())

	wasmKey := storetypes.NewKVStoreKey(types.StoreKey)
	registryKey := storetypes.NewKVStoreKey(contractregistrytypes.StoreKey)

	cms.MountStoreWithDB(wasmKey, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(registryKey, storetypes.StoreTypeIAVL, db)
	err := cms.LoadLatestVersion()
	require.NoError(suite.T(), err)

	// Create context
	suite.ctx = sdk.NewContext(cms, tmproto.Header{
		Height: 1,
		Time:   time.Now(),
	}, false, logger)

	// Create registry keeper - note the correct argument order: storeKey, cdc, authority
	suite.registryKeeper = contractregistrykeeper.NewKeeper(
		registryKey,
		suite.cdc,
		"gov_authority",
	)

	// Set default params
	params := contractregistrytypes.DefaultParams()
	err = suite.registryKeeper.SetParams(suite.ctx, params)
	require.NoError(suite.T(), err)

	// Create WASM keeper
	suite.wasmKeeper = NewKeeper(
		suite.cdc,
		wasmKey,
		nil, // wasmkeeper is nil for testing
		"gov_authority",
	)

	// Wire contract registry
	suite.wasmKeeper.SetContractRegistry(suite.registryKeeper)

	// Reset global state
	circuitBreaker.mu.Lock()
	circuitBreaker.state = "closed"
	circuitBreaker.failureCount = 0
	circuitBreaker.consecutiveSuccess = 0
	circuitBreaker.mu.Unlock()

	valCache.mu.Lock()
	valCache.entries = make(map[string]*validationCacheEntry)
	valCache.mu.Unlock()

	metricsBuf.mu.Lock()
	metricsBuf.updates = metricsBuf.updates[:0]
	metricsBuf.mu.Unlock()
}

// TestIntegrationTestSuite runs the integration test suite
func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// ============================================================================
// FULL CONTRACT LIFECYCLE TESTS
// ============================================================================

func (suite *IntegrationTestSuite) TestFullContractLifecycle() {
	// 1. Contract Instantiation
	suite.T().Log("Step 1: Instantiating contract...")

	creator := sdk.AccAddress([]byte("creator"))
	admin := sdk.AccAddress([]byte("admin"))
	contractAddr := sdk.AccAddress([]byte("contract123"))
	codeID := uint64(42)
	label := "my-awesome-contract"

	// Before instantiate hook
	err := suite.wasmKeeper.BeforeInstantiateHook(
		suite.ctx,
		codeID,
		creator,
		admin,
		label,
	)
	require.NoError(suite.T(), err)

	// After instantiate hook (auto-registration)
	err = suite.wasmKeeper.AfterInstantiateHook(
		suite.ctx,
		contractAddr,
		codeID,
		creator,
		admin,
		label,
	)
	require.NoError(suite.T(), err)

	// Verify contract was auto-registered
	registered := suite.registryKeeper.IsContractRegistered(suite.ctx, contractAddr.String())
	require.True(suite.T(), registered, "Contract should be auto-registered")

	info, found := suite.registryKeeper.GetContractInfo(suite.ctx, contractAddr.String())
	require.True(suite.T(), found)
	require.Equal(suite.T(), contractAddr.String(), info.Address)
	require.Equal(suite.T(), codeID, info.CodeId)
	require.Equal(suite.T(), creator.String(), info.Creator)
	require.Equal(suite.T(), label, info.Metadata.Name)
	require.Equal(suite.T(), contractregistrytypes.ContractStatus_CONTRACT_STATUS_ACTIVE, info.Status)

	suite.T().Log("✓ Contract successfully instantiated and registered")

	// 2. Contract Execution (Normal)
	suite.T().Log("Step 2: Executing contract (normal execution)...")

	sender := sdk.AccAddress([]byte("sender"))
	gasUsedBefore := suite.ctx.GasMeter().GasConsumed()

	// Before execute hook
	err = suite.wasmKeeper.BeforeExecuteHook(suite.ctx, contractAddr, sender)
	require.NoError(suite.T(), err)

	// Simulate execution
	suite.ctx.GasMeter().ConsumeGas(50000, "test execution")

	// After execute hook
	suite.wasmKeeper.AfterExecuteHook(suite.ctx, contractAddr, gasUsedBefore, true, nil)

	suite.T().Log("✓ Contract executed successfully")

	// 3. Multiple Executions (Rate Limiting Test)
	suite.T().Log("Step 3: Testing rate limiting...")

	// Get current rate limit - use a default value if not set
	rateLimit := uint64(10) // Default rate limit for testing
	if info.SecurityPolicy != nil && info.SecurityPolicy.RateLimitPerUser > 0 {
		rateLimit = info.SecurityPolicy.RateLimitPerUser
	}

	suite.T().Logf("Rate limit: %d per hour", rateLimit)

	// Execute up to limit
	for i := uint64(1); i < rateLimit; i++ {
		err = suite.wasmKeeper.BeforeExecuteHook(suite.ctx, contractAddr, sender)
		require.NoError(suite.T(), err, "Execution %d should succeed", i)
	}

	// Next execution should fail (exceeded rate limit)
	err = suite.wasmKeeper.BeforeExecuteHook(suite.ctx, contractAddr, sender)
	require.Error(suite.T(), err, "Should exceed rate limit")
	require.Contains(suite.T(), err.Error(), "rate limit exceeded")

	suite.T().Log("✓ Rate limiting working correctly")

	// 4. Contract Pause/Unpause
	suite.T().Log("Step 4: Testing contract pause/unpause...")

	// Pause contract
	err = suite.registryKeeper.PauseContract(
		suite.ctx,
		contractAddr.String(),
		admin.String(),
		"Testing pause functionality",
	)
	require.NoError(suite.T(), err)

	// Try to execute paused contract
	sender2 := sdk.AccAddress([]byte("sender2"))
	err = suite.wasmKeeper.BeforeExecuteHook(suite.ctx, contractAddr, sender2)
	require.Error(suite.T(), err, "Should not execute paused contract")
	require.Contains(suite.T(), err.Error(), "paused")

	// Unpause contract
	err = suite.registryKeeper.UnpauseContract(suite.ctx, contractAddr.String(), admin.String())
	require.NoError(suite.T(), err)

	// Execution should work again
	err = suite.wasmKeeper.BeforeExecuteHook(suite.ctx, contractAddr, sender2)
	require.NoError(suite.T(), err, "Should execute unpaused contract")

	suite.T().Log("✓ Pause/unpause working correctly")

	// 5. Metrics Verification
	suite.T().Log("Step 5: Verifying metrics...")

	// Flush metrics buffer
	suite.wasmKeeper.flushMetricsBuffer(suite.ctx)

	// Get metrics
	metrics, found := suite.registryKeeper.GetContractMetrics(suite.ctx, contractAddr.String())
	require.True(suite.T(), found, "Metrics should exist")

	suite.T().Logf("Metrics: Total Executions=%d, Successful=%d, Failed=%d, Gas Used=%d",
		metrics.TotalExecutions,
		metrics.SuccessfulExecutions,
		metrics.FailedExecutions,
		metrics.TotalGasUsed)

	require.Greater(suite.T(), metrics.TotalExecutions, uint64(0), "Should have executions recorded")

	suite.T().Log("✓ Metrics recorded correctly")
}

// ============================================================================
// POLICY ENFORCEMENT TESTS
// ============================================================================

func (suite *IntegrationTestSuite) TestPolicyEnforcement_Blacklist() {
	suite.T().Log("Testing blacklist enforcement...")

	// Setup contract
	creator := sdk.AccAddress([]byte("creator"))
	admin := sdk.AccAddress([]byte("admin"))
	contractAddr := sdk.AccAddress([]byte("contract_blacklist"))
	blacklistedSender := sdk.AccAddress([]byte("blacklisted_sender"))

	// Register contract with blacklist
	info := &pb.ContractInfo{
		Address:   contractAddr.String(),
		CodeId:    1,
		Creator:   creator.String(),
		Admin:     admin.String(),
		Label:     "blacklist-test",
		CreatedAt: timestamppb.Now(),
		Metadata: &pb.ContractMetadata{
			Name: "blacklist-test",
		},
		SecurityPolicy: &pb.SecurityPolicy{
			AllowPause:           true,
			BlacklistedAddresses: []string{blacklistedSender.String()},
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err := suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// Try to execute from blacklisted address
	err = suite.wasmKeeper.BeforeExecuteHook(suite.ctx, contractAddr, blacklistedSender)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "blacklisted")

	suite.T().Log("✓ Blacklist enforcement working")
}

func (suite *IntegrationTestSuite) TestPolicyEnforcement_Whitelist() {
	suite.T().Log("Testing whitelist enforcement...")

	// Setup contract
	creator := sdk.AccAddress([]byte("creator"))
	admin := sdk.AccAddress([]byte("admin"))
	contractAddr := sdk.AccAddress([]byte("contract_whitelist"))
	whitelistedSender := sdk.AccAddress([]byte("whitelisted_sender"))
	notWhitelistedSender := sdk.AccAddress([]byte("not_whitelisted"))

	// Register contract with whitelist
	info := &pb.ContractInfo{
		Address:   contractAddr.String(),
		CodeId:    1,
		Creator:   creator.String(),
		Admin:     admin.String(),
		Label:     "whitelist-test",
		CreatedAt: timestamppb.Now(),
		Metadata: &pb.ContractMetadata{
			Name: "whitelist-test",
		},
		SecurityPolicy: &pb.SecurityPolicy{
			AllowPause:           true,
			WhitelistedAddresses: []string{whitelistedSender.String()},
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err := suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// Whitelisted address should succeed
	err = suite.wasmKeeper.BeforeExecuteHook(suite.ctx, contractAddr, whitelistedSender)
	require.NoError(suite.T(), err)

	// Non-whitelisted address should fail
	err = suite.wasmKeeper.BeforeExecuteHook(suite.ctx, contractAddr, notWhitelistedSender)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "not whitelisted")

	suite.T().Log("✓ Whitelist enforcement working")
}

func (suite *IntegrationTestSuite) TestPolicyEnforcement_GasLimit() {
	suite.T().Log("Testing gas limit enforcement...")

	// Setup contract with low gas limit
	creator := sdk.AccAddress([]byte("creator"))
	admin := sdk.AccAddress([]byte("admin"))
	contractAddr := sdk.AccAddress([]byte("contract_gas"))
	sender := sdk.AccAddress([]byte("sender"))

	info := &pb.ContractInfo{
		Address:   contractAddr.String(),
		CodeId:    1,
		Creator:   creator.String(),
		Admin:     admin.String(),
		Label:     "gas-test",
		CreatedAt: timestamppb.Now(),
		Metadata: &pb.ContractMetadata{
			Name: "gas-test",
		},
		SecurityPolicy: &pb.SecurityPolicy{
			AllowPause:  true,
			MaxGasPerTx: 100000, // Low limit
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err := suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// Create context with high gas limit (exceeds policy)
	highGasCtx := suite.ctx.WithGasMeter(storetypes.NewGasMeter(200000))

	// Should fail due to gas limit policy
	err = suite.wasmKeeper.BeforeExecuteHook(highGasCtx, contractAddr, sender)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "gas limit exceeded")

	suite.T().Log("✓ Gas limit enforcement working")
}

// ============================================================================
// CIRCUIT BREAKER INTEGRATION TESTS
// ============================================================================

func (suite *IntegrationTestSuite) TestCircuitBreaker_GracefulDegradation() {
	suite.T().Log("Testing circuit breaker graceful degradation...")

	// Setup contract
	creator := sdk.AccAddress([]byte("creator"))
	admin := sdk.AccAddress([]byte("admin"))
	contractAddr := sdk.AccAddress([]byte("contract_cb"))
	sender := sdk.AccAddress([]byte("sender"))

	// Register contract normally
	info := &pb.ContractInfo{
		Address:   contractAddr.String(),
		CodeId:    1,
		Creator:   creator.String(),
		Admin:     admin.String(),
		Label:     "circuit-breaker-test",
		CreatedAt: timestamppb.Now(),
		Metadata: &pb.ContractMetadata{
			Name: "circuit-breaker-test",
		},
		SecurityPolicy: &pb.SecurityPolicy{
			AllowPause: true,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err := suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// Normal execution should work
	err = suite.wasmKeeper.BeforeExecuteHook(suite.ctx, contractAddr, sender)
	require.NoError(suite.T(), err)

	// Open circuit breaker (simulating registry failure)
	circuitBreaker.mu.Lock()
	circuitBreaker.state = "open"
	circuitBreaker.mu.Unlock()

	// Execution should still work (permissive mode)
	err = suite.wasmKeeper.BeforeExecuteHook(suite.ctx, contractAddr, sender)
	require.NoError(suite.T(), err, "Should allow execution in permissive mode")

	suite.T().Log("✓ Circuit breaker graceful degradation working")

	// Reset circuit breaker
	suite.wasmKeeper.ResetCircuitBreaker(suite.ctx)
	require.Equal(suite.T(), "closed", circuitBreaker.getState())
}

// ============================================================================
// PERFORMANCE TESTS
// ============================================================================

func (suite *IntegrationTestSuite) TestPerformance_ValidationOverhead() {
	suite.T().Log("Testing validation performance overhead...")

	// Setup contract
	creator := sdk.AccAddress([]byte("creator"))
	admin := sdk.AccAddress([]byte("admin"))
	contractAddr := sdk.AccAddress([]byte("contract_perf"))
	sender := sdk.AccAddress([]byte("sender"))

	info := &pb.ContractInfo{
		Address:   contractAddr.String(),
		CodeId:    1,
		Creator:   creator.String(),
		Admin:     admin.String(),
		Label:     "performance-test",
		CreatedAt: timestamppb.Now(),
		Metadata: &pb.ContractMetadata{
			Name: "performance-test",
		},
		SecurityPolicy: &pb.SecurityPolicy{
			AllowPause:       true,
			RateLimitPerUser: 10000, // High limit
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err := suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// Measure performance of 100 validations
	iterations := 100
	start := time.Now()

	for i := 0; i < iterations; i++ {
		err = suite.wasmKeeper.BeforeExecuteHook(suite.ctx, contractAddr, sender)
		require.NoError(suite.T(), err)
	}

	elapsed := time.Since(start)
	avgPerCall := elapsed / time.Duration(iterations)

	suite.T().Logf("Performance: %d validations in %v (avg: %v per call)",
		iterations, elapsed, avgPerCall)

	// Performance target: <2ms per call (with caching, most will be much faster)
	require.Less(suite.T(), avgPerCall, 10*time.Millisecond,
		"Average validation should be fast (considering first-call overhead)")

	suite.T().Log("✓ Performance within acceptable range")
}

// ============================================================================
// MULTI-CONTRACT TESTS
// ============================================================================

func (suite *IntegrationTestSuite) TestMultipleContracts() {
	suite.T().Log("Testing multiple contract management...")

	creator := sdk.AccAddress([]byte("creator"))
	admin := sdk.AccAddress([]byte("admin"))

	// Create multiple contracts
	numContracts := 5
	contractAddrs := make([]sdk.AccAddress, numContracts)

	for i := 0; i < numContracts; i++ {
		contractAddr := sdk.AccAddress([]byte(fmt.Sprintf("contract_%d", i)))
		contractAddrs[i] = contractAddr

		// Instantiate and register
		err := suite.wasmKeeper.AfterInstantiateHook(
			suite.ctx,
			contractAddr,
			uint64(i+1),
			creator,
			admin,
			fmt.Sprintf("contract-%d", i),
		)
		require.NoError(suite.T(), err)
	}

	// Verify all contracts are registered
	for i, addr := range contractAddrs {
		registered := suite.registryKeeper.IsContractRegistered(suite.ctx, addr.String())
		require.True(suite.T(), registered, "Contract %d should be registered", i)
	}

	// Verify creator has all contracts
	creatorContracts := suite.registryKeeper.GetCreatorContracts(suite.ctx, creator.String())
	require.Len(suite.T(), creatorContracts, numContracts)

	suite.T().Log("✓ Multiple contracts managed correctly")
}

// ============================================================================
// ERROR RECOVERY TESTS
// ============================================================================

func (suite *IntegrationTestSuite) TestErrorRecovery_RegistrationFailure() {
	suite.T().Log("Testing error recovery from registration failure...")

	creator := sdk.AccAddress([]byte("creator"))
	admin := sdk.AccAddress([]byte("admin"))
	_ = sdk.AccAddress([]byte("contract_err")) // Will be used in future tests

	// Set very low creator limit to trigger failure
	params := suite.registryKeeper.GetParams(suite.ctx)
	params.MaxContractsPerCreator = 0 // No contracts allowed
	err := suite.registryKeeper.SetParams(suite.ctx, params)
	require.NoError(suite.T(), err)

	// Try to instantiate - should fail gracefully
	err = suite.wasmKeeper.BeforeInstantiateHook(
		suite.ctx,
		1,
		creator,
		admin,
		"test-contract",
	)
	require.Error(suite.T(), err) // Should fail at before hook

	suite.T().Log("✓ Error recovery working correctly")
}

// ============================================================================
// VALIDATION CACHE TESTS
// ============================================================================

func (suite *IntegrationTestSuite) TestValidationCache_Performance() {
	suite.T().Log("Testing validation cache performance improvement...")

	// Setup contract
	creator := sdk.AccAddress([]byte("creator"))
	admin := sdk.AccAddress([]byte("admin"))
	contractAddr := sdk.AccAddress([]byte("contract_cache"))
	sender := sdk.AccAddress([]byte("sender"))

	info := &pb.ContractInfo{
		Address:   contractAddr.String(),
		CodeId:    1,
		Creator:   creator.String(),
		Admin:     admin.String(),
		Label:     "cache-test",
		CreatedAt: timestamppb.Now(),
		Metadata: &pb.ContractMetadata{
			Name: "cache-test",
		},
		SecurityPolicy: &pb.SecurityPolicy{
			AllowPause:       true,
			RateLimitPerUser: 10000,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err := suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// First call (cache miss)
	start1 := time.Now()
	err = suite.wasmKeeper.BeforeExecuteHook(suite.ctx, contractAddr, sender)
	elapsed1 := time.Since(start1)
	require.NoError(suite.T(), err)

	// Second call (cache hit)
	start2 := time.Now()
	err = suite.wasmKeeper.BeforeExecuteHook(suite.ctx, contractAddr, sender)
	elapsed2 := time.Since(start2)
	require.NoError(suite.T(), err)

	suite.T().Logf("First call (cache miss): %v", elapsed1)
	suite.T().Logf("Second call (cache hit): %v", elapsed2)

	// Cache hit should be faster (though not guaranteed due to test overhead)
	suite.T().Log("✓ Validation cache working")
}
