package keeper

import (
	"sync"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aura-bindings/types"
	vckeeper "github.com/aequitas/aura/chain/x/vcregistry/keeper"
)

// Keeper manages the aura-bindings module state and provides CosmWasm binding functionality
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	// Dependencies on other keepers
	vcKeeper *vckeeper.Keeper

	// Thread-safe state management
	mu sync.RWMutex

	// Query rate limiting
	queryRateLimits map[string]int // address -> query count per block
	currentBlock    int64

	// Statistics
	queryStats   map[string]uint64 // query type -> count
	messageStats map[string]uint64 // message type -> count
}

// NewKeeper creates a new aura-bindings Keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	vcKeeper *vckeeper.Keeper,
) *Keeper {
	return &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		vcKeeper:        vcKeeper,
		queryRateLimits: make(map[string]int),
		queryStats:      make(map[string]uint64),
		messageStats:    make(map[string]uint64),
		currentBlock:    0,
	}
}

// Logger returns a module-specific logger
func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetStore returns the module KVStore
func (k *Keeper) GetStore(ctx sdk.Context) storetypes.KVStore {
	return ctx.KVStore(k.storeKey)
}

// ResetQueryRateLimits resets query rate limits for a new block
func (k *Keeper) ResetQueryRateLimits(ctx sdk.Context) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if ctx.BlockHeight() > k.currentBlock {
		k.currentBlock = ctx.BlockHeight()
		k.queryRateLimits = make(map[string]int)
	}
}

// CheckQueryRateLimit checks if the address has exceeded query rate limits
func (k *Keeper) CheckQueryRateLimit(ctx sdk.Context, address string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Reset rate limits if we're in a new block (inline to avoid deadlock)
	if ctx.BlockHeight() > k.currentBlock {
		k.currentBlock = ctx.BlockHeight()
		k.queryRateLimits = make(map[string]int)
	}

	count := k.queryRateLimits[address]
	if count >= types.MaxQueriesPerBlock {
		return types.ErrQueryRateLimitExceeded
	}

	k.queryRateLimits[address]++
	return nil
}

// IncrementQueryStat increments the query statistics counter
func (k *Keeper) IncrementQueryStat(queryType string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.queryStats[queryType]++
}

// IncrementMessageStat increments the message statistics counter
func (k *Keeper) IncrementMessageStat(msgType string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.messageStats[msgType]++
}

// GetQueryStats returns the current query statistics
func (k *Keeper) GetQueryStats() map[string]uint64 {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Return a copy to prevent concurrent modification
	stats := make(map[string]uint64, len(k.queryStats))
	for k, v := range k.queryStats {
		stats[k] = v
	}
	return stats
}

// GetMessageStats returns the current message statistics
func (k *Keeper) GetMessageStats() map[string]uint64 {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Return a copy to prevent concurrent modification
	stats := make(map[string]uint64, len(k.messageStats))
	for k, v := range k.messageStats {
		stats[k] = v
	}
	return stats
}

// VCKeeper returns the VC registry keeper
func (k *Keeper) VCKeeper() *vckeeper.Keeper {
	return k.vcKeeper
}
