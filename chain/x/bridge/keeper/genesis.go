package keeper

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

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

	var maxTransferCounter uint64
	seenTransferIDs := make(map[uint64]bool)

	for _, transfer := range data.Transfers {
		if transfer == nil {
			continue
		}

		// Detect duplicate transfer IDs during import
		if seq, ok := parseTransferSequence(transfer.TransferId); ok {
			if seenTransferIDs[seq] {
				panic(fmt.Sprintf("duplicate transfer ID in genesis: %s", transfer.TransferId))
			}
			seenTransferIDs[seq] = true

			if seq > maxTransferCounter {
				maxTransferCounter = seq
			}
		}

		k.setTransfer(ctx, transfer)
	}

	// CRITICAL: Set counter to MAX+1 to prevent duplicate IDs
	// The next transfer will use this counter, so it must be one higher than the max
	if maxTransferCounter > 0 {
		bz := make([]byte, 8)
		binary.BigEndian.PutUint64(bz, maxTransferCounter+1)
		k.store(ctx).Set(types.TransferCounterKey, bz)
	}

	for _, cfg := range data.ChainConfigs {
		if cfg == nil {
			continue
		}
		k.setChainConfig(ctx, *cfg)
	}

	for _, validator := range data.Validators {
		k.setValidator(ctx, validator)
	}

	for _, token := range data.WrappedTokens {
		k.setWrappedToken(ctx, token)
	}

	for _, identity := range data.SharedIdentities {
		k.setSharedIdentity(ctx, identity)
	}

	for _, swap := range data.CrossChainSwaps {
		k.setSwap(ctx, swap)
	}

	for _, stats := range data.RelayerStats {
		k.setRelayerStats(ctx, stats)
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
	wrappedPtrs := make([]*types.WrappedToken, 0, len(wrappedTokens))
	for i := range wrappedTokens {
		token := wrappedTokens[i]
		tokenCopy := token
		wrappedPtrs = append(wrappedPtrs, &tokenCopy)
	}

	chainConfigs := k.getAllChainConfigs(ctx)
	chainConfigPtrs := make([]*types.ChainConfig, 0, len(chainConfigs))
	for i := range chainConfigs {
		cfg := chainConfigs[i]
		cfgCopy := cfg
		chainConfigPtrs = append(chainConfigPtrs, &cfgCopy)
	}

	// Export processed source hashes for replay attack prevention
	processedHashes := k.GetAllProcessedSourceHashes(ctx)
	processedHashList := make([]string, 0, len(processedHashes))
	for compositeKey := range processedHashes {
		processedHashList = append(processedHashList, compositeKey)
	}

	return types.GenesisState{
		Params:                 bridgeParamsToProto(params),
		Transfers:              k.getAllTransfers(ctx),
		ChainConfigs:           chainConfigPtrs,
		Validators:             k.getAllValidators(ctx),
		WrappedTokens:          wrappedPtrs,
		SharedIdentities:       k.getAllSharedIdentities(ctx),
		CrossChainSwaps:        k.getAllSwaps(ctx),
		RelayerStats:           k.getAllRelayerStats(ctx),
		ProcessedSourceHashes:  processedHashList,
	}
}

func bridgeParamsFromProto(params *types.BridgeParams) types.Params {
	if params == nil {
		return types.DefaultParams()
	}
	return types.Params{
		BridgeEnabled:                params.Enabled,
		MinConfirmations:             params.MinConfirmations,
		BridgeFeeBasisPoints:         params.BridgeFeeBasisPoints,
		MaxTransferAmount:            params.MaxTransferAmount,
		ValidatorThresholdPercentage: params.ValidatorThresholdPercentage,
	}
}

func bridgeParamsToProto(params types.Params) *types.BridgeParams {
	return &types.BridgeParams{
		Enabled:                      params.BridgeEnabled,
		MinConfirmations:             params.MinConfirmations,
		BridgeFeeBasisPoints:         params.BridgeFeeBasisPoints,
		MaxTransferAmount:            params.MaxTransferAmount,
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
