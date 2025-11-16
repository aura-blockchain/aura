# Pre-Validation Module

## Overview

The Pre-Validation module optimizes transaction processing by pre-validating common transaction types during off-peak electricity hours (default: 2am-6am). Pre-validated transactions are stored as encrypted smart contracts and executed instantly when needed, reducing transaction time and energy consumption.

## Architecture

### Core Components

#### 1. Transaction Pre-Validation System
- **Template-Based Validation**: Common transaction patterns are defined as templates
- **Encrypted Storage**: Pre-validated transactions are stored encrypted using AES-256-GCM
- **Nonce-Based Uniqueness**: Each pre-validated transaction includes a unique nonce
- **Expiry Management**: Pre-validations expire after a configurable period (default: 72 hours)

#### 2. Scheduler System
- **Off-Peak Execution**: Runs during configured off-peak hours to minimize energy costs
- **Interval-Based**: Configurable run intervals (default: 30 minutes)
- **Volume Control**: Maximum pre-validations per run to prevent resource exhaustion
- **Emergency Override**: Can run during peak hours if needed (admin function)

#### 3. Cache Management
- **Multiple Strategies**:
  - **FIFO**: First In, First Out
  - **LRU**: Least Recently Used
  - **LFU**: Least Frequently Used
  - **Adaptive**: Hybrid approach considering both frequency and recency
- **Size Limits**: Configurable maximum cache size
- **Automatic Eviction**: Evicts based on strategy when cache is full

#### 4. Auto-Scaling System
- **Metrics-Driven**: Adjusts amounts based on cache hit rates and execution patterns
- **Per-Type Scaling**: Each transaction type scales independently
- **Bounds Enforcement**: Respects min/max amounts per type
- **Cooldown Periods**: Prevents rapid scaling fluctuations

#### 5. Monitoring & Metrics
- **Real-Time Tracking**: Monitors cache hits, misses, and execution times
- **Control Group**: Small percentage of transactions bypass pre-validation for comparison
- **Energy Tracking**: Estimates energy savings based on execution patterns
- **Hourly Metrics**: Maintains 24-hour rolling window of metrics

## Transaction Types

The module supports pre-validation for these high-frequency transaction types:

1. **IR Completions** (Highest Priority)
   - Most frequent transaction type
   - Default template: 100 pre-validated per run
   - Gas estimate: 50,000

2. **DEX Swaps**
   - Common trading pairs (USDC-AURA, ETH-AURA, etc.)
   - Default template: 50 per run
   - Gas estimate: 100,000

3. **LP Deposits/Withdrawals**
   - Liquidity pool operations
   - Default template: 30 per run each
   - Gas estimate: 80,000

4. **VC Minting**
   - Verifiable credential creation
   - Default template: 20 per run
   - Gas estimate: 120,000
   - Requires confidence score ≥ 500

5. **Bridge Transfers**
   - Cross-chain transfers
   - Default template: 15 per run
   - Gas estimate: 150,000

6. **Confidence Score Updates**
   - Score adjustments
   - Default template: 25 per run
   - Gas estimate: 60,000

7. **Identity Changes**
   - Identity modification requests
   - Default template: 10 per run
   - Gas estimate: 90,000
   - Requires confidence score ≥ 1000

## Templates

### Template Structure

```protobuf
message ValidationTemplate {
  string id = 1;
  TransactionType tx_type = 2;
  string name = 3;
  string description = 4;
  string validation_rules = 5;      // JSON schema
  string parameter_schema = 6;      // Parameter definitions
  string gas_formula = 7;           // Gas estimation formula
  uint32 priority_weight = 8;       // Higher = more templates
  uint64 min_confidence_score = 9;  // Minimum required score
  bool active = 10;
  TemplateStats stats = 13;         // Usage statistics
}
```

### Default Templates

The module includes default templates for all transaction types. Custom templates can be added for specific use cases (e.g., "USDC-AURA-high-volume-swap").

## Configuration

### Parameters

