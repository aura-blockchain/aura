import { Coin, SigningStargateClient, StdFee, StargateClient } from '@cosmjs/stargate';
export { Coin } from '@cosmjs/stargate';
import { OfflineDirectSigner, EncodeObject } from '@cosmjs/proto-signing';
export { OfflineDirectSigner } from '@cosmjs/proto-signing';

/**
 * Authentication info
 */
interface AuthInfo {
    address: string;
    publicKey: string;
    accountNumber: number;
    sequence: number;
    authType: string;
}
/**
 * Account info
 */
interface AccountInfo {
    address: string;
    publicKey?: Uint8Array;
    accountNumber: number;
    sequence: number;
}
/**
 * Grant authorization
 */
interface GrantAuthorization {
    granter: string;
    grantee: string;
    authorization: any;
    expiration?: Date;
}
/**
 * Grant parameters
 */
interface GrantParams {
    granter: string;
    grantee: string;
    msgTypeUrl: string;
    expiration?: Date;
    spendLimit?: string;
}
/**
 * Revoke parameters
 */
interface RevokeParams {
    granter: string;
    grantee: string;
    msgTypeUrl: string;
}
/**
 * Execute parameters
 */
interface ExecuteParams {
    grantee: string;
    messages: any[];
}
/**
 * Auth module parameters
 */
interface AuthParams {
    maxMemoCharacters: number;
    txSigLimit: number;
    txSizeCostPerByte: number;
    sigVerifyCostED25519: number;
    sigVerifyCostSecp256k1: number;
}

/**
 * Bridge transfer status enum
 */
declare enum BridgeTransferStatus {
    PENDING = 0,
    COMPLETED = 1,
    FAILED = 2,
    REFUNDED = 3
}
/**
 * Bridge transfer information
 */
interface BridgeTransfer {
    id: string;
    sender: string;
    recipient: string;
    amount: Coin;
    sourceChain: string;
    targetChain: string;
    status: BridgeTransferStatus;
    createdAt: Date;
    completedAt?: Date;
    proof?: string;
    txHash?: string;
    error?: string;
}
/**
 * Parameters for initiating a bridge transfer
 */
interface InitiateBridgeParams {
    sender: string;
    recipient: string;
    amount: string;
    denom: string;
    targetChain: string;
    timeout?: number;
    memo?: string;
}
/**
 * Parameters for completing a bridge transfer
 */
interface CompleteBridgeParams {
    transferId: string;
    proof: string;
    height: number;
    signatures: string[];
}
/**
 * Bridge parameters
 */
interface BridgeParams {
    minTransferAmount: string;
    maxTransferAmount: string;
    supportedChains: string[];
    bridgeFee: string;
    confirmationsRequired: number;
    timeout: number;
    enabled: boolean;
}
/**
 * Bridge security configuration
 */
interface BridgeSecurity {
    merkleRoot: string;
    validators: string[];
    requiredSignatures: number;
    lastUpdateHeight: number;
}
/**
 * Bridge statistics
 */
interface BridgeStats {
    totalTransfers: number;
    totalVolume: string;
    activeTransfers: number;
    completedTransfers: number;
    failedTransfers: number;
}

/**
 * Compliance status enum
 */
declare enum ComplianceStatusType {
    UNKNOWN = 0,
    PENDING = 1,
    APPROVED = 2,
    REJECTED = 3,
    REVOKED = 4
}
/**
 * KYC level enum
 */
declare enum KYCLevel {
    NONE = 0,
    BASIC = 1,
    INTERMEDIATE = 2,
    ADVANCED = 3
}
/**
 * Compliance status for an address
 */
interface ComplianceStatus {
    address: string;
    status: ComplianceStatusType;
    kycLevel: KYCLevel;
    verifiedAt?: Date;
    expiresAt?: Date;
    country?: string;
    riskScore: number;
    sanctioned: boolean;
    flags: string[];
}
/**
 * KYC submission parameters
 */
interface SubmitKYCParams {
    address: string;
    level: KYCLevel;
    personalInfo: {
        firstName: string;
        lastName: string;
        dateOfBirth: string;
        nationality: string;
    };
    documents: {
        type: string;
        id: string;
        hash: string;
    }[];
    residenceInfo?: {
        country: string;
        city: string;
        postalCode: string;
    };
}
/**
 * Transaction report
 */
interface TransactionReport {
    txHash: string;
    sender: string;
    recipient: string;
    amount: string;
    denom: string;
    timestamp: Date;
    complianceChecks: {
        sanctionCheck: boolean;
        amlCheck: boolean;
        riskAssessment: number;
    };
    flags: string[];
    approved: boolean;
}
/**
 * Compliance parameters
 */
interface ComplianceParams {
    kycRequired: boolean;
    sanctionsCheckEnabled: boolean;
    amlEnabled: boolean;
    maxRiskScore: number;
    reportingThreshold: string;
    autoReportEnabled: boolean;
}
/**
 * Sanction check result
 */
interface SanctionCheckResult {
    address: string;
    sanctioned: boolean;
    lists: string[];
    details?: string;
    checkedAt: Date;
}

/**
 * Confidence score information
 */
interface ConfidenceScore {
    address: string;
    score: number;
    level: number;
    completedRoutines: number;
    totalRoutines: number;
    lastUpdated: Date;
    nextUpdate: Date;
    factors: {
        transactionHistory: number;
        stakingParticipation: number;
        governanceParticipation: number;
        routineCompletion: number;
        networkReputation: number;
    };
}
/**
 * Score history entry
 */
interface ScoreHistoryEntry {
    address: string;
    score: number;
    level: number;
    timestamp: Date;
    reason: string;
    change: number;
}
/**
 * Confidence score rewards
 */
interface ConfidenceRewards {
    address: string;
    totalRewards: string;
    claimableRewards: string;
    lockedRewards: string;
    lastClaim: Date;
    nextClaim: Date;
    multiplier: number;
}
/**
 * Reward claim parameters
 */
interface ClaimRewardsParams {
    address: string;
    amount?: string;
}
/**
 * Complete inclusion routine parameters (for confidence score module)
 */
interface ConfidenceCompleteRoutineParams {
    address: string;
    routineId: string;
    proof: string;
    metadata?: Record<string, any>;
}
/**
 * Confidence score parameters
 */
interface ConfidenceScoreParams {
    minScore: number;
    maxScore: number;
    decayRate: number;
    rewardRate: string;
    levelThresholds: number[];
    updateInterval: number;
}
/**
 * Score calculation factors
 */
interface ScoreFactors {
    transactionWeight: number;
    stakingWeight: number;
    governanceWeight: number;
    routineWeight: number;
    reputationWeight: number;
}
/**
 * Slash event for confidence score
 */
