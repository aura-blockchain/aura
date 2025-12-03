---
id: "044"
title: "Compliance No KYC Provider Authorization"
status: ready
priority: p1
category: security
module: compliance
severity: CRITICAL
cvss: 9.5
source: compliance-audit
---

# Compliance No KYC Provider Authorization

## Problem

Anyone can submit KYC records for any address. No verification that the submitter is an authorized KYC provider.

## Affected Files

- `chain/x/compliance/keeper/msg_server.go:27-53`

## Vulnerability

```go
func (s *msgServer) SubmitKYC(goCtx context.Context, req *types.MsgSubmitKYC) (*types.MsgSubmitKYCResponse, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
    }
    if req.Address == "" {
        return nil, status.Error(codes.InvalidArgument, "address is required")
    }
    // NO CHECK: Is sender authorized to submit KYC?
    // NO CHECK: Is req.Provider a registered provider?
    // NO CHECK: Is req.VerificationId valid?

    record := &types.KYCRecord{
        Address:        req.Address,        // Any address!
        KycLevel:       req.KycLevel,       // Any level!
        Provider:       req.Provider,       // Any provider name!
        VerificationId: req.VerificationId, // Any ID!
    }
    s.Keeper.SetKYCRecord(ctx, record)  // Stored!
}
```

## Attack Scenario

```go
// Attacker grants themselves KYC level ADVANCED
msg := MsgSubmitKYC{
    Address: "attacker_address",
    KycLevel: KYCLevel_KYC_LEVEL_ADVANCED,
    Provider: "TotallyLegitKYC",
    VerificationId: "FAKE123",
    Documents: []string{"passport", "utility_bill"},
    Jurisdiction: "US",
}
// Succeeds! Attacker now has highest KYC level
// Can bypass all KYC-gated features
```

## Required Fix

```go
// Add to params
message ComplianceParams {
    repeated string approved_kyc_providers = 20;
    map<string, string> kyc_provider_addresses = 21; // provider name -> address
}

func (s *msgServer) SubmitKYC(goCtx context.Context, req *types.MsgSubmitKYC) (*types.MsgSubmitKYCResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // 1. Validate inputs
    if req == nil || req.Address == "" {
        return nil, status.Error(codes.InvalidArgument, "invalid request")
    }

    // 2. Get params
    params := s.Keeper.GetParams(ctx)

    // 3. Verify provider is in approved list
    isApproved := false
    for _, provider := range params.ApprovedKYCProviders {
        if provider == req.Provider {
            isApproved = true
            break
        }
    }
    if !isApproved {
        return nil, status.Errorf(codes.PermissionDenied,
            "provider '%s' is not authorized", req.Provider)
    }

    // 4. Verify sender is the provider's registered address
    signers := req.GetSigners()
    if len(signers) == 0 {
        return nil, status.Error(codes.Unauthenticated, "no signers")
    }

    providerAddr := params.KycProviderAddresses[req.Provider]
    if signers[0].String() != providerAddr {
        return nil, status.Errorf(codes.PermissionDenied,
            "sender %s is not authorized address for provider %s",
            signers[0].String(), req.Provider)
    }

    // 5. Optionally: Verify with external provider API
    // verified, err := s.Keeper.VerifyExternalKYC(ctx, req.Provider, req.VerificationId)
    // if err != nil || !verified {
    //     return nil, status.Error(codes.FailedPrecondition, "external verification failed")
    // }

    // 6. Now safe to store
    record := &types.KYCRecord{
        Address:        req.Address,
        KycLevel:       req.KycLevel,
        Provider:       req.Provider,
        VerificationId: req.VerificationId,
        // ...
    }
    s.Keeper.SetKYCRecord(ctx, record)

    return &types.MsgSubmitKYCResponse{Success: true}, nil
}
```

## Acceptance Criteria

- [ ] Approved KYC providers list in params
- [ ] Provider address registration
- [ ] Signer verification matches provider address
- [ ] Governance proposal to add/remove providers
- [ ] Tests for unauthorized provider rejection
- [ ] Tests for address mismatch rejection
