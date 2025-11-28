import { Coin } from '@cosmjs/stargate';

/**
 * Bridge transfer status enum
 */
export enum BridgeTransferStatus {
  PENDING = 0,
  COMPLETED = 1,
  FAILED = 2,
  REFUNDED = 3,
}

/**
 * Bridge transfer information
 */
export interface BridgeTransfer {
  id: string;
  sender: string;
  recipient: string;
  amount: Coin;
  sourceChain: string;
  targetChain: string;
  status: BridgeTransferStatus;
  createdAt: Date;
  completedAt?: Date;
  proof?: string;
  txHash?: string;
  error?: string;
}

/**
 * Parameters for initiating a bridge transfer
 */
export interface InitiateBridgeParams {
  sender: string;
  recipient: string;
  amount: string;
  denom: string;
  targetChain: string;
  timeout?: number;
  memo?: string;
}

/**
 * Parameters for completing a bridge transfer
 */
export interface CompleteBridgeParams {
  transferId: string;
  proof: string;
  height: number;
  signatures: string[];
}

/**
 * Bridge parameters
 */
export interface BridgeParams {
  minTransferAmount: string;
  maxTransferAmount: string;
  supportedChains: string[];
  bridgeFee: string;
  confirmationsRequired: number;
  timeout: number;
  enabled: boolean;
}

/**
 * Bridge security configuration
 */
export interface BridgeSecurity {
  merkleRoot: string;
  validators: string[];
  requiredSignatures: number;
  lastUpdateHeight: number;
}

/**
 * Bridge statistics
 */
export interface BridgeStats {
  totalTransfers: number;
  totalVolume: string;
  activeTransfers: number;
  completedTransfers: number;
  failedTransfers: number;
}
