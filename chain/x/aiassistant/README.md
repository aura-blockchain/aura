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
