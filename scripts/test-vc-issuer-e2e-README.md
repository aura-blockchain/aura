# VC-Issuer E2E Test Script

**File**: `/home/decri/blockchain-projects/aura/scripts/test-vc-issuer-e2e.sh`

## Overview

End-to-end test script for the `vc-issuer` WASM contract that:
- Spins up an ephemeral local Aura node
- Uploads the optimized vc-issuer WASM contract
- Instantiates the contract
- Executes the full VC issuance flow (register issuer → request VC → fulfill VC)
- Validates all operations with preflight/postflight checks
- Measures gas consumption for each operation
- Preserves all logs and state for debugging

## Prerequisites

1. **aurad binary**: Built and available in PATH or set via `AURA_BINARY`
2. **Contract artifact**: `contracts/artifacts/vc_issuer.wasm` must exist
3. **System utilities**: `jq`, `curl`, `base64`
4. **Python 3**: For client.toml manipulation (standard on most systems)
5. **Running node**: The script starts its own node automatically

## Usage

### Basic Usage

```bash
cd /home/decri/blockchain-projects/aura
./scripts/test-vc-issuer-e2e.sh
```

The script will:
1. Create a temporary home directory
2. Initialize a fresh chain with 3 accounts (validator, issuer, subject)
3. Start aurad in the background
4. Execute the full E2E flow
5. Preserve all logs for inspection
6. Print the temp directory location at completion

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `AURA_BINARY` | Path to aurad binary | `aurad` |
| `AURA_CHAIN_ID` | Chain ID for test | `aura-vc-e2e-1` |
| `AURA_DENOM` | Token denomination | `uaura` |
| `AURA_GAS_PRICES` | Gas prices | `0.025uaura` |
| `AURA_RPC_PORT` | RPC port (auto-selected if not set) | Random port 26657-46657 |
| `AURA_P2P_PORT` | P2P port (auto-selected if not set) | Random port 26656-46656 |
| `AURA_API_PORT` | API port (auto-selected if not set) | Random port 1317-21317 |
| `AURA_GRPC_PORT` | gRPC port (auto-selected if not set) | Random port 19090-39090 |
| `AURA_GRPC_ENABLE` | Enable gRPC server | `true` |
| `CLEANUP_VC_E2E` | Delete temp directory on exit | Not set (preserve logs) |

### Examples

**Use custom binary and chain ID:**
```bash
AURA_BINARY=/path/to/aurad AURA_CHAIN_ID=test-1 ./scripts/test-vc-issuer-e2e.sh
```

**Specify ports explicitly:**
```bash
AURA_RPC_PORT=26657 AURA_API_PORT=1317 ./scripts/test-vc-issuer-e2e.sh
```

**Enable cleanup (delete temp directory):**
```bash
CLEANUP_VC_E2E=1 ./scripts/test-vc-issuer-e2e.sh
```

## Outputs

### Temp Directory Structure

After running, the temp directory (printed at exit) contains:

```
/tmp/aura-vc-e2e.XXXXXX/
├── config/
│   ├── genesis.json           # Chain genesis
│   ├── config.toml            # Node configuration
│   ├── app.toml               # Application configuration
│   └── client.toml            # Client configuration (amino-json signing)
├── data/                      # Blockchain state
├── keyring-test/              # Test keyring (validator, issuer, subject keys)
├── aurad.log                  # Node logs (stdout/stderr)
├── tx_operations.log          # Transaction event log with timestamps
├── gas_measurements.log       # Gas usage per operation
├── store_unsigned.json        # Unsigned store transaction
└── store_signed.json          # Signed store transaction
```

### Log Files

#### `aurad.log`
Complete node output including:
- Consensus logs
- Block production
- Transaction execution
- Module logs

