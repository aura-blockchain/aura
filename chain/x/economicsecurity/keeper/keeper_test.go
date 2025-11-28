package keeper

import (
	"math/big"
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/economicsecurity/params"
	"github.com/aequitas/aura/chain/x/economicsecurity/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestSupplyCapEnforcement tests Feature 1: Maximum supply cap enforcement
func TestSupplyCapEnforcement(t *testing.T) {
	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store)

	// Test within cap
	err := keeper.CheckSupplyCap("1000000")
	if err != nil {
		t.Fatalf("CheckSupplyCap should allow minting within cap: %v", err)
	}

	// Test exceeding cap
	err = keeper.CheckSupplyCap("2000000000000000")
	if err != types.ErrMaxSupplyExceeded {
		t.Fatalf("Expected ErrMaxSupplyExceeded, got: %v", err)
	}

	// Test updating circulating supply
	err = keeper.UpdateCirculatingSupply("1000000", true)
	if err != nil {
		t.Fatalf("UpdateCirculatingSupply failed: %v", err)
	}

	p := keeper.GetParams()
	// Default circulating supply is "100000000", adding "1000000" = "101000000"
	if p.Tokenomics.CirculatingSupply != "101000000" {
		t.Fatalf("Expected circulating supply 101000000, got: %s", p.Tokenomics.CirculatingSupply)
	}

	// Test remaining supply
	remaining := keeper.GetRemainingSupply()
	expected := new(big.Int)
	// MaxSupply "1000000000" - CirculatingSupply "101000000" = "899000000"
	expected.SetString("899000000", 10)
	actual := new(big.Int)
	actual.SetString(remaining, 10)

	if expected.Cmp(actual) != 0 {
		t.Fatalf("Expected remaining supply %s, got: %s", expected.String(), actual.String())
	}
}

// TestInflationMonitoring tests Feature 2: Inflation rate monitoring and alerts
func TestInflationMonitoring(t *testing.T) {
	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store)
	keeper.SetCurrentTime(time.Now().Unix())

	// Test normal inflation check
	err := keeper.CheckInflation()
	if err != nil {
		t.Fatalf("CheckInflation should not error on valid rate: %v", err)
	}

	// Test inflation above max
	p := keeper.GetParams()
	// Default max is 2000 (20%), set to 2500 (25%) to exceed max
	p.Tokenomics.InflationRate = 2500 // 25% (above max of 20%)
	keeper.SetParams(p)

	err = keeper.CheckInflation()
	if err != types.ErrInflationRateTooHigh {
		t.Fatalf("Expected ErrInflationRateTooHigh, got: %v", err)
	}

	// Check that alert was created
	alerts := keeper.GetInflationAlerts(10)
	if len(alerts) == 0 {
		t.Fatalf("Expected inflation alert to be created")
	}

	// Test adjusting inflation rate
	err = keeper.AdjustInflationRate(700, "test adjustment")
	if err != nil {
		t.Fatalf("AdjustInflationRate failed: %v", err)
	}

	p = keeper.GetParams()
	if p.Tokenomics.InflationRate != 700 {
		t.Fatalf("Expected inflation rate 700, got: %d", p.Tokenomics.InflationRate)
	}
}

