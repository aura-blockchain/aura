# Auth Module CLI Commands

This document provides comprehensive examples for all CLI commands in the x/auth module.

## Transaction Commands

### Role Management

#### Create Role
Create a new role with specified permissions.

```bash
aurad tx auth create-role admin "CREATE_ROLE,ASSIGN_ROLE,MANAGE_PARAMS" "Administrator role" --from mykey
aurad tx auth create-role operator "EXECUTE_TX,VIEW_LOGS" "Operator role" --from mykey
aurad tx auth create-role viewer "VIEW_LOGS,VIEW_STATUS" "Read-only viewer role" --from mykey
```

#### Assign Role
Assign a role to an address with optional expiration.

```bash
# Permanent assignment
aurad tx auth assign-role aura1abc... admin --from mykey

# Temporary assignment (expires in 24 hours)
aurad tx auth assign-role aura1abc... operator --expires-in 86400 --from mykey
```

#### Revoke Role
Remove a role assignment from an address.

```bash
aurad tx auth revoke-role aura1abc... admin --from mykey
```

### Multisig Wallet Operations

#### Create Multisig Wallet
Create a multisig wallet with specified signers and threshold.

```bash
# Create a 2-of-3 custom multisig wallet
aurad tx auth create-multisig-wallet "aura1abc...,aura1def...,aura1ghi..." 2 --wallet-type custom --from mykey

# Create a 3-of-5 standard multisig wallet
aurad tx auth create-multisig-wallet "aura1abc...,aura1def...,aura1ghi...,aura1jkl...,aura1mno..." 3 --wallet-type 3-of-5 --from mykey
```

#### Create Multisig Proposal
Create a proposal that requires multiple signatures.

```bash
# Proposal expires in 24 hours
aurad tx auth create-multisig-proposal wallet123 "Transfer funds" "Transfer 100 tokens to treasury" abcd1234 --expires-in 86400 --from mykey

# Proposal expires in 7 days
aurad tx auth create-multisig-proposal wallet456 "Update parameters" "Change fee params" ef567890 --expires-in 604800 --from mykey
```

#### Sign Multisig Proposal
Sign a pending multisig proposal.

```bash
aurad tx auth sign-multisig-proposal proposal123 --from signer1
aurad tx auth sign-multisig-proposal proposal123 --from signer2
```

#### Execute Multisig Proposal
Execute an approved multisig proposal.

```bash
aurad tx auth execute-multisig-proposal proposal123 --from mykey
```

### Time-Locked Actions

#### Propose Time-Locked Action
Propose an admin action with a time delay.

```bash
# Propose parameter update with 24-hour delay
aurad tx auth propose-timelocked-action UPDATE_PARAMS abcd1234 86400 --from admin

# Propose admin change with 7-day delay
aurad tx auth propose-timelocked-action CHANGE_ADMIN ef567890 604800 --from admin
```

#### Execute Time-Locked Action
Execute a time-locked action after the delay period.

```bash
aurad tx auth execute-timelocked-action action123 --from admin
```

#### Cancel Time-Locked Action
Cancel a pending time-locked action.

```bash
aurad tx auth cancel-timelocked-action action123 --from admin
```

### Emergency Admin Management

#### Activate Emergency Admin
Activate an emergency admin with specific privileges.

```bash
# Activate emergency admin with system pause privilege (expires in 1 hour)
aurad tx auth activate-emergency-admin aura1abc... "PAUSE_SYSTEM,EMERGENCY_WITHDRAWAL" --expires-in 3600 --from admin

# Activate emergency admin with no expiration
aurad tx auth activate-emergency-admin aura1abc... "EMERGENCY_WITHDRAWAL,OVERRIDE_LIMITS" --from admin
```

#### Deactivate Emergency Admin
Deactivate an active emergency admin.

```bash
aurad tx auth deactivate-emergency-admin aura1abc... --from admin
```

### Validator Key Rotation

#### Initiate Key Rotation
Initiate the rotation of a validator's consensus public key.

```bash
aurad tx auth initiate-key-rotation auravaloper1abc... '{"@type":"/cosmos.crypto.ed25519.PubKey","key":"base64key..."}' --from validator
```

#### Complete Key Rotation
Complete the validator key rotation process.

```bash
aurad tx auth complete-key-rotation auravaloper1abc... --from validator
```

### Session Management

#### Create Session
Create a new API session.

```bash
# Create session with metadata
aurad tx auth create-session 192.168.1.1 --metadata "device=mobile,app=wallet,version=1.0" --from mykey

# Create session without metadata
aurad tx auth create-session 10.0.0.1 --from mykey
```

#### Revoke Session
Revoke an active session.

```bash
aurad tx auth revoke-session session123 --from mykey
```

## Query Commands

### Role Queries

#### Query Role
Get details of a specific role.

```bash
aurad query auth role admin
aurad query auth role operator
```

#### List All Roles
List all defined roles.

```bash
aurad query auth roles
```

#### Query Role Assignments
Get role assignments for an address.

```bash
aurad query auth role-assignments aura1abc...
```

#### Check Permission
Check if an address has a specific permission.

```bash
aurad query auth has-permission aura1abc... CREATE_ROLE
aurad query auth has-permission aura1abc... EXECUTE_TX
```