#### `tx_operations.log`
Timestamped transaction flow:
```
[2024-12-05 10:23:15] === Starting E2E Test Flow ===
[2024-12-05 10:23:15] Loading initial account state from running node...
[2024-12-05 10:23:15] Loaded validator via CLI: account_number=0, sequence=0
[2024-12-05 10:23:16] === STORE CONTRACT ===
[2024-12-05 10:23:16] ✓ Account state validated for validator: account_number=0, sequence=0
[2024-12-05 10:23:16] Generating unsigned store transaction...
[2024-12-05 10:23:16] Signing transaction with amino-json mode...
[2024-12-05 10:23:16] Signed transaction uses mode: SIGN_MODE_LEGACY_AMINO_JSON
[2024-12-05 10:23:16] Broadcasting signed transaction...
[2024-12-05 10:23:16] Waiting for tx store (ABC123...)...
[2024-12-05 10:23:18] ✓ tx store succeeded at height 42
[2024-12-05 10:23:18] ✓ Account state validated for validator: account_number=0, sequence=1
[2024-12-05 10:23:18] ✓ Stored contract with code_id=1
...
```

#### `gas_measurements.log`
Gas usage by operation:
```
store: gas_used=2847329, gas_wanted=5000000
instantiate: gas_used=143256, gas_wanted=5000000
register-issuer: gas_used=89421, gas_wanted=5000000
request-vc: gas_used=102349, gas_wanted=5000000
fulfill-vc: gas_used=118734, gas_wanted=5000000
```

## Validation Features

### Preflight Checks (Before Transactions)

Before each transaction, the script:
1. Queries current account state from the node
2. Validates account_number matches expected value
3. Validates sequence matches expected value
4. Fails fast if mismatch detected

**Example validation:**
```bash
# Before validator's second transaction
validate_account_state validator 0 1  # Expect account 0, sequence 1
```

### Postflight Checks (After Transactions)

After each transaction confirms, the script:
1. Re-queries account state from the node
2. Validates sequence incremented by exactly 1
3. Fails fast if sequence didn't increment
4. Tracks cumulative sequence for next transaction

**Example validation:**
```bash
# After validator's second transaction
validate_account_state validator 0 2  # Expect account 0, sequence 2
```

### Benefits

- **Catches stale state**: Prevents using outdated account/sequence values
- **Detects signing issues**: Identifies transactions that don't execute
- **Fail-fast**: Stops immediately on first validation failure
- **Clear diagnostics**: Logs exact mismatch values for debugging

## Signing Mode Fix

### The Problem

Original script used `--sign-mode legacy-amino-json` which caused authentication
rejection in some SDK versions:

```
Error: rpc error: code = InvalidArgument desc = failed to execute message;
message index: 0: failed to verify signature: unauthorized
```

### The Solution

1. **Use `amino-json` instead of `legacy-amino-json`**:
   ```bash
   --sign-mode amino-json
   ```

2. **Explicit sign flow for store transaction**:
   ```bash
   # Generate unsigned transaction
   aurad tx aura_wasm_security store contract.wasm --generate-only > unsigned.json

   # Sign with explicit mode
   aurad tx sign unsigned.json --sign-mode amino-json > signed.json

   # Broadcast signed transaction
   aurad tx broadcast signed.json
   ```

3. **Add signing mode to run_tx_and_wait()** for other transactions:
   ```bash
   run_tx_and_wait() {
     # ...
     res=$("$@" "${TX_FLAGS[@]}" --sign-mode amino-json 2>&1)
     # ...
   }
   ```

4. **Verify signing mode in transaction**:
   ```bash
   jq -r '.auth_info.signer_infos[0].mode_info.single.mode' signed.json
   # Should output: SIGN_MODE_LEGACY_AMINO_JSON
   ```

### Why This Works

- `amino-json` is the canonical signing mode name in recent Cosmos SDK versions
- `legacy-amino-json` may not be recognized by all SDK versions
- Explicit sign + broadcast flow gives more control and better error messages
- Verification step catches signing mode mismatches early

## Gas Measurement

The script automatically tracks gas usage for all operations:

### Baseline Metrics

After running the script, check `gas_measurements.log` for:

| Operation | Typical Gas | Notes |
|-----------|-------------|-------|
| `store` | ~2.8M | Large WASM binary upload |
| `instantiate` | ~140K | Contract initialization |
| `register-issuer` | ~90K | State write + validation |
| `request-vc` | ~100K | Create request record |
| `fulfill-vc` | ~120K | Process and store credential |

