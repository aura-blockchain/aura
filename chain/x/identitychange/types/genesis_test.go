package types

import (
	"testing"
)

func TestDefaultGenesisState(t *testing.T) {
	genesis := DefaultGenesisState()

	if genesis == nil {
		t.Fatal("DefaultGenesisState should not return nil")
	}

	if genesis.Params == nil {
		t.Fatal("expected Params to be set")
	}

	if genesis.Records == nil {
		t.Error("expected Records to be initialized")
	}

	if len(genesis.Records) != 0 {
		t.Errorf("expected empty Records, got length %d", len(genesis.Records))
	}

	if genesis.Requests == nil {
		t.Error("expected Requests to be initialized")
	}

	if len(genesis.Requests) != 0 {
		t.Errorf("expected empty Requests, got length %d", len(genesis.Requests))
	}

	if genesis.History == nil {
		t.Error("expected History to be initialized")
	}

	if genesis.Suspended {
		t.Error("expected Suspended to be false")
	}
}

func TestValidateGenesisState_Valid(t *testing.T) {
	genesis := DefaultGenesisState()

	err := ValidateGenesisState(genesis)
	if err != nil {
		t.Errorf("ValidateGenesisState should not return error for default genesis: %v", err)
	}
}

func TestValidateGenesisState_Nil(t *testing.T) {
	err := ValidateGenesisState(nil)
	if err == nil {
		t.Error("ValidateGenesisState should return error for nil state")
	}
}

func TestValidateGenesisState_MissingRequestID(t *testing.T) {
	genesis := DefaultGenesisState()
	genesis.Requests = append(genesis.Requests, &IdentityChangeRequest{
		RequestId: "",
		TargetDid: "did:aura:123",
	})

	err := ValidateGenesisState(genesis)
	if err == nil {
		t.Error("ValidateGenesisState should return error for missing request ID")
	}
}

func TestValidateGenesisState_MissingTargetDID(t *testing.T) {
	genesis := DefaultGenesisState()
	genesis.Requests = append(genesis.Requests, &IdentityChangeRequest{
		RequestId: "req1",
		TargetDid: "",
	})

	err := ValidateGenesisState(genesis)
	if err == nil {
		t.Error("ValidateGenesisState should return error for missing target DID")
	}
}

func TestValidateGenesisState_DuplicateRequestID(t *testing.T) {
	genesis := DefaultGenesisState()
	genesis.Requests = append(genesis.Requests,
		&IdentityChangeRequest{
			RequestId: "req1",
			TargetDid: "did:aura:123",
		},
		&IdentityChangeRequest{
			RequestId: "req1",
			TargetDid: "did:aura:456",
		},
	)

	err := ValidateGenesisState(genesis)
	if err == nil {
		t.Error("ValidateGenesisState should return error for duplicate request ID")
	}
}

func TestValidateGenesisState_MissingRecordDID(t *testing.T) {
	genesis := DefaultGenesisState()
	genesis.Records = append(genesis.Records, &IdentityRecord{
		Did: "",
	})

	err := ValidateGenesisState(genesis)
	if err == nil {
		t.Error("ValidateGenesisState should return error for missing record DID")
	}
}

func TestValidateGenesisState_ValidRecords(t *testing.T) {
	genesis := DefaultGenesisState()
	genesis.Records = append(genesis.Records, &IdentityRecord{
		Did: "did:aura:123",
	})

	err := ValidateGenesisState(genesis)
	if err != nil {
		t.Errorf("ValidateGenesisState should not return error for valid records: %v", err)
	}
}
