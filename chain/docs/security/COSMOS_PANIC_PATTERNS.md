# Cosmos SDK Panic Patterns

## Genesis Panics (InitGenesis/ExportGenesis)

Panics in `InitGenesis` and `ExportGenesis` are **standard Cosmos SDK practice**.
These functions run during chain startup - if they fail, the chain cannot operate safely.

**Why panics are correct here:**
- Genesis state is set once at chain creation
- Invalid genesis = fundamentally broken chain
- No recovery possible if genesis is malformed
- Returning an error would be silently ignored

Reference: All Cosmos SDK core modules (bank, staking, gov) use this pattern.

## Keeper Constructor Panics

Panics in `NewKeeper()` are acceptable when:
- Missing required dependencies (codec, store key, bank keeper)
- App wiring is fundamentally broken
- No sensible default exists

These panics fire during app initialization, not during transaction processing.
