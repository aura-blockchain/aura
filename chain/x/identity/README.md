# Identity Module

## Overview

The Identity module provides comprehensive identity management, authentication, role-based access control (RBAC), multisig wallets, time-locked actions, emergency admin capabilities, and session management for the Aura blockchain. It implements decentralized identity (DID) management with support for identity change workflows, validator key rotation, and GDPR-compliant data erasure.

## Features

- **Identity Change Management**: Controlled workflow for identity modifications with assistant proof verification
- **Role-Based Access Control (RBAC)**: Flexible permission system with role assignments and expirations
- **Multisig Wallets**: Multi-signature wallet management with proposal-based operations
- **Time-Locked Actions**: Delayed execution of sensitive operations with cancellation capability
- **Emergency Admin**: Temporary elevated privileges for emergency response
- **Validator Key Rotation**: Secure validator consensus key rotation
- **Session Management**: User session tracking with device fingerprinting
- **DID Key Rotation**: Grace-period based key rotation for DIDs
- **GDPR Compliance**: Right to erasure implementation

## State

### Identity Records
- **IdentityRecord**: Core identity information with DID, controller address, status, and metadata
- **ChangeRequest**: Pending identity change requests with verification requirements
- **ChangeHistory**: Immutable audit trail of identity modifications

### Roles and Permissions
- **Role**: Named permission sets (e.g., "admin", "operator")
- **RoleAssignment**: Role-to-address mappings with expiration timestamps
- **Permissions**: Granular capabilities (manage_multisig, manage_timelock, manage_emergency, etc.)

### Multisig Wallets
- **MultisigWallet**: Wallet with multiple signers and threshold requirement
- **MultisigProposal**: Proposals requiring threshold signatures before execution
- **Wallet Types**: Individual, corporate, government, DAO

### Time-Locked Actions
- **TimeLockedAction**: Actions with mandatory delay period before execution
- **Action Status**: Pending, executed, cancelled

### Emergency Admin
- **EmergencyAdmin**: Temporary admin with specific privileges and expiration
- **Privileges**: Array of elevated permissions

### Sessions
- **Session**: User authentication session with device fingerprint and IP metadata
- **Session Expiration**: Configurable timeout period

### Validator Key Rotation
- **ValidatorKeyRotation**: Consensus public key rotation with status tracking

## Messages

### Identity Change Operations

#### MsgRequestIdentityChange
Request an identity change with incident response validation.

**Example**:
```json
{
  "requester": "aura1...",
  "target_did": "did:aura:...",
  "metadata_hash": "sha256_hash_of_change_data",
  "ir_id": "incident_response_id",
  "proof_hash": "sha256_proof"
}
```

**Response**:
```json
{
  "request_id": "req_123456"
}
```

#### MsgSubmitAssistantProof
AI assistant submits proof for identity verification.

**Example**:
```json
{
  "assistant": "aura1assistant...",
  "request_id": "req_123456",
  "proof_hash": "verification_proof_hash",
  "confidence_delta": 50,
  "success": true
}
```

#### MsgApplyIdentityChange
Apply an approved identity change.

**Example**:
```json
{
  "requester": "aura1...",
  "request_id": "req_123456"
}
```

#### MsgRejectIdentityChange
Reject an identity change request.

**Example**:
```json
{
  "actor": "aura1admin...",
  "request_id": "req_123456",
  "reason": "Insufficient verification"
}
```

#### MsgSuspendIdentityChanges
Suspend all identity changes system-wide (authority only).

**Example**:
```json
{
  "authority": "aura1gov...",
  "reason": "Security incident detected"
}
```

### Role Management Operations

#### MsgCreateRole
Create a new role with permissions.

**Example**:
```json
{
  "creator": "aura1admin...",
  "role_name": "operator",
  "description": "System operator role",
  "permissions": [
    "manage_multisig",
    "manage_timelock"
  ]
}
```

