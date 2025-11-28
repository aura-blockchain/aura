/**
 * Confidence score information
 */
export interface ConfidenceScore {
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
export interface ScoreHistoryEntry {
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
export interface ConfidenceRewards {
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
export interface ClaimRewardsParams {
  address: string;
  amount?: string; // If not specified, claim all
}

/**
 * Complete inclusion routine parameters (for confidence score module)
 */
export interface ConfidenceCompleteRoutineParams {
  address: string;
  routineId: string;
  proof: string;
  metadata?: Record<string, any>;
}

/**
 * Confidence score parameters
 */
export interface ConfidenceScoreParams {
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
export interface ScoreFactors {
  transactionWeight: number;
  stakingWeight: number;
  governanceWeight: number;
  routineWeight: number;
  reputationWeight: number;
}

/**
 * Slash event for confidence score
 */
export interface ConfidenceSlashEvent {
  address: string;
  oldScore: number;
  newScore: number;
  reason: string;
  amount: number;
  timestamp: Date;
}
