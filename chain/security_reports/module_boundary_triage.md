# Module Boundary Security Audit – Tooling Triage (2025-12-18)

This note captures the current static-analysis state for boundary-critical modules (bridge, dex, security, walletsecurity, validatorsecurity, contractregistry, wasm) and what remains to triage.

## Current Findings

- `gosec` (`gosec_module_boundary.json`):
  - 16 × `G115` remaining (int↔uint conversions): mostly key/ID helpers in bridge/dex/contractregistry/walletsecurity CLI and wasm hooks. See `gosec_module_boundary_g115.txt` for the exact list.
  - RIPEMD160 usage in bridge keeper documented/suppressed (Bitcoin/Cosmos address compatibility); `G507` cleared.
  - `G104` unchecked errors **resolved** in bridge keeper (setTransfer paths).
- `semgrep --config p/ci` (`semgrep_module_boundary.json`): 0 findings (ruleset too generic; Cosmos-specific rules still to add).

## Next Actions (tooling-focused)

1) Review `G115` sites and add explicit bounds checks where values originate from user input or external hashes; mark intentional deterministic conversions with comments/ignore directives where required for compatibility.
2) Decide on ripemd160 usage for bridge Merkle hashes; if protocol-bound, document and add inline justification; otherwise replace with SHA256 + compatibility notes.
3) Author a Cosmos-specific Semgrep policy for:
   - Unchecked keeper/module errors in Msg servers.
   - Missing signer/authority validation on Msg handlers.
   - Insecure randomness or time-based seeds crossing module boundaries.
   - Unsafe type assertions on protobuf/gogoproto types.
4) Re-run `gosec` and Semgrep after rule updates and record deltas here.

## Files of Interest

- Reports: `chain/security_reports/gosec_module_boundary.json`, `chain/security_reports/semgrep_module_boundary.json`
- Bridge keeper (ripemd160 + conversions): `chain/x/bridge/keeper/keeper.go`