// TestLiquidityMiningCaps tests Feature 3: Liquidity mining reward caps
func TestLiquidityMiningCaps(t *testing.T) {
	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store)
	keeper.SetCurrentHeight(100001)

	// Test checking reward cap
	// Default TotalRewardsAllocated is "1000000" (1M), so 500k is within cap
	err := keeper.CheckLiquidityRewardCap("500000") // 500k tokens
	if err != nil {
		t.Fatalf("CheckLiquidityRewardCap should allow amount within cap: %v", err)
	}

	// Test exceeding total allocated cap
	// Default TotalRewardsAllocated is "1000000" (1M), so 2M exceeds it
	err = keeper.CheckLiquidityRewardCap("2000000") // 2M tokens (exceeds 1M allocated)
	if err != types.ErrLiquidityRewardCapExceeded {
		t.Fatalf("Expected ErrLiquidityRewardCapExceeded, got: %v", err)
	}

	// Test distributing rewards
	// Keep amounts small to fit within MaxRewardsPerEpoch of "1000"
	recipients := map[string]string{
		"addr1": "300", // 300 tokens
		"addr2": "200", // 200 tokens
	}
	irVerified := map[string]bool{
		"addr1": true, // addr1 is IR-verified, gets 1.5x multiplier
		"addr2": false,
	}

	err = keeper.DistributeLiquidityRewards(recipients, irVerified)
	if err != nil {
		t.Fatalf("DistributeLiquidityRewards failed: %v", err)
	}

	// Verify stats
	enabled, _, distributed, _, epoch, _ := keeper.GetLiquidityMiningStats()
	if !enabled {
		t.Fatalf("Expected liquidity mining to be enabled")
	}
	if epoch != 1 {
		t.Fatalf("Expected epoch 1, got: %d", epoch)
	}

	distAmt := new(big.Int)
	distAmt.SetString(distributed, 10)
	if distAmt.Cmp(big.NewInt(0)) <= 0 {
		t.Fatalf("Expected distributed amount > 0, got: %s", distributed)
	}
}

// TestVestingSchedules tests Feature 4: Vesting schedules
func TestVestingSchedules(t *testing.T) {
	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store)
	currentTime := time.Now().Unix()
	keeper.SetCurrentTime(currentTime)

	// Create vesting schedule
	startTime := timestamppb.New(time.Unix(currentTime-86400, 0)) // Started 1 day ago
	scheduleID, err := keeper.CreateVestingSchedule(
		"beneficiary1",
		"10000000000", // 10k tokens
		startTime,
		0,       // No cliff
		2592000, // 30 days vesting
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	if err != nil {
		t.Fatalf("CreateVestingSchedule failed: %v", err)
	}

	// Verify schedule created
	schedule, ok := keeper.GetVestingSchedule(scheduleID)
	if !ok {
		t.Fatalf("Schedule not found")
	}
	if schedule.BeneficiaryAddress != "beneficiary1" {
		t.Fatalf("Expected beneficiary1, got: %s", schedule.BeneficiaryAddress)
	}

	// Try to release vested tokens (should have ~1/30 vested after 1 day)
	amount, err := keeper.ReleaseVestedTokens("beneficiary1", scheduleID)
	if err != nil {
		t.Fatalf("ReleaseVestedTokens failed: %v", err)
	}

	releasedAmt := new(big.Int)
	releasedAmt.SetString(amount, 10)
	if releasedAmt.Cmp(big.NewInt(0)) <= 0 {
		t.Fatalf("Expected some tokens to be vested")
	}

	// Test revoking schedule
	unvested, err := keeper.RevokeVestingSchedule(scheduleID, "test revocation")
	if err != nil {
		t.Fatalf("RevokeVestingSchedule failed: %v", err)
	}

	unvestedAmt := new(big.Int)
	unvestedAmt.SetString(unvested, 10)
	if unvestedAmt.Cmp(big.NewInt(0)) <= 0 {
		t.Fatalf("Expected some unvested amount to be returned")
	}
}

