// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

const (
	EventTypeCreatePool                = "create_pool"
	EventTypeSwap                      = "swap"
	EventTypeAddLiquidity              = "add_liquidity"
	EventTypeRemoveLiquidity           = "remove_liquidity"
	EventTypeUpdateFees                = "update_fees"
	EventTypeLiquidityLocked           = "liquidity_locked"
	EventTypeManipulationDetected      = "manipulation_detected"
	EventTypeCircuitBreakerActivated   = "circuit_breaker_activated"
	EventTypeCircuitBreakerDeactivated = "circuit_breaker_deactivated"
	EventTypeOrderCommitted            = "order_committed"
	EventTypeOrderRevealed             = "order_revealed"
	EventTypeCommitmentExpired         = "commitment_expired"
	EventTypeBatchExecuted             = "batch_executed"
	EventTypeBatchOrderExecuted        = "batch_order_executed"

	AttributeKeyPoolID      = "pool_id"
	AttributeKeyCreator     = "creator"
	AttributeKeyTrader      = "trader"
	AttributeKeyProvider    = "provider"
	AttributeKeyTokenA      = "token_a"
	AttributeKeyTokenB      = "token_b"
	AttributeKeyAmountA     = "amount_a"
	AttributeKeyAmountB     = "amount_b"
	AttributeKeyAmountIn    = "amount_in"
	AttributeKeyAmountOut   = "amount_out"
	AttributeKeyLiquidity   = "liquidity"
	AttributeKeyFeeRate     = "fee_rate"
	AttributeKeyLPTokens    = "lp_tokens"
	AttributeKeyPoolShare   = "pool_share"
	AttributeKeySender      = "sender"
	AttributeKeyPriceImpact = "price_impact"
)
