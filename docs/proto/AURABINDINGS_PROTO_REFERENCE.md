# AURA Bindings Proto Reference

**Version:** v1beta1
**Package:** `aura.aurabindings.v1beta1`
**Go Package:** `github.com/aequitas/aura/proto/aura/aurabindings/v1beta1`

---

## Table of Contents

1. [Query Messages](#query-messages)
2. [Message (Tx) Messages](#message-tx-messages)
3. [Module Parameters](#module-parameters)
4. [Usage Examples](#usage-examples)

---

## Query Messages

All queries are accessed through the unified `AuraQueryRequest` message using the oneof pattern.

### VCRegistry Queries

#### 1. QueryVCStatus
Check the status of a specific verifiable credential.

**Request:**
```protobuf
message QueryVCStatus {
  string vc_id = 1;  // VC identifier to query
}
```

**Response:**
```protobuf
message QueryVCStatusResponse {
  aura.vcregistry.v1beta1.VCStatus status = 1;  // Current status
  bool valid = 2;  // True if active and not expired
  google.protobuf.Timestamp expires_at = 3;  // Expiration timestamp
  bool revoked = 4;  // True if revoked
  bytes merkle_proof = 5;  // Merkle proof for trustless verification
}
```

**Use Case:** Check if a user's credential is still valid before granting access.

---

#### 2. QueryUserVCs
Query all VCs for a user with optional filters.

**Request:**
```protobuf
message QueryUserVCs {
  string holder_address = 1;  // User's bech32 address
  aura.vcregistry.v1beta1.VCStatus status_filter = 2;  // Optional: filter by status
  aura.vcregistry.v1beta1.VCType type_filter = 3;  // Optional: filter by type
}
```

**Response:**
```protobuf
message QueryUserVCsResponse {
  repeated VCRecord vcs = 1;  // List of matching VCs
  uint64 total_count = 2;  // Total number of VCs (before pagination)
}

message VCRecord {
  string vc_id = 1;
  aura.vcregistry.v1beta1.VCType vc_type = 2;
  string vc_type_custom = 3;
  aura.vcregistry.v1beta1.VCStatus status = 4;
  google.protobuf.Timestamp issued_at = 5;
  google.protobuf.Timestamp expires_at = 6;
}
```

**Use Case:** List all professional credentials a user holds for a DAO membership check.

---

#### 3. QueryResolveDID
Resolve a DID to its document and associated VCs.

**Request:**
```protobuf
message QueryResolveDID {
  string did = 1;  // DID to resolve (e.g., "did:aura:mainnet:...")
}
```

**Response:**
```protobuf
message QueryResolveDIDResponse {
  bool exists = 1;  // True if DID exists
  string controller = 2;  // Controller address
  repeated VCRecord active_vcs = 3;  // Active VCs for this DID
  google.protobuf.Timestamp created = 4;
  google.protobuf.Timestamp updated = 5;
}
```

**Use Case:** Verify a DID-based identity and retrieve all associated credentials.

---

#### 4. QueryValidateMintEligibility
Check if a user can mint a specific VC type.

**Request:**
```protobuf
message QueryValidateMintEligibility {
  string holder_address = 1;  // User address to check
  aura.vcregistry.v1beta1.VCType vc_type = 2;  // VC type to mint
  string vc_type_custom = 3;  // Custom type name (if VC_TYPE_CUSTOM)
}
```

**Response:**
```protobuf
message QueryValidateMintEligibilityResponse {
  bool eligible = 1;  // True if user can mint this VC
  repeated string missing_requirements = 2;  // List of missing requirements
  uint64 current_cs = 3;  // User's current confidence score
  uint64 required_cs = 4;  // Required confidence score for this VC
  repeated string completed_ir_ids = 5;  // IRs user has completed
  repeated string required_ir_ids = 6;  // IRs required for this VC
}
```

**Use Case:** Show users what they need to complete before they can mint a credential.

---

#### 5. QueryCheckRevocation
Check if a VC is revoked.

**Request:**
```protobuf
message QueryCheckRevocation {
  string vc_id = 1;  // VC to check
}
```

**Response:**
```protobuf
message QueryCheckRevocationResponse {
  bool revoked = 1;  // True if revoked
  google.protobuf.Timestamp revoked_at = 2;  // When revoked
  string reason = 3;  // Revocation reason
  bytes merkle_proof = 4;  // Merkle proof
}
```

**Use Case:** Verify a credential hasn't been revoked before accepting it.

---

#### 6. QueryGetDisclosurePolicy
Get disclosure policy for a user.

**Request:**
```protobuf
message QueryGetDisclosurePolicy {
  string holder_address = 1;  // User address
}
```

**Response:**
```protobuf
message QueryGetDisclosurePolicyResponse {
  bool exists = 1;  // True if policy exists
  bool auto_approve_verified = 2;  // Auto-approve for verified requesters
  repeated string allowed_verifiers = 3;  // Whitelist of verifier addresses
  repeated string blocked_verifiers = 4;  // Blacklist of verifier addresses
}
```

**Use Case:** Check if a user will auto-approve disclosure requests from your contract.

---

#### 7. QueryListPendingDisclosures
List pending disclosure requests for a user.

**Request:**
```protobuf
message QueryListPendingDisclosures {
  string holder_address = 1;  // User address
}
```

**Response:**
```protobuf
message QueryListPendingDisclosuresResponse {
  repeated DisclosureRequest requests = 1;  // Pending requests
}

message DisclosureRequest {
  string request_id = 1;
  string verifier_address = 2;
  repeated string requested_attributes = 3;
  string purpose = 4;
  google.protobuf.Timestamp requested_at = 5;
}
```

**Use Case:** Show users pending disclosure requests they need to approve.

---

### Compliance Queries

#### 8. QueryKYCStatus
Query KYC verification status for a user.

**Request:**
```protobuf
message QueryKYCStatus {
  string address = 1;  // User address to check
}
```

**Response:**
```protobuf
message QueryKYCStatusResponse {
  aura.compliance.v1beta1.KYCLevel kyc_level = 1;  // Current KYC level
  bool verified = 2;  // True if any KYC level achieved
  google.protobuf.Timestamp verified_at = 3;  // When verified
  google.protobuf.Timestamp expires_at = 4;  // When expires
  string provider = 5;  // KYC provider name
  string jurisdiction = 6;  // User's jurisdiction
}
```

**Use Case:** Verify user has required KYC level before allowing large transactions.

---

#### 9. QuerySanctionsCheck
Check if user is on sanctions lists.

**Request:**
```protobuf
message QuerySanctionsCheck {
  string address = 1;  // User address to check
  bool force_refresh = 2;  // Force new check instead of using cache
}
```

**Response:**
```protobuf
message QuerySanctionsCheckResponse {
  aura.compliance.v1beta1.SanctionsStatus status = 1;  // Sanctions status
  bool clear = 2;  // True if clear (not on any list)
  repeated string matched_lists = 3;  // List names with matches
  google.protobuf.Timestamp screened_at = 4;  // When screened
  bool requires_manual_review = 5;  // True if manual review needed
}
```

**Use Case:** Screen users against OFAC and other sanctions lists before transactions.

---

#### 10. QueryComplianceVerify
Verify user meets compliance requirements.

**Request:**
```protobuf
message QueryComplianceVerify {
  string address = 1;  // User address to check
  aura.compliance.v1beta1.KYCLevel required_kyc_level = 2;  // Required KYC level
  bool require_sanctions_clear = 3;  // Require sanctions clearance
}
```

**Response:**
```protobuf
message QueryComplianceVerifyResponse {
  bool compliant = 1;  // True if all requirements met
  bool kyc_sufficient = 2;  // True if KYC level sufficient
  bool sanctions_clear = 3;  // True if sanctions clear
  repeated string violations = 4;  // List of compliance violations
}
```

**Use Case:** One-shot compliance check before executing a sensitive operation.

---

#### 11. QueryGDPRStatus
Query GDPR consent status.

**Request:**
```protobuf
message QueryGDPRStatus {
  string address = 1;  // User address
  string consent_type = 2;  // Type of consent (e.g., "data_processing")
}
```

**Response:**
```protobuf
message QueryGDPRStatusResponse {
  bool consented = 1;  // True if consent given
  google.protobuf.Timestamp consent_given_at = 2;  // When consent given
  string consent_version = 3;  // Privacy policy version
  bool withdrawn = 4;  // True if consent withdrawn
}
```

**Use Case:** Verify user has consented to data processing before collecting information.

---

### Auth Queries

#### 12. QueryHasRole
Check if user has a specific role.

**Request:**
```protobuf
message QueryHasRole {
  string address = 1;  // User address
  string role = 2;  // Role to check (e.g., "admin", "verifier")
}
```

**Response:**
```protobuf
message QueryHasRoleResponse {
  bool has_role = 1;  // True if user has the role
  google.protobuf.Timestamp granted_at = 2;  // When role was granted
  google.protobuf.Timestamp expires_at = 3;  // When role expires (if applicable)
}
```

**Use Case:** Check if user has admin role before allowing privileged operations.

---

#### 13. QueryCheckPermission
Check if user has a specific permission.

**Request:**
```protobuf
message QueryCheckPermission {
  string address = 1;  // User address
  string permission = 2;  // Permission to check
}
```

**Response:**
```protobuf
message QueryCheckPermissionResponse {
  bool has_permission = 1;  // True if user has the permission
  repeated string granted_by_roles = 2;  // Roles granting this permission
}
```

**Use Case:** Fine-grained permission checking for contract operations.

---

#### 14. QueryGetRoleAssignments
Get all role assignments for a user.

**Request:**
```protobuf
message QueryGetRoleAssignments {
  string address = 1;  // User address
}
```

**Response:**
```protobuf
message QueryGetRoleAssignmentsResponse {
  repeated RoleAssignment roles = 1;  // All role assignments
}

message RoleAssignment {
  string role = 1;
  google.protobuf.Timestamp granted_at = 2;
  google.protobuf.Timestamp expires_at = 3;
  string granted_by = 4;
}
```

**Use Case:** Display all roles a user has for profile or admin interfaces.

---

#### 15. QuerySessionStatus
Query active session status.

**Request:**
```protobuf
message QuerySessionStatus {
  string session_id = 1;  // Session identifier
}
```

**Response:**
```protobuf
message QuerySessionStatusResponse {
  bool active = 1;  // True if session is active
  string owner_address = 2;  // Session owner
  google.protobuf.Timestamp created_at = 3;
  google.protobuf.Timestamp expires_at = 4;
  bool biometric_verified = 5;  // True if biometric auth used
}
```

**Use Case:** Verify session is still active before allowing session-based operations.

---

### ConfidenceScore Queries

#### 16. QueryUserScore
Query user's confidence score.

**Request:**
```protobuf
message QueryUserScore {
  string wallet_address = 1;  // User address
}
```

**Response:**
```protobuf
message QueryUserScoreResponse {
  uint64 total_score = 1;  // Total confidence score
  bool verified = 2;  // True if score >= 10,000
  bool has_anchor = 3;  // True if IR-000 completed
  uint64 verification_achieved_height = 4;  // Block when verified
  google.protobuf.Timestamp verification_achieved_at = 5;
}
```

**Use Case:** Check if user has minimum confidence score for DAO membership.

---

#### 17. QueryHasCompletedIR
Check if user completed a specific IR.

**Request:**
```protobuf
message QueryHasCompletedIR {
  string wallet_address = 1;  // User address
  string ir_id = 2;  // IR to check (e.g., "IR-102")
}
```

**Response:**
```protobuf
message QueryHasCompletedIRResponse {
  bool completed = 1;  // True if IR completed
  google.protobuf.Timestamp completed_at = 2;  // When completed
  uint64 final_score = 3;  // Score awarded for this IR
  string assistant_address = 4;  // AI assistant who verified
}
```

**Use Case:** Require specific IR completion as prerequisite for VC minting.

---

#### 18. QueryArenaScore
Query score in a specific arena.

**Request:**
```protobuf
message QueryArenaScore {
  string wallet_address = 1;  // User address
  string arena = 2;  // Arena name (e.g., "Biometric", "Financial")
}
```

**Response:**
```protobuf
message QueryArenaScoreResponse {
  uint64 total_score = 1;  // Total score in this arena
  uint32 ir_count = 2;  // Number of IRs completed in arena
  bool focus_bonus_active = 3;  // True if >= 5000 CS in arena
}
```

**Use Case:** Award bonuses or access based on specialization in an arena.

---

#### 19. QueryAnchorInfo
Query anchor IR-000 completion info.

**Request:**
```protobuf
message QueryAnchorInfo {
  string wallet_address = 1;  // User address
}
```

**Response:**
```protobuf
message QueryAnchorInfoResponse {
  bool completed = 1;  // True if IR-000 completed
  google.protobuf.Timestamp completed_at = 2;  // When completed
  uint64 block_height = 3;  // Block height
  bytes verifier_plugin_hash = 4;  // Verifier plugin hash
}
```

**Use Case:** Verify user has completed basic identity anchor IR.

---

### DataRegistry Queries

#### 20. QueryGetDataItem
Query a specific data item.

**Request:**
```protobuf
message QueryGetDataItem {
  string data_id = 1;  // Data item identifier
}
```

**Response:**
```protobuf
message QueryGetDataItemResponse {
  bool exists = 1;  // True if data item exists
  string owner = 2;  // Owner address
  string data_type = 3;  // Type of data
  string status = 4;  // Status (active, archived, etc.)
  google.protobuf.Timestamp created_at = 5;
  bytes metadata_hash = 6;  // Hash of metadata
}
```

**Use Case:** Verify data item exists before purchasing or accessing.

---

#### 21. QueryCheckDataAccess
Check if requester can access data.

**Request:**
```protobuf
message QueryCheckDataAccess {
  string data_id = 1;  // Data item identifier
  string requester = 2;  // Address requesting access
}
```

**Response:**
```protobuf
message QueryCheckDataAccessResponse {
  bool has_access = 1;  // True if requester can access
  string access_level = 2;  // Access level granted (read, write, admin)
  google.protobuf.Timestamp granted_at = 3;
  google.protobuf.Timestamp expires_at = 4;
}
```

**Use Case:** Check access permissions before serving data to a user.

---

#### 22. QueryListUserDataItems
List all data items owned by user.

**Request:**
```protobuf
message QueryListUserDataItems {
  string owner = 1;  // Owner address
  string type_filter = 2;  // Optional: filter by type
  string status_filter = 3;  // Optional: filter by status
}
```

**Response:**
```protobuf
message QueryListUserDataItemsResponse {
  repeated DataItemSummary items = 1;  // List of data items
  uint64 total_count = 2;  // Total count
}

message DataItemSummary {
  string data_id = 1;
  string data_type = 2;
  string status = 3;
  google.protobuf.Timestamp created_at = 4;
}
```

**Use Case:** Display all data items a user owns in a marketplace interface.

---

### EconomicSecurity Queries

#### 23. QueryCheckSpendingLimit
Check if transaction is within spending limits.

**Request:**
```protobuf
message QueryCheckSpendingLimit {
  string address = 1;  // User address
  string amount = 2;  // Amount to check (as string to avoid overflow)
  string denom = 3;  // Denomination
}
```

**Response:**
```protobuf
message QueryCheckSpendingLimitResponse {
  bool within_limit = 1;  // True if within limits
  string daily_limit = 2;  // Daily spending limit
  string daily_spent = 3;  // Already spent today
  string daily_remaining = 4;  // Remaining daily limit
  string transaction_limit = 5;  // Per-transaction limit
}
```

**Use Case:** Enforce spending limits in a wallet or DeFi protocol.

---

#### 24. QueryIsWhaleTransaction
Check if transaction is considered a whale transaction.

**Request:**
```protobuf
message QueryIsWhaleTransaction {
  string address = 1;  // User address
  string amount = 2;  // Transaction amount
}
```

**Response:**
```protobuf
message QueryIsWhaleTransactionResponse {
  bool is_whale = 1;  // True if whale transaction
  string threshold = 2;  // Whale threshold amount
  bool requires_special_handling = 3;  // True if special handling needed
}
```

**Use Case:** Trigger additional verification or compliance for large transactions.

---

#### 25. QueryVestingSchedule
Query vesting schedule for user.

**Request:**
```protobuf
message QueryVestingSchedule {
  string address = 1;  // User address
}
```

**Response:**
```protobuf
message QueryVestingScheduleResponse {
  bool has_vesting = 1;  // True if user has vesting
  string total_vesting = 2;  // Total vesting amount
  string vested_amount = 3;  // Amount already vested
  string remaining_amount = 4;  // Amount still vesting
  google.protobuf.Timestamp next_vesting_date = 5;  // Next vesting event
}
```

**Use Case:** Display vesting schedule in a token dashboard.

---

## Message (Tx) Messages

All state-changing messages are accessed through the unified `AuraMsgRequest` message.

### VCRegistry Messages

#### 1. MsgRequestDisclosure
Request attribute disclosure from a credential holder.

**Request:**
```protobuf
message MsgRequestDisclosure {
  string contract_address = 1;  // Contract making the request
  string holder_address = 2;  // VC holder to request from
  string verifier_address = 3;  // Verifier address (typically the contract)
  repeated string requested_attributes = 4;  // Attributes to disclose (e.g., ["age", "country"])
  string purpose = 5;  // Purpose of disclosure (for audit trail)
  uint64 expiry_blocks = 6;  // Number of blocks until request expires
}
```

**Response:**
```protobuf
message MsgRequestDisclosureResponse {
  string request_id = 1;  // Unique request identifier
  bool auto_approved = 2;  // True if auto-approved by policy
  google.protobuf.Timestamp created_at = 3;
  google.protobuf.Timestamp expires_at = 4;
}
```

**Use Case:** Request age verification for age-restricted DAO membership.

---

#### 2. MsgVerifyPresentation
Verify a credential presentation.

**Request:**
```protobuf
message MsgVerifyPresentation {
  string contract_address = 1;  // Contract verifying the presentation
  string presentation_id = 2;  // Presentation to verify
  string context = 3;  // Context for verification (e.g., "loan_application")
}
```

**Response:**
```protobuf
message MsgVerifyPresentationResponse {
  bool valid = 1;  // True if presentation is valid
  bool all_vcs_active = 2;  // True if all VCs in presentation are active
  bool signature_valid = 3;  // True if signature is valid
  repeated string vc_ids = 4;  // VCs included in presentation
  map<string, string> disclosed_attributes = 5;  // Disclosed attribute values
  string holder_did = 6;  // DID of presentation holder
}
```

**Use Case:** Verify credential presentation submitted during loan application.

---

#### 3. MsgCreatePresentation
Create a new credential presentation.

**Request:**
```protobuf
message MsgCreatePresentation {
  string contract_address = 1;  // Contract creating the presentation
  string holder_address = 2;  // VC holder
  repeated string vc_ids = 3;  // VCs to include in presentation
  string context = 4;  // Presentation context
  string nonce = 5;  // Nonce for replay protection
}
```

**Response:**
```protobuf
message MsgCreatePresentationResponse {
  string presentation_id = 1;  // Created presentation ID
  bytes qr_code_data = 2;  // QR code data for presentation
  google.protobuf.Timestamp created_at = 3;
  google.protobuf.Timestamp expires_at = 4;
}
```

**Use Case:** Create presentation on behalf of user (with permission).

---

### InclusionRoutines Messages

#### 4. MsgRecordIRCompletion
Record completion of an Inclusion Routine.

**Request:**
```protobuf
message MsgRecordIRCompletion {
  string contract_address = 1;  // Contract recording the completion
  string wallet_address = 2;  // User completing the IR
  string ir_id = 3;  // IR identifier (e.g., "IR-102")
  bytes proof_hash = 4;  // SHA256 hash of proof data
  string metadata = 5;  // Optional metadata (JSON)
}
```

**Response:**
```protobuf
message MsgRecordIRCompletionResponse {
  bool success = 1;  // True if completion recorded
  uint64 base_score = 2;  // Base score for this IR
  uint64 final_score = 3;  // Final score after multipliers
  float velocity_bonus = 4;  // Velocity multiplier applied
  float arena_bonus = 5;  // Arena multiplier applied
  float jackpot_bonus = 6;  // Jackpot multiplier applied
  uint64 new_total_score = 7;  // User's new total confidence score
  bool verification_achieved = 8;  // True if user reached 10,000 CS
  google.protobuf.Timestamp completed_at = 9;
}
```

**Use Case:** Game contract records IR completion when user achieves in-game milestone.

**Authorization:** Only whitelisted contracts can call this (via `authorized_ir_contracts` param).

---

### ContractRegistry Messages

#### 5. MsgRegisterContract
Register a smart contract in the registry.

**Request:**
```protobuf
message MsgRegisterContract {
  string contract_address = 1;  // Contract to register
  string creator_address = 2;  // Contract creator/admin
  ContractMetadata metadata = 3;  // Contract metadata
  SecurityPolicy security_policy = 4;  // Security settings
  ComplianceRequirements compliance = 5;  // Compliance requirements
}

message ContractMetadata {
  string name = 1;  // Contract name
  string description = 2;  // Description
  string version = 3;  // Version string
  repeated string tags = 4;  // Searchable tags
  string source_code_url = 5;  // Source code repository URL
  string audit_report_url = 6;  // Audit report URL
  string license = 7;  // License identifier (e.g., "MIT", "Apache-2.0")
}

message SecurityPolicy {
  uint64 max_gas_per_execution = 1;  // Max gas per execution
  uint64 max_executions_per_block = 2;  // Max executions per block
  uint64 rate_limit_per_user = 3;  // Max executions per user per hour
  bool require_vc = 4;  // Require users to have VCs
  repeated string required_vc_types = 5;  // Required VC types
  uint64 min_confidence_score = 6;  // Minimum CS required to interact
  repeated string blacklisted_addresses = 7;  // Blacklisted addresses
  repeated string whitelisted_addresses = 8;  // Whitelisted addresses
  bool whitelist_mode = 9;  // True to enable whitelist mode
}

message ComplianceRequirements {
  bool require_kyc = 1;  // Require KYC verification
  string min_kyc_level = 2;  // Minimum KYC level
  bool require_sanctions_clear = 3;  // Require sanctions clearance
  bool require_gdpr_consent = 4;  // Require GDPR consent
  repeated string restricted_jurisdictions = 5;  // Blocked jurisdictions
  bool pep_allowed = 6;  // Allow Politically Exposed Persons
}
```

**Response:**
```protobuf
message MsgRegisterContractResponse {
  bool success = 1;  // True if registration successful
  google.protobuf.Timestamp registered_at = 2;
  string contract_id = 3;  // Contract registry ID
}
```

**Use Case:** New contract registers itself to enable custom bindings.

---

#### 6. MsgUpdateContractMetadata
Update contract metadata.

**Request:**
```protobuf
message MsgUpdateContractMetadata {
  string contract_address = 1;  // Contract to update
  string admin_address = 2;  // Contract admin (must be original creator)
  ContractMetadata metadata = 3;  // Updated metadata
}
```

**Response:**
```protobuf
message MsgUpdateContractMetadataResponse {
  bool success = 1;  // True if update successful
  google.protobuf.Timestamp updated_at = 2;
}
```

**Use Case:** Update contract description or audit report URL after new audit.

---

### Compliance Messages

#### 7. MsgReportSuspiciousActivity
Report suspicious activity to compliance module.

**Request:**
```protobuf
message MsgReportSuspiciousActivity {
  string contract_address = 1;  // Contract reporting the activity
  string subject_address = 2;  // Address of suspicious actor
  string activity_type = 3;  // Type (e.g., "unusual_pattern", "velocity_abuse")
  string description = 4;  // Human-readable description
  repeated string indicators = 5;  // Risk indicators triggered
  string severity = 6;  // Severity level (LOW, MEDIUM, HIGH, CRITICAL)
  string evidence = 7;  // Evidence data (JSON or IPFS hash)
  string transaction_hash = 8;  // Related transaction hash
}
```

**Response:**
```protobuf
message MsgReportSuspiciousActivityResponse {
  string activity_id = 1;  // Created activity report ID
  bool escalated = 2;  // True if escalated to compliance review
  google.protobuf.Timestamp reported_at = 3;
}
```

**Use Case:** DEX reports unusual trading patterns for compliance review.

---

### Monitoring Messages

#### 8. MsgReportContractEvent
Report a contract event to monitoring module.

**Request:**
```protobuf
message MsgReportContractEvent {
  string contract_address = 1;  // Contract reporting the event
  string event_type = 2;  // Event type (e.g., "user_action", "threshold_exceeded")
  string severity = 3;  // Severity (INFO, WARNING, ERROR, CRITICAL)
  string description = 4;  // Event description
  map<string, string> metadata = 5;  // Event metadata
  bool trigger_alert = 6;  // True to trigger monitoring alert
}
```

**Response:**
```protobuf
message MsgReportContractEventResponse {
  string event_id = 1;  // Created event ID
  bool alert_triggered = 2;  // True if alert was triggered
  google.protobuf.Timestamp recorded_at = 3;
}
```

**Use Case:** Contract reports critical event for operator alerting.

---

## Module Parameters

All parameters are contained in the `Params` message:

```protobuf
message Params {
  // Query Permissions (6)
  bool enable_vc_queries = 1;                    // Default: true
  bool enable_compliance_queries = 2;            // Default: true
  bool enable_auth_queries = 3;                  // Default: true
  bool enable_cs_queries = 4;                    // Default: true
  bool enable_data_queries = 5;                  // Default: true
  bool enable_economic_queries = 6;              // Default: true

  // Message Permissions (5)
  bool enable_vc_messages = 7;                   // Default: true
  bool enable_ir_messages = 8;                   // Default: true
  bool enable_registry_messages = 9;             // Default: true
  bool enable_compliance_messages = 10;          // Default: true
  bool enable_monitoring_messages = 11;          // Default: true

  // Rate Limiting (4)
  uint64 rate_limit_queries_per_block = 12;     // Default: 1000
  uint64 rate_limit_messages_per_block = 13;    // Default: 100
  uint64 rate_limit_queries_per_contract = 14;  // Default: 100
  uint64 rate_limit_messages_per_contract = 15; // Default: 10

  // Gas Limits (2)
  uint64 query_gas_limit = 16;                   // Default: 100000
  uint64 message_gas_limit = 17;                 // Default: 500000

  // Authorization (3)
  bool require_contract_registration = 18;       // Default: true
  bool allow_unaudited_contracts = 19;           // Default: false (mainnet), true (testnet)
  repeated string authorized_ir_contracts = 20;  // Default: empty (governance-managed)

  // Query-Specific Limits (3)
  uint64 max_vcs_per_query = 21;                // Default: 50
  uint64 max_data_items_per_query = 22;         // Default: 50
  uint64 max_roles_per_query = 23;              // Default: 20

  // Caching (3)
  bool enable_query_caching = 24;                // Default: true
  uint64 query_cache_ttl_seconds = 25;           // Default: 60
  uint64 max_cache_size_mb = 26;                 // Default: 100

  // Security (3)
  bool enable_circuit_breaker = 27;              // Default: true
  uint64 circuit_breaker_threshold = 28;         // Default: 50 (%)
  uint64 circuit_breaker_window_seconds = 29;    // Default: 300

  // Telemetry (2)
  bool enable_telemetry = 30;                    // Default: true
  uint64 telemetry_sample_rate = 31;             // Default: 100 (%)
}
```

---

## Usage Examples

### Example 1: VC-Gated DAO Membership Check

```rust
// Rust contract code (pseudocode)
use cosmwasm_std::{Deps, StdResult};
use aura_bindings::AuraQuerier;

fn check_membership(deps: Deps, user: String) -> StdResult<bool> {
    let querier = AuraQuerier::new(&deps.querier);

    // Check confidence score
    let score_response = querier.query_user_score(user.clone())?;
    if !score_response.verified {
        return Ok(false);
    }

    // Check for professional VC
    let vcs_response = querier.query_user_vcs(
        user.clone(),
        VCStatus::ACTIVE,
        VCType::PROFESSIONAL
    )?;

    if vcs_response.vcs.is_empty() {
        return Ok(false);
    }

    // Check KYC
    let kyc_response = querier.query_kyc_status(user)?;
    if kyc_response.kyc_level < KYCLevel::BASIC {
        return Ok(false);
    }

    Ok(true)
}
```

### Example 2: Compliance-Checked Swap

```rust
// Rust contract code (pseudocode)
fn execute_swap(
    deps: DepsMut,
    info: MessageInfo,
    amount: Uint128
) -> Result<Response, ContractError> {
    let querier = AuraQuerier::new(&deps.querier);

    // Check sanctions
    let sanctions = querier.query_sanctions_check(
        info.sender.to_string(),
        false  // use cache
    )?;

    if !sanctions.clear {
        return Err(ContractError::SanctionsViolation);
    }

    // Check spending limits
    let limits = querier.query_check_spending_limit(
        info.sender.to_string(),
        amount.to_string(),
        "uaura"
    )?;

    if !limits.within_limit {
        return Err(ContractError::SpendingLimitExceeded);
    }

    // Check if whale transaction
    let whale = querier.query_is_whale_transaction(
        info.sender.to_string(),
        amount.to_string()
    )?;

    if whale.is_whale {
        // Report for compliance monitoring
        let msg = AuraMsg::ReportSuspiciousActivity {
            contract_address: env.contract.address.to_string(),
            subject_address: info.sender.to_string(),
            activity_type: "large_swap".to_string(),
            description: format!("Whale swap: {}", amount),
            indicators: vec!["whale_threshold_exceeded".to_string()],
            severity: "MEDIUM".to_string(),
            evidence: "".to_string(),
            transaction_hash: "".to_string(),
        };

        // Execute swap and report
        return Ok(Response::new()
            .add_message(msg)
            .add_attribute("action", "whale_swap")
            .add_attribute("amount", amount));
    }

    // Execute normal swap
    Ok(Response::new().add_attribute("action", "swap"))
}
```

### Example 3: IR Completion Recording

```rust
// Rust contract code (pseudocode)
// Only authorized contracts can do this
fn complete_challenge(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    challenge_id: String
) -> Result<Response, ContractError> {
    // Verify challenge completed
    let challenge = CHALLENGES.load(deps.storage, challenge_id.clone())?;
    if !challenge.completed {
        return Err(ContractError::ChallengeNotComplete);
    }

    // Record IR completion
    let msg = AuraMsg::RecordIRCompletion {
        contract_address: env.contract.address.to_string(),
        wallet_address: info.sender.to_string(),
        ir_id: "IR-105".to_string(),  // Gaming IR
        proof_hash: challenge.proof_hash,
        metadata: challenge.metadata,
    };

    Ok(Response::new()
        .add_message(msg)
        .add_attribute("action", "ir_complete")
        .add_attribute("ir_id", "IR-105"))
}
```

---

## Notes

- All queries are **read-only** and do not modify state
- All messages are **state-changing** and require gas
- Rate limits apply **per block** and **per contract**
- Query caching is **enabled by default** with 60s TTL
- Circuit breaker activates on **50% error rate** over 5 minutes
- Contract registration is **required by default** (configurable)
- IR recording requires **whitelist authorization** via governance

---

## References

- Main Implementation Task: `SMART_CONTRACT_IMPLEMENTATION_TASKS.md` - Phase 3
- Proto Files: `/proto/aura/aurabindings/v1beta1/`
- Generated Go Code: `/proto/aura/aurabindings/v1beta1/*.pb.go`
