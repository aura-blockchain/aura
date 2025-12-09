package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidateGenesis performs comprehensive validation on bridge genesis state
// to prevent data corruption and invalid state on chain initialization.
func ValidateGenesis(gen *GenesisState) error {
	if gen == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}

	// Validate parameters
	if err := validateBridgeParams(&gen.Params); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate transfers
	transferIDs := make(map[string]bool)
	for i, transfer := range gen.Transfers {
		if err := validateTransfer(&transfer, i); err != nil {
			return fmt.Errorf("invalid transfer at index %d: %w", i, err)
		}
		// Check for duplicate transfer IDs
		if transferIDs[transfer.TransferId] {
			return fmt.Errorf("duplicate transfer ID %s at index %d", transfer.TransferId, i)
		}
		transferIDs[transfer.TransferId] = true
	}

	// Validate chain configurations
	chainIDs := make(map[string]bool)
	for i, config := range gen.ChainConfigs {
		if err := validateChainConfig(&config, i); err != nil {
			return fmt.Errorf("invalid chain config at index %d: %w", i, err)
		}
		// Check for duplicate chain IDs
		if chainIDs[config.ChainId] {
			return fmt.Errorf("duplicate chain ID %s at index %d", config.ChainId, i)
		}
		chainIDs[config.ChainId] = true
	}

	// Validate validators
	if err := validateValidators(gen.Validators, gen.Params.ValidatorThresholdPercentage); err != nil {
		return fmt.Errorf("invalid validators: %w", err)
	}

	// Validate wrapped tokens
	wrappedTokenKeys := make(map[string]bool)
	for i, token := range gen.WrappedTokens {
		if err := validateWrappedToken(&token, i); err != nil {
			return fmt.Errorf("invalid wrapped token at index %d: %w", i, err)
		}
		// Check for duplicate token keys (denom + chain)
		key := fmt.Sprintf("%s:%s", token.WrappedDenom, token.SourceChain)
		if wrappedTokenKeys[key] {
			return fmt.Errorf("duplicate wrapped token %s from chain %s at index %d",
				token.WrappedDenom, token.SourceChain, i)
		}
		wrappedTokenKeys[key] = true
	}

	// Validate shared identities
	identityAddrs := make(map[string]bool)
	for i, identity := range gen.SharedIdentities {
		if err := validateSharedIdentity(&identity, i); err != nil {
			return fmt.Errorf("invalid shared identity at index %d: %w", i, err)
		}
		// Check for duplicate identity addresses
		if identityAddrs[identity.Address] {
			return fmt.Errorf("duplicate shared identity address %s at index %d",
				identity.Address, i)
		}
		identityAddrs[identity.Address] = true
	}

	// Validate cross-chain swaps
	swapIDs := make(map[string]bool)
	for i, swap := range gen.CrossChainSwaps {
		if err := validateCrossChainSwap(&swap, i); err != nil {
			return fmt.Errorf("invalid cross-chain swap at index %d: %w", i, err)
		}
		// Check for duplicate swap IDs
		if swap.SwapId != "" {
			if swapIDs[swap.SwapId] {
				return fmt.Errorf("duplicate swap ID %s at index %d", swap.SwapId, i)
			}
			swapIDs[swap.SwapId] = true
		}
	}

	// Validate relayer stats
	relayerAddrs := make(map[string]bool)
	for i, stats := range gen.RelayerStats {
		if err := validateRelayerStats(&stats, i); err != nil {
			return fmt.Errorf("invalid relayer stats at index %d: %w", i, err)
		}
		// Check for duplicate relayer addresses
		if relayerAddrs[stats.RelayerAddress] {
			return fmt.Errorf("duplicate relayer stats for address %s at index %d",
				stats.RelayerAddress, i)
		}
		relayerAddrs[stats.RelayerAddress] = true
	}

	// Validate processed source hashes
	processedHashes := make(map[string]bool)
	for i, hash := range gen.ProcessedSourceHashes {
		if hash == "" {
			return fmt.Errorf("processed source hash cannot be empty at index %d", i)
		}
		if processedHashes[hash] {
			return fmt.Errorf("duplicate processed source hash at index %d", i)
		}
		processedHashes[hash] = true
	}

	return nil
}

