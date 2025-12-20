# Security Remediation Action Matrix
**Generated:** 2025-12-02
**Source:** SECURITY_ARCHITECTURE_AUDIT.md

---

## Critical Priority Fixes (DO FIRST)

| ID | Issue | Affected Modules | Effort | Impact | Status |
|----|-------|------------------|--------|--------|--------|
| CRITICAL-001 | Common security library not used | ALL | 3-5 days | EXTREME | 🔴 TODO |
| CRITICAL-002 | In-memory state (non-deterministic) | incidentresponse | 2 days | EXTREME | 🔴 TODO |
| CRITICAL-003 | No access control on emergency ops | security, incidentresponse | 1 day | EXTREME | 🔴 TODO |
| CRITICAL-004 | No reentrancy protection | ALL | 2-3 days | EXTREME | 🔴 TODO |
| CRITICAL-005 | Integer overflow in spending limits | walletsecurity | 1 day | HIGH | 🔴 TODO |
| CRITICAL-006 | No input validation | ALL | 2 days | HIGH | 🔴 TODO |
| CRITICAL-007 | No rate limiting on crypto ops | cryptography | 1 day | EXTREME | 🔴 TODO |
| CRITICAL-008 | Mock cryptography in privacy | privacy | 5-7 days | EXTREME | 🔴 TODO |
| CRITICAL-009 | No emergency pause propagation | ALL | 2 days | HIGH | 🔴 TODO |
| CRITICAL-010 | No slashing integration | validatorsecurity | 2 days | HIGH | 🔴 TODO |
| CRITICAL-011 | Rate limiters in memory only | networksecurity | 1 day | HIGH | 🔴 TODO |
| CRITICAL-012 | Missing msg_server | security | 2-3 days | HIGH | 🔴 TODO |

**Total Critical Effort:** 24-35 days (can be parallelized)
**Blocking Production:** YES - MUST fix before mainnet

---

## High Priority Fixes (DO NEXT)

| ID | Issue | Affected Modules | Effort | Impact | Status |
|----|-------|------------------|--------|--------|--------|
| HIGH-001 | No single source of truth | ALL | 3 days | MEDIUM | 🟡 TODO |
| HIGH-002 | Missing audit logging | ALL | 2 days | MEDIUM | 🟡 TODO |
| HIGH-003 | Sensitive keys stored on-chain | cryptography | 1 day | HIGH | 🟡 TODO |
| HIGH-004 | No compliance framework | privacy | 3 days | MEDIUM | 🟡 TODO |
| HIGH-005 | No runbook integration | incidentresponse | 2 days | MEDIUM | 🟡 TODO |
| HIGH-006 | No security metrics | ALL | 2 days | MEDIUM | 🟡 TODO |
| HIGH-007 | Missing DDoS protection | networksecurity | 2 days | MEDIUM | 🟡 TODO |
| HIGH-008 | No sentry enforcement | validatorsecurity | 1 day | MEDIUM | 🟡 TODO |

**Total High Effort:** 16 days
**Blocking Production:** Recommended before mainnet

---

## Medium Priority Improvements

| ID | Issue | Effort | Status |
|----|-------|--------|--------|
| MEDIUM-001 | Inconsistent error handling | 1 day | 🟢 TODO |
| MEDIUM-002 | No security testing framework | 3 days | 🟢 TODO |
| MEDIUM-003 | Missing security docs | 2 days | 🟢 TODO |
| MEDIUM-004 | No security upgrade path | 2 days | 🟢 TODO |
| MEDIUM-005 | Telemetry leaks sensitive data | 1 day | 🟢 TODO |
| MEDIUM-006 | No parameter governance | 2 days | 🟢 TODO |

**Total Medium Effort:** 11 days
**Blocking Production:** No, but recommended

---

## Detailed Remediation Steps

### CRITICAL-001: Integrate Common Security Library

**File:** All security module keepers

**Steps:**
1. Import common security package
   ```go
   import "github.com/aequitas/aura/chain/x/common/security"
   ```

2. Add guards to keeper struct
   ```go
   type Keeper struct {
       // Existing fields...
       reentrancyGuard *security.ReentrancyGuard
       pauseGuard      *security.PauseGuard
       inputValidator  *security.InputValidator
       safeMath        *security.SafeMath
       gasLimitGuard   *security.GasLimitGuard
       accessControl   *security.AccessControl
   }
   ```

