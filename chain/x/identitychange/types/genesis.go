package types

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

type GenesisState struct {
	Params   Params                  `json:"params"`
	Records  []IdentityRecord        `json:"records"`
	Requests []IdentityChangeRequest `json:"requests"`
	History  []IdentityChangeHistory `json:"history"`
	Suspended bool                   `json:"suspended"`
}

func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:   DefaultParams(),
		Records:  []IdentityRecord{},
		Requests: []IdentityChangeRequest{},
		History:  []IdentityChangeHistory{},
		Suspended: false,
	}
}

func ValidateGenesisState(state GenesisState) error {
	if err := state.Params.Validate(); err != nil {
		return fmt.Errorf("params validation failed: %w", err)
	}
	requestIDs := make(map[string]struct{}, len(state.Requests))
	for _, req := range state.Requests {
		if req.RequestID == "" {
			return fmt.Errorf("identity change request missing id")
		}
		if req.TargetDID == "" {
			return fmt.Errorf("identity change request %s target did required", req.RequestID)
		}
		if _, exists := requestIDs[req.RequestID]; exists {
			return fmt.Errorf("duplicate identity change request %s", req.RequestID)
		}
		requestIDs[req.RequestID] = struct{}{}
	}
	for _, record := range state.Records {
		if record.DID == "" {
			return fmt.Errorf("identity record missing did")
		}
	}
	return nil
}

func DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(DefaultGenesisState())
}
