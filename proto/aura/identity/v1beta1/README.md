# AURA Identity Module - Protocol Buffer Definitions

This directory contains the consolidated protocol buffer definitions for the AURA Identity module, which combines functionality from:
- **auth**: Role-based access control, sessions, multisig, time-locked actions
- **identitychange**: Decentralized identity management and change workflows

## Files

### 1. `identity.proto`
Main types file containing all core data structures:

**Parameters:**
- `Params`: Unified module parameters
- `AuthParams`: Authentication and authorization settings
- `IdentityChangeParams`: Identity change workflow settings

**Identity Management:**
- `IdentityRecord`: Decentralized identity with DID, status, confidence score
- `ChangeRequest`: Identity modification requests
- `ChangeHistory`: Historical record of identity changes
- `IdentityStatus`: Active, Suspended, Revoked, etc.
- `ChangeType`: Update attributes, verification methods, transfer control
- `ChangeStatus`: Pending, Approved, Rejected, Executed

**Authentication & Authorization:**
- `Role`: Named permission sets
- `RoleAssignment`: Role assignments to addresses
- `AuditLog`: Comprehensive audit trail
- `Session`: User session management
- `RateLimitConfig`: Per-user rate limiting

**Multisig:**
- `MultisigWallet`: Multi-signature wallet (3-of-5, 5-of-7, custom)
- `MultisigProposal`: Proposals requiring multiple signatures
- `WalletType`: Predefined and custom wallet types
- `ProposalStatus`: Pending, Approved, Executed, etc.

**Advanced Security:**
- `TimeLockedAction`: Admin actions with time delays
- `EmergencyAdmin`: Emergency privileges with expiration
- `ValidatorKeyRotation`: Validator key rotation tracking
- `ActionStatus`: Status tracking for time-locked operations

### 2. `genesis.proto`
Genesis state definition for the identity module:

```protobuf
message GenesisState {
  Params params
  repeated IdentityRecord identity_records
  repeated ChangeRequest change_requests
  repeated ChangeHistory change_history
  repeated Role roles
  repeated RoleAssignment role_assignments
  repeated MultisigWallet multisig_wallets
  repeated MultisigProposal multisig_proposals
  repeated TimeLockedAction time_locked_actions
  repeated EmergencyAdmin emergency_admins
  repeated ValidatorKeyRotation validator_rotations
  repeated Session sessions
  repeated RateLimitConfig rate_limits
  repeated AuditLog audit_logs
  bool identity_changes_suspended
  uint64 next_audit_log_id
}
```

### 3. `query.proto`
gRPC query service with 25+ query endpoints:

**Identity Queries:**
- `IdentityRecord` - Query identity by DID
- `IdentityRecordByAddress` - Query identity by address
- `AllIdentityRecords` - List all identities
- `ChangeRequest` - Query change request by ID
- `ChangeRequestsByDID` - Query change requests for a DID
- `ChangeHistory` - Query change history for a DID

**Auth Queries:**
- `Role` - Query role by name
- `AllRoles` - List all roles
- `RoleAssignments` - Query role assignments for address
- `HasPermission` - Check if address has permission

**Multisig Queries:**
- `MultisigWallet` - Query wallet by ID
- `AllMultisigWallets` - List all wallets
- `MultisigProposal` - Query proposal by ID
- `MultisigProposalsByWallet` - Query proposals for wallet

**Security Queries:**
- `TimeLockedAction` - Query time-locked action
- `AllTimeLockedActions` - List all time-locked actions
- `EmergencyAdmin` - Query emergency admin
- `AllEmergencyAdmins` - List all emergency admins
- `ValidatorRotation` - Query validator rotation

**Session & Rate Limit Queries:**
- `Session` - Query session by ID
- `SessionsByAddress` - Query sessions for address
- `RateLimit` - Query rate limit config

**Audit Queries:**
- `AuditLogs` - Query all audit logs
- `AuditLogsByActor` - Query audit logs by actor

All queries support pagination where applicable and follow REST conventions with HTTP annotations.

### 4. `tx.proto`
gRPC transaction service with 20+ message types:

**Identity Change Operations:**
- `RequestIdentityChange` - Request identity modification
- `SubmitAssistantProof` - Submit assistant verification proof
- `ApplyIdentityChange` - Apply approved change
- `RejectIdentityChange` - Reject change request
- `SuspendIdentityChanges` - Suspend all identity changes (emergency)

**Role Management:**
- `CreateRole` - Create new role with permissions
- `AssignRole` - Assign role to address
- `RevokeRole` - Revoke role from address

**Multisig Operations:**
- `CreateMultisigWallet` - Create multisig wallet (3-of-5, 5-of-7, custom)
- `CreateMultisigProposal` - Create proposal for wallet
- `SignMultisigProposal` - Sign proposal
- `ExecuteMultisigProposal` - Execute approved proposal