3. Initialize in NewKeeper
   ```go
   func NewKeeper(...) *Keeper {
       return &Keeper{
           // ...
           reentrancyGuard: security.NewReentrancyGuard(),
           pauseGuard:      security.NewPauseGuard(authority),
           inputValidator:  security.NewInputValidator(),
           safeMath:        security.NewSafeMath(),
           gasLimitGuard:   security.NewGasLimitGuard(maxGas),
           accessControl:   security.NewAccessControl(admins),
       }
   }
   ```

4. Wrap all public functions
   ```go
   func (k Keeper) PublicFunction(ctx sdk.Context, input string) error {
       // Step 1: Check not paused
       if err := k.pauseGuard.CheckNotPaused(); err != nil {
           return err
       }

       // Step 2: Validate input
       if err := k.inputValidator.ValidateString(input, "input"); err != nil {
           return err
       }

       // Step 3: Reentrancy guard
       return k.reentrancyGuard.WithReentrancyGuard(func() error {
           // Actual logic here
           return nil
       })
   }
   ```

**Files to modify:**
- chain/x/security/keeper/keeper.go
- chain/x/walletsecurity/keeper/keeper.go
- chain/x/validatorsecurity/keeper/keeper.go
- chain/x/networksecurity/keeper/keeper.go
- chain/x/cryptography/keeper/keeper.go
- chain/x/incidentresponse/keeper/keeper.go
- chain/x/privacy/keeper/keeper.go

**Testing:**
```go
func TestReentrancyProtection(t *testing.T) {
    // Verify reentrancy blocked
}

func TestPauseEnforcement(t *testing.T) {
    // Verify pause blocks operations
}

func TestAccessControl(t *testing.T) {
    // Verify unauthorized calls rejected
}
```

---

### CRITICAL-002: Migrate Incident Response to KV Store

**File:** chain/x/incidentresponse/keeper/keeper.go

**Steps:**
1. Add KV store fields
   ```go
   type Keeper struct {
       cdc      codec.BinaryCodec
       storeKey storetypes.StoreKey
       // REMOVE in-memory maps:
       // incidents      map[string]*types.Incident ❌
       // nextIncidentID uint64 ❌
       // pauseState     *types.ChainPauseState ❌
       // pauseVotes     map[string][]string ❌
       // walletLimits   map[string]*types.WalletLimits ❌
   }
   ```

2. Implement KV store operations
   ```go
   func (k Keeper) getStore(ctx sdk.Context) storetypes.KVStore {
       return ctx.KVStore(k.storeKey)
   }

   func (k Keeper) SetIncident(ctx sdk.Context, incident *types.Incident) error {
       store := k.getStore(ctx)
       bz := k.cdc.MustMarshal(incident)
       store.Set(types.GetIncidentKey(incident.ID), bz)
       return nil
   }

   func (k Keeper) GetIncident(ctx sdk.Context, id string) (*types.Incident, error) {
       store := k.getStore(ctx)
       bz := store.Get(types.GetIncidentKey(id))
       if bz == nil {
           return nil, types.ErrIncidentNotFound
       }
       var incident types.Incident
       k.cdc.MustUnmarshal(bz, &incident)
       return &incident, nil
   }
   ```

3. Implement counter using KV store
   ```go
   func (k Keeper) GetNextIncidentID(ctx sdk.Context) uint64 {
       store := k.getStore(ctx)
       bz := store.Get(types.NextIncidentIDKey)
       if bz == nil {
           return 1
       }
       return sdk.BigEndianToUint64(bz)
   }

   func (k Keeper) IncrementIncidentID(ctx sdk.Context) {
       store := k.getStore(ctx)
       next := k.GetNextIncidentID(ctx) + 1
       store.Set(types.NextIncidentIDKey, sdk.Uint64ToBigEndian(next))
   }
   ```

4. Update all methods to use KV store
5. Implement genesis import/export
6. Add migration logic for any existing state

**Files to modify:**
- chain/x/incidentresponse/keeper/keeper.go
- chain/x/incidentresponse/keeper/keeper_kv.go (new)
- chain/x/incidentresponse/types/keys.go (add key prefixes)
- chain/x/incidentresponse/keeper/genesis.go

