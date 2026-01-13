import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../config/chain', () => ({
  CHAIN_CONFIG: { chainId: 'aura-mvp-1', chainName: 'Aura', bech32Prefix: 'aura', slip44: 118 },
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

const wcMock = {
  connect: vi.fn().mockResolvedValue({ uri: 'wc:test-uri', approval: Promise.resolve({ topic: 't' }) }),
  on: vi.fn(),
  respond: vi.fn(),
};

vi.mock('@walletconnect/sign-client', () => ({
  __esModule: true,
  default: {
    init: vi.fn().mockResolvedValue(wcMock)
  }
}), { virtual: true });

describe('WalletConnect integration button', () => {
  let popup;
  let uriEl;
  let statusEl;

  beforeEach(async () => {
    vi.resetModules();
    uriEl = { textContent: '' };
    statusEl = { textContent: '', classList: { toggle: () => {} } };
    global.document = {
      querySelector: vi.fn((sel) => {
        if (sel === '#walletConnectUri') return uriEl;
        if (sel === '#walletConnectStatus') return statusEl;
        return null;
      }),
      addEventListener: vi.fn((evt, handler) => {
        if (evt === 'DOMContentLoaded') {
          handler();
        }
      })
    };
    global.window = {};
    global.chrome = { storage: { local: { get: (_keys, cb) => cb({}), set: (_data, cb) => cb && cb() } } };
    popup = await import('../../popup.js');
  });

  it('generates a WalletConnect URI and updates status', async () => {
    await popup.initiateWalletConnect();
    expect(uriEl.textContent).toBe('wc:test-uri');
  });
});
