package types

import (
	"fmt"

	v1beta1 "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

// ValidateGenesis performs basic validation for the DEX genesis state.
func ValidateGenesis(gen *v1beta1.GenesisState) error {
	if gen == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}
	// Params is a value type (not pointer) in protobuf, cannot be nil
	// Validate the actual parameter values instead
	if gen.Params.TradingFee.IsNil() {
		return fmt.Errorf("trading fee cannot be nil")
	}
	if gen.Params.TradingFee.IsNegative() {
		return fmt.Errorf("trading fee cannot be negative")
	}
	if gen.Params.ProtocolFee.IsNil() {
		return fmt.Errorf("protocol fee cannot be nil")
	}
	if gen.Params.ProtocolFee.IsNegative() {
		return fmt.Errorf("protocol fee cannot be negative")
	}
	if gen.Params.MaxSlippageBps == 0 {
		return fmt.Errorf("max slippage must be greater than zero")
	}
	if gen.Params.MinSwapAmount.IsNil() {
		return fmt.Errorf("min swap amount cannot be nil")
	}
	if gen.Params.MinSwapAmount.IsNegative() {
		return fmt.Errorf("min swap amount cannot be negative")
	}
	return nil
}
