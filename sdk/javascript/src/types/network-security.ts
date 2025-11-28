/**
 * Network security status
 */
export interface NetworkSecurityStatus {
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
export enum ThreatLevel {
  NONE = 0,
  LOW = 1,
  MEDIUM = 2,
  HIGH = 3,
  CRITICAL = 4,
}

/**
 * Threat report parameters
 */
export interface ReportThreatParams {
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
export interface SybilProtectionStatus {
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
export interface RateLimitsConfig {
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
export interface PeerReputation {
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
export interface NetworkSecurityParams {
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
export interface GossipNetworkStatus {
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
export interface DDoSMitigationStatus {
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
export interface ForkDetection {
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
export interface PartitionDetection {
  detected: boolean;
  affectedNodes: number;
  isolatedNodes: string[];
  detectedAt: Date;
  duration: number;
  resolved: boolean;
}
