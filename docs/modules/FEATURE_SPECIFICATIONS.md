# AURA New Feature Specifications
## Three Major Feature Additions

**Date:** November 13, 2025
**Version:** 1.0
**Status:** Ready for Implementation

---

## Overview

This document specifies three new features for the AURA blockchain:

1. **QR Code Verification API** - Real-time VC verification for businesses
2. **Selective Disclosure** - User-controlled attribute visibility
3. **Data Registry Module** - Verified data storage (NEW MODULE)

---

# Feature #1: QR Code Verification API

## Summary

Allow businesses/entities to scan a QR code displayed by an AURA verified person and confirm in real-time that the verification is genuine (not a fake screenshot or spoofed app).

## Module Assignment

**Add to:** `vcregistry` module

## Technical Design

### 1.1 New Proto Messages

**File:** `proto/aura/vcregistry/v1beta1/presentation.proto`

```protobuf
syntax = "proto3";
package aura.vcregistry.v1beta1;

import "google/protobuf/timestamp.proto";
import "aura/vcregistry/v1beta1/vc_registry.proto";

// VCPresentation represents a QR code presentation of VCs
message VCPresentation {
  string presentation_id = 1;              // Unique ID for this presentation
  string holder_did = 2;                   // DID of person presenting
  string holder_address = 3;               // Bech32 address
  repeated string vc_ids = 4;              // VCs being presented
  google.protobuf.Timestamp created_at = 5; // When QR was generated
  uint64 nonce = 6;                        // Anti-replay nonce
  uint64 expires_in_seconds = 7;           // QR expiration (e.g., 300s = 5min)
  bytes signature = 8;                     // Holder's signature over presentation
  PresentationContext context = 9;         // What's being shown
}

// PresentationContext defines what attributes are disclosed
message PresentationContext {
  bool show_full_name = 1;
  bool show_age = 2;
  bool show_age_over_18 = 3;
  bool show_age_over_21 = 4;
  bool show_address = 5;
  bool show_city_state_only = 6;
  bool show_professional_license = 7;
  map<string, bool> custom_attributes = 8; // Extensible
}

// VerificationResult returned to verifier
message VerificationResult {
  bool is_valid = 1;                       // Overall validity
  string holder_did = 2;                   // Who is presenting
  repeated VCVerificationDetail vc_details = 3; // Each VC's status
  google.protobuf.Timestamp verified_at = 4;    // When verification occurred
  string verification_error = 5;           // Error message if invalid
  DiscloseableAttributes attributes = 6;   // Attributes shown
}

// Individual VC verification details
message VCVerificationDetail {
  string vc_id = 1;
  VCType vc_type = 2;
  VCStatus status = 3;
  bool is_valid = 4;
  bool is_expired = 5;
  bool is_revoked = 6;
  google.protobuf.Timestamp issued_at = 7;
  google.protobuf.Timestamp expires_at = 8;
}

// Attributes disclosed to verifier
message DiscloseableAttributes {
  string full_name = 1;                    // Only if show_full_name = true
  uint32 age = 2;                          // Only if show_age = true
  bool is_over_18 = 3;                     // Only if show_age_over_18 = true
  bool is_over_21 = 4;                     // Only if show_age_over_21 = true
  string full_address = 5;                 // Only if show_address = true
  string city_state = 6;                   // Only if show_city_state_only = true
  map<string, string> custom_attributes = 7; // Extensible
}

// MsgCreatePresentation - User creates a QR code presentation
message MsgCreatePresentation {
  string creator = 1;                      // User's address
  repeated string vc_ids = 2;              // VCs to present
  PresentationContext context = 3;         // What to show
  uint64 expires_in_seconds = 4;           // Default: 300 (5 min)
}

message MsgCreatePresentationResponse {
  string presentation_id = 1;
  string qr_code_data = 2;                 // Base64 encoded QR data
  google.protobuf.Timestamp expires_at = 3;
}

// Query service additions
service Query {
  // ... existing queries ...

  // Verify a presentation (called by business/verifier)
  rpc VerifyPresentation(QueryVerifyPresentationRequest)
      returns (QueryVerifyPresentationResponse);
}

message QueryVerifyPresentationRequest {
  string qr_code_data = 1;                 // Data scanned from QR code
  string verifier_address = 2;             // Who is verifying (optional)
}

message QueryVerifyPresentationResponse {
  VerificationResult result = 1;
}
```

### 1.2 Implementation Requirements

**Anti-Spoofing Measures:**

1. **Time-limited QR codes** - Expire after 5 minutes (configurable)
2. **Nonce-based replay protection** - Each presentation has unique nonce
3. **Cryptographic signature** - Signed by holder's private key
4. **On-chain verification** - Verifier queries blockchain directly
5. **Fresh data** - QR generation pulls current VC status from chain

**QR Code Format:**

```json
{
  "v": "1.0",
  "p": "<presentation_id>",
  "h": "<holder_did>",
  "vcs": ["VC-001", "VC-002"],
  "ctx": {"show_age_over_21": true},
  "exp": 1699900000,
  "n": 12345678,
  "sig": "<signature>"
}
```

