# RFC-0008: Identity Change Module

- **Author(s):** Identity Team
- **Status:** Draft
- **Created:** 2025-11-12
- **Target Release:** Community Testnet

## Summary

Define an on-chain `IdentityChange` module that tracks a holder’s requests to update their DID, metadata, or confidence score after re-verification (e.g., new inclusion routine, lifestyle change, or recovery).  The module ensures every identity change is attested by the assistant network, auditable, and compatible with the `IdentityManager`/`InclusionRoutine` ecosystem.

## Motivation & Goals

- Provide a governed pathway for identity holders to rotate keys, refresh attributes, or change identity state without losing continuity.
- Keep a transparent, permissioned ledger of identity update attempts so fraud monitors and compliance auditors can trace the lineage of every DID.
- Prevent mass, unauthorized identity swaps (e.g., by re-verifying via stolen IR proofs) through staged approvals, nullifiers, and assistant attestations.

## Detailed Design

### State
- `IdentityRecord {did, owner, confidence_score, metadata_hash, latest_ir_version, last_changed_height, status}` keeps the canonical identity data referenced by wallets and governance.
- `IdentityChangeRequest {id, requester, target_did, request_meta_hash, assistant, ir_id, proof_hash, status, verdict_height, reason}` stores pending changes, their proofs, and audit trail.
- `IdentityChangeHistory` links past requests for governance/compliance audits; each record includes `prev_confidence_score`, `new_confidence_score`, and `transition_reason`.
- `IdentityChangeStatus` (enum): `Idle`, `PendingVerification`, `ReadyToApply`, `Rejected`, `Applied`, `Suspended`.

### Messages
- `MsgRequestIdentityChange` (signed by holder): submits metadata for the desired DID change plus which IR or proofs are attached. It mints a change request in `PendingVerification` state gated by cross-checks (e.g., `IRStatus == Active`, rate limits, locale restrictions).
- `MsgSubmitAssistantProof` (assistant-authority only): attaches the assistant’s signed attestation (`{request_id, proof_hash, confidence_delta}`) to the change request, shifting it to `ReadyToApply` if verification succeeded or `Rejected` if evidence contradicts.
- `MsgApplyIdentityChange` (holder or governance): finalizes the new DID/metadata, updates `IdentityRecord`, and emits `IdentityChanged` events. Holds until `ReadyToApply` and ensures confidence score thresholds (e.g., total `ConfidenceScore` still >= 1000 after change).
- `MsgRejectIdentityChange` (assistant or governance): marks a request as rejected or suspicious, optionally slashing the assistant stake if they misbehaved.
- `MsgSuspendIdentityChanges` (governance): temporarily halts new change requests (e.g., during a security emergency) while letting existing ones resolve.

### Events
- `EventIdentityChangeRequested`, `EventAssistantProofSubmitted`, `EventIdentityChangeApplied`, `EventIdentityChangeRejected`, `EventChangeSuspended` for off-chain listeners.

### Queries
- `QueryIdentityRecord(did)` returns the latest `IdentityRecord` + confidence score history.
- `QueryIdentityChangeRequest(request_id)` shows status, proof hashes, and assistant attestations.
- `QueryIdentityChangeHistory(did, pagination)` streams previous transitions for compliance.

### Access Control
- Holders can only request changes for `IdentityRecord.owner` matches their account; requests are rate limited (global defaults + per-locale overrides from `InclusionRoutine` module).
- Assistants on the `Active` whitelist (per RFC-0003) can submit proofs; malicious attestations can be slashed via governance hooks.
- Governance proposals can pause all change requests or forcibly reset suspicious records.

### Integration Points
- The module references `InclusionRoutine` definitions so only `Active` IRs whose `confidence_score` outputs meet the threshold can trigger a change. `IdentityManager` listens to `IdentityChangeApplied` events to update VC issuance and revoke outdated tokens.
- `Wallet` clients query `IdentityChangeRequest` to show users pending approvals and can optionally cancel (if still `PendingVerification`).
- `zk-governance` uses this module indirectly by verifying that only holders with a non-revoked, up-to-date identity record can produce governance commitments.

## Lifecycle State Machine

| State | Transition | Guard |
| --- | --- | --- |
| `Idle` | `RequestIdentityChange` | Rate limits + `IdentityRecord` match |
| `PendingVerification` | `SubmitAssistantProof` | Assistant attests (hash matches `proof_hash`) |
| `ReadyToApply` | `ApplyIdentityChange` | Confidence thresholds + no active suspensions |
| `Rejected` | `RequestIdentityChange` (new) | Fresh submission after addressing reason |
| `Applied` | (historical) | `IdentityRecord` updated; archived in history |
| Any | `SuspendIdentityChanges` | Governance emergency |

Requests can stay `PendingVerification` indefinitely until an assistant proof arrives; staleness triggers auditor alerts after `max_processing_height` (configurable parameter).

## Parameters
- `max_requests_per_wallet_per_month` (default 2).
- `min_confidence_after_change` (default 1000) ensures VC eligibility persists.
- `staleness_height_threshold` (default 10,000 blocks) triggers auto-rejection + investigator event.
- `assistant_slash_on_false_positive` bool toggled by governance to hold assistants accountable.

## Validation Plan

- Unit tests covering request creation, assistant proof acceptance/rejection, and confidence threshold enforcement.
- Integration test linking `InclusionRoutine` outputs to `IdentityChange` transitions (mock assistant proof flows).
- Fuzzing for `IdentityChangeRequest` history pagination and cycle detection (e.g., repeated cancels).

## Security / Privacy Considerations

- Proof hashes never expose raw biometric or IR data—we store only hashes + metadata references.
- Identity change requests are rate limited and require multiple attestations whenever `confidence_score` changes by more than `delta_threshold` to prevent takeover.
- Governance can freeze requests and audit change histories (no PII stored on-chain) to respond to suspicious patterns.

## Open Questions

- Should identity changes require multi-assistant consensus for high-sensitivity DIDs (e.g., gov IDs)?
- How do we handle partial revocations when a new DID is applied but old credentials still circulate?
- Do we need a ledger of `IdentityChangeAudits` for regulatory reporting, or are events + history queries sufficient?
