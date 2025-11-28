# Confidence Score Aggregation Module Design

**Status:** Design Specification
**Created:** 2025-11-13
**Target Release:** Devnet/Testnet
**Dependencies:** inclusion_routines, identitychange, tokenomics, vc_registry

---

## 1. MODULE PURPOSE

### 1.1 What Confidence Scores Represent

The **ConfidenceScore (CS)** is an aggregate numerical metric representing the statistical reliability of a user's verified identity claims. It quantifies the cumulative assurance gained from completing multiple Inclusion Routines (IRs), each contributing evidence that the wallet is controlled by a unique, real human.

**Key Properties:**
- **Additive Accumulation:** Each completed IR adds points to the total CS
- **Binary Threshold:** CS ≥ 10,000 = "Verified" status
- **Prerequisite-Gated:** All users must complete IR-000 (anchor) before earning CS points
- **Asynchronous:** Users build CS over time at their own pace
- **Immutable History:** Completed IRs cannot be removed (only new ones added)
- **Fraud-Resistant:** Slashing mechanisms reduce CS for fraudulent verifications

### 1.2 How Scores Aggregate from IR Completions

**Aggregation Formula:**
```
TotalCS = Σ(IRᵢ.score_value × arena_multiplier × velocity_bonus)
```

Where:
- `IRᵢ.score_value`: Base score defined in IR registry (50-2500 points)
- `arena_multiplier`: Optional bonus for arena focus (1.0-1.5×)
- `velocity_bonus`: Time-based completion bonus (1.0-1.25×)

**Completion Flow:**
1. User completes IR via AI Assistant
2. Assistant submits attestation to blockchain
3. Confidence Score module validates:
   - IR exists and is active
   - User has completed anchor (IR-000)
   - No rate limit violations
   - Prerequisites satisfied
   - Proof hash signature valid
4. Score added to user's total CS
5. Event emitted for wallet UI update

### 1.3 Threshold Requirements for Different VCs

The ConfidenceScore directly gates access to Verifiable Credentials:

| VC Type | CS Threshold | Additional Requirements |
|---------|--------------|-------------------------|
| `VC:isVerifiedHuman` | 10,000 | IR-000 completion |
| `VC:isAgeOver18` | 10,000 | IR-000 + IR-GOV_ID_01 |
| `VC:isAgeOver21` | 10,000 | IR-000 + IR-GOV_ID_01 (US-specific) |
| `VC:hasHighAssurance` | 15,000 | IR-000 + ≥2 High-Assurance arena IRs |
| `VC:isProfessional` | 12,000 | IR-000 + IR-206 or IR-604 |
| `VC:ArenaFocus:Biometric` | 10,000 | IR-000 + ≥5,000 CS from Biometric arena |
| `VC:ArenaFocus:GeoLocation` | 10,000 | IR-000 + ≥5,000 CS from GeoLocation arena |
| `VC:ArenaFocus:HighAssurance` | 10,000 | IR-000 + ≥5,000 CS from HighAssurance arena |

**Verification Policy Example:**
```json
{
  "required_credentials": [
    {
      "type": "VerifiedHuman",
      "constraints": {
        "minimum_cs": 10000,
        "required_anchor": "IR-000",
        "accepted_verifiers": ["sha256:..."]
      }
    }
  ]
}
```

---

## 2. STATE DESIGN

### 2.1 Core State Objects

**UserConfidenceRecord:**
```protobuf
message UserConfidenceRecord {
  string wallet_address = 1;              // Bech32 Aura address
  uint64 total_score = 2;                 // Aggregate CS value
  repeated IRCompletion completed_irs = 3; // Full completion history
  AnchorInfo anchor_info = 4;             // IR-000 metadata
  google.protobuf.Timestamp last_updated = 5;
  VerificationStatus status = 6;          // UNVERIFIED, VERIFIED, SUSPENDED
  uint64 verification_achieved_height = 7; // Block when CS ≥ 10,000
  ArenaBreakdown arena_scores = 8;        // Per-arena CS totals
}

message AnchorInfo {
  bool completed = 1;
  google.protobuf.Timestamp completed_at = 2;
  bytes verifier_plugin_hash = 3;        // SHA256 of verifier plug-in
  uint64 block_height = 4;
  bytes proof_hash = 5;
}

enum VerificationStatus {
  UNVERIFIED = 0;
  VERIFIED = 1;      // CS ≥ 10,000
  SUSPENDED = 2;     // Governance flag
  REVOKED = 3;       // Self-revocation or fraud
}

message ArenaBreakdown {
  uint64 biometric = 1;
  uint64 possession = 2;
  uint64 knowledge = 3;
  uint64 social = 4;
  uint64 geo_location = 5;
  uint64 high_assurance = 6;
  uint64 persistence = 7;
  uint64 specialized = 8;
}
```

