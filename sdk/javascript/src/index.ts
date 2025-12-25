// Main exports
export { AuraClient } from './client';
export { AuraWallet } from './wallet';
export { TxBuilder } from './tx';

// Module exports
export { BankModule } from './modules/bank';
export { DexModule } from './modules/dex';
export { StakingModule } from './modules/staking';
export { GovernanceModule } from './modules/governance';

// Type exports
export * from './types';

// Error exports
export * from './errors';

// Event exports
export * from './events';

// Batch operation exports
export * from './batch';

// Re-export commonly used CosmJS types
export type { Coin } from '@cosmjs/stargate';
export type { OfflineDirectSigner } from '@cosmjs/proto-signing';
