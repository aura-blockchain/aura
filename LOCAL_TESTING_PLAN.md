# Aura Project - Local Testing Plan (v4 - Definitive Edition)

This document is the definitive and most exhaustive local testing plan for the Aura project. It includes standard, advanced, and esoteric test cases to ensure maximum stability, security, and robustness. **This is the final version.**

## Phase 1: Primitives & Static Analysis

*   **[ ] 1.1: Linter and Static Analysis:** `make lint`
*   **[ ] 1.2: Unit Tests:** `go test ./...`
*   **[ ] 1.3: Integration Tests:** `go test -tags=integration ./...`
*   **[ ] 1.4: Verify Crypto Primitives:** Write an integration test that performs a sample signing and verification using the project's crypto libraries to ensure they are correctly linked and configured.
*   **[ ] 1.5: Verify Encoding Primitives:** Write an integration test to marshal and unmarshal key data structures (like `Block`, `Transaction`) to and from their binary/JSON representations to catch serialization bugs.

## Phase 2: Single-Node Lifecycle & Configuration

*   **[ ] 2.1: Genesis & Initialization:** Verify `aurad init`, `add-genesis-account`, `gentx`, `collect-gentxs`, and `validate-genesis` all work as expected.
*   **[ ] 2.2: Exhaustive Configuration Testing:**
    *   **Description:** For every parameter in `config.toml` and `app.toml`, modify it from the default and verify the node's behavior changes as expected.
    *   **Action:** Script the process of changing a single parameter (e.g., `p2p.max_num_inbound_peers`), restarting the node, and verifying the change (e.g., `aurad status` or by trying to connect more peers than the new limit).
    *   **Expected Outcome:** Node correctly respects all configuration values or fails gracefully on invalid ones.
*   **[ ] 2.3: CLI Command Verification:**
    *   **Description:** Test every single CLI command and subcommand provided by `aurad`, including all queries and transactions.
    *   **Action:** Write a script that iterates through `aurad --help`, `aurad query --help`, `aurad tx --help`, etc., and executes each command with valid and invalid parameters.
    *   **Expected Outcome:** All commands behave as documented. Errors for invalid usage are clear and actionable.

## Phase 3: Multi-Node Network & Consensus

*   **[ ] 3.1: 4-Node Network Baseline:** `./launch-testnet.sh`
*   **[ ] 3.2: Consensus Liveness & Halt:** Test 4-node, 3-node (live), and 2-node (halt) configurations.
*   **[ ] 3.3: Network Variable Latency/Bandwidth:** Use `tc` to simulate poor network conditions and test consensus stability.
*   **[ ] 3.4: Malicious Peer Ejection:** Test if a node bans a peer that sends invalid blocks, transactions, or consensus messages.

## Phase 4: Comprehensive Security & Attack Simulation

*   **[ ] 4.1: Smart Contract Security Analysis:** `cargo audit` and manual review of all custom contracts for common flaws (reentrancy, overflows, access control).
*   **[ ] 4.2: 51% Re-org Attack:** Simulate a majority partition building a longer chain to test fork-choice logic.
*   **[ ] 4.3: Validator Double-Sign Slashing:** Force a validator to double-sign and verify it gets jailed and slashed.
*   **[ ] 4.4: Validator Downtime Slashing:** Take a validator offline for the `signed_blocks_window` and verify it gets jailed and slashed.
*   **[ ] 4.5: RPC Endpoint Hardening & Fuzzing:**
    *   **Description:** Fuzz test all public-facing RPC and API endpoints with malformed requests and unexpected data types.
    *   **Action:** Use a fuzzing tool (e.g., custom Python script, `go-fuzz`) to send garbage/malicious data to `localhost:26657` and `localhost:1317`.
    *   **Expected Outcome:** The node does not crash and logs errors correctly.
*   **[ ] 4.6: Governance Exploit Scenarios:** Test submission of malicious proposals, voting power concentration, and vote spamming.

## Phase 5: Advanced State, Economics & Upgrades

*   **[ ] 5.1: State Snapshot & Restore:** Test a new node's ability to bootstrap from a state snapshot.
*   **[ ] 5.2: State Pruning:** Verify old state is correctly removed when pruning is enabled.
*   **[ ] 5.3: Staking & Rewards Logic:** Programmatically verify that staking rewards over N epochs match the expected calculations.
*   **[ ] 5.4: Fee Market Dynamics:** Test transaction acceptance/rejection based on `min-gas-prices`.
*   **[ ] 5.5: On-Chain Software Upgrade:** Test the full governance-based software upgrade process, including the chain halt, binary swap, and successful restart.
*   **[ ] 5.6: State Migration:** As part of the software upgrade, verify any custom state migration logic runs correctly.

