# Test 6.1: IBC Testing Guide (Aura <-> PAW)

## Overview

This document provides comprehensive guidance for testing IBC (Inter-Blockchain Communication) functionality between Aura and PAW chains. While a full end-to-end IBC test requires both chains to be running simultaneously, this guide documents the requirements, setup process, and expected test scenarios.

## Prerequisites

### Infrastructure Requirements

1. **Two Running Chains**
   - Aura testnet (currently running on ports 27657-27660)
   - PAW testnet (requires separate initialization)

2. **IBC Relayer**
   - Hermes v1.13.2 (installed at `/usr/local/bin/hermes`)
   - Configuration file: `~/.hermes/config.toml`

3. **Network Configuration**
   - No port conflicts between chains
   - Aura: RPC :27657, gRPC :10090
   - PAW: RPC :26657, gRPC :9090 (recommended)

### Software Versions

- **Cosmos SDK**: v0.50.x
- **IBC Go**: v8.x
- **Hermes**: 1.13.2

## Test 6.1.1: Hermes Relayer Configuration

### Hermes Config Template

```toml
[global]
log_level = 'info'

[mode]

[mode.clients]
enabled = true
refresh = true
misbehaviour = true

[mode.connections]
enabled = true

[mode.channels]
enabled = true

[mode.packets]
enabled = true
clear_interval = 100
clear_on_start = true
tx_confirmation = true

[rest]
enabled = true
host = '127.0.0.1'
port = 3000

[telemetry]
enabled = true
host = '127.0.0.1'
port = 3001

[[chains]]
id = 'aura-testnet-1'
rpc_addr = 'http://127.0.0.1:27657'
grpc_addr = 'http://127.0.0.1:10090'
event_source = { mode = 'push', url = 'ws://127.0.0.1:27657/websocket', batch_delay = '500ms' }
rpc_timeout = '10s'
account_prefix = 'aura'
key_name = 'aura-relayer'
store_prefix = 'ibc'
default_gas = 100000
max_gas = 400000
gas_price = { price = 0.025, denom = 'uaura' }
gas_multiplier = 1.1
max_msg_num = 30
max_tx_size = 180000
clock_drift = '5s'
max_block_time = '30s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
address_type = { derivation = 'cosmos' }

[[chains]]
id = 'paw-testnet-1'
rpc_addr = 'http://127.0.0.1:26657'
grpc_addr = 'http://127.0.0.1:9090'
event_source = { mode = 'push', url = 'ws://127.0.0.1:26657/websocket', batch_delay = '500ms' }
rpc_timeout = '10s'
account_prefix = 'paw'
key_name = 'paw-relayer'
store_prefix = 'ibc'
default_gas = 100000
max_gas = 400000
gas_price = { price = 0.025, denom = 'upaw' }
gas_multiplier = 1.1
max_msg_num = 30
max_tx_size = 180000
clock_drift = '5s'
max_block_time = '30s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
address_type = { derivation = 'cosmos' }
```

### Setup Commands

```bash
# 1. Create relayer keys for both chains
hermes keys add --chain aura-testnet-1 --mnemonic-file ~/.hermes/aura-relayer.mnemonic
hermes keys add --chain paw-testnet-1 --mnemonic-file ~/.hermes/paw-relayer.mnemonic

# 2. Fund relayer accounts (requires faucet or manual transfer)
# Aura:
aurad tx bank send <faucet-addr> <aura-relayer-addr> 10000000uaura --chain-id aura-testnet-1

# PAW:
pawd tx bank send <faucet-addr> <paw-relayer-addr> 10000000upaw --chain-id paw-testnet-1

# 3. Verify configuration
hermes config validate

# 4. Health check
hermes health-check
```

## Test 6.1.2: IBC Connection Creation

### Expected Process

1. **Create IBC Client**
   ```bash
   hermes create client --host-chain aura-testnet-1 --reference-chain paw-testnet-1
   hermes create client --host-chain paw-testnet-1 --reference-chain aura-testnet-1
   ```

   **Expected Output:**
   - Client ID: `07-tendermint-0` (on each chain)
   - Success message confirming client creation