**IRCompletion:**
```protobuf
message IRCompletion {
  string ir_id = 1;                      // e.g., "IR-102"
  uint64 score_earned = 2;               // CS points added
  google.protobuf.Timestamp completed_at = 3;
  bytes verifier_hash = 4;               // SHA256 of verifier plug-in
  bytes proof_hash = 5;                  // SHA256 of proof data
  string assistant_address = 6;          // AI assistant who verified
  uint64 block_height = 7;
  string tx_hash = 8;
  MultiplierInfo multipliers = 9;        // Applied bonuses
}

message MultiplierInfo {
  float arena_multiplier = 1;            // 1.0-1.5x
  float velocity_multiplier = 2;         // 1.0-1.25x
  float jackpot_multiplier = 3;          // 1.0, 5.0, or 25.0x
}
```

**IRCompletionEvent (Indexed for queries):**
```protobuf
message IRCompletionEvent {
  string wallet_address = 1;
  string ir_id = 2;
  uint64 score_added = 3;
  uint64 new_total_score = 4;
  google.protobuf.Timestamp timestamp = 5;
  string assistant_address = 6;
  bytes proof_hash = 7;
  string arena = 8;                      // For arena-specific queries
}
```

**ConfidenceHistory:**
```protobuf
message ConfidenceHistory {
  string wallet_address = 1;
  repeated ScoreChange changes = 2;
}

message ScoreChange {
  uint64 block_height = 1;
  int64 score_delta = 2;                 // Can be negative (slashing)
  uint64 new_total = 3;
  ChangeReason reason = 4;
  string related_ir_id = 5;              // If applicable
  string tx_hash = 6;
}

enum ChangeReason {
  IR_COMPLETION = 0;
  FRAUD_SLASH = 1;
  GOVERNANCE_ADJUSTMENT = 2;
  APPEAL_REVERSAL = 3;
}
```

### 2.2 State Storage Keys

```
# Primary indexes
confidence_scores/{wallet_address} -> UserConfidenceRecord
ir_completions/{wallet_address}/{ir_id} -> IRCompletion
anchor_registry/{wallet_address} -> AnchorInfo

# Secondary indexes (for queries)
verified_users/{wallet_address} -> bool  # CS ≥ 10,000
arena_completions/{arena}/{wallet_address} -> []IRCompletion
completion_events/{block_height}/{tx_hash} -> IRCompletionEvent

# History
score_history/{wallet_address}/{block_height} -> ScoreChange

# Rate limiting
ir_rate_limits/{wallet_address}/{ir_id}/{window_start} -> uint64
```

---

## 3. OPERATIONS

### 3.1 Messages

**MsgRecordIRCompletion:**
```protobuf
message MsgRecordIRCompletion {
  string wallet_address = 1;
  string ir_id = 2;
  bytes proof_hash = 3;                  // SHA256 of proof
  bytes verifier_hash = 4;               // SHA256 of verifier plug-in
  string assistant_address = 5;          // Signer (AI assistant)
  google.protobuf.Timestamp timestamp = 6;
}

message MsgRecordIRCompletionResponse {
  uint64 score_earned = 1;
  uint64 new_total_score = 2;
  bool verification_achieved = 3;        // True if crossed 10,000
}
```

**Validation Rules:**
- Signer must be registered AI Assistant with active stake
- IR must exist in registry and be Active
- User must have completed IR-000 (except for IR-000 itself)
- IR must not already be completed by this user
- Prerequisites must be satisfied
- Rate limits must not be exceeded
- Proof hash must be 32 bytes (SHA256)

