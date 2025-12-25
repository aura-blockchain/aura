# Common Package

## Overview

The Common package provides shared utilities for Aura blockchain modules to ensure determinism, optimize performance, and enforce security invariants. It is NOT a Cosmos SDK module but a library package imported by other modules.

## Components

### Cache (`cache/`)
Multi-layer caching with LRU eviction, compression, and metrics tracking for frequently accessed data.

**Key Features**:
- L1 in-memory cache with expiration
- L2 compressed cache for larger data
- Cache warming for predictable performance
- Hit/miss metrics and compression ratio tracking

### Determinism (`determinism/`)
Deterministic time and randomness utilities to ensure consensus safety.

**Key Features**:
- `GetBlockTime()`: Use block time instead of `time.Now()`
- `GetBlockHeight()`: Current block height
- `IsExpired()`: Check expiration using block time
- Deterministic random number generation with block seed

**CRITICAL**: All keeper code MUST use these utilities instead of `time.Now()` or `math/rand` to maintain consensus.

### Gas Metering (`gasmetering/`)
Custom gas configuration and metered KV store wrappers for fine-grained gas control.

**Key Features**:
- Custom gas configs per module
- Metered store wrapper with gas tracking
- Gas cost profiling and optimization

### Optimization (`optimization/`)
Performance optimization utilities for batch operations, memoization, and indexing.

**Key Features**:
- Batch write operations to reduce store overhead
- Memoization for expensive computations
- Secondary index management
- Integration utilities for cross-module optimization

### Validation (`validation/`)
Common validation functions for addresses, amounts, timestamps, and data structures.

**Key Features**:
- Address validation (AccAddress, ValAddress, ConsAddress)
- Amount validation (positive, non-negative, range checks)
- Timestamp validation (not future, within range)
- String sanitization and length checks

### Testing (`testing/`)
Inter-module security and integration tests.

**Key Features**:
- Fuzzing tests for cross-module interactions
- Message flow security tests
- Access control verification

## Usage

Import the common package in your module:

```go
import (
    "github.com/aequitas/aura/chain/x/common/cache"
    "github.com/aequitas/aura/chain/x/common/determinism"
    "github.com/aequitas/aura/chain/x/common/validation"
)

// Use deterministic time
blockTime := determinism.GetBlockTime(ctx)

// Validate address
if err := validation.ValidateAccAddress(address); err != nil {
    return err
}

// Use cache
cache := cache.NewCache(cache.DefaultConfig())
```

## Security Notes

- NEVER use `time.Now()` - always use `determinism.GetBlockTime(ctx)`
- NEVER use `math/rand` - always use `determinism` package for randomness
- All validation functions return errors - check them
- Cache is NOT deterministic - only use for read-heavy query optimization
