// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MigrateStakeToUaura migrates governance params from "stake" to "uaura" denomination.
// This should be called during a chain upgrade when fixing the denomination mismatch.
func (k *Keeper) MigrateStakeToUaura(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	if params == nil {
		ctx.Logger().Info("governance params not set, skipping migration")
		return nil
	}

	modified := false

	// Migrate main min_deposit
	if strings.Contains(params.MinDeposit, "stake") {
		params.MinDeposit = strings.ReplaceAll(params.MinDeposit, "stake", "uaura")
		modified = true
		ctx.Logger().Info("migrated governance min_deposit", "new_value", params.MinDeposit)
	}

	// Migrate category-specific params
	for category, categoryParams := range params.CategoryParams {
		if categoryParams != nil && strings.Contains(categoryParams.MinDeposit, "stake") {
			categoryParams.MinDeposit = strings.ReplaceAll(categoryParams.MinDeposit, "stake", "uaura")
			modified = true
			ctx.Logger().Info("migrated category min_deposit",
				"category", category,
				"new_value", categoryParams.MinDeposit)
		}
	}

	if modified {
		k.SetParams(ctx, params)
		ctx.Logger().Info("governance params migration complete")
	} else {
		ctx.Logger().Info("governance params already use uaura, no migration needed")
	}

	return nil
}