#### MsgAssignRole
Assign a role to an address with optional expiration.

**Example**:
```json
{
  "assigner": "aura1admin...",
  "address": "aura1user...",
  "role_name": "operator",
  "expires_at": "2025-12-31T23:59:59Z"
}
```

#### MsgRevokeRole
Revoke a role from an address.

**Example**:
```json
{
  "revoker": "aura1admin...",
  "address": "aura1user...",
  "role_name": "operator"
}
```

### Multisig Operations

#### MsgCreateMultisigWallet
Create a new multisig wallet.

**Example**:
```json
{
  "creator": "aura1...",
  "signers": [
    "aura1signer1...",
    "aura1signer2...",
    "aura1signer3..."
  ],
  "threshold": 2,
  "wallet_type": "WALLET_TYPE_INDIVIDUAL"
}
```

**Response**:
```json
{
  "wallet_id": "mswallet-aura1...-1702345678"
}
```

#### MsgCreateMultisigProposal
Create a proposal for multisig wallet action.

**Example**:
```json
{
  "proposer": "aura1signer1...",
  "wallet_id": "mswallet-...",
  "title": "Transfer funds to treasury",
  "description": "Quarterly treasury allocation",
  "payload": "base64_encoded_tx_data"
}
```

**Response**:
```json
{
  "proposal_id": "msprop-mswallet-...-1702345678"
}
```

#### MsgSignMultisigProposal
Sign a multisig proposal.

**Example**:
```json
{
  "signer": "aura1signer2...",
  "proposal_id": "msprop-..."
}
```

#### MsgExecuteMultisigProposal
Execute an approved multisig proposal.

**Example**:
```json
{
  "executor": "aura1signer1...",
  "proposal_id": "msprop-..."
}
```

### Time-Locked Operations

#### MsgProposeTimeLockedAction
Propose an action with mandatory delay.

**Example**:
```json
{
  "proposer": "aura1admin...",
  "action_type": "parameter_change",
  "payload": "base64_encoded_action_data",
  "delay_seconds": 172800
}
```

**Response**:
```json
{
  "action_id": "tlaction-...",
  "executable_at": "2025-12-11T10:00:00Z"
}
```

#### MsgExecuteTimeLockedAction
Execute a time-locked action after delay expires.

**Example**:
```json
{
  "executor": "aura1admin...",
  "action_id": "tlaction-..."
}
```

#### MsgCancelTimeLockedAction
Cancel a pending time-locked action.

**Example**:
```json
{
  "canceller": "aura1admin...",
  "action_id": "tlaction-..."
}
```

### Emergency Operations

#### MsgActivateEmergencyAdmin
Activate an emergency admin with temporary privileges.

**Example**:
```json
{
  "activator": "aura1gov...",
  "admin_address": "aura1emergency...",
  "privileges": [
    "emergency_pause",
    "emergency_upgrade"
  ],
  "expires_at": "2025-12-10T10:00:00Z"
}
```

#### MsgDeactivateEmergencyAdmin
Deactivate an emergency admin.

**Example**:
```json
{
  "deactivator": "aura1gov...",
  "admin_address": "aura1emergency..."
}
```

### Validator Operations

#### MsgRotateValidatorKey
Rotate validator consensus public key.

**Example**:
```json
{
  "validator_address": "auravaloper1...",
  "new_consensus_pubkey": "auravalconspub1..."
}
```

### Session Operations

#### MsgCreateSession
Create a new user session.

**Example**:
```json
{
  "address": "aura1...",
  "device_fingerprint": "device_hash_123",
  "ip_address": "192.168.1.1",
  "metadata": {
    "user_agent": "Mozilla/5.0...",
    "platform": "Linux"
  }
}
```

**Response**:
```json
{
  "session_id": "sess_abc123...",
  "expires_at": "2025-12-09T11:00:00Z"
}
```

#### MsgEndSession
End an active session.