Encoded as: `aura://verify?data=<base64_encoded_json>`

**Verification Flow:**

```
User (Mobile App)                 Blockchain                 Verifier (Business)
      |                                |                              |
      | 1. Generate Presentation       |                              |
      |------------------------------>|                              |
      |    MsgCreatePresentation       |                              |
      |                                |                              |
      | 2. Receive QR Code Data        |                              |
      |<------------------------------|                              |
      |                                |                              |
      | 3. Display QR Code             |                              |
      |                                |                              |
      |                                |    4. Scan QR Code          |
      |                                |<-----------------------------|
      |                                |                              |
      |                                |    5. Query VerifyPresentation
      |                                |<-----------------------------|
      |                                |                              |
      |                                | 6. Validate:                 |
      |                                |    - Signature               |
      |                                |    - Nonce fresh             |
      |                                |    - Not expired             |
      |                                |    - VCs valid/active        |
      |                                |    - Check revocation        |
      |                                |                              |
      |                                | 7. Return VerificationResult |
      |                                |----------------------------->|
      |                                |                              |
      |                                |    8. Show "✓ AURA Verified" |
      |                                |       or "✗ Invalid"         |
```

### 1.3 Security Considerations

**Threats Prevented:**

1. **Screenshot Attack** - QR expires after 5 min, screenshot becomes invalid
2. **Replay Attack** - Nonce prevents reusing old QR codes
3. **Fake App Attack** - Verifier queries blockchain, not trusting app UI
4. **MITM Attack** - Signature ensures data hasn't been tampered
5. **Revoked VC Attack** - Real-time blockchain query checks current status

**Privacy Protections:**

- User chooses what attributes to reveal
- Verifier only sees what user explicitly allows
- No PII stored in QR code (only references to on-chain VCs)
- Verifier address logged for audit (optional)

### 1.4 API Endpoints

**REST API:**
```
POST /aura/vcregistry/v1beta1/presentations
GET  /aura/vcregistry/v1beta1/presentations/{presentation_id}/verify
```

**gRPC:**
```
Msg/CreatePresentation
Query/VerifyPresentation
```

**CLI:**
```bash
# User creates QR code
aurad tx vcregistry create-presentation \
  --vc-ids="VC-001,VC-002" \
  --show-age-over-21 \
  --expires-in=300 \
  --from alice

# Business verifies QR code
aurad query vcregistry verify-presentation \
  --qr-data="aura://verify?data=eyJ2Ijo..." \
  --verifier-address=business1
```

---

# Feature #2: Selective Disclosure

## Summary

Allow users to control which identity attributes are stored separately and what gets displayed for each verification request. Enable voice commands like "AURA show my age" or web interface with checkboxes.

## Module Assignment

**Add to:** `vcregistry` module

## Technical Design

### 2.1 New Proto Messages

**File:** `proto/aura/vcregistry/v1beta1/attributes.proto`

