// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"
	"math/big"

	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
	economicsecuritypb "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
)

// QueryServer implements the economicsecurity Query service
type QueryServer struct {
	economicsecuritypb.UnimplementedQueryServer
	keeper *Keeper
}

// NewQueryServer returns a new QueryServer
func NewQueryServer(k *Keeper) economicsecuritypb.QueryServer {
	return &QueryServer{keeper: k}
}

var _ economicsecuritypb.QueryServer = &QueryServer{}

// Params returns the module parameters
func (qs *QueryServer) Params(ctx context.Context, req *economicsecuritypb.QueryParamsRequest) (*economicsecuritypb.QueryParamsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	params, _ := qs.keeper.GetParams(ctx)
	return &economicsecuritypb.QueryParamsResponse{
		Params: &params,
	}, nil
}

// VestingSchedule returns a vesting schedule by ID with calculated vested amounts
func (qs *QueryServer) VestingSchedule(ctx context.Context, req *economicsecuritypb.QueryVestingScheduleRequest) (*economicsecuritypb.QueryVestingScheduleResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	if req.ScheduleId == "" {
		return nil, fmt.Errorf("schedule_id cannot be empty")
	}

	schedule, err := qs.keeper.GetVestingSchedule(ctx, req.ScheduleId)
	if err != nil {
		return nil, err
	}

	if schedule == nil {
		return nil, types.ErrVestingScheduleNotFound
	}

	// Get current time for vesting calculation
	currentTime, err := qs.keeper.GetCurrentTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current time: %w", err)
	}

	// Calculate vested amount, remaining amount, and next vest time
	vestedAmount, remainingAmount, nextVestTime := calculateVestingDetails(schedule, currentTime)

	// Convert time.Time to gogoproto Timestamp
	nextVestTimeProto := &gogotypes.Timestamp{
		Seconds: nextVestTime,
		Nanos:   0,
	}

	return &economicsecuritypb.QueryVestingScheduleResponse{
		Schedule:        schedule,
		VestedAmount:    vestedAmount,
		RemainingAmount: remainingAmount,
		NextVestTime:    nextVestTimeProto,
	}, nil
}

// VestingSchedulesByBeneficiary returns all vesting schedules for a beneficiary
func (qs *QueryServer) VestingSchedulesByBeneficiary(ctx context.Context, req *economicsecuritypb.QueryVestingSchedulesByBeneficiaryRequest) (*economicsecuritypb.QueryVestingSchedulesByBeneficiaryResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	if req.BeneficiaryAddress == "" {
		return nil, fmt.Errorf("beneficiary_address cannot be empty")
	}

	// Get schedule IDs for beneficiary
	scheduleIDs, err := qs.keeper.GetUserVestingIndex(ctx, req.BeneficiaryAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get user vesting index: %w", err)
	}

	// Retrieve all schedules
	schedules := make([]*types.VestingSchedule, 0, len(scheduleIDs))
	totalVested := big.NewInt(0)
	totalVesting := big.NewInt(0)

	for _, scheduleID := range scheduleIDs {
		schedule, err := qs.keeper.GetVestingSchedule(ctx, scheduleID)
		if err != nil {
			// Skip schedules that can't be retrieved (defensive)
			continue
		}
		if schedule == nil {
			continue
		}

		schedules = append(schedules, schedule)

		// Accumulate totals
		if schedule.VestedAmount != "" {
			vested, ok := new(big.Int).SetString(schedule.VestedAmount, 10)
			if ok {
				totalVested.Add(totalVested, vested)
			}
		}

		if schedule.TotalAmount != "" {
			total, ok := new(big.Int).SetString(schedule.TotalAmount, 10)
			if ok {
				totalVesting.Add(totalVesting, total)
			}
		}
	}

	return &economicsecuritypb.QueryVestingSchedulesByBeneficiaryResponse{
		Schedules:    schedules,
		TotalVested:  totalVested.String(),
		TotalVesting: totalVesting.String(),
	}, nil
}

