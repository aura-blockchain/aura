package types

import (
	"encoding/json"
	"fmt"

	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
	"github.com/cosmos/cosmos-sdk/codec"
)

func DefaultGenesisState() *identitychangepb.GenesisState {
	return &identitychangepb.GenesisState{
		Params:    DefaultParamsProto(),
		Records:   []*identitychangepb.IdentityRecord{},
		Requests:  []*identitychangepb.IdentityChangeRequest{},
		History:   []*identitychangepb.IdentityChangeHistory{},
		Suspended: false,
	}
}

func ValidateGenesisState(state *identitychangepb.GenesisState) error {
	if state == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}
	if state.Params != nil {
		params := ParamsFromProto(state.Params)
		if err := params.Validate(); err != nil {
			return fmt.Errorf("params validation failed: %w", err)
		}
	}
	requestIDs := make(map[string]struct{}, len(state.Requests))
	for _, req := range state.Requests {
		if req.RequestId == "" {
			return fmt.Errorf("identity change request missing id")
		}
		if req.TargetDid == "" {
			return fmt.Errorf("identity change request %s target did required", req.RequestId)
		}
		if _, exists := requestIDs[req.RequestId]; exists {
			return fmt.Errorf("duplicate identity change request %s", req.RequestId)
		}
		requestIDs[req.RequestId] = struct{}{}
	}
	for _, record := range state.Records {
		if record.Did == "" {
			return fmt.Errorf("identity record missing did")
		}
	}
	return nil
}

func DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(DefaultGenesisState())
}