**MsgRecalculateScore (Admin/Governance):**
```protobuf
message MsgRecalculateScore {
  string wallet_address = 1;
  string authority = 2;                  // Governance module address
}

message MsgRecalculateScoreResponse {
  uint64 previous_score = 1;
  uint64 recalculated_score = 2;
  repeated string discrepancies = 3;     // If any found
}
```

**MsgSlashScore:**
```protobuf
message MsgSlashScore {
  string wallet_address = 1;
  string ir_id = 2;                      // IR being disputed
  string reason = 3;
  uint64 slash_amount = 4;               // CS points to deduct
  string authority = 5;                  // Governance or fraud detector
}

message MsgSlashScoreResponse {
  uint64 previous_score = 1;
  uint64 new_score = 2;
  bool verification_revoked = 3;         // True if dropped below 10,000
}
```

**MsgAppealSlash:**
```protobuf
message MsgAppealSlash {
  string wallet_address = 1;
  string slash_tx_hash = 2;
  string evidence = 3;                   // IPFS hash or metadata
  uint64 deposit = 4;                    // AEQ deposit for appeal
}
```

### 3.2 Query Operations

**QueryUserScore:**
```protobuf
service ConfidenceScoreQuery {
  rpc GetUserScore(QueryUserScoreRequest) returns (QueryUserScoreResponse);
  rpc GetUserCompletions(QueryUserCompletionsRequest) returns (QueryUserCompletionsResponse);
  rpc GetScoreHistory(QueryScoreHistoryRequest) returns (QueryScoreHistoryResponse);
  rpc GetThresholds(QueryThresholdsRequest) returns (QueryThresholdsResponse);
  rpc GetVerifiedUsers(QueryVerifiedUsersRequest) returns (QueryVerifiedUsersResponse);
  rpc GetArenaBreakdown(QueryArenaBreakdownRequest) returns (QueryArenaBreakdownResponse);
}

message QueryUserScoreRequest {
  string wallet_address = 1;
}

message QueryUserScoreResponse {
  uint64 total_score = 1;
  bool is_verified = 2;
  AnchorInfo anchor_info = 3;
  ArenaBreakdown arena_scores = 4;
  uint64 ir_count = 5;
  google.protobuf.Timestamp last_updated = 6;
}
```

**QueryUserCompletions:**
```protobuf
message QueryUserCompletionsRequest {
  string wallet_address = 1;
  string arena_filter = 2;               // Optional
  cosmos.base.query.v1beta1.PageRequest pagination = 3;
}

message QueryUserCompletionsResponse {
  repeated IRCompletion completions = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```

**QueryScoreHistory:**
```protobuf
message QueryScoreHistoryRequest {
  string wallet_address = 1;
  uint64 from_height = 2;                // Optional
  uint64 to_height = 3;                  // Optional
  cosmos.base.query.v1beta1.PageRequest pagination = 4;
}

message QueryScoreHistoryResponse {
  repeated ScoreChange changes = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```

**QueryThresholds:**
```protobuf
message QueryThresholdsRequest {}

message QueryThresholdsResponse {
  uint64 verified_human_threshold = 1;  // 10,000
  map<string, uint64> vc_thresholds = 2; // VC type -> threshold
  map<string, uint64> arena_focus_thresholds = 3;
}
```

---

## 4. BUSINESS LOGIC

### 4.1 Score Calculation Rules

**Base Score Addition:**
```go
func CalculateScoreEarned(irID string, userAddr string) (uint64, error) {
    ir := GetIR(irID)
    baseScore := ir.ScoreValue

    // Apply arena multiplier if focusing
    arenaMultiplier := CalculateArenaMultiplier(userAddr, ir.Arena)

    // Apply velocity bonus
    velocityMultiplier := CalculateVelocityBonus(userAddr)

    // Check for jackpot
    jackpotMultiplier := CheckJackpot(userAddr, currentHeight, irID)

    finalScore := float64(baseScore) * arenaMultiplier * velocityMultiplier * jackpotMultiplier

    return uint64(finalScore), nil
}
```

