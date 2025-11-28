# Network Security Module

The Network Security module provides comprehensive network-level security features for the Aura blockchain, protecting against various attack vectors and ensuring network resilience.

## Features

### 1. DDoS Protection with Rate Limiting

**Implementation**: `keeper/rate_limiter.go`

- **Token Bucket Algorithm**: Implements rate limiting using token bucket algorithm
- **Per-Peer Rate Limiting**: Each peer has individual rate limits
- **Burst Control**: Configurable burst size for temporary traffic spikes
- **Bandwidth Throttling**: Limits bandwidth usage per peer
- **Automatic Banning**: Peers exceeding limits are automatically banned

**Configuration**:
```go
RateLimitConfig{
    MaxRequestsPerSecond:  100,
    BurstSize:             200,
    WindowDuration:        time.Minute,
    BanDuration:           time.Hour,
    BandwidthLimitPerPeer: 10 * 1024 * 1024, // 10 MB/s
}
```

**Key Functions**:
- `CheckRateLimit(ctx, peerID)`: Validates if peer can make request
- `CheckBandwidthLimit(ctx, peerID, bytes)`: Checks bandwidth usage
- `DDosProtectionCheck(ctx, peerID, messageSize)`: Comprehensive DDoS check

### 2. Sybil Resistance

**Implementation**: `keeper/sybil_eclipse.go`

Detects and prevents Sybil attacks through:

- **IP Subnet Analysis**: Detects concentration from same /24 subnets (>30%)
- **ASN Distribution**: Monitors Autonomous System Number diversity (>40% threshold)
- **Geographic Distribution**: Tracks regional concentration (>50% threshold)
- **Minimum Diversity**: Requires at least 3 unique ASNs with >5 peers

**Key Functions**:
- `CheckSybilResistance(ctx)`: Analyzes peer distribution
- `AnalyzePeerDistribution(peers)`: Detects suspicious patterns
- `CalculatePeerDiversity(ctx)`: Computes diversity score

### 3. Eclipse Attack Prevention

**Implementation**: `keeper/sybil_eclipse.go`

Prevents Eclipse attacks through:

- **Trusted Peer Verification**: Ensures connections to trusted peers
- **ASN Diversity Checks**: Requires minimum diverse ASNs
- **IP Concentration Limits**: Max 25% from single /24 subnet
- **Connection Type Balance**: Prevents all-outbound connections
- **Peering Restrictions**: Configurable trusted-peers-only mode

**Configuration**:
```go
ConnectionConfig{
    MaxInboundConnections:  100,
    MaxOutboundConnections: 50,
    MaxConnectionsPerIp:    10,
    TrustedPeersOnly:       false,
    MinPeerDiversity:       5,
}
```

**Key Functions**:
- `CheckEclipseAttack(ctx)`: Detects Eclipse attack patterns
- `ValidateNewConnection(ctx, peerInfo)`: Validates new peer connections
- `PerformPeerDiversityCheck(ctx)`: Ensures sufficient diversity

### 4. Connection Management

**Implementation**: `keeper/sybil_eclipse.go`, `keeper/keeper.go`

- **Connection Limits**: Per-direction (inbound/outbound) limits
- **Per-IP Limits**: Prevents multiple connections from same IP
- **Connection Timeout**: Automatic timeout for stale connections
- **Reputation-Based Filtering**: Low-reputation peers are rejected
- **IP Validation**: Rejects private/loopback addresses from internet

**Key Functions**:
- `AcceptConnection(ctx, peerInfo)`: Accepts validated connection
- `DisconnectPeer(ctx, peerID)`: Handles disconnection
- `GetConnectionCount(ctx, ipAddress)`: Tracks connections per IP

### 5. Mempool Security

**Implementation**: `keeper/mempool.go`

- **Size Limits**: Maximum transaction count and byte size
- **Priority Fee Mechanism**: Prevents spam with minimum fees
- **Per-Account Limits**: Maximum transactions per account
- **Eviction Policies**: Configurable (lowest_fee, oldest, random)
- **Anti-Spam Protection**: Reputation-based transaction filtering

**Configuration**:
```go
MempoolConfig{
    MaxSize:            10000,
    MaxBytes:           100 * 1024 * 1024, // 100 MB
    MinPriorityFee:     math.NewInt(1000),
    MaxTxsPerAccount:   100,
    EvictionPolicy:     "lowest_fee",
    EnablePriorityFees: true,
}
```

**Key Functions**:
- `ValidateTransaction(ctx, tx, txBytes, sender)`: Validates mempool admission
- `MempoolSpamProtection(ctx, tx, txBytes, sender)`: Comprehensive spam protection
- `CalculatePriorityScore(ctx, tx, txSize)`: Computes transaction priority

