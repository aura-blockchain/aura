// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ConfidenceScoreKeeper defines the expected interface for interacting with the confidencescore module
type ConfidenceScoreKeeper interface {
	// GetUserScore retrieves the total confidence score for a user
	GetUserScore(ctx sdk.Context, walletAddr string) (uint64, bool)

	// HasCompletedIR checks if a user has completed a specific inclusion routine
	HasCompletedIR(ctx sdk.Context, walletAddr string, irID string) bool

	// GetArenaScore retrieves the score for a specific arena
	GetArenaScore(ctx sdk.Context, walletAddr string, arena string) (uint64, error)

	// GetAnchorInfo retrieves anchor completion information for a user
	GetAnchorInfo(ctx sdk.Context, walletAddr string) (interface{}, bool)

	// IsVerified checks if a user has verified status (CS >= threshold)
	IsVerified(ctx sdk.Context, walletAddr string) bool
}