### Multisig Wallet Queries

#### Query Multisig Wallet
Get details of a multisig wallet.

```bash
aurad query auth multisig-wallet wallet123
```

#### List Multisig Wallets
List all multisig wallets.

```bash
aurad query auth multisig-wallets
```

#### Query Multisig Proposal
Get details of a multisig proposal.

```bash
aurad query auth multisig-proposal proposal123
```

#### List Multisig Proposals
List proposals for a wallet with optional status filter.

```bash
# List all proposals for a wallet
aurad query auth multisig-proposals wallet123

# List only pending proposals
aurad query auth multisig-proposals wallet123 --status pending

# List only executed proposals
aurad query auth multisig-proposals wallet123 --status executed
```

### Time-Locked Action Queries

#### Query Time-Locked Action
Get details of a time-locked action.

```bash
aurad query auth timelocked-action action123
```

#### List Time-Locked Actions
List all time-locked actions with optional status filter.

```bash
# List all actions
aurad query auth timelocked-actions

# List only pending actions
aurad query auth timelocked-actions --status pending

# List only ready actions
aurad query auth timelocked-actions --status ready
```

### Emergency Admin Queries

#### Query Emergency Admin
Get emergency admin status for an address.

```bash
aurad query auth emergency-admin aura1abc...
```

#### List Emergency Admins
List all emergency admins.

```bash
aurad query auth emergency-admins
```

### Validator Key Rotation Queries

#### Query Validator Key Rotation
Get key rotation status for a validator.

```bash
aurad query auth validator-key-rotation auravaloper1abc...
```

### Session Queries

#### Query Session
Get details of a session.

```bash
aurad query auth session session123
```

#### List Sessions
List all sessions for a user.

```bash
aurad query auth sessions aura1abc...
```

### Rate Limit Queries

#### Query Rate Limit Status
Get current rate limit status for a user.

```bash
aurad query auth rate-limit-status aura1abc...
```

### Audit Log Queries

#### Query Audit Logs
Query audit logs with various filters.

```bash
# Get all recent logs (default limit 100)
aurad query auth audit-logs

# Filter by actor
aurad query auth audit-logs --actor aura1abc...

# Filter by action type
aurad query auth audit-logs --action CREATE_ROLE

# Filter by time range (Unix timestamps)
aurad query auth audit-logs --start-time 1609459200 --end-time 1609545600

# Complex filter with all parameters
aurad query auth audit-logs --actor aura1abc... --action CREATE_ROLE --start-time 1609459200 --end-time 1609545600 --limit 50
```

### Parameters Query

#### Query Module Parameters
Get current auth module parameters.

```bash
aurad query auth params
```

## Common Usage Patterns

### Setting up a Multisig Admin Workflow

1. Create roles:
```bash
aurad tx auth create-role multisig_admin "CREATE_MULTISIG_WALLET,CREATE_PROPOSAL" "Multisig admin role" --from admin
```

2. Create multisig wallet:
```bash
aurad tx auth create-multisig-wallet "admin1,admin2,admin3" 2 --wallet-type custom --from admin1
```

3. Create and execute proposal:
```bash
# Create proposal
aurad tx auth create-multisig-proposal wallet1 "Treasury Transfer" "Send 1000 tokens" <payload-hex> --expires-in 86400 --from admin1

# Sign proposal (2 of 3 needed)
aurad tx auth sign-multisig-proposal proposal1 --from admin1
aurad tx auth sign-multisig-proposal proposal1 --from admin2

# Execute proposal
aurad tx auth execute-multisig-proposal proposal1 --from admin1
```

### Emergency Pause and Recovery

1. Activate emergency admin:
```bash
aurad tx auth activate-emergency-admin emergency_key "PAUSE_SYSTEM,EMERGENCY_ACTIONS" --expires-in 7200 --from governance
```

2. Execute emergency actions as emergency admin

3. Deactivate when crisis resolved:
```bash
aurad tx auth deactivate-emergency-admin emergency_key --from governance
```

### Rotating Validator Keys

1. Initiate rotation:
```bash
aurad tx auth initiate-key-rotation valoper1abc... '{"@type":"/cosmos.crypto.ed25519.PubKey","key":"new_key"}' --from validator
```

2. Check rotation status:
```bash
aurad query auth validator-key-rotation valoper1abc...
```

3. Complete rotation:
```bash
aurad tx auth complete-key-rotation valoper1abc... --from validator
```

### Monitoring and Auditing

1. Check rate limits:
```bash
aurad query auth rate-limit-status user1
```

2. Review audit logs:
```bash
aurad query auth audit-logs --action CREATE_ROLE --limit 100
aurad query auth audit-logs --actor admin1 --start-time 1609459200
```

3. List active sessions:
```bash
aurad query auth sessions user1
```

## Error Handling

Common errors and solutions:

- **"permission denied"**: User doesn't have required role/permission
- **"role not found"**: Role must be created before assignment
- **"proposal expired"**: Multisig proposal deadline has passed
- **"insufficient signatures"**: Multisig threshold not met
- **"action not ready"**: Time-lock delay not yet elapsed
- **"rate limit exceeded"**: User has hit API rate limits
