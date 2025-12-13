# Phase 3.4: Malicious Peer Handling

## Overview

This document describes expected behavior and testing procedures for handling malicious peers in the Aura blockchain network. Tendermint Core includes built-in peer management and security features to detect and ban misbehaving nodes.

## Threat Model

### Types of Malicious Behavior

1. **Invalid Block Proposals**
   - Blocks with invalid signatures
   - Blocks with incorrect state transitions
   - Blocks with invalid transactions
   - Blocks that don't match the block hash

2. **Invalid Consensus Messages**
   - Malformed prevote/precommit messages
   - Messages with invalid signatures
   - Messages for wrong height/round
   - Duplicate messages (double-voting attempts)

3. **Invalid Transactions**
   - Transactions with invalid signatures
   - Malformed transaction data
   - Transactions that fail validation
   - Replay attacks

4. **P2P Protocol Violations**
   - Excessive message flooding
   - Invalid message formats
   - Protocol version mismatches
   - Handshake failures

5. **Data Withholding**
   - Not responding to block requests
   - Not forwarding transactions
   - Ignoring peer discovery requests

## Tendermint Security Features

### 1. Peer Scoring System

Tendermint maintains a reputation score for each peer:
- **Starting Score**: New peers start with a neutral score
- **Good Behavior**: +1 for successful block/tx relay
- **Bad Behavior**: -10 to -100 depending on severity
- **Ban Threshold**: Peers with score < -100 are banned

### 2. Automatic Peer Banning

Peers are banned for:
- Sending invalid blocks (immediate ban)
- Sending malformed consensus messages (immediate ban)
- Repeated protocol violations (cumulative)
- Excessive message rate (flooding)

Ban duration: 24 hours by default (configurable)

### 3. Message Validation

Every incoming message is validated:
```
Receive Message
    ↓
Verify Message Format
    ↓
Verify Signature
    ↓
Verify Message Context (height, round, etc.)
    ↓
Process or Reject
    ↓
Update Peer Score
```

### 4. Rate Limiting

Protects against DoS attacks:
- Maximum messages per second per peer
- Maximum bytes per second per peer
- Maximum concurrent connections

## Expected Behavior

### Scenario 1: Invalid Block Received

**Action**: Peer sends block with invalid signature

**Expected Response**:
1. Validator detects invalid signature
2. Block is rejected immediately
3. Error logged: `"Received invalid block from peer X"`
4. Peer reputation decreases by -100
5. Peer is banned from validator's peer list
6. Connection to peer is closed

**Result**: Network unaffected, malicious peer isolated

### Scenario 2: Invalid Consensus Message

**Action**: Peer sends prevote with forged signature

**Expected Response**:
1. Message fails signature verification
2. Message is discarded
3. Error logged: `"Invalid signature on prevote from peer X"`
4. Peer reputation decreases
5. Peer may be banned if repeated

**Result**: Consensus continues without disruption

### Scenario 3: Transaction Flooding

**Action**: Peer sends 10,000 transactions rapidly

**Expected Response**:
1. Rate limiter detects excessive messages
2. Excess transactions are dropped
3. Warning logged: `"Peer X exceeding rate limit"`
4. Peer is temporarily throttled or disconnected
5. Mempool remains functional

**Result**: Network performance unaffected

### Scenario 4: Double-Sign Attempt

**Action**: Validator tries to sign two different blocks at same height

**Expected Response**:
1. Evidence of double-signing is detected
2. Evidence is gossiped to network
3. Malicious validator is slashed (stake reduced 5%)
4. Malicious validator is jailed (removed from active set)
5. Evidence is recorded in blockchain

**Result**: Byzantine behavior punished, network security maintained

### Scenario 5: Invalid State Transition

**Action**: Peer proposes block with invalid state change

**Expected Response**:
1. Block proposal validation fails
2. Block is rejected before consensus voting
3. Error logged with specific validation failure
4. Peer reputation decreases
5. Network continues to next round with different proposer

**Result**: Invalid state never reaches consensus

### Scenario 6: Network Partition Attack

**Action**: Malicious node tries to split network

