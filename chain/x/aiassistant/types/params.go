// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
)

const (
	defaultMinStakeAmount  int64 = 5_000_000
	defaultHeartbeatWindow       = 600
	defaultHeartbeatGrace        = 120
	defaultMaxLocales            = 5
)

func DefaultParams() Params {
	return Params{
		MinStake: Balance{
			Denom:  DefaultStakeDenom,
			Amount: sdkmath.NewInt(defaultMinStakeAmount),
		},
		HeartbeatWindowSeconds:   defaultHeartbeatWindow,
		HeartbeatGraceSeconds:    defaultHeartbeatGrace,
		SlashFractionDowntime:    sdkmath.LegacyMustNewDecFromStr("0.02"),
		SlashFractionMisbehavior: sdkmath.LegacyMustNewDecFromStr("0.10"),
		MaxLocales:               defaultMaxLocales,
	}
}

func ValidateParams(p Params) error {
	if err := validateBalance(p.MinStake, true); err != nil {
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
	if p.SlashFractionDowntime.IsNil() {
		return fmt.Errorf("slash_fraction_downtime must be set")
	}
	if p.SlashFractionDowntime.IsNegative() {
		return fmt.Errorf("slash_fraction_downtime cannot be negative")
	}
	if p.SlashFractionMisbehavior.IsNil() {
		return fmt.Errorf("slash_fraction_misbehavior must be set")
	}
	if p.SlashFractionMisbehavior.IsNegative() {
		return fmt.Errorf("slash_fraction_misbehavior cannot be negative")
	}
	if p.MaxLocales == 0 {
		return fmt.Errorf("max_locales must be positive")
	}
	return nil
}

func validateBalance(balance Balance, requirePositive bool) error {
	if strings.TrimSpace(balance.Denom) == "" {
		return fmt.Errorf("denom cannot be empty")
	}
	if balance.Amount.IsNil() {
		return fmt.Errorf("amount must be provided")
	}
	if requirePositive && !balance.Amount.IsPositive() {
		return fmt.Errorf("amount must be positive")
	}
	if !requirePositive && balance.Amount.IsNegative() {
		return fmt.Errorf("amount cannot be negative")
	}
	return nil
}