interface ConfidenceSlashEvent {
    address: string;
    oldScore: number;
    newScore: number;
    reason: string;
    amount: number;
    timestamp: Date;
}

/**
 * Key pair interface
 */
interface KeyPair {
    publicKey: string;
    privateKey: string;
    algorithm: string;
    createdAt: Date;
    expiresAt?: Date;
}
/**
 * Encrypted data
 */
interface EncryptedData {
    ciphertext: string;
    nonce: string;
    algorithm: string;
    publicKey: string;
    timestamp: Date;
}
/**
 * Key rotation parameters
 */
interface KeyRotationParams {
    address: string;
    oldPublicKey: string;
    newPublicKey: string;
    signature: string;
    reason?: string;
}
/**
 * Encryption parameters
 */
interface EncryptParams {
    data: string;
    publicKey: string;
    algorithm?: string;
}
/**
 * Decryption parameters
 */
interface DecryptParams {
    encryptedData: EncryptedData;
    privateKey: string;
}
/**
 * Quantum-resistant key pair
 */
interface QuantumKeyPair extends KeyPair {
    quantumSafe: boolean;
    latticeParams?: {
        dimension: number;
        modulus: number;
    };
}
/**
 * Key info
 */
interface KeyInfo {
    address: string;
    publicKey: string;
    algorithm: string;
    createdAt: Date;
    rotatedAt?: Date;
    rotationCount: number;
    isQuantumSafe: boolean;
    status: KeyStatus;
}
/**
 * Key status enum
 */
declare enum KeyStatus {
    ACTIVE = 0,
    ROTATED = 1,
    REVOKED = 2,
    EXPIRED = 3
}
/**
 * Cryptography parameters
 */
interface CryptographyParams {
    defaultAlgorithm: string;
    keyRotationInterval: number;
    quantumSafeEnabled: boolean;
    maxKeyAge: number;
    supportedAlgorithms: string[];
}
/**
 * Secure enclave configuration
 */
interface SecureEnclaveConfig {
    enabled: boolean;
    provider: string;
    attestation: string;
    keyStorageLevel: string;
}
/**
 * Random number request
 */
interface RandomRequest {
    length: number;
    encoding?: 'hex' | 'base64' | 'bytes';
}
/**
 * Random number response
 */
interface RandomResponse {
    value: string;
    entropy: number;
    timestamp: Date;
}

/**
 * Data item status enum
 */
declare enum DataItemStatus {
    ACTIVE = 0,
    ARCHIVED = 1,
    DELETED = 2
}
/**
 * Data item
 */
interface DataItem {
    id: string;
    owner: string;
    data: string;
    metadata: Record<string, any>;
    hash: string;
    size: number;
    status: DataItemStatus;
    encrypted: boolean;
    createdAt: Date;
    updatedAt: Date;
    accessList?: string[];
    version: number;
}
/**
 * Register data parameters
 */
interface RegisterDataParams {
    owner: string;
    data: string;
    metadata?: Record<string, any>;
    encrypted?: boolean;
    accessList?: string[];
}
/**
 * Update data parameters
 */
interface UpdateDataParams {
    id: string;
    owner: string;
    data?: string;
    metadata?: Record<string, any>;
    accessList?: string[];
}
/**
 * Delete data parameters
 */
interface DeleteDataParams {
    id: string;
    owner: string;
    permanent?: boolean;
}
/**
 * Data query filters
 */
interface DataQueryFilters {
    owner?: string;
    status?: DataItemStatus;
    encrypted?: boolean;
    fromDate?: Date;
    toDate?: Date;
    tags?: string[];
    limit?: number;
    offset?: number;
}
/**
 * Data registry parameters
 */
interface DataRegistryParams {
    maxDataSize: number;
    storageFee: string;
    updateFee: string;
    deleteFee: string;
    encryptionRequired: boolean;
    versioning: boolean;
    maxVersions: number;
}
/**
 * Data access grant
 */
interface DataAccessGrant {
    dataId: string;
    grantee: string;
    grantor: string;
    permissions: string[];
    expiresAt?: Date;
    createdAt: Date;
}
/**
 * Data statistics
 */
interface DataStats {
    totalItems: number;
    activeItems: number;
    archivedItems: number;
    totalSize: number;
    ownerCount: number;
    avgItemSize: number;
}

/**
 * Dynamic fee structure
 */
interface DynamicFeeStructure {
    baseFee: string;
    priorityFee: string;
    congestionMultiplier: number;
    timeBasedMultiplier: number;
    lastUpdated: Date;
    nextUpdate: Date;
}
/**
 * MEV protection status
 */
interface MEVProtectionStatus {
    enabled: boolean;
    protectionLevel: MEVProtectionLevel;
    detectedAttempts: number;
    preventedAttempts: number;
    lastDetection?: Date;
    penalties: {
        address: string;
        amount: string;
        reason: string;
        timestamp: Date;
    }[];
}
/**
 * MEV protection level enum
 */
declare enum MEVProtectionLevel {
    NONE = 0,
    LOW = 1,
    MEDIUM = 2,
    HIGH = 3
}
/**
 * Whale protection configuration
 */
interface WhaleProtectionConfig {
    enabled: boolean;
    transactionLimit: string;
    dailyLimit: string;
    cooldownPeriod: number;
    exemptAddresses: string[];
}
/**
 * Whale transaction info
 */
interface WhaleTransaction {
    address: string;
    amount: string;
    timestamp: Date;
    delayed: boolean;
    releaseTime?: Date;
}
/**
 * Economic security parameters
 */
interface EconomicSecurityParams {
    dynamicFeesEnabled: boolean;
    mevProtectionEnabled: boolean;
    whaleProtectionEnabled: boolean;
    minBaseFee: string;
    maxBaseFee: string;
    congestionThreshold: number;
    whaleThreshold: string;
}
/**
 * Fee market metrics
 */
interface FeeMarketMetrics {
    currentBaseFee: string;
    averageFee: string;
    medianFee: string;
    congestionLevel: number;
    pendingTransactions: number;
    throughput: number;
    timestamp: Date;
}
/**
 * Economic attack detection
 */
interface EconomicAttackDetection {
    attackType: string;
    severity: number;
    affectedAddresses: string[];
    detectedAt: Date;
    mitigated: boolean;
    details: string;
}

/**
 * Identity change request status enum
 */
declare enum ChangeRequestStatus {
    PENDING = 0,
    APPROVED = 1,
    REJECTED = 2,
    EXPIRED = 3,
    CANCELLED = 4
}
/**
 * Identity change type enum
 */
declare enum ChangeType {
    NAME = 0,
    EMAIL = 1,
    PHONE = 2,
    ADDRESS = 3,
    DOCUMENTS = 4,
    OTHER = 5
}
/**
 * Identity change request
 */
