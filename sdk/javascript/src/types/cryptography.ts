/**
 * Key pair interface
 */
export interface KeyPair {
  publicKey: string;
  privateKey: string;
  algorithm: string;
  createdAt: Date;
  expiresAt?: Date;
}

/**
 * Encrypted data
 */
export interface EncryptedData {
  ciphertext: string;
  nonce: string;
  algorithm: string;
  publicKey: string;
  timestamp: Date;
}

/**
 * Key rotation parameters
 */
export interface KeyRotationParams {
  address: string;
  oldPublicKey: string;
  newPublicKey: string;
  signature: string;
  reason?: string;
}

/**
 * Encryption parameters
 */
export interface EncryptParams {
  data: string;
  publicKey: string;
  algorithm?: string;
}

/**
 * Decryption parameters
 */
export interface DecryptParams {
  encryptedData: EncryptedData;
  privateKey: string;
}

/**
 * Quantum-resistant key pair
 */
export interface QuantumKeyPair extends KeyPair {
  quantumSafe: boolean;
  latticeParams?: {
    dimension: number;
    modulus: number;
  };
}

/**
 * Key info
 */
export interface KeyInfo {
  address: string;
  publicKey: string;
  algorithm: string;
  createdAt: Date;
  rotatedAt?: Date;
  rotationCount: number;
  isQuantumSafe: boolean;
  status: KeyStatus;
}

/**
 * Key status enum
 */
export enum KeyStatus {
  ACTIVE = 0,
  ROTATED = 1,
  REVOKED = 2,
  EXPIRED = 3,
}

/**
 * Cryptography parameters
 */
export interface CryptographyParams {
  defaultAlgorithm: string;
  keyRotationInterval: number;
  quantumSafeEnabled: boolean;
  maxKeyAge: number;
  supportedAlgorithms: string[];
}

/**
 * Secure enclave configuration
 */
export interface SecureEnclaveConfig {
  enabled: boolean;
  provider: string;
  attestation: string;
  keyStorageLevel: string;
}

/**
 * Random number request
 */
export interface RandomRequest {
  length: number;
  encoding?: 'hex' | 'base64' | 'bytes';
}

/**
 * Random number response
 */
export interface RandomResponse {
  value: string;
  entropy: number;
  timestamp: Date;
}
