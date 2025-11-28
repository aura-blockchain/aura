# Identity Module - Quick Reference

## Proto Files Overview

| File | Lines | Definitions | Purpose |
|------|-------|-------------|---------|
| `identity.proto` | 505 | 27 types | Core data structures and enums |
| `genesis.proto` | 58 | 1 type | Genesis state definition |
| `query.proto` | 365 | 51 types | Query service (25+ endpoints) |
| `tx.proto` | 284 | 43 types | Transaction service (20+ messages) |
| **Total** | **1,212** | **122** | **Complete module definition** |

## Key Message Types

### Identity Management
```protobuf
IdentityRecord          // DID, status, confidence score
ChangeRequest          // Identity modification request
ChangeHistory          // Historical change tracking
```

### Authentication & Authorization
```protobuf
Role                   // Named permission sets
RoleAssignment         // Role assignments to addresses
Session                // User session management
AuditLog              // Comprehensive audit trail
RateLimitConfig       // Per-user rate limiting
```

### Security
```protobuf
MultisigWallet        // 3-of-5, 5-of-7, custom wallets
MultisigProposal      // Proposals requiring signatures
TimeLockedAction      // Actions with mandatory delays
EmergencyAdmin        // Time-limited emergency powers
ValidatorKeyRotation  // Validator key rotation tracking
```

## Query Endpoints (25+)

### Identity
- `IdentityRecord(did)` → Identity details
- `IdentityRecordByAddress(address)` → Identity by address
- `ChangeRequest(request_id)` → Change request details
- `ChangeHistory(did)` → Historical changes

### Roles & Permissions
- `Role(role_name)` → Role details
- `RoleAssignments(address)` → User roles
- `HasPermission(address, permission)` → Permission check

### Multisig
- `MultisigWallet(wallet_id)` → Wallet details
- `MultisigProposal(proposal_id)` → Proposal status
- `MultisigProposalsByWallet(wallet_id)` → Wallet proposals

### Security
- `TimeLockedAction(action_id)` → Action status
- `EmergencyAdmin(address)` → Admin details
- `ValidatorRotation(validator_address)` → Rotation status

### Sessions & Audit
- `Session(session_id)` → Session details
- `SessionsByAddress(address)` → User sessions
- `AuditLogs()` → All audit logs
- `AuditLogsByActor(actor)` → Actor's logs

## Transaction Types (20+)

### Identity Operations
```bash
RequestIdentityChange      # Request identity modification
SubmitAssistantProof      # Assistant verification
ApplyIdentityChange       # Execute approved change
RejectIdentityChange      # Reject request
SuspendIdentityChanges    # Emergency suspension
```

### Role Management
```bash
CreateRole      # Define new role
AssignRole      # Grant role to address
RevokeRole      # Remove role from address
```

### Multisig
```bash
CreateMultisigWallet      # Create threshold wallet
CreateMultisigProposal    # Propose action
SignMultisigProposal      # Sign proposal
ExecuteMultisigProposal   # Execute approved
```

### Time-Locked
```bash
ProposeTimeLockedAction   # Propose delayed action
ExecuteTimeLockedAction   # Execute ready action
CancelTimeLockedAction    # Cancel pending action
```

### Emergency & Admin
```bash
ActivateEmergencyAdmin    # Grant emergency powers
DeactivateEmergencyAdmin  # Revoke emergency powers
RotateValidatorKey        # Update validator key
UpdateParams              # Update module params
```

### Sessions
```bash
CreateSession   # Establish session
EndSession      # Terminate session
```

## Status Enums

### IdentityStatus
```
UNSPECIFIED, ACTIVE, SUSPENDED, REVOKED, IDLE, PENDING_VERIFICATION
```

### ChangeStatus
```
UNSPECIFIED, PENDING, APPROVED, REJECTED, EXECUTED, READY_TO_APPLY, PENDING_VERIFICATION
```

### ProposalStatus
```
UNSPECIFIED, PENDING, APPROVED, EXECUTED, REJECTED, EXPIRED
```

### ActionStatus
```
UNSPECIFIED, PENDING, READY, EXECUTED, CANCELLED
```

### AuditResult
```
UNSPECIFIED, SUCCESS, FAILURE, DENIED
```

## CLI Quick Commands

### Query
```bash
# Identity
aurad query identity identity-record <did>
aurad query identity change-request <request-id>

# Roles
aurad query identity role <role-name>
aurad query identity role-assignments <address>
aurad query identity has-permission <address> <permission>

# Multisig
aurad query identity multisig-wallet <wallet-id>
aurad query identity multisig-proposal <proposal-id>

# Audit
aurad query identity audit-logs
aurad query identity audit-logs-by-actor <actor>
```

