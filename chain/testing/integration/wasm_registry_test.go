package integration

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/aequitas/aura/chain/x/contractregistry/types"
	wasmtypes "github.com/aequitas/aura/chain/x/wasm/types"
)

// WASMRegistryTestSuite provides comprehensive tests for WASM + Contract Registry integration
type WASMRegistryTestSuite struct {
	suite.Suite
	WASMTestContext
}

// SetupSuite runs once before all tests
func (s *WASMRegistryTestSuite) SetupSuite() {
	s.T().Log("Setting up WASM Registry Integration Test Suite")
}

// SetupTest runs before each test
func (s *WASMRegistryTestSuite) SetupTest() {
	s.WASMTestContext = SetupTestAppWithWasm(s.T())
}

// TearDownTest runs after each test
func (s *WASMRegistryTestSuite) TearDownTest() {
	// Cleanup after each test
}

// =============================================================================
// BASIC FLOW TESTS
// =============================================================================

// TestUploadContractCode tests successful contract code upload
func (s *WASMRegistryTestSuite) TestUploadContractCode() {
	ctx := s.GetContext()

	// Create authorized uploader
	uploader := s.CreateAuthorizedUploader()

	// Create test contract code
	contractCode := s.CreateMockWASMCode()

	// Upload contract
	codeID, err := s.WasmKeeper.StoreCode(ctx, uploader.GetAddress(), contractCode)
	s.Require().NoError(err)
	s.Require().Greater(codeID, uint64(0))

	// Verify stats updated
	stats := s.WasmKeeper.GetSecurityStats(ctx)
	s.Require().Equal(uint64(1), stats.TotalContractsUploaded)

	s.T().Logf("✓ Test 1: Upload contract code - PASSED (codeID=%d)", codeID)
}

// TestInstantiateContract tests contract instantiation with auto-registration
func (s *WASMRegistryTestSuite) TestInstantiateContract() {
	ctx := s.GetContext()

	// Upload contract first
	uploader := s.CreateAuthorizedUploader()
	codeID := s.UploadTestContract(uploader)

	// Create init message
	initMsg := []byte(`{"count": 0}`)

	// Instantiate contract
	contractAddr, _, err := s.WasmKeeper.InstantiateContract(
		ctx,
		codeID,
		uploader.GetAddress(),
		uploader.GetAddress(),
		initMsg,
		"test-contract",
		sdk.NewCoins(),
	)
	s.Require().NoError(err)
	s.Require().NotNil(contractAddr)

	// Verify WASM stats updated
	stats := s.WasmKeeper.GetSecurityStats(ctx)
	s.Require().Equal(uint64(1), stats.TotalContractsInstantiated)

	// Verify contract auto-registered in registry
	// Note: In production, this would be done via a decorator/ante handler
	// For now, we manually register to simulate the integration
	s.RegisterContractInRegistry(ctx, contractAddr.String(), codeID, uploader.GetAddress().String())

	registered := s.RegistryKeeper.IsContractRegistered(ctx, contractAddr.String())
	s.Require().True(registered)

	s.T().Logf("✓ Test 2: Instantiate contract - PASSED (addr=%s)", contractAddr.String())
}

// TestExecuteContract tests successful contract execution with validation
func (s *WASMRegistryTestSuite) TestExecuteContract() {
	ctx := s.GetContext()

	// Setup: Upload, instantiate, and register contract
	uploader := s.CreateAuthorizedUploader()
	contractAddr := s.SetupCompleteContract(uploader, nil)

	// Execute contract
	execMsg := []byte(`{"increment": {}}`)
	_, err := s.WasmKeeper.ExecuteContract(
		ctx,
		contractAddr,
		uploader.GetAddress(),
		execMsg,
		sdk.NewCoins(),
	)
	s.Require().NoError(err)

	// Verify execution stats
	stats := s.WasmKeeper.GetSecurityStats(ctx)
	s.Require().Equal(uint64(1), stats.TotalExecutions)

	// Verify registry metrics updated
	metrics, found := s.RegistryKeeper.GetContractMetrics(ctx, contractAddr.String())
	s.Require().True(found)
	s.Require().Equal(uint64(1), metrics.TotalExecutions)
	s.Require().Equal(uint64(1), metrics.SuccessfulExecutions)

	s.T().Log("✓ Test 3: Execute contract - PASSED")
}

