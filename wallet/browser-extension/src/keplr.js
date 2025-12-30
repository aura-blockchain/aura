const { CHAIN_CONFIG, COIN, GAS_PRICE_TIERS, RPC_ENDPOINTS, REST_ENDPOINTS } = require('../../config/chain');

function buildKeplrChainInfo() {
  return {
    chainId: CHAIN_CONFIG.chainId,
    chainName: CHAIN_CONFIG.chainName || 'Aura',
    rpc: RPC_ENDPOINTS[0]?.address || 'http://localhost:26657',
    rest: REST_ENDPOINTS[0]?.address || 'http://localhost:1317',
    bip44: {
      coinType: CHAIN_CONFIG.slip44 || 118,
    },
    bech32Config: {
      bech32PrefixAccAddr: CHAIN_CONFIG.bech32Prefix,
      bech32PrefixAccPub: `${CHAIN_CONFIG.bech32Prefix}pub`,
      bech32PrefixValAddr: `${CHAIN_CONFIG.bech32Prefix}valoper`,
      bech32PrefixValPub: `${CHAIN_CONFIG.bech32Prefix}valoperpub`,
      bech32PrefixConsAddr: `${CHAIN_CONFIG.bech32Prefix}valcons`,
      bech32PrefixConsPub: `${CHAIN_CONFIG.bech32Prefix}valconspub`,
    },
    currencies: [
      {
        coinDenom: COIN.symbol || 'AURA',
        coinMinimalDenom: COIN.base,
        coinDecimals: COIN.exponent || 6,
      },
    ],
    feeCurrencies: [
      {
        coinDenom: COIN.symbol || 'AURA',
        coinMinimalDenom: COIN.base,
        coinDecimals: COIN.exponent || 6,
        gasPriceStep: {
          low: GAS_PRICE_TIERS.low || 0.015,
          average: GAS_PRICE_TIERS.average || 0.025,
          high: GAS_PRICE_TIERS.high || 0.04,
        },
      },
    ],
    stakeCurrency: {
      coinDenom: COIN.symbol || 'AURA',
      coinMinimalDenom: COIN.base,
      coinDecimals: COIN.exponent || 6,
    },
    features: ['stargate', 'ibc-transfer'],
  };
}

export { buildKeplrChainInfo };
