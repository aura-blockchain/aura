package types

import (
    "fmt"
    "time"

    paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Default values for security parameters
var (
    DefaultTimelockDuration  = 24 * time.Hour     // 24 hour timelock for withdrawals
    DefaultFraudProofWindow  = 7 * 24 * time.Hour // 7 day window for fraud proofs
)

// Parameter store keys
var (
    KeyBridgeEnabled                = []byte("BridgeEnabled")
    KeyMinConfirmations             = []byte("MinConfirmations")
    KeyBridgeFeeBasisPoints         = []byte("BridgeFeeBasisPoints")
    KeyCoreMaxTransferAmount            = []byte("MaxTransferAmount")
    KeyValidatorThresholdPercentage = []byte("ValidatorThresholdPercentage")
)

// Params defines the parameters persisted in the Cosmos SDK param store.
type Params struct {
    BridgeEnabled                bool   `json:"bridge_enabled"`
    MinConfirmations             uint64 `json:"min_confirmations"`
    BridgeFeeBasisPoints         uint64 `json:"bridge_fee_basis_points"`
    MaxTransferAmount            string `json:"max_transfer_amount"`
    ValidatorThresholdPercentage uint64 `json:"validator_threshold_percentage"`
}

// DefaultParams returns default parameters used by the param store.
func DefaultParams() Params {
    return Params{
        BridgeEnabled:                true,
        MinConfirmations:             1,
        BridgeFeeBasisPoints:         30,
        MaxTransferAmount:            "1000000000",
        ValidatorThresholdPercentage: 67,
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
        paramtypes.NewParamSetPair(KeyMinConfirmations, &p.MinConfirmations, validateUint64Core),
        paramtypes.NewParamSetPair(KeyBridgeFeeBasisPoints, &p.BridgeFeeBasisPoints, validateUint64Core),
        paramtypes.NewParamSetPair(KeyCoreMaxTransferAmount, &p.MaxTransferAmount, validateStringNotEmpty),
        paramtypes.NewParamSetPair(KeyValidatorThresholdPercentage, &p.ValidatorThresholdPercentage, validateUint64Core),
    }
}

func validateBool(i interface{}) error {
    _, ok := i.(bool)
    if !ok {
        return ErrInvalidParam
    }
    return nil
}

func validateUint64Core(i interface{}) error {
    _, ok := i.(uint64)
    if !ok {
        return ErrInvalidParam
    }
    return nil
}

func validateStringNotEmpty(i interface{}) error {
    s, ok := i.(string)
    if !ok {
        return ErrInvalidParam
    }
    if s == "" {
        return fmt.Errorf("value cannot be empty")
    }
    return nil
}

// DefaultGenesis returns the default genesis state for the bridge module
func DefaultGenesis() *GenesisState {
    params := DefaultParams()
    return &GenesisState{
        Params: &BridgeParams{
            Enabled:                      params.BridgeEnabled,
            MinConfirmations:             params.MinConfirmations,
            BridgeFeeBasisPoints:         params.BridgeFeeBasisPoints,
            MaxTransferAmount:            params.MaxTransferAmount,
            ValidatorThresholdPercentage: params.ValidatorThresholdPercentage,
        },
    }
}


