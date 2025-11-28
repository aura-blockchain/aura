/**
 * Inclusion routine status enum
 */
export enum RoutineStatus {
  PENDING = 0,
  IN_PROGRESS = 1,
  COMPLETED = 2,
  FAILED = 3,
  EXPIRED = 4,
}

/**
 * Routine type enum
 */
export enum RoutineType {
  VERIFICATION = 0,
  STAKING = 1,
  GOVERNANCE = 2,
  SOCIAL = 3,
  EDUCATIONAL = 4,
  CUSTOM = 5,
}

/**
 * Inclusion routine
 */
export interface InclusionRoutine {
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
export interface RoutineRequirement {
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
export interface CreateRoutineParams {
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
export interface CompleteRoutineParams {
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
export interface RoutinePrerequisites {
  minScore: number;
  requiredRoutines: string[];
  kycLevel: number;
  minStake?: string;
}

/**
 * Inclusion routine parameters
 */
export interface InclusionRoutineParams {
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
export interface RoutineStats {
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
export interface RoutineLeaderboardEntry {
  address: string;
  completedRoutines: number;
  totalScore: number;
  rank: number;
  badges: string[];
}
