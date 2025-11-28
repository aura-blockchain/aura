/**
 * Privacy level enum
 */
export enum PrivacyLevel {
  PUBLIC = 0,
  PSEUDONYMOUS = 1,
  PRIVATE = 2,
  ANONYMOUS = 3,
}

/**
 * Privacy settings
 */
export interface PrivacySettings {
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
export interface PrivateTransactionParams {
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
export interface ConfidentialTransaction {
  txHash: string;
  commitment: string;
  rangeProof: string;
  sender: string; // Stealth address
  recipient: string; // Stealth address
  timestamp: Date;
  confirmations: number;
}

/**
 * Ring signature parameters
 */
export interface RingSignatureParams {
  message: string;
  signers: string[]; // Public keys
  realSignerIndex: number;
  privateKey: string;
}

/**
 * Ring signature
 */
export interface RingSignature {
  signature: string;
  ringMembers: string[];
  keyImage: string;
  timestamp: Date;
}

/**
 * Mixing configuration
 */
export interface MixingConfig {
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
export interface StealthAddress {
  address: string;
  publicViewKey: string;
  publicSpendKey: string;
  createdAt: Date;
  used: boolean;
}

/**
 * Privacy parameters
 */
export interface PrivacyParams {
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
export interface ZeroKnowledgeProof {
  proof: string;
  publicInputs: string[];
  verificationKey: string;
  proofSystem: string;
  createdAt: Date;
}

/**
 * Privacy audit log
 */
export interface PrivacyAuditLog {
  address: string;
  action: string;
  privacyLevel: PrivacyLevel;
  timestamp: Date;
  metadata: Record<string, any>;
}
