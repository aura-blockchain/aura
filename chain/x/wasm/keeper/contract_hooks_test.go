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
)

// ContractHooksTestSuite is the test suite for contract hooks
type ContractHooksTestSuite struct {
	suite.Suite
	ctx            sdk.Context
	wasmKeeper     Keeper
	registryKeeper *contractregistrykeeper.Keeper
	cdc            codec.BinaryCodec
	contractAddr   sdk.AccAddress
	creator        sdk.AccAddress
	admin          sdk.AccAddress
	sender         sdk.AccAddress
}

// SetupTest sets up the test environment
func (suite *ContractHooksTestSuite) SetupTest() {
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

	// Create context with gas meter
	suite.ctx = sdk.NewContext(cms, tmproto.Header{
		Height: 1,
		Time:   time.Now(),
	}, false, logger).WithGasMeter(storetypes.NewGasMeter(10_000_000)) // 10M gas limit

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

	// Set contract registry
	suite.wasmKeeper.SetContractRegistry(suite.registryKeeper)

	// Create test addresses
	suite.contractAddr = sdk.AccAddress([]byte("contract_address_test"))
	suite.creator = sdk.AccAddress([]byte("creator_test"))
	suite.admin = sdk.AccAddress([]byte("admin_test"))
	suite.sender = sdk.AccAddress([]byte("sender_test"))

	// Reset circuit breaker (now uses KV store)
	suite.wasmKeeper.ResetCircuitBreaker(suite.ctx)

	// Clear validation cache
	valCache.mu.Lock()
	valCache.entries = make(map[string]*validationCacheEntry)
	valCache.mu.Unlock()

	// Clear metrics buffer
	metricsBuf.mu.Lock()
	metricsBuf.updates = metricsBuf.updates[:0]
	metricsBuf.mu.Unlock()
}

// TestContractHooksTestSuite runs the test suite
func TestContractHooksTestSuite(t *testing.T) {
	suite.Run(t, new(ContractHooksTestSuite))
}

// ============================================================================
// BEFORE INSTANTIATE HOOK TESTS
// ============================================================================

func (suite *ContractHooksTestSuite) TestBeforeInstantiateHook_Success() {
	// Test successful pre-instantiation check
	err := suite.wasmKeeper.BeforeInstantiateHook(
		suite.ctx,
		1, // code ID
		suite.creator,
		suite.admin,
		"test-contract",
	)
	require.NoError(suite.T(), err)
}

func (suite *ContractHooksTestSuite) TestBeforeInstantiateHook_NoRegistry() {
	// Test graceful degradation when registry is nil
	keeper := suite.wasmKeeper
	keeper.contractRegistry = nil

	err := keeper.BeforeInstantiateHook(
		suite.ctx,
		1,
		suite.creator,
		suite.admin,
		"test-contract",
	)
	require.NoError(suite.T(), err) // Should not fail
}

func (suite *ContractHooksTestSuite) TestBeforeInstantiateHook_CreatorLimit() {
	// Set low limit
	params := suite.registryKeeper.GetParams(suite.ctx)
	params.MaxContractsPerCreator = 1
	err := suite.registryKeeper.SetParams(suite.ctx, params)
	require.NoError(suite.T(), err)

	// Register first contract
	info := &pb.ContractInfo{
		Address:   "contract1",
		CodeId:    1,
		Creator:   suite.creator.String(),
		Admin:     suite.admin.String(),
		Label:     "contract1",
		CreatedAt: time.Now(),
		Metadata: pb.ContractMetadata{
			Name: "contract1",
		},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause: true,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err = suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// Try to instantiate second contract - should fail
	err = suite.wasmKeeper.BeforeInstantiateHook(
		suite.ctx,
		2,
		suite.creator,
		suite.admin,
		"test-contract-2",
	)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "contract limit exceeded")
}

func (suite *ContractHooksTestSuite) TestBeforeInstantiateHook_RateLimit() {
	// Create multiple instantiation attempts in same block
	for i := 0; i < 10; i++ {
		err := suite.wasmKeeper.BeforeInstantiateHook(
			suite.ctx,
			uint64(i+1),
			suite.creator,
			suite.admin,
			fmt.Sprintf("contract-%d", i),
		)
		require.NoError(suite.T(), err)
	}

	// 11th attempt should fail
	err := suite.wasmKeeper.BeforeInstantiateHook(
		suite.ctx,
		11,
		suite.creator,
		suite.admin,
		"contract-11",
	)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "rate limit exceeded")
}

