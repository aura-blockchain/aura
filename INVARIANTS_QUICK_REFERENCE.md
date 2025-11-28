# Invariants Quick Reference Guide

## Privacy Module Invariants

### 1. ParamsInvariant
- **Route**: `privacy/params-valid`
- **Purpose**: Validates module parameters
- **Key Checks**:
  - Parameters pass validation

### 2. EncryptionKeyValidityInvariant
- **Route**: `privacy/encryption-key-validity`
- **Purpose**: Validates encryption keys
- **Key Checks**:
  - ✓ Key ID not empty
  - ✓ Owner is valid address
  - ✓ Public key not empty
  - ✓ Algorithm in [aes-256-gcm, chacha20-poly1305, xchacha20-poly1305]
  - ✓ CreatedAt set
  - ✓ If rotated, RotatedAt must be set

### 3. MixingStateConsistencyInvariant
- **Route**: `privacy/mixing-state-consistency`
- **Purpose**: Validates mixing pool state
- **Key Checks**:
  - ✓ Pool ID not empty
  - ✓ Participant count >= 0
  - ✓ Min participants > 0
  - ✓ Max participants >= min participants
  - ✓ Participant count <= max participants
  - ✓ CreatedAt set
  - ✓ If active, StartedAt must be set

### 4. RingSignatureValidityInvariant
- **Route**: `privacy/ring-signature-validity`
- **Purpose**: Validates ring signatures
- **Key Checks**:
  - ✓ Signature ID not empty
  - ✓ Signature data not empty
  - ✓ Ring size > 0
  - ✓ Public keys count == ring size
  - ✓ All public keys not empty
  - ✓ Message hash not empty
  - ✓ CreatedAt set

---

## Contract Registry Module Invariants

### 1. ParamsInvariant
- **Route**: `contractregistry/params-valid`
- **Purpose**: Validates module parameters
- **Key Checks**:
  - Parameters pass validation

### 2. ContractMetadataConsistencyInvariant
- **Route**: `contractregistry/contract-metadata-consistency`
- **Purpose**: Validates contract metadata
- **Key Checks**:
  - ✓ Contract address is valid
  - ✓ Name not empty
  - ✓ Version not empty
  - ✓ Code hash not empty
  - ✓ Creator address valid (if set)
  - ✓ CreatedAt set

### 3. CodeHashValidityInvariant
- **Route**: `contractregistry/code-hash-validity`
- **Purpose**: Validates code hash format
- **Key Checks**:
  - ✓ Code hash length == 32 bytes (SHA256)

### 4. ContractAddressValidityInvariant
- **Route**: `contractregistry/contract-address-validity`
- **Purpose**: Validates contract addresses
- **Key Checks**:
  - ✓ All indexed addresses are valid

### 5. VersionConsistencyInvariant
- **Route**: `contractregistry/version-consistency`
- **Purpose**: Validates version information
- **Key Checks**:
  - ✓ Version length <= 100 chars
  - ✓ If UpdatedAt set, UpdatedAt >= CreatedAt

---

## Test Coverage Summary

| Module | Invariants | Test Functions | Lines of Test Code |
|--------|-----------|----------------|-------------------|
| Privacy | 4 | 25 | 601 |
| Contract Registry | 5 | 25 | 638 |
| **Total** | **9** | **50** | **1,239** |

---

## Common Test Patterns

### ✅ Valid State Tests
- `Test{Invariant}_EmptyStore` - Empty store always passes
- `Test{Invariant}_Valid{Data}` - Valid data passes

### ❌ Invalid State Tests
- `Test{Invariant}_Empty{Field}` - Missing required fields fail
- `Test{Invariant}_Invalid{Field}` - Malformed fields fail
- `Test{Invariant}_{LogicViolation}` - Logic constraint violations fail

### 🔄 Integration Tests
- `TestAllInvariants` - All invariants on empty store
- `TestAllInvariantsWithMultipleInvalidStates` - Detects any broken state
- `TestAllInvariantsWithValidData` - Validates correct states

---

## Running Tests

```bash
# Privacy module
go test -v ./x/privacy/keeper/... -run TestInvariantsTestSuite

# Contract Registry module
go test -v ./x/contractregistry/keeper/... -run TestInvariantsTestSuite

# All invariant tests
go test -v ./x/privacy/keeper/... ./x/contractregistry/keeper/... -run TestInvariantsTestSuite
```

---

## File Locations

### Privacy Module
- **Invariants**: `/chain/x/privacy/keeper/invariants.go`
- **Tests**: `/chain/x/privacy/keeper/invariants_test.go`
- **Keys**: `/chain/x/privacy/types/keys.go`

### Contract Registry Module
- **Invariants**: `/chain/x/contractregistry/keeper/invariants.go`
- **Tests**: `/chain/x/contractregistry/keeper/invariants_test.go`
- **Keys**: `/chain/x/contractregistry/types/keys.go`

---

## Key Store Prefixes

### Privacy Module
```go
EncryptionKeyKeyPrefix   = []byte{0x09}
RingSignatureKeyPrefix   = []byte{0x0a}
MixingPoolKeyPrefix      = []byte{0x07}
```

### Contract Registry Module
```go
ContractMetadataKeyPrefix        = []byte{0x14}
ContractAddressIndexKeyPrefix    = []byte{0x15}
```
