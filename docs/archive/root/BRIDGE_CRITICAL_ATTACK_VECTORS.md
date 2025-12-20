# Bridge Module - Critical Attack Vectors

**CONFIDENTIAL - SECURITY DOCUMENT**

This document describes the most severe attack vectors that could be exploited to drain the Aura bridge.

---

## Attack Vector 1: Validator Set Manipulation

### Exploitability: CRITICAL
### Impact: CATASTROPHIC (Total bridge compromise)

```
┌─────────────────────────────────────────────────────────┐
│ ATTACKER CONTROLS GENESIS FILE                          │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
         ┌─────────────────────┐
         │ Add Malicious       │
         │ Validators to       │
         │ Genesis             │
         └──────────┬──────────┘
                    │
                    ▼
        ┌──────────────────────┐
        │ Chain Launches       │
        │ 5 Validators:        │
        │ - 3 Attacker         │
        │ - 2 Honest           │
        └──────────┬───────────┘
                   │
                   ▼
     ┌───────────────────────────┐
     │ Attacker Submits          │
     │ MsgMintTokens:            │
     │ Amount: 1,000,000 tokens  │
     │ No signature required!    │
     └───────────┬───────────────┘
                 │
                 ▼
       ┌──────────────────────┐
       │ 3 Malicious          │
       │ Validators Attest    │
       │ (No crypto proof!)   │
       └──────────┬───────────┘
                  │
                  ▼
         ┌────────────────────┐
         │ Threshold Reached  │
         │ (3/5 validators)   │
         │ TOKENS MINTED!     │
         └────────┬───────────┘
                  │
                  ▼
    ┌──────────────────────────────┐
    │ Attacker Dumps Tokens        │
    │ Drains All Bridge Liquidity  │
    │ GAME OVER                    │
    └──────────────────────────────┘
```

### Why This Works
1. **No genesis validator validation** - Any validator set accepted
2. **No minimum validator count** - Even 1 validator would work
3. **No signature verification in MintTokens** - Just attestation count
4. **No cryptographic proof required** - Validators can lie

### Proof of Concept
```go
// In genesis.json
"bridge": {
  "validators": [
    {
      "address": "attacker1",
      "public_key": "...",
      "power": 1,
      "active": true
    },
    {
      "address": "attacker2",
      "public_key": "...",
      "power": 1,
      "active": true
    },
    {
      "address": "attacker3",
      "public_key": "...",
      "power": 1,
      "active": true
    }
  ]
}

// After chain launch
tx := MsgMintTokens{
  Validator: "attacker1",
  SourceChain: "ethereum",
  SourceTxHash: "0xFAKE_HASH",
  Recipient: "attacker_aura_address",
  Amount: "1000000000000", // 1 million tokens
  Denom: "uaura",
  ValidatorSignature: nil, // Not checked anyway!
}
// Submit from attacker1, attacker2, attacker3
// Threshold reached, tokens minted, no real deposit on Ethereum
```

### Maximum Loss
**UNLIMITED** - All bridge liquidity + inflated supply

### Mitigation Required
See CRITICAL-1, CRITICAL-4, CRITICAL-7 in main report

---

## Attack Vector 2: Replay Attack - Infinite Minting

### Exploitability: CRITICAL
### Impact: CATASTROPHIC (Unlimited token creation)

