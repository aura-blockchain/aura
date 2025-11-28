# Contract Registry Proto - Quick Reference

## Message Types Summary

### Core Types (contract_registry.proto)

| Message Type | Fields | Purpose |
|-------------|--------|---------|
| `ContractInfo` | 11 fields | Complete contract registration information |
| `ContractMetadata` | 11 fields | Identity requirements and contract details |
| `SecurityPolicy` | 6 fields | Security controls and access restrictions |
| `ComplianceRequirements` | 7 fields | Regulatory and audit requirements |
| `ContractMetrics` | 10 fields | Usage statistics and monitoring data |
| `ContractStatus` (enum) | 5 values | Contract operational state |

### Query Types (query.proto)

| RPC Method | Request | Response | HTTP Endpoint |
|-----------|---------|----------|---------------|
| `ContractInfo` | contract_address | ContractInfo | GET /contracts/{address} |
| `ContractsByCreator` | creator_address, pagination | []ContractInfo | GET /contracts/creator/{address} |
| `ContractsByTag` | tag, pagination | []ContractInfo | GET /contracts/tag/{tag} |
| `RegisteredContracts` | status, pagination | []ContractInfo | GET /contracts |
| `ContractMetrics` | contract_address | ContractMetrics | GET /metrics/{address} |

### Message Types (msg.proto)

| RPC Method | Signer Required | Purpose |
|-----------|----------------|---------|
| `RegisterContract` | Admin/Creator | Register new contract with full metadata |
| `UpdateContractMetadata` | Admin | Update contract metadata |
| `UpdateSecurityPolicy` | Admin | Update security policy |
| `PauseContract` | Admin/Governance | Pause contract execution |
| `UnpauseContract` | Admin/Governance | Resume contract execution |
| `DeprecateContract` | Admin/Governance | Mark contract as deprecated |

### Genesis Types (genesis.proto)

| Message Type | Fields | Purpose |
|-------------|--------|---------|
| `GenesisState` | params, contracts[], metrics[] | Module genesis state |
| `ContractRegistryParams` | 8 fields | Module configuration parameters |

## Field Details

### ContractInfo Fields
```
address              string
code_id              uint64
creator              string
admin                string
label                string
created_at           timestamp
updated_at           timestamp
metadata             ContractMetadata
security_policy      SecurityPolicy
compliance           ComplianceRequirements
status               ContractStatus
```

### ContractMetadata Fields
```
name                 string
description          string
version              string
homepage             string
source_code_url      string
tags                 []string
requires_vc          bool
required_vc_types    []string
min_confidence_score uint64
required_kyc_level   uint32
check_sanctions      bool
```

### SecurityPolicy Fields
```
allow_pause              bool
allow_migration          bool
max_gas_per_tx           uint64
rate_limit_per_user      uint64
blacklisted_addresses    []string
whitelisted_addresses    []string
```

### ComplianceRequirements Fields
```
enforce_kyc              bool
min_kyc_level            uint32
enforce_sanctions_check  bool
enforce_spending_limits  bool
require_audit            bool
last_audit_date          timestamp
audit_report_uri         string
```

### ContractMetrics Fields
```
contract_address         string
total_executions         uint64
successful_executions    uint64
failed_executions        uint64
total_gas_used           uint64
avg_gas_per_execution    uint64
unique_users             uint64
rate_limit_violations    uint64
compliance_failures      uint64
last_execution           timestamp
```

## ContractStatus Enum Values

| Value | Name | Description |
|-------|------|-------------|
| 0 | CONTRACT_STATUS_UNSPECIFIED | Invalid/uninitialized |
| 1 | CONTRACT_STATUS_ACTIVE | Normal operation |
| 2 | CONTRACT_STATUS_PAUSED | Temporarily suspended |
| 3 | CONTRACT_STATUS_DEPRECATED | Discouraged use, migration recommended |
| 4 | CONTRACT_STATUS_FROZEN | Emergency stop |

## Module Parameters