// TestWhaleProtection tests Feature 5: Anti-whale mechanisms
func TestWhaleProtection(t *testing.T) {
	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store)
	keeper.SetCurrentTime(time.Now().Unix())

	// Update circulating supply to a known amount
	// Start fresh - default is 100000000, let's set to a round number
	p := keeper.GetParams()
	p.Tokenomics.CirculatingSupply = "100000000000" // 100B tokens
	keeper.SetParams(p)

	// Test normal transfer (below 1% threshold)
	// MaxTxPercentage is 100 basis points (1%), so max tx = 100000000000 * 100 / 10000 = 1000000000
	err := keeper.CheckWhaleProtection("sender1", "recipient1", "500000000") // 500M tokens (0.5%)
	if err != nil {
		t.Fatalf("CheckWhaleProtection should allow normal transfer: %v", err)
	}

	// Test transfer exceeding single tx limit (1% of supply)
	// Sending 2B tokens which is 2% of supply, exceeds MaxTxPercentage of 1%
	err = keeper.CheckWhaleProtection("sender1", "recipient1", "2000000000") // 2B tokens (2%)
	if err != types.ErrWhaleTxLimitExceeded {
		t.Fatalf("Expected ErrWhaleTxLimitExceeded, got: %v", err)
	}

	// Test holding limit
	// MaxHoldingPercentage is 500 basis points (5%), so max holding = 100000000000 * 500 / 10000 = 5000000000
	keeper.UpdateAddressHolding("whale1", "5000000000")                 // 5B tokens (5% of supply - at limit)
	err = keeper.CheckWhaleProtection("sender1", "whale1", "100000000") // Try to send 100M more (would exceed 5%)
	if err != types.ErrWhaleHoldingLimitExceeded {
		t.Fatalf("Expected ErrWhaleHoldingLimitExceeded, got: %v", err)
	}

	// Test large tx cooldown
	// LargeTxThreshold is 100 basis points (1%), MaxTxPercentage is also 100 (1%)
	// Since the check uses > for largeTxThreshold (line 50 in whale_protection.go)
	// We need to send an amount > largeTxThreshold but <= maxTxAmount
	// This is only possible if we temporarily adjust the limits
	p = keeper.GetParams()
	p.WhaleProtection.MaxTxPercentage = 200  // Increase max tx to 2%
	p.WhaleProtection.LargeTxThreshold = 100 // Keep large tx threshold at 1%
	keeper.SetParams(p)

	// Now 1B tokens (1%) will trigger large tx tracking but not exceed max tx limit
	err = keeper.CheckWhaleProtection("sender2", "recipient1", "1000000001") // 1B + 1 tokens (just over 1%)
	if err != nil {
		t.Fatalf("First large tx should succeed: %v", err)
	}

	err = keeper.CheckWhaleProtection("sender2", "recipient1", "1000000001") // Second large tx immediately
	if err != types.ErrLargeTxCooldownActive {
		t.Fatalf("Expected ErrLargeTxCooldownActive, got: %v", err)
	}
}

// TestTransferTax tests Feature 6: Transfer tax options
func TestTransferTax(t *testing.T) {
	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store)

	// Enable transfer tax
	p := keeper.GetParams()
	p.TransferTax.Enabled = true
	p.TransferTax.BaseTaxRate = 100 // 1%
	keeper.SetParams(p)

	// Calculate tax for transfer
	tax, burn, treasury, err := keeper.CalculateTransferTax("sender1", "10000000000") // 10k tokens
	if err != nil {
		t.Fatalf("CalculateTransferTax failed: %v", err)
	}

	taxAmt := new(big.Int)
	taxAmt.SetString(tax, 10)

	expected := big.NewInt(100000000) // 1% of 10k = 100 tokens
	if taxAmt.Cmp(expected) != 0 {
		t.Fatalf("Expected tax %s, got: %s", expected.String(), taxAmt.String())
	}

	// Verify distribution
	burnAmt := new(big.Int)
	burnAmt.SetString(burn, 10)

	treasuryAmt := new(big.Int)
	treasuryAmt.SetString(treasury, 10)

	// Should be 50% burn, 30% treasury based on default config
	expectedBurn := big.NewInt(50000000) // 50% of 100 tokens
	if burnAmt.Cmp(expectedBurn) != 0 {
		t.Fatalf("Expected burn %s, got: %s", expectedBurn.String(), burnAmt.String())
	}

	// Process tax
	err = keeper.ProcessTransferTax(burn, treasury)
	if err != nil {
		t.Fatalf("ProcessTransferTax failed: %v", err)
	}
}

// TestGovernanceStakeRequirement tests Feature 7: Minimum stake requirements
func TestGovernanceStakeRequirement(t *testing.T) {
	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store)

	// Test insufficient stake
	err := keeper.CheckProposalStake("proposer1", "100") // 100 tokens (below minimum of 1000)
	if err != types.ErrInsufficientStake {
		t.Fatalf("Expected ErrInsufficientStake, got: %v", err)
	}

	// Test sufficient stake
	err = keeper.CheckProposalStake("proposer1", "10000000000") // 10k tokens
	if err != nil {
		t.Fatalf("CheckProposalStake should allow sufficient stake: %v", err)
	}
}

