# Identity Module

The Identity module consolidates authentication, authorization, and identity change management into a single, cohesive module. It merges functionality from the `auth` and `identitychange` modules.

## Overview

This module provides:

1. **Role-Based Access Control (RBAC)**: Create roles with specific permissions and assign them to addresses
2. **Session Management**: Create, track, and manage user sessions
3. **Rate Limiting**: Configure rate limits per user
4. **Multisig Wallets**: Support for multisignature wallets and proposals
5. **Time-Locked Actions**: Actions that execute after a specified delay
6. **Emergency Admin**: Temporary emergency administrator privileges
7. **Validator Key Rotation**: Manage validator consensus key rotation
8. **Identity Management**: Decentralized identity (DID) records
9. **Identity Change Requests**: Workflow for identity changes with verification
10. **Identity Recovery**: Social recovery for identity accounts
11. **Identity Verification**: Multi-level identity verification
12. **Cross-Chain Identity**: Link identities across different chains
13. **Audit Logging**: Comprehensive audit trail of all operations

## Architecture

### Module Structure

```
x/identity/
├── types/
│   ├── keys.go       - Store keys and prefixes
│   ├── errors.go     - Error definitions
│   ├── genesis.go    - Genesis state and parameters
│   └── types.go      - Core type definitions
├── keeper/
│   ├── keeper.go     - Main keeper with genesis methods
│   ├── auth.go       - Role and permission management
│   ├── changes.go    - Identity change request functions
│   └── sessions.go   - Session, multisig, and auxiliary functions
├── module.go         - AppModule implementation
└── README.md         - This file
```

### Key Store Prefixes

The module uses the following key prefixes for data storage:

- `0x00`: Module parameters
- `0x01-0x08`: Auth-related data (roles, permissions, accounts, sessions, rate limits)
- `0x09-0x0e`: Multisig, time-lock, emergency admin, validator rotation
- `0x10-0x17`: Identity records, change requests, history, recovery, verification, delegation, federation, cross-chain
- `0x20-0x21`: Counters for audit logs and change requests
- `0x30`: Suspended flag

## Core Concepts

### 1. Roles and Permissions

**Roles** are collections of permissions that can be assigned to addresses. The module includes default roles:

- **Admin**: Full administrative access to all module functions
- **Moderator**: Can manage roles and sessions, verify identities
- **Validator**: Can rotate validator keys
- **User**: Basic user with no special permissions

**Permissions** are granular capabilities such as:
- `create_role`, `assign_role`, `revoke_role`
- `manage_multisig`, `manage_timelock`, `manage_emergency`
- `manage_identity`, `verify_identity`, `approve_change_request`

### 2. Identity Records

Identity records represent decentralized identities (DIDs) with:
- DID identifier
- Owner address
- Metadata hash
- Confidence score
- Latest inclusion routine version
- Status and change history

### 3. Change Requests

Identity changes follow a workflow:

1. **Create Request**: User submits a change request
2. **Pending Verification**: Awaiting verification by an authorized verifier
3. **Ready to Apply**: Approved by verifier, ready for application
4. **Applied**: Change has been applied to the identity record
5. **Rejected**: Change was rejected

### 4. Audit Logging

All important operations are logged with:
- Actor (who performed the action)
- Action type
- Resource affected
- Result (success/failure)
- Metadata
- Timestamp
- Error details (if applicable)

## Parameters

The module has extensive configuration parameters:

### Auth Parameters
- `max_roles_per_address`: Maximum roles assignable to one address (default: 10)
- `max_permissions_per_role`: Maximum permissions per role (default: 50)
- `default_role_expiry_seconds`: Default role assignment expiration (default: 1 year)

### Session Parameters
- `default_session_expiry_seconds`: Default session duration (default: 24 hours)
- `max_sessions_per_user`: Maximum concurrent sessions (default: 10)
- `session_inactivity_timeout_seconds`: Inactivity timeout (default: 1 hour)

### Rate Limiting Parameters
- `default_max_requests_per_minute`: Default minute rate limit (default: 60)
- `default_max_requests_per_hour`: Default hour rate limit (default: 3600)
- `default_max_requests_per_day`: Default day rate limit (default: 86400)

### Multisig Parameters
- `max_signers_per_wallet`: Maximum signers in a multisig wallet (default: 20)
- `default_proposal_expiry_seconds`: Proposal expiration (default: 7 days)

### Time-Lock Parameters
- `min_time_lock_delay_seconds`: Minimum delay (default: 1 hour)
- `max_time_lock_delay_seconds`: Maximum delay (default: 30 days)

