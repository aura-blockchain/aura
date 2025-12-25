# Auth Module

## Overview

The Auth module provides advanced authentication and authorization capabilities for the Aura blockchain, including role-based access control (RBAC), multisig wallets, time-locked admin actions, emergency admin privileges, validator key rotation, session management, and rate limiting.

## Features

- **Role-Based Access Control**: Create roles with granular permissions and assign to addresses with expiration
- **Multisig Wallets**: Multi-signature wallets with configurable thresholds (3-of-5, 5-of-7, custom)
- **Time-Locked Actions**: Delayed execution of admin actions with mandatory wait periods
- **Emergency Admin**: Temporary elevated privileges for emergency response with auto-expiration
- **Validator Key Rotation**: Secure rotation of validator consensus public keys
- **Session Management**: API session tracking with timeout and device authentication
- **Rate Limiting**: Per-user configurable rate limits (per-minute, per-hour, per-day)
- **Audit Logging**: Comprehensive event tracking for all auth operations

## State

### Roles and Permissions
- **Role**: Named permission sets (admin, moderator, validator, user)
- **RoleAssignment**: Role-to-address mappings with optional expiration timestamps
- **Permissions**: Granular capabilities (create_role, assign_role, manage_multisig, manage_timelock, etc.)

### Multisig Wallets
- **MultisigWallet**: Wallet with signers array and threshold requirement
- **MultisigProposal**: Proposals requiring threshold signatures before execution
- **WalletType**: Enum - 3_OF_5, 5_OF_7, CUSTOM

### Time-Locked Actions
- **TimeLockedAction**: Actions with mandatory delay period before execution
- **ActionStatus**: Pending, executed, cancelled

### Emergency Admin
- **EmergencyAdmin**: Temporary admin with specific privileges and auto-expiration

### Sessions
- **Session**: User authentication session with timeout and device tracking
- **RateLimitConfig**: Per-user rate limit configuration

### Validator Key Rotation
- **ValidatorKeyRotation**: Consensus key rotation tracking with status

## Messages

### MsgCreateRole
Create a new role with permissions.

**Fields**: `creator`, `name`, `permissions`, `description`

### MsgAssignRole
Assign role to address with optional expiration.

**Fields**: `assigner`, `address`, `role_name`, `expires_in_seconds`

### MsgRevokeRole
Revoke role from address.

**Fields**: `revoker`, `address`, `role_name`

### MsgCreateMultisigWallet
Create multisig wallet with signers and threshold.

**Fields**: `creator`, `signers`, `threshold`, `wallet_type`

### MsgCreateMultisigProposal
Create proposal for multisig wallet action.

**Fields**: `proposer`, `wallet_id`, `title`, `description`, `payload`, `expires_in_seconds`

### MsgSignMultisigProposal
Sign a multisig proposal.

**Fields**: `signer`, `proposal_id`

### MsgExecuteMultisigProposal
Execute approved multisig proposal.

**Fields**: `executor`, `proposal_id`

### MsgProposeTimeLockedAction
Propose time-locked admin action.

**Fields**: `proposer`, `action_type`, `payload`, `delay_seconds`

### MsgExecuteTimeLockedAction
Execute ready time-locked action.

**Fields**: `executor`, `action_id`

### MsgCancelTimeLockedAction
Cancel pending time-locked action.

**Fields**: `canceller`, `action_id`

### MsgActivateEmergencyAdmin
Activate emergency admin with privileges.

**Fields**: `activator`, `admin_address`, `privileges`, `expires_in_seconds`

### MsgDeactivateEmergencyAdmin
Deactivate emergency admin.

**Fields**: `deactivator`, `admin_address`

### MsgInitiateValidatorKeyRotation
Initiate validator key rotation.

**Fields**: `initiator`, `validator_address`, `new_consensus_pubkey`

### MsgCompleteValidatorKeyRotation
Complete validator key rotation.

**Fields**: `completer`, `validator_address`

### MsgCreateSession
Create API session with timeout.

**Fields**: `user_address`

### MsgRevokeSession
Revoke active session.

**Fields**: `user_address`, `session_id`
