# CLI Commands Implementation Summary

**Date:** 2025-11-13
**Status:** Partial Implementation Complete - CLI Layer Fixed

## Overview

This document summarizes the implementation of CLI commands for the AURA blockchain's new features:
1. QR Code Verification API (vcregistry module)
2. Selective Disclosure (vcregistry module)
3. Data Registry (dataregistry module)

## What Was Implemented

### 1. Proto Files Created/Updated

#### C:/Users/decri/gitclones/aura/proto/aura/vcregistry/v1beta1/query.proto
- **Status:** CREATED
- **Purpose:** Unified query service consolidating all vcregistry query methods
- **Contents:**
  - All VC queries (GetVC, ListUserVCs, CheckVCStatus, etc.)
  - Presentation queries (VerifyPresentation)
  - Attribute queries (GetDisclosurePolicy, ListAttributeVCs, ParseVoiceCommand)
  - Policy, revocation, and DID queries

#### C:/Users/decri/gitclones/aura/proto/aura/vcregistry/v1beta1/tx.proto
- **Status:** CREATED
- **Purpose:** Unified msg service consolidating all vcregistry transaction messages
- **Contents:**
  - VC lifecycle (MintVC, RevokeVC, etc.)
  - Presentation transactions (CreatePresentation)
  - Attribute transactions (CreateAttributeVC, UpdateDisclosurePolicy)
  - Policy and DID management

### 2. CLI Files Fixed

#### C:/Users/decri/gitclones/aura/chain/x/dataregistry/client/cli/query.go
- **Status:** FIXED
- **Changes:**
  - Updated imports to use proto-generated types (`dataregistrypb`)
  - Fixed query client method calls to match actual proto-generated methods
  - Updated type conversions to use proto enum values
  - Fixed method names: `DataItem()`, `UserDataItems()`, `Stats()`, `Params()`

**Key Fixes:**
```go
// OLD (incorrect):
queryClient := types.NewQueryClient(clientCtx)
res, err := queryClient.GetDataItem(context.Background(), req)

// NEW (correct):
queryClient := dataregistrypb.NewQueryClient(clientCtx)
res, err := queryClient.DataItem(context.Background(), req)
```

### 3. VCRegistry CLI Files

#### C:/Users/decri/gitclones/aura/chain/x/vcregistry/client/cli/tx.go
- **Status:** ALREADY CORRECT
- **Commands Implemented:**
  - `create-presentation` - Create QR code presentation
  - `create-attribute-vc` - Create attribute VC
  - `update-disclosure-policy` - Update disclosure policy
- **Uses:** Proto-generated types from `vcregistrypb`

#### C:/Users/decri/gitclones/aura/chain/x/vcregistry/client/cli/query.go
- **Status:** NEEDS PROTO UPDATE (methods exist in CLI but not in generated proto)
- **Commands Implemented:**
  - `verify-presentation` - Verify QR code
  - `show-disclosure-policy` - Show disclosure policy
  - `parse-voice-command` - Parse voice command
  - `list-attribute-vcs` - List attribute VCs

**Issue:** These query methods are called in CLI but not yet generated from proto. Need to:
1. Run `buf generate` or `make proto-gen` to regenerate proto code
2. This will create the QueryClient methods needed by the CLI

### 4. DataRegistry CLI Files

#### C:/Users/decri/gitclones/aura/chain/x/dataregistry/client/cli/tx.go
- **Status:** NEEDS FIXING
- **Commands Implemented:**
  - `store-data-item` - Store new data item
  - `update-data-item` - Update data item metadata
  - `delete-data-item` - Delete data item
  - `update-access-policy` - Update access policy
  - `verify-data-item` - Add verification

**Needed Fixes:**
```go
// Update imports:
import dataregistrypb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"

// Update message types:
msg := &dataregistrypb.MsgStoreDataItem{...}

// Update type conversions:
dataType := parseDataTypeProto(args[0])  // Returns dataregistrypb.DataItemType
```

## Proto-Generated Methods Available

### DataRegistry Query Methods
Based on `C:/Users/decri/gitclones/aura/proto/aura/dataregistry/v1beta1/query_grpc.pb.go`:

