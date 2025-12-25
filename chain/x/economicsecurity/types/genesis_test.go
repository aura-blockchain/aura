// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
)

func TestDefaultGenesis(t *testing.T) {
	genesis := DefaultGenesis()

	if genesis == nil {
		t.Fatal("DefaultGenesis should not return nil")
	}

	if genesis.Params == nil {
		t.Fatal("expected Params to be set")
	}

	if genesis.VestingSchedules == nil {
		t.Error("expected VestingSchedules to be initialized")
	}

	if len(genesis.VestingSchedules) != 0 {
		t.Errorf("expected empty VestingSchedules, got length %d", len(genesis.VestingSchedules))
	}

	if genesis.VoteLocks == nil {
		t.Error("expected VoteLocks to be initialized")
	}

	if genesis.PendingTreasuryTxs == nil {
		t.Error("expected PendingTreasuryTxs to be initialized")
	}

	if genesis.InflationAlerts == nil {
		t.Error("expected InflationAlerts to be initialized")
	}

	if genesis.LargeTxRecords == nil {
		t.Error("expected LargeTxRecords to be initialized")
	}

	if genesis.LastLargeTxTimes == nil {
		t.Error("expected LastLargeTxTimes to be initialized")
	}

	if genesis.UserMevBalances == nil {
		t.Error("expected UserMevBalances to be initialized")
	}
}

func TestValidateGenesis_Valid(t *testing.T) {
	genesis := DefaultGenesis()

	err := ValidateGenesis(genesis)
	if err != nil {
		t.Errorf("ValidateGenesis should not return error for default genesis: %v", err)
	}
}

func TestValidateGenesis_WithSchedules(t *testing.T) {
	genesis := DefaultGenesis()
	genesis.VestingSchedules = []*VestingSchedule{
		{
			ScheduleId:         "schedule1",
			BeneficiaryAddress: "aura1testaddr",
			TotalAmount:        "1000",
		},
	}

	err := ValidateGenesis(genesis)
	if err != nil {
		t.Errorf("ValidateGenesis should not return error for valid schedules: %v", err)
	}
}

func TestValidateGenesis_DuplicateScheduleID(t *testing.T) {
	genesis := DefaultGenesis()
	genesis.VestingSchedules = []*VestingSchedule{
		{
			ScheduleId:         "schedule1",
			BeneficiaryAddress: "aura1testaddr",
			TotalAmount:        "1000",
		},
		{
			ScheduleId:         "schedule1",
			BeneficiaryAddress: "aura1testaddr2",
			TotalAmount:        "2000",
		},
	}

	err := ValidateGenesis(genesis)
	if err == nil {
		t.Error("ValidateGenesis should return error for duplicate schedule ID")
	}
}

func TestValidateGenesis_EmptyScheduleID(t *testing.T) {
	genesis := DefaultGenesis()
	genesis.VestingSchedules = []*VestingSchedule{
		{
			ScheduleId:         "",
			BeneficiaryAddress: "aura1testaddr",
			TotalAmount:        "1000",
		},
	}

	err := ValidateGenesis(genesis)
	if err != ErrInvalidScheduleID {
		t.Errorf("ValidateGenesis should return ErrInvalidScheduleID, got: %v", err)
	}
}

func TestValidateGenesis_EmptyBeneficiary(t *testing.T) {
	genesis := DefaultGenesis()
	genesis.VestingSchedules = []*VestingSchedule{
		{
			ScheduleId:         "schedule1",
			BeneficiaryAddress: "",
			TotalAmount:        "1000",
		},
	}

	err := ValidateGenesis(genesis)
	if err != ErrInvalidBeneficiary {
		t.Errorf("ValidateGenesis should return ErrInvalidBeneficiary, got: %v", err)
	}
}

func TestValidateGenesis_InvalidAmount(t *testing.T) {
	genesis := DefaultGenesis()
	genesis.VestingSchedules = []*VestingSchedule{
		{
			ScheduleId:         "schedule1",
			BeneficiaryAddress: "aura1testaddr",
			TotalAmount:        "0",
		},
	}

	err := ValidateGenesis(genesis)
	if err != ErrInvalidAmount {
		t.Errorf("ValidateGenesis should return ErrInvalidAmount, got: %v", err)
	}
}

func TestValidateGenesis_WithVoteLocks(t *testing.T) {
	genesis := DefaultGenesis()
	genesis.VoteLocks = []*VoteLock{
		{
			LockId: "lock1",
			Owner:  "aura1testaddr",
			Amount: "1000",
		},
	}

	err := ValidateGenesis(genesis)
	if err != nil {
		t.Errorf("ValidateGenesis should not return error for valid locks: %v", err)
	}
}

func TestValidateGenesis_DuplicateLockID(t *testing.T) {
	genesis := DefaultGenesis()
	genesis.VoteLocks = []*VoteLock{
		{
			LockId: "lock1",
			Owner:  "aura1testaddr",
			Amount: "1000",
		},
		{
			LockId: "lock1",
			Owner:  "aura1testaddr2",
			Amount: "2000",
		},
	}

	err := ValidateGenesis(genesis)
	if err == nil {
		t.Error("ValidateGenesis should return error for duplicate lock ID")
	}
}

func TestValidateGenesis_EmptyLockOwner(t *testing.T) {
	genesis := DefaultGenesis()
	genesis.VoteLocks = []*VoteLock{
		{
			LockId: "lock1",
			Owner:  "",
			Amount: "1000",
		},
	}

	err := ValidateGenesis(genesis)
	if err != ErrInvalidAddress {
		t.Errorf("ValidateGenesis should return ErrInvalidAddress, got: %v", err)
	}
}

func TestValidateGenesis_WithPendingTx(t *testing.T) {
	genesis := DefaultGenesis()
	genesis.PendingTreasuryTxs = []*PendingTreasuryTx{
		{
			TxId:      "tx1",
			Recipient: "aura1testaddr",
			Amount:    "1000",
		},
	}

	err := ValidateGenesis(genesis)
	if err != nil {
		t.Errorf("ValidateGenesis should not return error for valid pending tx: %v", err)
	}
}

func TestValidateGenesis_DuplicateTxID(t *testing.T) {
	genesis := DefaultGenesis()
	genesis.PendingTreasuryTxs = []*PendingTreasuryTx{
		{
			TxId:      "tx1",
			Recipient: "aura1testaddr",
			Amount:    "1000",
		},
		{
			TxId:      "tx1",
			Recipient: "aura1testaddr2",
			Amount:    "2000",
		},
	}

	err := ValidateGenesis(genesis)
	if err == nil {
		t.Error("ValidateGenesis should return error for duplicate tx ID")
	}
}

func TestValidateGenesis_EmptyTxRecipient(t *testing.T) {
	genesis := DefaultGenesis()
	genesis.PendingTreasuryTxs = []*PendingTreasuryTx{
		{
			TxId:      "tx1",
			Recipient: "",
			Amount:    "1000",
		},
	}

	err := ValidateGenesis(genesis)
	if err != ErrInvalidAddress {
		t.Errorf("ValidateGenesis should return ErrInvalidAddress, got: %v", err)
	}
}
