---
status: pending
priority: p0
issue_id: "100"
tags: [code-review, build, critical, testnet-blocker]
dependencies: []
---

# P0 CRITICAL: Build Compilation Completely Broken

## Problem Statement

The Aura blockchain project **cannot compile**. Running `go build` or `go test` fails with hundreds of errors. This is a **COMPLETE BLOCKER** for any testnet deployment.

**Why it matters:** Without a buildable project, there is no way to run a 4-node testnet, execute tests, or verify any functionality.

## Findings

### Root Causes Identified

1. **Protobuf Type Incompatibility**
   - `github_com_cosmos_cosmos_sdk_types.Int` is a constant string, not a type
   - `github_com_cosmos_cosmos_sdk_types.Dec` is undefined
   - Files affected: `proto/aura/bridge/v1beta1/bridge.pb.go`, `proto/aura/dex/v1beta1/*.pb.go`

2. **gRPC Gateway ProtoMessage Interface**
   - Generated `.pb.gw.go` files don't implement `protoreflect.ProtoMessage`
   - Files affected: All `query.pb.gw.go` files across modules

3. **Service Descriptor Naming**
   - `Msg_ServiceDesc` vs `Msg_serviceDesc` case mismatch
   - Files affected: `chain/x/*/types/codec.go` (vcregistry, auth, compliance, etc.)

4. **Genesis Type Mismatches**
   - Pointer vs value type inconsistencies in genesis structs
   - `*v1beta1.Params` vs `v1beta1.Params` mismatches
   - Files affected: `chain/x/auth/types/genesis.go`, `chain/x/compliance/types/genesis.go`

5. **Timestamp/Duration Type Conflicts**
   - `durationpb.Duration` vs `gogoproto/types.Duration` incompatibility
   - `timestamppb.Timestamp` vs `gogoproto/types.Timestamp` incompatibility
   - Files affected: `chain/x/walletsecurity/client/cli/tx.go`, `chain/x/dataregistry/keeper/*.go`

6. **Missing Type Definitions**
   - `RawContractMessage` undefined in `proto/aura/wasm/v1beta1/tx.pb.go`

### Build Output Sample

```
# github.com/aequitas/aura/proto/aura/bridge/v1beta1
bridge.pb.go:80:9: github_com_cosmos_cosmos_sdk_types.Int is not a type

# github.com/aequitas/aura/chain/x/vcregistry/types
codec.go:20:60: undefined: vcregistrypb.Msg_ServiceDesc (but have Msg_serviceDesc)

# github.com/aequitas/aura/proto/aura/wasm/v1beta1
tx.pb.go:134:6: undefined: RawContractMessage
```

## Proposed Solutions

### Solution A: Regenerate All Protobufs (Recommended)
**Effort:** 1-2 days | **Risk:** Medium

1. Update proto files to use correct Cosmos SDK 0.53.4 types
2. Regenerate all `.pb.go` files with correct protoc plugins
3. Fix any remaining type mismatches

```bash
cd /home/decri/blockchain-projects/aura/chain
make proto-gen  # Or equivalent protobuf regeneration
go mod tidy
go build ./...
```

**Pros:**
- Fixes root cause
- Ensures SDK compatibility

**Cons:**
- May require proto file updates
- Risk of breaking API changes

### Solution B: Manual Type Fixes
**Effort:** 3-5 days | **Risk:** High

Fix each type error individually:
- Replace `github_com_cosmos_cosmos_sdk_types.Int` with `sdkmath.Int`
- Fix service descriptor names
- Convert pointer/value types

**Pros:**
- Targeted fixes

**Cons:**
- Time-consuming
- Error-prone
- Doesn't fix root cause

### Solution C: Pin Working SDK Version
**Effort:** 1 day | **Risk:** Medium

Check if there's a version mismatch between SDK and generated code:
- Verify `go.mod` SDK version matches proto generation
- Consider downgrading/upgrading SDK version

## Recommended Action

**GO WITH SOLUTION A**: Regenerate protobufs with correct Cosmos SDK 0.53.4 settings.

## Technical Details

### Affected Files (Sample)

**Proto Generated Files:**
- `proto/aura/bridge/v1beta1/bridge.pb.go`
- `proto/aura/bridge/v1beta1/genesis.pb.go`
- `proto/aura/bridge/v1beta1/security.pb.go`
- `proto/aura/dex/v1beta1/*.pb.go`
- `proto/aura/wasm/v1beta1/*.pb.go`
- `proto/aura/*/v1beta1/query.pb.gw.go` (all modules)

**Chain Type Files:**
- `chain/x/vcregistry/types/codec.go`
- `chain/x/auth/types/codec.go`
- `chain/x/auth/types/genesis.go`
- `chain/x/compliance/types/codec.go`
- `chain/x/compliance/types/genesis.go`
- `chain/x/walletsecurity/client/cli/tx.go`
- `chain/x/dataregistry/keeper/*.go`

### Database/State Changes
None - this is a compilation issue.

## Acceptance Criteria

- [ ] `go build ./cmd/aurad` succeeds without errors
- [ ] `go test ./...` runs (may have test failures, but compiles)
- [ ] `./aurad version` outputs version info
- [ ] Node can start: `./aurad start --home ~/.aura`

## Work Log

| Date | Action | Result |
|------|--------|--------|
| 2025-12-08 | Identified build failure | Compilation errors across 80+ packages |

## Resources

- [Cosmos SDK 0.53 Migration Guide](https://docs.cosmos.network/v0.53/migrations)
- [Protobuf Generation for Cosmos](https://docs.cosmos.network/v0.53/building-apps/proto)
- Related: ROADMAP_PRODUCTION.md claims build is passing - needs verification
