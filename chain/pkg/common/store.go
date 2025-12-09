package common

import (
	"fmt"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Marshallable is a constraint for types that can be marshalled/unmarshalled.
// All protobuf-generated types satisfy this constraint.
type Marshallable interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

// GetObject retrieves an object from the store using generics.
// This eliminates repetitive marshal/unmarshal code across modules.
//
// Type parameters:
//   - T: Type of object to retrieve (must be a pointer to a protobuf type)
//
// Parameters:
//   - ctx: SDK context for store access
//   - store: KVStore to read from
//   - cdc: Binary codec for unmarshalling
//   - key: Store key for the object
//
// Returns:
//   - T: Retrieved object (zero value if not found)
//   - bool: True if object exists, false if not found
//   - error: Unmarshal error if object exists but cannot be decoded
//
// Security considerations:
//   - Returns error on unmarshal failure (prevents using corrupted data)
//   - Clear indication of not found vs error states
//
// Example usage:
//   var pool dextypes.LiquidityPool
//   pool, found, err := common.GetObject(ctx, store, k.cdc, poolKey, &pool)
//   if err != nil {
//       return nil, err
//   }
//   if !found {
//       return nil, types.ErrPoolNotFound
//   }
func GetObject[T any](ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec, key []byte, obj T) (T, bool, error) {
	bz := store.Get(key)
	if bz == nil {
		var zero T
		return zero, false, nil
	}

	// obj must be a pointer for Unmarshal to work
	if err := cdc.Unmarshal(bz, obj.(codec.ProtoMarshaler)); err != nil {
		var zero T
		return zero, false, fmt.Errorf("failed to unmarshal object: %w", err)
	}

	return obj, true, nil
}

// SetObject stores an object in the store using generics.
// This eliminates repetitive marshal/set code across modules.
//
// Type parameters:
//   - T: Type of object to store (must be a protobuf type)
//
// Parameters:
//   - ctx: SDK context for store access
//   - store: KVStore to write to
//   - cdc: Binary codec for marshalling
//   - key: Store key for the object
//   - obj: Object to store
//
// Returns:
//   - error: Marshal error if object cannot be encoded
//
// Security considerations:
//   - Returns error on marshal failure (prevents storing invalid data)
//   - Atomic operation (no partial writes on error)
//
// Example usage:
//   pool := &dextypes.LiquidityPool{...}
//   if err := common.SetObject(ctx, store, k.cdc, poolKey, pool); err != nil {
//       return err
//   }
func SetObject[T codec.ProtoMarshaler](ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec, key []byte, obj T) error {
	bz, err := cdc.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal object: %w", err)
	}

	store.Set(key, bz)
	return nil
}

// DeleteObject removes an object from the store.
// This provides consistent delete semantics across modules.
//
// Parameters:
//   - store: KVStore to delete from
//   - key: Store key of the object to delete
//
// Note:
//   - Deleting a non-existent key is a no-op (not an error)
//
// Example usage:
//   common.DeleteObject(store, poolKey)
func DeleteObject(store storetypes.KVStore, key []byte) {
	store.Delete(key)
}

// HasObject checks if an object exists in the store without deserializing it.
// This is more efficient than GetObject when you only need to check existence.
//
// Parameters:
//   - store: KVStore to check
//   - key: Store key to check
//
// Returns:
//   - bool: True if key exists, false otherwise
//
// Example usage:
//   if common.HasObject(store, poolKey) {
//       // Pool exists
//   }
func HasObject(store storetypes.KVStore, key []byte) bool {
	return store.Has(key)
}

// IterateObjects iterates over all objects with a given prefix using a callback.
// This provides a consistent iteration pattern across modules.
//
// Type parameters:
//   - T: Type of objects to iterate (must be a pointer to a protobuf type)
//
// Parameters:
//   - ctx: SDK context for store access
//   - store: KVStore to iterate over
//   - cdc: Binary codec for unmarshalling
//   - prefix: Key prefix to iterate over
//   - callback: Function called for each object (return error to stop iteration)
//
// Returns:
//   - error: First error returned by callback, or unmarshal error
//
// Security considerations:
//   - Stops iteration on first error
//   - Properly closes iterator on error or success
//   - Returns unmarshal errors (prevents processing corrupted data)
//
// Example usage:
//   err := common.IterateObjects(ctx, store, k.cdc, PoolPrefix, func(key []byte, pool *dextypes.LiquidityPool) error {
//       // Process pool
//       return nil
//   })
func IterateObjects[T codec.ProtoMarshaler](
	ctx sdk.Context,
	store storetypes.KVStore,
	cdc codec.BinaryCodec,
	keyPrefix []byte,
	callback func(key []byte, obj T) error,
) error {
	prefixStore := prefix.NewStore(store, keyPrefix)
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Create new instance for each iteration
		var obj T
		if err := cdc.Unmarshal(iterator.Value(), obj); err != nil {
			return fmt.Errorf("failed to unmarshal object at key %x: %w", iterator.Key(), err)
		}

		if err := callback(iterator.Key(), obj); err != nil {
			return err
		}
	}

	return nil
}