```protobuf
message Params {
  bool enabled = 1;
  SchedulerConfig scheduler_config = 2;
  AutoScalingConfig auto_scaling_config = 3;
  CacheStrategy cache_strategy = 4;
  uint64 max_cache_size = 5;                    // Default: 10,000
  uint32 expiry_hours = 6;                      // Default: 72
  string encryption_algorithm = 7;              // Default: "AES-256-GCM"
  double control_group_percentage = 8;          // Default: 5.0%
  uint64 min_confidence_score = 9;              // Default: 100
  double energy_cost_per_validation_kwh = 10;   // Default: 0.0001
  double energy_cost_per_execution_kwh = 11;    // Default: 0.001
  bool metrics_enabled = 12;                    // Default: true
  bool detailed_logging = 13;                   // Default: false
}
```

### Scheduler Configuration

```protobuf
message SchedulerConfig {
  repeated uint32 off_peak_hours = 1;      // Default: [2, 3, 4, 5, 6]
  string timezone = 2;                     // Default: "UTC"
  bool enabled = 3;                        // Default: true
  uint32 run_interval_minutes = 4;         // Default: 30
  uint64 max_per_run = 5;                  // Default: 1000
  bool allow_peak_hours = 6;               // Default: false
}
```

### Auto-Scaling Configuration

```protobuf
message AutoScalingConfig {
  bool enabled = 1;                        // Default: true
  map<string, uint64> initial_amounts = 2; // Per-type starting amounts
  double target_cache_hit_rate = 3;        // Default: 0.80 (80%)
  double min_cache_hit_rate = 4;           // Default: 0.50 (50%)
  map<string, uint64> max_amounts = 5;     // Per-type maximums
  double scale_up_factor = 6;              // Default: 1.5
  double scale_down_factor = 7;            // Default: 0.75
  uint32 cooldown_minutes = 8;             // Default: 60
  uint32 evaluation_period_hours = 9;      // Default: 24
}
```

## Usage Flow

### 1. Pre-Validation Creation (Off-Peak Hours)

```
Scheduler triggers →
  For each transaction type:
    Get templates →
    Generate synthetic transactions →
    Validate transaction →
    Encrypt data →
    Store in cache →
    Update metrics
```

### 2. Transaction Execution (Any Time)

```
User submits transaction →
  Check control group (5%) →
    If control group: Normal execution (for comparison)
    If not control group:
      Search cache for matching pre-validation →
        If found (cache hit):
          Decrypt pre-validated data →
          Execute instantly →
          Record time savings →
          Update metrics
        If not found (cache miss):
          Normal execution →
          Record miss →
          Update metrics
```

### 3. Auto-Scaling (Periodic)

```
Cooldown period elapsed →
  For each transaction type:
    Analyze metrics →
      If cache hit rate > target: Scale up
      If cache hit rate < minimum: Scale down
      If execution rate high: Scale up
      If expiration rate high: Scale down
    Apply new amounts →
    Record adjustment
```

## Metrics & Monitoring

### Key Metrics

1. **Cache Performance**
   - Hit rate (overall and per-type)
   - Miss rate
   - Cache utilization

2. **Time Savings**
   - Average time saved per transaction
   - Total time saved
   - Per-type time savings

3. **Energy Efficiency**
   - Total energy saved (kWh)
   - Cost savings

4. **Control Group**
   - Average execution time
   - Median, P95, P99 percentiles
   - Standard deviation

5. **Template Performance**
   - Usage count per template
   - Execution vs. expiration ratio
   - Cache hit rate per template

### Events

The module emits several event types for monitoring:

- `EventPreValidationCreated`: New pre-validation created
- `EventPreValidationExecuted`: Pre-validation used
- `EventPreValidationExpired`: Pre-validation expired
- `EventCacheHit`: Cache hit occurred
- `EventCacheMiss`: Cache miss occurred
- `EventSchedulerRun`: Scheduler executed
- `EventAutoScaling`: Auto-scaling adjustment
- `EventMetricsUpdate`: Periodic metrics update

## Security Considerations

### Encryption

- **Algorithm**: AES-256-GCM (authenticated encryption)
- **Key Rotation**: Support for multiple encryption keys
- **Key Management**: Keys stored in keeper, rotatable for security

