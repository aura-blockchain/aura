package types

import (
	"errors"

	economicsecuritypb "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:             DefaultParams(),
		VestingSchedules:   []*VestingSchedule{},
		VoteLocks:          []*VoteLock{},
		PendingTreasuryTxs: []*PendingTreasuryTx{},
		InflationAlerts:    []*InflationAlert{},
		LargeTxRecords:     []*LargeTxRecord{},
		LastLargeTxTimes:   make(map[string]int64),
		UserMevBalances:    make(map[string]string),
	}
}

// ValidateGenesis performs basic validation of genesis data
func ValidateGenesis(gs *GenesisState) error {
	// Validate params
	if err := ValidateParams(gs.Params); err != nil {
		return err
	}

	// Validate vesting schedules
	scheduleIDs := make(map[string]bool)
	for _, schedule := range gs.VestingSchedules {
		if schedule.ScheduleId == "" {
			return ErrInvalidScheduleID
		}
		if scheduleIDs[schedule.ScheduleId] {
			return errors.New("duplicate schedule ID: " + schedule.ScheduleId)
		}
		scheduleIDs[schedule.ScheduleId] = true

		if schedule.BeneficiaryAddress == "" {
			return ErrInvalidBeneficiary
		}
		if schedule.TotalAmount == "" || schedule.TotalAmount == "0" {
			return ErrInvalidAmount
		}
	}

	// Validate vote locks
	lockIDs := make(map[string]bool)
	for _, lock := range gs.VoteLocks {
		if lock.LockId == "" {
			return errors.New("invalid lock ID")
		}
		if lockIDs[lock.LockId] {
			return errors.New("duplicate lock ID: " + lock.LockId)
		}
		lockIDs[lock.LockId] = true

		if lock.Owner == "" {
			return ErrInvalidAddress
		}
		if lock.Amount == "" || lock.Amount == "0" {
			return ErrInvalidAmount
		}
	}

	// Validate pending treasury transactions
	txIDs := make(map[string]bool)
	for _, tx := range gs.PendingTreasuryTxs {
		if tx.TxId == "" {
			return errors.New("invalid transaction ID")
		}
		if txIDs[tx.TxId] {
			return errors.New("duplicate transaction ID: " + tx.TxId)
		}
		txIDs[tx.TxId] = true

		if tx.Recipient == "" {
			return ErrInvalidAddress
		}
		if tx.Amount == "" || tx.Amount == "0" {
			return ErrInvalidAmount
		}
	}

	return nil
}

// GenesisStateFromProto converts proto GenesisState to internal type
func GenesisStateFromProto(pb *economicsecuritypb.GenesisState) *GenesisState {
	if pb == nil {
		return DefaultGenesis()
	}
	return pb
}

// GenesisStateToProto converts internal GenesisState to proto type
func GenesisStateToProto(gs *GenesisState) *economicsecuritypb.GenesisState {
	return gs
}