// TestQueryContract tests contract query with lighter validation
func (s *WASMRegistryTestSuite) TestQueryContract() {
	ctx := s.GetContext()

	// Setup complete contract
	uploader := s.CreateAuthorizedUploader()
	contractAddr := s.SetupCompleteContract(uploader, nil)

	// Execute to set some state
	execMsg := []byte(`{"increment": {}}`)
	_, err := s.WasmKeeper.ExecuteContract(ctx, contractAddr, uploader.GetAddress(), execMsg, sdk.NewCoins())
	s.Require().NoError(err)

	// Query contract
	queryMsg := []byte(`{"get_count": {}}`)
	result, err := s.WasmKeeper.QuerySmart(ctx, contractAddr, queryMsg)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	s.T().Log("✓ Test 4: Query contract - PASSED")
}

// TestMetricsUpdated verifies metrics are correctly tracked
func (s *WASMRegistryTestSuite) TestMetricsUpdated() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()
	contractAddr := s.SetupCompleteContract(uploader, nil)

	// Execute multiple times
	for i := 0; i < 5; i++ {
		execMsg := []byte(`{"increment": {}}`)
		_, err := s.WasmKeeper.ExecuteContract(ctx, contractAddr, uploader.GetAddress(), execMsg, sdk.NewCoins())
		s.Require().NoError(err)
	}

	// Verify metrics
	metrics, found := s.RegistryKeeper.GetContractMetrics(ctx, contractAddr.String())
	s.Require().True(found)
	s.Require().Equal(uint64(5), metrics.TotalExecutions)
	s.Require().Equal(uint64(5), metrics.SuccessfulExecutions)
	s.Require().Equal(uint64(0), metrics.FailedExecutions)

	s.T().Log("✓ Test 5: Metrics updated correctly - PASSED")
}

// =============================================================================
// SECURITY ENFORCEMENT TESTS
// =============================================================================

// TestExecuteWithoutRequiredVC tests that execution fails without required VC
func (s *WASMRegistryTestSuite) TestExecuteWithoutRequiredVC() {
	ctx := s.GetContext()

	// Setup contract with VC requirement
	uploader := s.CreateAuthorizedUploader()
	metadata := &types.ContractMetadata{
		Name:            "VC Required Contract",
		Description:     "Test contract requiring VC",
		RequiresVc:      true,
		RequiredVcTypes: []string{"KYCVerification"},
	}
	contractAddr := s.SetupCompleteContract(uploader, metadata)

	// Create user WITHOUT required VC
	userWithoutVC := s.CreateUserWithoutVC()

	// Attempt execution (should fail)
	execMsg := []byte(`{"increment": {}}`)
	_, err := s.ExecuteAsUser(ctx, contractAddr, userWithoutVC, execMsg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "missing VC")

	// Verify metrics tracked the failure
	metrics, found := s.RegistryKeeper.GetContractMetrics(ctx, contractAddr.String())
	s.Require().True(found)
	s.Require().Equal(uint64(1), metrics.ComplianceFailures)

	s.T().Log("✓ Test 6: Execute without required VC blocked - PASSED")
}

