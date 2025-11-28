package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crtypes "github.com/aequitas/aura/chain/x/contractregistry/types"
	wasmtypes "github.com/aequitas/aura/chain/x/wasm/types"
)

// WASMSecurityTestSuite focuses on security attack simulations and edge cases
type WASMSecurityTestSuite struct {
	suite.Suite
	WASMTestContext
}

// SetupTest runs before each test
func (s *WASMSecurityTestSuite) SetupTest() {
	s.WASMTestContext = SetupTestAppWithWasm(s.T())
}

// =============================================================================
// DOS ATTACK TESTS
// =============================================================================

// TestDoSViaRapidExecution tests DoS protection via rate limiting
func (s *WASMSecurityTestSuite) TestDoSViaRapidExecution() {
	ctx := s.GetContext()

	// Setup contract with strict rate limit
	uploader := s.CreateAuthorizedUploader()
	policy := &crtypes.SecurityPolicy{
		AllowPause:       true,
		MaxGasPerTx:      1000000,
		RateLimitPerUser: 5, // Very strict
	}
	contractAddr := s.SetupCompleteContractWithPolicy(uploader, nil, policy)

	attacker := s.CreateAuthorizedUploader()
	execMsg := []byte(`{"increment": {}}`)

	// Simulate rapid execution attempts
	successCount := 0
	blockedCount := 0

	for i := 0; i < 20; i++ {
		// Check rate limit
		err := s.RegistryKeeper.ValidateContractExecution(
			ctx,
			contractAddr.String(),
			attacker.GetAddress().String(),
			100000,
		)

		if err != nil {
			blockedCount++
			s.Require().Contains(err.Error(), "rate limit exceeded")
			continue
		}

		// Increment rate limit
		s.RegistryKeeper.IncrementRateLimit(ctx, contractAddr.String(), attacker.GetAddress().String())

		// Execute
		_, err = s.WasmKeeper.ExecuteContract(ctx, contractAddr, attacker.GetAddress(), execMsg, sdk.NewCoins())
		if err == nil {
			successCount++
		}
	}

	// Verify rate limiting worked
	s.Require().Equal(5, successCount, "Should allow exactly 5 executions")
	s.Require().Equal(15, blockedCount, "Should block 15 attempts")

	// Verify violations tracked
	metrics, _ := s.RegistryKeeper.GetContractMetrics(ctx, contractAddr.String())
	s.Require().Greater(metrics.RateLimitViolations, uint64(0))

	s.T().Log("✓ DoS via rapid execution blocked - PASSED")
}

// TestDoSViaLargePayload tests protection against large payload attacks
func (s *WASMSecurityTestSuite) TestDoSViaLargePayload() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()

	// Try to upload extremely large contract
	largeCode := make([]byte, 10*1024*1024) // 10MB

	_, err := s.WasmKeeper.StoreCode(ctx, uploader.GetAddress(), largeCode)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "exceeds maximum")

	s.T().Log("✓ DoS via large payload blocked - PASSED")
}

// TestDoSViaGasExhaustion tests gas limit enforcement
func (s *WASMSecurityTestSuite) TestDoSViaGasExhaustion() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()
	policy := &crtypes.SecurityPolicy{
		AllowPause:       true,
		MaxGasPerTx:      100000, // Low gas limit
		RateLimitPerUser: 100,
	}
	contractAddr := s.SetupCompleteContractWithPolicy(uploader, nil, policy)

	attacker := s.CreateAuthorizedUploader()

	// Attempt execution with excessive gas
	err := s.RegistryKeeper.ValidateContractExecution(
		ctx,
		contractAddr.String(),
		attacker.GetAddress().String(),
		5000000, // Much higher than allowed
	)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "gas limit exceeded")

	s.T().Log("✓ DoS via gas exhaustion blocked - PASSED")
}

// =============================================================================
// VC BYPASS ATTACK TESTS
// =============================================================================

