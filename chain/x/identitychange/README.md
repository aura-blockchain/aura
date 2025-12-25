# Identity Change Module

## Overview

The Identity Change module tracks holder-driven DID rotations, metadata refreshes, and confidence score updates once AI assistants re-verify an inclusion routine. It provides a complete lifecycle for identity change requests with verification, approval, rejection, and suspension capabilities.

## Features

- **Identity Change Requests**: Submit DID rotation and metadata update requests
- **AI Assistant Verification**: AI assistants provide verification proofs for identity changes
- **Change History Tracking**: Complete audit trail of all identity modifications
- **Rate Limiting**: Configurable rate limits to prevent spam requests
- **Emergency Suspension**: Governance can suspend all identity changes during security events
- **Status Lifecycle**: IDLE → PENDING_VERIFICATION → READY_TO_APPLY → APPLIED/REJECTED

## State

### Core Types
- **IdentityRecord**: Current identity state for each holder (DID, metadata, confidence score)
- **IdentityChangeRequest**: Pending change request with status and verification data
- **IdentityChangeHistory**: Historical record of completed identity changes
- **Params**: Module parameters including rate limits and verification requirements

### Status Enum
- **IDLE**: No active change request
- **PENDING_VERIFICATION**: Awaiting AI assistant verification
- **READY_TO_APPLY**: Verified and ready for application
- **REJECTED**: Change request rejected
- **APPLIED**: Change successfully applied
- **SUSPENDED**: Changes suspended by governance

## Messages

### MsgRequestIdentityChange
Submit identity change request with new DID or metadata.

**Fields**: `holder`, `new_did`, `new_metadata`, `reason`

### MsgSubmitAssistantProof
AI assistant submits verification proof for pending request.

**Fields**: `assistant`, `request_id`, `verification_proof`, `confidence_score`

### MsgApplyIdentityChange
Apply verified identity change to permanent record.

**Fields**: `holder`, `request_id`

### MsgRejectIdentityChange
Reject identity change request with reason.

**Fields**: `authority`, `request_id`, `reason`

### MsgSuspendIdentityChanges
Suspend all identity change processing (governance only).

**Fields**: `authority`, `suspend`, `reason`

## Queries

### QueryIdentityRecord
Get current identity record for holder.

### QueryIdentityChangeRequest
Get pending change request by ID.

### QueryIdentityChangeHistory
Get complete change history for holder.

### QueryParams
Get module parameters.

## Events

### EventIdentityChangeRequested
Emitted when new identity change is requested.

**Attributes**: `holder`, `request_id`, `timestamp`

### EventVerificationSubmitted
Emitted when AI assistant submits verification.

**Attributes**: `request_id`, `assistant`, `confidence_score`

### EventIdentityChangeApplied
Emitted when identity change is applied.

**Attributes**: `holder`, `request_id`, `new_did`

### EventIdentityChangeRejected
Emitted when identity change is rejected.

**Attributes**: `request_id`, `reason`

### EventIdentityChangesSuspended
Emitted when all changes are suspended.

**Attributes**: `suspended`, `reason`

## Integration Notes

- Rate limits enforce maximum requests per holder per time period
- AI assistant verification required before changes can be applied
- Governance can suspend all changes during upgrades or security incidents
- Complete audit trail maintained for compliance and forensics
- Integrates with identity module for DID verification
