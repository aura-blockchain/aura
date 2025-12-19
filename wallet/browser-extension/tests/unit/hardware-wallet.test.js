import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Create mock objects before vi.mock calls
const mockTransport = {
  device: { productName: 'Ledger', manufacturerName: 'Ledger' },
  close: vi.fn(),
};

const mockLedgerApp = {
  getAddress: vi.fn().mockResolvedValue({
    bech32_address: 'aura1hwaddress',
    publicKey: new Uint8Array([1, 2, 3]),
  }),
  sign: vi.fn().mockResolvedValue({
    signature: new Uint8Array([4, 5, 6]),
  }),
};

const mockKeystoneSdk = {
  getConnectedWalletAddress: vi.fn().mockReturnValue({
    address: 'aura1keystone',
    publicKey: 'pub_keystone',
  }),
  sign: vi.fn().mockResolvedValue({
    signature: 'keystone_sig',
  }),
};

const trezorMock = {
  manifest: vi.fn(),
  cosmosGetAddress: vi.fn().mockResolvedValue({
    success: true,
    payload: {
      address: 'aura1trezor',
      publicKey: new Uint8Array([9, 9, 9]),
    },
  }),
  cosmosSignTx: vi.fn().mockResolvedValue({
    success: true,
    payload: {
      signature: 'deadbeef',
    },
  }),
  getAddress: vi.fn(),
  signTransaction: vi.fn(),
};

vi.mock('@ledgerhq/hw-transport-webhid', () => ({
  default: {
    create: vi.fn().mockResolvedValue({
      device: { productName: 'Ledger', manufacturerName: 'Ledger' },
      close: vi.fn(),
    }),
    list: vi.fn().mockResolvedValue([{ productName: 'Ledger' }]),
  },
}));

vi.mock('@ledgerhq/hw-app-cosmos', () => ({
  default: class MockCosmosApp {
    constructor() {}
    getAddress = vi.fn().mockResolvedValue({
      bech32_address: 'aura1hwaddress',
      publicKey: new Uint8Array([1, 2, 3]),
    });
    sign = vi.fn().mockResolvedValue({
      signature: new Uint8Array([4, 5, 6]),
    });
  },
}));

vi.mock('@keystonehq/keystone-sdk', () => ({
  default: class MockKeystoneSDK {
    constructor() {}
    getConnectedWalletAddress = vi.fn().mockReturnValue({
      address: 'aura1keystone',
      publicKey: 'pub_keystone',
    });
    sign = vi.fn().mockResolvedValue({
      signature: 'keystone_sig',
    });
  },
}));

vi.mock('trezor-connect', () => ({ default: trezorMock }));

describe('HardwareWalletManager (ledgerjs)', () => {
  let HardwareWalletManager;

  beforeEach(async () => {
    vi.resetModules();
    ({ default: HardwareWalletManager } = await import('../../hardware-wallet.js'));
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('detects ledger devices via webhid', async () => {
    const mgr = new HardwareWalletManager();
    const devices = await mgr.detectDevices();
    expect(devices.length).toBe(1);
    expect(devices[0].type).toBe('ledger');
  });

  it('connects and fetches address', async () => {
    const mgr = new HardwareWalletManager();
    await mgr.connect();
    const addr = await mgr.getAddress("m/44'/118'/0'/0/0", false);
    expect(addr.address).toBe('aura1hwaddress');
    expect(addr.publicKey).toBe('010203');
    expect(mgr.isConnected()).toBe(true);
  });

  it('signs bytes and returns hex signature', async () => {
    const mgr = new HardwareWalletManager();
    await mgr.connect();
    const signed = await mgr.signWithLedger(new Uint8Array([1, 2, 3]), "m/44'/118'/0'/0/0");
    expect(signed.signature).toBe('040506');
  });

  it('throws for unsupported device types', async () => {
    const mgr = new HardwareWalletManager();
    mgr.connectedDevice = { type: 'unknown' };
    await expect(mgr.getAddress()).rejects.toThrow();
  });

  it('connects to trezor via Connect and fetches address', async () => {
    const mgr = new HardwareWalletManager();
    await mgr.connect('trezor');
    const addr = await mgr.getTrezorAddress("m/44'/118'/0'/0/0");
    expect(addr.address).toBe('aura1trezor');
    expect(addr.publicKey).toBe('090909');
    const pathUsed = trezorMock.cosmosGetAddress.mock.calls[1]?.[0]?.path || [];
    expect(pathUsed[0]).toBe(2147483692); // 44'
    expect(mgr.connectedDevice.type).toBe('trezor');
  });

  it('signs transactions with trezor using cosmosSignTx', async () => {
    const mgr = new HardwareWalletManager();
    await mgr.connect('trezor');
    const signed = await mgr.signWithTrezor(new Uint8Array([10, 11, 12]), "m/44'/118'/0'/0/1");
    expect(signed.signature).toBe('deadbeef');
    expect(trezorMock.cosmosSignTx).toHaveBeenCalled();
    const call = trezorMock.cosmosSignTx.mock.calls.pop()[0];
    expect(call.path[4]).toBe(1);
  });

  it('supports keystone address + signing', async () => {
    const mgr = new HardwareWalletManager();
    await mgr.connect('keystone');
    const addr = await mgr.getKeystoneAddress();
    expect(addr.address).toBe('aura1keystone');
    const sig = await mgr.signWithKeystone(new Uint8Array([1, 2, 3]));
    expect(sig.signature).toBe('keystone_sig');
  });
});
