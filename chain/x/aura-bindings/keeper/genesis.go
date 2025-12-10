package keeper

import (
	"sort"

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

	// Import query statistics (deterministic ordering)
	if data.QueryStats != nil {
		k.queryStats = make(map[string]uint64, len(data.QueryStats))
		queryTypes := make([]string, 0, len(data.QueryStats))
		for queryType := range data.QueryStats {
			queryTypes = append(queryTypes, queryType)
		}
		sort.Strings(queryTypes)
		for _, queryType := range queryTypes {
			k.queryStats[queryType] = data.QueryStats[queryType]
		}
	} else {
		k.queryStats = make(map[string]uint64)
	}

	// Import message statistics (deterministic ordering)
	if data.MessageStats != nil {
		k.messageStats = make(map[string]uint64, len(data.MessageStats))
		msgTypes := make([]string, 0, len(data.MessageStats))
		for msgType := range data.MessageStats {
			msgTypes = append(msgTypes, msgType)
		}
		sort.Strings(msgTypes)
		for _, msgType := range msgTypes {
			k.messageStats[msgType] = data.MessageStats[msgType]
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

	// Export query statistics (deterministic ordering)
	queryStats := make(map[string]uint64, len(k.queryStats))
	queryTypes := make([]string, 0, len(k.queryStats))
	for queryType := range k.queryStats {
		queryTypes = append(queryTypes, queryType)
	}
	sort.Strings(queryTypes)
	for _, queryType := range queryTypes {
		queryStats[queryType] = k.queryStats[queryType]
	}

	// Export message statistics (deterministic ordering)
	messageStats := make(map[string]uint64, len(k.messageStats))
	msgTypes := make([]string, 0, len(k.messageStats))
	for msgType := range k.messageStats {
		msgTypes = append(msgTypes, msgType)
	}
	sort.Strings(msgTypes)
	for _, msgType := range msgTypes {
		messageStats[msgType] = k.messageStats[msgType]
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