**Arena Focus Multiplier:**
```go
func CalculateArenaMultiplier(userAddr string, arena string) float64 {
    record := GetUserRecord(userAddr)
    arenaScore := record.ArenaScores[arena]

    // No bonus until user has 3,000+ CS in this arena
    if arenaScore < 3000 {
        return 1.0
    }

    // Graduated bonus: 1.1x at 3k, 1.2x at 4k, 1.5x at 5k+
    if arenaScore >= 5000 {
        return 1.5
    } else if arenaScore >= 4000 {
        return 1.2
    } else {
        return 1.1
    }
}
```

**Velocity Bonus (from tokenomics spec):**
```go
func CalculateVelocityBonus(userAddr string) float64 {
    record := GetUserRecord(userAddr)
    if !record.AnchorInfo.Completed {
        return 1.0
    }

    anchorTime := record.AnchorInfo.CompletedAt
    currentTime := ctx.BlockTime()
    duration := currentTime.Sub(anchorTime)
    days := duration.Hours() / 24

    if record.TotalScore >= 10000 {
        // Bonus already applied
        return 1.0
    }

    switch {
    case days <= 7:
        return 1.25  // 25% bonus
    case days <= 30:
        return 1.10  // 10% bonus
    default:
        return 1.00  // No bonus
    }
}
```

**Jackpot Check (probabilistic):**
```go
func CheckJackpot(userAddr string, blockHeight uint64, irID string) float64 {
    // Deterministic but unpredictable
    seed := sha256.Sum256([]byte(userAddr + strconv.FormatUint(blockHeight, 10) + irID))
    seedInt := binary.BigEndian.Uint64(seed[:8])

    // 1 in 100 chance for 5x
    if seedInt % 100 == 77 {
        return 5.0
    }

    // 1 in 1000 chance for 25x
    if seedInt % 1000 == 888 {
        return 25.0
    }

    return 1.0
}
```

### 4.2 IR Prerequisite Validation

**Prerequisite Graph Traversal:**
```go
func ValidatePrerequisites(userAddr string, irID string) error {
    ir := GetIR(irID)
    prerequisites := GetIRPrerequisites(irID)

    userRecord := GetUserRecord(userAddr)
    completedIRs := make(map[string]bool)
    for _, completion := range userRecord.CompletedIRs {
        completedIRs[completion.IRID] = true
    }

    // Check all prerequisites
    for _, prereqID := range prerequisites {
        if !completedIRs[prereqID] {
            return fmt.Errorf("prerequisite %s not completed", prereqID)
        }
    }

    return nil
}
```

**Circular Dependency Prevention:**
```
- IR registry module enforces DAG structure
- Cycle detection during IR creation/update
- Prerequisites only reference older/existing IRs
```

### 4.3 Arena Focus Bonuses

**Arena Focus Badge Requirements:**
```
Arena Focus Badge Criteria:
1. User must be Verified (CS ≥ 10,000)
2. User must have ≥ 5,000 CS from single arena
3. Badge grants social signaling, not additional CS
```

**VC Issuance for Focus:**
```protobuf
message VCArenaFocus {
  string arena = 1;                      // e.g., "Biometric"
  uint64 arena_score = 2;                // CS from that arena
  google.protobuf.Timestamp achieved_at = 3;
}
```

### 4.4 Score Degradation Over Time (Staleness)

**Design Decision:** Initial implementation does NOT include time-based degradation.

**Rationale:**
- Identity verification is binary, not time-sensitive
- User can self-revoke VC if needed
- Verifiers can set their own freshness policies
- Governance can adjust via parameter change if needed

**Future Enhancement (if governance approves):**
```yaml
staleness_params:
  enabled: false
  degradation_rate: 0  # CS points per year
  minimum_retained: 0  # Never drop below this
  exempted_arenas: []  # High-assurance IRs exempt
```

### 4.5 Fraud/Slash Handling