interface IdentityChangeRequest {
    id: string;
    requester: string;
    changeType: ChangeType;
    oldValue: string;
    newValue: string;
    status: ChangeRequestStatus;
    reason: string;
    evidence: {
        type: string;
        hash: string;
        url?: string;
    }[];
    submittedAt: Date;
    reviewedAt?: Date;
    approvedBy?: string;
    expiresAt: Date;
    comments?: string;
}
/**
 * Submit change request parameters
 */
interface SubmitChangeRequestParams {
    requester: string;
    changeType: ChangeType;
    oldValue: string;
    newValue: string;
    reason: string;
    evidence?: {
        type: string;
        hash: string;
        url?: string;
    }[];
}
/**
 * Approve change request parameters
 */
interface ApproveChangeRequestParams {
    requestId: string;
    approver: string;
    approved: boolean;
    comments?: string;
}
/**
 * Identity change parameters
 */
interface IdentityChangeParams {
    reviewPeriod: number;
    requiredApprovals: number;
    maxPendingRequests: number;
    evidenceRequired: boolean;
    autoApprovalEnabled: boolean;
    approvers: string[];
}
/**
 * Identity history entry
 */
interface IdentityHistoryEntry {
    address: string;
    changeType: ChangeType;
    oldValue: string;
    newValue: string;
    timestamp: Date;
    requestId: string;
    approver: string;
}
/**
 * Identity verification status
 */
interface IdentityVerificationStatus {
    address: string;
    verified: boolean;
    verificationLevel: number;
    lastVerified?: Date;
    pendingChanges: number;
    totalChanges: number;
}

/**
 * Inclusion routine status enum
 */
declare enum RoutineStatus {
    PENDING = 0,
    IN_PROGRESS = 1,
    COMPLETED = 2,
    FAILED = 3,
    EXPIRED = 4
}
/**
 * Routine type enum
 */
declare enum RoutineType {
    VERIFICATION = 0,
    STAKING = 1,
    GOVERNANCE = 2,
    SOCIAL = 3,
    EDUCATIONAL = 4,
    CUSTOM = 5
}
/**
 * Inclusion routine
 */
interface InclusionRoutine {
    id: string;
    address: string;
    type: RoutineType;
    name: string;
    description: string;
    status: RoutineStatus;
    progress: number;
    requirements: RoutineRequirement[];
    rewards: {
        scoreIncrease: number;
        tokenReward: string;
        badges: string[];
    };
    createdAt: Date;
    startedAt?: Date;
    completedAt?: Date;
    expiresAt: Date;
    metadata: Record<string, any>;
}
/**
 * Routine requirement
 */
interface RoutineRequirement {
    id: string;
    description: string;
    type: string;
    completed: boolean;
    verifiable: boolean;
    proof?: string;
}
/**
 * Create routine parameters
 */
interface CreateRoutineParams {
    creator: string;
    type: RoutineType;
    name: string;
    description: string;
    requirements: {
        description: string;
        type: string;
        verifiable: boolean;
    }[];
    rewards: {
        scoreIncrease: number;
        tokenReward: string;
        badges?: string[];
    };
    duration: number;
}
/**
 * Complete routine parameters
 */
interface CompleteRoutineParams {
    routineId: string;
    address: string;
    proofs: {
        requirementId: string;
        proof: string;
    }[];
}
/**
 * Routine prerequisites
 */
interface RoutinePrerequisites {
    minScore: number;
    requiredRoutines: string[];
    kycLevel: number;
    minStake?: string;
}
/**
 * Inclusion routine parameters
 */
interface InclusionRoutineParams {
    creationEnabled: boolean;
    maxActiveRoutines: number;
    defaultDuration: number;
    verificationRequired: boolean;
    minScoreReward: number;
    maxScoreReward: number;
    rateLimitWindow: number;
    rateLimitCount: number;
}
/**
 * Routine statistics
 */
interface RoutineStats {
    totalRoutines: number;
    activeRoutines: number;
    completedRoutines: number;
    failedRoutines: number;
    totalParticipants: number;
    averageCompletionRate: number;
    totalRewardsDistributed: string;
}
/**
 * Routine leaderboard entry
 */
interface RoutineLeaderboardEntry {
    address: string;
    completedRoutines: number;
    totalScore: number;
    rank: number;
    badges: string[];
}

/**
 * Alert severity enum
 */
declare enum AlertSeverity {
    INFO = 0,
    WARNING = 1,
    ERROR = 2,
    CRITICAL = 3
}
/**
 * Alert status enum
 */
declare enum AlertStatus {
    ACTIVE = 0,
    ACKNOWLEDGED = 1,
    RESOLVED = 2
}
/**
 * System metrics
 */
interface SystemMetrics {
    blockHeight: number;
    blockTime: number;
    transactionsPerSecond: number;
    activeValidators: number;
    totalStaked: string;
    networkLoad: number;
    memoryUsage: number;
    cpuUsage: number;
    diskUsage: number;
    peerCount: number;
    timestamp: Date;
}
/**
 * Alert
 */
interface Alert {
    id: string;
    type: string;
    severity: AlertSeverity;
    status: AlertStatus;
    message: string;
    details: Record<string, any>;
    source: string;
    triggeredAt: Date;
    acknowledgedAt?: Date;
    resolvedAt?: Date;
    acknowledgedBy?: string;
}
/**
 * Monitoring configuration
 */
interface MonitoringConfig {
    enabled: boolean;
    metricsInterval: number;
    alertThresholds: {
        cpuUsage: number;
        memoryUsage: number;
        diskUsage: number;
        blockTime: number;
        transactionRate: number;
    };
    alertRecipients: string[];
    retentionPeriod: number;
}
/**
 * Performance metrics
 */
interface PerformanceMetrics {
    avgBlockTime: number;
    avgTransactionsPerBlock: number;
    avgGasPrice: string;
    networkThroughput: number;
    validatorUptime: number;
    missedBlocks: number;
    timestamp: Date;
}
/**
 * Node health
 */
interface NodeHealth {
    nodeId: string;
    status: string;
    uptime: number;
    version: string;
    syncStatus: {
        catching_up: boolean;
        latest_block_height: number;
        latest_block_time: Date;
    };
    peers: number;
    validatorInfo?: {
        address: string;
        votingPower: string;
        proposerPriority: number;
    };
}
/**
 * Anomaly detection result
 */
interface AnomalyDetection {
    detected: boolean;
    anomalyType: string;
    severity: number;
    confidence: number;
    affectedMetrics: string[];
    timestamp: Date;
    details: string;
}
/**
 * Monitoring parameters
 */