### Transactions
```bash
# Identity change
aurad tx identity request-identity-change \
  --target-did <did> --metadata-hash <hash> --ir-id <id> --proof-hash <hash>

# Roles
aurad tx identity create-role --role-name <name> --permissions <p1,p2>
aurad tx identity assign-role <address> <role-name>
aurad tx identity revoke-role <address> <role-name>

# Multisig
aurad tx identity create-multisig-wallet \
  --signers <addr1,addr2,...> --threshold <n> --wallet-type <type>
aurad tx identity create-multisig-proposal \
  --wallet-id <id> --title <title> --payload <data>
aurad tx identity sign-multisig-proposal --proposal-id <id>

# Time-locked
aurad tx identity propose-time-locked-action \
  --action-type <type> --payload <data> --delay-seconds <sec>
```

## HTTP REST Endpoints

All queries available via REST:
```
GET /aura/identity/v1beta1/identities/{did}
GET /aura/identity/v1beta1/identities/address/{address}
GET /aura/identity/v1beta1/roles/{role_name}
GET /aura/identity/v1beta1/role-assignments/{address}
GET /aura/identity/v1beta1/multisig-wallets/{wallet_id}
GET /aura/identity/v1beta1/sessions/{session_id}
GET /aura/identity/v1beta1/audit-logs
```

## Module Parameters

### AuthParams
- `enable_rbac` - Enable role-based access control
- `max_roles_per_account` - Max roles per account
- `session_timeout` - Session expiration duration
- `enable_audit_logging` - Enable audit logs
- `default_timelock_delay_seconds` - Time-lock delay
- `default_requests_per_minute/hour/day` - Rate limits
- `multisig_proposal_expiry_seconds` - Proposal TTL

### IdentityChangeParams
- `max_requests_per_wallet_per_month` - Monthly request limit
- `min_confidence_after_change` - Min confidence score
- `staleness_height_threshold` - Staleness threshold
- `assistant_slash_on_false_positive` - Slash flag
- `staleness_investigator_chain` - Investigator chain

## Storage Key Prefixes

```
0x01 - Roles
0x02 - Role assignments
0x03 - Multisig wallets
0x04 - Multisig proposals
0x05 - Time-locked actions
0x06 - Emergency admins
0x07 - Validator rotations
0x08 - Sessions
0x09 - User sessions index
0x0A - Rate limits
0x0B - Audit logs
0x0C - Params
0x0D - Audit log counter
0x10 - Identity records
0x11 - Change requests
0x12 - Change history
0x13 - DID to address mapping
```

## Code Generation

```bash
# Generate Go code from proto
cd /home/decri/blockchain-projects/aura/proto
buf generate

# Or using make
cd /home/decri/blockchain-projects/aura
make proto-gen
```

## Import Paths

```go
import (
    identitytypes "github.com/aequitas/aura/proto/aura/identity/v1beta1"
)
```

## Package Info

- **Package**: `aura.identity.v1beta1`
- **Go Package**: `github.com/aequitas/aura/proto/aura/identity/v1beta1`
- **Proto Version**: proto3
- **Location**: `/home/decri/blockchain-projects/aura/proto/aura/identity/v1beta1/`

## Features Consolidated

From **auth** module:
- Role-based access control (RBAC)
- Multisig wallets and proposals
- Time-locked actions
- Emergency admin privileges
- Validator key rotation
- Session management
- Rate limiting
- Audit logging

From **identitychange** module:
- Decentralized identifiers (DIDs)
- Identity records with confidence scores
- Change request workflows
- Assistant proof verification
- Change history tracking
- Identity status management

## Testing Checklist

- [ ] Identity record CRUD
- [ ] Change request workflow
- [ ] Role creation and assignment
- [ ] Permission checking
- [ ] Multisig operations
- [ ] Time-locked action delays
- [ ] Emergency admin activation/deactivation
- [ ] Session creation/expiration
- [ ] Rate limiting enforcement
- [ ] Audit log generation
- [ ] Validator key rotation
- [ ] Genesis import/export
- [ ] Query endpoints
- [ ] Transaction execution

## Next Steps

1. Generate Go code: `buf generate`
2. Create module structure: `/chain/x/identity/`
3. Implement keeper
4. Implement msg/query servers
5. Create CLI commands
6. Write tests
7. Integrate with app
8. Migrate from auth/identitychange
9. Deploy and test

## Support

- Proto definitions: `/proto/aura/identity/v1beta1/`
- Documentation: `README.md` in same directory
- Full report: `/IDENTITY_PROTO_IMPLEMENTATION.md`