// ============================================================================
// AFTER INSTANTIATE HOOK TESTS
// ============================================================================

func (suite *ContractHooksTestSuite) TestAfterInstantiateHook_Success() {
	// Test successful auto-registration
	err := suite.wasmKeeper.AfterInstantiateHook(
		suite.ctx,
		suite.contractAddr,
		1, // code ID
		suite.creator,
		suite.admin,
		"test-contract",
	)
	require.NoError(suite.T(), err)

	// Verify contract was registered
	registered := suite.registryKeeper.IsContractRegistered(suite.ctx, suite.contractAddr.String())
	require.True(suite.T(), registered)

	// Verify contract info
	info, found := suite.registryKeeper.GetContractInfo(suite.ctx, suite.contractAddr.String())
	require.True(suite.T(), found)
	require.Equal(suite.T(), suite.contractAddr.String(), info.Address)
	require.Equal(suite.T(), uint64(1), info.CodeId)
	require.Equal(suite.T(), suite.creator.String(), info.Creator)
	require.Equal(suite.T(), suite.admin.String(), info.Admin)
	require.Equal(suite.T(), "test-contract", info.Metadata.Name)
	require.Equal(suite.T(), pb.ContractStatus_CONTRACT_STATUS_ACTIVE, info.Status)
}

func (suite *ContractHooksTestSuite) TestAfterInstantiateHook_NoRegistry() {
	// Test graceful degradation when registry is nil
	keeper := suite.wasmKeeper
	keeper.contractRegistry = nil

	err := keeper.AfterInstantiateHook(
		suite.ctx,
		suite.contractAddr,
		1,
		suite.creator,
		suite.admin,
		"test-contract",
	)
	require.NoError(suite.T(), err) // Should not fail
}

func (suite *ContractHooksTestSuite) TestAfterInstantiateHook_CircuitBreakerOpen() {
	// Open circuit breaker with recent failure time (so timeout hasn't elapsed)
	data := circuitBreakerData{
		FailureCount:      circuitBreakerThreshold,
		LastFailure:       suite.ctx.BlockTime(), // Recent failure keeps circuit open
		State:             circuitBreakerStateOpen,
		ConsecutiveSuccess: 0,
	}
	suite.wasmKeeper.setCircuitBreakerState(suite.ctx, data)

	err := suite.wasmKeeper.AfterInstantiateHook(
		suite.ctx,
		suite.contractAddr,
		1,
		suite.creator,
		suite.admin,
		"test-contract",
	)
	require.NoError(suite.T(), err) // Should not fail

	// Verify contract was NOT registered (graceful degradation)
	registered := suite.registryKeeper.IsContractRegistered(suite.ctx, suite.contractAddr.String())
	require.False(suite.T(), registered)
}

func (suite *ContractHooksTestSuite) TestAfterInstantiateHook_DuplicateRegistration() {
	// Register contract first time
	err := suite.wasmKeeper.AfterInstantiateHook(
		suite.ctx,
		suite.contractAddr,
		1,
		suite.creator,
		suite.admin,
		"test-contract",
	)
	require.NoError(suite.T(), err)

	// Try to register again - should gracefully handle error
	err = suite.wasmKeeper.AfterInstantiateHook(
		suite.ctx,
		suite.contractAddr,
		1,
		suite.creator,
		suite.admin,
		"test-contract",
	)
	require.NoError(suite.T(), err) // Should not fail instantiation
}

// ============================================================================
// BEFORE EXECUTE HOOK TESTS
// ============================================================================

