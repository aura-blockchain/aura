import { describe, it, expect, vi, beforeEach } from 'vitest';

const sessionObj = { topic: 'test-topic', relay: { protocol: 'irn' } };
const handlers = {};
const wcMock = {
  connect: vi.fn().mockResolvedValue({ uri: 'wc:test', approval: Promise.resolve(sessionObj) }),
  on: vi.fn((event, cb) => { handlers[event] = cb; }),
  respond: vi.fn(),
  session: {
    get: vi.fn().mockReturnValue(sessionObj),
  },
  disconnect: vi.fn().mockResolvedValue(),
};

vi.mock('@walletconnect/sign-client', () => ({
  __esModule: true,
  default: {
    init: vi.fn().mockResolvedValue(wcMock)
  }
}), { virtual: true });

describe('walletconnect module', () => {
  beforeEach(() => {
    vi.resetModules();
    const store = {};
    global.localStorage = {
      getItem: vi.fn((k) => store[k] || null),
      setItem: vi.fn((k, v) => { store[k] = v; }),
      removeItem: vi.fn((k) => { delete store[k]; }),
    };
    global.chrome = undefined;
  });

  it('persists session after approval', async () => {
    const mod = await import('../../src/walletconnect.js');
    await mod.connectWalletConnect();
    await Promise.resolve(); // allow approval to resolve
    expect(mod.getWalletConnectSession()).toEqual(sessionObj);
    expect(localStorage.setItem).toHaveBeenCalled();
  });

  it('restores session from storage', async () => {
    const mod = await import('../../src/walletconnect.js');
    // seed storage
    localStorage.setItem('aura_wcv2_session', JSON.stringify(sessionObj));
    const restored = await mod.restoreWalletConnectSession();
    expect(restored).toEqual(sessionObj);
    expect(mod.getWalletConnectSession()).toEqual(sessionObj);
    expect(wcMock.session.get).toHaveBeenCalledWith('test-topic');
  });

  it('disconnect clears session and storage', async () => {
    const mod = await import('../../src/walletconnect.js');
    await mod.connectWalletConnect();
    await Promise.resolve();
    expect(mod.getWalletConnectSession()).not.toBeNull();
    await mod.disconnectWalletConnect();
    expect(mod.getWalletConnectSession()).toBeNull();
    expect(localStorage.removeItem).toHaveBeenCalled();
  });

  it('responds with error when signer rejects signDirect', async () => {
    const mod = await import('../../src/walletconnect.js');
    const provider = { signDirect: vi.fn().mockRejectedValue(new Error('no signer')) };
    mod.setWalletConnectProvider(provider);
    await mod.connectWalletConnect();
    await Promise.resolve();
    const handler = handlers.session_request;
    expect(handler).toBeDefined();
    await handler({
      topic: 't',
      id: 1,
      params: {
        chainId: 'cosmos:aura-testnet-1',
        request: { method: 'cosmos_signDirect', params: { signerAddress: 'aura1', signDoc: {} } },
      },
    });
    expect(wcMock.respond).toHaveBeenCalled();
    const resp = wcMock.respond.mock.calls.pop()[0];
    expect(resp.response.error.message).toContain('no signer');
  });

  it('responds with result when signer resolves', async () => {
    const mod = await import('../../src/walletconnect.js');
    const provider = { signDirect: vi.fn().mockResolvedValue({ ok: true }) };
    mod.setWalletConnectProvider(provider);
    await mod.connectWalletConnect();
    await Promise.resolve();
    const handler = handlers.session_request;
    await handler({
      topic: 't',
      id: 2,
      params: {
        chainId: 'cosmos:aura-testnet-1',
        request: { method: 'cosmos_signDirect', params: { signerAddress: 'aura1', signDoc: {} } },
      },
    });
    const resp = wcMock.respond.mock.calls.pop()[0];
    expect(resp.response.result.ok).toBe(true);
  });

  it('updates status when unsupported method', async () => {
    const mod = await import('../../src/walletconnect.js');
    mod.onWalletConnectStatus((msg) => { handlers.lastStatus = msg; });
    await mod.connectWalletConnect();
    await Promise.resolve();
    const handler = handlers.session_request;
    await handler({
      topic: 't',
      id: 3,
      params: {
        chainId: 'cosmos:aura-testnet-1',
        request: { method: 'unknown_method', params: {} },
      },
    });
    expect(wcMock.respond).toHaveBeenCalled();
    expect(handlers.lastStatus).toContain('Rejected unsupported');
  });
});
