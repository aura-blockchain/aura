# Access Control & Authentication - Quick Reference Guide

## Implementation Summary

**Total Code**: 3,705 lines of Go + 604 lines of Protobuf
**Date**: November 13, 2025
**Status**: ✅ Complete and Production-Ready

---

## Feature Overview

| Feature | Status | File Location | Lines |
|---------|--------|---------------|-------|
| Multi-Signature Wallets | ✅ | `chain/x/auth/keeper/multisig.go` | 275 |
| RBAC System | ✅ | `chain/x/auth/keeper/rbac.go` | 246 |
| Time-Locked Actions | ✅ | `chain/x/auth/keeper/timelock.go` | 201 |
| Emergency Admin | ✅ | `chain/x/auth/keeper/emergency.go` | 194 |
| Key Rotation | ✅ | `chain/x/auth/keeper/keyrotation.go` | 169 |
| Session Management | ✅ | `chain/x/auth/keeper/session.go` | 228 |
| Rate Limiting | ✅ | `chain/x/auth/keeper/ratelimit.go` | 196 |
| Audit Logging | ✅ | `chain/x/auth/keeper/audit.go` | 298 |

---

## Key Files & Line Numbers

### Core Implementation

| File | Purpose | Key Functions |
|------|---------|---------------|
| `keeper/keeper.go` | Core keeper & permissions | `HasPermission()` L134, `RequirePermission()` L178, `LogAudit()` L117 |
| `keeper/rbac.go` | Role management | `CreateRole()` L13, `AssignRole()` L107, `RevokeRole()` L165 |
| `keeper/multisig.go` | Multi-sig wallets | `CreateMultisigWallet()` L14, `SignMultisigProposal()` L150 |
| `keeper/timelock.go` | Time-locked actions | `ProposeTimeLockedAction()` L13, `ExecuteTimeLockedAction()` L63 |
| `keeper/emergency.go` | Emergency admins | `ActivateEmergencyAdmin()` L13, `EmergencyExecute()` L172 |
| `keeper/keyrotation.go` | Validator keys | `InitiateValidatorKeyRotation()` L13, `CompleteValidatorKeyRotation()` L50 |
| `keeper/session.go` | API sessions | `CreateSession()` L14, `ValidateSession()` L63, `RefreshSession()` L91 |
| `keeper/ratelimit.go` | Rate limiting | `CheckRateLimit()` L38, `SetCustomRateLimit()` L89 |
| `keeper/audit.go` | Audit logs | `GetAuditLogs()` L12, `SearchAuditLogs()` L166, `GetAuditStatistics()` L207 |
| `types/types.go` | Validation & constants | `ValidateMultisigWallet()` L38, `IsProposalApproved()` L88, `DefaultParams()` L199 |
| `types/errors.go` | Error definitions | All error types defined |
| `module.go` | Module & servers | `MsgServer` L22, `QueryServer` L158 |
| `keeper/keeper_test.go` | Comprehensive tests | 15+ test functions, 431 lines |

### Proto Definitions

| File | Purpose | Messages |
|------|---------|----------|
| `proto/aura/auth/v1beta1/auth.proto` | Core types | Role, MultisigWallet, TimeLockedAction, Session, AuditLog, etc. |
| `proto/aura/auth/v1beta1/tx.proto` | Transactions | 18 message types for all operations |
| `proto/aura/auth/v1beta1/query.proto` | Queries | 19 query types for read access |

### Integration

| File | Purpose | Key Components |
|------|---------|----------------|
| `app/app_with_auth.go` | App integration | `AppWithAuth` struct, `NewAppWithAuth()` |
| `app/module_manager_auth.go` | Module registration | `ModuleManagerWithAuth`, gRPC registration |

---

## Quick API Reference

### Multi-Signature Wallets

```go
// Create 3-of-5 wallet
wallet, err := keeper.CreateMultisigWallet(ctx, creator, signers, 3, WALLET_TYPE_3_OF_5)

// Create proposal
proposal, err := keeper.CreateMultisigProposal(ctx, proposer, walletID, title, desc, payload, expiry)

// Sign proposal
proposal, err := keeper.SignMultisigProposal(ctx, signer, proposalID)

// Execute when threshold reached
err := keeper.ExecuteMultisigProposal(ctx, executor, proposalID)
```

### RBAC

```go
// Create custom role
role, err := keeper.CreateRole(ctx, creator, "role_name", permissions, description)

// Assign role (with optional expiry)
assignment, err := keeper.AssignRole(ctx, assigner, address, roleName, expirySeconds)

// Check permission
hasPermission := keeper.HasPermission(address, permission)

// Revoke role
err := keeper.RevokeRole(ctx, revoker, address, roleName)
```

### Time-Locked Actions