interface MonitoringParams {
    enabled: boolean;
    metricsRetention: number;
    alertingEnabled: boolean;
    anomalyDetectionEnabled: boolean;
    mlModelVersion: string;
    samplingRate: number;
}

/**
 * Network security status
 */
interface NetworkSecurityStatus {
    threatLevel: ThreatLevel;
    activePeers: number;
    blockedPeers: number;
    sybilAttackDetected: boolean;
    eclipseAttackDetected: boolean;
    ddosAttackDetected: boolean;
    lastThreatDetection?: Date;
    mitigationActive: boolean;
}
/**
 * Threat level enum
 */
declare enum ThreatLevel {
    NONE = 0,
    LOW = 1,
    MEDIUM = 2,
    HIGH = 3,
    CRITICAL = 4
}
/**
 * Threat report parameters
 */
interface ReportThreatParams {
    reporter: string;
    threatType: string;
    severity: ThreatLevel;
    description: string;
    evidence: {
        type: string;
        data: string;
    }[];
    affectedPeers?: string[];
}
/**
 * Sybil protection status
 */
interface SybilProtectionStatus {
    enabled: boolean;
    detectionAlgorithm: string;
    suspiciousNodes: number;
    blockedNodes: number;
    confidence: number;
    lastCheck: Date;
}
/**
 * Rate limits configuration
 */
interface RateLimitsConfig {
    enabled: boolean;
    requestsPerSecond: number;
    requestsPerMinute: number;
    requestsPerHour: number;
    burstSize: number;
    blockDuration: number;
    whitelistedAddresses: string[];
}
/**
 * Peer reputation
 */
interface PeerReputation {
    peerId: string;
    reputation: number;
    trustScore: number;
    uptime: number;
    successfulConnections: number;
    failedConnections: number;
    reportedThreats: number;
    maliciousActivity: boolean;
    lastSeen: Date;
}
/**
 * Network security parameters
 */
interface NetworkSecurityParams {
    sybilProtectionEnabled: boolean;
    rateLimitingEnabled: boolean;
    reputationSystemEnabled: boolean;
    minReputation: number;
    maxPeersPerIp: number;
    blocklistUpdateInterval: number;
    threatResponseEnabled: boolean;
}
/**
 * Gossip network status
 */
interface GossipNetworkStatus {
    connectedPeers: number;
    messageRate: number;
    averageLatency: number;
    droppedMessages: number;
    duplicateMessages: number;
    healthScore: number;
}
/**
 * DDoS mitigation status
 */
interface DDoSMitigationStatus {
    active: boolean;
    mitigationType: string;
    attackDuration: number;
    requestsBlocked: number;
    legitimateRequests: number;
    falsePositives: number;
    startedAt?: Date;
}
/**
 * Fork detection
 */
interface ForkDetection {
    detected: boolean;
    forkHeight: number;
    chainALength: number;
    chainBLength: number;
    detectedAt: Date;
    resolved: boolean;
    resolvedAt?: Date;
}
/**
 * Partition detection
 */
interface PartitionDetection {
    detected: boolean;
    affectedNodes: number;
    isolatedNodes: string[];
    detectedAt: Date;
    duration: number;
    resolved: boolean;
}

/**
 * Validation result
 */
interface ValidationResult {
    valid: boolean;
    errors: ValidationError[];
    warnings: ValidationWarning[];
    gasEstimate?: number;
    feeEstimate?: string;
    confidence: number;
}
/**
 * Validation error
 */
interface ValidationError {
    code: string;
    message: string;
    field?: string;
    severity: 'error' | 'critical';
}
/**
 * Validation warning
 */
interface ValidationWarning {
    code: string;
    message: string;
    field?: string;
    suggestion?: string;
}
/**
 * Transaction validation rules
 */
interface ValidationRules {
    minGasPrice: string;
    maxGasLimit: number;
    maxMemoLength: number;
    allowedMsgTypes: string[];
    requiredFields: string[];
    customRules: CustomRule[];
}
/**
 * Custom validation rule
 */
interface CustomRule {
    name: string;
    description: string;
    validator: string;
    enabled: boolean;
    severity: 'error' | 'warning';
}
/**
 * Prevalidation parameters
 */
interface PrevalidationParams {
    enabled: boolean;
    strictMode: boolean;
    cacheResults: boolean;
    cacheDuration: number;
    maxValidationTime: number;
    asyncValidation: boolean;
}
/**
 * Balance check result
 */
interface BalanceCheckResult {
    sufficient: boolean;
    required: string;
    available: string;
    deficit?: string;
}
/**
 * Signature validation result
 */
interface SignatureValidationResult {
    valid: boolean;
    publicKey: string;
    algorithm: string;
    message?: string;
}
/**
 * Transaction structure validation
 */
interface StructureValidationResult {
    valid: boolean;
    issues: {
        field: string;
        issue: string;
        fix?: string;
    }[];
}
/**
 * Compliance validation result
 */
interface ComplianceValidationResult {
    compliant: boolean;
    checks: {
        name: string;
        passed: boolean;
        details?: string;
    }[];
    requiredActions?: string[];
}

/**
 * Privacy level enum
 */
declare enum PrivacyLevel {
    PUBLIC = 0,
    PSEUDONYMOUS = 1,
    PRIVATE = 2,
    ANONYMOUS = 3
}
/**
 * Privacy settings
 */
interface PrivacySettings {
    address: string;
    defaultLevel: PrivacyLevel;
    transactionPrivacy: boolean;
    balancePrivacy: boolean;
    identityPrivacy: boolean;
    mixingEnabled: boolean;
    ringSignatureSize: number;
    stealthAddressEnabled: boolean;
}
/**
 * Private transaction parameters
 */
interface PrivateTransactionParams {
    sender: string;
    recipient: string;
    amount: string;
    denom: string;
    privacyLevel: PrivacyLevel;
    memo?: string;
}
/**
 * Confidential transaction
 */
interface ConfidentialTransaction {
    txHash: string;
    commitment: string;
    rangeProof: string;
    sender: string;
    recipient: string;
    timestamp: Date;
    confirmations: number;
}
/**
 * Ring signature parameters
 */
interface RingSignatureParams {
    message: string;
    signers: string[];
    realSignerIndex: number;
    privateKey: string;
}
/**
 * Ring signature
 */
interface RingSignature {
    signature: string;
    ringMembers: string[];
    keyImage: string;
    timestamp: Date;
}
/**
 * Mixing configuration
 */
interface MixingConfig {
    enabled: boolean;
    minMixSize: number;
    maxMixSize: number;
    mixingRounds: number;
    delay: number;
    fee: string;
}
/**
 * Stealth address
 */
interface StealthAddress {
    address: string;
    publicViewKey: string;
    publicSpendKey: string;
    createdAt: Date;
    used: boolean;
}
/**
 * Privacy parameters
 */