```
┌──────────────────────────────────────────┐
│ LEGITIMATE USER BRIDGES 10,000 TOKENS   │
│ From Ethereum → Aura                     │
│ TX Hash: 0xABCD1234                      │
└─────────────────┬────────────────────────┘
                  │
                  ▼
       ┌──────────────────────┐
       │ Validators Submit    │
       │ MsgMintTokens        │
       │ TxHash: 0xABCD1234   │
       └──────────┬───────────┘
                  │
                  ▼
         ┌────────────────────┐
         │ Code Checks Index  │
         │ transferIDByHash   │
         │ (0xABCD1234)       │
         └────────┬───────────┘
                  │
                  ▼
    ┌────────────────────────────┐
    │ If Not Found:              │
    │ Creates NEW transferID     │
    │ transferID = "transfer-1"  │
    └────────────┬───────────────┘
                 │
                 ▼
      ┌──────────────────────┐
      │ Mints 10,000 Tokens  │
      │ Status: COMPLETED    │
      └──────────┬───────────┘
                 │
                 ▼
    ┌───────────────────────────────┐
    │ ATTACKER WAITS 1 HOUR         │
    └───────────┬───────────────────┘
                │
                ▼
    ┌──────────────────────────────────┐
    │ ATTACKER'S MALICIOUS VALIDATORS  │
    │ Submit MsgMintTokens AGAIN       │
    │ TxHash: 0xABCD1234 (SAME!)       │
    └───────────┬──────────────────────┘
                │
                ▼
       ┌─────────────────────────┐
       │ Code Checks Index       │
       │ transferIDByHash        │
       │ FINDS: "transfer-1"     │
       └───────────┬─────────────┘
                   │
                   ▼
      ┌──────────────────────────┐
      │ Transfer Already Exists  │
      │ Status: COMPLETED        │
      │ BUT...                   │
      └───────────┬──────────────┘
                  │
                  ▼
    ┌────────────────────────────────┐
    │ SubmitAttestation Still Works! │
    │ Attestation counter increments │
    └────────────┬───────────────────┘
                 │
                 ▼
       ┌──────────────────────────┐
       │ CheckAttestationThreshold│
       │ Threshold Reached AGAIN! │
       └──────────┬───────────────┘
                  │
                  ▼
         ┌────────────────────┐
         │ Mints ANOTHER      │
         │ 10,000 Tokens!     │
         └────────┬───────────┘
                  │
                  ▼
    ┌──────────────────────────────┐
    │ REPEAT INFINITELY            │
    │ Same TxHash, Unlimited Mints │
    │ 10,000 real deposit becomes  │
    │ 10,000,000 minted tokens     │
    └──────────────────────────────┘
```

### Why This Works
1. **No nonce tracking** - Despite protobuf definition, never implemented
2. **Attestation system is stateful** - Can be triggered multiple times
3. **No permanent marking of processed transactions** - Hash index can be manipulated
4. **No transfer status check before minting** - Already-completed transfers can trigger mints

### Proof of Concept
```go
// Step 1: Legitimate mint
msg1 := &MsgMintTokens{
    Validator: "validator1",
    SourceChain: "ethereum",
    SourceTxHash: "0xABCD1234",
    Recipient: "aura1user",
    Amount: "10000000000",
    Denom: "uaura",
}
// Gets minted → transferID: "transfer-1", Status: COMPLETED

// Step 2: Wait for completion

// Step 3: Replay attack
msg2 := &MsgMintTokens{
    Validator: "validator2", // Different validator
    SourceChain: "ethereum",
    SourceTxHash: "0xABCD1234", // SAME HASH!
    Recipient: "aura1attacker",
    Amount: "10000000000",
    Denom: "uaura",
}
// Finds existing transfer but attestation logic still executes
// Threshold reached again → MINTS AGAIN

// Repeat with validator3, validator4, etc.
```

### Maximum Loss
**UNLIMITED** - Each source transaction can be replayed infinitely

### Mitigation Required
See CRITICAL-2 in main report

---

## Attack Vector 3: Signature Replay - Unlock Without Burn

### Exploitability: HIGH
### Impact: CRITICAL (All locked tokens stolen)

```
┌────────────────────────────────────────────┐
│ USER BURNS 50,000 WRAPPED TOKENS ON AURA  │
│ To unlock native tokens                    │
└─────────────────┬──────────────────────────┘
                  │
                  ▼
       ┌──────────────────────────┐
       │ Validators Create        │
       │ Signatures for Unlock:   │
       │ SHA256(chain:hash:       │
       │ sender:amount:denom)     │
       └──────────┬───────────────┘
                  │
                  ▼
      ┌────────────────────────────┐
      │ User Submits               │
      │ MsgUnlockTokens with       │
      │ ValidatorSignatures[]      │
      └──────────┬─────────────────┘
                 │
                 ▼
    ┌──────────────────────────────┐
    │ Code Verifies Signatures     │
    │ pubKey.VerifySignature() ✓   │
    │ Threshold Met ✓              │
    └──────────┬───────────────────┘
                │
                ▼
       ┌──────────────────────┐
       │ Sends 50,000 Tokens  │
       │ Status: COMPLETED    │
       │ EMITS EVENT WITH     │
       │ SIGNATURES!          │
       └──────────┬───────────┘
                  │
                  ▼
    ┌───────────────────────────────────┐
    │ ATTACKER OBSERVES BLOCKCHAIN      │
    │ Extracts signatures from event    │
    └───────────┬───────────────────────┘
                │
                ▼
    ┌──────────────────────────────────────┐
    │ ATTACKER SUBMITS NEW MsgUnlockTokens │
    │ With SAME signatures                 │
    │ But different transfer context       │
    └───────────┬──────────────────────────┘
                │
                ▼
       ┌─────────────────────────┐
       │ Code Verifies Signatures│
       │ Signatures are VALID!   │
       │ (They're real validator │
       │  signatures)            │
       └──────────┬──────────────┘
                  │
                  ▼
         ┌────────────────────┐
         │ NO STATUS CHECK    │
         │ before SendCoins!  │
         └────────┬───────────┘
                  │
                  ▼
      ┌──────────────────────────┐
      │ Sends ANOTHER            │
      │ 50,000 Tokens!           │
      └──────────┬───────────────┘
                 │
                 ▼
    ┌──────────────────────────────┐
    │ REPEAT with same signatures  │
    │ Until module account empty   │
    │                              │
    │ 1 legitimate burn becomes:   │
    │ - 50,000 tokens              │
    │ - 50,000 tokens (replay 1)   │
    │ - 50,000 tokens (replay 2)   │
    │ - ... (unlimited)            │
    └──────────────────────────────┘
```

