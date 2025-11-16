package keeper

import (
	"fmt"
	"math/big"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// InitGenesis initializes the keeper from genesis state
func (k *Keeper) InitGenesis(genesis types.GenesisState) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Set params
	if genesis.Params != nil {
		if err := k.paramsStore.SetParams(*genesis.Params); err != nil {
			return fmt.Errorf("failed to set params: %w", err)
		}
	}

	// Load vesting schedules
	for _, schedule := range genesis.VestingSchedules {
		k.vestingSchedules[schedule.ScheduleId] = schedule
		k.userVestings[schedule.BeneficiaryAddress] = append(
			k.userVestings[schedule.BeneficiaryAddress],
			schedule.ScheduleId,
		)
	}

	// Load vote locks
	for _, lock := range genesis.VoteLocks {
		k.voteLocks[lock.LockId] = lock
		k.userVoteLocks[lock.Owner] = append(k.userVoteLocks[lock.Owner], lock.LockId)
	}

	// Load pending treasury transactions
	for _, tx := range genesis.PendingTreasuryTxs {
		k.pendingTreasuryTxs[tx.TxId] = tx
	}

	// Load inflation alerts
	k.inflationAlerts = genesis.InflationAlerts

	// Load large transaction records
	k.largeTxRecords = genesis.LargeTxRecords

	// Load last large tx times
	k.lastLargeTxTimes = genesis.LastLargeTxTimes

	// Load user MEV balances
	for addr, balanceStr := range genesis.UserMevBalances {
		balance := new(big.Int)
		balance.SetString(balanceStr, 10)
		k.userMEVBalances[addr] = balance
	}

	return nil
}

// ExportGenesis exports the current state for genesis
func (k *Keeper) ExportGenesis() types.GenesisState {
	k.mu.RLock()
	defer k.mu.RUnlock()

	vestingSchedules := make([]*types.VestingSchedule, 0, len(k.vestingSchedules))
	for _, schedule := range k.vestingSchedules {
		vestingSchedules = append(vestingSchedules, schedule)
	}

	voteLocks := make([]*types.VoteLock, 0, len(k.voteLocks))
	for _, lock := range k.voteLocks {
		voteLocks = append(voteLocks, lock)
	}

	pendingTxs := make([]*types.PendingTreasuryTx, 0, len(k.pendingTreasuryTxs))
	for _, tx := range k.pendingTreasuryTxs {
		pendingTxs = append(pendingTxs, tx)
	}

	userMEVBalances := make(map[string]string)
	for addr, balance := range k.userMEVBalances {
		userMEVBalances[addr] = balance.String()
	}

	params := k.GetParams()
	return types.GenesisState{
		Params:             &params,
		VestingSchedules:   vestingSchedules,
		VoteLocks:          voteLocks,
		PendingTreasuryTxs: pendingTxs,
		InflationAlerts:    k.inflationAlerts,
		LargeTxRecords:     k.largeTxRecords,
		LastLargeTxTimes:   k.lastLargeTxTimes,
		UserMevBalances:    userMEVBalances,
	}
}