### Access Control

- **Confidence Score Gating**: Minimum scores required for certain transaction types
- **Template Validation**: All templates validated before use
- **Nonce Uniqueness**: Prevents replay attacks

### Data Integrity

- **Transaction Hashing**: Each transaction includes cryptographic hash
- **Validation Proofs**: Cryptographic proof of pre-validation
- **State Root Tracking**: Links pre-validation to specific state

## Performance Characteristics

### Time Savings

Expected time savings based on transaction complexity:

- Simple transactions (IR completions): 50-100ms
- Medium transactions (DEX swaps): 100-200ms
- Complex transactions (Bridge transfers): 200-500ms

### Energy Savings

Based on configuration:
- Pre-validation cost: 0.1 Wh
- Normal execution cost: 1.0 Wh
- Net savings per execution: 0.9 Wh

At scale (10,000 transactions/day with 80% hit rate):
- Daily savings: 7,200 Wh = 7.2 kWh
- Annual savings: ~2,628 kWh

### Cache Efficiency

Target metrics:
- Cache hit rate: 80% (configurable)
- Cache utilization: 70-90%
- Expiration rate: <20%

## Integration with Other Modules

### Confidence Score Module
- Validates user confidence scores before pre-validation
- Higher scores enable access to more transaction types

### Inclusion Routines Module
- IR completions are highest priority for pre-validation
- Optimizes the most frequent transaction type

### VC Registry Module
- VC minting pre-validated for qualified users
- Requires higher confidence scores

### DEX Module
- Common trading pairs pre-validated
- Reduces swap latency for users

### Bridge Module
- Cross-chain transfers pre-validated
- Improves UX for multi-chain operations

## Administration

### Admin Functions

1. **Force Scheduler Run**: Trigger scheduler outside normal schedule
2. **Adjust Type Amounts**: Manually set pre-validation amounts
3. **Reset Auto-Scaling**: Return to initial amounts
4. **Reset Metrics**: Clear all metrics (testing only)
5. **Register Templates**: Add new transaction templates

### Monitoring Dashboards

Recommended dashboard panels:

1. **Cache Performance**
   - Hit rate over time
   - Hit/miss ratio by type
   - Cache size utilization

2. **Scaling Behavior**
   - Type amounts over time
   - Scaling events timeline
   - Hit rate vs. target

3. **Energy & Cost**
   - Cumulative energy savings
   - Cost savings (if pricing configured)
   - Savings by transaction type

4. **Control Group Analysis**
   - Execution time distribution
   - Comparison: pre-validated vs. normal
   - Percentile charts

## Future Enhancements

### Planned Features

1. **Machine Learning Integration**
   - Predict transaction patterns
   - Optimize template selection
   - Dynamic parameter tuning

2. **Cross-Chain Pre-Validation**
   - Pre-validate bridge transactions
   - Coordinate with destination chains

3. **User-Specific Templates**
   - Personalized pre-validation
   - Based on user history and preferences

4. **Advanced Encryption**
   - Homomorphic encryption support
   - Zero-knowledge proofs for validation

5. **Distributed Scheduling**
   - Multi-validator pre-validation
   - Load distribution across validators

## Troubleshooting

### Common Issues

**Low Cache Hit Rate**
- Check if transaction patterns match templates
- Review template parameter schemas
- Consider adding more specific templates

**High Expiration Rate**
- Reduce expiry hours
- Scale down amounts for affected types
- Review transaction volume predictions

**Scheduler Not Running**
- Verify enabled in configuration
- Check off-peak hours configuration
- Ensure not blocked by cooldown period

**Memory Issues**
- Reduce max_cache_size
- Increase cleanup frequency
- Enable more aggressive eviction strategy

## References

- [Protobuf Definitions](../../proto/aura/prevalidation/v1beta1/prevalidation.proto)
- [Keeper Implementation](../../chain/x/prevalidation/keeper/)
- [Types & Errors](../../chain/x/prevalidation/types/)
- [Parameters](../../chain/x/prevalidation/params/)
