# Remaining Tests and Coverage Gaps

This project now builds and all Go tests pass in segmented runs:
```
cd chain
go test ./app/... ./cmd/... ./testing/... ./x/...
```
The suite is large (420+ Go test files), but several production-grade gaps remain. Use the items below to drive the next wave of test authoring. Each item is written so another agent can pick it up directly.

## P0 – Security & Consensus-Critical Gaps
- `chain/x/networksecurity` – Add deeper invariant coverage beyond current cases:
  - Peer reputation bounding under heavy churn (fuzz preferred), ban list persistence across blocks, connection throttling edge cases.
  - Inject malformed gossip/fork alerts to ensure invariants catch empty IDs, negative heights, and out-of-range scores.
  - Command: `cd chain && go test ./x/networksecurity/...`.

- `chain/x/economicsecurity` (root + params) – Keeper has light tests only.
  - Add invariant and params validation tests for security budgets, stake weights, and threshold schemes.
  - Include fuzz tests on economic thresholds to prevent overflow/underflow on large stakes.
  - Command: `cd chain && go test ./x/economicsecurity/... -run Invariant`.

## P1 – Protocol Surface & CLI Coverage
- CLI packages with `[no test files]`: `x/governance/client/cli`, `x/identity/client/cli`, `x/incidentresponse/client/cli`, `x/contractregistry/client/cli`, `x/networksecurity/client/cli`, `x/prevalidation/client/cli`, `x/wasm/client/cli`, `x/walletsecurity/client/cli`, `x/vcregistry/params`, `x/economicsecurity/params`, `x/dataregistry/params`.
  - Add cobra command tests using in-memory contexts to assert flag validation, required args, default fees/gas, and generated messages.
  - Verify sign/broadcast path wiring, especially for governance proposal flows and security pause/guard commands.
  - Command: `cd chain && go test ./x/<module>/client/cli -run Test*CLI`.

- Module roots lacking coverage (`x/auth`, `x/confidencescore`, `x/contractregistry`, `x/cryptography`, `x/dataregistry`, `x/dex`, `x/economics`, `x/governance`, `x/identity`, `x/incidentresponse`, `x/monitoring`, `x/prevalidation`, `x/security`, `x/validatorsecurity`, `x/walletsecurity`, `x/vcregistry`):
  - Add module-level genesis round-trip tests (Init/Export/Validate), Params validation tests, and invariants where applicable.
  - For `x/cryptography`, add tests around key registration/rotation and failure cases for invalid schemes.
  - For `x/monitoring`, add coverage for alerting/metrics wiring beyond `alerting` and `ml` subpackages (test SIEM and params serialization).

- `chain/cmd/aurad` – No tests for top-level command wiring.
  - Add smoke tests for `aurad start --home <tmpdir> --rpc.laddr ...` in dry-run mode (no sockets) to validate flag parsing and configuration loading.

## P1 – Cross-Module Integration & Invariants
- Add multi-module integration tests under `chain/testing/integration` to cover:
  - Bridge ↔ Governance: proposal to pause bridge, ensure messages blocked while paused.
  - DEX ↔ Security: reentrancy guard prevents nested swaps when pause guard is active.
  - Compliance ↔ Prevalidation: AML rule violations reject txs before mempool admit.
  - Include state machine invariants for token conservation and bridge transfer counters after restarts.

- Add fuzz/invariant suites:
  - DEX AMM math fuzz (slippage bounds, rounding) under `x/dex/keeper`.
  - Bridge signature verification fuzz for malformed payloads and replay windows.
  - Walletsecurity rate-limiting invariant: no account exceeds configured tx caps across blocks.

## P2 – SDKs, Frontends, and Tooling
- SDKs: go/js/python unit suites now green. Still add golden tests for Tx/Query builders, signing, amino vs proto JSON, and REST/GRPC client mocks.
  - Add E2E tests against in-process `aurad` (simapp) to verify broadcast/query flows.