2. **Create IBC Connection**
   ```bash
   hermes create connection \
     --a-chain aura-testnet-1 \
     --b-chain paw-testnet-1
   ```

   **Expected Output:**
   - Connection ID: `connection-0` (on both chains)
   - State: OPEN

3. **Verify Connection**
   ```bash
   hermes query connections --chain aura-testnet-1
   hermes query connections --chain paw-testnet-1
   ```

## Test 6.1.3: IBC Channel Creation

### Transfer Channel Setup

```bash
# Create transfer channel
hermes create channel \
  --a-chain aura-testnet-1 \
  --a-connection connection-0 \
  --a-port transfer \
  --b-port transfer
```

**Expected Output:**
- Channel ID: `channel-0` (on both chains)
- Port: `transfer`
- State: OPEN
- Ordering: UNORDERED
- Version: `ics20-1`

### Channel Verification

```bash
# Query channels
hermes query channels --chain aura-testnet-1
hermes query channels --chain paw-testnet-1

# Query specific channel
aurad query ibc channel end transfer channel-0 --node http://localhost:27657
pawd query ibc channel end transfer channel-0 --node http://localhost:26657
```

## Test 6.1.4: Token Transfer Testing

### Test 1: Aura → PAW Transfer

```bash
# Transfer 1000uaura from Aura to PAW
aurad tx ibc-transfer transfer \
  transfer \
  channel-0 \
  <paw-recipient-addr> \
  1000uaura \
  --from <aura-sender> \
  --chain-id aura-testnet-1 \
  --node http://localhost:27657 \
  --timeout-height-offset 1000
```

**Expected Results:**
1. Transaction succeeds on Aura
2. Packet sent event emitted
3. Relayer picks up packet
4. Packet acknowledged on PAW
5. PAW recipient receives IBC tokens:
   - Denom: `ibc/<hash>` (where hash = SHA256(transfer/channel-0/uaura))
   - Amount: 1000

**Verification:**
```bash
# On PAW chain
pawd query bank balances <paw-recipient-addr> --node http://localhost:26657

# Should show: ibc/<hash> with amount 1000
```

### Test 2: PAW → Aura Transfer

```bash
# Transfer 500upaw from PAW to Aura
pawd tx ibc-transfer transfer \
  transfer \
  channel-0 \
  <aura-recipient-addr> \
  500upaw \
  --from <paw-sender> \
  --chain-id paw-testnet-1 \
  --node http://localhost:26657 \
  --timeout-height-offset 1000
```

**Expected Results:**
Similar to Test 1, but in reverse direction.

### Test 3: Token Redemption (Return Path)

```bash
# Send IBC tokens back to original chain
pawd tx ibc-transfer transfer \
  transfer \
  channel-0 \
  <original-aura-addr> \
  1000ibc/<hash> \
  --from <paw-holder> \
  --chain-id paw-testnet-1
```

**Expected Results:**
- IBC tokens burned on PAW
- Original uaura tokens released on Aura
- Recipient receives 1000uaura (original denom)

## Test 6.1.5: Timeout and Error Scenarios

### Test 1: Timeout on Send

```bash
# Send with very short timeout
aurad tx ibc-transfer transfer \
  transfer \
  channel-0 \
  <paw-addr> \
  100uaura \
  --from <aura-sender> \
  --timeout-height-offset 1 \
  --chain-id aura-testnet-1
```

**Expected Results:**
- Packet times out
- Timeout packet processed by relayer
- Tokens refunded to sender on Aura

### Test 2: Channel Closure

```bash
# Close channel (requires governance or special conditions)
# This is an edge case test
```

**Expected Results:**
- Pending packets should be handled
- No new packets accepted
- Existing tokens can still be redeemed

## Test 6.1.6: Relayer Failure and Recovery

### Test 1: Relayer Restart

```bash
# Stop relayer
pkill hermes

# Wait for packets to accumulate (send several transfers)

# Restart relayer
hermes start
```

**Expected Results:**
- Relayer catches up on missed packets
- All pending packets processed
- No packet loss

### Test 2: Relayer with One Chain Down

