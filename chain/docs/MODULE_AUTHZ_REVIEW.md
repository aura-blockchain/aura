# Module AuthZ Surface Review (2026-02-14)

- Scope: Msg/Query signers, authority params, governance overrides, emergency/pause/circuit hooks across 27 modules.
- Approach: traced Msg servers + keeper UpdateParams/Admin paths; confirmed signer checks, governance `authority` enforcement, and pause guards where applicable.

**Validated (pass):**
- aiassistant, governance, dex, bridge, wasm, contractregistry, validatorsecurity, security, monitoring, inclusionroutines, confidencescore, economics, economicsecurity, dataregistry, identity, identitychange, privacy, vcregistry, aurabindings, networksecurity, incidentresponse, cryptography, aiassistant (params), auth (roles/sessions), walletsecurity params (no ext deps).

**Notable patterns:**
- Governance authority consistently enforced on MsgUpdateParams where present; signer equality enforced on Msgs with single owner/admin semantics.
- Pause/circuit paths: security + bridge + wasm use authority or admin checks; dex operations depend on security pause guard.
- Contractregistry enforces admin on metadata/security policy updates; pause allows governance override as designed.

**Gaps / follow-ups:**
- walletsecurity: now enforces active auth sessions when an auth keeper is injected; consider wiring auth keeper in app wiring for production. 2FA/session checks present; biometric remains deprecated.
- prevalidation: [resolved] now calls compliance keeper sanctions check before accepting txs (MEDIUM-004 addressed).
- Authorization error codes mixed (`PermissionDenied` vs `Unauthorized`); standardize to gRPC `codes.PermissionDenied`.
- Rate limit for auth attempts added via `AuthRateLimitDecorator` in ante handler, backed by walletsecurity keeper storage.
- PermissionDenied audit checklist for new/updated Msg servers:
  - Map all auth/permission failures (wrong authority/admin/signer) to `status.Error(codes.PermissionDenied, ...)`.
  - Table-driven tests that exercise: nil/empty signer, wrong signer, wrong authority, and happy-path signer to guard regressions.
  - Add quick property/fuzz test for signer mismatch mapping to PermissionDenied.
  - Emit structured audit events/telemetry for denied actions to aid observability.
  - Wire into CI via `make test-authz` (targeted auth test suite) once added.

Artifacts: dependency map `chain/docs/MODULE_BOUNDARY_INVENTORY.md`; authz review notes (this file).
