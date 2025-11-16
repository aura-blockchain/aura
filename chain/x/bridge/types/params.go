package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyBridgeEnabled = []byte("BridgeEnabled")
)

// Params defines the parameters for the Bridge module
type Params struct {
	BridgeEnabled bool `json:"bridge_enabled"`
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		BridgeEnabled: true,
	}
}

// ParamKeyTable returns the parameter key table
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyBridgeEnabled, &p.BridgeEnabled, validateBool),
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return ErrInvalidParam
	}
	return nil
}
