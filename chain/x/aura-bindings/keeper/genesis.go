package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/aequitas/aura/chain/x/aura-bindings/types"
)

// InitGenesis initializes the module state from genesis data
func (k *Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Validate genesis state
	if err := data.Validate(); err != nil {
		return err
	}

	// Import query statistics
	if data.QueryStats != nil {
		k.queryStats = make(map[string]uint64, len(data.QueryStats))
		for queryType, count := range data.QueryStats {
			k.queryStats[queryType] = count
		}
	} else {
		k.queryStats = make(map[string]uint64)
	}

	// Import message statistics
	if data.MessageStats != nil {
		k.messageStats = make(map[string]uint64, len(data.MessageStats))
		for msgType, count := range data.MessageStats {
			k.messageStats[msgType] = count
		}
	} else {
		k.messageStats = make(map[string]uint64)
	}

	// Reset rate limits for the current block
	k.currentBlock = ctx.BlockHeight()
	k.queryRateLimits = make(map[string]int)

	k.Logger(ctx).Info("initialized aura-bindings genesis",
		"query_stat_count", len(k.queryStats),
		"message_stat_count", len(k.messageStats),
	)

	return nil
}

// ExportGenesis exports the current module state to genesis
func (k *Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Export query statistics
	queryStats := make(map[string]uint64, len(k.queryStats))
	for queryType, count := range k.queryStats {
		queryStats[queryType] = count
	}

	// Export message statistics
	messageStats := make(map[string]uint64, len(k.messageStats))
	for msgType, count := range k.messageStats {
		messageStats[msgType] = count
	}

	k.Logger(ctx).Info("exported aura-bindings genesis",
		"query_stat_count", len(queryStats),
		"message_stat_count", len(messageStats),
	)

	return types.GenesisState{
		QueryStats:   queryStats,
		MessageStats: messageStats,
	}
}