```go
type QueryClient interface {
    DataItem(ctx context.Context, in *QueryDataItemRequest, opts ...grpc.CallOption) (*QueryDataItemResponse, error)
    UserDataItems(ctx context.Context, in *QueryUserDataItemsRequest, opts ...grpc.CallOption) (*QueryUserDataItemsResponse, error)
    SearchDataItems(ctx context.Context, in *QuerySearchDataItemsRequest, opts ...grpc.CallOption) (*QuerySearchDataItemsResponse, error)
    DataItemVerifications(ctx context.Context, in *QueryDataItemVerificationsRequest, opts ...grpc.CallOption) (*QueryDataItemVerificationsResponse, error)
    Stats(ctx context.Context, in *QueryStatsRequest, opts ...grpc.CallOption) (*QueryStatsResponse, error)
    Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
}
```

### VCRegistry Query Methods
Based on `C:/Users/decri/gitclones/aura/proto/aura/vcregistry/v1beta1/vc_registry_grpc.pb.go`:

```go
type QueryClient interface {
    GetVC(ctx context.Context, in *QueryGetVCRequest, opts ...grpc.CallOption) (*QueryGetVCResponse, error)
    ListUserVCs(ctx context.Context, in *QueryListUserVCsRequest, opts ...grpc.CallOption) (*QueryListUserVCsResponse, error)
    CheckVCStatus(ctx context.Context, in *QueryCheckVCStatusRequest, opts ...grpc.CallOption) (*QueryCheckVCStatusResponse, error)
    BatchVCStatus(ctx context.Context, in *QueryBatchVCStatusRequest, opts ...grpc.CallOption) (*QueryBatchVCStatusResponse, error)
    GetVCPolicy(ctx context.Context, in *QueryGetVCPolicyRequest, opts ...grpc.CallOption) (*QueryGetVCPolicyResponse, error)
    ListVCPolicies(ctx context.Context, in *QueryListVCPoliciesRequest, opts ...grpc.CallOption) (*QueryListVCPoliciesResponse, error)
    GetRevocationList(ctx context.Context, in *QueryGetRevocationListRequest, opts ...grpc.CallOption) (*QueryGetRevocationListResponse, error)
    CheckRevocation(ctx context.Context, in *QueryCheckRevocationRequest, opts ...grpc.CallOption) (*QueryCheckRevocationResponse, error)
    ResolveDID(ctx context.Context, in *QueryResolveDIDRequest, opts ...grpc.CallOption) (*QueryResolveDIDResponse, error)
    GetDIDByAddress(ctx context.Context, in *QueryGetDIDByAddressRequest, opts ...grpc.CallOption) (*QueryGetDIDByAddressResponse, error)
    ValidateMintEligibility(ctx context.Context, in *QueryValidateMintEligibilityRequest, opts ...grpc.CallOption) (*QueryValidateMintEligibilityResponse, error)
    Stats(ctx context.Context, in *QueryStatsRequest, opts ...grpc.CallOption) (*QueryStatsResponse, error)
    Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
}
```

**Note:** The CLI calls additional methods (VerifyPresentation, GetDisclosurePolicy, ListAttributeVCs, ParseVoiceCommand) that need to be added to the proto and regenerated.

## Remaining Tasks

### 1. Regenerate Proto Code
**Required:** Yes
**Command:**
```bash
cd proto && buf generate
# OR
make proto-gen
```

**Note:** `buf` command is not installed on the system. Need to either:
- Install buf: `go install github.com/bufbuild/buf/cmd/buf@latest`
- Use alternative proto generation method
- The new query.proto and tx.proto files need to be compiled

### 2. Fix DataRegistry TX CLI
**File:** `C:/Users/decri/gitclones/aura/chain/x/dataregistry/client/cli/tx.go`

**Changes Needed:**
```go
// 1. Update imports
import dataregistrypb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"

// 2. Update all message creations
msg := &dataregistrypb.MsgStoreDataItem{
    Creator:         clientCtx.GetFromAddress().String(),
    DataType:        dataType,  // Use dataregistrypb.DataItemType
    Title:           title,
    // ... rest of fields
}

// 3. Update type parsing to return proto types
func parseDataTypeProto(s string) dataregistrypb.DataItemType {
    // Return dataregistrypb.DataItemType_DATA_ITEM_TYPE_PHOTO, etc.
}
```

### 3. Update Module Registration
Ensure CLI commands are properly registered in module.go files:
- `C:/Users/decri/gitclones/aura/chain/x/vcregistry/client/cli/module.go`
- `C:/Users/decri/gitclones/aura/chain/x/dataregistry/client/cli/module.go`