**Expected Response**:
1. Nodes detect conflicting block hashes
2. Consensus voting reveals the partition
3. Nodes on minority side (< 1/3 voting power) halt
4. Nodes on majority side (> 2/3 voting power) continue
5. Minority nodes sync to majority chain when partition resolves

**Result**: Network maintains safety (no double-spend)

## Configuration Options

### config.toml Settings

```toml
[p2p]
# Maximum number of inbound peers
max_num_inbound_peers = 40

# Maximum number of outbound peers
max_num_outbound_peers = 10

# Peer exchange enabled (for discovery)
pex = true

# Seed mode (for seed nodes only)
seed_mode = false

# Private peer IDs (won't be gossiped)
private_peer_ids = ""

# Unconditional peer IDs (never disconnected)
unconditional_peer_ids = ""

[mempool]
# Maximum number of transactions in mempool
size = 5000

# Maximum size of a single transaction
max_tx_bytes = 1048576

# Maximum total size of mempool
max_txs_bytes = 1073741824

# Broadcast enabled
broadcast = true

# Cache size for rejected transactions
cache_size = 10000
```

### app.toml Settings

```toml
[mempool]
# Maximum gas wanted per transaction
max-tx-gas-wanted = 0

# Maximum number of transactions in mempool (override)
max-txs = -1
```

## Logging and Monitoring

### Key Log Messages

**Invalid Block**:
```
ERR Received invalid block module=blockchain err="invalid signature" peer=X
```

**Invalid Transaction**:
```
ERR Failed to add tx to mempool err="invalid signature" tx=Y
```

**Peer Banned**:
```
INF Banning peer module=p2p peer=X reason="invalid block"
```

**Double-Sign Evidence**:
```
WRN Received evidence of double-signing validator=Z height=N
```

**Rate Limit Exceeded**:
```
WRN Peer exceeding rate limit module=p2p peer=X rate=N/s
```

### Monitoring Metrics

**Prometheus Metrics** (exposed on `:26660/metrics`):

```
# Peer count
tendermint_p2p_peers{peer_id="..."}

# Messages received
tendermint_p2p_message_receive_bytes_total{message_type="..."}

# Messages sent
tendermint_p2p_message_send_bytes_total{message_type="..."}

# Consensus state
tendermint_consensus_height
tendermint_consensus_validators
tendermint_consensus_missing_validators

# Mempool
tendermint_mempool_size
tendermint_mempool_tx_size_bytes
```

## Testing Methodology

### Manual Testing (Not Recommended)

Testing malicious peer behavior manually is complex and requires:
- Custom Tendermint client to send invalid messages
- Deep understanding of Tendermint P2P protocol
- Risk of disrupting testnet

### Automated Testing (Recommended)

Use Tendermint's built-in test suite:

```bash
# Test invalid block handling
go test -v ./consensus -run TestInvalidBlock

# Test invalid vote handling
go test -v ./consensus -run TestInvalidVote

# Test peer banning
go test -v ./p2p -run TestBanPeer
```

### Chaos Engineering Approach

Use toxiproxy to simulate Byzantine behavior:

```bash
# Add toxiproxy for validator-1
docker-compose -f docker-compose.toxiproxy.yml up -d

# Create proxy
toxiproxy-cli -h http://localhost:10800 create val1-consensus \
    -l 0.0.0.0:26657 -u aura-validator-1:26657

# Add toxic: corrupt data
toxiproxy-cli -h http://localhost:10800 toxic add val1-consensus \
    -t slicer -a average_size=100 -a size_variation=10
```

### Observational Testing

Monitor existing testnet for natural occurrences:

1. **Check logs for peer bans**:
   ```bash
   docker logs aura-validator-1 2>&1 | grep -i "ban"
   ```

2. **Check for invalid transactions**:
   ```bash
   docker logs aura-validator-1 2>&1 | grep -i "invalid"
   ```

3. **Monitor peer scores** (if exposed via RPC)

## Verification Checklist

- [ ] Logs show rejected invalid blocks
- [ ] Logs show rejected invalid transactions
- [ ] Malicious peers are banned (appear in ban list)
- [ ] Network consensus continues despite malicious activity
- [ ] No invalid data enters the blockchain
- [ ] Peer count decreases when malicious peers are banned
- [ ] Metrics show message rejection rate
- [ ] Evidence of double-signing is properly recorded

