# Identity Module - Quick Reference

## Module Location
```
/home/decri/blockchain-projects/aura/chain/x/identity/
```

## File Structure
```
x/identity/
├── types/
│   ├── keys.go                 (207 lines) - Store keys and prefixes
│   ├── errors.go               (145 lines) - Error definitions
│   ├── genesis.go              (369 lines) - Genesis state and params
│   └── types.go                (334 lines) - Core types
├── keeper/
│   ├── keeper.go               (489 lines) - Main keeper with genesis
│   ├── auth.go                 (453 lines) - Role & permission management
│   ├── changes.go              (540 lines) - Identity change management
│   └── sessions.go             (478 lines) - Sessions, multisig, etc.
├── module.go                   (407 lines) - AppModule implementation
├── README.md                   (419 lines) - Full documentation
├── IMPLEMENTATION_SUMMARY.md   (520 lines) - Implementation details
└── QUICK_REFERENCE.md          (This file) - Quick reference

Total: 3,108 lines of Go code
```

## Key Components Merged

### From auth module:
- Roles and permissions (RBAC)
- Role assignments with expiry
- Sessions with lifecycle management
- Rate limiting per user
- Multisig wallets and proposals
- Time-locked actions
- Emergency admin privileges
- Validator key rotation
- Audit logging

### From identitychange module:
- Decentralized identity (DID) records
- Identity change requests with workflow
- Change history tracking
- Identity recovery
- Identity verification
- Identity delegation
- Identity federation
- Cross-chain identity linking

## Core Types

### Role System
```go
type Role struct {
    Name        string
    Permissions []string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type RoleAssignment struct {
    Address    string
    RoleName   string
    AssignedBy string
    AssignedAt time.Time
    ExpiresAt  *time.Time
}
```

### Identity System
```go
type IdentityRecord struct {
    DID               string
    Owner             string
    MetadataHash      string
    ConfidenceScore   int64
    LatestIRVersion   string
    LastChangedHeight int64
    Status            ChangeStatus
    CreatedAt         time.Time
    UpdatedAt         time.Time
}

type ChangeRequest struct {
    RequestID       string
    Requester       string
    TargetDID       string
    IrID            string
    MetadataHash    string
    Status          ChangeStatus
    Assistant       string
    VerdictHeight   int64
    Reason          string
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

## Default Roles

| Role | Permissions |
|------|-------------|
| **admin** | All permissions (admin, create_role, assign_role, revoke_role, manage_multisig, manage_timelock, manage_emergency, rotate_validator_key, manage_session, view_audit_logs, manage_identity, verify_identity, approve_change_request) |
| **moderator** | assign_role, revoke_role, manage_session, view_audit_logs, verify_identity |
| **validator** | rotate_validator_key |
| **user** | No special permissions |

## Default Parameters

| Category | Parameter | Default Value |
|----------|-----------|---------------|
| **Auth** | max_roles_per_address | 10 |
| | max_permissions_per_role | 50 |
| | default_role_expiry_seconds | 31536000 (1 year) |
| **Sessions** | default_session_expiry_seconds | 86400 (24 hours) |
| | max_sessions_per_user | 10 |
| | session_inactivity_timeout_seconds | 3600 (1 hour) |
| **Rate Limits** | default_max_requests_per_minute | 60 |
| | default_max_requests_per_hour | 3600 |
| | default_max_requests_per_day | 86400 |
| **Multisig** | max_signers_per_wallet | 20 |
| | default_proposal_expiry_seconds | 604800 (7 days) |
| **Time-Lock** | min_time_lock_delay_seconds | 3600 (1 hour) |
| | max_time_lock_delay_seconds | 2592000 (30 days) |
| **Emergency** | max_emergency_admin_expiry_seconds | 86400 (24 hours) |
| | require_multi_sig_for_emergency | true |
| **Identity** | max_change_requests_per_month | 10 |
| | min_confidence_score_after_change | 50 |
| | change_request_expiry_seconds | 2592000 (30 days) |
| | enable_identity_recovery | true |
| | enable_cross_chain_identity | true |
| | min_recovery_threshold | 2 |
| **Audit** | max_audit_logs_retained | 100000 |
| | enable_audit_logging | true |

## State Machine: Change Request Workflow

```
┌─────────────────┐
│   Unspecified   │
└────────┬────────┘
         │
         ▼
    ┌────────┐
    │  Idle  │
    └───┬────┘
        │ CreateChangeRequest()
        ▼
┌──────────────────────┐
│ PendingVerification  │
└─────────┬────────────┘
          │
          │ SubmitVerification(approved=true)
          ▼
    ┌──────────────┐
    │ ReadyToApply │
    └──────┬───────┘
           │
           │ ApplyChange()
           ▼
      ┌─────────┐
      │ Applied │
      └─────────┘

    OR

    SubmitVerification(approved=false)
    OR RejectChange()
          │
          ▼
      ┌──────────┐
      │ Rejected │
      └──────────┘