## Phase 6: Cross-Chain Interoperability

*   **[ ] 6.1: IBC (Aura <-> Paw):** Setup a Hermes relayer and test token transfers, channel creation/closing, and relayer failure/restart scenarios.
*   **[ ] 6.2: Atomic Swaps (Aura <-> BTC):** If the module exists, test successful swaps and failed/refunded swaps with a local `bitcoind` in `regtest` mode.

## Phase 7: Destructive & Long-Running Tests

*   **[ ] 7.1: Database Corruption Test:**
    *   **Description:** Test the node's ability to detect and recover from on-disk database corruption.
    *   **Action:** While a node is stopped, use `dd` or a hex editor to write garbage data into a random location in its `data/application.db` or `data/state.db`.
    *   **Expected Outcome:** Upon restart, the node should fail with a clear "database is corrupt" error and not enter an undefined state. Test recovery procedures, if they exist.
*   **[ ] 7.2: Resource Constraint Test:**
    *   **Description:** Test node performance and stability under heavy resource constraints.
    *   **Action:** Run a node container with restricted resources (`--memory="512m"`, `--cpus="0.5"`). Observe its ability to sync and process blocks.
    *   **Expected Outcome:** The node should run slower but remain stable. This helps define minimum system requirements.
*   **[ ] 7.3: Long-Running Stability (Soak Test):**
    *   **Description:** Run the 4-node testnet under a moderate, continuous, and varied load for an extended period (24-48 hours).
    *   **Action:** Use a script to continuously send a mix of bank transfers, smart contract interactions, and governance proposals. Monitor with Prometheus.
    *   **Expected Outcome:** The network remains stable with no memory leaks, performance degradation, or unexpected halts.

This v4 plan represents the full scope of local testing that can be performed.
---

# Test Plan Gap Analysis - aura

Generated: Sat Dec 13 07:36:50 AM UTC 2025
Source: LOCAL_TESTING_PLAN.md

## Identified Gaps

### Missing Essential Tests

- [ ] **Encoding/Serialization**: Not found in current test plan
- [ ] **Consensus Testing**: Not found in current test plan
- [ ] **Security Testing**: Not found in current test plan
- [ ] **Slashing Tests**: Not found in current test plan
- [ ] **RPC Endpoint Testing**: Not found in current test plan
- [ ] **State Management**: Not found in current test plan
- [ ] **Economic Testing**: Not found in current test plan
- [ ] **Upgrade Testing**: Not found in current test plan
- [ ] **Cross-Chain/IBC**: Not found in current test plan
- [ ] **Database Testing**: Not found in current test plan
- [ ] **Destructive Tests**: Not found in current test plan

### Missing Advanced Tests

- [ ] **Load Testing**: Consider adding this test category
- [ ] **Performance Profiling**: Consider adding this test category
- [ ] **Memory Leak Detection**: Consider adding this test category
- [ ] **Byzantine Behavior**: Consider adding this test category
- [ ] **Double-Spend Prevention**: Consider adding this test category
- [ ] **Timestamp Validation**: Consider adding this test category
- [ ] **Fee Market Testing**: Consider adding this test category
- [ ] **State Snapshots**: Consider adding this test category
- [ ] **Replay Protection**: Consider adding this test category
- [ ] **Nonce Management**: Consider adding this test category
- [ ] **Gas Optimization**: Consider adding this test category
- [ ] **Oracle Testing**: Consider adding this test category
- [ ] **DEX Testing**: Consider adding this test category
- [ ] **Governance Testing**: Consider adding this test category
- [ ] **Chain Reorganization**: Consider adding this test category
- [ ] **Orphan Blocks**: Consider adding this test category

## Recommendations

### Infrastructure Tooling

The following tools have been created to support local testing:

- `scripts/load-tests/` - Load testing with k6
- `scripts/testnet-scenarios.sh` - Multi-node test scenarios
- `scripts/snapshot-manager.sh` - State snapshot management
- `scripts/network-sim.sh` - Network condition simulation
- `scripts/profile-*.sh` - Performance profiling tools
- `scripts/db-benchmark.sh` - Database benchmarking

### Test Coverage Improvements

1. **Fuzzing**: Add property-based and fuzzing tests for critical paths
2. **Load Testing**: Implement realistic load scenarios with k6
3. **Chaos Engineering**: Use testnet-scenarios.sh for failure testing
4. **Performance Regression**: Set up automated profiling
5. **Security Scanning**: Integrate static analysis and vulnerability scanning

### Next Steps

1. Review missing essential tests and add to test plan
2. Implement infrastructure-assisted tests using new tooling
3. Set up automated test execution for CI/CD
4. Document test procedures and expected outcomes
5. Create test data generators for realistic scenarios

