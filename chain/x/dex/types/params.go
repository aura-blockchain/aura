package types

import (
	"cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyIRBoostEnabled      = []byte("IRBoostEnabled")
	KeyIRBoostPercent      = []byte("IRBoostPercent")
	KeyAuthority           = []byte("Authority")
	KeyAuthorityExpiration = []byte("AuthorityExpiration")
	KeyGovernanceEnabled   = []byte("GovernanceEnabled")
)

// Params defines the parameters for the DEX module
type Params struct {
	IrBoostEnabled      bool            `json:"ir_boost_enabled"`
	IrBoostPercent      uint32          `json:"ir_boost_percent"`
	Authority           string          `json:"authority"`
	AuthorityExpiration int64           `json:"authority_expiration"`
	GovernanceEnabled   bool            `json:"governance_enabled"`
	MinLiquidityTiers   []LiquidityTier `json:"min_liquidity_tiers"`
}

// LiquidityTier defines a tier for dynamic minimum liquidity
type LiquidityTier struct {
	MaxAuraPriceUsd math.LegacyDec `json:"max_aura_price_usd"`
	MinLiquidityUsd math.LegacyDec `json:"min_liquidity_usd"`
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		IrBoostEnabled:      true,
		IrBoostPercent:      40,
		Authority:           "",
		AuthorityExpiration: 0,
		GovernanceEnabled:   false,
		MinLiquidityTiers:   DefaultLiquidityTiers(),
	}
}

// DefaultLiquidityTiers returns the default liquidity tier configuration
func DefaultLiquidityTiers() []LiquidityTier {
	return []LiquidityTier{
		{MaxAuraPriceUsd: math.LegacyNewDecWithPrec(50, 2), MinLiquidityUsd: math.LegacyNewDec(1000)}, // < $0.50: $1,000
		{MaxAuraPriceUsd: math.LegacyNewDec(1), MinLiquidityUsd: math.LegacyNewDec(2500)},             // < $1.00: $2,500
		{MaxAuraPriceUsd: math.LegacyNewDec(5), MinLiquidityUsd: math.LegacyNewDec(5000)},             // < $5.00: $5,000
		{MaxAuraPriceUsd: math.LegacyZeroDec(), MinLiquidityUsd: math.LegacyNewDec(10000)},            // >= $5.00: $10,000
	}
}

// ParamKeyTable returns the parameter key table
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyIRBoostEnabled, &p.IrBoostEnabled, validateBool),
		paramtypes.NewParamSetPair(KeyIRBoostPercent, &p.IrBoostPercent, validateUint32),
		paramtypes.NewParamSetPair(KeyAuthority, &p.Authority, validateString),
		paramtypes.NewParamSetPair(KeyAuthorityExpiration, &p.AuthorityExpiration, validateInt64),
		paramtypes.NewParamSetPair(KeyGovernanceEnabled, &p.GovernanceEnabled, validateBool),
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return ErrInvalidParam
	}
	return nil
}

func validateUint32(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return ErrInvalidParam
	}
	return nil
}

func validateString(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return ErrInvalidParam
	}
	return nil
}

func validateInt64(i interface{}) error {
	_, ok := i.(int64)
	if !ok {
		return ErrInvalidParam
	}
	return nil
}
