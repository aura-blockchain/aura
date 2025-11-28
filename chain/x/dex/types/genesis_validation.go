package types

import "fmt"

// ValidateGenesis performs basic validation for the DEX genesis state.
func ValidateGenesis(gen *GenesisState) error {
    if gen == nil {
        return fmt.Errorf("genesis state cannot be nil")
    }
    if gen.Params == nil {
        return fmt.Errorf("params cannot be nil")
    }
    if gen.Params.TradingFee == "" {
        return fmt.Errorf("trading fee cannot be empty")
    }
    if gen.Params.ProtocolFee == "" {
        return fmt.Errorf("protocol fee cannot be empty")
    }
    if gen.Params.MaxSlippageBps == 0 {
        return fmt.Errorf("max slippage must be greater than zero")
    }
    if gen.Params.MinSwapAmount == "" {
        return fmt.Errorf("min swap amount must be set")
    }
    return nil
}
