# AURA Blockchain Module Consolidation Plan

**Status:** In Progress
**Date:** 2025-11-27
**Author:** Architecture Team

---

## Executive Summary

This document outlines the consolidation of AURA's 24 custom modules into 8 focused modules. This restructuring reduces maintenance burden, improves auditability, and presents a professional, focused architecture to the blockchain community.

---

## Before/After Overview

### Before (24 Modules)
```
chain/x/
├── aiassistant/          # REMOVE - off-chain service
├── aura-bindings/        # MERGE → wasm
├── auth/                 # MERGE → identity
├── bridge/               # KEEP
├── common/               # KEEP as utility package (not a module)
├── compliance/           # KEEP (evaluate off-chain later)
├── confidencescore/      # KEEP
├── contractregistry/     # MERGE → wasm
├── cryptography/         # MERGE → security
├── dataregistry/         # MERGE → vcregistry
├── dex/                  # KEEP
├── economicsecurity/     # MERGE → economics
├── governance/           # MERGE → economics
├── identitychange/       # MERGE → identity
├── incidentresponse/     # MERGE → security
├── inclusionroutines/    # KEEP
├── monitoring/           # REMOVE - off-chain infrastructure
├── networksecurity/      # MERGE → security
├── prevalidation/        # REMOVE - move to ante handlers
├── privacy/              # MERGE → security
├── validatorsecurity/    # MERGE → security
├── vcregistry/           # KEEP
├── walletsecurity/       # MERGE → security
└── wasm/                 # KEEP (absorbs contractregistry, aura-bindings)
```

### After (8 Modules)
```
chain/x/
├── bridge/               # Cross-chain bridge functionality
├── common/               # Shared utilities (not a Cosmos module)
├── compliance/           # KYC/AML/regulatory (evaluate off-chain later)
├── confidencescore/      # Reputation and trust scoring
├── dex/                  # Decentralized exchange
├── economics/            # NEW: Token economics + governance
├── identity/             # NEW: Auth + identity management
├── inclusionroutines/    # Inclusion routine execution
├── security/             # NEW: All security functionality
├── vcregistry/           # Verifiable credentials + data registry
└── wasm/                 # Smart contracts + bindings + registry
```

---

## Detailed Consolidation Mapping

### 1. security Module (NEW)

**Merges:** networksecurity, validatorsecurity, walletsecurity, incidentresponse, cryptography, privacy

**Store Keys:**
- `security` (primary)
- Legacy prefixes maintained for migration

**Subpackages:**
```
chain/x/security/
├── keeper/
│   ├── keeper.go              # Main keeper
│   ├── network.go             # Rate limiting, Sybil, DDoS protection
│   ├── validator.go           # Jailing, slashing, sentry nodes
│   ├── wallet.go              # Multisig, social recovery
│   ├── incident.go            # Circuit breakers, incident response
│   ├── crypto.go              # Key rotation, encryption
│   ├── privacy.go             # Ring signatures, mixing
│   ├── msg_server.go
│   └── query_server.go
├── types/
│   ├── keys.go
│   ├── errors.go
│   ├── network_types.go
│   ├── validator_types.go
│   ├── wallet_types.go
│   ├── incident_types.go
│   ├── crypto_types.go
│   └── privacy_types.go
└── module.go
```

**Genesis State:**
```json
{
  "security": {
    "params": {},
    "network": { "rate_limits": [], "reputations": [], "trusted_peers": [] },
    "validator": { "infractions": [], "sentry_nodes": [], "alerts": [] },
    "wallet": { "multisigs": [], "recovery_configs": [], "sessions": [] },
    "incident": { "incidents": [], "pause_state": {} },
    "crypto": { "key_rotations": [], "threshold_schemes": [] },
    "privacy": { "mixing_pools": [], "view_keys": [] }
  }
}
```

---

### 2. identity Module (NEW)

**Merges:** auth, identitychange

**Store Keys:**
- `identity` (primary)

**Subpackages:**
```
chain/x/identity/
├── keeper/
│   ├── keeper.go              # Main keeper
│   ├── auth.go                # Role-based access, permissions
│   ├── accounts.go            # Account management
│   ├── changes.go             # Identity change requests
│   ├── audit.go               # Audit trail
│   ├── msg_server.go
│   └── query_server.go
├── types/
│   ├── keys.go
│   ├── errors.go
│   ├── auth_types.go
│   └── change_types.go
└── module.go
```

**Genesis State:**
```json
{
  "identity": {
    "params": {},
    "roles": [],
    "permissions": [],
    "change_requests": [],
    "audit_logs": []
  }
}
```

---

### 3. economics Module (NEW)

**Merges:** economicsecurity, governance

**Store Keys:**
- `economics` (primary)

**Subpackages:**
```
chain/x/economics/
├── keeper/
│   ├── keeper.go              # Main keeper
│   ├── fees.go                # Dynamic fees
│   ├── mev.go                 # MEV protection
│   ├── vesting.go             # Vesting schedules
│   ├── treasury.go            # Treasury management
│   ├── whale.go               # Whale protection
│   ├── governance.go          # Proposals, voting
│   ├── msg_server.go
│   └── query_server.go
├── types/
│   ├── keys.go
│   ├── errors.go
│   ├── fee_types.go
│   ├── vesting_types.go
│   └── governance_types.go
└── module.go
```