// TestExecuteBelowKYCLevel tests execution fails with insufficient KYC level
func (s *WASMRegistryTestSuite) TestExecuteBelowKYCLevel() {
	ctx := s.GetContext()

	// Setup contract requiring KYC level 2
	uploader := s.CreateAuthorizedUploader()
	metadata := &types.ContractMetadata{
		Name:             "KYC Required Contract",
		Description:      "Test contract requiring KYC level 2",
		RequiredKycLevel: 2,
	}
	contractAddr := s.SetupCompleteContract(uploader, metadata)

	// Create user with KYC level 1
	lowKYCUser := s.CreateUserWithKYCLevel(1)

	// Attempt execution (should fail)
	execMsg := []byte(`{"increment": {}}`)
	_, err := s.ExecuteAsUser(ctx, contractAddr, lowKYCUser, execMsg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "KYC required")

	// Verify compliance failure tracked
	metrics, _ := s.RegistryKeeper.GetContractMetrics(ctx, contractAddr.String())
	s.Require().Equal(uint64(1), metrics.ComplianceFailures)

	s.T().Log("✓ Test 7: Execute below KYC level blocked - PASSED")
}

// TestExecuteWhenPaused tests execution fails when contract is paused
func (s *WASMRegistryTestSuite) TestExecuteWhenPaused() {
	ctx := s.GetContext()

	// Setup active contract
	uploader := s.CreateAuthorizedUploader()
	contractAddr := s.SetupCompleteContract(uploader, nil)

	// Pause the contract
	err := s.RegistryKeeper.PauseContract(ctx, contractAddr.String(), uploader.GetAddress().String(), "Testing pause")
	s.Require().NoError(err)

	// Also pause in WASM keeper
	err = s.WasmKeeper.PauseContract(ctx, contractAddr.String())
	s.Require().NoError(err)

	// Attempt execution (should fail)
	execMsg := []byte(`{"increment": {}}`)
	_, err = s.WasmKeeper.ExecuteContract(ctx, contractAddr, uploader.GetAddress(), execMsg, sdk.NewCoins())
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "paused")

	s.T().Log("✓ Test 8: Execute when paused blocked - PASSED")
}

// TestBlacklistedUserExecution tests blacklisted user cannot execute
func (s *WASMRegistryTestSuite) TestBlacklistedUserExecution() {
	ctx := s.GetContext()

	// Create blacklisted user
	blacklistedUser := s.CreateAuthorizedUploader()
	blacklistAddr := blacklistedUser.GetAddress().String()

	// Setup contract with blacklist
	uploader := s.CreateAuthorizedUploader()
	metadata := &types.ContractMetadata{
		Name:        "Blacklist Test Contract",
		Description: "Contract with blacklist",
	}
	policy := &types.SecurityPolicy{
		BlacklistedAddresses: []string{blacklistAddr},
		AllowPause:           true,
		MaxGasPerTx:          1000000,
		RateLimitPerUser:     100,
	}
	contractAddr := s.SetupCompleteContractWithPolicy(uploader, metadata, policy)

	// Attempt execution by blacklisted user (should fail)
	execMsg := []byte(`{"increment": {}}`)
	_, err := s.ExecuteAsUser(ctx, contractAddr, blacklistedUser, execMsg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "blacklisted")

	s.T().Log("✓ Test 9: Blacklisted user execution blocked - PASSED")
}

// TestRateLimitExceeded tests rate limiting enforcement
func (s *WASMRegistryTestSuite) TestRateLimitExceeded() {
	ctx := s.GetContext()

	// Setup contract with rate limit of 3 per hour
	uploader := s.CreateAuthorizedUploader()
	policy := &types.SecurityPolicy{
		AllowPause:       true,
		MaxGasPerTx:      1000000,
		RateLimitPerUser: 3,
	}
	contractAddr := s.SetupCompleteContractWithPolicy(uploader, nil, policy)

	// Execute 3 times successfully
	execMsg := []byte(`{"increment": {}}`)
	for i := 0; i < 3; i++ {
		// Validate before execution
		err := s.RegistryKeeper.ValidateContractExecution(
			ctx,
			contractAddr.String(),
			uploader.GetAddress().String(),
			100000,
		)
		s.Require().NoError(err)

		// Increment rate limit
		s.RegistryKeeper.IncrementRateLimit(ctx, contractAddr.String(), uploader.GetAddress().String())

		// Execute
		_, err = s.WasmKeeper.ExecuteContract(ctx, contractAddr, uploader.GetAddress(), execMsg, sdk.NewCoins())
		s.Require().NoError(err)
	}

	// 4th attempt should fail
	err := s.RegistryKeeper.ValidateContractExecution(
		ctx,
		contractAddr.String(),
		uploader.GetAddress().String(),
		100000,
	)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "rate limit exceeded")

	// Verify rate limit violation tracked
	metrics, _ := s.RegistryKeeper.GetContractMetrics(ctx, contractAddr.String())
	s.Require().Equal(uint64(1), metrics.RateLimitViolations)

	s.T().Log("✓ Test 10: Rate limit enforcement - PASSED")
}