```protobuf
syntax = "proto3";
package aura.vcregistry.v1beta1;

import "google/protobuf/timestamp.proto";

// AttributeVC represents a single identity attribute as a VC
message AttributeVC {
  string attribute_vc_id = 1;              // Unique ID
  string holder_address = 2;               // Owner
  AttributeType attribute_type = 3;        // What attribute
  bytes encrypted_value = 4;               // Encrypted attribute value
  bytes value_hash = 5;                    // Hash for ZK proofs
  google.protobuf.Timestamp issued_at = 6;
  google.protobuf.Timestamp expires_at = 7;
  VCStatus status = 8;
  string issuer = 9;                       // Who verified this attribute
  uint64 verification_level = 10;          // 1-100 confidence
}

// AttributeType defines types of identity attributes
enum AttributeType {
  ATTRIBUTE_TYPE_UNSPECIFIED = 0;

  // Personal attributes
  ATTRIBUTE_TYPE_FULL_NAME = 1;
  ATTRIBUTE_TYPE_FIRST_NAME = 2;
  ATTRIBUTE_TYPE_LAST_NAME = 3;
  ATTRIBUTE_TYPE_DATE_OF_BIRTH = 4;
  ATTRIBUTE_TYPE_AGE = 5;
  ATTRIBUTE_TYPE_GENDER = 6;

  // Contact attributes
  ATTRIBUTE_TYPE_EMAIL = 10;
  ATTRIBUTE_TYPE_PHONE = 11;
  ATTRIBUTE_TYPE_ADDRESS_FULL = 12;
  ATTRIBUTE_TYPE_ADDRESS_STREET = 13;
  ATTRIBUTE_TYPE_ADDRESS_CITY = 14;
  ATTRIBUTE_TYPE_ADDRESS_STATE = 15;
  ATTRIBUTE_TYPE_ADDRESS_ZIP = 16;
  ATTRIBUTE_TYPE_ADDRESS_COUNTRY = 17;

  // Government IDs
  ATTRIBUTE_TYPE_PASSPORT_NUMBER = 20;
  ATTRIBUTE_TYPE_DRIVERS_LICENSE = 21;
  ATTRIBUTE_TYPE_SSN = 22;
  ATTRIBUTE_TYPE_TAX_ID = 23;

  // Physical attributes
  ATTRIBUTE_TYPE_HEIGHT = 30;
  ATTRIBUTE_TYPE_WEIGHT = 31;
  ATTRIBUTE_TYPE_EYE_COLOR = 32;
  ATTRIBUTE_TYPE_HAIR_COLOR = 33;

  // Professional attributes
  ATTRIBUTE_TYPE_OCCUPATION = 40;
  ATTRIBUTE_TYPE_EMPLOYER = 41;
  ATTRIBUTE_TYPE_PROFESSIONAL_LICENSE = 42;
  ATTRIBUTE_TYPE_EDUCATION_LEVEL = 43;
  ATTRIBUTE_TYPE_DEGREE = 44;

  // Special certifications
  ATTRIBUTE_TYPE_SCUBA_CERTIFIED = 50;
  ATTRIBUTE_TYPE_PILOTS_LICENSE = 51;
  ATTRIBUTE_TYPE_SECURITY_CLEARANCE = 52;

  // Custom
  ATTRIBUTE_TYPE_CUSTOM = 100;
}

// DisclosurePolicy defines user's disclosure preferences
message DisclosurePolicy {
  string holder_address = 1;
  repeated AttributeDisclosureRule rules = 2;
  DisclosurePolicyMode default_mode = 3;  // Default: deny all
  google.protobuf.Timestamp updated_at = 4;
}

// Rule for disclosing a specific attribute
message AttributeDisclosureRule {
  AttributeType attribute_type = 1;
  DisclosurePolicyMode mode = 2;
  repeated string allowed_verifiers = 3;   // Whitelist (optional)
  uint64 max_disclosures_per_day = 4;      // Rate limit
}

enum DisclosurePolicyMode {
  DISCLOSURE_POLICY_MODE_DENY = 0;         // Never disclose
  DISCLOSURE_POLICY_MODE_ASK = 1;          // Prompt user each time
  DISCLOSURE_POLICY_MODE_ALLOW = 2;        // Always disclose
  DISCLOSURE_POLICY_MODE_CONDITIONAL = 3;  // Allow if conditions met
}

// DisclosureRequest - What verifier is asking for
message DisclosureRequest {
  string request_id = 1;
  string verifier_address = 2;
  string verifier_name = 3;                // "Joe's Bar", "DMV", etc.
  repeated AttributeType requested_attributes = 4;
  string purpose = 5;                      // "Age verification for alcohol"
  google.protobuf.Timestamp requested_at = 6;
  uint64 expires_in_seconds = 7;
}

// DisclosureResponse - User's answer to request
message DisclosureResponse {
  string request_id = 1;
  string holder_address = 2;
  bool approved = 3;
  repeated AttributeDisclosure disclosed_attributes = 4;
  google.protobuf.Timestamp responded_at = 5;
}

// Single attribute disclosure
message AttributeDisclosure {
  AttributeType attribute_type = 1;
  string revealed_value = 2;               // Decrypted value (if full disclosure)
  bytes zk_proof = 3;                      // ZK proof (if selective)
  bool is_zk_proof = 4;                    // True if using ZK, false if full reveal
}

// Voice command structure
message VoiceCommand {
  string holder_address = 1;
  string command_text = 2;                 // "AURA show my age"
  repeated AttributeType parsed_attributes = 3;
  google.protobuf.Timestamp issued_at = 4;
}

// Messages for transactions

message MsgCreateAttributeVC {
  string creator = 1;
  AttributeType attribute_type = 2;
  bytes encrypted_value = 3;
  string issuer = 4;
  uint64 expires_in_seconds = 5;
}

message MsgCreateAttributeVCResponse {
  string attribute_vc_id = 1;
}

message MsgUpdateDisclosurePolicy {
  string creator = 1;
  repeated AttributeDisclosureRule rules = 2;
  DisclosurePolicyMode default_mode = 3;
}

message MsgUpdateDisclosurePolicyResponse {}

message MsgRespondToDisclosureRequest {
  string creator = 1;
  string request_id = 2;
  bool approved = 3;
  repeated AttributeType disclosed_attributes = 4;
}

message MsgRespondToDisclosureRequestResponse {
  DisclosureResponse response = 1;
}

// Query service additions
service Query {
  // Get user's disclosure policy
  rpc DisclosurePolicy(QueryDisclosurePolicyRequest)
      returns (QueryDisclosurePolicyResponse);

  // Get user's attribute VCs
  rpc AttributeVCs(QueryAttributeVCsRequest)
      returns (QueryAttributeVCsResponse);

  // Parse voice command
  rpc ParseVoiceCommand(QueryParseVoiceCommandRequest)
      returns (QueryParseVoiceCommandResponse);
}

message QueryDisclosurePolicyRequest {
  string holder_address = 1;
}

message QueryDisclosurePolicyResponse {
  DisclosurePolicy policy = 1;
}

message QueryAttributeVCsRequest {
  string holder_address = 1;
  repeated AttributeType filter_types = 2; // Optional filter
}

message QueryAttributeVCsResponse {
  repeated AttributeVC attribute_vcs = 1;
}

message QueryParseVoiceCommandRequest {
  string command_text = 1;                 // "AURA show my age and address"
}

message QueryParseVoiceCommandResponse {
  repeated AttributeType parsed_attributes = 1;
  bool is_valid = 2;
  string error_message = 3;
}
```