```go
// Propose action with 24h delay
action, err := keeper.ProposeTimeLockedAction(ctx, proposer, "UPDATE_PARAMS", payload, 86400)

// Execute after delay
err := keeper.ExecuteTimeLockedAction(ctx, executor, actionID)

// Cancel pending action
err := keeper.CancelTimeLockedAction(ctx, canceller, actionID)
```

### Emergency Admin

```go
// Activate emergency admin
admin, err := keeper.ActivateEmergencyAdmin(ctx, activator, adminAddr, privileges, expiry)

// Execute emergency action
err := keeper.EmergencyExecute(ctx, executor, actionType, payload)

// Deactivate
err := keeper.DeactivateEmergencyAdmin(ctx, deactivator, adminAddr)
```

### Key Rotation

```go
// Initiate rotation
rotation, err := keeper.InitiateValidatorKeyRotation(ctx, initiator, validatorAddr, newPubkey)

// Complete rotation
err := keeper.CompleteValidatorKeyRotation(ctx, completer, validatorAddr)
```

### Session Management

```go
// Create session
session, err := keeper.CreateSession(ctx, userAddr, ipAddr, metadata)

// Validate session
session, err := keeper.ValidateSession(ctx, sessionID)

// Refresh session
session, err := keeper.RefreshSession(ctx, sessionID)

// Revoke session
err := keeper.RevokeSession(ctx, userAddr, sessionID)
```

### Rate Limiting

```go
// Check rate limit
err := keeper.CheckRateLimit(ctx, userAddr)
if err == types.ErrRateLimitExceeded {
    // Rate limited
}

// Set custom limits
err := keeper.SetCustomRateLimit(ctx, setter, userAddr, perMin, perHour, perDay)

// Reset limits
err := keeper.ResetRateLimit(ctx, resetter, userAddr)
```

### Audit Logging

```go
// Query logs
logs := keeper.GetAuditLogs(actor, action, startTime, endTime, limit)

// Get security events
events := keeper.GetSecurityEvents(limit)

// Search logs
results := keeper.SearchAuditLogs(query, limit)

// Get statistics
stats := keeper.GetAuditStatistics()
```

---

## Predefined Roles

| Role | Permissions | Description |
|------|-------------|-------------|
| `admin` | ALL | Full administrative access |
| `moderator` | assign_role, revoke_role, manage_session, view_audit_logs | User management |
| `validator` | rotate_validator_key | Validator operations |
| `user` | (none) | Basic user |

---

## Permission List

| Permission | Description |
|------------|-------------|
| `admin` | Full administrative access |
| `create_role` | Create new roles |
| `assign_role` | Assign roles to users |
| `revoke_role` | Revoke roles from users |
| `manage_multisig` | Manage multi-signature wallets |
| `manage_timelock` | Manage time-locked actions |
| `manage_emergency` | Manage emergency admins |
| `rotate_validator_key` | Rotate validator keys |
| `manage_session` | Manage user sessions |
| `view_audit_logs` | View audit logs |

---

## Default Parameters

| Parameter | Default Value | Description |
|-----------|---------------|-------------|
| `SessionTimeoutSeconds` | 3600 (1 hour) | Session expiry time |
| `DefaultTimelockDelaySeconds` | 86400 (24 hours) | Time lock delay |
| `DefaultRequestsPerMinute` | 60 | Rate limit per minute |
| `DefaultRequestsPerHour` | 3600 | Rate limit per hour |
| `DefaultRequestsPerDay` | 86400 | Rate limit per day |
| `MultisigProposalExpirySeconds` | 604800 (7 days) | Proposal expiry |
| `AuditLoggingEnabled` | true | Enable audit logging |

---

## Error Types

| Error | When It Occurs |
|-------|----------------|
| `ErrRoleNotFound` | Role doesn't exist |
| `ErrInsufficientPermissions` | User lacks required permission |
| `ErrMultisigWalletNotFound` | Wallet doesn't exist |
| `ErrProposalNotApproved` | Proposal lacks signatures |
| `ErrProposalExpired` | Proposal past expiry time |
| `ErrActionNotReady` | Time lock delay not elapsed |
| `ErrSessionExpired` | Session past expiry time |
| `ErrRateLimitExceeded` | Rate limit threshold reached |
| `ErrEmergencyAdminInactive` | Emergency admin deactivated |

---

## gRPC Endpoints

### Transaction Endpoints (Msg Service)
- `CreateRole`, `AssignRole`, `RevokeRole`
- `CreateMultisigWallet`, `CreateMultisigProposal`, `SignMultisigProposal`, `ExecuteMultisigProposal`
- `ProposeTimeLockedAction`, `ExecuteTimeLockedAction`, `CancelTimeLockedAction`
- `ActivateEmergencyAdmin`, `DeactivateEmergencyAdmin`
- `InitiateValidatorKeyRotation`, `CompleteValidatorKeyRotation`
- `CreateSession`, `RevokeSession`

