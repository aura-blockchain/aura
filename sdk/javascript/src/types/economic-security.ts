/**
 * Dynamic fee structure
 */
export interface DynamicFeeStructure {
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
export interface MEVProtectionStatus {
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
export enum MEVProtectionLevel {
  NONE = 0,
  LOW = 1,
  MEDIUM = 2,
  HIGH = 3,
}

/**
 * Whale protection configuration
 */
export interface WhaleProtectionConfig {
  enabled: boolean;
  transactionLimit: string;
  dailyLimit: string;
  cooldownPeriod: number;
  exemptAddresses: string[];
}

/**
 * Whale transaction info
 */
export interface WhaleTransaction {
  address: string;
  amount: string;
  timestamp: Date;
  delayed: boolean;
  releaseTime?: Date;
}

/**
 * Economic security parameters
 */
export interface EconomicSecurityParams {
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
export interface FeeMarketMetrics {
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
export interface EconomicAttackDetection {
  attackType: string;
  severity: number;
  affectedAddresses: string[];
  detectedAt: Date;
  mitigated: boolean;
  details: string;
}
