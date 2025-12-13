package keeper

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strconv"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

const transferIDPrefix = "transfer-"

// InitGenesis initializes the bridge module state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) error {
	if err := types.ValidateGenesis(&data); err != nil {
		return err
	}
	params := bridgeParamsFromProto(data.Params)
	if err := k.SetParams(ctx, params); err != nil {
		return err
	}

	// MIGRATION NOTE: With deterministic IDs based on block height + tx hash,
	// we no longer need a counter for NEW transfers. However, for backward
	// compatibility, we must properly restore the counter for chains that have
	// legacy sequential IDs (e.g., "transfer-1", "transfer-2").
	//
	// This prevents the critical off-by-one error where:
	//   WRONG: counter = max(existing IDs)     → next ID duplicates last ID
	//   RIGHT: counter = max(existing IDs) + 1 → next ID is unique
	//
	// For backward compatibility during migration:
	//   - Old counter-based IDs (e.g., "transfer-1", "transfer-2") are valid
	//   - New IDs use format: "transfer-{hash}" (large 64-bit values)
	//   - Both formats coexist in the same chain
	//   - Counter tracks legacy IDs only (for chains being imported/restored)

	seenTransferIDs := make(map[string]bool)
	var maxTransferCounter uint64 = 0
	legacyIDsFound := false

	for i := range data.Transfers {
		transfer := &data.Transfers[i]

		// Skip empty/nil transfers gracefully - they may appear in malformed genesis
		// but should not cause import failure
		if transfer.TransferId == "" {
			continue
		}

		// CRITICAL SECURITY: Detect duplicate transfer IDs during import
		// Duplicate IDs would cause silent overwrites of existing transfers
		if seenTransferIDs[transfer.TransferId] {
			return fmt.Errorf("duplicate transfer ID in genesis: %s", transfer.TransferId)
		}
		seenTransferIDs[transfer.TransferId] = true

		// Track the maximum counter value from legacy sequential IDs
		// This is needed for proper counter restoration (max+1, not max)
		if seq, ok := parseTransferSequence(transfer.TransferId); ok {
			// Legacy sequential IDs are small values (< 1 trillion threshold)
			// Hash-based IDs are large (> 1 trillion)
			const legacyIDThreshold = uint64(1 << 40) // 1 trillion
			if seq < legacyIDThreshold {
				legacyIDsFound = true
				if seq > maxTransferCounter {
					maxTransferCounter = seq
				}
			}
		}

		k.setTransfer(ctx, transfer)
	}

	// CRITICAL FIX: Restore counter to MAX + 1 (not MAX) to prevent ID collision
	// This ensures the next transfer gets a unique ID, not a duplicate
	//
	// Example scenario (the bug this fixes):
	//   Genesis has: transfer-1, transfer-2, transfer-5
	//   WRONG: counter = 5 → next transfer gets ID 5 → COLLISION with existing transfer-5
	//   RIGHT: counter = 6 → next transfer gets ID 6 → no collision
	//
	// Edge case: transfer-0 exists → counter should be 1 (0+1)
	//
	// Only restore counter if legacy sequential IDs were found in genesis
	// (chains using only new hash-based IDs don't need counter restoration)
	if legacyIDsFound {
		bz := make([]byte, 8)
		binary.BigEndian.PutUint64(bz, maxTransferCounter+1) // +1 CRITICAL: next available ID
		k.store(ctx).Set(types.TransferCounterKey, bz)
	}

	for _, cfg := range data.ChainConfigs {
		if err := k.setChainConfig(ctx, cfg); err != nil {
			return err
		}
	}

	for i := range data.Validators {
		k.setValidator(ctx, &data.Validators[i])
	}

	for i := range data.WrappedTokens {
		k.setWrappedToken(ctx, &data.WrappedTokens[i])
	}

	for i := range data.SharedIdentities {
		k.setSharedIdentity(ctx, &data.SharedIdentities[i])
	}

	for i := range data.CrossChainSwaps {
		k.setSwap(ctx, &data.CrossChainSwaps[i])
	}

	for i := range data.RelayerStats {
		k.setRelayerStats(ctx, &data.RelayerStats[i])
	}

	// Import processed source hashes for replay attack prevention
	for _, compositeKey := range data.ProcessedSourceHashes {
		if compositeKey != "" {
			k.SetProcessedSourceHash(ctx, compositeKey)
		}
	}

	return nil
}