**Example**:
```json
{
  "address": "aura1...",
  "session_id": "sess_abc123..."
}
```

### GDPR Operations

#### MsgEraseIdentity
Request identity erasure (GDPR Right to Erasure).

**Example**:
```json
{
  "requester": "aura1...",
  "did": "did:aura:...",
  "reason": "User requested data deletion"
}
```

**Response**:
```json
{
  "erased_at": "2025-12-09T10:00:00Z"
}
```

#### MsgRotateDIDKey
Rotate DID verification key with grace period.

**Example**:
```json
{
  "initiator": "aura1...",
  "did": "did:aura:...",
  "new_verification_method": "new_pubkey_id",
  "reason": "Key compromise suspected"
}
```

**Response**:
```json
{
  "rotation_time": "2025-12-09T10:00:00Z",
  "grace_period_end": "2025-12-16T10:00:00Z"
}
```

### Admin Operations

#### MsgUpdateParams
Update module parameters (authority only).

**Example**:
```json
{
  "authority": "aura1gov...",
  "params": {
    "auth": {
      "session_timeout": "3600s",
      "multisig_proposal_expiry_seconds": 604800,
      "max_roles_per_address": 10
    },
    "change": {
      "verification_threshold": 2,
      "change_cooldown_seconds": 86400
    }
  }
}
```

## Queries

### QueryParams
Get module parameters.

**Request**:
```bash
aurad query identity params
```

**Response**:
```json
{
  "params": {
    "auth": {
      "session_timeout": "3600s",
      "multisig_proposal_expiry_seconds": 604800
    },
    "change": {
      "verification_threshold": 2
    }
  }
}
```

### QueryIdentityRecord
Query identity record by DID.

**Request**:
```bash
aurad query identity identity-record did:aura:mainnet:abc123
```

**Response**:
```json
{
  "record": {
    "did": "did:aura:mainnet:abc123",
    "controller": "aura1...",
    "status": "active",
    "created_at": "2025-01-01T00:00:00Z",
    "metadata": {}
  }
}
```

### QueryChangeRequest
Query a change request by ID.

**Request**:
```bash
aurad query identity change-request req_123456
```

**Response**:
```json
{
  "request": {
    "id": "req_123456",
    "requester": "aura1...",
    "target_did": "did:aura:...",
    "status": "pending",
    "verifications_received": 1,
    "verifications_required": 2
  }
}
```

### QueryRole
Query role by name.

**Request**:
```bash
aurad query identity role operator
```

**Response**:
```json
{
  "role": {
    "name": "operator",
    "permissions": ["manage_multisig", "manage_timelock"],
    "description": "System operator role"
  }
}
```

### QueryRoleAssignments
Query role assignments for an address.

**Request**:
```bash
aurad query identity role-assignments aura1...
```

**Response**:
```json
{
  "assignments": [
    {
      "address": "aura1...",
      "role_name": "operator",
      "assigned_at": "2025-01-01T00:00:00Z",
      "expires_at": "2025-12-31T23:59:59Z"
    }
  ]
}
```

### QueryHasPermission
Check if address has specific permission.

**Request**:
```bash
aurad query identity has-permission aura1... manage_multisig
```

**Response**:
```json
{
  "has_permission": true
}
```

### QueryMultisigWallet
Query multisig wallet by ID.

**Request**:
```bash
aurad query identity multisig-wallet mswallet-...
```

**Response**:
```json
{
  "wallet": {
    "id": "mswallet-...",
    "signers": ["aura1signer1...", "aura1signer2...", "aura1signer3..."],
    "threshold": 2,
    "created_at": "2025-01-01T00:00:00Z",
    "wallet_type": "WALLET_TYPE_INDIVIDUAL"
  }
}
```

### QueryMultisigProposal
Query multisig proposal by ID.

**Request**:
```bash
aurad query identity multisig-proposal msprop-...
```

