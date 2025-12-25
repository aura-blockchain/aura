# Auth Module Pagination Implementation

## Summary
Added pagination support to 6 query endpoints in the auth module following the proven pattern from the identity module.

## Modified Files

### Proto Definitions
- `proto/aura/auth/v1beta1/query.proto`
  - Added pagination import
  - Added PageRequest/PageResponse to 6 List methods

### Generated Code
- Ran `make proto-gen` to regenerate Go bindings

### Implementation
- `chain/x/auth/keeper/query_server.go`
  - Added pagination to all List query methods
  - Used query.Paginate() with prefix stores
  - Maintained existing filtering logic

## Paginated Endpoints

1. **ListRoles** - All roles with pagination
2. **ListMultisigWallets** - All multisig wallets with pagination
3. **ListMultisigProposals** - Proposals with wallet_id/status filters + pagination
4. **ListTimeLockedActions** - Actions with status filter + pagination
5. **ListEmergencyAdmins** - All emergency admins with pagination
6. **ListSessions** - Sessions filtered by user_address + pagination

## Technical Details

All implementations:
- Use `prefix.NewStore()` with appropriate key prefixes from keeper.go
- Use `query.Paginate()` for efficient iteration
- Apply filters within pagination callback
- Return `PageResponse` with pagination metadata

## Verification

```bash
cd chain && go build ./x/auth/...  # Success
```

## Pattern Source
Identity module: `chain/x/identity/keeper/query_server.go` (14 paginated methods)
