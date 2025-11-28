import { Coin } from '@cosmjs/stargate';

export interface AuraChainConfig {
  rpcEndpoint: string;
  restEndpoint?: string;
  chainId: string;
  prefix?: string;
  gasPrice?: string;
  gasAdjustment?: number;
}

export interface WalletAccount {
  address: string;
  pubkey: Uint8Array;
  algo: string;
}

export interface Pool {
  id: string;
  tokenA: string;
  tokenB: string;
  reserveA: string;
  reserveB: string;
  totalShares: string;
  swapFee: string;
}

export interface PoolParams {
  tokenA: string;
  tokenB: string;
  amountA: string;
  amountB: string;
}

export interface SwapParams {
  poolId: string;
  tokenIn: string;
  amountIn: string;
  minAmountOut: string;
  recipient?: string;
}

export interface AddLiquidityParams {
  poolId: string;
  amountA: string;
  amountB: string;
  minShares: string;
}

export interface RemoveLiquidityParams {
  poolId: string;
  shares: string;
  minAmountA: string;
  minAmountB: string;
}

export interface Validator {
  operatorAddress: string;
  consensusPubkey: string;
  jailed: boolean;
  status: number;
  tokens: string;
  delegatorShares: string;
  description: {
    moniker: string;
    identity: string;
    website: string;
    securityContact: string;
    details: string;
  };
  commission: {
    rate: string;
    maxRate: string;
    maxChangeRate: string;
  };
}

export interface DelegateParams {
  validatorAddress: string;
  amount: string;
  denom?: string;
}

export interface UndelegateParams {
  validatorAddress: string;
  amount: string;
  denom?: string;
}

export interface RedelegateParams {
  srcValidatorAddress: string;
  dstValidatorAddress: string;
  amount: string;
  denom?: string;
}

export interface Proposal {
  proposalId: string;
  content: {
    typeUrl: string;
    value: Uint8Array;
  };
  status: number;
  finalTallyResult: {
    yes: string;
    abstain: string;
    no: string;
    noWithVeto: string;
  };
  submitTime: Date;
  depositEndTime: Date;
  totalDeposit: Coin[];
  votingStartTime: Date;
  votingEndTime: Date;
}

export interface VoteParams {
  proposalId: string;
  option: VoteOption;
  metadata?: string;
}

export interface DepositParams {
  proposalId: string;
  amount: string;
  denom?: string;
}

export enum VoteOption {
  UNSPECIFIED = 0,
  YES = 1,
  ABSTAIN = 2,
  NO = 3,
  NO_WITH_VETO = 4,
}

export interface TxResult {
  transactionHash: string;
  height: number;
  code: number;
  rawLog?: string;
  gasUsed: number;
  gasWanted: number;
}

export interface QueryBalance {
  denom: string;
  amount: string;
}

export interface SendParams {
  recipient: string;
  amount: string;
  denom?: string;
  memo?: string;
}

export interface GasOptions {
  gasLimit?: number;
  gasPrice?: string;
  memo?: string;
}

// Export all module-specific types
export * from './auth';
export * from './bridge';
export * from './compliance';
export * from './confidence-score';
export * from './cryptography';
export * from './data-registry';
export * from './economic-security';
export * from './identity-change';
export * from './inclusion-routines';
export * from './monitoring';
export * from './network-security';
export * from './prevalidation';
export * from './privacy';
export * from './validator-security';
export * from './vc-registry';
export * from './wallet-security';
