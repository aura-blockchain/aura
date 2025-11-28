# AURA Blockchain Architecture Documentation

Welcome to the AURA blockchain architecture documentation. This directory contains comprehensive documentation about the consolidated module architecture.

## Quick Navigation

### 🚀 Getting Started
**Start here if you're new to the project:**
- [**ARCHITECTURE_QUICK_START.md**](./ARCHITECTURE_QUICK_START.md) - Quick reference guide (5 min read)
- [**CONSOLIDATION_VISUAL.md**](./CONSOLIDATION_VISUAL.md) - Visual diagrams and charts (10 min read)

### 📚 Comprehensive Documentation
**Deep dive into the architecture:**
- [**CONSOLIDATED_ARCHITECTURE.md**](./CONSOLIDATED_ARCHITECTURE.md) - Complete architecture specification (30 min read)

## What's Inside

### CONSOLIDATED_ARCHITECTURE.md (37 KB, 1,219 lines)
The complete architecture specification covering:
- Module architecture overview
- Detailed breakdown of 3 new consolidated modules (security, identity, economics)
- Documentation of 7 retained modules
- Explanation of removed modules
- Keeper dependency graph
- Genesis state structure
- Migration notes
- File structure
- Developer guidelines

### ARCHITECTURE_QUICK_START.md (4.6 KB)
A quick reference guide with:
- TL;DR summary
- Module lookup table
- Import path changes
- Key prefix organization
- Common questions and answers
- Quick file structure overview

### CONSOLIDATION_VISUAL.md (23 KB)
Visual representations including:
- Before/after comparison
- Consolidation flow diagrams
- Benefits summary
- Migration checklist
- Key metrics

## Module Consolidation Summary

### The Big Picture
- **Before:** 24 modules
- **After:** 11 modules
- **Reduction:** 54%

### New Consolidated Modules

1. **Security Module** - Consolidates 6 modules
   - networksecurity, validatorsecurity, walletsecurity
   - incidentresponse, cryptography, privacy

2. **Identity Module** - Consolidates 2 modules
   - auth (custom), identitychange

3. **Economics Module** - Consolidates 2 modules
   - economicsecurity, governance

### Expanded Modules

1. **VCRegistry** - Absorbs dataregistry concepts
2. **WASM** - Absorbs contractregistry and aura-bindings

### Removed Modules

1. **monitoring** → Off-chain (Prometheus/Grafana)
2. **aiassistant** → Off-chain service
3. **prevalidation** → Merged into ante handler

## Directory Structure

```
docs/architecture/
├── README.md                        # This file
├── CONSOLIDATED_ARCHITECTURE.md     # Complete specification
├── ARCHITECTURE_QUICK_START.md      # Quick reference
└── CONSOLIDATION_VISUAL.md          # Visual diagrams
```

## For Different Audiences

### New Developers
1. Start with [ARCHITECTURE_QUICK_START.md](./ARCHITECTURE_QUICK_START.md)
2. Look at visual diagrams in [CONSOLIDATION_VISUAL.md](./CONSOLIDATION_VISUAL.md)
3. Read specific sections of [CONSOLIDATED_ARCHITECTURE.md](./CONSOLIDATED_ARCHITECTURE.md) as needed

### Existing Developers
1. Review [ARCHITECTURE_QUICK_START.md](./ARCHITECTURE_QUICK_START.md) for import path changes
2. Check Section 7 of [CONSOLIDATED_ARCHITECTURE.md](./CONSOLIDATED_ARCHITECTURE.md) for migration notes
3. Use Section 9 for developer guidelines

### Architects
1. Read [CONSOLIDATED_ARCHITECTURE.md](./CONSOLIDATED_ARCHITECTURE.md) in full
2. Review the keeper dependency graph (Section 5)
3. Study the genesis state structure (Section 6)

### DevOps/Infrastructure
1. Check removed modules in [CONSOLIDATED_ARCHITECTURE.md](./CONSOLIDATED_ARCHITECTURE.md) Section 4
2. Review genesis structure in Section 6
3. Note the store key changes in Section 7

## Related Documentation

- `/home/decri/blockchain-projects/aura/chain/docs/modules/` - Individual module documentation
- `/home/decri/blockchain-projects/aura/chain/docs/development/` - Development guides
- `/home/decri/blockchain-projects/aura/chain/docs/ops/` - Operations documentation

## Contributing

When adding new features to consolidated modules:

1. Identify the appropriate domain within the module
2. Use the correct key prefix for that domain
3. Follow the naming conventions in Section 9.2
4. Add comprehensive tests
5. Update this documentation

## Questions?

- For architecture questions, review the FAQ in ARCHITECTURE_QUICK_START.md
- For implementation details, see CONSOLIDATED_ARCHITECTURE.md Section 9
- For visual understanding, see CONSOLIDATION_VISUAL.md

## Version

- **Documentation Version:** 1.0
- **Last Updated:** 2025-11-27
- **AURA Version:** Development Phase

---

**Quick Stats:**
- Total documentation: ~65 KB
- Total lines: ~2,800
- Reading time: ~45 minutes (all docs)
- Quick reference time: ~5 minutes
