---
id: "062"
title: "Compliance Messages Missing GetSigners Implementation"
status: done
priority: p2
category: security
module: compliance
severity: HIGH
source: compliance-audit
completed: 2025-12-02
---

# Compliance Messages Missing GetSigners Implementation

## Problem

Protobuf messages don't implement GetSigners() method required for Cosmos SDK transaction verification.

## Affected Files

- `proto/aura/compliance/v1beta1/compliance.proto:285-371`
- `proto/aura/compliance/v1beta1/msg_signers.go` (implementation)
- `proto/aura/compliance/v1beta1/msg_validation.go` (validation)
- `chain/x/compliance/types/codec.go` (registration)

## Impact

- Cannot verify message signatures
- Transaction security compromised
- SDK validation bypassed

## Resolution

### Implementation Approach

Used Option 1 (proto annotations) + manual Go implementation:
- Proto files have `cosmos.msg.v1.signer` annotations
- Manual GetSigners() implementation in `msg_signers.go` for fine-grained control
- Comprehensive ValidateBasic() for all messages in `msg_validation.go`

### Files Modified

1. **proto/aura/compliance/v1beta1/msg_signers.go** - GetSigners implementations
2. **proto/aura/compliance/v1beta1/msg_validation.go** - ValidateBasic implementations
3. **proto/aura/compliance/v1beta1/msg_validation_test.go** - Comprehensive validation tests
4. **chain/x/compliance/types/codec.go** - Registered MsgEraseGDPRData
5. **chain/x/compliance/keeper/msg_server_signer_test.go** - Added EraseGDPRData tests

### Messages Implemented

All compliance messages now have GetSigners() and ValidateBasic():

1. **MsgSubmitKYC** - Provider is signer
   - Validates: address, provider, KYC level, PII commitment (32 bytes), jurisdiction (ISO 3166-1 alpha-2)

2. **MsgReportSuspiciousActivity** - Reporter is signer
   - Validates: reporter, address, transaction hash, activity type, description length

3. **MsgScreenSanctions** - Address being screened is signer
   - Validates: address format

4. **MsgRecordGDPRConsent** - Consent subject is signer
   - Validates: address, consent type, consent version

5. **MsgRequestGDPRData** - Data subject is signer
   - Validates: address, request type (access, rectification, erasure, portability, restriction, objection)

6. **MsgEraseGDPRData** - Data subject is signer
   - Validates: address, erasure reason length (max 500 chars)

7. **MsgGenerateTaxReport** - Report requester is signer
   - Validates: address, tax year (YYYY format), jurisdiction (ISO 3166-1), report type, jurisdiction-reportType compatibility

### Security Properties

- Empty addresses rejected
- Invalid bech32 addresses rejected
- GetSigners returns empty array on invalid address (prevents panic)
- Signer must match the actor in the message (provider, reporter, or subject)
- All validation happens before transaction execution

### Test Coverage

**ValidateBasic Tests** (proto/aura/compliance/v1beta1/msg_validation_test.go):
- 7 message types × average 10 test cases = 70+ test cases
- Tests cover: valid messages, empty fields, invalid addresses, format validation, boundary conditions

**GetSigners Tests** (chain/x/compliance/keeper/msg_server_signer_test.go):
- Tests for all 7 message types
- Invalid address handling
- Empty field handling
- Signer verification enforcement

**Test Results**: All tests pass ✅

## Acceptance Criteria

- [x] GetSigners() implemented for all message types
- [x] ValidateBasic() implemented for all messages
- [x] Tests for signer verification
- [x] MsgEraseGDPRData registered in codec
- [x] All tests passing
