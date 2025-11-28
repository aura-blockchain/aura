# Keeper Consolidation - Line Reference

## File: /home/decri/blockchain-projects/aura/chain/app/app.go

### Imports Section (Lines 71-86)
- **Lines 71-80**: Added imports for consolidated modules (security, identity, economics)
- **Lines 82-86**: Legacy module imports (marked for future removal)

### App Struct - Keeper Fields (Lines 142-157)
- **Lines 142-145**: Consolidated keeper declarations
  - Line 143: `securityKeeper` - Replaces 6 modules
  - Line 144: `identityKeeper` - Replaces identitychange
  - Line 145: `economicsKeeper` - Replaces economicsecurity + governance
  
- **Lines 147-157**: Core module keepers (unchanged)
  - inclusionroutines, confidencescore, vcregistry, dataregistry, compliance
  - dex, bridge, ai, contractregistry, wasmSecurity

### App Struct - Store Keys (Lines 160-187)
- **Lines 161-168**: Cosmos SDK standard keys
- **Lines 170-173**: Consolidated module keys
  - Line 171: `security` store key
  - Line 172: `identity` store key  
  - Line 173: `economics` store key
- **Lines 175-186**: Core AURA module keys (unchanged)

### App Struct - Memory Keys (Lines 188-190)
- **Line 189**: Only `vc` memory key retained
- Removed: `validatorSecurity` memory key (now in consolidated security module)

## Modules Replaced

### Security Module (securityKeeper)
Replaces 6 individual keepers:
1. networksecurityKeeper
2. validatorSecurityKeeper
3. walletSecurityKeeper
4. incidentResponseKeeper
5. cryptographyKeeper
6. privacyKeeper

Also removes:
- monitoringKeeper (absorbed into security)
- prevalidationKeeper (absorbed into security)

### Identity Module (identityKeeper)
Replaces 1 keeper:
1. idKeeper (identitychange)

### Economics Module (economicsKeeper)
Replaces 2 keepers:
1. economicSecurityKeeper
2. govKeeper (governance)

## Total Reduction
- **Before**: 11 individual keepers for consolidated modules
- **After**: 3 consolidated keepers
- **Reduction**: 8 fewer keeper fields (73% reduction)

## Status
✅ Keeper field declarations updated
✅ Store keys consolidated
✅ Memory keys updated
✅ Imports added for consolidated modules
⏳ Keeper initialization code (not yet updated - to be done separately)
⏳ Module registration (not yet updated - to be done separately)
⏳ Store key creation/mounting (not yet updated - to be done separately)
