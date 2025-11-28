package types

import "fmt"

// GenesisState holds the governance module genesis configuration.
type GenesisState struct {
    Params *GovernanceParams `json:"params" yaml:"params"`
}

// DefaultGenesis returns default governance genesis state.
func DefaultGenesis() *GenesisState {
    params := DefaultParams()
    return &GenesisState{Params: params}
}

// Validate validates the governance genesis configuration.
func (g GenesisState) Validate() error {
    if g.Params == nil {
        return fmt.Errorf("governance params cannot be nil")
    }

    if g.Params.MinDeposit == "" {
        return fmt.Errorf("min deposit must be set")
    }
    if g.Params.Quorum == "" || g.Params.Threshold == "" || g.Params.VetoThreshold == "" {
        return fmt.Errorf("voting thresholds must be set")
    }
    if g.Params.MaxDepositPeriod == nil || g.Params.VotingPeriod == nil {
        return fmt.Errorf("deposit and voting periods must be set")
    }
    return nil
}