### Why This Works
1. **Signatures are not bound to transfer ID** - Same signatures work for any transfer
2. **No signature uniqueness tracking** - Used signatures not stored
3. **No status check before SendCoins** - Completed transfers can be unlocked again
4. **Signatures don't expire** - Valid forever
5. **Message hash is predictable** - Easy to reconstruct and replay

### Proof of Concept
```go
// Step 1: Legitimate unlock (observe on-chain)
legitimateMsg := &MsgUnlockTokens{
    Sender: "aura1user",
    SourceChain: "ethereum",
    BurnTxHash: "0xBURN123",
    Amount: "50000000000",
    Denom: "uaura",
    ValidatorSignatures: [][]byte{sig1, sig2, sig3}, // Extract these from events
}
// Unlocks successfully, emits event with signatures

// Step 2: Attacker extracts signatures from event logs
observedSignatures := [][]byte{sig1, sig2, sig3}

// Step 3: Attacker replays with different transfer
attackMsg := &MsgUnlockTokens{
    Sender: "aura1attacker", // Attacker's address
    SourceChain: "ethereum",
    BurnTxHash: "0xFAKE456", // Different hash
    Amount: "50000000000", // Same amount
    Denom: "uaura",
    ValidatorSignatures: observedSignatures, // SAME SIGNATURES!
}
// Signatures verify (they're real!) → Tokens sent to attacker

// Repeat until module account is empty
```

### Maximum Loss
**ALL LOCKED TOKENS** - Limited only by module account balance

### Mitigation Required
See CRITICAL-3 in main report

---

## Attack Vector 4: Circuit Breaker Bypass - Rapid Drain

### Exploitability: MEDIUM (Requires exploiting another vulnerability)
### Impact: CRITICAL (All liquidity stolen in seconds)

```
┌──────────────────────────────────────┐
│ ATTACKER EXPLOITS MINTING VULN       │
│ (e.g., malicious validators)         │
└─────────────────┬────────────────────┘
                  │
                  ▼
       ┌──────────────────────────┐
       │ Prepares 1000 Transactions│
       │ Each minting 100,000 AURA│
       └──────────┬───────────────┘
                  │
                  ▼
    ┌────────────────────────────────┐
    │ BLOCK N: All 1000 Txs Included │
    └────────────┬───────────────────┘
                 │
                 ▼
       ┌──────────────────────────┐
       │ Tx 1: MintTokens         │
       │ Check: MaxTransferAmount │
       │ 100,000 < 100,000,000 ✓  │
       │ Check: CircuitBreaker?   │
       │ NOT ENFORCED!            │
       │ → MINTS 100,000          │
       └──────────┬───────────────┘
                  │
                  ▼
       ┌──────────────────────────┐
       │ Tx 2: MintTokens         │
       │ → MINTS 100,000          │
       └──────────┬───────────────┘
                  │
                  ▼
       ┌──────────────────────────┐
       │ Tx 3-1000: All Succeed   │
       │ No hourly volume check   │
       │ No rate limiting         │
       └──────────┬───────────────┘
                  │
                  ▼
    ┌──────────────────────────────┐
    │ TOTAL MINTED IN BLOCK:       │
    │ 100,000,000 tokens           │
    │ Time elapsed: ~6 seconds     │
    │ No circuit breaker tripped   │
    └──────────┬───────────────────┘
                │
                ▼
     ┌──────────────────────────────┐
     │ BLOCK N+1: Attacker Dumps    │
     │ Drains All Bridge Liquidity  │
     │ Before Anyone Can React      │
     └──────────────────────────────┘
```

