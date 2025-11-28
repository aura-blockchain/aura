/**
 * Identity change request status enum
 */
export enum ChangeRequestStatus {
  PENDING = 0,
  APPROVED = 1,
  REJECTED = 2,
  EXPIRED = 3,
  CANCELLED = 4,
}

/**
 * Identity change type enum
 */
export enum ChangeType {
  NAME = 0,
  EMAIL = 1,
  PHONE = 2,
  ADDRESS = 3,
  DOCUMENTS = 4,
  OTHER = 5,
}

/**
 * Identity change request
 */
export interface IdentityChangeRequest {
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
export interface SubmitChangeRequestParams {
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
export interface ApproveChangeRequestParams {
  requestId: string;
  approver: string;
  approved: boolean;
  comments?: string;
}

/**
 * Identity change parameters
 */
export interface IdentityChangeParams {
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
export interface IdentityHistoryEntry {
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
export interface IdentityVerificationStatus {
  address: string;
  verified: boolean;
  verificationLevel: number;
  lastVerified?: Date;
  pendingChanges: number;
  totalChanges: number;
}
