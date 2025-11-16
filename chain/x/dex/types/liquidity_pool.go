package types

import (
	"cosmossdk.io/math"
)

// LiquidityPool represents a liquidity pool
type LiquidityPool struct {
	PoolId    string              `json:"pool_id"`
	DenomA    string              `json:"denom_a"`
	DenomB    string              `json:"denom_b"`
	ReserveA  math.Int            `json:"reserve_a"`
	ReserveB  math.Int            `json:"reserve_b"`
	Providers []LiquidityProvider `json:"providers"`
}

// LiquidityProvider represents a liquidity provider's position
type LiquidityProvider struct {
	Address  string   `json:"address"`
	LpTokens math.Int `json:"lp_tokens"`
}
