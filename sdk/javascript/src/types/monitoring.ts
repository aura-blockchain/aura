/**
 * Alert severity enum
 */
export enum AlertSeverity {
  INFO = 0,
  WARNING = 1,
  ERROR = 2,
  CRITICAL = 3,
}

/**
 * Alert status enum
 */
export enum AlertStatus {
  ACTIVE = 0,
  ACKNOWLEDGED = 1,
  RESOLVED = 2,
}

/**
 * System metrics
 */
export interface SystemMetrics {
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
export interface Alert {
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
export interface MonitoringConfig {
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
export interface PerformanceMetrics {
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
export interface NodeHealth {
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
export interface AnomalyDetection {
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
export interface MonitoringParams {
  enabled: boolean;
  metricsRetention: number;
  alertingEnabled: boolean;
  anomalyDetectionEnabled: boolean;
  mlModelVersion: string;
  samplingRate: number;
}