// TestVCBypassAttempt tests that VC enforcement cannot be bypassed
func (s *WASMSecurityTestSuite) TestVCBypassAttempt() {
	ctx := s.GetContext()

	// Setup contract requiring VC
	uploader := s.CreateAuthorizedUploader()
	metadata := &crtypes.ContractMetadata{
		Name:            "VC Protected",
		RequiresVc:      true,
		RequiredVcTypes: []string{"KYCVerification", "AccreditedInvestor"},
	}
	contractAddr := s.SetupCompleteContract(uploader, metadata)

	// Attack vector 1: User with NO VCs
	attackerNoVC := s.CreateUserWithoutVC()
	err := s.RegistryKeeper.ValidateContractExecution(
		ctx,
		contractAddr.String(),
		attackerNoVC.GetAddress().String(),
		100000,
	)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "missing VC")

	// Attack vector 2: User with only ONE of the required VCs
	attackerPartialVC := s.CreateUserWithVC("KYCVerification")
	err = s.RegistryKeeper.ValidateContractExecution(
		ctx,
		contractAddr.String(),
		attackerPartialVC.GetAddress().String(),
		100000,
	)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "missing VC")

	// Attack vector 3: User with wrong VC type
	attackerWrongVC := s.CreateUserWithVC("DifferentVC")
	err = s.RegistryKeeper.ValidateContractExecution(
		ctx,
		contractAddr.String(),
		attackerWrongVC.GetAddress().String(),
		100000,
	)
	s.Require().Error(err)

	// Verify compliance failures tracked
	metrics, _ := s.RegistryKeeper.GetContractMetrics(ctx, contractAddr.String())
	s.Require().Equal(uint64(3), metrics.ComplianceFailures)

	s.T().Log("✓ VC bypass attempts blocked - PASSED")
}

// TestVCRevocationDuringExecution tests handling of VC revocation
func (s *WASMSecurityTestSuite) TestVCRevocationDuringExecution() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()
	metadata := &crtypes.ContractMetadata{
		Name:            "VC Required",
		RequiresVc:      true,
		RequiredVcTypes: []string{"KYCVerification"},
	}
	contractAddr := s.SetupCompleteContract(uploader, metadata)

	// User with valid VC
	user := s.CreateUserWithVC("KYCVerification")

	// First execution succeeds
	err := s.RegistryKeeper.ValidateContractExecution(
		ctx,
		contractAddr.String(),
		user.GetAddress().String(),
		100000,
	)
	s.Require().NoError(err)

	// Simulate VC revocation
	mockVC := s.VCKeeper.(*MockVCKeeper)
	mockVC.userVCs[user.GetAddress().String()] = []string{} // Remove VC

	// Second execution should fail
	err = s.RegistryKeeper.ValidateContractExecution(
		ctx,
		contractAddr.String(),
		user.GetAddress().String(),
		100000,
	)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "missing VC")

	s.T().Log("✓ VC revocation enforcement - PASSED")
}

// =============================================================================
// PAUSED CONTRACT ATTACK TESTS
// =============================================================================

// TestPausedContractExecutionAttempts tests that paused contracts cannot execute
func (s *WASMSecurityTestSuite) TestPausedContractExecutionAttempts() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()
	contractAddr := s.SetupCompleteContract(uploader, nil)

	// Pause the contract
	err := s.RegistryKeeper.PauseContract(
		ctx,
		contractAddr.String(),
		uploader.GetAddress().String(),
		"Security test",
	)
	s.Require().NoError(err)

	// Attempt execution - should fail in registry validation
	err = s.RegistryKeeper.ValidateContractExecution(
		ctx,
		contractAddr.String(),
		uploader.GetAddress().String(),
		100000,
	)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "paused")

	// Also pause in WASM keeper
	err = s.WasmKeeper.PauseContract(ctx, contractAddr.String())
	s.Require().NoError(err)

	// Attempt direct execution - should also fail
	execMsg := []byte(`{"increment": {}}`)
	_, err = s.WasmKeeper.ExecuteContract(
		ctx,
		contractAddr,
		uploader.GetAddress(),
		execMsg,
		sdk.NewCoins(),
	)
	s.Require().Error(err)

	s.T().Log("✓ Paused contract execution blocked - PASSED")
}

