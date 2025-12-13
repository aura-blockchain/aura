package testutil

import (
	"fmt"
	"testing"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// InvariantFunc is a function that checks an invariant
type InvariantFunc func(ctx sdk.Context) (string, bool)

// InvariantChecker provides utilities for checking module invariants
type InvariantChecker struct {
	t          *testing.T
	invariants []InvariantFunc
}

// NewInvariantChecker creates a new invariant checker
func NewInvariantChecker(t *testing.T) *InvariantChecker {
	return &InvariantChecker{
		t:          t,
		invariants: make([]InvariantFunc, 0),
	}
}

// RegisterInvariant registers an invariant to check
func (ic *InvariantChecker) RegisterInvariant(inv InvariantFunc) {
	ic.invariants = append(ic.invariants, inv)
}

// CheckAll checks all registered invariants
func (ic *InvariantChecker) CheckAll(ctx sdk.Context) {
	for i, inv := range ic.invariants {
		msg, broken := inv(ctx)
		require.False(ic.t, broken, "Invariant %d broken: %s", i, msg)
	}
}

// Common invariants that can be used across modules

// NonNegativeBalanceInvariant checks that all balances are non-negative
func NonNegativeBalanceInvariant(balances map[string]sdk.Coins) InvariantFunc {
	return func(ctx sdk.Context) (string, bool) {
		for addr, coins := range balances {
			if coins.IsAnyNegative() {
				return fmt.Sprintf("negative balance for %s: %s", addr, coins), true
			}
		}
		return "", false
	}
}

// TotalSupplyInvariant checks that total supply equals sum of all balances
func TotalSupplyInvariant(totalSupply sdk.Coins, balances map[string]sdk.Coins) InvariantFunc {
	return func(ctx sdk.Context) (string, bool) {
		sum := sdk.NewCoins()
		for _, coins := range balances {
			sum = sum.Add(coins...)
		}
		if !totalSupply.Equal(sum) {
			return fmt.Sprintf("total supply %s != sum of balances %s", totalSupply, sum), true
		}
		return "", false
	}
}

// ModuleAccountInvariant checks module account balances
func ModuleAccountInvariant(moduleAccounts map[string]sdk.Coins) InvariantFunc {
	return func(ctx sdk.Context) (string, bool) {
		for module, coins := range moduleAccounts {
			if coins.IsAnyNegative() {
				return fmt.Sprintf("module %s has negative balance: %s", module, coins), true
			}
		}
		return "", false
	}
}

// CountInvariant checks that a count matches expected value
func CountInvariant(name string, expected, actual int) InvariantFunc {
	return func(ctx sdk.Context) (string, bool) {
		if expected != actual {
			return fmt.Sprintf("%s count mismatch: expected %d, got %d", name, expected, actual), true
		}
		return "", false
	}
}

// StoreKeyExistsInvariant checks that specific keys exist in store
func StoreKeyExistsInvariant(store storetypes.KVStore, keys [][]byte) InvariantFunc {
	return func(ctx sdk.Context) (string, bool) {
		for _, key := range keys {
			if !store.Has(key) {
				return fmt.Sprintf("store missing required key: %x", key), true
			}
		}
		return "", false
	}
}

// NoOrphanedDataInvariant checks for orphaned data references
func NoOrphanedDataInvariant(parentKeys, childKeys [][]byte) InvariantFunc {
	return func(ctx sdk.Context) (string, bool) {
		parentSet := make(map[string]bool)
		for _, key := range parentKeys {
			parentSet[string(key)] = true
		}

		for _, childKey := range childKeys {
			if !parentSet[string(childKey)] {
				return fmt.Sprintf("orphaned child key found: %x", childKey), true
			}
		}
		return "", false
	}
}

// ValidatorPowerInvariant checks validator power consistency
func ValidatorPowerInvariant(validators []TestValidator) InvariantFunc {
	return func(ctx sdk.Context) (string, bool) {
		for _, val := range validators {
			if val.Tokens.IsNegative() {
				return fmt.Sprintf("validator %s has negative tokens: %s", val.Address, val.Tokens), true
			}
		}
		return "", false
	}
}

// SequentialIDInvariant checks that IDs are sequential with no gaps
func SequentialIDInvariant(ids []uint64) InvariantFunc {
	return func(ctx sdk.Context) (string, bool) {
		if len(ids) == 0 {
			return "", false
		}

		for i := 0; i < len(ids)-1; i++ {
			if ids[i+1] != ids[i]+1 {
				return fmt.Sprintf("ID gap found: %d -> %d", ids[i], ids[i+1]), true
			}
		}
		return "", false
	}
}

// TimestampOrderingInvariant checks that timestamps are in order
func TimestampOrderingInvariant(timestamps []int64, mustBeStrict bool) InvariantFunc {
	return func(ctx sdk.Context) (string, bool) {
		for i := 0; i < len(timestamps)-1; i++ {
			if mustBeStrict && timestamps[i+1] <= timestamps[i] {
				return fmt.Sprintf("timestamps not strictly ordered: %d >= %d", timestamps[i], timestamps[i+1]), true
			}
			if !mustBeStrict && timestamps[i+1] < timestamps[i] {
				return fmt.Sprintf("timestamps not ordered: %d > %d", timestamps[i], timestamps[i+1]), true
			}
		}
		return "", false
	}
}
