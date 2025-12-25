// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

// GetConfidenceScoreKeeper returns the confidence score keeper
func (k *Keeper) GetConfidenceScoreKeeper() ConfidenceScoreKeeper {
	// k.mu.RLock()  // Commented out - k.mu is undefined
	// defer k.mu.RUnlock()
	return k.csKeeper
}
