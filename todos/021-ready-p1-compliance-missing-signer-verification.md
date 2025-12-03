---
id: "021"
title: "Missing Signer Verification in Compliance Module"
status: ready
priority: p1
category: security
module: compliance
severity: CRITICAL
cvss: 8.9
source: security-audit-report
---

# Missing Signer Verification in Compliance Module

## Problem

All compliance message handlers lack signer verification. Attackers can submit KYC records, report suspicious activity, or manipulate compliance data for any address.

## Affected Files

- `chain/x/compliance/keeper/msg_server.go`

## Affected Functions

1. `SubmitKYC()` - Line 27
2. `ReportSuspiciousActivity()` - Line 55
3. `ScreenSanctions()` - Line 82

## Attack Scenarios

### KYC Spoofing
```go
// Attacker submits fake KYC for victim
msg := &MsgSubmitKYC{
    Address: "victim_address",
    KycLevel: KYC_LEVEL_ADVANCED,  // Highest level
    Provider: "fake-kyc-provider",
    VerificationId: "fake-verification",
}
// Victim now appears KYC verified
// Can bypass regulatory controls
```

### False Flagging
```go
// Competitor flags legitimate business as suspicious
msg := &MsgReportSuspiciousActivity{
    Address: "competitor_address",
    TransactionHash: "legitimate-tx-123",
    ActivityType: "MONEY_LAUNDERING",
    Description: "suspicious pattern detected",
}
// Competitor gets flagged and frozen
```

## Impact

- Regulatory compliance bypass
- False identity verification
- Malicious flagging of legitimate users
- AML/KYC system compromise

## Required Fix

```go
func (s *msgServer) SubmitKYC(goCtx context.Context, req *types.MsgSubmitKYC) (*types.MsgSubmitKYCResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // CRITICAL: Verify submitter is authorized KYC provider
    params := s.Keeper.GetParams(ctx)

    // Check if provider is in approved list
    isAuthorized := false
    for _, authorizedProvider := range params.ApprovedKYCProviders {
        if authorizedProvider == req.Provider {
            isAuthorized = true
            break
        }
    }
    if !isAuthorized {
        return nil, status.Error(codes.PermissionDenied, "provider not authorized")
    }

    // Get registered provider address
    providerAddr, err := s.Keeper.GetKYCProviderAddress(ctx, req.Provider)
    if err != nil {
        return nil, status.Error(codes.NotFound, "provider not registered")
    }

    // Verify sender matches provider address
    signers := req.GetSigners()
    if len(signers) == 0 {
        return nil, status.Error(codes.Unauthenticated, "no signers")
    }

    if !signers[0].Equals(providerAddr) {
        return nil, status.Error(codes.PermissionDenied, "sender not authorized provider")
    }

    // Rest of function...
}
```

## Acceptance Criteria

- [ ] KYC provider authorization system implemented
- [ ] Signer verification on all message handlers
- [ ] Approved providers list in module params
- [ ] Tests for unauthorized KYC submission rejection
- [ ] Tests for unauthorized SAR filing rejection
