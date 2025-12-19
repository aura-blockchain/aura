# Module Data Validation Review (2026-02-14)

Scope: boundary inputs across modules (denoms/amounts/addresses/params/byte payloads) with focus on gogoproto value semantics and shared validators.

Completed:
- Introduced shared jurisdiction validator in `chain/x/common/validation.ValidateJurisdictionCode` (ISO 3166-1 alpha-2 + subdivisions), with unit coverage.
- Compliance now reuses the shared validator, eliminating bespoke regex logic and aligning future jurisdiction checks across modules.
- Verified gogoproto non-nullable patterns remain enforced via `chain/docs/GOGOPROTO_TYPES.md` helpers; math.Int/Dec comparisons use `.Is*`/`.Equal` in boundary code.
- WASM upload path already enforces `MaxWasmCodeSize` via params and rejects empty payloads; dataregistry enforces `MaxStorageBytes` on content.
- Prevalidation now consults compliance keeper to reject sanctioned senders before execution (sanctions surfaced as `ErrSanctionedSender`).
- Walletsecurity now enforces active auth sessions (when auth keeper provided) before lock/unlock flows to ensure session-based authorization.

Follow-ups:
- Add byte-size guards for any user-supplied blob fields outside wasm/dataregistry (e.g., future VC payloads) using shared max constants.
- Extend shared validation to cover VC type identifiers and memo limits where applicable.
- Ensure future modules import the common validator instead of duplicating regex/state.