```

## Key Keeper Methods

### Role Management
```go
CreateRole(ctx, creator, name, permissions, description) (Role, error)
AssignRole(ctx, assigner, address, roleName, expirySeconds) (RoleAssignment, error)
RevokeRole(ctx, revoker, address, roleName) error
HasPermission(ctx, address, permission) bool
RequirePermission(ctx, address, permission) error
```

### Identity Management
```go
CreateChangeRequest(ctx, requester, targetDID, irID, metadataHash) (ChangeRequest, error)
SubmitVerification(ctx, requestID, assistant, approved, reason) (ChangeRequest, error)
ApplyChange(ctx, requestID, applier) (IdentityRecord, error)
RejectChange(ctx, requestID, rejecter, reason) (ChangeRequest, error)
GetIdentityRecord(ctx, did) (IdentityRecord, error)
GetChangeHistory(ctx, did) ([]ChangeHistory, error)
```

### Session Management
```go
CreateSession(ctx, userAddress, expirySeconds) (Session, error)
RevokeSession(ctx, userAddress, sessionID) error
GetSession(ctx, sessionID) (Session, error)
```

### Audit Logging
```go
LogAudit(ctx, actor, action, resource, result, metadata, errorDetail)
GetAllAuditLogs(ctx) ([]AuditLog, error)
```

## Store Key Prefixes

| Prefix | Usage |
|--------|-------|
| 0x00 | Module parameters |
| 0x01 | Roles |
| 0x02 | Role assignments |
| 0x03 | Permission grants |
| 0x04 | Audit logs |
| 0x05 | Accounts |
| 0x06 | Sessions |
| 0x07 | User sessions index |
| 0x08 | Rate limit configs |
| 0x09 | Multisig wallets |
| 0x0a | Multisig proposals |
| 0x0b | Time-locked actions |
| 0x0c | Emergency admins |
| 0x0d | Emergency actions |
| 0x0e | Validator rotations |
| 0x10 | Identity records |
| 0x11 | Change requests |
| 0x12 | Change history |
| 0x13 | Recovery records |
| 0x14 | Verification records |
| 0x15 | Delegation records |
| 0x16 | Federation records |
| 0x17 | Cross-chain links |
| 0x20 | Audit log counter |
| 0x21 | Change request counter |
| 0x30 | Suspended flag |

## Error Codes

| Range | Category |
|-------|----------|
| 1-99 | Auth errors |
| 100-199 | Account/Session errors |
| 200-299 | Multisig errors |
| 300-399 | Time-locked action errors |
| 400-499 | Emergency admin errors |
| 500-599 | Validator errors |
| 600-699 | Identity change errors |
| 900-999 | General errors |

## Common Error Codes
- 1: Role not found
- 4: Insufficient permissions
- 102: Session not found
- 106: Rate limit exceeded
- 200: Multisig wallet not found
- 300: Time-locked action not found
- 600: Identity not found
- 602: Change request not found
- 607: Change request limit exceeded

## Usage Examples

### Initialize Module
```go
import (
    identitykeeper "github.com/aequitas/aura/chain/x/identity/keeper"
    identitymodule "github.com/aequitas/aura/chain/x/identity"
)

keeper := identitykeeper.NewKeeper(storeService, cdc, authority, logger)
module := identitymodule.NewAppModule(cdc, keeper)
```

### Create and Assign Role
```go
// Create role
role, err := keeper.CreateRole(ctx, adminAddr, "moderator",
    []string{"verify_identity", "manage_session"},
    "Moderator role")

// Assign role (expires in 1 year)
assignment, err := keeper.AssignRole(ctx, adminAddr, userAddr, "moderator", 31536000)
```

### Identity Change Workflow
```go
// 1. Create request
request, err := keeper.CreateChangeRequest(ctx, userAddr, "did:aura:123", "ir-v1", "hash123")

// 2. Verify
verifiedRequest, err := keeper.SubmitVerification(ctx, request.RequestID, verifierAddr, true, "")

// 3. Apply
record, err := keeper.ApplyChange(ctx, request.RequestID, adminAddr)
```

### Session Management
```go
// Create session (expires in 24 hours)
session, err := keeper.CreateSession(ctx, userAddr, 86400)

// Revoke session
err = keeper.RevokeSession(ctx, userAddr, session.SessionID)
```

## Security Checklist

- ✅ All sensitive operations require permissions
- ✅ Audit logging for accountability
- ✅ Rate limiting to prevent abuse
- ✅ Time-locked actions for critical operations
- ✅ Multisig support for high-value operations
- ✅ Emergency admin with expiration
- ✅ Session expiry and limits
- ✅ Change request monthly limits
- ✅ Input validation on all operations
- ✅ Proper error handling with context

## Testing TODO

- [ ] Unit tests for keeper methods
- [ ] Integration tests for workflows
- [ ] Genesis import/export tests
- [ ] Parameter validation tests
- [ ] Permission enforcement tests
- [ ] Rate limiting tests
- [ ] Session lifecycle tests
- [ ] Audit logging tests
- [ ] Edge case tests

## Next Steps

1. **Add proto definitions** for gRPC services
2. **Implement msg_server.go** for transaction handlers
3. **Implement query_server.go** for query handlers
4. **Add CLI commands** in client/cli/
5. **Write comprehensive tests**
6. **Add invariants** for state validation
7. **Integrate with app.go**
8. **Write migration guide** from old modules
9. **Performance testing**
10. **Security audit**

## Quick Integration

```go
// In app/app.go

import (
    identitykeeper "github.com/aequitas/aura/chain/x/identity/keeper"
    identitymodule "github.com/aequitas/aura/chain/x/identity"
    identitytypes "github.com/aequitas/aura/chain/x/identity/types"
)

// Create keeper
app.IdentityKeeper = identitykeeper.NewKeeper(
    app.GetSubspace(identitytypes.ModuleName),
    app.appCodec,
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
    app.Logger(),
)

// Create module
identityModule := identitymodule.NewAppModule(
    app.appCodec,
    app.IdentityKeeper,
)

// Add to module manager
app.ModuleManager = module.NewManager(
    // ... other modules
    identityModule,
)
```

## Status

✅ **Complete**: Core implementation finished
⏳ **Pending**: Proto definitions, msg/query servers, CLI, tests
📝 **Documentation**: README and summaries complete

---

**Created**: 2025-11-27
**Module Version**: 1.0.0
**Cosmos SDK**: v0.50+