// validateBridgeParams validates bridge module parameters
func validateBridgeParams(params *BridgeParams) error {
	if params.MinConfirmations == 0 {
		return fmt.Errorf("min confirmations must be greater than zero")
	}
	if params.MinConfirmations > 1000 {
		return fmt.Errorf("min confirmations cannot exceed 1000 (unreasonable)")
	}

	if params.BridgeFeeBasisPoints > 10_000 {
		return fmt.Errorf("bridge fee basis points must be 10000 or less (100%%)")
	}

	if params.ValidatorThresholdPercentage == 0 {
		return fmt.Errorf("validator threshold percentage must be greater than zero")
	}
	if params.ValidatorThresholdPercentage > 100 {
		return fmt.Errorf("validator threshold percentage cannot exceed 100")
	}

	if params.MaxTransferAmount.IsNil() {
		return fmt.Errorf("max transfer amount cannot be nil")
	}
	if params.MaxTransferAmount.IsNegative() {
		return fmt.Errorf("max transfer amount cannot be negative")
	}
	if params.MaxTransferAmount.IsZero() {
		return fmt.Errorf("max transfer amount must be positive")
	}

	return nil
}

// validateTransfer validates a cross-chain transfer
func validateTransfer(transfer *CrossChainTransfer, index int) error {
	if transfer.TransferId == "" {
		return fmt.Errorf("transfer ID cannot be empty")
	}
	if transfer.SourceChain == "" {
		return fmt.Errorf("source chain cannot be empty")
	}
	if transfer.TargetChain == "" {
		return fmt.Errorf("target chain cannot be empty")
	}
	if transfer.SourceChain == transfer.TargetChain {
		return fmt.Errorf("source chain and target chain must be different")
	}
	if transfer.Sender == "" {
		return fmt.Errorf("sender address cannot be empty")
	}
	if transfer.Recipient == "" {
		return fmt.Errorf("recipient address cannot be empty")
	}
	if transfer.Amount.IsNil() || transfer.Amount.LTE(sdkmath.ZeroInt()) {
		return fmt.Errorf("transfer amount must be positive")
	}
	if transfer.Denom == "" {
		return fmt.Errorf("denom cannot be empty")
	}

	// Validate confirmations do not exceed required confirmations
	if transfer.Confirmations > transfer.RequiredConfirmations {
		return fmt.Errorf("confirmations (%d) cannot exceed required confirmations (%d)",
			transfer.Confirmations, transfer.RequiredConfirmations)
	}

	// Validate validator signatures
	signerAddrs := make(map[string]bool)
	for j, sig := range transfer.ValidatorSignatures {
		if sig.ValidatorAddress == "" {
			return fmt.Errorf("validator address cannot be empty at signature index %d", j)
		}
		if signerAddrs[sig.ValidatorAddress] {
			return fmt.Errorf("duplicate validator signature from %s at index %d",
				sig.ValidatorAddress, j)
		}
		signerAddrs[sig.ValidatorAddress] = true
		if len(sig.Signature) == 0 {
			return fmt.Errorf("signature cannot be empty at index %d", j)
		}
	}

	return nil
}

// validateChainConfig validates a chain configuration
func validateChainConfig(config *ChainConfig, index int) error {
	if config.ChainId == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	if config.ChainName == "" {
		return fmt.Errorf("chain name cannot be empty")
	}
	if config.AddressPrefix == "" {
		return fmt.Errorf("address prefix cannot be empty")
	}
	if config.MinConfirmations == 0 {
		return fmt.Errorf("min confirmations must be greater than zero")
	}

	return nil
}

