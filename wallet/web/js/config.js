/**
 * Aura Web Wallet Configuration
 * Chain settings and API endpoints for Aura network
 */
const AuraConfig = {
  // Chain identification
  chainId: 'aura-mvp-1',
  chainName: 'Aura',
  bech32Prefix: 'aura',
  slip44: 118,

  // Token configuration
  coin: {
    base: 'uaura',
    display: 'aura',
    symbol: 'AURA',
    exponent: 6
  },

  // Gas settings
  gasPrice: {
    low: 0.015,
    average: 0.025,
    high: 0.04
  },
  defaultGas: '200000',

  // API endpoints - configurable for different environments
  endpoints: {
    rest: 'http://localhost:1317',
    rpc: 'http://localhost:26657'
  },

  // Storage keys
  storageKeys: {
    wallet: 'aura_wallet_data',
    settings: 'aura_wallet_settings',
    history: 'aura_tx_history'
  },

  // UI settings
  ui: {
    refreshInterval: 30000, // 30 seconds
    maxHistoryItems: 50,
    addressTruncateLength: 12
  }
};

// Make config immutable
Object.freeze(AuraConfig);
Object.freeze(AuraConfig.coin);
Object.freeze(AuraConfig.gasPrice);
Object.freeze(AuraConfig.endpoints);
Object.freeze(AuraConfig.storageKeys);
Object.freeze(AuraConfig.ui);

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
  module.exports = AuraConfig;
}
