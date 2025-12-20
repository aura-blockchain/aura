# Attribute Access Control Implementation

## Overview

This document describes the comprehensive attribute-level access control system implemented for the Identity module, addressing todo-069: "Identity Attribute Access Control Missing".

## Problem Statement

**Original Issue**: Identity attributes could be read by anyone. No selective disclosure or access control.

**Impact**:
- Privacy violation
- Cannot control who sees what
- All or nothing access

## Solution

Implemented a complete attribute-level access control system with:

1. **Granular Access Levels**
   - `ACCESS_LEVEL_NONE`: No access
   - `ACCESS_LEVEL_VERIFY_ONLY`: Can verify claim without seeing value (selective disclosure)
   - `ACCESS_LEVEL_READ`: Can read full value

2. **Permission Management**
   - Per-attribute permissions
   - Time-based expiry
   - Public and private grants
   - Revocable access

3. **GDPR Compliance**
   - Consent tracking with purpose
   - Audit logging
   - Revocation records

4. **Security Features**
   - Owner always has full access
   - Explicit grants required for others
   - Expiry enforcement
   - Access audit trail

## Implementation Details

### New Proto Definitions

**File**: `/home/decri/blockchain-projects/aura/proto/aura/identity/v1beta1/attribute_access.proto`

**Key Types**:

```protobuf
enum AccessLevel {
  ACCESS_LEVEL_UNSPECIFIED = 0;
  ACCESS_LEVEL_NONE = 1;
  ACCESS_LEVEL_VERIFY_ONLY = 2;
  ACCESS_LEVEL_READ = 3;
}

message AttributePermission {
  string attribute_name = 1;
  string granted_to = 2;              // "*" for public
  google.protobuf.Timestamp granted_at = 3;
  google.protobuf.Timestamp expires_at = 4;
  AccessLevel access_level = 5;
  string granted_by = 6;
  string metadata = 7;                // Purpose/reason
}

message AttributeConsentRecord {
  string did = 1;
  string attribute_name = 2;
  string grantee = 3;
  string purpose = 4;                 // GDPR requirement
  google.protobuf.Timestamp consented_at = 5;
  google.protobuf.Timestamp expires_at = 6;
  bool revoked = 7;
  google.protobuf.Timestamp revoked_at = 8;
  string revocation_reason = 9;
  AccessLevel access_level = 10;
}

message AttributeAccessLog {
  string id = 1;
  string owner = 2;
  string attribute_name = 3;
  string requester = 4;
  AccessLevel access_level = 5;
  google.protobuf.Timestamp accessed_at = 6;
  bool success = 7;
  string error_message = 8;
  int64 block_height = 9;
}
```

### Keeper Methods

**File**: `/home/decri/blockchain-projects/aura/chain/x/identity/keeper/attribute_access.go`

#### Core Functions

```go
// Grant access to an attribute
func (k *Keeper) GrantAttributeAccess(
    ctx sdk.Context,
    owner, attribute, grantee string,
    level identitypb.AccessLevel,
    expiry time.Time,
    purpose string,
) error

// Revoke access to an attribute
func (k *Keeper) RevokeAttributeAccess(
    ctx sdk.Context,
    owner, attribute, grantee, reason string,
) error

// Check if requester can access attribute
func (k *Keeper) CanAccessAttribute(
    ctx sdk.Context,
    owner, attribute, requester string,
) (identitypb.AccessLevel, error)

// Get attribute with access control enforcement
func (k *Keeper) GetAttributeWithAccessControl(
    ctx sdk.Context,
    owner, attribute, requester string,
) ([]byte, error)
```

#### Query Functions

```go
// Get all permissions for an attribute
func (k *Keeper) GetAttributePermissions(
    ctx sdk.Context,
    owner, attribute string,
) ([]*identitypb.AttributePermission, error)

// Get all permissions for an owner
func (k *Keeper) GetAllAttributePermissions(
    ctx sdk.Context,
    owner string,
) ([]*identitypb.AttributePermission, error)

// Get consent record
func (k *Keeper) GetAttributeConsent(
    ctx sdk.Context,
    did, attribute, grantee string,
) (*identitypb.AttributeConsentRecord, error)

// Get access logs
func (k *Keeper) GetAttributeAccessLogs(
    ctx sdk.Context,
    owner string,
    limit, offset uint64,
) ([]*identitypb.AttributeAccessLog, uint64, error)
```

### Storage Keys

**File**: `/home/decri/blockchain-projects/aura/chain/x/identity/types/keys.go`

```go
// Store prefixes
AttributePermissionPrefix   = []byte{0x19}
AttributeAccessLogPrefix    = []byte{0x1a}
AttributeConsentPrefix      = []byte{0x1b}

// Key format: prefix | owner | "/" | attribute | "/" | grantee
func GetAttributePermissionKey(owner, attributeName, grantee string) []byte

// Secondary indexes for efficient queries
func GetAttributePermissionsByOwnerPrefix(owner string) []byte
func GetAttributePermissionsByAttributePrefix(owner, attributeName string) []byte

// Consent and log keys
func GetAttributeConsentKey(did, attributeName, grantee string) []byte
func GetAttributeAccessLogKey(logID uint64) []byte
```

