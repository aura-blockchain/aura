# Contract Registry Proto Architecture

## Message Hierarchy

```
ContractInfo (Root Message)
├── address: string
├── code_id: uint64
├── creator: string
├── admin: string
├── label: string
├── created_at: timestamp
├── updated_at: timestamp
├── metadata: ContractMetadata
│   ├── name: string
│   ├── description: string
│   ├── version: string
│   ├── homepage: string
│   ├── source_code_url: string
│   ├── tags: []string
│   ├── requires_vc: bool
│   ├── required_vc_types: []string
│   ├── min_confidence_score: uint64
│   ├── required_kyc_level: uint32
│   └── check_sanctions: bool
├── security_policy: SecurityPolicy
│   ├── allow_pause: bool
│   ├── allow_migration: bool
│   ├── max_gas_per_tx: uint64
│   ├── rate_limit_per_user: uint64
│   ├── blacklisted_addresses: []string
│   └── whitelisted_addresses: []string
├── compliance: ComplianceRequirements
│   ├── enforce_kyc: bool
│   ├── min_kyc_level: uint32
│   ├── enforce_sanctions_check: bool
│   ├── enforce_spending_limits: bool
│   ├── require_audit: bool
│   ├── last_audit_date: timestamp
│   └── audit_report_uri: string
└── status: ContractStatus (enum)
    ├── UNSPECIFIED (0)
    ├── ACTIVE (1)
    ├── PAUSED (2)
    ├── DEPRECATED (3)
    └── FROZEN (4)
```

## Service Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Contract Registry Module                  │
└─────────────────────────────────────────────────────────────┘
                            │
            ┌───────────────┴───────────────┐
            │                               │
            ▼                               ▼
    ┌──────────────┐               ┌──────────────┐
    │ Query Service│               │  Msg Service │
    └──────────────┘               └──────────────┘
            │                               │
    ┌───────┴────────┐              ┌──────┴──────────┐
    │                │              │                 │
    ▼                ▼              ▼                 ▼
ContractInfo    ContractsByCreator  RegisterContract  UpdateMetadata
ContractsByTag  RegisteredContracts UpdateSecurity    PauseContract
ContractMetrics                     UnpauseContract   DeprecateContract
```

## Data Flow

### Registration Flow
```
User/Admin
    │
    ├─> MsgRegisterContract
    │       ├─> ContractMetadata (identity requirements)
    │       ├─> SecurityPolicy (access controls)
    │       └─> ComplianceRequirements (regulatory)
    │
    ├─> Validation
    │       ├─> Check metadata completeness
    │       ├─> Validate security parameters
    │       └─> Verify compliance config
    │
    ├─> Store ContractInfo
    │       ├─> KV Store (contract address -> ContractInfo)
    │       ├─> Index by Creator
    │       └─> Index by Tags
    │
    └─> Initialize ContractMetrics
            ├─> Set counters to zero
            └─> Record creation timestamp
```

### Query Flow
```
User/App
    │
    ├─> QueryContractInfo(address)
    │       └─> Retrieve from KV Store
    │
    ├─> QueryContractsByCreator(creator, pagination)
    │       └─> Scan Creator Index
    │
    ├─> QueryContractsByTag(tag, pagination)
    │       └─> Scan Tag Index
    │
    ├─> QueryRegisteredContracts(status, pagination)
    │       └─> Scan All Contracts (filtered)
    │
    └─> QueryContractMetrics(address)
            └─> Retrieve Metrics from KV Store
```

### Update Flow
```
Admin/Governance
    │
    ├─> MsgUpdateContractMetadata
    │       ├─> Verify signer is admin
    │       ├─> Validate new metadata
    │       ├─> Update ContractInfo
    │       └─> Update updated_at timestamp
    │
    ├─> MsgUpdateSecurityPolicy
    │       ├─> Verify signer is admin
    │       ├─> Validate policy parameters
    │       └─> Update SecurityPolicy
    │
    ├─> MsgPauseContract
    │       ├─> Verify signer authority
    │       ├─> Set status = PAUSED
    │       └─> Emit pause event
    │
    └─> MsgDeprecateContract
            ├─> Verify signer authority
            ├─> Set status = DEPRECATED
            ├─> Record migration target
            └─> Emit deprecation event