**Response**:
```json
{
  "proposal": {
    "id": "msprop-...",
    "wallet_id": "mswallet-...",
    "title": "Transfer funds",
    "status": "pending",
    "signatures": ["aura1signer1..."],
    "expires_at": "2025-12-16T00:00:00Z"
  }
}
```

### QueryTimeLockedAction
Query time-locked action by ID.

**Request**:
```bash
aurad query identity time-locked-action tlaction-...
```

**Response**:
```json
{
  "action": {
    "id": "tlaction-...",
    "action_type": "parameter_change",
    "proposer": "aura1admin...",
    "proposed_at": "2025-12-09T10:00:00Z",
    "executable_at": "2025-12-11T10:00:00Z",
    "status": "pending"
  }
}
```

### QueryEmergencyAdmin
Query emergency admin by address.

**Request**:
```bash
aurad query identity emergency-admin aura1emergency...
```

**Response**:
```json
{
  "admin": {
    "address": "aura1emergency...",
    "privileges": ["emergency_pause", "emergency_upgrade"],
    "activated_at": "2025-12-09T10:00:00Z",
    "expires_at": "2025-12-10T10:00:00Z",
    "is_active": true
  }
}
```

## Events

| Event Type | Attributes | Description |
|------------|------------|-------------|
| `identity_change_requested` | `request_id`, `requester`, `target_did` | Identity change request created |
| `assistant_proof_submitted` | `request_id`, `assistant`, `success` | Assistant verification submitted |
| `identity_change_applied` | `request_id`, `did` | Identity change successfully applied |
| `identity_change_rejected` | `request_id`, `reason` | Identity change request rejected |
| `identity_changes_suspended` | `authority`, `reason` | All identity changes suspended |
| `role_created` | `role_name`, `creator` | New role created |
| `role_assigned` | `address`, `role_name`, `expires_at` | Role assigned to address |
| `role_revoked` | `address`, `role_name` | Role revoked from address |
| `multisig_wallet_created` | `wallet_id`, `threshold`, `signers` | Multisig wallet created |
| `multisig_proposal_created` | `proposal_id`, `wallet_id`, `proposer` | Multisig proposal created |
| `multisig_proposal_signed` | `proposal_id`, `signer`, `signatures_count` | Proposal signed |
| `multisig_proposal_executed` | `proposal_id`, `executor` | Proposal executed |
| `time_locked_action_proposed` | `action_id`, `action_type`, `delay_seconds` | Time-locked action proposed |
| `time_locked_action_executed` | `action_id`, `executor` | Time-locked action executed |
| `time_locked_action_cancelled` | `action_id`, `canceller` | Time-locked action cancelled |
| `emergency_admin_activated` | `admin_address`, `privileges`, `expires_at` | Emergency admin activated |
| `emergency_admin_deactivated` | `admin_address` | Emergency admin deactivated |
| `validator_key_rotated` | `validator_address`, `new_consensus_pubkey` | Validator key rotated |
| `session_created` | `session_id`, `address`, `expires_at` | User session created |
| `session_ended` | `session_id`, `address` | User session ended |
| `identity_erased` | `did`, `requester`, `erased_at` | Identity data erased (GDPR) |
| `did_key_rotated` | `did`, `rotation_time`, `grace_period_end` | DID key rotated |

## Errors

