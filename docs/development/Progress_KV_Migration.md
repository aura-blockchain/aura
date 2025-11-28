# VC Registry KV Migration Progress

## Plan
- Stage 1: Introduce KV-enabled revocation storage (records, Merkle root, proofs) with optional fallback to in-memory for legacy wiring; keep API stable for callers while adding context-aware persistence hooks.
- Stage 2: Move VC records, policies, and DID documents to KV with codec-backed marshaling; expose typed getters/setters that operate on KV first; migrate genesis import/export to KV formats.
- Stage 3: Move mint counts, presentations, attribute VCs, disclosure policies/requests/responses, and user indices to KV; ensure pagination and proofs derive from committed state.
- Stage 4: Remove in-memory fallbacks, require KV wiring (store key + codec) in keeper construction, and align module wiring/tests with full store-backed state.

## Current Status
- Proto/GenesisState already expanded (regenerated via `buf generate`) to include presentations, user presentation index, attribute VCs, user attribute index, disclosure policies/requests/responses, and pending disclosure index.
- Keeper now uses real structs for attribute/disclosure state. Attribute VC creation enforces required fields (type, holder, ciphertext/hash), default status/expiry sanity, and singleton per attribute type per holder. Disclosure policy validation enforces default mode + unique typed rules. Disclosure requests check expiry bounds; responses verify request existence, pending status, expiry, and disclosed attributes subset.
- Keeper now generates deterministic IDs for attribute VCs and disclosure requests; pending disclosure index uses stable `holder:request` keys with iterate support.
- Message server implements create/revoke AttributeVC, update disclosure policy, create/respond disclosure request with KV-first persistence. Query server implements get/list attribute VCs, get disclosure policy, and get disclosure request.
- Genesis init/export extended to include presentations, presentation/user indices, attribute VCs/user indices, disclosure policies/requests/responses, and pending disclosures with KV+map fallback helpers. Validation covers uniqueness and cross-references for these new collections.
- Presentation creation now records into KV or in-memory maps and retrieval supports KV/map fallback.
- Mint rate limiting now reads/writes KV-backed counters whenever an SDK context is available, while retaining the in-memory fallback for legacy/unit test flows.

-## Work Completed
- Wired the module manager / CosmosApp BeginBlock path to deliver SDK contexts so `CleanupOldMintCounts` runs against the KV store and block metadata reaches the keeper.
- Added `SyncContextMetadata` hooks to the keeper and refreshed msg server flows to call it before state mutations, consolidating timestamp sourcing.
- Added regression coverage for the new BeginBlock wiring and msg server behaviors, including mint freshening, attribute/disclosure lifecycle syncing, and pending index cleanup under real SDK contexts.

## Next Steps
- Extend context propagation through the remaining message/server flows so every keeper timestamp, rate-limit counter, and request index leans on the SDK block metadata we now surface.
- Grow integration coverage (mint eligibility edge cases, attribute/disclosure indexing, DID/event sync) before removing the in-memory fallbacks in Stage 4.
- Document any outstanding TODOs (events, nonce storage, full policy enforcement) and triage follow-ups once the KV wiring is stable.
 - Begin the next migration sprint by tackling the individual pieces we outlined earlier: msg/query handlers, keeper validation, genesis inclusion, and end-to-end tests for the attribute/disclosure/presentation flows.
> Heads up: current work cycle is approaching the rate-limit window—plan to continue the remaining items once the next run starts.