### 6. Gossip Protocol Validation

**Implementation**: `keeper/gossip.go`

- **Message Signature Verification**: Validates sender signatures
- **Size Limits**: Enforces maximum message size
- **TTL Enforcement**: Rejects expired messages
- **Redundancy Filtering**: Deduplicates messages via cache
- **Malicious Packet Filtering**: Detects and blocks malformed data

**Configuration**:
```go
GossipConfig{
    VerifySignatures:       true,
    MaxMessageSize:         1024 * 1024, // 1 MB
    MessageTtl:             time.Minute * 5,
    EnableRedundancyFilter: true,
    MaxFanout:              8,
}
```

**Key Functions**:
- `ValidateGossipMessage(ctx, msg)`: Validates gossip messages
- `PacketFilter(ctx, packetData, senderID)`: Filters malicious packets
- `PropagateMessage(ctx, msg)`: Determines propagation strategy

### 7. Fork Detection and Alerting

**Implementation**: `keeper/fork_partition.go`

- **Block Hash Tracking**: Detects different hashes at same height
- **Automatic Alerts**: Creates fork alerts when detected
- **Confirmation Depth**: Configurable confirmation requirements
- **Auto-Resolution**: Optional automatic fork resolution

**Configuration**:
```go
ForkDetectionConfig{
    EnableDetection:      true,
    HeightDiffThreshold:  10,
    EnableAutoResolution: false,
    ConfirmationDepth:    6,
}
```

**Key Functions**:
- `DetectFork(ctx, height, blockHash)`: Detects blockchain forks
- `ResolveFork(ctx, alertID)`: Resolves fork alerts
- `GetAllForkAlerts(ctx, includeResolved)`: Retrieves fork alerts

### 8. Sync Attack Prevention

**Implementation**: `keeper/fork_partition.go`

- **Block Data Validation**: Validates sync data integrity
- **Hash Verification**: Ensures block hashes match data
- **Suspicious Pattern Detection**: Identifies malicious sync peers
- **Invalid Block Tracking**: Monitors invalid block count per peer

**Key Functions**:
- `ValidateSyncData(ctx, peerID, height, hash, data)`: Validates sync data
- `RecordInvalidBlock(ctx, peerID)`: Records invalid blocks
- `IsSuspiciousSyncPeer(ctx, peerID)`: Detects suspicious behavior

### 9. Node Reputation Tracking

**Implementation**: `keeper/reputation.go`

- **Score-Based System**: Reputation scores (0 to max_score)
- **Behavior Tracking**: Monitors valid/invalid messages
- **Automatic Rewards/Penalties**: Updates based on behavior
- **Reputation Decay**: Gradual score decay over time
- **Uptime Tracking**: Monitors connection stability

**Configuration**:
```go
ReputationConfig{
    EnableTracking:     true,
    InitialScore:       100,
    MinScoreToConnect:  0,
    DecayRate:          1,
    MaxScore:           1000,
    MisbehaviorPenalty: 50,
    GoodBehaviorReward: 1,
}
```

**Key Functions**:
- `UpdateReputation(ctx, peerID, score, reason)`: Updates reputation
- `PenalizeReputation(ctx, peerID, penalty)`: Penalizes misbehavior
- `RewardReputation(ctx, peerID, reward)`: Rewards good behavior
- `GetHealthyPeers(ctx)`: Returns peers above threshold

### 10. Network Partitioning Detection

**Implementation**: `keeper/fork_partition.go`

- **Peer Count Monitoring**: Tracks connected peer count
- **Expected Peer Tracking**: Maintains running average
- **Sudden Drop Detection**: Alerts on >50% peer drop
- **Missing Peer Identification**: Identifies disconnected peers
- **Automatic Alerts**: Creates partition alerts

**Configuration**:
```go
PartitionDetectionConfig{
    EnableDetection:    true,
    MinConnectedPeers:  10,
    CheckInterval:      time.Minute,
    PartitionThreshold: 5,
}
```

**Key Functions**:
- `DetectPartition(ctx)`: Checks for network partition
- `GetMissingPeerIDs(ctx)`: Identifies missing peers
- `PerformNetworkHealthCheck(ctx)`: Comprehensive health check

## ABCI Integration

**File**: `abci.go`

The module integrates with Cosmos SDK's ABCI lifecycle:

### BeginBlocker
Runs every block to:
- Perform network health checks
- Decay reputation scores (every 100 blocks)
- Cleanup expired rate limits (every 50 blocks)
- Cleanup message cache (every 200 blocks)
- Update peer uptimes
- Check mempool health
- Cleanup resolved alerts (every 1000 blocks)
- Prune low reputation peers (every 500 blocks)