// validateValidators validates the validator set
func validateValidators(validators []BridgeValidator, thresholdPercentage uint64) error {
	if len(validators) == 0 {
		return nil // Empty validator set is allowed (will be populated later)
	}

	validatorAddrs := make(map[string]bool)
	totalPower := uint64(0)

	for i, validator := range validators {
		// Validate validator address is valid bech32
		if validator.Address == "" {
			return fmt.Errorf("validator address cannot be empty at index %d", i)
		}

		// Validate using SDK address validation
		if _, err := sdk.AccAddressFromBech32(validator.Address); err != nil {
			return fmt.Errorf("invalid validator address at index %d: %w", i, err)
		}

		// Check for duplicate validators
		if validatorAddrs[validator.Address] {
			return fmt.Errorf("duplicate validator address %s at index %d",
				validator.Address, i)
		}
		validatorAddrs[validator.Address] = true

		// Validate voting power
		if validator.Power == 0 {
			return fmt.Errorf("validator voting power must be greater than zero at index %d", i)
		}
		totalPower += validator.Power
	}

	// Validate threshold is achievable with current validator set
	// Threshold must be <= total voting power
	if thresholdPercentage > 0 && totalPower > 0 {
		requiredPower := (totalPower * thresholdPercentage) / 100
		if requiredPower > totalPower {
			return fmt.Errorf("threshold percentage %d%% requires more power than available (required: %d, total: %d)",
				thresholdPercentage, requiredPower, totalPower)
		}
	}

	return nil
}

// validateWrappedToken validates a wrapped token entry
func validateWrappedToken(token *WrappedToken, index int) error {
	if token.WrappedDenom == "" {
		return fmt.Errorf("wrapped denom cannot be empty")
	}
	if token.SourceChain == "" {
		return fmt.Errorf("source chain cannot be empty")
	}
	if token.OriginalDenom == "" {
		return fmt.Errorf("original denom cannot be empty")
	}

	// Validate total supply is not negative
	if token.TotalSupply.IsNil() || token.TotalSupply.IsNegative() {
		return fmt.Errorf("total supply cannot be nil or negative")
	}

	return nil
}

// validateSharedIdentity validates a shared identity
func validateSharedIdentity(identity *SharedIdentity, index int) error {
	if identity.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	// Validate reputation score is within bounds (0-1000)
	if identity.ReputationScore > 1000 {
		return fmt.Errorf("reputation score cannot exceed 1000")
	}

	// Validate IR score is within bounds (0-100)
	if identity.AuraIrScore > 100 {
		return fmt.Errorf("AURA IR score cannot exceed 100")
	}

	return nil
}

// validateCrossChainSwap validates a cross-chain swap
func validateCrossChainSwap(swap *CrossChainSwap, index int) error {
	if swap.SwapId == "" {
		return fmt.Errorf("swap ID cannot be empty")
	}
	if swap.Sender == "" {
		return fmt.Errorf("sender address cannot be empty")
	}
	if swap.SourceChain == "" {
		return fmt.Errorf("source chain cannot be empty")
	}
	if swap.TargetChain == "" {
		return fmt.Errorf("target chain cannot be empty")
	}
	if swap.SourceChain == swap.TargetChain {
		return fmt.Errorf("source chain and target chain must be different")
	}
	// Note: source_coin and final_amount are composite/custom types, validation is complex
	// We validate that basic fields are present
	if swap.TargetDenom == "" {
		return fmt.Errorf("target denom cannot be empty")
	}
	if swap.MinTargetAmount.IsNil() || swap.MinTargetAmount.IsNegative() {
		return fmt.Errorf("min target amount cannot be nil or negative")
	}

	return nil
}

// validateRelayerStats validates relayer statistics
func validateRelayerStats(stats *RelayerStats, index int) error {
	if stats.RelayerAddress == "" {
		return fmt.Errorf("relayer address cannot be empty")
	}

	// Validate using SDK address validation
	if _, err := sdk.AccAddressFromBech32(stats.RelayerAddress); err != nil {
		return fmt.Errorf("invalid relayer address: %w", err)
	}

	// Validate logical consistency - successful + failed should equal total
	if stats.SuccessfulTransfers+stats.FailedTransfers != stats.TotalTransfersRelayed {
		return fmt.Errorf("sum of successful (%d) and failed (%d) transfers must equal total (%d)",
			stats.SuccessfulTransfers, stats.FailedTransfers, stats.TotalTransfersRelayed)
	}

	// Validate total volume is not negative
	if stats.TotalVolume.IsNil() || stats.TotalVolume.IsNegative() {
		return fmt.Errorf("total volume cannot be nil or negative")
	}

	return nil
}
