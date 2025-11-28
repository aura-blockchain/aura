/**
 * Verifiable Credential status enum
 */
export enum VCStatus {
  ACTIVE = 0,
  REVOKED = 1,
  EXPIRED = 2,
  SUSPENDED = 3,
}

/**
 * Verifiable Credential
 */
export interface VerifiableCredential {
  id: string;
  issuer: string;
  subject: string;
  type: string[];
  issuanceDate: Date;
  expirationDate?: Date;
  claims: Record<string, any>;
  proof: {
    type: string;
    created: Date;
    proofPurpose: string;
    verificationMethod: string;
    signature: string;
  };
  status: VCStatus;
  revocationReason?: string;
  revokedAt?: Date;
}

/**
 * Mint VC parameters
 */
export interface MintVCParams {
  issuer: string;
  subject: string;
  credentialType: string[];
  claims: Record<string, any>;
  expirationDate?: Date;
  transferable?: boolean;
}

/**
 * Revoke VC parameters
 */
export interface RevokeVCParams {
  vcId: string;
  issuer: string;
  reason: string;
}

/**
 * Verifiable Presentation
 */
export interface VerifiablePresentation {
  id: string;
  holder: string;
  verifiableCredentials: string[]; // VC IDs
  type: string[];
  purpose: string;
  createdAt: Date;
  expiresAt?: Date;
  proof: {
    type: string;
    created: Date;
    proofPurpose: string;
    verificationMethod: string;
    signature: string;
    challenge?: string;
  };
}

/**
 * Create presentation parameters
 */
export interface CreatePresentationParams {
  holder: string;
  vcIds: string[];
  purpose: string;
  expirationDuration?: number;
  challenge?: string;
}

/**
 * Verification result
 */
export interface VCVerificationResult {
  valid: boolean;
  vcId: string;
  issuer: string;
  subject: string;
  status: VCStatus;
  verified: boolean;
  signatureValid: boolean;
  notExpired: boolean;
  notRevoked: boolean;
  issuerTrusted: boolean;
  errors: string[];
  verifiedAt: Date;
}

/**
 * Presentation verification result
 */
export interface PresentationVerificationResult {
  valid: boolean;
  presentationId: string;
  holder: string;
  vcResults: VCVerificationResult[];
  signatureValid: boolean;
  notExpired: boolean;
  challengeVerified: boolean;
  errors: string[];
  verifiedAt: Date;
}

/**
 * VC Registry statistics
 */
export interface VCStats {
  totalVCs: number;
  activeVCs: number;
  revokedVCs: number;
  expiredVCs: number;
  totalIssuers: number;
  totalSubjects: number;
  totalPresentations: number;
  credentialTypes: {
    type: string;
    count: number;
  }[];
}

/**
 * VC query filters
 */
export interface VCQueryFilters {
  issuer?: string;
  subject?: string;
  type?: string;
  status?: VCStatus;
  fromDate?: Date;
  toDate?: Date;
  limit?: number;
  offset?: number;
}

/**
 * VC Registry parameters
 */
export interface VCRegistryParams {
  issuanceEnabled: boolean;
  revocationEnabled: boolean;
  maxExpirationDuration: number;
  requireExpirationDate: boolean;
  trustedIssuers: string[];
  supportedTypes: string[];
  verificationFee: string;
}

/**
 * Issuer info
 */
export interface IssuerInfo {
  address: string;
  name: string;
  credentialsIssued: number;
  credentialsRevoked: number;
  trusted: boolean;
  registeredAt: Date;
  publicKey: string;
}
