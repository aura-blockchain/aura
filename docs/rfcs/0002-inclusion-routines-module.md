# RFC-0002: Inclusion Routine Registry Module

- **Author(s):** Spec Team
- **Status:** Draft
- **Created:** 2025-11-11
- **Target Release:** Devnet

## Summary

Standardize Inclusion Routine (IR) definitions, scoring weights, locale tags, and governance-controlled lifecycle in an on-chain registry module referenced by the identity keeper and AI assistants.

## Motivation & Goals

- Single source of truth for IR metadata (score, PoI rewards, privacy tier).
- Governance-managed CRUD with versioning, enabling upgrades/retirements.
- Enforce prerequisite graphs and anti-abuse rate limits.

Non-goals: off-chain AI workflows, PoI treasury payouts (handled elsewhere).

## Detailed Design

- **State:**
  - `IRDefinition`: `{id, name, arena, score, poi_reward, locale_tags[], privacy_tier, version, metadata_hash}`
  - `IRPrerequisites`: adjacency list for dependency graph.
  - `IRRateLimits`: per-wallet/per-block caps.
- **Messages:**
  - `MsgCreateIR`, `MsgUpdateIR`, `MsgDeleteIR` (gated by governance authority).
  - `MsgSetIRPrereqs`, `MsgSetIRRateLimit`.
- **Events:** emitted on create/update/delete for off-chain sync.
- **Access:** read-only queries for wallet/assistant clients to fetch active IR set.

## Lifecycle State Machine

Reference flow: `docs/architecture/flows/ir-completion.puml` (SVG available for docs embeds).

| State | Description | Allowed Transitions |
| ----- | ----------- | ------------------- |
| `Draft` | Definition staged, not yet usable by assistants. | `Draft`, `Reviewing`, `Discarded` |
| `Reviewing` | Governance proposal under vote; definition immutable. | `Approved`, `Rejected` |
| `Approved` | Vote passed but activation height not reached yet. | `Active`, `Deprecated` |
| `Active` | Assistants/wallets allowed to run IR; rate limits enforced. | `Deprecated`, `Suspended` |
| `Suspended` | Temporarily disabled (bug/abuse) but kept for audit. | `Active`, `Deprecated`, `Retired` |
| `Deprecated` | Replaced by newer version; read-only for audits. | `Retired` |
| `Retired` | Fully removed from active set; history retained for compliance. | — |

Transition guards:
- `Draft -> Reviewing`: requires governance proposal with audit report + locale plan.
- `Reviewing -> Approved/Rejected`: based on proposal outcome.
- `Approved -> Active`: executes at `activation_height` once prerequisites & rate limits configured.
- `Active -> Suspended`: emergency governance hook; emits `IRSuspended` event and halts PoI rewards.
- `Deprecated -> Retired`: only after `sunset_height` reached (default 90 days) to allow outstanding verifications to settle.

State stored as `IRStatus` enum on each `IRDefinition`; events mirror changes for assistant caches.

## Parameters & Queries

- `max_ir_per_locale` (default 24) to avoid bloated lists in wallet UI.
- `default_rate_limit` (per wallet/per hour) until specialized overrides applied.
- `suspension_fee` to discourage frivolous suspension proposals.

Queries (gRPC/REST):
- `ListIRs(status_filter, locale, pagination)` returns filtered set plus version metadata.
- `GetIR(id)` detailed record + prerequisite hashes.
- `GetIRGraph()` returns full dependency DAG for assistants pre-fetching tasks.
- `GetRateLimit(id)` returns global + per-locale overrides for enforcement clients.

## Security / Privacy Considerations

- Metadata should not include any PII or raw instructions; link to off-chain templates via hashes/URIs.
- Governance proposals modifying IRs must include audit notes and locale impact statements.

## Validation Plan

- Unit tests for CRUD, version conflicts, prereq cycle detection.
- Fuzz tests on dependency graphs.
- Simulations for state transitions ensuring guards enforce activation/suspension rules.

## Backwards Compatibility

- Provide migration handler from genesis list to module-managed entries.

## Open Questions

- Should deactivated IRs remain queryable for historical audits?
- How to encode privacy tier taxonomy for regulators?