**Fraud Detection Flow:**
```
1. AI Assistant network flags suspicious IR completion
2. Consensus of assistants votes on fraud claim
3. If >66% vote fraud, governance proposal auto-created
4. Proposal includes evidence, slash amount, affected wallets
5. Governance votes on slash execution
6. If approved, slash applied and user notified
```

**Slash Execution:**
```go
func ExecuteSlash(walletAddr string, irID string, slashAmount uint64) error {
    record := GetUserRecord(walletAddr)

    // Deduct score
    if record.TotalScore < slashAmount {
        record.TotalScore = 0
    } else {
        record.TotalScore -= slashAmount
    }

    // Update status if dropped below threshold
    if record.TotalScore < 10000 && record.Status == VERIFIED {
        record.Status = UNVERIFIED
        RevokeVC(walletAddr, "VC:isVerifiedHuman")
    }

    // Log to history
    LogScoreChange(walletAddr, -int64(slashAmount), FRAUD_SLASH, irID)

    return SaveUserRecord(record)
}
```

**Appeal Process:**
```
1. User submits MsgAppealSlash with deposit
2. Governance reviews evidence
3. If valid, slash reversed + deposit returned
4. If invalid, deposit slashed to treasury
```

---

## 5. INTEGRATION POINTS

### 5.1 Consumes from Inclusion Routines Module

**Dependencies:**
- IR definitions (score values, prerequisites, arena)
- IR status (Active, Suspended, Retired)
- Rate limit configurations

**Integration:**
```go
type IRRegistry interface {
    GetIR(irID string) (*IRDefinition, error)
    GetIRPrerequisites(irID string) ([]string, error)
    GetRateLimit(irID string) (*RateLimitConfig, error)
    IsIRActive(irID string) bool
}

func (k Keeper) SetIRRegistry(registry IRRegistry) {
    k.irRegistry = registry
}
```

### 5.2 Provides to VC Registry Module

**Exports:**
- User verification status (CS ≥ 10,000)
- Arena focus achievements
- Completed IR list for policy checks

**Integration:**
```go
type ConfidenceScoreProvider interface {
    GetUserScore(walletAddr string) (uint64, error)
    IsVerified(walletAddr string) bool
    HasCompletedIR(walletAddr string, irID string) bool
    GetArenaScore(walletAddr string, arena string) (uint64, error)
}

// VC Registry module uses this to gate VC issuance
func (vcKeeper) IssueVC(walletAddr string, vcType string) error {
    if vcType == "VerifiedHuman" {
        if !csProvider.IsVerified(walletAddr) {
            return ErrInsufficientConfidenceScore
        }
    }

    if vcType == "AgeOver21" {
        if !csProvider.HasCompletedIR(walletAddr, "IR-GOV_ID_01") {
            return ErrMissingRequiredIR
        }
    }

    // Issue VC...
}
```

### 5.3 Receives from AI Assistant Network

**Attestation Format:**
```protobuf
message IRAttestation {
  string wallet_address = 1;
  string ir_id = 2;
  bytes proof_hash = 3;                  // SHA256 of proof
  bytes verifier_hash = 4;               // SHA256 of verifier plug-in
  google.protobuf.Timestamp timestamp = 5;
  bytes assistant_signature = 6;
}
```

**Validation:**
```go
func (k Keeper) ValidateAttestation(attestation IRAttestation) error {
    // Verify assistant is registered and bonded
    assistant := k.assistantKeeper.GetAssistant(attestation.AssistantAddress)
    if assistant == nil || !assistant.Active {
        return ErrInvalidAssistant
    }

    // Verify signature
    if !VerifySignature(attestation, assistant.PublicKey) {
        return ErrInvalidSignature
    }

    // Verify timestamp freshness (within 5 minutes)
    if time.Since(attestation.Timestamp) > 5*time.Minute {
        return ErrStaleAttestation
    }

    return nil
}
```

### 5.4 Hooks into Tokenomics Module

