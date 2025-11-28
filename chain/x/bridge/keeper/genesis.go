package keeper

import (
	"encoding/binary"
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
	for _, transfer := range data.Transfers {
		if transfer == nil {
			continue
		}
		k.setTransfer(ctx, transfer)
		if seq, ok := parseTransferSequence(transfer.TransferId); ok && seq > maxTransferCounter {
			maxTransferCounter = seq
		}
	}
	if maxTransferCounter > 0 {
		bz := make([]byte, 8)
		binary.BigEndian.PutUint64(bz, maxTransferCounter)
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

	return types.GenesisState{
		Params:           bridgeParamsToProto(params),
		Transfers:        k.getAllTransfers(ctx),
		ChainConfigs:     chainConfigPtrs,
		Validators:       k.getAllValidators(ctx),
		WrappedTokens:    wrappedPtrs,
		SharedIdentities: k.getAllSharedIdentities(ctx),
		CrossChainSwaps:  k.getAllSwaps(ctx),
		RelayerStats:     k.getAllRelayerStats(ctx),
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
