# Inclusion Routines Module

The Inclusion Routines module manages the definitions, prerequisites, and rate limits for Inclusion Routines (IRs) in the Aequitas Protocol. Inclusion Routines are verification tasks that users complete to build their identity confidence scores.

## Overview

This module provides:

- **IR Definition Management**: Create, update, delete, suspend, and activate IR definitions
- **Prerequisite Graph**: Define and enforce prerequisite relationships between IRs
- **Rate Limiting**: Configure and enforce rate limits on IR usage per wallet
- **Query Interface**: Query IR definitions, prerequisites, rate limits, and the dependency graph

## Key Concepts

### Inclusion Routine (IR)

An IR is a verification task that contributes to a user's identity confidence score. Each IR has:

- **ID**: Unique identifier (e.g., "IR-101")
- **Name**: Human-readable name
- **Arena**: Category of verification (Anchor, Biometric, Knowledge, Social, etc.)
- **Score**: Confidence points awarded upon completion
- **POI Reward**: Proof-of-Identity token rewards
- **Privacy Tier**: Sensitivity level (Low, Medium, High)
- **Locale Tags**: Geographic applicability
- **Status**: Lifecycle state (Draft, Active, Suspended, Retired, etc.)

### Prerequisites

IRs can require other IRs to be completed first. For example:
- IR-000 (Government ID Anchor) is a prerequisite for most other IRs
- Advanced biometric IRs may require basic liveness checks first

The module ensures:
- No circular dependencies in the prerequisite graph
- Prerequisites exist before being assigned
- Graph traversal for dependency resolution

### Rate Limits

Rate limits prevent abuse and ensure fair usage:

- **Per Wallet Per Hour**: Maximum attempts per wallet per hour
- **Per Wallet Per Day**: Maximum attempts per wallet per day
- **Per Block Global**: Maximum global attempts per block

The keeper automatically cleans up expired rate limit counters.

## State

The module maintains the following state:

- **IRs**: Map of IR ID to IR definition
- **Prerequisites**: Map of IR ID to prerequisite relationships
- **Rate Limits**: Map of IR ID to rate limit configuration
- **Rate Limit Usage**: Temporary counters for rate limit enforcement

## Messages

### MsgCreateIR

Creates a new IR definition. Requires governance authority.

```protobuf
message MsgCreateIR {
  string authority = 1;
  string id = 2;
  string name = 3;
  Arena arena = 4;
  string description = 5;
  int64 score = 6;
  int64 poi_reward = 7;
  repeated string locale_tags = 8;
  PrivacyTier privacy_tier = 9;
  string version = 10;
  string metadata_hash = 11;
  int64 activation_height = 12;
  int64 sunset_height = 13;
}
```

### MsgUpdateIR

Updates an existing IR definition. Requires governance authority.

### MsgDeleteIR

Deletes an IR definition. Requires governance authority. Fails if other IRs depend on it.

### MsgSetIRPrerequisites

Sets prerequisite relationships for an IR. Requires governance authority.

### MsgSetIRRateLimit

Configures rate limits for an IR. Requires governance authority.

### MsgSuspendIR

Suspends an IR, preventing new attempts. Requires governance authority.

### MsgActivateIR

Activates a suspended or approved IR. Requires governance authority.

## Queries

### IR

Query a single IR definition by ID.

```bash
aura query inclusionroutines ir IR-101
```

### ListIRs

Query a list of IRs with optional filters:

- Status filter (Active, Suspended, etc.)
- Arena filter (Biometric, Knowledge, etc.)
- Locale filter (e.g., "US", "global")
- Pagination

```bash
aura query inclusionroutines list-irs --status=active --arena=biometric
```

### IRGraph

Query the prerequisite dependency graph for an IR or the entire graph.

```bash
aura query inclusionroutines ir-graph IR-101
```

### RateLimit

Query the rate limit configuration for an IR.

```bash
aura query inclusionroutines rate-limit IR-101
```

### Params

Query module parameters.

```bash
aura query inclusionroutines params
```

## Parameters

- **max_ir_per_locale**: Maximum IRs allowed per locale (default: 50)
- **default_rate_limit_hour**: Default hourly rate limit (default: 10)
- **suspension_fee**: Fee required to suspend an IR (default: "1000000uaura")
- **min_governance_deposit**: Minimum deposit for governance proposals (default: "10000000uaura")

