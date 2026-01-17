// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var _ stakingtypes.StakingHooks = Hooks{}

// Hooks wraps the validatorsecurity keeper to implement staking hooks
type Hooks struct {
	k Keeper
}

// NewHooks creates a new Hooks instance
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// AfterValidatorCreated auto-registers validator in validatorsecurity module
func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	validatorAddress := valAddr.String()

	// Check if already registered
	_, err := h.k.GetValidatorSecurityInfo(ctx, validatorAddress)
	if err == nil {
		// Already registered
		return nil
	}

	// Auto-register with minimal security info
	h.k.Logger(sdkCtx).Info("auto-registering validator from staking hook",
		"validator", validatorAddress)

	return h.k.RegisterValidator(ctx, validatorAddress, "", "", "", "", 0, 0, nil)
}

// BeforeValidatorModified is called before a validator is modified
func (h Hooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	return nil
}

// AfterValidatorRemoved is called after a validator is removed
func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

// AfterValidatorBonded tracks when validator becomes active
func (h Hooks) AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	validatorAddress := valAddr.String()

	info, err := h.k.GetValidatorSecurityInfo(ctx, validatorAddress)
	if err != nil {
		// Not registered in validatorsecurity, skip
		return nil
	}

	// Clear jail state when bonded
	if info.IsJailed {
		h.k.Logger(sdkCtx).Info("clearing jail state on validator bond",
			"validator", validatorAddress)
		info.IsJailed = false
		info.JailedUntil = nil
		h.k.SetValidatorSecurityInfo(ctx, info)
	}

	return nil
}

// AfterValidatorBeginUnbonding syncs jail state when validator begins unbonding
// This is critical: slashing module jails by calling staking.Jail which triggers unbonding
func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	validatorAddress := valAddr.String()

	info, err := h.k.GetValidatorSecurityInfo(ctx, validatorAddress)
	if err != nil {
		// Not registered in validatorsecurity, auto-register with jailed state
		h.k.Logger(sdkCtx).Info("auto-registering jailed validator from unbonding hook",
			"validator", validatorAddress)
		if err := h.k.RegisterValidator(ctx, validatorAddress, "", "", "", "", 0, 0, nil); err != nil {
			return err
		}
		// Get the newly created info
		info, err = h.k.GetValidatorSecurityInfo(ctx, validatorAddress)
		if err != nil {
			return err
		}
	}

	// Check if this unbonding is due to jailing by querying slashing module
	if h.k.slashingKeeper.IsTombstoned(ctx, consAddr) {
		// Validator is tombstoned, sync that state
		if !info.IsTombstoned {
			h.k.Logger(sdkCtx).Warn("syncing tombstoned state from slashing module",
				"validator", validatorAddress)
			tombstonedAt := sdkCtx.BlockTime()
			info.IsTombstoned = true
			info.TombstonedAt = &tombstonedAt
			info.IsJailed = true
			h.k.SetValidatorSecurityInfo(ctx, info)
		}
	} else if !info.IsJailed {
		// Validator is jailed but not tombstoned, sync jail state
		h.k.Logger(sdkCtx).Info("syncing jailed state from staking hook",
			"validator", validatorAddress)
		info.IsJailed = true
		h.k.SetValidatorSecurityInfo(ctx, info)
	}

	return nil
}

// BeforeDelegationCreated is called before a delegation is created
func (h Hooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeDelegationSharesModified is called before delegation shares are modified
func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeDelegationRemoved is called before a delegation is removed
func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// AfterDelegationModified is called after a delegation is modified
func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeValidatorSlashed syncs state BEFORE slashing occurs
// This is the most critical hook for jail state synchronization
func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction math.LegacyDec) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	validatorAddress := valAddr.String()

	info, err := h.k.GetValidatorSecurityInfo(ctx, validatorAddress)
	if err != nil {
		// Not registered, auto-register
		h.k.Logger(sdkCtx).Info("auto-registering validator from slashing hook",
			"validator", validatorAddress)
		if err := h.k.RegisterValidator(ctx, validatorAddress, "", "", "", "", 0, 0, nil); err != nil {
			return err
		}
		info, err = h.k.GetValidatorSecurityInfo(ctx, validatorAddress)
		if err != nil {
			return err
		}
	}

	// Mark as jailed before the slash is applied
	if !info.IsJailed {
		h.k.Logger(sdkCtx).Warn("marking validator as jailed before slash",
			"validator", validatorAddress,
			"slash_fraction", fraction.String())
		info.IsJailed = true
		h.k.SetValidatorSecurityInfo(ctx, info)
	}

	return nil
}

// AfterUnbondingInitiated is called after unbonding is initiated
func (h Hooks) AfterUnbondingInitiated(ctx context.Context, id uint64) error {
	return nil
}
