# Module & Accessory Security TODOs

This list captures current hardening gaps per module/accessory so future agents can continue making every seam airtight. Paths refer to repository-relative locations.

## chain/x/auth
- `keeper/msg_server.go` lacks multi-sig authority checks for some role operations; ensure every state change validates governance/authz ACLs.
- Add integration tests covering `client/cli` flows to confirm CLI arguments surface validation errors before tx submission.

## chain/x/bridge
- Wire fraud-proof Msg/Query handlers to the keeper’s new `ResolveFraudProof` logic (`keeper/keeper.go`); currently CLI can’t submit/resolve proofs.
- Implement withdrawal queueing + relayer stats enforcement with dedicated telemetry so paused withdraws are observable.
- Extend CLI/e2e tests to cover honest transfer, stalled attestation, and fraud-proof resolution paths.

## chain/x/compliance *(done but monitor)*
- Continue fuzzing the sanctions/AML filters—`keeper/alerts.go` still has TODOs for rate-limiting alert spam.

## chain/x/confidencescore
- Proto lacks reward metadata/timestamps; update `proto/aura/confidencescore/v1beta1` once downstream SDKs are ready.
- Ensure reward distribution emits telemetry similar to walletsecurity so payout spikes are visible (see `keeper/rewards.go`).

## chain/x/cryptography
- Keeper methods `RegisterCircuit`, `SubmitProof`, `VerifyProof` (lines 333-404) still return `ErrNotImplemented`; flesh them out with mock verifiers + state storage.
- Add Msg/Query services + CLI coverage once keeper logic exists.

## chain/x/dataregistry
- Finish authority checks + reward minting in `keeper/msg_server.go:100-220` to prevent unauthorized revocations and to pay verifiers.
- Emit events for every mutation so auditors can reconstruct histories.

## chain/x/dex
- Add a native `QuerySpotPrice` gRPC/proto endpoint so SDK clients can fetch prices without CLI parsing.
- Add telemetry/logging for validation failures on `Orderbook` / `Pool` queries similar to the new user-orders/market-price coverage.
- Build end-to-end CLI test harness (mini network) to exercise swaps + queries via `aurad` processes.

## chain/x/economicsecurity
- Review `keeper/params.go` for TODO validations; decimals should be range-checked similar to validatorsecurity.

## chain/x/governance
- Module still lacks `ModuleServices` logging on invalid queries; mirror the pattern used in DEX query server.

## chain/x/identitychange
- TODO: add unit tests for `client/cli` validators (currently only structure tests exist).

## chain/x/incidentresponse
- Implement Msg/Query servers + CLI (proto references exist but keepers still use TODO events).

## chain/x/inclusionroutines
- Revisit keeper hooks to ensure cross-module reentrancy checks exist when issuing incentives.

## chain/x/monitoring
- `metrics/` and `siem/` directories need implementation; connect emitted telemetry (wallet/dex/networksecurity) to alerting rules.

## chain/x/networksecurity
- Maintain coverage for gossip cache TTL edge cases; add property tests ensuring dedup eviction works under bursty traffic.

## chain/x/prevalidation
- Complete gRPC client descriptions (currently TODO) so pre-validation flows can be audited via CLI.

## chain/x/privacy
- Ensure shielded swap APIs include circuit validation once cryptography module lands; currently placeholders exist.

## chain/x/validatorsecurity
- Add logging similar to DEX for query validation failures (e.g., invalid validator address) so SOC can monitor misuse.

## chain/x/vcregistry
- Finish KV migration stage 4 by removing in-memory fallbacks (`keeper/query.go` TODOs). Ensure events + nonce storage are persisted.

## chain/x/walletsecurity
- Continue monitoring telemetry to ensure log volume is manageable; consider rate-limiting to prevent spam from compromised wallets.

## Accessories

### wallet/, wallet-tools/, verifier-portal/
- Directories remain empty; build CLI + toolchains promised in README (seed management, audits, verifier UI). Until populated, wallet onboarding can’t proceed.

### ai-assistant/
- Only README exists; implement on-chain registry + off-chain components per doc or remove references.

### docs/
- Update runbooks when new telemetry/logging sources are added (recent DEX validation logs are documented but keep others in sync).