---

### CRITICAL-003: Implement Access Control on Emergency Ops

**File:** chain/x/security/keeper/msg_server.go (NEW)

**Steps:**
1. Create msg_server.go
   ```go
   package keeper

   import (
       "context"
       sdk "github.com/cosmos/cosmos-sdk/types"
       "github.com/aequitas/aura/chain/x/security/types"
   )

   type msgServer struct {
       Keeper
   }

   func NewMsgServerImpl(keeper Keeper) types.MsgServer {
       return &msgServer{Keeper: keeper}
   }

   var _ types.MsgServer = msgServer{}
   ```

2. Implement emergency pause handler
   ```go
   func (s msgServer) EmergencyPause(goCtx context.Context, msg *types.MsgEmergencyPause) (*types.MsgEmergencyPauseResponse, error) {
       ctx := sdk.UnwrapSDKContext(goCtx)

       // Access control check
       if !s.accessControl.HasRole(msg.Authority, "emergency_operator") {
           return nil, types.ErrUnauthorized
       }

       // Reentrancy guard
       if err := s.reentrancyGuard.Enter(); err != nil {
           return nil, err
       }
       defer s.reentrancyGuard.Exit()

       // Validate inputs
       if err := msg.ValidateBasic(); err != nil {
           return nil, err
       }

       // Execute pause
       if err := s.pauseGuard.Pause(ctx, msg.Authority); err != nil {
           return nil, err
       }

       // Emit event
       ctx.EventManager().EmitEvent(
           sdk.NewEvent(
               types.EventTypeEmergencyPause,
               sdk.NewAttribute(types.AttributeKeyAuthority, msg.Authority),
               sdk.NewAttribute(types.AttributeKeyReason, msg.Reason),
           ),
       )

       return &types.MsgEmergencyPauseResponse{}, nil
   }
   ```

3. Add multi-sig requirement
   ```go
   type PauseProposal struct {
       ID          string
       Proposer    string
       Approvals   []string
       Required    uint32
       ExpiresAt   time.Time
   }

   func (s msgServer) ProposeEmergencyPause(...) {
       // Create proposal
       // Store in KV
       // If threshold met, execute
   }

   func (s msgServer) ApproveEmergencyPause(...) {
       // Add approval
       // Check threshold
       // Execute if ready
   }
   ```

**Files to create:**
- chain/x/security/keeper/msg_server.go
- chain/x/security/types/msgs.go (define MsgEmergencyPause, etc.)
- chain/x/security/types/codec.go (register messages)

---

### CRITICAL-008: Replace Mock Cryptography in Privacy

**File:** chain/x/privacy/keeper/keeper.go

**Steps:**
1. **Option A: Disable privacy features until real implementation**
   ```go
   // In InitGenesis or params validation
   params.EnableZkProofs = false
   params.EnableConfidentialTransactions = false
   params.EnableRingSignatures = false

   // Add deprecation warnings
   func (k Keeper) VerifyZKProof(...) bool {
       panic("ZK proofs are not yet implemented in production")
   }
   ```

2. **Option B: Integrate real ZK proof library** (RECOMMENDED)
   ```go
   import (
       "github.com/consensys/gnark"
       "github.com/consensys/gnark/backend/groth16"
   )

   func (k Keeper) VerifyZKProof(ctx context.Context, proof *Proof) bool {
       // Load circuit
       circuit := k.GetCircuit(ctx, proof.CircuitID)

       // Load verifying key
       vk := k.GetVerifyingKey(ctx, proof.CircuitID)

       // Verify using gnark
       publicWitness := k.GetPublicWitness(ctx, proof.ProofID)
       err := groth16.Verify(proof.Data, vk, publicWitness)

       return err == nil
   }
   ```

3. Replace commitment verification
   ```go
   func (k Keeper) VerifyCommitment(ctx context.Context, commitment []byte, secret []byte) bool {
       // Compute commitment from secret
       computed := ComputePedersenCommitment(secret, randomness)

       // Constant-time comparison
       return subtle.ConstantTimeCompare(commitment, computed) == 1
   }
   ```

**Dependencies to add:**
```
require (
    github.com/consensys/gnark v0.9.0
    github.com/consensys/gnark-crypto v0.12.0
)
```