### 2.2 Use Cases

**Use Case 1: Bar Entry (Age Only)**

```
User: "I need to show I'm over 21"
App: Displays QR with only age_over_21 = true
Bar: Scans QR, sees "✓ Over 21" (no name, no address)
```

**Use Case 2: Facebook Marketplace (Name + Address)**

```
User: "Show full name and address"
App: Displays QR with name="John Doe", address="123 Main St"
Buyer: Scans QR, verifies seller's identity and location
```

**Use Case 3: Voice Command**

```
User: "AURA show my age, address, and height"
App: Parses command → [AGE, ADDRESS_FULL, HEIGHT]
App: Generates QR with those 3 attributes
```

### 2.3 Implementation Requirements

**Attribute Storage:**

- Each attribute stored as separate `AttributeVC`
- Encrypted on-chain (user's key)
- Can be selectively disclosed via ZK proofs

**Disclosure Flow:**

```
Verifier                      User                      Blockchain
   |                           |                              |
   | 1. Request attributes     |                              |
   |-------------------------->|                              |
   |  (age, address)           |                              |
   |                           |                              |
   |                           | 2. Check policy              |
   |                           |----------------------------->|
   |                           |                              |
   |                           | 3. Policy: "ask user"        |
   |                           |<-----------------------------|
   |                           |                              |
   |                           | 4. User approves             |
   |                           |                              |
   |                           | 5. Generate presentation     |
   |                           |----------------------------->|
   |                           |                              |
   |                           | 6. QR code with attributes   |
   |                           |<-----------------------------|
   |                           |                              |
   | 7. Scan QR                |                              |
   |<--------------------------|                              |
   |                           |                              |
   | 8. Verify presentation    |                              |
   |------------------------------------------------->|       |
   |                           |                              |
   | 9. Verification result    |                              |
   |<------------------------------------------------|       |
   | (age=25, address=123 Main)|                              |
```

**Web Interface:**

```html
<div class="disclosure-control">
  <h3>Select what to show:</h3>
  <label><input type="checkbox" name="full_name"> Full Name</label>
  <label><input type="checkbox" name="age"> Age (exact)</label>
  <label><input type="checkbox" name="age_over_21"> Over 21 (yes/no only)</label>
  <label><input type="checkbox" name="address"> Full Address</label>
  <label><input type="checkbox" name="city_state"> City & State only</label>
  <label><input type="checkbox" name="height"> Height</label>
  <label><input type="checkbox" name="eye_color"> Eye Color</label>
  <label><input type="checkbox" name="scuba_cert"> Scuba Certification</label>

  <button onclick="generateQR()">Generate QR Code</button>
</div>
```

**Voice Command Parsing:**

```
"AURA show my age" → [ATTRIBUTE_TYPE_AGE]
"AURA show my age and address" → [ATTRIBUTE_TYPE_AGE, ATTRIBUTE_TYPE_ADDRESS_FULL]
"AURA show name address height" → [FULL_NAME, ADDRESS_FULL, HEIGHT]
"AURA show everything" → [ALL_ATTRIBUTES]
"AURA show only that I'm over 21" → [AGE_OVER_21] (ZK proof, not actual age)
```

### 2.4 Privacy Features

**Zero-Knowledge Proofs:**

For sensitive disclosures like "over 21" without revealing exact age:

```
Attribute: date_of_birth = "1990-05-15"
ZK Proof: PROVE(date_of_birth < 2004-11-13) WITHOUT revealing date_of_birth
Verifier sees: "✓ Over 21" (no actual birthdate)
```

**Encryption:**

- Attributes encrypted with user's private key
- Only decrypted when user explicitly discloses
- On-chain storage shows only encrypted bytes + hash

---

# Feature #3: Data Registry Module

## Summary

Allow verified users to store additional verified data beyond identity attributes: car registration, golf scores (geotagged/timestamped), photos, NFTs, and other documents that require trust verification with peers or authorities.

## Module Assignment

**NEW MODULE:** `dataregistry`

## Technical Design

### 3.1 Module Structure

```
chain/x/dataregistry/
├── keeper/
│   ├── keeper.go
│   ├── data_item.go
│   ├── query.go
│   └── msg_server.go
├── types/
│   ├── data_item.go
│   ├── params.go
│   ├── genesis.go
│   ├── keys.go
│   ├── errors.go
│   └── events.go
├── client/
│   └── cli/
│       ├── query.go
│       └── tx.go
└── module.go

proto/aura/dataregistry/v1beta1/
├── data_registry.proto
├── query.proto
└── tx.proto
```

### 3.2 Proto Definitions

**File:** `proto/aura/dataregistry/v1beta1/data_registry.proto`

```protobuf
syntax = "proto3";
package aura.dataregistry.v1beta1;

import "google/protobuf/timestamp.proto";
import "cosmos/base/query/v1beta1/pagination.proto";

option go_package = "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1;dataregistrypb";

// DataItemType categorizes stored data
enum DataItemType {
  DATA_ITEM_TYPE_UNSPECIFIED = 0;

  // Documents
  DATA_ITEM_TYPE_VEHICLE_REGISTRATION = 1;
  DATA_ITEM_TYPE_VEHICLE_INSURANCE = 2;
  DATA_ITEM_TYPE_PROPERTY_DEED = 3;
  DATA_ITEM_TYPE_LEASE_AGREEMENT = 4;
  DATA_ITEM_TYPE_CONTRACT = 5;
  DATA_ITEM_TYPE_RECEIPT = 6;
  DATA_ITEM_TYPE_WARRANTY = 7;

  // Media
  DATA_ITEM_TYPE_PHOTO = 10;
  DATA_ITEM_TYPE_VIDEO = 11;
  DATA_ITEM_TYPE_AUDIO = 12;
  DATA_ITEM_TYPE_DOCUMENT_PDF = 13;

  // Scores & Achievements
  DATA_ITEM_TYPE_GOLF_SCORE = 20;
  DATA_ITEM_TYPE_TEST_SCORE = 21;
  DATA_ITEM_TYPE_CERTIFICATION = 22;
  DATA_ITEM_TYPE_ACHIEVEMENT = 23;

  // Digital Assets
  DATA_ITEM_TYPE_NFT = 30;
  DATA_ITEM_TYPE_DIGITAL_ART = 31;
  DATA_ITEM_TYPE_MUSIC_LICENSE = 32;

  // Health
  DATA_ITEM_TYPE_VACCINATION_RECORD = 40;
  DATA_ITEM_TYPE_MEDICAL_RECORD = 41;
  DATA_ITEM_TYPE_PRESCRIPTION = 42;

  // Custom
  DATA_ITEM_TYPE_CUSTOM = 100;
}

// DataItemStatus represents lifecycle
enum DataItemStatus {
  DATA_ITEM_STATUS_UNSPECIFIED = 0;
  DATA_ITEM_STATUS_PENDING_VERIFICATION = 1;
  DATA_ITEM_STATUS_VERIFIED = 2;
  DATA_ITEM_STATUS_REJECTED = 3;
  DATA_ITEM_STATUS_EXPIRED = 4;
  DATA_ITEM_STATUS_REVOKED = 5;
}

// VerificationLevel indicates trust level
enum VerificationLevel {
  VERIFICATION_LEVEL_UNSPECIFIED = 0;
  VERIFICATION_LEVEL_SELF_ATTESTED = 1;     // User claims, not verified
  VERIFICATION_LEVEL_PEER_VERIFIED = 2;      // Verified by another user
  VERIFICATION_LEVEL_AI_VERIFIED = 3;        // Verified by AI agent
  VERIFICATION_LEVEL_AUTHORITY_VERIFIED = 4; // Verified by official authority
  VERIFICATION_LEVEL_BLOCKCHAIN_ANCHORED = 5; // Anchored with external proof
}

// DataItem represents a stored data item
message DataItem {
  string data_id = 1;                       // Unique identifier
  string owner_address = 2;                 // Owner's wallet
  DataItemType data_type = 3;               // Type of data
  string data_type_custom = 4;              // Custom type name
  DataItemStatus status = 5;                // Current status
  VerificationLevel verification_level = 6; // How verified

  // Content (stored off-chain or encrypted)
  bytes content_hash = 7;                   // SHA256 of content
  string storage_location = 8;              // IPFS CID, Arweave TX, or URL
  bool is_encrypted = 9;                    // Is content encrypted
  bytes encryption_key_hash = 10;           // Hash of encryption key (if encrypted)

  // Metadata
  string title = 11;                        // Human-readable title
  string description = 12;                  // Description
  map<string, string> metadata = 13;        // Extensible metadata
  repeated string tags = 14;                // Search tags

  // Temporal/Geospatial
  google.protobuf.Timestamp created_at = 15;
  google.protobuf.Timestamp verified_at = 16;
  google.protobuf.Timestamp expires_at = 17; // 0 = never
  GeoLocation geo_location = 18;            // Where data was created

  // Verification
  repeated Verification verifications = 19; // Who verified it
  string verified_by = 20;                  // Primary verifier address

  // Access control
  AccessPolicy access_policy = 21;          // Who can see it

  // Provenance
  string previous_data_id = 22;             // For versioning/updates
  uint64 version = 23;                      // Version number
}

// GeoLocation for geotagged data
message GeoLocation {
  double latitude = 1;
  double longitude = 2;
  double altitude = 3;                      // Optional
  double accuracy_meters = 4;               // GPS accuracy
  google.protobuf.Timestamp timestamp = 5;  // When location was captured
  string location_name = 6;                 // "Pebble Beach Golf Course"
}

// Verification record
message Verification {
  string verifier_address = 1;              // Who verified
  VerificationLevel level = 2;              // What level
  google.protobuf.Timestamp verified_at = 3;
  string verification_method = 4;           // "AI OCR", "Manual review", etc.
  uint64 confidence_score = 5;              // 0-100
  string notes = 6;                         // Verifier notes
  bytes proof = 7;                          // Cryptographic proof (optional)
}

// AccessPolicy controls who can view data
message AccessPolicy {
  AccessMode mode = 1;
  repeated string allowed_addresses = 2;    // Whitelist
  repeated string denied_addresses = 3;     // Blacklist
  bool require_verified_identity = 4;       // Must have AURA VC
  uint64 min_confidence_score = 5;          // Minimum CS required
}

enum AccessMode {
  ACCESS_MODE_PRIVATE = 0;                  // Owner only
  ACCESS_MODE_WHITELIST = 1;                // Allowed addresses only
  ACCESS_MODE_PUBLIC = 2;                   // Anyone can view
  ACCESS_MODE_VERIFIED_USERS = 3;           // Any AURA-verified user
}

// SpecializedDataItems for specific types

// Golf score with geotagging and timestamping
message GolfScoreData {
  string course_name = 1;
  GeoLocation location = 2;
  google.protobuf.Timestamp played_at = 3;
  uint32 total_score = 4;
  repeated uint32 hole_scores = 5;          // 18 holes
  uint32 handicap = 6;
  string scorecard_image_cid = 7;           // IPFS CID of photo
  repeated string playing_partners = 8;     // Witness addresses
}

// Vehicle registration data
message VehicleRegistrationData {
  string vin = 1;
  string make = 2;
  string model = 3;
  uint32 year = 4;
  string license_plate = 5;
  string state = 6;
  google.protobuf.Timestamp registered_at = 7;
  google.protobuf.Timestamp expires_at = 8;
  string registration_image_cid = 9;        // IPFS CID
}

// Photo with metadata
message PhotoData {
  string photo_cid = 1;                     // IPFS CID
  GeoLocation location = 2;
  google.protobuf.Timestamp taken_at = 3;
  string camera_model = 4;
  string description = 5;
  repeated string people_tagged = 6;        // AURA addresses
  uint32 width = 7;
  uint32 height = 8;
}

// Transaction messages

message MsgStoreDataItem {
  string creator = 1;
  DataItemType data_type = 2;
  string title = 3;
  string description = 4;
  bytes content_hash = 5;
  string storage_location = 6;              // IPFS CID
  bool is_encrypted = 7;
  GeoLocation geo_location = 8;             // Optional
  map<string, string> metadata = 9;
  AccessPolicy access_policy = 10;
  bytes specialized_data = 11;              // Encoded GolfScoreData, etc.
}

message MsgStoreDataItemResponse {
  string data_id = 1;
}

message MsgUpdateDataItem {
  string creator = 1;
  string data_id = 2;
  string title = 3;
  string description = 4;
  map<string, string> metadata = 5;
  AccessPolicy access_policy = 6;
}

message MsgUpdateDataItemResponse {}

message MsgDeleteDataItem {
  string creator = 1;
  string data_id = 2;
}

message MsgDeleteDataItemResponse {}

message MsgVerifyDataItem {
  string verifier = 1;
  string data_id = 2;
  VerificationLevel level = 3;
  uint64 confidence_score = 4;
  string notes = 5;
}

message MsgVerifyDataItemResponse {}

message MsgRevokeDataItem {
  string authority = 1;
  string data_id = 2;
  string reason = 3;
}

message MsgRevokeDataItemResponse {}

// Query service

service Query {
  // Get data item by ID
  rpc DataItem(QueryDataItemRequest) returns (QueryDataItemResponse);

  // List user's data items
  rpc UserDataItems(QueryUserDataItemsRequest) returns (QueryUserDataItemsResponse);

  // Search data items
  rpc SearchDataItems(QuerySearchDataItemsRequest) returns (QuerySearchDataItemsResponse);

  // Get verifications for a data item
  rpc DataItemVerifications(QueryDataItemVerificationsRequest)
      returns (QueryDataItemVerificationsResponse);
}

message QueryDataItemRequest {
  string data_id = 1;
}

message QueryDataItemResponse {
  DataItem data_item = 1;
}

message QueryUserDataItemsRequest {
  string owner_address = 1;
  DataItemType type_filter = 2;             // Optional
  cosmos.base.query.v1beta1.PageRequest pagination = 3;
}

message QueryUserDataItemsResponse {
  repeated DataItem data_items = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}

message QuerySearchDataItemsRequest {
  string search_query = 1;
  repeated string tags = 2;
  DataItemType type_filter = 3;
  GeoLocation near_location = 4;            // Optional geo search
  double radius_km = 5;                     // Radius for geo search
  cosmos.base.query.v1beta1.PageRequest pagination = 6;
}

message QuerySearchDataItemsResponse {
  repeated DataItem data_items = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}

message QueryDataItemVerificationsRequest {
  string data_id = 1;
}

message QueryDataItemVerificationsResponse {
  repeated Verification verifications = 1;
}

// Params for module configuration
message Params {
  uint64 max_data_items_per_user = 1;       // Default: 1000
  uint64 max_storage_bytes = 2;             // Max size per item
  string storage_fee = 3;                   // Fee in uaura
  uint64 verification_reward = 4;           // Reward for verifiers
  repeated string authorized_verifiers = 5; // Official verifier addresses
}

// Genesis state
message GenesisState {
  Params params = 1;
  repeated DataItem data_items = 2;
  uint64 next_data_id = 3;
}
```

### 3.3 Use Cases

**Use Case 1: Car Registration for Facebook Marketplace**

```
Seller stores car registration:
- data_type: VEHICLE_REGISTRATION
- VIN, make, model, year
- Photo of registration (IPFS)
- Verified by AI (OCR reads registration)
- Access: Public

Buyer scans QR:
- Sees verified car registration
- Confirms seller owns the vehicle
- Trust established for transaction
```

**Use Case 2: Golfing Scores (Geotagged & Timestamped)**

```
Golfer completes round at Pebble Beach:
- data_type: GOLF_SCORE
- Score: 78
- Location: GPS coordinates of Pebble Beach
- Timestamp: 2025-11-13 14:30:00
- Photo of scorecard (IPFS)
- Verified by: Playing partners (peer verification)
- Access: Public (for leaderboards)

Later:
- Golfer proves handicap to tournament
- Historical scores are verifiable
- Location proves course difficulty
```

**Use Case 3: Medical Vaccination Record**

```
Patient stores vaccination record:
- data_type: VACCINATION_RECORD
- Vaccine: COVID-19 Booster
- Date: 2025-10-15
- Provider: Kaiser Permanente
- Encrypted: Yes (HIPAA compliance)
- Verified by: Healthcare provider (authority verification)
- Access: Private (share via QR when needed)

Travel:
- Airport requests vaccination proof
- Patient generates QR with decrypted record
- Airport verifies authenticity via blockchain
```

**Use Case 4: NFT Storage with Provenance**

```
Artist stores NFT:
- data_type: NFT
- Title: "Digital Sunset #42"
- Image: IPFS CID
- Metadata: Creator, creation date, edition
- Blockchain anchored: Yes (Ethereum TX hash)
- Access: Public (for viewing), Owner only (for transfer)

Collector verifies:
- NFT is authentic
- Provenance tracked on AURA
- Creator verified via AURA identity
```

### 3.4 Storage Architecture

**Hybrid Approach:**

1. **On-chain (AURA blockchain):**
   - DataItem metadata
   - Content hash
   - Access policy
   - Verifications
   - Geo/temporal data

2. **Off-chain (IPFS/Arweave):**
   - Actual content (photos, PDFs, videos)
   - Large files
   - Referenced by CID/TX hash

**Storage Flow:**

```
User                        AURA App                 IPFS                Blockchain
  |                              |                     |                       |
  | 1. Upload photo              |                     |                       |
  |----------------------------->|                     |                       |
  |                              |                     |                       |
  |                              | 2. Upload to IPFS   |                       |
  |                              |-------------------->|                       |
  |                              |                     |                       |
  |                              | 3. Receive CID      |                       |
  |                              |<--------------------|                       |
  |                              |                     |                       |
  |                              | 4. Hash content     |                       |
  |                              | SHA256(photo) = hash|                       |
  |                              |                     |                       |
  |                              | 5. MsgStoreDataItem |                       |
  |                              |     (hash, CID, metadata)                   |
  |                              |----------------------------------------->|  |
  |                              |                     |                       |
  |                              | 6. DataItem created |                       |
  |                              |<-----------------------------------------|  |
  |                              |     data_id = "DATA-001"                    |
  |                              |                     |                       |
  | 7. Success: DATA-001         |                     |                       |
  |<-----------------------------|                     |                       |


Retrieval:
  |                              |                     |                       |
  | 8. View photo (DATA-001)     |                     |                       |
  |----------------------------->|                     |                       |
  |                              |                     |                       |
  |                              | 9. Query blockchain |                       |
  |                              |----------------------------------------->|  |
  |                              |                     |                       |
  |                              | 10. Get CID + hash  |                       |
  |                              |<-----------------------------------------|  |
  |                              |                     |                       |
  |                              | 11. Fetch from IPFS |                       |
  |                              |-------------------->|                       |
  |                              |                     |                       |
  |                              | 12. Content + verify hash                  |
  |                              |<--------------------|                       |
  |                              |                     |                       |
  | 13. Display photo            |                     |                       |
  |<-----------------------------|                     |                       |
```

### 3.5 Module Parameters

```go
type Params struct {
    MaxDataItemsPerUser  uint64 // 1000
    MaxStorageBytes      uint64 // 100MB per item
    StorageFee           string // "100000uaura" per MB
    VerificationReward   uint64 // POI reward for verifying
    AuthorizedVerifiers  []string // Official verifier addresses
}
```

### 3.6 CLI Examples

```bash
# Store car registration
aurad tx dataregistry store \
  --type=vehicle-registration \
  --title="2024 Tesla Model 3 Registration" \
  --description="CA Registration" \
  --storage-location="ipfs://Qm..." \
  --metadata="vin:5YJ3E1EA1KF123456,make:Tesla,model:Model 3,year:2024" \
  --geo-location="37.7749,-122.4194" \
  --access-policy=public \
  --from alice

# Store golf score
aurad tx dataregistry store \
  --type=golf-score \
  --title="Pebble Beach Round - Nov 2025" \
  --specialized-data='{"course_name":"Pebble Beach","total_score":78,...}' \
  --geo-location="36.5697,-121.9489" \
  --storage-location="ipfs://Qm..." \
  --from bob

# Verify data item
aurad tx dataregistry verify \
  --data-id="DATA-001" \
  --level=ai-verified \
  --confidence=95 \
  --notes="AI verified VIN matches registration image" \
  --from ai-agent

# Query user's data
aurad query dataregistry user-items alice

# Search for golf scores near location
aurad query dataregistry search \
  --type=golf-score \
  --near-location="36.5697,-121.9489" \
  --radius-km=50
```

---

## Implementation Plan

### Phase 1: Foundation (Week 1-2)

**Feature #1: QR Verification API**
- [ ] Create `presentation.proto` file
- [ ] Implement `MsgCreatePresentation` handler
- [ ] Implement `QueryVerifyPresentation` handler
- [ ] Add anti-spoofing (nonce, signature, expiration)
- [ ] Write unit tests
- [ ] Integration tests with vcregistry

**Feature #2: Selective Disclosure (Part 1)**
- [ ] Create `attributes.proto` file
- [ ] Implement `AttributeVC` storage
- [ ] Implement `DisclosurePolicy` management
- [ ] Write basic CLI commands
- [ ] Unit tests

### Phase 2: Core Features (Week 3-4)

**Feature #2: Selective Disclosure (Part 2)**
- [ ] Implement presentation with selective attributes
- [ ] Add voice command parsing
- [ ] Integrate with Feature #1 (QR codes)
- [ ] Web interface for attribute selection
- [ ] ZK proof integration (basic)
- [ ] Integration tests

**Feature #3: Data Registry (Part 1)**
- [ ] Scaffold new `dataregistry` module
- [ ] Create all proto files
- [ ] Implement basic CRUD (store, update, delete)
- [ ] Implement keeper logic
- [ ] Unit tests for keeper

### Phase 3: Advanced Features (Week 5-6)

**Feature #3: Data Registry (Part 2)**
- [ ] IPFS integration for storage
- [ ] Geolocation indexing
- [ ] Search functionality
- [ ] Verification system
- [ ] Access control enforcement
- [ ] Specialized data types (golf, vehicle, photo)
- [ ] Integration tests

### Phase 4: Integration & Testing (Week 7)

- [ ] End-to-end testing all features
- [ ] Performance testing
- [ ] Security audit
- [ ] Documentation
- [ ] Example frontend apps
- [ ] Deployment guides

### Phase 5: Deployment (Week 8)

- [ ] Testnet deployment
- [ ] Mainnet upgrade proposal
- [ ] Community testing
- [ ] Bug fixes
- [ ] Final documentation

---

## Security Considerations

### Feature #1: QR Verification
- ✅ Time-limited QR codes prevent screenshot attacks
- ✅ Nonce prevents replay attacks
- ✅ Signature prevents tampering
- ✅ On-chain verification prevents fake apps

### Feature #2: Selective Disclosure
- ✅ User controls what's revealed
- ✅ Encrypted attribute storage
- ✅ ZK proofs for sensitive data
- ✅ Audit trail of disclosures

### Feature #3: Data Registry
- ✅ Off-chain storage prevents blockchain bloat
- ✅ Content hashing ensures integrity
- ✅ Access control protects privacy
- ✅ Verification levels prevent fraud
- ✅ Encryption for sensitive data

---

## Success Metrics

**Feature #1:**
- 10,000+ QR verifications in first month
- <500ms verification latency
- 99.9%+ uptime
- Zero successful spoofing attacks

**Feature #2:**
- 50% of users use selective disclosure
- Average 3.2 attributes disclosed per verification
- 90%+ user satisfaction with privacy controls

**Feature #3:**
- 100,000+ data items stored in first quarter
- 50+ different data types in use
- 80%+ verification rate for stored data
- <1% fraudulent data items

---

**Document Status:** Complete & Ready for Implementation
**Next Step:** Launch implementation agents in parallel
