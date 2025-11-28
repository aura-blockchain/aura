# Contract Registry - Code Samples and Examples

## CLI Query Client - Before and After

### BEFORE (Incorrect)
```go
// This would fail because types doesn't have NewQueryClient
queryClient := types.NewQueryClient(clientCtx)

res, err := queryClient.ContractInfo(
    context.Background(),
    &types.QueryContractInfoRequest{
        ContractAddress: args[0],
    },
)
```

### AFTER (Correct)
```go
// Import the protobuf package
pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"

// Use protobuf generated client
queryClient := pb.NewQueryClient(clientCtx)

res, err := queryClient.ContractInfo(
    context.Background(),
    &pb.QueryContractInfoRequest{
        ContractAddress: args[0],
    },
)
```

## Keeper Usage - Type Definitions

### BEFORE (Compilation Error)
```go
// In keeper/audit_trail.go
// Error: undefined: types.AuditEntry
func (k Keeper) AddAuditEntry(ctx sdk.Context, entry *types.AuditEntry) {
    // ... implementation
}
```

### AFTER (With Proto Definition)
```proto
// In proto/aura/contractregistry/v1beta1/contract_registry.proto
message AuditEntry {
  uint64 id = 1;
  string contract_address = 2;
  int64 timestamp = 3;
  string action = 4;
  string actor = 5;
  string details = 6;
  bool success = 7;
}
```

```go
// Now in types/types.go
type AuditEntry = pb.AuditEntry

// And in keeper/audit_trail.go
func (k Keeper) AddAuditEntry(ctx sdk.Context, entry *types.AuditEntry) {
    // ... implementation - compiles successfully
}
```

## Migration Record Example

### Proto Definition
```proto
message MigrationRecord {
  uint64 id = 1;
  string old_contract_address = 2;
  string new_contract_address = 3;
  uint64 old_code_id = 4;
  uint64 new_code_id = 5;
  google.protobuf.Timestamp migrated_at = 6;
  string migrated_by = 7;
  string reason = 8;
}
```

### Usage in Keeper
```go
// In keeper/migration.go
func (k Keeper) RecordMigration(ctx sdk.Context, oldContractAddr, newContractAddr, admin, reason string) error {
    // Create migration record
    migration := &types.MigrationRecord{
        Id:                 k.GetNextMigrationID(ctx),
        OldContractAddress: oldContractAddr,
        NewContractAddress: newContractAddr,
        MigratedAt:         timestamppb.New(ctx.BlockTime()),
        MigratedBy:         admin,
        Reason:             reason,
        OldCodeId:          oldInfo.CodeId,
        NewCodeId:          newInfo.CodeId,
    }

    // Store migration record
    k.SetMigrationRecord(ctx, migration)

    return nil
}
```

## ContractInfo Enhancement

### BEFORE (Missing Fields)
```proto
message ContractInfo {
  string address = 1;
  uint64 code_id = 2;
  string creator = 3;
  string admin = 4;
  string label = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
  ContractMetadata metadata = 8;
  SecurityPolicy security_policy = 9;
  ComplianceRequirements compliance = 10;
  ContractStatus status = 11;
}
```

### AFTER (With Migration Tracking)
```proto
message ContractInfo {
  string address = 1;
  uint64 code_id = 2;
  string creator = 3;
  string admin = 4;
  string label = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
  ContractMetadata metadata = 8;
  SecurityPolicy security_policy = 9;
  ComplianceRequirements compliance = 10;
  ContractStatus status = 11;

  // NEW: Migration tracking fields
  string migration_target = 12;
  string migrated_from = 13;
  google.protobuf.Timestamp migrated_at = 14;
}
```

### Usage in Keeper
```go
// In keeper/migration.go
func (k Keeper) RecordMigration(ctx sdk.Context, oldContractAddr, newContractAddr, admin, reason string) error {
    // Get contract info
    oldInfo, found := k.GetContractInfo(ctx, oldContractAddr)
    if !found {
        return types.ErrContractNotFound
    }

    newInfo, found := k.GetContractInfo(ctx, newContractAddr)
    if !found {
        return types.ErrContractNotFound
    }

    // Create migration record
    migration := &types.MigrationRecord{
        Id:                 k.GetNextMigrationID(ctx),
        OldContractAddress: oldContractAddr,
        NewContractAddress: newContractAddr,
        MigratedAt:         timestamppb.New(ctx.BlockTime()),
        MigratedBy:         admin,
        Reason:             reason,
        OldCodeId:          oldInfo.CodeId,
        NewCodeId:          newInfo.CodeId,
    }

    // Store migration record
    k.SetMigrationRecord(ctx, migration)

    // Update old contract with migration info (NOW COMPILES!)
    oldInfo.MigrationTarget = newContractAddr
    oldInfo.MigratedAt = migration.MigratedAt
    k.SetContractInfo(ctx, oldInfo)

    // Update new contract with migration source (NOW COMPILES!)
    newInfo.MigratedFrom = oldContractAddr
    k.SetContractInfo(ctx, newInfo)

    return nil
}
```