// ExportGenesis exports the bridge module state to genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	params := k.GetParams(ctx)
	wrappedTokens := k.getAllWrappedTokens(ctx)

	chainConfigs := k.getAllChainConfigs(ctx)

	// Export processed source hashes for replay attack prevention
	// CONSENSUS-CRITICAL: Sort keys for deterministic iteration order
	processedHashes := k.GetAllProcessedSourceHashes(ctx)
	processedHashList := make([]string, 0, len(processedHashes))
	for compositeKey := range processedHashes {
		processedHashList = append(processedHashList, compositeKey)
	}
	sort.Strings(processedHashList)

	// Get all transfers and convert to non-pointer slice
	transfersPtrs := k.getAllTransfers(ctx)
	transfers := make([]types.CrossChainTransfer, 0, len(transfersPtrs))
	for _, t := range transfersPtrs {
		if t != nil {
			transfers = append(transfers, *t)
		}
	}

	// Get all validators and convert to non-pointer slice
	validatorsPtrs := k.getAllValidators(ctx)
	validators := make([]types.BridgeValidator, 0, len(validatorsPtrs))
	for _, v := range validatorsPtrs {
		if v != nil {
			validators = append(validators, *v)
		}
	}

	// Get all shared identities and convert to non-pointer slice
	identitiesPtrs := k.getAllSharedIdentities(ctx)
	identities := make([]types.SharedIdentity, 0, len(identitiesPtrs))
	for _, id := range identitiesPtrs {
		if id != nil {
			identities = append(identities, *id)
		}
	}

	// Get all swaps and convert to non-pointer slice
	swapsPtrs := k.getAllSwaps(ctx)
	swaps := make([]types.CrossChainSwap, 0, len(swapsPtrs))
	for _, s := range swapsPtrs {
		if s != nil {
			swaps = append(swaps, *s)
		}
	}

	// Get all relayer stats and convert to non-pointer slice
	statsPtrs := k.getAllRelayerStats(ctx)
	stats := make([]types.RelayerStats, 0, len(statsPtrs))
	for _, st := range statsPtrs {
		if st != nil {
			stats = append(stats, *st)
		}
	}

	return types.GenesisState{
		Params:                 bridgeParamsToProto(params),
		Transfers:              transfers,
		ChainConfigs:           chainConfigs,
		Validators:             validators,
		WrappedTokens:          wrappedTokens,
		SharedIdentities:       identities,
		CrossChainSwaps:        swaps,
		RelayerStats:           stats,
		ProcessedSourceHashes:  processedHashList,
	}
}

func bridgeParamsFromProto(params types.BridgeParams) types.Params {
	// Start with defaults to ensure all required fields are populated
	result := types.DefaultParams()
	// Override with genesis values
	result.BridgeEnabled = params.Enabled
	result.MinConfirmations = params.MinConfirmations
	result.BridgeFeeBasisPoints = params.BridgeFeeBasisPoints
	result.MaxTransferAmount = params.MaxTransferAmount.String()
	result.ValidatorThresholdPercentage = params.ValidatorThresholdPercentage
	return result
}

func bridgeParamsToProto(params types.Params) types.BridgeParams {
	// Parse MaxTransferAmount from string to math.Int
	maxTransferAmt, ok := sdkmath.NewIntFromString(params.MaxTransferAmount)
	if !ok {
		// Default fallback if parsing fails
		maxTransferAmt = sdkmath.NewInt(1000000000)
	}
	return types.BridgeParams{
		Enabled:                      params.BridgeEnabled,
		MinConfirmations:             params.MinConfirmations,
		BridgeFeeBasisPoints:         params.BridgeFeeBasisPoints,
		MaxTransferAmount:            maxTransferAmt,
		ValidatorThresholdPercentage: params.ValidatorThresholdPercentage,
	}
}

func parseTransferSequence(transferID string) (uint64, bool) {
	if !strings.HasPrefix(transferID, transferIDPrefix) {
		return 0, false
	}
	seqStr := strings.TrimPrefix(transferID, transferIDPrefix)
	seq, err := strconv.ParseUint(seqStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return seq, true
}
