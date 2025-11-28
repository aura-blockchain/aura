/**
 * Validator security information
 */
export interface ValidatorSecurityInfo {
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
export interface ConfigureValidatorSecurityParams {
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
export interface SlashingEvent {
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
export enum SlashReason {
  DOUBLE_SIGN = 0,
  DOWNTIME = 1,
  BYZANTINE = 2,
  CENSORSHIP = 3,
  OTHER = 4,
}

/**
 * Jailing info
 */
export interface JailingInfo {
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
export interface SentryNodeConfig {
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
export interface ValidatorMonitoringMetrics {
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
export interface DoubleSignDetection {
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
export interface ValidatorSecurityParams {
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
export interface ValidatorSecurityAudit {
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
