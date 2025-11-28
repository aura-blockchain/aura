# Genesis Modules Quick Reference

## Module Consolidation Map

### Security Module (NEW)
**Replaces:**
- networksecurity
- validatorsecurity
- walletsecurity
- incidentresponse
- cryptography
- privacy

**Genesis Sections:**
```json
"security": {
  "params": {...},
  "network": {...},
  "validator": {...},
  "wallet": {...},
  "incident": {...},
  "crypto": {...},
  "privacy": {...}
}
```

### Identity Module (NEW)
**Replaces:**
- Custom auth module
- identitychange

**Genesis Sections:**
```json
"identity": {
  "params": {...},
  "records": [],
  "change_requests": [],
  "change_history": [],
  "suspended": false,
  "audit_trail": []
}
```

### Economics Module (NEW)
**Replaces:**
- economicsecurity
- governance

**Genesis Sections:**
```json
"economics": {
  "params": {...},
  "vesting_schedules": [],
  "treasury_state": {...},
  "governance_state": {...},
  "dynamic_fees": {...},
  "whale_protection": {...},
  "mev_state": {...}
}
```

### VCRegistry Module (ENHANCED)
**Now absorbs:**
- dataregistry

**Added Field:**
```json
"vcregistry": {
  ...,
  "data_items": []  // NEW - from dataregistry
}
```

### Wasm Module (NEW)
**Replaces:**
- contractregistry
- aura-bindings

**Genesis Sections:**
```json
"wasm": {
  "params": {...},
  "codes": [],
  "contracts": [],
  "sequences": []
}
```

## Removed Modules
These modules no longer appear in genesis:
- ❌ networksecurity
- ❌ validatorsecurity
- ❌ walletsecurity
- ❌ incidentresponse
- ❌ cryptography
- ❌ privacy
- ❌ economicsecurity
- ❌ governance (custom)
- ❌ auth (custom)
- ❌ identitychange
- ❌ monitoring
- ❌ aiassistant
- ❌ prevalidation
- ❌ contractregistry
- ❌ aura-bindings
- ❌ dataregistry

## Unchanged Modules
Standard Cosmos modules (kept as-is):
- ✅ auth (Cosmos standard)
- ✅ bank
- ✅ staking
- ✅ slashing
- ✅ distribution
- ✅ genutil

AURA core modules (kept as-is):
- ✅ vcregistry (enhanced with data_items)
- ✅ confidencescore
- ✅ inclusionroutines
- ✅ dex
- ✅ bridge
- ✅ compliance

## Complete Genesis Module List (16 modules)

1. auth (Cosmos)
2. bank (Cosmos)
3. staking (Cosmos)
4. slashing (Cosmos)
5. distribution (Cosmos)
6. genutil (Cosmos)
7. **security** (NEW consolidated)
8. **identity** (NEW consolidated)
9. **economics** (NEW consolidated)
10. vcregistry (enhanced)
11. confidencescore
12. inclusionroutines
13. dex
14. bridge
15. compliance
16. **wasm** (NEW)

## Migration Notes

### From Old to New Structure:

**Security features:**
```
networksecurity → security.network
validatorsecurity → security.validator
walletsecurity → security.wallet
incidentresponse → security.incident
cryptography → security.crypto
privacy → security.privacy
```

**Identity features:**
```
identitychange → identity
(custom auth merged into Cosmos auth)
```

**Economics features:**
```
economicsecurity → economics (vesting, fees, whale protection)
governance → economics.governance_state
```

**Smart contracts:**
```
contractregistry → wasm
aura-bindings → wasm
```

**Data:**
```
dataregistry → vcregistry.data_items
```
