/**
 * ApiService - Bank Transaction Tests
 * Validates sendTokens happy/error paths for the desktop wallet.
 */

import { ApiService } from '../src/services/api';
import { SigningStargateClient, GasPrice } from '@cosmjs/stargate';
import { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';

jest.mock('@cosmjs/stargate', () => ({
  SigningStargateClient: {
    connectWithSigner: jest.fn()
  },
  GasPrice: {
    fromString: jest.fn()
  }
}));

jest.mock('@cosmjs/proto-signing', () => {
  const actual = jest.requireActual('@cosmjs/proto-signing');
  return {
    ...actual,
    DirectSecp256k1HdWallet: {
      fromMnemonic: jest.fn()
    }
  };
});

describe('ApiService transaction flows', () => {
  let apiService;
  let mockClient;

  beforeEach(async () => {
    jest.clearAllMocks();
    await window.electron.store.clear();

    mockClient = {
      sendTokens: jest.fn(),
      delegateTokens: jest.fn(),
      undelegateTokens: jest.fn(),
      signAndBroadcast: jest.fn(),
      withdrawRewards: jest.fn()
    };

    GasPrice.fromString.mockReturnValue({ amount: '0.025', denom: 'uaura' });
    SigningStargateClient.connectWithSigner.mockResolvedValue(mockClient);
    DirectSecp256k1HdWallet.fromMnemonic.mockResolvedValue({ getAccounts: jest.fn() });

    apiService = new ApiService();
  });

  describe('Bank Send Transactions', () => {
    test('should broadcast a send transaction with converted amount and default fees', async () => {
      const mockResult = {
        code: 0,
        transactionHash: 'TXHASH123',
        gasUsed: 90000,
        gasWanted: 200000
      };

      mockClient.sendTokens.mockResolvedValue(mockResult);

      const result = await apiService.sendTokens(
        'aura1senderaddress',
        'aura1recipientaddress',
        1500000,
        'uaura',
        'Test payment',
        'sample mnemonic phrase for testing'
      );

      expect(DirectSecp256k1HdWallet.fromMnemonic).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({ prefix: 'aura' })
      );
      expect(GasPrice.fromString).toHaveBeenCalledWith('0.025uaura');
      expect(SigningStargateClient.connectWithSigner).toHaveBeenCalledWith(
        'http://localhost:26657',
        expect.any(Object),
        expect.objectContaining({ gasPrice: { amount: '0.025', denom: 'uaura' } })
      );
      expect(mockClient.sendTokens).toHaveBeenCalledWith(
        'aura1senderaddress',
        'aura1recipientaddress',
        [{ denom: 'uaura', amount: '1500000' }],
        expect.objectContaining({
          amount: expect.arrayContaining([expect.objectContaining({ denom: 'uaura', amount: '5000' })]),
          gas: '200000'
        }),
        'Test payment'
      );
      expect(result).toEqual(mockResult);
    });

    test('should surface broadcast failures from the bank send path', async () => {
      mockClient.sendTokens.mockRejectedValue(new Error('insufficient funds'));

      await expect(
        apiService.sendTokens(
          'aura1senderaddress',
          'aura1recipientaddress',
          250000,
          'uaura',
          '',
          'sample mnemonic'
        )
      ).rejects.toThrow('insufficient funds');
    });
  });

  describe('Staking & Governance Transactions', () => {
    test('delegation path enforces signing key and relays to staking client', async () => {
      const mockResult = { transactionHash: 'ABC123', code: 0 };
      mockClient.delegateTokens.mockResolvedValue(mockResult);

      const result = await apiService.delegate(
        'aura1delegator',
        'auravaloper1node',
        250000,
        'uaura',
        '',
        'mnemonic words here'
      );

      expect(mockClient.delegateTokens).toHaveBeenCalledWith(
        'aura1delegator',
        'auravaloper1node',
        { denom: 'uaura', amount: '250000' },
        expect.any(Object),
        ''
      );
      expect(result).toEqual(mockResult);
    });

    test('delegate rejects missing signing key', async () => {
      await expect(
        apiService.delegate('aura1delegator', 'auravaloper1node', 1, 'uaura', '', '')
      ).rejects.toThrow('Missing signing key');
    });

    test('redelegate prevents identical validators', async () => {
      await expect(
        apiService.redelegate(
          'aura1delegator',
          'auravaloper1node',
          'auravaloper1node',
          1,
          'uaura',
          '',
          'mnemonic'
        )
      ).rejects.toThrow('Source and destination validators must be different');
    });

    test('vote rejects invalid option and enforces voter address', async () => {
      await expect(
        apiService.vote('aura1voter', 1, 'maybe', '', 'mnemonic phrase')
      ).rejects.toThrow('Invalid vote option. Must be: yes, abstain, no, or no_with_veto');

      await expect(
        apiService.vote('invalid', 1, 'yes', '', 'mnemonic')
      ).rejects.toThrow('Invalid voter address');
    });

    test('vote relays signAndBroadcast with normalized option', async () => {
      mockClient.signAndBroadcast.mockResolvedValue({ code: 0, txhash: 'VOTE123' });

      const result = await apiService.vote('aura1valid', 5, 'YES', 'approve', 'mnemonic words');

      expect(mockClient.signAndBroadcast).toHaveBeenCalledWith(
        'aura1valid',
        [
          expect.objectContaining({
            typeUrl: '/cosmos.gov.v1beta1.MsgVote',
            value: expect.objectContaining({
              proposalId: BigInt(5),
              option: 1
            })
          })
        ],
        expect.any(Object),
        'approve'
      );
      expect(result).toEqual({ code: 0, txhash: 'VOTE123' });
    });

    test('withdraw rewards path validates addresses before broadcast', async () => {
      await expect(
        apiService.withdrawRewards('invalid', 'auravaloper1', '', 'mnemonic')
      ).rejects.toThrow('Invalid delegator address');

      mockClient.withdrawRewards.mockResolvedValue({ code: 0, txhash: 'REWARD' });

      const result = await apiService.withdrawRewards(
        'aura1delegator',
        'auravaloper1',
        'memo',
        'mnemonic words'
      );

      expect(mockClient.withdrawRewards).toHaveBeenCalledWith(
        'aura1delegator',
        'auravaloper1',
        expect.any(Object),
        'memo'
      );
      expect(result).toEqual({ code: 0, txhash: 'REWARD' });
    });
  });

  describe('DEX Transactions', () => {
    test('createDexPool broadcasts correct message', async () => {
      const response = { code: 0, txhash: 'POOL123' };
      mockClient.signAndBroadcast.mockResolvedValue(response);

      const result = await apiService.createDexPool(
        'aura1creator',
        'uaura',
        'usdt',
        '1000000',
        '500000',
        'create pool memo',
        'sample mnemonic words'
      );

      expect(mockClient.signAndBroadcast).toHaveBeenCalledWith(
        'aura1creator',
        [
          expect.objectContaining({
            typeUrl: '/aura.dex.v1beta1.MsgCreatePool',
            value: expect.objectContaining({
              creator: 'aura1creator',
              denomA: 'uaura',
              denomB: 'usdt',
              amountA: { denom: 'uaura', amount: '1000000' },
              amountB: { denom: 'usdt', amount: '500000' }
            })
          })
        ],
        expect.objectContaining({ gas: '500000' }),
        'create pool memo'
      );
      expect(result).toEqual(response);
    });

    test('swapDexExactIn uses default minAmountOut and clamps slippage', async () => {
      mockClient.signAndBroadcast.mockResolvedValue({ code: 0 });

      await apiService.swapDexExactIn(
        'aura1trader',
        '42',
        'uaura',
        1250000,
        null,
        75,
        'swap memo',
        'mnemonic words here'
      );

      expect(mockClient.signAndBroadcast).toHaveBeenCalledWith(
        'aura1trader',
        [
          expect.objectContaining({
            typeUrl: '/aura.dex.v1beta1.MsgSwapExactIn',
            value: expect.objectContaining({
              poolId: '42',
              coinIn: { denom: 'uaura', amount: '1250000' },
              minAmountOut: '1250000',
              maxSlippageBps: 75
            })
          })
        ],
        expect.objectContaining({ gas: '300000' }),
        'swap memo'
      );
    });

    test('commitDexOrder converts commit hash to bytes', async () => {
      mockClient.signAndBroadcast.mockResolvedValue({ code: 0 });

      await apiService.commitDexOrder(
        'aura1sender',
        '0xaabbccdd',
        'commit memo',
        'words words words'
      );

      expect(mockClient.signAndBroadcast).toHaveBeenCalledWith(
        'aura1sender',
        [
          expect.objectContaining({
            typeUrl: '/aura.dex.v1beta1.MsgCommitOrder',
            value: expect.objectContaining({
              sender: 'aura1sender',
              commitHash: new Uint8Array([0xaa, 0xbb, 0xcc, 0xdd])
            })
          })
        ],
        expect.objectContaining({ gas: '150000' }),
        'commit memo'
      );
    });

    test('createDexHTLC builds correct payload and enforces recipients', async () => {
      const response = { code: 0, txhash: 'HTLC123' };
      mockClient.signAndBroadcast.mockResolvedValue(response);

      const result = await apiService.createDexHTLC(
        'aura1sender',
        'aura1recipient',
        'uaura',
        750000,
        'secret-hash',
        3600n,
        'lock memo',
        'mnemonic words'
      );

      expect(mockClient.signAndBroadcast).toHaveBeenCalledWith(
        'aura1sender',
        [
          expect.objectContaining({
            typeUrl: '/aura.dex.v1beta1.MsgCreateHTLC',
            value: expect.objectContaining({
              sender: 'aura1sender',
              recipient: 'aura1recipient',
              amount: { denom: 'uaura', amount: '750000' },
              secretHash: 'secret-hash',
              timelockDuration: 3600n
            })
          })
        ],
        expect.objectContaining({ gas: '320000' }),
        'lock memo'
      );
      expect(result).toEqual(response);

      await expect(
        apiService.createDexHTLC('invalid', 'aura1rcpt', 'uaura', 1, 'hash', 10n, '', 'mnemonic')
      ).rejects.toThrow('Invalid sender address');
    });

    test('claimDexHTLC validates identifiers and secret', async () => {
      mockClient.signAndBroadcast.mockResolvedValue({ code: 0 });

      await apiService.claimDexHTLC(
        'aura1recipient',
        42,
        'unlock-secret',
        'claim memo',
        'mnemonic words'
      );

      expect(mockClient.signAndBroadcast).toHaveBeenCalledWith(
        'aura1recipient',
        [
          expect.objectContaining({
            typeUrl: '/aura.dex.v1beta1.MsgClaimHTLC',
            value: expect.objectContaining({
              recipient: 'aura1recipient',
              htlcId: '42',
              secret: 'unlock-secret'
            })
          })
        ],
        expect.objectContaining({ gas: '200000' }),
        'claim memo'
      );

      await expect(
        apiService.claimDexHTLC('invalid', '', '', '', 'mnemonic')
      ).rejects.toThrow('Invalid recipient address');
    });

    test('refundDexHTLC enforces sender and HTLC id', async () => {
      mockClient.signAndBroadcast.mockResolvedValue({ code: 0 });

      await apiService.refundDexHTLC(
        'aura1sender',
        99,
        'refund memo',
        'mnemonic words'
      );

      expect(mockClient.signAndBroadcast).toHaveBeenCalledWith(
        'aura1sender',
        [
          expect.objectContaining({
            typeUrl: '/aura.dex.v1beta1.MsgRefundHTLC',
            value: expect.objectContaining({
              sender: 'aura1sender',
              htlcId: '99'
            })
          })
        ],
        expect.objectContaining({ gas: '200000' }),
        'refund memo'
      );

      await expect(
        apiService.refundDexHTLC('', null, '', 'mnemonic words')
      ).rejects.toThrow('Invalid sender address');
    });
  });
});