### Performance Regression Detection

1. Run script to establish baseline: `./scripts/test-vc-issuer-e2e.sh`
2. Save gas measurements: `cp /tmp/aura-vc-e2e.XXXXXX/gas_measurements.log baseline_gas.log`
3. After code changes, run again and compare
4. Investigate if gas usage increases significantly (>10%)

### Optimization Tips

- **Store operation**: Gas is proportional to WASM size - optimize with `wasm-opt`
- **Execute operations**: Reduce state reads/writes in contract logic
- **Event emission**: Excessive events increase gas - emit only essential data

## Troubleshooting

### Script Fails Immediately

**Symptom**: Script exits with "missing required command" error

**Solution**: Install missing tools:
```bash
# Debian/Ubuntu
sudo apt-get install jq curl

# macOS
brew install jq curl
```

### Port Already in Use

**Symptom**: "selected ports are already in use" error

**Solution**: Either:
1. Stop the conflicting node
2. Specify different ports:
   ```bash
   AURA_RPC_PORT=36657 AURA_P2P_PORT=36656 ./scripts/test-vc-issuer-e2e.sh
   ```

### Node Fails to Start

**Symptom**: "aurad exited immediately; see /tmp/aura-vc-e2e.XXXXXX/aurad.log"

**Solution**:
1. Check the log file path printed by the script
2. Common issues:
   - Invalid genesis file
   - Corrupted state database
   - Missing modules in binary

### Contract Not Found

**Symptom**: "artifact not found: contracts/artifacts/vc_issuer.wasm"

**Solution**: Build and optimize the contract:
```bash
cd /home/decri/blockchain-projects/aura/contracts/vc-issuer
cargo wasm
cd /home/decri/blockchain-projects/aura/contracts
make optimize-wasm
```

### Signing Mode Mismatch

**Symptom**: "signed store tx did not use LEGACY_AMINO_JSON" warning

**Solution**: This is now logged as WARNING (not error). The script continues and
attempts broadcast. If broadcast fails, check:
1. aurad version supports amino-json signing
2. client.toml has correct sign-mode setting
3. Binary was built with auth module properly configured

### Transaction Fails with Code != 0

**Symptom**: "tx failed with code 5: insufficient funds"

**Solution**: The script automatically funds accounts from genesis. If this fails:
1. Check genesis.json in temp directory
2. Verify accounts have balances in bank module
3. Increase default allocation in `add_key()` function

### Sequence Mismatch

**Symptom**: "Postflight validation failed - sequence did not increment correctly"

**Solution**: This indicates a transaction didn't execute. Check:
1. Transaction was actually included in a block (check tx_operations.log)
2. Transaction didn't fail (code=0 in response)
3. Account state query is timing out or returning stale data

### Query Returns Empty Data

**Symptom**: "Query returned empty data" for credentials query

**Solution**: The contract state may not be persisted. Check:
1. Fulfill transaction succeeded (check tx hash and logs)
2. Block containing fulfill tx was committed (check node height)
3. Query is using correct contract address
4. Contract logic is storing credentials correctly

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: WASM E2E Tests

on: [push, pull_request]

jobs:
  e2e-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y jq curl

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Build aurad
        run: |
          cd chain
          go build -o aurad ./cmd/aurad
          echo "$(pwd)" >> $GITHUB_PATH

      - name: Setup Rust
        uses: actions-rust-lang/setup-rust-toolchain@v1
        with:
          target: wasm32-unknown-unknown

      - name: Build and optimize contracts
        run: |
          cd contracts/vc-issuer
          cargo wasm
          cd ..
          make optimize-wasm

      - name: Run E2E test
        run: ./scripts/test-vc-issuer-e2e.sh
        env:
          AURA_RPC_PORT: 26657
          AURA_API_PORT: 1317
          CLEANUP_VC_E2E: "1"

      - name: Check gas baseline
        run: |
          # Parse gas measurements and fail if over threshold
          grep "store:" /tmp/aura-vc-e2e.*/gas_measurements.log | \
            awk '{print $2}' | sed 's/gas_used=//' | \
            awk '$1 > 3000000 {exit 1}'
