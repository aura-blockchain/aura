---
id: "018"
title: "Multi-Signature Weight Verification Missing"
status: ready
priority: p1
category: security
module: walletsecurity
severity: CRITICAL
cvss: 9.3
source: security-audit-report
---

# Multi-Signature Weight Verification Missing

## Problem

The multi-sig implementation increments `CurrentWeight` without actually checking if the signer has weight assigned or validating signature weights. Line 156 blindly increments: `tx.CurrentWeight++`

## Affected Files

- `chain/x/walletsecurity/keeper/msg_server.go:156`

## Vulnerability

```go
// Current code at line 156:
tx.Signatures = append(tx.Signatures, string(msg.Signature))
tx.SignedBy = append(tx.SignedBy, msg.Signer)
tx.CurrentWeight++  // WRONG: Should use actual signer weight

// What if wallet uses weighted signatures?
// SignerWeights: {"alice": 40, "bob": 60}
// WeightThreshold: 100
//
// Current code allows alice to sign twice and get weight = 2
// Should require alice (40) + bob (60) = 100
```

## Impact

- Threshold bypass
- Unauthorized transaction execution
- Fund theft

## Required Fix

```go
func (ms msgServer) SignMultiSigTransaction(goCtx context.Context, msg *wspb.MsgSignMultiSigTransaction) (*wspb.MsgSignMultiSigTransactionResponse, error) {
    // ... signer verification ...

    // Get wallet configuration
    walletBytes, err := ms.Keeper.GetMultiSigWallet(ctx, tx.WalletId)
    if err != nil {
        return nil, status.Error(codes.NotFound, "wallet not found")
    }

    var wallet wspb.MultiSigWallet
    ms.Keeper.cdc.Unmarshal(walletBytes, &wallet)

    // Verify signer is authorized for this wallet
    isAuthorized := false
    for _, authorizedSigner := range wallet.Signers {
        if authorizedSigner == msg.Signer {
            isAuthorized = true
            break
        }
    }
    if !isAuthorized {
        return nil, status.Error(codes.PermissionDenied, "signer not authorized for this wallet")
    }

    // Check for duplicate signature
    for _, existingSigner := range tx.SignedBy {
        if existingSigner == msg.Signer {
            return nil, status.Error(codes.AlreadyExists, "signer already signed this transaction")
        }
    }

    // Calculate actual weight to add
    signerWeight := int32(1) // Default weight
    if wallet.SignerWeights != nil {
        if weight, ok := wallet.SignerWeights[msg.Signer]; ok {
            signerWeight = weight
        }
    }

    // CRYPTOGRAPHICALLY VERIFY THE SIGNATURE
    msgToSign := fmt.Sprintf("%s:%s:%s", tx.WalletId, msg.TxId, tx.TxData)
    msgHash := sha256.Sum256([]byte(msgToSign))

    if !verifySignature(msg.Signer, msgHash[:], msg.Signature) {
        return nil, status.Error(codes.Unauthenticated, "invalid signature")
    }

    // Add signature and weight
    tx.Signatures = append(tx.Signatures, string(msg.Signature))
    tx.SignedBy = append(tx.SignedBy, msg.Signer)
    tx.CurrentWeight += signerWeight  // Use actual weight

    // Check threshold (use weight threshold if configured)
    requiredWeight := wallet.Threshold
    if wallet.WeightThreshold > 0 {
        requiredWeight = wallet.WeightThreshold
    }

    readyToExecute := tx.CurrentWeight >= requiredWeight

    // Rest of function...
}
```

## Acceptance Criteria

- [ ] Proper weight calculation implemented
- [ ] Duplicate signature prevention added
- [ ] Cryptographic signature verification added
- [ ] Signer authorization check added
- [ ] Tests for weighted multi-sig scenarios
- [ ] Tests for duplicate signature rejection
