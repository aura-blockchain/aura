/**
 * Unit Tests for Staking Module
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';

// Mock COSMOS_SDK
global.COSMOS_SDK = {
  config: {
    bech32Prefix: 'aura',
    coinDenom: 'uaura',
    coinDecimals: 6,
    restEndpoint: 'http://localhost:1317',
  },
};

// Mock fetch
global.fetch = vi.fn();

describe('StakingModule', () => {
  let StakingModule;

  beforeEach(async () => {
    StakingModule = (await import('../../src/staking.js')).default || (await import('../../src/staking.js'));
    vi.clearAllMocks();
  });

  describe('buildDelegateTx', () => {
    it('should build correct delegate transaction', () => {
      const params = {
        delegatorAddress: 'aura1test',
        validatorAddress: 'auravaloper1test',
        amount: 1000000,
        denom: 'uaura',
      };

      const tx = StakingModule.buildDelegateTx(params);

      expect(tx.body.messages).toHaveLength(1);
      expect(tx.body.messages[0]['@type']).toBe('/cosmos.staking.v1beta1.MsgDelegate');
      expect(tx.body.messages[0].delegator_address).toBe(params.delegatorAddress);
      expect(tx.body.messages[0].validator_address).toBe(params.validatorAddress);
      expect(tx.body.messages[0].amount.amount).toBe('1000000');
      expect(tx.body.messages[0].amount.denom).toBe('uaura');
    });

    it('should use default fee and gas', () => {
      const params = {
        delegatorAddress: 'aura1test',
        validatorAddress: 'auravaloper1test',
        amount: 1000000,
      };

      const tx = StakingModule.buildDelegateTx(params);

      expect(tx.auth_info.fee).toBeDefined();
      expect(tx.auth_info.fee.gas).toBe('250000');
    });
  });

  describe('buildUndelegateTx', () => {
    it('should build correct undelegate transaction', () => {
      const params = {
        delegatorAddress: 'aura1test',
        validatorAddress: 'auravaloper1test',
        amount: 500000,
      };

      const tx = StakingModule.buildUndelegateTx(params);

      expect(tx.body.messages).toHaveLength(1);
      expect(tx.body.messages[0]['@type']).toBe('/cosmos.staking.v1beta1.MsgUndelegate');
      expect(tx.body.messages[0].delegator_address).toBe(params.delegatorAddress);
      expect(tx.body.messages[0].validator_address).toBe(params.validatorAddress);
      expect(tx.body.messages[0].amount.amount).toBe('500000');
    });
  });

  describe('buildRedelegateTx', () => {
    it('should build correct redelegate transaction', () => {
      const params = {
        delegatorAddress: 'aura1test',
        validatorSrcAddress: 'auravaloper1test1',
        validatorDstAddress: 'auravaloper1test2',
        amount: 750000,
      };

      const tx = StakingModule.buildRedelegateTx(params);

      expect(tx.body.messages).toHaveLength(1);
      expect(tx.body.messages[0]['@type']).toBe('/cosmos.staking.v1beta1.MsgBeginRedelegate');
      expect(tx.body.messages[0].delegator_address).toBe(params.delegatorAddress);
      expect(tx.body.messages[0].validator_src_address).toBe(params.validatorSrcAddress);
      expect(tx.body.messages[0].validator_dst_address).toBe(params.validatorDstAddress);
    });
  });

  describe('buildWithdrawRewardsTx', () => {
    it('should build correct withdraw rewards transaction', () => {
      const params = {
        delegatorAddress: 'aura1test',
        validatorAddress: 'auravaloper1test',
      };

      const tx = StakingModule.buildWithdrawRewardsTx(params);

      expect(tx.body.messages).toHaveLength(1);
      expect(tx.body.messages[0]['@type']).toBe('/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward');
      expect(tx.body.messages[0].delegator_address).toBe(params.delegatorAddress);
      expect(tx.body.messages[0].validator_address).toBe(params.validatorAddress);
    });
  });

  describe('buildWithdrawAllRewardsTx', () => {
    it('should build transaction for multiple validators', () => {
      const params = {
        delegatorAddress: 'aura1test',
        validatorAddresses: ['auravaloper1test1', 'auravaloper1test2', 'auravaloper1test3'],
      };

      const tx = StakingModule.buildWithdrawAllRewardsTx(params);

      expect(tx.body.messages).toHaveLength(3);
      tx.body.messages.forEach((msg, index) => {
        expect(msg['@type']).toBe('/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward');
        expect(msg.delegator_address).toBe(params.delegatorAddress);
        expect(msg.validator_address).toBe(params.validatorAddresses[index]);
      });
    });
  });

  describe('queryDelegations', () => {
    it('should query delegations successfully', async () => {
      const mockDelegations = [
        {
          delegation: { delegator_address: 'aura1test', validator_address: 'auravaloper1test1' },
          balance: { denom: 'uaura', amount: '1000000' },
        },
        {
          delegation: { delegator_address: 'aura1test', validator_address: 'auravaloper1test2' },
          balance: { denom: 'uaura', amount: '2000000' },
        },
      ];

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ delegation_responses: mockDelegations }),
      });

      const delegations = await StakingModule.queryDelegations('aura1test');

      expect(delegations).toEqual(mockDelegations);
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/cosmos/staking/v1beta1/delegations/aura1test')
      );
    });

    it('should handle query errors', async () => {
      fetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
      });

      await expect(StakingModule.queryDelegations('aura1test')).rejects.toThrow();
    });

    it('should return empty array if no delegations', async () => {
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      });

      const delegations = await StakingModule.queryDelegations('aura1test');

      expect(delegations).toEqual([]);
    });
  });

  describe('queryRewards', () => {
    it('should query rewards for specific validator', async () => {
      const mockRewards = {
        rewards: [{ denom: 'uaura', amount: '123456' }],
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockRewards,
      });

      const rewards = await StakingModule.queryRewards('aura1test', 'auravaloper1test');

      expect(rewards).toEqual(mockRewards);
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/delegators/aura1test/rewards/auravaloper1test')
      );
    });

    it('should query rewards for all validators', async () => {
      const mockRewards = {
        total: [{ denom: 'uaura', amount: '500000' }],
        rewards: [
          { validator_address: 'auravaloper1test1', reward: [{ denom: 'uaura', amount: '250000' }] },
          { validator_address: 'auravaloper1test2', reward: [{ denom: 'uaura', amount: '250000' }] },
        ],
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockRewards,
      });

      const rewards = await StakingModule.queryRewards('aura1test');

      expect(rewards).toEqual(mockRewards);
      expect(fetch).toHaveBeenCalledWith(
        expect.stringMatching(/\/delegators\/aura1test\/rewards$/)
      );
    });
  });

  describe('queryValidators', () => {
    it('should query bonded validators', async () => {
      const mockValidators = [
        { operator_address: 'auravaloper1test1', status: 'BOND_STATUS_BONDED' },
        { operator_address: 'auravaloper1test2', status: 'BOND_STATUS_BONDED' },
      ];

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ validators: mockValidators }),
      });

      const validators = await StakingModule.queryValidators('BOND_STATUS_BONDED');

      expect(validators).toEqual(mockValidators);
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('status=BOND_STATUS_BONDED')
      );
    });
  });

  describe('calculateTotalStaked', () => {
    it('should calculate total staked amount correctly', () => {
      const delegations = [
        { balance: { amount: '1000000' } },
        { balance: { amount: '2000000' } },
        { balance: { amount: '3000000' } },
      ];

      const total = StakingModule.calculateTotalStaked(delegations);

      expect(total).toBe(6000000);
    });

    it('should handle empty delegations', () => {
      const total = StakingModule.calculateTotalStaked([]);
      expect(total).toBe(0);
    });

    it('should handle missing balance fields', () => {
      const delegations = [
        { balance: { amount: '1000000' } },
        {},
        { balance: {} },
      ];

      const total = StakingModule.calculateTotalStaked(delegations);

      expect(total).toBe(1000000);
    });
  });

  describe('calculateTotalRewards', () => {
    it('should calculate total rewards correctly', () => {
      const rewardsData = {
        total: [
          { denom: 'uaura', amount: '123456.789' },
          { denom: 'stake', amount: '1000' },
        ],
      };

      const total = StakingModule.calculateTotalRewards(rewardsData);

      expect(total).toBeCloseTo(123456.789, 2);
    });

    it('should handle missing total field', () => {
      const total = StakingModule.calculateTotalRewards({});
      expect(total).toBe(0);
    });
  });

  describe('validateValidatorAddress', () => {
    it('should validate correct validator addresses', () => {
      expect(StakingModule.validateValidatorAddress('auravaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqjkvhyq')).toBe(true);
    });

    it('should reject invalid validator addresses', () => {
      expect(StakingModule.validateValidatorAddress('aura1test')).toBe(false);
      expect(StakingModule.validateValidatorAddress('cosmosvaloper1test')).toBe(false);
      expect(StakingModule.validateValidatorAddress('auravaloper1')).toBe(false);
      expect(StakingModule.validateValidatorAddress('')).toBe(false);
      expect(StakingModule.validateValidatorAddress(null)).toBe(false);
    });
  });
});
