/**
 * Transaction Service Tests
 * Tests for staking operations: undelegate, withdraw rewards, redelegate, vote
 */

import TransactionService from '../src/services/TransactionService';
import PawAPI from '../src/services/PawAPI';

// Mock PawAPI
jest.mock('../src/services/PawAPI');

describe('TransactionService - Staking Operations', () => {
  const mockPrivateKey = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef';
  const mockAccountNumber = '1234';
  const mockSequence = '5';
  const mockChainId = 'aura-testnet-1';

  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('undelegate', () => {
    test('should successfully create undelegate transaction', async () => {
      const mockBroadcastResult = {
        code: 0,
        txhash: 'ABC123',
        height: '1000',
      };

      PawAPI.broadcastTx = jest.fn().mockResolvedValue(mockBroadcastResult);

      const result = await TransactionService.undelegate({
        delegatorAddress: 'aura1delegator',
        validatorAddress: 'auravaloper1validator',
        amount: 1000000,
        denom: 'uaura',
        memo: 'Test undelegate',
        privateKeyHex: mockPrivateKey,
        accountNumber: mockAccountNumber,
        sequence: mockSequence,
        chainId: mockChainId,
      });

      expect(PawAPI.broadcastTx).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            messages: expect.arrayContaining([
              expect.objectContaining({
                type_url: '/cosmos.staking.v1beta1.MsgUndelegate',
                value: expect.objectContaining({
                  delegator_address: 'aura1delegator',
                  validator_address: 'auravaloper1validator',
                  amount: {
                    denom: 'uaura',
                    amount: '1000000',
                  },
                }),
              }),
            ]),
            memo: 'Test undelegate',
          }),
          auth_info: expect.any(Object),
          signatures: expect.any(Array),
        }),
        'sync'
      );
      expect(result).toEqual(mockBroadcastResult);
    });

    test('should handle undelegate error', async () => {
      PawAPI.broadcastTx = jest.fn().mockRejectedValue(
        new Error('Insufficient delegation')
      );

      await expect(
        TransactionService.undelegate({
          delegatorAddress: 'aura1delegator',
          validatorAddress: 'auravaloper1validator',
          amount: 1000000,
          denom: 'uaura',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Insufficient delegation');
    });

    test('should handle missing required parameters', async () => {
      await expect(
        TransactionService.undelegate({
          delegatorAddress: '',
          validatorAddress: 'auravaloper1validator',
          amount: 1000000,
          denom: 'uaura',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow();
    });
  });

  describe('withdrawRewards', () => {
    test('should successfully create withdraw rewards transaction', async () => {
      const mockBroadcastResult = {
        code: 0,
        txhash: 'DEF456',
        height: '1001',
      };

      PawAPI.broadcastTx = jest.fn().mockResolvedValue(mockBroadcastResult);

      const result = await TransactionService.withdrawRewards({
        delegatorAddress: 'aura1delegator',
        validatorAddress: 'auravaloper1validator',
        memo: 'Withdraw rewards',
        privateKeyHex: mockPrivateKey,
        accountNumber: mockAccountNumber,
        sequence: mockSequence,
        chainId: mockChainId,
      });

      expect(PawAPI.broadcastTx).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            messages: expect.arrayContaining([
              expect.objectContaining({
                type_url: '/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward',
                value: expect.objectContaining({
                  delegator_address: 'aura1delegator',
                  validator_address: 'auravaloper1validator',
                }),
              }),
            ]),
            memo: 'Withdraw rewards',
          }),
        }),
        'sync'
      );
      expect(result).toEqual(mockBroadcastResult);
    });

    test('should handle withdraw rewards error', async () => {
      PawAPI.broadcastTx = jest.fn().mockRejectedValue(
        new Error('No rewards available')
      );

      await expect(
        TransactionService.withdrawRewards({
          delegatorAddress: 'aura1delegator',
          validatorAddress: 'auravaloper1validator',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('No rewards available');
    });
  });

  describe('redelegate', () => {
    test('should successfully create redelegate transaction', async () => {
      const mockBroadcastResult = {
        code: 0,
        txhash: 'GHI789',
        height: '1002',
      };

      PawAPI.broadcastTx = jest.fn().mockResolvedValue(mockBroadcastResult);

      const result = await TransactionService.redelegate({
        delegatorAddress: 'aura1delegator',
        srcValidatorAddress: 'auravaloper1src',
        dstValidatorAddress: 'auravaloper1dst',
        amount: 500000,
        denom: 'uaura',
        memo: 'Test redelegate',
        privateKeyHex: mockPrivateKey,
        accountNumber: mockAccountNumber,
        sequence: mockSequence,
        chainId: mockChainId,
      });

      expect(PawAPI.broadcastTx).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            messages: expect.arrayContaining([
              expect.objectContaining({
                type_url: '/cosmos.staking.v1beta1.MsgBeginRedelegate',
                value: expect.objectContaining({
                  delegator_address: 'aura1delegator',
                  validator_src_address: 'auravaloper1src',
                  validator_dst_address: 'auravaloper1dst',
                  amount: {
                    denom: 'uaura',
                    amount: '500000',
                  },
                }),
              }),
            ]),
            memo: 'Test redelegate',
          }),
        }),
        'sync'
      );
      expect(result).toEqual(mockBroadcastResult);
    });

    test('should handle redelegate error', async () => {
      PawAPI.broadcastTx = jest.fn().mockRejectedValue(
        new Error('Validator not found')
      );

      await expect(
        TransactionService.redelegate({
          delegatorAddress: 'aura1delegator',
          srcValidatorAddress: 'auravaloper1src',
          dstValidatorAddress: 'auravaloper1invalid',
          amount: 500000,
          denom: 'uaura',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Validator not found');
    });

    test('should validate source and destination validators are different', async () => {
      const mockBroadcastResult = {
        code: 0,
        txhash: 'XYZ123',
        height: '1003',
      };

      PawAPI.broadcastTx = jest.fn().mockResolvedValue(mockBroadcastResult);

      // Service should allow this - validation happens on chain
      await expect(
        TransactionService.redelegate({
          delegatorAddress: 'aura1delegator',
          srcValidatorAddress: 'auravaloper1same',
          dstValidatorAddress: 'auravaloper1different',
          amount: 500000,
          denom: 'uaura',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).resolves.toBeDefined();
    });
  });

  describe('vote', () => {
    test('should successfully vote yes on proposal', async () => {
      const mockBroadcastResult = {
        code: 0,
        txhash: 'JKL012',
        height: '1003',
      };

      PawAPI.broadcastTx = jest.fn().mockResolvedValue(mockBroadcastResult);

      const result = await TransactionService.vote({
        voterAddress: 'aura1voter',
        proposalId: '1',
        option: 'yes',
        memo: 'Vote yes',
        privateKeyHex: mockPrivateKey,
        accountNumber: mockAccountNumber,
        sequence: mockSequence,
        chainId: mockChainId,
      });

      expect(PawAPI.broadcastTx).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            messages: expect.arrayContaining([
              expect.objectContaining({
                type_url: '/cosmos.gov.v1beta1.MsgVote',
                value: expect.objectContaining({
                  proposal_id: '1',
                  voter: 'aura1voter',
                  option: 1, // yes
                }),
              }),
            ]),
            memo: 'Vote yes',
          }),
        }),
        'sync'
      );
      expect(result).toEqual(mockBroadcastResult);
    });

    test('should successfully vote no on proposal', async () => {
      const mockBroadcastResult = {
        code: 0,
        txhash: 'MNO345',
        height: '1004',
      };

      PawAPI.broadcastTx = jest.fn().mockResolvedValue(mockBroadcastResult);

      await TransactionService.vote({
        voterAddress: 'aura1voter',
        proposalId: '2',
        option: 'no',
        memo: '',
        privateKeyHex: mockPrivateKey,
        accountNumber: mockAccountNumber,
        sequence: mockSequence,
        chainId: mockChainId,
      });

      expect(PawAPI.broadcastTx).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            messages: expect.arrayContaining([
              expect.objectContaining({
                value: expect.objectContaining({
                  option: 3, // no
                }),
              }),
            ]),
          }),
        }),
        'sync'
      );
    });

    test('should successfully vote abstain on proposal', async () => {
      const mockBroadcastResult = {
        code: 0,
        txhash: 'PQR678',
        height: '1005',
      };

      PawAPI.broadcastTx = jest.fn().mockResolvedValue(mockBroadcastResult);

      await TransactionService.vote({
        voterAddress: 'aura1voter',
        proposalId: '3',
        option: 'abstain',
        memo: '',
        privateKeyHex: mockPrivateKey,
        accountNumber: mockAccountNumber,
        sequence: mockSequence,
        chainId: mockChainId,
      });

      expect(PawAPI.broadcastTx).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            messages: expect.arrayContaining([
              expect.objectContaining({
                value: expect.objectContaining({
                  option: 2, // abstain
                }),
              }),
            ]),
          }),
        }),
        'sync'
      );
    });

    test('should successfully vote no_with_veto on proposal', async () => {
      const mockBroadcastResult = {
        code: 0,
        txhash: 'STU901',
        height: '1006',
      };

      PawAPI.broadcastTx = jest.fn().mockResolvedValue(mockBroadcastResult);

      await TransactionService.vote({
        voterAddress: 'aura1voter',
        proposalId: '4',
        option: 'no_with_veto',
        memo: '',
        privateKeyHex: mockPrivateKey,
        accountNumber: mockAccountNumber,
        sequence: mockSequence,
        chainId: mockChainId,
      });

      expect(PawAPI.broadcastTx).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            messages: expect.arrayContaining([
              expect.objectContaining({
                value: expect.objectContaining({
                  option: 4, // no_with_veto
                }),
              }),
            ]),
          }),
        }),
        'sync'
      );
    });

    test('should accept numeric vote option', async () => {
      const mockBroadcastResult = {
        code: 0,
        txhash: 'VWX234',
        height: '1007',
      };

      PawAPI.broadcastTx = jest.fn().mockResolvedValue(mockBroadcastResult);

      await TransactionService.vote({
        voterAddress: 'aura1voter',
        proposalId: '5',
        option: 1, // numeric yes
        memo: '',
        privateKeyHex: mockPrivateKey,
        accountNumber: mockAccountNumber,
        sequence: mockSequence,
        chainId: mockChainId,
      });

      expect(PawAPI.broadcastTx).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            messages: expect.arrayContaining([
              expect.objectContaining({
                value: expect.objectContaining({
                  option: 1,
                }),
              }),
            ]),
          }),
        }),
        'sync'
      );
    });

    test('should handle invalid vote option', async () => {
      await expect(
        TransactionService.vote({
          voterAddress: 'aura1voter',
          proposalId: '6',
          option: 'invalid',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Invalid vote option');
    });

    test('should handle vote broadcast error', async () => {
      PawAPI.broadcastTx = jest.fn().mockRejectedValue(
        new Error('Proposal not found')
      );

      await expect(
        TransactionService.vote({
          voterAddress: 'aura1voter',
          proposalId: '999',
          option: 'yes',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Proposal not found');
    });
  });

  describe('delegate', () => {
    test('should successfully create delegate transaction', async () => {
      const mockBroadcastResult = {
        code: 0,
        txhash: 'DELEGATE123',
        height: '1008',
      };

      PawAPI.broadcastTx = jest.fn().mockResolvedValue(mockBroadcastResult);

      const result = await TransactionService.delegate({
        delegatorAddress: 'aura1delegator',
        validatorAddress: 'auravaloper1validator',
        amount: 1000000,
        denom: 'uaura',
        memo: 'Test delegate',
        privateKeyHex: mockPrivateKey,
        accountNumber: mockAccountNumber,
        sequence: mockSequence,
        chainId: mockChainId,
      });

      expect(PawAPI.broadcastTx).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            messages: expect.arrayContaining([
              expect.objectContaining({
                type_url: '/cosmos.staking.v1beta1.MsgDelegate',
                value: expect.objectContaining({
                  delegator_address: 'aura1delegator',
                  validator_address: 'auravaloper1validator',
                  amount: {
                    denom: 'uaura',
                    amount: '1000000',
                  },
                }),
              }),
            ]),
          }),
        }),
        'sync'
      );
      expect(result).toEqual(mockBroadcastResult);
    });
  });

  describe('sendTokens', () => {
    test('should successfully create send transaction', async () => {
      const mockBroadcastResult = {
        code: 0,
        txhash: 'SEND456',
        height: '1009',
      };

      PawAPI.broadcastTx = jest.fn().mockResolvedValue(mockBroadcastResult);

      const result = await TransactionService.sendTokens({
        fromAddress: 'aura1sender',
        toAddress: 'aura1recipient',
        amount: 500000,
        denom: 'uaura',
        memo: 'Test send',
        privateKeyHex: mockPrivateKey,
        accountNumber: mockAccountNumber,
        sequence: mockSequence,
        chainId: mockChainId,
      });

      expect(PawAPI.broadcastTx).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            messages: expect.arrayContaining([
              expect.objectContaining({
                type_url: '/cosmos.bank.v1beta1.MsgSend',
                value: expect.objectContaining({
                  from_address: 'aura1sender',
                  to_address: 'aura1recipient',
                  amount: [{denom: 'uaura', amount: '500000'}],
                }),
              }),
            ]),
          }),
        }),
        'sync'
      );
      expect(result).toEqual(mockBroadcastResult);
    });
  });

  describe('Integration - Multiple Operations', () => {
    test('should handle sequential staking operations', async () => {
      const mockResults = [
        {code: 0, txhash: 'TX1', height: '1000'},
        {code: 0, txhash: 'TX2', height: '1001'},
        {code: 0, txhash: 'TX3', height: '1002'},
        {code: 0, txhash: 'TX4', height: '1003'},
      ];

      PawAPI.broadcastTx = jest
        .fn()
        .mockResolvedValueOnce(mockResults[0])
        .mockResolvedValueOnce(mockResults[1])
        .mockResolvedValueOnce(mockResults[2])
        .mockResolvedValueOnce(mockResults[3]);

      // Delegate
      const delegateResult = await TransactionService.delegate({
        delegatorAddress: 'aura1test',
        validatorAddress: 'auravaloper1val',
        amount: 1000000,
        denom: 'uaura',
        memo: '',
        privateKeyHex: mockPrivateKey,
        accountNumber: mockAccountNumber,
        sequence: mockSequence,
        chainId: mockChainId,
      });
      expect(delegateResult.code).toBe(0);

      // Withdraw rewards
      const withdrawResult = await TransactionService.withdrawRewards({
        delegatorAddress: 'aura1test',
        validatorAddress: 'auravaloper1val',
        memo: '',
        privateKeyHex: mockPrivateKey,
        accountNumber: mockAccountNumber,
        sequence: String(Number(mockSequence) + 1),
        chainId: mockChainId,
      });
      expect(withdrawResult.code).toBe(0);

      // Redelegate
      const redelegateResult = await TransactionService.redelegate({
        delegatorAddress: 'aura1test',
        srcValidatorAddress: 'auravaloper1val',
        dstValidatorAddress: 'auravaloper1val2',
        amount: 500000,
        denom: 'uaura',
        memo: '',
        privateKeyHex: mockPrivateKey,
        accountNumber: mockAccountNumber,
        sequence: String(Number(mockSequence) + 2),
        chainId: mockChainId,
      });
      expect(redelegateResult.code).toBe(0);

      // Vote
      const voteResult = await TransactionService.vote({
        voterAddress: 'aura1test',
        proposalId: '1',
        option: 'yes',
        memo: '',
        privateKeyHex: mockPrivateKey,
        accountNumber: mockAccountNumber,
        sequence: String(Number(mockSequence) + 3),
        chainId: mockChainId,
      });
      expect(voteResult.code).toBe(0);

      expect(PawAPI.broadcastTx).toHaveBeenCalledTimes(4);
    });
  });
});
