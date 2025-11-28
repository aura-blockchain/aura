package v1beta1

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/aequitas/aura/chain/x/common/validation"
)

const (
	// MaxChainIDLength is the maximum length for chain identifiers
	MaxChainIDLength = 64
	// MinChainIDLength is the minimum length for chain identifiers
	MinChainIDLength = 2
	// MaxTransferIDLength is the maximum length for transfer IDs
	MaxTransferIDLength = 128
	// MaxSignatureSize is the maximum size for a signature
	MaxSignatureSize = 256
	// MinSignatureSize is the minimum size for a signature (64 bytes for ed25519)
	MinSignatureSize = 64
	// MaxSignatures is the maximum number of validator signatures
	MaxSignatures = 100
	// MinSignatures is the minimum number of validator signatures for unlock
	MinSignatures = 1
	// MaxSlippageBps is the maximum allowed slippage in basis points (100%)
	MaxSlippageBps = uint64(10000)
	// MaxStatusLength is the maximum length for status strings
	MaxStatusLength = 64
)

// parseAndValidatePositiveInt parses a string to Int and validates it's positive
func parseAndValidatePositiveInt(s string, fieldName string) error {
	if s == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	val, ok := sdkmath.NewIntFromString(s)
	if !ok {
		return fmt.Errorf("%s must be a valid integer, got: %s", fieldName, s)
	}
	return validation.ValidatePositiveInt(val, fieldName)
}

