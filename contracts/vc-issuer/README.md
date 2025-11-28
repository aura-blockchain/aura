# VC Issuer Contract

This CosmWasm contract implements a policy-driven issuance authority for Aura verifiable credentials. It exercises the `AuraMsg::VCRegistry` bindings and acts as the canonical example for integrating Cosmos contracts with Aura’s VC registry.

## Features

- Admin-managed issuer registry with activation toggles and policy metadata.
- Queue of issuance requests per subject/issuer.
- Execution path for issuers to fulfill a request, which emits an Aura VC registration message.
- Query endpoints for issuers, pending requests, and issued credentials (audit trail).

## Messages

| ExecuteMsg | Description |
| --- | --- |
| `RegisterIssuer { issuer, policy_id, daily_limit }` | Admin-only registration of issuers. |
| `UpdateIssuerStatus { issuer, active }` | Admin-only toggle. |
| `RequestVc { issuer, subject, vc_type, metadata }` | Subjects (or dApps) request issuance from an approved issuer. |
| `FulfillRequest { request_id, credential_base64 }` | Issuer mints a VC via Aura bindings and closes the request. |
| `RevokeVc { vc_id, reason }` | Emits metadata for off-chain tracking (future binding integration). |

| QueryMsg | Description |
| --- | --- |
| `Issuer { address }` | Fetch the issuer profile. |
| `PendingRequests { issuer }` | List pending requests for an issuer (up to 25). |
| `IssuedBySubject { subject }` | Audit previously issued credentials for a subject. |

## Build & Test

```bash
# Run unit tests
cargo test -p vc-issuer

# Generate JSON schema (output in schema/)
cargo run -p vc-issuer --example schema

# Build optimized wasm (requires rust-optimizer or similar)
cargo wasm -p vc-issuer
```

## Roadmap

Planned enhancements include:
- Rate-limit enforcement (per issuer per day).
- Integration with disclosure bindings for selective VC issuance.
- Configurable VC metadata templates.
