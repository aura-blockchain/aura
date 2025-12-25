# Data Registry Module

## Overview

The Data Registry module provides decentralized data storage and verification for the Aura blockchain. It manages on-chain metadata for off-chain data stored on IPFS, supporting various data types from vehicle registrations to golf scores with multi-level verification and access control.

## Features

- **Multi-Type Data Storage**: Support for 20+ data types including vehicle registration, property deeds, photos, medical records, NFTs, and custom data
- **IPFS Integration**: Content-addressed storage with automatic pinning and unpinning
- **Multi-Level Verification**: Self-attested, peer-verified, AI-verified, authority-verified, and blockchain-anchored verification levels
- **Access Control**: Private, whitelist, public, and verified-user access modes with policy enforcement
- **Geo-Location Support**: Geographic tagging with radius-based search capabilities
- **Verification Tracking**: Immutable verification history with confidence scoring
- **Search & Discovery**: Tag-based and type-based data item search with access control filtering

## State

### Data Items
- **DataItem**: Core data object with metadata, content hash, storage location, verification status, and access policy
- **DataItemType**: Enum for data categories (vehicle, property, medical, media, achievements, etc.)
- **DataItemStatus**: Lifecycle states (pending, verified, rejected, expired, revoked)

### Verification
- **Verification**: Verification record with verifier address, verification level, timestamp, and notes
- **VerificationLevel**: Trust levels from self-attested to blockchain-anchored

### Access Control
- **AccessPolicy**: Defines who can access data via mode, allowed/denied addresses, and verification requirements
- **AccessMode**: Private, whitelist, public, or verified-users-only

### Specialized Data
- **VehicleRegistrationData**: VIN, make, model, year, license plate
- **PhotoData**: Camera metadata, GPS coordinates, timestamp
- **GolfScoreData**: Course, date, score, handicap

## Messages

### MsgStoreDataItem
Store a new data item with metadata and IPFS content.

### MsgUpdateDataItem
Update metadata or access policy for existing data.

### MsgDeleteDataItem
Delete data item and unpin from IPFS.

### MsgVerifyDataItem
Add verification to a data item (authority, peer, or AI).

### MsgRevokeDataItem
Revoke verification or mark data as invalid.

## Queries

### QueryDataItem
Retrieve data item by ID with access control check.

### QueryUserDataItems
List all data items owned by an address with optional type/status filters.

### QuerySearchDataItems
Search data items by tags, type, geo-location, and access permissions.

### QueryDataItemVerifications
Get verification history for a data item.

### QueryStats
Get registry statistics (total items, verified items, storage bytes, items by type).

## Events

### EventDataItemStored
Emitted when new data item is stored.

**Attributes**: `data_id`, `data_type`, `owner`, `storage_location`

### EventDataItemUpdated
Emitted when data item metadata is updated.

**Attributes**: `data_id`, `owner`

### EventDataItemDeleted
Emitted when data item is deleted and unpinned.

**Attributes**: `data_id`, `owner`

### EventDataItemVerified
Emitted when verification is added to data item.

**Attributes**: `data_id`, `verifier`, `verification_level`, `confidence_score`

### EventDataItemRevoked
Emitted when data item verification is revoked.

**Attributes**: `data_id`, `authority`

## Integration Notes

- Data content is stored off-chain on IPFS; only metadata and hashes are on-chain
- Access control is enforced at query time
- Verification levels build on each other (higher levels imply lower levels)
- Geo-location search uses simplified distance calculation (production should use Haversine formula)
- IPFS pinning/unpinning happens automatically during store/delete operations
