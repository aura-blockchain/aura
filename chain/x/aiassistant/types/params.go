package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

const (
	defaultMinStakeAmount int64 = 5_000_000
	defaultHeartbeatWindow      = 600
	defaultHeartbeatGrace       = 120
	defaultMaxLocales           = 5
)

func DefaultParams() Params {
	minStake := Balance{
		Denom:  DefaultStakeDenom,
		Amount: fmt.Sprintf("%d", defaultMinStakeAmount),
	}
	return Params{
		MinStake:               &minStake,
		HeartbeatWindowSeconds: defaultHeartbeatWindow,
		HeartbeatGraceSeconds:  defaultHeartbeatGrace,
		SlashFractionDowntime:  "0.02",
		SlashFractionMisbehavior:"0.10",
		MaxLocales:             defaultMaxLocales,
	}
}

func ValidateParams(p Params) error {
	if err := validateBalance(p.MinStake); err != nil {
		return fmt.Errorf("min_stake: %w", err)
	}
	if p.HeartbeatWindowSeconds == 0 {
		return fmt.Errorf("heartbeat_window_seconds must be positive")
	}
	if p.HeartbeatGraceSeconds == 0 {
		return fmt.Errorf("heartbeat_grace_seconds must be positive")
	}
	if p.HeartbeatGraceSeconds >= p.HeartbeatWindowSeconds {
		return fmt.Errorf("heartbeat_grace_seconds must be less than heartbeat window")
	}
	if _, err := sdkmath.LegacyNewDecFromStr(p.SlashFractionDowntime); err != nil {
		return fmt.Errorf("slash_fraction_downtime: %w", err)
	}
	if _, err := sdkmath.LegacyNewDecFromStr(p.SlashFractionMisbehavior); err != nil {
		return fmt.Errorf("slash_fraction_misbehavior: %w", err)
	}
	if p.MaxLocales == 0 {
		return fmt.Errorf("max_locales must be positive")
	}
	return nil
}

func validateBalance(balance *Balance) error {
	if balance == nil {
		return fmt.Errorf("balance required")
	}
	if balance.Denom == "" {
		return fmt.Errorf("denom cannot be empty")
	}
	if balance.Amount == "" {
		return fmt.Errorf("amount must be provided")
	}
	amount, ok := sdkmath.NewIntFromString(balance.Amount)
	if !ok || !amount.IsPositive() {
		return fmt.Errorf("amount must be a positive integer")
	}
	return nil
}