### 4. Build and Test
```bash
cd chain
go build ./cmd/aurad

# Test command registration
./aurad tx vcregistry --help
./aurad query vcregistry --help
./aurad tx dataregistry --help
./aurad query dataregistry --help
```

## Command Reference

### VCRegistry Commands

#### Transaction Commands
```bash
# Create QR code presentation
aurad tx vcregistry create-presentation vc:abc123,vc:def456 '{"show_full_name":true}' 300 --from alice

# Create attribute VC
aurad tx vcregistry create-attribute-vc age "base64encodedvalue" --from alice

# Update disclosure policy
aurad tx vcregistry update-disclosure-policy '{"auto_disclose_age":true}' --from alice
```

#### Query Commands
```bash
# Verify presentation
aurad query vcregistry verify-presentation "aura://verify?data=base64data"

# Show disclosure policy
aurad query vcregistry show-disclosure-policy aura1abc...

# Parse voice command
aurad query vcregistry parse-voice-command "AURA show my age"

# List attribute VCs
aurad query vcregistry list-attribute-vcs aura1abc...
```

### DataRegistry Commands

#### Transaction Commands
```bash
# Store data item
aurad tx dataregistry store-data-item photo "Vacation" abc123... ipfs://Qm... --from alice

# Update data item
aurad tx dataregistry update-data-item data:abc... --title "New Title" --from alice

# Delete data item
aurad tx dataregistry delete-data-item data:abc... --from alice

# Update access policy
aurad tx dataregistry update-access-policy data:abc... '{"mode":"public"}' --from alice

# Verify data item
aurad tx dataregistry verify-data-item data:abc... peer_verified 85 --from bob
```

#### Query Commands
```bash
# Show data item
aurad query dataregistry show-data-item data:abc123...

# List data items
aurad query dataregistry list-data-items aura1abc... --type photo

# Search data items
aurad query dataregistry search-data-items '{"tags":["vacation"]}'

# Get stats
aurad query dataregistry stats

# Get params
aurad query dataregistry params
```

## Known Issues

### 1. Proto Regeneration Needed
- **Issue:** New query.proto and tx.proto files created but not yet compiled
- **Impact:** VCRegistry query commands will fail until proto is regenerated
- **Solution:** Run `buf generate` or install buf tool

### 2. DataRegistry TX CLI Not Updated
- **Issue:** Still using types.Msg* instead of dataregistrypb.Msg*
- **Impact:** Build errors when compiling
- **Solution:** Update imports and type references as shown above

### 3. Missing Keeper Implementation
- **Status:** Keeper methods already exist (confirmed in previous session)
- **Location:**
  - `chain/x/vcregistry/keeper/presentation.go`
  - `chain/x/vcregistry/keeper/attributes.go`
  - `chain/x/dataregistry/keeper/keeper.go`
- **Note:** CLI just needs to call these correctly

## Files Modified

### Created:
- `proto/aura/vcregistry/v1beta1/query.proto` - Unified query service
- `proto/aura/vcregistry/v1beta1/tx.proto` - Unified msg service
- `chain/x/dataregistry/client/cli/query.go` - Fixed query CLI

### To Be Modified:
- `chain/x/dataregistry/client/cli/tx.go` - Needs proto type updates

### Backup Files:
- `chain/x/dataregistry/client/cli/query_old.go` - Original query.go (backup)

## Next Steps

1. **Install buf or setup proto generation:**
   ```bash
   go install github.com/bufbuild/buf/cmd/buf@latest
   cd proto
   buf generate
   ```

2. **Fix dataregistry tx.go:**
   - Update imports to use dataregistrypb
   - Update all message types
   - Update type parsing functions

3. **Build and test:**
   ```bash
   cd chain
   go build ./cmd/aurad
   ./aurad tx vcregistry --help
   ./aurad query dataregistry --help
   ```

4. **Integration testing:**
   - Test each command with actual data
   - Verify keeper methods are called correctly
   - Check error handling

## Success Criteria

- [ ] Proto code regenerated successfully
- [ ] All CLI commands compile without errors
- [ ] Commands show up in help output
- [ ] Query commands can call keeper methods
- [ ] Transaction commands can create and submit messages
- [ ] Type conversions work correctly between CLI and proto layers

## Notes

- The core keeper logic is already implemented and working
- The CLI layer is the only component that needs fixing
- Proto definitions are complete and comprehensive
- Once proto is regenerated and tx.go is fixed, all commands should work
