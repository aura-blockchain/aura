export const DEFAULTS = {
  restEndpoint: process.env.AURA_REST_ENDPOINT || 'http://localhost:1317',
  rpcEndpoint: process.env.AURA_RPC_ENDPOINT || 'http://localhost:26657',
  chainId: process.env.AURA_CHAIN_ID || 'aura-localnet',
  denom: 'uaura',
  displayDenom: 'AURA',
  bech32Prefix: 'aura',
  decimals: 6,
  gasPrice: process.env.AURA_GAS_PRICE || '0.025uaura',
};

export function getFee(gasLimit, gasPrice = DEFAULTS.gasPrice) {
  const match = gasPrice.match(/^([0-9.]+)([a-zA-Z/]+)$/);
  if (!match) {
    throw new Error(`Invalid gas price format: ${gasPrice}`);
  }
  const [, priceAmount, priceDenom] = match;
  const feeAmount = (Number(priceAmount) * gasLimit).toFixed(0);
  return {
    amount: [{ denom: priceDenom, amount: feeAmount }],
    gas: gasLimit.toString(),
  };
}
