# IdentityChange Module Integration

This document outlines how the `identitychange` module should be wired into the wider Cosmos SDK app once the chain framework is in place.

## Params Registration

- Instantiate `params.Store` (defaulting to `types.DefaultParams()`) and pass it to `keeper.NewKeeper` so all request handlers read the live configuration.
- When the SDK `params` module is available, replace the in-memory store with the `paramskeeper.Subspace` so governance proposals can update the values listed in `types.Params`.

## App Module Wiring

1. Create a `module.go` defining `AppModuleBasic` (codec registration, genesis validation) and `AppModule` (route/handler registration).
2. Register the gRPC `MsgServer`/`QueryServer` implementations from `msg_server.go`/`query_server.go` using the Cosmos SDK module wiring helpers (`module.RegisterServices` or `app/router` entries).
3. Plug the keeper into the module’s `AppModule` so handlers call `keeper.CreateRequest`, `SubmitProof`, `ApplyChange`, etc.
4. Include `identitychange` in the app’s `ModuleBasics` map, `ModuleManager`, and `GenesisState` registration so it participates in init, begin/end blocks, and simulations.

## Testing & Governance Hooks

- Use the provided keeper tests to lock down rate limits and confidence thresholds; extend them with gRPC integration tests once the module is part of the app.
- Add hooks to emit events (`IdentityChangeRequested`, `IdentityChangeApplied`, etc.) and to connect with governance proposals for suspending requests.

## Next Actions

- Generate the real protobuf bindings from `proto/aura/identitychange/v1beta1/identity_change.proto` using `buf generate` or `protoc` so the module shares the same message/query contract as other SDK modules.
- Replace the placeholder `proto_stub.go` with the generated file and rerun `go test ./...` once the Cosmos SDK dependencies have been added to `chain/go.mod`.