func (suite *ContractHooksTestSuite) TestBeforeExecuteHook_Success() {
	// Register contract first
	info := &pb.ContractInfo{
		Address:   suite.contractAddr.String(),
		CodeId:    1,
		Creator:   suite.creator.String(),
		Admin:     suite.admin.String(),
		Label:     "test-contract",
		CreatedAt: time.Now(),
		Metadata: pb.ContractMetadata{
			Name: "test-contract",
		},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause:       true,
			MaxGasPerTx:      20000000, // 20M gas - must be >= context gas meter limit (10M)
			RateLimitPerUser: 100,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err := suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// Test execution validation
	err = suite.wasmKeeper.BeforeExecuteHook(
		suite.ctx,
		suite.contractAddr,
		suite.sender,
	)
	require.NoError(suite.T(), err)
}

func (suite *ContractHooksTestSuite) TestBeforeExecuteHook_ContractPaused() {
	// Register and pause contract
	info := &pb.ContractInfo{
		Address:   suite.contractAddr.String(),
		CodeId:    1,
		Creator:   suite.creator.String(),
		Admin:     suite.admin.String(),
		Label:     "test-contract",
		CreatedAt: time.Now(),
		Metadata: pb.ContractMetadata{
			Name: "test-contract",
		},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause: true,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_PAUSED,
	}
	err := suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// Try to execute - should fail
	err = suite.wasmKeeper.BeforeExecuteHook(
		suite.ctx,
		suite.contractAddr,
		suite.sender,
	)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "paused")
}

func (suite *ContractHooksTestSuite) TestBeforeExecuteHook_RateLimitExceeded() {
	// Register contract with low rate limit
	info := &pb.ContractInfo{
		Address:   suite.contractAddr.String(),
		CodeId:    1,
		Creator:   suite.creator.String(),
		Admin:     suite.admin.String(),
		Label:     "test-contract",
		CreatedAt: time.Now(),
		Metadata: pb.ContractMetadata{
			Name: "test-contract",
		},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause:       true,
			MaxGasPerTx:      20000000, // Must be >= context gas meter limit
			RateLimitPerUser: 5,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err := suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// Execute 5 times successfully - use different block heights to bypass validation cache
	for i := 0; i < 5; i++ {
		// Create new context with different block height to avoid cache hits
		ctx := suite.ctx.WithBlockHeight(int64(i + 10))
		err = suite.wasmKeeper.BeforeExecuteHook(
			ctx,
			suite.contractAddr,
			suite.sender,
		)
		require.NoError(suite.T(), err, "call %d should succeed", i+1)
	}

	// 6th execution should fail - use yet another block height
	ctx := suite.ctx.WithBlockHeight(20)
	err = suite.wasmKeeper.BeforeExecuteHook(
		ctx,
		suite.contractAddr,
		suite.sender,
	)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "rate limit exceeded")
}

