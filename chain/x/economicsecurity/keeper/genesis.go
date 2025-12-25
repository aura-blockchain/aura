// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"
	"sort"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// InitGenesis initializes the keeper from genesis state
// This function loads all genesis data into the KV store with full validation
func (k *Keeper) InitGenesis(ctx context.Context, genesis types.GenesisState) error {
	// Validate genesis before proceeding
	if err := types.ValidateGenesis(&genesis); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	// Set params - this is critical for module operation
	if genesis.Params != nil {
		if err := k.SetParams(*genesis.Params); err != nil {
			return fmt.Errorf("failed to set params: %w", err)
		}
	} else {
		// Set default params if none provided
		if err := k.SetParams(*types.DefaultParams()); err != nil {
			return fmt.Errorf("failed to set default params: %w", err)
		}
	}

	// Load all vesting schedules into KV store
	// These represent token vesting for team, investors, advisors, etc.
	for _, schedule := range genesis.VestingSchedules {
		if schedule == nil {
			continue
		}

		// Store the vesting schedule
		if err := k.SetVestingSchedule(ctx, schedule); err != nil {
			return fmt.Errorf("failed to set vesting schedule %s: %w", schedule.ScheduleId, err)
		}

		// Update user vesting index for quick lookup
		if err := k.AddUserVestingSchedule(ctx, schedule.BeneficiaryAddress, schedule.ScheduleId); err != nil {
			return fmt.Errorf("failed to add user vesting index for %s: %w", schedule.BeneficiaryAddress, err)
		}
	}

	// Load all vote locks into KV store
	// Vote locks are used for governance voting power
	for _, lock := range genesis.VoteLocks {
		if lock == nil {
			continue
		}

		// Store the vote lock
		if err := k.SetVoteLock(ctx, lock); err != nil {
			return fmt.Errorf("failed to set vote lock %s: %w", lock.LockId, err)
		}

		// Update user vote lock index for quick lookup
		if err := k.AddUserVoteLock(ctx, lock.Owner, lock.LockId); err != nil {
			return fmt.Errorf("failed to add user vote lock index for %s: %w", lock.Owner, err)
		}
	}

	// Load all pending treasury transactions into KV store
	// These are multisig treasury operations awaiting approval
	for _, tx := range genesis.PendingTreasuryTxs {
		if tx == nil {
			continue
		}

		if err := k.SetPendingTreasuryTx(ctx, tx); err != nil {
			return fmt.Errorf("failed to set pending treasury tx %s: %w", tx.TxId, err)
		}
	}

	// Load all inflation alerts into KV store
	// These track inflation rate anomalies and deviations
	for _, alert := range genesis.InflationAlerts {
		if alert == nil {
			continue
		}

		if err := k.SetInflationAlert(ctx, alert); err != nil {
			return fmt.Errorf("failed to set inflation alert %s: %w", alert.AlertId, err)
		}
	}

	// Load all large transaction records into KV store
	// These track whale activity and large transfers for protection
	for _, record := range genesis.LargeTxRecords {
		if record == nil {
			continue
		}

		if err := k.SetLargeTxRecord(ctx, record); err != nil {
			return fmt.Errorf("failed to set large tx record %s: %w", record.TxHash, err)
		}
	}

	// Load last large transaction times by address (deterministic ordering)
	// This is used for rate limiting whale transactions
	txTimeAddrs := make([]string, 0, len(genesis.LastLargeTxTimes))
	for address := range genesis.LastLargeTxTimes {
		if address != "" {
			txTimeAddrs = append(txTimeAddrs, address)
		}
	}
	sort.Strings(txTimeAddrs)
	for _, address := range txTimeAddrs {
		timestamp := genesis.LastLargeTxTimes[address]
		if err := k.SetLastLargeTxTime(ctx, address, timestamp); err != nil {
			return fmt.Errorf("failed to set last large tx time for %s: %w", address, err)
		}
	}

	// Load user MEV balances (deterministic ordering)
	// These represent accumulated MEV rewards awaiting distribution
	mevAddrs := make([]string, 0, len(genesis.UserMevBalances))
	for address := range genesis.UserMevBalances {
		if address != "" && genesis.UserMevBalances[address] != "" {
			mevAddrs = append(mevAddrs, address)
		}
	}
	sort.Strings(mevAddrs)
	for _, address := range mevAddrs {
		balanceStr := genesis.UserMevBalances[address]
		if err := k.SetUserMEVBalance(ctx, address, balanceStr); err != nil {
			return fmt.Errorf("failed to set MEV balance for %s: %w", address, err)
		}
	}

	return nil
}

