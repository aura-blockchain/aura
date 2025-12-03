---
id: "040"
title: "Weak ZK Proof Verification - Accepts Any Bytes"
status: ready
priority: p1
category: security
module: identity
severity: CRITICAL
cvss: 10.0
source: identity-privacy-audit
---

# Weak ZK Proof Verification - Accepts Any Bytes

## Problem

The ZK proof verification only checks `len(proof) > 0`. Any non-empty byte array is accepted as a valid proof.

## Affected Files

- `chain/x/identity/keeper/keeper.go:578-586`

## Vulnerability

```go
func (k *Keeper) VerifyZKProof(ctx sdk.Context, proof []byte, publicInputs []byte) (bool, error) {
    // BUG: Only checks length, not actual proof validity
    if len(proof) == 0 {
        return false, fmt.Errorf("empty proof")
    }

    // MISSING: Actual cryptographic verification
    // Should verify proof against verification key and public inputs

    return true, nil  // ALWAYS RETURNS TRUE IF PROOF NOT EMPTY
}
```

## Impact

- Anyone can bypass identity verification
- Zero-knowledge proofs provide zero security
- Identity system completely compromised
- Can claim any identity without proof

## Required Fix

```go
import (
    "github.com/consensys/gnark-crypto/ecc"
    "github.com/consensys/gnark/backend/groth16"
    "github.com/consensys/gnark/frontend"
)

// Store verification keys for different proof types
type ZKVerificationKey struct {
    ProofType string
    Key       groth16.VerifyingKey
    Circuit   frontend.Circuit
}

func (k *Keeper) VerifyZKProof(ctx sdk.Context, proofType string, proof []byte, publicInputs []byte) (bool, error) {
    if len(proof) == 0 {
        return false, fmt.Errorf("empty proof")
    }

    if len(publicInputs) == 0 {
        return false, fmt.Errorf("empty public inputs")
    }

    // Get verification key for this proof type
    vk, err := k.GetVerificationKey(ctx, proofType)
    if err != nil {
        return false, fmt.Errorf("unknown proof type: %s", proofType)
    }

    // Deserialize proof
    zkProof := groth16.NewProof(ecc.BN254)
    if _, err := zkProof.ReadFrom(bytes.NewReader(proof)); err != nil {
        return false, fmt.Errorf("invalid proof format: %w", err)
    }

    // Deserialize public inputs
    witness, err := frontend.NewWitness(nil, ecc.BN254.ScalarField())
    if err != nil {
        return false, fmt.Errorf("failed to create witness: %w", err)
    }

    // Populate witness with public inputs
    if err := deserializePublicInputs(witness, publicInputs, vk.Circuit); err != nil {
        return false, fmt.Errorf("invalid public inputs: %w", err)
    }

    // ACTUALLY VERIFY THE PROOF
    publicWitness, err := witness.Public()
    if err != nil {
        return false, fmt.Errorf("failed to extract public witness: %w", err)
    }

    if err := groth16.Verify(zkProof, vk.Key, publicWitness); err != nil {
        // Proof verification failed - this is expected for invalid proofs
        return false, nil
    }

    return true, nil
}

// Example verification key registration
func (k *Keeper) RegisterVerificationKey(ctx sdk.Context, proofType string, vkBytes []byte, circuitType string) error {
    // Parse verification key
    vk := groth16.NewVerifyingKey(ecc.BN254)
    if _, err := vk.ReadFrom(bytes.NewReader(vkBytes)); err != nil {
        return fmt.Errorf("invalid verification key: %w", err)
    }

    // Store
    k.SetVerificationKey(ctx, proofType, &ZKVerificationKey{
        ProofType: proofType,
        Key:       vk,
    })

    return nil
}
```

## Acceptance Criteria

- [ ] Actual ZK proof verification implemented
- [ ] Verification key storage and management
- [ ] Support for multiple proof types (groth16, plonk, etc.)
- [ ] Public input validation
- [ ] Tests with valid and invalid proofs
- [ ] Integration with gnark or similar ZK library
