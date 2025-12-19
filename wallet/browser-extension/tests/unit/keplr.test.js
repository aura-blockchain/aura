import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../config/chain', () => ({
  CHAIN_CONFIG: { chainId: 'aura-testnet-1', chainName: 'Aura', bech32Prefix: 'aura', slip44: 118 },
  COIN: { base: 'uaura', display: 'aura', symbol: 'AURA', exponent: 6 },
  GAS_PRICE_TIERS: { low: 0.015, average: 0.025, high: 0.04 },
  REST_ENDPOINTS: [{ address: 'http://localhost:1317' }],
  RPC_ENDPOINTS: [{ address: 'http://localhost:26657' }]
}), { virtual: true });

describe('Keplr chain info builder', () => {
  let buildKeplrChainInfo;

  beforeEach(async () => {
    vi.resetModules();
    ({ buildKeplrChainInfo } = await import('../../src/keplr.js'));
  });

  it('produces expected chain metadata', () => {
    const info = buildKeplrChainInfo();
    expect(info.chainId).toBe('aura-testnet-1');
    expect(info.bech32Config.bech32PrefixAccAddr).toBe('aura');
    expect(info.currencies[0].coinMinimalDenom).toBe('uaura');
    expect(info.feeCurrencies[0].gasPriceStep.low).toBe(0.015);
  });
});