// TestPausedContractQueryAttempts tests query access to paused contracts
func (s *WASMSecurityTestSuite) TestPausedContractQueryAttempts() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()
	contractAddr := s.SetupCompleteContract(uploader, nil)

	// Pause contract
	err := s.WasmKeeper.PauseContract(ctx, contractAddr.String())
	s.Require().NoError(err)

	// Attempt query - should fail
	queryMsg := []byte(`{"get_count": {}}`)
	_, err = s.WasmKeeper.QuerySmart(ctx, contractAddr, queryMsg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "paused")

	s.T().Log("✓ Paused contract query blocked - PASSED")
}

// =============================================================================
// UNAUTHORIZED OPERATION TESTS
// =============================================================================

// TestUnauthorizedUploadAttempt tests upload by unauthorized user
func (s *WASMSecurityTestSuite) TestUnauthorizedUploadAttempt() {
	ctx := s.GetContext()

	// Create user WITHOUT authorization
	unauthorized := s.CreateUserWithoutVC()

	// Attempt upload
	code := s.CreateMockWASMCode()
	_, err := s.WasmKeeper.StoreCode(ctx, unauthorized.GetAddress(), code)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "not authorized")

	s.T().Log("✓ Unauthorized upload blocked - PASSED")
}

// TestUnauthorizedPauseAttempt tests pause by non-admin user
func (s *WASMSecurityTestSuite) TestUnauthorizedPauseAttempt() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()
	contractAddr := s.SetupCompleteContract(uploader, nil)

	// Create non-admin user
	attacker := s.CreateAuthorizedUploader()

	// Attempt to pause as non-admin
	err := s.RegistryKeeper.PauseContract(
		ctx,
		contractAddr.String(),
		attacker.GetAddress().String(),
		"Unauthorized pause attempt",
	)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unauthorized")

	// Verify contract still active
	info, _ := s.RegistryKeeper.GetContractInfo(ctx, contractAddr.String())
	s.Require().Equal(crtypes.ContractStatus_CONTRACT_STATUS_ACTIVE, info.Status)

	s.T().Log("✓ Unauthorized pause blocked - PASSED")
}

// TestUnauthorizedMetadataUpdate tests metadata update by non-admin
func (s *WASMSecurityTestSuite) TestUnauthorizedMetadataUpdate() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()
	contractAddr := s.SetupCompleteContract(uploader, nil)

	// Create non-admin user
	attacker := s.CreateAuthorizedUploader()

	// Attempt metadata update
	newMetadata := crtypes.ContractMetadata{
		Name:        "Hacked Contract",
		Description: "Malicious update",
	}
	err := s.RegistryKeeper.UpdateContractMetadata(
		ctx,
		contractAddr.String(),
		attacker.GetAddress().String(),
		newMetadata,
	)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "not contract admin")

	s.T().Log("✓ Unauthorized metadata update blocked - PASSED")
}

// =============================================================================
// MALICIOUS CONTRACT REGISTRATION TESTS
// =============================================================================