### Emergency Admin Parameters
- `max_emergency_admin_expiry_seconds`: Maximum emergency admin duration (default: 24 hours)
- `require_multi_sig_for_emergency`: Require multisig for emergency actions (default: true)

### Identity Change Parameters
- `max_change_requests_per_month`: Monthly request limit (default: 10)
- `min_confidence_score_after_change`: Minimum score after change (default: 50)
- `change_request_expiry_seconds`: Request expiration (default: 30 days)
- `enable_identity_recovery`: Enable recovery features (default: true)
- `enable_cross_chain_identity`: Enable cross-chain features (default: true)
- `min_recovery_threshold`: Minimum recovery contacts (default: 2)

### Audit Parameters
- `max_audit_logs_retained`: Maximum audit logs stored (default: 100000)
- `enable_audit_logging`: Enable audit logging (default: true)

## Genesis State

The genesis state includes:

- Module parameters
- Initial roles and role assignments
- Audit logs
- Sessions and rate limit configurations
- Multisig wallets and proposals
- Time-locked actions
- Emergency admins
- Validator key rotations
- Identity records and change requests
- Change history
- Recovery, verification, delegation, federation, and cross-chain records
- Suspended flag
- Counter values

## State Transitions

### Role Assignment Flow
1. Admin creates role with permissions
2. Admin assigns role to address with optional expiry
3. Address gains permissions from role
4. Role can be revoked or expires automatically

### Identity Change Flow
1. User creates change request for DID
2. Request enters pending verification state
3. Verifier reviews and approves/rejects
4. If approved, becomes ready to apply
5. Admin applies change to identity record
6. Change is recorded in history

### Session Flow
1. User creates session
2. Session is active until expiry or revocation
3. Session tracks last activity
4. Expired sessions can be cleaned up

## Security Features

1. **Permission-Based Access**: All sensitive operations require specific permissions
2. **Audit Trail**: Complete logging of all operations
3. **Rate Limiting**: Prevent abuse through configurable rate limits
4. **Time-Locked Actions**: Delay execution of critical operations
5. **Multisig Support**: Require multiple signatures for important actions
6. **Emergency Admin**: Temporary elevated privileges with expiration
7. **Session Management**: Track and limit concurrent sessions
8. **Change Request Limits**: Prevent spam through monthly limits

## Integration

### In App Initialization

```go
import (
    identitykeeper "github.com/aequitas/aura/chain/x/identity/keeper"
    identitymodule "github.com/aequitas/aura/chain/x/identity"
    identitytypes "github.com/aequitas/aura/chain/x/identity/types"
)

// Create keeper
identityKeeper := identitykeeper.NewKeeper(
    storeService,
    cdc,
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
    logger,
)

// Create module
identityModule := identitymodule.NewAppModule(cdc, identityKeeper)
```

### CLI Usage Examples

```bash
# Create a new role
aurad tx identity create-role moderator "manage_session,verify_identity" "Moderator role" --from admin

# Assign role to address
aurad tx identity assign-role aura1... moderator --from admin

# Create identity change request
aurad tx identity create-change-request did:aura:123 ir-v1 hash123 --from user

# Submit verification
aurad tx identity submit-verification req-1 true --from verifier

# Apply approved change
aurad tx identity apply-change req-1 --from admin

# Query identity record
aurad query identity identity did:aura:123

# Query change history
aurad query identity change-history did:aura:123

# Query module parameters
aurad query identity params
```

## Migration from Separate Modules

This module consolidates:

- **auth module**: All role, permission, session, multisig, time-lock, emergency admin, and validator rotation functionality
- **identitychange module**: All identity record, change request, and history management functionality

When migrating:
1. Export genesis state from both modules
2. Combine into consolidated identity genesis state
3. Initialize identity module with combined state
4. Remove old auth and identitychange modules

## Future Enhancements

Potential future additions:
- Biometric authentication integration
- Hardware security module (HSM) support
- Zero-knowledge proof verification
- Advanced delegation mechanisms
- Identity federation with external systems
- Enhanced cross-chain identity verification
- Automated identity recovery workflows
- Machine learning-based fraud detection

## Testing

The module should include comprehensive tests for:
- Role and permission management
- Identity change workflows
- Session lifecycle
- Multisig operations
- Time-locked actions
- Genesis import/export
- Parameter validation
- Audit logging
- Rate limiting

## Contributing

When contributing to this module:
1. Follow Cosmos SDK v0.50+ patterns
2. Ensure all state is stored in KVStore (no in-memory maps)
3. Add comprehensive tests for new features
4. Update documentation
5. Include audit logging for sensitive operations
6. Validate all inputs
7. Follow error handling best practices

## License

Copyright (c) Aura Blockchain