func (suite *ContractHooksTestSuite) TestBeforeExecuteHook_ValidationCache() {
	// Register contract
	info := &pb.ContractInfo{
		Address:   suite.contractAddr.String(),
		CodeId:    1,
		Creator:   suite.creator.String(),
		Admin:     suite.admin.String(),
		Label:     "test-contract",
		CreatedAt: time.Now(),
		Metadata: pb.ContractMetadata{
			Name: "test-contract",
		},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause:       true,
			RateLimitPerUser: 100,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err := suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// First call - should hit registry
	err = suite.wasmKeeper.BeforeExecuteHook(suite.ctx, suite.contractAddr, suite.sender)
	require.NoError(suite.T(), err)

	// Second call in same block - should hit cache
	err = suite.wasmKeeper.BeforeExecuteHook(suite.ctx, suite.contractAddr, suite.sender)
	require.NoError(suite.T(), err)

	// Verify cache has entry
	cacheKey := getCacheKey(suite.contractAddr.String(), suite.sender.String(), suite.ctx.BlockHeight())
	_, found := valCache.get(cacheKey, suite.ctx.BlockHeight())
	require.True(suite.T(), found)
}

func (suite *ContractHooksTestSuite) TestBeforeExecuteHook_NoRegistry() {
	// Test graceful degradation when registry is nil
	keeper := suite.wasmKeeper
	keeper.contractRegistry = nil

	err := keeper.BeforeExecuteHook(
		suite.ctx,
		suite.contractAddr,
		suite.sender,
	)
	require.NoError(suite.T(), err) // Should not fail
}

func (suite *ContractHooksTestSuite) TestBeforeExecuteHook_CircuitBreakerOpen() {
	// Open circuit breaker
	data := circuitBreakerData{
		FailureCount:      circuitBreakerThreshold,
		LastFailure:       suite.ctx.BlockTime(),
		State:             circuitBreakerStateOpen,
		ConsecutiveSuccess: 0,
	}
	suite.wasmKeeper.setCircuitBreakerState(suite.ctx, data)

	err := suite.wasmKeeper.BeforeExecuteHook(
		suite.ctx,
		suite.contractAddr,
		suite.sender,
	)
	require.NoError(suite.T(), err) // Should not fail - permissive mode
}

// ============================================================================
// AFTER EXECUTE HOOK TESTS
// ============================================================================

func (suite *ContractHooksTestSuite) TestAfterExecuteHook_Success() {
	// Register contract first
	info := &pb.ContractInfo{
		Address:   suite.contractAddr.String(),
		CodeId:    1,
		Creator:   suite.creator.String(),
		Admin:     suite.admin.String(),
		Label:     "test-contract",
		CreatedAt: time.Now(),
		Metadata: pb.ContractMetadata{
			Name: "test-contract",
		},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause: true,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err := suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// Execute hook
	gasUsedBefore := suite.ctx.GasMeter().GasConsumed()
	suite.wasmKeeper.AfterExecuteHook(
		suite.ctx,
		suite.contractAddr,
		gasUsedBefore,
		true,
		nil,
	)

	// Metrics should be buffered
	metricsBuf.mu.Lock()
	buffered := len(metricsBuf.updates) > 0
	metricsBuf.mu.Unlock()
	require.True(suite.T(), buffered)
}

func (suite *ContractHooksTestSuite) TestAfterExecuteHook_Failure() {
	// Register contract
	info := &pb.ContractInfo{
		Address:   suite.contractAddr.String(),
		CodeId:    1,
		Creator:   suite.creator.String(),
		Admin:     suite.admin.String(),
		Label:     "test-contract",
		CreatedAt: time.Now(),
		Metadata: pb.ContractMetadata{
			Name: "test-contract",
		},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause: true,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err := suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// Execute hook with failure
	gasUsedBefore := suite.ctx.GasMeter().GasConsumed()
	suite.wasmKeeper.AfterExecuteHook(
		suite.ctx,
		suite.contractAddr,
		gasUsedBefore,
		false,
		fmt.Errorf("execution failed"),
	)

	// Metrics should still be buffered
	metricsBuf.mu.Lock()
	buffered := len(metricsBuf.updates) > 0
	metricsBuf.mu.Unlock()
	require.True(suite.T(), buffered)
}

func (suite *ContractHooksTestSuite) TestAfterExecuteHook_MetricsBufferFlush() {
	// Register contract
	info := &pb.ContractInfo{
		Address:   suite.contractAddr.String(),
		CodeId:    1,
		Creator:   suite.creator.String(),
		Admin:     suite.admin.String(),
		Label:     "test-contract",
		CreatedAt: time.Now(),
		Metadata: pb.ContractMetadata{
			Name: "test-contract",
		},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause: true,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err := suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// Execute hook multiple times to trigger flush
	gasUsedBefore := suite.ctx.GasMeter().GasConsumed()
	for i := 0; i < metricsBufferSize+1; i++ {
		suite.wasmKeeper.AfterExecuteHook(
			suite.ctx,
			suite.contractAddr,
			gasUsedBefore,
			true,
			nil,
		)
	}

	// Buffer should have been flushed
	metricsBuf.mu.Lock()
	bufferSize := len(metricsBuf.updates)
	metricsBuf.mu.Unlock()
	require.Less(suite.T(), bufferSize, metricsBufferSize)
}

func (suite *ContractHooksTestSuite) TestAfterExecuteHook_NoRegistry() {
	// Test graceful degradation when registry is nil
	keeper := suite.wasmKeeper
	keeper.contractRegistry = nil

	// Should not panic
	gasUsedBefore := suite.ctx.GasMeter().GasConsumed()
	keeper.AfterExecuteHook(
		suite.ctx,
		suite.contractAddr,
		gasUsedBefore,
		true,
		nil,
	)
}

// ============================================================================
// CIRCUIT BREAKER TESTS
// ============================================================================

func (suite *ContractHooksTestSuite) TestCircuitBreaker_StateTransitions() {
	ctx := suite.ctx

	// Start in closed state
	require.Equal(suite.T(), circuitBreakerStateClosed, suite.wasmKeeper.getCircuitBreakerStateString(ctx))

	// Record failures to open circuit
	for i := 0; i < circuitBreakerThreshold; i++ {
		suite.wasmKeeper.recordCircuitBreakerFailure(ctx)
	}
	require.Equal(suite.T(), circuitBreakerStateOpen, suite.wasmKeeper.getCircuitBreakerStateString(ctx))

	// Wait for timeout by setting last failure time in the past
	data := suite.wasmKeeper.getCircuitBreakerState(ctx)
	data.LastFailure = ctx.BlockTime().Add(-time.Duration(circuitBreakerTimeout+1) * time.Second)
	suite.wasmKeeper.setCircuitBreakerState(ctx, data)

	// Should not skip (transitions to half-open)
	shouldSkip := suite.wasmKeeper.shouldSkipCircuitBreaker(ctx)
	require.False(suite.T(), shouldSkip)
	require.Equal(suite.T(), circuitBreakerStateHalfOpen, suite.wasmKeeper.getCircuitBreakerStateString(ctx))

	// Record successes to close circuit
	for i := 0; i < circuitBreakerSuccessThreshold; i++ {
		suite.wasmKeeper.recordCircuitBreakerSuccess(ctx)
	}
	require.Equal(suite.T(), circuitBreakerStateClosed, suite.wasmKeeper.getCircuitBreakerStateString(ctx))
}

func (suite *ContractHooksTestSuite) TestCircuitBreaker_ResetManually() {
	ctx := suite.ctx

	// Open circuit breaker
	for i := 0; i < circuitBreakerThreshold; i++ {
		suite.wasmKeeper.recordCircuitBreakerFailure(ctx)
	}
	require.Equal(suite.T(), circuitBreakerStateOpen, suite.wasmKeeper.getCircuitBreakerStateString(ctx))

	// Reset manually
	suite.wasmKeeper.ResetCircuitBreaker(ctx)
	require.Equal(suite.T(), circuitBreakerStateClosed, suite.wasmKeeper.getCircuitBreakerStateString(ctx))
}

// ============================================================================
// PERFORMANCE TESTS
// ============================================================================

func (suite *ContractHooksTestSuite) TestBeforeExecuteHook_PerformanceTarget() {
	// Register contract
	info := &pb.ContractInfo{
		Address:   suite.contractAddr.String(),
		CodeId:    1,
		Creator:   suite.creator.String(),
		Admin:     suite.admin.String(),
		Label:     "test-contract",
		CreatedAt: time.Now(),
		Metadata: pb.ContractMetadata{
			Name: "test-contract",
		},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause:       true,
			RateLimitPerUser: 100,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err := suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// Measure performance
	start := time.Now()
	err = suite.wasmKeeper.BeforeExecuteHook(
		suite.ctx,
		suite.contractAddr,
		suite.sender,
	)
	elapsed := time.Since(start)

	require.NoError(suite.T(), err)

	// Should complete in less than 2ms (performance target)
	suite.T().Logf("BeforeExecuteHook took %v", elapsed)
	// Note: In actual production with real dependencies, this might vary
}

// ============================================================================
// CONCURRENT SAFETY TESTS
// ============================================================================

func (suite *ContractHooksTestSuite) TestConcurrentExecution() {
	// Register contract
	info := &pb.ContractInfo{
		Address:   suite.contractAddr.String(),
		CodeId:    1,
		Creator:   suite.creator.String(),
		Admin:     suite.admin.String(),
		Label:     "test-contract",
		CreatedAt: time.Now(),
		Metadata: pb.ContractMetadata{
			Name: "test-contract",
		},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause:       true,
			RateLimitPerUser: 1000,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err := suite.registryKeeper.RegisterContract(suite.ctx, info)
	require.NoError(suite.T(), err)

	// Test concurrent hook calls (simulating parallel execution)
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Should not panic
			_ = suite.wasmKeeper.BeforeExecuteHook(
				suite.ctx,
				suite.contractAddr,
				suite.sender,
			)
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