// validateChainID validates a chain identifier (paw, xai, aura, etc.)
func validateChainID(chainID string, fieldName string) error {
	if err := validation.ValidateBoundedString(chainID, MinChainIDLength, MaxChainIDLength, fieldName); err != nil {
		return err
	}
	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgLockTokens
func (m *MsgLockTokens) ValidateBasic() error {
	// Validate sender address
	if err := validation.ValidateAccAddress(m.Sender); err != nil {
		return fmt.Errorf("sender: %w", err)
	}

	// Validate target chain
	if err := validateChainID(m.TargetChain, "target_chain"); err != nil {
		return err
	}

	// Validate recipient (non-empty string)
	if err := validation.ValidateBoundedString(m.Recipient, 1, 256, "recipient"); err != nil {
		return err
	}

	// Validate amount (pointer type due to protoc-gen-go)
	if m.Amount == nil {
		return fmt.Errorf("amount cannot be nil")
	}
	if err := validation.ValidateCoin(*m.Amount, "amount"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgMintTokens
func (m *MsgMintTokens) ValidateBasic() error {
	// Validate validator address
	if err := validation.ValidateAccAddress(m.Validator); err != nil {
		return fmt.Errorf("validator: %w", err)
	}

	// Validate source chain
	if err := validateChainID(m.SourceChain, "source_chain"); err != nil {
		return err
	}

	// Validate source transaction hash
	if err := validation.ValidateHash(m.SourceTxHash); err != nil {
		return fmt.Errorf("source_tx_hash: %w", err)
	}

	// Validate recipient
	if err := validation.ValidateAccAddress(m.Recipient); err != nil {
		return fmt.Errorf("recipient: %w", err)
	}

	// Validate amount (string type - customtype annotation ignored by protoc-gen-go)
	if err := parseAndValidatePositiveInt(m.Amount, "amount"); err != nil {
		return err
	}

	// Validate denom
	if err := validation.ValidateDenom(m.Denom); err != nil {
		return fmt.Errorf("denom: %w", err)
	}

	// Validate validator signature
	if err := validation.ValidateBytes(m.ValidatorSignature, MinSignatureSize, MaxSignatureSize, "validator_signature"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUnlockTokens
func (m *MsgUnlockTokens) ValidateBasic() error {
	// Validate sender address
	if err := validation.ValidateAccAddress(m.Sender); err != nil {
		return fmt.Errorf("sender: %w", err)
	}

	// Validate source chain
	if err := validateChainID(m.SourceChain, "source_chain"); err != nil {
		return err
	}

	// Validate burn transaction hash
	if err := validation.ValidateHash(m.BurnTxHash); err != nil {
		return fmt.Errorf("burn_tx_hash: %w", err)
	}

	// Validate amount (string type - customtype annotation ignored by protoc-gen-go)
	if err := parseAndValidatePositiveInt(m.Amount, "amount"); err != nil {
		return err
	}

	// Validate denom
	if err := validation.ValidateDenom(m.Denom); err != nil {
		return fmt.Errorf("denom: %w", err)
	}

	// Validate validator signatures
	if len(m.ValidatorSignatures) < MinSignatures {
		return fmt.Errorf("validator_signatures: must have at least %d signature", MinSignatures)
	}

	if len(m.ValidatorSignatures) > MaxSignatures {
		return fmt.Errorf("validator_signatures: cannot exceed %d signatures, got %d", MaxSignatures, len(m.ValidatorSignatures))
	}

	// Validate each signature
	for i, sig := range m.ValidatorSignatures {
		if err := validation.ValidateBytes(sig, MinSignatureSize, MaxSignatureSize, fmt.Sprintf("validator_signatures[%d]", i)); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgBurnTokens
func (m *MsgBurnTokens) ValidateBasic() error {
	// Validate sender address
	if err := validation.ValidateAccAddress(m.Sender); err != nil {
		return fmt.Errorf("sender: %w", err)
	}

	// Validate target chain
	if err := validateChainID(m.TargetChain, "target_chain"); err != nil {
		return err
	}

	// Validate recipient (non-empty string)
	if err := validation.ValidateBoundedString(m.Recipient, 1, 256, "recipient"); err != nil {
		return err
	}

	// Validate amount (pointer type due to protoc-gen-go)
	if m.Amount == nil {
		return fmt.Errorf("amount cannot be nil")
	}
	if err := validation.ValidateCoin(*m.Amount, "amount"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgLinkAddress
func (m *MsgLinkAddress) ValidateBasic() error {
	// Validate AURA address
	if err := validation.ValidateAccAddress(m.AuraAddress); err != nil {
		return fmt.Errorf("aura_address: %w", err)
	}

	// Validate signer address
	if err := validation.ValidateAccAddress(m.Signer); err != nil {
		return fmt.Errorf("signer: %w", err)
	}

	// At least one of PAW or XAI address must be provided
	hasPaw := m.PawAddress != ""
	hasXai := m.XaiAddress != ""

	if !hasPaw && !hasXai {
		return fmt.Errorf("at least one of paw_address or xai_address must be provided")
	}

	// Validate PAW address if provided
	if hasPaw {
		if err := validation.ValidateBoundedString(m.PawAddress, 1, 256, "paw_address"); err != nil {
			return err
		}

		// Validate PAW signature if PAW address is provided
		if err := validation.ValidateBytes(m.PawSignature, MinSignatureSize, MaxSignatureSize, "paw_signature"); err != nil {
			return err
		}
	}

	// Validate XAI address if provided
	if hasXai {
		if err := validation.ValidateBoundedString(m.XaiAddress, 1, 256, "xai_address"); err != nil {
			return err
		}

		// Validate XAI signature if XAI address is provided
		if err := validation.ValidateBytes(m.XaiSignature, MinSignatureSize, MaxSignatureSize, "xai_signature"); err != nil {
			return err
		}
	}

	// Ensure signer is one of the addresses being linked
	if m.Signer != m.AuraAddress && m.Signer != m.PawAddress && m.Signer != m.XaiAddress {
		return fmt.Errorf("signer must be one of the addresses being linked")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCrossChainSwap
func (m *MsgCrossChainSwap) ValidateBasic() error {
	// Validate sender address
	if err := validation.ValidateAccAddress(m.Sender); err != nil {
		return fmt.Errorf("sender: %w", err)
	}

	// Validate source chain
	if err := validateChainID(m.SourceChain, "source_chain"); err != nil {
		return err
	}

	// Validate input coin (pointer type due to protoc-gen-go)
	if m.InputCoin == nil {
		return fmt.Errorf("input_coin cannot be nil")
	}
	if err := validation.ValidateCoin(*m.InputCoin, "input_coin"); err != nil {
		return err
	}

	// Validate target chain
	if err := validateChainID(m.TargetChain, "target_chain"); err != nil {
		return err
	}

	// Source and target chains must be different
	if m.SourceChain == m.TargetChain {
		return fmt.Errorf("source_chain and target_chain must be different")
	}

	// Validate target denom
	if err := validation.ValidateDenom(m.TargetDenom); err != nil {
		return fmt.Errorf("target_denom: %w", err)
	}

	// Validate minimum target amount (string type - customtype annotation ignored)
	if err := parseAndValidatePositiveInt(m.MinTargetAmount, "min_target_amount"); err != nil {
		return err
	}

	// Validate recipient (optional, but if present must be valid)
	if m.Recipient != "" {
		if err := validation.ValidateBoundedString(m.Recipient, 1, 256, "recipient"); err != nil {
			return err
		}
	}

	// Validate max slippage (basis points, max 10000 = 100%)
	if err := validation.ValidateBasisPoints(m.MaxSlippageBps, "max_slippage_bps"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgRelayTransfer
func (m *MsgRelayTransfer) ValidateBasic() error {
	// Validate relayer address
	if err := validation.ValidateAccAddress(m.Relayer); err != nil {
		return fmt.Errorf("relayer: %w", err)
	}

	// Validate transfer ID
	if err := validation.ValidateID(m.TransferId, "transfer_id"); err != nil {
		return err
	}

	// Validate target transaction hash
	if err := validation.ValidateHash(m.TargetTxHash); err != nil {
		return fmt.Errorf("target_tx_hash: %w", err)
	}

	// Validate status
	if err := validation.ValidateBoundedString(m.Status, 1, MaxStatusLength, "status"); err != nil {
		return err
	}

	return nil
}