**Scenario:**
1. Stop PAW chain
2. Send transfer from Aura
3. Restart PAW chain

**Expected Results:**
- Packet remains pending while PAW is down
- Packet delivered when PAW restarts
- OR packet times out and refunded

## Test 6.1.7: Advanced Scenarios

### Multi-Hop Transfer

While not directly tested, document the capability:
- Aura → PAW → Another Chain
- Requires multiple channels and connections
- Token denomination becomes nested: `ibc/<hash1>/ibc/<hash2>/uaura`

### High-Volume Testing

```bash
# Send many transfers in parallel
for i in {1..100}; do
  aurad tx ibc-transfer transfer transfer channel-0 \
    <addr> 10uaura --from sender --yes &
done
```

**Metrics to Monitor:**
- Packet delivery rate
- Average confirmation time
- Relayer performance
- Gas costs

## Implementation Status

### Current Status: ✓ INFRASTRUCTURE READY

1. **Aura Chain**
   - ✓ IBC module enabled
   - ✓ Transfer module available
   - ✓ Running and producing blocks

2. **PAW Chain**
   - ✓ Binary available (`/home/hudson/blockchain-projects/paw/pawd`)
   - ⚠ Needs initialization for testnet
   - ⚠ Needs port configuration to avoid conflicts

3. **Hermes Relayer**
   - ✓ Installed and available
   - ⚠ Configuration needed
   - ⚠ Requires funded relayer accounts

### Setup Time Estimate

- PAW testnet initialization: 30 minutes
- Hermes configuration: 15 minutes
- IBC connection setup: 10 minutes
- Testing and validation: 60 minutes
- **Total: ~2 hours**

## Automated Test Script (Future Implementation)

```bash
#!/bin/bash
# test_6.1_ibc_full.sh
# Automated IBC testing between Aura and PAW

# 1. Verify both chains running
# 2. Configure Hermes
# 3. Create clients, connections, channels
# 4. Execute test transfers
# 5. Verify balances
# 6. Test timeout scenarios
# 7. Test relayer recovery
# 8. Generate test report
```

## Monitoring and Debugging

### Useful Commands

```bash
# Relayer logs
hermes -c ~/.hermes/config.toml start --debug

# Query packet commitments
aurad query ibc channel packet-commitments transfer channel-0

# Query packet acknowledgments
aurad query ibc channel packet-acks transfer channel-0

# Query next sequence
aurad query ibc channel next-sequence-receive transfer channel-0

# Check client state
aurad query ibc client state 07-tendermint-0

# Check connection state
aurad query ibc connection end connection-0
```

### Common Issues and Solutions

1. **"Client expired"**
   - Solution: Update client with `hermes update client`

2. **"Channel not found"**
   - Solution: Verify channel was created successfully
   - Check connection state

3. **"Packet timeout"**
   - Solution: Increase timeout height/timestamp
   - Verify both chains are producing blocks

4. **"Insufficient fees"**
   - Solution: Fund relayer accounts
   - Adjust gas_price in Hermes config

## Success Criteria

### Test 6.1: Overall PASS if:

- ✓ Hermes successfully configured for both chains
- ✓ IBC clients created on both chains
- ✓ IBC connection established (state: OPEN)
- ✓ Transfer channel created (state: OPEN)
- ✓ Token transfer Aura → PAW successful
- ✓ Token transfer PAW → Aura successful
- ✓ IBC token denom correctly formatted
- ✓ Token redemption works (return path)
- ✓ Timeout handling works correctly
- ✓ Relayer can recover from restart

## Conclusion

IBC functionality between Aura and PAW requires coordinated setup of both chains and the Hermes relayer. While the infrastructure is ready and the IBC modules are properly configured, full end-to-end testing requires:

1. PAW testnet deployment
2. Hermes configuration with funded accounts
3. Sequential execution of connection/channel creation
4. Systematic testing of all scenarios

The Aura chain is IBC-ready and has been tested in isolation. The PAW chain binary exists and can be deployed. The primary blockers are operational (chain deployment and relayer setup) rather than technical.

**Recommendation:** Defer full IBC testing to integration testing phase when both chains can be deployed simultaneously for extended testing periods.
