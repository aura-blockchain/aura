import WalletConnectService from '../src/services/WalletConnectService';
import TransactionService from '../src/services/TransactionService';
import AsyncStorage from '@react-native-async-storage/async-storage';
import SignClient from '@walletconnect/sign-client';

const handlers = {};
const mockClient = {
  connect: jest.fn(() => ({ uri: 'wc:test', approval: Promise.resolve({ topic: 'topic-1' }) })),
  on: jest.fn((event, cb) => { handlers[event] = cb; }),
  respond: jest.fn(),
  session: { get: jest.fn(() => ({ topic: 'topic-1' })) },
  disconnect: jest.fn(() => Promise.resolve()),
};

jest.mock('@walletconnect/sign-client', () => {
  const defaultExport = {
    init: jest.fn(() => Promise.resolve(mockClient)),
    __client: mockClient,
    __handlers: handlers,
  };
  return { __esModule: true, default: defaultExport };
});

jest.mock('@react-native-async-storage/async-storage', () => ({
  __esModule: true,
  default: {
    setItem: jest.fn(() => Promise.resolve()),
    getItem: jest.fn(() => Promise.resolve(null)),
    removeItem: jest.fn(() => Promise.resolve()),
  },
}));

jest.mock('../src/services/TransactionService', () => ({
  __esModule: true,
  default: {
    signDirect: jest.fn(),
    signAmino: jest.fn(),
  },
}));

describe('WalletConnectService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    handlers.session_request = undefined;
    mockClient.session.get.mockReturnValue({ topic: 'topic-1' });
    AsyncStorage.getItem.mockResolvedValue(null);
    WalletConnectService.client = null;
    WalletConnectService.session = null;
    WalletConnectService.clearSigner();
  });

  it('persists session after connect approval', async () => {
    WalletConnectService.setSigner({ address: 'aura1abc', privateKeyHex: 'deadbeef' });
    const { uri } = await WalletConnectService.connect();
    await Promise.resolve(); // wait for approval
    expect(uri).toBe('wc:test');
    expect(WalletConnectService.session?.topic).toBe('topic-1');
    expect(AsyncStorage.setItem).toHaveBeenCalledWith(
      'aura_wcv2_session',
      JSON.stringify({ topic: 'topic-1' }),
    );
  });

  it('restores persisted session', async () => {
    AsyncStorage.getItem.mockResolvedValue(JSON.stringify({ topic: 'topic-restore' }));
    mockClient.session.get.mockReturnValue({ topic: 'topic-restore' });
    WalletConnectService.setSigner({ address: 'aura1restore', privateKeyHex: 'cafebabe' });
    const restored = await WalletConnectService.restore();
    expect(restored?.topic).toBe('topic-restore');
    expect(WalletConnectService.session?.topic).toBe('topic-restore');
  });

  it('handles signDirect request with configured signer', async () => {
    WalletConnectService.setSigner({ address: 'aura1signed', privateKeyHex: 'feedface' });
    TransactionService.signDirect.mockReturnValue({ signed: {}, signature: { signature: 'sig' } });
    await WalletConnectService.init();
    const handler = SignClient.__handlers.session_request;
    expect(handler).toBeDefined();

    await handler({
      topic: 'topic-1',
      id: 7,
      params: {
        chainId: 'cosmos:aura-testnet-1',
        request: {
          method: 'cosmos_signDirect',
          params: {
            signerAddress: 'aura1signed',
            signDoc: { chain_id: 'aura-testnet-1', account_number: '1', sequence: '0', fee: {}, msgs: [] },
          },
        },
      },
    });

    expect(TransactionService.signDirect).toHaveBeenCalled();
    expect(mockClient.respond).toHaveBeenCalled();
  });

  it('rejects request when signer address mismatches', async () => {
    WalletConnectService.setSigner({ address: 'aura1me', privateKeyHex: 'abcd' });
    await WalletConnectService.init();
    const handler = SignClient.__handlers.session_request;
    await handler({
      topic: 'topic-1',
      id: 8,
      params: {
        chainId: 'cosmos:aura-testnet-1',
        request: { method: 'cosmos_signDirect', params: { signerAddress: 'aura1other', signDoc: {} } },
      },
    });
    expect(mockClient.respond).toHaveBeenCalled();
    const lastResponse = mockClient.respond.mock.calls.pop()?.[0];
    expect(lastResponse?.response?.error?.message).toMatch(/mismatch/i);
  });
});