// TestQuadraticVoting tests Feature 8: Quadratic voting
func TestQuadraticVoting(t *testing.T) {
	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store)

	// Enable quadratic voting
	p := keeper.GetParams()
	p.Governance.QuadraticVotingEnabled = true
	keeper.SetParams(p)

	// Test quadratic voting power calculation
	votingPower, err := keeper.CalculateQuadraticVotingPower("10000000000") // 10k tokens
	if err != nil {
		t.Fatalf("CalculateQuadraticVotingPower failed: %v", err)
	}

	power := new(big.Int)
	power.SetString(votingPower, 10)

	// sqrt(10000000000) = 100000
	expected := big.NewInt(100000)
	if power.Cmp(expected) != 0 {
		t.Fatalf("Expected voting power %s, got: %s", expected.String(), power.String())
	}
}

// TestVoteLocking tests Feature 9: Vote locking mechanisms
func TestVoteLocking(t *testing.T) {
	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store)
	keeper.SetCurrentTime(time.Now().Unix())

	// Lock tokens for 1 year
	oneYear := uint64(31536000)
	lockID, votingPower, err := keeper.LockVotingTokens("voter1", "10000000000", oneYear)
	if err != nil {
		t.Fatalf("LockVotingTokens failed: %v", err)
	}

	if lockID == "" {
		t.Fatalf("Expected non-empty lock ID")
	}

	// Verify voting power is boosted (should be 2x for 1 year lock)
	power := new(big.Int)
	power.SetString(votingPower, 10)
	originalAmount := big.NewInt(10000000000)

	if power.Cmp(originalAmount) <= 0 {
		t.Fatalf("Expected voting power to be boosted, got: %s", power.String())
	}

	// Try to unlock before time
	_, err = keeper.UnlockVotingTokens("voter1", lockID)
	if err != types.ErrVoteLockNotExpired {
		t.Fatalf("Expected ErrVoteLockNotExpired, got: %v", err)
	}

	// Fast forward time
	keeper.SetCurrentTime(keeper.currentTime + int64(oneYear) + 1)

	// Now unlock should work
	amount, err := keeper.UnlockVotingTokens("voter1", lockID)
	if err != nil {
		t.Fatalf("UnlockVotingTokens failed: %v", err)
	}

	if amount != "10000000000" {
		t.Fatalf("Expected amount 10000000000, got: %s", amount)
	}
}

// TestTreasuryMultisig tests Feature 10: Treasury multi-signature controls
func TestTreasuryMultisig(t *testing.T) {
	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store)
	keeper.SetCurrentTime(time.Now().Unix())

	// Setup treasury multisig
	p := keeper.GetParams()
	p.TreasuryMultisig.TreasuryAddress = "treasury1"
	p.TreasuryMultisig.Threshold = 3
	p.TreasuryMultisig.Signers = []string{"signer1", "signer2", "signer3", "signer4"}
	keeper.SetParams(p)

	// Propose spend
	txID, execTime, err := keeper.ProposeTreasurySpend("signer1", "recipient1", "1000000000", "test spend")
	if err != nil {
		t.Fatalf("ProposeTreasurySpend failed: %v", err)
	}

	if txID == "" {
		t.Fatalf("Expected non-empty tx ID")
	}

	// Sign by second signer
	currentSigs, required, err := keeper.SignTreasurySpend("signer2", txID)
	if err != nil {
		t.Fatalf("SignTreasurySpend failed: %v", err)
	}

	if currentSigs != 2 {
		t.Fatalf("Expected 2 signatures, got: %d", currentSigs)
	}

	if required != 3 {
		t.Fatalf("Expected threshold 3, got: %d", required)
	}

	// Try to execute before threshold
	err = keeper.ExecuteTreasurySpend("signer1", txID, "10000000000")
	if err != types.ErrInsufficientSignatures {
		t.Fatalf("Expected ErrInsufficientSignatures, got: %v", err)
	}

	// Add third signature
	keeper.SignTreasurySpend("signer3", txID)

	// Try to execute before timelock
	err = keeper.ExecuteTreasurySpend("signer1", txID, "10000000000")
	if err != types.ErrTimelockNotExpired {
		t.Fatalf("Expected ErrTimelockNotExpired, got: %v", err)
	}

	// Fast forward past timelock
	keeper.SetCurrentTime(execTime.AsTime().Unix() + 1)

	// Now execution should work
	err = keeper.ExecuteTreasurySpend("signer1", txID, "10000000000")
	if err != nil {
		t.Fatalf("ExecuteTreasurySpend failed: %v", err)
	}

	// Verify executed
	tx, ok := keeper.GetPendingTreasuryTx(txID)
	if !ok {
		t.Fatalf("Transaction not found")
	}
	if !tx.Executed {
		t.Fatalf("Expected transaction to be executed")
	}
}