interface PrivacyParams {
    privacyEnabled: boolean;
    defaultPrivacyLevel: PrivacyLevel;
    ringSignatureEnabled: boolean;
    confidentialTxEnabled: boolean;
    mixingEnabled: boolean;
    stealthAddressEnabled: boolean;
    minRingSize: number;
    maxRingSize: number;
}
/**
 * Zero-knowledge proof
 */
interface ZeroKnowledgeProof {
    proof: string;
    publicInputs: string[];
    verificationKey: string;
    proofSystem: string;
    createdAt: Date;
}
/**
 * Privacy audit log
 */
interface PrivacyAuditLog {
    address: string;
    action: string;
    privacyLevel: PrivacyLevel;
    timestamp: Date;
    metadata: Record<string, any>;
}

/**
 * Validator security information
 */
interface ValidatorSecurityInfo {
    operatorAddress: string;
    securityScore: number;
    sentryNodesEnabled: boolean;
    sentryNodeCount: number;
    doubleSignProtection: boolean;
    keyRotationEnabled: boolean;
    lastKeyRotation?: Date;
    uptime: number;
    missedBlocks: number;
    jailed: boolean;
    tombstoned: boolean;
}
/**
 * Validator security configuration parameters
 */
interface ConfigureValidatorSecurityParams {
    operatorAddress: string;
    sentryNodesEnabled?: boolean;
    sentryNodes?: string[];
    doubleSignProtection?: boolean;
    keyRotationInterval?: number;
    alertThresholds?: {
        missedBlocks: number;
        downtimeMinutes: number;
    };
}
/**
 * Slashing event
 */
interface SlashingEvent {
    id: string;
    validatorAddress: string;
    reason: string;
    slashedAmount: string;
    slashedPercentage: number;
    height: number;
    timestamp: Date;
    infractionHeight: number;
    jailDuration?: number;
}
/**
 * Slash reason enum
 */
declare enum SlashReason {
    DOUBLE_SIGN = 0,
    DOWNTIME = 1,
    BYZANTINE = 2,
    CENSORSHIP = 3,
    OTHER = 4
}
/**
 * Jailing info
 */
interface JailingInfo {
    validatorAddress: string;
    jailed: boolean;
    jailReason: string;
    jailedAt?: Date;
    releaseTime?: Date;
    missedBlocksCounter: number;
}
/**
 * Sentry node configuration
 */
interface SentryNodeConfig {
    nodeId: string;
    address: string;
    enabled: boolean;
    lastHeartbeat: Date;
    status: string;
    latency: number;
}
/**
 * Validator monitoring metrics
 */
interface ValidatorMonitoringMetrics {
    operatorAddress: string;
    uptime: number;
    missedBlocks: number;
    signedBlocks: number;
    proposedBlocks: number;
    votingPower: string;
    commission: string;
    jailed: boolean;
    bondStatus: string;
    lastActive: Date;
}
/**
 * Double sign detection
 */
interface DoubleSignDetection {
    detected: boolean;
    validatorAddress: string;
    height: number;
    evidenceA: string;
    evidenceB: string;
    timestamp: Date;
    actionTaken: string;
}
/**
 * Validator security parameters
 */
interface ValidatorSecurityParams {
    sentryNodesRequired: boolean;
    minSentryNodes: number;
    keyRotationRequired: boolean;
    keyRotationInterval: number;
    doubleSignSlashPercentage: number;
    downtimeSlashPercentage: number;
    maxMissedBlocks: number;
    jailDuration: number;
    tombstonePermanent: boolean;
}
/**
 * Security audit
 */
interface ValidatorSecurityAudit {
    validatorAddress: string;
    auditDate: Date;
    securityScore: number;
    findings: {
        category: string;
        severity: string;
        description: string;
        recommendation: string;
    }[];
    passed: boolean;
}

/**
 * Verifiable Credential status enum
 */
declare enum VCStatus {
    ACTIVE = 0,
    REVOKED = 1,
    EXPIRED = 2,
    SUSPENDED = 3
}
/**
 * Verifiable Credential
 */
interface VerifiableCredential {
    id: string;
    issuer: string;
    subject: string;
    type: string[];
    issuanceDate: Date;
    expirationDate?: Date;
    claims: Record<string, any>;
    proof: {
        type: string;
        created: Date;
        proofPurpose: string;
        verificationMethod: string;
        signature: string;
    };
    status: VCStatus;
    revocationReason?: string;
    revokedAt?: Date;
}
/**
 * Mint VC parameters
 */
interface MintVCParams {
    issuer: string;
    subject: string;
    credentialType: string[];
    claims: Record<string, any>;
    expirationDate?: Date;
    transferable?: boolean;
}
/**
 * Revoke VC parameters
 */
interface RevokeVCParams {
    vcId: string;
    issuer: string;
    reason: string;
}
/**
 * Verifiable Presentation
 */
interface VerifiablePresentation {
    id: string;
    holder: string;
    verifiableCredentials: string[];
    type: string[];
    purpose: string;
    createdAt: Date;
    expiresAt?: Date;
    proof: {
        type: string;
        created: Date;
        proofPurpose: string;
        verificationMethod: string;
        signature: string;
        challenge?: string;
    };
}
/**
 * Create presentation parameters
 */
interface CreatePresentationParams {
    holder: string;
    vcIds: string[];
    purpose: string;
    expirationDuration?: number;
    challenge?: string;
}
/**
 * Verification result
 */
interface VCVerificationResult {
    valid: boolean;
    vcId: string;
    issuer: string;
    subject: string;
    status: VCStatus;
    verified: boolean;
    signatureValid: boolean;
    notExpired: boolean;
    notRevoked: boolean;
    issuerTrusted: boolean;
    errors: string[];
    verifiedAt: Date;
}
/**
 * Presentation verification result
 */
interface PresentationVerificationResult {
    valid: boolean;
    presentationId: string;
    holder: string;
    vcResults: VCVerificationResult[];
    signatureValid: boolean;
    notExpired: boolean;
    challengeVerified: boolean;
    errors: string[];
    verifiedAt: Date;
}
/**
 * VC Registry statistics
 */
interface VCStats {
    totalVCs: number;
    activeVCs: number;
    revokedVCs: number;
    expiredVCs: number;
    totalIssuers: number;
    totalSubjects: number;
    totalPresentations: number;
    credentialTypes: {
        type: string;
        count: number;
    }[];
}
/**
 * VC query filters
 */
interface VCQueryFilters {
    issuer?: string;
    subject?: string;
    type?: string;
    status?: VCStatus;
    fromDate?: Date;
    toDate?: Date;
    limit?: number;
    offset?: number;
}
/**
 * VC Registry parameters
 */
