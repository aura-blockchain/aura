import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../config/chain', () => ({
  CHAIN_CONFIG: { chainId: 'aura-testnet-1', chainName: 'Aura', bech32Prefix: 'aura', slip44: 118 },
  COIN: { base: 'uaura', display: 'aura', symbol: 'AURA', exponent: 6 },
  GAS_PRICE_TIERS: { low: 0.015, average: 0.025, high: 0.04 },
  REST_ENDPOINTS: [{ address: 'http://localhost:1317' }],
  RPC_ENDPOINTS: [{ address: 'http://localhost:26657' }]
}), { virtual: true });

vi.mock('../../hardware-wallet', () => {
  function HardwareWalletManager() {}
  HardwareWalletManager.prototype.isConnected = () => false;
  const exported = () => {};
  exported.default = HardwareWalletManager;
  exported.HardwareWalletManager = HardwareWalletManager;
  return exported;
}, { virtual: true });

describe('Keplr registration flow', () => {
  let suggestSpy;
  let popup;

  beforeEach(async () => {
    vi.resetModules();
    suggestSpy = vi.fn();
    global.window = {
      keplr: {
        experimentalSuggestChain: suggestSpy,
        enable: vi.fn(),
        getKey: vi.fn(),
        signAmino: vi.fn(),
        signDirect: vi.fn(),
      },
      getOfflineSignerAuto: vi.fn(),
    };
    global.document = {
      querySelector: vi.fn(() => null),
      addEventListener: vi.fn((evt, handler) => {
        if (evt === 'DOMContentLoaded') {
          handler();
        }
      })
    };
    global.chrome = { storage: { local: { get: (_keys, cb) => cb({}), set: (_data, cb) => cb && cb() } } };
    popup = await import('../../popup.js');
  });

  it('exposes auraKeplrProvider and suggests chain on handler call', async () => {
    expect(global.window.auraKeplrProvider).toBeDefined();
    await popup.suggestKeplrChain();
    expect(suggestSpy).toHaveBeenCalled();
  });
});
