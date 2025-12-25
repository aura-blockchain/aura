# AI Assistant Module

## Overview

The AI Assistant module manages registration, lifecycle, and reputation tracking for AI assistants on the Aura blockchain. It provides proof-of-verification mechanisms, locale-based assistant discovery, heartbeat monitoring, and misbehavior reporting with slashing capabilities.

## Features

- **Assistant Registration**: Register AI assistants with locale support, verification proofs, and endpoint configuration
- **Status Management**: Active, jailed, and tombstoned status tracking with automated slashing
- **Locale Support**: Multi-language/region support for assistant discovery and filtering
- **Heartbeat Monitoring**: Periodic liveness checks with automatic jailing for non-responsive assistants
- **Misbehavior Reporting**: Reputation-based slashing system for low-quality or malicious assistants
- **Balance Tracking**: Per-assistant token balance management for staking and rewards
- **Query by Locale**: Discover assistants supporting specific languages and regions

## State

### Assistant Records
- **Assistant**: Core assistant information including address, locales, status, reputation score, and endpoint
- **AssistantStatus**: Enum values - UNSPECIFIED, ACTIVE, JAILED, TOMBSTONED
- **Balance**: Token balance tracking per assistant
- **Params**: Module parameters including minimum stake, slash percentages, and heartbeat intervals

### Status Tracking
- **Reputation Score**: Integer score decremented on misbehavior reports
- **Jailed/Tombstoned**: Assistants can be temporarily jailed or permanently tombstoned
- **Last Heartbeat**: Timestamp of most recent liveness check

## Messages

### MsgRegisterAssistant
Register a new AI assistant with locale support.

**Fields**: `address`, `locales`, `verification_proof`, `endpoint`

### MsgUpdateLocales
Update supported locales for an existing assistant.

**Fields**: `address`, `locales`

### MsgHeartbeat
Submit liveness proof to avoid jailing.

**Fields**: `address`, `timestamp`

### MsgReportMisbehavior
Report assistant misbehavior with evidence hash.

**Fields**: `reporter`, `assistant_address`, `evidence_hash`, `description`

### MsgUpdateParams
Update module parameters (authority only).

**Fields**: `authority`, `params`

## Queries

### QueryAssistant
Get assistant details by address.

### QueryAssistants
List all registered assistants.

### QueryAssistantsByLocale
Filter assistants by supported locale.

### QueryParams
Get module parameters.

## Events

### EventRegisterAssistant
Emitted when new AI assistant is registered.

**Attributes**: `assistant_address`, `owner_address`, `model_hash`, `stake_amount`, `locales`

### EventUpdateLocales
Emitted when assistant updates supported locales.

**Attributes**: `assistant_address`, `old_locales`, `new_locales`

### EventHeartbeat
Emitted on successful heartbeat submission.

**Attributes**: `assistant_address`, `heartbeat_latency`, `next_slash_time`

### EventHeartbeatFailure
Emitted when assistant misses heartbeat and is slashed.

**Attributes**: `assistant_address`, `failure_count`, `slashed_amount`

### EventReportMisbehavior
Emitted when misbehavior is reported and slashing occurs.

**Attributes**: `assistant_address`, `reporter_address`, `reason`, `slashed_amount`, `new_status`

### EventSlashAssistant
Emitted when assistant is slashed for any reason.

**Attributes**: `assistant_address`, `slash_fraction`, `slashed_amount`, `remaining_stake`

### EventStatusChange
Emitted when assistant status changes (ACTIVE/JAILED/TOMBSTONED).

**Attributes**: `assistant_address`, `old_status`, `new_status`

## Integration Notes

- Assistants must maintain minimum stake to remain active
- Heartbeat failures result in automatic slashing and jailing
- Misbehavior reports require evidence and can lead to permanent tombstoning
- Locale filtering enables users to find assistants in their language/region
- Reputation scoring affects assistant visibility and rewards
