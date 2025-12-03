# MockBankKeeper Balance Tracking

## Critical Implementation Detail

The `MockBankKeeper.SendCoinsFromAccountToModule` method MUST deduct balances when balance tracking is enabled for an address.

### Why Balance Deduction is Required

1. **Commit-Reveal Tests**: Tests like `TestBatchExecution_FailedLocks` verify that orders fail when users have insufficient funds. Without deduction, all orders appear to have sufficient balance.

2. **Realistic Simulation**: Real BankKeeper deducts balances on transfer. The mock must behave the same way for tests to be meaningful.

3. **Batch Execution**: When multiple orders compete for the same funds, deduction ensures only properly funded orders succeed.

### Balance Tracking Modes

- **Tracking Enabled**: If an address has any balance set (`len(m.balances[addr]) > 0`), then check AND deduct on sends
- **Permissive Mode**: If an address has never been funded, allow unlimited sends (backward compatibility for tests that don't explicitly set up balances)

### DO NOT MODIFY THIS LOGIC

The linter/formatter may try to "optimize" by removing the deduction. This breaks tests. The deduction is intentional and required.
