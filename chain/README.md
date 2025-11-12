# Chain Module

## Purpose
Anchor the identity manager, inclusion routine registry, and confidence scoring logic in a Cosmos SDK module so AI assistants, wallets, and governance can share a single source of truth.

## Anchors
- `docs/rfcs/0002-inclusion-routines-module.md` defines the IR lifecycle, rate limits, and governance-managed metadata.
- `docs/architecture/flows/ir-completion.puml` / `vc-mint-revoke.puml` describe how proofs, VC issuance, and confidence scores flow through the modules.
- `PROJECT_STATUS.md` / `progress.md` summarize the expectations for this module being one of the first implementation milestones.

## Next steps
1. Translate `IRDefinition`, prerequisite graphs, and rate limits into Protobuf state and message types.
2. Implement CRUD handlers with governance gating, event emission, and query surfaces for wallets/assistants.
3. Build unit/fuzz tests for lifecycle transitions (draft, active, suspended, retired) and prerequisite DAG validation.
4. Integrate with the AI assistant attestation pipeline once the on-chain/off-chain complementing contracts are ready.

