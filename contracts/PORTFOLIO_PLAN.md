# CosmWasm Contract Portfolio Plan

This document captures the design of the production-grade CosmWasm contracts we will ship to prove Aura’s custom bindings. Each contract description enumerates its state, entry points, keeper interactions, and coverage expectations so that implementation can proceed without ambiguity.

---

## 1. VC Issuer (Contract ID: `vc-issuer`)

**Purpose**  
Provide an on-chain authority that mints verifiable credentials (VCs) for accounts that satisfy policy requirements. This contract exercises the Aura VC registry bindings end-to-end.

**State Model**
| Key | Type | Description |
| --- | ---- | ----------- |
| `issuers` | Map<Address, IssuerProfile> | Approved issuers and their metadata (policy id, status, limits). |
| `requests` | Map<RequestID, IssueRequest> | Pending issuance requests keyed by deterministic hash. |
| `issued` | Map<VcId, IssueRecord> | Audit log of VCs minted via this contract (mirrors keeper for cross-checking). |

**Core Entry Points**
1. `execute::register_issuer { issuer, policy, daily_limit }`
   - Validates caller is governance multisig.
   - Writes to `issuers`.
2. `execute::request_vc { issuer, subject, vc_type, metadata }`
   - Any user can request issuance from an approved issuer.
   - Writes to `requests`.
3. `execute::fulfill_request { request_id }`
   - Only the target issuer can call.
   - Calls AuraMsg `vc_registry.register_vc`.
   - Records VC in `issued`.
4. `execute::revoke_vc { vc_id, reason }`
   - Wraps Aura VC revocation binding for off-chain parity.

**Queries**
1. `query::issuer { address }` – returns IssuerProfile.
2. `query::pending_requests { issuer }` – paginated view into `requests`.
3. `query::issued_vcs { issuer, subject? }` – for auditing.

**Aura Interactions**
- `AuraMsg::vc_registry.register_vc`
- Optional: `MsgRevokeVC`
- Queries: `AuraQuery::vc_registry.get_vc`

**Testing Checklist**
- Unit tests for access control, limits, and queue handling.
- Schema export for Execute/Query messages.
- Integration test hitting Aura keeper via go/sim to ensure minted VC appears in chain state.

---

## 2. Disclosure Verifier (Contract ID: `vc-disclosure-verifier`)

**Purpose**  
Manage selective disclosure workflows: issue requests to holders, accept proof responses, and enforce policies before granting access tokens. Exercises Aura selective-disclosure helpers once implemented.

**State Model**
| Key | Type | Description |
| --- | ---- | ----------- |
| `policies` | Map<PolicyID, DisclosurePolicy> | Required attributes, expiry, response window. |
| `requests` | Map<RequestID, DisclosureSession> | Tracks outstanding disclosure flows per holder. |
| `verifier_tokens` | Map<Address, AccessGrant> | Issued grants after successful disclosure. |

**Core Entry Points**
1. `execute::set_policy { policy }`
   - Governance only.
   - Writes to `policies`.
2. `execute::request_disclosure { holder, policy_id, purpose }`
   - Verifier calls to start session.
   - Calls AuraMsg (future) `create_disclosure_request`.
3. `execute::ingest_response { request_id, approved, attributes }`
   - Called by holder via off-chain front end.
   - Validates response vs. `policies`.
   - Issues `AccessGrant` token if approved.
4. `execute::revoke_grant { holder }`

**Queries**
1. `query::policy { id }`
2. `query::request { id }`
3. `query::grant { holder }`

**Aura Interactions**
- `AuraMsg::disclosure.request`
- `AuraMsg::disclosure.respond`
- `AuraQuery::disclosure.policy`

**Testing Checklist**
- Unit tests covering approval/denial paths, expirations, and grant issuance.
- Schema export verifying JSON structure consumers expect.
- Integration test that round-trips request/response through Aura keeper mocks.

---

## Deliverables Matrix

| Item | VC Issuer | Disclosure Verifier |
| ---- | --------- | ------------------- |
| Contract implementation | ✅ planned | ✅ planned |
| Unit tests | ✅ (policy, limits, issuance) | ✅ (policy eval, expiry, grants) |
| Schema export (`cargo schema`) | ✅ | ✅ |
| README + usage docs | ✅ | ✅ |
| Integration test (Go harness) | ✅ (mint VC) | ✅ (request/response) |
| CI hooks (fmt/clippy/test/wasm) | shared workflows | shared workflows |

---

## CI Enhancements
1. Add `contracts/Makefile` (or cargo workspace commands) to run:
   - `cargo fmt --all`
   - `cargo clippy --all-targets -- -D warnings`
   - `cargo test --workspace`
   - `cargo schema` per contract (output to `schema/`)
   - `cargo wasm` (optimizer pipeline) per contract.
2. GitHub Workflow `contracts.yml` that invokes above on PRs and publishes artifacts/schemas.
3. Hash check for `third_party/wasmvm/internal/api/libwasmvm.x86_64.so` to ensure artifacts stay in sync.

This plan should be implemented incrementally: start with the VC Issuer (state machine + messages), then layer in the Disclosure Verifier, and finally wire the CI/doc automation. Each milestone should update this file and `AGENT_PROGRESS`.