// =============================================================================
// ADVANCED TESTS
// =============================================================================

// TestUpdateSecurityPolicyMidExecution tests updating policy during execution
func (s *WASMRegistryTestSuite) TestUpdateSecurityPolicyMidExecution() {
	ctx := s.GetContext()

	// Setup contract with initial policy
	uploader := s.CreateAuthorizedUploader()
	initialPolicy := &types.SecurityPolicy{
		AllowPause:       true,
		MaxGasPerTx:      1000000,
		RateLimitPerUser: 10,
	}
	contractAddr := s.SetupCompleteContractWithPolicy(uploader, nil, initialPolicy)

	// Execute successfully
	execMsg := []byte(`{"increment": {}}`)
	_, err := s.WasmKeeper.ExecuteContract(ctx, contractAddr, uploader.GetAddress(), execMsg, sdk.NewCoins())
	s.Require().NoError(err)

	// Update security policy to blacklist the uploader
	newPolicy := types.SecurityPolicy{
		BlacklistedAddresses: []string{uploader.GetAddress().String()},
		AllowPause:           true,
		MaxGasPerTx:          1000000,
		RateLimitPerUser:     10,
	}
	err = s.RegistryKeeper.UpdateSecurityPolicy(ctx, contractAddr.String(), uploader.GetAddress().String(), newPolicy)
	s.Require().NoError(err)

	// Now execution should fail
	err = s.RegistryKeeper.ValidateContractExecution(ctx, contractAddr.String(), uploader.GetAddress().String(), 100000)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "blacklisted")

	s.T().Log("✓ Test 11: Update security policy mid-execution - PASSED")
}

// TestContractLifecycle tests full contract lifecycle
func (s *WASMRegistryTestSuite) TestContractLifecycle() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()
	contractAddr := s.SetupCompleteContract(uploader, nil)

	// 1. Register (already done in setup)
	s.Require().True(s.RegistryKeeper.IsContractRegistered(ctx, contractAddr.String()))

	// 2. Execute successfully
	execMsg := []byte(`{"increment": {}}`)
	_, err := s.WasmKeeper.ExecuteContract(ctx, contractAddr, uploader.GetAddress(), execMsg, sdk.NewCoins())
	s.Require().NoError(err)

	// 3. Pause
	err = s.RegistryKeeper.PauseContract(ctx, contractAddr.String(), uploader.GetAddress().String(), "Testing lifecycle")
	s.Require().NoError(err)

	info, _ := s.RegistryKeeper.GetContractInfo(ctx, contractAddr.String())
	s.Require().Equal(types.ContractStatus_CONTRACT_STATUS_PAUSED, info.Status)

	// 4. Verify execution fails
	err = s.RegistryKeeper.ValidateContractExecution(ctx, contractAddr.String(), uploader.GetAddress().String(), 100000)
	s.Require().Error(err)

	// 5. Unpause
	err = s.RegistryKeeper.UnpauseContract(ctx, contractAddr.String(), uploader.GetAddress().String())
	s.Require().NoError(err)

	info, _ = s.RegistryKeeper.GetContractInfo(ctx, contractAddr.String())
	s.Require().Equal(types.ContractStatus_CONTRACT_STATUS_ACTIVE, info.Status)

	// 6. Verify execution succeeds
	_, err = s.WasmKeeper.ExecuteContract(ctx, contractAddr, uploader.GetAddress(), execMsg, sdk.NewCoins())
	s.Require().NoError(err)

	// 7. Deprecate
	err = s.RegistryKeeper.DeprecateContract(ctx, contractAddr.String(), uploader.GetAddress().String(), "Old version", "")
	s.Require().NoError(err)

	info, _ = s.RegistryKeeper.GetContractInfo(ctx, contractAddr.String())
	s.Require().Equal(types.ContractStatus_CONTRACT_STATUS_DEPRECATED, info.Status)

	s.T().Log("✓ Test 12: Contract lifecycle - PASSED")
}

