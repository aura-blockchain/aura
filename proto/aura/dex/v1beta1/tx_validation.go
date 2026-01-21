package v1beta1

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/aequitas/aura/proto/common/validation"
)

const (
	// MaxSecretLength is the maximum size for an HTLC secret
	MaxSecretLength = 256
	// MinSecretLength is the minimum size for an HTLC secret (32 bytes for hash)
	MinSecretLength = 32
	// MaxSecretHashLength is the maximum size for an HTLC secret hash
	MaxSecretHashLength = 128
	// MinSecretHashLength is the minimum size for an HTLC secret hash
	MinSecretHashLength = 32
	// MaxTimelockDuration is the maximum HTLC timelock duration (30 days in seconds)
	MaxTimelockDuration = uint64(30 * 24 * 60 * 60)
	// MinTimelockDuration is the minimum HTLC timelock duration (1 hour in seconds)
	MinTimelockDuration = uint64(60 * 60)
	// MaxSlippageBps is the maximum allowed slippage in basis points (100%)
	MaxSlippageBps = uint64(10000)
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

// ValidateBasic implements the sdk.Msg interface for MsgCreatePool
func (m *MsgCreatePool) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Validate denom A
	if err := validation.ValidateDenom(m.DenomA); err != nil {
		return fmt.Errorf("denom_a: %w", err)
	}

	// Validate denom B
	if err := validation.ValidateDenom(m.DenomB); err != nil {
		return fmt.Errorf("denom_b: %w", err)
	}

	// Denoms must be different
	if m.DenomA == m.DenomB {
		return fmt.Errorf("denom_a and denom_b must be different")
	}

	// Validate amount A (value type with gogoproto.nullable = false)
	if err := validation.ValidateCoin(m.AmountA, "amount_a"); err != nil {
		return err
	}

	// Validate amount B (value type with gogoproto.nullable = false)
	if err := validation.ValidateCoin(m.AmountB, "amount_b"); err != nil {
		return err
	}

	// Ensure denoms match
	if m.AmountA.Denom != m.DenomA {
		return fmt.Errorf("amount_a denom must match denom_a")
	}

	if m.AmountB.Denom != m.DenomB {
		return fmt.Errorf("amount_b denom must match denom_b")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAddLiquidity
func (m *MsgAddLiquidity) ValidateBasic() error {
	// Validate provider address
	if err := validation.ValidateAccAddress(m.Provider); err != nil {
		return fmt.Errorf("provider: %w", err)
	}

	// Validate pool ID
	if err := validation.ValidateID(m.PoolId, "pool_id"); err != nil {
		return err
	}

	// Validate amount A (value type with gogoproto.nullable = false)
	if err := validation.ValidateCoin(m.AmountA, "amount_a"); err != nil {
		return err
	}

	// Validate amount B (value type with gogoproto.nullable = false)
	if err := validation.ValidateCoin(m.AmountB, "amount_b"); err != nil {
		return err
	}

	// Amounts must have different denoms
	if m.AmountA.Denom == m.AmountB.Denom {
		return fmt.Errorf("amount_a and amount_b must have different denoms")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgRemoveLiquidity
func (m *MsgRemoveLiquidity) ValidateBasic() error {
	// Validate provider address
	if err := validation.ValidateAccAddress(m.Provider); err != nil {
		return fmt.Errorf("provider: %w", err)
	}

	// Validate pool ID
	if err := validation.ValidateID(m.PoolId, "pool_id"); err != nil {
		return err
	}

	// Validate LP tokens amount (cosmossdk.io/math.Int value type)
	if err := validation.ValidatePositiveInt(m.LpTokens, "lp_tokens"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgSwapExactIn
func (m *MsgSwapExactIn) ValidateBasic() error {
	// Validate sender address
	if err := validation.ValidateAccAddress(m.Sender); err != nil {
		return fmt.Errorf("sender: %w", err)
	}

	// Validate pool ID
	if err := validation.ValidateID(m.PoolId, "pool_id"); err != nil {
		return err
	}

	// Validate coin in (value type with gogoproto.nullable = false)
	if err := validation.ValidateCoin(m.CoinIn, "coin_in"); err != nil {
		return err
	}

	// Validate minimum amount out (cosmossdk.io/math.Int value type)
	if err := validation.ValidatePositiveInt(m.MinAmountOut, "min_amount_out"); err != nil {
		return err
	}

	// Validate max slippage (basis points, max 10000 = 100%)
	if err := validation.ValidateBasisPoints(m.MaxSlippageBps, "max_slippage_bps"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCreateOrder
func (m *MsgCreateOrder) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Validate order type (enum validation - will be validated at protobuf level)
	// OrderType is an enum, so we just check it's within valid range
	if m.OrderType < 0 {
		return fmt.Errorf("order_type must be valid")
	}

	// Validate AURA amount (cosmossdk.io/math.Int value type)
	if err := validation.ValidatePositiveInt(m.AuraAmount, "aura_amount"); err != nil {
		return err
	}

	// Validate other coin denom
	if err := validation.ValidateDenom(m.OtherCoin); err != nil {
		return fmt.Errorf("other_coin: %w", err)
	}

	// Validate other amount (cosmossdk.io/math.Int value type)
	if err := validation.ValidatePositiveInt(m.OtherAmount, "other_amount"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCancelOrder
func (m *MsgCancelOrder) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Validate order ID
	if err := validation.ValidateID(m.OrderId, "order_id"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgExecuteSwap
func (m *MsgExecuteSwap) ValidateBasic() error {
	// Validate initiator address
	if err := validation.ValidateAccAddress(m.Initiator); err != nil {
		return fmt.Errorf("initiator: %w", err)
	}

	// Validate order ID
	if err := validation.ValidateID(m.OrderId, "order_id"); err != nil {
		return err
	}

	// Validate secret (must be non-empty for HTLC)
	if err := validation.ValidateBoundedString(m.Secret, MinSecretLength, MaxSecretLength, "secret"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCreateHTLC
func (m *MsgCreateHTLC) ValidateBasic() error {
	// Validate sender address
	if err := validation.ValidateAccAddress(m.Sender); err != nil {
		return fmt.Errorf("sender: %w", err)
	}

	// Validate recipient address
	if err := validation.ValidateAccAddress(m.Recipient); err != nil {
		return fmt.Errorf("recipient: %w", err)
	}

	// Sender and recipient must be different
	if m.Sender == m.Recipient {
		return fmt.Errorf("sender and recipient must be different")
	}

	// Validate amount (value type with gogoproto.nullable = false)
	if err := validation.ValidateCoin(m.Amount, "amount"); err != nil {
		return err
	}

	// Validate secret hash
	if err := validation.ValidateHash(m.SecretHash); err != nil {
		return fmt.Errorf("secret_hash: %w", err)
	}

	// Validate timelock duration (must be reasonable)
	if err := validation.ValidateBoundedUint64(m.TimelockDuration, MinTimelockDuration, MaxTimelockDuration, "timelock_duration"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgClaimHTLC
func (m *MsgClaimHTLC) ValidateBasic() error {
	// Validate recipient address
	if err := validation.ValidateAccAddress(m.Recipient); err != nil {
		return fmt.Errorf("recipient: %w", err)
	}

	// Validate HTLC ID
	if err := validation.ValidateID(m.HtlcId, "htlc_id"); err != nil {
		return err
	}

	// Validate secret (must be non-empty)
	if err := validation.ValidateBoundedString(m.Secret, MinSecretLength, MaxSecretLength, "secret"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgRefundHTLC
func (m *MsgRefundHTLC) ValidateBasic() error {
	// Validate sender address
	if err := validation.ValidateAccAddress(m.Sender); err != nil {
		return fmt.Errorf("sender: %w", err)
	}

	// Validate HTLC ID
	if err := validation.ValidateID(m.HtlcId, "htlc_id"); err != nil {
		return err
	}

	return nil
}