interface VCRegistryParams {
    issuanceEnabled: boolean;
    revocationEnabled: boolean;
    maxExpirationDuration: number;
    requireExpirationDate: boolean;
    trustedIssuers: string[];
    supportedTypes: string[];
    verificationFee: string;
}
/**
 * Issuer info
 */
interface IssuerInfo {
    address: string;
    name: string;
    credentialsIssued: number;
    credentialsRevoked: number;
    trusted: boolean;
    registeredAt: Date;
    publicKey: string;
}

/**
 * Wallet security configuration
 */
interface WalletSecurityConfig {
    address: string;
    multisigEnabled: boolean;
    requiredSignatures?: number;
    signers?: string[];
    sessionTimeout: number;
    biometricEnabled: boolean;
    twoFactorEnabled: boolean;
    whitelistedAddresses: string[];
    dailyTransactionLimit?: string;
    requireApprovalAbove?: string;
    lastUpdated: Date;
}
/**
 * Configure wallet security parameters
 */
interface ConfigureWalletSecurityParams {
    address: string;
    sessionTimeout?: number;
    biometricEnabled?: boolean;
    twoFactorEnabled?: boolean;
    whitelistedAddresses?: string[];
    dailyTransactionLimit?: string;
    requireApprovalAbove?: string;
}
/**
 * Enable multisig parameters
 */
interface EnableMultisigParams {
    address: string;
    threshold: number;
    signers: string[];
}
/**
 * Multisig transaction
 */
interface MultisigTransaction {
    id: string;
    creator: string;
    multisigAddress: string;
    messages: any[];
    signatures: {
        signer: string;
        signature: string;
        timestamp: Date;
    }[];
    requiredSignatures: number;
    status: MultisigTransactionStatus;
    createdAt: Date;
    expiresAt: Date;
    executedAt?: Date;
}
/**
 * Multisig transaction status enum
 */
declare enum MultisigTransactionStatus {
    PENDING = 0,
    APPROVED = 1,
    EXECUTED = 2,
    REJECTED = 3,
    EXPIRED = 4
}
/**
 * Sign multisig transaction parameters
 */
interface SignMultisigParams {
    transactionId: string;
    signer: string;
    signature: string;
}
/**
 * Wallet session
 */
interface WalletSession {
    id: string;
    address: string;
    deviceId: string;
    ipAddress: string;
    userAgent: string;
    createdAt: Date;
    expiresAt: Date;
    lastActivity: Date;
    biometricVerified: boolean;
    twoFactorVerified: boolean;
}
/**
 * Transaction approval request
 */
interface TransactionApprovalRequest {
    id: string;
    address: string;
    transaction: any;
    amount: string;
    recipient: string;
    status: ApprovalStatus;
    requestedAt: Date;
    expiresAt: Date;
    approvedAt?: Date;
    approver?: string;
}
/**
 * Approval status enum
 */
declare enum ApprovalStatus {
    PENDING = 0,
    APPROVED = 1,
    REJECTED = 2,
    EXPIRED = 3
}
/**
 * Security alert
 */
interface SecurityAlert {
    id: string;
    address: string;
    type: string;
    severity: 'low' | 'medium' | 'high' | 'critical';
    message: string;
    details: Record<string, any>;
    triggeredAt: Date;
    acknowledged: boolean;
    acknowledgedAt?: Date;
}
/**
 * Wallet security parameters
 */
interface WalletSecurityParams {
    sessionTimeoutEnabled: boolean;
    defaultSessionTimeout: number;
    maxSessionDuration: number;
    biometricAuthEnabled: boolean;
    twoFactorAuthEnabled: boolean;
    multisigEnabled: boolean;
    transactionLimitsEnabled: boolean;
    defaultDailyLimit: string;
    approvalThreshold: string;
}
/**
 * Biometric verification
 */
interface BiometricVerification {
    address: string;
    deviceId: string;
    biometricType: string;
    verified: boolean;
    verifiedAt: Date;
    expiresAt: Date;
}
/**
 * Two-factor authentication
 */
interface TwoFactorAuth {
    address: string;
    enabled: boolean;
    method: string;
    secret?: string;
    backupCodes?: string[];
    lastUsed?: Date;
}

interface AuraChainConfig {
    rpcEndpoint: string;
    restEndpoint?: string;
    chainId: string;
    prefix?: string;
    gasPrice?: string;
    gasAdjustment?: number;
}
interface WalletAccount {
    address: string;
    pubkey: Uint8Array;
    algo: string;
}
interface Pool {
    id: string;
    tokenA: string;
    tokenB: string;
    reserveA: string;
    reserveB: string;
    totalShares: string;
    swapFee: string;
}
interface PoolParams {
    tokenA: string;
    tokenB: string;
    amountA: string;
    amountB: string;
}
interface SwapParams {
    poolId: string;
    tokenIn: string;
    amountIn: string;
    minAmountOut: string;
    recipient?: string;
}
interface AddLiquidityParams {
    poolId: string;
    amountA: string;
    amountB: string;
    minShares: string;
}
interface RemoveLiquidityParams {
    poolId: string;
    shares: string;
    minAmountA: string;
    minAmountB: string;
}
interface Validator {
    operatorAddress: string;
    consensusPubkey: string;
    jailed: boolean;
    status: number;
    tokens: string;
    delegatorShares: string;
    description: {
        moniker: string;
        identity: string;
        website: string;
        securityContact: string;
        details: string;
    };
    commission: {
        rate: string;
        maxRate: string;
        maxChangeRate: string;
    };
}
interface DelegateParams {
    validatorAddress: string;
    amount: string;
    denom?: string;
}
interface UndelegateParams {
    validatorAddress: string;
    amount: string;
    denom?: string;
}
interface RedelegateParams {
    srcValidatorAddress: string;
    dstValidatorAddress: string;
    amount: string;
    denom?: string;
}
interface Proposal {
    proposalId: string;
    content: {
        typeUrl: string;
        value: Uint8Array;
    };
    status: number;
    finalTallyResult: {
        yes: string;
        abstain: string;
        no: string;
        noWithVeto: string;
    };
    submitTime: Date;
    depositEndTime: Date;
    totalDeposit: Coin[];
    votingStartTime: Date;
    votingEndTime: Date;
}
interface VoteParams {
    proposalId: string;
    option: VoteOption;
    metadata?: string;
}
interface DepositParams {
    proposalId: string;
    amount: string;
    denom?: string;
}
declare enum VoteOption {
    UNSPECIFIED = 0,
    YES = 1,
    ABSTAIN = 2,
    NO = 3,
    NO_WITH_VETO = 4
}
interface TxResult {
    transactionHash: string;
    height: number;
    code: number;
    rawLog?: string;
    gasUsed: number;
    gasWanted: number;
}
interface QueryBalance {
    denom: string;
    amount: string;
}
interface SendParams {
    recipient: string;
    amount: string;
    denom?: string;
    memo?: string;
}
interface GasOptions {
    gasLimit?: number;
    gasPrice?: string;
    memo?: string;
}