// TestDynamicFees tests Feature 11: Dynamic fee adjustment
func TestDynamicFees(t *testing.T) {
	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store)

	// Record high utilization
	for i := 0; i < 10; i++ {
		keeper.RecordBlockUtilization(9000) // 90% utilization
	}

	// Adjust fees
	err := keeper.AdjustDynamicFees()
	if err != nil {
		t.Fatalf("AdjustDynamicFees failed: %v", err)
	}

	// Fee multiplier should have increased
	multiplier := keeper.GetCurrentFeeMultiplier()
	if multiplier <= 10000 { // Should be higher than base 1x
		t.Fatalf("Expected fee multiplier to increase, got: %d", multiplier)
	}

	// Record low utilization
	for i := 0; i < 20; i++ {
		keeper.RecordBlockUtilization(2000) // 20% utilization
	}

	keeper.AdjustDynamicFees()

	// Fee multiplier should have decreased
	newMultiplier := keeper.GetCurrentFeeMultiplier()
	if newMultiplier >= multiplier {
		t.Fatalf("Expected fee multiplier to decrease")
	}

	// Calculate dynamic fee
	fee := keeper.CalculateDynamicFee()
	feeAmt := new(big.Int)
	feeAmt.SetString(fee, 10)

	if feeAmt.Cmp(big.NewInt(0)) <= 0 {
		t.Fatalf("Expected positive fee")
	}
}

// TestMEVRedistribution tests Feature 12: MEV redistribution
func TestMEVRedistribution(t *testing.T) {
	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store)

	// Capture MEV
	err := keeper.CaptureMEV("10000000000") // 10k tokens
	if err != nil {
		t.Fatalf("CaptureMEV failed: %v", err)
	}

	// Distribute MEV
	activeUsers := []string{"user1", "user2", "user3"}
	userActivity := map[string]uint64{
		"user1": 100,
		"user2": 200,
		"user3": 150,
	}
	userIRScores := map[string]uint64{
		"user1": 800,
		"user2": 900,
		"user3": 850,
	}

	validatorShare, _, _, err := keeper.DistributeMEV(activeUsers, userActivity, userIRScores)
	if err != nil {
		t.Fatalf("DistributeMEV failed: %v", err)
	}

	// Verify shares are distributed
	valAmt := new(big.Int)
	valAmt.SetString(validatorShare, 10)
	if valAmt.Cmp(big.NewInt(0)) <= 0 {
		t.Fatalf("Expected validator share > 0")
	}

	// Check user balances
	balance := keeper.GetUserMEVBalance("user1")
	balAmt := new(big.Int)
	balAmt.SetString(balance, 10)
	if balAmt.Cmp(big.NewInt(0)) <= 0 {
		t.Fatalf("Expected user1 to have MEV balance")
	}

	// Claim rewards
	claimed, err := keeper.ClaimMEVRewards("user1")
	if err != nil {
		t.Fatalf("ClaimMEVRewards failed: %v", err)
	}

	claimedAmt := new(big.Int)
	claimedAmt.SetString(claimed, 10)
	if claimedAmt.Cmp(balAmt) != 0 {
		t.Fatalf("Expected to claim full balance")
	}

	// Balance should be zero after claim
	newBalance := keeper.GetUserMEVBalance("user1")
	if newBalance != "0" {
		t.Fatalf("Expected balance to be 0 after claim, got: %s", newBalance)
	}

	// Verify stats
	enabled, captured, _, _, _, _ := keeper.GetMEVStats()
	if !enabled {
		t.Fatalf("Expected MEV redistribution to be enabled")
	}
	if captured == "0" {
		t.Fatalf("Expected captured MEV > 0")
	}
}
