---
status: pending
priority: p2
issue_id: "114"
tags: [code-review, security, zkp, cryptography]
dependencies: ["100"]
---

# P2 HIGH: ZK Circuit Constraints Not Cryptographically Verified

## Problem Statement

The ZKP module implements zero-knowledge proofs but the circuit constraints are not formally verified, meaning the proofs might not actually prove what they claim.

**Why it matters:** Unsound ZK circuits can allow attackers to generate fake proofs, defeating the entire privacy/compliance system.

## Findings

### Current Implementation

**File:** `/home/decri/blockchain-projects/aura/chain/x/zkp/keeper/verifier.go`

```go
func (k Keeper) VerifyProof(ctx sdk.Context, proofType string, proof []byte, publicInputs []byte) (bool, error) {
    // Load circuit for proof type
    circuit, err := k.loadCircuit(ctx, proofType)
    if err != nil {
        return false, err
    }

    // Verify using gnark
    err = groth16.Verify(proof, circuit.VerifyingKey, publicInputs)
    if err != nil {
        return false, nil // Invalid proof
    }

    return true, nil
}
```

### Issues Identified

1. **No constraint soundness verification** - Circuit constraints not formally proven
2. **No test vectors** - No test cases from known-good implementations
3. **Missing range checks** - Integer constraints may overflow
4. **No negative tests** - No tests for invalid proofs being rejected

### Specific Circuit Concerns

**Age Verification Circuit:**
```go
// CONCERN: Does this actually prove age > 18 without revealing birthdate?
type AgeVerificationCircuit struct {
    CurrentDate   frontend.Variable `gnark:",public"`
    AgeThreshold  frontend.Variable `gnark:",public"`
    BirthDate     frontend.Variable `gnark:",secret"`
    Commitment    frontend.Variable `gnark:",public"`
}
```

Questions:
- Is the commitment binding?
- Can the birthdate be extracted from commitment?
- Are there overflow attacks on date arithmetic?

## Proposed Solutions

### Solution A: Add Test Vectors and Negative Tests (Recommended)
**Effort:** 2-3 days | **Risk:** Low

```go
func TestAgeCircuit_SoundnessPositive(t *testing.T) {
    // Test that valid proof passes
    circuit := &AgeVerificationCircuit{
        CurrentDate:  2024_01_01,
        AgeThreshold: 18,
        BirthDate:    2000_01_01,
    }
    proof := generateProof(t, circuit)
    require.True(t, verifyProof(proof, circuit.PublicInputs()))
}

func TestAgeCircuit_SoundnessNegative(t *testing.T) {
    // Test that invalid birthdate fails
    circuit := &AgeVerificationCircuit{
        CurrentDate:  2024_01_01,
        AgeThreshold: 18,
        BirthDate:    2010_01_01, // Only 14 years old
    }
    _, err := generateProof(t, circuit)
    require.Error(t, err) // Should fail to generate valid proof
}

func TestAgeCircuit_Malleability(t *testing.T) {
    // Test that modified proofs fail
    circuit := &AgeVerificationCircuit{...}
    proof := generateProof(t, circuit)

    // Modify proof bytes
    proof[10] ^= 0xff
    require.False(t, verifyProof(proof, circuit.PublicInputs()))
}
```

### Solution B: Formal Verification
**Effort:** 2-4 weeks | **Risk:** Medium

Engage cryptographic auditor to formally verify circuit constraints.

## Recommended Action

**GO WITH SOLUTION A FIRST**: Add comprehensive tests immediately. Consider Solution B before mainnet.

## Technical Details

### Affected Files

- `chain/x/zkp/keeper/verifier.go`
- `chain/x/zkp/keeper/circuits/`
- `chain/x/zkp/keeper/verifier_test.go`

### Test Categories Needed

1. **Completeness** - Valid proofs verify
2. **Soundness** - Invalid statements rejected
3. **Zero-knowledge** - Secrets not leaked
4. **Malleability** - Modified proofs fail
5. **Range checks** - Overflow prevented

## Acceptance Criteria

- [ ] Test vectors for all circuit types
- [ ] Negative tests for each constraint
- [ ] Overflow/range check tests
- [ ] Malleability resistance tests
- [ ] Documentation of security assumptions
- [ ] Consider formal audit for mainnet

## Work Log

| Date | Action | Result |
|------|--------|--------|
| 2025-12-08 | Security audit identified verification gap | P2 High |

## Resources

- [gnark Documentation](https://docs.gnark.consensys.net/)
- [ZK Circuit Vulnerabilities](https://blog.trailofbits.com/tag/zero-knowledge/)
- [Soundness in ZK Proofs](https://www.zeroknowledgeblog.com/index.php/the-basics-of-zkps)
