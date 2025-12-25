// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetIRPrerequisites returns the list of prerequisite IR IDs for the given IR
// This method is part of the IRRegistry interface used by the confidencescore module
func (k *Keeper) GetIRPrerequisites(ctx sdk.Context, irID string) ([]string, error) {
	// Check if IR exists
	if _, exists := k.GetIR(ctx, irID); !exists {
		return nil, types.ErrIRNotFound
	}

	// Get prerequisites
	prereq, hasPrereqs := k.GetPrerequisite(ctx, irID)
	if !hasPrereqs {
		return []string{}, nil // No prerequisites
	}

	return prereq.RequiredIrIds, nil
}

// IsIRActive checks if an IR is currently active
// This method is part of the IRRegistry interface used by the confidencescore module
func (k *Keeper) IsIRActive(ctx sdk.Context, irID string) bool {
	ir, exists := k.GetIR(ctx, irID)
	if !exists {
		return false
	}

	// Check status
	if ir.Status != types.IRStatus_IR_STATUS_ACTIVE {
		return false
	}

	// Check activation height
	currentHeight := ctx.BlockHeight()
	if currentHeight < ir.ActivationHeight {
		return false
	}

	// Check sunset height
	if ir.SunsetHeight > 0 && currentHeight >= ir.SunsetHeight {
		return false
	}

	return true
}

// GetIRScore returns the score value for completing an IR
// This method is part of the IRRegistry interface used by the confidencescore module
func (k *Keeper) GetIRScore(ctx sdk.Context, irID string) (uint64, error) {
	ir, exists := k.GetIR(ctx, irID)
	if !exists {
		return 0, types.ErrIRNotFound
	}

	if ir.Score < 0 {
		return 0, fmt.Errorf("invalid score for IR %s: %d", irID, ir.Score)
	}

	return uint64(ir.Score), nil
}

// GetIRArena returns the arena type for an IR
// This method is part of the IRRegistry interface used by the confidencescore module
func (k *Keeper) GetIRArena(ctx sdk.Context, irID string) (string, error) {
	ir, exists := k.GetIR(ctx, irID)
	if !exists {
		return "", types.ErrIRNotFound
	}

	// Convert Arena enum to string
	arenaStr := ir.Arena.String()
	return arenaStr, nil
}

// GetIRDifficulty returns the difficulty level for an IR
func (k *Keeper) GetIRDifficulty(ctx sdk.Context, irID string) (int32, error) {
	ir, exists := k.GetIR(ctx, irID)
	if !exists {
		return 0, types.ErrIRNotFound
	}

	// Difficulty can be derived from score
	// Higher score indicates higher difficulty
	difficulty := int32(ir.Score / 100)
	if difficulty < 1 {
		difficulty = 1
	}
	return difficulty, nil
}

// GetIRReward returns the POI reward for completing an IR
func (k *Keeper) GetIRReward(ctx sdk.Context, irID string) (int64, error) {
	ir, exists := k.GetIR(ctx, irID)
	if !exists {
		return 0, types.ErrIRNotFound
	}

	return ir.PoiReward, nil
}