**Genesis State:**
```json
{
  "economics": {
    "params": {},
    "vesting_schedules": [],
    "treasury": {},
    "proposals": [],
    "votes": [],
    "deposits": []
  }
}
```

---

### 4. wasm Module (EXPANDED)

**Absorbs:** contractregistry, aura-bindings

**Store Keys:**
- `wasm` (primary)

**Subpackages:**
```
chain/x/wasm/
├── keeper/
│   ├── keeper.go              # Main keeper
│   ├── contracts.go           # Contract deployment
│   ├── registry.go            # Contract registry (from contractregistry)
│   ├── bindings.go            # AURA bindings (from aura-bindings)
│   ├── msg_server.go
│   └── query_server.go
├── bindings/                   # Custom message/query bindings
│   ├── msg.go
│   ├── query.go
│   └── types.go
├── types/
└── module.go
```

---

### 5. vcregistry Module (EXPANDED)

**Absorbs:** dataregistry

**Store Keys:**
- `vcregistry` (primary)

**Subpackages:**
```
chain/x/vcregistry/
├── keeper/
│   ├── keeper.go              # Main keeper
│   ├── credentials.go         # VC management
│   ├── presentations.go       # VP management
│   ├── did.go                 # DID resolution
│   ├── data.go                # Data registry (from dataregistry)
│   ├── msg_server.go
│   └── query_server.go
├── types/
└── module.go
```

---

### 6-8. Unchanged Modules

These modules remain as-is:
- **bridge/** - Cross-chain functionality
- **confidencescore/** - Reputation scoring
- **dex/** - Decentralized exchange
- **inclusionroutines/** - IR execution
- **compliance/** - Regulatory compliance (evaluate off-chain migration later)

---

## Modules to Remove

### aiassistant
**Reason:** AI functionality does not require consensus. Run as off-chain service.
**Migration:** Create separate `aura-ai-service` repository.

### monitoring
**Reason:** Monitoring is infrastructure, not consensus. Use Prometheus/Grafana.
**Migration:** Export metrics via standard Cosmos SDK telemetry.

### prevalidation
**Reason:** Pre-validation can be handled in ante handlers.
**Migration:** Move logic to `app/ante.go`.

---

## Migration Strategy

### Phase 1: Create New Module Structures
1. Create `security/`, `identity/`, `economics/` directories
2. Set up keeper scaffolding with subpackage imports
3. Define combined types and genesis structures

### Phase 2: Migrate Keeper Logic
1. Move keeper methods to appropriate subpackages
2. Update import paths
3. Maintain backward-compatible store key prefixes

### Phase 3: Update App Integration
1. Update `app/app.go` module registration
2. Update `app/module_manager.go`
3. Update genesis template in `cmd/aurad/cmd/init.go`

### Phase 4: Remove Old Modules
1. Delete old module directories
2. Update proto definitions
3. Clean up unused dependencies

### Phase 5: Documentation
1. Update architecture docs
2. Update README files
3. Create migration guide for existing deployments

---

## Store Key Migration

To maintain backward compatibility with existing chain state:

```go
// security/types/keys.go
const (
    ModuleName = "security"
    StoreKey   = "security"

    // Legacy prefixes for migration
    LegacyNetworkSecurityPrefix   = "networksecurity"
    LegacyValidatorSecurityPrefix = "validatorsecurity"
    LegacyWalletSecurityPrefix    = "walletsecurity"
    LegacyIncidentResponsePrefix  = "incidentresponse"
    LegacyCryptographyPrefix      = "cryptography"
    LegacyPrivacyPrefix           = "privacy"
)
```

---

## Testing Requirements

1. **Unit Tests:** All migrated keeper methods must have test coverage
2. **Integration Tests:** Module interactions must be tested
3. **Genesis Tests:** Import/export round-trip must work
4. **Migration Tests:** Upgrade from old structure must preserve state

---

## Rollback Plan

If consolidation causes issues:
1. Git revert to pre-consolidation commit
2. Old module structure is preserved in git history
3. No on-chain migration needed (this is pre-launch)

---

## Success Criteria

- [ ] 8 modules compile successfully
- [ ] All existing tests pass
- [ ] Node initializes and produces blocks
- [ ] Genesis export/import works correctly
- [ ] Documentation is complete and accurate

---

## Appendix: Module Dependency Graph (After)

```
                    ┌─────────────┐
                    │   bank      │ (SDK)
                    └──────┬──────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
         ▼                 ▼                 ▼
   ┌───────────┐    ┌───────────┐    ┌───────────┐
   │ identity  │    │ economics │    │ security  │
   └─────┬─────┘    └─────┬─────┘    └─────┬─────┘
         │                │                │
         └────────┬───────┴────────┬───────┘
                  │                │
                  ▼                ▼
           ┌───────────┐    ┌───────────┐
           │ vcregistry│    │   dex     │
           └─────┬─────┘    └─────┬─────┘
                 │                │
                 ▼                ▼
          ┌────────────┐   ┌───────────┐
          │confidences.│   │  bridge   │
          └─────┬──────┘   └───────────┘
                │
                ▼
         ┌─────────────┐
         │ inclusion   │
         │ routines    │
         └─────────────┘
                │
                ▼
         ┌─────────────┐
         │    wasm     │
         └─────────────┘
```