### Why This Works
1. **Circuit breaker not enforced in MintTokens** - Only checked in LockTokens
2. **No hourly volume tracking** - Parameter defined but never checked
3. **No failed transfer counting** - Could spam failed txs to DoS
4. **No automatic pause on anomalies** - Circuit breaker never trips
5. **Parameters not enforced** - MaxHourlyVolume exists but unused

### Proof of Concept
```go
// Prepare attack
for i := 0; i < 1000; i++ {
    msg := &MsgMintTokens{
        Validator: fmt.Sprintf("validator%d", i%3),
        SourceChain: "ethereum",
        SourceTxHash: fmt.Sprintf("0x%d", i), // Unique hashes
        Recipient: "aura1attacker",
        Amount: "100000000000", // 100k tokens
        Denom: "uaura",
    }
    // Submit all in same block via mempool flooding
}

// Result: 100 million tokens minted in single block
// No circuit breaker enforcement
// All liquidity drained before governance can respond
```

### Maximum Loss
**ALL BRIDGE LIQUIDITY** - No upper bound per block

### Mitigation Required
See CRITICAL-5 in main report

---

## Attack Vector 5: Validator Collusion - Coordinated Theft

### Exploitability: MEDIUM (Requires controlling 67%+ of validators)
### Impact: CATASTROPHIC (Complete bridge compromise)

```
┌──────────────────────────────────────────┐
│ ATTACKER CONTROLS 4 OF 5 VALIDATORS      │
│ (Could be through compromise, bribery,   │
│  or being malicious operators)           │
└─────────────────┬────────────────────────┘
                  │
                  ▼
       ┌──────────────────────────┐
       │ Validators Coordinate    │
       │ To Execute Attack        │
       └──────────┬───────────────┘
                  │
                  ▼
    ┌────────────────────────────────┐
    │ PHASE 1: Mint Fake Tokens      │
    │ - Submit MsgMintTokens         │
    │ - Claim deposits on Ethereum   │
    │ - 4/5 validators attest        │
    │ - Threshold met (67%)          │
    │ - Tokens minted                │
    └────────────┬───────────────────┘
                 │
                 ▼
    ┌────────────────────────────────┐
    │ PHASE 2: Prevent Detection     │
    │ - 4/5 validators ignore fraud  │
    │ - Fraud proofs submitted but   │
    │ - No validator action          │
    │ - Honest validator outvoted    │
    └────────────┬───────────────────┘
                 │
                 ▼
    ┌────────────────────────────────┐
    │ PHASE 3: Drain Liquidity       │
    │ - Dump minted tokens           │
    │ - Extract all real locked value│
    │ - Leave worthless wrapped      │
    │   tokens behind                │
    └────────────┬───────────────────┘
                 │
                 ▼
    ┌────────────────────────────────┐
    │ PHASE 4: Cover Tracks          │
    │ - 4/5 validators approve       │
    │ - "Invalid" fraud proofs       │
    │ - Update chain configs to      │
    │   disable affected chains      │
    │ - No slashing (they control    │
    │   governance)                  │
    └────────────────────────────────┘
```

### Why This Works
1. **No slashing implemented** - Malicious validators face no consequences
2. **No stake requirements** - Validators don't risk their own funds
3. **No reputation system** - Past behavior not tracked
4. **Simple majority threshold** - 67% is achievable with few validators
5. **No external verification** - Only validators validate

### Maximum Loss
**UNLIMITED** - Complete control of bridge

### Mitigation Required
- Increase validator set size (20+)
- Implement severe slashing (50%+ of stake)
- Require substantial stake ($1M+ per validator)
- Add external verification (light clients, Merkle proofs)
- Implement reputation-based weighting

---

## Likelihood vs Impact Matrix

```
Impact →
High   │ CRITICAL-1 │ CRITICAL-2 │ CRITICAL-5 │
       │ Genesis    │ Replay     │ No Circuit │
       │ Validator  │ Attack     │ Breaker    │
       ├────────────┼────────────┼────────────┤
       │ CRITICAL-4 │ CRITICAL-3 │            │
       │ No Sig     │ Sig Replay │            │
       │ Verify     │            │            │
Medium │            │            │ Vector 5   │
       │            │            │ Collusion  │
       │            │            │            │
Low    │            │            │            │
       │            │            │            │
       └────────────┴────────────┴────────────┘
         Low        Medium       High
                Likelihood →
```

---

## Defense in Depth Analysis

### Current Defenses (What Exists)

Layer 1: Input Validation
- ✅ Basic nil checks
- ✅ Address format validation
- ❌ Insufficient for production