**PoI Reward Calculation:**
```go
func (k Keeper) RecordIRCompletion(ctx sdk.Context, msg MsgRecordIRCompletion) error {
    // Calculate CS score earned
    scoreEarned := CalculateScoreEarned(msg.IRID, msg.WalletAddress)

    // Update user record
    UpdateUserScore(msg.WalletAddress, scoreEarned, msg)

    // Trigger PoI reward via tokenomics module
    k.tokenomicsKeeper.DistributePoIReward(ctx, PoIRewardRequest{
        WalletAddress: msg.WalletAddress,
        IRID: msg.IRID,
        AssistantAddress: msg.AssistantAddress,
        Multipliers: multipliers,
    })

    return nil
}
```

---

## 6. PARAMETERS

### 6.1 Module Parameters

```protobuf
message Params {
  // Verification thresholds
  uint64 verified_human_threshold = 1;   // Default: 10,000
  uint64 high_assurance_threshold = 2;   // Default: 15,000
  uint64 arena_focus_threshold = 3;      // Default: 5,000

  // Score adjustment limits
  uint64 max_slash_percentage = 4;       // Default: 50%
  uint64 min_appeal_deposit = 5;         // Default: 1000 AEQ

  // Rate limiting
  uint64 max_irs_per_day = 6;            // Default: 10
  uint64 max_irs_per_hour = 7;           // Default: 3

  // Staleness (future)
  bool staleness_enabled = 8;            // Default: false
  uint64 degradation_rate_per_year = 9;  // Default: 0

  // Arena multipliers
  float arena_multiplier_tier1 = 10;     // Default: 1.1 (at 3k)
  float arena_multiplier_tier2 = 11;     // Default: 1.2 (at 4k)
  float arena_multiplier_tier3 = 12;     // Default: 1.5 (at 5k)

  // Velocity bonuses
  uint64 velocity_tier1_days = 13;       // Default: 7
  float velocity_tier1_bonus = 14;       // Default: 1.25
  uint64 velocity_tier2_days = 15;       // Default: 30
  float velocity_tier2_bonus = 16;       // Default: 1.10

  // Jackpot probabilities
  uint64 jackpot_5x_odds = 17;           // Default: 100 (1 in 100)
  uint64 jackpot_25x_odds = 18;          // Default: 1000 (1 in 1000)
}
```

### 6.2 VC-Specific Thresholds

```yaml
vc_thresholds:
  "VerifiedHuman":
    cs_required: 10000
    required_irs: ["IR-000"]

  "AgeOver18":
    cs_required: 10000
    required_irs: ["IR-000", "IR-GOV_ID_01"]

  "AgeOver21":
    cs_required: 10000
    required_irs: ["IR-000", "IR-GOV_ID_01"]
    region_specific: true

  "HighAssurance":
    cs_required: 15000
    required_irs: ["IR-000"]
    required_arena_count:
      high_assurance: 2

  "ArenaFocus:*":
    cs_required: 10000
    arena_specific_cs: 5000
```

### 6.3 Slash Parameters

```yaml
slash_config:
  fraud_detected:
    amount_percentage: 25  # 25% of total CS
    min_amount: 1000
    max_amount: 5000

  appeal_process:
    deposit_required: 1000  # AEQ
    review_period: 14  # days
    refund_on_success: true

  assistant_slash:
    false_positive: 5  # % of stake
    false_negative: 2  # % of stake
    downtime: 0.1  # % per day
```

---

## 7. EVENTS

### 7.1 Event Definitions

**EventIRCompleted:**
```protobuf
message EventIRCompleted {
  string wallet_address = 1;
  string ir_id = 2;
  uint64 score_earned = 3;
  uint64 new_total_score = 4;
  string assistant_address = 5;
  string arena = 6;
  uint64 block_height = 7;
  MultiplierInfo multipliers = 8;
}
```

**EventVerificationAchieved:**
```protobuf
message EventVerificationAchieved {
  string wallet_address = 1;
  uint64 final_score = 2;
  google.protobuf.Timestamp achieved_at = 3;
  uint64 ir_count = 4;
  uint64 days_since_anchor = 5;
}
```

**EventArenaFocusAchieved:**
```protobuf
message EventArenaFocusAchieved {
  string wallet_address = 1;
  string arena = 2;
  uint64 arena_score = 3;
  google.protobuf.Timestamp achieved_at = 4;
}
```