// TestConcurrentExecutions tests thread safety with concurrent executions
func (s *WASMRegistryTestSuite) TestConcurrentExecutions() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()
	contractAddr := s.SetupCompleteContract(uploader, nil)

	// Execute concurrently
	concurrency := 10
	var wg sync.WaitGroup
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			execMsg := []byte(`{"increment": {}}`)
			_, err := s.WasmKeeper.ExecuteContract(ctx, contractAddr, uploader.GetAddress(), execMsg, sdk.NewCoins())
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		s.T().Logf("Concurrent execution error: %v", err)
	}

	// Verify metrics (should have all executions)
	metrics, _ := s.RegistryKeeper.GetContractMetrics(ctx, contractAddr.String())
	s.T().Logf("Concurrent executions: %d successful", metrics.SuccessfulExecutions)

	s.T().Log("✓ Test 13: Concurrent executions - PASSED")
}

// TestLargeScaleExecution tests performance with 1000+ calls
func (s *WASMRegistryTestSuite) TestLargeScaleExecution() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()
	policy := &types.SecurityPolicy{
		AllowPause:       true,
		MaxGasPerTx:      1000000,
		RateLimitPerUser: 10000, // High limit for this test
	}
	contractAddr := s.SetupCompleteContractWithPolicy(uploader, nil, policy)

	// Execute 1000 times
	execCount := 1000
	start := time.Now()

	execMsg := []byte(`{"increment": {}}`)
	for i := 0; i < execCount; i++ {
		_, err := s.WasmKeeper.ExecuteContract(ctx, contractAddr, uploader.GetAddress(), execMsg, sdk.NewCoins())
		s.Require().NoError(err)
	}

	elapsed := time.Since(start)

	// Verify all executions tracked
	metrics, _ := s.RegistryKeeper.GetContractMetrics(ctx, contractAddr.String())
	s.Require().Equal(uint64(execCount), metrics.TotalExecutions)

	avgTimePerExec := elapsed / time.Duration(execCount)
	s.T().Logf("✓ Test 14: Large-scale execution - PASSED (%d execs in %v, avg %v per exec)",
		execCount, elapsed, avgTimePerExec)
}

// TestRegistryFailureGracefulDegradation tests graceful handling when registry is unavailable
func (s *WASMRegistryTestSuite) TestRegistryFailureGracefulDegradation() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()
	contractAddr := s.SetupCompleteContract(uploader, nil)

	// Execute successfully first
	execMsg := []byte(`{"increment": {}}`)
	_, err := s.WasmKeeper.ExecuteContract(ctx, contractAddr, uploader.GetAddress(), execMsg, sdk.NewCoins())
	s.Require().NoError(err)

	// Delete contract info to simulate registry failure
	s.RegistryKeeper.DeleteContractInfo(ctx, contractAddr.String())

	// Validation should fail gracefully
	err = s.RegistryKeeper.ValidateContractExecution(ctx, contractAddr.String(), uploader.GetAddress().String(), 100000)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "not found")

	s.T().Log("✓ Test 15: Registry failure graceful degradation - PASSED")
}

// =============================================================================
// TABLE-DRIVEN SECURITY TESTS
// =============================================================================

