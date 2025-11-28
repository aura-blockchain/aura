# Validator Security Module

## Overview

The Validator Security module provides comprehensive security features for validators on the Aura blockchain, including slashing, jailing, tombstoning, monitoring, and automated failover capabilities.

## Features

### 1. Slashing Conditions for Validator Misbehavior

**File:** `keeper/slashing.go` (Lines 11-201)

- **Double-sign slashing**: Validators who sign two different blocks at the same height are slashed and permanently tombstoned
- **Downtime slashing**: Validators who miss too many blocks are slashed and temporarily jailed
- **Configurable slash fractions**: Separate slash percentages for double-signing (default 5%) and downtime (default 0.01%)

**Key Functions:**
- `HandleDoubleSign()`: Processes double-signing evidence and applies slashing
- `HandleDowntime()`: Detects downtime violations and applies penalties
- `ValidateMinimumStake()`: Ensures validators meet minimum staking requirements

### 2. Double-Sign Detection Mechanisms

**File:** `keeper/slashing.go` (Lines 14-98)

- Validates and stores double-signing evidence
- Compares two votes at the same height
- Automatically tombstones violating validators
- Creates critical alerts for monitoring

**Evidence Storage:**
- Validator address
- Block height
- Both conflicting votes
- Timestamp
- Slash fraction applied

### 3. Downtime Penalty System

**File:** `keeper/slashing.go` (Lines 101-159), `keeper/monitoring.go` (Lines 11-76)

- Tracks block signing using a sliding window approach
- Configurable window size (default 10,000 blocks)
- Minimum signed blocks threshold (default 50%)
- Automatic jailing when threshold is exceeded

**Key Functions:**
- `TrackBlockSign()`: Records each block signing event
- Maintains missed block counter
- Uses bitmap for efficient storage

### 4. Tombstoning for Permanent Validator Bans

**File:** `keeper/jailing.go` (Lines 93-148)

- Permanent ban for severe violations (double-signing)
- Cannot be reversed
- Validator removed from active set
- Region count decremented if geo-distribution enabled

**Functions:**
- `TombstoneValidator()`: Permanently bans a validator
- `IsValidatorTombstoned()`: Checks tombstone status
- `GetTombstonedValidators()`: Returns all tombstoned validators

### 5. Validator Key Separation (Hot/Cold Keys)

**File:** `keeper/keeper.go` (Lines 155-166), `types/validator_security.proto` (Lines 55-58)

- Separate hot keys for signing and cold keys for staking
- Validation ensures keys are different
- Enhances security by isolating signing operations
- Stored in ValidatorSecurityInfo

**Validation:**
- Hot and cold keys must be different
- Both must be set if key separation is enabled
- Validated during registration and updates

### 6. Sentry Node Architecture for DDoS Protection

**File:** `keeper/sentry.go` (Lines 1-240)

- Register multiple sentry nodes per validator
- Track sentry node status and heartbeats
- Monitor request counts and blocked requests
- Automatic failover if sentry nodes go offline

**Key Functions:**
- `RegisterSentryNode()`: Registers a new sentry node
- `UpdateSentryHeartbeat()`: Updates node health status
- `RecordSentryRequest()`: Tracks request metrics
- `DeactivateSentryNode()`: Handles node failures

**Sentry Node Info:**
- IP address and port
- Active status
- Last heartbeat timestamp
- Request statistics (total and blocked)

### 7. Validator Monitoring and Alerting System

**File:** `keeper/monitoring.go` (Lines 1-266)

- Continuous monitoring of validator health
- Multiple alert types and severity levels
- Alert acknowledgment system
- Automatic alert generation for violations

**Alert Types:**
- DOWNTIME: Validator inactive
- DOUBLE_SIGN: Double-signing detected
- LOW_STAKE: Below minimum stake
- SENTRY_NODE_OFFLINE: Sentry node unavailable
- GEOGRAPHIC_VIOLATION: Region capacity exceeded
- KEY_COMPROMISE: Security breach detected
- FAILOVER_TRIGGERED: Backup activated