### EndBlocker
Runs at block end to:
- Cleanup mempool if needed

## Query Endpoints

### Queries
- `GET /aura/networksecurity/v1beta1/params` - Get module parameters
- `GET /aura/networksecurity/v1beta1/peer/{peer_id}` - Get peer info
- `GET /aura/networksecurity/v1beta1/peers` - Get all peers
- `GET /aura/networksecurity/v1beta1/trusted_peers` - Get trusted peers
- `GET /aura/networksecurity/v1beta1/reputation/{peer_id}` - Get peer reputation
- `GET /aura/networksecurity/v1beta1/ratelimit/{peer_id}` - Get rate limit status
- `GET /aura/networksecurity/v1beta1/mempool/stats` - Get mempool stats
- `GET /aura/networksecurity/v1beta1/fork_alerts` - Get fork alerts
- `GET /aura/networksecurity/v1beta1/partition_alerts` - Get partition alerts
- `GET /aura/networksecurity/v1beta1/health` - Get network health

## Transaction Messages

### Messages
- `MsgUpdateParams` - Update module parameters (governance)
- `MsgAddTrustedPeer` - Add a trusted peer
- `MsgRemoveTrustedPeer` - Remove a trusted peer
- `MsgBanPeer` - Manually ban a peer
- `MsgUnbanPeer` - Unban a peer
- `MsgUpdatePeerReputation` - Manually update reputation
- `MsgResolveForkAlert` - Resolve a fork alert
- `MsgResolvePartitionAlert` - Resolve a partition alert

## Error Codes

| Code | Error | Description |
|------|-------|-------------|
| 1 | ErrInvalidPeerID | Invalid peer ID provided |
| 2 | ErrPeerNotFound | Peer not found in registry |
| 3 | ErrPeerBanned | Peer is currently banned |
| 4 | ErrRateLimitExceeded | Rate limit exceeded |
| 5 | ErrConnectionLimitExceeded | Connection limit exceeded |
| 20 | ErrSybilDetected | Potential Sybil attack detected |
| 21 | ErrEclipseDetected | Potential Eclipse attack detected |
| 22 | ErrSyncAttack | Sync attack detected |

## Testing

Comprehensive test suite in `keeper/keeper_test.go`:

- Parameter management tests
- Peer info management tests
- Trusted peer management tests
- Ban/unban functionality tests
- Reputation system tests
- Rate limiting tests
- Connection management tests
- Fork detection tests
- Partition detection tests
- Sybil attack detection tests
- Mempool validation tests
- Gossip protocol validation tests

Run tests:
```bash
go test -v ./x/networksecurity/keeper/...
```

## Security Considerations

### Production Deployment

1. **Configure Rate Limits**: Adjust based on network capacity
2. **Set Trusted Peers**: Add known reliable peers
3. **Monitor Alerts**: Set up monitoring for fork/partition alerts
4. **Tune Reputation**: Adjust penalties/rewards for your network
5. **Review Logs**: Regularly check security-related logs

### Performance Impact

- Rate limiting: Minimal (in-memory maps)
- Reputation tracking: Low (periodic updates)
- Fork detection: Low (simple hash comparison)
- Partition detection: Low (runs every 100 blocks)
- Gossip validation: Medium (signature verification if enabled)

### Recommended Configuration

For production networks:
```go
Params{
    RateLimit: RateLimitConfig{
        MaxRequestsPerSecond: 100,
        BurstSize: 200,
        BandwidthLimitPerPeer: 10 * 1024 * 1024,
    },
    Connection: ConnectionConfig{
        MaxInboundConnections: 100,
        MaxOutboundConnections: 50,
        MinPeerDiversity: 5,
    },
    Reputation: ReputationConfig{
        EnableTracking: true,
        InitialScore: 100,
        MinScoreToConnect: 20,
    },
}
```

## Integration Example

```go
import (
    "github.com/aura/chain/x/networksecurity"
    networksecuritykeeper "github.com/aura/chain/x/networksecurity/keeper"
)

// In app.go
app.NetworkSecurityKeeper = networksecuritykeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[networksecuritytypes.StoreKey]),
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
    logger,
)

// Register module
app.ModuleManager = module.NewManager(
    // ...
    networksecurity.NewAppModule(appCodec, app.NetworkSecurityKeeper),
)
```

## Future Enhancements

1. Machine learning-based anomaly detection
2. Advanced peer scoring algorithms
3. Geolocation-based peer selection
4. Dynamic parameter adjustment based on network conditions
5. Integration with external threat intelligence feeds
6. Advanced traffic pattern analysis
7. Distributed denial of service (DDoS) mitigation coordination
8. Automated incident response workflows
