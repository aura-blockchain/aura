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
5. When a real Cosmos SDK application exists, hook `chain/app/module_manager.go` into the app so `RegisterGRPCServices` wires the `identitychange` Msg/Query servers into the canonical gRPC server used by wallets and assistants.
- The minimal Cosmos SDK app shell lives under `chain/app` (`CosmosApp` + `ModuleManager`) so you can boot up a baseapp instance, register the module’s Msg/Query gRPC services, and expose the gRPC server for transport wiring.
