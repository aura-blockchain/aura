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
      ).rejects.toThrow('Invalid delegator address');
    });

    test('should reject zero or negative undelegate amount', async () => {
      await expect(
        TransactionService.undelegate({
          delegatorAddress: 'aura1delegator',
          validatorAddress: 'auravaloper1validator',
          amount: 0,
          denom: 'uaura',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Undelegate amount must be greater than zero');
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

    test('should reject missing reward addresses', async () => {
      await expect(
        TransactionService.withdrawRewards({
          delegatorAddress: '',
          validatorAddress: '',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Invalid delegator address');
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

    test('should reject identical validators during redelegation', async () => {
      await expect(
        TransactionService.redelegate({
          delegatorAddress: 'aura1delegator',
          srcValidatorAddress: 'auravaloper1same',
          dstValidatorAddress: 'auravaloper1same',
          amount: 500000,
          denom: 'uaura',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Source and destination validators must be different');
    });

    test('should reject non-positive redelegate amount', async () => {
      await expect(
        TransactionService.redelegate({
          delegatorAddress: 'aura1delegator',
          srcValidatorAddress: 'auravaloper1src',
          dstValidatorAddress: 'auravaloper1dst',
          amount: -1,
          denom: 'uaura',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Redelegate amount must be greater than zero');
    });
  });

  describe('DEX transactions', () => {
    test('createDexPool builds correct message payload', async () => {
      const mockBroadcastResult = {code: 0, txhash: 'POOL123'};
      PawAPI.broadcastTx = jest.fn().mockResolvedValue(mockBroadcastResult);

      const result = await TransactionService.createDexPool({
        creatorAddress: 'aura1creator',
        denomA: 'uaura',
        denomB: 'usdt',
        amountA: 1_000_000,
        amountB: 500_000,
        memo: 'create pool',
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
                type_url: '/aura.dex.v1beta1.MsgCreatePool',
                value: expect.objectContaining({
                  creator: 'aura1creator',
                  denom_a: 'uaura',
                  denom_b: 'usdt',
                  amount_a: {denom: 'uaura', amount: '1000000'},
                  amount_b: {denom: 'usdt', amount: '500000'},
                }),
              }),
            ]),
          }),
        }),
        'sync',
      );
      expect(result).toEqual(mockBroadcastResult);
    });

    test('swapDexExactIn defaults minAmountOut and clamps slippage', async () => {
      PawAPI.broadcastTx = jest.fn().mockResolvedValue({code: 0});

      await TransactionService.swapDexExactIn({
        senderAddress: 'aura1trader',
        poolId: '42',
        denomIn: 'uaura',
        amountIn: 1_500_000,
        minAmountOut: null,
        maxSlippageBps: 75,
        memo: 'swap now',
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
                type_url: '/aura.dex.v1beta1.MsgSwapExactIn',
                value: expect.objectContaining({
                  pool_id: '42',
                  coin_in: {denom: 'uaura', amount: '1500000'},
                  min_amount_out: '1500000',
                  max_slippage_bps: 75,
                }),
              }),
            ]),
          }),
        }),
        'sync',
      );
    });

    test('commitDexOrder normalizes commit hash to base64', async () => {
      PawAPI.broadcastTx = jest.fn().mockResolvedValue({code: 0});
      const expectedBytes = Buffer.from('aabbccdd', 'hex').toString('base64');

      await TransactionService.commitDexOrder({
        senderAddress: 'aura1sender',
        commitHash: '0xaabbccdd',
        memo: 'commit',
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
                type_url: '/aura.dex.v1beta1.MsgCommitOrder',
                value: expect.objectContaining({
                  sender: 'aura1sender',
                  commit_hash: expectedBytes,
                }),
              }),
            ]),
          }),
        }),
        'sync',
      );
    });

    test('createDexHTLC builds correct HTLC payload', async () => {
      const mockBroadcastResult = {code: 0, txhash: 'HTLC123'};
      PawAPI.broadcastTx = jest.fn().mockResolvedValue(mockBroadcastResult);

      const result = await TransactionService.createDexHTLC({
        senderAddress: 'aura1sender',
        recipientAddress: 'aura1recipient',
        denom: 'uaura',
        amount: 750000,
        secretHash: 'secret-hash',
        timelockDuration: 3600,
        memo: 'lock funds',
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
                type_url: '/aura.dex.v1beta1.MsgCreateHTLC',
                value: expect.objectContaining({
                  sender: 'aura1sender',
                  recipient: 'aura1recipient',
                  amount: {denom: 'uaura', amount: '750000'},
                  secret_hash: 'secret-hash',
                  timelock_duration: '3600',
                }),
              }),
            ]),
            memo: 'lock funds',
          }),
        }),
        'sync',
      );
      expect(result).toEqual(mockBroadcastResult);
    });

    test('claimDexHTLC requires valid identifiers and secrets', async () => {
      PawAPI.broadcastTx = jest.fn().mockResolvedValue({code: 0});

      await TransactionService.claimDexHTLC({
        recipientAddress: 'aura1recipient',
        htlcId: 42,
        secret: 'unseal-secret',
        memo: 'claim',
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
                type_url: '/aura.dex.v1beta1.MsgClaimHTLC',
                value: expect.objectContaining({
                  recipient: 'aura1recipient',
                  htlc_id: '42',
                  secret: 'unseal-secret',
                }),
              }),
            ]),
          }),
        }),
        'sync',
      );

      await expect(
        TransactionService.claimDexHTLC({
          recipientAddress: 'invalid',
          htlcId: '',
          secret: '',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        }),
      ).rejects.toThrow('Invalid recipient address');
    });

    test('refundDexHTLC validates sender and HTLC id', async () => {
      PawAPI.broadcastTx = jest.fn().mockResolvedValue({code: 0});

      await TransactionService.refundDexHTLC({
        senderAddress: 'aura1sender',
        htlcId: 84,
        memo: 'refund',
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
                type_url: '/aura.dex.v1beta1.MsgRefundHTLC',
                value: expect.objectContaining({
                  sender: 'aura1sender',
                  htlc_id: '84',
                }),
              }),
            ]),
            memo: 'refund',
          }),
        }),
        'sync',
      );

      await expect(
        TransactionService.refundDexHTLC({
          senderAddress: '',
          htlcId: null,
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        }),
      ).rejects.toThrow('Invalid sender address');
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

    test('should reject missing voter address or proposal id', async () => {
      await expect(
        TransactionService.vote({
          voterAddress: '',
          proposalId: '',
          option: 'yes',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Invalid voter address');

      await expect(
        TransactionService.vote({
          voterAddress: 'aura1voter',
          proposalId: '',
          option: 'yes',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Proposal ID is required');

      await expect(
        TransactionService.vote({
          voterAddress: 'aura1voter',
          proposalId: '1',
          option: 'yes',
          memo: '',
          privateKeyHex: '',
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Private key is required');
    });
  });

  describe('deposit', () => {
    test('should successfully deposit to proposal', async () => {
      const mockBroadcastResult = {
        code: 0,
        txhash: 'DEP123',
        height: '2001',
      };

      PawAPI.broadcastTx = jest.fn().mockResolvedValue(mockBroadcastResult);

      const result = await TransactionService.deposit({
        depositorAddress: 'aura1depositor',
        proposalId: 9,
        amount: 2500000,
        denom: 'uaura',
        memo: 'fund proposal',
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
                type_url: '/cosmos.gov.v1beta1.MsgDeposit',
                value: expect.objectContaining({
                  proposal_id: '9',
                  depositor: 'aura1depositor',
                  amount: [{denom: 'uaura', amount: '2500000'}],
                }),
              }),
            ]),
            memo: 'fund proposal',
          }),
        }),
        'sync'
      );
      expect(result).toEqual(mockBroadcastResult);
    });

    test('should validate deposit inputs', async () => {
      await expect(
        TransactionService.deposit({
          depositorAddress: '',
          proposalId: 1,
          amount: 10,
          denom: 'uaura',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Invalid depositor address');

      await expect(
        TransactionService.deposit({
          depositorAddress: 'aura1depositor',
          proposalId: 0,
          amount: 10,
          denom: 'uaura',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Invalid proposal id');

      await expect(
        TransactionService.deposit({
          depositorAddress: 'aura1depositor',
          proposalId: 2,
          amount: 0,
          denom: 'uaura',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Deposit amount must be greater than zero');

      await expect(
        TransactionService.deposit({
          depositorAddress: 'aura1depositor',
          proposalId: 2,
          amount: 10,
          denom: 'uaura',
          memo: '',
          privateKeyHex: '',
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Private key is required');
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

    test('should reject invalid delegate inputs', async () => {
      await expect(
        TransactionService.delegate({
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
      ).rejects.toThrow('Invalid delegator address');

      await expect(
        TransactionService.delegate({
          delegatorAddress: 'aura1delegator',
          validatorAddress: '',
          amount: 1000000,
          denom: 'uaura',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Invalid validator address');

      await expect(
        TransactionService.delegate({
          delegatorAddress: 'aura1delegator',
          validatorAddress: 'auravaloper1validator',
          amount: 0,
          denom: 'uaura',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Delegate amount must be greater than zero');

      await expect(
        TransactionService.delegate({
          delegatorAddress: 'aura1delegator',
          validatorAddress: 'auravaloper1validator',
          amount: 1,
          denom: 'uaura',
          memo: '',
          privateKeyHex: '',
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Private key is required');
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

    test('should reject invalid send inputs', async () => {
      await expect(
        TransactionService.sendTokens({
          fromAddress: '',
          toAddress: 'aura1recipient',
          amount: 1,
          denom: 'uaura',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Invalid sender address');

      await expect(
        TransactionService.sendTokens({
          fromAddress: 'aura1sender',
          toAddress: '',
          amount: 1,
          denom: 'uaura',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Invalid recipient address');

      await expect(
        TransactionService.sendTokens({
          fromAddress: 'aura1sender',
          toAddress: 'aura1recipient',
          amount: 0,
          denom: 'uaura',
          memo: '',
          privateKeyHex: mockPrivateKey,
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Send amount must be greater than zero');

      await expect(
        TransactionService.sendTokens({
          fromAddress: 'aura1sender',
          toAddress: 'aura1recipient',
          amount: 1,
          denom: 'uaura',
          memo: '',
          privateKeyHex: '',
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
        })
      ).rejects.toThrow('Private key is required');
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

  describe('createSignedTransaction validation', () => {
    test('should reject missing private key', async () => {
      await expect(
        TransactionService.createSignedTransaction({
          messages: [{}],
          fee: {},
          memo: '',
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
          privateKeyHex: '',
        })
      ).rejects.toThrow('Private key is required');
    });

    test('should reject empty message list', async () => {
      await expect(
        TransactionService.createSignedTransaction({
          messages: [],
          fee: {},
          memo: '',
          accountNumber: mockAccountNumber,
          sequence: mockSequence,
          chainId: mockChainId,
          privateKeyHex: mockPrivateKey,
        })
      ).rejects.toThrow('At least one message is required');
    });

    test('should reject missing chain context', async () => {
      await expect(
        TransactionService.createSignedTransaction({
          messages: [{}],
          fee: {},
          memo: '',
          accountNumber: undefined,
          sequence: undefined,
          chainId: '',
          privateKeyHex: mockPrivateKey,
        })
      ).rejects.toThrow('Account number and sequence are required');
    });
  });
});
