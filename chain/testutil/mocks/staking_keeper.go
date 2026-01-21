// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"bytes"
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MockStakingKeeper implements a mock StakingKeeper for testing
type MockStakingKeeper struct {
	Validators      map[string]stakingtypes.Validator
	Delegations     map[string]map[string]stakingtypes.Delegation
	UnbondingTime   int64
	BondDenomValue  string
	ValidatorPowers map[string]int64
}

// NewMockStakingKeeper creates a new mock staking keeper
func NewMockStakingKeeper() *MockStakingKeeper {
	return &MockStakingKeeper{
		Validators:      make(map[string]stakingtypes.Validator),
		Delegations:     make(map[string]map[string]stakingtypes.Delegation),
		UnbondingTime:   21 * 24 * 60 * 60, // 21 days in seconds
		BondDenomValue:  "uaura",
		ValidatorPowers: make(map[string]int64),
	}
}

// GetValidator returns a validator by operator address
func (m *MockStakingKeeper) GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, bool) {
	val, ok := m.Validators[addr.String()]
	return val, ok
}

// SetValidator sets a validator
func (m *MockStakingKeeper) SetValidator(ctx context.Context, validator stakingtypes.Validator) {
	m.Validators[validator.OperatorAddress] = validator
}

// GetDelegation returns a delegation
func (m *MockStakingKeeper) GetDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.Delegation, bool) {
	if delMap, ok := m.Delegations[delAddr.String()]; ok {
		if del, ok := delMap[valAddr.String()]; ok {
			return del, true
		}
	}
	return stakingtypes.Delegation{}, false
}

// SetDelegation sets a delegation
func (m *MockStakingKeeper) SetDelegation(ctx context.Context, delegation stakingtypes.Delegation) {
	delAddr := delegation.DelegatorAddress
	valAddr := delegation.ValidatorAddress

	if _, ok := m.Delegations[delAddr]; !ok {
		m.Delegations[delAddr] = make(map[string]stakingtypes.Delegation)
	}
	m.Delegations[delAddr][valAddr] = delegation
}

// GetAllValidators returns all validators
func (m *MockStakingKeeper) GetAllValidators(ctx context.Context) []stakingtypes.Validator {
	validators := make([]stakingtypes.Validator, 0, len(m.Validators))
	for _, val := range m.Validators {
		validators = append(validators, val)
	}
	return validators
}

// GetBondedValidatorsByPower returns bonded validators sorted by power
func (m *MockStakingKeeper) GetBondedValidatorsByPower(ctx context.Context) []stakingtypes.Validator {
	// Simple implementation - return all bonded validators
	validators := make([]stakingtypes.Validator, 0)
	for _, val := range m.Validators {
		if val.IsBonded() {
			validators = append(validators, val)
		}
	}
	return validators
}

// Jail jails a validator
func (m *MockStakingKeeper) Jail(ctx context.Context, valAddr sdk.ValAddress) error {
	if val, ok := m.Validators[valAddr.String()]; ok {
		val.Jailed = true
		m.Validators[valAddr.String()] = val
	}
	return nil
}

// Unjail unjails a validator
func (m *MockStakingKeeper) Unjail(ctx context.Context, valAddr sdk.ValAddress) error {
	if val, ok := m.Validators[valAddr.String()]; ok {
		val.Jailed = false
		m.Validators[valAddr.String()] = val
	}
	return nil
}

// Slash slashes a validator
func (m *MockStakingKeeper) Slash(ctx context.Context, valAddr sdk.ValAddress, fraction math.LegacyDec) error {
	if val, ok := m.Validators[valAddr.String()]; ok {
		slashAmount := fraction.MulInt(val.Tokens).TruncateInt()
		val.Tokens = val.Tokens.Sub(slashAmount)
		m.Validators[valAddr.String()] = val
	}
	return nil
}

// GetParams returns staking params (mock)
func (m *MockStakingKeeper) GetParams(ctx context.Context) stakingtypes.Params {
	return stakingtypes.DefaultParams()
}

// BondDenom returns the bond denom
func (m *MockStakingKeeper) BondDenom(ctx context.Context) (string, error) {
	return m.BondDenomValue, nil
}

// PowerReduction returns the power reduction constant
func (m *MockStakingKeeper) PowerReduction(ctx context.Context) math.Int {
	return math.NewInt(1000000) // 1 token = 1,000,000 utoken
}

// ValidatorByConsAddr returns a validator by consensus address
func (m *MockStakingKeeper) ValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (stakingtypes.Validator, bool) {
	// Simple implementation - iterate through validators
	for _, val := range m.Validators {
		valConsAddr, err := val.GetConsAddr()
		if err == nil && bytes.Equal(valConsAddr, consAddr.Bytes()) {
			return val, true
		}
	}
	return stakingtypes.Validator{}, false
}

// RemoveValidator removes a validator (test helper)
func (m *MockStakingKeeper) RemoveValidator(valAddr sdk.ValAddress) {
	delete(m.Validators, valAddr.String())
}

// SetValidatorPower sets validator power (test helper)
func (m *MockStakingKeeper) SetValidatorPower(valAddr sdk.ValAddress, power int64) {
	m.ValidatorPowers[valAddr.String()] = power
}
