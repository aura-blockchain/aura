# IBC Status for E2E Test Harness

The multi-chain helpers in `chain/tests/e2e/chain.go` were written for the long-term goal of running genuine cross-chain tests (via `SimulateIBCTransfer`, `WaitForRelayer`, etc.). However, the current Aura testnet intentionally keeps the real IBC subsystem disabled (see `chain/docs/IBC_STATUS.md`), so there is no active relayer or channel to exercise today.

## Current Behavior

- `SimulateIBCTransfer` and `WaitForRelayer` are stubs. By default they `t.Skip()` unless `AURA_E2E_ENABLE_IBC=1` is set.
- They do **not** create clients, connections, channels, or deliver packets yet.
- Any production IBC attempts on Aura still return `ErrIBCNotEnabled` until the PAW companion chain + Hermes relayer are deployed.

## Enabling Real IBC E2E Tests (Future Work)

1. Bring up both Aura and PAW chains using the scripts in `chain/testing/local/phase6`.
2. Follow `test_6.1_ibc_setup_guide.md` to create clients/connections/channels through Hermes.
3. Extend `SimulateIBCTransfer` to call into a running relayer (or Go relayer APIs) once those channels exist.
4. Update the tests in `ibc_e2e_test.go` to assert on real acknowledgements once the plumbing is in place.

Until the infrastructure above is online, the E2E suite keeps the IBC helpers as no-ops so the rest of the tests remain runnable.