| Code | Name | Description |
|------|------|-------------|
| 1 | `ErrInvalidDID` | DID format is invalid |
| 2 | `ErrIdentityNotFound` | Identity record not found |
| 3 | `ErrUnauthorized` | Caller lacks required permission |
| 4 | `ErrChangeRequestNotFound` | Change request ID not found |
| 5 | `ErrInsufficientVerifications` | Not enough verifications for change |
| 6 | `ErrIdentityChangesSuspended` | Identity changes globally suspended |
| 7 | `ErrRoleNotFound` | Role name not found |
| 8 | `ErrRoleAlreadyExists` | Role name already exists |
| 9 | `ErrRoleAssignmentNotFound` | Role assignment not found |
| 10 | `ErrInvalidPermission` | Permission name invalid |
| 11 | `ErrMultisigWalletNotFound` | Multisig wallet not found |
| 12 | `ErrMultisigProposalNotFound` | Multisig proposal not found |
| 13 | `ErrNotWalletSigner` | Address not in wallet signers list |
| 14 | `ErrProposalNotApproved` | Proposal threshold not met |
| 15 | `ErrProposalExpired` | Proposal expired |
| 16 | `ErrTimeLockedActionNotFound` | Time-locked action not found |
| 17 | `ErrActionDelayNotElapsed` | Time-lock delay not elapsed yet |
| 18 | `ErrActionNotPending` | Action not in pending status |
| 19 | `ErrEmergencyAdminNotFound` | Emergency admin not found |
| 20 | `ErrEmergencyAdminExpired` | Emergency admin expired |
| 21 | `ErrSessionNotFound` | Session ID not found |
| 22 | `ErrSessionExpired` | Session expired |
| 23 | `ErrInvalidThreshold` | Multisig threshold invalid |
| 24 | `ErrAlreadySigned` | Signer already signed proposal |
| 25 | `ErrIdentityAlreadyExists` | Identity DID already exists |

## CLI Examples

### Create a multisig wallet
```bash
aurad tx identity create-multisig-wallet \
  --signers aura1signer1,aura1signer2,aura1signer3 \
  --threshold 2 \
  --wallet-type individual \
  --from mykey
```

### Create a multisig proposal
```bash
aurad tx identity create-multisig-proposal \
  --wallet-id mswallet-... \
  --title "Treasury allocation" \
  --description "Q1 2025 allocation" \
  --payload $(cat proposal.json | base64) \
  --from signer1
```

### Sign a multisig proposal
```bash
aurad tx identity sign-multisig-proposal \
  --proposal-id msprop-... \
  --from signer2
```

### Propose a time-locked action
```bash
aurad tx identity propose-time-locked-action \
  --action-type parameter_change \
  --payload $(cat action.json | base64) \
  --delay-seconds 172800 \
  --from admin
```

### Create a session
```bash
aurad tx identity create-session \
  --device-fingerprint device_hash_123 \
  --ip-address 192.168.1.1 \
  --from mykey
```

### Rotate DID key
```bash
aurad tx identity rotate-did-key \
  --did did:aura:mainnet:abc123 \
  --new-verification-method new_pubkey_id \
  --reason "Scheduled rotation" \
  --from mykey
```

## Integration Notes

### For Wallet Developers

1. **Session Management**: Always create a session when user logs in, check expiration before operations
2. **Multisig Support**: Implement UI for proposal creation, signing, and execution workflows
3. **Time-Locked Actions**: Display countdown timers for pending actions
4. **Role Display**: Show user's roles and permissions in account view
5. **DID Resolution**: Use QueryIdentityRecord to resolve DIDs to addresses

### Security Considerations

- **Multisig Threshold**: Ensure threshold ≤ number of signers
- **Time-Lock Duration**: Use appropriate delays for sensitive operations (e.g., 48+ hours for parameter changes)
- **Emergency Admin**: Limit activation to true emergencies, set short expiration times
- **Session Timeout**: Default 1 hour, adjust based on risk profile
- **Permission Checks**: Always verify permissions before allowing UI actions

### Best Practices

- **Role Management**: Use least-privilege principle, create specific roles for specific tasks
- **Audit Trail**: All operations emit events for off-chain indexing and monitoring
- **Proposal Expiration**: Set reasonable expiration times for multisig proposals (default 7 days)
- **Key Rotation**: Implement grace periods for key rotation to allow dependent systems to update
- **GDPR Compliance**: Implement erasure workflows, store PII off-chain