```

## Integration Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              Contract Registry Module (Core)                 │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐            │
│  │ContractInfo│  │  Security  │  │ Compliance │            │
│  │  Storage   │  │   Policy   │  │Requirements│            │
│  └────────────┘  └────────────┘  └────────────┘            │
└─────────────────────────────────────────────────────────────┘
         │              │              │              │
    ┌────┴────┐    ┌───┴───┐    ┌────┴────┐    ┌────┴────┐
    │         │    │       │    │         │    │         │
    ▼         ▼    ▼       ▼    ▼         ▼    ▼         ▼
┌─────────┐ ┌────┐ ┌─────┐ ┌───┐ ┌─────┐ ┌───┐ ┌─────┐ ┌────┐
│VCRegistry│ │Auth│ │ Conf│ │Eco│ │Comp │ │Mon│ │Data │ │Inc │
│         │ │    │ │Score│ │Sec│ │lianc│ │   │ │Reg  │ │Rout│
└─────────┘ └────┘ └─────┘ └───┘ └─────┘ └───┘ └─────┘ └────┘

VCRegistry:       VC verification, required types
Auth:             Role-based access control, admin checks
ConfidenceScore:  Minimum score requirements
EconomicSecurity: Spending limits enforcement
Compliance:       KYC levels, sanctions screening
Monitoring:       Metrics collection, alerting
DataRegistry:     Audit document storage
InclusionRoutines: Integration with IR system
```

## Storage Layout

```
KV Store Keys:
├─> 0x01 | contractAddress -> ContractInfo
├─> 0x02 | creatorAddress | contractAddress -> 0x01
├─> 0x03 | tag | contractAddress -> 0x01
├─> 0x04 | contractAddress -> ContractMetrics
└─> 0x05 | status | contractAddress -> 0x01

Indices:
├─> Creator Index: Fast lookup by creator
├─> Tag Index: Fast lookup by tag
└─> Status Index: Fast lookup by status
```

## Status State Machine

```
                    RegisterContract
                          │
                          ▼
                    ┌──────────┐
                    │  ACTIVE  │◄───┐
                    └──────────┘    │
                      │      ▲      │
            PauseContract    │      │
                      │      │      │
                      ▼      │      │
                    ┌──────────┐    │
                    │  PAUSED  │    │
                    └──────────┘    │
                          │         │
                    UnpauseContract │
                          │         │
              ┌───────────┴─────────┘
              │
    DeprecateContract
              │
              ▼
        ┌──────────┐
        │DEPRECATED│
        └──────────┘
              │
        (Admin/Governance)
              │
              ▼
         ┌─────────┐
         │ FROZEN  │
         └─────────┘
```

## Security Enforcement Flow

```
Contract Execution Request
    │
    ├─> Query ContractInfo
    │
    ├─> Check Status
    │   ├─> ACTIVE: Continue
    │   ├─> PAUSED: Reject
    │   ├─> FROZEN: Reject
    │   └─> DEPRECATED: Warn & Continue
    │
    ├─> Check SecurityPolicy
    │   ├─> Whitelist: Check if user in whitelist
    │   ├─> Blacklist: Check if user NOT in blacklist
    │   ├─> Rate Limit: Check user rate limit
    │   └─> Gas Limit: Check max gas
    │
    ├─> Check ComplianceRequirements
    │   ├─> KYC: Query Compliance Module
    │   ├─> Sanctions: Query Compliance Module
    │   └─> Spending Limits: Query EconomicSecurity Module
    │
    ├─> Check Metadata Requirements
    │   ├─> VCs: Query VCRegistry Module
    │   └─> Confidence Score: Query ConfidenceScore Module
    │
    └─> Execution Allowed / Denied
            ├─> Update ContractMetrics
            └─> Emit events
```

## Module Parameters

```
ContractRegistryParams
├─> Registration Control
│   ├─> open_registration: bool
│   └─> max_contracts_per_creator: uint64
│
├─> Validation Requirements
│   ├─> require_metadata: bool
│   ├─> require_security_policy: bool
│   └─> require_compliance_config: bool
│
├─> Audit Management
│   └─> audit_warning_days: uint64
│
└─> Default Limits
    ├─> default_rate_limit: uint64
    └─> default_max_gas: uint64
```

## Genesis State

```
GenesisState
├─> params: ContractRegistryParams
├─> contracts: []ContractInfo
│   ├─> All registered contracts
│   └─> With complete metadata & policies
└─> metrics: []ContractMetrics
    └─> All contract execution metrics
```

---

**Architecture Design Principles:**
1. Security-first approach
2. Modular integration with existing AURA modules
3. Comprehensive compliance enforcement
4. Flexible identity-based access control
5. Detailed metrics and monitoring
6. Admin control with governance override
7. Clear status management
8. Efficient storage and indexing