- Wallets (`wallet/desktop`, `wallet/mobile`, `wallet/browser-extension`, `wallet/web`) – minimal or zero automated tests.
  - Add unit tests for key management flows, signing, and offline/online switching.
  - Add Cypress/Playwright smoke tests for web/extension; add Detox/React Native Testing Library for mobile.

- Explorer (`explorer/`) – almost no test coverage.
  - Add API decoder tests (`explorer/tx_decoder.py`) for all custom Msgs (bridge, dex, aiassistant, wasm security) and pagination handling.
  - Add frontend component snapshot and integration tests for block/tx detail views.

- Smart contracts (`contracts/*`) – no contract-level tests present (only artifacts).
  - Add Rust unit tests and CosmWasm integration tests for `binding-tester` and `vc-issuer`, covering instantiate/execute/query and failure paths.
  - Command: `cd contracts/vc-issuer && cargo test`.

## P3 – Observability & DevEx
- `chain/testing/coverage` & `chain/testing/fuzz` exist but are light. Add:
  - Coverage gate script to fail if module coverage <90% (per-module profiles).
  - Additional fuzz targets for cryptography key parsing and wasm message decoding.
  - Command: `cd chain/testing/fuzz && go test -run Fuzz`.

- Add race detection CI-equivalent scripts (local): `go test -race ./...` broken into batches to avoid timeouts.

## How to Execute
1. Pick a bullet, implement tests, and run the indicated command(s).
2. Keep runs short by module (e.g., `go test ./x/security/...`), not `./...`.
3. When adding fuzz/invariant tests, gate them with build tags if they are slow (`//go:build fuzz`).
4. Record coverage improvements and new commands in this file after each batch.

### Progress Log
- 2026-02-13: Added economicsecurity params store validation tests, param invariant corruption cases, and fuzz coverage for staking incentive thresholds; command: `cd chain && go test ./x/economicsecurity/...`.
- 2026-02-13: Expanded networksecurity invariants with fuzzed reputation bounds, ban persistence, connection count stress, and malformed alert coverage; command: `cd chain && go test ./x/networksecurity/...`.
- 2026-02-13: Added prevalidation batch/scheduler lifecycle tests, security privacy module CRUD/validation coverage, and common determinism/gasmetering deterministic behavior tests; commands: `cd chain && go test ./x/prevalidation/... ./x/security/... ./x/common/...`.
- 2026-02-13: Added prevalidation rate-limit, signature validation, and nonce bounds tests; command: `cd chain && go test ./x/prevalidation/...`.
- 2026-02-13: Added governance CLI command registration and argument parsing tests (tx/query, category/vote parsing, weighted options); command: `cd chain && go test ./x/governance/...`.
- 2026-02-13: Added CLI structure/argument tests for identity, incidentresponse, contractregistry, and wasm modules; commands: `cd chain && go test ./x/identity/... ./x/incidentresponse/... ./x/contractregistry/... ./x/wasm/...`.
- 2026-02-13: Added params store validation/immutability tests for dataregistry and vcregistry; command: `cd chain && go test ./x/dataregistry/... ./x/vcregistry/...`.
- 2026-02-13: Added walletsecurity CLI registration and arg validation (tx/query) and ensured checksum/enclave validation paths are covered; command: `cd chain && go test ./x/walletsecurity/...`.
- 2026-02-13: Added aurad root command bootstrap/config tests to verify flags, defaults, and help execution wiring; command: `cd chain && go test ./cmd/aurad/...`.
- 2026-02-13: Added networksecurity bandwidth tracker regression coverage and economicsecurity params fuzz bounds; commands: `cd chain && go test ./x/networksecurity/... ./x/economicsecurity/...`.
- 2026-02-13: Added DEX AMM GetQuote fuzz coverage for slippage/fee bounds; command: `cd chain && go test ./x/dex/...`.
