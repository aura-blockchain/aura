# AURA Documentation Structure

**Last Updated:** November 25, 2025

This document describes the organization of AURA blockchain documentation.

## Root Level Documentation (5 files)

The root directory contains only essential, user-facing documentation:

1. **README.md** - Project overview and introduction
2. **LICENSE** - Project license
3. **CODE_OF_CONDUCT.md** - Community guidelines
4. **QUICK_START.md** - Consolidated quick start guide
5. **AGENT_PROGRESS** - Working file for development tracking

## Documentation Directory Structure

```
docs/
├── architecture/           # System architecture and design
├── archive/               # Historical phase reports
├── compliance/            # Legal and compliance documents
├── developers/            # Developer integration guides
├── development/           # Development, testing, and coverage reports
├── economics/             # Economic models and tokenomics
├── modules/               # Module-specific documentation
│   ├── auth/
│   ├── confidencescore/
│   ├── dex/
│   ├── economicsecurity/
│   └── inclusionroutines/
├── ops/                   # Operations and deployment
├── proto/                 # Protocol buffer documentation
├── testing/               # Testing guides
└── wallet/                # Wallet documentation
```

## Directory Contents

### /docs/architecture/
System architecture, design documents, and smart contract proposals.

**Files:**
- CONTRACTREGISTRY_ARCHITECTURE.md
- SMART_CONTRACT_IMPLEMENTATION_TASKS.md
- SMART_CONTRACT_INTEGRATION_PROPOSAL.md
- README.md

### /docs/archive/
Historical phase reports and completed project documentation.

**Files:**
- PHASE3_PROTO_DEFINITIONS_COMPLETE.md
- PHASE3_TASK3.1_SUMMARY.md
- PHASE4_TASK6.1_PROTO_DEFINITIONS_REPORT.md

### /docs/developers/
Integration guides and SDK documentation for developers.

**Files:**
- SDK_INTEGRATION_QUICK_REFERENCE.md

### /docs/development/
Development guides, test coverage reports, and implementation summaries.

**Files:**
- BATCH3_TYPES_COVERAGE_REPORT.md
- CLI_COMMANDS_REPORT.md
- CLI_COVERAGE_REPORT.md
- CLI_TEST_COVERAGE_REPORT.md
- COMPREHENSIVE_TEST_SUITE_COMPLETION.md
- GRPC_IMPLEMENTATION_SUMMARY.md
- GRPC_SERVERS_IMPLEMENTATION_SUMMARY.md
- MODULE_SECURITY_TODO.md
- Progress_KV_Migration.md
- TESTING.md
- TESTING_INFRASTRUCTURE_COMPLETE.md
- TESTING_OPTIMIZATIONS.md
- TEST_COVERAGE_IMPROVEMENT_REPORT.md
- TEST_COVERAGE_REPORT.md
- TEST_IMPLEMENTATION_SUMMARY.md
- TYPES_BATCH2_COVERAGE_REPORT.md

### /docs/modules/
Module-specific documentation organized by module name.