**EventScoreSlashed:**
```protobuf
message EventScoreSlashed {
  string wallet_address = 1;
  string ir_id = 2;
  uint64 slash_amount = 3;
  uint64 new_score = 4;
  string reason = 5;
  bool verification_revoked = 6;
}
```

**EventJackpotTriggered:**
```protobuf
message EventJackpotTriggered {
  string wallet_address = 1;
  string ir_id = 2;
  float multiplier = 3;                  // 5.0 or 25.0
  uint64 bonus_score = 4;
}
```

---

## 8. SECURITY CONSIDERATIONS

### 8.1 Sybil Resistance

**Multi-Layered Protection:**
1. **IR-000 Anchor:** Government ID verification prevents easy identity creation
2. **High CS Threshold:** 10,000 points requires significant effort/cost
3. **Prerequisite Chains:** Forces sequential completion, time investment
4. **Rate Limiting:** Prevents rapid farming
5. **Cost of Fraud:** PoI rewards diminish relative to effort after ~2M users
6. **AI Assistant Staking:** Assistants risk stake for false attestations

### 8.2 Replay Attack Prevention

```go
func (k Keeper) CheckReplay(walletAddr string, irID string, proofHash []byte) error {
    // Check if this exact proof hash already used
    key := fmt.Sprintf("proof_hashes/%s/%x", walletAddr, proofHash)
    if k.store.Has(key) {
        return ErrReplayDetected
    }

    // Store hash to prevent reuse
    k.store.Set(key, []byte{1})

    return nil
}
```

### 8.3 Front-Running Protection

- IR completions are atomic transactions
- No mempool-visible state changes
- Assistant signatures time-bound (5 min expiry)
- Block-based jackpot randomness (not predictable pre-commit)

### 8.4 State Integrity

**Invariants Checked:**
```go
func (k Keeper) ValidateInvariants(ctx sdk.Context) error {
    // 1. Total CS = sum of all IR scores
    // 2. No completions without anchor
    // 3. No duplicate IR completions
    // 4. Arena scores sum to total score
    // 5. Verified users have CS ≥ 10,000
    // 6. All IRs in completions exist in registry
}
```

---

## 9. TESTING STRATEGY

### 9.1 Unit Tests

```
- Score calculation accuracy (base + multipliers)
- Prerequisite validation (valid/invalid chains)
- Rate limit enforcement
- Slash execution (boundary cases)
- Arena focus calculations
- Velocity bonus timing
- Jackpot probability distribution
```

### 9.2 Integration Tests

```
- End-to-end IR completion flow
- VC issuance gating
- Cross-module interactions (IR registry, tokenomics)
- Event emission verification
- Query response accuracy
```

### 9.3 Simulation Tests

```
- 10M user verification simulation
- Arena distribution analysis
- Fraud detection efficacy
- Slash/appeal process
- Economic sustainability
```

---

## 10. MIGRATION & DEPLOYMENT

### 10.1 Genesis State

```json
{
  "confidence_score": {
    "params": {
      "verified_human_threshold": "10000",
      "high_assurance_threshold": "15000",
      "arena_focus_threshold": "5000",
      "max_slash_percentage": "50",
      "min_appeal_deposit": "1000000000",
      "max_irs_per_day": "10",
      "max_irs_per_hour": "3"
    },
    "user_records": [],
    "completion_events": []
  }
}
```

### 10.2 Upgrade Path

**From identitychange module:**
```go
func MigrateFromIdentityChange(ctx sdk.Context, oldKeeper identitychange.Keeper) error {
    // Iterate all existing identity records
    oldKeeper.IterateIdentityRecords(ctx, func(record identitychange.IdentityRecord) {
        // Create new UserConfidenceRecord
        newRecord := types.UserConfidenceRecord{
            WalletAddress: record.Owner,
            TotalScore: uint64(record.ConfidenceScore),
            // Migrate other fields...
        }

        SaveUserRecord(ctx, newRecord)
    })
}
```

---

## 11. FUTURE ENHANCEMENTS

### 11.1 Planned Features

