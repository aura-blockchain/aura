package gasmetering

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MeteredStore wraps a KVStore with automatic gas metering
type MeteredStore struct {
	store  storetypes.KVStore
	ctx    sdk.Context
	config GasConfig
}

// NewMeteredStore creates a new gas-metered store wrapper
func NewMeteredStore(ctx context.Context, store storetypes.KVStore, config GasConfig) *MeteredStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &MeteredStore{
		store:  store,
		ctx:    sdkCtx,
		config: config,
	}
}

// Get reads a value and charges gas
func (ms *MeteredStore) Get(key []byte) []byte {
	ms.ctx.GasMeter().ConsumeGas(ms.config.StoreReadCost, "store read")
	value := ms.store.Get(key)

	// Charge for data size
	if value != nil {
		ms.ctx.GasMeter().ConsumeGas(uint64(len(value)), "store read data")
	}

	return value
}

// Set writes a value and charges gas
func (ms *MeteredStore) Set(key, value []byte) {
	// Charge for write operation
	ms.ctx.GasMeter().ConsumeGas(ms.config.StoreWriteCost, "store write")

	// Charge for data size
	totalSize := uint64(len(key) + len(value))
	ms.ctx.GasMeter().ConsumeGas(totalSize, "store write data")

	ms.store.Set(key, value)
}

// Delete removes a value and charges gas
func (ms *MeteredStore) Delete(key []byte) {
	ms.ctx.GasMeter().ConsumeGas(ms.config.StoreDeleteCost, "store delete")
	ms.store.Delete(key)
}

// Has checks existence and charges gas
func (ms *MeteredStore) Has(key []byte) bool {
	ms.ctx.GasMeter().ConsumeGas(ms.config.StoreHasCost, "store has")
	return ms.store.Has(key)
}

// Iterator creates a gas-metered iterator
func (ms *MeteredStore) Iterator(start, end []byte) storetypes.Iterator {
	ms.ctx.GasMeter().ConsumeGas(ms.config.IterationBaseCost, "iterator creation")
	return newMeteredIterator(ms.ctx, ms.store.Iterator(start, end), ms.config)
}

// ReverseIterator creates a gas-metered reverse iterator
func (ms *MeteredStore) ReverseIterator(start, end []byte) storetypes.Iterator {
	ms.ctx.GasMeter().ConsumeGas(ms.config.IterationBaseCost, "reverse iterator creation")
	return newMeteredIterator(ms.ctx, ms.store.ReverseIterator(start, end), ms.config)
}

// meteredIterator wraps an iterator with gas metering
type meteredIterator struct {
	iter      storetypes.Iterator
	ctx       sdk.Context
	config    GasConfig
	count     uint32
	maxCount  uint32
}

func newMeteredIterator(ctx sdk.Context, iter storetypes.Iterator, config GasConfig) *meteredIterator {
	return &meteredIterator{
		iter:     iter,
		ctx:      ctx,
		config:   config,
		count:    0,
		maxCount: config.MaxIterationResults,
	}
}

// Valid checks if iterator is valid
func (mi *meteredIterator) Valid() bool {
	// Check iteration limit
	if mi.count >= mi.maxCount {
		return false
	}
	return mi.iter.Valid()
}

// Next advances the iterator and charges gas
func (mi *meteredIterator) Next() {
	mi.ctx.GasMeter().ConsumeGas(mi.config.StoreIterationCost, "iterator next")
	mi.count++
	mi.iter.Next()
}

// Key returns the current key and charges gas
func (mi *meteredIterator) Key() []byte {
	key := mi.iter.Key()
	mi.ctx.GasMeter().ConsumeGas(uint64(len(key)), "iterator key")
	return key
}

// Value returns the current value and charges gas
func (mi *meteredIterator) Value() []byte {
	value := mi.iter.Value()
	mi.ctx.GasMeter().ConsumeGas(uint64(len(value)), "iterator value")
	return value
}

// Close closes the iterator
func (mi *meteredIterator) Close() error {
	return mi.iter.Close()
}

// Domain returns the iteration domain
func (mi *meteredIterator) Domain() ([]byte, []byte) {
	return mi.iter.Domain()
}

// Error returns any error
func (mi *meteredIterator) Error() error {
	return mi.iter.Error()
}

// GetCount returns the number of iterations performed
func (mi *meteredIterator) GetCount() uint32 {
	return mi.count
}

// IterateWithLimit iterates with a result limit and gas metering
func IterateWithLimit(
	ctx context.Context,
	store storetypes.KVStore,
	prefix []byte,
	limit uint32,
	config GasConfig,
	callback func(key, value []byte) error,
) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Charge for starting iteration
	sdkCtx.GasMeter().ConsumeGas(config.IterationBaseCost, "begin iteration")

	it := storetypes.KVStorePrefixIterator(store, prefix)
	defer it.Close()

	count := uint32(0)
	for ; it.Valid(); it.Next() {
		// Check limit
		if count >= limit {
			return fmt.Errorf("iteration limit exceeded: %d", limit)
		}

		// Charge gas per iteration
		sdkCtx.GasMeter().ConsumeGas(config.StoreIterationCost, "iterate")

		// Charge for key/value size
		key := it.Key()
		value := it.Value()
		sdkCtx.GasMeter().ConsumeGas(uint64(len(key)+len(value)), "iteration data")

		// Execute callback
		if err := callback(key, value); err != nil {
			return err
		}

		count++
	}

	return nil
}

// ConsumeGasForOperation charges gas for a named operation
func ConsumeGasForOperation(ctx context.Context, operation string, cost uint64) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.GasMeter().ConsumeGas(cost, operation)
}

// ConsumeGasForCrypto charges gas for cryptographic operations
func ConsumeGasForCrypto(ctx context.Context, operation string, config GasConfig) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	var cost uint64
	switch operation {
	case "hash":
		cost = config.HashCost
	case "signature_verify":
		cost = config.SignatureVerifyCost
	case "pubkey_derive":
		cost = config.PublicKeyDeriveCost
	case "ecdsa_verify":
		cost = config.ECDSAVerifyCost
	case "ed25519_verify":
		cost = config.ED25519VerifyCost
	case "merkle_verify":
		cost = config.MerkleProofVerifyCost
	case "ring_signature_verify":
		cost = config.RingSignatureVerifyCost
	case "zk_proof_verify":
		cost = config.ZKProofVerifyCost
	default:
		cost = 1000 // Default
	}

	sdkCtx.GasMeter().ConsumeGas(cost, fmt.Sprintf("crypto: %s", operation))
}

// ConsumeGasForValidation charges gas for validation operations
func ConsumeGasForValidation(ctx context.Context, validationType string, config GasConfig) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	var cost uint64
	switch validationType {
	case "address":
		cost = config.AddressValidationCost
	case "signature":
		cost = config.SignatureCheckCost
	case "amount":
		cost = config.AmountValidationCost
	case "string":
		cost = config.StringValidationCost
	default:
		cost = 500
	}

	sdkCtx.GasMeter().ConsumeGas(cost, fmt.Sprintf("validate: %s", validationType))
}