**Subdirectories:**
- **auth/** - CLI_COMMANDS.md
- **confidencescore/** - CLI_COMMANDS_SUMMARY.md, CLI_QUICK_REFERENCE.md, COMMAND_TREE.txt
- **dex/** - DEX_CLI_COMMANDS.md
- **economicsecurity/** - ECONOMICSECURITY_GRPC_COMPLETE.md, ECONOMICSECURITY_RPC_MAPPING.md
- **inclusionroutines/** - CLI_IMPLEMENTATION_SUMMARY.md, CLI_QUICK_REFERENCE.md

### /docs/proto/
Protocol buffer definitions and references.

**Files:**
- AURABINDINGS_PROTO_REFERENCE.md
- CONTRACTREGISTRY_PROTO_QUICK_REFERENCE.md

### /docs/wallet/
Wallet integration and status reports.

**Files:**
- BROWSER_WALLET_STATUS_INTEGRATION_REPORT.md

## Files Removed

The following temporary and redundant files were deleted:

1. **AGENT_ACCESS.md** - Temporary agent documentation
2. **AGENT_TESTING_GUIDE.md** - Temporary testing guide
3. **economicsecurity_grpc_implementation_report.md** - Consolidated into module docs
4. **QUICK_START_ECOSYSTEM.md** - Consolidated into QUICK_START.md
5. **QUICK_START_ECOSYSTEM_TOOLS.md** - Consolidated into QUICK_START.md
6. **QUICK_FIX_GUIDE.md** - Consolidated into QUICK_START.md

## Documentation Migration Summary

### From Root → /docs/archive/
- PHASE3_PROTO_DEFINITIONS_COMPLETE.md
- PHASE3_TASK3.1_SUMMARY.md
- PHASE4_TASK6.1_PROTO_DEFINITIONS_REPORT.md

### From Root → /docs/development/
- CLI_COMMANDS_REPORT.md
- MODULE_SECURITY_TODO.md
- Progress_KV_Migration.md

### From Root → /docs/architecture/
- CONTRACTREGISTRY_ARCHITECTURE.md
- SMART_CONTRACT_IMPLEMENTATION_TASKS.md
- SMART_CONTRACT_INTEGRATION_PROPOSAL.md

### From Root → /docs/proto/
- AURABINDINGS_PROTO_REFERENCE.md
- CONTRACTREGISTRY_PROTO_QUICK_REFERENCE.md

### From Root → /docs/modules/
- DEX_CLI_COMMANDS.md → modules/dex/
- ECONOMICSECURITY_GRPC_COMPLETE.md → modules/economicsecurity/
- ECONOMICSECURITY_RPC_MAPPING.md → modules/economicsecurity/

### From Root → /docs/developers/
- SDK_INTEGRATION_QUICK_REFERENCE.md

### From Root → /docs/wallet/
- BROWSER_WALLET_STATUS_INTEGRATION_REPORT.md

### From /chain/ → /docs/development/
- BATCH3_TYPES_COVERAGE_REPORT.md
- CLI_COVERAGE_REPORT.md
- CLI_TEST_COVERAGE_REPORT.md
- COMPREHENSIVE_TEST_SUITE_COMPLETION.md
- GRPC_IMPLEMENTATION_SUMMARY.md
- GRPC_SERVERS_IMPLEMENTATION_SUMMARY.md
- TESTING.md
- TESTING_INFRASTRUCTURE_COMPLETE.md
- TESTING_OPTIMIZATIONS.md
- TEST_COVERAGE_IMPROVEMENT_REPORT.md
- TEST_COVERAGE_REPORT.md
- TEST_IMPLEMENTATION_SUMMARY.md
- TYPES_BATCH2_COVERAGE_REPORT.md

### From /chain/x/[module]/ → /docs/modules/[module]/
- chain/x/auth/CLI_COMMANDS.md → modules/auth/
- chain/x/confidencescore/CLI_*.md → modules/confidencescore/
- chain/x/inclusionroutines/CLI_*.md → modules/inclusionroutines/

## Quick Navigation

### For Users
- Start with: **/README.md**
- Quick setup: **/QUICK_START.md**
- Community: **/CODE_OF_CONDUCT.md**

### For Developers
- SDK integration: **/docs/developers/**
- Module docs: **/docs/modules/**
- Proto reference: **/docs/proto/**

### For Contributors
- Development guide: **/docs/development/**
- Testing guide: **/docs/testing/**
- Architecture: **/docs/architecture/**

### For Operators
- Deployment: **/docs/ops/**
- Monitoring: **/docs/ops/runbooks/**

## Documentation Standards

1. **File Naming**: Use SCREAMING_SNAKE_CASE for documentation files
2. **Organization**: Place files in the most specific relevant directory
3. **Cross-References**: Use absolute paths from repository root
4. **Updates**: Update DOCUMENTATION_STRUCTURE.md when adding new directories
5. **Archive**: Move completed phase/project docs to /docs/archive/

## Statistics

- **Root-level docs**: 5 files (down from 29)
- **Total documentation files**: 64 markdown files
- **Documentation directories**: 15 directories
- **Module docs**: 5 modules documented

## Maintenance

To maintain this structure:

1. Keep root level minimal (5-7 essential files only)
2. Organize new docs into appropriate subdirectories
3. Archive completed phase reports to /docs/archive/
4. Update this file when structure changes
5. Use consistent naming conventions

---

**This structure prioritizes discoverability, organization, and maintainability.**