```

### Local Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Run E2E test before allowing commit to WASM contract code
if git diff --cached --name-only | grep -q "^contracts/"; then
    echo "WASM contract modified, running E2E test..."
    if ! ./scripts/test-vc-issuer-e2e.sh; then
        echo "E2E test failed! Commit aborted."
        exit 1
    fi
fi
```

## Advanced Usage

### Custom Gas Limits

Modify `TX_FLAGS` in the script:
```bash
TX_FLAGS=(--chain-id "${CHAIN_ID}" ... --gas 10000000)  # 10M gas
```

### Debug Signing Issues

Set verbose logging in the script:
```bash
set -x  # Enable bash trace mode
```

Then check what command-line flags are actually passed to aurad.

### Parallel Test Runs

Run multiple instances with different ports:
```bash
# Terminal 1
AURA_RPC_PORT=26657 AURA_API_PORT=1317 ./scripts/test-vc-issuer-e2e.sh &

# Terminal 2
AURA_RPC_PORT=36657 AURA_API_PORT=2317 ./scripts/test-vc-issuer-e2e.sh &
```

### Custom Contract Initialization

Modify the `INIT_MSG` construction:
```bash
INIT_MSG=$(jq -n --arg admin "${VALIDATOR_ADDR}" \
  '{admin: $admin, extra_config: "custom_value"}')
```

## Security Considerations

### Test Keyring

The script uses `--keyring-backend test` which stores keys **unencrypted**. This is
acceptable for E2E tests but:

- **Never use test keyring in production**
- **Never commit test keys to git**
- **Temp directories are world-readable** - don't run on shared systems with sensitive data

### Ephemeral Chain

Each run creates a fresh chain that is destroyed (unless CLEANUP_VC_E2E not set).
This means:

- **No state persists between runs** - good for test isolation
- **Keys are regenerated each time** - addresses will differ
- **Genesis accounts have large balances** - unrealistic for mainnet

### Port Exposure

The script binds to `127.0.0.1` (localhost only) by default. If you modify to bind
to `0.0.0.0`:

- **Firewall rules apply** - ensure test ports not exposed to internet
- **No authentication** - anyone who can reach the port can transact
- **Use only in trusted networks**

## Performance Tuning

### Faster Block Times

Edit the script to modify `config.toml` after init:
```bash
sed -i.bak 's/timeout_commit = "5s"/timeout_commit = "1s"/' "${CONFIG_TOML}"
```

This reduces time waiting for transactions to confirm.

### Disable Unnecessary Modules

If testing doesn't require all 27 Aura modules, comment out non-essential modules
in the genesis to reduce chain startup time.

### Increase Gas Limits

If transactions run out of gas:
```bash
TX_FLAGS=(... --gas 10000000)  # Increase from default 5M
```

Or use `--gas auto` with a multiplier:
```bash
TX_FLAGS=(... --gas auto --gas-adjustment 1.5)
```

## Related Documentation

- **WASM Module**: `/home/decri/blockchain-projects/aura/chain/x/wasm/README.md`
- **Contract Source**: `/home/decri/blockchain-projects/aura/contracts/vc-issuer/`
- **Signing Verification Report**: `/home/decri/blockchain-projects/aura/chain/WASM_SIGNING_VERIFICATION_REPORT.md`
- **Roadmap**: `/home/decri/blockchain-projects/aura/ROADMAP_PRODUCTION.md`

## Contributing

When modifying this script:

1. **Test thoroughly** - Run at least 3 times to ensure consistency
2. **Preserve backward compatibility** - Use environment variables for new config
3. **Update this README** - Document new features and options
4. **Maintain fail-fast behavior** - Validate aggressively, don't mask errors
5. **Log extensively** - Future maintainers will thank you

## License

This script is part of the Aura blockchain project. See repository root for license.