// VoteLock returns a vote lock by ID
func (qs *QueryServer) VoteLock(ctx context.Context, req *economicsecuritypb.QueryVoteLockRequest) (*economicsecuritypb.QueryVoteLockResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	if req.LockId == "" {
		return nil, fmt.Errorf("lock_id cannot be empty")
	}

	lock, err := qs.keeper.GetVoteLock(ctx, req.LockId)
	if err != nil {
		return nil, err
	}

	if lock == nil {
		return nil, types.ErrVoteLockNotFound
	}

	return &economicsecuritypb.QueryVoteLockResponse{
		Lock: lock,
	}, nil
}

// VoteLocksByOwner returns all vote locks for an owner
func (qs *QueryServer) VoteLocksByOwner(ctx context.Context, req *economicsecuritypb.QueryVoteLocksByOwnerRequest) (*economicsecuritypb.QueryVoteLocksByOwnerResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	if req.Owner == "" {
		return nil, fmt.Errorf("owner cannot be empty")
	}

	// Get lock IDs for owner
	lockIDs, err := qs.keeper.GetUserVoteLockIndex(ctx, req.Owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get user vote lock index: %w", err)
	}

	// Retrieve all locks
	locks := make([]*types.VoteLock, 0, len(lockIDs))
	totalLocked := big.NewInt(0)
	totalVotingPower := big.NewInt(0)

	for _, lockID := range lockIDs {
		lock, err := qs.keeper.GetVoteLock(ctx, lockID)
		if err != nil {
			// Skip locks that can't be retrieved (defensive)
			continue
		}
		if lock == nil || lock.Withdrawn {
			continue
		}

		locks = append(locks, lock)

		// Accumulate totals
		if lock.Amount != "" {
			amount, ok := new(big.Int).SetString(lock.Amount, 10)
			if ok {
				totalLocked.Add(totalLocked, amount)
			}
		}

		if lock.VotingPower != "" {
			power, ok := new(big.Int).SetString(lock.VotingPower, 10)
			if ok {
				totalVotingPower.Add(totalVotingPower, power)
			}
		}
	}

	return &economicsecuritypb.QueryVoteLocksByOwnerResponse{
		Locks:            locks,
		TotalLocked:      totalLocked.String(),
		TotalVotingPower: totalVotingPower.String(),
	}, nil
}

// VotingPower returns the voting power for an address
func (qs *QueryServer) VotingPower(ctx context.Context, req *economicsecuritypb.QueryVotingPowerRequest) (*economicsecuritypb.QueryVotingPowerResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	if req.Address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	// Get lock IDs for address
	lockIDs, err := qs.keeper.GetUserVoteLockIndex(ctx, req.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to get user vote lock index: %w", err)
	}

	totalVotingPower := big.NewInt(0)
	totalLocked := big.NewInt(0)
	activeLocks := uint64(0)

	// Aggregate voting power from all active locks
	for _, lockID := range lockIDs {
		lock, err := qs.keeper.GetVoteLock(ctx, lockID)
		if err != nil || lock == nil || lock.Withdrawn {
			continue
		}

		activeLocks++

		if lock.Amount != "" {
			amount, ok := new(big.Int).SetString(lock.Amount, 10)
			if ok {
				totalLocked.Add(totalLocked, amount)
			}
		}

		if lock.VotingPower != "" {
			power, ok := new(big.Int).SetString(lock.VotingPower, 10)
			if ok {
				totalVotingPower.Add(totalVotingPower, power)
			}
		}
	}

	return &economicsecuritypb.QueryVotingPowerResponse{
		VotingPower:  totalVotingPower.String(),
		LockedAmount: totalLocked.String(),
		ActiveLocks:  activeLocks,
	}, nil
}

// PendingTreasuryTx returns a pending treasury transaction
func (qs *QueryServer) PendingTreasuryTx(ctx context.Context, req *economicsecuritypb.QueryPendingTreasuryTxRequest) (*economicsecuritypb.QueryPendingTreasuryTxResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	if req.TxId == "" {
		return nil, fmt.Errorf("tx_id cannot be empty")
	}

	tx, err := qs.keeper.GetPendingTreasuryTx(ctx, req.TxId)
	if err != nil {
		return nil, err
	}

	if tx == nil {
		return nil, types.ErrTxNotFound
	}

	return &economicsecuritypb.QueryPendingTreasuryTxResponse{
		Transaction: tx,
	}, nil
}