## Known Limitations

### Current Testnet (Single Validator)

With only 1 active validator, Byzantine behavior testing is limited:

**Cannot Test**:
- Multi-validator consensus attacks
- 51% attacks
- Double-signing detection (requires 2+ validators)
- Byzantine fault tolerance thresholds

**Can Test**:
- Invalid transaction rejection
- Malformed message handling
- P2P protocol violations
- Rate limiting
- Peer banning for protocol violations

### Future Enhancements

For comprehensive malicious peer testing, need:

1. **Multi-Validator Setup**: 4+ validators with equal voting power
2. **Chaos Engineering Tools**: Automated Byzantine behavior injection
3. **Custom Test Harness**: Scripts to generate invalid messages
4. **Monitoring Dashboard**: Real-time peer reputation visualization

## Security Recommendations

### For Production

1. **Sentry Node Architecture**
   - Validators should NOT have public IPs
   - Validators connect only to trusted sentry nodes
   - Sentry nodes handle untrusted P2P traffic

2. **Firewall Rules**
   - Restrict P2P port (26656) to known peers
   - Use `private_peer_ids` for validator interconnection
   - Use `unconditional_peer_ids` for critical peers

3. **Rate Limiting**
   - Configure conservative rate limits
   - Monitor rate limit violations
   - Auto-ban on repeated violations

4. **Monitoring**
   - Alert on peer bans
   - Alert on consensus timeouts
   - Alert on unusual message patterns
   - Track peer reputation over time

5. **Regular Audits**
   - Review peer list regularly
   - Analyze ban logs
   - Test failover scenarios
   - Update peer discovery seeds

## Example: Checking Peer Status

### View Current Peers

```bash
curl -s http://localhost:26657/net_info | jq '.result.peers[] | {node_id: .node_info.id, moniker: .node_info.moniker, remote_ip: .remote_ip}'
```

### Check Peer Connection Info

```bash
curl -s http://localhost:26657/net_info | jq '{
  listening: .result.listening,
  n_peers: .result.n_peers,
  peers: [.result.peers[] | {
    moniker: .node_info.moniker,
    channels: .node_info.channels
  }]
}'
```

### Monitor Consensus State

```bash
curl -s http://localhost:26657/consensus_state | jq '{
  height: .result.round_state.height,
  round: .result.round_state.round,
  step: .result.round_state.step,
  validators: .result.round_state.validators
}'
```

## Incident Response

### If Malicious Activity Detected

1. **Immediate**:
   - Identify the malicious peer ID
   - Verify peer is banned
   - Check if any invalid data entered chain

2. **Short-term**:
   - Add peer to permanent ban list
   - Update firewall rules if needed
   - Notify other operators

3. **Long-term**:
   - Analyze attack vector
   - Update security policies
   - Improve monitoring
   - Share findings with community

### Manual Peer Ban

```bash
# Get peer ID from logs or net_info
PEER_ID="abc123..."

# Add to persistent peers config (ban)
# Edit config.toml:
# [p2p]
# persistent_peers = ""
# private_peer_ids = "$PEER_ID"

# Restart node
docker restart aura-validator-1
```

## References

- [Tendermint P2P Documentation](https://docs.tendermint.com/v0.35/tendermint-core/p2p.html)
- [Tendermint Security Best Practices](https://docs.tendermint.com/v0.35/tendermint-core/validators.html)
- [Cosmos SDK Security](https://docs.cosmos.network/main/core/security)
- [Byzantine Fault Tolerance](https://en.wikipedia.org/wiki/Byzantine_fault)

## Conclusion

The Aura blockchain, built on Tendermint Core, includes robust mechanisms for handling malicious peers. While comprehensive testing requires a multi-validator setup, the security features are well-established and battle-tested in production Cosmos networks.

**Key Takeaway**: Trust but verify. Monitor peer behavior, maintain up-to-date banlists, and use sentry architecture for production validators.

## Next Steps

- Phase 4: Security & Attack Simulation (requires multi-validator setup)
- Configure multi-validator testnet for comprehensive Byzantine testing
- Implement automated chaos engineering tests
- Set up advanced monitoring and alerting