declare class AuraWallet {
    private wallet;
    private prefix;
    constructor(prefix?: string);
    /**
     * Generate a new 24-word mnemonic
     */
    static generateMnemonic(): string;
    /**
     * Validate a mnemonic phrase
     */
    static validateMnemonic(mnemonic: string): boolean;
    /**
     * Create wallet from mnemonic
     */
    fromMnemonic(mnemonic: string, hdPath?: string): Promise<void>;
    /**
     * Get wallet accounts
     */
    getAccounts(): Promise<WalletAccount[]>;
    /**
     * Get first account address
     */
    getAddress(): Promise<string>;
    /**
     * Get the offline signer for transaction signing
     */
    getSigner(): OfflineDirectSigner;
    /**
     * Export mnemonic (use with caution!)
     */
    exportMnemonic(): Promise<string>;
}

declare class TxBuilder {
    private client;
    private defaultGasPrice;
    private defaultGasAdjustment;
    constructor(client: SigningStargateClient, gasPrice?: string, gasAdjustment?: number);
    /**
     * Sign and broadcast a transaction
     */
    signAndBroadcast(signerAddress: string, messages: readonly EncodeObject[], options?: GasOptions): Promise<TxResult>;
    /**
     * Calculate transaction fee
     */
    calculateFee(signerAddress: string, messages: readonly EncodeObject[], options?: GasOptions): Promise<StdFee>;
    /**
     * Simulate transaction
     */
    simulate(signerAddress: string, messages: readonly EncodeObject[]): Promise<number>;
    /**
     * Extract denom from gas price string
     */
    private extractDenom;
    /**
     * Calculate fee amount
     */
    private calculateAmount;
    /**
     * Format transaction result
     */
    private formatResult;
}

declare class BankModule {
    private client;
    constructor(client: AuraClient);
    /**
     * Get account balance for a specific denom
     */
    getBalance(address: string, denom: string): Promise<Coin | null>;
    /**
     * Get all account balances
     */
    getAllBalances(address: string): Promise<readonly Coin[]>;
    /**
     * Send tokens to another address
     */
    send(senderAddress: string, params: SendParams, options?: GasOptions): Promise<TxResult>;
    /**
     * Multi-send tokens to multiple recipients
     */
    multiSend(senderAddress: string, recipients: Array<{
        address: string;
        amount: string;
        denom?: string;
    }>, options?: GasOptions): Promise<TxResult>;
    /**
     * Get total supply of a denom
     */
    getTotalSupply(denom: string): Promise<Coin | null>;
    /**
     * Get all denoms
     */
    getAllDenoms(): Promise<string[]>;
    /**
     * Get spendable balance (available for spending)
     */
    getSpendableBalance(address: string, denom: string): Promise<Coin | null>;
    /**
     * Format balance for display
     */
    formatBalance(balance: Coin, decimals?: number): string;
}

declare class DexModule {
    private client;
    constructor(client: AuraClient);
    /**
     * Create a new liquidity pool
     */
    createPool(creator: string, params: PoolParams, options?: GasOptions): Promise<TxResult>;
    /**
     * Add liquidity to an existing pool
     */
    addLiquidity(sender: string, params: AddLiquidityParams, options?: GasOptions): Promise<TxResult>;
    /**
     * Remove liquidity from a pool
     */
    removeLiquidity(sender: string, params: RemoveLiquidityParams, options?: GasOptions): Promise<TxResult>;
    /**
     * Swap tokens
     */
    swap(sender: string, params: SwapParams, options?: GasOptions): Promise<TxResult>;
    /**
     * Get pool by ID
     */
    getPool(poolId: string): Promise<Pool | null>;
    /**
     * Get all pools
     */
    getAllPools(): Promise<Pool[]>;
    /**
     * Get pool for token pair
     */
    getPoolByTokens(tokenA: string, tokenB: string): Promise<Pool | null>;
    /**
     * Calculate swap output amount
     */
    calculateSwapOutput(amountIn: string, reserveIn: string, reserveOut: string, swapFee?: string): string;
    /**
     * Calculate price impact
     */
    calculatePriceImpact(amountIn: string, reserveIn: string, reserveOut: string): number;
    /**
     * Calculate shares for liquidity addition
     */
    calculateShares(amountA: string, amountB: string, reserveA: string, _reserveB: string, totalShares: string): string;
}

declare class StakingModule {
    private client;
    constructor(client: AuraClient);
    /**
     * Delegate tokens to a validator
     */
    delegate(delegator: string, params: DelegateParams, options?: GasOptions): Promise<TxResult>;
    /**
     * Undelegate tokens from a validator
     */
    undelegate(delegator: string, params: UndelegateParams, options?: GasOptions): Promise<TxResult>;
    /**
     * Redelegate tokens from one validator to another
     */
    redelegate(delegator: string, params: RedelegateParams, options?: GasOptions): Promise<TxResult>;
    /**
     * Withdraw delegation rewards from a validator
     */
    withdrawRewards(delegator: string, validatorAddress: string, options?: GasOptions): Promise<TxResult>;
    /**
     * Withdraw all delegation rewards
     */
    withdrawAllRewards(delegator: string, options?: GasOptions): Promise<TxResult>;
    /**
     * Get all validators
     */
    getValidators(): Promise<Validator[]>;
    /**
     * Get validator by address
     */
    getValidator(validatorAddress: string): Promise<Validator | null>;
    /**
     * Get delegations for a delegator
     */
    getDelegations(delegator: string): Promise<any[]>;
    /**
     * Get delegation to a specific validator
     */
    getDelegation(delegator: string, validatorAddress: string): Promise<any | null>;
    /**
     * Get unbonding delegations
     */
    getUnbondingDelegations(delegator: string): Promise<any[]>;
    /**
     * Get rewards for a delegator
     */
    getRewards(delegator: string): Promise<Coin[]>;
    /**
     * Get staking pool
     */
    getPool(): Promise<any | null>;
    /**
     * Calculate APY for a validator
     */
    calculateAPY(validator: Validator, annualProvisions: string, totalBondedTokens: string): number;
}