// PendingTreasuryTxs returns all pending treasury transactions
func (qs *QueryServer) PendingTreasuryTxs(ctx context.Context, req *economicsecuritypb.QueryPendingTreasuryTxsRequest) (*economicsecuritypb.QueryPendingTreasuryTxsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	transactions := make([]*types.PendingTreasuryTx, 0)

	// Iterate through all pending transactions
	err := qs.keeper.IteratePendingTreasuryTxs(ctx, func(tx *types.PendingTreasuryTx) bool {
		// Only include non-executed, non-rejected transactions
		if !tx.Executed && !tx.Rejected {
			transactions = append(transactions, tx)
		}
		return false // Continue iteration
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate pending treasury txs: %w", err)
	}

	return &economicsecuritypb.QueryPendingTreasuryTxsResponse{
		Transactions: transactions,
	}, nil
}

// InflationMetrics returns current inflation metrics
func (qs *QueryServer) InflationMetrics(ctx context.Context, req *economicsecuritypb.QueryInflationMetricsRequest) (*economicsecuritypb.QueryInflationMetricsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	params, _ := qs.keeper.GetParams(ctx)
	if params.Tokenomics == nil {
		return nil, fmt.Errorf("tokenomics config not found")
	}

	// Get previous inflation for 24h change calculation
	previousInflation, err := qs.keeper.GetPreviousInflation(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous inflation: %w", err)
	}

	// Calculate 24h inflation change
	var inflationChange24h uint64
	if params.Tokenomics.InflationRate > previousInflation {
		inflationChange24h = params.Tokenomics.InflationRate - previousInflation
	} else {
		inflationChange24h = previousInflation - params.Tokenomics.InflationRate
	}

	// Calculate next check time
	currentHeight, err := qs.keeper.GetCurrentHeight(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current height: %w", err)
	}

	currentTime, err := qs.keeper.GetCurrentTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current time: %w", err)
	}

	// Estimate next check time (assuming ~6 second block time)
	blocksUntilCheck := params.InflationCheckInterval - (currentHeight % params.InflationCheckInterval)
	nextCheckTime := currentTime + int64(blocksUntilCheck*6)

	// Convert time.Time to gogoproto Timestamp
	lastAdjustmentProto := &gogotypes.Timestamp{
		Seconds: params.Tokenomics.LastInflationAdjustment.Unix(),
		Nanos:   int32(params.Tokenomics.LastInflationAdjustment.Nanosecond()),
	}
	nextCheckProto := &gogotypes.Timestamp{
		Seconds: nextCheckTime,
		Nanos:   0,
	}

	return &economicsecuritypb.QueryInflationMetricsResponse{
		CurrentInflationRate: params.Tokenomics.InflationRate,
		TargetInflationRate:  params.Tokenomics.TargetInflationRate,
		InflationChange_24H:  inflationChange24h,
		LastAdjustment:       lastAdjustmentProto,
		NextCheck:            nextCheckProto,
	}, nil
}

