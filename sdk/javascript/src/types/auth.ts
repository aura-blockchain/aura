/**
 * Authentication info
 */
export interface AuthInfo {
  address: string;
  publicKey: string;
  accountNumber: number;
  sequence: number;
  authType: string;
}

/**
 * Account info
 */
export interface AccountInfo {
  address: string;
  publicKey?: Uint8Array;
  accountNumber: number;
  sequence: number;
}

/**
 * Grant authorization
 */
export interface GrantAuthorization {
  granter: string;
  grantee: string;
  authorization: any;
  expiration?: Date;
}

/**
 * Grant parameters
 */
export interface GrantParams {
  granter: string;
  grantee: string;
  msgTypeUrl: string;
  expiration?: Date;
  spendLimit?: string;
}

/**
 * Revoke parameters
 */
export interface RevokeParams {
  granter: string;
  grantee: string;
  msgTypeUrl: string;
}

/**
 * Execute parameters
 */
export interface ExecuteParams {
  grantee: string;
  messages: any[];
}

/**
 * Auth module parameters
 */
export interface AuthParams {
  maxMemoCharacters: number;
  txSigLimit: number;
  txSizeCostPerByte: number;
  sigVerifyCostED25519: number;
  sigVerifyCostSecp256k1: number;
}