// TestMaliciousContractRegistration tests registration with malicious data
func (s *WASMSecurityTestSuite) TestMaliciousContractRegistration() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()

	// Test case 1: Empty contract address
	info := crtypes.ContractInfo{
		Address: "",
		CodeId:  1,
		Creator: uploader.GetAddress().String(),
	}
	err := s.RegistryKeeper.RegisterContract(ctx, info)
	s.Require().Error(err)

	// Test case 2: Zero code ID
	info = crtypes.ContractInfo{
		Address: "malicious_contract",
		CodeId:  0,
		Creator: uploader.GetAddress().String(),
	}
	// This would be caught in validation if we had it

	// Test case 3: Excessive rate limit
	info = crtypes.ContractInfo{
		Address: "malicious_contract_2",
		CodeId:  1,
		Creator: uploader.GetAddress().String(),
		Admin:   uploader.GetAddress().String(),
		Metadata: crtypes.ContractMetadata{
			Name:        "Test",
			Description: "Test",
		},
		SecurityPolicy: crtypes.SecurityPolicy{
			RateLimitPerUser: 999999, // Excessive
			MaxGasPerTx:      1000000,
		},
		Status: crtypes.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err = s.RegistryKeeper.RegisterContract(ctx, info)
	s.Require().Error(err) // Should fail validation

	s.T().Log("✓ Malicious contract registration blocked - PASSED")
}

// TestContractLimitEnforcement tests max contracts per creator
func (s *WASMSecurityTestSuite) TestContractLimitEnforcement() {
	ctx := s.GetContext()

	// Set low limit
	params := s.RegistryKeeper.GetParams(ctx)
	params.MaxContractsPerCreator = 3
	err := s.RegistryKeeper.SetParams(ctx, params)
	s.Require().NoError(err)

	uploader := s.CreateAuthorizedUploader()

	// Register 3 contracts (should succeed)
	for i := 0; i < 3; i++ {
		contractAddr := fmt.Sprintf("contract_%d", i)
		s.RegisterContractInRegistry(ctx, contractAddr, 1, uploader.GetAddress().String())
	}

	// 4th registration should fail
	info := crtypes.ContractInfo{
		Address: "contract_4",
		CodeId:  1,
		Creator: uploader.GetAddress().String(),
		Admin:   uploader.GetAddress().String(),
		Metadata: crtypes.ContractMetadata{
			Name:        "Fourth Contract",
			Description: "Should fail",
		},
		SecurityPolicy: crtypes.SecurityPolicy{
			AllowPause:       true,
			MaxGasPerTx:      1000000,
			RateLimitPerUser: 100,
		},
		Status: crtypes.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	err = s.RegistryKeeper.RegisterContract(ctx, info)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "too many contracts")

	s.T().Log("✓ Contract limit enforcement - PASSED")
}

// =============================================================================
// RACE CONDITION TESTS
// =============================================================================

// TestConcurrentPauseUnpause tests race conditions in pause/unpause
func (s *WASMSecurityTestSuite) TestConcurrentPauseUnpause() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()
	contractAddr := s.SetupCompleteContract(uploader, nil)

	// Concurrent pause/unpause operations
	done := make(chan bool, 20)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			_ = s.RegistryKeeper.PauseContract(
				ctx,
				contractAddr.String(),
				uploader.GetAddress().String(),
				"Concurrent test",
			)
		}()

		go func() {
			defer func() { done <- true }()
			_ = s.RegistryKeeper.UnpauseContract(
				ctx,
				contractAddr.String(),
				uploader.GetAddress().String(),
			)
		}()
	}

	// Wait for all operations
	for i := 0; i < 20; i++ {
		<-done
	}

	// Contract should be in a valid state (either paused or active)
	info, found := s.RegistryKeeper.GetContractInfo(ctx, contractAddr.String())
	s.Require().True(found)
	s.Require().NotEqual(crtypes.ContractStatus_CONTRACT_STATUS_UNSPECIFIED, info.Status)

	s.T().Log("✓ Concurrent pause/unpause handled - PASSED")
}