**Severity Levels:**
- INFO: Informational messages
- WARNING: Potential issues
- CRITICAL: Immediate attention required

### 8. Automated Failover to Backup Validators

**File:** `keeper/sentry.go` (Lines 167-240)

- Configure backup validators during registration
- Automatic failover when primary fails
- Restore from failover when primary recovers
- Alerts generated for all failover events

**Functions:**
- `TriggerFailover()`: Activates backup validator
- `RestoreFromFailover()`: Returns to primary validator
- Verifies backup validator health before activation

### 9. Geographical Distribution Requirements

**File:** `keeper/keeper.go` (Lines 155-178, 236-259)

- Enforce maximum validators per region
- Track regional validator counts
- Prevent over-concentration in single region
- Configurable limits per region

**Features:**
- Region-based capacity limits
- Latitude/longitude validation
- Country code tracking
- Automatic alerts for violations

### 10. Minimum Staking Requirements

**File:** `keeper/slashing.go` (Lines 162-190)

- Configurable minimum stake amount
- Validation during validator operations
- Alerts when stake falls below minimum
- Required for unjailing

**Default:** 1000 tokens (with 9 decimals)

### 11. Jailing Mechanism for Temporary Suspensions

**File:** `keeper/jailing.go` (Lines 1-208)

- Temporary jail for downtime violations
- Configurable jail duration (default 24 hours)
- Manual unjailing after period expires
- Requirements verification before unjailing

**Functions:**
- `JailValidator()`: Temporarily jails a validator
- `UnjailValidator()`: Unjails after requirements met
- `IsValidatorJailed()`: Check jail status
- `GetJailedValidators()`: List all jailed validators

**Unjail Requirements:**
- Jail period expired
- Minimum stake met
- Sufficient sentry nodes (if required)
- Not tombstoned

## Configuration Parameters

### Default Values

```go
DoubleSignSlashFraction:  5%        // Slash 5% for double-signing
DowntimeSlashFraction:    0.01%     // Slash 0.01% for downtime
SignedBlocksWindow:       10000     // Track last 10k blocks
MinSignedPerWindow:       50%       // Must sign 50% of blocks
DowntimeJailDuration:     24h       // Jail for 24 hours
MinimumStakeAmount:       1000      // Minimum 1000 tokens
EnableGeoDistribution:    true      // Enforce geo limits
MaxValidatorsPerRegion:   10        // Max 10 per region
RequireSentryNodes:       true      // Require sentry nodes
MinSentryNodes:           2         // Minimum 2 sentry nodes
MonitoringInterval:       5m        // Monitor every 5 minutes
EnableAutoFailover:       true      // Auto-failover enabled
FailoverTimeout:          10m       // Failover after 10 minutes
```

## Usage Examples

### Register Validator with Security Info

```go
msg := MsgRegisterValidator{
    ValidatorAddress: "auravaloper1...",
    HotKey: "hot_key_pub",
    ColdKey: "cold_key_pub",
    Region: "us-west-2",
    CountryCode: "US",
    Latitude: 37.7749,
    Longitude: -122.4194,
    BackupValidatorAddresses: []string{"auravaloper2...", "auravaloper3..."},
}
```

### Register Sentry Node

```go
msg := MsgRegisterSentryNode{
    ValidatorAddress: "auravaloper1...",
    SentryAddress: "aurasentry1...",
    IpAddress: "192.168.1.100",
    Port: 26656,
}
```

### Report Double Signing

```go
msg := MsgReportDoubleSign{
    ReporterAddress: "aura1...",
    ValidatorAddress: "auravaloper1...",
    Height: 12345,
    VoteA: []byte("..."),
    VoteB: []byte("..."),
}
```

### Unjail Validator