1. **Score Delegation:** Allow users to delegate verification status to sub-wallets
2. **VC Templates:** Pre-configured VC policies for common use cases
3. **Cross-Chain Verification:** IBC-enabled CS verification
4. **Dynamic Thresholds:** Adjust thresholds based on network fraud rates
5. **ML-Based Fraud Detection:** Real-time anomaly detection

### 11.2 Research Areas

1. **Privacy-Preserving CS Queries:** ZK proofs of CS without revealing exact score
2. **Reputation Decay Models:** Time-based adjustments for stale identities
3. **Multi-Sig Verification:** Require multiple assistants for high-value IRs
4. **Biometric Uniqueness Scoring:** Advanced deduplication via ML embeddings

---

## APPENDIX A: EXAMPLE FLOWS

### Flow 1: New User Verification Journey

```
Day 0: User downloads wallet
  → Completes IR-000 (Anchor) → CS = 0 (prerequisite only)

Day 1: User completes 3 biometric IRs
  → IR-101 (50 CS) + IR-102 (300 CS) + IR-105 (350 CS)
  → Total CS = 700

Day 3: User completes knowledge IRs
  → IR-304 (500 CS) + IR-307 (300 CS) + IR-315 (800 CS)
  → Total CS = 2,300

Day 5: User completes high-assurance IR
  → IR-601 (2000 CS)
  → Total CS = 4,300

Day 6: User completes geo-location + social IRs
  → IR-501 (500 CS) + IR-402 (350 CS) + IR-507 (700 CS)
  → Total CS = 5,850

Day 7: User completes possession + knowledge IRs
  → IR-203 (400 CS) + IR-308 (250 CS) + IR-320 (1500 CS)
  → IR-204 (400 CS) + IR-322 (500 CS) + IR-323 (300 CS)
  → Total CS = 9,200

Day 8: Final push to verification
  → IR-606 (1400 CS)
  → Total CS = 10,600
  → Velocity Bonus Applied: 1.25x on final reward
  → EventVerificationAchieved emitted
  → VC:isVerifiedHuman issued
```

### Flow 2: Arena Focus Achievement

```
User at CS = 8,500 (Verified status pending)
Arena Scores:
  - Biometric: 2,100
  - Knowledge: 3,200
  - Geo-Location: 500
  - Others: 2,700

Goal: Achieve Biometric Arena Focus

Step 1: Complete IR-106 (400 CS)
  → Biometric arena: 2,500 CS
  → Total: 8,900 CS

Step 2: Complete IR-108 (700 CS)
  → Biometric arena: 3,200 CS
  → Total: 9,600 CS
  → Arena multiplier activates: 1.1x

Step 3: Complete IR-115 (600 CS)
  → Base: 600 × 1.2 (arena multiplier) = 720 CS
  → Biometric arena: 3,920 CS
  → Total: 10,320 CS
  → EventVerificationAchieved emitted

Step 4: Complete IR-116 (400 CS)
  → Base: 400 × 1.2 = 480 CS
  → Biometric arena: 4,400 CS
  → Total: 10,800 CS

Step 5: Complete IR-112 (450 CS)
  → Base: 450 × 1.5 (max multiplier) = 675 CS
  → Biometric arena: 5,075 CS
  → Total: 11,475 CS
  → EventArenaFocusAchieved emitted
  → VC:ArenaFocus:Biometric issued
```

---

## APPENDIX B: GLOSSARY

**ConfidenceScore (CS):** Aggregate numerical score representing identity verification strength

**IR (Inclusion Routine):** Verifiable task that earns CS points

**Arena:** Category of IRs (Biometric, Knowledge, etc.)

**Anchor (IR-000):** Mandatory government ID verification prerequisite

**Arena Focus:** Specialization bonus for concentrating CS in single arena

**Velocity Bonus:** Time-based multiplier for rapid verification completion

**Jackpot:** Probabilistic bonus multiplier (5x or 25x)

**Slash:** CS point deduction for fraud detection

**PoI (Proof-of-Identity):** Token rewards for completing IRs

**VC (Verifiable Credential):** Privacy-preserving proof of attribute

**Verification Status:** Binary state (CS ≥ 10,000 = Verified)

---

**END OF DESIGN SPECIFICATION**