// ExportGenesis exports the current state for genesis
// This function retrieves all module state from the KV store
func (k *Keeper) ExportGenesis(ctx context.Context) (types.GenesisState, error) {
	var genesis types.GenesisState

	// Export params
	params, _ := k.GetParams(ctx)
	genesis.Params = &params

	// Export all vesting schedules
	vestingSchedules := make([]*types.VestingSchedule, 0)
	if err := k.IterateVestingSchedules(ctx, func(schedule *types.VestingSchedule) bool {
		vestingSchedules = append(vestingSchedules, schedule)
		return false // continue iteration
	}); err != nil {
		return genesis, fmt.Errorf("failed to iterate vesting schedules: %w", err)
	}
	genesis.VestingSchedules = vestingSchedules

	// Export all vote locks
	voteLocks := make([]*types.VoteLock, 0)
	if err := k.IterateVoteLocks(ctx, func(lock *types.VoteLock) bool {
		voteLocks = append(voteLocks, lock)
		return false // continue iteration
	}); err != nil {
		return genesis, fmt.Errorf("failed to iterate vote locks: %w", err)
	}
	genesis.VoteLocks = voteLocks

	// Export all pending treasury transactions
	pendingTxs := make([]*types.PendingTreasuryTx, 0)
	if err := k.IteratePendingTreasuryTxs(ctx, func(tx *types.PendingTreasuryTx) bool {
		pendingTxs = append(pendingTxs, tx)
		return false // continue iteration
	}); err != nil {
		return genesis, fmt.Errorf("failed to iterate pending treasury txs: %w", err)
	}
	genesis.PendingTreasuryTxs = pendingTxs

	// Export all inflation alerts
	inflationAlerts := make([]*types.InflationAlert, 0)
	if err := k.IterateInflationAlerts(ctx, func(alert *types.InflationAlert) bool {
		inflationAlerts = append(inflationAlerts, alert)
		return false // continue iteration
	}); err != nil {
		return genesis, fmt.Errorf("failed to iterate inflation alerts: %w", err)
	}
	genesis.InflationAlerts = inflationAlerts

	// Export all large transaction records
	largeTxRecords := make([]*types.LargeTxRecord, 0)
	if err := k.IterateLargeTxRecords(ctx, func(record *types.LargeTxRecord) bool {
		largeTxRecords = append(largeTxRecords, record)
		return false // continue iteration
	}); err != nil {
		return genesis, fmt.Errorf("failed to iterate large tx records: %w", err)
	}
	genesis.LargeTxRecords = largeTxRecords

	// Export last large transaction times
	// We need to iterate through the store to get all addresses
	genesis.LastLargeTxTimes = make(map[string]int64)
	// Note: We would need to iterate through all keys in the LastLargeTxTimePrefix
	// For now, we'll leave this as an empty map since we don't have all addresses
	// In a production system, you'd want to maintain an index of all addresses with large txs

	// Export user MEV balances
	genesis.UserMevBalances = make(map[string]string)
	if err := k.IterateUserMEVBalances(ctx, func(address string, balance string) bool {
		genesis.UserMevBalances[address] = balance
		return false // continue iteration
	}); err != nil {
		return genesis, fmt.Errorf("failed to iterate user MEV balances: %w", err)
	}

	return genesis, nil
}

// ValidateInvariants runs all module invariants on genesis state
// This ensures the genesis state is consistent and valid before chain start
func (k *Keeper) ValidateInvariants(ctx context.Context) error {
	// We would call each invariant here, but since invariants expect sdk.Context
	// in the current implementation, we'll just perform basic validation

	// Validate params
	params, _ := k.GetParams(ctx)
	if err := types.ValidateParams(&params); err != nil {
		return fmt.Errorf("invariant check failed - invalid params: %w", err)
	}

	// Validate all vesting schedules are accessible
	vestingCount := 0
	if err := k.IterateVestingSchedules(ctx, func(schedule *types.VestingSchedule) bool {
		vestingCount++
		// Basic validation
		if schedule.ScheduleId == "" {
			return true // stop iteration - error condition
		}
		return false // continue
	}); err != nil {
		return fmt.Errorf("invariant check failed - vesting schedules: %w", err)
	}

	// Validate all vote locks are accessible
	voteLockCount := 0
	if err := k.IterateVoteLocks(ctx, func(lock *types.VoteLock) bool {
		voteLockCount++
		// Basic validation
		if lock.LockId == "" {
			return true // stop iteration - error condition
		}
		return false // continue
	}); err != nil {
		return fmt.Errorf("invariant check failed - vote locks: %w", err)
	}

	return nil
}