```go
msg := MsgUnjail{
    ValidatorAddress: "auravaloper1...",
}
```

## State Storage

### Keys

- **Params**: Module parameters
- **ValidatorSecurityInfo**: Security information per validator
- **DoubleSignEvidence**: Evidence of double-signing
- **DowntimeInfraction**: Downtime violation records
- **ValidatorAlert**: Monitoring alerts
- **SentryNodeInfo**: Sentry node information
- **ValidatorSigningInfo**: Block signing statistics
- **JailedValidators**: Index of jailed validators
- **TombstonedValidators**: Index of tombstoned validators
- **RegionValidatorCount**: Validators per region

## Events

The module emits the following events:

- `validator_registered`: New validator registered
- `validator_security_updated`: Security info updated
- `sentry_node_registered`: Sentry node added
- `double_sign_reported`: Double-signing detected
- `validator_unjailed`: Validator unjailed
- `alert_acknowledged`: Alert acknowledged
- `params_updated`: Parameters updated

## Queries

- `Params`: Get module parameters
- `ValidatorSecurityInfo`: Get validator security info
- `AllValidators`: List all validators
- `JailedValidators`: List jailed validators
- `TombstonedValidators`: List tombstoned validators
- `DoubleSignEvidences`: List all evidences
- `ValidatorAlerts`: Get validator alerts
- `SentryNodes`: Get validator sentry nodes

## Integration

To integrate this module into your application:

1. Add to `app/app.go`:
```go
import validatorsecuritykeeper "github.com/aura/aura/x/validatorsecurity/keeper"
import validatorsecuritytypes "github.com/aura/aura/x/validatorsecurity/types"
```

2. Add keeper to app struct
3. Initialize keeper in NewApp()
4. Add to module manager
5. Set up begin/end blockers

## Testing

Run tests with:
```bash
cd x/validatorsecurity
go test ./...
```

Test coverage includes:
- Parameter validation
- Validator registration
- Key separation
- Geographic validation
- Sentry node management
- Double-sign detection
- Downtime tracking
- Jailing/unjailing
- Tombstoning
- Alert system
- Failover mechanisms

## Security Considerations

1. **Double-signing is permanent**: Validators who double-sign are permanently tombstoned
2. **Key separation is enforced**: Hot and cold keys must be different
3. **Geographic distribution**: Prevents centralization in single region
4. **Sentry nodes required**: DDoS protection mandatory
5. **Minimum stake enforced**: Validators must maintain minimum stake
6. **Monitoring required**: Regular health checks mandatory
7. **Failover automatic**: System handles validator failures automatically

## File Summary

### Proto Files
- `/proto/aura/validatorsecurity/v1beta1/validator_security.proto`: Core data structures
- `/proto/aura/validatorsecurity/v1beta1/tx.proto`: Transaction messages
- `/proto/aura/validatorsecurity/v1beta1/query.proto`: Query definitions

### Types Files
- `types/params.go`: Parameter definitions and validation
- `types/keys.go`: Store keys and key constructors
- `types/errors.go`: Error definitions
- `types/genesis.go`: Genesis state handling
- `types/genesis_test.go`: Genesis validation tests

### Keeper Files
- `keeper/keeper.go`: Core keeper with validator registration
- `keeper/slashing.go`: Slashing and double-sign detection
- `keeper/jailing.go`: Jailing and tombstoning logic
- `keeper/monitoring.go`: Monitoring and alerting system
- `keeper/sentry.go`: Sentry node and failover management
- `keeper/msg_server.go`: Message handlers
- `keeper/query_server.go`: Query handlers
- `keeper/expected_keepers.go`: Interface definitions
- `keeper/keeper_test.go`: Comprehensive keeper tests
- `keeper/slashing_test.go`: Slashing-specific tests

### Module Files
- `module.go`: Module definition and registration
- `genesis.go`: Genesis initialization and export
- `abci.go`: Begin/end block logic
- `README.md`: This documentation file
