/**
 * Staking Operations Tests
 * Tests for undelegate, withdraw rewards, redelegate, and vote functions
 */

import { ApiService } from '../src/services/api';
import { SigningStargateClient } from '@cosmjs/stargate';
import { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';

// Mock CosmJS
jest.mock('@cosmjs/stargate');
jest.mock('@cosmjs/proto-signing');

describe('ApiService - Staking Transactions', () => {
  let apiService;
  let mockClient;
  let mockWallet;

  beforeEach(async () => {
    jest.clearAllMocks();
    await window.electron.store.clear();

    apiService = new ApiService();

    // Mock wallet
    mockWallet = {
      getAccounts: jest.fn().mockResolvedValue([
        {
          address: 'aura1test',
          pubkey: new Uint8Array([1, 2, 3])
        }
      ])
    };
    DirectSecp256k1HdWallet.fromMnemonic = jest.fn().mockResolvedValue(mockWallet);

    // Mock client
    mockClient = {
      delegateTokens: jest.fn(),
      undelegateTokens: jest.fn(),
      withdrawRewards: jest.fn(),
      signAndBroadcast: jest.fn()
    };
    SigningStargateClient.connectWithSigner = jest.fn().mockResolvedValue(mockClient);
  });

  describe('undelegate', () => {
    test('should successfully undelegate tokens', async () => {
      const mockResult = {
        code: 0,
        transactionHash: 'ABC123',
        height: 1000,
        gasUsed: 150000,
        gasWanted: 200000
      };

      mockClient.undelegateTokens.mockResolvedValue(mockResult);

      const result = await apiService.undelegate(
        'aura1delegator',
        'auravaloper1validator',
        1000000,
        'uaura',
        'Test undelegate',
        'test mnemonic phrase twelve words seed phrase here now today example'
      );

      expect(DirectSecp256k1HdWallet.fromMnemonic).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({ prefix: 'aura' })
      );
      expect(SigningStargateClient.connectWithSigner).toHaveBeenCalled();
      expect(mockClient.undelegateTokens).toHaveBeenCalledWith(
        'aura1delegator',
        'auravaloper1validator',
        { denom: 'uaura', amount: '1000000' },
        expect.objectContaining({
          amount: expect.arrayContaining([
            expect.objectContaining({ denom: 'uaura' })
          ]),
          gas: '200000'
        }),
        'Test undelegate'
      );
      expect(result).toEqual(mockResult);
    });

    test('should handle undelegate error', async () => {
      mockClient.undelegateTokens.mockRejectedValue(new Error('Insufficient delegation'));

      await expect(
        apiService.undelegate(
          'aura1delegator',
          'auravaloper1validator',
          1000000,
          'uaura',
          '',
          'test mnemonic'
        )
      ).rejects.toThrow('Insufficient delegation');
    });

    test('should handle invalid mnemonic error', async () => {
      DirectSecp256k1HdWallet.fromMnemonic.mockRejectedValue(
        new Error('Invalid mnemonic')
      );

      await expect(
        apiService.undelegate(
          'aura1delegator',
          'auravaloper1validator',
          1000000,
          'uaura',
          '',
          'invalid mnemonic'
        )
      ).rejects.toThrow();
    });
  });

  describe('withdrawRewards', () => {
    test('should successfully withdraw rewards', async () => {
      const mockResult = {
        code: 0,
        transactionHash: 'DEF456',
        height: 1001,
        gasUsed: 120000,
        gasWanted: 200000
      };

      mockClient.withdrawRewards.mockResolvedValue(mockResult);

      const result = await apiService.withdrawRewards(
        'aura1delegator',
        'auravaloper1validator',
        'Withdraw rewards',
        'test mnemonic phrase'
      );

      expect(mockClient.withdrawRewards).toHaveBeenCalledWith(
        'aura1delegator',
        'auravaloper1validator',
        expect.objectContaining({
          amount: expect.arrayContaining([
            expect.objectContaining({ denom: 'uaura' })
          ]),
          gas: '200000'
        }),
        'Withdraw rewards'
      );
      expect(result).toEqual(mockResult);
    });

    test('should handle withdraw rewards error', async () => {
      mockClient.withdrawRewards.mockRejectedValue(new Error('No rewards available'));

      await expect(
        apiService.withdrawRewards(
          'aura1delegator',
          'auravaloper1validator',
          '',
          'test mnemonic'
        )
      ).rejects.toThrow('No rewards available');
    });
  });

  describe('redelegate', () => {
    test('should successfully redelegate tokens', async () => {
      const mockResult = {
        code: 0,
        transactionHash: 'GHI789',
        height: 1002,
        gasUsed: 180000,
        gasWanted: 200000
      };

      mockClient.signAndBroadcast.mockResolvedValue(mockResult);

      const result = await apiService.redelegate(
        'aura1delegator',
        'auravaloper1src',
        'auravaloper1dst',
        500000,
        'uaura',
        'Test redelegate',
        'test mnemonic phrase'
      );

      expect(mockClient.signAndBroadcast).toHaveBeenCalledWith(
        'aura1delegator',
        expect.arrayContaining([
          expect.objectContaining({
            typeUrl: '/cosmos.staking.v1beta1.MsgBeginRedelegate',
            value: expect.objectContaining({
              delegatorAddress: 'aura1delegator',
              validatorSrcAddress: 'auravaloper1src',
              validatorDstAddress: 'auravaloper1dst',
              amount: { denom: 'uaura', amount: '500000' }
            })
          })
        ]),
        expect.objectContaining({
          amount: expect.arrayContaining([
            expect.objectContaining({ denom: 'uaura' })
          ]),
          gas: '200000'
        }),
        'Test redelegate'
      );
      expect(result).toEqual(mockResult);
    });

    test('should handle redelegate error', async () => {
      mockClient.signAndBroadcast.mockRejectedValue(
        new Error('Validator not found')
      );

      await expect(
        apiService.redelegate(
          'aura1delegator',
          'auravaloper1src',
          'auravaloper1invalid',
          500000,
          'uaura',
          '',
          'test mnemonic'
        )
      ).rejects.toThrow('Validator not found');
    });

    test('should handle network error during redelegate', async () => {
      SigningStargateClient.connectWithSigner.mockRejectedValue(
        new Error('Network error')
      );

      await expect(
        apiService.redelegate(
          'aura1delegator',
          'auravaloper1src',
          'auravaloper1dst',
          500000,
          'uaura',
          '',
          'test mnemonic'
        )
      ).rejects.toThrow();
    });
  });

  describe('vote', () => {
    test('should successfully vote yes on proposal', async () => {
      const mockResult = {
        code: 0,
        transactionHash: 'JKL012',
        height: 1003,
        gasUsed: 100000,
        gasWanted: 200000
      };

      mockClient.signAndBroadcast.mockResolvedValue(mockResult);

      const result = await apiService.vote(
        'aura1voter',
        '1',
        'yes',
        'Vote yes',
        'test mnemonic phrase'
      );

      expect(mockClient.signAndBroadcast).toHaveBeenCalledWith(
        'aura1voter',
        expect.arrayContaining([
          expect.objectContaining({
            typeUrl: '/cosmos.gov.v1beta1.MsgVote',
            value: expect.objectContaining({
              proposalId: BigInt(1),
              voter: 'aura1voter',
              option: 1 // yes
            })
          })
        ]),
        expect.any(Object),
        'Vote yes'
      );
      expect(result).toEqual(mockResult);
    });

    test('should successfully vote no on proposal', async () => {
      const mockResult = {
        code: 0,
        transactionHash: 'MNO345',
        height: 1004
      };

      mockClient.signAndBroadcast.mockResolvedValue(mockResult);

      await apiService.vote(
        'aura1voter',
        '2',
        'no',
        '',
        'test mnemonic'
      );

      expect(mockClient.signAndBroadcast).toHaveBeenCalledWith(
        'aura1voter',
        expect.arrayContaining([
          expect.objectContaining({
            typeUrl: '/cosmos.gov.v1beta1.MsgVote',
            value: expect.objectContaining({
              option: 3 // no
            })
          })
        ]),
        expect.any(Object),
        ''
      );
    });

    test('should successfully vote abstain on proposal', async () => {
      const mockResult = {
        code: 0,
        transactionHash: 'PQR678',
        height: 1005
      };

      mockClient.signAndBroadcast.mockResolvedValue(mockResult);

      await apiService.vote(
        'aura1voter',
        '3',
        'abstain',
        '',
        'test mnemonic'
      );

      expect(mockClient.signAndBroadcast).toHaveBeenCalledWith(
        'aura1voter',
        expect.arrayContaining([
          expect.objectContaining({
            value: expect.objectContaining({
              option: 2 // abstain
            })
          })
        ]),
        expect.any(Object),
        ''
      );
    });

    test('should successfully vote no_with_veto on proposal', async () => {
      const mockResult = {
        code: 0,
        transactionHash: 'STU901',
        height: 1006
      };

      mockClient.signAndBroadcast.mockResolvedValue(mockResult);

      await apiService.vote(
        'aura1voter',
        '4',
        'no_with_veto',
        '',
        'test mnemonic'
      );

      expect(mockClient.signAndBroadcast).toHaveBeenCalledWith(
        'aura1voter',
        expect.arrayContaining([
          expect.objectContaining({
            value: expect.objectContaining({
              option: 4 // no_with_veto
            })
          })
        ]),
        expect.any(Object),
        ''
      );
    });

    test('should accept numeric vote option', async () => {
      const mockResult = {
        code: 0,
        transactionHash: 'VWX234',
        height: 1007
      };

      mockClient.signAndBroadcast.mockResolvedValue(mockResult);

      await apiService.vote(
        'aura1voter',
        '5',
        1, // numeric yes
        '',
        'test mnemonic'
      );

      expect(mockClient.signAndBroadcast).toHaveBeenCalledWith(
        'aura1voter',
        expect.arrayContaining([
          expect.objectContaining({
            value: expect.objectContaining({
              option: 1
            })
          })
        ]),
        expect.any(Object),
        ''
      );
    });

    test('should handle invalid vote option', async () => {
      await expect(
        apiService.vote(
          'aura1voter',
          '6',
          'invalid',
          '',
          'test mnemonic'
        )
      ).rejects.toThrow('Invalid vote option');
    });

    test('should handle vote error', async () => {
      mockClient.signAndBroadcast.mockRejectedValue(
        new Error('Proposal not found')
      );

      await expect(
        apiService.vote(
          'aura1voter',
          '999',
          'yes',
          '',
          'test mnemonic'
        )
      ).rejects.toThrow('Proposal not found');
    });
  });

  describe('Integration - Multiple Operations', () => {
    test('should handle sequential staking operations', async () => {
      const mockDelegateResult = {
        code: 0,
        transactionHash: 'DELEGATE123',
        height: 1000
      };

      const mockWithdrawResult = {
        code: 0,
        transactionHash: 'WITHDRAW456',
        height: 1001
      };

      const mockUndelegateResult = {
        code: 0,
        transactionHash: 'UNDELEGATE789',
        height: 1002
      };

      mockClient.delegateTokens.mockResolvedValue(mockDelegateResult);
      mockClient.withdrawRewards.mockResolvedValue(mockWithdrawResult);
      mockClient.undelegateTokens.mockResolvedValue(mockUndelegateResult);

      // Delegate
      const delegateResult = await apiService.delegate(
        'aura1test',
        'auravaloper1val',
        1000000,
        'uaura',
        '',
        'test mnemonic'
      );
      expect(delegateResult.code).toBe(0);

      // Withdraw rewards
      const withdrawResult = await apiService.withdrawRewards(
        'aura1test',
        'auravaloper1val',
        '',
        'test mnemonic'
      );
      expect(withdrawResult.code).toBe(0);

      // Undelegate
      const undelegateResult = await apiService.undelegate(
        'aura1test',
        'auravaloper1val',
        500000,
        'uaura',
        '',
        'test mnemonic'
      );
      expect(undelegateResult.code).toBe(0);

      expect(mockClient.delegateTokens).toHaveBeenCalledTimes(1);
      expect(mockClient.withdrawRewards).toHaveBeenCalledTimes(1);
      expect(mockClient.undelegateTokens).toHaveBeenCalledTimes(1);
    });

    test('should handle vote and redelegate operations', async () => {
      const mockVoteResult = {
        code: 0,
        transactionHash: 'VOTE123',
        height: 1000
      };

      const mockRedelegateResult = {
        code: 0,
        transactionHash: 'REDELEGATE456',
        height: 1001
      };

      mockClient.signAndBroadcast
        .mockResolvedValueOnce(mockVoteResult)
        .mockResolvedValueOnce(mockRedelegateResult);

      // Vote
      const voteResult = await apiService.vote(
        'aura1test',
        '1',
        'yes',
        '',
        'test mnemonic'
      );
      expect(voteResult.code).toBe(0);

      // Redelegate
      const redelegateResult = await apiService.redelegate(
        'aura1test',
        'auravaloper1src',
        'auravaloper1dst',
        1000000,
        'uaura',
        '',
        'test mnemonic'
      );
      expect(redelegateResult.code).toBe(0);

      expect(mockClient.signAndBroadcast).toHaveBeenCalledTimes(2);
    });
  });

  describe('deposit', () => {
    test('should successfully deposit to proposal', async () => {
      const mockResult = {
        code: 0,
        transactionHash: 'DEPOSIT001',
        height: 1010
      };

      mockClient.signAndBroadcast.mockResolvedValue(mockResult);

      const result = await apiService.deposit(
        'aura1depositor',
        7,
        2500000,
        'uaura',
        'funding proposal',
        'test mnemonic words'
      );

      expect(mockClient.signAndBroadcast).toHaveBeenCalledWith(
        'aura1depositor',
        [
          expect.objectContaining({
            typeUrl: '/cosmos.gov.v1beta1.MsgDeposit',
            value: expect.objectContaining({
              proposalId: BigInt(7),
              depositor: 'aura1depositor',
              amount: [{ denom: 'uaura', amount: '2500000' }]
            })
          })
        ],
        expect.objectContaining({
          amount: expect.arrayContaining([
            expect.objectContaining({ denom: 'uaura', amount: '5000' })
          ]),
          gas: '200000'
        }),
        'funding proposal'
      );
      expect(result).toEqual(mockResult);
    });

    test('should validate deposit inputs', async () => {
      await expect(
        apiService.deposit('', 1, 10, 'uaura', '', 'mnemonic words')
      ).rejects.toThrow('Invalid depositor address');

      await expect(
        apiService.deposit('aura1depositor', 0, 10, 'uaura', '', 'mnemonic')
      ).rejects.toThrow('Invalid proposal id');

      await expect(
        apiService.deposit('aura1depositor', 1, 0, 'uaura', '', 'mnemonic')
      ).rejects.toThrow('Deposit amount must be greater than zero');

      await expect(
        apiService.deposit('aura1depositor', 1, 10, 'uaura', '', '')
      ).rejects.toThrow('Missing signing key');
    });
  });
});
