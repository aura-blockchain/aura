// Shared Aura chain constants consumed by all wallets (desktop, mobile, extension).
// Derived from docs/chain-registry/aura.json (primary) or wallet/config/chain.json (local fallback) to keep dApps and wallets aligned.
let registry;
try {
  // Prefer canonical chain-registry metadata if present
  // eslint-disable-next-line import/no-dynamic-require, global-require
  registry = require('../../docs/chain-registry/aura.json');
} catch (err) {
  // eslint-disable-next-line import/no-dynamic-require, global-require
  registry = require('./chain.json');
}

const feeToken = registry.fees?.fee_tokens?.[0] || {};
const GAS_PRICE_TIERS = {
  low: feeToken.low_gas_price ?? 0.015,
  average: feeToken.average_gas_price ?? feeToken.fixed_min_gas_price ?? 0.025,
  high: feeToken.high_gas_price ?? 0.04
};

const COIN = {
  base: registry.assets?.[0]?.base || 'uaura',
  display: registry.assets?.[0]?.display || 'aura',
  symbol: registry.assets?.[0]?.symbol || 'AURA',
  exponent: registry.assets?.[0]?.display_exponent || 6
};

const RPC_ENDPOINTS = registry.apis?.rpc || [];
const REST_ENDPOINTS = registry.apis?.rest || [];

const CHAIN_CONFIG = {
  chainId: registry.chain_id,
  chainName: registry.pretty_name || 'Aura',
  bech32Prefix: registry.bech32_prefix,
  slip44: registry.slip44,
  coin: COIN,
  gasPrice: GAS_PRICE_TIERS.average,
  gasPriceTiers: GAS_PRICE_TIERS,
  rpc: RPC_ENDPOINTS,
  rest: REST_ENDPOINTS
};

module.exports = {
  CHAIN_CONFIG,
  GAS_PRICE_TIERS,
  COIN,
  RPC_ENDPOINTS,
  REST_ENDPOINTS
};