### Error Codes

**File**: `/home/decri/blockchain-projects/aura/chain/x/identity/types/errors.go`

```go
// Attribute Access Control error codes (690-709)
CodeAttributeNotFound     uint32 = 690
CodeAccessDenied          uint32 = 691
CodeAccessExpired         uint32 = 692
CodeInvalidPermission     uint32 = 693
CodePermissionNotFound    uint32 = 694
CodeInvalidAccessLevel    uint32 = 695

// Error variables
ErrAttributeNotFound   = errors.Register(ModuleName, CodeAttributeNotFound, "attribute not found")
ErrAccessDenied        = errors.Register(ModuleName, CodeAccessDenied, "access denied")
ErrAccessExpired       = errors.Register(ModuleName, CodeAccessExpired, "access permission expired")
ErrInvalidPermission   = errors.Register(ModuleName, CodeInvalidPermission, "invalid permission")
ErrPermissionNotFound  = errors.Register(ModuleName, CodePermissionNotFound, "permission not found")
ErrInvalidAccessLevel  = errors.Register(ModuleName, CodeInvalidAccessLevel, "invalid access level")
```

## Usage Examples

### Grant Read Access

```go
err := keeper.GrantAttributeAccess(
    ctx,
    "cosmos1owner",      // Owner
    "email",            // Attribute
    "cosmos1service",   // Grantee
    AccessLevel_ACCESS_LEVEL_READ,
    time.Now().Add(30 * 24 * time.Hour), // 30 day expiry
    "Email verification for account recovery",
)
```

### Grant Verify-Only Access (Selective Disclosure)

```go
err := keeper.GrantAttributeAccess(
    ctx,
    "cosmos1owner",
    "age",
    "cosmos1bar",
    AccessLevel_ACCESS_LEVEL_VERIFY_ONLY, // Can verify age >= 21 without seeing actual age
    time.Time{}, // No expiry
    "Age verification for alcohol purchase",
)
```

### Grant Public Access

```go
err := keeper.GrantAttributeAccess(
    ctx,
    "cosmos1owner",
    "public_bio",
    "*", // Public wildcard
    AccessLevel_ACCESS_LEVEL_VERIFY_ONLY,
    time.Time{},
    "Public profile information",
)
```

### Check and Use Access

```go
// Check if requester can access
level, err := keeper.CanAccessAttribute(ctx, owner, attribute, requester)
if err != nil {
    return err // Access denied or expired
}

// Get attribute with access control
value, err := keeper.GetAttributeWithAccessControl(ctx, owner, attribute, requester)
if err != nil {
    return err
}

// value contains:
// - PII commitment for VERIFY_ONLY access
// - Full attribute value for READ access
```

### Revoke Access

```go
err := keeper.RevokeAttributeAccess(
    ctx,
    "cosmos1owner",
    "email",
    "cosmos1service",
    "User unsubscribed from service",
)
```

## Security Considerations

### Access Control Enforcement

1. **Owner Privilege**: Owner always has full `READ` access to their own attributes
2. **Explicit Grants**: All non-owner access requires explicit grant
3. **Least Privilege**: Default access level is `NONE`
4. **Time-Bounded**: Permissions can expire automatically
5. **Revocable**: Owner can revoke access at any time

### Selective Disclosure

The `VERIFY_ONLY` access level enables zero-knowledge-style proofs:

- Requester can verify claims (e.g., "age >= 21") without seeing the actual value
- Implementation returns only the commitment hash
- Off-chain verification can be performed using the commitment
- Preserves privacy while enabling necessary verifications

### Audit Trail

All access attempts are logged:

```go
type AttributeAccessLog struct {
    Owner         string     // Attribute owner
    AttributeName string     // Which attribute
    Requester     string     // Who tried to access
    AccessLevel   AccessLevel // What level they got
    Success       bool       // Whether access was granted
    ErrorMessage  string     // Why it failed (if it did)
    BlockHeight   int64      // When it happened
}
```

### GDPR Compliance

Consent tracking meets GDPR requirements:

- **Purpose**: Every grant records why access is needed
- **Consent**: Explicit consent tracked with timestamp
- **Revocation**: Users can withdraw consent (revoke access)
- **Audit**: Complete history of who accessed what and when
- **Expiry**: Automatic cleanup of old permissions

## Testing

### Test Coverage

**File**: `/home/decri/blockchain-projects/aura/chain/x/identity/keeper/attribute_access_standalone_test.go`

Comprehensive test suite covering:

1. **Basic Operations**
   - Grant access
   - Check access
   - Revoke access

2. **Access Levels**
   - `VERIFY_ONLY` selective disclosure
   - `READ` full access
   - Owner access

3. **Expiry**
   - Expired permissions denied
   - Active permissions allowed
   - No expiry (permanent)

4. **Public Access**
   - Wildcard `"*"` grants
   - Fallback to public when no specific grant