## Genesis

The module supports loading IR definitions from a JSON file at genesis:

```json
{
  "version": "1.0",
  "total_count": 181,
  "inclusion_routines": [
    {
      "id": "IR-000",
      "arena": "ANCHOR",
      "name": "Government ID Anchor",
      "description": "...",
      "score": 0,
      "privacy_tier": "high",
      "locale_tags": ["global"]
    }
  ]
}
```

Load from file:

```go
genesis, err := types.LoadGenesisFromFile("data/inclusion_routines/ir_definitions.json")
```

## Keeper Interface

The keeper provides the following key methods:

### IR CRUD
- `GetIR(id) (IRDefinition, bool)`
- `SetIR(ir) error`
- `DeleteIR(id) error`
- `CreateIR(ir) error`
- `UpdateIR(ir) error`
- `ListIRs(statusFilter, arenaFilter, localeFilter, offset, limit) ([]IRDefinition, int)`

### Status Management
- `SuspendIR(id) error`
- `ActivateIR(id) error`

### Prerequisites
- `GetPrerequisites(irID) (IRPrerequisite, bool)`
- `SetPrerequisites(irID, requiredIRIDs) error`
- `ValidatePrerequisites(irID, completedIRs) error`
- `GetIRGraph(irID) []IRGraphNode`

### Rate Limits
- `GetRateLimit(irID) (IRRateLimit, bool)`
- `SetRateLimit(limit) error`
- `CheckRateLimit(wallet, irID) error`
- `IncrementRateLimit(wallet, irID) error`
- `CleanupExpiredRateLimits()`

## Testing

Run the module tests:

```bash
go test ./chain/x/inclusionroutines/...
```

## Integration

To integrate the module into the application:

1. Create the keeper:
```go
import (
    "github.com/aequitas/aura/chain/x/inclusionroutines"
    "github.com/aequitas/aura/chain/x/inclusionroutines/keeper"
    "github.com/aequitas/aura/chain/x/inclusionroutines/params"
    "github.com/aequitas/aura/chain/x/inclusionroutines/types"
)

paramsStore := params.NewStore(types.DefaultParams())
irKeeper := keeper.NewKeeper(paramsStore)
```

2. Create the module:
```go
irModule := inclusionroutines.NewAppModule(irKeeper)
```

3. Register services:
```go
irModule.RegisterServices(moduleServices)
```

4. Initialize genesis:
```go
genesis, _ := types.LoadGenesisFromFile("data/inclusion_routines/ir_definitions.json")
irs := make([]types.IRDefinition, len(genesis.Irs))
for i, ir := range genesis.Irs {
    irs[i] = types.IRDefinitionFromProto(ir)
}
prereqs := make([]types.IRPrerequisite, len(genesis.Prerequisites))
for i, p := range genesis.Prerequisites {
    prereqs[i] = types.IRPrerequisiteFromProto(p)
}
limits := make([]types.IRRateLimit, len(genesis.RateLimits))
for i, l := range genesis.RateLimits {
    limits[i] = types.IRRateLimitFromProto(l)
}
irKeeper.InitGenesis(irs, prereqs, limits)
```

## Security Considerations

- All state-modifying operations require governance authority
- Rate limits prevent denial-of-service attacks
- Prerequisite validation prevents bypassing required verification steps
- Circular dependency detection prevents invalid graph states
- Comprehensive input validation prevents malformed data

## Events

### EventIRCreated
Emitted when inclusion routine is created.

**Attributes**: `ir_id`, `creator`, `name`

### EventIRUpdated
Emitted when inclusion routine is updated.

**Attributes**: `ir_id`, `updated_by`

### EventIRActivated
Emitted when inclusion routine is activated.

**Attributes**: `ir_id`, `activated_by`

### EventIRDeactivated
Emitted when inclusion routine is deactivated.

**Attributes**: `ir_id`, `reason`

### EventIRCompleted
Emitted when inclusion routine completes execution.

**Attributes**: `ir_id`, `completion_time`, `result`

### EventArenaAssigned
Emitted when arena is assigned to inclusion routine.

**Attributes**: `ir_id`, `arena_id`

## Future Enhancements

- Persistent storage backend (currently in-memory)
- Metrics and monitoring hooks
- Advanced querying (by score range, multiple filters)
- IR versioning and migration support
- Dynamic rate limit adjustment based on network conditions
