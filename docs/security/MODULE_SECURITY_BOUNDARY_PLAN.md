# Module Security Boundary Audit Plan

Goal: expand coverage beyond guard-level checks into cross-module message-flow fuzzing and adversarial scenarios, as called out in ROADMAP_PRODUCTION.md.

## Scope & Targets
- **High-risk modules:** authz/authn, bank, staking/slashing, governance, IBC (ICS20/ICS27), wasm, incidentresponse, economicsecurity, bridge/htlc paths.
- **Flows to combine:** staking → governance changes → authz grants; wasm contract calls invoking bank/staking; IBC transfers with fee/gov interplay; bridge/htlc state + slashing; incidentresponse + auth boundaries.

## Test Strategy
1) **Cross-module sequences:** table-driven Go tests that execute multi-step message sequences (e.g., grant authz → wasm executes bank send → governance param change → staking slash) and assert invariants (no unauthorized state changes, supply conservation, correct jailing/slashing).
2) **Property/fuzz harness:** use `testing/quick` or `rapid` to fuzz message ordering across modules with seeded RNG, checking invariants: total supply, validator power monotonicity (mod slashing), bank balance non-negative, authz scope respected, wasm module cannot bypass authz.
3) **Simulation hooks:** extend existing simapp/simulation tests to include adversarial ops across modules; add regression seeds for found issues.
4) **IBC/bridge edges:** fuzz packet ordering/timeouts, fee adjustments, and relayer misbehavior; ensure escrow/fee pools reconcile.

## Immediate Next Actions
- [ ] Inventory existing cross-module tests in `chain/` (simulation, keeper tests) and map gaps vs. above flows.
- [ ] Draft table-driven cross-module sequences in a new `_test.go` (e.g., `chain/app/cross_module_security_test.go`) covering authz+bank+wasm+gov and staking+slashing+incidentresponse.
- [ ] Add a `rapid`-based fuzzer skeleton to permute message orderings over a small actor set; seed and record regressions.
- [ ] Define invariants to assert per run: total supply conservation, no negative balances, jailed validators cannot sign/earn, authz scope enforcement, wasm cannot bypass bank/staking/gov checks, IBC escrow correctness.
- [ ] Document findings and seeds under `docs/security/` (update this plan with results/seeds as they appear).

## Deliverables
- New cross-module Go test files with deterministic sequences.
- Fuzz harness with reproducible seeds and CI-safe runtime limits.
- Short findings log + seeds in `docs/security/` once issues are uncovered/fixed.