### ContractRegistryParams
```
open_registration          bool    - Allow anyone to register contracts
max_contracts_per_creator  uint64  - Limit per creator (0 = unlimited)
require_metadata           bool    - Metadata must be provided
require_security_policy    bool    - Security policy must be specified
require_compliance_config  bool    - Compliance config must be specified
audit_warning_days         uint64  - Days before audit warning
default_rate_limit         uint64  - Default rate limit per user/hour
default_max_gas            uint64  - Default max gas per transaction
```

## Integration Examples

### Query Contract Information
```bash
# Get contract info
aura query contractregistry contract-info [contract_address]

# Get contracts by creator
aura query contractregistry contracts-by-creator [creator_address]

# Get contracts by tag
aura query contractregistry contracts-by-tag "defi"

# Get all contracts
aura query contractregistry registered-contracts

# Get contract metrics
aura query contractregistry contract-metrics [contract_address]
```

### Register Contract
```bash
aura tx contractregistry register-contract \
  --contract-address [address] \
  --code-id [code_id] \
  --creator [creator] \
  --admin [admin] \
  --label "My Contract" \
  --metadata [metadata_json] \
  --security-policy [policy_json] \
  --compliance [compliance_json] \
  --from [signer]
```

### Update Operations
```bash
# Update metadata
aura tx contractregistry update-metadata [contract_address] [metadata_json]

# Update security policy
aura tx contractregistry update-security-policy [contract_address] [policy_json]

# Pause contract
aura tx contractregistry pause-contract [contract_address] "Security issue"

# Unpause contract
aura tx contractregistry unpause-contract [contract_address]

# Deprecate contract
aura tx contractregistry deprecate-contract [contract_address] "Replaced by v2" [new_contract]
```

## gRPC Endpoints

### Query Service
- `aura.contractregistry.v1beta1.Query/ContractInfo`
- `aura.contractregistry.v1beta1.Query/ContractsByCreator`
- `aura.contractregistry.v1beta1.Query/ContractsByTag`
- `aura.contractregistry.v1beta1.Query/RegisteredContracts`
- `aura.contractregistry.v1beta1.Query/ContractMetrics`

### Msg Service
- `aura.contractregistry.v1beta1.Msg/RegisterContract`
- `aura.contractregistry.v1beta1.Msg/UpdateContractMetadata`
- `aura.contractregistry.v1beta1.Msg/UpdateSecurityPolicy`
- `aura.contractregistry.v1beta1.Msg/PauseContract`
- `aura.contractregistry.v1beta1.Msg/UnpauseContract`
- `aura.contractregistry.v1beta1.Msg/DeprecateContract`

## REST API Endpoints

- GET `/aura/contractregistry/v1beta1/contracts/{contract_address}`
- GET `/aura/contractregistry/v1beta1/contracts/creator/{creator_address}`
- GET `/aura/contractregistry/v1beta1/contracts/tag/{tag}`
- GET `/aura/contractregistry/v1beta1/contracts`
- GET `/aura/contractregistry/v1beta1/metrics/{contract_address}`

## File Locations

| File | Location | Size |
|------|----------|------|
| contract_registry.proto | proto/aura/contractregistry/v1beta1/ | 175 lines |
| query.proto | proto/aura/contractregistry/v1beta1/ | 115 lines |
| msg.proto | proto/aura/contractregistry/v1beta1/ | 172 lines |
| genesis.proto | proto/aura/contractregistry/v1beta1/ | 46 lines |
| contract_registry.pb.go | proto/aura/contractregistry/v1beta1/ | 29 KB |
| query.pb.go | proto/aura/contractregistry/v1beta1/ | 27 KB |
| msg.pb.go | proto/aura/contractregistry/v1beta1/ | 32 KB |
| genesis.pb.go | proto/aura/contractregistry/v1beta1/ | 11 KB |
| query_grpc.pb.go | proto/aura/contractregistry/v1beta1/ | 12 KB |
| msg_grpc.pb.go | proto/aura/contractregistry/v1beta1/ | 14 KB |

**Total:** 508 proto lines, ~125 KB generated code