**Time-Locked Operations:**
- `ProposeTimeLockedAction` - Propose delayed action
- `ExecuteTimeLockedAction` - Execute ready action
- `CancelTimeLockedAction` - Cancel pending action

**Emergency Operations:**
- `ActivateEmergencyAdmin` - Activate emergency admin with privileges
- `DeactivateEmergencyAdmin` - Deactivate emergency admin

**Validator Operations:**
- `RotateValidatorKey` - Rotate validator consensus key

**Session Operations:**
- `CreateSession` - Create user session
- `EndSession` - End user session

**Admin Operations:**
- `UpdateParams` - Update module parameters (governance)

## Key Features

### 1. Unified Identity Management
- Decentralized Identifiers (DIDs)
- Confidence score tracking
- Verification methods
- Identity change workflows with assistant proofs
- Historical change tracking

### 2. Role-Based Access Control (RBAC)
- Flexible permission system
- System and custom roles
- Time-limited role assignments
- Permission checking

### 3. Multisig Security
- Multiple wallet types (3-of-5, 5-of-7, custom)
- Proposal-based execution
- Threshold signatures
- Expiration handling

### 4. Time-Locked Actions
- Delayed execution for critical operations
- Cancellation support
- Status tracking
- Ready-to-execute queries

### 5. Emergency Administration
- Time-limited emergency powers
- Specific privilege grants
- Activation/deactivation tracking

### 6. Session Management
- Device fingerprinting
- IP tracking
- Expiration handling
- Session metadata

### 7. Rate Limiting
- Per-minute/hour/day limits
- Per-user configuration
- Sliding window tracking

### 8. Comprehensive Auditing
- All actions logged
- Actor and target tracking
- Result recording
- Queryable audit trail

### 9. Validator Key Rotation
- Consensus key updates
- Rotation status tracking
- Historical rotation records

## Proto Generation

To generate Go code from these proto files:

```bash
cd /home/decri/blockchain-projects/aura/proto
buf generate
```

Or use the Cosmos SDK proto generation:

```bash
cd /home/decri/blockchain-projects/aura
make proto-gen
```

## Integration Notes

### Module Consolidation
This module consolidates:
- `/chain/x/auth/` → Identity auth features
- `/chain/x/identitychange/` → Identity change workflows

### Backwards Compatibility
- Existing auth functionality preserved
- Identity change workflows maintained
- All proto message names follow conventions

### gRPC Services
- Query service: Read-only operations
- Msg service: State-changing transactions
- All endpoints have HTTP/REST annotations

### Storage Keys
The module uses the following key prefixes:
- `0x01` - Roles
- `0x02` - Role assignments
- `0x03` - Multisig wallets
- `0x04` - Multisig proposals
- `0x05` - Time-locked actions
- `0x06` - Emergency admins
- `0x07` - Validator rotations
- `0x08` - Sessions
- `0x09` - User sessions index
- `0x0A` - Rate limits
- `0x0B` - Audit logs
- `0x0C` - Params
- `0x0D` - Audit log counter
- `0x10` - Identity records
- `0x11` - Change requests
- `0x12` - Change history
- `0x13` - DID to address mapping

## Usage Examples

### Query an Identity
```bash
aurad query identity identity-record did:aura:1234567890
```

### Request Identity Change
```bash
aurad tx identity request-identity-change \
  --target-did did:aura:1234567890 \
  --metadata-hash abc123 \
  --ir-id ir-001 \
  --proof-hash def456 \
  --from alice
```

### Assign Role
```bash
aurad tx identity assign-role \
  aura1address \
  moderator \
  --from admin
```

### Create Multisig Wallet
```bash
aurad tx identity create-multisig-wallet \
  --signers aura1addr1,aura1addr2,aura1addr3,aura1addr4,aura1addr5 \
  --threshold 3 \
  --wallet-type 3-of-5 \
  --from creator
```

### Propose Time-Locked Action
```bash
aurad tx identity propose-time-locked-action \
  --action-type UPDATE_PARAMS \
  --payload <encoded-data> \
  --delay-seconds 86400 \
  --from admin
```

## Statistics

- **Total Messages**: 122 (messages, enums, services)
- **Query Endpoints**: 25+
- **Transaction Types**: 20+
- **Enums**: 7
- **Main Message Types**: 25+

## Related Documentation

- Module implementation: `/chain/x/identity/`
- Keeper: `/chain/x/identity/keeper/`
- Types: `/chain/x/identity/types/`
- CLI: `/chain/x/identity/client/cli/`
- Tests: `/chain/x/identity/keeper/*_test.go`

## Version

- Package: `aura.identity.v1beta1`
- Go Package: `github.com/aequitas/aura/proto/aura/identity/v1beta1`
- Proto3 syntax