declare class GovernanceModule {
    private client;
    constructor(client: AuraClient);
    /**
     * Submit a text proposal
     */
    submitTextProposal(proposer: string, title: string, description: string, initialDeposit: string, denom?: string, options?: GasOptions): Promise<TxResult>;
    /**
     * Vote on a proposal
     */
    vote(voter: string, params: VoteParams, options?: GasOptions): Promise<TxResult>;
    /**
     * Deposit to a proposal
     */
    deposit(depositor: string, params: DepositParams, options?: GasOptions): Promise<TxResult>;
    /**
     * Get all proposals
     */
    getProposals(status?: number): Promise<Proposal[]>;
    /**
     * Get proposal by ID
     */
    getProposal(proposalId: string): Promise<Proposal | null>;
    /**
     * Get votes for a proposal
     */
    getVotes(proposalId: string): Promise<any[]>;
    /**
     * Get vote for a specific voter
     */
    getVote(proposalId: string, voter: string): Promise<any | null>;
    /**
     * Get deposits for a proposal
     */
    getDeposits(proposalId: string): Promise<any[]>;
    /**
     * Get tally for a proposal
     */
    getTally(proposalId: string): Promise<any | null>;
    /**
     * Get governance parameters
     */
    getParams(paramsType: 'voting' | 'tallying' | 'deposit'): Promise<any | null>;
    /**
     * Check if proposal has passed quorum
     */
    hasQuorum(proposalId: string): Promise<boolean>;
    /**
     * Get vote option name
     */
    getVoteOptionName(option: VoteOption): string;
}

declare class AuraClient {
    private config;
    private client;
    private signingClient;
    private txBuilder;
    readonly bank: BankModule;
    readonly dex: DexModule;
    readonly staking: StakingModule;
    readonly governance: GovernanceModule;
    constructor(config: AuraChainConfig);
    /**
     * Connect to the blockchain without signing capabilities
     */
    connect(): Promise<void>;
    /**
     * Connect with a wallet for signing transactions
     */
    connectWithWallet(wallet: AuraWallet): Promise<void>;
    /**
     * Get the read-only client
     */
    getClient(): StargateClient;
    /**
     * Get the signing client
     */
    getSigningClient(): SigningStargateClient;
    /**
     * Get the transaction builder
     */
    getTxBuilder(): TxBuilder;
    /**
     * Get chain configuration
     */
    getConfig(): AuraChainConfig;
    /**
     * Get current block height
     */
    getHeight(): Promise<number>;
    /**
     * Get chain ID
     */
    getChainId(): Promise<string>;
    /**
     * Disconnect from the blockchain
     */
    disconnect(): Promise<void>;
    /**
     * Check if client is connected
     */
    isConnected(): boolean;
    /**
     * Check if client has signing capabilities
     */
    canSign(): boolean;
}

export { type AccountInfo, type AddLiquidityParams, type Alert, AlertSeverity, AlertStatus, type AnomalyDetection, ApprovalStatus, type ApproveChangeRequestParams, type AuraChainConfig, AuraClient, AuraWallet, type AuthInfo, type AuthParams, type BalanceCheckResult, BankModule, type BiometricVerification, type BridgeParams, type BridgeSecurity, type BridgeStats, type BridgeTransfer, BridgeTransferStatus, ChangeRequestStatus, ChangeType, type ClaimRewardsParams, type CompleteBridgeParams, type CompleteRoutineParams, type ComplianceParams, type ComplianceStatus, ComplianceStatusType, type ComplianceValidationResult, type ConfidenceCompleteRoutineParams, type ConfidenceRewards, type ConfidenceScore, type ConfidenceScoreParams, type ConfidenceSlashEvent, type ConfidentialTransaction, type ConfigureValidatorSecurityParams, type ConfigureWalletSecurityParams, type CreatePresentationParams, type CreateRoutineParams, type CryptographyParams, type CustomRule, type DDoSMitigationStatus, type DataAccessGrant, type DataItem, DataItemStatus, type DataQueryFilters, type DataRegistryParams, type DataStats, type DecryptParams, type DelegateParams, type DeleteDataParams, type DepositParams, DexModule, type DoubleSignDetection, type DynamicFeeStructure, type EconomicAttackDetection, type EconomicSecurityParams, type EnableMultisigParams, type EncryptParams, type EncryptedData, type ExecuteParams, type FeeMarketMetrics, type ForkDetection, type GasOptions, type GossipNetworkStatus, GovernanceModule, type GrantAuthorization, type GrantParams, type IdentityChangeParams, type IdentityChangeRequest, type IdentityHistoryEntry, type IdentityVerificationStatus, type InclusionRoutine, type InclusionRoutineParams, type InitiateBridgeParams, type IssuerInfo, type JailingInfo, KYCLevel, type KeyInfo, type KeyPair, type KeyRotationParams, KeyStatus, MEVProtectionLevel, type MEVProtectionStatus, type MintVCParams, type MixingConfig, type MonitoringConfig, type MonitoringParams, type MultisigTransaction, MultisigTransactionStatus, type NetworkSecurityParams, type NetworkSecurityStatus, type NodeHealth, type PartitionDetection, type PeerReputation, type PerformanceMetrics, type Pool, type PoolParams, type PresentationVerificationResult, type PrevalidationParams, type PrivacyAuditLog, PrivacyLevel, type PrivacyParams, type PrivacySettings, type PrivateTransactionParams, type Proposal, type QuantumKeyPair, type QueryBalance, type RandomRequest, type RandomResponse, type RateLimitsConfig, type RedelegateParams, type RegisterDataParams, type RemoveLiquidityParams, type ReportThreatParams, type RevokeParams, type RevokeVCParams, type RingSignature, type RingSignatureParams, type RoutineLeaderboardEntry, type RoutinePrerequisites, type RoutineRequirement, type RoutineStats, RoutineStatus, RoutineType, type SanctionCheckResult, type ScoreFactors, type ScoreHistoryEntry, type SecureEnclaveConfig, type SecurityAlert, type SendParams, type SentryNodeConfig, type SignMultisigParams, type SignatureValidationResult, SlashReason, type SlashingEvent, StakingModule, type StealthAddress, type StructureValidationResult, type SubmitChangeRequestParams, type SubmitKYCParams, type SwapParams, type SybilProtectionStatus, type SystemMetrics, ThreatLevel, type TransactionApprovalRequest, type TransactionReport, type TwoFactorAuth, TxBuilder, type TxResult, type UndelegateParams, type UpdateDataParams, type VCQueryFilters, type VCRegistryParams, type VCStats, VCStatus, type VCVerificationResult, type ValidationError, type ValidationResult, type ValidationRules, type ValidationWarning, type Validator, type ValidatorMonitoringMetrics, type ValidatorSecurityAudit, type ValidatorSecurityInfo, type ValidatorSecurityParams, type VerifiableCredential, type VerifiablePresentation, VoteOption, type VoteParams, type WalletAccount, type WalletSecurityConfig, type WalletSecurityParams, type WalletSession, type WhaleProtectionConfig, type WhaleTransaction, type ZeroKnowledgeProof };
