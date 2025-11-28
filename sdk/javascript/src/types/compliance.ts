/**
 * Compliance status enum
 */
export enum ComplianceStatusType {
  UNKNOWN = 0,
  PENDING = 1,
  APPROVED = 2,
  REJECTED = 3,
  REVOKED = 4,
}

/**
 * KYC level enum
 */
export enum KYCLevel {
  NONE = 0,
  BASIC = 1,
  INTERMEDIATE = 2,
  ADVANCED = 3,
}

/**
 * Compliance status for an address
 */
export interface ComplianceStatus {
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
export interface SubmitKYCParams {
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
export interface TransactionReport {
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
export interface ComplianceParams {
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
export interface SanctionCheckResult {
  address: string;
  sanctioned: boolean;
  lists: string[];
  details?: string;
  checkedAt: Date;
}