// InflationAlerts returns inflation alerts
func (qs *QueryServer) InflationAlerts(ctx context.Context, req *economicsecuritypb.QueryInflationAlertsRequest) (*economicsecuritypb.QueryInflationAlertsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	alerts := make([]*types.InflationAlert, 0)
	limit := req.Limit
	if limit == 0 {
		limit = 100 // Default limit
	}

	count := uint64(0)

	// Iterate through all inflation alerts
	err := qs.keeper.IterateInflationAlerts(ctx, func(alert *types.InflationAlert) bool {
		if count >= limit {
			return true // Stop iteration
		}

		alerts = append(alerts, alert)
		count++
		return false // Continue iteration
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate inflation alerts: %w", err)
	}

	return &economicsecuritypb.QueryInflationAlertsResponse{
		Alerts: alerts,
	}, nil
}

// LiquidityMiningStats returns liquidity mining statistics
func (qs *QueryServer) LiquidityMiningStats(ctx context.Context, req *economicsecuritypb.QueryLiquidityMiningStatsRequest) (*economicsecuritypb.QueryLiquidityMiningStatsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	params, _ := qs.keeper.GetParams(ctx)
	if params.LiquidityMining == nil {
		return nil, fmt.Errorf("liquidity mining config not found")
	}

	lm := params.LiquidityMining

	// Calculate remaining rewards
	allocated, ok1 := new(big.Int).SetString(lm.TotalRewardsAllocated, 10)
	distributed, ok2 := new(big.Int).SetString(lm.TotalRewardsDistributed, 10)

	remaining := "0"
	if ok1 && ok2 {
		remainingBig := new(big.Int).Sub(allocated, distributed)
		if remainingBig.Sign() > 0 {
			remaining = remainingBig.String()
		}
	}

	// Calculate current epoch rewards
	maxPerEpoch, ok3 := new(big.Int).SetString(lm.MaxRewardsPerEpoch, 10)
	rewardsThisEpoch := "0"
	if ok3 {
		rewardsThisEpoch = maxPerEpoch.String()
	}

	// Calculate next distribution height
	nextDistributionHeight := lm.LastDistributionHeight + lm.EpochDurationBlocks

	return &economicsecuritypb.QueryLiquidityMiningStatsResponse{
		Enabled:                lm.Enabled,
		TotalAllocated:         lm.TotalRewardsAllocated,
		TotalDistributed:       lm.TotalRewardsDistributed,
		RemainingRewards:       remaining,
		CurrentEpoch:           lm.CurrentEpoch,
		RewardsThisEpoch:       rewardsThisEpoch,
		NextDistributionHeight: nextDistributionHeight,
	}, nil
}

// MEVStats returns MEV redistribution statistics
func (qs *QueryServer) MEVStats(ctx context.Context, req *economicsecuritypb.QueryMEVStatsRequest) (*economicsecuritypb.QueryMEVStatsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	params, _ := qs.keeper.GetParams(ctx)
	if params.Mev == nil {
		return nil, fmt.Errorf("MEV config not found")
	}

	mev := params.Mev

	// Get total pending MEV
	pendingRedistribution, err := qs.keeper.GetTotalMEVPending(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total MEV pending: %w", err)
	}

	return &economicsecuritypb.QueryMEVStatsResponse{
		Enabled:                      mev.Enabled,
		TotalCaptured:                mev.TotalMevCaptured,
		TotalRedistributed:           mev.TotalMevRedistributed,
		PendingRedistribution:        pendingRedistribution,
		UserRedistributionPercentage: mev.UserRedistributionPercentage,
		Strategy:                     mev.Strategy,
	}, nil
}

// UserMEVBalance returns a user's MEV redistribution balance
func (qs *QueryServer) UserMEVBalance(ctx context.Context, req *economicsecuritypb.QueryUserMEVBalanceRequest) (*economicsecuritypb.QueryUserMEVBalanceResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	if req.Address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	balance, err := qs.keeper.GetUserMEVBalance(ctx, req.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to get user MEV balance: %w", err)
	}

	// For lifetime_received, we currently return the same as balance
	// In a production system, this would be tracked separately
	lifetimeReceived := balance

	return &economicsecuritypb.QueryUserMEVBalanceResponse{
		Balance:          balance,
		LifetimeReceived: lifetimeReceived,
	}, nil
}

// TokenomicsStats returns overall tokenomics statistics
func (qs *QueryServer) TokenomicsStats(ctx context.Context, req *economicsecuritypb.QueryTokenomicsStatsRequest) (*economicsecuritypb.QueryTokenomicsStatsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	params, _ := qs.keeper.GetParams(ctx)
	if params.Tokenomics == nil {
		return nil, fmt.Errorf("tokenomics config not found")
	}

	tokenomics := params.Tokenomics

	// Calculate total vested and vesting amounts
	totalVested := big.NewInt(0)
	totalVesting := big.NewInt(0)

	err := qs.keeper.IterateVestingSchedules(ctx, func(schedule *types.VestingSchedule) bool {
		if schedule.VestedAmount != "" {
			vested, ok := new(big.Int).SetString(schedule.VestedAmount, 10)
			if ok {
				totalVested.Add(totalVested, vested)
			}
		}

		if schedule.TotalAmount != "" {
			total, ok := new(big.Int).SetString(schedule.TotalAmount, 10)
			if ok {
				totalVesting.Add(totalVesting, total)
			}
		}
		return false // Continue iteration
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate vesting schedules: %w", err)
	}

	// Calculate total locked governance tokens
	totalLockedGovernance := big.NewInt(0)

	err = qs.keeper.IterateVoteLocks(ctx, func(lock *types.VoteLock) bool {
		if !lock.Withdrawn && lock.Amount != "" {
			amount, ok := new(big.Int).SetString(lock.Amount, 10)
			if ok {
				totalLockedGovernance.Add(totalLockedGovernance, amount)
			}
		}
		return false // Continue iteration
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate vote locks: %w", err)
	}

	// Get treasury balance (from params treasury address)
	treasuryBalance := "0"
	if params.TreasuryMultisig != nil && params.TreasuryMultisig.TreasuryAddress != "" {
		// In production, query bank module for actual balance
		// For now, we return 0 as placeholder
		treasuryBalance = "0"
	}

	// Get total burned
	totalBurned, err := qs.keeper.GetTotalBurned(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total burned: %w", err)
	}

	// Calculate whale protection triggers in last 24h
	whaleProtectionTriggers := uint64(0)
	currentTime, err := qs.keeper.GetCurrentTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current time: %w", err)
	}

	oneDayAgo := currentTime - 86400 // 24 hours in seconds

	err = qs.keeper.IterateLargeTxRecords(ctx, func(record *types.LargeTxRecord) bool {
		if !record.Timestamp.IsZero() && record.Timestamp.Unix() >= oneDayAgo {
			whaleProtectionTriggers++
		}
		return false // Continue iteration
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate large tx records: %w", err)
	}

	// Calculate transfer tax collected in last 24h
	// This would require additional tracking in production
	transferTaxCollected24h := "0"

	return &economicsecuritypb.QueryTokenomicsStatsResponse{
		MaxSupply:                   tokenomics.MaxSupply,
		CirculatingSupply:           tokenomics.CirculatingSupply,
		TotalVested:                 totalVested.String(),
		TotalVesting:                totalVesting.String(),
		TotalLockedGovernance:       totalLockedGovernance.String(),
		TreasuryBalance:             treasuryBalance,
		CurrentInflationRate:        tokenomics.InflationRate,
		TotalBurned:                 totalBurned,
		WhaleProtectionTriggers_24H: whaleProtectionTriggers,
		TransferTaxCollected_24H:    transferTaxCollected24h,
	}, nil
}

// Helper function to calculate vesting details
func calculateVestingDetails(schedule *types.VestingSchedule, currentTime int64) (vestedAmount, remainingAmount string, nextVestTime int64) {
	if schedule == nil || schedule.Revoked {
		return "0", "0", 0
	}

	// Parse total amount
	totalAmount, ok := new(big.Int).SetString(schedule.TotalAmount, 10)
	if !ok || totalAmount.Sign() <= 0 {
		return "0", schedule.TotalAmount, 0
	}

	// Calculate time-based vesting
	startTime := schedule.StartTime.Unix()
	cliffEnd := startTime + int64(schedule.CliffDuration)
	vestingEnd := startTime + int64(schedule.VestingDuration)

	// Before cliff: nothing vested
	if currentTime < cliffEnd {
		return "0", totalAmount.String(), cliffEnd
	}

	// After vesting period: fully vested
	if currentTime >= vestingEnd {
		return totalAmount.String(), "0", 0
	}

	// During vesting: calculate linear vesting
	elapsed := currentTime - startTime
	totalDuration := int64(schedule.VestingDuration)

	// Linear vesting calculation: (elapsed / total_duration) * total_amount
	vested := new(big.Int).Mul(totalAmount, big.NewInt(elapsed))
	vested.Div(vested, big.NewInt(totalDuration))

	// Remaining = total - vested
	remaining := new(big.Int).Sub(totalAmount, vested)
	if remaining.Sign() < 0 {
		remaining = big.NewInt(0)
	}

	// Next vest time is end of vesting period
	return vested.String(), remaining.String(), vestingEnd
}