**Files to modify:**
- chain/x/privacy/keeper/keeper.go
- chain/x/privacy/keeper/zk_proofs.go (new)
- chain/x/privacy/keeper/commitments.go (new)
- go.mod

---

## Testing Requirements

### Security Test Suite Template
```go
// security_test_suite.go
type SecurityTestSuite struct {
    suite.Suite
    keeper       Keeper
    ctx          sdk.Context
    validAuthority string
    invalidCaller  string
}

func (s *SecurityTestSuite) TestReentrancyProtection() {
    // Attempt reentrancy attack
    s.Require().Error(s.keeper.ReentrantFunction(s.ctx))
}

func (s *SecurityTestSuite) TestUnauthorizedAccess() {
    // Attempt unauthorized operation
    err := s.keeper.EmergencyPause(s.ctx, s.invalidCaller)
    s.Require().ErrorIs(err, types.ErrUnauthorized)
}

func (s *SecurityTestSuite) TestInputValidation() {
    // Test with invalid inputs
    err := s.keeper.Operation(s.ctx, "")
    s.Require().Error(err)

    err = s.keeper.Operation(s.ctx, strings.Repeat("a", 10000))
    s.Require().Error(err)
}

func (s *SecurityTestSuite) TestPauseEnforcement() {
    // Pause module
    s.keeper.pauseGuard.Pause(s.ctx, s.validAuthority)

    // Verify operations blocked
    err := s.keeper.NormalOperation(s.ctx)
    s.Require().ErrorIs(err, security.ErrModulePaused)
}

func (s *SecurityTestSuite) TestIntegerOverflow() {
    // Test with max values
    maxInt := math.MaxInt64
    err := s.keeper.AddAmount(s.ctx, maxInt, 1)
    s.Require().ErrorIs(err, security.ErrIntegerOverflow)
}
```

---

## Progress Tracking

### Week 1: Critical Fixes
- [ ] CRITICAL-001: Common security library integration
- [ ] CRITICAL-002: Incident response KV migration
- [ ] CRITICAL-003: Access control implementation
- [ ] CRITICAL-004: Reentrancy guards

### Week 2: Critical Fixes (Cont.)
- [ ] CRITICAL-005: Safe math implementation
- [ ] CRITICAL-006: Input validation
- [ ] CRITICAL-007: Rate limiting on crypto
- [ ] CRITICAL-012: msg_server implementation

### Week 3: Critical Fixes (Cont.) + High Priority
- [ ] CRITICAL-008: Privacy cryptography (or disable)
- [ ] CRITICAL-009: Pause propagation
- [ ] CRITICAL-010: Slashing integration
- [ ] CRITICAL-011: Rate limiter persistence
- [ ] HIGH-001: State consolidation
- [ ] HIGH-002: Audit logging

### Week 4: High Priority
- [ ] HIGH-003: Key storage security
- [ ] HIGH-004: Compliance framework
- [ ] HIGH-005: Runbook integration
- [ ] HIGH-006: Security metrics
- [ ] HIGH-007: DDoS protection
- [ ] HIGH-008: Sentry enforcement

### Week 5: Medium Priority + Testing
- [ ] MEDIUM-001 through MEDIUM-006
- [ ] Comprehensive security test suite
- [ ] Penetration testing
- [ ] Security documentation

---

## Sign-off Checklist

Before considering remediation complete:

- [ ] All CRITICAL issues resolved
- [ ] Security test suite passing
- [ ] External security audit completed
- [ ] Security documentation complete
- [ ] Incident response procedures tested
- [ ] Emergency pause tested on testnet
- [ ] All modules using common security library
- [ ] No sensitive data stored on-chain
- [ ] Access control enforced on all privileged operations
- [ ] Reentrancy protection on all external-facing functions
- [ ] Input validation on all user inputs
- [ ] Integer overflow protection on all arithmetic
- [ ] Rate limiting on expensive operations
- [ ] Audit logging on all security events
- [ ] Real cryptography (no mocks) in privacy module

---

**Status Legend:**
- 🔴 TODO (Critical - blocking production)
- 🟡 TODO (High - strongly recommended)
- 🟢 TODO (Medium - recommended)
- ✅ DONE
- ⏸️ BLOCKED
