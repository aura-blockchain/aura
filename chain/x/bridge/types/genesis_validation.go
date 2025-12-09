package types

import "fmt"

// ValidateGenesis performs sanity checks on bridge genesis state.
func ValidateGenesis(gen *GenesisState) error {
	if gen == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}
	// Params is a value type (not pointer) in protobuf, cannot be nil
	// Validate the actual parameter values instead
	if gen.Params.MinConfirmations == 0 {
		return fmt.Errorf("min confirmations must be greater than zero")
	}
	if gen.Params.BridgeFeeBasisPoints > 10_000 {
		return fmt.Errorf("bridge fee basis points must be 10000 or less")
	}
	if gen.Params.ValidatorThresholdPercentage == 0 || gen.Params.ValidatorThresholdPercentage > 100 {
		return fmt.Errorf("validator threshold must be between 1 and 100")
	}
	if gen.Params.MaxTransferAmount.IsNil() {
		return fmt.Errorf("max transfer amount cannot be nil")
	}
	if gen.Params.MaxTransferAmount.IsNegative() {
		return fmt.Errorf("max transfer amount cannot be negative")
	}
	return nil
}
