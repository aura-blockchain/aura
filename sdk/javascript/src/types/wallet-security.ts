/**
 * Wallet security configuration
 */
export interface WalletSecurityConfig {
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
export interface ConfigureWalletSecurityParams {
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
export interface EnableMultisigParams {
  address: string;
  threshold: number;
  signers: string[];
}

/**
 * Multisig transaction
 */
export interface MultisigTransaction {
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
export enum MultisigTransactionStatus {
  PENDING = 0,
  APPROVED = 1,
  EXECUTED = 2,
  REJECTED = 3,
  EXPIRED = 4,
}

/**
 * Sign multisig transaction parameters
 */
export interface SignMultisigParams {
  transactionId: string;
  signer: string;
  signature: string;
}

/**
 * Wallet session
 */
export interface WalletSession {
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
export interface TransactionApprovalRequest {
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
export enum ApprovalStatus {
  PENDING = 0,
  APPROVED = 1,
  REJECTED = 2,
  EXPIRED = 3,
}

/**
 * Security alert
 */
export interface SecurityAlert {
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
export interface WalletSecurityParams {
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
export interface BiometricVerification {
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
export interface TwoFactorAuth {
  address: string;
  enabled: boolean;
  method: string; // 'totp', 'sms', 'email'
  secret?: string;
  backupCodes?: string[];
  lastUsed?: Date;
}
