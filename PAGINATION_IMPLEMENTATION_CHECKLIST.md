# Pagination Implementation Checklist

## Quick Reference
- **Total**: 12 modules, 40 methods
- **Files to modify**: 12 proto files + 12 keeper files = 24 files
- **Pattern**: Follow contractregistry, dex, identity, governance, or wasm

---

## HIGH PRIORITY (28 methods across 4 modules)

### 1. economics (9 methods)
**Files**: `proto/aura/economics/v1beta1/query.proto`, `chain/x/economics/keeper/query_server.go`

- [ ] AllVestingSchedules (line 127)
- [ ] Deposits (line 288)
- [ ] PendingTreasuryTxs (needs line number check)
- [ ] Proposals (line 163)
- [ ] TokenomicsStats (needs line number check)
- [ ] VestingSchedulesByAddress (line 77)
- [ ] VoteDelegations (line 403)
- [ ] VoteLocksByOwner (line 338)
- [ ] Votes (line 240)

### 2. auth (7 methods)
**Files**: `proto/aura/auth/v1beta1/query.proto`, `chain/x/auth/keeper/query_server.go`

- [ ] HasPermission (returns role assignments - check if truly a list)
- [ ] ListEmergencyAdmins
- [ ] ListMultisigProposals
- [ ] ListMultisigWallets
- [ ] ListRoles
- [ ] ListSessions
- [ ] ListTimeLockedActions

### 3. bridge (7 methods)
**Files**: `proto/aura/bridge/v1beta1/query.proto`, `chain/x/bridge/keeper/query_server.go`

- [ ] AllChains
- [ ] AllTransfers
- [ ] AllWrappedTokens
- [ ] UserTransfers
- [ ] Validators
- [ ] collectValidators (internal helper - may not need proto)
- [ ] validatorsFromChains (internal helper - may not need proto)

### 4. networksecurity (5 methods)
**Files**: `proto/aura/networksecurity/v1beta1/query.proto`, `chain/x/networksecurity/keeper/query_server.go`

- [ ] AllPeers
- [ ] ForkAlerts
- [ ] NetworkHealth (check if truly returns list)
- [ ] PartitionAlerts
- [ ] TrustedPeers

---

## MEDIUM PRIORITY (6 methods across 2 modules)

### 5. security (4 methods)
**Files**: `proto/aura/security/v1beta1/query.proto`, `chain/x/security/keeper/query_server.go`

- [ ] AllIncidents
- [ ] AllKeyRotationSchedules
- [ ] AllMixingPools
- [ ] AllPeers

### 6. validatorsecurity (2 methods)
**Files**: `proto/aura/validatorsecurity/v1beta1/query.proto`, `chain/x/validatorsecurity/keeper/query_server.go`

- [ ] AllValidators
- [ ] DoubleSignEvidences

---

## LOW PRIORITY (6 methods across 6 modules - 1 each)

### 7. aiassistant
**Files**: `proto/aura/aiassistant/v1beta1/query.proto`, `chain/x/aiassistant/keeper/query_server.go`
- [ ] Assistants

### 8. aura-bindings
**Files**: `proto/aura/aura-bindings/v1beta1/query.proto`, `chain/x/aura-bindings/keeper/query_server.go`
- [ ] AllStats

### 9. compliance
**Files**: `proto/aura/compliance/v1beta1/query.proto`, `chain/x/compliance/keeper/query_server.go`
- [ ] TransactionAlerts

### 10. incidentresponse
**Files**: `proto/aura/incidentresponse/v1beta1/query.proto`, `chain/x/incidentresponse/keeper/query_server.go`
- [ ] GetAllIncidents

### 11. prevalidation
**Files**: `proto/aura/prevalidation/v1beta1/query.proto`, `chain/x/prevalidation/keeper/query_server.go`
- [ ] AllTemplates

### 12. privacy
**Files**: `proto/aura/privacy/v1beta1/query.proto`, `chain/x/privacy/keeper/query_server.go`
- [ ] MixingPools

---

## Implementation Steps (Per Module)

1. **Update Proto** (query.proto)
   - Add `cosmos.base.query.v1beta1.PageRequest pagination = N;` to Request
   - Add `cosmos.base.query.v1beta1.PageResponse pagination = N;` to Response

2. **Regenerate Code**
   ```bash
   make proto-gen
   ```

3. **Update Keeper** (query_server.go)
   - Import: `"github.com/cosmos/cosmos-sdk/types/query"`
   - Replace manual iteration with `query.Paginate()`
   - Return `PageResponse` with total count

4. **Test**
   ```bash
   go test ./x/<module>/keeper/... -v
   ```

---

## Reference Implementation

See these modules for working examples:
- `x/contractregistry/keeper/query_server.go` (manual pagination)
- `x/identity/keeper/query_server.go` (query.Paginate pattern)
- `x/dex/keeper/query_server.go`
- `x/governance/keeper/query_server.go`
- `x/wasm/keeper/query_server.go`