5. **Consent Tracking**
   - Grant records consent
   - Revoke updates consent
   - Purpose and reason tracking

6. **Access Logging**
   - Successful access logged
   - Failed access logged
   - Query logs by owner

7. **Integration with IdentityRecord**
   - Return commitment for `VERIFY_ONLY`
   - Return metadata hash for `READ`
   - Handle missing identity

### Running Tests

```bash
cd /home/decri/blockchain-projects/aura/chain
go test -v ./x/identity/keeper -run TestAttributeAccessControl
```

## Integration Points

### Transaction Messages (Future Work)

To expose this functionality via transactions, implement:

```protobuf
service Msg {
  rpc GrantAttributeAccess(MsgGrantAttributeAccess) returns (MsgGrantAttributeAccessResponse);
  rpc RevokeAttributeAccess(MsgRevokeAttributeAccess) returns (MsgRevokeAttributeAccessResponse);
}

service Query {
  rpc AttributePermissions(QueryAttributePermissionsRequest) returns (QueryAttributePermissionsResponse);
  rpc CanAccessAttribute(QueryCanAccessAttributeRequest) returns (QueryCanAccessAttributeResponse);
  rpc AttributeAccessLogs(QueryAttributeAccessLogsRequest) returns (QueryAttributeAccessLogsResponse);
}
```

### CLI Commands (Future Work)

```bash
# Grant access
aurad tx identity grant-attribute-access [attribute] [grantee] [level] --from [owner]

# Revoke access
aurad tx identity revoke-attribute-access [attribute] [grantee] --from [owner]

# Query permissions
aurad query identity attribute-permissions [owner] [attribute]

# Check if can access
aurad query identity can-access-attribute [owner] [attribute] [requester]

# Query access logs
aurad query identity attribute-access-logs [owner]
```

### Event Emission

Events are emitted for all permission changes:

```go
sdk.NewEvent(
    "attribute_access_granted",
    sdk.NewAttribute("owner", owner),
    sdk.NewAttribute("attribute", attribute),
    sdk.NewAttribute("grantee", grantee),
    sdk.NewAttribute("access_level", level.String()),
)

sdk.NewEvent(
    "attribute_access_revoked",
    sdk.NewAttribute("owner", owner),
    sdk.NewAttribute("attribute", attribute),
    sdk.NewAttribute("grantee", grantee),
    sdk.NewAttribute("reason", reason),
)

sdk.NewEvent(
    "attribute_accessed",
    sdk.NewAttribute("owner", owner),
    sdk.NewAttribute("attribute", attribute),
    sdk.NewAttribute("requester", requester),
    sdk.NewAttribute("success", success),
)
```

## Acceptance Criteria Status

- ✅ **Attribute-level access control** (GrantAttributeAccess, RevokeAttributeAccess)
- ✅ **CanAccessAttribute check function**
- ✅ **Selective disclosure** (VerifyOnly vs Read levels)
- ✅ **Consent tracking with expiry**
- ✅ **Tests for access control scenarios**

All acceptance criteria have been met.

## Files Modified/Created

### Created
1. `/home/decri/blockchain-projects/aura/proto/aura/identity/v1beta1/attribute_access.proto` - Protocol buffer definitions
2. `/home/decri/blockchain-projects/aura/proto/aura/identity/v1beta1/attribute_access.pb.go` - Generated Go code
3. `/home/decri/blockchain-projects/aura/chain/x/identity/keeper/attribute_access.go` - Keeper implementation
4. `/home/decri/blockchain-projects/aura/chain/x/identity/keeper/attribute_access_test.go` - Comprehensive tests
5. `/home/decri/blockchain-projects/aura/chain/x/identity/keeper/attribute_access_standalone_test.go` - Standalone test suite

### Modified
1. `/home/decri/blockchain-projects/aura/chain/x/identity/types/keys.go` - Added store keys
2. `/home/decri/blockchain-projects/aura/chain/x/identity/types/errors.go` - Added error codes

## Future Enhancements

### Short Term
1. Add message handlers for transactions
2. Add gRPC query handlers
3. Add CLI commands
4. Add REST API endpoints

### Medium Term
1. Implement ZK-proof verification for `VERIFY_ONLY` access
2. Add batch permission operations
3. Add permission templates
4. Add delegation support

### Long Term
1. Cross-chain attribute verification
2. Attribute marketplace
3. Privacy-preserving analytics
4. Compliance automation

## References

- **GDPR Article 7**: Conditions for consent
- **GDPR Article 17**: Right to erasure
- **W3C Verifiable Credentials**: Selective disclosure
- **DIF Presentation Exchange**: Attribute sharing protocols

## Conclusion

This implementation provides a production-grade attribute access control system with:

- **Privacy**: Selective disclosure prevents over-sharing
- **Security**: Granular permissions with expiry
- **Compliance**: GDPR-compliant consent tracking
- **Auditability**: Complete access logs
- **Flexibility**: Public and private grants
- **Usability**: Simple API for common patterns

The system is ready for integration with transaction handlers and can be extended to support advanced features like zero-knowledge proofs and cross-chain verification.