Layer 2: Cryptographic Verification
- ⚠️ Partial (UnlockTokens only)
- ❌ Not in MintTokens
- ❌ No Merkle proofs

Layer 3: Access Control
- ❌ None

Layer 4: Rate Limiting
- ❌ None enforced

Layer 5: Monitoring
- ⚠️ Basic events
- ❌ Insufficient for detection

Layer 6: Emergency Response
- ❌ No pause mechanism
- ❌ No emergency withdrawal

**Total Defense Layers Active: 0.5 of 6**

### Required Defense Layers

```
┌─────────────────────────────────────┐
│ Layer 6: Emergency Response         │
│ - Pause, emergency withdrawal       │
├─────────────────────────────────────┤
│ Layer 5: Monitoring & Alerting      │
│ - Real-time anomaly detection       │
├─────────────────────────────────────┤
│ Layer 4: Rate Limiting              │
│ - Circuit breaker, daily limits     │
├─────────────────────────────────────┤
│ Layer 3: Access Control             │
│ - Governance, validator management  │
├─────────────────────────────────────┤
│ Layer 2: Cryptographic Verification │
│ - Signatures, Merkle proofs, nonces │
├─────────────────────────────────────┤
│ Layer 1: Input Validation           │
│ - Comprehensive validation          │
└─────────────────────────────────────┘
```

---

## Time to Exploit (TTE)

| Attack Vector | TTE (Skilled Attacker) | TTE (Novice) |
|---------------|------------------------|--------------|
| Vector 1: Genesis Manipulation | Hours | N/A (requires chain control) |
| Vector 2: Replay Attack | Minutes | Days |
| Vector 3: Signature Replay | Hours | Weeks |
| Vector 4: Rapid Drain | Seconds (after Vector 1-3) | N/A |
| Vector 5: Collusion | Days (coordination) | N/A |

---

## Detection Difficulty

| Attack Vector | Detection Difficulty | Time to Detect |
|---------------|---------------------|----------------|
| Vector 1 | HARD (looks legitimate) | Hours-Days |
| Vector 2 | MEDIUM (unusual patterns) | Minutes-Hours |
| Vector 3 | MEDIUM (duplicate unlocks) | Minutes |
| Vector 4 | EASY (volume spike) | Seconds |
| Vector 5 | HARD (authorized actions) | Days-Weeks |

---

## Recommended Monitoring Rules

### Critical Alerts (Immediate Response)
1. **Mint volume spike** - > 10x average in 1 hour
2. **Repeated source tx hash** - Same hash minted > 1 time
3. **Signature reuse** - Same signature used > 1 time
4. **Validator collusion** - Same validators attesting to suspicious patterns
5. **Supply divergence** - Wrapped token supply ≠ locked amount

### High Alerts (Investigate within 1 hour)
1. **Large single transfer** - > $100,000
2. **Rapid small transfers** - > 100 transfers/hour from single address
3. **New validator** - Validator set changes
4. **Failed fraud proof** - Fraud proof submitted but rejected
5. **Parameter changes** - Circuit breaker, limits modified

### Medium Alerts (Daily review)
1. **Validator performance** - < 95% attestation rate
2. **Stuck transfers** - Pending > 24 hours
3. **Fee anomalies** - Fee revenue deviates from volume

---

## Response Procedures

### If Attack Detected

**IMMEDIATE (< 1 minute):**
1. Trigger emergency pause (if implemented)
2. Alert all validators
3. Alert dev team
4. Alert community (Discord, Twitter)

**SHORT-TERM (< 1 hour):**
1. Analyze attack vector
2. Estimate damage
3. Prepare patch
4. Coordinate governance response

**MEDIUM-TERM (< 24 hours):**
1. Deploy emergency patch
2. Conduct full audit
3. File insurance claims
4. Communicate with users

**LONG-TERM (< 1 week):**
1. Implement permanent fixes
2. Compensate affected users
3. Security review
4. Post-mortem report

---

## Insurance Fund Requirements

Based on attack vectors:

**Minimum Insurance Fund:**
- Scenario 1 (Genesis): Not insurable (chain control)
- Scenario 2 (Replay): $10M (assumes max 10 replays per tx)
- Scenario 3 (Sig Replay): $50M (all locked tokens)
- Scenario 4 (Rapid Drain): $100M (all liquidity)
- Scenario 5 (Collusion): $100M (complete compromise)

**Recommended Insurance: $100M** to cover worst-case scenarios

---

**End of Critical Attack Vectors Document**
