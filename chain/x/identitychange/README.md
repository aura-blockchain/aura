# Identity Change Module

The Cosmos SDK `identitychange` module tracks holder-driven DID rotations, metadata refreshes, and confidence score updates once AI assistants re-verify an inclusion routine. It is the implementation companion to `docs/rfcs/0008-identity-change-module.md`.

## Subsystems
- `proto/aura/identitychange/v1beta1/identity_change.proto`: Protobuf definitions for state (`IdentityRecord`, `IdentityChangeRequest`, `IdentityChangeHistory`, `Params`) plus `Msg`/`Query` services.
- `types/`: basic in-memory types, status enum, and parameter defaults/validation used before the proto-generated code becomes available.
- `keeper/`: lightweight keeper that stores records/requests in maps, enforces rate limits, tracks history, and exposes getters; it currently simulates the handler logic needed by message servers.

## Integration points
- Future `module.go`, `msg_server.go`, and `query_server.go` files will wire the keeper into the Cosmos SDK message routing.
- Wallet clients and compliance tooling will query `IdentityChangeRequest` and `IdentityChangeHistory` via the query service that is described in the protobuf file.
- Governance will invoke `SuspendIdentityChanges` before upgrades or security events to halt request intake.

## Next work
1. Wire the stub `MsgServer`/`QueryServer` (already added) into a full Cosmos module with Protobuf-generated bindings.
2. Add parameter store wiring and config validation once the Cosmos `params` module is available.
3. Hook the module into the app’s module manager and register gRPC query routes.
4. Implement the message handlers described in the RFC (`RequestIdentityChange`, `SubmitAssistantProof`, `ApplyIdentityChange`, `RejectIdentityChange`, `SuspendIdentityChanges`) plus the unit/integration tests outlined in the RFC.