## Query Types Example

### Proto Definition
```proto
service Query {
  rpc ContractAudits(QueryContractAuditsRequest) returns (QueryContractAuditsResponse) {
    option (google.api.http).get = "/aura/contractregistry/v1beta1/audits/{contract_address}";
  }
}

message QueryContractAuditsRequest {
  string contract_address = 1;
  uint64 limit = 2;
}

message QueryContractAuditsResponse {
  repeated AuditEntry entries = 1;
  AuditStatistics statistics = 2;
}
```

### CLI Command Implementation
```go
// In client/cli/query.go
func GetCmdQueryContractAudits() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "audits [contract-address]",
        Short: "Query audit reports for a contract",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            clientCtx, err := client.GetClientQueryContext(cmd)
            if err != nil {
                return err
            }

            queryClient := pb.NewQueryClient(clientCtx)

            res, err := queryClient.ContractAudits(
                context.Background(),
                &pb.QueryContractAuditsRequest{
                    ContractAddress: args[0],
                },
            )
            if err != nil {
                return err
            }

            return clientCtx.PrintProto(res)
        },
    }

    flags.AddQueryFlagsToCmd(cmd)
    return cmd
}
```

## Module Parameters

### Proto Definition
```proto
message ContractRegistryParams {
  bool open_registration = 1;
  uint32 max_contracts_per_creator = 2;
  bool require_metadata = 3;
  bool require_security_policy = 4;
  bool require_compliance_config = 5;
  uint32 audit_warning_days = 6;
  uint32 default_rate_limit = 7;
  uint64 default_max_gas = 8;
}
```

### Default Parameters in Go
```go
// In types/types.go
func DefaultParams() ContractRegistryParams {
    return ContractRegistryParams{
        OpenRegistration:        true,
        MaxContractsPerCreator:  100,
        RequireMetadata:         true,
        RequireSecurityPolicy:   true,
        RequireComplianceConfig: false,
        AuditWarningDays:        180,
        DefaultRateLimit:        100,
        DefaultMaxGas:           5000000,
    }
}
```

## Type Aliases in types Package

### Complete Type Export Pattern
```go
// In types/types.go
package types

import pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"

type (
    // Core types
    ContractInfo            = pb.ContractInfo
    ContractMetadata        = pb.ContractMetadata
    SecurityPolicy          = pb.SecurityPolicy
    ComplianceRequirements  = pb.ComplianceRequirements
    ContractStatus          = pb.ContractStatus
    ContractMetrics         = pb.ContractMetrics
    ContractRegistryParams  = pb.ContractRegistryParams
    AuditEntry              = pb.AuditEntry
    AuditStatistics         = pb.AuditStatistics
    MigrationRecord         = pb.MigrationRecord

    // Query types...
    QueryContractAuditsRequest           = pb.QueryContractAuditsRequest
    QueryContractAuditsResponse          = pb.QueryContractAuditsResponse
    // ... more query types
)
```

## Audit Trail Usage Example

### Recording an Audit Entry
```go
// In keeper/audit_trail.go
func (k Keeper) RecordContractExecution(ctx sdk.Context, contractAddr, executor string, gasUsed uint64, success bool, errorMsg string) {
    details := fmt.Sprintf("Gas used: %d", gasUsed)
    if !success && errorMsg != "" {
        details = fmt.Sprintf("%s, Error: %s", details, errorMsg)
    }

    entry := &types.AuditEntry{
        ContractAddress: contractAddr,
        Timestamp:       ctx.BlockTime().Unix(),
        Action:          "EXECUTE_CONTRACT",
        Actor:           executor,
        Details:         details,
        Success:         success,
    }

    k.AddAuditEntry(ctx, entry)
    k.UpdateMetricsOnExecution(ctx, contractAddr, gasUsed, success)
}
```

### Querying Audit Trail
```go
// In client/cli/query.go
func GetCmdQueryContractAudits() *cobra.Command {
    return &cobra.Command{
        Use:   "audits [contract-address]",
        Short: "Query audit reports for a contract",
        RunE: func(cmd *cobra.Command, args []string) error {
            clientCtx, err := client.GetClientQueryContext(cmd)
            if err != nil {
                return err
            }

            queryClient := pb.NewQueryClient(clientCtx)
            res, err := queryClient.ContractAudits(
                context.Background(),
                &pb.QueryContractAuditsRequest{
                    ContractAddress: args[0],
                },
            )
            if err != nil {
                return err
            }

            // Display results
            return clientCtx.PrintProto(res)
        },
    }
}
```

## Summary of Patterns

1. **Type Definitions**: Always define in proto files, re-export via Go type aliases
2. **Query Clients**: Use `pb.NewQueryClient()` from protobuf package, never from `types`
3. **Request/Response Types**: Use `pb.*Request` and `pb.*Response` from protobuf
4. **Proto Field Numbers**: Use correct numbers for backward compatibility
5. **Go Imports**: Import protobuf package as `pb` for clarity
6. **Type Safety**: Leverage protobuf generated code for strong typing