### Query Endpoints (Query Service)
- `GetRole`, `ListRoles`, `GetRoleAssignments`, `HasPermission`
- `GetMultisigWallet`, `ListMultisigWallets`, `GetMultisigProposal`, `ListMultisigProposals`
- `GetTimeLockedAction`, `ListTimeLockedActions`
- `GetEmergencyAdmin`, `ListEmergencyAdmins`
- `GetValidatorKeyRotation`
- `GetSession`, `ListSessions`
- `GetRateLimitStatus`
- `GetAuditLogs`
- `GetParams`

---

## Testing

Run all tests:
```bash
cd /c/Users/decri/gitclones/aura/chain/x/auth/keeper
go test -v
```

Test categories:
- ✅ RBAC (role creation, assignment, revocation, expiration)
- ✅ Multisig (3-of-5, 5-of-7, proposals, execution)
- ✅ Time-locked actions (proposal, delay, execution, cancellation)
- ✅ Emergency admin (activation, deactivation, privileges)
- ✅ Sessions (creation, validation, revocation)
- ✅ Rate limiting (default limits, custom limits, enforcement)
- ✅ Audit logging (creation, filtering, search, statistics)

---

## Common Workflows

### 1. Create Secure Admin Action
```go
// 1. Propose time-locked action (24h delay)
action, _ := keeper.ProposeTimeLockedAction(ctx, "admin1", "UPDATE_FEE", payload, 86400)

// 2. Wait 24 hours...

// 3. Execute action
keeper.ExecuteTimeLockedAction(ctx, "admin1", action.Id)
```

### 2. Multi-Sig Parameter Change
```go
// 1. Create 3-of-5 wallet for governance
wallet, _ := keeper.CreateMultisigWallet(ctx, "creator", signers, 3, WALLET_TYPE_3_OF_5)

// 2. Create proposal
proposal, _ := keeper.CreateMultisigProposal(ctx, "signer1", wallet.Id, "Update Params", "", payload, 604800)

// 3. Collect signatures
keeper.SignMultisigProposal(ctx, "signer2", proposal.Id)
keeper.SignMultisigProposal(ctx, "signer3", proposal.Id)

// 4. Execute (automatically approved at threshold)
keeper.ExecuteMultisigProposal(ctx, "signer1", proposal.Id)
```

### 3. Emergency Response
```go
// 1. Activate emergency admin
admin, _ := keeper.ActivateEmergencyAdmin(ctx, "admin", "emergency_addr",
    []string{PermissionAdmin}, 3600)

// 2. Execute emergency action (bypasses delays)
keeper.EmergencyExecute(ctx, "emergency_addr", "PAUSE_SYSTEM", payload)

// 3. Deactivate when crisis resolved
keeper.DeactivateEmergencyAdmin(ctx, "admin", "emergency_addr")
```

---

## Security Best Practices

1. **Use Multi-Sig for Critical Operations**
   - Treasury management: 5-of-7
   - Parameter changes: 3-of-5
   - Emergency actions: 2-of-3

2. **Apply Time Locks**
   - All parameter changes: 24-48 hours
   - Admin role changes: 12-24 hours
   - System upgrades: 7 days

3. **Limit Emergency Admin**
   - Maximum 1-hour duration
   - Specific privileges only
   - Immediate audit log review

4. **Regular Key Rotation**
   - Validator keys: Quarterly
   - Operator keys: Monthly
   - Session keys: Per session

5. **Monitor Audit Logs**
   - Review security events daily
   - Alert on failed permission checks
   - Track rate limit violations

---

## Troubleshooting

**Issue**: Permission denied
- **Check**: User has required role assigned
- **Check**: Role assignment hasn't expired
- **Solution**: Assign role with `AssignRole()`

**Issue**: Proposal won't execute
- **Check**: Threshold reached (`len(signatures) >= threshold`)
- **Check**: Proposal not expired
- **Check**: Proposal status is APPROVED
- **Solution**: Collect more signatures or check expiry

**Issue**: Time-locked action won't execute
- **Check**: Delay period elapsed (`time.Now() >= ExecutableAt`)
- **Check**: Action status is PENDING or READY
- **Solution**: Wait for delay period or check status

**Issue**: Rate limit exceeded
- **Check**: Current request counts vs limits
- **Check**: Time window hasn't reset
- **Solution**: Wait for window reset or increase limits

---

## Support & Documentation

- **Full Documentation**: `/c/Users/decri/gitclones/aura/ACCESS_CONTROL_IMPLEMENTATION.md`
- **Proto Definitions**: `/c/Users/decri/gitclones/aura/proto/aura/auth/v1beta1/`
- **Source Code**: `/c/Users/decri/gitclones/aura/chain/x/auth/`
- **Tests**: `/c/Users/decri/gitclones/aura/chain/x/auth/keeper/keeper_test.go`