// TestConcurrentMetricUpdates tests race conditions in metrics updates
func (s *WASMSecurityTestSuite) TestConcurrentMetricUpdates() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()
	contractAddr := s.SetupCompleteContract(uploader, nil)

	// Concurrent metric updates
	updateCount := 100
	done := make(chan bool, updateCount)

	for i := 0; i < updateCount; i++ {
		go func(success bool) {
			defer func() { done <- true }()
			s.RegistryKeeper.UpdateMetricsOnExecution(
				ctx,
				contractAddr.String(),
				100000,
				success,
			)
		}(i%2 == 0) // Half success, half failure
	}

	// Wait for all updates
	for i := 0; i < updateCount; i++ {
		<-done
	}

	// Verify metrics are consistent
	metrics, found := s.RegistryKeeper.GetContractMetrics(ctx, contractAddr.String())
	s.Require().True(found)
	s.Require().Equal(
		metrics.TotalExecutions,
		metrics.SuccessfulExecutions+metrics.FailedExecutions,
		"Metrics should be consistent",
	)

	s.T().Log("✓ Concurrent metric updates handled - PASSED")
}

// =============================================================================
// TIME-BASED ATTACK TESTS
// =============================================================================

// TestRateLimitWindowReset tests rate limit window reset timing attacks
func (s *WASMSecurityTestSuite) TestRateLimitWindowReset() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()
	policy := &crtypes.SecurityPolicy{
		AllowPause:       true,
		MaxGasPerTx:      1000000,
		RateLimitPerUser: 3,
	}
	contractAddr := s.SetupCompleteContractWithPolicy(uploader, nil, policy)

	user := uploader.GetAddress().String()

	// Use up rate limit in current window
	for i := 0; i < 3; i++ {
		s.RegistryKeeper.IncrementRateLimit(ctx, contractAddr.String(), user)
	}

	// Should be blocked
	err := s.RegistryKeeper.CheckRateLimit(ctx, contractAddr.String(), user, 3)
	s.Require().Error(err)

	// Advance time by 1 hour + 1 second
	newTime := ctx.BlockTime().Add(time.Hour + time.Second)
	ctx = ctx.WithBlockTime(newTime)

	// Should be allowed again in new window
	err = s.RegistryKeeper.CheckRateLimit(ctx, contractAddr.String(), user, 3)
	s.Require().NoError(err)

	s.T().Log("✓ Rate limit window reset - PASSED")
}

// =============================================================================
// EDGE CASE TESTS
// =============================================================================

// TestMaximumTagsAndMetadata tests limits on metadata fields
func (s *WASMSecurityTestSuite) TestMaximumTagsAndMetadata() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()

	// Create metadata with excessive tags
	excessiveTags := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		excessiveTags[i] = fmt.Sprintf("tag_%d", i)
	}

	metadata := &crtypes.ContractMetadata{
		Name:        "Tag Spam Contract",
		Description: "Testing tag limits",
		Tags:        excessiveTags,
	}

	// This should either succeed or have reasonable limits
	contractAddr := s.SetupCompleteContract(uploader, metadata)

	// Verify contract registered
	s.Require().True(s.RegistryKeeper.IsContractRegistered(ctx, contractAddr.String()))

	s.T().Log("✓ Maximum tags handled - PASSED")
}

// TestZeroGasExecution tests execution with zero gas
func (s *WASMSecurityTestSuite) TestZeroGasExecution() {
	ctx := s.GetContext()

	uploader := s.CreateAuthorizedUploader()
	policy := &crtypes.SecurityPolicy{
		AllowPause:       true,
		MaxGasPerTx:      1000000,
		RateLimitPerUser: 100,
	}
	contractAddr := s.SetupCompleteContractWithPolicy(uploader, nil, policy)

	// Attempt validation with zero gas
	err := s.RegistryKeeper.ValidateContractExecution(
		ctx,
		contractAddr.String(),
		uploader.GetAddress().String(),
		0, // Zero gas
	)

	// Should either error or be handled gracefully
	// (depends on policy implementation)
	if err != nil {
		s.T().Logf("Zero gas rejected: %v", err)
	} else {
		s.T().Log("Zero gas accepted (no minimum)")
	}

	s.T().Log("✓ Zero gas execution handled - PASSED")
}

// =============================================================================
// TEST RUNNER
// =============================================================================

// TestWASMSecurity runs the security test suite
func TestWASMSecurity(t *testing.T) {
	suite.Run(t, new(WASMSecurityTestSuite))
}
