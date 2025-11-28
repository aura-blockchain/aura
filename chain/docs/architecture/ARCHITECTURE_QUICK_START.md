# AURA Consolidated Architecture - Quick Start Guide

This is a quick reference guide to help you navigate the comprehensive [CONSOLIDATED_ARCHITECTURE.md](./CONSOLIDATED_ARCHITECTURE.md) document.

## TL;DR

- **Modules reduced from 24 to 11** (54% reduction)
- **3 new consolidated modules:** security, identity, economics
- **No state migration needed** (development phase)
- **All functionality preserved**, just better organized

## The Three New Consolidated Modules

### 1. Security Module
**What it consolidates:** networksecurity, validatorsecurity, walletsecurity, incidentresponse, cryptography, privacy

**When to use:** Anything related to protecting the network, validators, wallets, handling incidents, cryptographic operations, or privacy features.

**Store key:** `security`

### 2. Identity Module  
**What it consolidates:** auth (custom), identitychange

**When to use:** User authentication, role management, DID operations, identity changes, audit trails.

**Store key:** `identity`

### 3. Economics Module
**What it consolidates:** economicsecurity, governance

**When to use:** Fees, vesting, treasury, governance proposals, voting, MEV protection, whale protection.

**Store key:** `economics`

## Quick Module Lookup

| Need to... | Use this module |
|-----------|----------------|
| Issue/verify credentials | vcregistry |
| Calculate verifier scores | confidencescore |
| Manage inclusion routines | inclusionroutines |
| DEX operations | dex |
| Cross-chain bridging | bridge |
| Compliance tracking | compliance |
| Smart contracts | wasm |
| Network security | security |
| Validator protection | security |
| Wallet security | security |
| Emergency response | security |
| Cryptography | security |
| Privacy features | security |
| User roles/auth | identity |
| DID management | identity |
| Identity changes | identity |
| Dynamic fees | economics |
| Governance | economics |
| Vesting | economics |
| Treasury | economics |

## Import Paths Changed

**Old way:**
```go
import "github.com/aequitas/aura/chain/x/networksecurity"
import "github.com/aequitas/aura/chain/x/validatorsecurity"
import "github.com/aequitas/aura/chain/x/walletsecurity"
```

**New way:**
```go
import "github.com/aequitas/aura/chain/x/security"
```

## Key Prefix Organization

Each consolidated module uses prefixed key ranges to maintain logical separation:

### Security Module
- `0x01-0x09`: Network security
- `0x10-0x17`: Validator security  
- `0x20-0x2A`: Wallet security
- `0x30-0x35`: Incident response
- `0x40-0x48`: Cryptography
- `0x50-0x55`: Privacy

### Identity Module
- `0x01-0x0e`: Auth/roles
- `0x10-0x17`: Identity changes
- `0x20-0x21`: Counters

### Economics Module
- `0x01-0x04`: Fee management
- `0x10-0x13`: Vesting
- `0x20-0x23`: Treasury
- `0x30-0x38`: Governance
- `0x40-0x44`: Monitoring
- `0x50-0x52`: MEV protection

## Common Questions

**Q: Do I need to migrate existing state?**
A: No, we're in development phase. No migration needed.

**Q: Where did monitoring go?**
A: Removed. Use Prometheus + Grafana instead.

**Q: Where did aiassistant go?**
A: Removed. Deploy as separate off-chain service.

**Q: Where did prevalidation go?**
A: Merged into the ante handler (`chain/app/ante.go`).

**Q: Can I still access network security features?**
A: Yes, they're in the security module now.

**Q: How do I add a new network security feature?**
A: Add it to `chain/x/security/keeper/network.go` with prefix `0x01-0x09`.

**Q: Will this break my existing code?**
A: Yes, import paths need updating. But functionality is preserved.

## File Structure

```
chain/x/
├── security/       # NEW - 6 modules consolidated
├── identity/       # NEW - 2 modules consolidated  
├── economics/      # NEW - 2 modules consolidated
├── vcregistry/     # Retained (expanded)
├── confidencescore/# Retained
├── inclusionroutines/ # Retained
├── dex/           # Retained
├── bridge/        # Retained
├── compliance/    # Retained
└── wasm/          # Retained (expanded)
```

## Next Steps

1. **Read the full doc:** [CONSOLIDATED_ARCHITECTURE.md](./CONSOLIDATED_ARCHITECTURE.md)
2. **Update imports:** Change old module imports to new consolidated modules
3. **Review keeper dependencies:** Section 5 of main doc
4. **Check genesis structure:** Section 6 of main doc
5. **Follow dev guidelines:** Section 9 of main doc

## Get Help

- Full documentation: [CONSOLIDATED_ARCHITECTURE.md](./CONSOLIDATED_ARCHITECTURE.md)
- Architecture diagrams: Sections 1 and 5
- Migration guide: Section 7
- Developer guidelines: Section 9