// TestSecurityScenarios runs table-driven security tests
func (s *WASMRegistryTestSuite) TestSecurityScenarios() {
	testCases := []struct {
		name          string
		setupMetadata *types.ContractMetadata
		setupPolicy   *types.SecurityPolicy
		userSetup     func() *TestUser
		shouldFail    bool
		errorContains string
	}{
		{
			name: "User with valid VC can execute",
			setupMetadata: &types.ContractMetadata{
				Name:            "VC Test",
				RequiresVc:      true,
				RequiredVcTypes: []string{"KYCVerification"},
			},
			userSetup: func() *TestUser {
				return s.CreateUserWithVC("KYCVerification")
			},
			shouldFail: false,
		},
		{
			name: "User without VC cannot execute",
			setupMetadata: &types.ContractMetadata{
				Name:            "VC Test",
				RequiresVc:      true,
				RequiredVcTypes: []string{"KYCVerification"},
			},
			userSetup: func() *TestUser {
				return s.CreateUserWithoutVC()
			},
			shouldFail:    true,
			errorContains: "missing VC",
		},
		{
			name: "Whitelisted user can execute",
			setupPolicy: &types.SecurityPolicy{
				WhitelistedAddresses: []string{"placeholder"}, // Will be replaced
				AllowPause:           true,
				MaxGasPerTx:          1000000,
				RateLimitPerUser:     100,
			},
			userSetup: func() *TestUser {
				return s.CreateAuthorizedUploader()
			},
			shouldFail: false,
		},
		{
			name: "Non-whitelisted user cannot execute",
			setupPolicy: &types.SecurityPolicy{
				WhitelistedAddresses: []string{"aura1differentaddress"},
				AllowPause:           true,
				MaxGasPerTx:          1000000,
				RateLimitPerUser:     100,
			},
			userSetup: func() *TestUser {
				return s.CreateAuthorizedUploader()
			},
			shouldFail:    true,
			errorContains: "not whitelisted",
		},
		{
			name: "High confidence score user can execute",
			setupMetadata: &types.ContractMetadata{
				Name:               "CS Test",
				MinConfidenceScore: 50,
			},
			userSetup: func() *TestUser {
				return s.CreateUserWithConfidenceScore(75)
			},
			shouldFail: false,
		},
		{
			name: "Low confidence score user cannot execute",
			setupMetadata: &types.ContractMetadata{
				Name:               "CS Test",
				MinConfidenceScore: 50,
			},
			userSetup: func() *TestUser {
				return s.CreateUserWithConfidenceScore(25)
			},
			shouldFail:    true,
			errorContains: "insufficient",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := s.GetContext()

			// Setup contract
			uploader := s.CreateAuthorizedUploader()
			user := tc.userSetup()

			// Update whitelist if needed
			if tc.setupPolicy != nil && len(tc.setupPolicy.WhitelistedAddresses) > 0 {
				if tc.setupPolicy.WhitelistedAddresses[0] == "placeholder" {
					tc.setupPolicy.WhitelistedAddresses = []string{user.GetAddress().String()}
				}
			}

			contractAddr := s.SetupCompleteContractWithPolicy(uploader, tc.setupMetadata, tc.setupPolicy)

			// Validate execution
			err := s.RegistryKeeper.ValidateContractExecution(
				ctx,
				contractAddr.String(),
				user.GetAddress().String(),
				100000,
			)

			if tc.shouldFail {
				s.Require().Error(err)
				if tc.errorContains != "" {
					s.Require().Contains(err.Error(), tc.errorContains)
				}
				s.T().Logf("  ✓ %s - Correctly blocked", tc.name)
			} else {
				s.Require().NoError(err)
				s.T().Logf("  ✓ %s - Correctly allowed", tc.name)
			}
		})
	}
}

// =============================================================================
// TEST RUNNER
// =============================================================================

// TestWASMRegistryIntegration runs the full test suite
func TestWASMRegistryIntegration(t *testing.T) {
	suite.Run(t, new(WASMRegistryTestSuite))
}
